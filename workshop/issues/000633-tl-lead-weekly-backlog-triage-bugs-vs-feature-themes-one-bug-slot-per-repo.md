---
id: 000633
status: open
deps: []
github_issue:
created: 2026-10-25
updated: 2026-10-25
estimate_hours:
card_mirror: '192d5c893ccebc8284d7263c2cf7ed0177787930' # card fields mirrored from issue-cards; edit via sdlc
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

- [ ]

## Log

### 2026-10-25
