Create or overwrite a file with given content; auto-creates parent dirs. Cannot append.

Usage:
- If this is an existing file, you MUST read it first (`view`) before writing; this tool rejects stale-overwrite attempts.
- Prefer `edit`/`multiedit` for partial modifications; use `write` for new files or complete rewrites.
- NEVER create documentation files (`*.md`) or README files unless the user explicitly requests them.
- Avoid writing emojis unless the user explicitly requests them.
