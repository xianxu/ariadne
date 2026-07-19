---
type: plan
issue: 000186
slug: testfix-package
created: 2026-07-19
---

# Plan: shared `testfix` git-fixture package for cmd/sdlc (#186)

## Goal

Apply **ARCH-DRY** to the cmd/sdlc test suite: the "init a throwaway git repo,
config a test identity, gpgsign off, (optionally) make an initial commit" idiom
is copy-pasted across the suite, and the "run git in a dir, fatal on error"
runner exists in several near-identical variants. Extract one internal package
that owns both; delegate every private copy to it. **Test-only,
behavior-preserving** — the suite staying green IS the acceptance test.

Not a base-layer/binary concern (that convention governs runtime code and is
intact) — this is test scaffolding sprawl, flagged by the #171 M3 boundary
review and the #171 close review.

## Two duplications being consolidated

1. **The runner** — "run `git <args>` in a dir, fail the test on error":
   `git()` (merge_e2e_test.go:30, already shared by 6 files but housed oddly),
   `gitIn()` (peerwrite_apply_test.go:14), `captureGit()`
   (milestonewindow_test.go:88), and inline `run`/`runGit` closures in
   `closeRepo`/`windowRepo`/`publishRepo`/`close_test.go` **and in the
   sub-packages** `internal/gitx/window_test.go` + `internal/activetime/parity_test.go`.

2. **The repo-setup sequence** — `init [-b main]` + `config user.email/name` +
   `config commit.gpgsign false` (+ optional initial commit / chdir / subdir):
   `hermeticRepo`, `initFleetRepo`, `windowRepo`, `closeRepo`, `publishRepo`,
   `close_test.go` (×3), `migrate_test.go` (×3), **plus two sub-package copies
   the first plan draft missed** (found by the plan-quality gate) —
   `internal/gitx/window_test.go:20-32` and `internal/activetime/parity_test.go`
   (`gitInit`, 12-30). The sub-package copies use bare `init -q` (no `-b main`);
   standardizing them on `testfix.Repo`'s `-b main` is a *determinism improvement*
   (branch name is irrelevant to CommitWindow / active-time, which key off commit
   dates) — verify green, not a behavior change.

## Design — `cmd/sdlc/internal/testfix`

Location follows the existing `cmd/sdlc/internal/{gitx,project,activetime,…}`
precedent. It is importable by `cmd/sdlc` and by sibling `cmd/sdlc/internal/*`
packages (Go internal rule); it imports only stdlib, so no cycle. Small, focused
API (Simplicity-First — cover the real idiom, not every git incantation):

```go
package testfix

// Git runs `git <args>` in dir (dir=="" → cwd), fatal on error; returns combined output.
func Git(t *testing.T, dir string, args ...string) string
// Capture runs `git <args>` in dir and returns stdout, fatal on error.
func Capture(t *testing.T, dir string, args ...string) string

// Repo inits a throwaway repo (init -q -b main + test identity + gpgsign off)
// and returns its path. Options tune placement, chdir, and an initial commit.
func Repo(t *testing.T, opts ...Option) string
func Chdir() Option                 // chdir into the repo; restore cwd on cleanup
func InitialCommit() Option         // write README + commit "init" (so first #N commit has a parent)
func At(parent, name string) Option // create at parent/name (default: t.TempDir())
```

