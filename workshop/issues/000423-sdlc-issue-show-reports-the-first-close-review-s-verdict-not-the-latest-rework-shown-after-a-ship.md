---
id: 000423
status: working
deps: []
github_issue:
created: 2026-10-17
updated: 2026-10-17
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'ca4867a1cdb6ac087c0157ee9ee7f88e48e86e4f' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-17T02:22:23-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "66fd2d89", done: "70b81bcc"}
---

# sdlc issue show reports the first close review's verdict, not the latest (REWORK shown after a SHIP)

## Problem

## Spec

`sdlc issue show` reports an issue's review verdict from the first close review, not the latest one. Four slots reported it independently:
- #408;
- ariadne#418 (first FIX-THEN-SHIP, then SHIP);
- pair#617 (round 1 shown after a round 2 SHIP);
- ariadne#416 (first REWORK shown after a round 2 SHIP).

The completion binding, the ledger and the landed status are all correct. Only the readout picks the wrong review.

A reader of `issue show`, whether the TL, the operator or an agent deciding whether to land, sees a stale REWORK or FIX-THEN-SHIP on an issue that shipped.

## Done when

- `issue show`, text and JSON, reports the verdict of the review bound to the issue's completion, or the latest review when there is no completion, along with its round. A test covers an issue with a REWORK round 1 and a SHIP round 2.
- Earlier rounds stay visible as history, labelled as earlier.

## Plan

- [x] Reproduce appended REWORK→SHIP readout with completion selection, round/history, and quoted metadata coverage.
- [x] Select the completion-bound entry (latest without completion), using ledger round numbers when available; preserve legacy sidecars and explicit unknown reads.
- [ ] Verify text/JSON through the collector and archived evidence, run the recorded suite, close and land on SHIP.

## Log

### 2026-10-17

- Root cause: observe.sidecarRows deliberately keeps the first metadata table, but the writer appends Re-review sections. Chosen fix parses only each entry’s metadata before its Review body, reuses markdoc fence filtering (ARCH-DRY), and keeps selection pure (ARCH-PURE). Match completion ReviewedHEAD for close; unmatched bound evidence is unknown rather than silently substituting a later review. No new persisted state or artifacts; ledger timestamps map entries to existing round numbers, with entry order for legacy files.

- RED: TestReviewRoundsSelectCompletionOrLatest returned round-1 REWORK for latest and bound round-2 SHIP, and accepted unbound completion evidence. GREEN: observe package passes, including quoted tables/fenced headings, missing latest verdict, completion selection past later malformed evidence, ledger round numbers and current open findings. Collector/writer targeted tests passed in 53.896s; the real close/archive/branch-deletion lifecycle now includes REWORK then SHIP and preserves both rounds. Golden changes add optional round/history fields; existing schema version remains 1 (additive fields).

- TL-granted :0 live check: staged 016625f7550d30c5708c3e803e31f5dd1b4b85e1, all seven tools verified vcs.revision exact and vcs.modified=false. Native issue show 416 text/JSON reported SHIP round 2 at 2026-10-17T00:40:55-07:00, window d1a4e571..412b0dc4 matching completion; history labelled earlier round 1 REWORK. Cleared :0 to clean main 5fbaf9730a5717a6194544b527e6c573d6a017bc and rebuilt/verified all seven tools there; hub notified it is free. PR #281; recorded suite attempt d521ca0e272e7faa25d9d40f9fc536a7 queued on 016625f7.

- Pre-close edge sweep reproduced and fixed stale-tracker completion selection: a stale but readable completion must still select its bound review (using the existing valued-read predicate). Round lookup consumes ledger numbers in append order when timestamps share one second. Cancelled only my queued suite PID 62893 before execution, then requeued the corrected head; no peer run touched.

- Corrected head ac4f0623f60856d090eff40ece50deafbf55fa7d passed the observer package and all targeted collector/writer tests (40.920s). Recorded suite attempt 26619cadc5bf3d0eec52a5e8bd03c4f0 is queued through the weighted gate; no suite failure occurred. Local-gates guard passes; no protected configuration changed.
