---
id: 000300
status: working
deps: []
github_issue:
created: 2026-10-06
updated: 2026-10-09
estimate_hours: 2.29
card_mirror: 'f57df793f5bb3e022d6145a2b82f58617db15588' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-09T18:33:10-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
---

# sdlc judge: a reviewer's late message replaces its verdict, so close records 'unknown'

## Problem

The boundary reviewer runs as `claude -p` (`cmd/sdlc/internal/judge/dispatch.go`,
`BuildArgs`). Plain `-p` prints only the agent's final assistant message, and the
gate parses the verdict and findings from that output.

When the reviewer starts a background job, such as a long test run, it writes its
verdict, then receives the job's completion and writes a short postscript. The
postscript becomes the final message, so the gate sees no `verdict` block and no
`findings` block. It records the verdict as `unknown`, and the close stops with
"consult a human". The real verdict and findings are lost from the sidecar.

This happened twice in two days, both in pair:
- **pair#387 close, round 7.** The prose said FIX-THEN-SHIP with one Minor
  finding; no block was captured.
- **pair#362 close, round 1.** The reviewer's earlier message held a full
  FIX-THEN-SHIP verdict with one Important and five Minor findings. The final
  message said "All four findings stand as written". The verdict was recovered
  by reading the reviewer's transcript
  (`~/.claude/projects/<worktree>/<session>.jsonl`) by hand.

The fail-safe worked: neither close finalized on a missing verdict. The cost is a
wasted round, a human interrupt and, without the manual transcript read, lost
findings.

## Spec

Captured for operator review; no implementation is authorized by this issue
creation.

- **Capture the whole run, not the last message.** Read the reviewer's output as a
  structured stream (`claude -p --output-format stream-json`, with the equivalents
  for codex and gemini where they exist). Take the latest assistant message that
  carries a `verdict` block, and the latest `findings` block. A later message
  without a block never erases an earlier one.
- **Keep the reviewer in the foreground.** Tell it in the contract preamble not to
  start background work, and, where the agent supports it, withhold the tools
  that wait on background work. This makes the case rare; the capture fix makes it
  harmless.
- **Keep the fail-safe.** If no message in the run carries a verdict, the verdict
  is still `unknown` and the close does not finalize. The sidecar then keeps the
  full transcript, or a pointer to it, so the human has something to read.
- **Related, separate:** pair#392 (the review pane notifies when an agent round
  comes back empty) concerns the interactive review workbench, not the sdlc
  judge.

## Done when

- A reviewer run whose verdict is followed by a block-less postscript is parsed
  to the earlier verdict and findings. A test feeds a recorded stream with that
  shape: verdict and findings, then a late message.
- A run with no verdict anywhere still records `unknown` and refuses to finalize,
  and its sidecar points at the run's full output.
