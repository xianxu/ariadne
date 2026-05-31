# `merge-checks.d/` — repo-local publish gate (ariadne #52)

Drop executable checks here. The base-layer runner (`scripts/run-merge-checks.sh`,
symlinked from ariadne) executes every executable in this directory at the publish
boundary — from CI on a PR to `main`, and from a repo's local `pre-push` hook.

## Contract

Each check is an executable invoked as:

```
./<check> <BASE_SHA> <HEAD_SHA>
```

- The range `BASE_SHA..HEAD_SHA` is what's being published (in CI it's
  `merge-base(base, head)..head`, i.e. exactly the PR's changes).
- **Exit 0 = pass, non-zero = fail.** A failing check fails the whole gate.
- Print human-readable findings to **stderr**.
- Checks run in **filename order** — prefix with `10-`, `20-` to order them.
- `README*` and `*.md` files are ignored.

An **empty** directory (no executable checks) is a **no-op pass** — repos with no
publish gate are unaffected.

## Example

`10-review-gate.sh` (you-decide) wraps its substrate gate:

```bash
#!/usr/bin/env bash
exec "$(git rev-parse --show-toplevel)/scripts/review-gate.sh" "$@"
```

## Where it runs

- **CI**: `.github/workflows/merge-check.yml` (seeded) calls the runner over the PR range.
- **Local**: a `pre-push` hook calls the same runner — so local and server gates never drift.

See `atlas/workflow/ci-merge-check.md` and `workshop/issues/000052`.
