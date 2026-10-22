---
id: 000531
status: working
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-22
estimate_hours: 1.38
collaboration-mode: collaboration
card_mirror: '7aad8d75d3016078e9df3aead092cf3fda7fc1e0' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-22T06:11:42-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
---

# sdlc actual: a milestone after a superseded one measures 0 h (pair#737 M3)

## Problem

## Spec

pair#737 (pair:1, 10-21): after the operator-approved re-scope (old M2 superseded, new work tagged M3 per the TL's
instruction), `sdlc` measures **0.00 h** active time for M3 despite 21 admitted sessions and substantive M3 commits
since the re-scope. The worker correctly refuses to hand-type hours. Likely cause: the milestone's measurement window
is anchored on a boundary (the superseded M2's start/close, or the M3 tag's first appearance) that excludes the
actual M3 work, or a superseded milestone isn't modelled at all.

Fix: measure a milestone that follows a superseded one from the supersession point (the last closed boundary or the
revision that superseded it) to the milestone's close, and make the measurement explain its window (start commit,
end, session count) so a 0 h result is visibly wrong rather than silently adopted.

## Done when

- A fixture with M1 closed, M2 superseded (never closed), M3 worked on measures M3 from after M1's close (or the
  supersession revision) and attributes the M3 sessions; a regression for the 0 h case.
- `sdlc actual` prints its window and session count.

