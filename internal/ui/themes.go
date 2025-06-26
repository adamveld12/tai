package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Theme defines the interface for color themes
type Theme interface {
	// Core colors
	Primary() lipgloss.Color
	Secondary() lipgloss.Color
	Accent() lipgloss.Color
	Background() lipgloss.Color
	Surface() lipgloss.Color
	Text() lipgloss.Color
	TextSubtle() lipgloss.Color

	// Semantic colors
	Error() lipgloss.Color
	Warning() lipgloss.Color
	Success() lipgloss.Color
	Info() lipgloss.Color

	// UI-specific colors
	Border() lipgloss.Color
	Highlight() lipgloss.Color
	Selection() lipgloss.Color

	// Get pre-configured styles
	Styles() *ThemeStyles
}

// ThemeStyles contains pre-configured styles for common UI elements
type ThemeStyles struct {
	Primary   lipgloss.Style
	Secondary lipgloss.Style
	Accent    lipgloss.Style
	Header    lipgloss.Style
	Highlight lipgloss.Style
	Border    lipgloss.Style
	Error     lipgloss.Style
	Warning   lipgloss.Style
	Success   lipgloss.Style
	Info      lipgloss.Style
	Subtle    lipgloss.Style
	CodeBlock lipgloss.Style
	Input     lipgloss.Style
}

// BaseTheme provides common functionality for all themes
type BaseTheme struct {
	styles *ThemeStyles
}

func (t *BaseTheme) Styles() *ThemeStyles {
	return t.styles
}

// buildStyles creates all the pre-configured styles for a theme
func buildStyles(theme Theme) *ThemeStyles {
	return &ThemeStyles{
		Primary: lipgloss.NewStyle().
			Foreground(theme.Primary()),

		Secondary: lipgloss.NewStyle().
			Foreground(theme.Secondary()),

		Accent: lipgloss.NewStyle().
			Foreground(theme.Accent()),

		Header: lipgloss.NewStyle().
			Foreground(theme.Text()).
			Background(theme.Surface()).
			Bold(true).
			Padding(0, 1),

		Highlight: lipgloss.NewStyle().
			Foreground(theme.Background()).
			Background(theme.Highlight()).
			Padding(0, 1),

		Border: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(theme.Border()),

		Error: lipgloss.NewStyle().
			Foreground(theme.Error()).
			Bold(true),

		Warning: lipgloss.NewStyle().
			Foreground(theme.Warning()).
			Bold(true),

		Success: lipgloss.NewStyle().
			Foreground(theme.Success()).
			Bold(true),

		Info: lipgloss.NewStyle().
			Foreground(theme.Info()),

		Subtle: lipgloss.NewStyle().
			Foreground(theme.TextSubtle()).
			Faint(true),

		CodeBlock: lipgloss.NewStyle().
			Background(theme.Surface()).
			Foreground(theme.Text()).
			Padding(1).
			MarginTop(1).
			MarginBottom(1),

		Input: lipgloss.NewStyle().
			Foreground(theme.Text()).
			Background(theme.Surface()).
			Padding(0, 1),
	}
}

// RetroTheme implements the original retro color scheme
type RetroTheme struct {
	BaseTheme
}

func NewRetroTheme() Theme {
	t := &RetroTheme{}
	t.styles = buildStyles(t)
	return t
}

func (t *RetroTheme) Primary() lipgloss.Color    { return lipgloss.Color("#028391") } // Teal
func (t *RetroTheme) Secondary() lipgloss.Color  { return lipgloss.Color("#F85525") } // Orange-Red
func (t *RetroTheme) Accent() lipgloss.Color     { return lipgloss.Color("#FAA698") } // Peach
func (t *RetroTheme) Background() lipgloss.Color { return lipgloss.Color("#01204E") } // Dark Navy
func (t *RetroTheme) Surface() lipgloss.Color    { return lipgloss.Color("#02356B") } // Slightly lighter navy
func (t *RetroTheme) Text() lipgloss.Color       { return lipgloss.Color("#F6DCAC") } // Cream
func (t *RetroTheme) TextSubtle() lipgloss.Color { return lipgloss.Color("#D4C29A") } // Darker cream
func (t *RetroTheme) Error() lipgloss.Color      { return lipgloss.Color("#F85525") } // Orange-Red
func (t *RetroTheme) Warning() lipgloss.Color    { return lipgloss.Color("#FAA698") } // Peach
func (t *RetroTheme) Success() lipgloss.Color    { return lipgloss.Color("#028391") } // Teal
func (t *RetroTheme) Info() lipgloss.Color       { return lipgloss.Color("#028391") } // Teal
func (t *RetroTheme) Border() lipgloss.Color     { return lipgloss.Color("#028391") } // Teal
func (t *RetroTheme) Highlight() lipgloss.Color  { return lipgloss.Color("#FAA698") } // Peach
func (t *RetroTheme) Selection() lipgloss.Color  { return lipgloss.Color("#02356B") } // Slightly lighter navy

// DarkTheme implements a dark purple theme
type DarkTheme struct {
	BaseTheme
}

