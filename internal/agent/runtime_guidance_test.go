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

	guidance := buildRuntimeSystemGuidance(agentTools, false, tools.EditToolName, false)
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