- The contract preamble tells reviewers not to start background work.

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: greenfield-go-module   design=0.2 impl=0.32
item: smaller-go-module      design=0.1 impl=0.2
item: smaller-go-module      design=0.05 impl=0.12
item: smaller-go-module      design=0.05 impl=0.12
item: smaller-go-module      design=0.05 impl=0.1
item: smaller-go-module      design=0.05 impl=0.16
item: real-api-discovery     design=0.0 impl=0.18
item: atlas-docs             design=0.05 impl=0.04
item: milestone-review       design=0.0 impl=0.14
item: milestone-review       design=0.0 impl=0.14
item: milestone-review       design=0.0 impl=0.14
design-buffer: 0.15
total: 2.29
```

Items, in order:
- `ReadStream`;
- `Dispatch` through the stream, with the retry;
- `HasVerdict`, the last-match token and the unattended rule with goldens;
- `APIFailure`;
- the boundary ledger fix;
- the sized timeout and its plumbing;
- the live stream-format check;
- atlas;
- the M1, M2 and close reviews.

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only.*

## Plan

Durable plan: `workshop/plans/000300-sdlc-judge-a-reviewer-s-late-message-replaces-its-verdict-so-close-records-unknown-plan.md`.

- [x] M1 — Verdict capture: claude stream-json, the latest verdict/findings block wins, one retry when no verdict, the unattended rule in the contract (D1, #271). Plan Tasks 1–3.
- [ ] M2 — Honest failures: an unreachable API is "review did not run", a protocol-error round is never "passed", the timeout scales with the window (D2, D3). Plan Tasks 4–6.

## Log

### 2026-10-06

Filed at the operator's request after the second occurrence (pair#387, pair#362).
The recovered pair#362 round is in pair's
`workshop/plans/000362-couch-schedule-this-close-review.md` history. Details left
local for the operator to refine.

### 2026-10-09
- 2026-10-09: closed M1 — make test green except sandbox-only processgroup. M1: TestDispatchKeepsAVerdictBeforeAPostscript, TestDispatchRetriesARunWithoutAVerdict, TestDispatchTwoRunsWithoutAVerdictStayUnknown, TestDispatchRetryDeadlineKeepsTheFirstAttempt, TestDispatchRetryErrorKeepsTheFirstAttempt, TestDispatchAPIErrorIsReviewDidNotRun, TestRunFailure, TestDispatchParsesOnlyClaudeAsAStream, TestCloseRecordsTheVerdictBeforeAPostscript, TestCloseWithoutAVerdictKeepsBothRuns; unattended rule in all 8 goldens; live claude stream conformance passes.; review verdict: SHIP
- 2026-10-09: unclaimed: TL slot hands off: shaped scope recorded in Log; implement in :2-:4

Claimed in ariadne:1 as the first issue of project `ariadne-robustness-1` (`workshop/projects/ariadne-robustness-1.md`). The project file and the evidence pensive (`workshop/pensive/2026-10-09-01-pensive-sdlc-robustness-evidence.md`) land separately, on the plain-git branch `project-ariadne-robustness-1`.

Scope event: absorbs #271 (a backgrounded command leaves no verdict; same failure family, designed together), plus two transcript findings:
- **D2:** the sandbox blocks the judge's API host, and the ledger records the failed round as `blocked: false … passed`;
- **D3:** the 30-minute review timeout is routinely overridden to `WF_REVIEW_TIMEOUT=2h`.

Evidence and citations are in the pensive, Part 2 §D.

Taken over in ariadne:2 on the ariadne:1 lead slot's dispatch (Couch receipt verified). Survey of the judge: claude runs `-p` in text mode, so only the final message is captured; allowed tools are `Read,Grep,Glob,Bash`, with nothing against background Bash; nothing reads stderr or looks for API errors; a boundary round without a findings block is stamped `blocked: false` by `stampAndPersist` (D2's "passed"); the 30-minute timeout is fixed, and no size reaches dispatch. The durable plan sets out D1–D7.

M1 built. Live stream samples were recorded on 2026-10-09: a clean run, and the same command inside the sandbox, which gives `system/api_retry` events, then a synthetic assistant 'API Error … ERR_PROXY_TUNNEL' message, then `result.is_error: true` and exit 1. The fixtures mirror both. `ReadStream` with `AgentRun` keeps every message and the unparsed tail. `HasVerdict` is the retry predicate, recognising everything any parser reads; an existing test caught that the legacy sentinels were missing. Dispatch retries once within one deadline; a stream error is `ErrAPIUnreachable` ('review did not run', names the sandbox fix, no retry). `ParseVerdictToken` is now last-match. `UnattendedRule` is in both contracts: the golden diff is exactly the rule in 8 prompts, the human contract doc is updated, and the live conformance test passes against the real claude in stream mode. Retry, capture and is_error are each mutation-checked. Full suite: two wall-clock concurrency tests failed under load average ~50 (other slots running suites) and pass alone.

M1 review round 1 (FIX-THEN-SHIP). BR-1: a run that already gave its verdict is never a failure; a late error result no longer erases the verdict, which is the same family as the postscript (`verdict_then_error.jsonl`). An error without a network cause is `ErrReviewDidNotRun` and doesn't point at the sandbox. BR-2: only claude's stdout is read as a stream, so a quoted event line in codex or gemini prose can't fail the review. Minors: the remedy names each agent's API host; a retry that runs out of time keeps attempt 1 as evidence and leaves the fail-safe to the caller; a stderr signature counts only with a non-zero exit, so a clean-exit verdict-less run is retried. The close-level tests now exist (`TestCloseRecordsTheVerdictBeforeAPostscript`, `TestCloseWithoutAVerdictKeepsBothRuns`).
