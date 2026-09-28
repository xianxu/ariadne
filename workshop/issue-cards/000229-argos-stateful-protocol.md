---
id: 000229
status: open
created: 2026-09-15
updated: 2026-09-15
estimate_hours:
github_issue:
---

# Design Argos stateful protocol binary

## Problem

The current `sdlc` binary manages repository lifecycle checkpoints, but it does
not manage a long-running backlog execution session. Small fixes need a higher
level process that can select and group suitable issues, assign them to a stable
worktree/Couch slot, advance one task at a time, recover after context loss, and
produce a durable session report.

The model must remain grounded in durable state rather than asking the LLM to
remember which task is current or infer whether a transition is legal. This is a
different layer from repository execution: Argos coordinates the session, while
the admitted task enters the normal `sdlc quick` flow. Argos must not reimplement
issue status, claims, verification, or close semantics.
