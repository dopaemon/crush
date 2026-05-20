package prompt

import "testing"

func TestSystemPromptDynamicBoundary(t *testing.T) {
	if SystemPromptDynamicBoundary == "" {
		t.Fatalf("boundary constant must not be empty")
	}
	if SystemPromptDynamicBoundary != "__SYSTEM_PROMPT_DYNAMIC_BOUNDARY__" {
		t.Fatalf("unexpected boundary value: %q", SystemPromptDynamicBoundary)
	}
}

