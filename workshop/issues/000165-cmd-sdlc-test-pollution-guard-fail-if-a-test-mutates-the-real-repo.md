---
id: 000165
status: open
deps: []
github_issue:
created: 2026-07-05
updated: 2026-07-05
estimate_hours:
---

# cmd/sdlc test-pollution guard: fail if a test mutates the real repo

## Problem

`cmd/sdlc`'s tests build throwaway git repos (merge e2e, issue e2e, close, repolock,
activetime, …). They isolate via `cmd.Dir=<temp>` / `-C <temp>` and/or `os.Chdir`
into a `t.TempDir()`. But there is **no backstop**: a test that forgets to isolate
(a `git(t, "", …)` with empty dir → process cwd, a `chdir`-without-restore, a
relative path, or a git op that finds the enclosing repo) silently operates on the
**live ariadne repo** — and `go test ./cmd/sdlc/` runs from `cmd/sdlc/`, which is
*inside* it.

This is not hypothetical. During #148 (2026-07-05) a test-harness git sequence
(`add -A && commit -m seed` → `switch -c feature` → `commit -m work` → `switch main`
→ `merge feature`) leaked into the real repo: it committed a stray `f` + a bare
`origin.git/` + swept the untracked `workshop/` files, then fast-forwarded `main`
onto the junk — corrupting **local main AND `origin/main`** (recovery required a
force-push + cherry-pick salvage of the real #148 commits). The committed code was
later proven clean (a severed-clone re-run of the full suite produced zero
pollution), so the trigger was a transient intermediate state — but nothing would
have *caught* it, and nothing prevents the next one.

## Spec

Add a cheap, once-per-package harness assertion (a `TestMain` in `cmd/sdlc`, the
natural home) that FAILS LOUDLY if the package's tests mutate the real repository:

- **Resolve the real repo up front** — in `TestMain`, before `m.Run()`, capture the
  real repo path via `git rev-parse --show-toplevel` (resolved from the initial cwd,
  BEFORE any test `os.Chdir`). Snapshot: `HEAD` sha, current branch, and
  `git status --porcelain` (all `-C <realroot>` so a test's chdir can't fool it).
- **Re-check after `m.Run()`** — recompute the same three; if HEAD moved, the branch
  changed, or the porcelain set gained entries not present before, fail with a loud,
  specific message naming what changed (so the offending test run is obvious) and a
  non-zero exit. Pre-existing untracked files present in the *before* snapshot must
  not trip it (only NEW mutations).
- **Cheap** — three `git` calls before + after the whole package, not per-test.
- Consider a shared helper so the other repo-building packages (`internal/gitx`,
  `internal/activetime`, `cmd/sdlc/internal/…`) can opt in with one line.

Design subtlety: the guard must resolve the real root from the ORIGINAL cwd
(`TestMain` runs before any test chdir), and always query with `-C <realroot>` — a
test that chdir'd into a temp dir and forgot to restore would otherwise make the
after-check read the temp repo and miss the real-repo damage.

## Done when

- `cmd/sdlc` `TestMain` fails the package run (non-zero) if the real repo's HEAD,
  branch, or tracked/untracked set changed across `m.Run()`; pre-existing untracked
  files don't trip it.
- A deliberately-polluting fake test (guarded, e.g. behind an env flag) proves the
  guard FIRES — the guard is itself tested, not just asserted to work.
- The guard is cheap (≈constant git calls per package, not per test).
- (Optional) a shared helper offered to the other repo-building test packages.

## Plan

- [ ]

## Log

### 2026-07-05

Filed from the #148 landing session. The incident: a transient test-harness git
sequence corrupted local + origin `main` (force-push recovery). The committed code
is clean (severed-clone repro), so this is a *backstop* against the class, not a
fix for a known-live test — the exact polluting trigger was never pinpointed, which
is itself the argument for a guard: a future leak should FAIL a test, not silently
rewrite `main`. Related: this surfaced while shipping #148 (the `sdlc merge`
reused-branch guard) — thematically adjacent (both are "refuse loudly instead of
silently doing the wrong thing"), but a distinct, separable concern.
