package repl

import "jeeves/api"

type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

type loginDoneMsg struct{ token string }
type historyLoadedMsg struct{ months []api.MonthEntry }
type postsLoadedMsg struct {
	posts  []api.Post
	source string // "recent" | "month:<ym>" | "search:<q>" | "onthisday"
}
type postCreatedMsg struct{ post api.Post }
type postUpdatedMsg struct{ post api.Post }
type labelsLoadedMsg struct{ labels []api.Label }
type clearStatusMsg struct{}
