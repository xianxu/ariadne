---
id: 000357
status: working
deps: []
github_issue:
created: 2026-10-13
updated: 2026-10-13
estimate_hours:
card_mirror: '9c934f155e7da97c7daa7516b1c234fe3e3d8f24' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-13T19:46:01-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
---

# test gate: the suite queue is polled, not FIFO, so a waiter can starve (51+ min) while later suites take the head

## Problem

ariadne#346's weighted test lock lets only the queue head take suite slots, but the head is whoever wins the poll on the queue file when it frees, not the oldest waiter. On 2026-10-13 pair:6's full suite (pair#538) waited 51+ minutes while later suites kept taking the head; pair:7 saw the same.

## Spec

Make admission FIFO: each waiter takes a ticket (an append-only sequence, e.g. a monotonically numbered file or a counter under the queue lock) and only the lowest live ticket may become head; dead waiters' tickets are skipped (flock or pid check), so a crash never blocks the line. Keep weights, the atomic multi-slot acquire and the no-deadlock property; the waiting message shows the waiter's position.

## Done when

- With three waiters arriving in order, they acquire in that order even when a later one polls first; a dead waiter's ticket is skipped; tests use private lock files and sleep holders only (never real load).

## Plan

- [x] Tickets in `scripts/test-lock.py`: each waiter creates `<slot 0>.tickets/NNNNNNNNNNNN`
  (max live + 1) and holds an exclusive flock on it while waiting; creation and pruning
  happen under a short `alloc` flock, so a just-created ticket is never read as dead.
  A ticket that can be locked has a dead waiter and is unlinked. Only the oldest live
  waiter may take the head, and the head still takes the old `.queue` file, so an older
  copy's head goes first in a mixed fleet. Weights, all-or-none acquire and the
  head-holds-no-slot (no deadlock) rule are unchanged. A nested run passes through
  without joining the line. The wait shows `position K in line`.
- [x] Tests (private lock files, sleep holders): arrival order with a stopped waiter
  (a later waiter polling first does not overtake), a `kill -9` waiter skipped, no
  ticket left behind, an older copy's queue head going first.
- [x] Move `scripts/test-lock.test.sh` to `scripts/test/`, where the
  `20-ci-workflow-tests` merge check runs it (it had no runner); atlas updated.

## Log

### 2026-10-13

### 2026-10-13 — ariadne:3

Implemented as planned above. Proof the FIFO cases have teeth: with main's polled
`test-lock.py` swapped in, "a later waiter that polls first does not overtake" and the
position case fail (the dead-waiter case passes there too, since the old gate has no
tickets to block on). My first test draft backgrounded the `lock` shell function, so `$!`
was a subshell: `kill -STOP`/`kill -9` never reached the waiter, and the stopped-waiter
case passed without exercising the scenario. The waiters now start as `python3` itself
and the case also asserts the stopped waiter has not acquired. Lifecycle (ARCH-FUNERAL):
a ticket lives exactly as long as its waiter's wait; it is unlinked on acquire, timeout
or nesting, and a crashed waiter's ticket by the next allocation or position check;
`alloc` is one persistent empty file. The moved test takes 39s (local-gates-guard.test.sh
takes 87s).
