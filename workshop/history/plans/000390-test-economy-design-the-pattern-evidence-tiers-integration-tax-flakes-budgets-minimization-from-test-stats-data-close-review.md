# Boundary Review — ariadne#390 (whole-issue close)

| field | value |
|-------|-------|
| issue | 390 — test-economy: design the pattern (evidence tiers, integration tax, flakes, budgets, minimization) from test-stats data |
| repo | ariadne |
| issue file | workshop/issues/000390-test-economy-design-the-pattern-evidence-tiers-integration-tax-flakes-budgets-minimization-from-test-stats-data.md |
| boundary | whole-issue close |
| milestone | — |
| window | 09ead92ed893c343fc81bdc9d7025f96aace862d..453d71f3d0b29eb6761fe198b88c6cd409fbf4e1 |
| command | sdlc close --issue 390 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-15T14:47:47-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The design covers all five requested areas, and the published implementation cards and ops project linkage are present. No executable code or gates change in this window. One classification rule needs correction before implementation: failure on main alone does not establish a flake.

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: failure-classification-requires-evidence
    title: |
      Main failure alone must not establish flake eligibility
    detail: |
      workshop/plans/000390-test-economy-pattern-plan.side/02-flakes-and-landing.md:19 classifies a failure as a new flake when it “fails on main too,” then makes it eligible for the known-flake rerun rule. A deterministic regression already present on main satisfies that predicate without establishing intermittency. Preserve persistent-main and suspected faults as separate classifications; require evidence of intermittency before granting flake eligibility. Family sweep: this is the only explicit offending predicate; reconcile it with the parent plan’s section 3 (lines 82–89), its classification revision (line 193), and #394’s acceptance criterion distinguishing persistent main failure. ARCH-PURPOSE: the classification must preserve the design’s trustworthy-evidence goal.
```

1. **Strengths**
   - The baseline reports incomplete sample coverage and separates CI failures from confirmed same-head flakes.
   - Exact-head landing, drift invalidation and uncertain-publication recovery are explicit.
   - The dated revisions clearly withdraw the premature 600/240-second budget policy.
   - Verified #392–#398 and pair#578 in published Git refs, and their entries in the ops project.

2. **Critical findings:** None.

3. **Important findings:** The classifier inconsistency above. Existing `cmd/sdlc/teststats.go` distinguishes `suspected-main-baseline` from `confirmed`, supporting the required distinction.

4. **Minor findings:** None.

5. **Test coverage notes:** No runtime changes require new regression tests here. Inspected existing aggregation tests and follow-on acceptance criteria. `git diff --check` passed. No suite was run for this design-only range.

6. **Architectural notes**
   - **ARCH-DRY — pass:** Reuses existing attempt records, suite locks and Git/GitHub seams.
   - **ARCH-PURE — pass:** Proposed policy entities are separated from IO adapters; they are explicitly future entities.
   - **ARCH-PURPOSE — flag:** Correct the failure-classification rule above. Otherwise, the design and filed follow-ons cover the issue’s purpose.
   - Atlas and README updates are appropriately assigned to implementation; this range introduces no runnable surface.

7. **Plan revision recommendation:** Append a dated `## Revisions` entry clarifying that failure on main establishes neither intermittency nor known-flake eligibility, and reconcile side doc 02 with that rule.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"e5ffad9c-0ee4-44a3-b6d6-83bd73f52b22","repository":"github.com/xianxu/ariadne","issue_id":"000390","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/034e09d57b770819","reviewed_head":"453d71f3d0b29eb6761fe198b88c6cd409fbf4e1","slot":"ariadne:1","requested_at":"2026-10-15T21:46:15.058939Z"},"phase":"finished","pid":40213,"started_at":"2026-10-15T21:46:15.693268Z","finished_at":"2026-10-15T21:47:45.380634Z","session_id":"01a14187-b99a-7322-b91a-e4dc7f599343","outcome":"returned"}] -->

---

## Re-review — 2026-10-15T14:50:57-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 390 — test-economy: design the pattern (evidence tiers, integration tax, flakes, budgets, minimization) from test-stats data |
| repo | ariadne |
| issue file | workshop/issues/000390-test-economy-design-the-pattern-evidence-tiers-integration-tax-flakes-budgets-minimization-from-test-stats-data.md |
| boundary | whole-issue close |
| milestone | — |
| window | 09ead92ed893c343fc81bdc9d7025f96aace862d..bd83e139d3b606b79372a3719f5ee61f84a06088 |
| command | sdlc close --issue 390 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-15T14:50:57-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The design covers the five requested concerns, records the operator’s decisions, and has published implementation cards linked from the ops project. BR-1 is addressed consistently across the parent plan and side document. No executable behavior changes in this range.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Inspected the pinned before/after correction: side doc 02:19–22 and parent plan:88 now require same-exact-head fail→pass evidence and exclude persistent-main faults; plan:213 records the governing revision. Existing cmd/sdlc/teststats.go:136–143 distinguishes suspected-main-baseline from confirmed flakes, and published #394 requires persistent-main classification coverage. This corrects a design record; executable implementation remains assigned to #394.
```

Done-when: current — the contract still requires an operator-reviewed design, filed implementation issues, and ops project linkage; all are present.

1. **Strengths**
   - Baseline data includes sample coverage and separates CI failures from confirmed flakes.
   - Landing design specifies drift invalidation, crash recovery, and uncertain-publication handling.
   - Dated revisions explicitly withdraw premature budget values.
   - Verified published #392–#398, pair#578, tracker cards, and ops project links.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes:** Inspected existing aggregation tests and follow-on acceptance criteria. No suite run for this design-only range. `git diff --check` passed; checkout remains unchanged.

6. **Architectural notes**
   - **ARCH-DRY — pass:** Reuses attempt records, weighted suite locks, and existing Git/GitHub seams.
   - **ARCH-PURE — pass:** Proposed policy entities are separated from IO adapters and explicitly identified as future implementation.
   - **ARCH-PURPOSE — pass:** Design, implementation ownership, and project linkage fulfill this issue’s scope.
   - Atlas and README updates belong to implementation: this window adds no current command, configuration, or architectural runtime surface.

7. **Plan revision recommendations:** None.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"92e08c70-7583-4d85-8207-fc444a8e535d","repository":"github.com/xianxu/ariadne","issue_id":"000390","recipe":"small-diff-review","round":2,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/1e4e233cb539fdf6","reviewed_head":"bd83e139d3b606b79372a3719f5ee61f84a06088","slot":"ariadne:1","requested_at":"2026-10-15T21:49:50.295698Z"},"phase":"finished","pid":34510,"started_at":"2026-10-15T21:49:50.613872Z","finished_at":"2026-10-15T21:50:55.554569Z","session_id":"01a1418b-0030-7271-96d2-a7d59b506ee9","outcome":"returned"}] -->
