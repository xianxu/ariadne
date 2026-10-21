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
