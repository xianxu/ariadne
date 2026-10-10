---
id: 000271
status: wontfix
deps: []
github_issue:
created: 2026-09-28
updated: 2026-10-09
estimate_hours:
card_mirror: 'de4ff529372391e91e5dd66fdf4636f3b377e2fc' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-09T22:06:04-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
---

# Boundary reviewer can background a command and end with no verdict

## Problem

The boundary reviewer dispatched by `sdlc close` / `milestone-close` (`claude -p` with an
`--allowedTools` list, `cmd/sdlc/internal/judge/dispatch.go:134`) can start a command in
the background and end its turn waiting for a notification that never comes in `-p` mode.
Its whole output is then that last message, with no verdict line and no `findings` block,
so the gate records `Review-Verdict: unknown` and refuses to finalize ("consult a human").

Observed: parley.nvim #282 close, 2026-09-28 23:28 (re-review round 3). The sidecar
`workshop/plans/000282-answer-single-undo-close-review.md` in parley.nvim holds the entire
review as: "I'll wait for the background test run to notify." Prior rounds on the same
issue produced normal verdicts, so it is intermittent: it depends on whether the reviewer
chooses to background its test run (parley's full suite takes minutes).

## Spec

- The reviewer cannot end on a backgrounded command: either the dispatch denies background
  execution (Bash `run_in_background`, Monitor, and similar) in its allowed tools, or the
  prompt states that the reviewer runs non-interactively, gets no notifications, and must
  run commands in the foreground (within the review timeout) before giving its verdict.
  Prefer the tool restriction; a prompt rule alone is advisory.
- Defense in depth at the gate: an output with neither a verdict nor a `findings` block is
  a dispatch failure, not a verdict. Retry the dispatch once, and only then stop with the
  existing "consult a human" message, naming the likely cause (the reviewer ended without
  a verdict).

## Done when

- A reviewer that tries to background a command either cannot (tool denied) or is told not
  to, covered by a test on the dispatch arguments or the prompt golden.
- A verdict-less review output is retried once before the gate stops, covered by a test
  with a fake reviewer that returns a verdict-less first answer.

## Plan

- [ ] Deny background execution in the reviewer's allowed tools, or add the prompt rule
- [ ] Retry a verdict-less dispatch once in the gate; test with a fake reviewer

## Log


- 2026-10-09: abandoned (wontfix): fixed by #300 (landed PR #173): the reviewer is denied background tools and its stream-json run keeps the latest verdict block, with one retry when none arrives; #300 tests cover #271's Done-when. wontfix stands in for a missing fixed/dup status (#212, #312)
- Filed from parley.nvim #282's close (operator request, 2026-09-28).

### 2026-09-28
