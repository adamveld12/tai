package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/adamveld12/tai/internal/llm"
	"github.com/adamveld12/tai/internal/state"
)

type agentImpl struct {
	output chan AgentStatus
	name   string
	llm.Provider
	state.Dispatcher
}

func (a *agentImpl) messageHandler(ctx context.Context, input state.Message) {
	dispatcher := a.Dispatcher

	dispatcher.Dispatch(ChatCompletionStartedAction{
		Role:    input.Role,
		Content: input.Content,
	})

	os := dispatcher.GetState()
	req := llm.ChatRequest{
		Messages:     os.Context.Messages,
		Model:        os.Model.Name,
		SystemPrompt: state.SystemPrompt(os),
	}

	var err error
	var res <-chan llm.ChatStreamChunk
	if res, err = a.StreamChatCompletion(context.Background(), req); err != nil {
		a.Dispatch(ChatCompletionCompletedAction{
			Success: false,
			Error:   err,
		})
		return
	}

	go func() {
		outMsgTime := time.Now()
		var outMsg state.Message
		for chunk := range res {
			select {
			case <-ctx.Done():
				err = ctx.Err()
				a.Dispatch(ChatCompletionCompletedAction{
					Success: false,
					Error:   err,
				})
				return
			default:
				if chunk.Error != nil {
					err = ctx.Err()
					a.Dispatch(ChatCompletionCompletedAction{
						Success: false,
						Error:   err,
					})
					return
				} else {
					chunkMsg := state.Message{
						Role:      state.RoleAssistant,
						Content:   chunk.Delta,
						Timestamp: outMsgTime,
						Usage: state.TokenUsage{
							Prompt:     chunk.Usage.PromptTokens,
							Completion: chunk.Usage.CompletionTokens,
							Total:      chunk.Usage.TotalTokens,
						},
					}
					updatedContent := fmt.Sprintf("%s%s", outMsg.Content, chunk.Delta)
					outMsg = chunkMsg
					outMsg.Content = updatedContent

					a.Dispatch(MessageChunkAction{Message: chunkMsg})
				}
			}
		}

		a.Dispatch(ChatCompletionCompletedAction{Success: true, Message: outMsg})
	}()
}

func (a *agentImpl) OnStateChange(action state.Action, newState, oldState state.AppState) {
	switch action := action.(type) {
	case ChatCompletionStartedAction:
	case ChatCompletionCompletedAction:
		a.output <- AgentStatus{Success: action.Success, Error: action.Error, Message: action.Message}
	case TerminateAgentAction:
		close(a.output)
	}
}

func (a *agentImpl) Start(ctx context.Context, input chan state.Message) <-chan AgentStatus {
	go func() {
		for {
			select {
			case msg, ok := <-input:
				if !ok {
					return
				}
				a.messageHandler(ctx, msg)
			case <-ctx.Done():
				return
			}
		}
	}()

	return a.output
}

func (a *agentImpl) String() string {
	return fmt.Sprintf("Agent(%s)", a.name)
}

type TaskInput struct {
	llm.Provider
	state.Dispatcher
	Name             string
	SystemPrompt     string
	WorkingDirectory string
}

func Task(input TaskInput) (Agent, error) {
	if input.Provider == nil {
		return nil, fmt.Errorf("no provider specified")
	}

	if input.Name == "" {
		return nil, fmt.Errorf("no name specified")
	}

	ag := &agentImpl{
		output:   make(chan AgentStatus),
		name:     input.Name,
		Provider: input.Provider,
	}

	if input.Dispatcher != nil {
		state := input.Dispatcher.GetState()
		if input.WorkingDirectory == "" {
			input.WorkingDirectory = state.Context.WorkingDirectory
		}
	} else {
		input.Dispatcher = state.NewMemoryState(
			input.SystemPrompt,
			input.WorkingDirectory,
			ag.String(),
		)
	}

	if input.WorkingDirectory == "" {
		return nil, fmt.Errorf("no working directory specified")
	}

	ag.Dispatcher = input.Dispatcher

	return ag, nil
}
