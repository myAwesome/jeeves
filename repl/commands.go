package repl

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"jeeves/api"
	"jeeves/auth"
)

func (h *handler) login() {
	fmt.Print("Email: ")
	email, err := h.rl.Readline()
	if err != nil {
		return
	}
	email = strings.TrimSpace(email)

	password, err := h.rl.ReadPassword("Password: ")
	if err != nil {
		return
	}

	token, err := h.client().Login(email, string(password))
	if err != nil {
		fmt.Printf("Login failed: %v\n", err)
		return
	}

	if err := auth.Save(token); err != nil {
		fmt.Printf("Warning: could not save session: %v\n", err)
	}
	fmt.Println("Logged in.")
}

func (h *handler) logout() {
	if err := auth.Clear(); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("Logged out.")
}

func (h *handler) post(args []string) {
	date := parseDateArg(args)
	fmt.Printf("Date: %s\n", date.Format("2006-01-02"))
	fmt.Println("Body (enter '.' on empty line to finish):")

	h.rl.SetPrompt("  ")
	defer h.rl.SetPrompt("> ")

	var lines []string
	for {
		line, err := h.rl.Readline()
		if err != nil || line == "." {
			break
		}
		lines = append(lines, line)
	}

	body := strings.TrimSpace(strings.Join(lines, "\n"))
	if body == "" {
		fmt.Println("Empty entry, cancelled.")
		return
	}

	p, err := h.client().CreatePost(body, date)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Posted! (id: %d)\n", p.ID)
}

func parseDateArg(args []string) time.Time {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if len(args) == 0 {
		return today
	}
	switch args[0] {
	case "yesterday", "y":
		return today.AddDate(0, 0, -1)
	default:
		if t, err := time.Parse("2006-01-02", args[0]); err == nil {
			return t
		}
	}
	return today
}

func (h *handler) read(args []string) {
	limit := 10
	if len(args) > 0 {
		if n, err := strconv.Atoi(args[0]); err == nil && n > 0 {
			limit = n
		}
	}

	posts, err := h.client().GetPosts(limit)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	if len(posts) == 0 {
		fmt.Println("No posts yet.")
		return
	}

	fmt.Println()
	for _, p := range posts {
		printPost(p)
	}
}

func (h *handler) search(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: search <query>")
		return
	}
	query := strings.Join(args, " ")

	posts, err := h.client().SearchPosts(query, 20)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	if len(posts) == 0 {
		fmt.Printf("No results for '%s'.\n", query)
		return
	}

	fmt.Printf("\nFound %d result(s):\n\n", len(posts))
	for _, p := range posts {
		printPost(p)
	}
}

func (h *handler) history() {
	data, err := h.client().GetHistory()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	if len(data) == 0 {
		fmt.Println("No history yet.")
		return
	}

	fmt.Println()
	for year, months := range data {
		fmt.Printf("  %s: ", year)
		if monthList, ok := months.([]any); ok {
			parts := make([]string, 0, len(monthList))
			for _, m := range monthList {
				parts = append(parts, fmt.Sprintf("%v", m))
			}
			fmt.Println(strings.Join(parts, ", "))
		}
	}
	fmt.Println()
}

func printPost(p api.Post) {
	sep := strings.Repeat("─", 60)
	fmt.Println(sep)
	fmt.Printf("  [%s]  #%d\n\n", formatDate(p.Date), p.ID)

	body := p.Body
	if len(body) > 600 {
		body = body[:600] + "..."
	}
	for _, line := range strings.Split(body, "\n") {
		fmt.Printf("  %s\n", line)
	}
	fmt.Println()
}

func formatDate(s string) string {
	for _, layout := range []string{
		time.RFC3339,
		"2006-01-02T15:04:05.000Z",
		"2006-01-02 15:04:05",
		"2006-01-02",
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.Format("Mon, 02 Jan 2006  15:04")
		}
	}
	return s
}

