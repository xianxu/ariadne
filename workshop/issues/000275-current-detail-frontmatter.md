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

- [ ] Close: a follow-up commit mirrors the codecomplete card onto the issue branch. It cannot go into the evidence commit, because the card embeds that commit's SHA.
- [ ] Checkout archives (merge, push, interrupted recovery): refresh the moved history file from the done card before staging it.
- [ ] Slot-landing archive: a deterministic projection of the done card. The retry proof pins the card through the archived file's `card_mirror`.
- [ ] Documentation: the refresh points and the remaining stale copies, including main's active copy before landing.

## Log

### 2026-09-30

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
