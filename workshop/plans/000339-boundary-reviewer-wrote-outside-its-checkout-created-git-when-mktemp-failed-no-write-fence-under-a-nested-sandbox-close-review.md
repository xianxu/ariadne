# Boundary Review — ariadne#339 (whole-issue close)

| field | value |
|-------|-------|
| issue | 339 — boundary reviewer wrote outside its checkout (created ~/.git) when mktemp failed: no write fence under a nested sandbox |
| repo | ariadne |
| issue file | workshop/issues/000339-boundary-reviewer-wrote-outside-its-checkout-created-git-when-mktemp-failed-no-write-fence-under-a-nested-sandbox.md |
| boundary | whole-issue close |
| milestone | — |
| window | e1a463ec6875969047968d05b01e05031f0776af..25731ea0ea232b7a289d4279c2552fd0a9a0ea47 |
| command | sdlc close --issue 339 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-12T14:18:33-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Status: the confinement is live in my own session (TMPDIR, shims, env file all set). Next I'll run the package tests and check a few edge cases in the shims.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The fix does what #339 needs and works. Every reviewer starts through `judge.Dispatch`: `milestoneclose.go:690`, `judge.go:180` and `planningreview.go:107,116`. Each one now gets a temp dir that is created and tested for writes before it runs. It also gets a git ceiling, `git` and `mktemp` shims on PATH, and an env file that puts the shims on PATH inside Claude Code's Bash. Afterward a check looks for changes outside and fails loudly if it finds any.

This review ran under the new confinement, so I could check it live:
- `TMPDIR`, `GIT_CEILING_DIRECTORIES` and `CLAUDE_ENV_FILE` were set as described.
- `command -v git` and `command -v mktemp` found the shims.
- `mktemp -d` landed in `.sdlc-review/tmp`.
- `git init` from `/` was refused.

`go test ./cmd/sdlc/internal/reviewbox/ ./cmd/sdlc/internal/judge/` passes. A post-check failure becomes `VerdictNotRun` in `milestoneclose.go:691-700`, so the gate halts.

One thing should be fixed before shipping. I confirmed two cheap ways around the git shim by running them, and the post-check does not catch either one. I deleted both test repos afterward.

**Strengths**
- One confinement point, `Dispatch` → `confine` (`dispatch.go:267-279`), covers all four reviewer types. The environment travels in the context, so the argv and the `Run` seam are unchanged (ARCH-DRY, ARCH-PURPOSE).
- The cause was measured, not guessed. macOS `mktemp` ignores TMPDIR when given no template. The finding that Claude Code's Bash rebuilds PATH led to the `CLAUDE_ENV_FILE` route (`confine.go:133-141`), and the plan's Revisions record both corrections honestly.
- The pure parts are tested directly with no IO: `Diff`, `GitShim`, `MktempShim`, `claudeEnvFile` and `overlayEnv`.
- `judge/confine_unix_test.go` is a real repro: a subprocess fake claude runs the incident script. Tests also cover the aborted-setup path and the clean path, including that the base is gone afterward.
- A failed retry re-runs `Prepare` (`dispatch.go:328-331`), so attempt 2 after a checkout reset is still confined.

**Critical:** none.

**Important**
1. **The git shim can be bypassed, and the post-check misses it** (ARCH-SECURE, ARCH-PURPOSE), `confine.go:262-293`.
   - **Relative paths:** after the subcommand, the shim checks only absolute arguments and skips relative ones (`*) continue`). From `$TMPDIR/a/b`, `git init ../../../../../../relesc` created `/Users/xianxu/Library/Caches/sdlc/relesc/.git`.
   - **Path options before the subcommand:** `--git-dir=/abs` and `--work-tree=...` before `init` are skipped. `git --git-dir=<review-checkouts>/escgd init` created a repository there.
   - **Why the check misses them:** both land in a sibling or cousin directory, not an ancestor or `$HOME`, so `Check` doesn't see them.
   - **Fix sketch:**
     - Resolve relative operands against `$dir`, i.e. check `inside "$(resolve "$dir/$a")"`.
     - Treat `--git-dir`/`--work-tree` values, both `=` and separate-argument forms, and `$GIT_DIR`/`$GIT_WORK_TREE` as paths that must be inside.
     - Add these cases to `TestGitShimRefusesRepositoriesOutside`.

