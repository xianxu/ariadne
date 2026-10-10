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

## Open findings

- **BR-1** [Important] `late-signal-erases-verdict` APIFailure treats any is_error result as "review did not run" even when the run already carries a verdict
- **BR-2** [Important] `channel-vs-prose-separation` ReadStream runs on codex/gemini text output, so a quoted stream-event JSON line can trip ErrAPIUnreachable
- **BR-3** [Minor] `agent-specific-remedy` The "allow api.anthropic.com" remedy is hard-coded for every agent
- **BR-4** [Minor] `fail-safe-keeps-evidence` A retry that hits the deadline drops the first attempt's text from the error/sidecar
- **BR-5** [Minor] `plan-claims-match-tests` Task 2 claims close-level end-to-end and sidecar tests; the tests stop at Dispatch
- **BR-6** [Minor] `late-signal-erases-verdict` A stderr API signature on a clean-exit, verdict-less run suppresses the intended retry
