package prompt

import "testing"

func TestSplitByDynamicBoundary_NoBoundary(t *testing.T) {
	staticPart, dynamicPart := SplitByDynamicBoundary("hello")
	if staticPart != "hello" || dynamicPart != "" {
		t.Fatalf("unexpected split: static=%q dynamic=%q", staticPart, dynamicPart)
	}
}

func TestSplitByDynamicBoundary_WithBoundary(t *testing.T) {
	input := "A\nB\n" + SystemPromptDynamicBoundary + "\nC\nD\n"
	staticPart, dynamicPart := SplitByDynamicBoundary(input)
	if staticPart != "A\nB" {
		t.Fatalf("unexpected static: %q", staticPart)
	}
	if dynamicPart != "C\nD\n" {
		t.Fatalf("unexpected dynamic: %q", dynamicPart)
	}
}

