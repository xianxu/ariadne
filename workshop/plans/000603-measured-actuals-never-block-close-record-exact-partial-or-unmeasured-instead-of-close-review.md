# Boundary Review — ariadne#603 (whole-issue close)

| field | value |
|-------|-------|
| issue | 603 — Measured actuals never block close: record exact, partial or unmeasured instead of refusing |
| repo | ariadne |
| issue file | workshop/issues/000603-measured-actuals-never-block-close-record-exact-partial-or-unmeasured-instead-of.md |
| boundary | whole-issue close |
| milestone | — |
| window | 41238256b57c82978e4ba83a4b17a8f7969ce0e8..f880c83bad3e6618d7fb8f05637ca6e294a744d2 |
| command | sdlc close --issue 603 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-24T22:17:26-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Disposing BR-4/BR-6 now; running the gatestate and sdlc tests, then a revert check on the BR-4 fix.

Revert check for BR-4: drop `bytes = size.Bytes` and rerun the oversize test.

BR-6's regression test fails without the fix, as it should. BR-4's test still passes with the fix reverted, so I'm checking whether the structural-cap path can be reached at all.

The added assertion sits in the wrong test. I'm checking that the fix holds on the actual structural-cap path.

The fix works on the structural-cap path, but no committed test fails without it. Next I'm running the full `cmd/sdlc` suite and going through the close path and the docs.

While the suite runs, I'm reading the plan and the close path.

The package hit Go's 600s timeout and my `tail` cut off the header. I'm rerunning only the tests relevant to #603.

The #603-specific tests pass. I'm running the other actual/exclusion/queue/correction tests verbosely to find what ran into the timeout.

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

This round fixed the third repeat of `writer-bound-matches-validator` by enforcing a rule, not by patching one more site. `Validate` and the writers now check strings with one shared predicate (`persistedLine`). Every writer that builds a `MeasurementUnattributed` record persists it through `Normalised()`. A property test checks that any normalised record validates, and I confirmed it fails with the fix reverted. The BR-4 fix is also correct: on the encoder's structural-cap path, `OversizeError.Bytes` now carries the measured size (703431 in my probe) instead of 0. But the new assertion went into a test that never reaches that path, and it still passes with the fix reverted. So BR-4 has no regression evidence and stays `not-addressed`. It is Minor and cheap to fix by moving the assertion. Nothing else blocks the close.

1. **Strengths**
   - `cmd/sdlc/internal/gatestate/measurement_unattributed.go:27-56`: one predicate `persistedLine` is shared by `Validate`, `BoundLine` and `Normalised`. A grep for `MeasurementUnattributed{` outside tests finds exactly three writers (`actualbranch.go:222`, `actualexclusion.go:464`, `actualinputs.go:558`), and all three go through `.Normalised()`.
   - `measurement_summary_test.go` `TestUnattributedNormalisedAlwaysValidates` runs 2000 seeded adversarial trials (NUL, CR/LF, invalid UTF-8, over-long strings). It goes red when NUL/UTF-8 cleaning is removed from `BoundLine`.
   - `ValidateShape`/`ValidateBounds` are now separate (`measurement.go:225-270`), and an oversize snapshot is turned into a summary instead of refusing the close (`actualmeasurement.go:198-210`).
   - Ledger evolution is additive: older rows have no `Confidence` and read as exact, which matches the old close behaviour (`estimate/ledger.go:93-108`).
   - `actualcorrection.go:739-743`: a correction row carries no confidence columns, so it is only accepted on an exact measurement.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - BR-4 (re-disposed, not new): copy the `errors.As(err, &over) && over.Bytes > SnapshotLimitBytesV4` assertion into `TestMeasurementPastStructuralCapCountedRefusal` (`measurement_branch_v4_test.go:261`). With the fix reverted, that test then fails with `got measurement snapshot is at least 703431 bytes…`. The comment at line 168 also wrongly says the current test covers the structural cap.
   - Plan state table rows (plan:177/179/181) still name `correct-actual` as the recovery for unmeasured rows. The Revisions entry at plan:299 supersedes them, so this is only stale text.

