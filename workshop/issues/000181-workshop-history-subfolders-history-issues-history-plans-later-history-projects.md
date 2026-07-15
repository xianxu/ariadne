---
id: 000181
status: working
deps: []
github_issue:
created: 2026-07-15
updated: 2026-07-15
estimate_hours: 0.94
started: 2026-07-15T16:23:08-07:00
---

# workshop/history subfolders: history/issues, history/plans, later history/projects

## Problem

`workshop/history/` is FLAT: 258 files mixing archived issues, their plan
docs, and review sidecars in one directory. It doesn't scale visually or
navigationally, and #180 is about to add a third artifact kind (archived
projects, if archive-on-done is chosen there). The archive layout is
consumed by real machinery, so restructuring is a code change, not a mkdir:

- the merge/push archive step moves the id-keyed family
  (`issues|plans/ → history/`, #160) as one flat destination;
- `sdlc resolve`'s `familyFiles` globs the archive dir from the vocab model
  (`issue.cue` `discovery: archive: "workshop/history"`) — the single
  source consumers derive from;
- the #163 filename-grammar helper and any history-scanning tooling assume
  the flat layout.

## Spec

Subfolder the archive by artifact kind (operator, 2026-07-15):

- `workshop/history/issues/` — archived issue files
- `workshop/history/plans/` — archived plan docs + review sidecars (they
  are plans-dir residents today; settle at design whether sidecars get
  their own `history/reviews/` or ride with plans)
- `workshop/history/projects/` — later, when #180 settles archive-on-done
  for projects

Mechanics (settle at design time):

- **Model first:** `issue.cue`'s `discovery:` block encodes the new archive
  layout (per-kind archive dirs, or an archive ROOT + kind convention);
  `pkg/vocab` exposes it; consumers derive — no hardcoded paths (the #163
  lesson: turn the sweep into a source guard).
- **Archive step** (merge/push) writes into the per-kind subfolders.
- **Resolution stays correct across the transition:** `familyFiles` globs
  the new subfolders; decide whether it ALSO globs the flat root during a
  transition window, or whether the 258 existing files migrate in one
  `git mv` commit (preserves history; simpler — likely this).
- Downstream repos: the layout ships via the vocab model, so peers pick it
  up on rebuild; their own history dirs need the same one-time migration
  (a `propagate`-adjacent sweep, or lazily on their next archive).

Related: #180 (adds `history/projects/` once project archive-on-done is
decided); #160 (the archive flow being restructured); #163 (filename
grammar single-source — extend, don't fork).

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: smaller-go-module    design=0.15  impl=0.2
item: smaller-go-module    design=0.1   impl=0.15
item: atlas-docs           design=0.05  impl=0.1
item: milestone-review     design=0.0   impl=0.15
design-buffer: 0.15
total: 0.94
```

Σdesign 0.3 × 1.15 + Σimpl 0.6 × 1.0 = 0.94. First
smaller-go-module = ArchiveSubdirs + predicates + resolve reads (Tasks 1–2);
second = the writer sweep across push+merge + updating the ~15-test archive
inventory (Task 3); atlas-docs = migration commit + cue comment + 7-file
helptext sweep + atlas; milestone-review = close-time boundary review.
Design hours are not ×0.2-discounted: the plan was authored in this issue's
active window (claim-early #113), so the estimate carries its authoring.
*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md`
against `baseline-v3.1.md`. Method A only.*

## Done when

- New archives land in `workshop/history/{issues,plans}/`; the 258 existing
  files are migrated (git mv, one commit) or the transition glob covers
  both — no resolution breakage either way (`sdlc resolve` finds archived
  families; the resolve test suite covers the new layout).
- No consumer hardcodes the layout: the vocab model is the single source,
  with a guard test (the #163 pattern).
- `workshop/history/projects/` is a one-line addition when #180 lands.

## Plan

Durable design: `workshop/plans/000181-history-subfolders-plan.md`
(fresh-eyes reviewed; folded: guard-test checkbox, merge.go's duplicate
issue-dest write site + second mutant, plan-family assessDirty case, test
inventory). Decisions: reads tolerant/writes strict; ArchiveSubdirs
Go-owned (cue keeps the root string — downstream JSON compat); sidecars
ride with history/plans/; one-commit git mv migration (159+99).
Single-pass, plain checkboxes.

- [x] design at start-plan: discovery-model shape (per-kind archive dirs vs
      root+convention), sidecar placement, migrate-vs-transition-glob,
      downstream rollout
- [ ] vocab.ArchiveSubdirs + layout-tolerant predicates (isHistoryPath,
      NextID) + guard test
- [ ] resolve reads across both layouts (familyFiles)
- [ ] writers → subfolders (push archiveDoneIssues + merge
      archiveDoneIssuesInDir + archivePlanArtifacts) + test inventory sweep
      + both mutants
- [ ] migrate ariadne's 258 files (git mv, one commit) + docs sweep +
      bookkeeping

## Log

### 2026-07-15

Filed from the #180 brainstorm (operator): history should have subfolders —
issues/, plans/, later projects/. Current state: 258 flat files; archive
path single-sourced in issue.cue discovery (good — the change is a model
edit + consumer derivation, not a path hunt).
