---
id: 000391
status: working
deps: []
github_issue:
created: 2026-10-15
updated: 2026-10-15
estimate_hours:
card_mirror: '7a476c6e06b10d7d3a2984d679cf51d6a5935b5a' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-15T13:31:36-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
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

- [ ] Reproduce native nested subagent layout and unclassifiable-file failures with fixtures.
- [ ] Warn and skip unclassifiable Qoder sources; keep positively owned failures and cancellation fatal through shared discovery (ARCH-DRY). Identity already scans only top-level sessions.
- [ ] Verify actual/close adoption, run recorded suite, close on SHIP and land; notify ariadne:4 and hub.

## Log

### 2026-10-15

Measured native child: project/parent-session/subagents/agent-aExplore-ID.jsonl contains user/assistant rows with parent sessionId, agentId and isSidechain=true, but no workspace-directories or runtime-config. Recursively requiring standalone session metadata makes this unresolved file fatal fleet-wide. Use the operator-approved skip-with-diagnostic approach rather than inventing inherited metadata. Existing limits, immutable prefix validation, known-owner failures and cancellation remain enforced; no new state or retained artifacts (ARCH-PURE, ARCH-SECURE, ARCH-FUNERAL).
