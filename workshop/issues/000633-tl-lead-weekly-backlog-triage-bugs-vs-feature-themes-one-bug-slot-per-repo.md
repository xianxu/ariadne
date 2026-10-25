---
id: 000633
status: working
deps: []
github_issue:
created: 2026-10-25
updated: 2026-10-25
estimate_hours: 2.29
card_mirror: '816dd746d24059d1ac01d1f4e45aa8ecb48b1ea8' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-25T12:24:34-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
---

# tl-lead: weekly backlog triage (bugs vs feature themes, one bug slot per repo, closing-velocity insight, Needs-you output)

## Problem

## Spec

**Weekly backlog triage as a TL duty (operator, 10-25).** Today the TL's dispatching is reactive: it files what it
finds and dispatches from memory or a slot's "next". There's no regular look at each repository's open issues, no
split between bugs and features, and no measure of whether bugs are under control. The operator's design:

- **Weekly, per repository, the TL triages every open issue:** **bug** or **feature**; and **ready** (relevant, deps
  met, small enough), **waiting** (on what), **stale** (propose closing, with the reason) or **duplicate**.
- **Capacity:** one bug-fix slot per repository by default. From the TL's measured per-slot velocity (issues closed
  per slot per week, from landing history), pick a fixed number of bugs that slot can close in the week; the rest wait.
- **Features are organized into themes**, and the TL proposes themes to the operator as candidate projects (the
  project datatype), rather than dispatching features one by one.
- **The result goes to the operator in Needs you before anything proceeds**, organized by theme: the week's bug list
  per repo, the close proposals, the feature themes proposed as projects.
- **Insight, not just lists:** track issue closing velocity per repository (bugs opened vs closed per week, open-bug
  count trend), so the TL can answer "is bug volume under control, or are we falling behind and should shift capacity
  from features to fixes?" and present the trade-off; the operator decides by business need.
- **Prompted by a scheduled task** (couch-scheduling, pair#737) once a week; the 21:00 ready-queue check (ops#11)
  counts against the triaged ready list.

## Done when

- `tl-lead --help` states the triage duty, the bug/feature and ready/waiting/stale/duplicate marks, the default of one
  bug slot per repo, velocity-based sizing, themes-as-project proposals, and the Needs-you output shape.
- A read-only helper (e.g. `tl-lead triage stats`) reports per-repo opened/closed per week and open-bug trend from
  issue history, with tests over a fake history.

## Plan

Durable plan: [000633 plan](../plans/000633-tl-lead-weekly-backlog-triage-bugs-vs-feature-themes-one-bug-slot-per-repo-plan.md).

- [ ] M1 — the `kind` card field, `sdlc issue set-kind`, and `sdlc issue list --json` over every card
- [ ] M2 — `tl-lead triage stats` and the WEEKLY BACKLOG TRIAGE duty in help

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: smaller-go-module        design=0.1 impl=0.15
item: smaller-go-module        design=0.2 impl=0.2
item: atlas-docs               design=0.05 impl=0.05
item: milestone-review         design=0.0 impl=0.2
item: greenfield-go-module     design=0.5 impl=0.3
item: atlas-docs               design=0.1 impl=0.1
item: milestone-review         design=0.0 impl=0.2
design-buffer: 0.15
total: 2.29
```

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only.*

Items:
- **M1:**
  - `kind` plus `set-kind` mirrors `severity`, so it is a smaller module with little design.
  - `issue list --json` walks the full tracker record set.
  - The atlas update.
  - The milestone review.
- **M2:**
  - `triage stats` is a greenfield single-concern module: the pure `weekStats` and `slotVelocity`, the seam method and the command, with the generated-input property test.
  - The help section and its order test.
  - The close review.

`impl=` is 40% of the v2 midpoints. Design comes from the v2 ranges with a +15% buffer, because the plan resolves the schema, the week definition and the quota rule.

## Log

### 2026-10-25
- 2026-10-25: plan approved by the TL. kind is a card field with no heuristics; the TL back-fills the last 4 weeks of closed cards once after M1 lands; the unknown-key test runs first.
- 2026-10-25: ARCH-EVOLVE gate (M1 step 1) holds, so the work continues. `TestOlderCardReaderPreservesUnknownFieldsThroughSetters` (`cmd/sdlc/internal/issue/cardunknown_test.go`, #460) already proves that a top-level unknown key survives every card setter, and it passes. Pair pins ariadne 5a17f789 (10-24), which contains both #460 commits (429930268, f6124b5ef), so pair's sdlc keeps `kind` through any rewrite. No new test is needed.
