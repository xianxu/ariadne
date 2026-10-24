# Boundary Review — ariadne#594 (whole-issue close)

| field | value |
|-------|-------|
| issue | 594 — Actual attribution excludes witnessed local staging intervals |
| repo | ariadne |
| issue file | workshop/issues/000594-actual-attribution-excludes-witnessed-local-staging-intervals.md |
| boundary | whole-issue close |
| milestone | — |
| window | 6caf95c6fcfdb977b62c347bc36ff3eec47c36b4..6c77a6d37dbd2451ea82f0f0494bf60f0b6f9fa7 |
| command | sdlc close --issue 594 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-24T09:32:56-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The implementation matches the approved staging-exclusion policy. Selection, retained-proof validation, visible warnings, and immutable replay agree, and focused normal/race tests pass. No runtime defect was found. Update README’s attribution guidance before closing; the plan also needs a minor classification cleanup.

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: user-facing-contract-doc-parity
    title: |
      README update appears missing for witnessed staging exclusions
    detail: |
      README.md:1206–1216 describes the actual/active-time attribution contract but omits the newly supported stage/* exclusion and its visible, unassigned branch-window minutes. README.md is unchanged in this range, while cmd/sdlc/helptext/actual.md:54 and active-time.md:85 document the changed behavior. This family's affected README consumers are actual/close and active-time, both covered by this paragraph. Update it with the concurrent-claim exclusion, unchanged sole-reservation fallback, and strict unknown/detached refusal.
  - id: new
    severity: Minor
    kind: defect
    family: core-concept-classification
    title: |
      Core concepts tables omit PURE/INTEGRATION classifications
    detail: |
      The plan's tables at lines 13 and 21 omit the required Kind column. Enumerated entries are IsStagingBranch, SelectBranchActivity, MeasurementBranchInterval, measurementBranchWarning/validateOwnedBranchProof, and Existing HEAD capture/decoder. Add explicit classifications; separate capture from decoding, since capture performs IO while decoding, warning rendering, and proof validation are pure.
```

1. **Strengths**

   - `branch.go:165` excludes staging only after the existing own/foreign attribution and sole-reservation fallback.
   - Live selection and saved validation share `IsStagingBranch`, preventing predicate drift.
   - Tests verify the exact 172-second exclusion for both concurrent issues, malformed/unknown witnesses, credited-window tampering, and visible minutes.
   - The Git/tracker integration test verifies retained replay after deleting live HEAD history.

2. **Critical findings**

   None.

3. **Important findings**

   README parity finding above. Extend `README.md:1206` to describe the changed attribution contract.

4. **Minor findings**

   Add explicit entity classifications to both Core concepts tables.

5. **Test coverage notes**

   Passed:

   - Complete `activetime` and `gatestate` package tests.
   - Exact command-level staging, resting, and neutral-exclusion replay tests.
   - Focused race tests across all three packages.
   - `git diff --check`.

   The broader command-test regex was stopped after several minutes without a result; it is not counted as passing. No full-suite result was independently established.

6. **Architectural notes**

   - **ARCH-DRY — pass:** staging classification is shared across live and retained consumers; the consumer sweep found no missing runtime handling.
   - **ARCH-PURE — pass:** classification and interval selection remain deterministic; Git capture stays in the existing IO boundary.
   - **ARCH-PURPOSE — pass:** both concurrent candidates exclude staging, sole-reservation behavior remains intact, and replay preserves the exclusion.
   - **ARCH-EVOLVE — pass:** the new audit kind preserves existing meanings and codecs. Unknown nested members round-trip; strict validation remains. The pinned base reader rejects unknown kinds rather than silently reinterpreting them.

7. **Plan revision recommendations**

   Add a `## Revisions` entry recording explicit PURE/INTEGRATION classifications and separating HEAD capture from decoding. Include the README attribution update in the documentation checklist.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"03b6fa5f-cf43-4fbf-96f5-87bcc85df82c","repository":"github.com/xianxu/ariadne","issue_id":"000594","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/07bf2fc457c06052","reviewed_head":"6c77a6d37dbd2451ea82f0f0494bf60f0b6f9fa7","slot":"ariadne:3","requested_at":"2026-10-24T16:27:51.193187Z"},"phase":"finished","pid":54651,"started_at":"2026-10-24T16:27:51.363338Z","finished_at":"2026-10-24T16:32:55.873585Z","session_id":"01a16ebd-73ca-70f0-9da3-fdbf9769a9fa","outcome":"returned"}] -->
