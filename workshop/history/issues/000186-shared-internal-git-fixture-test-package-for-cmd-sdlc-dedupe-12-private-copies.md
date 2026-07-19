---
id: 000186
status: done
deps: []
github_issue:
created: 2026-07-17
updated: 2026-07-19
estimate_hours: 2.07
started: 2026-07-19T11:17:40-07:00
actual_hours: 3.67
---

# shared internal git-fixture test package for cmd/sdlc (dedupe ~12 private copies)

## Problem

cmd/sdlc tests carry ~12 private copies of the same git-fixture idiom
(init temp repo on main, config user, initial commit) — closeRepo,
hermeticRepo, initFleetRepo/gitIn (peerwrite_apply_test.go), migrate/resolve
fixtures, plus near-identical writeProject helpers in discover_test.go vs
projectfind_test.go. Flagged by the #171 M3 boundary review and again at the
#171 issue-close review.

## Spec

Apply **ARCH-DRY** to the cmd/sdlc test suite. Two duplications:

1. **The git runner** — "run `git <args>` in a dir, fatal on error" exists as
   near-identical variants: `git()` (merge_e2e_test.go:30, already shared by 6
   files but oddly housed), `gitIn()` (peerwrite_apply_test.go:14), `captureGit()`
   (milestonewindow_test.go:88), and inline `run`/`runGit` closures — including in
   the sub-packages `internal/gitx/window_test.go` + `internal/activetime/parity_test.go`.
2. **The repo-setup sequence** — `init [-b main]` + config identity + gpgsign off
   (+ optional initial commit / chdir / subdir): copy-pasted across `hermeticRepo`,
   `initFleetRepo`, `windowRepo`, `closeRepo`, `publishRepo`, `close_test.go`×3,
   `migrate_test.go`×3, **plus two sub-package copies** — `internal/gitx/window_test.go`
   and `internal/activetime/parity_test.go` (`gitInit`). The sub-package copies use
   bare `init -q`; standardizing on `testfix.Repo`'s `-b main` is a determinism
   improvement (branch name is irrelevant to CommitWindow / active-time), not a
   behavior change.

Extract `cmd/sdlc/internal/testfix` (following the `internal/{gitx,project,…}`
precedent; stdlib-only, so no import cycle) exposing `Git`, `Capture`, and
`Repo(t, opts...)` with `Chdir()`, `InitialCommit()`, `At(parent,name)` options;
delegate every private copy. Test-only, behavior-preserving — the suite staying
green IS the acceptance test.

