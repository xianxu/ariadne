---
id: 000175
status: working
deps: []
github_issue:
created: 2026-07-14
updated: 2026-07-14
estimate_hours: 1.04
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

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: smaller-go-module    design=0.2  impl=0.15
item: smaller-go-module    design=0.1  impl=0.2
item: atlas-docs           design=0.1  impl=0.08
item: milestone-review     design=0.0  impl=0.15
design-buffer: 0.15
total: 1.04
```

Σdesign 0.4 × 1.15 + Σimpl 0.58 × 1.0 = 1.04. First smaller-go-module = pure
core + formatters (partition fn, signature widen, message contracts); second =
gate rewire + 3 integration tests on existing seams; atlas-docs = plan-quality
prompt bullet + atlas close-gates prose; milestone-review = the close-time
boundary review. +15% buffer: thorough plan doc exists. *Produced via
`brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against
`baseline-v3.1.md`. Method A only.*

## Plan

Durable design: `workshop/plans/000175-no-verdict-single-pass-recovery-plan.md`.
Single-pass work — plain checkboxes (§3), one close boundary.

- [x] design the acceptance rule (issue-close review covers trailing unclosed milestones) + refusal-text recovery hint
- [ ] `partitionMissingVerdicts` (pure trailing-vs-midstream split) + table test
- [ ] `findMilestonesMissingVerdict` returns plan order; update 5 call sites
- [ ] message formatters: §3 citation + fold recovery in `formatMissingVerdicts`; new `formatTrailingVerdictAccepted` / `formatTrailingNeedsJudge` (gatesig closing line preserved)
- [ ] rewire verdict-gate block in `computeClose` + 3 integration tests (accept trailing / refuse midstream / refuse trailing+--no-judge)
- [ ] plan-quality prompt: over-split failure-mode bullet + prompt-content pin
- [ ] atlas (`atlas/workflow/sdlc-binary.md` close-gates prose) + bookkeeping

## Log

### 2026-07-14

Filed from #172 M4 (T2 no-verdict triage).

Design (claim → start-plan → plan reviewed fresh-eyes): the whole-issue close
review window is branch-point→HEAD (`boundaryWindowBase` with milestone "") —
it covers ALL branch commits, so accepting *trailing* unclosed milestones is
real coverage, while a *mid-stream* miss (a later milestone closed WITH review)
genuinely crossed a boundary unreviewed and stays refused. Acceptance requires
the close review to actually run: `--no-judge` voids the premise → refuse.
Partition is a pure function (ARCH-PURE); all three Spec candidates ship
(ARCH-PURPOSE). Plan reviewer confirmed gate reachability via the
closeRepo/stubJudge/expectDie seams and caught one sketch bug (close writes
`codecomplete`, not `done`, #160).
