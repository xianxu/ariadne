# Boundary Review — ariadne#149 (whole-issue close)

| field | value |
|-------|-------|
| issue | 149 — sdlc command-tree tests should use an isolated repo lock, not the cwd lock |
| repo | ariadne |
| issue file | workshop/issues/000149-sdlc-command-tree-tests-should-use-an-isolated-repo-lock-not-the-cwd-lock.md |
| boundary | whole-issue close |
| milestone | — |
| window | b3d4412f1593b46d12317ac1c679ad69ff8824b7..HEAD |
| command | sdlc close --issue 149 |
| reviewer | claude |
| timestamp | 2026-07-05T22:12:55-07:00 |
| verdict | SHIP |

## Review

I have everything I need. My independent sweep confirms the plan's central claim, the guard fires end-to-end, and the real repo is untouched. Let me compile the review.

Summary of what I verified independently:
- `go build ./...` + `go vet ./cmd/sdlc/` clean; full `cmd/sdlc` suite green (18.9s), real repo HEAD/branch/porcelain unchanged before→after.
- New tests pass individually (`TestSnapshotDiff`, `TestGuardVerdict`, `TestRepoLock_IsolatedFromRealRepo`, `TestSetStatusAlias_BothPathsMutate`).
- **Anti-collusion proof**: added a throwaway test that pollutes the real repo → the `TestMain` guard printed the loud hermeticity message and FAILED an otherwise-passing run, proving the untested `os.Exit(exit)` wiring works. Cleaned up; repo left pristine.
- **ARCH-PURPOSE shadow-sweep**: enumerated every command-tree test (`buildRoot().Execute()` / `.Execute()`) and confirmed set-status was the sole un-isolated mutating offender (fetch chdir's into temp; process-manual/arch-principles/issue-validate are read-only; close uses unwrapped `NewCloseCmd()` + early refusal; repolock tests stub the acquire or chdir).

```verdict
verdict: SHIP
confidence: high
```

This boundary cleanly delivers both #149 (isolate the command-tree lock offender into a temp repo) and the folded #165 (a `TestMain` backstop that fails a passing run leaving durable real-repo damage). The architecture is textbook ARCH-PURE — a pure, unit-tested decision core (`repoSnapshot`/`snapshotDiff`/`guardVerdict`) behind a thin git-IO shell (`readSnapshot`/`realRepoRoot`/`TestMain`). I verified build/vet/full-suite green with the guard active and the real repo untouched, proved the guard actually fires end-to-end (the one piece the unit tests don't cover), and independently confirmed the "sole offender" claim via a full command-tree sweep. Nothing blocks SHIP; all findings are Minor/forward-looking.

**1. Strengths**
- **Clean pure/IO split (ARCH-PURE).** `snapshotDiff`/`guardVerdict` (`testmain_test.go:44,80`) are deterministic and unit-tested without IO; `TestMain`/`readSnapshot` are the injected thin shell. If a test needs mocks to run a "pure" entity it isn't pure — here the pure core genuinely runs on plain data. Textbook.
- **NET-state diffing done right.** `snapshotDiff` reports only *new* porcelain lines (set difference), so pre-existing session untracked notes don't false-trip — verified: the full suite ran green under the guard with the working tree in its normal state.
- **Honest, non-overclaiming comments.** `testmain_test.go:38-42` explicitly states the `lockFile` check catches a *leaked* lock, not #149's *transient* contention, and that the real #149 fix is offender isolation. The docs match what the code actually does.
- **Correct failure semantics.** `guardVerdict` preserves a failing run's own exit code rather than masking it, while still surfacing the mutation (`testmain_test.go:80`, proven by `TestGuardVerdict`).
- **Acceptance test targets the root cause.** `TestRepoLock_IsolatedFromRealRepo` (`hermeticrepo_test.go:40`) asserts the lock common-dir never resolves under the real checkout — the exact #149 property.
- Atlas updated accurately (`atlas/workflow/sdlc-binary.md`) — the new test-hermeticity paragraph faithfully describes both guards.

**2. Critical findings**
None.

**3. Important findings**
None.

**4. Minor findings**
- `testmain_test.go:110` (`readSnapshot`): the `.git/sdlc.lock` presence check `os.Stat(root/.git/sdlc.lock)` silently no-ops in a *linked worktree*, where `<toplevel>/.git` is a file and the real lock lives at the shared common-dir (as `atlas/workflow/sdlc-binary.md` itself documents). HEAD/branch/porcelain checks still work in a worktree; only the leaked-lock detection degrades. Since the primary #149 fix is offender isolation, this is a robustness note, not a hole. Consider resolving the lock via `git -C root rev-parse --git-common-dir` to match production.
- `hermeticrepo_test.go:56` (`TestRepoLock_IsolatedFromRealRepo`): asserts only *not-under-real-root*; it discards `hermeticRepo`'s returned `dir` and never positively asserts the lock is under the temp dir. The negative check is sufficient for #149, but `strings.HasPrefix(got, dir)` would be a marginally stronger pin.
- ARCH-DRY: the `git init -b main` + identity + gpgsign-off + chdir boilerplate now exists in ~11 test files (`closeRepo`, `collectdiff`, `fetch`'s `gitInit`, and the new `hermeticRepo`). Adding the minimal `hermeticRepo` was the right call for this issue, but it's now the canonical minimal isolator — a future cleanup could have the heavier helpers delegate their init+chdir to it rather than keep parallel copies.

**5. Test coverage notes**
- The pure decision (`snapshotDiff`/`guardVerdict`) is well covered, including the pre-existing-untracked-not-tripped, HEAD-move, branch-switch, and leaked-lock cases. The plan deliberately chose the pure `guardVerdict` test over a live env-gated variant to avoid real-repo risk — a defensible tradeoff. The consequence is that the `TestMain` → `os.Exit(exit)` glue itself has no automated regression guard; I verified it fires manually this session. If cheap, an env-gated subprocess test (spawn the compiled test binary in a temp git repo with a pollute flag, assert non-zero exit) would close that last gap and guard against a refactor silently disabling the backstop. Non-blocking.
- Offender-isolation completeness is currently pinned by the single-offender `TestSetStatusAlias` fix plus the always-on guard. That's adequate; the guard catches any *future* durable leak automatically.

**6. Architectural notes for upcoming work**
- ARCH-DRY: pass with the Minor note above — `hermeticRepo` should be treated as the one canonical temp-repo isolator going forward; don't add a 12th hand-rolled `git init` block.
- ARCH-PURE: pass — keep new guard surfaces (if `snapshotDiff` grows to cover config/refs, as the plan foresees) inside the pure core with the IO staying in `readSnapshot`.
- ARCH-PURPOSE: pass — both #149 and #165 Done-when items are met (lock isolation proven by the acceptance test; mutation backstop proven to fire). The shadow-sweep of command-tree consumers confirmed no deferred offender remains hand-waved as "follow-up."

**7. Plan revision recommendations**
None. The Core concepts table matches the code (`repoSnapshot`/`snapshotDiff`/`guardVerdict`/`TestMain` in `testmain_test.go`; the `hermeticRepo` isolator + the modified `TestSetStatusAlias`), and the plan's checkboxes are all genuinely delivered. The one refinement worth capturing lives in the issue Log, not the plan: the leaked-lock check's worktree blind spot (Minor above) — optional to note if a follow-up touches worktree lock semantics.

---

## Re-review — 2026-07-05T22:21:39-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 149 — sdlc command-tree tests should use an isolated repo lock, not the cwd lock |
| repo | ariadne |
| issue file | workshop/issues/000149-sdlc-command-tree-tests-should-use-an-isolated-repo-lock-not-the-cwd-lock.md |
| boundary | whole-issue close |
| milestone | — |
| window | b3d4412f1593b46d12317ac1c679ad69ff8824b7..HEAD |
| command | sdlc close --issue 149 |
| reviewer | claude |
| timestamp | 2026-07-05T22:21:39-07:00 |
| verdict | SHIP |

## Review

I have everything I need. My independent sweep confirms the diff, the guard is active and passing, the real repo is untouched, and the "sole offender" claim is empirically confirmed by the green run under the live guard. Compiling the review.

```verdict
verdict: SHIP
confidence: high
```

This boundary cleanly delivers both #149 (isolate the one command-tree lock offender into a temp repo) and the folded #165 (a package `TestMain` backstop that fails a passing run leaving durable real-repo damage). The architecture is textbook ARCH-PURE: a pure, unit-tested decision core (`repoSnapshot`/`snapshotDiff`/`guardVerdict`) behind a thin git-IO shell (`readSnapshot`/`realRepoRoot`/`TestMain`). I independently verified `go build ./...` + `go vet ./cmd/sdlc/` clean, the full `cmd/sdlc` suite green (18.996s, exit 0) **with the guard active**, and the real repo HEAD/branch/porcelain unchanged before→after — which is itself the strongest possible proof that offender-isolation is complete (any lingering real-repo mutation would have failed the run). I also re-ran the "sole offender" shadow-sweep from scratch and confirmed it. Nothing blocks SHIP; all findings are Minor/forward-looking.

**1. Strengths**
- **Clean pure/IO split (ARCH-PURE).** `snapshotDiff` (`testmain_test.go:41`) and `guardVerdict` (`testmain_test.go:72`) are deterministic, take plain data, and are unit-tested with zero IO; `readSnapshot`/`realRepoRoot`/`TestMain` are the injected thin git shell. A genuinely pure core, not a mock-driven one.
- **NET-state diffing done right.** `snapshotDiff` reports only *new* porcelain lines via set difference (`testmain_test.go:52-57`), so pre-existing session untracked notes don't false-trip — empirically confirmed (suite green, porcelain 0→0 under the guard).
- **Correct failure semantics.** `guardVerdict` preserves an already-failing run's own exit code rather than masking it, while still surfacing the mutation (`testmain_test.go:72-78`, pinned by `TestGuardVerdict`).
- **Acceptance test targets the root cause.** `TestRepoLock_IsolatedFromRealRepo` (`hermeticrepo_test.go:44`) asserts the lock common-dir never resolves under the real checkout — the exact #149 property, and it passed live.
- **Review loop actually closed.** The close review's worktree-lock Minor was *applied*, not just acknowledged: `readSnapshot` resolves the lock via `git rev-parse --git-common-dir` (`testmain_test.go:110-120`) rather than a hardcoded `<root>/.git`, so leaked-lock detection also works in a linked worktree. Commit "apply review Minor (worktree lock resolution)" is real.
- **Empirically-verified single-offender claim.** Independently swept every `buildRoot()`/`.Execute()` test: `TestSetStatusAlias` (now `hermeticRepo`-isolated) was the only un-isolated mutating verb; `fetch` chdir's, `close`/`issue-validate` use unwrapped direct constructors (`NewCloseCmd`/`newIssueValidateCmd`), the metadata/alias-shape tests never `Execute()`, and every `executeSDLCTestCommand` caller stubs the acquire (`repolock_test.go:324,385`).

**2. Critical findings**
None.

**3. Important findings**
None.

**4. Minor findings**
- `testmain_test.go` — the `TestMain → os.Exit(exit)` glue has no automated regression guard: the pure `guardVerdict` fire-path is unit-tested, but nothing exercises the end-to-end "passing run + real mutation → non-zero exit." I verified the clean path (exit 0, no false-fire) but can't trigger the fail path in a read-only review without polluting the tree. An env-gated subprocess test (spawn the compiled test binary in a temp git repo with a pollute flag, assert non-zero exit) would close it. Non-blocking.
- ARCH-DRY (`hermeticrepo_test.go:16`) — `hermeticRepo` is now the ~4th git-init+chdir shape alongside `gitInit` (fetch_test), `closeRepo` (closereview_test), and `git`/init blocks elsewhere. Adding the minimal isolator was the right call for this issue, but it's now the canonical minimal form; a future cleanup could have the heavier helpers delegate their init+chdir to it. Forward-looking.
- `testmain_test.go:35-40` — the guard is net-state, so a test that transiently corrupts `main` and reverts *within* `m.Run()` wouldn't be flagged. This is documented and intentional (durable damage is the threat that mattered in the #165 incident); noting only so the limitation is on record.

**5. Test coverage notes**
- The pure decision is well covered: identical-snapshot, pre-existing-untracked-not-tripped, new-change, HEAD-move, branch-switch, leaked-lock, and unresolved-repo-skip cases (`testmain_guard_test.go:15-52`), plus the three `guardVerdict` paths (`testmain_guard_test.go:56-72`). Assertions check real content (`d[0] != "new working-tree change: ?? f"`), not tautologies.
- Offender-isolation completeness is pinned by the always-on `TestMain` guard rather than a static enumeration — the right choice: it catches any *future* durable leak automatically, and it empirically passed this run.
- `abbrevSHA` (`state.go:371`) is safe on the short/empty SHAs the guard's tests feed it (`len > 8` guard).

**6. Architectural notes for upcoming work**
- ARCH-PURE — **pass.** If `snapshotDiff` later grows to cover config/refs (as the plan foresees), keep the decision in the pure core and the new git reads in `readSnapshot`.
- ARCH-DRY — **pass with the Minor above.** Treat `hermeticRepo` as the one canonical temp-repo isolator; don't add a 5th hand-rolled init+chdir block.
- ARCH-PURPOSE — **pass.** Shadow-sweep of command-tree consumers confirmed no deferred offender remains hand-waved as "follow-up"; both #149 and #165 Done-when items are met (lock isolation proven by the acceptance test + green run; mutation backstop's decision proven by unit tests). Production lock behavior is genuinely unchanged — `repolock.go` is not in this diff.

**7. Plan revision recommendations**
- Small table completeness (not a contradiction, so not blocking): the plan's "Pure entities" Core-concepts table lists `repoSnapshot` and `snapshotDiff` but not `guardVerdict`, which was extracted later (Task 4) and is a third pure entity that ships in `testmain_test.go`. Consider a one-line `## Revisions` entry adding `guardVerdict(before, after, code) (int, []string)` to the pure-entities table so the table fully reflects the delivered code. Every row the table *does* list matches the code at the stated path, so this is completeness, not drift.
