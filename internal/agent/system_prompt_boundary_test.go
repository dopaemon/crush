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

func TestComposeSystemPrompt(t *testing.T) {
	if got := composeSystemPrompt("", "dyn"); got != "dyn" {
		t.Fatalf("unexpected compose result: %q", got)
	}
	if got := composeSystemPrompt("sta", ""); got != "sta" {
		t.Fatalf("unexpected compose result: %q", got)
	}
	if got := composeSystemPrompt("sta", "dyn"); got != "sta\n\ndyn" {
		t.Fatalf("unexpected compose result: %q", got)
	}
}

func TestAppendDynamicSection(t *testing.T) {
	if got := appendDynamicSection("", "x"); got != "x" {
		t.Fatalf("unexpected append result: %q", got)
	}
	if got := appendDynamicSection("a", ""); got != "a" {
		t.Fatalf("unexpected append result: %q", got)
	}
	if got := appendDynamicSection("a", "b"); got != "a\n\nb" {
		t.Fatalf("unexpected append result: %q", got)
	}
	if got := appendDynamicSection("a\n", "\nb\n"); got != "a\n\nb" {
		t.Fatalf("unexpected append result with trim: %q", got)
	}
}
