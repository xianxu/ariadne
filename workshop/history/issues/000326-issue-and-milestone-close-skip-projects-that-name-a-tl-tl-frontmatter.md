---
id: 000326
status: done
deps: []
github_issue:
created: 2026-10-10
updated: 2026-10-11
estimate_hours:
card_mirror: '31e1ba49b1b34e2007964a10cbb95767b19fe6d0' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-10T23:14:00-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "a83ae72a", done: "4eebc762"}
actual_hours: 0.93
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

Design: one filter at the single fleet-discovery loop in `close.go` (`buildCloseResult`'s project section). Both close and
milestone-close run through it, and everything downstream — tick, detail-block upsert and its `--no-project` gate, the
peer-write, and the review snapshot (`captureCloseReviewSnapshot` reads `projectEdits`) — consumes that loop's output.
So a TL-owned match `continue`s before any of it (ARCH-DRY: one decision point, not four).

- [x] `project.TLOwners(text) ([]string, error)` in `internal/project/metadata.go`: `tl:` as a string or a list of strings;
  no frontmatter or no/empty `tl:` → not owned; any other shape → error. Decodes only the `tl` key (see Revisions).
  Unit tests.
- [x] Close loop: TL-owned → `cinfo` skip line naming project + TL(s), `continue`. Unreadable `tl:` → warn and skip (fail
  closed: don't write a file whose ownership can't be read).
- [x] Tests: close + milestone-close leave a `tl:` project byte-identical and print the skip line (milestone-close also
  must not refuse on the missing detail block); an edit to a `tl:` project during the review doesn't void SHIP.
- [x] Docs: `construct/vocabulary/project.cue` (`tl?`), `construct/datatype/project.md` table, `sdlc close --help` /
  `milestone-close --help` mention the skip; atlas line if the close flow is mapped there.
- [ ] Close.

Overlap: ariadne#308 reworks the peer project tick (writes to the peer's origin instead of its checkout). This issue
only decides *which* projects a close touches; #308 keeps *how* the remaining ones are written. Not absorbed.

## Log


- 2026-10-11: closed — TestTLOwners (incl. unrelated-malformed-field row, BR-1) + TestCloseCommands_TLOwnedProjectUntouched (close & milestone-close, with/without a TL edit mid-review: tl: file byte-identical, skip line, SHIP finalizes, sibling non-tl project still ticked; filter mutation fails it). make test: all green except TestCancellationKillsDescendants (sandbox blocks /bin/ps; passes unsandboxed) and TestPlanningReviewConcurrencySchedules (load timing flake; passes alone). vocab vet green. --no-plan-check: the only unticked row is this close itself.; review verdict: SHIP
### 2026-10-10
- Implemented. `TLOwners` tells "no frontmatter" apart via a new `frontmatter.ErrNoFrontmatter` sentinel instead of
  string-matching the error. Unit test `TestTLOwners` (owned / not-owned / malformed rows, each asserting its own error).
- `TestCloseCommands_TLOwnedProjectUntouched` (close + milestone-close × with/without a TL edit during the review): the
  `tl:` file stays byte-identical, the skip line prints, SHIP finalizes, and a sibling project without `tl:` is still
  ticked. It passes no `--no-project` (the plain project carries its M1 detail block), so it proves the TL project
  doesn't trip that gate. Mutation (filter disabled) fails it.
- `project.cue`'s `#Project` is a definition, so the exported `pkg/vocab/project.json` doesn't change (`go generate` → no
  diff); vet_test.sh green.
- Close review round 1, FIX-THEN-SHIP. BR-1 (Important): `TLOwners` decoded the full typed `Metadata`, so a malformed
  unrelated field (`deps:` as a scalar) made a project without `tl:` unreadable and skipped it, which broke "projects
  without `tl:` behave as today". Fixed: decode only the `tl` key; `Metadata.TL` dropped (the Minor). New not-owned row
  `unrelated field malformed` fails against the old decode. CRLF frontmatter (Minor) is pre-existing in every
  frontmatter reader and stays out of scope.

## Revisions

- 2026-10-11: plan row 1 originally said `Metadata` gains `TL any`; the BR-1 fix replaced that with a decode of only the
  `tl` key, so a malformed unrelated field can't skip a project without `tl:`.
