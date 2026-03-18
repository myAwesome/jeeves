package repl

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"jeeves/api"
	"jeeves/auth"
	"jeeves/config"
)

// ---- screen enum ----

type screen int

const (
	screenLogin   screen = iota
	screenHistory        // left=months, right=posts for selected month
	screenRecent         // left=posts, right=post viewer
	screenSearch         // left=search input+results, right=post viewer
	screenCompose        // full-screen textarea
)

// ---- list item types ----

type monthItem struct{ m api.MonthEntry }

func (i monthItem) Title() string       { return i.m.Month }
func (i monthItem) Description() string { return fmt.Sprintf("%s  ·  %d posts", i.m.Year, i.m.Count) }
func (i monthItem) FilterValue() string { return i.m.Month + " " + i.m.Year }

type postItem struct{ p api.Post }

func (i postItem) Title() string { return formatDate(i.p.Date) }
func (i postItem) Description() string {
	runes := []rune(i.p.Body)
	if len(runes) > 72 {
		return string(runes[:72]) + "…"
	}
	return string(runes)
}
func (i postItem) FilterValue() string { return i.p.Body }

// ---- model ----

type Model struct {
	cfg           *config.Config
	width, height int
	mode          screen
	focusLeft     bool // true = left panel focused
	viewingPost   bool // true = right panel shows postView

	// lists
	leftList  list.Model // months (history) or posts (recent/search)
	rightList list.Model // posts for selected month (history)

	// post viewer
	postView viewport.Model

	// compose / edit
	composeArea   textarea.Model
	composeDate   textinput.Model
	dateActive    bool // true = date input focused
	editingPostID int  // 0 = new post, >0 = editing existing

	// search
	searchInput  textinput.Model
	searching    bool
	searchQuery  string // query used for current results (drives highlighting)

	// login
	emailInput textinput.Model
	passInput  textinput.Model
	passActive bool
	loginErr   string

	// data
	months []api.MonthEntry
	posts  []api.Post            // backing data for current list
	labels map[int]api.Label     // loaded once at startup

	// status
	loading   bool
	statusMsg string
}

func newList(items []list.Item, title string, w, h int) list.Model {
	d := list.NewDefaultDelegate()
	l := list.New(items, d, w, h)
	l.Title = title
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	return l
}

func newMonthList(months []api.MonthEntry, w, h int) list.Model {
	items := make([]list.Item, len(months))
	for i, m := range months {
		items[i] = monthItem{m}
	}
	return newList(items, "History", w, h)
}

func newPostList(posts []api.Post, title string, w, h int) list.Model {
	items := make([]list.Item, len(posts))
	for i, p := range posts {
		items[i] = postItem{p}
	}
	return newList(items, title, w, h)
}

func newModel(cfg *config.Config) Model {
	email := textinput.New()
	email.Placeholder = "email@example.com"
	email.Focus()
	email.CharLimit = 200

	pass := textinput.New()
	pass.Placeholder = "password"
	pass.EchoMode = textinput.EchoPassword
	pass.CharLimit = 200

	compDate := textinput.New()
	compDate.SetValue(time.Now().Format("2006-01-02"))
	compDate.CharLimit = 10
	compDate.Width = 12

	srch := textinput.New()
	srch.Placeholder = "type to search…"
	srch.CharLimit = 200

	compArea := textarea.New()
	compArea.Placeholder = "Write your entry here…"
	compArea.ShowLineNumbers = false

	leftL := newList(nil, "History", 0, 0)
	rightL := newList(nil, "Posts", 0, 0)
	pv := viewport.New(0, 0)
	pv.MouseWheelEnabled = true

	mode := screenLogin
	if auth.Token() != "" {
		mode = screenRecent
	}

	return Model{
		cfg:         cfg,
		mode:        mode,
		focusLeft:   true,
		leftList:    leftL,
		rightList:   rightL,
		postView:    pv,
		emailInput:  email,
		passInput:   pass,
		composeDate: compDate,
		searchInput: srch,
		composeArea: compArea,
	}
}

// ---- tea.Model interface ----

