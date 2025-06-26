package cli

import (
	"flag"
	"fmt"
	"os"
)

// Mode represents the execution mode of the application
type Mode string

const (
	ModeREPL    Mode = "repl"
	ModeOneShot Mode = "oneshot"
)

// Config holds the configuration for the CLI application
type Config struct {
	WorkingDirectory string
	Mode             Mode
	Prompt           string
	SystemPrompt     string
	Verbose          bool
	Help             bool
	Provider         string
}

// ParseArgs parses command line arguments and returns a Config
func ParseArgs() (*Config, error) {
	config := &Config{}
	var oneshot bool

	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %w", err)
	}

	flag.BoolVar(&oneshot, "oneshot", false, "Run in one-shot mode (single prompt and exit)")
	flag.BoolVar(&config.Verbose, "verbose", false, "Enable verbose logging")
	flag.BoolVar(&config.Help, "help", false, "Show help message")
	flag.StringVar(&config.Provider, "provider", "lmstudio", "Specify the LLM provider to use (e.g., lmstudio)")
	flag.StringVar(&config.SystemPrompt, "system", "", "Specify the system prompt to use")
	flag.StringVar(&config.WorkingDirectory, "dir", wd, "Set the working directory (default: current directory)")

	flag.Parse()

	if oneshot {
		config.Mode = ModeOneShot
		// Get input from remaining args or stdin
		args := flag.Args()
		if len(args) > 0 {
			config.Prompt = args[0]
		}
	} else {
		config.Mode = ModeREPL
	}

	return config, nil
}

// ShowHelp displays the help message
func ShowHelp() {
	fmt.Fprintf(os.Stderr, `TAI - Terminal AI Assistant

Usage:
  tai                          Start interactive REPL mode
  tai -oneshot "your prompt"  Run in one-shot mode (read from stdin)

Options:
  -oneshot         Run in one-shot mode
  -verbose         Enable verbose logging
  -help            Show this help message
  -provider        LLM provider to use (default: lmstudio)
  -system          System prompt to use
  -dir             Working directory (default: current directory)

Examples:
  tai                                                    # Start REPL mode
  tai -oneshot "Hello, world!"                           # One-shot with prompt
  echo "Hello" | tai -oneshot                            # One-shot from stdin
  echo "Hello" | tai -oneshot 'what comes after Hello?' # One-shot from stdin with additional prompt
  tai -provider ollama -system "You are a poet"          # REPL with custom provider and system prompt
  tai -dir /path/to/project -oneshot "analyze this"     # One-shot with custom working directory

`)
}
