# Boundary Review — ariadne#524 (milestone M1)

| field | value |
|-------|-------|
| issue | 524 — Scoped future evidence and exact-package flake reruns |
| repo | ariadne |
| issue file | workshop/issues/000524-scoped-evidence-reruns.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 63ebb3385af267d64084921c6e8d1150a188c734..be03c359a52d9f73806509cb326155edd7e70967 |
| command | sdlc milestone-close --issue 524 --milestone M1 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-21T05:10:29-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

M1 introduces a bounded, scoped evidence reader and documents the fleet rollout barrier well. Focused and broader targeted tests pass. However, implicit saved-proof selection can bypass unsupported evidence linked to the selected parent when its check differs. The checked-off pre-change reader/guard control is also missing.

1. **Strengths**

   - Unsupported records retain opaque bytes and cannot become accepted run measurements; duplicate keys, malformed identities and filename mismatches are rejected.
   - Production store readers now consume `SuiteEvidenceSnapshot`; legacy tuple projections are confined to tests.
   - Indexed authority lookups avoid repeated filesystem scans. The N/2N tests verify single opens and IO-free projection.
   - The atlas explicitly documents stale-reader behavior, fleet binary verification and the prohibition on activating the later writer.

```findings
findings:
  - id: new
    severity: Critical
    kind: defect
    family: linked-proof-authority
    title: |
      Implicit saved-proof selection bypasses unsupported evidence linked to its parent
    detail: |
      cmd/sdlc/localcheck_republish.go:73 checks unsupported authority before selecting a supported attempt, using an empty ID for implicit selection. A decoder-accepted future record with the same repository/head, check "vm", and rerun_of naming a supported suite parent is therefore missed: proofError only consults byParent when an ID is supplied (teststats_evidence.go:72-78), and the suite subject lookup excludes "vm". loadSavedAttempt returns the parent as newest, and guardSuitePass repeats the ID-less lookup, allowing success publication despite linked uncertainty. Check the selected candidate's ID before returning it and make the success guard enforce linked-parent authority, or reject inconsistent parent/check identities as unscopable. Add a production-consumer regression covering explicit and implicit selection of this fixture. ARCH-SECURE / ARCH-PURPOSE.
  - id: new
    severity: Important
    kind: defect
    family: compatibility-control-evidence
    title: |
      The completed M1 old-reader control is not retained
    detail: |
      workshop/plans/000524-scoped-evidence-reruns-plan.md:90 marks a pre-change decoder/guard control complete, but teststats_evidence_test.go:65-71 only checks decodeSuiteAttempt rejection before invoking the upgraded consumers. Decoder rejection does not prove the old reader's global warning, saved-proof incompleteness or unrelated success refusal. Retain a reproducible control against the pinned pre-change reader/guard and compare it with the upgraded reader on equivalent mixed-store fixtures; record the commands and results before marking this requirement complete. ARCH-PURPOSE.
```

2. **Critical findings**

   - **Linked-parent bypass:** [localcheck_republish.go:73](/Volumes/TestScratch/review-checkouts/9ea6798693905b79/cmd/sdlc/localcheck_republish.go:73). The existing relation test already expects a cross-check link to block an explicitly identified parent, but actual consumer tests never combine that link with implicit selection. Enforce the same rule after selecting the candidate and through publication.

3. **Important findings**

   - **Missing compatibility control:** [plan:90](/Volumes/TestScratch/review-checkouts/9ea6798693905b79/workshop/plans/000524-scoped-evidence-reruns-plan.md:90). Preserve executable evidence of the old whole-store behavior, alongside the upgraded behavior.

4. **Minor findings**

   None.

5. **Test coverage notes**

   - Focused new-reader tests passed in **2.890s**.
   - Broader targeted store, statistics, saved-proof, republish, flake, actuals and weekly tests passed in **51.876s**.
   - `git diff --check` passed.
   - No full suite or mutation tests were run. The checkout remains unchanged.
   - Add the linked-parent regression through `loadSavedAttempt`, `republishSavedPass` and status publication, asserting refusal and zero success posts.

6. **Architectural notes**

   | Principle | Assessment |
   |---|---|
   | ARCH-DRY | Pass: shared snapshot and indexed authority projection. |
   | ARCH-PURE | Pass: identity decoding and relation decisions are tested directly without IO. |
   | ARCH-PURPOSE | **Flag:** linked-parent consumer coverage and the promised old-reader control are incomplete. |
   | ARCH-MOCK | Pass: no new external dependency; integration fixtures use existing seams and private resources. |
   | ARCH-CONSTRAINTS | Pass: existing record/byte budgets remain, with counted scan tests. |
   | ARCH-SECURE | **Flag:** validated linked identity loses authority during implicit selection. |
   | ARCH-ORDER | Pass for M1: snapshots hold no state between external events; managed execution changes remain M2 work. |
   | ARCH-FUNERAL | Pass: no new durable artifact family or increased persisted record size. |
   | ARCH-EVOLVE | Pass for the changed reader: future bytes remain opaque and untouched; no new rewrite path or version is introduced. |

   The atlas gate is satisfied. M1 adds no command, flag or configuration key requiring a README update.

7. **Plan revision recommendations**

   Add a `## Revisions` entry specifying that linked-parent authority must survive both explicit and implicit selection, including inconsistent parent/check fixtures. Record the retained pre-change control and its measured comparison. Add explicit PURE/INTEGRATION and milestone columns to the core-concepts table so current M1 entities are distinguishable from planned M2 entities.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"bb628093-9896-4c9a-bff2-484e823157c2","repository":"github.com/xianxu/ariadne","issue_id":"000524","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/9ea6798693905b79","reviewed_head":"be03c359a52d9f73806509cb326155edd7e70967","slot":"ariadne:3","requested_at":"2026-10-21T12:07:53.965993Z"},"phase":"finished","pid":75537,"started_at":"2026-10-21T12:07:54.11021Z","finished_at":"2026-10-21T12:10:28.998035Z","session_id":"01a15e5c-5f47-7f20-8abb-b2547dabed9f","outcome":"returned"}] -->

