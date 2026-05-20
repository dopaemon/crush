Read a file from the local filesystem by path.

Usage:
- Supports offset and line limit for targeted reads; default reads up to {{ .DefaultReadLimit }} lines from start.
- Returns line-numbered text output; preserve exact whitespace after the line-number prefix when preparing edits.
- Maximum returned content section is {{ .MaxViewSizeKB }}KB.
- Can render images (PNG, JPEG, GIF, WebP) for visual inspection.
- Can read notebook files (`.ipynb`) and return cell content with outputs.
- This tool reads files only; for directories use `ls`.
- For screenshots provided by path, use this tool directly on that file path.
