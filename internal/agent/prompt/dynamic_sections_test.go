package prompt

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/crush/internal/config"
)

func TestPromptBuild_IncludesDynamicSections(t *testing.T) {
	tmpl := "{{.DynamicBoundary}}\n{{.LanguageSection}}\n{{.OutputStyleSection}}\n{{.MemorySection}}\n{{.SessionGuidanceSection}}\n{{.EnvInfoSection}}\n{{.SummarizeToolResultsSection}}\n{{.ScratchpadSection}}\n{{.TokenBudgetSection}}\n{{.BriefSection}}"
	p, err := NewPrompt(
		"test",
		tmpl,
		WithTimeFunc(func() time.Time { return time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC) }),
		WithWorkingDir(t.TempDir()),
	)
	if err != nil {
		t.Fatalf("new prompt: %v", err)
	}

	brief := true
	store := config.NewTestStore(&config.Config{
		Options: &config.Options{
			Language:          "Vietnamese",
			OutputStylePrompt: "Keep responses terse.",
			ContextPaths:      []string{},
			ScratchpadDirectory: "/tmp/crush-scratch/session-1",
			TokenBudgetTarget: "500k",
			BriefMode: &brief,
		},
	})

	out, err := p.Build(context.Background(), "openai", "gpt-5", store)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if !strings.Contains(out, "# Language\nAlways respond in Vietnamese.") {
		t.Fatalf("missing language section: %q", out)
	}
	if !strings.Contains(out, "# Output Style\nKeep responses terse.") {
		t.Fatalf("missing output style section: %q", out)
	}
	if strings.Contains(out, "# Memory") {
		t.Fatalf("unexpected memory section without context files: %q", out)
	}
	if !strings.Contains(out, "# Session Guidance") {
		t.Fatalf("missing session guidance section: %q", out)
	}
	if !strings.Contains(out, "# Environment Info") {
		t.Fatalf("missing env info section: %q", out)
	}
	if !strings.Contains(out, "# Tool Result Summaries") {
		t.Fatalf("missing summarize tool results section: %q", out)
	}
	if !strings.Contains(out, "# Scratchpad") {
		t.Fatalf("missing scratchpad section: %q", out)
	}
	if !strings.Contains(out, "# Token Budget") {
		t.Fatalf("missing token budget section: %q", out)
	}
	if !strings.Contains(out, "# Brief Mode") {
		t.Fatalf("missing brief section: %q", out)
	}
}
