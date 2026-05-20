A powerful search tool built on ripgrep.

Usage:
- ALWAYS use this `grep` tool for search tasks; do not run `grep` or `rg` through shell when this tool is available.
- Supports full regex syntax (for example `log.*Error` or `function\\s+\\w+`).
- Filter files with `include` patterns (for example `*.js`, `**/*.{ts,tsx}`).
- Use `literal_text=true` when searching exact strings that contain regex metacharacters.
- For open-ended multi-round investigation, delegate with `agent` instead of repeatedly shell-grepping.
- Returns matching file paths sorted by modification time (max {{ .MaxResults }} results), respects ignore files, and skips hidden files.
