Inspect the repository fleet from one caller path.

`sdlc fleet inventory` reports typed facts for every discovered Git worktree.
`sdlc fleet policy` resolves the admission policy for one prospective path.
Both commands accept `--json`; without it they render the same typed result for
humans. `--path` defaults to `.`.

This command is read-only and does not acquire the Git transaction lock.
