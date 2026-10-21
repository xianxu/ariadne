# Boundary Review — ariadne#524 (milestone M2)

| field | value |
|-------|-------|
| issue | 524 — Scoped future evidence and exact-package flake reruns |
| repo | ariadne |
| issue file | workshop/issues/000524-scoped-evidence-reruns.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 0783e8acd98577a05be2d112ccd65aa5003770b0..1ff8b62a01066df5152816acc8c2d2b6880bd005 |
| command | sdlc milestone-close --issue 524 --milestone M2 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-21T07:47:37-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

Exact-package selection, shared resource ownership, and documentation are implemented well. However, the new PASS collector rejects legitimate successful Go reruns in two reproducible cases. Both reach the production gate and produce `known-flake-rerun-inconclusive`, blocking M2.

```findings
findings:
  - id: new
    severity: Critical
    kind: defect
    family: external-event-contract-fidelity
    title: |
      PASS collection rejects valid Go subtest events and invents required ancestors
    detail: |
      cmd/sdlc/internal/flakes/rerun.go:104-108 validates every terminal Test value using the narrower registry-name grammar: a passing t.Run("case(value)") invalidates the entire log, even when only its valid top-level test requires proof. Lines 124-132 also require a PASS event for every slash prefix, but t.Run("path/file") emits TestSlash/path/file and TestSlash, with no TestSlash/path event. Real go test -json runs exited zero in both cases; temporary overlays extending TestKnownFlakeExactPackageGate and TestKnownFlakeGateSubtest reproduced production outcomes of known-flake-rerun-inconclusive. ARCH-PURPOSE/ARCH-MOCK: distinguish registry eligibility from valid tool output, and require actual test outcomes without fabricating ancestors from name separators. Sweep legal subtest-name forms and add real-Go regressions for both cases while preserving package binding, conflict rejection, and sibling-failure blocking.
```

1. **Strengths**

   - `RerunCommands` produces deterministic, anchored, package-specific argv. The real two-package decoy test passes.
   - The batch retains one admission, lease, scratch directory, and aggregate deadline. Failure and controlled cancellation tests pass.
   - Runner cleanup and environment-recovery regressions remain green.
   - README, helptext, and atlas document the new execution surface and private conformance command.

2. **Critical findings**

   The finding above covers two incorrect assumptions in [rerun.go](/Volumes/TestScratch/review-checkouts/94256da7abb451c1/cmd/sdlc/internal/flakes/rerun.go:104). Valid incidental subtest names must not invalidate required proof, and slash-separated names do not establish actual `t.Run` ancestry.

   Fix the collector’s event model and retain regressions using real Go output. The existing helper-generated events synthesize every slash-prefix ancestor, masking the second defect.

3. **Important findings**

   None additional.

4. **Minor findings**

   None.

5. **Test coverage notes**

   Independently passed:

   - Entire `internal/flakes` package.
   - Focused gate, affected-flake, batch, environment, and runner-cleanup checks: 39.219s.
   - Opt-in production CLI conformance: 14.953s.
   - Focused race checks: flakes 1.160s; sdlc 9.205s.
   - Pinned-range whitespace check.

   Additional real-Go proof tests and both production gate fixture extensions failed as described above. Temporary overlays were removed; the repository remains unchanged. The central recorded suite remains pending in the supplied tracker and was not independently confirmed.

6. **Architectural notes**

   | Marker | Result |
   |---|---|
   | ARCH-DRY | Pass — shared selection and managed execution paths. |
   | ARCH-PURE | Pass — M2’s listed PURE entities exist and their unit tests require no IO. |
   | ARCH-PURPOSE | **Flag** — legitimate whole-parent/subtest reruns lose passing authority. |
   | ARCH-MOCK | **Flag** — synthetic event fixtures encode assumptions contradicted by actual Go output. |
   | ARCH-CONSTRAINTS | Pass — aggregate timeout and bounded production selection retained. |
   | ARCH-SECURE | Pass — literal argv and quoted diagnostics preserve input separation. |
   | ARCH-ORDER | Pass — tested failure/cancellation paths stop subsequent launches and join descendants. |
   | ARCH-FUNERAL | Pass — existing artifact lifecycle retained; conformance fixtures clean automatically. |
   | ARCH-EVOLVE | Pass — no new persisted format or rewrite path introduced. |

7. **Plan revision recommendations**

   Add a `## Revisions` entry defining PASS authority against actual Go events: distinguish registry-name restrictions from emitted names, and distinguish slash prefixes from real test ancestry. Update the ancestor claims in the plan, helptext, and atlas accordingly. Retain the central-suite/delivery checklist item until its evidence is recorded.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"d26bc2cc-ad3e-42d2-91cd-7632fe3243be","repository":"github.com/xianxu/ariadne","issue_id":"000524","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/94256da7abb451c1","reviewed_head":"1ff8b62a01066df5152816acc8c2d2b6880bd005","slot":"ariadne:3","requested_at":"2026-10-21T14:43:52.827599Z"},"phase":"finished","pid":19966,"started_at":"2026-10-21T14:43:52.933755Z","finished_at":"2026-10-21T14:47:37.188881Z","session_id":"01a15eeb-2cf5-7ab1-ab83-2cde8a2f21f9","outcome":"returned"}] -->

