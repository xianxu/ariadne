# Boundary Review — ariadne#423 (whole-issue close)

| field | value |
|-------|-------|
| issue | 423 — sdlc issue show reports the first close review's verdict, not the latest (REWORK shown after a SHIP) |
| repo | ariadne |
| issue file | workshop/issues/000423-sdlc-issue-show-reports-the-first-close-review-s-verdict-not-the-latest-rework-shown-after-a-ship.md |
| boundary | whole-issue close |
| milestone | — |
| window | 2bfbc4fa27ed30da2455ad38d4eaea0ddf82a774..4bea66dbd7c3275d35560124f274834b1505e16b |
| command | sdlc close --issue 423 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-17T04:24:25-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The main REWORK→SHIP selection and archived-history paths work, but stale tracker data exposes a blocking edge case: an issue without a completion binding loses its latest verdict. A temporary regression probe reproduced this without changing repository files.

```findings
findings:
  - id: new
    severity: Critical
    kind: defect
    family: binding-presence-independent-of-read-quality
    title: |
      Stale cards without completion bindings suppress the latest review
    detail: |
      cmd/sdlc/internal/observe/checkpoints.go:186 uses valued(comp.Read) as evidence that a completion binding exists. assembleCompletion preserves State=stale when a stale card has no binding (assemble.go:198), leaving ReviewedHead empty. A working issue with a valid latest SHIP consequently becomes unknown with an empty verdict and error "no review entry matches the completion reviewed head ". The complete family in this window is this selection guard and its shared text/JSON consumers. Check actual binding presence separately from read quality; preserve latest-review selection when absent and completion-bound selection when present. Add regression coverage for both stale states.
```

1. **Strengths**
   - Reuses `markdoc.StripFenced` to exclude fenced examples.
   - Tests completion selection, malformed latest evidence, ledger numbering, and duplicate timestamps.
   - The lifecycle test verifies REWORK→SHIP through archival and branch deletion.

2. **Critical findings**
   - The binding-presence defect above blocks shipment. Check binding existence before entering completion matching.

3. **Important findings**
   - None.

4. **Minor findings**
   - None.

5. **Test coverage notes**
   - Observer package passed.
   - Targeted lifecycle and sidecar create/append tests passed.
   - Temporary overlay test for a stale working card without completion failed: expected latest SHIP; received unknown.
   - `git diff --check` passed.

6. **Architectural notes**
   - **ARCH-DRY: pass.** Fence handling reuses the existing helper.
   - **ARCH-PURE: pass.** Selection remains deterministic; ledger collection stays at the IO boundary.
   - **ARCH-PURPOSE: flag.** The required “latest review when there is no completion” behavior fails for stale cards.
   - Atlas documents the changed readout. No new command, flag, or configuration surface requires a README usage update.

7. **Plan revision recommendations**
   - Add a `## Revisions` entry distinguishing completion-binding presence from tracker freshness, with fresh/stale × present/absent regression coverage.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"e2b7f2f1-a74d-4196-937d-412659227373","repository":"github.com/xianxu/ariadne","issue_id":"000423","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/8975df82cc57282c","reviewed_head":"4bea66dbd7c3275d35560124f274834b1505e16b","slot":"ariadne:5","requested_at":"2026-10-17T11:23:17.606778Z"},"phase":"finished","pid":88350,"started_at":"2026-10-17T11:23:17.763941Z","finished_at":"2026-10-17T11:24:25.184821Z","session_id":"01a1499a-1936-71b1-bd81-4532363ae12c","outcome":"returned"}] -->

---

## Re-review — 2026-10-17T05:29:35-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 423 — sdlc issue show reports the first close review's verdict, not the latest (REWORK shown after a SHIP) |
| repo | ariadne |
| issue file | workshop/issues/000423-sdlc-issue-show-reports-the-first-close-review-s-verdict-not-the-latest-rework-shown-after-a-ship.md |
| boundary | whole-issue close |
| milestone | — |
| window | 2bfbc4fa27ed30da2455ad38d4eaea0ddf82a774..ea1a9ee9a2c0855563cc269aac6dd3cd8e216565 |
| command | sdlc close --issue 423 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-17T05:29:35-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

BR-1 is addressed, with a regression test confirmed to fail without its fix. The main verdict-selection and archival paths pass. One blocking defect remains: independently sampled timestamps can produce incorrect round numbers and reverse history labels, violating the Done-when contract.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      The guard now requires comp.Token alongside valued read quality. The four fresh/stale × binding-present/absent cases pass. Removing the token guard in a disposable pinned-head copy makes TestReviewRoundsSelectCompletionOrLatest/stale_absent fail with the original unknown-verdict error.
findings:
  - id: new
    severity: Critical
    kind: defect
    family: review-provenance-requires-stable-identity
    title: |
      Timestamp drift corrupts review round numbers and history labels
    detail: |
      cmd/sdlc/internal/observe/checkpoints.go:238-255 joins entries to ledger rounds by exact timestamp, silently substituting entry order on a miss. Sidecar persistence samples nowRFC3339 separately from ledger persistence (reviewsidecar.go:211 and close.go:1526); the pinned #423 sidecar records 04:24:25 while its ledger records 04:24:26. A disposable regression probe with ledger rounds 4 and 6 and a one-second mismatch on the completion-bound second entry reports round 2 and labels the earlier round 4 “later, not completion-bound.” The complete family is the collector’s timestamp map (observe.go:318-322), both historical/current number lookups, and the numeric relation comparison (checkpoints.go:197), affecting close and milestone reviews and their shared text/JSON consumers. Use stable review identity with compatibility handling for existing artifacts; derive chronological labels from entry order. Add coverage for timestamp drift where ledger numbers differ from entry ordinals.