func (m Model) Init() tea.Cmd {
	if m.mode == screenRecent {
		return tea.Batch(cmdFetchRecent(m.cfg, 25), cmdFetchLabels(m.cfg))
	}
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m.recalcSizes(), nil

	case tea.MouseMsg:
		if m.viewingPost {
			var cmd tea.Cmd
			m.postView, cmd = m.postView.Update(msg)
			return m, cmd
		}
		return m, nil

	case tea.KeyMsg:
		// ctrl+c always quits
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		// 'q' quits only outside text-entry screens
		if m.mode != screenSearch && m.mode != screenCompose && m.mode != screenLogin {
			if key.Matches(msg, keys.Quit) {
				return m, tea.Quit
			}
		}
		return m.updateMode(msg)

	case historyLoadedMsg:
		m.loading = false
		m.searchQuery = ""
		m.months = msg.months
		lw := contentLeftW(m.width)
		h := m.contentH()
		m.leftList = newMonthList(msg.months, lw, h)
		m.mode = screenHistory
		m.focusLeft = true
		m.viewingPost = false
		return m, nil

	case postsLoadedMsg:
		m.loading = false
		m.posts = msg.posts
		lw := contentLeftW(m.width)
		rw := contentRightW(m.width)
		h := m.contentH()
		switch {
		case strings.HasPrefix(msg.source, "month:"):
			m.rightList = newPostList(msg.posts, "Posts", rw, h)
			m.viewingPost = false
			m.focusLeft = false
		case msg.source == "recent":
			m.searchQuery = ""
			m.leftList = newPostList(msg.posts, "Recent", lw, h)
			m.mode = screenRecent
			m.focusLeft = true
			m.viewingPost = false
			if len(msg.posts) > 0 {
				m.postView.SetContent(renderPost(msg.posts[0], "", m.labels))
				m.postView.GotoTop()
				m.viewingPost = true
			}
		case msg.source == "onthisday":
			m.searchQuery = ""
			m.leftList = newPostList(msg.posts, "On This Day", lw, h)
			m.mode = screenRecent
			m.focusLeft = true
			m.viewingPost = false
			if len(msg.posts) > 0 {
				m.postView.SetContent(renderPost(msg.posts[0], "", m.labels))
				m.postView.GotoTop()
				m.viewingPost = true
			}
		case strings.HasPrefix(msg.source, "search:"):
			m.searchQuery = strings.TrimPrefix(msg.source, "search:")
			m.leftList = newPostList(msg.posts, "Results", lw, h-2)
			m.searching = false
			if len(msg.posts) > 0 {
				m.postView.SetContent(renderPost(msg.posts[0], m.searchQuery, m.labels))
				m.postView.GotoTop()
				m.viewingPost = true
			} else {
				m.viewingPost = false
			}
		}
		return m, nil

	case loginDoneMsg:
		if err := auth.Save(msg.token); err != nil {
			m.statusMsg = "Warning: could not save session"
		}
		m.mode = screenRecent
		m.loading = true
		return m, tea.Batch(cmdFetchRecent(m.cfg, 25), cmdFetchLabels(m.cfg))

	case labelsLoadedMsg:
		m.labels = make(map[int]api.Label, len(msg.labels))
		for _, l := range msg.labels {
			m.labels[l.ID] = l
		}
		return m, nil

	case postCreatedMsg:
		m.loading = false
		m.statusMsg = greenStyle.Render("Posted! #" + fmt.Sprintf("%d", msg.post.ID))
		m.mode = screenRecent
		m.loading = true
		return m, tea.Batch(cmdFetchRecent(m.cfg, 25), clearStatusAfter(3*time.Second))

	case postUpdatedMsg:
		m.loading = false
		m.statusMsg = greenStyle.Render("Updated! #" + fmt.Sprintf("%d", msg.post.ID))
		m.mode = screenRecent
		m.loading = true
		return m, tea.Batch(cmdFetchRecent(m.cfg, 25), clearStatusAfter(3*time.Second))

	case errMsg:
		m.loading = false
		m.searching = false
		m.loginErr = msg.err.Error()
		m.statusMsg = msg.err.Error()
		return m, clearStatusAfter(5 * time.Second)

	case clearStatusMsg:
		m.statusMsg = ""
		m.loginErr = ""
		return m, nil
	}

	return m, nil
}

