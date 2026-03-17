package repl

import (
	"fmt"
	"os"
	"strings"

	"github.com/chzyer/readline"

	"jeeves/api"
	"jeeves/auth"
	"jeeves/config"
)

const bannerText = `
     _
    | | ___  _____   _____  ___
    | |/ _ \/ _ \ \ / / _ \/ __|
 _  | |  __/  __/\ V /  __/\__ \
(_)_/ |\___|\___| \_/ \___||___/
    |__/

  Your personal diary. Type 'help' for commands.
`

type handler struct {
	rl  *readline.Instance
	cfg *config.Config
}

func (h *handler) client() *api.Client {
	return api.NewClient(h.cfg.BaseURL, auth.Token(), h.cfg.Dev)
}

func Run(cfg *config.Config) {
	fmt.Print(colored(colorCyan, bannerText))

	loggedIn := auth.Token() != ""
	if loggedIn {
		fmt.Println(colored(colorGreen, "Welcome back!"))
	} else {
		fmt.Println(colored(colorYellow, "Type 'login' to get started."))
	}
	fmt.Println()

	rl, err := readline.New(promptFor(loggedIn))
	if err != nil {
		fmt.Fprintf(os.Stderr, "readline: %v\n", err)
		os.Exit(1)
	}
	defer rl.Close()

	h := &handler{rl: rl, cfg: cfg}

	for {
		line, err := rl.Readline()
		if err != nil {
			// EOF or Ctrl+D
			fmt.Println("\nGoodbye!")
			return
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		cmd := parts[0]
		args := parts[1:]

		switch cmd {
		case "login":
			h.login()
		case "logout":
			h.logout()
		case "post", "write", "new":
			h.requireAuth(func() { h.post(args) })
		case "read", "list":
			h.requireAuth(func() { h.read(args) })
		case "search":
			h.requireAuth(func() { h.search(args) })
		case "history":
			h.requireAuth(func() { h.history(args) })
		case "today", "onthisday":
			h.requireAuth(h.todayInHistory)
		case "help", "?":
			printHelp()
		case "exit", "quit", "q":
			fmt.Println("Goodbye!")
			return
		default:
			fmt.Printf(colored(colorYellow, "Unknown command '%s'. Type 'help' for available commands.\n"), cmd)
		}
	}
}

func (h *handler) requireAuth(fn func()) {
	if auth.Token() == "" {
		fmt.Println(colored(colorYellow, "Not logged in. Use 'login' first."))
		return
	}
	fn()
}

func printHelp() {
	fmt.Print(`
Commands:
  login          Log in with email and password
  logout         Clear current session

  post [date]    Write a new diary entry (date: today/yesterday/YYYY-MM-DD)
  read [N]       Show last N posts (default: 10)
  search <text>  Search posts by content
  history [ym]   Show months with entries, or posts for a month (e.g. 10-02)
  today          Show entries written on this day in previous years

  help           Show this help
  exit           Exit Jeeves

`)
}
