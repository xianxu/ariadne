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

---

## Re-review — 2026-10-09T21:46:02-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 300 — sdlc judge: a reviewer's late message replaces its verdict, so close records 'unknown' |
| repo | ariadne |
| issue file | workshop/issues/000300-sdlc-judge-a-reviewer-s-late-message-replaces-its-verdict-so-close-records-unknown.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 6eae019e1525207f085e65687e3f64ba4d4f09b8..ecee5aabaf6142c76b3c3faf029769d170a47848 |
| command | sdlc milestone-close --issue 300 --milestone M2 |
| reviewer | claude |
| timestamp | 2026-10-09T21:46:02-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Reviewing the M2 diff for #300. Next I'm checking the plan rows and the ledger edge cases.

All the test names the plan cites exist. Next I'm running the targeted tests and a mutation check on the BR-9 fix.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

Round 1 asked for three fixes, and all three landed with real evidence. BR-9 now computes the round's outcome once in `stampAndPersist`, and that one value sets the stored flag, the waiver stamp and the reported line. `ConvergenceLine` no longer says "Converging." for a findings-less round. I reverted each half of that fix in a scratch worktree, and `TestBoundaryRoundWithoutFindingsIsNotPassed` failed both times. BR-10: all 21 tests named in Plan Tasks 1–6 exist under those names, and the close steps moved to their own unticked row. BR-11: `reviewLimitLabel` now reads `judge.EffectiveTimeout`, and an invalid override shows as invalid. Nothing blocks the boundary, but two cheap items should be fixed first. `README.md:65` still says the timeout defaults to 30m. And `ConvergenceLine` decides "no readable findings" with its own test, separate from the one that sets the stored flag, so the readers can still disagree in one edge case.

1. **Strengths**
   - `cmd/sdlc/gatepersist.go:55` — `blocked := d.Block || g.Blocked` is the single statement of the round's outcome. The decision returned to the caller is deliberately unchanged, and the Revisions entry explains why: forcing it would block a SHIP close whose dispositions only failed validation.
   - `cmd/sdlc/closestream_test.go:73` `TestCloseSizesTheReviewTimeout` checks the reviewer's real context deadline through the `judge.Run` seam, not a mocked value. It covers both the sized limit and the override (ARCH-MOCK pass).
   - `judge.ReviewTimeout` is a pure function with a table test over the boundary values: 500, 501, 1500 and the cap (ARCH-PURE pass).
   - The help text, the atlas and `lessons.md` were updated in the same commit, and the two new lessons state rules rather than single cases.

2. **Critical:** none.

3. **Important**
   - `README.md:65` still says "`WF_REVIEW_TIMEOUT` defaults to `30m`". Boundary reviews now scale with the window (`judge.ReviewTimeout`), and every other description (`root.md`, `close.md`, `milestone-close.md`, `change-code.md`, the atlas) was updated. This is the 3rd finding in family `single-source-timeout`. Earlier rounds fixed instances. The rule: when the timeout policy changes, run `grep -rn WF_REVIEW_TIMEOUT` across the repo and update every hit in the same commit.

4. **Minor**
   - **Two tests for one outcome (3rd finding in family `one-round-outcome-all-readers`).**
     - The boundary sets `Blocked` from `review.Round == nil`.
     - `ConvergenceLine` (`gatestate/family.go:158`) uses a different test: `ProtocolError != "" && no New && no Dispositions`.
     - They disagree in one case: a findings block whose only content is an invalid disposition. `ApplyChecked` drops it and sets `ProtocolError`. The convergence line then says "Not converging: the review produced no readable findings (…dropped 1 invalid disposition…)", while the gate prints `[ok]`.
     - The rule: make it one predicate, for example a `gatestate.Round` method or a stored flag, that both the stamp and `ConvergenceLine` read.
     - Separately, the waiver-stamp reader, `forcedRationale(forced, blocked)`, has no test. If it were reverted to `d.Block`, the D6 test would still pass, because it passes no forced rationale.
   - `stampAndPersist` is shared by both gates, but its warning hard-codes "this round's review produced no findings block". A `BlockedReason` field next to `Blocked` would keep the gate-specific wording out of the shared tail.
   - `dispatch.go` writes the 30m default and the 2h ceiling as literals twice each (`ReviewTimeout` and `reviewTimeout`). Named constants would fix it. This is in one file and harmless for now.

5. **Test coverage**
   - Mutation checks (reverting the `ConvergenceLine` branch, and reverting `blocked` to `d.Block`) both turn the D6 test red, so the fix is pinned.
   - `ConvergenceLine`'s new branch has no unit test in `family_test.go`; only the end-to-end test covers it.
   - Targeted tests pass: the `cmd/sdlc` subset, and the `judge` and `gatestate` packages.

6. **Architecture**
   - ARCH-DRY: flag (Minor), the duplicated literals above. The precedence itself is now defined once (`EffectiveTimeout`).
   - ARCH-PURE: pass.
   - ARCH-PURPOSE: pass. D6 and D7 are delivered.
   - ARCH-MOCK: pass. The test goes through the `judge.Run` seam.
   - ARCH-CONSTRAINTS: pass. The timeout scales with the window and stays capped at 2h; measuring the window costs one extra git call.
   - ARCH-SECURE: pass. The environment variable is validated, and an invalid value is labeled as such rather than displayed as a limit.
   - ARCH-ORDER: flag (Minor), the two tests for one outcome above.
   - ARCH-FUNERAL: pass. Nothing new is durable.

7. **Plan revisions:** none needed. The plan matches the code.

```findings
dispose:
  - id: BR-9
    disposition: addressed
    note: |
      stampAndPersist computes blocked once for the flag, the Forced stamp and the warn/ok line, and ConvergenceLine reports not converging. Mutation-checked: reverting either half fails TestBoundaryRoundWithoutFindingsIsNotPassed.
  - id: BR-10
    disposition: addressed
    note: |
      All 21 tests named in the Plan Tasks 1-6 rows exist under those names; the close steps sit in their own unticked row.
  - id: BR-11
    disposition: addressed
    note: |
      reviewLimitLabel reads judge.EffectiveTimeout; an invalid override is labelled invalid rather than printed as the limit.
findings:
  - id: new
    severity: Important
    family: single-source-timeout
    title: |
      README.md:65 still says WF_REVIEW_TIMEOUT defaults to 30m; boundary reviews now scale with the window
    detail: |
      3rd finding in family single-source-timeout. Rule: a timeout-policy change sweeps every hit of grep -rn WF_REVIEW_TIMEOUT in the same commit; helptext and atlas were swept, README was missed.
  - id: new
    severity: Minor
    family: one-round-outcome-all-readers
    title: |
      ConvergenceLine decides "no readable findings" with its own predicate, separate from the Blocked stamp's review.Round == nil
    detail: |
      3rd finding in family. A findings block holding only an invalid disposition gets ProtocolError from ApplyChecked, so ConvergenceLine says "Not converging: no readable findings" while the gate prints [ok]. Rule: one predicate (a gatestate Round method or a stored flag) read by the stamp and by ConvergenceLine. The Forced-stamp reader also has no test: reverting it to d.Block keeps the D6 test green.
  - id: new
    severity: Minor
    family: channel-vs-prose-separation
    title: |
      The shared stampAndPersist hard-codes the boundary-specific "produced no findings block" wording for gatePersist.Blocked
    detail: |
      Carry a BlockedReason alongside Blocked so the shared tail holds no gate-specific prose.
```
