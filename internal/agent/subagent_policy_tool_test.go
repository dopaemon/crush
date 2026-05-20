package agent

import (
	"context"
	"testing"

	"charm.land/fantasy"
	"github.com/charmbracelet/crush/internal/agent/tools"
	"github.com/stretchr/testify/require"
)

func TestSubagentPolicyTool_BlocksAgentToolInSubagent(t *testing.T) {
	inner := &mockAgentTool{name: AgentToolName}
	wrapped := newSubagentPolicyTool(inner, true)
	resp, err := wrapped.Run(context.Background(), fantasy.ToolCall{Name: AgentToolName})
	require.NoError(t, err)
	require.Contains(t, resp.Content, "do not re-delegate")
}

func TestSubagentPolicyTool_AllowsOtherToolsInSubagent(t *testing.T) {
	inner := &mockAgentTool{name: tools.ViewToolName}
	wrapped := newSubagentPolicyTool(inner, true)
	resp, err := wrapped.Run(context.Background(), fantasy.ToolCall{Name: tools.ViewToolName})
	require.NoError(t, err)
	require.Contains(t, resp.Content, "ok")
}

func TestSubagentPolicyTool_BlocksAgenticFetchInSubagent(t *testing.T) {
	inner := &mockAgentTool{name: tools.AgenticFetchToolName}
	wrapped := newSubagentPolicyTool(inner, true)
	resp, err := wrapped.Run(context.Background(), fantasy.ToolCall{Name: tools.AgenticFetchToolName})
	require.NoError(t, err)
	require.Contains(t, resp.Content, "do not re-delegate")
}
