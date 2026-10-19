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

---

## Re-review — 2026-10-19T08:31:18-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 465 — dev-aliases script exceeds its per-script budget under suite load |
| repo | ariadne |
| issue file | workshop/issues/000465-dev-aliases-script-exceeds-its-per-script-budget-under-suite-load.md |
| boundary | whole-issue close |
| milestone | — |
| window | 45a96385ad4d24bda920822c60d7879d643a2a99..7b64fba896b2b67f054c10e6ab6efd9b31d84566 |
| command | sdlc close --issue 465 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-19T08:31:18-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

Both prior findings are addressed. The path optimization and alias scripts pass, and removing timestamp restoration now fails the intended regression. One Important coverage regression remains: strict-list comparisons substitute expected behavior for an actual command result.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      construct/scripts/test/dev-aliases.test.sh:446-454 establishes start=2000, edit=2001, output=2002. Normal execution passes 82/82; replacing timestamp restoration with a no-op makes the mid-build-edit assertion at line 499 fail (79 passed, three failures including two shape checks).
  - id: BR-2
    disposition: addressed
    note: |
      The pinned before/after diff updates the issue's Done when at lines 35-37 to require path conformance, N/2N zero-launch checks, counted reductions, repeated targeted checks, and a loaded recorded-suite pass, matching Revisions at lines 54-56 while retaining the two-minute ceiling.
findings:
  - id: new
    severity: Important
    kind: defect
    family: preserve-regression-sensitivity
    title: |
      Strict-list comparison no longer observes ordinary strict output
    detail: |
      construct/scripts/test/dev-aliases.test.sh:35-44 skips ordinary --strict --list execution, substitutes a supplied status, and compares verbose strict output against a non-strict baseline. All instances are the same-origin and distinct-origin strict calls at lines 280 and 301. A production mutation suppressing ordinary strict-list output only on collisions passes 82/82 at head; the identical mutation fails the original test (81 passed, one failure).
      This is the 2nd finding in family preserve-regression-sensitivity. Earlier rounds fixed instances. Do not fix only the demonstrated collision case: enforce the rule that fixture reuse must preserve observed output and status for every originally tested mode. Capture or reuse actual strict results for both enumerated cases and verify the mutation fails. ARCH-PURPOSE requires preserving behavioral coverage, not merely assertion counts.
```

1. **Strengths**

   - `construct/dev-aliases.sh:110` centralizes equivalent path operations without subprocesses.
   - Path conformance and N/2N zero-launch checks pass.
   - FIFO barriers exercise partial concurrent builds deterministically.
   - The repaired timestamp fixture detects missing restoration.

2. **Critical findings:** None.

3. **Important findings:** The strict-list coverage regression above. Restore actual strict baselines for both fixture cases.

4. **Minor findings:** None.

5. **Test coverage notes**

   All three alias scripts passed; the main script reported **82 passed, zero failed**. Timestamp and strict-list mutations were tested in disposable copies, then removed. Syntax checks and `git diff --check` passed. A loaded recorded-suite pass remains required before landing; it was not independently verified here.

6. **Architectural notes**

   - **ARCH-DRY: pass** — shared path helper; no duplicate implementation found.
   - **ARCH-PURE: pass** — path splitting remains free of IO.
   - **ARCH-PURPOSE: flag** — strict-mode coverage weakened during optimization.
   - **ARCH-EVOLVE: pass** — no persisted-format behavior changes.

   Atlas covers the helper and fixture changes. No new user-facing command, flag, or configuration requires a README update. No Core concepts table is present.

7. **Plan revision recommendations**

   Add a `## Revisions` entry establishing the fixture-reuse rule above and recording restored comparisons against actual strict output/status for both origin cases.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"b70b1f4e-f683-4548-a58a-a7a1da42fe4e","repository":"github.com/xianxu/ariadne","issue_id":"000465","recipe":"small-diff-review","round":2,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/43d958124e9bd312","reviewed_head":"7b64fba896b2b67f054c10e6ab6efd9b31d84566","slot":"ariadne:3","requested_at":"2026-10-19T15:28:38.159182Z"},"phase":"finished","pid":67808,"started_at":"2026-10-19T15:28:38.340237Z","finished_at":"2026-10-19T15:31:17.622585Z","session_id":"01a154c7-6f68-7d42-af66-f7acbd01040d","outcome":"returned"}] -->

