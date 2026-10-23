# Boundary Review — ariadne#574 (whole-issue close)

| field | value |
|-------|-------|
| issue | 574 — Keep slot-sibling ariadne checkouts current so queue-helper rollouts finish |
| repo | ariadne |
| issue file | workshop/issues/000574-keep-slot-sibling-ariadne-checkouts-current-so-queue-helper-rollouts-finish.md |
| boundary | whole-issue close |
| milestone | — |
| window | db77404e3257507f41c9d6fc3d819a8723bb36fa..a492fb052f9d058acef4f8eeabc41f62c8003482 |
| command | sdlc close --issue 574 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-23T09:37:48-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The focused tests passed, and the implementation preserves checkout ownership, preflight ordering, cancellation, and honest reporting of uncertain updates. Two Important gaps remain: the new flag lacks the required README documentation, and several supplier-provenance controls marked complete in the plan have no corresponding regression tests.

```findings
findings:
  - id: new
    severity: Important
    family: user-surface-documentation
    title: |
      README update appears missing for weave refresh --suite-helper
    detail: |
      cmd/weave/refresh.go:95 introduces a user-facing flag, but README.md is unchanged in the pinned range. Add the flag, automatic suite-setup boundary, eligibility limits, and skip/recovery behavior to README.md's existing refresh documentation around line 405.
  - id: new
    severity: Important
    family: planned-behavior-regression-coverage
    title: |
      Supplier provenance tests do not deliver the controls marked complete in the plan
    detail: |
      workshop/plans/000574-sibling-helper-refresh-plan.md:82 marks multiple revisions, fast-forward after ticket launch, missing worktree, symlink/identity mismatch, and Git failure controls complete. cmd/sdlc/queue_suite_supplier_test.go exercises one native dirty checkout plus injected constant observations; it does not exercise those production resolver/read paths. Add private fixtures covering distinct real supplier revisions, a supplier advancing while its legacy ticket remains unchanged, and unavailable results for missing or redirected owners and failed Git observations.
```

1. **Strengths**

   - Refresh runs before helper preflight through the shared launch seam, including saved resumes; runner checks and proof-only reposts are excluded.
   - `cmd/weave/internal/refresh/helper.go:76` preserves completed progress and reports uncertain effects without rollback.
   - Supplier diagnostics explicitly distinguish current checkout revision from unknown launch revision.
   - Request-local caches bound repeated owner resolution and supplier reads; counted tests cover repeated and distinct suppliers.

2. **Critical findings**

   None.

3. **Important findings**

   The two findings above require README documentation and completion of the promised provenance regression controls. The latter should exercise `resolveQueueSupplier` and `readQueueSupplier`, rather than returning prepared observations through their injection seam.

4. **Minor findings**

   None.

5. **Test coverage notes**

   Passed the focused tests across `cmd/weave/internal/refresh`, `cmd/weave`, `cmd/sdlc`, and `pkg/workspace`. Coverage includes eligibility, deterministic mutations before update, cancellation before/after effects, ignored-file protection, executable fallback, resume exclusions, and probe deduplication. No prior findings required disposition. The checkout remained unchanged.

6. **Architectural notes**

   - **ARCH-DRY — Pass:** shared workflow/helper resolution and existing refresh safety helpers are reused.
   - **ARCH-PURE — Pass:** eligibility and diagnostic formatting remain deterministic; IO has identifiable seams.
   - **ARCH-PURPOSE — Flag:** the implementation addresses both requested behaviors, but the plan overstates delivered verification.
   - **ARCH-MOCK — Pass:** private Git repositories and injected operation boundaries provide stateful, reproducible failure controls.
   - **ARCH-CONSTRAINTS — Pass:** optional refresh has a deadline; status deduplicates probes and inherits the status deadline.
   - **ARCH-SECURE — Pass:** exact checkout checks, bounded Git output, and conservative unavailable results protect provenance.
   - **ARCH-ORDER — Pass:** sequential revalidation and confirmed/uncertain outcomes preserve partial progress; tests inject ordering around Git effects.
   - **ARCH-FUNERAL — Pass:** no new growing artifact family or background updater; the existing setup lease is reused.
   - **ARCH-EVOLVE — Pass:** ticket storage and frozen JSON remain unchanged; new observations are transient.

   Core-concept entities exist at the declared locations. The pure decision and formatting responsibilities are distinguishable from their integration callers.

7. **Plan revision recommendations**

   Add `## Revisions` entries recording the README documentation requirement and the completed provenance-test matrix. Until those tests exist, amend the checked verification item at line 82 to reflect the narrower evidence actually delivered.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"95463c2e-f4cd-4c59-98c8-01e947963491","repository":"github.com/xianxu/ariadne","issue_id":"000574","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/8000f35af00287d4","reviewed_head":"a492fb052f9d058acef4f8eeabc41f62c8003482","slot":"ariadne:3","requested_at":"2026-10-23T16:35:27.002991Z"},"phase":"finished","pid":44319,"started_at":"2026-10-23T16:35:27.180476Z","finished_at":"2026-10-23T16:37:47.399515Z","session_id":"01a1699e-0be8-7082-ba69-ec46c70c3b3e","outcome":"returned"}] -->