---

## Re-review — 2026-10-21T05:18:19-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 524 — Scoped future evidence and exact-package flake reruns |
| repo | ariadne |
| issue file | workshop/issues/000524-scoped-evidence-reruns.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 63ebb3385af267d64084921c6e8d1150a188c734..744236f93156b5662e61fffdd8bfbad206e33b96 |
| command | sdlc milestone-close --issue 524 --milestone M1 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-21T05:18:19-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

M1’s scoped reader behavior and both prior corrections are verified. Unrelated proof remains usable, linked unsupported evidence blocks publication, and the retained old-reader control demonstrates the rollout requirement. No Critical defects remain; the README gate needs a small correction.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      TestScopedSavedProofConsumers covers explicit/implicit selection, reuse and publication. Removing the parent-index invocation in a disposable pinned-head clone made the cross-check regression fail with unauthorized publication; the unchanged head passes.
  - id: BR-2
    disposition: addressed
    note: |
      Ran sh scripts/verify-pre524-reader.sh successfully against pinned pre-change source. Its retained test asserts the global warning, incomplete saved proof and guard/publication refusal; the equivalent upgraded-reader cases pass.
findings:
  - id: new
    severity: Important
    kind: defect
    family: user-surface-documentation
    title: |
      README update appears missing for the compatibility verification command
    detail: |
      scripts/verify-pre524-reader.sh:1 introduces an operator-run verification command, documented at atlas/workflow/sdlc-binary.md:2646, but README.md has no change in the pinned range. Add a brief README usage entry or link covering the command and reader-first rollout barrier to satisfy the explicit README gate.
  - id: new
    severity: Minor
    kind: defect
    family: artifact-lifecycle
    title: |
      Compatibility-control clones have no defined cleanup lifecycle
    detail: |
      scripts/verify-pre524-reader.sh:7 creates a fresh clone on every invocation and retains it on success or failure without a cleanup path. ARCH-FUNERAL: define retention through inspection and its removal owner/command, or clean by default with an explicit retain option; update the plan's claim that this work creates no new artifact family.
```

1. **Strengths**

   - Strict envelope validation rejects ambiguous identities while preserving opaque bytes (`cmd/sdlc/teststats_evidence.go:136`).
   - Parent indexing closes the production selection/publication bypass without repeated store scans (`cmd/sdlc/teststats_evidence.go:54`).
   - Legacy tuple readers are test-only; production consumers use the scoped snapshot.
   - Atlas explicitly documents stale-reader behavior and fleet verification before writer activation (`atlas/workflow/sdlc-binary.md:2637`).

2. **Critical findings**

   None.

3. **Important findings**

   Add README discovery for the compatibility verification command, as recorded above.

4. **Minor findings**

   Define the retained verification clone’s cleanup lifecycle.

5. **Test coverage notes**

   - Focused reader, consumer, statistics and evidence tests passed.
   - Focused race tests, including the actuals and statistics regressions, passed.
   - The real pre-change compatibility control passed.
   - Removing parent indexing made the BR-1 regression fail, confirming meaningful regression coverage.
   - `git diff --check` passed. No full central suite was run during this review.
   - Disposable verification checkouts were removed; the reviewed checkout remains clean.

6. **Architectural notes for upcoming work**

   - **ARCH-DRY — pass:** shared snapshot projections serve the audited production readers.
   - **ARCH-PURE — pass:** identity and authority projections operate on supplied values; filesystem access stays in the store shell. M1 Core concepts match the implementation; M2 entities remain explicitly planned.
   - **ARCH-PURPOSE — pass:** M1 delivers scoped compatibility and reproducible old-reader evidence; rerun execution remains M2.
   - **ARCH-MOCK — pass:** no new external dependency; production-consumer tests use existing private Git/store and fake GitHub seams.
   - **ARCH-CONSTRAINTS — pass:** bounded reads remain enforced, with budget and counted-scan tests.
   - **ARCH-SECURE — pass:** malformed identity remains visible uncertainty; unsupported bodies never become passing facts.
   - **ARCH-ORDER — pass:** M1 introduces no cross-event execution state; snapshot construction completes before authority consumption.
   - **ARCH-FUNERAL — flag:** verification-clone retention needs the lifecycle noted above.
   - **ARCH-EVOLVE — pass:** unsupported records remain opaque and untouched; no new persisted writer or format reinterpretation is introduced.

7. **Plan revision recommendations**

   Add a `## Revisions` entry defining compatibility-control scratch retention, cleanup ownership and removal. Correct the lifecycle paragraph’s “no new artifact family” claim to account for those private clones.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"38541188-1e97-43fd-bbe1-2a6102f6957c","repository":"github.com/xianxu/ariadne","issue_id":"000524","recipe":"milestone-review","round":2,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/d83c023049a68299","reviewed_head":"744236f93156b5662e61fffdd8bfbad206e33b96","slot":"ariadne:3","requested_at":"2026-10-21T12:15:58.292339Z"},"phase":"finished","pid":76466,"started_at":"2026-10-21T12:15:58.40264Z","finished_at":"2026-10-21T12:18:18.619684Z","session_id":"01a15e63-c2c9-7fb3-80c2-4b3934137f9c","outcome":"returned"}] -->
