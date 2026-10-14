---
id: 000357
status: open
deps: []
github_issue:
created: 2026-10-13
updated: 2026-10-13
estimate_hours:
card_mirror: '4f96201d0f2b491675f33712eaddb34d72792084' # card fields mirrored from issue-cards; edit via sdlc
---

# test gate: the suite queue is polled, not FIFO, so a waiter can starve (51+ min) while later suites take the head

## Problem

ariadne#346's weighted test lock lets only the queue head take suite slots, but the head is whoever wins the poll on the queue file when it frees, not the oldest waiter. On 2026-10-13 pair:6's full suite (pair#538) waited 51+ minutes while later suites kept taking the head; pair:7 saw the same.

## Spec

Make admission FIFO: each waiter takes a ticket (an append-only sequence, e.g. a monotonically numbered file or a counter under the queue lock) and only the lowest live ticket may become head; dead waiters' tickets are skipped (flock or pid check), so a crash never blocks the line. Keep weights, the atomic multi-slot acquire and the no-deadlock property; the waiting message shows the waiter's position.

## Done when

- With three waiters arriving in order, they acquire in that order even when a later one polls first; a dead waiter's ticket is skipped; tests use private lock files and sleep holders only (never real load).

## Plan

- [ ]

## Log

### 2026-10-13
