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

