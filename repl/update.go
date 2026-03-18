package repl

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"jeeves/api"
	"jeeves/auth"
	"jeeves/config"
)

func (m Model) updateMode(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch m.mode {
	case screenLogin:
		return m.updateLogin(msg)
	case screenHistory:
		return m.updateHistory(msg)
	case screenRecent:
		return m.updateRecent(msg)
	case screenSearch:
		return m.updateSearch(msg)
	case screenCompose:
		return m.updateCompose(msg)
	}
	return m, nil
}

// ---- login ----

func (m Model) updateLogin(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Tab), key.Matches(msg, keys.Down):
		if !m.passActive {
			m.emailInput.Blur()
			m.passActive = true
			m.passInput.Focus()
			return m, textinput.Blink
		}
		m.passInput.Blur()
		m.passActive = false
		m.emailInput.Focus()
		return m, textinput.Blink

	case key.Matches(msg, keys.Enter):
		if !m.passActive {
			m.emailInput.Blur()
			m.passActive = true
			m.passInput.Focus()
			return m, textinput.Blink
		}
		email := strings.TrimSpace(m.emailInput.Value())
		pass := m.passInput.Value()
		if email == "" || pass == "" {
			m.loginErr = "Email and password are required."
			return m, clearStatusAfter(3 * time.Second)
		}
		m.loading = true
		m.loginErr = ""
		return m, cmdLogin(m.cfg, email, pass)
	}

	var cmd tea.Cmd
	if m.passActive {
		m.passInput, cmd = m.passInput.Update(msg)
	} else {
		m.emailInput, cmd = m.emailInput.Update(msg)
	}
	return m, cmd
}

// ---- history ----

func (m Model) updateHistory(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Compose):
		return m.enterCompose()
	case key.Matches(msg, keys.Search):
		return m.enterSearch()
	case key.Matches(msg, keys.Recent):
		m.loading = true
		return m, cmdFetchRecent(m.cfg, 30)
	case key.Matches(msg, keys.Today):
		m.loading = true
		return m, cmdFetchToday(m.cfg)
	case key.Matches(msg, keys.Logout):
		_ = auth.Clear()
		m.mode = screenLogin
		m.passActive = false
		m.emailInput.SetValue("")
		m.passInput.SetValue("")
		m.emailInput.Focus()
		return m, textinput.Blink
	case key.Matches(msg, keys.Tab):
		m.focusLeft = !m.focusLeft
		return m, nil
	case key.Matches(msg, keys.Escape):
		if m.viewingPost {
			m.viewingPost = false
			m.focusLeft = false
		}
		return m, nil
	case key.Matches(msg, keys.Enter):
		if m.focusLeft {
			i := m.leftList.Index()
			if i >= 0 && i < len(m.months) {
				m.loading = true
				return m, cmdFetchPostsByMonth(m.cfg, m.months[i].YM)
			}
		} else {
			i := m.rightList.Index()
			if i >= 0 && i < len(m.posts) {
				m.postView.SetContent(renderPost(m.posts[i]))
				m.postView.GotoTop()
				m.viewingPost = true
			}
		}
		return m, nil
	}

	var cmd tea.Cmd
	if m.viewingPost && !m.focusLeft {
		m.postView, cmd = m.postView.Update(msg)
	} else if m.focusLeft {
		m.leftList, cmd = m.leftList.Update(msg)
	} else {
		m.rightList, cmd = m.rightList.Update(msg)
	}
	return m, cmd
}

// ---- recent / onthisday (left=posts, right=post viewer) ----

func (m Model) updateRecent(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Compose):
		return m.enterCompose()
	case key.Matches(msg, keys.Search):
		return m.enterSearch()
	case key.Matches(msg, keys.History):
		m.loading = true
		return m, cmdFetchHistory(m.cfg)
	case key.Matches(msg, keys.Today):
		m.loading = true
		return m, cmdFetchToday(m.cfg)
	case key.Matches(msg, keys.Logout):
		_ = auth.Clear()
		m.mode = screenLogin
		m.passActive = false
		m.emailInput.SetValue("")
		m.passInput.SetValue("")
		m.emailInput.Focus()
		return m, textinput.Blink
	case key.Matches(msg, keys.Tab):
		m.focusLeft = !m.focusLeft
		return m, nil
	case key.Matches(msg, keys.Escape):
		if m.viewingPost && !m.focusLeft {
			m.focusLeft = true
		}
		return m, nil
	case key.Matches(msg, keys.Enter):
		i := m.leftList.Index()
		if i >= 0 && i < len(m.posts) {
			m.postView.SetContent(renderPost(m.posts[i]))
			m.postView.GotoTop()
			m.viewingPost = true
			m.focusLeft = false
		}
		return m, nil
	}

	var cmd tea.Cmd
	if m.viewingPost && !m.focusLeft {
		m.postView, cmd = m.postView.Update(msg)
	} else {
		prevIdx := m.leftList.Index()
		m.leftList, cmd = m.leftList.Update(msg)
		// auto-preview on navigation
		if m.leftList.Index() != prevIdx {
			i := m.leftList.Index()
			if i >= 0 && i < len(m.posts) {
				m.postView.SetContent(renderPost(m.posts[i]))
				m.postView.GotoTop()
				m.viewingPost = true
			}
		}
	}
	return m, cmd
}

// ---- search ----