func NewDarkTheme() Theme {
	t := &DarkTheme{}
	t.styles = buildStyles(t)
	return t
}

func (t *DarkTheme) Primary() lipgloss.Color    { return lipgloss.Color("#B794F4") } // Light purple
func (t *DarkTheme) Secondary() lipgloss.Color  { return lipgloss.Color("#805AD5") } // Medium purple
func (t *DarkTheme) Accent() lipgloss.Color     { return lipgloss.Color("#E9D8FD") } // Very light purple
func (t *DarkTheme) Background() lipgloss.Color { return lipgloss.Color("#1A0033") } // Deep dark purple
func (t *DarkTheme) Surface() lipgloss.Color    { return lipgloss.Color("#2D0E4F") } // Dark purple
func (t *DarkTheme) Text() lipgloss.Color       { return lipgloss.Color("#F7FAFC") } // Almost white
func (t *DarkTheme) TextSubtle() lipgloss.Color { return lipgloss.Color("#CBD5E0") } // Light gray
func (t *DarkTheme) Error() lipgloss.Color      { return lipgloss.Color("#FC8181") } // Light red
func (t *DarkTheme) Warning() lipgloss.Color    { return lipgloss.Color("#F6E05E") } // Yellow
func (t *DarkTheme) Success() lipgloss.Color    { return lipgloss.Color("#68D391") } // Light green
func (t *DarkTheme) Info() lipgloss.Color       { return lipgloss.Color("#63B3ED") } // Light blue
func (t *DarkTheme) Border() lipgloss.Color     { return lipgloss.Color("#553C9A") } // Purple
func (t *DarkTheme) Highlight() lipgloss.Color  { return lipgloss.Color("#9F7AEA") } // Medium light purple
func (t *DarkTheme) Selection() lipgloss.Color  { return lipgloss.Color("#44337A") } // Dark purple

// LightTheme implements Solarized Light color scheme
type LightTheme struct {
	BaseTheme
}

func NewLightTheme() Theme {
	t := &LightTheme{}
	t.styles = buildStyles(t)
	return t
}

func (t *LightTheme) Primary() lipgloss.Color    { return lipgloss.Color("#268bd2") } // Blue
func (t *LightTheme) Secondary() lipgloss.Color  { return lipgloss.Color("#cb4b16") } // Orange
func (t *LightTheme) Accent() lipgloss.Color     { return lipgloss.Color("#d33682") } // Magenta
func (t *LightTheme) Background() lipgloss.Color { return lipgloss.Color("#fdf6e3") } // Base3
func (t *LightTheme) Surface() lipgloss.Color    { return lipgloss.Color("#eee8d5") } // Base2
func (t *LightTheme) Text() lipgloss.Color       { return lipgloss.Color("#657b83") } // Base00
func (t *LightTheme) TextSubtle() lipgloss.Color { return lipgloss.Color("#93a1a1") } // Base1
func (t *LightTheme) Error() lipgloss.Color      { return lipgloss.Color("#dc322f") } // Red
func (t *LightTheme) Warning() lipgloss.Color    { return lipgloss.Color("#b58900") } // Yellow
func (t *LightTheme) Success() lipgloss.Color    { return lipgloss.Color("#859900") } // Green
func (t *LightTheme) Info() lipgloss.Color       { return lipgloss.Color("#2aa198") } // Cyan
func (t *LightTheme) Border() lipgloss.Color     { return lipgloss.Color("#93a1a1") } // Base1
func (t *LightTheme) Highlight() lipgloss.Color  { return lipgloss.Color("#586e75") } // Base01
func (t *LightTheme) Selection() lipgloss.Color  { return lipgloss.Color("#eee8d5") } // Base2

// ThemeManager manages the current theme
type ThemeManager struct {
	current Theme
	themes  map[string]Theme
}

// NewThemeManager creates a new theme manager with all available themes
func NewThemeManager() *ThemeManager {
	tm := &ThemeManager{
		themes: make(map[string]Theme),
	}

	// Register all themes
	tm.themes["retro"] = NewRetroTheme()
	tm.themes["dark"] = NewDarkTheme()
	tm.themes["light"] = NewLightTheme()

	// Set default theme
	tm.current = tm.themes["retro"]

	return tm
}

// Current returns the current active theme
func (tm *ThemeManager) Current() Theme {
	return tm.current
}

// SetTheme changes the current theme
func (tm *ThemeManager) SetTheme(name string) error {
	theme, exists := tm.themes[name]
	if !exists {
		return fmt.Errorf("theme '%s' not found", name)
	}

	tm.current = theme
	return nil
}

// ListThemes returns all available theme names
func (tm *ThemeManager) ListThemes() []string {
	names := make([]string, 0, len(tm.themes))
	for name := range tm.themes {
		names = append(names, name)
	}
	return names
}

// Global theme manager instance
var ThemeManagerInstance = NewThemeManager()

// Convenience function to get current theme
func CurrentTheme() Theme {
	return ThemeManagerInstance.Current()
}

// Convenience function to get current theme styles
func CurrentStyles() *ThemeStyles {
	return ThemeManagerInstance.Current().Styles()
}
