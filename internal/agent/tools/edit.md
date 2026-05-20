Performs exact string replacements in files.

Usage:
- You must read the target file first (`view`) before editing.
- Preserve exact indentation and whitespace from the file content; do not include line-number prefixes from view output in `old_string`/`new_string`.
- ALWAYS prefer editing existing files in the codebase. NEVER write new files unless explicitly required.
- Avoid adding emojis to files unless the user explicitly requests it.
- The edit fails if `old_string` is not unique in the file; provide more unique context or use `replace_all`.
- Use `replace_all` for file-wide renames/replacements when all matches should change.