5. **Test coverage notes**
   - **Passing:** `gatestate`, `estimate` and `activetime` all pass. In `cmd/sdlc`, a run filtered to `Confidence|SourceCap|Unattributed|Summar|Adopt` passes (149s). A run filtered to `Actual|Exclusion|Queue|Correction|Measurement` passed 202 tests with no failures before hitting the 540s package timeout.
   - **Timed out:** at the timeout, `TestSuiteQueuePriorityRealHelper` was running. That's a real-helper test and probably just slow under this sandbox, but it didn't finish, so I can't say it passes.
   - **Sandbox failure, unrelated:** `testfix` `TestQuietBackgroundGitReachesTheReceiveSide` fails because the sandbox refuses `git init` outside the reviewer's directories.
   - **Revert checks:** BR-6 goes red when reverted; BR-4 stays green (see above). The checkout was restored clean.

6. **Architecture**
   - **ARCH-DRY: pass.** The single predicate replaces the old `BoundReason` and the separate inline check in `Validate`.
   - **ARCH-PURE: pass.** Normalisation, validation and classification are pure and tested without IO.
   - **ARCH-PURPOSE: pass.** Every unmeasurable status now records unmeasured instead of exiting (`close.go:459-462`), and the fix was swept across all three writers, not just the one the finding named.
   - **ARCH-MOCK: pass.** No new external seams.
   - **ARCH-CONSTRAINTS: pass.** Unattributed records are capped at 32, summary credits are capped, and `Source` is now bounded at 256 bytes.
   - **ARCH-SECURE: pass.** Untrusted transcript session IDs reach `Source` only through `BoundLine`.
   - **ARCH-ORDER: pass.** Close stays single-shot and writes nothing before the evidence commit.
   - **ARCH-FUNERAL: pass.** Growth is bounded by the snapshot and record caps.
   - **ARCH-EVOLVE: pass.** The new ledger columns and the `unattributed` member are additive, and unknown fields are kept through `encodeMeasurementObject`. The stricter `Source` check applies only to a type introduced in this issue, so no older records are affected.
   - **Docs gate:** README, atlas (`sdlc-binary.md`, `issue-lifecycle.md`, `ledger-landscape.md`), help text and the issue vocabulary (`issue.cue`) are all updated.

7. **Plan revisions:** optionally add a Revisions note that rows 177/179/181 of the state table are superseded by the plan:299 entry ("Unmeasured recovery").

