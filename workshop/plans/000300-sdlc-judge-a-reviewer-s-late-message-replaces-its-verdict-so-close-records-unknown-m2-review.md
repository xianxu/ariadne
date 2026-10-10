# Boundary Review — ariadne#300 (milestone M2)

| field | value |
|-------|-------|
| issue | 300 — sdlc judge: a reviewer's late message replaces its verdict, so close records 'unknown' |
| repo | ariadne |
| issue file | workshop/issues/000300-sdlc-judge-a-reviewer-s-late-message-replaces-its-verdict-so-close-records-unknown.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 6eae019e1525207f085e65687e3f64ba4d4f09b8..00104a0157c37a206d86776219a2640ae40e54da |
| command | sdlc milestone-close --issue 300 --milestone M2 |
| reviewer | claude |
| timestamp | 2026-10-09T21:36:38-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Checking how the convergence line and ledger render read the `Blocked` stamp, then the timeout wiring.

Tests run next to confirm the new cases pass; then I'll write up.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

**Verdict: fix then ship.** Both M2 changes work. D6 now records a boundary round that produced no findings block as `BLOCKED` instead of "passed". D7 sizes the boundary review's timeout by the window's added lines, and `WF_REVIEW_TIMEOUT` still overrides it. I read the full diff and ran `TestReviewTimeout`, `TestBoundaryRoundWithoutFindingsIsNotPassed` and `TestCloseSizesTheReviewTimeout`; all pass.

One problem should be fixed before the boundary. Plan Task 5 is ticked as "doesn't count as convergence", but only the ledger record and `PassesUnchanged` read the new stamp. Other places that read the round's outcome still treat it as a pass:
- The convergence line reports "Converging."
- The gate's terminal line is the success line (`cok`).
- A `--force` rationale for that round is not recorded.

### Strengths
- **The stamp doesn't change the gate decision.** `gatepersist.go:54` puts the protocol-error case on the ledger record only (`d.Block || g.Blocked`). The M2 build revision explains why: changing the decision would have blocked a SHIP close whose dispositions merely failed validation, with an empty "open findings" list. That reasoning holds.
- **`judge.ReviewTimeout` is pure and table-tested** (`dispatch.go:352`), including the boundary values 500/501 and the 2-hour cap.
- **`TestCloseSizesTheReviewTimeout` checks the real deadline.** It reads the deadline from the reviewer's actual `ctx` through the `judge.Run` seam, end to end through `close`, so it fails if the sizing is removed. It also covers the override case.
- **A window that can't be measured falls back visibly.** `boundaryReviewTimeout` warns and returns 0 (the default) instead of failing the close.
- Help text (root, close, milestone-close, change-code) and the atlas were updated in the same range.

### Critical
None.

### Important
1. **Other readers of the round outcome still treat a findings-less round as a pass** (`boundaryledger.go:193`, `gatepersist.go:54-67`, `gatestate/family.go:154`).
   - `ConvergenceLine` builds its verdict only from new findings and repeated families. A findings-less round has 0 of each, so it prints "… 0 new findings … Converging." That is the opposite of Task 5's ticked claim, and of the warning printed just above it ("the gate cannot converge on it").
   - `stampAndPersist` still chooses between the warning and success terminal lines by `d.Block` alone, so the terminal says ok while the ledger says BLOCKED.
   - `forcedRationale(forced, d.Block)` skips the `--force` waiver on a round that is now recorded as blocked.
   - **Rule (ARCH-PURPOSE, the same class swept once):** a round's outcome is decided in one place, and every reader of it uses that one value.
   - **Fix sketch:** compute `blocked := d.Block || g.Blocked` once in `stampAndPersist` and use it for the stamp, `Forced` and the warning/ok choice. Pass the protocol-error state into `ConvergenceLine`, for example "No findings block — not converging." Then extend the D6 test to assert on the convergence line in stderr.

