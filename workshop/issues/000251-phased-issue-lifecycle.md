---
id: 000251
status: open
deps: []
github_issue:
created: 2026-09-24
updated: 2026-09-24
estimate_hours:
---

# Phased issue lifecycle: spec, plan and code phases land on main

## Problem

Issue files keep getting stranded or diverging, because `sdlc` publishes
*copies* of commits and never brings the copy back into the branch the
original came from. Evidence:

- **Resting branch diverges** (#249): `issue sync` makes local-only commits on
  `main-slotN`; any peer publish then leaves the branch ahead and behind.
- **Published copies strand, and drop content** (#249, re-diagnosed): `change-code`
  publishes one selected commit's diff (`PublishCommit`: `merge-tree` +
  `commit-tree` onto `origin/main`) and leaves the local branch alone. In pair,
  `15f5220f` (1 line) published as `6bc2340e`. The earlier syncs `24199071`
  (+24) and `6aa7d819` (+41/−7) never reached origin, and none of them is an
  ancestor of the #311 close commit. #311's design text was lost from `main` and
  had to be re-added on the feature branch (`dc94d4be`). They weren't duplicates.
- **Wrong-branch commits** (#240, #220): issue updates made while on a feature
  branch are committed there. Seen again on 2026-09-24 in ariadne slot 1:
  `issue new` for #251 reserved on origin correctly, *and* committed a local copy
  on top of #250's closed feature branch. That moved HEAD past #250's reviewed
  head, which would have made #250's `sdlc merge` refuse. It was removed by hand.
  The same slot's `main-slot1` also shows the #249 pattern (`ff972bf`,
  `d2640b3`).

Underneath: the issue system is a design artifact with its own history (specs,
plans, cross-referencing docs), but the lifecycle only models the code half.
`open → working → codecomplete → done` puts everything from claim to close in
one `working` state. Design never lands on `main` in its own right. It reaches
`main` only as a side effect of `change-code`, one copied commit at a time.

## Spec

### Principle

- **Design artifacts (`workshop/`) land on `main` at defined points. Code
  reaches `main` through a PR.** Atlas documents the product as built, so it
  lands with the code.
- **Every local commit either reaches `main` by its own SHA or is explicitly
  abandoned.** This can be checked exactly: `git log origin/main..HEAD` lists
  only commits that are on a defined path to `main`.
- **Integrate by merging and fix conflicts forward.** Never rewrite commits that
  have been shared.

### Phases

An issue moves through **phases**. Every phase has the same state machine:

```
         claim(phase)            land + verdict
open ──────────────► working ──────────────► complete
                       │  ▲
          land, no     │  │ claim(phase)
          verdict      ▼  │
                     parked
                                   (or: skipped)
```

| Phase | Output | `complete` gate |
|---|---|---|
| spec | `## Spec` / PRD, `## Done when` | prd gate (#237: hard structure + soft judge) |
| plan | `## Plan`, durable plan, estimate | plan-quality judge (today's `change-code`) |
| code | code + tests | boundary review → PR → merge (today's `codecomplete` → `done`) |

- **One structure, whatever the issue's size.** Filing a bug is `issue new`,
  which starts with every phase `open`. Triage takes it to spec-complete, design
  to plan-complete (the point where a reasonable estimate exists), then
  implementation. Small issues move through quickly. Keeping the same shape
  gives measured time per phase, including design time, and bounds long-running
  phases where the operator's attention has moved elsewhere.
- **The lock is per phase.** `claim --phase X` records the holder: a PM on spec,
  a TL on plan, an implementer on code. The status is a tuple `(phase, status)`
  in `construct/vocabulary/issue.cue`.
- **Phases may overlap.** Development is iterative. A plan can start while the
  spec is still moving, and a spec change can partly invalidate a plan (a 10%
  spec change should cost roughly 10% of the plan). Only the *gates* order the
  phases (see below).
- **`parked`** lands the current state on `main` with no verdict and releases
  the phase lock. Nothing is stranded, anyone can pick the phase up, and the
  phase's active-time window ends at the park.
- **`complete` is backed by a verdict trailer whose hash matches the content it
  approved** (following the #237 prd-ready design). An edit after completion
  makes the verdict stale. Once code has started, spec and plan revisions go in
  `## Revisions` and ride the code PR.
- **A sufficiency judge replaces "optional".** Phases aren't simply skippable.
  A base-layer judge decides whether the spec/plan is *rough enough to proceed*
  for this issue's size. "Change that button's color" can go from a sketch to
  ship in one pass. A feature needs a real spec and plan. A phase the judge lets
  through becomes `skipped`, with the reason recorded. This subsumes the #231
  quick/full flow inference and #237's `--no-prd`.

### Landing mechanics

- **Each phase claim cuts a short-lived branch from `origin/main`**, and the
  branch ends when the phase lands. Handoff between phases, and between people,
  goes through `main`.
- **Branching early also signals that the slot is in use.** A slot sitting on its
  resting branch looks free to other agents, which may take it over. On
  2026-09-24 this conversation was designing in slot 1 on `main-slot1`, and
  another session took the slot for #250 in the meantime. Design work in
  progress has to be on a branch so the slot shows as occupied.
- **Spec and plan branches touch only `workshop/`.** Landing is: fetch, merge
  `origin/main` into the branch, then push the branch head to `main` as a
  fast-forward. The commits keep their SHAs. Nothing is copied and there is
  nothing to merge back.
- **The code branch is today's PR branch.** Design revisions made during code
  ride the PR.
- **Snapshot + merge-back** is only for docs that must land from a code branch
  (the #240 case). Steps:
  1. 3-way merge the branch's `workshop/` changes onto `origin/main` as commit `D`.
  2. Push `D`.
  3. Merge `origin/main` into the branch at once, so `D` is in both histories.

  This replaces today's publish-a-copy with no merge-back, which is the root
  cause above.
- **A landing carries the landing issue's own files.** Stray `workshop/` edits
  to other issues are tolerated with a warning, not refused.
- **Resting branches stay clean.** They only ever fast-forward to `origin/main`.

### Out of scope here

- Renaming `issue sync` to `checkpoint`, plus off-machine branch backup. It
  doesn't change the key flow, so it is deferred.

### Open questions

1. With overlapping phases, is the status a single `(phase, status)` tuple for
   the lead phase, or one status per phase (spec: working, plan: working)?
2. How is `issue.cue` structured (a phase list plus a shared state machine), and
   is the issue-level status derived from it?
3. What exactly are the sufficiency judge's inputs and verdict, and how does it
   relate to #237's prd gate and #231's flow inference?
4. Migration: today's `working` corresponds to working-on-plan, or to
   working-on-code for simple issues. Existing issues start without historical
   baggage (this issue replaces rather than extends #249's design).
5. Base-layer impact on downstream repos. This changes the SDLC contract every
   ariadne-styled repo follows.

## Done when

- The design is settled: open questions answered, phase model written into
  `issue.cue`, and a project file splits the work into slices (landing
  mechanics, phase lock/status model, sufficiency judge, migration).
- Landing any phase leaves no local-only commit behind. Test fixture: three
  design checkpoints, then plan-complete → `origin/main..HEAD` is empty on the
  phase branch, **and `origin/main` contains the content of all three**.
- A peer publish between two checkpoints does not strand or lose anything
  (two-clone fixture).
- Filing or editing a *different* issue from a code branch never leaves a
  commit on that branch that blocks its reviewed-head guard (the slot 1
  #250/#251 case above).
- Relation to #249, #240, #220, #222 and #237 is written down: which are
  superseded, which become slices, and which stay separate.

## Plan

- [ ]

## Log

### 2026-09-24

Designed in conversation from ariadne slot 1 while reviewing #249. Key steps:
- #249's "case 2 duplicates" were actually lost content (pair `6bc2340e`
  carries 1 line; the earlier syncs never reached origin).
- The root cause is copy-without-merge-back.
- Design is its own history and deserves phases that mirror code-complete.
- Every phase has the same shape (claim/working/parked/complete), with
  per-phase locks for a future multi-role team (PM/TL/implementer).
- Short-lived phase branches make design landings plain fast-forwards of real
  commits.

Pensive consulted: `brain/workshop/pensive/2026-09-18-01-pensive-product-lens-user-model.md`
§"How prd fits with sdlc" (prd-ready as a derived verdict, not a status;
enforced at start-plan; `--no-prd` escape).

Filed from slot 1 while it was on #250's feature branch; the local copy of
this issue's creation commit was removed from that branch
(`git reset --keep fa5396e`) and the spec written in slot 2.

## Revisions

- 2026-09-25: storage and landing design (Problem, own-SHA landing,
  snapshot + merge-back, clean resting branches) superseded by #252 (issue
  cards on a tracker ref, details on the branch). The phased lifecycle half
  (phases, per-phase claims, `parked`, verdict-backed `complete`, sufficiency
  judge) is not covered by #252 and remains open for the operator to keep,
  narrow or drop.
