Apply multiple exact find-and-replace edits to one file in a single operation; edits run sequentially.

Usage:
- Read the target file first (`view`) before applying multi-edit changes.
- Preserve exact indentation and whitespace; never include view line-number prefixes in edit strings.
- Prefer editing existing files; avoid creating new files unless explicitly required.
- Avoid adding emojis to files unless the user explicitly requests it.
- Each `old_string` must be unique unless intentionally using `replace_all`.
- Prefer `multiedit` over repeated single `edit` calls when changing multiple locations in the same file.