---

## Re-review — 2026-10-19T08:37:50-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 465 — dev-aliases script exceeds its per-script budget under suite load |
| repo | ariadne |
| issue file | workshop/issues/000465-dev-aliases-script-exceeds-its-per-script-budget-under-suite-load.md |
| boundary | whole-issue close |
| milestone | — |
| window | 45a96385ad4d24bda920822c60d7879d643a2a99..f874054852a15da9f76d9d2fc23482a05f99a148 |
| command | sdlc close --issue 465 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-19T08:37:50-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The strict-list coverage regression is fixed, and no new defects were found. All targeted checks passed, including mutation checks for strict-list output and build-start timestamps. The loaded recorded-suite pass remains required before landing under the documented review/queue overlap; this review does not establish that pass.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      construct/scripts/test/dev-aliases.test.sh:440-448 establishes ordered start/edit/output timestamps. Disabling timestamp restoration makes the intended mid-build-edit assertion at line 493 fail; the mutated run reports 79 passed, three failed.
  - id: BR-2
    disposition: addressed
    note: |
      The pinned before/after diff updates the issue's Done when at lines 35-37 to include path conformance, N/2N zero-launch checks, counted reductions and loaded-suite evidence, matching its Revisions while retaining the two-minute ceiling.
  - id: BR-3
    disposition: addressed
    note: |
      construct/scripts/test/dev-aliases.test.sh:35-44 captures actual output and status for all four ordinary/verbose comparisons, including strict cases at lines 274 and 295. Suppressing ordinary strict-list output on collisions makes both repaired comparisons fail (79 passed, three failed). Fresh gen_def and nous_def generation also restores observations after fixture changes; retained dupe.defs reuse observes the same mode on unchanged inputs.
```

1. **Strengths**

   - `construct/dev-aliases.sh:110` consolidates equivalent path operations into one subprocess-free helper.
   - Path conformance and N/2N owner/CLI checks pass.
   - FIFO barriers exercise partial concurrent builds deterministically; successful output copies preserve distinct inodes.
   - Fixture reuse now preserves actual CLI output and status.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   - Main alias script: **82 passed, zero failed**.
   - All three alias scripts passed twice in each of the three runner environments; shell lint passed.
   - Strict-list and timestamp mutations failed the intended behavioral assertions in disposable copies, subsequently removed.
   - `git diff --check` passed. The two-minute runner ceiling is unchanged.

6. **Architectural notes**

   - **ARCH-DRY: pass** — path splitting has one shared implementation.
   - **ARCH-PURE: pass** — path logic contains no IO; integration fixtures exercise real layout behavior.
   - **ARCH-PURPOSE: pass** — optimization retains behavioral coverage, including the repaired strict comparisons.
   - **ARCH-EVOLVE: pass** — no persisted-format changes.

   Atlas documents the helper and fixture approach. No new user-facing command, flag or configuration requires a README update. No Core concepts table is present.

7. **Plan revision recommendations:** None; the fixture-reuse rule and acceptance revisions match the implementation.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"e3684d9e-2063-4047-8a2f-561e7bbd96ee","repository":"github.com/xianxu/ariadne","issue_id":"000465","recipe":"small-diff-review","round":3,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/94e7533771bd0b4a","reviewed_head":"f874054852a15da9f76d9d2fc23482a05f99a148","slot":"ariadne:3","requested_at":"2026-10-19T15:34:54.975909Z"},"phase":"finished","pid":60090,"started_at":"2026-10-19T15:34:55.087171Z","finished_at":"2026-10-19T15:37:40.0536Z","session_id":"01a154cd-2e71-7271-a4e8-794f335b2981","outcome":"returned"}] -->
