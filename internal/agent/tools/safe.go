package tools

import (
	"runtime"
	"slices"
	"strings"
)

var safeCommands = []string{
	// Bash builtins and core utils
	"cal",
	"date",
	"df",
	"du",
	"echo",
	"env",
	"free",
	"groups",
	"hostname",
	"id",
	"kill",
	"killall",
	"ls",
	"nice",
	"nohup",
	"printenv",
	"ps",
	"pwd",
	"set",
	"time",
	"timeout",
	"top",
	"type",
	"uname",
	"unset",
	"uptime",
	"whatis",
	"whereis",
	"which",
	"whoami",

	// Git
	"git blame",
	"git branch",
	"git config --get",
	"git config --list",
	"git describe",
	"git diff",
	"git grep",
	"git log",
	"git ls-files",
	"git ls-remote",
	"git remote",
	"git rev-parse",
	"git shortlog",
	"git show",
	"git status",
	"git tag",
}

var chainingMetacharacters = []string{
	";",
	"|",
	"&&",
	"$(",
	"`",
}

// containsCommandChaining reports whether s contains shell metacharacters
// that enable command chaining or substitution.
func containsCommandChaining(s string) bool {
	return slices.ContainsFunc(chainingMetacharacters, func(c string) bool {
		return strings.Contains(s, c)
	})
}

var dedicatedToolPreferredCommands = map[string]string{
	"cat":  "Use `view` to read files.",
	"head": "Use `view` with an appropriate `limit`.",
	"tail": "Use `view` with an appropriate `offset` and `limit`.",
	"find": "Use `glob` or `ls` for file discovery.",
	"grep": "Use `grep` tool for content search.",
	"rg":   "Use `grep` tool for content search.",
	"sed":  "Use `edit`/`multiedit` for file edits and `view` for reading.",
	"awk":  "Use dedicated tools (`view`/`grep`/`edit`) instead of shell text processing.",
}

func dedicatedToolHintForCommand(cmd string) string {
	trimmed := strings.TrimSpace(strings.ToLower(cmd))
	if trimmed == "" {
		return ""
	}
	parts := strings.Fields(trimmed)
	if len(parts) == 0 {
		return ""
	}
	if hint, ok := dedicatedToolPreferredCommands[parts[0]]; ok {
		return hint
	}
	return ""
}

func init() {
	if runtime.GOOS == "windows" {
		safeCommands = append(
			safeCommands,
			// Windows-specific commands
			"ipconfig",
			"nslookup",
			"ping",
			"systeminfo",
			"tasklist",
			"where",
		)
	}
}
