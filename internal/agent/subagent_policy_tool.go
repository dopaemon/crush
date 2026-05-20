package agent

import (
	"context"

	"charm.land/fantasy"
	"github.com/charmbracelet/crush/internal/agent/tools"
)

type subagentPolicyTool struct {
	inner      fantasy.AgentTool
	isSubAgent bool
}

func newSubagentPolicyTool(inner fantasy.AgentTool, isSubAgent bool) fantasy.AgentTool {
	return &subagentPolicyTool{inner: inner, isSubAgent: isSubAgent}
}

func (s *subagentPolicyTool) Info() fantasy.ToolInfo {
	return s.inner.Info()
}

func (s *subagentPolicyTool) ProviderOptions() fantasy.ProviderOptions {
	return s.inner.ProviderOptions()
}

func (s *subagentPolicyTool) SetProviderOptions(opts fantasy.ProviderOptions) {
	s.inner.SetProviderOptions(opts)
}

func (s *subagentPolicyTool) Run(ctx context.Context, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
	if s.isSubAgent && (call.Name == AgentToolName || call.Name == tools.AgenticFetchToolName) {
		return fantasy.NewTextErrorResponse("Subagent policy: do not re-delegate from within a subagent turn."), nil
	}
	return s.inner.Run(ctx, call)
}
