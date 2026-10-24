---
id: 000601
status: open
deps: []
github_issue:
created: 2026-10-24
updated: 2026-10-24
estimate_hours:
card_mirror: '565a4aac645f321403cb3bf0dc7b7ea268eda2ef' # card fields mirrored from issue-cards; edit via sdlc
---

# Notify the owning Couch slot when an sdlc background operation finishes (best effort; ariadne works without Couch)

## Problem

## Spec

**Wake the owning slot when an sdlc background operation finishes (operator request, 10-24).** An agent starts a
suite, resume or merge as a background process and ends its turn; when the process finishes nothing starts a new turn,
so the slot sits "waiting" on a result that already arrived. Overnight 10-23/24 about a dozen results went unnoticed
until the TL's hourly sweep, and two merge runners exited unreported.

sdlc already knows when a run reaches a terminal state and which slot owns it (queue leases and merge runners record
their owner). Proposal: on a terminal state (suite passed or failed, resume done or cancelled, merge landed or exited),
sdlc sends the owning slot a short advisory notice through Couch, e.g. `suite attempt X on head Y: passed; next: sdlc
merge`. The notice is a pointer only: the agent reads the real result from sdlc, so a lost or repeated notice is
harmless.

**Constraint (operator):** ariadne must keep working on its own, outside Couch; only Couch depends on ariadne, never the
reverse. So sdlc *may* call `couch` when it's available, and if Couch isn't there (not installed, not running, no
owning address, the call fails), that's fine: sdlc carries on exactly as today, with no error and no retry loop.

## Done when

- A terminal state of a suite, resume or merge sends one notice to the recorded owning Couch address when Couch is
  available, best effort, with a bounded timeout.
- Without Couch (binary absent, socket absent, no owner address, call fails), every sdlc verb behaves exactly as
  before; tests cover each absence case with no Couch dependency in ariadne's build or tests.
- The notice text names the operation, head, outcome and the next verb.

## Plan

- [ ]

## Log

### 2026-10-24
