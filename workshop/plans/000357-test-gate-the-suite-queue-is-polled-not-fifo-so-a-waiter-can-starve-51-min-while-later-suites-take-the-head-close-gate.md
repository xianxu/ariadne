---
gate: boundary-review
issue: 357
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-13T20:19:22-07:00"
      dispatched: "2026-10-13T20:14:40-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Minor
          title: FIFO tests order waiter arrival with sleep 0.5, which can flake under CI load
          detail: scripts/test/test-lock.test.sh:119-121,134-135 start waiters 0.5s apart; each runs python startup plus git rev-parse before taking a ticket. Poll until the tickets dir holds N entries before starting the next waiter.
          family: sleep-ordered-test-arrival
          round: 1
        - id: BR-2
          severity: Minor
          title: A stopped (Ctrl-Z) non-head waiter now blocks every later waiter indefinitely
          detail: Before, only the head could stall the line. Either document in the atlas that a suspended waiter holds its place, or skip tickets whose waiters stop updating them (e.g. a stale mtime heartbeat).
          family: strict-fifo-liveness
          round: 1
        - id: BR-3
          severity: Minor
          title: Tickets.locked() defines an ad-hoc context-manager class; contextlib.contextmanager is simpler
          family: hand-rolled-context-manager
          round: 1
      recipe: milestone-review
      isolation: checkout
      reviewed: 83634985882f1566e69c81d746218cacb9ce1d50
      blocked: false
    - "n": 2
      timestamp: "2026-10-13T20:23:07-07:00"
      dispatched: "2026-10-13T20:21:56-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: await_tickets polls the numbered-ticket count before each next waiter, in both FIFO blocks (test-lock.test.sh:123-125,139-140); suite passes at head.
          round: 2
        - id: BR-2
          disposition: addressed
          note: Documented as the rule in the script header (test-lock.py:27-29) and atlas sdlc-binary.md:1644-1645; existing SIGSTOP test pins the behavior.
          round: 2
        - id: BR-3
          disposition: addressed
          note: Tickets.locked now @contextlib.contextmanager with try/finally unlock (test-lock.py:128-134).
          round: 2
      recipe: fix-review
      isolation: checkout
      reviewed: 603f60e74f8b2e3036374febe5eb650dd47ef1ed
      blocked: false
---

# Gate ledger — ariadne#357 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-13T20:19:22-07:00 (claude) — passed

### Raised

- **BR-1** [Minor] `sleep-ordered-test-arrival` FIFO tests order waiter arrival with sleep 0.5, which can flake under CI load
  scripts/test/test-lock.test.sh:119-121,134-135 start waiters 0.5s apart; each runs python startup plus git rev-parse before taking a ticket. Poll until the tickets dir holds N entries before starting the next waiter.
- **BR-2** [Minor] `strict-fifo-liveness` A stopped (Ctrl-Z) non-head waiter now blocks every later waiter indefinitely
  Before, only the head could stall the line. Either document in the atlas that a suspended waiter holds its place, or skip tickets whose waiters stop updating them (e.g. a stale mtime heartbeat).
- **BR-3** [Minor] `hand-rolled-context-manager` Tickets.locked() defines an ad-hoc context-manager class; contextlib.contextmanager is simpler

## Round 2 — 2026-10-13T20:23:07-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — await_tickets polls the numbered-ticket count before each next waiter, in both FIFO blocks (test-lock.test.sh:123-125,139-140); suite passes at head.
- BR-2 — addressed — Documented as the rule in the script header (test-lock.py:27-29) and atlas sdlc-binary.md:1644-1645; existing SIGSTOP test pins the behavior.
- BR-3 — addressed — Tickets.locked now @contextlib.contextmanager with try/finally unlock (test-lock.py:128-134).

## Open findings

(none — every finding has been disposed)
