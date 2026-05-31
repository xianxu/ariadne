# CI merge-check — the pluggable publish gate

A generic, server-side gate every ariadne derivative inherits: on a PR to `main`,
run the repo's own checks over the PR's changes and report (and, if made required,
block). Each derivative plugs in its own logic; a repo with no checks is a no-op.
Introduced in #52; pairs with the in-place-branch workflow (#51) and the LLM
judges (`pre-merge-checks.md`).

## Three layers (and why)

Two CI-isolation facts force the split: GitHub Actions does **not** discover
symlinked *workflow* files, and an isolated CI checkout (the consumer repo only)
does **not** contain the base-layer sibling that a *script* symlink points at:

| Layer | File | Delivery | Why |
|---|---|---|---|
| Shim | `.github/workflows/merge-check.yml` | `seed` | Real file per repo (Actions ignores symlinked workflows); thin + stable so it rarely changes |
| Runner | `scripts/run-merge-checks.sh` | `symlink` → `../<upstream>` | Generic logic, single-source via the symlink; but the target is the base-layer *sibling*, absent in an isolated CI checkout — so the shim clones peers first (below) |
| Checks | `scripts/merge-checks.d/*` | `scaffold` | Repo-owned; each derivative drops its own (or none) |

The shim first runs `BOOTSTRAP_CLONE_ONLY=1 ./bootstrap.sh` to clone the upstream
peer chain as siblings (so the runner symlink resolves; `CLONE_ONLY` skips the
heavy `make bootstrap` handoff; guarded so the root repo is a no-op). Then it
computes the PR range and calls the runner; the runner discovers and runs every
executable in `merge-checks.d/`, aggregating exit codes. **Empty dir = pass.**
(The clone step was added after the first live PR exercise — you-decide#4 / #53
Phase E — showed the symlinked runner `exit 127`ing in CI.)

## The check contract

```
scripts/merge-checks.d/10-foo.sh  <BASE_SHA>  <HEAD_SHA>     # exit 0 pass, non-0 fail
```

Run in filename order; `README*`/`*.md` ignored; findings to stderr. The shim passes
`merge-base(base, head)` as `<BASE_SHA>` so a two-dot `base..head` diff is exactly the
PR's changes.

## One runner, two call sites

`run-merge-checks.sh` is invoked by **CI** (the shim) *and* by a repo's local
**`pre-push` hook** — so local fast-feedback and the server gate run the identical
check-set and can't drift.

## Strictness is opt-in

Without branch protection the check is **advisory** (runs + reports; doesn't block the
merge button; direct pushes to `main` stay open). Making `merge-check` a *required*
status check — and creating the remote repo — is the `make remote-init` operator verb
(#52 M2; needs `gh` auth). Mechanism works advisory-only without it.

## Not the same as the LLM judges

These are **deterministic, server-side, repo-specific** (a substrate gate, a lint, a
test). The `sdlc judge` / `parallel-checks.sh` judges are **LLM, local, generic**
(plan/specs/lessons). Two complementary tiers — see `pre-merge-checks.md`.

## Example consumer

you-decide drops `merge-checks.d/10-review-gate.sh` (wraps its `review: passed`
substrate gate) and points its `pre-push` hook at the runner. See you-decide#4 M3.
