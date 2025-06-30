package ui

import (
	"context"
	"fmt"
	"log"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/adamveld12/tai/internal/agent"
	"github.com/adamveld12/tai/internal/llm"
	"github.com/adamveld12/tai/internal/state"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/stopwatch"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/muesli/reflow/wordwrap"
)

// REPLScreen represents the REPLScreen UI model
type REPLScreen struct {
	agent.Agent
	state.Dispatcher
	prompt     chan<- state.Message
	response   <-chan agent.AgentStatus
	input      textinput.Model
	viewport   viewport.Model
	swatch     stopwatch.Model
	spinner    spinner.Model
	width      int
	height     int
	ready      bool
	autoscroll bool
	mu         sync.Mutex
}

// NewREPL creates a new REPL instance
func NewREPL(d state.Dispatcher, p llm.Provider) *REPLScreen {
	agent, err := agent.Task(agent.TaskInput{
		Provider:   p,
		Dispatcher: d,
		Name:       "orchestrator",
	})

	if err != nil {
		log.Fatalf("💩 failed to create agent: %v", err)
	}

	prompts := make(chan state.Message, 1)
	output := agent.Start(context.Background(), prompts)

	repl := &REPLScreen{
		Agent:      agent,
		Dispatcher: d,
		prompt:     prompts,
		response:   output,
		swatch:     stopwatch.New(),
		input:      textinput.Model(ElementInput(">", "")),
		spinner:    spinner.New(spinner.WithSpinner(spinner.Points), spinner.WithStyle(CurrentStyles().Accent)),
		viewport:   ElementViewport(80, 20),
		autoscroll: true,
	}

	repl.swatch.Interval = time.Millisecond * 16
	go func() {
		for status := range output {
			if status.Error == nil {
				repl.setViewport()
			}
		}
	}()

	return repl
}

// Init initializes the REPL
func (r *REPLScreen) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen, r.viewport.Init())
}

func (r *REPLScreen) OnStateChange(action state.Action, newState, oldState state.AppState) (msg tea.Msg) {
	msg = action
	switch action.(type) {
	case agent.ChatCompletionStartedAction, agent.ChatCompletionCompletedAction, agent.MessageChunkAction, ClearMessagesAction:
		r.setViewport()
	}

	return
}

