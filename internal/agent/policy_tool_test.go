package agent

import (
	"context"
	"testing"

	"charm.land/fantasy"
	"github.com/charmbracelet/crush/internal/agent/tools"
	"github.com/charmbracelet/crush/internal/message"
	"github.com/charmbracelet/crush/internal/pubsub"
	"github.com/stretchr/testify/require"
)

type fakeMessageService struct {
	listFn func(context.Context, string) ([]message.Message, error)
}

func (f *fakeMessageService) Subscribe(context.Context) <-chan pubsub.Event[message.Message] { return nil }
func (f *fakeMessageService) Create(context.Context, string, message.CreateMessageParams) (message.Message, error) {
	return message.Message{}, nil
}
func (f *fakeMessageService) Update(context.Context, message.Message) error { return nil }
func (f *fakeMessageService) Get(context.Context, string) (message.Message, error) {
	return message.Message{}, nil
}
func (f *fakeMessageService) List(ctx context.Context, sessionID string) ([]message.Message, error) {
	if f.listFn == nil {
		return nil, nil
	}
	return f.listFn(ctx, sessionID)
}
func (f *fakeMessageService) ListUserMessages(context.Context, string) ([]message.Message, error) {
	return nil, nil
}
func (f *fakeMessageService) ListAllUserMessages(context.Context) ([]message.Message, error) {
	return nil, nil
}
func (f *fakeMessageService) Delete(context.Context, string) error { return nil }
func (f *fakeMessageService) DeleteSessionMessages(context.Context, string) error {
	return nil
}
func (f *fakeMessageService) Flush(context.Context, string) error { return nil }
func (f *fakeMessageService) FlushAll(context.Context) error      { return nil }

func TestDenyRetryPolicyTool_BlocksAfterDenied(t *testing.T) {
	inner := &mockAgentTool{name: tools.EditToolName}
	svc := &fakeMessageService{
		listFn: func(context.Context, string) ([]message.Message, error) {
			return []message.Message{
				{
					Parts: []message.ContentPart{
						message.ToolResult{
							ToolCallID: "c1",
							Name:    tools.EditToolName,
							Content: "User denied permission",
							IsError: true,
						},
					},
				},
				{
					Parts: []message.ContentPart{
						message.ToolCall{
							ID:    "c1",
							Name:  tools.EditToolName,
							Input: `{"file_path":"/tmp/a.go","old_string":"x","new_string":"y"}`,
						},
					},
				},
			}, nil
		},
	}
	wrapped := newDenyRetryPolicyTool(inner, svc)
	ctx := context.WithValue(context.Background(), tools.SessionIDContextKey, "s1")
	resp, err := wrapped.Run(ctx, fantasy.ToolCall{
		Name:  tools.EditToolName,
		Input: `{"file_path":"/tmp/a.go","old_string":"x","new_string":"y"}`,
	})
	require.NoError(t, err)
	require.Contains(t, resp.Content, "identical call")
}

func TestDenyRetryPolicyTool_AllowsWhenNoDeniedHistory(t *testing.T) {
	inner := &mockAgentTool{name: tools.EditToolName}
	svc := &fakeMessageService{
		listFn: func(context.Context, string) ([]message.Message, error) {
			return []message.Message{}, nil
		},
	}
	wrapped := newDenyRetryPolicyTool(inner, svc)
	ctx := context.WithValue(context.Background(), tools.SessionIDContextKey, "s1")
	resp, err := wrapped.Run(ctx, fantasy.ToolCall{Name: tools.EditToolName})
	require.NoError(t, err)
	require.Contains(t, resp.Content, "ok")
}

func TestDenyRetryPolicyTool_AllowsDifferentInputAfterDenied(t *testing.T) {
	inner := &mockAgentTool{name: tools.EditToolName}
	svc := &fakeMessageService{
		listFn: func(context.Context, string) ([]message.Message, error) {
			return []message.Message{
				{
					Parts: []message.ContentPart{
						message.ToolResult{
							ToolCallID: "c1",
							Name:       tools.EditToolName,
							Content:    "User denied permission",
							IsError:    true,
						},
					},
				},
				{
					Parts: []message.ContentPart{
						message.ToolCall{
							ID:    "c1",
							Name:  tools.EditToolName,
							Input: `{"file_path":"/tmp/a.go","old_string":"x","new_string":"y"}`,
						},
					},
				},
			}, nil
		},
	}
	wrapped := newDenyRetryPolicyTool(inner, svc)
	ctx := context.WithValue(context.Background(), tools.SessionIDContextKey, "s1")
	resp, err := wrapped.Run(ctx, fantasy.ToolCall{
		Name:  tools.EditToolName,
		Input: `{"file_path":"/tmp/a.go","old_string":"x","new_string":"z"}`,
	})
	require.NoError(t, err)
	require.Contains(t, resp.Content, "ok")
}

