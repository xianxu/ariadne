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

---

## Re-review — 2026-10-12T14:31:19-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 339 — boundary reviewer wrote outside its checkout (created ~/.git) when mktemp failed: no write fence under a nested sandbox |
| repo | ariadne |
| issue file | workshop/issues/000339-boundary-reviewer-wrote-outside-its-checkout-created-git-when-mktemp-failed-no-write-fence-under-a-nested-sandbox.md |
| boundary | whole-issue close |
| milestone | — |
| window | e1a463ec6875969047968d05b01e05031f0776af..717a39a265dd89af4772469f88326bafea25b1ec |
| command | sdlc close --issue 339 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-12T14:31:19-07:00 |
| verdict | SHIP |

## Review

Tests pass; now checking that the round-1 shim tests go red when the fix is reverted, and probing the remaining git paths.

```verdict
verdict: SHIP
confidence: high
```

All five round-1 findings are fixed, and I checked each one against the code. The tests pass: `go test ./cmd/sdlc/internal/reviewbox/ ./cmd/sdlc/internal/judge/` is green. I reverted the new git-dir/work-tree/`GIT_DIR` check in the shim, and both `TestGitShimRefusesRepositoriesOutside` and `TestGitShimChecksGitDirEnv` failed. Then I restored the file and the tree is clean. So BR-1 has a regression test that catches a revert. I also ran the rendered shim against real git. `git init ../out/rel` and `GIT_DIR=../out/gd2 git init` are refused. Two siblings outside init/clone still get through: `git worktree add <outside>` and an alias that runs `init`. That is one Minor finding below. It doesn't block, because the post-run check backs up the shim for this incident's class (a new `.git` above the root or in `$HOME`).

1. **Strengths**
   - `confine.go` `resolve()` handles relative paths and `..` correctly. If a path climbs out of its deepest existing directory, it resolves to nothing, so the shim refuses it.
   - The env file now re-sets every variable from `Env()` after sourcing an inherited env file (`claudeEnvFile`). The test sets an inherited `TMPDIR` and checks that ours wins.
   - `TakeSnapshot` now returns an error when home is unknown or can't be listed, and both `Confine` and `Check` pass that error up. The new `TestTakeSnapshotFailsWhenHomeCannotBeListed` covers it.
   - All four reviewer callers get confinement through the one `judge.Dispatch` path (ARCH-DRY). The retry calls `Prepare()` again after the checkout is reset.
   - The plan's Revisions section records the corrected cause (macOS `mktemp` ignoring TMPDIR), the move to `CLAUDE_ENV_FILE` and the signature change. The atlas and `close.md` both describe the new behaviour.

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings**
   - The git shim only checks `init` and `clone`. I ran it live: `git worktree add $OUT/wt` and `git -c alias.mk=init mk $OUT/al` both created a repository or working tree outside the reviewer's directories. `Check` misses both unless the path is a new top-level entry in `$HOME` or an ancestor of the root or base. A reviewer adding a worktree at `/tmp/base` to compare against the base commit is a plausible, non-malicious case. This is the 2nd finding in this family (below).

5. **Test coverage notes**
   - The shim tests use a fake git that records its arguments, plus the real-subprocess incident test in `judge/confine_unix_test.go`. That is good coverage.
   - No test covers `worktree add` or alias expansion.
   - `TakeSnapshot` still treats an `Lstat` error on an ancestor's `.git` as "absent". That is harmless for readable ancestors, so I'm not raising it.

6. **Architectural notes**
   - **ARCH-DRY: pass.** `gitOutput` sits next to the existing `git()` helper.
   - **ARCH-PURE: pass.** `Diff`, `GitShim`, `MktempShim`, `claudeEnvFile` and `overlayEnv` are pure.
   - **ARCH-PURPOSE: flag (Minor).** The shim covers a list of commands rather than the class "git commands that create a repository or working tree".
   - **ARCH-MOCK: pass.** BR-5 moved the call to `gitOutput`. It still takes no `ctx`, which is acceptable for a single `rev-parse`.
   - **ARCH-CONSTRAINTS: pass.** The deadline test was widened, with the reason written down.
   - **ARCH-SECURE: pass.** Values written into the shims and env file are quoted with `shQuote`, and setup fails closed.
   - **ARCH-ORDER: pass.** Dispatch scopes the confinement's lifetime: it is created first, the retry calls `Prepare()` again, then `Check` runs and `Close` is deferred. A concurrent writer to `$HOME` produces a false positive, and the error message tells the operator to rerun.
   - **ARCH-FUNERAL: pass.** `Close` is deferred. A killed process leaves one small directory in the OS temp dir, which the OS cleans up.

7. **Plan revision recommendations**
   - Leave the plan's Core concepts and Task 3 text as they are. They still describe the old shim ("any absolute-path argument"), the old `pathPrefix` signature and the CLAUDE_CODE_TMPDIR stand-in, but the Revisions entries already record those changes, so this is not a new finding.
   - If the shim is widened to cover `worktree add` and alias expansion, add a Revisions entry for it.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Relative operands, --git-dir/--work-tree (both forms) and GIT_DIR/GIT_WORK_TREE are now checked; reverting the check turns TestGitShimRefusesRepositoriesOutside and TestGitShimChecksGitDirEnv red; a live run against real git refused ../out/rel and a relative GIT_DIR.
  - id: BR-2
    disposition: addressed
    note: |
      claudeEnvFile re-exports every Env() var after sourcing the inherited file; TestClaudeEnvFileSourcesAnInheritedOne sets an inherited TMPDIR and asserts ours wins.
  - id: BR-3
    disposition: addressed
    note: |
      TakeSnapshot returns an error for an unknown or unlistable home, Confine and Check propagate it, and TestTakeSnapshotFailsWhenHomeCannotBeListed covers it.
  - id: BR-4
    disposition: addressed
    note: |
      The plan Cause section is marked superseded and points to Revisions; the incidentScript comment in judge/confine_unix_test.go now states the mktemp/var-folders cause.
  - id: BR-5
    disposition: addressed
    note: |
      Confine now calls the reviewbox gitOutput helper next to git(); it takes no ctx, which is acceptable for one rev-parse at setup.
findings:
  - id: new
    severity: Minor
    family: confinement-shim-path-coverage
    title: |
      Git shim guards only init/clone; worktree add and alias-expanded init still create repos outside
    detail: |
      This is the 2nd finding in family confinement-shim-path-coverage. Verified live against real git: `git worktree add $OUT/wt` and `git -c alias.mk=init mk $OUT/al` both succeeded outside the root, and Check misses them unless they land as a new top-level entry in $HOME or an ancestor of the root or base. Rule: the shim guards the class "git commands that create a repository or working tree at a path" (init, clone, worktree add, and any alias that expands to one of them, by refusing -c alias.* and config-defined aliases of those), not a hand-kept list of two commands. Add test cases for each.
```
