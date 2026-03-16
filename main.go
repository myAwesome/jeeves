package main

import (
	"flag"
	"fmt"
	"os"

	"jeeves/config"
	"jeeves/repl"
)

func main() {
	dev := flag.Bool("dev", false, "enable dev mode: print all HTTP requests and responses")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}
	cfg.Dev = *dev
	repl.Run(cfg)
}
