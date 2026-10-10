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

## Open findings

- **BR-7** [Important] `plan-claims-match-tests` Atlas, contract doc and plan Core concepts/D5 still describe the code before round 1
- **BR-8** [Minor] `fail-safe-keeps-evidence` A retry that fails with an error (not a timeout) drops attempt 1's text
