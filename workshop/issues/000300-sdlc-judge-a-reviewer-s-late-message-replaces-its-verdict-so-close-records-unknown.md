---
id: 000300
status: open
deps: []
github_issue:
created: 2026-10-06
updated: 2026-10-06
estimate_hours:
card_mirror: 'c11ebf4089f282088ac168dc63e8835e87c51094' # card fields mirrored from issue-cards; edit via sdlc
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

## Plan

- [ ]

## Log

### 2026-10-06

Filed at the operator's request after the second occurrence (pair#387, pair#362).
The recovered pair#362 round is in pair's
`workshop/plans/000362-couch-schedule-this-close-review.md` history. Details left
local for the operator to refine.
