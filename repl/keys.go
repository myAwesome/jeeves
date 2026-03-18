package repl

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Up      key.Binding
	Down    key.Binding
	Tab     key.Binding
	Enter   key.Binding
	Escape  key.Binding
	Compose key.Binding
	Search  key.Binding
	History key.Binding
	Recent  key.Binding
	Today   key.Binding
	Logout  key.Binding
	Submit  key.Binding
	Quit    key.Binding
}

var keys = keyMap{
	Up:      key.NewBinding(key.WithKeys("up", "k"),      key.WithHelp("↑/k", "up")),
	Down:    key.NewBinding(key.WithKeys("down", "j"),    key.WithHelp("↓/j", "down")),
	Tab:     key.NewBinding(key.WithKeys("tab"),          key.WithHelp("tab", "switch panel")),
	Enter:   key.NewBinding(key.WithKeys("enter"),        key.WithHelp("↵", "select")),
	Escape:  key.NewBinding(key.WithKeys("esc"),          key.WithHelp("esc", "back")),
	Compose: key.NewBinding(key.WithKeys("n"),            key.WithHelp("n", "new post")),
	Search:  key.NewBinding(key.WithKeys("/"),            key.WithHelp("/", "search")),
	History: key.NewBinding(key.WithKeys("h"),            key.WithHelp("h", "history")),
	Recent:  key.NewBinding(key.WithKeys("r"),            key.WithHelp("r", "recent")),
	Today:   key.NewBinding(key.WithKeys("t"),            key.WithHelp("t", "today")),
	Logout:  key.NewBinding(key.WithKeys("L"),            key.WithHelp("L", "logout")),
	Submit:  key.NewBinding(key.WithKeys("ctrl+s"),       key.WithHelp("ctrl+s", "submit")),
	Quit:    key.NewBinding(key.WithKeys("q", "ctrl+c"),  key.WithHelp("q", "quit")),
}
