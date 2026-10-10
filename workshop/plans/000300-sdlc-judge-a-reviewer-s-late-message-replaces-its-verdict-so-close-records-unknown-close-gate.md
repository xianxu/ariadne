---
gate: boundary-review
issue: 300
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-09T20:50:49-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Important
          title: APIFailure treats any is_error result as "review did not run" even when the run already carries a verdict
          detail: apifailure.go:25 checks is_error before HasVerdict, so a reviewer that wrote its verdict and then hit error_during_execution/max_turns/an overload loses it, and the error misdirects to a sandbox fix. Check HasVerdict first, or narrow the message when no signature matches; add a fixture test.
          family: late-signal-erases-verdict
          round: 1
        - id: BR-2
          severity: Important
          title: ReadStream runs on codex/gemini text output, so a quoted stream-event JSON line can trip ErrAPIUnreachable
          detail: dispatch.go:287 stream-parses every agent's stdout; a standalone {"type":"result","is_error":true} line in text-mode prose sets Stream=true, is dropped from Text(), and fails the dispatch. Gate stream parsing on AgentClaude and add a test.
          family: channel-vs-prose-separation
          round: 1
        - id: BR-3
          severity: Minor
          title: The "allow api.anthropic.com" remedy is hard-coded for every agent
          family: agent-specific-remedy
          round: 1
        - id: BR-4
          severity: Minor
          title: A retry that hits the deadline drops the first attempt's text from the error/sidecar
          family: fail-safe-keeps-evidence
          round: 1
        - id: BR-5
          severity: Minor
          title: Task 2 claims close-level end-to-end and sidecar tests; the tests stop at Dispatch
          family: plan-claims-match-tests
          round: 1
        - id: BR-6
          severity: Minor
          title: A stderr API signature on a clean-exit, verdict-less run suppresses the intended retry
          family: late-signal-erases-verdict
          round: 1
      boundary: M1
      recipe: milestone-review
      blocked: true
    - "n": 2
      timestamp: "2026-10-09T21:06:47-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: RunFailure checks HasVerdict first; verdict_then_error.jsonl and maxTurns cases in TestRunFailure; reverting the guard turns the test red.
          round: 2
        - id: BR-2
          disposition: addressed
          note: classifyRunResult reads only claude stdout as a stream; TestDispatchParsesOnlyClaudeAsAStream goes red when the gate is removed.
          round: 2
        - id: BR-3
          disposition: addressed
          note: apiHosts maps each agent to its host; TestRunFailure asserts the codex and gemini remedies.
          round: 2
        - id: BR-4
          disposition: addressed
          note: The retry's DeadlineExceeded branch returns attempt 1 labelled; TestDispatchRetryDeadlineKeepsTheFirstAttempt pins it.
          round: 2
        - id: BR-5
          disposition: addressed
          note: TestCloseRecordsTheVerdictBeforeAPostscript and TestCloseWithoutAVerdictKeepsBothRuns exist in cmd/sdlc/closestream_test.go and pass.
          round: 2
        - id: BR-6
          disposition: addressed
          note: The stderr case requires nonZeroExit; dropping that condition turns TestRunFailure red.
          round: 2
      findings:
        - id: BR-7
          severity: Important
          title: Atlas, contract doc and plan Core concepts/D5 still describe the code before round 1
          detail: 'atlas sdlc-binary.md:892 and construct/judge-output-contract.md say any is_error result is ErrAPIUnreachable with the sandbox fix. The plan table lists APIFailure(...)(cause, ok) with HasVerdict in classify.go, and D5 hard-codes api.anthropic.com. The code now keeps a verdict before an error, has ErrReviewDidNotRun, reads only claude as a stream, and names a host per agent. 2nd in family. Rule: a commit that changes a contract also updates every artifact describing it (plan, atlas, contract doc) in the same commit.'
          family: plan-claims-match-tests
          round: 2
        - id: BR-8
          severity: Minor
          title: A retry that fails with an error (not a timeout) drops attempt 1's text
          detail: 'Dispatch returns (second, err) for ErrAPIUnreachable, ErrReviewDidNotRun, an interrupt or a launch failure on attempt 2, and milestoneclose.go:641 discards the output on any error. 2nd in family. Rule: once a retry has started, nothing that comes back without a verdict may drop attempt 1; any error from attempt 2 carries attempt 1''s text.'
          family: fail-safe-keeps-evidence
          round: 2
      boundary: M1
      recipe: milestone-review
      blocked: true
    - "n": 3
      timestamp: "2026-10-09T21:19:42-07:00"
      agent: claude
      dispose:
        - id: BR-7
          disposition: addressed
          note: Atlas sdlc-binary.md (#300 paragraph), judge-output-contract.md, plan D5 and the Core concepts row now describe RunFailure, ErrReviewDidNotRun, per-agent hosts, claude-only stream and the verdict-first guard; Revisions records the change. Checked against apifailure.go/dispatch.go.
          round: 3
        - id: BR-8
          disposition: addressed
          note: Dispatch returns attempt 1's labelled text with any error from attempt 2; TestDispatchRetryErrorKeepsTheFirstAttempt fails under the old (second, err) return; all four callers print the output on error.
          round: 3
      boundary: M1
      recipe: milestone-review
      blocked: false
    - "n": 4
      timestamp: "2026-10-09T21:36:38-07:00"
      agent: claude
      findings:
        - id: BR-9
          severity: Important
          title: A findings-less boundary round is BLOCKED in the ledger but still reported as Converging, ok, and unforced
          detail: 'gatePersist.Blocked only feeds the stored Blocked flag. ConvergenceLine (gatestate/family.go:154) prints "Converging." for 0 new / 0 repeats, stampAndPersist picks cok vs cwarn by d.Block alone, and forcedRationale(forced, d.Block) drops a --force waiver on a round now recorded blocked. Plan Task 5 is ticked as "doesn''t count as convergence". Rule: one effective outcome per round, consumed by every reader. Compute blocked := d.Block || g.Blocked once in stampAndPersist for Blocked, Forced and the warn/ok line, and have ConvergenceLine report a protocol-error round as not converging. Extend the D6 test to assert the stderr convergence line.'
          family: one-round-outcome-all-readers
          round: 4
        - id: BR-10
          severity: Minor
          title: Plan Task 5 names TestProtocolErrorRoundBlocks; the shipped test is TestBoundaryRoundWithoutFindingsIsNotPassed
          detail: '3rd finding in this family. Rule: ticking a Plan row requires that every named test exists under that name and asserts every clause of the row; check it when the row is ticked rather than fixing rows one at a time.'
          family: plan-claims-match-tests
          round: 4
        - id: BR-11
          severity: Minor
          title: reviewLimitLabel restates dispatchTimeout's precedence and the 30m default (ARCH-DRY)
          detail: milestoneclose.go reviewLimitLabel repeats the env-over-sized precedence and the 30m default, which now lives in three places. Export judge.EffectiveTimeout(env, sized) and label from its result, which also stops an invalid override from being printed as the limit.
          family: single-source-timeout
          round: 4
      boundary: M2
      recipe: milestone-review
      blocked: true
    - "n": 5
      timestamp: "2026-10-09T21:46:02-07:00"
      agent: claude
      dispose:
        - id: BR-9
          disposition: addressed
          note: 'stampAndPersist computes blocked once for the flag, the Forced stamp and the warn/ok line, and ConvergenceLine reports not converging. Mutation-checked: reverting either half fails TestBoundaryRoundWithoutFindingsIsNotPassed.'
          round: 5
        - id: BR-10
          disposition: addressed
          note: All 21 tests named in the Plan Tasks 1-6 rows exist under those names; the close steps sit in their own unticked row.
          round: 5
        - id: BR-11
          disposition: addressed
          note: reviewLimitLabel reads judge.EffectiveTimeout; an invalid override is labelled invalid rather than printed as the limit.
          round: 5
      findings:
        - id: BR-12
          severity: Important
          title: README.md:65 still says WF_REVIEW_TIMEOUT defaults to 30m; boundary reviews now scale with the window
          detail: '3rd finding in family single-source-timeout. Rule: a timeout-policy change sweeps every hit of grep -rn WF_REVIEW_TIMEOUT in the same commit; helptext and atlas were swept, README was missed.'
          family: single-source-timeout
          round: 5
        - id: BR-13
          severity: Minor
          title: ConvergenceLine decides "no readable findings" with its own predicate, separate from the Blocked stamp's review.Round == nil
          detail: '3rd finding in family. A findings block holding only an invalid disposition gets ProtocolError from ApplyChecked, so ConvergenceLine says "Not converging: no readable findings" while the gate prints [ok]. Rule: one predicate (a gatestate Round method or a stored flag) read by the stamp and by ConvergenceLine. The Forced-stamp reader also has no test: reverting it to d.Block keeps the D6 test green.'
          family: one-round-outcome-all-readers
          round: 5
        - id: BR-14
          severity: Minor
          title: The shared stampAndPersist hard-codes the boundary-specific "produced no findings block" wording for gatePersist.Blocked
          detail: Carry a BlockedReason alongside Blocked so the shared tail holds no gate-specific prose.
          family: channel-vs-prose-separation
          round: 5
      boundary: M2
      recipe: milestone-review
      blocked: true
    - "n": 6
      timestamp: "2026-10-09T21:55:59-07:00"
      agent: claude
      dispose:
        - id: BR-12
          disposition: addressed
          note: README.md:65 now describes the window-sized limit; a repo-wide git grep for WF_REVIEW_TIMEOUT finds no remaining "defaults to 30m" wording.
          round: 6
        - id: BR-13
          disposition: addressed
          note: Round.ProducedNothing is read by both the Blocked stamp (boundaryledger.go:195) and ConvergenceLine (family.go:158); reverting either the forcedRationale reader or the ConvergenceLine check turns TestBoundaryRoundWithoutFindingsIsNotPassed red (scratch mutation).
          round: 6
        - id: BR-14
          disposition: addressed
          note: gatePersist.BlockedReason carries the gate's own wording; stampAndPersist holds no boundary-specific text.
          round: 6
      boundary: M2
      recipe: milestone-review
      blocked: false
---

# Gate ledger — ariadne#300 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-09T20:50:49-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `late-signal-erases-verdict` APIFailure treats any is_error result as "review did not run" even when the run already carries a verdict
  apifailure.go:25 checks is_error before HasVerdict, so a reviewer that wrote its verdict and then hit error_during_execution/max_turns/an overload loses it, and the error misdirects to a sandbox fix. Check HasVerdict first, or narrow the message when no signature matches; add a fixture test.
- **BR-2** [Important] `channel-vs-prose-separation` ReadStream runs on codex/gemini text output, so a quoted stream-event JSON line can trip ErrAPIUnreachable
  dispatch.go:287 stream-parses every agent's stdout; a standalone {"type":"result","is_error":true} line in text-mode prose sets Stream=true, is dropped from Text(), and fails the dispatch. Gate stream parsing on AgentClaude and add a test.
- **BR-3** [Minor] `agent-specific-remedy` The "allow api.anthropic.com" remedy is hard-coded for every agent
- **BR-4** [Minor] `fail-safe-keeps-evidence` A retry that hits the deadline drops the first attempt's text from the error/sidecar
- **BR-5** [Minor] `plan-claims-match-tests` Task 2 claims close-level end-to-end and sidecar tests; the tests stop at Dispatch
- **BR-6** [Minor] `late-signal-erases-verdict` A stderr API signature on a clean-exit, verdict-less run suppresses the intended retry

## Round 2 — 2026-10-09T21:06:47-07:00 (claude) — BLOCKED

### Disposed

- BR-1 — addressed — RunFailure checks HasVerdict first; verdict_then_error.jsonl and maxTurns cases in TestRunFailure; reverting the guard turns the test red.
- BR-2 — addressed — classifyRunResult reads only claude stdout as a stream; TestDispatchParsesOnlyClaudeAsAStream goes red when the gate is removed.
- BR-3 — addressed — apiHosts maps each agent to its host; TestRunFailure asserts the codex and gemini remedies.
- BR-4 — addressed — The retry's DeadlineExceeded branch returns attempt 1 labelled; TestDispatchRetryDeadlineKeepsTheFirstAttempt pins it.
- BR-5 — addressed — TestCloseRecordsTheVerdictBeforeAPostscript and TestCloseWithoutAVerdictKeepsBothRuns exist in cmd/sdlc/closestream_test.go and pass.
- BR-6 — addressed — The stderr case requires nonZeroExit; dropping that condition turns TestRunFailure red.

### Raised

- **BR-7** [Important] `plan-claims-match-tests` Atlas, contract doc and plan Core concepts/D5 still describe the code before round 1
  atlas sdlc-binary.md:892 and construct/judge-output-contract.md say any is_error result is ErrAPIUnreachable with the sandbox fix. The plan table lists APIFailure(...)(cause, ok) with HasVerdict in classify.go, and D5 hard-codes api.anthropic.com. The code now keeps a verdict before an error, has ErrReviewDidNotRun, reads only claude as a stream, and names a host per agent. 2nd in family. Rule: a commit that changes a contract also updates every artifact describing it (plan, atlas, contract doc) in the same commit.
- **BR-8** [Minor] `fail-safe-keeps-evidence` A retry that fails with an error (not a timeout) drops attempt 1's text
  Dispatch returns (second, err) for ErrAPIUnreachable, ErrReviewDidNotRun, an interrupt or a launch failure on attempt 2, and milestoneclose.go:641 discards the output on any error. 2nd in family. Rule: once a retry has started, nothing that comes back without a verdict may drop attempt 1; any error from attempt 2 carries attempt 1's text.

## Round 3 — 2026-10-09T21:19:42-07:00 (claude) — passed

### Disposed

- BR-7 — addressed — Atlas sdlc-binary.md (#300 paragraph), judge-output-contract.md, plan D5 and the Core concepts row now describe RunFailure, ErrReviewDidNotRun, per-agent hosts, claude-only stream and the verdict-first guard; Revisions records the change. Checked against apifailure.go/dispatch.go.
- BR-8 — addressed — Dispatch returns attempt 1's labelled text with any error from attempt 2; TestDispatchRetryErrorKeepsTheFirstAttempt fails under the old (second, err) return; all four callers print the output on error.

## Round 4 — 2026-10-09T21:36:38-07:00 (claude) — BLOCKED

### Raised

- **BR-9** [Important] `one-round-outcome-all-readers` A findings-less boundary round is BLOCKED in the ledger but still reported as Converging, ok, and unforced
  gatePersist.Blocked only feeds the stored Blocked flag. ConvergenceLine (gatestate/family.go:154) prints "Converging." for 0 new / 0 repeats, stampAndPersist picks cok vs cwarn by d.Block alone, and forcedRationale(forced, d.Block) drops a --force waiver on a round now recorded blocked. Plan Task 5 is ticked as "doesn't count as convergence". Rule: one effective outcome per round, consumed by every reader. Compute blocked := d.Block || g.Blocked once in stampAndPersist for Blocked, Forced and the warn/ok line, and have ConvergenceLine report a protocol-error round as not converging. Extend the D6 test to assert the stderr convergence line.
- **BR-10** [Minor] `plan-claims-match-tests` Plan Task 5 names TestProtocolErrorRoundBlocks; the shipped test is TestBoundaryRoundWithoutFindingsIsNotPassed
  3rd finding in this family. Rule: ticking a Plan row requires that every named test exists under that name and asserts every clause of the row; check it when the row is ticked rather than fixing rows one at a time.
- **BR-11** [Minor] `single-source-timeout` reviewLimitLabel restates dispatchTimeout's precedence and the 30m default (ARCH-DRY)
  milestoneclose.go reviewLimitLabel repeats the env-over-sized precedence and the 30m default, which now lives in three places. Export judge.EffectiveTimeout(env, sized) and label from its result, which also stops an invalid override from being printed as the limit.

## Round 5 — 2026-10-09T21:46:02-07:00 (claude) — BLOCKED

### Disposed

- BR-9 — addressed — stampAndPersist computes blocked once for the flag, the Forced stamp and the warn/ok line, and ConvergenceLine reports not converging. Mutation-checked: reverting either half fails TestBoundaryRoundWithoutFindingsIsNotPassed.
- BR-10 — addressed — All 21 tests named in the Plan Tasks 1-6 rows exist under those names; the close steps sit in their own unticked row.
- BR-11 — addressed — reviewLimitLabel reads judge.EffectiveTimeout; an invalid override is labelled invalid rather than printed as the limit.

### Raised

- **BR-12** [Important] `single-source-timeout` README.md:65 still says WF_REVIEW_TIMEOUT defaults to 30m; boundary reviews now scale with the window
  3rd finding in family single-source-timeout. Rule: a timeout-policy change sweeps every hit of grep -rn WF_REVIEW_TIMEOUT in the same commit; helptext and atlas were swept, README was missed.
- **BR-13** [Minor] `one-round-outcome-all-readers` ConvergenceLine decides "no readable findings" with its own predicate, separate from the Blocked stamp's review.Round == nil
  3rd finding in family. A findings block holding only an invalid disposition gets ProtocolError from ApplyChecked, so ConvergenceLine says "Not converging: no readable findings" while the gate prints [ok]. Rule: one predicate (a gatestate Round method or a stored flag) read by the stamp and by ConvergenceLine. The Forced-stamp reader also has no test: reverting it to d.Block keeps the D6 test green.
- **BR-14** [Minor] `channel-vs-prose-separation` The shared stampAndPersist hard-codes the boundary-specific "produced no findings block" wording for gatePersist.Blocked
  Carry a BlockedReason alongside Blocked so the shared tail holds no gate-specific prose.

## Round 6 — 2026-10-09T21:55:59-07:00 (claude) — passed

### Disposed

- BR-12 — addressed — README.md:65 now describes the window-sized limit; a repo-wide git grep for WF_REVIEW_TIMEOUT finds no remaining "defaults to 30m" wording.
- BR-13 — addressed — Round.ProducedNothing is read by both the Blocked stamp (boundaryledger.go:195) and ConvergenceLine (family.go:158); reverting either the forcedRationale reader or the ConvergenceLine check turns TestBoundaryRoundWithoutFindingsIsNotPassed red (scratch mutation).
- BR-14 — addressed — gatePersist.BlockedReason carries the gate's own wording; stampAndPersist holds no boundary-specific text.

## Open findings

(none — every finding has been disposed)