func (m Model) contentH() int {
	h := m.height - 2
	if h < 1 {
		return 1
	}
	return h
}

func (m Model) recalcSizes() Model {
	lw := contentLeftW(m.width)
	rw := contentRightW(m.width)
	h := m.contentH()

	leftH := h
	if m.mode == screenSearch {
		leftH = h - 2
	}
	m.leftList.SetSize(lw, leftH)
	m.rightList.SetSize(rw, h)
	m.postView.Width = rw
	m.postView.Height = h
	m.composeArea.SetWidth(m.width - 6)
	m.composeArea.SetHeight(h - 4)
	m.searchInput.Width = lw - 4
	return m
}

// ---- View ----

func (m Model) View() string {
	if m.width == 0 {
		return "Loading…"
	}
	switch m.mode {
	case screenLogin:
		return m.viewLogin()
	case screenCompose:
		return m.viewCompose()
	default:
		return m.viewTwoColumn()
	}
}

func (m Model) viewTwoColumn() string {
	h := m.contentH()
	plw := panelLeftW(m.width)
	prw := panelRightW(m.width)

	// -- left panel content --
	var leftContent string
	if m.mode == screenSearch {
		leftContent = "  / " + m.searchInput.View() + "\n\n"
		if m.searching {
			leftContent += dimStyle.Render("  Searching…")
		} else {
			leftContent += m.leftList.View()
		}
	} else {
		leftContent = m.leftList.View()
	}

	ls := panelLeft
	if m.focusLeft {
		ls = panelLeftFocused
	}
	left := ls.Width(plw).Height(h).Render(leftContent)

	// -- right panel content --
	var rightContent string
	if m.loading {
		rightContent = "\n  " + dimStyle.Render("Loading…")
	} else if m.viewingPost {
		rightContent = m.postView.View()
	} else if m.mode == screenHistory {
		if len(m.posts) > 0 {
			rightContent = m.rightList.View()
		} else {
			rightContent = "\n  " + dimStyle.Render("Select a month and press ↵")
		}
	} else {
		rightContent = "\n  " + dimStyle.Render("No posts.")
	}

	right := panelRight.Width(prw).Height(h).Render(rightContent)

	columns := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	return lipgloss.JoinVertical(lipgloss.Left,
		m.viewTitle(),
		columns,
		m.viewStatus(),
	)
}

func (m Model) viewTitle() string {
	loggedIn := ""
	if auth.Token() != "" {
		loggedIn = "  " + greenStyle.Render("●")
	}
	modeLabel := ""
	switch m.mode {
	case screenHistory:
		modeLabel = "  " + dimStyle.Render("history")
	case screenRecent:
		modeLabel = "  " + dimStyle.Render("recent")
	case screenSearch:
		modeLabel = "  " + dimStyle.Render("search")
	}
	return " " + titleStyle.Render("Jeeves") + loggedIn + modeLabel
}

func (m Model) viewStatus() string {
	if m.statusMsg != "" {
		if strings.HasPrefix(m.statusMsg, "Posted") {
			return " " + m.statusMsg
		}
		return " " + errStyle.Render(m.statusMsg)
	}
	switch m.mode {
	case screenSearch:
		if m.viewingPost {
			return " " + dimStyle.Render("↵ search  ·  ↑↓ navigate  ·  e edit  ·  esc back")
		}
		return " " + dimStyle.Render("↵ search  ·  ↑↓ navigate  ·  esc back")
	default:
		if m.viewingPost {
			return " " + dimStyle.Render("n new  ·  e edit  ·  / search  ·  r recent  ·  t today  ·  h history  ·  tab switch  ·  q quit")
		}
		return " " + dimStyle.Render("n new  ·  / search  ·  r recent  ·  t today  ·  h history  ·  tab switch  ·  q quit")
	}
}

