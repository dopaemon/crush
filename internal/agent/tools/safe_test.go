package tools

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainsCommandChaining(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"plain ls", "ls -la", false},
		{"plain echo", "echo hello world", false},
		{"plain pwd", "pwd", false},
		{"plain git status", "git status", false},
		{"ls with redirect", "ls > /tmp/out", false},
		{"ls with pipe", "ls | grep foo", true},
		{"ls with double ampersand", "ls && echo done", true},
		{"ls with semicolon", "ls; echo done", true},
		{"ls with pipe pipe", "ls || echo fail", true},
		{"ls with backticks", "ls `echo foo`", true},
		{"ls with subshell", "ls $(echo foo)", true},
		{"ls with background ampersand", "ls & echo done", false},
		{"rm -rf with && ls (rm first)", "rm -rf / && ls", true},
		{"redirect with ampersand gt", "ls &> /dev/null", false},
		{"redirect with gt ampersand", "ls >& /dev/null", false},
		{"simple kill", "kill 1234", false},
		{"kill with pipe", "kill 1234 | echo foo", true},
		{"git log", "git log --oneline", false},
		{"git log with pipe", "git log | head", true},
		{"empty string", "", false},
		{"dollar sign in argument", "echo $HOME", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := containsCommandChaining(tt.input)
			assert.Equal(t, tt.expected, got, "containsCommandChaining(%q)", tt.input)
		})
	}
}

func TestDedicatedToolHintForCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		cmd       string
		wantHint  bool
		wantMatch string
	}{
		{name: "cat blocked", cmd: "cat foo.txt", wantHint: true, wantMatch: "view"},
		{name: "grep blocked", cmd: "grep foo bar.txt", wantHint: true, wantMatch: "grep` tool"},
		{name: "find blocked", cmd: "find . -name '*.go'", wantHint: true, wantMatch: "glob"},
		{name: "sed blocked", cmd: "sed -n '1,10p' file", wantHint: true, wantMatch: "edit"},
		{name: "ls allowed", cmd: "ls -la", wantHint: false},
		{name: "chained still blocked by first command", cmd: "cat a.txt | wc -l", wantHint: true, wantMatch: "view"},
		{name: "empty", cmd: "   ", wantHint: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			hint := dedicatedToolHintForCommand(tt.cmd)
			if tt.wantHint {
				assert.NotEmpty(t, hint)
				assert.Contains(t, strings.ToLower(hint), strings.ToLower(tt.wantMatch))
				return
			}
			assert.Empty(t, hint)
		})
	}
}
