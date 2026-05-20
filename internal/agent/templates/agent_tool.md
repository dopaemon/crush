Use the agent tool for independent side tasks that benefit from parallelism or context isolation.

Guidelines:
- Prefer direct local tools for simple, targeted lookups.
- Delegate broader exploration, long-running searches, or bounded implementation chunks.
- Do not duplicate delegated work in the main thread.
- If multiple delegated tasks are independent, run them in parallel.
- Keep delegated prompts concrete: objective, scope, expected output.
