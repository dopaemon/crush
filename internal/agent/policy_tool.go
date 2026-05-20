package agent

import (
	"context"
	"encoding/json"
	"strings"

	"charm.land/fantasy"
	"github.com/charmbracelet/crush/internal/agent/tools"
	"github.com/charmbracelet/crush/internal/message"
)

func toolFilePath(toolName, input string) string {
	var payload map[string]any
	if err := json.Unmarshal([]byte(input), &payload); err != nil {
		return ""
	}
	path, _ := payload["file_path"].(string)
	if path == "" {
		return ""
	}
	switch toolName {
	case tools.ViewToolName, tools.EditToolName, tools.MultiEditToolName:
		return path
	default:
		return ""
	}
}

func hasRecentViewForPath(msgs []message.Message, path string) bool {
	if path == "" {
		return false
	}
	for i := len(msgs) - 1; i >= 0 && len(msgs)-i <= 60; i-- {
		for _, tc := range msgs[i].ToolCalls() {
			if tc.Name != tools.ViewToolName {
				continue
			}
			if toolFilePath(tc.Name, tc.Input) == path {
				return true
			}
		}
	}
	return false
}

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
			recentFailedIdenticalCalls := 0
			deniedCallIDs := map[string]struct{}{}
			for i := len(msgs) - 1; i >= 0 && len(msgs)-i <= 30; i-- {
				for _, tr := range msgs[i].ToolResults() {
					if tr.Name != call.Name || tr.ToolCallID == "" {
						continue
					}
					content := strings.ToLower(tr.Content)
					if strings.Contains(content, "user denied permission") ||
						strings.Contains(content, "tool call blocked by hook") ||
						strings.Contains(content, "turn halted by hook") {
						deniedCallIDs[tr.ToolCallID] = struct{}{}
						continue
					}
					if (tr.Name == tools.EditToolName || tr.Name == tools.MultiEditToolName || tr.Name == tools.WriteToolName) &&
						(strings.Contains(content, "old_string") && strings.Contains(content, "not found") ||
							(strings.Contains(content, "exact") && strings.Contains(content, "match") && strings.Contains(content, "fail"))) {
						deniedCallIDs[tr.ToolCallID] = struct{}{}
					}
					if tr.IsError {
						deniedCallIDs[tr.ToolCallID] = struct{}{}
					}
				}
			}
			if len(deniedCallIDs) > 0 {
				for i := len(msgs) - 1; i >= 0 && len(msgs)-i <= 30; i-- {
					for _, tc := range msgs[i].ToolCalls() {
						if tc.Name != call.Name || tc.ID == "" {
							continue
						}
						if _, denied := deniedCallIDs[tc.ID]; !denied {
							continue
						}
						// Block only identical retries of denied calls.
						if tc.Input == call.Input {
							recentFailedIdenticalCalls++
							if recentFailedIdenticalCalls >= 2 {
								return fantasy.NewTextErrorResponse("Repeated identical failed call detected. Change approach (different input/tool) instead of retrying unchanged."), nil
							}
						}
					}
				}
				if recentFailedIdenticalCalls == 1 {
					return fantasy.NewTextErrorResponse("Previous identical call was denied or failed. Use a different approach or clarify intent before retrying."), nil
				}
			}
			switch call.Name {
			case tools.EditToolName, tools.MultiEditToolName:
				path := toolFilePath(call.Name, call.Input)
				if path != "" && !hasRecentViewForPath(msgs, path) {
					return fantasy.NewTextErrorResponse("Read-before-edit policy: call view on this file before editing."), nil
				}
			}
		}
	}
	return d.inner.Run(ctx, call)
}
