---
id: '000175'
status: done
started: 2026-07-14T18:42:44-07:00
created: 2026-07-14
updated: 2026-07-14
estimate_hours: 1.04
actual_hours: 0.65
---

# no-verdict gate: accept the issue-close review for trailing unclosed milestones (single-pass Mx recovery)

## Problem

#172's friction audit: `no-verdict` has the highest claude route-around rate —
4 of its 8 refusals resolved VIA BYPASS. Cause: a plan whose `## Plan` rows are
tagged `Mx` but whose work landed in ONE pass (no per-milestone
`milestone-close`) cannot satisfy the per-milestone Review-Verdict-trailer
demand retroactively — the issue-close boundary review covers the whole window,
but the gate still refuses, so agents pass `--no-verdict`. AGENTS.md §3 now
warns against over-splitting atomic work into Mx rows, but the gate punishes
the already-recovered case.
