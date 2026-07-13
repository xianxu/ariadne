---
id: 000167
status: working
deps: []
github_issue:
created: 2026-07-07
updated: 2026-07-12
estimate_hours:
started: 2026-07-12T22:02:15-07:00
---

# institutionalize using continuation for long running session

## Problem

Ariadne has an agent-neutral `continuation` datatype for preserving the human-
meaningful state of a long-running session, and Pair's continuation writer can
durably write it and restart the session from it. Agents currently initiate that
flow only after an explicit operator request. They are not told to use it
proactively as context fills, so long sessions can reach compaction with less
room to create a good handoff. In metis-v2, an agent learned this behavior after
the operator demonstrated it once; the behavior should be institutionalized for
every Ariadne consumer instead of relearned per project.

## Spec

Add the proactive context-pressure policy to `AGENTS.base.md`, the always-loaded,
agent-neutral constitution exported to every harness entry file.

- When the harness reports that the active context is more than 60% full, the
  agent proactively checkpoints the session before starting another substantial
  unit of work. If an exact percentage is unavailable, a harness context-pressure
  or compaction warning is the equivalent trigger.
- The checkpoint completes the current atomic action, updates relevant durable
  work records, and applies the canonical `continuation` datatype. The
  constitution owns **when** to checkpoint; `construct/datatype/continuation.md`
  remains the single source for **what** to preserve and **how** to finalize it
  (`ARCH-DRY`).
- The agent invokes the available continuation writer and does not separately
  perform a restart. In Pair, the existing writer owns write, commit, push, and
  automatic restart from the continuation. The policy must not duplicate that
  mechanism or add Pair-specific implementation to Ariadne (`ARCH-PURE`).
- Keep this policy in `AGENTS.base.md`, not the continuation prototype or a
  harness-specific entry file, so every derived repository and supported agent
  sees the trigger before it needs to discover the datatype (`ARCH-PURPOSE`).
- Update the workflow atlas to map the proactive trigger without duplicating the
  datatype's section-level procedure.
- Add a regression guard that pins the constitution's threshold, route to the
  continuation datatype, and writer-owned restart boundary.

Out of scope: context-meter/model-window catalogs, automatic threshold polling
inside Pair, changes to the continuation document shape, and new restart
machinery.

## Done when

- `AGENTS.base.md` tells agents to proactively create a continuation above 60%
  context utilization, with a warning-based fallback when percentage is unknown.
- The rule routes to the canonical continuation datatype and leaves restart to
  the writer.
- The rule propagates through the existing base-layer prose composition rather
  than being copied into individual harness files.
- Automated tests guard the policy boundary, and the atlas describes the
  resulting lifecycle.

## Plan

- [ ] Add the session-continuity policy to `AGENTS.base.md`.
- [ ] Add a focused regression guard for the policy contract.
- [ ] Update the workflow atlas and verify base-layer composition.

## Log

### 2026-07-07

### 2026-07-12

Claimed the issue and inspected the continuation datatype plus Pair's existing
writer/restart path. Design decision: `AGENTS.base.md` owns the proactive trigger;
the datatype continues to own the handoff procedure, and Pair's writer continues
to own restart. This avoids duplicating either contract (`ARCH-DRY`, `ARCH-PURE`)
while delivering the agent-neutral behavior the issue asks for (`ARCH-PURPOSE`).
