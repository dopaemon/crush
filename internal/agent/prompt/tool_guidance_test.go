package prompt

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRenderToolGuidance(t *testing.T) {
	g := renderToolGuidance([]string{"view", "edit", "bash", "agent"}, true)
	require.True(t, strings.Contains(g, "Read file contents with `view`"))
	require.True(t, strings.Contains(g, "Use `edit`/`multiedit` for modifications"))
	require.True(t, strings.Contains(g, "Use `bash` only for shell/system operations"))
	require.True(t, strings.Contains(g, "Delegate independent side tasks with `agent`"))
	require.True(t, strings.Contains(g, "You are a subagent"))
}
