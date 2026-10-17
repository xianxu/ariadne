---
id: 000423
status: working
deps: []
github_issue:
created: 2026-10-17
updated: 2026-10-17
estimate_hours: 1.05
collaboration-mode: collaboration
card_mirror: '33c22e1e0e6ee6d778fceb1469415f3d2b9e3591' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-17T02:22:23-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
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
- Fresh and stale tracker reads each preserve latest-review selection without a completion and bound-review selection with one.
- Round identity uses reviewed commits, tolerates independently sampled timestamps, and reports unknown when neither identity nor complete legacy alignment is provable. History labels follow entry order.

## Plan

- [x] Reproduce appended REWORK→SHIP readout with completion selection, round/history, and quoted metadata coverage.
- [x] Select the completion-bound entry (latest without completion), using ledger round numbers when available; preserve legacy sidecars and explicit unknown reads.
- [x] Verify text/JSON through the collector and archived evidence and run the recorded suite; proceed to the close/landing gates on this evidence.

- [x] BR-1: separate binding presence from freshness and verify all four fresh/stale × present/absent cases.
- [ ] BR-2: complete the approved durable plan, recorded suite, re-close and landing.

## Estimate

Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only. Calibration is provisional/stale (#127). This estimates the final full-flow issue scope, including already implemented parsing/selection; it is not fitted to elapsed suite queues or an actual preview. Three existing Go extensions cover entry parsing/selection, pure round identity, and coherent collection/rendering. Each uses v2 design 0.2h discounted ×0.2 for the settled spec, and implementation 0.5h ×0.40. Existing Go/Git seams require no new-library discovery; familiarity 1.0. Docs use design 0.1h ×0.2 and implementation 0.1h ×0.40; one review uses design 0.2h ×0.2 and implementation 0.5h ×0.40. Thorough-plan buffer 15% applies to design only. Design 0.18h ×1.15 + implementation 0.84h = 1.047h, rounded to 1.05h. This forward model does not legitimize the known queue-polluted actual (#420).

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: smaller-go-module design=0.04 impl=0.20
item: smaller-go-module design=0.04 impl=0.20
item: smaller-go-module design=0.04 impl=0.20
item: atlas-docs design=0.02 impl=0.04
item: milestone-review design=0.04 impl=0.20
design-buffer: 0.15
total: 1.05
```

## Log

- BR-2: full flow admitted after plan-quality CLEAN (PQ-1/PQ-2 addressed) and estimate-quality INFO. Estimate item implementation budgets include each extension's direct tests; the review item includes one planned close and the measured ~8–9-minute suite execution, excluding inactive queue waiting. The model is a final-scope baseline, not a reconstruction of already spent review/reintegration cycles; those repeats are variance rather than silently back-fitting the estimate. Actual stays N/A pending #420.
- RED confirmed the collector mixes generations when a ref advances between reads. GREEN: pure observer tests cover stable commits, gaps, duplicates, absent identity and headless legacy alignment. A headless/conflicting ledger outside complete compatible alignment cannot prove any group's multiplicity, so it remains unknown. The focused collector/writer suite passed in 77.048s before the final conservative partial-legacy guard and retained-run assertion; rerun follows on final code.

### 2026-10-17

- Root cause: observe.sidecarRows deliberately keeps the first metadata table, but the writer appends Re-review sections. Chosen fix parses only each entry’s metadata before its Review body, reuses markdoc fence filtering (ARCH-DRY), and keeps selection pure (ARCH-PURE). Match completion ReviewedHEAD for close; unmatched bound evidence is unknown rather than silently substituting a later review. No new persisted state or artifacts; ledger timestamps map entries to existing round numbers, with entry order for legacy files.

- RED: TestReviewRoundsSelectCompletionOrLatest returned round-1 REWORK for latest and bound round-2 SHIP, and accepted unbound completion evidence. GREEN: observe package passes, including quoted tables/fenced headings, missing latest verdict, completion selection past later malformed evidence, ledger round numbers and current open findings. Collector/writer targeted tests passed in 53.896s; the real close/archive/branch-deletion lifecycle now includes REWORK then SHIP and preserves both rounds. Golden changes add optional round/history fields; existing schema version remains 1 (additive fields).

- TL-granted :0 live check: staged 016625f7550d30c5708c3e803e31f5dd1b4b85e1, all seven tools verified vcs.revision exact and vcs.modified=false. Native issue show 416 text/JSON reported SHIP round 2 at 2026-10-17T00:40:55-07:00, window d1a4e571..412b0dc4 matching completion; history labelled earlier round 1 REWORK. Cleared :0 to clean main 5fbaf9730a5717a6194544b527e6c573d6a017bc and rebuilt/verified all seven tools there; hub notified it is free. PR #281; recorded suite attempt d521ca0e272e7faa25d9d40f9fc536a7 queued on 016625f7.

- Pre-close edge sweep reproduced and fixed stale-tracker completion selection: a stale but readable completion must still select its bound review (using the existing valued-read predicate). Round lookup consumes ledger numbers in append order when timestamps share one second. Cancelled only my queued suite PID 62893 before execution, then requeued the corrected head; no peer run touched.

- Corrected head ac4f0623f60856d090eff40ece50deafbf55fa7d passed the observer package and all targeted collector/writer tests (40.920s). Recorded suite attempt 26619cadc5bf3d0eec52a5e8bd03c4f0 is queued through the weighted gate; no suite failure occurred. Local-gates guard passes; no protected configuration changed.

- Recorded suite 26619cadc5bf3d0eec52a5e8bd03c4f0 passed on ac4f0623: all 12 shards over 1543 cmd/sdlc tests and 66 other packages, 516.414s wall, 1724.956s queued, log SHA256 dd62414f23cb11ee2007cacad4f2fe8c622aee16336f91419623b9e75bed74d6. Main advanced with #421 while it ran; merged current main cleanly and re-running the recorded suite on the integrated head before close. Actual preview measured 0.16h through the last work checkpoint; no hand-entered actual.

- Integrated recorded suite 2e0137236f6c98802b3328f0539ac922 passed on 0227eeb1ddb632300d3ca903bb616cd4b37c9bcc: all 12 shards over 1547 cmd/sdlc tests and 66 other packages, 552.651s wall, 2107.431s queued, log SHA256 f4ebb7d127d3e74bc386eff278a23a92c56389fd84423ef90f4db2068ea9aad9. CI is green. All implementation/verification plan items are complete; close review and landing are the remaining gates.

- Main advanced with #422 during the integrated suite. Merged it cleanly; a third recorded suite will verify current main. TL has coordinated a hold on main writes through #423 landing to prevent another drift cycle. Actual preview now says 1.49h, including the queue-polling interval tracked by #420; do not adopt that as focused work or hand-subtract an unbound wait.

- Final integrated suite c896f6a83d2fce9a92d66e0558871d4d passed on bfcfa3954220dcaa401c095717a5f73f8a632393: 12 shards over 1547 cmd/sdlc tests, 66 other packages, 533.026s wall, 1254.843s queued; log SHA256 131323a7722899993361ed4e6ffe854d8bb0ffb3c6633870f14c8840d5fbfb90. This includes current main (#421 and #422). Close records actual_hours N/A via the precise no-actual acknowledgment: preview had expanded from 0.16h to 1.49h while queue polling bridged idle time (#420); adopting that sample or hand-subtracting waits without retained ownership provenance would misstate focused work. No numeric actual is invented.

- BR-1 RED: stale absent case returned unknown with an empty completion head. GREEN after checking the parsed token as well as valued read quality: full observe package and targeted collector/writer tests pass (43.977s). Four-case matrix covers both stale-with-binding and stale-without-binding through the shared text/JSON selection. Added the read-quality/presence lesson. Waiting only for #420 to land, per TL correction; integrate it before the next recorded suite, then re-close.

- TL released the #420 dependency because it needs a re-plan. Current main is already integrated; queued the post-BR-1 recorded suite on pushed head 2cf29e61ee883d6316d625b7364292d2a55e5586. Continue to re-close and land on SHIP without waiting for #420.

- Post-BR-1 recorded suite 566a8fde2ee05a22964093a6b59cc9da passed on 2cf29e61ee883d6316d625b7364292d2a55e5586: 12 shards over 1547 cmd/sdlc tests, 66 other packages, 495.778s wall, 1550.958s queued, log SHA256 101f015fa77fbd02772ccc4375ebf555fac1f59240724dc191ccea92da6973db. Re-close now reviews the four-state presence/freshness fix. Actual remains N/A because #420 is deferred and the existing engine attributes queue polling as focused work.

## Revisions

### 2026-10-17 — BR-1: binding presence is independent of read quality

Round 1 REWORK reproduced a stale working card with no completion: assembleCompletion retains a stale Read but has no parsed binding, so a valued-read guard matched an empty reviewed head and suppressed the latest verdict. Require parsed binding presence as well as readable provenance; enumerate fresh/stale × present/absent for the shared text/JSON selection (ARCH-PURPOSE). No new state or mechanism. TL approved this revision.

### 2026-10-17 — BR-2: stable round identity

Round 2 addressed BR-1 but found timestamp equality is not identity: the actual round-1 sidecar is stamped 04:24:25 and ledger 04:24:26. A RED probe with ledger rounds 4 and 6 reported round 2 and called earlier round 4 later. TL approved joining sidecar window heads with existing ledger Reviewed/ReviewerRuns heads, complete known-head-compatible sequence alignment for legacy entries, otherwise explicit round unknown. Labels derive from entry positions, never round-number magnitude. No writer changes or durable state. Durable implementation plan: workshop/plans/000423-sdlc-issue-show-reports-the-first-close-review-s-verdict-not-the-latest-rework-shown-after-a-ship-plan.md. Use full flow if the final patch leaves the shell.

### 2026-10-17 — plan-quality snapshot refinement

PQ-1 confirmed that the collector reads mutable refs independently, which can mix generations during a concurrent update. Refine BR-2 by resolving each evidence ref once and using its immutable commit for all artifacts; add controlled ref-advance integration proof (ARCH-ORDER). PQ-2 compresses test prose into named adversarial strategies. The reviewed-commit matching contract is unchanged; hub notified before implementation.

### BR-2 implementation verification

Final pure observer package passes (0.094s); focused observer/collector/writer tests pass (69.130s). Controlled real-Git ref advancement is RED before pinning and GREEN after; its ledger also proves nonfinal retained reviewer-run identity despite timestamp drift. The real close/archive lifecycle reports ledger rounds 4 and 6 before and after branch deletion. Unknown round identity remains independent of the readable verdict, exposed as round_error in JSON and round unknown in text. No writer or durable artifact format changes. Current origin/main 2bfbc4fa27ed30da2455ad38d4eaea0ddf82a774 is already integrated. Next: recorded suite on pushed head, then re-close.
