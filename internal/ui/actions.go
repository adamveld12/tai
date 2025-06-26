package ui

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/adamveld12/tai/internal/llm"
	"github.com/adamveld12/tai/internal/state"
)

func NewMessage(d state.Dispatcher, provider llm.Provider, role state.Role, content string) error {
	d.Dispatch(ChatCompletionStartedAction{})

	d.Dispatch(MessageAction{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	})

	go func() {
		s := d.GetState()
		req := llm.ChatRequest{
			Messages:     s.Context.Messages,
			Model:        s.Model.Name,
			SystemPrompt: state.SystemPrompt(s),
		}

		startedAt := time.Now()
		d.Dispatch(MessageAction{
			Role:      state.RoleAssistant,
			Timestamp: startedAt,
		})

		// res, err := provider.ChatCompletion(context.Background(), req)
		res, err := provider.StreamChatCompletion(context.Background(), req)

		if err != nil {
			log.Fatalf("Failed to get chat completion: %v", err)
		}

		for chunk := range res {
			if chunk.Error != nil {
				break
			} else {
				d.Dispatch(MessageChunkAction{
					Message: state.Message{
						Role:      state.RoleAssistant,
						Content:   chunk.Delta,
						Timestamp: startedAt,
						Usage: state.TokenUsage{
							Prompt:     chunk.Usage.PromptTokens,
							Completion: chunk.Usage.CompletionTokens,
							Total:      chunk.Usage.TotalTokens,
						},
					},
				})
			}
		}

		d.Dispatch(ChatCompletionCompletedAction{})
	}()

	return nil
}

type MessageChunkAction struct {
	state.Message
}

func (a MessageChunkAction) Execute(s state.AppState) (state.AppState, error) {
	// find last assistant message with same ID and append to it
	for idx, msg := range s.Context.Messages {
		if msg.Role == a.Role && msg.Timestamp.Equal(a.Timestamp) {
			a.Content = fmt.Sprintf("%s%s", msg.Content, a.Content)
			s.Context.Messages = append(s.Context.Messages[:idx], a.Message)
			s.Context.Updated = time.Now()
			break
		}
	}

	return s, nil
}

type MessageAction state.Message

func (a MessageAction) Execute(s state.AppState) (state.AppState, error) {
	s.Context.Messages = append(s.Context.Messages, state.Message(a))
	s.Context.Updated = time.Now()
	return s, nil
}

func ClearMessages(d state.Dispatcher) {
	d.Dispatch(ClearMessagesAction{})
}

type ClearMessagesAction struct{}

type SwitchThemeAction struct {
	Name string
}

func (a ClearMessagesAction) Execute(s state.AppState) (state.AppState, error) {
	s.Context.Messages = []state.Message{}
	s.Context.Updated = time.Now()
	return s, nil
}

type ChatCompletionStartedAction struct{}

func (a ChatCompletionStartedAction) Execute(s state.AppState) (state.AppState, error) {
	s.Model.Busy = true
	return s, nil
}

type ChatCompletionCompletedAction struct{}

func (a ChatCompletionCompletedAction) Execute(s state.AppState) (state.AppState, error) {
	s.Model.Busy = false
	return s, nil
}

type ChangeProviderAction struct {
	Provider string
	Name     string
}

func (a ChangeProviderAction) Execute(s state.AppState) (state.AppState, error) {
	s.Model.Provider = a.Provider
	s.Model.Name = a.Name
	s.Model.Busy = true
	return s, nil

}
