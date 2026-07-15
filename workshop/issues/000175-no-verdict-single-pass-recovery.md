---
id: 000175
status: codecomplete
deps: []
github_issue:
created: 2026-07-14
updated: 2026-07-14
estimate_hours: 1.04
started: 2026-07-14T18:42:44-07:00
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
- [x] `partitionMissingVerdicts` (pure trailing-vs-midstream split) + table test
- [x] `findMilestonesMissingVerdict` returns plan order; update 5 call sites
- [x] message formatters: §3 citation + fold recovery in `formatMissingVerdicts`; new `formatTrailingVerdictAccepted` / `formatTrailingNeedsJudge` (gatesig closing line preserved)
- [x] rewire verdict-gate block in `computeClose` + 3 integration tests (accept trailing / refuse midstream / refuse trailing+--no-judge)
- [x] plan-quality prompt: over-split failure-mode bullet + prompt-content pin
- [x] atlas (`atlas/workflow/sdlc-binary.md` close-gates prose) + bookkeeping

## Log

### 2026-07-14
- 2026-07-14: closed — go test ./cmd/sdlc/... 11/11 pkgs green incl. 3 new gate integration tests (accept trailing / refuse midstream / refuse trailing+--no-judge) + mutation-checked partition; gatesig pins hold (processmanual green); plan-quality golden regenerated; review verdict: SHIP

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

Implementation: all six plan rows landed. New refusal/acceptance messages
share `verdictNextActionLines` + `verdictBypassClosingLine` (ARCH-DRY —
gatesig.go pins the closing line for #172 friction attribution; the new
acceptance info line is pinned non-colliding via assertNoGatesigCollision,
#177 precedent). Mutation check: an everything-midstream mutant fails both
the partition table test and the acceptance integration test. plan-quality
golden regenerated (byte-pin, intentional prompt change). Shadow-doc sweep:
helptext/close.md gate description + atlas/workflow/sdlc-binary.md updated;
AGENTS.base.md bypass prose still accurate. Done-when line 2 (via-bypass
drop from 4/8) is a lagging metric — measurable only as future closes
accrue; `TestClose_TrailingUnclosedMilestones_AcceptedByCloseReview` is the
leading proof that the friction case now passes without `--no-verdict`.
