# Boundary Review — ariadne#465 (whole-issue close)

| field | value |
|-------|-------|
| issue | 465 — dev-aliases script exceeds its per-script budget under suite load |
| repo | ariadne |
| issue file | workshop/issues/000465-dev-aliases-script-exceeds-its-per-script-budget-under-suite-load.md |
| boundary | whole-issue close |
| milestone | — |
| window | 45a96385ad4d24bda920822c60d7879d643a2a99..4e8fedec83276f8fb2c014b07d2b25ba9dad3834 |
| command | sdlc close --issue 465 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-19T08:25:00-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The path helper and fixture consolidation look sound, and all three alias scripts passed. Two Important fixes remain: restore regression sensitivity for build-start timestamps, and synchronize Done when with the revised acceptance contract. No production correctness defect was found.

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: preserve-regression-sensitivity
    title: |
      Future-dated edits no longer test build-start timestamp preservation
    detail: |
      construct/scripts/test/dev-aliases.test.sh:443-444 sets the mid-build edit to 2030, so it remains newer than the completed binary even if timestamp restoration breaks. An in-memory mutation removing touch -r from both generated skipd definitions still passed 82/82 assertions, including lines 491 and 493. These are all instances of this weakened timestamp fixture in the window. Use ordered start/edit/output timestamps and verify the regression fails without restoration. ARCH-PURPOSE: retaining assertion counts must preserve their behavioral coverage.
  - id: new
    severity: Important
    kind: defect
    family: acceptance-contract-synchronization
    title: |
      Done when does not reflect the revised acceptance contract
    detail: |
      workshop/issues/000465-dev-aliases-script-exceeds-its-per-script-budget-under-suite-load.md:35-36 retains the original wall-time criteria, while Revisions at lines 53-55 adds production path conformance and N/2N zero-path-utility requirements and accepts counted reductions plus a loaded recorded-suite pass. This issue's Done when section is the sole instance in the window. Update it to name those required checks and reconcile the performance evidence, preserving the recorded-suite requirement.
```

Done-when: stale — the production path contract and revised performance evidence appear only in Revisions.

1. **Strengths**

   - `construct/dev-aliases.sh:110` centralizes path splitting across startup and collection.
   - `dev-aliases-path.test.sh` passed platform-tool conformance and N/2N zero-launch checks.
   - FIFO barriers deterministically exercise concurrent partial builds; successful sentinel copies preserve distinct output inodes.

2. **Critical findings:** None.

3. **Important findings:** The two findings above. Restore meaningful timestamp coverage and update the current acceptance criteria.

4. **Minor findings:** None.

5. **Test coverage notes**

   All three alias scripts passed; the main script reported **82 passed, 0 failed**. The timestamp mutation also passed 82/82, demonstrating the coverage gap. `git diff --check` passed. A loaded recorded full-suite pass was not independently verified; the pinned issue records it as queued.

6. **Architectural notes**

   - **ARCH-DRY: pass** — equivalent path operations share one helper.
   - **ARCH-PURE: pass** — path splitting contains no IO; layout checks remain in integration code.
   - **ARCH-PURPOSE: flag** — the timestamp fixture preserves the count while weakening existing behavioral coverage.
   - **ARCH-EVOLVE: pass** — no persisted format change.

   Atlas documents the new helper and fixture approach. No new user-facing command, flag, or configuration requires a README update. No Core concepts table is present.

7. **Plan revision recommendations**

   Append a `## Revisions` entry recording the timestamp-fixture correction and synchronize Done when with path conformance, N/2N counts, revised performance evidence, repeated targeted passes, and the required recorded-suite pass.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"af0ef016-bea9-4bcf-a294-6a71259082fb","repository":"github.com/xianxu/ariadne","issue_id":"000465","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/65c64d79c4c83df2","reviewed_head":"4e8fedec83276f8fb2c014b07d2b25ba9dad3834","slot":"ariadne:3","requested_at":"2026-10-19T15:22:48.307457Z"},"phase":"finished","pid":97669,"started_at":"2026-10-19T15:22:48.494841Z","finished_at":"2026-10-19T15:24:59.038878Z","session_id":"01a154c2-18ff-7342-921b-226179de5ba1","outcome":"returned"}] -->