### Minor
1. **The plan names a test that doesn't exist.** Task 5 names `TestProtocolErrorRoundBlocks`; the shipped test is `TestBoundaryRoundWithoutFindingsIsNotPassed`. **This is the 3rd finding in family `plan-claims-match-tests`.** The rule that covers all of them: when a Plan row is ticked, every test it names must exist under that name and assert every clause of the row. The fix is to check that at tick time, by grepping the named test and confirming each clause, rather than editing this one row.
2. **The precedence and 30-minute default are restated (ARCH-DRY).** `reviewLimitLabel` (`milestoneclose.go:802`) repeats `dispatchTimeout`'s precedence and its 30-minute default, so that default now lives in three places. Export one `judge.EffectiveTimeout(env, sized)` and label from that.
3. **The limit label trusts an invalid override.** An invalid `WF_REVIEW_TIMEOUT` is printed as the "limit" before dispatch rejects it. It's harmless, but the label isn't a validated value.

### Test coverage notes
- D6 is pinned on the rendered ledger and the stored record, but not on the convergence line or the `--force` path.
- D7 is pinned both as a pure table and end to end.
- One deadline spanning the retry is covered from M1 (`TestDispatchRetryDeadlineKeepsTheFirstAttempt`).

### Architecture
- **ARCH-DRY:** flagged (Minor 2).
- **ARCH-PURE:** pass. `ReviewTimeout` is pure; the IO stays in `boundaryReviewTimeout`.
- **ARCH-PURPOSE:** flagged (Important 1, the convergence and `--force` readers).
- **ARCH-MOCK:** pass. `judge.Run` and `git` run through the existing seams in tests.
- **ARCH-CONSTRAINTS:** pass. The timeout is bounded at 2 hours and visible in the dispatch line, and the deadline covers both attempts.
- **ARCH-SECURE:** pass, with Minor 3 noted.
- **ARCH-ORDER:** pass. This adds no new carried state; it is a stamp computed once per round.
- **ARCH-FUNERAL:** pass. Nothing new is durable; the ledger rows are unchanged in size apart from a boolean value.

### Architectural notes for upcoming work
- `stampAndPersist` now has two sources for "blocked". Folding them into one `effectiveBlocked` value keeps the plan gate and the boundary gate from drifting apart.

### Plan revision recommendations
- **Task 5:** rename the test reference to `TestBoundaryRoundWithoutFindingsIsNotPassed`. Either deliver "not convergence" (Important 1) or record a revision that drops that clause.

```findings
findings:
  - id: new
    severity: Important
    family: one-round-outcome-all-readers
    title: |
      A findings-less boundary round is BLOCKED in the ledger but still reported as Converging, ok, and unforced
    detail: |
      gatePersist.Blocked only feeds the stored Blocked flag. ConvergenceLine (gatestate/family.go:154) prints "Converging." for 0 new / 0 repeats, stampAndPersist picks cok vs cwarn by d.Block alone, and forcedRationale(forced, d.Block) drops a --force waiver on a round now recorded blocked. Plan Task 5 is ticked as "doesn't count as convergence". Rule: one effective outcome per round, consumed by every reader. Compute blocked := d.Block || g.Blocked once in stampAndPersist for Blocked, Forced and the warn/ok line, and have ConvergenceLine report a protocol-error round as not converging. Extend the D6 test to assert the stderr convergence line.
  - id: new
    severity: Minor
    family: plan-claims-match-tests
    title: |
      Plan Task 5 names TestProtocolErrorRoundBlocks; the shipped test is TestBoundaryRoundWithoutFindingsIsNotPassed
    detail: |
      3rd finding in this family. Rule: ticking a Plan row requires that every named test exists under that name and asserts every clause of the row; check it when the row is ticked rather than fixing rows one at a time.
  - id: new
    severity: Minor
    family: single-source-timeout
    title: |
      reviewLimitLabel restates dispatchTimeout's precedence and the 30m default (ARCH-DRY)
    detail: |
      milestoneclose.go reviewLimitLabel repeats the env-over-sized precedence and the 30m default, which now lives in three places. Export judge.EffectiveTimeout(env, sized) and label from its result, which also stops an invalid override from being printed as the limit.
```
