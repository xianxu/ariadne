---
id: 000633
status: working
deps: []
github_issue:
created: 2026-10-25
updated: 2026-10-25
estimate_hours:
card_mirror: '4f2c50d381f8c0ded482318ff79c77f67363d9c1' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-25T12:24:34-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
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

## Log

### 2026-10-25
