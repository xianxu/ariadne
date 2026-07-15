---
id: 000181
status: open
deps: []
github_issue:
created: 2026-07-15
updated: 2026-07-15
estimate_hours:
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

## Done when

- New archives land in `workshop/history/{issues,plans}/`; the 258 existing
  files are migrated (git mv, one commit) or the transition glob covers
  both — no resolution breakage either way (`sdlc resolve` finds archived
  families; the resolve test suite covers the new layout).
- No consumer hardcodes the layout: the vocab model is the single source,
  with a guard test (the #163 pattern).
- `workshop/history/projects/` is a one-line addition when #180 lands.

## Plan

- [ ] design at start-plan: discovery-model shape (per-kind archive dirs vs
      root+convention), sidecar placement, migrate-vs-transition-glob,
      downstream rollout

## Log

### 2026-07-15

Filed from the #180 brainstorm (operator): history should have subfolders —
issues/, plans/, later projects/. Current state: 258 flat files; archive
path single-sourced in issue.cue discovery (good — the change is a model
edit + consumer derivation, not a path hunt).