```findings
dispose:
  - id: BR-4
    disposition: not-addressed
    note: |
      Fix is correct (probe: Bytes=703431 on the structural-cap path), but the added assertion sits in TestMeasurementOversizeCountedRefusal, which still passes with the fix reverted; move it into TestMeasurementPastStructuralCapCountedRefusal (measurement_branch_v4_test.go:261), which goes red without the fix.
  - id: BR-6
    disposition: addressed
    note: |
      Rule enforced: Validate and BoundLine share persistedLine; all 3 non-test MeasurementUnattributed writers persist via Normalised(), which also bounds Source; TestUnattributedNormalisedAlwaysValidates goes red when NUL/UTF-8 cleaning is reverted.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"cfc357fe-54a3-40bd-b1c5-ca7a6de12866","repository":"github.com/xianxu/ariadne","issue_id":"000603","recipe":"milestone-review","round":4,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/ece5817acbfcebbb","reviewed_head":"f880c83bad3e6618d7fb8f05637ca6e294a744d2","slot":"ariadne:2","requested_at":"2026-10-25T04:44:28.483989Z"},"phase":"finished","pid":8823,"started_at":"2026-10-25T04:44:28.573841Z","finished_at":"2026-10-25T05:17:25.378954Z","session_id":"32156633-0cfb-4682-8bb9-9da85de02d27","outcome":"returned"}] -->

---

## Re-review — 2026-10-24T22:22:42-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 603 — Measured actuals never block close: record exact, partial or unmeasured instead of refusing |
| repo | ariadne |
| issue file | workshop/issues/000603-measured-actuals-never-block-close-record-exact-partial-or-unmeasured-instead-of.md |
| boundary | whole-issue close |
| milestone | — |
| window | 7c2047f2cbc7724104c833b23fb62457fb19c49d..3f77a47a8ac4c70c8191da54bb72504e5041584a |
| command | sdlc close --issue 603 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-24T22:22:42-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

This window contains one change, to `cmd/sdlc/internal/gatestate/measurement_branch_v4_test.go`. No tracker, plan or Spec files changed, so there is no scope change. The fix for BR-4 was already in the code at the pinned head. `ValidateBounds` in `cmd/sdlc/internal/gatestate/measurement.go:247-250` sets `bytes = size.Bytes` when the encoder stops at its structural cap. Every `OversizeError` it returns carries that size. `summariseMeasurement` then copies `over.Bytes` into `Summary.Bytes` (`cmd/sdlc/actualmeasurement.go:268`), so it no longer records a made-up 0. The window adds the regression test that was missing. I ran it in the disposable checkout. The suite passed at head. I then disabled `bytes = size.Bytes`, and `TestMeasurementPastStructuralCapCountedRefusal` failed at `measurement_branch_v4_test.go:276`. Restoring the line made it pass again. `go vet` and `gofmt -l` report nothing.

1. **Strengths**
   - The new assertion at `measurement_branch_v4_test.go:272-277` covers the exact path BR-4 named: the encoder stops past its structural cap at 703431 bytes, above the 131072 bound. I confirmed it fails when the fix is removed.
   - The comment on the counted-size test (`:168`) was narrowed so it no longer claims to cover the structural-cap case. The cap case now has its own test.
   - All three `OversizeError` returns in `ValidateBounds` share one `bytes` value, so the size is set in one place for every caller, not only the cited one.

2. **Critical:** none.
3. **Important:** none.
4. **Minor:** The new test pins the value in gatestate. It does not pin `Summary.Bytes` in `summariseMeasurement` itself. That copy is a direct field assignment, so I accept the risk.
5. **Test coverage:** Removing the fix turns the new test red, which meets ariadne#194. No other tests changed.
6. **Architecture** (interdiff only)
   - ARCH-DRY: pass. One `bytes` value feeds every return.
   - ARCH-PURE: pass. Pure validation with a pure test, no IO.
   - ARCH-PURPOSE: pass. The fix covers the whole class of oversize paths, not just one site.
   - ARCH-EVOLVE: pass. No persisted format changed.
7. **Plan revisions:** none.

```findings
dispose:
  - id: BR-4
    disposition: addressed
    note: |
      ValidateBounds takes bytes from size.Bytes on the structural-cap path (measurement.go:247-250) and summariseMeasurement copies over.Bytes into Summary.Bytes (actualmeasurement.go:268). The new test at measurement_branch_v4_test.go:272-277 fails when that line is disabled (measured: at least 703431 bytes) and passes when it is restored.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"799a3ac6-7360-4a04-a1f7-3efb61631361","repository":"github.com/xianxu/ariadne","issue_id":"000603","recipe":"fix-review","round":5,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/29492748be80d90b","reviewed_head":"3f77a47a8ac4c70c8191da54bb72504e5041584a","slot":"ariadne:2","requested_at":"2026-10-25T05:21:53.228655Z"},"phase":"finished","pid":93374,"started_at":"2026-10-25T05:21:53.366989Z","finished_at":"2026-10-25T05:22:41.216657Z","session_id":"522039a6-6250-48fe-9069-8a5cd81e15e0","outcome":"returned"}] -->

---

## Re-review — 2026-10-25T04:00:22-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 603 — Measured actuals never block close: record exact, partial or unmeasured instead of refusing |
| repo | ariadne |
| issue file | workshop/issues/000603-measured-actuals-never-block-close-record-exact-partial-or-unmeasured-instead-of.md |
| boundary | whole-issue close |
| milestone | — |
| window | d44120397014f3cb407aeda803ea37b141626dad..9e84f7def5418b0b4144f2e98609b3c7f8320f0f |
| command | sdlc close --issue 603 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-25T04:00:22-07:00 |
| verdict | SHIP |

## Review

Checking the guard's nil handling, then running the new test with and without the fix.

Now I'll revert the guard and confirm the test fails.

```verdict
verdict: SHIP
confidence: high
```

