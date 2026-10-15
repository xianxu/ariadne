---
id: 000387
status: open
deps: []
github_issue:
created: 2026-10-15
updated: 2026-10-15
estimate_hours:
card_mirror: 'daba20f49597c76f0d2350a4d36e9289a3b56391' # card fields mirrored from issue-cards; edit via sdlc
---

# Agent: trailer says 'human' (not 'sdlc') for a commit made outside an agent session

## Problem

ariadne#367's `Agent:` trailer says `Agent: sdlc` when `sdlc` runs in a plain shell with no agent session, i.e. when
the operator makes the commit by hand. Velocity and usage numbers then credit the operator's hand work to "sdlc". The
operator (2026-10-15): "say human in that case".

## Spec

- The no-agent-session fallback renders `Agent: human` (agentid's fallback constant), and `Parse` accepts `human`;
  existing `sdlc` trailers in history still parse (treated as human for attribution).
- `sdlc agent-id` in a plain shell prints `Agent: human`.
- Docs (AGENTS.base.md §12, help) updated.

## Done when

- Tests: plain-shell fallback → `human`; Parse of `human` and of legacy `sdlc`; the commit-site sweep still passes.

## Plan

- [ ]

## Log

### 2026-10-15