```

1. **Strengths**

   - Completion-bound selection preserves current blocking-finding counts.
   - Fresh/stale binding coverage verifies BR-1’s full matrix.
   - The lifecycle test exercises the real writer, collector, archival, and branch deletion.
   - Fence handling reuses `markdoc.StripFenced`.

2. **Critical findings**

   - The timestamp-identity defect above blocks shipment. Entry-order fallback cannot reliably identify ledger rounds.

3. **Important findings**

   - None.

4. **Minor findings**

   - None.

5. **Test coverage notes**

   - Observer package passed.
   - Targeted lifecycle and sidecar tests passed.
   - BR-1 mutation test failed as expected.
   - Timestamp-drift probe failed, reproducing incorrect numbering and history labels.
   - `git diff --check` passed. Disposable copies were removed; checkout remains clean.

6. **Architectural notes**

   - **ARCH-DRY: pass.** Existing fence processing is reused.
   - **ARCH-PURE: pass.** Selection remains deterministic; collection stays at the IO boundary.
   - **ARCH-PURPOSE: flag.** Required round identity and earlier-history labels remain unreliable.
   - Atlas was updated. No new command, flag, configuration, or usage step requires a README change.

7. **Plan revision recommendations**

   - Add a `## Revisions` entry requiring stable sidecar-to-ledger identity, compatibility for existing timestamp differences, and regression coverage with nonconsecutive ledger numbers.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"232ab55e-a8d6-4bd9-935d-82c57640c5e0","repository":"github.com/xianxu/ariadne","issue_id":"000423","recipe":"small-diff-review","round":2,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/31c8cc98ca1f4ce2","reviewed_head":"ea1a9ee9a2c0855563cc269aac6dd3cd8e216565","slot":"ariadne:5","requested_at":"2026-10-17T12:28:08.51601Z"},"phase":"finished","pid":62630,"started_at":"2026-10-17T12:28:08.613898Z","finished_at":"2026-10-17T12:29:35.460322Z","session_id":"01a149d5-7708-7312-96c4-6374b766dbd6","outcome":"returned"}] -->

---

## Re-review — 2026-10-17T06:13:35-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 423 — sdlc issue show reports the first close review's verdict, not the latest (REWORK shown after a SHIP) |
| repo | ariadne |
| issue file | workshop/issues/000423-sdlc-issue-show-reports-the-first-close-review-s-verdict-not-the-latest-rework-shown-after-a-ship.md |
| boundary | whole-issue close |
| milestone | — |
| window | 2bfbc4fa27ed30da2455ad38d4eaea0ddf82a774..8838469f4f7e1a03761b118b16e19f7e3a5fa314 |
| command | sdlc close --issue 423 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-17T06:13:35-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned range satisfies the issue’s selection, round-identity, and history requirements. BR-2 is addressed with reachable regression coverage; no new blocking findings were found. One broader test exits identically on both pinned base and head, so that failure is pre-existing.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Confirmed the binding-presence guard and passing fresh/stale × present/absent regression matrix.
  - id: BR-2
    disposition: addressed
    note: |
      Reviewed-commit matching replaces timestamp identity across collection and assembly; history labels use entry order. The timestamp-drift regression passes at head and fails with assignment disabled. The snapshot regression passes at head and fails with mutable-ref reads. Both disposable checkouts were removed.
```

1. **Strengths**
   - Conservative matching preserves unknown round identity instead of inventing numbers (`cmd/sdlc/internal/observe/checkpoints.go:282`).
   - Completion selection preserves chronological history and current blocking counts (`checkpoints.go:184`).
   - Evidence reads share a pinned commit, with a controlled ref-advance regression (`cmd/sdlc/observe.go:246`).
   - Atlas documentation describes compatibility handling and additive JSON fields.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage**
   - Passed the full pure observer package.
   - Passed lifecycle/archive, snapshot, milestone, and collector command-bound tests.
   - Both negative controls failed as expected.
   - Broader testing encountered `TestObserveAnchorsOnTheIssuesDir` exiting without an assertion message; isolated runs reproduce this on both base and head.
   - Full repository suite was not independently rerun.

6. **Architecture**
   - **ARCH-DRY — pass:** one shared round-assignment rule.
   - **ARCH-PURE — pass:** matching and selection remain IO-free; collection supplies facts.
   - **ARCH-PURPOSE — pass:** collector, historical/current numbering, and text/JSON consumers use the correction.
   - **ARCH-MOCK — pass:** existing injected Git seam and portable repository fixtures remain in use.
   - **ARCH-CONSTRAINTS — pass:** bounded local ref resolution; command-bound test passes.
   - **ARCH-SECURE — pass:** malformed or conflicting identities remain explicitly unknown.
   - **ARCH-ORDER — pass:** immutable snapshots prevent mixed-generation evidence; controlled ordering test proves it.
   - **ARCH-FUNERAL — pass:** no new durable artifacts or background work.

   Core-concept paths and classifications match the code. Atlas is updated; no new command, flag, or configuration requires a README usage update.

7. **Plan revision recommendations:** None.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"cb4c4b0b-c096-4c49-a464-23b4c1d13c76","repository":"github.com/xianxu/ariadne","issue_id":"000423","recipe":"milestone-review","round":3,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/b85b9c4da9ab46d5","reviewed_head":"8838469f4f7e1a03761b118b16e19f7e3a5fa314","slot":"ariadne:5","requested_at":"2026-10-17T13:10:50.396948Z"},"phase":"finished","pid":95110,"started_at":"2026-10-17T13:10:50.572298Z","finished_at":"2026-10-17T13:13:34.897437Z","session_id":"01a149fc-8f1e-75e3-93ad-bb536ba6e7e9","outcome":"returned"}] -->
