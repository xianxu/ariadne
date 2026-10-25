---
id: 000603
status: working
deps: []
github_issue:
created: 2026-10-24
updated: 2026-10-24
estimate_hours: 4.17
card_mirror: '835c5f4881314e3edbf6c3d04fd39bb2d0285bb8' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-24T15:56:22-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
---

# Measured actuals never block close: record exact, partial or unmeasured instead of refusing

## Problem

## Spec

**Measured actuals must never block a close; degrade instead of refusing (operator, 10-24).** The operator's framing:
"the goal of the estimation system is both a soft nudge for AI to break things down and reason about how complex the
parts are, and to be used for future estimates. It is auxiliary, and shouldn't be blocking landing, generally
speaking."

Since ariadne#580, close refuses measured actuals in several shapes (resting branch, fixed by #591; 64 KiB proof, fixed
by #593; stage branches, #594; detached checkouts; more than 128 sources in the window). Each refusal stops the close,
needs a TL authorization for `--no-actual`, and loses the hours entirely. Overnight 10-23/24 that happened to roughly a
dozen closes.

Change:
1. **Degrade, don't refuse.** Attribute what can be attributed; report unattributed minutes and why; record the result
   with a confidence: `exact`, `partial` (with the unattributed share) or `unmeasured` (with the reason). Close proceeds
   in every case; no TL authorization needed.
2. **No source-count refusal.** Past the source cap, merge or summarise sources rather than refuse.
3. **Calibration reads confidence.** Velocity and estimate calibration use `exact` fully, `partial` weighted by its
   attributed share, and skip `unmeasured`.
4. **Retire the `--no-actual` ritual** for measurement failures (keep the flag for genuinely inapplicable windows).

## Done when

- Every refusal shape listed above closes with a recorded confidence instead of refusing; tests replay the real
  pair#856, #812, #863, #830, #851 and #838 windows.
- Calibration output distinguishes and weights the three confidences.
- Close help and the atlas describe the confidence levels.

## Plan

Durable plan: [000603-measured-actuals-never-block-close-plan.md](../plans/000603-measured-actuals-never-block-close-record-exact-partial-or-unmeasured-instead-of-plan.md).

