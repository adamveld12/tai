package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

const (
	borderTL = '╭'
	borderTP = '─'
	borderTR = '╮'
	borderBL = '╰'
	borderBM = '─'
	borderBR = '╯'
)

type ChatInput textinput.Model

func (i ChatInput) View() string {
	input := textinput.Model(i).View()
	trueWidth := lipgloss.Width(input)
	borderTop := strings.Repeat(string(borderTP), trueWidth)
	borderBottom := strings.Repeat(string(borderBM), trueWidth)
	return fmt.Sprintf(" %c%s%c\n |%s|\n %c%s%c", borderTL, borderTop, borderTR, input, borderBL, borderBottom, borderBR)
}

func ElementViewport(width, height int) viewport.Model {
	vp := viewport.New(width, height)
	vp.SetContent("")
	vp.MouseWheelEnabled = true
	// vp.Style = CurrentTheme().Styles().CodeBlock

	return vp
}

func ElementInput(prompt, placeholder string) ChatInput {
	ti := ChatInput(textinput.New())
	ti.Width = 100

	ti.ShowSuggestions = true

	ti.Cursor.BlinkSpeed = time.Second
	ti.Cursor.Blink = true

	if placeholder == "" {
		placeholder = "Ask me anything..."
	}
	ti.Placeholder = placeholder

	if prompt == "" {
		prompt = ">"
	}
	ti.Prompt = CurrentTheme().Styles().Primary.Bold(true).Render(fmt.Sprintf("%s ", prompt))

	return ti
}

var (
	RoleStyle = lipgloss.NewStyle().Bold(true).Foreground(CurrentTheme().Text())
	DimStyle  = CurrentTheme().Styles().Subtle
)