func (m Model) updateSearch(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Escape):
		m.searchInput.Blur()
		m.searchInput.SetValue("")
		m.mode = screenHistory
		m = m.recalcSizes()
		return m, nil

	case key.Matches(msg, keys.Enter):
		if m.searching {
			return m, nil
		}
		query := strings.TrimSpace(m.searchInput.Value())
		if query == "" {
			return m, nil
		}
		m.searching = true
		return m, cmdSearch(m.cfg, query)

	case key.Matches(msg, keys.Up), key.Matches(msg, keys.Down):
		if len(m.posts) > 0 {
			prevIdx := m.leftList.Index()
			var listCmd tea.Cmd
			m.leftList, listCmd = m.leftList.Update(msg)
			if m.leftList.Index() != prevIdx {
				i := m.leftList.Index()
				if i >= 0 && i < len(m.posts) {
					m.postView.SetContent(renderPost(m.posts[i]))
					m.postView.GotoTop()
					m.viewingPost = true
				}
			}
			return m, listCmd
		}
	}

	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	return m, cmd
}

// ---- compose ----

func (m Model) updateCompose(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Escape):
		m.composeArea.Blur()
		m.mode = screenHistory
		return m, nil

	case key.Matches(msg, keys.Submit):
		body := strings.TrimSpace(m.composeArea.Value())
		if body == "" {
			m.statusMsg = "Empty entry, cancelled."
			m.mode = screenHistory
			m.composeArea.Blur()
			return m, clearStatusAfter(2 * time.Second)
		}
		date, err := time.Parse("2006-01-02", m.composeDate.Value())
		if err != nil {
			date = time.Now()
		}
		m.loading = true
		m.composeArea.Reset()
		m.composeArea.Blur()
		m.mode = screenHistory
		return m, cmdCreatePost(m.cfg, body, date)

	case key.Matches(msg, keys.Tab):
		if m.dateActive {
			m.composeDate.Blur()
			m.dateActive = false
			m.composeArea.Focus()
			return m, textarea.Blink
		}
		m.composeArea.Blur()
		m.dateActive = true
		m.composeDate.Focus()
		return m, textinput.Blink
	}

	var cmd tea.Cmd
	if m.dateActive {
		m.composeDate, cmd = m.composeDate.Update(msg)
	} else {
		m.composeArea, cmd = m.composeArea.Update(msg)
	}
	return m, cmd
}

// ---- mode transition helpers ----

func (m Model) enterCompose() (Model, tea.Cmd) {
	m.mode = screenCompose
	m.dateActive = false
	m.composeDate.SetValue(time.Now().Format("2006-01-02"))
	m.composeDate.Blur()
	m.composeArea.Reset()
	m.composeArea.Focus()
	m = m.recalcSizes()
	return m, textarea.Blink
}

func (m Model) enterSearch() (Model, tea.Cmd) {
	m.mode = screenSearch
	m.posts = nil
	m.viewingPost = false
	m.searchInput.SetValue("")
	m.searchInput.Focus()
	m = m.recalcSizes()
	return m, textinput.Blink
}

// ---- API tea.Cmd wrappers ----

func clearStatusAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg { return clearStatusMsg{} })
}

func cmdFetchHistory(cfg *config.Config) tea.Cmd {
	return func() tea.Msg {
		months, err := api.NewClient(cfg.BaseURL, auth.Token(), cfg.Dev).GetHistory()
		if err != nil {
			return errMsg{err}
		}
		return historyLoadedMsg{months}
	}
}

func cmdFetchPostsByMonth(cfg *config.Config, ym string) tea.Cmd {
	return func() tea.Msg {
		posts, err := api.NewClient(cfg.BaseURL, auth.Token(), cfg.Dev).GetPostsByMonth(ym)
		if err != nil {
			return errMsg{err}
		}
		return postsLoadedMsg{posts: posts, source: "month:" + ym}
	}
}

func cmdFetchRecent(cfg *config.Config, n int) tea.Cmd {
	return func() tea.Msg {
		posts, err := api.NewClient(cfg.BaseURL, auth.Token(), cfg.Dev).GetPosts(n)
		if err != nil {
			return errMsg{err}
		}
		return postsLoadedMsg{posts: posts, source: "recent"}
	}
}

func cmdFetchToday(cfg *config.Config) tea.Cmd {
	return func() tea.Msg {
		now := time.Now()
		mo := fmt.Sprintf("%02d", now.Month())
		day := fmt.Sprintf("%02d", now.Day())
		posts, err := api.NewClient(cfg.BaseURL, auth.Token(), cfg.Dev).GetPostsHistory(mo, day)
		if err != nil {
			return errMsg{err}
		}
		return postsLoadedMsg{posts: posts, source: "onthisday"}
	}
}

func cmdSearch(cfg *config.Config, query string) tea.Cmd {
	return func() tea.Msg {
		posts, err := api.NewClient(cfg.BaseURL, auth.Token(), cfg.Dev).SearchPosts(query, 20)
		if err != nil {
			return errMsg{err}
		}
		return postsLoadedMsg{posts: posts, source: "search:" + query}
	}
}

func cmdLogin(cfg *config.Config, email, password string) tea.Cmd {
	return func() tea.Msg {
		token, err := api.NewClient(cfg.BaseURL, "", cfg.Dev).Login(email, password)
		if err != nil {
			return errMsg{err}
		}
		return loginDoneMsg{token}
	}
}

func cmdCreatePost(cfg *config.Config, body string, date time.Time) tea.Cmd {
	return func() tea.Msg {
		p, err := api.NewClient(cfg.BaseURL, auth.Token(), cfg.Dev).CreatePost(body, date)
		if err != nil {
			return errMsg{err}
		}
		return postCreatedMsg{*p}
	}
}