- [x] M1 — close never refuses on measurement: confidence classifier, unknown-branch and per-source notes, oversize summary, queue degradation, close adopts every outcome, replay fixtures (#856/#812/#863/#830/#851/#838)
- [x] M2 — calibration weights by confidence (ledger columns, weighted forecast/converge/drift); close/actual help and atlas

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: greenfield-go-module   design=0.3 impl=0.2
item: smaller-go-module      design=0.1 impl=0.15
item: smaller-go-module      design=0.05 impl=0.12
item: smaller-go-module      design=0.05 impl=0.12
item: smaller-go-module      design=0.1 impl=0.15
item: smaller-go-module      design=0.15 impl=0.2
item: smaller-go-module      design=0.15 impl=0.2
item: smaller-go-module      design=0.05 impl=0.12
item: smaller-go-module      design=0.1 impl=0.15
item: greenfield-go-module   design=0.4 impl=0.3
item: smaller-go-module      design=0.1 impl=0.2
item: atlas-docs             design=0.1 impl=0.06
item: milestone-review       design=0.0 impl=0.15
item: milestone-review       design=0.0 impl=0.15
design-buffer: 0.15
total: 4.17
```

Items, in order: confidence classifier (Task 1); unattributed member (2); Validate split (3); branch Unknown (4); per-source
drop (5); queue degradation at four sites (8); close adoption, trailer and cue field (9); ledger columns (11); weighted
calibration (12); summary with credits, coalescing and exclusion (6, 7); replay fixtures (10); help and atlas (13); M1 and M2
reviews. Impl values are 40% of the v2 table (v3.1); design is low because the durable plan already settled it.

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only.*

## Log

### 2026-10-24
- 2026-10-24: closed — BR-4 fixed: the OversizeError size assertion moved into TestMeasurementPastStructuralCapCountedRefusal, which reaches the encoder structural cap; with bytes = size.Bytes reverted it fails (at least 703431 bytes, want the size reached), restored it passes; gatestate package green. Full suite green on bacff4e6; since then only checkbox edits and this test move. (fixed-then-ship); review verdict: SHIP
- 2026-10-24: closed — Full suite green on bacff4e6 (sdlc-local/suite, 0 failed, 17m50s); since then only plan/issue checkbox edits. M1 closed (a00a4f59c). Close never refuses on measurement: exact/partial/unmeasured confidence (classifyActual table), unknown-branch and per-source attribution recorded as unattributed, Validate split (pair#838 fixture reproduces its refusal under mutation), pair#851 and pair#866 exact plus summary through the real resolver, summary credits serve exclusion, queue unknowns upper bound weight 0, corrections refuse non-exact; legacy close writes a valid actual_confidence token (mutation fails YAML), tracker close the Close-Actual-Confidence trailer; ledger confidence columns round-trip, weighted calibration property tests; one persistedLine predicate for unattributed strings (property test, mutation caught). Help, atlas and README updated.; review verdict: FIX-THEN-SHIP
- 2026-10-24: closed M1 — M1 review round 2 fixed (BR-2 widened gaps bounded on the writer, TestCoalesceCreditsBoundsWidenedForTheValidator validates 3x the bound; BR-3 one rune-safe BoundReason for every unattributed writer, TestBoundReason; summary bytes from the measured OversizeError). Round 1: valid actual_confidence token (mutation fails YAML parse), Close-Actual-Confidence trailer e2e, corrections refuse non-exact, ambiguous owner unattributed. Source-cap fixtures pair#838 (mutation reproduces its refusal), #851, #866 exact plus summary; targeted batch green at ecaf2d671; touched-test runs green at c6af4ce68.; review verdict: FIX-THEN-SHIP
- Planned after #593 landed (TL: take #603, plan to TL). Code survey: (b)/(e) oversize and source cap already hold exact hours at refusal; (d) unknown branch, (f) per-source attribution and (g) queue evidence fail inside the input resolver and discard partial windows. Open decisions D1-D5 in the durable plan go to the TL.
- TL approved the plan (D1 reversed: resting/stage keep a close exact), then three design deltas: the uncounted minutes as a top-level `unattributed` member, the summary keeping its credits, `Validate` split into shape and bounds (plan-quality round 1); and every unknown queue attempt taking the D4 path (Task 8: the engine binds queue waits to sources by slot, so an unknown attempt's wait cannot be subtracted, and replaying such a subtraction would make the records unignorable). The TL set a limit: a fourth delta means cutting scope.
- pair#838 and #851 both refused at the branch proof (`measurement branch proof: measurement sources exceed snapshot limit 128`, pair 09f4314e9 and the #851 Log), on untrimmed sources; #593 did not reach that check. They have no raw captures here, so their fixtures are shaped generators (ARCH-PURPOSE: the shape, not the bytes).
- The summary keeps the coalesced gaps (`widened`) beside its credits, so a later issue can record what over-protection cost it; its byte bound is 448 KiB (2048 credits plus 2048 gaps, second precision), inside the 512 KiB structural cap.
- Unattributed minutes subtract only this issue's own counted cover: a minute another issue's source counted in parallel is still this issue's lost minute.
- A correction keeps refusing on incomplete queue evidence; only close and preview degrade.
- pair#866 (TL data point): closed refused under the source cap on bytes, "measurement snapshot is 182432 bytes, over the version 4 bound of 131072 (119 sources, 1 branch records, 860 branch intervals)". Fixtures: gatestate `TestMeasurementPair866ShapeIsTypedOversize` (that exact shape is valid and its bound is the typed oversize) and the pair#866 case of `TestSourceCapMeasuresInsteadOfRefusing` (119 contributing sources over the byte bound through the real resolver: exact plus summary).
- M1 boundary review round 1 (FAILED on a stray path from another process, `~/couch-profile-871-restage-20261024`, profiles written 18:38 by an unrelated restage; not removed): one Critical (unquoted `actual_confidence` frontmatter), five Important (unsupported `correct-actual` advice, corrections adopting partial measurements, ambiguous-owner claims filed under the sentinel, missing Task 5/9 tests, stale README) and five Minors, all fixed in the round. New tests: `TestLegacyCloseRecordsEveryConfidence` (mutation: the old full-string write fails YAML parsing), `TestTrackerCloseCarriesConfidenceTrailer`, `TestConfidenceLinesNoGatesigCollision`, `TestOwnedActualAmbiguousOwnerIsUnattributed`, `TestConfidenceReasonIsBoundedAndHomeRelative`.
- M1 boundary review round 2: FIX-THEN-SHIP with BR-2 (widened gaps unbounded on the writer: 5000 credits gave 2952 gaps against a 2048 validator bound, dropping the close to unmeasured) and BR-3 (unattributed reasons unbounded on the writer, three truncation rules); one class, writer bound below the validator's. Fixed together: widened coalesced to the bound, one `gatestate.BoundReason`, summary bytes from the measured `OversizeError`; tests `TestCoalesceCreditsBoundsWidenedForTheValidator` (3x the bound validates) and `TestBoundReason` (rune-safe cut).
- M1 closed (evidence a00a4f59c, FIX-THEN-SHIP, no blocking findings after round 3). Its advisories fixed in the M2 window: the writer-bound class as a rule, one `persistedLine` predicate shared by `Validate` and `BoundLine`/`Normalised` (NUL, invalid UTF-8 and Source now bounded too), pinned by `TestUnattributedNormalisedAlwaysValidates` (mutation: keeping NUL fails it); `OversizeError.Bytes` asserted on the structural-cap path. Lesson recorded.
- Verification so far: targeted batch (Actual|Queue|Branch|Exclusion|Measurement|Owned|Calib|Close|... over ./cmd/sdlc/...) passed at 0134ef86d; at 98f6537cf four tests failed that pinned the retired --no-actual guidance and the pre-correction header width, now updated. Mutation: restoring the full `Validate` at the branch proof makes the pair#838 case fail with pair#838's exact refusal.
