package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"charm.land/fantasy"
	"github.com/charmbracelet/crush/internal/agent/tools"
	"github.com/charmbracelet/crush/internal/message"
)

var riskyBashPatterns = []string{
	"rm -rf",
	"rm -fr",
	"kill -9",
	"pkill -9",
	"git reset --hard",
	"git checkout --",
	"git checkout -f",
	"git restore --source",
	"git branch -d",
	"git branch -D",
	"git clean -fd",
	"git clean -f",
	"git clean -xdf",
	"git rebase",
	"git commit --amend",
	"--no-verify",
	"git push",
	"git push --force",
	"git push --force-with-lease",
	"git push -f",
}

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
	case tools.ViewToolName, tools.EditToolName, tools.MultiEditToolName, tools.WriteToolName:
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

func toolCommand(toolName, input string) string {
	if toolName != tools.BashToolName {
		return ""
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(input), &payload); err != nil {
		return ""
	}
	cmd, _ := payload["command"].(string)
	return strings.TrimSpace(strings.ToLower(cmd))
}

func isRiskyBashCommand(cmd string) bool {
	if cmd == "" {
		return false
	}
	for _, pattern := range riskyBashPatterns {
		if strings.Contains(cmd, pattern) {
			return true
		}
	}
	return false
}

func isVerificationAgentCall(input string) bool {
	var payload map[string]any
	if err := json.Unmarshal([]byte(input), &payload); err != nil {
		return false
	}
	raw, _ := payload["subagent_type"].(string)
	return strings.EqualFold(strings.TrimSpace(raw), verificationAgentType)
}

func verificationPromptHasRequiredContext(input string) bool {
	var payload map[string]any
	if err := json.Unmarshal([]byte(input), &payload); err != nil {
		return false
	}
	prompt, _ := payload["prompt"].(string)
	p := strings.ToLower(prompt)
	hasOriginalRequest := strings.Contains(p, "original request")
	hasChangedFiles := strings.Contains(p, "changed files") || strings.Contains(p, "files changed")
	hasApproach := strings.Contains(p, "approach")
	return hasOriginalRequest && hasChangedFiles && hasApproach
}

func verifierContractSatisfied(msgs []message.Message) bool {
	if recentVerificationVerdict(msgs) != "pass" {
		return false
	}
	if !hasVerifierCommandEvidence(msgs) {
		return false
	}
	if hasVerifierDivergenceSignal(msgs) {
		return false
	}
	return true
}

func verifierContractNeedsFollowup(msgs []message.Message) bool {
	v := recentVerificationVerdict(msgs)
	if v == "fail" || v == "partial" {
		return true
	}
	if v == "pass" && (!hasVerifierCommandEvidence(msgs) || hasVerifierDivergenceSignal(msgs)) {
		return true
	}
	return false
}

func normalizeJSONInput(input string) string {
	var v any
	if err := json.Unmarshal([]byte(input), &v); err != nil {
		return input
	}
	normalized, err := json.Marshal(v)
	if err != nil {
		return input
	}
	return string(normalized)
}

func sameToolInput(a, b string) bool {
	return normalizeJSONInput(a) == normalizeJSONInput(b)
}

func maxStructuredOutputRetries() int {
	raw := strings.TrimSpace(os.Getenv("MAX_STRUCTURED_OUTPUT_RETRIES"))
	if raw == "" {
		return 5
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return 5
	}
	return n
}

func looksLikePromptInjection(content string) bool {
	c := strings.ToLower(content)
	return strings.Contains(c, "ignore previous instructions") ||
		strings.Contains(c, "disregard all prior instructions") ||
		(strings.Contains(c, "system prompt") && strings.Contains(c, "reveal")) ||
		(strings.Contains(c, "you are now") && strings.Contains(c, "assistant"))
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
					if looksLikePromptInjection(content) {
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
						// Block only semantically identical retries of denied calls.
						if sameToolInput(tc.Input, call.Input) {
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
			if call.Name == "structured_output" {
				retries := 0
				for i := len(msgs) - 1; i >= 0 && len(msgs)-i <= 200; i-- {
					for _, tc := range msgs[i].ToolCalls() {
						if tc.Name == "structured_output" {
							retries++
						}
					}
				}
				maxRetries := maxStructuredOutputRetries()
				if retries >= maxRetries {
					return fantasy.NewTextErrorResponse(fmt.Sprintf("Structured output retry limit reached (%d). Stop retrying unchanged and adjust approach.", maxRetries)), nil
				}
			}
			if call.Name == AgentToolName &&
				hasNonTrivialRecentImplementation(msgs) &&
				!verifierContractSatisfied(msgs) &&
				!isVerificationAgentCall(call.Input) {
				return fantasy.NewTextErrorResponse("Verification contract active: run `agent` with `subagent_type=\"verification\"` before other delegation."), nil
			}
			if call.Name == AgentToolName &&
				verifierContractNeedsFollowup(msgs) &&
				!isVerificationAgentCall(call.Input) {
				return fantasy.NewTextErrorResponse("Verification follow-up required: latest verifier verdict is not fully satisfied. Continue with verification subagent before other delegation."), nil
			}
			if call.Name == AgentToolName &&
				isVerificationAgentCall(call.Input) &&
				hasNonTrivialRecentImplementation(msgs) &&
				!verificationPromptHasRequiredContext(call.Input) {
				return fantasy.NewTextErrorResponse("Verification contract active: verification prompt must include original request, changed files, and approach."), nil
			}
			switch call.Name {
			case tools.EditToolName, tools.MultiEditToolName:
				path := toolFilePath(call.Name, call.Input)
				if path != "" {
					if _, statErr := os.Stat(path); statErr == nil && !hasRecentViewForPath(msgs, path) {
						return fantasy.NewTextErrorResponse("Read-before-edit policy: call view on this file before editing."), nil
					}
				}
			case tools.WriteToolName:
				path := toolFilePath(tools.WriteToolName, call.Input)
				if path != "" {
					if _, statErr := os.Stat(path); statErr == nil && !hasRecentViewForPath(msgs, path) {
						return fantasy.NewTextErrorResponse("Safe-overwrite policy: view existing file before write overwrite."), nil
					}
				}
			case tools.BashToolName:
				if isRiskyBashCommand(toolCommand(call.Name, call.Input)) {
					return fantasy.NewTextErrorResponse("Risky shell action blocked pending explicit user confirmation. Ask the user to confirm before running destructive commands."), nil
				}
			}
		}
	}
	return d.inner.Run(ctx, call)
}