**Minor**
- **Inherited env file can undo TMPDIR:** `claudeEnvFile` sources the inherited `CLAUDE_ENV_FILE` first and then re-exports only PATH. If that file exports `TMPDIR`, it silently replaces the confined TMPDIR in claude's Bash. Re-export TMPDIR/TMP/TEMP/GIT_CEILING_DIRECTORIES after sourcing (`confine.go:133-141`).
- **The check can silently do nothing:** `TakeSnapshot` ignores a `ReadDir(home)` error. An unreadable HOME makes the home check a no-op with no warning (`confine.go:213`).
- **False positives:** any new top-level `$HOME` entry made by another process during a long review fails the gate. The message tells the operator to rerun, but the cost is a whole review.
- **Direct git call outside the existing helper:** `Confine` runs `exec.Command("git", "rev-parse", …)` directly, without ctx and outside reviewbox's `git` helper. Not cancellable, and outside the seam (ARCH-MOCK).
- **Stale cause text:**
  - The plan's "## Cause (for the Log)" section still states the superseded CLAUDE_CODE_TMPDIR cause; only Revisions corrects it.
  - The `incidentScript` comment (`confine_unix_test.go:44`) repeats it.
- **Plan checkboxes:** the plan's task checkboxes are still `- [ ]`, while the issue's Plan is all `[x]`.
- **Long socket path:** the base inside the checkout lengthens `$TMPDIR`. This session's nested sandbox socket path was 105 bytes, over macOS's 104-byte `sun_path` limit. The log calls the EPERM a pre-existing problem, but a long TMPDIR makes it worse.

**Test coverage**
- The repro, abort, clean-path, retry-prepare and shim tests all exist and pass.
- The log mentions a mutation check (shim dropped from PATH → incident test goes red). That is plausible: without the shim, `git init` in the temp HOME succeeds and the `.git` assertion fires.
- Missing: shim cases for relative-path and `--git-dir` escapes, and an env-file test where the inherited file sets TMPDIR.

**Architecture**
| Marker | Result | Notes |
|---|---|---|
| ARCH-DRY | pass | Small note: `rev-parse` duplicates the reviewbox `git` helper's role. |
| ARCH-PURE | pass | |
| ARCH-PURPOSE | pass on consumers; flag | Every Dispatch caller is confined. The shim gap is the "refuse init/clone outside" class only partly covered (Important 1). |
| ARCH-MOCK | minor flag | Direct `exec.Command("git")` in `Confine`. |
| ARCH-CONSTRAINTS | pass | Setup latency is now visible: the cancellation test deadline went from 20 ms to 500 ms. Acceptable for a review-time path. |
| ARCH-SECURE | flag | Shim bypass; inherited env file is trusted to keep TMPDIR intact. |
| ARCH-ORDER | pass | Snapshot once, prepare on retry, check after the last attempt, close deferred. |
| ARCH-FUNERAL | pass | Base removed by defer. Inside the checkout it dies with the clone; the outside fallback is reaped by the OS temp-dir sweep. |

**Upcoming work:** the shim is a guard against accidents, not a fence. If an OS write fence becomes possible under nested sandboxes, it should replace the shim. Until then, the post-check is the real backstop; consider widening it beyond ancestors and `$HOME`, for example to the review checkout's parent directory.

**Plan revisions**
- Add a Revisions entry saying that "## Cause (for the Log)" is superseded, or rewrite that section.
- Add a Limits line saying the shim checks only cwd-relative and absolute operands, plus whatever the fix adds.

```findings
findings:
  - id: new
    severity: Important
    family: confinement-shim-path-coverage
    title: |
      git shim lets relative-path and --git-dir init create repositories outside, unseen by Check
    detail: |
      confine.go gitShimTemplate skips relative operands after init/clone and leading --git-dir/--work-tree values; verified live: `git init ../../../../../../relesc` from $TMPDIR/a/b created ~/Library/Caches/sdlc/relesc/.git and `git --git-dir=<review-checkouts>/escgd init` created a repo there, neither in an ancestor or $HOME so Check passes. Resolve relative args against $dir, check --git-dir/--work-tree (both forms) and GIT_DIR, add shim test cases.
  - id: new
    severity: Minor
    family: confinement-env-reassert
    title: |
      env.sh sources the inherited CLAUDE_ENV_FILE but re-exports only PATH, so an inherited TMPDIR wins
    detail: |
      claudeEnvFile (confine.go:133) should re-export TMPDIR/TMP/TEMP/GIT_CEILING_DIRECTORIES after sourcing the inherited file.
  - id: new
    severity: Minor
    family: silent-degraded-check
    title: |
      TakeSnapshot ignores ReadDir(home) errors, making the home check a silent no-op
  - id: new
    severity: Minor
    family: plan-stale-after-revision
    title: |
      Plan "Cause (for the Log)" section and incidentScript comment still state the superseded CLAUDE_CODE_TMPDIR cause; plan task boxes unchecked
  - id: new
    severity: Minor
    family: external-call-outside-seam
    title: |
      Confine runs exec.Command("git","rev-parse") directly, without ctx and outside reviewbox's git helper
```