**Kept local (genuine specializations, not copies of the idiom):** bare-origin
repos (`init --bare`, issue_test.go / merge_e2e_test.go), remote-only init
(`gitInit` in fetch_test.go), and the dated-commit helpers that inject explicit
`GIT_AUTHOR_DATE`/`GIT_COMMITTER_DATE` (`activetime/parity_test.go`'s `gitCommit`,
`gitx/window_test.go`'s dated commit). Their *runner* and *init+config* halves
still delegate to `testfix.Git` / `testfix.Repo`; only the date-injecting or
bare/remote step stays local. Forcing these into `Repo` options would over-fit
the API.

**Out of scope entirely — the `writeProject` red herring (plan-quality Finding 2):**
the issue's original Problem cited "near-identical `writeProject` helpers in
discover_test.go vs projectfind_test.go." Verified false: `projectfind_test.go`
has no `writeProject`; the real pair in `discover_test.go:12,26` *already*
delegates (`writeProject` → `writeProjectStatus`), and both write project
markdown, not git fixtures. Nothing to consolidate; not a `testfix` concern.

## Plan

- [ ] create `cmd/sdlc/internal/testfix/testfix.go` — `Git`, `Capture`, `Repo` + `Chdir`/`InitialCommit`/`At` options; package doc crediting #186
- [ ] runner dedup: replace `git()`/`gitIn()`/`captureGit()`/inline `run`/`runGit` with `testfix.Git`/`testfix.Capture` at every call site — top-level `cmd/sdlc/*_test.go` **and** `internal/gitx`, `internal/activetime`
- [ ] setup dedup: delegate `hermeticRepo`/`initFleetRepo`/`windowRepo`/`closeRepo`/`publishRepo` bodies + inline `close_test`×3/`migrate`×3 setups + `internal/gitx/window_test.go` + `internal/activetime` `gitInit` to `testfix.Repo(...)`; dated-commit / bare / remote steps stay local
- [ ] `go build ./... && go test ./cmd/sdlc/... ./pkg/vocab/` green (`./cmd/sdlc/...` covers the internal sub-packages) — behavior-identical, no test logic changed beyond the fixture call

## Verification

- `go build ./...` clean.
- `go test ./cmd/sdlc/... ./pkg/vocab/` green — the whole point; the fixtures back
  real tests, so green proves behavior-identical. Compare pass/package list vs the
  pre-refactor baseline (established green before starting).
- `rg` confirms no remaining private `init`+`config commit.gpgsign` setup copies
  or `git()`/`gitIn()`/`captureGit()`/`run` runner variants across `cmd/sdlc`
  (top-level AND `internal/gitx`, `internal/activetime`), **except** the
  documented locals: bare-origin init, remote-only `gitInit`, and the two
  dated-commit helpers.

## ARCH notes

- **ARCH-DRY** — the driver; one source for the fixture idiom.
- **ARCH-PURPOSE** — consolidate *all* the copies (both sweeps, top-level AND the
  two sub-package copies the first draft missed), not a convenient subset. The
  plan-quality gate caught exactly this under-scope; option (a) — include them —
  is the fix.
- **ARCH-PURE** — n/a in spirit (fixtures are IO scaffolding by nature); the win
  is single-sourcing the IO glue, not purifying it.

## Revisions

### 2026-07-19 — plan-quality gate (change-code) findings folded in

Reason: `sdlc change-code`'s plan-quality judge returned FAILURE on the first
draft. Deltas:
- **Finding 1 (blocking, ARCH-PURPOSE):** added the two sub-package copies
  (`internal/gitx/window_test.go`, `internal/activetime/parity_test.go`) to
  scope — the first draft's rg acceptance gate contradicted its top-level-only
  enumeration. Chose inclusion over exclusion. Noted the `-b main` determinism
  nuance and the dated-commit specializations that stay local. Estimate bumped
  1.8h → 2.07h (the judge sized this at +0.2–0.3h).
- **Finding 2 (minor):** documented the stale `writeProject` motivation as a
  verified red herring (already-DRY; markdown not git fixtures) and corrected the
  issue Spec.

### 2026-07-19 — extra copies found during implementation

The plan-quality gate surfaced the two `internal/*` copies; implementation
surfaced three MORE that both the first draft and the gate missed —
`activetime_test.go`'s `atGitRepo`, `collectdiff_test.go`, and `issue_test.go`'s
second (no-origin) site. All delegated (real count ~15, not ~12). Bare-origin
harnesses (`merge_e2e` `tempRepo`, `issue_test` AutoSyncsToMainCleanTree) +
remote-only `fetch` `gitInit` stayed local per the documented exception; the rg
gate confirms only those two bare-origin harnesses + the single `testfix.go`
source carry the `commit.gpgsign` setup line. See the issue `## Log` for the
full delegation inventory and verification.

### 2026-07-19 — close-review: one runner twin missed by the rg gate (FIX-THEN-SHIP)

The close boundary review (verdict FIX-THEN-SHIP, no Critical) found one
Important gap: `issuefiles_test.go`'s `runGitCommand` — a byte-identical twin of
`testfix.Git` whose *setup* already delegates via `hermeticRepo` (so it carries
no `commit.gpgsign` line) — fell outside BOTH verification filters (the
`gpgsign` grep AND the enumerated named runners) and stayed unconsolidated.
Folded in the delegation. Also (minor): delegated two inline add/commit loops in
`closereview_test.go`, restored `t.Helper()` in the delegating wrapper closures
(behavior-preservation — the inline originals had it), and tightened
`testfix.Repo`'s `At()` path to skip a discarded `t.TempDir()`. **Gate lesson:**
the acceptance grep must also match `exec.Command("git"` in `*_test.go` outside
the documented dated/bare/remote locals — keying only on `commit.gpgsign` + named
runners let a runner-only twin (setup-delegated, so no gpgsign line) slip
through.