func TestDenyRetryPolicyTool_BlocksAfterHookBlock(t *testing.T) {
	inner := &mockAgentTool{name: tools.EditToolName}
	svc := &fakeMessageService{
		listFn: func(context.Context, string) ([]message.Message, error) {
			return []message.Message{
				{
					Parts: []message.ContentPart{
						message.ToolResult{
							ToolCallID: "c1",
							Name:       tools.EditToolName,
							Content:    "Tool call blocked by hook. Reason: policy denied",
							IsError:    true,
						},
					},
				},
				{
					Parts: []message.ContentPart{
						message.ToolCall{
							ID:    "c1",
							Name:  tools.EditToolName,
							Input: `{"file_path":"/tmp/a.go","old_string":"x","new_string":"y"}`,
						},
					},
				},
			}, nil
		},
	}
	wrapped := newDenyRetryPolicyTool(inner, svc)
	ctx := context.WithValue(context.Background(), tools.SessionIDContextKey, "s1")
	resp, err := wrapped.Run(ctx, fantasy.ToolCall{
		Name:  tools.EditToolName,
		Input: `{"file_path":"/tmp/a.go","old_string":"x","new_string":"y"}`,
	})
	require.NoError(t, err)
	require.Contains(t, resp.Content, "identical call")
}

func TestDenyRetryPolicyTool_BlocksAfterExactMatchFailure(t *testing.T) {
	inner := &mockAgentTool{name: tools.EditToolName}
	svc := &fakeMessageService{
		listFn: func(context.Context, string) ([]message.Message, error) {
			return []message.Message{
				{
					Parts: []message.ContentPart{
						message.ToolResult{
							ToolCallID: "c1",
							Name:       tools.EditToolName,
							Content:    "Edit failed: old_string not found in file",
							IsError:    true,
						},
					},
				},
				{
					Parts: []message.ContentPart{
						message.ToolCall{
							ID:    "c1",
							Name:  tools.EditToolName,
							Input: `{"file_path":"/tmp/a.go","old_string":"x","new_string":"y"}`,
						},
					},
				},
			}, nil
		},
	}
	wrapped := newDenyRetryPolicyTool(inner, svc)
	ctx := context.WithValue(context.Background(), tools.SessionIDContextKey, "s1")
	resp, err := wrapped.Run(ctx, fantasy.ToolCall{
		Name:  tools.EditToolName,
		Input: `{"file_path":"/tmp/a.go","old_string":"x","new_string":"y"}`,
	})
	require.NoError(t, err)
	require.Contains(t, resp.Content, "identical call")
}

func TestDenyRetryPolicyTool_BlocksAfterRepeatedIdenticalFailures(t *testing.T) {
	inner := &mockAgentTool{name: tools.EditToolName}
	svc := &fakeMessageService{
		listFn: func(context.Context, string) ([]message.Message, error) {
			return []message.Message{
				{
					Parts: []message.ContentPart{
						message.ToolResult{ToolCallID: "c1", Name: tools.EditToolName, Content: "error 1", IsError: true},
						message.ToolResult{ToolCallID: "c2", Name: tools.EditToolName, Content: "error 2", IsError: true},
					},
				},
				{
					Parts: []message.ContentPart{
						message.ToolCall{ID: "c1", Name: tools.EditToolName, Input: `{"file_path":"/tmp/a.go","old_string":"x","new_string":"y"}`},
						message.ToolCall{ID: "c2", Name: tools.EditToolName, Input: `{"file_path":"/tmp/a.go","old_string":"x","new_string":"y"}`},
					},
				},
			}, nil
		},
	}
	wrapped := newDenyRetryPolicyTool(inner, svc)
	ctx := context.WithValue(context.Background(), tools.SessionIDContextKey, "s1")
	resp, err := wrapped.Run(ctx, fantasy.ToolCall{
		Name:  tools.EditToolName,
		Input: `{"file_path":"/tmp/a.go","old_string":"x","new_string":"y"}`,
	})
	require.NoError(t, err)
	require.Contains(t, resp.Content, "Repeated identical failed call")
}
