package agent

import (
	"context"
	"strings"
	"testing"

	"charm.land/fantasy"
	"github.com/charmbracelet/crush/internal/agent/tools"
	"github.com/charmbracelet/crush/internal/message"
	"github.com/stretchr/testify/require"
)

type mockAgentTool struct {
	name string
	opts fantasy.ProviderOptions
}

func (m *mockAgentTool) SetProviderOptions(opts fantasy.ProviderOptions) {
	m.opts = opts
}

func (m *mockAgentTool) ProviderOptions() fantasy.ProviderOptions {
	return m.opts
}

func (m *mockAgentTool) Name() string {
	return m.name
}

func (m *mockAgentTool) Info() fantasy.ToolInfo {
	return fantasy.ToolInfo{Name: m.name}
}

func (m *mockAgentTool) Run(context.Context, fantasy.ToolCall) (fantasy.ToolResponse, error) {
	return fantasy.NewTextResponse("ok"), nil
}

func TestLastDeniedToolFromHistory(t *testing.T) {
	msgs := []message.Message{
		{
			Parts: []message.ContentPart{
				message.ToolResult{
					Name:    tools.ViewToolName,
					Content: "ok",
					IsError: false,
				},
			},
		},
		{
			Parts: []message.ContentPart{
				message.ToolResult{
					Name:    tools.BashToolName,
					Content: "User denied permission",
					IsError: true,
				},
			},
		},
	}

	require.Equal(t, tools.BashToolName, lastDeniedToolFromHistory(msgs))
}

func TestBuildRuntimeSystemGuidance_IncludesDeniedToolHint(t *testing.T) {
	agentTools := []fantasy.AgentTool{
		&mockAgentTool{name: tools.ViewToolName},
		&mockAgentTool{name: tools.EditToolName},
	}

	guidance := buildRuntimeSystemGuidance(agentTools, false, tools.EditToolName, false, false, false, false, false, false, false, false)
	require.True(t, strings.Contains(guidance, "recently denied permission"))
	require.True(t, strings.Contains(guidance, tools.EditToolName))
}

func TestHasNonTrivialRecentImplementation(t *testing.T) {
	msgs := []message.Message{
		{
			Parts: []message.ContentPart{
				message.ToolCall{Name: tools.ViewToolName},
				message.ToolCall{Name: tools.EditToolName},
			},
		},
		{
			Parts: []message.ContentPart{
				message.ToolCall{Name: tools.MultiEditToolName},
			},
		},
		{
			Parts: []message.ContentPart{
				message.ToolCall{Name: tools.WriteToolName},
			},
		},
	}
	require.True(t, hasNonTrivialRecentImplementation(msgs))
}

func TestHasRecentVerificationFailure(t *testing.T) {
	msgs := []message.Message{
		{
			Parts: []message.ContentPart{
				message.ToolResult{
					Name:    tools.BashToolName,
					Content: "ok   github.com/charmbracelet/crush/internal/agent/tools  (cached)\nFAIL\tgithub.com/charmbracelet/crush/internal/agent\t0.05s",
					IsError: false,
				},
			},
		},
	}
	require.True(t, hasRecentVerificationFailure(msgs))
}

func TestBuildRuntimeSystemGuidance_NonInteractive(t *testing.T) {
	agentTools := []fantasy.AgentTool{
		&mockAgentTool{name: tools.ViewToolName},
	}
	guidance := buildRuntimeSystemGuidance(agentTools, false, "", false, false, false, false, false, false, false, true)
	require.True(t, strings.Contains(guidance, "non-interactive run"))
}

func TestHasRecentHookBlock(t *testing.T) {
	msgs := []message.Message{
		{
			Parts: []message.ContentPart{
				message.ToolResult{
					Name:    tools.EditToolName,
					Content: "Tool call blocked by hook. Reason: policy denied",
					IsError: true,
				},
			},
		},
	}
	require.True(t, hasRecentHookBlock(msgs))
}

func TestHasRecentAuthBlock(t *testing.T) {
	msgs := []message.Message{
		{
			Parts: []message.ContentPart{
				message.ToolResult{
					Name:    tools.BashToolName,
					Content: "Error: authentication required. Please login first.",
					IsError: true,
				},
			},
		},
	}
	require.True(t, hasRecentAuthBlock(msgs))
}

func TestHasRecentExactMatchEditFailure(t *testing.T) {
	msgs := []message.Message{
		{
			Parts: []message.ContentPart{
				message.ToolResult{
					Name:    tools.EditToolName,
					Content: "Edit failed: old_string not found in file",
					IsError: true,
				},
			},
		},
	}
	require.True(t, hasRecentExactMatchEditFailure(msgs))
}

func TestHasRecentRepeatedToolPattern(t *testing.T) {
	msgs := []message.Message{
		{
			Parts: []message.ContentPart{
				message.ToolCall{Name: tools.GrepToolName, Input: `{"pattern":"foo"}`},
			},
		},
		{
			Parts: []message.ContentPart{
				message.ToolCall{Name: tools.GrepToolName, Input: `{"pattern":"foo"}`},
			},
		},
		{
			Parts: []message.ContentPart{
				message.ToolCall{Name: tools.GrepToolName, Input: `{"pattern":"foo"}`},
			},
		},
	}
	require.True(t, hasRecentRepeatedToolPattern(msgs))
}

func TestHasRecentRiskyShellIntent(t *testing.T) {
	msgs := []message.Message{
		{
			Parts: []message.ContentPart{
				message.ToolCall{
					Name:  tools.BashToolName,
					Input: `{"command":"git reset --hard HEAD~1"}`,
				},
			},
		},
	}
	require.True(t, hasRecentRiskyShellIntent(msgs))
}
