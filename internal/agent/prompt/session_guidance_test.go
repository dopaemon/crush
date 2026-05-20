package prompt

import (
	"strings"
	"testing"

	"github.com/charmbracelet/crush/internal/config"
	"github.com/stretchr/testify/require"
)

func TestRenderSessionGuidance_IncludesInteractiveCommandAndNoColonRule(t *testing.T) {
	g := renderSessionGuidance([]string{"bash", "agent"}, false, &config.Options{})
	require.True(t, strings.Contains(g, "`! <command>`"))
	require.True(t, strings.Contains(g, "Do not use a trailing colon right before a tool call preamble sentence"))
}

func TestRenderSessionGuidance_IncludesAskUserQuestionHintWhenAvailable(t *testing.T) {
	g := renderSessionGuidance([]string{"bash", "ask_user_question"}, false, &config.Options{})
	require.True(t, strings.Contains(g, "use `ask_user_question` to ask a focused clarification"))
}
