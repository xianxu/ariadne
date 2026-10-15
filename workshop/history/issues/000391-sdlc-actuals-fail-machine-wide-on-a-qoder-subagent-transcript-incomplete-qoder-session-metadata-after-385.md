---
id: 000391
status: done
deps: []
github_issue:
created: 2026-10-15
updated: 2026-10-15
estimate_hours:
card_mirror: 'd4df68a0136b9711f5b69ae4b7fde2cd7f207f14' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-15T13:31:36-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: operator, spec: "e0dbb6a8", done: "c4c4b671"}
actual_hours: N/A
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
- Tests: the native session/subagents/agent-*.jsonl layout preserves the valid parent through actual and close adoption, with a skipped path/reason warning. Unclassifiable files and unreadable directories warn; positively requested-workspace damage and cancellation still refuse adoption.

## Plan

- [x] Reproduce native nested subagent layout and unclassifiable-file failures with fixtures.
- [x] Warn and skip unclassifiable Qoder sources; keep positively owned failures and cancellation fatal through shared discovery (ARCH-DRY). Identity already scans only top-level sessions.
- [x] Verify actual/close adoption and run the recorded full suite. SDLC close/landing and peer notifications follow the verification checkpoint.

## Log

### 2026-10-15
- 2026-10-15: closed — Native-layout regression reproduced before fix; actual/close and identity tests pass; transcript race tests pass; candidate actual #387 computes with eight skipped child warnings. Recorded full suite passed 66708c61 (1495 cmd/sdlc tests, 63 packages, zero failures, 612.5s, log b0e9c8ae0658). No actual recorded: installed shared sdlc fails on the same native child, so hours cannot be measured by the gate without fabricating; remeasure after landing/rebuild.; review verdict: SHIP

Measured native child: project/parent-session/subagents/agent-aExplore-ID.jsonl contains user/assistant rows with parent sessionId, agentId and isSidechain=true, but no workspace-directories or runtime-config. Recursively requiring standalone session metadata makes this unresolved file fatal fleet-wide. Use the operator-approved skip-with-diagnostic approach rather than inventing inherited metadata. Existing limits, immutable prefix validation, known-owner failures and cancellation remain enforced; no new state or retained artifacts (ARCH-PURE, ARCH-SECURE, ARCH-FUNERAL).

Regression red: native child and four unclassifiable scenarios failed actual/close adoption before the fix. Green: all root TestQoder tests passed (42.575s); transcripts normal and race suites passed. Candidate actual --issue 387 computes with eight explicit skipped-source warnings. Shared actual --issue 391 still fails on the first child; close will use shared sdlc with --no-actual (known installed measurement regression, no invented hours), then remeasurement can follow the rebuild.

Full suite passed on 66708c61fbe9beeefb981afddda9d271c60a88ac: 12 shards / 1,495 cmd/sdlc tests and 63 other packages, zero failures, 612.5s runtime after 967s queue. Recorded sdlc-local/suite; log sha256 b0e9c8ae0658. PR #261 CI green before close.

## Revisions

- 2026-10-15 — Clarify acceptance with the measured native layout and the known-owner/cancellation controls. Scope unchanged; the final plan checkbox covers pre-close verification rather than circularly requiring landing before close.
