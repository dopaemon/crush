Fast file pattern matching for large codebases.

Usage:
- Use this tool to find files by name/path pattern (for example `**/*.js` or `src/**/*.ts`).
- Results are sorted by modification time, skip hidden files, and are capped at {{ .MaxResults }}.
- Use `grep` for content search inside files.
- For open-ended multi-round discovery that needs repeated glob+grep cycles, prefer delegating to `agent`.
