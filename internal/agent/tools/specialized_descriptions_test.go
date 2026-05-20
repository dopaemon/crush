package tools

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGlobDescription_IncludesSpecializedGuidance(t *testing.T) {
	t.Parallel()
	desc := globDescription()
	require.Contains(t, desc, "Fast file pattern matching for large codebases")
	require.Contains(t, desc, "prefer delegating to `agent`")
}

func TestWebSearchDescription_IncludesSourcesGuidance(t *testing.T) {
	t.Parallel()
	desc := renderToolDescription(webSearchDescriptionTpl)
	require.Contains(t, desc, "include a `Sources:` section")
	require.Contains(t, desc, "Use current year/month context in queries")
}

func TestWebFetchDescription_IncludesRedirectAndLargePageGuidance(t *testing.T) {
	t.Parallel()
	desc := renderToolDescription(webFetchDescriptionTpl)
	require.Contains(t, desc, "Large pages (>50KB) are saved to a temp markdown file")
	require.Contains(t, desc, "rerun fetch with the final redirected URL")
}

func TestViewDescription_IncludesReadSpecificGuidance(t *testing.T) {
	t.Parallel()
	desc := viewDescription()
	require.Contains(t, desc, "Supports offset and line limit for targeted reads")
	require.Contains(t, desc, "Can read notebook files (`.ipynb`)")
	require.Contains(t, desc, "for directories use `ls`")
}
