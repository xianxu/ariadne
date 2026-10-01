---
id: 000275
status: working
deps: []
github_issue:
created: 2026-09-30
updated: 2026-09-30
estimate_hours:
card_mirror: 'bc23fdac81d2e2836676a16475cd62ffb233e336' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-09-30T15:23:01-07:00
flow: {kind: full, provenance: inferred}
---

# Keep issue detail frontmatter current

## Problem

The issue-tracker branch is authoritative, but people also read issue detail
Markdown directly and through editors such as Parley. Stale mirrored
frontmatter makes an issue look open or working after it has closed, creating
confusion even when the tracker is correct.

Observed with pair#358: after pulling main in pair:2, the active detail file
still said `open` while the card said `codecomplete`. After merge, the archived
detail file on remote main said `working` while the card said `done`.

## Spec

Keep issue detail headers, especially mirrored frontmatter, as current as
practical while retaining issue-tracker as the sole authority for card fields.
Refresh copies from the card; never infer or write authoritative card state
from a stale detail header.

Audit lifecycle and publication boundaries for refresh opportunities, including
claim/start-plan, card setters, close, merge/push, archive, and recovery. Close
and archive currently leave stale mirrors; the final archived copy should
reflect the finalized card, including status and measured hours. Define how
active detail copies become current without silently dirtying unrelated
worktrees or overwriting branch-owned content. Document when a local snapshot
can still lag; plain `git pull` is not itself an SDLC refresh operation.

This is broader than changing Parley's display: directly reading an issue or
its archive should not unnecessarily contradict its authoritative card.

## Done when

- Issue-tracker remains the sole authority; frontmatter refresh is a one-way
  projection that preserves detail bodies and branch-owned fields.
- Safe refresh points are defined and implemented across the lifecycle,
  including successful close and final archive publication.
- A closed-and-merged issue's archived frontmatter reflects its finalized card
  (`done`, current dates and actual hours where present), with a current mirror
  baseline.
- Stateful regression tests reproduce pair#358's stale close/archive headers
  and cover retry/recovery without overwriting unrelated edits or newer cards.
- Documentation explains refresh timing and any remaining stale-copy cases.

## Plan

Durable plan: `workshop/plans/000275-current-detail-frontmatter-plan.md`.
All refreshes reuse `issue.RefreshMirror`, which is one-way and preserves the body.

- [x] Close: a follow-up commit mirrors the codecomplete card onto the issue branch. It cannot go into the evidence commit, because the card embeds that commit's SHA.
- [x] Checkout archives (merge, push, interrupted recovery): refresh the moved history file from the done card before staging it.
- [x] Slot-landing archive: a deterministic projection of the done card. The retry proof pins the card through the archived file's `card_mirror`.
- [x] Documentation: the refresh points and the remaining stale copies, including main's active copy before landing.

## Log

### 2026-09-30
- 2026-09-30: closed — New real-git tests: close mirrors the codecomplete card (clean, dirty, staged; reconcile retries after a crash); merge/push/recovery archives mirror the done card; hand-edited mirror or unreadable baseline archived unchanged; landing projection (pure table) plus a proof that survives a later card change (mutation-checked); TestTrackerFullSlotCycle asserts archived details on main say done (pair#358 repro). make test: all cmd/sdlc shards green; processgroup ps failure is a sandbox artifact in an untouched package.; review verdict: SHIP
- 2026-09-30: flow upgraded quick → full — 260 added lines in code files (limit 100); an earlier round of this close already ran the full review

- Filed at the operator's request after closing pair#358. The user accepts
  issue-tracker authority and wants detail frontmatter kept as up to date as
  possible to reduce confusion. ARCH-DRY: derive mirrors from the existing
  authority rather than creating a second status owner.
- Design: the #252 M3 commit 0c9ad8ef deliberately dropped the archive refresh
  so that the landing archive proof would not depend on the live card. This
  design restores the refresh without losing that property: the archived
  file's `card_mirror` names the done card it was projected from, and the
  proof re-derives from that pinned blob.
- Main's active copy is never refreshed, because the rest branch is never
  edited and a main-side commit would conflict with the branch's frontmatter.
  It stays documented as a stale case until landing.
- Implemented in four commits: close mirror, checkout archive refresh, deterministic
  landing archive, docs. Close tests now locate the evidence commit at `HEAD^`
  (`evidenceRev`), because the mirror commit sits on top of it.
- Pair#358 repro: `TestTrackerFullSlotCycle` asserts that the archived details on
  main mirror the done card. The pin test was mutation-checked: with
  `pinArchivedCard` disabled, the proof fails with "archive generation differs".
- `make test`: every cmd/sdlc shard passes.
  `internal/processgroup` `TestCancellationKillsDescendants` fails in the sandbox
  (`fork/exec /bin/ps: operation not permitted`). That is environmental; the
  package is untouched.
  `TestClose_MilestoneRefusesWithRedirect` failed intermittently in one shard,
  passes in isolation, and passed on the next full run.
- Close round 1: the reviewer's network was blocked by the sandbox, so there
  was no verdict. Round 2 returned FIX-THEN-SHIP, but the reviewer had
  detached this checkout's HEAD (#204), so close refused it as stale. Fixed
  the findings it reported:
  - An unreadable mirror baseline no longer fails the landing. The details
    are archived unchanged and the proof expects that.
  - The baseline and the pin comparison now come from main's copy, which the
    planner projects.
  - The tracker is opened once per archive run.
  - The merge path has a test.
  - The atlas no longer over-claims what a reopen does, and documents the
    crash window between codecomplete and the mirror commit.
- Close round 3 returned FIX-THEN-SHIP with one Important finding (BR-1, plan
  table drift on planner purity) and three Minors. Fixes:
  - BR-1: recorded in the plan's `## Revisions` and table, because the reader
    is an injected seam.
  - Minors: fixed the merge help reflow; added `retryCloseMirror` in reconcile
    for the crash window; added a test that the staged-index branch keeps a
    staged edit.
