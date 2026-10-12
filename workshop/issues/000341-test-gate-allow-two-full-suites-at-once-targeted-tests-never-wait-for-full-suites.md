---
id: 000341
status: open
deps: []
github_issue:
created: 2026-10-12
updated: 2026-10-12
estimate_hours:
card_mirror: 'e7f337ea9ace5dc95986b1c34344fd5cd4750366' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-12T16:13:39-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
---

# test gate: allow two full suites at once; targeted tests never wait for full suites

## Problem

10-12: the machine-wide test gate serializes everything. One full suite at a time (scripts/test-lock.py, flock on ~/.local/state/ariadne/test-suite.lock, shared by every repo), and targeted test runs (pkg/testgate, test-gate.<k>.lock slots) also wait while any full suite runs. With 4–5 slots each needing a full suite to close, work went single-file (5–10 min per suite; slots waited half an hour) while the machine sat at load ~2. The strict rule was right at load ~200 this morning; with slots' spawned work moving to nice 10 (pair#502) and agent churn on a no-journal scratch volume (pair#496) it wastes the machine.

## Spec

Decided by the operator, 10-12:
1. **Full suites: up to 2 concurrently** machine-wide (configurable, e.g. `WF_TEST_SUITES=2`), still flock-based so a dead holder releases automatically; the waiting message names both holders.
2. **Targeted tests never wait for full suites:** they keep only their own small concurrency cap (the existing test-gate slots, configurable), independent of the suite lock.
3. The atlas states the rule (what takes which lock, the defaults, the env vars), and the agent-facing docs say targeted runs are always fine.
Keep `WF_TEST_LOCK=off` as the logged emergency hatch.

## Done when

- Two full suites run at once and a third waits; a targeted test run starts immediately while two suites run; tests for both (with fakes for the lock files, no real load generated).
- The atlas updated.

## Plan

- [ ]

## Log

### 2026-10-12
