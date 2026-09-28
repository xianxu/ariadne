---
id: '000039'
status: done
created: 2026-05-27
updated: 2026-05-27
estimate_hours: 8
actual_hours: 4
---

# Defer worktree decision to implementation-start time

## Problem

`sdlc start --issue N` today does two distinct things in one shot:
1. Commits + pushes the issue file (workstream-claim primitive).
2. Creates a worktree + branch from HEAD.

For a single operator with a single-threaded mind, the second step
slows down development and is often the wrong default — small fixes,
docs touch-ups, and early planning work all happen better on main with
no worktree overhead. The decision *whether* to worktree is also poorly
timed: at `sdlc start` you don't yet know the change's complexity,
blast radius, or test surface. Those signals become legible only after
the plan is written.

Beyond the worktree question, the planning → implementation transition
is also currently un-gated: there's no checkpoint that enforces a
filled-in Spec, a non-empty Plan, or judges whether the plan is
actually executable. `sdlc judge plan` runs only at merge time (to
catch unchecked-done items), never as a pre-implementation gate.
