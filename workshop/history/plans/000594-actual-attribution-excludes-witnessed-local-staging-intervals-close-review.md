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

---

## Re-review — 2026-10-24T09:40:26-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 594 — Actual attribution excludes witnessed local staging intervals |
| repo | ariadne |
| issue file | workshop/issues/000594-actual-attribution-excludes-witnessed-local-staging-intervals.md |
| boundary | whole-issue close |
| milestone | — |
| window | 6caf95c6fcfdb977b62c347bc36ff3eec47c36b4..0e1eeeb84505767cb8aab13fee3756943c5886ec |
| command | sdlc close --issue 594 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-24T09:40:26-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned implementation matches the approved policy: witnessed staging intervals remain unassigned under concurrent claims, sole-reservation fallback remains intact, and unknown evidence still refuses measurement. Both prior findings are addressed. No new blocking findings were found.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      README.md:1206–1220 now documents concurrent resting/staging exclusion, visible branch-window minutes, sole-reservation fallback, strict missing/detached/ambiguous refusal, and retained compatibility. The pinned correction agrees with branch.go:153–169 and actualbranch.go:298–303.
  - id: BR-2
    disposition: addressed
    note: |
      The plan's Core concepts tables at lines 13–25 now classify all enumerated entities and separate IO capture from pure decoding, selection, rendering, and proof validation. Inspection of the named definitions confirms these classifications; the correction is prose-only.
```

1. **Strengths**

   - Live selection and retained validation share `IsStagingBranch`.
   - Selection preserves own/foreign branch precedence and sole-reservation fallback.
   - Regression tests verify exactly 172 excluded seconds for both concurrent issues.
   - Git/tracker integration coverage verifies visible exclusions and retained replay after deleting live HEAD history.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   Passed complete `activetime` and `gatestate` package tests, five focused command-level tests, focused race checks across all three packages, and `git diff --check`. Coverage includes malformed/unknown witnesses, credited-window tampering, nested unknown-field preservation, and older-reader refusal. The separate pre-close suite result was not independently verified.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared predicate; consumer sweep found selection, validation, warning, and replay handling aligned.
   - **ARCH-PURE — pass:** new classification remains deterministic; IO stays in existing capture code.
   - **ARCH-PURPOSE — pass:** concurrent candidates exclude staging without inferring ownership; actual/close and active-time share the behavior.
   - **ARCH-EVOLVE — pass:** additive audit kind, preserved existing meanings and unknown fields, strict validation, unchanged adopted replay.

7. **Plan revision recommendations:** None.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"8ed56abf-ba2c-4345-a5c3-60c1daef6a9a","repository":"github.com/xianxu/ariadne","issue_id":"000594","recipe":"small-diff-review","round":2,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/c5b4d26f99a57385","reviewed_head":"0e1eeeb84505767cb8aab13fee3756943c5886ec","slot":"ariadne:3","requested_at":"2026-10-24T16:38:58.446722Z"},"phase":"finished","pid":83623,"started_at":"2026-10-24T16:38:58.599924Z","finished_at":"2026-10-24T16:40:24.417659Z","session_id":"01a16ec7-a101-7523-b121-2a08fd18fd94","outcome":"returned"}] -->
