package main

import (
	"fmt"
	"os"

	"jeeves/config"
	"jeeves/repl"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}
	repl.Run(cfg)
}