**Kept local** (genuine specializations, not copies): bare-origin (`init --bare`)
and remote-only (`gitInit`, fetch_test.go) init, and the dated-commit helpers
that inject `GIT_AUTHOR_DATE`/`GIT_COMMITTER_DATE` (`activetime`'s `gitCommit`,
`gitx`'s dated commit) — their runner/init halves still delegate, only the
date-injecting/bare/remote step stays local (Simplicity-First).

**Red herring** (plan-quality Finding 2): the original Problem's "near-identical
`writeProject` in discover_test.go vs projectfind_test.go" is stale —
`projectfind_test.go` has no `writeProject`; the real `discover_test.go` pair
already delegates (`writeProject` → `writeProjectStatus`) and writes markdown, not
git fixtures. Nothing to consolidate there.

Design detail: `workshop/plans/000186-testfix-package-plan.md`. Not a base-layer
concern — that convention governs runtime code and is intact; this is test
scaffolding.

## Estimate

Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against
`baseline-v3.1.md`. Method A only. Plan-doc-backed → +15% design buffer.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: greenfield-go-module    design=0.5 impl=0.25
item: cross-cutting-refactor  design=0.3 impl=0.25
item: cross-cutting-refactor  design=0.3 impl=0.3
design-buffer: 0.15
total: 2.07
```

The greenfield item is the small `testfix` package; the two refactor items are
the runner-unification sweep and the setup-sequence delegation sweep — each now
spanning the two `internal/*` sub-packages the plan-quality gate surfaced (+0.27h
vs the first 1.80h draft).

## Done when

- One internal test-fixture package (`cmd/sdlc/internal/testfix`) owns the
  git-repo/runner idiom; the private copies delegate to it.
- No behavior change; `go test ./cmd/sdlc/...` stays green.

## Plan

- [x] create `cmd/sdlc/internal/testfix/testfix.go` — `Git`, `Capture`, `Repo` + `Chdir`/`InitialCommit`/`At` options; package doc crediting #186
- [x] runner dedup: replace `git()`/`gitIn()`/`captureGit()`/inline `run`/`runGit` with `testfix.Git`/`testfix.Capture` at every call site — top-level `cmd/sdlc/*_test.go` **and** `internal/gitx`, `internal/activetime`
- [x] setup dedup: delegate `hermeticRepo`/`initFleetRepo`/`windowRepo`/`closeRepo`/`publishRepo` bodies + inline `close_test`×3/`migrate`×3 setups + `internal/gitx/window_test.go` + `internal/activetime` `gitInit` to `testfix.Repo(...)`; dated-commit/bare/remote steps stay local
- [x] `go build ./... && go test ./cmd/sdlc/... ./pkg/vocab/` green (`./cmd/sdlc/...` covers the internal sub-packages); `rg` confirms no remaining private setup/runner copies except the documented locals (bare-origin, remote-only `gitInit`, dated-commit helpers)

## Log

### 2026-07-19 — implemented (ARCH-DRY)
- 2026-07-19: closed — go build ./... clean; go vet ./cmd/sdlc/... clean; go test ./cmd/sdlc/... ./pkg/vocab/ -count=1 all green (cmd/sdlc 101s + every sub-package). Behavior-preserving test-only refactor: extracted cmd/sdlc/internal/testfix (Git/Capture/Repo), delegated ~15 private git-fixture copies, net -255 lines across 13 files. rg confirms only the documented bare-origin/remote-only locals + the single testfix source remain. --no-atlas: no new product surface/flow/terminology — testfix is internal test scaffolding.; review verdict: FIX-THEN-SHIP

Extracted `cmd/sdlc/internal/testfix` (`Git`, `Capture`, `Repo` + `Chdir`/
`InitialCommit`/`At`); delegated every private copy. Net **−255 lines across 13
test files** + the new ~120-line package.

- **Runner unification** (`git`/`gitIn`/`captureGit`/`gitOut` + inline `run`/
  `runGit` closures): converted to one-line delegations to `testfix.Git`/
  `Capture` — the ~127 call sites (`git(t,` ×86, `gitIn` ×28, `captureGit` ×13)
  stay untouched, exactly the "the copies delegate to it" Done-when. Preserved
  the two trim contracts (`merge_e2e`'s `git`, `gitx`'s `git` wrap
  `strings.TrimSpace` over the untrimmed `testfix.Git`).
- **Setup delegation**: `hermeticRepo`, `initFleetRepo`, `windowRepo`,
  `closeRepo`, `publishRepo` + inline `close_test`×3 / `migrate` (`mkRepo` + 2
  inline) / `internal/gitx` (×2) / `internal/activetime` `gitInit`.
- **Extra copies found during impl** (beyond the plan-quality gate's gitx +
  activetime): `activetime_test.go`'s `atGitRepo`, `collectdiff_test.go`, and
  `issue_test.go`'s second (no-origin) site — so the real count was ~15, not
  ~12. All delegated. The bound-closure helpers (`windowRepo`/`publishRepo`)
  keep their public `func(...string)` signature via a small closure over
  `testfix.Git` (the plan-quality advisory).
- **Kept local (documented)**: bare-origin push harnesses (`merge_e2e`
  `tempRepo`, `issue_test`'s AutoSyncsToMainCleanTree) and remote-only
  `fetch_test` `gitInit`; the explicit-date commit helpers
  (`activetime`/`gitx`) — only their init+config halves delegate. `rg` confirms
  the sole remaining `commit.gpgsign` copies are those two bare-origin harnesses
  + the single `testfix.go` source.
- **Determinism nudge**: the two sub-package copies used bare `init -q`;
  `testfix.Repo` standardizes on `-b main` (branch name is irrelevant to
  CommitWindow / active-time — verified green).

Verification: `go build ./...` clean; `go vet` clean; `go test ./cmd/sdlc/...
./pkg/vocab/ -count=1` all green (`cmd/sdlc` 101s + every sub-package).

### 2026-07-17
