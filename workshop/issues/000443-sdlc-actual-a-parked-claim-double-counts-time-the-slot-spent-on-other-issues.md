---
id: 000443
status: working
deps: []
github_issue:
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
collaboration-mode: collaboration
card_mirror: '112a88c64ce5127589fe38b48fcf0f45633dabe0' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-18T10:57:47-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "82ec2867", done: "d88d1c56"}
---

# sdlc actual: a parked claim double-counts time the slot spent on other issues

## Problem

A slot that holds a claim on one issue while it works others gets the same time counted twice. pair:6 held pair#636 from 2026-10-17 10:14 to 2026-10-18 08:41, parked behind pair#647, and meanwhile worked pair#649 and pair#656, which each adopted their own measured actuals. `sdlc actual --issue 636` then reported 9.28 h over 11 admitted sessions, mostly time already counted toward #649 and #656. The slot had to close with --no-actual.

## Spec

Active time a slot spends is attributed once. When measuring an issue, exclude intervals that are already attributed to another issue's adopted actual from the same slot, or, more simply, attribute each interval to the issue whose branch or commits it touched. Pick the rule that needs the least new evidence and say why. A parked claim (owner kept, no work) should accrue nothing while its slot works something else.

## Done when

- Measuring an issue whose claim overlaps another adopted issue from the same slot does not count the overlap twice; a test reproduces the #636 shape.
- The refusal or report names the excluded intervals and why.

## Plan

- [x] Use another issue's adopted actual and pinned measurement inputs as exclusion authority, never a live claim or an unscoped main commit. For each overlapping same-slot adopted issue, replay its pinned inputs, take the source-scoped activity runs credited to that issue, and subtract their time interval from the candidate before attribution. An ambiguous/missing pin refuses a numeric actual rather than silently counting overlap. The exclusion is conservative when a credited run was split fractionally among issues; report its issue, slot and interval.
- [x] Add a red #636-shaped integration fixture: a parked issue with intervening adopted work in the same slot. Assert preview and close candidates use the shared measurement path; no credited interval is counted again. Exercise differing slots and overlapping adopted intervals in the pure core, missing historical evidence in the integration fixture, and replay after a later transcript append.
- [x] Implement the pure interval subtraction in `internal/activetime`, the pinned adopted-issue evidence loader in the actual-time IO seam, and concise interval/refusal diagnostics. Preserve current branch commit scoping (#270).
- [x] Run targeted tests, `make -k`, the recorded full suite on the pushed head, then close review and land on SHIP.

## Log

### 2026-10-18

- Planning choice: branch-only attribution cannot guarantee exclusivity because #270 intentionally hides unrelated main commits from a branch's segment boundaries. Reusing completion-pinned actual inputs needs no new durable ledger (ARCH-DRY, ARCH-FUNERAL). A replay gap remains unknown, not zero (ARCH-PURE). This can conservatively undercount activity where two issues genuinely overlap in the same slot; the report names every excluded interval.
- The #636-shaped fixture measured 20 minutes for the parked issue before exclusion and less after replaying #649's adopted 10-minute run. The transcript grows after #649's close, so its saved byte prefix is the replay authority. Removing that source makes the candidate refuse and names #649 plus the overlapping time window. A fully excluded candidate reports `--no-actual` with the protected interval instead of telling the slot to restore evidence.
- Resumed after #437 and #444. The first full-suite run exposed a legacy close fixture whose candidate source had no slot pin and no other adopted actual. The loader now scans for temporally relevant adopted actuals before requiring candidate slot pins. The legacy close passes; a #636-shaped candidate with a relevant adopted actual still refuses an unpinned source.
- The first ariadne:0 live check on `5f724b3c` refused #443 because #417's claimant remained on its done card even though #417's adopted measurement ended Oct 17. The reviewed redesign reads the completion-pinned time window before source replay. A read-only run of the revised branch measures #443 at 1.09h and names intervals already adopted by #437 and #444; the earlier #417 window no longer blocks it. The initial stage was cleared immediately, restoring all seven tools to main `ce8cbf95daeef8153d651da7f21f612ed7940ac1`.
- Final verification on pushed `1f6267838e642b827d04bbbf2a8d240ff58e60f1`: targeted actual-time tests and `make -k` passed; recorded suite passed 12 shards (1,656 sdlc tests) and 68 other packages, zero failures, log SHA-256 prefix `4e8f3f425345`. The final ariadne:0 stage built all seven unmodified binaries at that head; installed `sdlc actual --issue 443` measured 1.37h, named already-adopted #437/#444 intervals, and excluded 202.41 minutes of recorded suite queue wait. Clear rebuilt all seven unmodified binaries at main `ce8cbf95daeef8153d651da7f21f612ed7940ac1`.

## Revisions

- 2026-10-18 11:12 PDT — The integration fixture uses one intervening adopted issue and a later transcript append; the pure core separately tests the union of two adopted issues. This keeps the tracker fixture focused while still defending the multi-issue rule. Close and preview share `computeActualForPurpose`; correction enters the same resolver and measurement path.
- 2026-10-18 11:15 PDT — A queued suite exposed a replay gap during self-review: new exclusions must travel with the existing completion-pinned `MeasurementInputs` or a later historical preview would drop them. Cancelled that queued suite before it acquired the machine lock, and will rerun it on the corrected head. The plan now adds an optional, bounded snapshot field and a candidate-to-replay test; no separate ledger is created.
- 2026-10-18 16:53 PDT — A full-suite fixture with no other adopted actual exposed an over-strict candidate slot check. Filter temporally relevant numeric adoptions first; only then require source-slot pins for the same-slot overlap decision. This preserves refusal for missing identity when overlap could matter while allowing legacy measurements with nothing to exclude.
- 2026-10-18 17:21 PDT — Live staging exposed a false refusal: #417's adopted work ended Oct 17, before #443's Oct 18 claim, yet its retained claimant made engagement history extend through #443's window. Re-plan against the adopted actual's saved measurement window before slot comparison or transcript replay; a disjoint saved window cannot exclude later work. Missing saved window remains unknown when temporal overlap cannot be ruled out. The hub approved this direction; disk-heavy work waits for the operator's trace hold to lift.
