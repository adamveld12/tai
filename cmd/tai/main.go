package main

import (
	"fmt"
	"os"

	"github.com/adamveld12/tai/internal/cli"
)

func main() {
	// Parse command line arguments
	config, err := cli.ParseArgs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing arguments: %v\n", err)
		os.Exit(1)
	}

	// Show help if requested
	if config.Help {
		cli.ShowHelp()
		os.Exit(0)
	}

	var handler cli.Executor

	// Execute based on mode
	switch config.Mode {
	case cli.ModeOneShot:
		handler = cli.NewOneShotHandler(config)
	case cli.ModeREPL:
		handler = cli.NewReplHandler(config)
	default:
		fmt.Fprintf(os.Stderr, "Unknown mode: %v\n", config.Mode)
		os.Exit(1)
	}

	if err := handler.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running %v: %v\n", config.Mode, err)
		os.Exit(1)
	}
}
