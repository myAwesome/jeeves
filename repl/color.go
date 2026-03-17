package repl

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorBold   = "\033[1m"
	colorDim    = "\033[2m"
)

func colored(color, s string) string {
	return color + s + colorReset
}

func promptFor(loggedIn bool) string {
	if loggedIn {
		return colored(colorGreen, "●") + " > "
	}
	return "> "
}
