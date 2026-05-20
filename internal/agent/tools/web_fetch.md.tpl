Fetch a web URL and return content as markdown for analysis workflows.

Usage:
- URL must be fully formed; this tool is read-only and does not modify project files.
- Large pages (>50KB) are saved to a temp markdown file; then use `view`/`grep` for targeted analysis.
- If redirected to a different host, rerun fetch with the final redirected URL.
{{- if .GhAvailable }} For GitHub content when an exact repo, issue, or PR link is provided, use `gh` CLI in bash instead.{{- end }}