// Update handles messages and updates the model
func (r *REPLScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case agent.ChatCompletionStartedAction:
		cmds = append(cmds, r.swatch.Reset(), r.swatch.Start(), r.spinner.Tick)
	case agent.ChatCompletionCompletedAction:
		r.spinner = spinner.New(spinner.WithSpinner(spinner.Points), spinner.WithStyle(CurrentStyles().Accent))
		cmds = append(cmds, r.swatch.Stop())
	case ClearMessagesAction:
		r.viewport.GotoTop()
	case tea.WindowSizeMsg:
		r.width = msg.Width
		r.height = msg.Height

		// Update text input width
		r.input.Width = msg.Width - 7 // Account for prompt and padding

		// Update viewport size
		headerHeight := 2 // Header and separator
		footerHeight := 5 // Input area and footer
		viewportHeight := msg.Height - headerHeight - footerHeight
		r.viewport.Width = msg.Width
		r.viewport.Height = viewportHeight
		r.setViewport()
		r.input.Focus()
		r.ready = true

	case tea.MouseMsg:
		switch msg.Button {
		case tea.MouseButtonWheelUp:
			r.autoscroll = false
		case tea.MouseButtonWheelDown:
			if r.viewport.ScrollPercent() >= .95 {
				r.autoscroll = true
			}
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "ctrl+d":
			return r, tea.Quit
		case "esc":
			r.input.Reset()
			r.setViewport()
		case "enter":
			if input, ok := r.handleTextInput(r.input.Value()); ok {
				if strings.HasPrefix(input, ":") {
					return r.handleCommand(input)
				} else {
					r.prompt <- state.Message{
						Role:      state.RoleUser,
						Content:   input,
						Timestamp: time.Now(),
						ToolCalls: []state.ToolCall{},
					}
				}
			}
		default:
			// Let the text input handle other keys
			r.input, cmd = r.input.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	// if r.autoscroll {
	// 	r.viewport.GotoBottom()
	// }

	r.swatch, cmd = r.swatch.Update(msg)
	cmds = append(cmds, cmd)
	r.viewport, cmd = r.viewport.Update(msg)
	cmds = append(cmds, cmd)
	r.spinner, cmd = r.spinner.Update(msg)
	cmds = append(cmds, cmd)

	return r, tea.Batch(cmds...)
}

// View renders the REPL interface
func (r *REPLScreen) View() string {
	if !r.ready {
		return "Initializing..."
	}

	var b strings.Builder

	// Header
	header := CurrentStyles().Header.Render("TAI - Terminal AI Assistant")

	b.WriteString(header)
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", r.width))
	b.WriteString("\n")

	b.WriteString(r.viewport.View())

	// Input area
	b.WriteString(strings.Repeat("─", r.width))
	b.WriteString("\n")

	b.WriteString(CurrentStyles().Subtle.Render(fmt.Sprintf("%s %s", r.spinner.View(), r.swatch.View())))
	b.WriteString("\n")
	b.WriteString(ChatInput(r.input).View())

	footer := CurrentStyles().Subtle.Render("\n:help, :clear, :quit, :theme | Ctrl+C to exit")
	b.WriteString(footer)

	return b.String()
}

// handleCommand processes colon commands
func (r *REPLScreen) handleCommand(cmd string) (tea.Model, tea.Cmd) {
	wrapWidth := int(math.Max(40, float64(r.viewport.Width)-10))

	switch strings.ToLower(strings.TrimSpace(cmd)) {
	case ":quit", ":q", ":exit":
		return r, tea.Quit
	case ":clear", ":c":
		r.Dispatcher.Dispatch(ClearMessagesAction{})
		return r, nil
	case ":help", ":h":
		helpText := `# TAI Commands

## Available Commands

| Command | Shortcut | Description |
|---------|----------|-------------|
| **:help** | **:h** | Show this help |
| **:clear** | **:c** | Clear conversation |
| **:quit** | **:q** | Exit application |

## Usage Tips

- Type your message and press **Enter** to send
- Use **mouse wheel** or **arrow keys** to scroll through history
- Messages support **markdown formatting**
`
		wrappedHelp := wordwrap.String(helpText, wrapWidth)
		if renderer, err := glamour.NewTermRenderer(glamour.WithWordWrap(wrapWidth)); err == nil {
			if renderedHelp, err := renderer.Render(helpText); err == nil {
				wrappedHelp = strings.TrimSpace(renderedHelp)
			}
		}

		r.viewport.SetContent(wrappedHelp)
		r.input.Reset()
		return r, nil
	default:
		// Wrap error message based on viewport width
		errorMsg := fmt.Sprintf("Unknown command: %s (type :help for available commands)\n", cmd)
		wrappedError := wordwrap.String(errorMsg, wrapWidth)
		r.viewport.SetContent(wrappedError)
		r.input.SetValue("")
		return r, nil
	}
}

func (r *REPLScreen) handleTextInput(content string) (input string, ok bool) {
	if input = strings.TrimSpace(content); input != "" {
		ok = true
	}

	defer r.input.SetValue("")
	return
}

// addToViewport adds content to the viewport
func (r *REPLScreen) setViewport() {

	newState := r.GetState()
	var builder strings.Builder
	var renderer *glamour.TermRenderer
	var err error

	// Apply additional wordwrap if needed (glamour should handle most of it)
	wrapWidth := 40 // Minimum wrap width
	if ww := r.viewport.Width - 10; ww > wrapWidth {
		wrapWidth = ww
	}

	// Create glamour renderer with dark theme
	if renderer, err = glamour.NewTermRenderer(
		glamour.WithStandardStyle("dracula"),
		glamour.WithWordWrap(wrapWidth),
	); err != nil {
		// Fallback to plain rendering if glamour fails
		renderer = nil
	}

	msgs := append([]state.Message{{
		Timestamp: newState.Context.Created,
		Role:      state.RoleSystem,
		Content:   newState.Context.SystemPrompt,
		ToolCalls: []state.ToolCall{},
	}}, newState.Context.Messages...)

	for _, msg := range msgs {
		role := string(msg.Role)
		renderedContent := wordwrap.String(msg.Content, wrapWidth)

		switch msg.Role {
		case state.RoleUser:
			role = CurrentStyles().Subtle.Render(role)
			renderedContent = CurrentStyles().Subtle.Render(renderedContent)
		case state.RoleSystem:
			role = CurrentStyles().Accent.Render("System >")
			renderedContent = CurrentStyles().Primary.Render(renderedContent)
		case state.RoleTool:
			role = CurrentStyles().Primary.Render(role)
		case state.RoleAssistant:
			role = CurrentStyles().Primary.Bold(true).Render(fmt.Sprintf("%s ~> %s", newState.Model.Provider, newState.Model.Name))
			fallthrough
		default:
			role = CurrentStyles().Primary.Render(role)
			if rendered, err := renderer.Render(msg.Content); err == nil {
				renderedContent = rendered
			}
		}

		fmt.Fprintf(
			&builder,
			"%s\n\t%s\n\n",
			role,
			renderedContent,
		)
	}

	r.mu.Lock()
	r.viewport.SetContent(builder.String())

	if r.autoscroll {
		r.viewport.GotoBottom()
	}
	r.mu.Unlock()
}
