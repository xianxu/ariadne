---
id: 000619
status: open
deps: []
github_issue:
created: 2026-10-24
updated: 2026-10-24
estimate_hours:
card_mirror: 'd7aabd9dd1a5603c20afaecb35cd5deefbaa557b' # card fields mirrored from issue-cards; edit via sdlc
---

# ARCH-EVOLVE: transitional compatibility for new invariants, and live state counts as persisted state

## Problem

## Spec

**Extend ARCH-EVOLVE to transitional compatibility and live state (operator, 10-24).** ARCH-EVOLVE today covers
persisted formats evolving additively. On 10-24 pair#866 added a new requirement on *existing live state* (a readiness
marker on running Zellij sessions) with no transition, so every session created before it was refused and the
operator's warm start broke. The operator's guidance, to be recorded as the principle:

- **Transitional phase:** a change to what existing state must look like ships with a period in which the system
  works with both old and new state; the old path is removed only once no old state can remain. With one operator
  today that period is roughly the longest-running session on the host; once released to others it becomes very
  hard to retire an old format, because you don't control who still uses it, so real systems carry historical
  compatibility code, and designs should expect that.
- **Live state counts:** state that crosses a build boundary is persisted state, even when it isn't in our files:
  running sessions and sockets (Zellij), other processes' memory, OS objects (SysV segments, locks), the thread store,
  registries, and remote services. A new build must accept what the previous build left running.
- **at-plan:** for each change that adds or tightens an invariant on existing persisted or live state, say how state
  from the previous version is handled during the transition, how the transition ends, and which test proves the
  upgrade (pair#879's fixture where it applies).

## Done when

- The ARCH-EVOLVE registry entry (or a sibling entry) carries the transitional-phase and live-state rules and the
  at-plan question; `sdlc arch-principles` and start-plan show it.
- ariadne#618's close-review question references it.

## Plan

- [ ]

## Log

### 2026-10-24
