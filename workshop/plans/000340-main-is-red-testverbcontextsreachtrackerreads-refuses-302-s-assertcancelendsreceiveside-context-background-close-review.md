# Boundary Review — ariadne#340 (whole-issue close)

| field | value |
|-------|-------|
| issue | 340 — main is red: TestVerbContextsReachTrackerReads refuses #302's AssertCancelEndsReceiveSide (context.Background) |
| repo | ariadne |
| issue file | workshop/issues/000340-main-is-red-testverbcontextsreachtrackerreads-refuses-302-s-assertcancelendsreceiveside-context-background.md |
| boundary | whole-issue close |
| milestone | — |
| window | 9aa008096b173acd2bdbdd61dfd1dab7fcb23312..ee93f7511b6ad73d4523da37369c47624881b9a4 |
| command | sdlc close --issue 340 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-12T16:41:06-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

This is a one-line fix and it does what the issue's Spec asks. `AssertCancelEndsReceiveSide` now builds its cancel context from `t.Context()` instead of `context.Background()`, and the context guard itself is unchanged. I checked it in the pinned checkout. `TestVerbContextsReachTrackerReads` passes at head. With base's `blockingorigin.go` put back, the same test fails with `internal/testfix/blockingorigin.go:59:36: testfix.AssertCancelEndsReceiveSide uses context.Background; thread the caller's context instead`. I then restored the file and the tree is clean. Both #302 pins pass at head: `TestTrackerGitCancelEndsTheReceiveSide` in cmd/sdlc and `TestGitCancelEndsTheReceiveSide` in gitx. I did not re-run the check where the group kill is reverted; that evidence comes from the Log only. That check is about #302's own fix, which this diff does not touch. Nothing blocks SHIP.

1. **Strengths**
   - `cmd/sdlc/internal/testfix/blockingorigin.go:59`: the fix brings the helper into line with the guard instead of weakening the guard, as the Spec asked. `t.Context()` is cancelled when the test ends, so the test-scoped lifetime is kept, and the existing `defer cancel()` still sets the cancel point the assertion relies on.
   - The sweep claim holds. `grep` finds no `context.Background` or `context.TODO` in the non-test files of `internal/testfix`. Only two callers exist (`internal/gitx/cancel_test.go:14` and `trackerenv_cancel_test.go:15`), and both are covered.
   - The Log gives a precise cause: 47a4e463 moved the assertion from `_test.go` files into a non-test file that the guard scans, and the re-close relied on "prior evidence stands".
   - `t.Context()` needs Go 1.24; `go.mod` declares `go 1.26.3`.

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings**
   - The inline comment at `blockingorigin.go:59` is long for an end-of-line comment. It's fine as is.

5. **Test coverage notes**
   - The existing guard test is the regression test, and I confirmed it goes red without the fix.
   - Both #302 pins pass at head.

6. **Architectural notes**
   - **ARCH-DRY: pass.** It is a single shared helper with no duplication.
   - **ARCH-PURE: pass.** The change only affects test infrastructure.
   - **ARCH-PURPOSE: pass.** The class is "non-test helpers in packages the guard scans that create a root context". Within the Spec's scope, `internal/testfix`, nothing else remains.
   - **ARCH-MOCK: pass.** No new external calls.
   - **ARCH-CONSTRAINTS: pass.** No runtime change.
   - **ARCH-SECURE: pass.** No untrusted input or secrets are involved.
   - **ARCH-ORDER: pass.** The context still ends at the earlier of explicit cancel or test end. The goroutine's lifetime is still bounded by `done` and the test.
   - **ARCH-FUNERAL: pass.** The change creates nothing durable.
   - Process lesson from the Log: a re-close should re-run `make test` whenever code has moved since the last evidence. This belongs to the workflow, not to this diff.

7. **Plan revision recommendations:** none. The plan matches the code.

```findings
findings:
  - id: new
    severity: Minor
    family: comment-density
    title: |
      Long trailing comment on blockingorigin.go:59; it could be a short line above the statement
    detail: |
      Style only; the content is accurate.
```
