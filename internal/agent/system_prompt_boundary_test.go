package agent

import (
	"testing"

	"charm.land/fantasy"
	"github.com/charmbracelet/crush/internal/csync"
)

func TestSetSystemPromptSplitsBoundary(t *testing.T) {
	a := &sessionAgent{
		systemPrompt:        csync.NewValue(""),
		staticSystemPrompt:  csync.NewValue(""),
		dynamicSystemPrompt: csync.NewValue(""),
		tools:               csync.NewSlice[fantasy.AgentTool](),
	}
	input := "static part\n__SYSTEM_PROMPT_DYNAMIC_BOUNDARY__\ndynamic part"
	a.SetSystemPrompt(input)
	if got := a.staticSystemPrompt.Get(); got != "static part" {
		t.Fatalf("unexpected static prompt: %q", got)
	}
	if got := a.dynamicSystemPrompt.Get(); got != "dynamic part" {
		t.Fatalf("unexpected dynamic prompt: %q", got)
	}
}

