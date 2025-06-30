package agent

import (
	"context"

	"github.com/adamveld12/tai/internal/state"
)

type AgentStatus struct {
	Success bool
	Message state.Message
	Error   error
}

type Agent interface {
	Start(context.Context, chan state.Message) <-chan AgentStatus
	String() string
}
