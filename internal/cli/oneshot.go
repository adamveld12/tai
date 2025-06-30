package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/adamveld12/tai/internal/llm"
	"github.com/adamveld12/tai/internal/state"
)

// OneShotHandler handles one-shot mode execution
type OneShotHandler struct {
	state.Dispatcher
	llm.Provider
	config *Config
}

// NewOneShotHandler creates a new one-shot handler
func NewOneShotHandler(config *Config) *OneShotHandler {
	// Determine provider based on config
	var providerType state.SupportedProvider
	switch config.Provider {
	case "openai":
		providerType = state.ProviderOpenAI
	case "lmstudio":
		providerType = state.ProviderLMStudio
	default:
		providerType = state.ProviderLMStudio // default
	}

	s := state.NewMemoryState("", config.WorkingDirectory, time.Now().Format("20060102150405"))

	provider, err := llm.GetProvider(s, providerType, "")
	if err != nil {
		log.Fatalf("Failed to initialize LLM provider: %v", err)
	}

	return &OneShotHandler{
		Dispatcher: s,
		Provider:   provider,
		config:     config,
	}
}

// Execute runs the one-shot mode
func (h *OneShotHandler) Execute() error {
	var input string
	var err error

	input, err = h.readFromStdin()
	if err != nil {
		return fmt.Errorf("failed to read from stdin: %w", err)
	}

	stdin := strings.TrimSpace(input)
	prompt := h.config.Prompt
	if prompt == "" && stdin == "" {
		return nil
	} else if prompt == "" && stdin != "" {
		prompt = stdin
	} else if prompt != "" && stdin == "" {
		prompt = strings.TrimSpace(prompt)
	} else {
		prompt = fmt.Sprintf("%s\n%s", strings.TrimSpace(prompt), stdin)
	}

	s := h.GetState()
	response, err := h.Provider.ChatCompletion(context.Background(), llm.ChatRequest{
		Messages: []state.Message{
			{Role: state.RoleUser, Content: prompt, Timestamp: time.Now()},
		},
		SystemPrompt: s.Context.SystemPrompt,
	})

	if err != nil {
		return fmt.Errorf("failed to get chat completion:\n\t%w", err)
	}

	// Output the response
	fmt.Println(response.Content)
	return nil
}

// readFromStdin reads input from stdin
func (h *OneShotHandler) readFromStdin() (string, error) {
	// Check if stdin has data
	stat, err := os.Stdin.Stat()
	if err != nil {
		return "", err
	}

	if (stat.Mode() & os.ModeCharDevice) != 0 {
		// Terminal mode - no piped input
		return "", nil
	}

	// Read from stdin
	var lines []string
	reader := bufio.NewReader(os.Stdin)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				if line != "" {
					lines = append(lines, line)
				}
				break
			}
			return "", err
		}
		lines = append(lines, strings.TrimRight(line, "\n"))
	}

	return strings.Join(lines, "\n"), nil
}