## Estimate

Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only; source is flagged stale. Pinned Plan resolver uses existing Git and Markdown libraries (library availability discount); selector, diagnostics and consumer wiring extend existing modules. Detailed approved design applies the 0.2 design discount, familiarity 1.0 and 15% design buffer. Implementation entries use 40% of v2 ranges; review includes isolated production proof.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: greenfield-go-module design=0.05 impl=0.32
item: smaller-go-module design=0.06 impl=0.20
item: smaller-go-module design=0.04 impl=0.12
item: smaller-go-module design=0.06 impl=0.20
item: milestone-review design=0.02 impl=0.20
item: atlas-docs design=0.03 impl=0.04
design-buffer: 0.15
total: 1.38
```

## Plan

Full-flow proposal: [implementation plan](../plans/000531-retired-boundary-accounting-plan.md).

- [x] Replay pair#737 and define retired-boundary accounting from the pinned canonical Plan.
- [x] Share cut selection across fresh tracker and legacy measurements; preserve retained replay cuts.
- [x] Explain effective milestone phases and source counts; verify adversarial and full-suite coverage.
Follow-through (pending): close through the mandatory review and land.

## Log

### 2026-10-21

### 2026-10-22 — claim and root-cause investigation

Claimed in ariadne:2 and started planning. BuildTimeline sends an unfinalized boundary's convergence to the complete measurement right edge, so its monotonic emitter leaves no M3 phases. Neither BoundaryRound nor gatestate.Round / MeasurementCut has supersession metadata.

Read-only pair#737 evidence in pair-slot1: M1 finalized 10-20 21:23:47 -07; M2 had four REWORK rounds, first dispatched 10-21 00:01:03 and last returned 10:31:26, never finalized. The issue's canonical Plan replaces M2's executable checkbox with a historical plain bullet; M3 is its own executable row. The durable plan explicitly preserves findings/cap/history without a SHIP claim. Git records that approved retirement at 10-21 10:59:39 -07. Its Log records M3 measuring 0.00h through 11:48:39 despite 21 admitted sessions and work during 10:53–11:31. The first M3 dispatch occurred later, 14:05:10, proving that waiting for another review is insufficient.

Proposal for hub review: use the existing last-finalized-boundary fallback allowed by Done when, rather than inventing a prose-derived supersession timestamp. Before producing fresh cuts, omit unfinalized named boundaries absent from the exact code pin's canonical issue Plan. Keep finalized boundaries, current executable boundaries and whole-issue rounds. Report the retired boundaries and effective selected phase range. Missing/ambiguous required Plan evidence refuses with an exact restore/commit/retry action; absence is never an empty authority set. Saved snapshots continue replaying their exact original Cuts, without applying today's Plan. ARCH-DRY/PURE: one pure selector before both fresh consumers, no duplicated timeline implementation. No code or test implementation has started; awaiting hub scope approval.

Hub approved the full-flow membership/fallback policy and isolated production-path proof (authenticated submitted delivery a5d4ada9-9b0c-47c0-836b-6fafdb35a465). No supersession event or schema change; proceed through change-code.

Plan-quality round 1 raised PQ-1 (function-level test contract). Named selector, pinned resolver, phase diagnostic and production compute/replay functions; compressed the executable-case inventories into adversarial strategies and mechanical guards. Policy unchanged; re-running admission.

Plan-quality round 2 CLEAN: PQ-1 addressed; no remaining findings. Estimate derived only after acceptance: 1.38h, provisional stale calibration. Baseline pure timeline/boundary/conservation tests PASS.

### 2026-10-22 — implementation and isolated proof

Admission completed: plan-quality CLEAN after PQ-1 correction (third invocation rechecked the estimate-bearing body); estimate-quality INFO, no blocking findings. Estimate module allowances include their unit/native fixtures; review allowance is tight for suite elapsed time and remains provisional rather than retrofitting the number.

Implemented fresh split-cut selection shared by tracker and legacy paths. Plan membership is read at the exact measurement code pin, with identity/uniqueness/fence validation and actionable refusal. Finalized history, executable open reviews and whole-issue rounds remain; retired unfinalized named rounds are omitted only from newly retained ordinary Cuts. Timeline engine, review ledger, findings/caps and persisted formats are unchanged. Legacy code history is pinned once for both its engine and Plan read. Rendering shows selected phase bounds, code pin and truthful session provenance, including empty/blocking cases.

TDD: native pre-M3-review regression first failed at 0h; after selection it measures exactly 25/60h. Silent-zero diagnostic control also failed before rendering was added. Identity adversarial tests exposed duplicate-ID acceptance and quoted-ID rejection; reused issue.FieldString to reject malformed YAML while accepting scalar quoting. Targeted controls then PASS. Captured actual pair#737's pre-review Plan from Git-confirmed dd1109784863d347b580f46e2d0bf4d4e7d1a54b and sanitized real round timestamps in testdata/actual/pair737-retired-boundary.json. The capture reproduces the old zero and attributes the Log's minute-resolution 10:53–11:31 range as 38 minutes with the new ordinary cuts; it contains no raw transcripts and does not claim to replay 21 original sessions.

Tracker-native Codex fixture confirms 0.1h M3 and that fresh retained Cuts exclude M2; old unfiltered snapshots still replay zero and new filtered snapshots replay 0.1h. Claude legacy production fixture supplies the independent provider path. Pure property selection checks 500 generated ledgers against independently stated keep rules. Native Git wrapper counts exactly two Plan reads, one listing and one body, at both 100 and 200 rounds. Adversarial pinned Plan tests cover source authority and visible refusals.

Verification: go test ./cmd/sdlc -run 'Test.*Actual|Test.*Split|Test.*Measurement|TestCapturedPair737' -count=1 -timeout 10m PASS (374.272s); full activetime and gatestate package tests PASS (9.694s / 0.342s). Broad CLI run retained to completion without cancellation. Recorded full pushed-head suite and close review remain.

Final targeted verification after pinning the legacy engine and identity parsing changes: retired-boundary / original split controls PASS (7.339s). The captured real sequence and tracker replay controls are included. Preparing the exact pushed head for its recorded suite.

### 2026-10-22 — full-suite guard correction

Recorded pushed-head suite attempt f5e4adb1f14c8d5dd1c3330389679e87 / lease-3108312308 completed normally after 3550.675s queue and 853.942s execution. Across 2241 SDLC tests and 76 other packages, the sole named failure was TestIssueHintsRenderDecimalIDs: both new --issue recovery messages needed issue.CLIRef. Seven SDLC shards and all other packages passed. Log SHA256 002ea27e66fbe4350237d4963fe46b021faf41f0b017a2115d8422e2693eda4a. This is a code guard failure, not a flake or infrastructure failure; no flake registration or resume was used.

Extended native refusal coverage to padded 000009 and observed red for the emitted padded command, alongside the existing source guard. Fixed both hint sites with the shared issue.CLIRef formatter and asserted decimal replacement guidance as well. Targeted issue-hint guard, retired-boundary tests, captured pair replay and owned replay PASS (7.188s). Repeating the full suite on the corrected pushed head; the original runner was retained to exit and is not duplicated.

### 2026-10-22 — corrected recorded suite PASS

Attempt 3d171593dcdf291b51b28c47dd4e440b / lease-1530253159 passed on the exact pushed code head after 6265.356s queued and 1101.583s execution. All eight shards (2241 SDLC tests) and all 76 other packages exited zero, with no failures. Log SHA256 7213a4340974d2459072ee959c9d0baad1bc20f29657131e806ff7d7528d6a0a. The retained runner survived Couch restart; nothing was restarted or duplicated. The prior source-guard failure remains recorded and is fixed by the decimal-reference change.

GitHub Actions Results delayed publication despite successful run conclusions; the hub confirmed the upstream incident. PR checks subsequently published success without bypass or Actions rerun. No global recovery claim is made. Updating verification bookkeeping and entering the mandatory close review; close and landing remain pending.

## Revisions

### 2026-10-22 — completion bookkeeping

Separated the pending close/landing procedure from implementation acceptance checkboxes so the close gate can review completed implementation without claiming that its own review or landing already happened. No design or acceptance change.
