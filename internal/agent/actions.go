package agent

import (
	"fmt"
	"time"

	"github.com/adamveld12/tai/internal/state"
)

type ChatCompletionStartedAction struct {
	Role    state.Role `json:"role"`
	Content string     `json:"content"`
}

func (a ChatCompletionStartedAction) Execute(s state.AppState) (state.AppState, error) {
	s.Model.Busy = true
	s.Context.Messages = append(s.Context.Messages, state.Message{
		Role:      a.Role,
		Content:   a.Content,
		Timestamp: time.Now(),
	})
	return s, nil
}

type ChatCompletionCompletedAction struct {
	Success bool
	state.Message
	Error error
}

func (a ChatCompletionCompletedAction) Execute(s state.AppState) (state.AppState, error) {
	s.Model.Busy = false
	return s, nil
}

type MessageChunkAction struct{ state.Message }

func (a MessageChunkAction) Execute(s state.AppState) (state.AppState, error) {
	// find last assistant message with same ID and append to it
	for idx, msg := range s.Context.Messages {
		if msg.Role == a.Role && msg.Timestamp.Equal(a.Timestamp) {
			a.Content = fmt.Sprintf("%s%s", msg.Content, a.Content)
			s.Context.Messages = append(s.Context.Messages[:idx], a.Message)
			s.Context.Updated = time.Now()
			return s, nil
		}
	}

	// If no existing message found, append as new
	s.Context.Messages = append(s.Context.Messages, a.Message)

	return s, nil
}

type TerminateAgentAction struct{ Reason string }

func (a TerminateAgentAction) Execute(s state.AppState) (state.AppState, error) {
	// TODO: No op for now, but could be used to clean up resources or notify the user in the state
	return s, nil
}
