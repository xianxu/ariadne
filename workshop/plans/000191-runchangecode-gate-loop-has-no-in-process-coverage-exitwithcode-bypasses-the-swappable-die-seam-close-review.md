# Boundary Review — ariadne#191 (whole-issue close)

| field | value |
|-------|-------|
| issue | 191 — runChangeCode gate loop has no in-process coverage — exitWithCode bypasses the swappable die seam |
| repo | ariadne |
| issue file | workshop/issues/000191-runchangecode-gate-loop-has-no-in-process-coverage-exitwithcode-bypasses-the-swappable-die-seam.md |
| boundary | whole-issue close |
| milestone | — |
| window | ab362ac4c35f137125f25e36a45f7b994d7493ad..e2ffaca01aba4238a63fc4b8476663263521a812 |
| command | sdlc close --issue 191 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-12T07:44:55-07:00 |
| verdict | SHIP |

## Review

Targeted tests pass; now running mutation checks and verifying the "one process-exit seam" claim.

```verdict
verdict: SHIP
confidence: high
```

This change does what the issue asked. `exitWithCode` is now a package var in the same style as `die` (`cmd/sdlc/term.go:60`), and `die` still exits through it. `changecode.go` is not touched in this window, so no gate's behavior moved. The two branches that had no tests are now driven in-process through `runChangeCode` itself. I checked that both tests catch real regressions by changing the code in the disposable checkout. Adding a `break` after the bypass line makes `TestChangeCodeGateLoopForcedFailureContinues` fail. Removing the unforced `exitWithCode(1)` makes both tests fail. I restored the checkout afterwards and confirmed it was clean. The targeted suite passes on head: the gate-loop tests, `TestForceAckMatchesGateCatalog`, `TestGateOrder*`, `TestActiveChangeCodeGates`, the wiring guards and the die tests. Nothing blocks SHIP. One Minor finding: a comment overstates the "one process-exit seam" claim.

**1. Strengths**
- **One seam, as specified.** `exitWithCode` is a var like `die`, and `die` still goes through it (`term.go:52`). The other existing callers (`close.go`, `push.go:108`, `issuelintids.go`, and the ask-exit at `changecode.go:959`) also become swappable, and none of them changed.
- **`expectExit` mirrors `expectDie` closely** (`die_test.go:60-76`). It uses a typed payload, restores the previous var in the defer, and re-panics on any panic that isn't its own, so genuine crashes are not hidden.
- **The tests go in through the real entry point.** They call `runChangeCode`, not a hand-built gate list. "Reached the next gate" can be observed because the `estimate` and `estimate-recon` gates are adjacent and both fail on #9. The unforced test also checks that HEAD and the porcelain status are unchanged.
- **Stale references were swept.** Every "not in-process drivable (#191)" claim was updated, across the atlas, `commitpathspec_guard_test.go`, `changecode_flow_test.go`, `observe_test.go` and `planreplay_test.go`.

**2. Critical findings:** none.

**3. Important findings:** none.

**4. Minor findings**
- **The "one process-exit seam" comment is overstated** (`term.go:55-59`). `judge.go:200` still calls `os.Exit(1)` directly when a judge fails. That skips the die cleanups and cannot be swapped by `expectExit`.
  - Other direct `os.Exit` calls are legitimate top-level exits: `main.go:70`, `activetime.go:245` (cobra wrapper), and `repolock.go:242/244` (signal handlers).
  - No new exit path was added, so the Done-when clause is met. The fix is either to scope the comment, or to route `judge.go:200` through `exitWithCode` in a follow-up.
- **ARCH-DRY (acceptable):** `expectExit` and `expectDie` have the same swap-and-recover shape. They swap different vars, so keeping two small helpers is reasonable.

**5. Test coverage notes**
- Both Done-when states are covered: unforced (exit 1, next gate never runs, checkout untouched) and forced (bypass recorded, next gate runs, the run gets past the loop to "Would create branch").
- The order check in the forced test relies on the earlier `Contains` loop to catch a missing string, because `Index` returns -1 for an absent string. That works.
- Package tests run serially (no `t.Parallel`), so swapping a global var is safe.

**6. Architectural notes**
- **ARCH-DRY:** pass.
- **ARCH-PURE:** pass. The IO exit is now an injected seam.
- **ARCH-PURPOSE:** pass. The whole purpose is delivered, nothing was deferred. The only sibling of the same class still in the tree is `judge.go:200`, listed above.

**7. Plan revision recommendations:** none. The plan matches the code.

```findings
findings:
  - id: new
    severity: Minor
    family: doc-claim-unverified-against-code
    title: |
      term.go comment says exitWithCode "stays the one process-exit seam", but judge.go:200 still calls os.Exit(1) directly
    detail: |
      judge.go:200 exits directly on a judge failure, so the die cleanups are skipped and expectExit cannot swap it. Every other direct os.Exit is a top-level or signal exit: main.go:70, activetime.go:245, repolock.go:242/244. Either scope the comment or route judge.go:200 through exitWithCode.
```
