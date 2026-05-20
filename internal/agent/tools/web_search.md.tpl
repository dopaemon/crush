Search the web via DuckDuckGo for up-to-date information beyond local knowledge.

Usage:
- Returns titles, URLs, and snippets; follow up with `web_fetch` to inspect full page content.
- After using these results in a final user-facing answer, include a `Sources:` section listing relevant URLs as markdown links.
- Use current year/month context in queries when user asks for latest/current information.
{{- if .GhAvailable }} For GitHub searches when an exact repo name, issue, or link is provided, use `gh search` in bash instead.{{- end }}
