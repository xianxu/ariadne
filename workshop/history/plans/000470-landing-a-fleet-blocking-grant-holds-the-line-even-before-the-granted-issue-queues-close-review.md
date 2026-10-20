# Boundary Review — ariadne#470 (whole-issue close)

| field | value |
|-------|-------|
| issue | 470 — landing: a fleet-blocking grant holds the line even before the granted issue queues |
| repo | ariadne |
| issue file | workshop/issues/000470-landing-a-fleet-blocking-grant-holds-the-line-even-before-the-granted-issue-queues.md |
| boundary | whole-issue close |
| milestone | — |
| window | 259c9b93331fcb4d8bdea1716eadcc109f106e92..ea1e7197580fcd4c9dea6df83f0550beeef0b57a |
| command | sdlc close --issue 470 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-19T20:04:05-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: medium
```

The implementation fulfills the revised grant-hold policy. No Critical or Important defects found. Directly affected tests pass, including race detection for `landingline`. Broader validation remains incomplete because the command-suite run reached its default 10-minute timeout. One minor acceptance-text inconsistency remains.

Done-when: current — covers the added hold, FIFO, recovery, reporting, and bounded-expiry behavior; its “tests unchanged” wording is inaccurate.

```findings
findings:
  - id: new
    severity: Minor
    kind: defect
    family: acceptance-text-matches-delivery
    title: |
      Done-when still promises unchanged grant-ordering tests
    detail: |
      workshop/issues/000470-landing-a-fleet-blocking-grant-holds-the-line-even-before-the-granted-issue-queues.md:43 says existing grant-ordering tests remain unchanged. Both instances in this window intentionally change: order_test.go:40 reverses the absent-grant expectation, and line_test.go:109 clears the completed grant before normal admission. Record a Revisions entry replacing the literal unchanged-test promise with preservation of ordering and non-preemption guarantees under the revised hold policy.
```

1. **Strengths**

   - `GrantHold` centralizes admission and reporting policy in the pure core (`order.go:96`).
   - The hold runs after acquiring the free head lock, preserving running landings (`line.go:254`).
   - Deterministic fixtures cover absent issues, arrival, exact expiry, revocation, cancellation, multiple grants, and hand-off gaps.
   - README, help, and atlas document the behavior and recovery command.

2. **Critical findings**

   None.

3. **Important findings**

   None.

4. **Minor findings**

   The Done-when wording described above.

5. **Test coverage notes**

   Passed:

   - `go test ./cmd/sdlc/internal/landingline`
   - `go test -race ./cmd/sdlc/internal/landingline`
   - Focused command tests covering grant-hold status, serial/train waiter output, existing train-prefix behavior, and granted train departure.
   - Pinned-range `git diff --check`.

   The broader `Test(Landing|Line|Train|EnterLandingLine)` run timed out after 600 seconds while executing `TestLandingTicksPeerProject`. This is incomplete validation, not evidence of a grant-hold defect. I did not independently reproduce the claimed full-suite result.

6. **Architectural notes**

   - **ARCH-DRY: pass.** Admission and reporting reuse one hold rule.
   - **ARCH-PURE: pass.** Grant decisions remain deterministic; filesystem and locking stay in the existing IO shell.
   - **ARCH-PURPOSE: pass.** Both initial absence and between-issue hand-off retain the hold; granted tickets preserve FIFO.
   - **ARCH-EVOLVE: pass.** No persisted-format change; existing log sweeps retain raw records, including unknown nested members.

7. **Plan revision recommendations**

   Add a `## Revisions` entry clarifying that existing ordering guarantees remain covered while test expectations intentionally change to reflect the new hold policy.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"94b96ee8-2870-4045-a2dd-1959433e500a","repository":"github.com/xianxu/ariadne","issue_id":"000470","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/ee510c4184d641cf","reviewed_head":"ea1e7197580fcd4c9dea6df83f0550beeef0b57a","slot":"ariadne:1","requested_at":"2026-10-20T02:52:56.591363Z"},"phase":"finished","pid":17160,"started_at":"2026-10-20T02:52:56.677342Z","finished_at":"2026-10-20T03:04:04.628176Z","session_id":"01a15739-eecd-7833-94fb-9f6a71822bb1","outcome":"returned"}] -->
