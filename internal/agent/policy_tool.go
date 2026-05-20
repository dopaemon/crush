package agent

import (
	"context"
	"strings"

	"charm.land/fantasy"
	"github.com/charmbracelet/crush/internal/agent/tools"
	"github.com/charmbracelet/crush/internal/message"
)

type denyRetryPolicyTool struct {
	inner    fantasy.AgentTool
	messages message.Service
}

func newDenyRetryPolicyTool(inner fantasy.AgentTool, messages message.Service) fantasy.AgentTool {
	return &denyRetryPolicyTool{inner: inner, messages: messages}
}

func (d *denyRetryPolicyTool) Info() fantasy.ToolInfo {
	return d.inner.Info()
}

func (d *denyRetryPolicyTool) ProviderOptions() fantasy.ProviderOptions {
	return d.inner.ProviderOptions()
}

func (d *denyRetryPolicyTool) SetProviderOptions(opts fantasy.ProviderOptions) {
	d.inner.SetProviderOptions(opts)
}

func (d *denyRetryPolicyTool) Run(ctx context.Context, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
	sessionID := tools.GetSessionFromContext(ctx)
	if sessionID != "" {
		msgs, err := d.messages.List(ctx, sessionID)
		if err == nil {
			for i := len(msgs) - 1; i >= 0 && len(msgs)-i <= 30; i-- {
				for _, tr := range msgs[i].ToolResults() {
					if tr.Name != call.Name {
						continue
					}
					if strings.Contains(strings.ToLower(tr.Content), "user denied permission") {
						return fantasy.NewTextErrorResponse("Previous call to this tool was denied by user. Use a different approach or ask user intent before retrying."), nil
					}
				}
			}
		}
	}
	return d.inner.Run(ctx, call)
}

