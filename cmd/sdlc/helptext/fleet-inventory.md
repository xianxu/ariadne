Collect every eligible sibling repository and canonical Git worktree from the
fleet containing the caller. `--path` selects the caller vantage and defaults to `.`.
Nested directories, linked worktrees, and symlinked vantages normalize to
the same fleet identity.

The default output is a deterministic human view of the typed inventory.
`--json` emits the JSON contract. Repository-scoped failures remain in
`diagnostics` while unaffected repositories and worktrees continue to render.

This command reports measured Git facts and declared policy capability only. It
does not infer coldness, drift, actor liveness, staleness, or admission keys.
