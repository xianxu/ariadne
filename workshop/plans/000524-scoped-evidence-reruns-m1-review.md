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
