package tools

import "context"

// FileTool represents a file manipulation tool
type FileTool interface {
	// ReadFile reads the contents of a file
	ReadFile(ctx context.Context, path string) (string, error)
	// WriteFile writes content to a file
	WriteFile(ctx context.Context, path string, content string) error
	// SearchFile searches for a term in a file
	SearchFile(ctx context.Context, path string, term string) ([]string, error)
}

// ShellTool represents a shell command execution tool
type ShellTool interface {
	// RunCommand executes a shell command and returns its output
	RunCommand(ctx context.Context, command string) (string, error)
	// StreamCommand executes a shell command and streams its output
	StreamCommand(ctx context.Context, command string) (<-chan string, error)
}

// WebTool represents a web data fetching tool
type WebTool interface {
	// FetchURL fetches data from a URL
	FetchURL(ctx context.Context, url string) (string, error)
}

// GitTool represents a Git operation tool
type GitTool interface {
	// Status checks the current status of the repository
	Status(ctx context.Context) (string, error)
	// Diff shows changes between commits, commit and working tree, etc.
	Diff(ctx context.Context) (string, error)
	// Commit records changes to the repository
	Commit(ctx context.Context, message string) error
	// Branch shows the current branch status
	Branch(ctx context.Context) (string, error)
}
