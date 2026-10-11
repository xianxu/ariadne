---
id: 000326
status: open
deps: []
github_issue:
created: 2026-10-10
updated: 2026-10-10
estimate_hours:
card_mirror: 'e5ea94db9005806284ee27e0d95c0e2b9d50792b' # card fields mirrored from issue-cards; edit via sdlc
---

# Issue and milestone close skip projects that name a TL (tl: frontmatter)

## Problem

`sdlc close` and `sdlc milestone-close` find every project across the fleet that references the issue, tick its
task row, upsert `**actual:**`/`**closed:**` in its detail block, and snapshot the project file as review input.
Once a TL drives a project (ops' ariadne-robustness-1, 10-10), that causes three failures:

- **Stray ticks in the TL's checkout.** The close edits the project file in ops, which isn't the worker's repo
  and is often dirty, so the edit can't be committed and is left behind for the TL to sort out (ops kink 38;
  pair#324's close did it again at 19:1x on 10-10).
- **Ticks before the merge.** A close is codecomplete, not landed; the project row reads done while the PR is
  still open or red (ariadne#320, ariadne#189).
- **Voided reviews.** The TL edits the project file while a worker's close review runs; the snapshot check then
  refuses to finalize a SHIP verdict (ariadne#183, ops kink 55).

The operator decided (10-10) that the TL owns `project.md` and workers never write or read it for closing.

## Spec

A project whose frontmatter has a `tl:` field (e.g. `tl: xian-xu@ops:0`, or a list) is **TL-owned**.

- `sdlc close` and `sdlc milestone-close` skip TL-owned projects entirely: no tick, no detail-block upsert, no
  peer-write, and the project file is not part of the review's prose snapshot.
- The close prints one info line per skipped project, naming it and the TL, so the worker knows to report the
  landing to the TL rather than tick anything.
- The project-detail-block gate (`--no-project`) doesn't refuse on account of a TL-owned project.
- Projects without `tl:` behave exactly as today.
- Read paths (`sdlc project find`, `sdlc resolve --kind project`) still find TL-owned projects.
- `construct/datatype/project.md` documents `tl:`; the vocabulary model accepts it.

Out of scope: the rest of the project datatype pivot (lifecycle without a deadline, Trains and Decisions sections,
forecast). That's a separate issue after ops' batch-2 recap; ariadne#308's project-tick rework should settle with
this one which owns what.

## Done when

- [ ] A close and a milestone-close of an issue referenced by a `tl:` project leave that project file byte-identical, and print the skip line (test).
- [ ] An edit to a `tl:` project during a close review does not void the verdict (test).
- [ ] A project without `tl:` is still ticked and snapshotted (existing tests stay green).
- [ ] `construct/datatype/project.md` and the vocabulary document `tl:`; `sdlc close --help` mentions the skip.

## Plan

- [ ]

## Log

### 2026-10-10