---

## Re-review — 2026-10-21T07:59:25-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 524 — Scoped future evidence and exact-package flake reruns |
| repo | ariadne |
| issue file | workshop/issues/000524-scoped-evidence-reruns.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 0783e8acd98577a05be2d112ccd65aa5003770b0..2750b1ed88b29c30a74bb781578b2ba59150f7c1 |
| command | sdlc milestone-close --issue 524 --milestone M2 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-21T07:59:25-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned M2 implementation matches the documented execution contract. BR-5 is addressed with independently verified regressions, and no new findings arose. README, helptext, and atlas cover the new surface. The central recorded suite and final delivery remain explicitly pending; this review does not certify those steps.

```findings
dispose:
  - id: BR-5
    disposition: addressed
    note: |
      Real-Go TestKnownFlakeExactPackageGate and TestKnownFlakeGateFlatSlashSubtest pass on the pinned head. Restoring the pre-fix collector in a disposable checkout makes both fail with known-flake-rerun-inconclusive; the pure regression also fails. Package binding, required-event conflicts, and sibling-failure controls remain green. The disposable checkout was removed.
```

1. **Strengths**

   - `cmd/sdlc/internal/flakes/rerun.go:17` produces deterministic, anchored commands for individual packages; real decoy fixtures verify selection.
   - `cmd/sdlc/internal/flakes/rerun.go:105` separates emitted Go identities from registry eligibility and requires actual leaf, top-level, and package outcomes.
   - `cmd/sdlc/localcheck_run.go:318` retains shared admission, lease, scratch, and aggregate deadline. Failure and controlled cancellation stop subsequent launches.
   - Private CLI conformance verifies production wiring, scoped future evidence, publication effects, and original/child attempt linkage.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   Independently passed:

   - Entire flakes package, normal and race.
   - Focused flake gates, real subtests, managed batches, and failure attribution: 26.489s.
   - Private production-CLI conformance: 11.419s.
   - Runner, environment, queue, and scoped-proof regressions: 13.459s.
   - Affected-flake controls: 3.502s.
   - Focused real-Go/batch race tests: 9.698s.
   - Pinned-range whitespace check.

   Two supplemental selections timed out in unchanged `TestLocalCheckCancelOnlyOwnRun` cleanup at `localcheck_lease_test.go:251`, which waits again on consumed completion channels. Those runs are not counted as passing. The new batch cancellation test passed normally and under race detection. The reviewed checkout remains unchanged.

6. **Architectural notes**

   | Marker | Result |
   |---|---|
   | ARCH-DRY | Pass — one production batch builder and shared execution path; obsolete helpers removed. |
   | ARCH-PURE | Pass — M2 table entities exist and their unit tests use supplied values without IO. |
   | ARCH-PURPOSE | Pass — suite and affected paths use exact-package batches; BR-5’s event-contract class is covered. |
   | ARCH-MOCK | Pass — private stateful publisher fixtures and real-Go conformance exercise production seams. |
   | ARCH-CONSTRAINTS | Pass — bounded selection, sequential launches, and aggregate timeout retained. |
   | ARCH-SECURE | Pass — literal argv, separate diagnostics, and strict interpreted JSON fields preserve proof boundaries. |
   | ARCH-ORDER | Pass — existing queue authority retained; failure/cancellation sequencing and descendant cleanup tested. |
   | ARCH-FUNERAL | Pass — existing attempt/log retention retained; temporary conformance artifacts clean automatically. |
   | ARCH-EVOLVE | Pass — no new persisted format or rewrite path; opaque future bytes remain unchanged. |

7. **Plan revision recommendations**

   None. The BR-5 revision accurately describes actual Go events. Keep the central-suite and delivery checklist pending until their evidence is recorded.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"95170fca-1f42-4faa-904f-b382dfd98888","repository":"github.com/xianxu/ariadne","issue_id":"000524","recipe":"milestone-review","round":2,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/d12f61852ea4d6be","reviewed_head":"2750b1ed88b29c30a74bb781578b2ba59150f7c1","slot":"ariadne:3","requested_at":"2026-10-21T14:53:04.523798Z"},"phase":"finished","pid":50917,"started_at":"2026-10-21T14:53:04.632039Z","finished_at":"2026-10-21T14:58:52.410133Z","session_id":"01a15ef3-9826-78f1-8df0-9996f5cc3104","outcome":"returned"}] -->
