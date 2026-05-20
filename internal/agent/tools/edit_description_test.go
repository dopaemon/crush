package tools

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEditToolDescription_IncludesSpecializedGuards(t *testing.T) {
	t.Parallel()
	require.Contains(t, editDescription, "You must read the target file first (`view`) before editing")
	require.Contains(t, editDescription, "NEVER write new files unless explicitly required")
	require.Contains(t, editDescription, "Avoid adding emojis to files unless the user explicitly requests it")
	require.Contains(t, editDescription, "old_string` is not unique")
}