This window integrates main (#617 adoption guard, #620) into #603. It resolves one conflict in `cmd/sdlc/close.go` and makes one behavioral change: the #617 adoption guard now applies only to measured closes. The resolution keeps both sides. The struct keeps #617's `Adoptions` and `effects` alongside #603's `Confidence`. `adoptOmittedActual` assigns both `f.Confidence` and `f.Adoptions` before `resolveOmittedActual`, so main's intent holds: the read-set is captured even when measurement fails. The duplicate post-resolve assignment is gone. No conflict markers remain in `cmd/` or `internal/`, the changed files are gofmt-clean, and the package builds and passes `go vet`. The guard scoping is not scope creep. Without it, the two features combined would refuse any close whose measurement failed while loading exclusions. #617's `compareAdoptionReadSet` returns "read-set is unavailable" for a non-nil read-set with nil cards, and that contradicts #603's purpose that measurement never blocks a close. Nothing in the issue file's `## Spec`, `## Revisions` or `## Done when` changed. The only edits there are the status flip, `actual_hours` and Log lines.

1. **Strengths**
   - `cmd/sdlc/closeprepare.go:123-127`: the guard condition uses the existing `f.unmeasured()` (`close.go:1838`). That is exactly `Status != actualMeasured`, the same condition under which `resolveOmittedActual` adopts nothing. An unmeasured close therefore never carries authority it lacks, and a measured close keeps full revalidation.
   - `cmd/sdlc/close_confidence_test.go:144-170`: the regression test runs the real command path (capture, then flags, then the prepare guard) for both `close` and `milestone-close`. That is the command-wiring proof the #617 BR-1 lesson asks for.
   - I reverted the guard to an unconditional `guard.adopted = f.Adoptions` and the test went red (`FAIL`). With the guard restored it passes, and so do the targeted `Confidence` and `Adoption` tests (`ok … 99.460s`).
   - `workshop/lessons.md`: the #603 lesson was moved to sit with the other #603 entries, not duplicated. The text is unchanged.

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings:** none.

5. **Test coverage notes:** The test covers the unavailable read-set (`&adoptionReadSet{}`, nil cards) for both verbs. It does not cover a measured close whose adoptions change during review. That is #617's existing coverage, and this window does not touch the measured branch.

6. **Architectural notes**
   - **ARCH-DRY: pass.** The guard reuses the existing `unmeasured()` predicate, with no parallel condition.
   - **ARCH-PURE: pass.** The change is one conditional in the prepare shell, and the comparison logic stays in `compareAdoptionReadSet`.
   - **ARCH-PURPOSE: pass.** It keeps #603's never-block purpose across the new #617 consumer. The only place the guard reads the read-set is `closeprepare.go:80`, and that is now gated.
   - **ARCH-EVOLVE: pass.** No persisted format changes. `Adoptions` is transient and never serialized.

7. **Plan revision recommendations:** none.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Disposed in an earlier round; nothing in this interdiff reopens it.
  - id: BR-2
    disposition: addressed
    note: |
      Disposed in an earlier round; this interdiff does not touch CoalesceCredits.
  - id: BR-3
    disposition: addressed
    note: |
      Disposed in an earlier round; this interdiff does not touch the unattributed writer.
  - id: BR-4
    disposition: addressed
    note: |
      Disposed in an earlier round; the Log records the OversizeError size assertion and a mutation check.
  - id: BR-5
    disposition: addressed
    note: |
      Disposed in an earlier round; no plan Task text changes in this interdiff.
  - id: BR-6
    disposition: addressed
    note: |
      Disposed in an earlier round; persistedLine is unchanged here.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"fd67477f-d1cb-4c72-80f4-b3edd3afef7e","repository":"github.com/xianxu/ariadne","issue_id":"000603","recipe":"fix-review","round":6,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/9713b66009b62b84","reviewed_head":"9e84f7def5418b0b4144f2e98609b3c7f8320f0f","slot":"ariadne:2","requested_at":"2026-10-25T10:57:39.870773Z"},"phase":"finished","pid":34531,"started_at":"2026-10-25T10:57:39.961483Z","finished_at":"2026-10-25T11:00:20.849172Z","session_id":"5a304653-f9cc-4dd1-aaed-61277d245763","outcome":"returned"}] -->
