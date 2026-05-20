package prompt

import "strings"

// SplitByDynamicBoundary splits a rendered system prompt into static and
// dynamic sections using [SystemPromptDynamicBoundary]. The returned static
// part excludes the boundary marker. If marker is not present, whole prompt is
// treated as static and dynamic is empty.
func SplitByDynamicBoundary(rendered string) (staticPart, dynamicPart string) {
	idx := strings.Index(rendered, SystemPromptDynamicBoundary)
	if idx < 0 {
		return rendered, ""
	}
	staticPart = strings.TrimRight(rendered[:idx], "\n")
	after := rendered[idx+len(SystemPromptDynamicBoundary):]
	dynamicPart = strings.TrimLeft(after, "\n")
	return staticPart, dynamicPart
}

