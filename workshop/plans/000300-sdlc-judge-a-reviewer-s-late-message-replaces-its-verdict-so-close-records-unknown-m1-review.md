# Boundary Review — ariadne#300 (milestone M1)

| field | value |
|-------|-------|
| issue | 300 — sdlc judge: a reviewer's late message replaces its verdict, so close records 'unknown' |
| repo | ariadne |
| issue file | workshop/issues/000300-sdlc-judge-a-reviewer-s-late-message-replaces-its-verdict-so-close-records-unknown.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | b51060b253657fa9a2c6827e330eb1e4c995f12e..c71732d979813be4b1c3057c2476795da09889ba |
| command | sdlc milestone-close --issue 300 --milestone M1 |
| reviewer | claude |
| timestamp | 2026-10-09T20:50:48-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Inspected the dispatch path; now checking whether the close-level sidecar promise (Done-when #2) has a test, then running the judge tests.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

M1 does what it set out to do. The claude reviewer now runs with `--output-format stream-json --verbose`. The pure `ReadStream` turns every assistant message into one joined text, so the existing parsers (each takes the last block) find the latest verdict and findings. `ParseVerdictToken` is now last-match. `Dispatch` retries a run with no verdict once, inside one deadline. `UnattendedRule` reaches all 8 goldens, and those golden diffs contain the rule and nothing else. `go test ./internal/judge/` passes.

Two cheap correctness gaps remain in the new failure classification, and neither blocks the boundary:
- **A verdict is thrown away when the result is an error.** A run whose terminal `result` has `is_error: true` is classified "review did not run" before the code checks for a verdict. A run that wrote its verdict and then failed on a later turn loses that verdict, and the error tells the operator to fix a sandbox the run never hit.
- **Codex and gemini output is parsed as a stream.** `ReadStream` is applied to every agent. A text-mode reviewer that quotes a stream-event JSON line (likely when reviewing this very diff) is misread as a stream, and the quoted line can trip `APIFailure`.

**Strengths**
- `Dispatch` still returns a plain string and reuses the parsers that already take the last block, so no per-message parser was added (ARCH-DRY; `dispatch.go:191-212`).
- `HasVerdict` (`stream.go:99`) matches everything any parser reads, including the legacy sentinels and the bare boundary token. The existing test (`"No DRY violations found."`) shows why that breadth is needed.
- The retry shares the deadline with the first attempt, and when both attempts fail the returned text keeps both runs, labelled.
- `ReadStream` falls back gracefully: a truncated stream keeps its unparsed tail, and output with no events is read as plain text (`stream.go:46-79`). All the fixtures are tested.
- The live conformance test now runs in stream mode and checks that the result is not an error.

**Critical:** none.

**Important**
1. `apifailure.go:25`: an `is_error` result wins over a verdict that is already present. It also covers more than API failures: in claude's stream, `error_during_execution` and `error_max_turns` set it too. **Fix:** check `HasVerdict` first, and when a verdict is present, record the error as a note rather than a failure. Otherwise keep D5's channel rule, but say "the run reported an error" instead of naming the sandbox, unless a signature such as `ERR_PROXY_TUNNEL` actually matches.
2. `dispatch.go:287`: `ReadStream` runs for every agent, but D1 says codex and gemini stay in text mode. Any standalone JSON line with a `type` key in their prose sets `Stream=true` and is dropped from `Text()`. A line like `{"type":"result","is_error":true,…}` (the fixtures' shape) produces a false `ErrAPIUnreachable`. That breaks D5's promise that the review's prose is never scanned (ARCH-SECURE). **Fix:** stream-parse only when the agent is claude; otherwise use `AgentRun{Tail: stdout}`. Add a test with a codex output that quotes `api_error.jsonl`.

**Minor**
- The error's remedy text "allow api.anthropic.com" is hard-coded for all agents (`dispatch.go:290`); codex and gemini use other hosts.
- If the retry hits the deadline, the error drops the first attempt's text, so the sidecar has nothing to read.
- The last-match change to `ParseVerdictToken` also affects the pre-merge `Classify` judges. A `VERDICT:` line quoted later in a findings section now overrides the leading one. This is acceptable given the contract, but it is a behavior change.
- Plan Task 2 says the tests drive "the close end to end … with its sidecar". The new tests stop at `Dispatch`. Done-when #2 (the sidecar points at the full output) holds only by following the code (`rr.Output = output`); no close-level test checks it.
- A stderr signature without a verdict is classified as a failure even after a clean exit. A transient retry warning on a run that then waited on a background job would skip the intended retry.

**Test coverage**
- The postscript, retry, double-failure, API-error and timeout-precedence cases are covered against the `Run` seam fake (ARCH-MOCK: pass).
- Two cases are missing: an `is_error` result after a verdict, and a text-mode agent quoting a stream line. Both match Important findings 1 and 2.
- No test checks that the retry's prompt carries `retryNotice`. A one-line check in the fake would cover it.

**Architecture**
- **ARCH-DRY: pass.** The parsers are reused. `HasVerdict` repeats `Classify`'s sentinel checks, but on purpose, to serve as a predicate.
- **ARCH-PURE: pass.** `ReadStream`, `HasVerdict` and `APIFailure` are pure, and `classifyRunResult` stays thin.
- **ARCH-PURPOSE: pass.** Codex and gemini are an explicitly declared non-goal; the sized timeout and the ledger fix are scheduled for M2.
- **ARCH-MOCK: pass.** The tests use the process-seam fake and the live check is updated.
- **ARCH-CONSTRAINTS: pass.** The retry is bounded to one attempt within one deadline.
- **ARCH-SECURE: flag.** See Important 2: text-mode prose is read on the stream channel.
- **ARCH-ORDER: pass.** The retry runs in sequence and no state is carried between events.
- **ARCH-FUNERAL: pass.** Nothing new is durable; the sidecar grows by at most one retry's text.

**Architectural notes for upcoming work**
- M2's D6 ledger fix and D7 sizing should plumb `DispatchOptions.Timeout` from `milestoneclose.go:640` and `planningreview.go:116`.
- `judge.go:177` (the pre-merge path) wraps `ErrAPIUnreachable` as "dispatch failed", which is fine.

**Plan revisions**
- Add a Revisions note that Task 2's tests run at the `Dispatch` level only. Either add a close-level sidecar test for Done-when #2 in M2, or reword the Task 2 claim.
- D5: record that a run carrying a verdict is not reclassified because of `is_error` (if adopted), and that stream parsing is limited to claude.

```findings
findings:
  - id: new
    severity: Important
    family: late-signal-erases-verdict
    title: |
      APIFailure treats any is_error result as "review did not run" even when the run already carries a verdict
    detail: |
      apifailure.go:25 checks is_error before HasVerdict, so a reviewer that wrote its verdict and then hit error_during_execution/max_turns/an overload loses it, and the error misdirects to a sandbox fix. Check HasVerdict first, or narrow the message when no signature matches; add a fixture test.
  - id: new
    severity: Important
    family: channel-vs-prose-separation
    title: |
      ReadStream runs on codex/gemini text output, so a quoted stream-event JSON line can trip ErrAPIUnreachable
    detail: |
      dispatch.go:287 stream-parses every agent's stdout; a standalone {"type":"result","is_error":true} line in text-mode prose sets Stream=true, is dropped from Text(), and fails the dispatch. Gate stream parsing on AgentClaude and add a test.
  - id: new
    severity: Minor
    family: agent-specific-remedy
    title: |
      The "allow api.anthropic.com" remedy is hard-coded for every agent
  - id: new
    severity: Minor
    family: fail-safe-keeps-evidence
    title: |
      A retry that hits the deadline drops the first attempt's text from the error/sidecar
  - id: new
    severity: Minor
    family: plan-claims-match-tests
    title: |
      Task 2 claims close-level end-to-end and sidecar tests; the tests stop at Dispatch
  - id: new
    severity: Minor
    family: late-signal-erases-verdict
    title: |
      A stderr API signature on a clean-exit, verdict-less run suppresses the intended retry
```
