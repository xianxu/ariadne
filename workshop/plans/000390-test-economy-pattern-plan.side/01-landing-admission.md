---
type: side-doc
parent: ../000390-test-economy-pattern-plan.md
process-instruction: "[free-form] answer the operator's 🤖 comments in place"
---

# Landing admission: what it is, and is it a new "landing" workflow in sdlc?

_Side doc of [the plan](../000390-test-economy-pattern-plan.md) (Architecture line, and §2's state machine). Two of
your questions: "tell me more about this admission system" and "so you're modeling an additional workflow in sdlc,
called landing?"_

## What happens today

`sdlc merge --yes` already does the landing work: it integrates main into the branch, runs the declared full suite on
that exact head, checks CI and review, and merges the PR. The problem is that **every ready branch does this at the same
time**: each tests itself against today's main, and the moment one lands, every other in-flight run is stale and has to
integrate and rerun. Last night #376 did this three times (and hit merge's 3-retry cap); #528 several times.

## What admission changes

**One landing at a time per repo, in order.** Nothing new for agents to call: `sdlc merge --yes` stays the verb.
- When a branch is ready (reviewed, CI green), its `sdlc merge` **enqueues** it and waits its turn instead of starting
  its integration at once.
- The **head of the queue** integrates main, runs the full suite on that exact head, and lands. No one else is testing
  against main meanwhile, so its result can't go stale (except for an emergency direct push, which marks it `stale`
  and it re-integrates once).
- Then the next one starts from the new main.
- The **priority lane** (fleet-blocking fixes, granted by you or me) jumps waiting candidates but never interrupts a
  running suite.

## So, yes: a modeled "landing" lifecycle, inside the existing verb

The steps already exist inside `sdlc merge`; what's new is making them **explicit, durable state** so a crash, a lost
response or a restart resumes exactly where it was (`waiting → integrating → testing → verified → landing → landed`,
plus `stale`, `failed`, `uncertain`). Today those steps live only in one process's memory, which is why a dead merge
process leaves things unclear. Agents keep calling `sdlc merge`; `sdlc landing status` (or similar) shows the queue.

**No daemon:** whichever `sdlc merge` process is at the head of the queue does the work; a waiting one just watches the
queue. If the head process dies, the next merge (or a retry) finds the durable state and resumes it.

## Why serial first, batching later

Batching (integrate several ready branches together, one suite for all) saves more, but when that suite fails you have
to bisect which branch broke it. Serial is simpler and removes most reruns already. The pilot's stop rule checks that it
actually helps: **full suites per merged PR** must fall and **p90 ready-to-landed delay** must not rise over seven days.

## Open for you

- Queue scope: per repo (proposed). One machine-wide queue would also serialize pair vs ariadne landings, which don't
  conflict; per repo is enough.
  **Decided (operator): per-repo queue.** Cross-repo dependencies between tasks are the TL's job to sequence before the
  tasks start (as with #556 → #544 → #571 today), not the queue's.
- Queue state location: local, outside git (like the other runtime state; see the plan's retention note), since all
  slots run on this machine today. A cross-machine queue would need origin-hosted state later.
  **Decided (operator): in Couch's storage directory** (`~/.local/share/pair/…`, next to its other runtime state):
  Couch is ariadne's runtime environment. One constraint keeps that crash-only: sdlc reads and writes the queue
  **files** directly (with the same atomic-write and lock discipline), so landing still works while Couch itself is
  down or restarting; Couch only displays the queue. And when sdlc runs outside a Couch setup (a plain shell), it
  falls back to the repo's git common dir.
