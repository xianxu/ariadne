---
id: 000391
status: open
deps: []
github_issue:
created: 2026-10-15
updated: 2026-10-15
estimate_hours:
card_mirror: '9adc76f86846e68a4fbda5713c8617b80574a841' # card fields mirrored from issue-cards; edit via sdlc
---

# sdlc actuals fail machine-wide on a Qoder subagent transcript ('incomplete qoder session metadata') after #385

## Problem

After ariadne#385 (Qoder transcripts in sdlc) landed, `sdlc close` refuses to compute actuals on this machine:
"can't auto-compute, source ~/.qoder/projects/-Users-xianxu-workspace-pair/6d3388c7-…/subagents/agent-aExplore-….jsonl:
incomplete qoder session metadata" (ariadne:4 closing #387, 2026-10-15). A Qoder **subagent** transcript (under a
session's `subagents/` directory) lacks the session metadata the new reader requires, and one unreadable file makes
the whole measurement fail. Since every close/actual scans the transcript roots, this likely blocks **every**
`sdlc close` and `sdlc actual` on any machine with Qoder subagent files.

## Spec

- Treat Qoder subagent transcripts correctly: either attribute them to their parent session (the directory says which)
  or skip them for session identity, but never fail the whole measurement because of them.
- More generally: a transcript file the reader can't classify is reported as skipped with its path and reason, and
  measurement proceeds over the rest, unless the unreadable file is one the issue's own window provably depends on.
- A fixture from the real subagent layout (trimmed, no secrets).

## Done when

- `sdlc actual --issue 387` (or any issue) on this machine computes again with the subagent file present.
- Tests: a subagent transcript next to a session is attributed or skipped as specified; an unclassifiable file is
  reported, not fatal.

## Plan

- [ ]

## Log

### 2026-10-15