func (m Model) viewLogin() string {
	errLine := ""
	if m.loginErr != "" {
		errLine = "\n\n  " + errStyle.Render(m.loginErr)
	}
	box := loginBoxStyle.Render(
		titleStyle.Render("Jeeves") + "\n\n" +
			"Email\n" + m.emailInput.View() + "\n\n" +
			"Password\n" + m.passInput.View() +
			errLine + "\n\n" +
			dimStyle.Render("tab · next field    ↵ · login"),
	)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func (m Model) viewCompose() string {
	titleText := "New Entry"
	if m.editingPostID != 0 {
		titleText = fmt.Sprintf("Edit Entry #%d", m.editingPostID)
	}
	title := " " + titleStyle.Render(titleText)
	dateRow := "  Date: " + m.composeDate.View()
	help := " " + dimStyle.Render("ctrl+s · save    tab · switch field    esc · cancel")
	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		dateRow,
		"",
		m.composeArea.View(),
		"",
		help,
	)
}

// ---- helpers ----

func renderPost(p api.Post, searchQuery string, labels map[int]api.Label) string {
	header := postDateStyle.Render(formatDate(p.Date)) +
		"  " + postIDStyle.Render(fmt.Sprintf("#%d", p.ID))
	sep := postIDStyle.Render(strings.Repeat("─", 50))

	var meta strings.Builder
	if len(p.Periods) > 0 {
		names := make([]string, len(p.Periods))
		for i, per := range p.Periods {
			names[i] = per.Name
		}
		meta.WriteString(postIDStyle.Render("⏱ "+strings.Join(names, " · ")) + "\n")
	}
	if len(p.Labels) > 0 {
		names := make([]string, 0, len(p.Labels))
		for _, id := range p.Labels {
			if l, ok := labels[id]; ok {
				names = append(names, l.Name)
			}
		}
		if len(names) > 0 {
			meta.WriteString(tagStyle.Render("⚑ "+strings.Join(names, " · ")) + "\n")
		}
	}
	metaStr := meta.String()
	if metaStr != "" {
		metaStr = "\n" + metaStr
	}

	body := highlightBody(p.Body, searchQuery)

	var comments strings.Builder
	if len(p.Comments) > 0 {
		comments.WriteString("\n\n" + postIDStyle.Render(strings.Repeat("─", 50)) + "\n")
		for _, c := range p.Comments {
			comments.WriteString(postIDStyle.Render(formatDate(c.Date)+"  ") + c.Body + "\n")
		}
	}

	return header + "\n" + sep + metaStr + "\n" + body + comments.String()
}

var tagRe = regexp.MustCompile(`#[\p{L}\p{N}_]+`)

type highlightSpan struct {
	start, end int
	style      lipgloss.Style
}

// highlightBody applies tag (#word) and search-query highlighting to body text.
// Search query matches take visual priority over tag matches.
func highlightBody(body, searchQuery string) string {
	var spans []highlightSpan

	// Search matches first (higher priority — processed first when ties occur)
	if searchQuery != "" {
		re := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(searchQuery))
		for _, loc := range re.FindAllStringIndex(body, -1) {
			spans = append(spans, highlightSpan{loc[0], loc[1], searchHighlightStyle})
		}
	}

	// Tags second (lower priority)
	for _, loc := range tagRe.FindAllStringIndex(body, -1) {
		spans = append(spans, highlightSpan{loc[0], loc[1], tagStyle})
	}

	if len(spans) == 0 {
		return body
	}

	// Stable sort by start position; search spans were added first so they win ties.
	sort.SliceStable(spans, func(i, j int) bool {
		return spans[i].start < spans[j].start
	})

	var sb strings.Builder
	pos := 0
	for _, s := range spans {
		if s.start < pos {
			continue // overlapped by a higher-priority span
		}
		sb.WriteString(body[pos:s.start])
		sb.WriteString(s.style.Render(body[s.start:s.end]))
		pos = s.end
	}
	sb.WriteString(body[pos:])
	return sb.String()
}

func formatDate(s string) string {
	for _, layout := range []string{
		time.RFC3339,
		"2006-01-02T15:04:05.000Z",
		"2006-01-02 15:04:05",
		"2006-01-02",
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.Format("Mon, 02 Jan 2006")
		}
	}
	return s
}
