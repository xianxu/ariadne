---
id: 000175
status: working
deps: []
github_issue:
created: 2026-07-14
updated: 2026-07-14
estimate_hours:
started: 2026-07-14T18:42:44-07:00
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

## Spec

Candidates:
- the no-verdict gate accepts the issue-close boundary review as covering
  trailing milestones that never got their own `milestone-close` (the close
  review window IS prev-boundary→HEAD, so coverage is real, not fictional);
- and/or the refusal text cites §3's don't-over-split rule and offers the
  concrete recovery ("fold the unclosed Mx rows into plain checkboxes, or run
  milestone-close per row").
- forward fix: the change-code plan-quality judge flags over-split Mx plans at
  design time (before the trap is set).

## Done when

- Single-pass work with legacy Mx tags closes without `--no-verdict`, OR the
  refusal names the sanctioned recovery.
- Re-measure: no-verdict via-bypass resolutions drop from 4/8.

## Plan

- [ ] design the acceptance rule (issue-close review covers trailing unclosed milestones) + refusal-text recovery hint

## Log

### 2026-07-14

Filed from #172 M4 (T2 no-verdict triage).
