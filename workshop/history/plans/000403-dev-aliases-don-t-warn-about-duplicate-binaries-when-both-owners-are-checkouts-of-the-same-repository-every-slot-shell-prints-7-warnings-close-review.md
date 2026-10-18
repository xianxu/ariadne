# Boundary Review — ariadne#403 (whole-issue close)

| field | value |
|-------|-------|
| issue | 403 — dev-aliases: don't warn about duplicate binaries when both owners are checkouts of the same repository (every slot shell prints 7 warnings) |
| repo | ariadne |
| issue file | workshop/issues/000403-dev-aliases-don-t-warn-about-duplicate-binaries-when-both-owners-are-checkouts-of-the-same-repository-every-slot-shell-prints-7-warnings.md |
| boundary | whole-issue close |
| milestone | — |
| window | 75b21892013595a0101c044ec83446353b789415..c64466ea43415a2281cd773e9aa50be67d1c6451 |
| command | sdlc close --issue 403 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-17T16:00:47-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The original duplicate-warning bug is fixed, with meaningful regression coverage and unchanged owner selection. No runtime correctness defect surfaced. Before shipping, document the new flag and extend acceptance coverage for the revised origin policy.

Done-when: stale — it covers the original slot collision but omits the revised non-GitHub and unrecognized-origin rules.

```findings
findings:
  - id: new
    severity: Important
    family: user-surface-documentation
    title: |
      README update appears missing for dev-aliases --verbose
    detail: |
      construct/dev-aliases.sh:62 introduces --verbose, but README.md is unchanged. This is the only new CLI flag in the window. Add its purpose and a diagnostic invocation near README.md:342; explain the revised --strict collision semantics there as well.
  - id: new
    severity: Important
    family: acceptance-contract-coverage
    title: |
      Acceptance criteria and fixtures omit the revised origin policy
    detail: |
      Issue 000403:44-45 retains the original Done-when despite the identity-policy revision at :67-69. construct/scripts/test/dev-aliases.test.sh:198-244 covers GitHub HTTPS/scp-style equivalence, distinct GitHub origins, and absent origins, but not identical non-GitHub URLs, non-GitHub transport differences, non-GitHub path-case differences, or unrecognized origins. Enumerate these cases in Done-when and add fixtures asserting stderr and strict status; the GitHub ssh:// branch at construct/dev-aliases.sh:213 also lacks a fixture.
  - id: new
    severity: Minor
    family: single-source-collision-policy
    title: |
      ARCH-DRY: warning and strict status independently classify collisions
    detail: |
      construct/dev-aliases.sh:267 counts identities using sort -u, while :274 independently groups them in awk for strict status. These are both instances of the same collision decision. Compute classification once and derive diagnostics and exit status from it.
```

1. **Strengths**

   - Same-origin fixtures verify silence, successful strict status, and selection of the slot substrate.
   - The new regression tests fail against the base script: 67 passed, 3 failed.
   - Identity metadata stays internal; public `--list` output and final owner mapping retain their format.
   - Atlas documentation explains normalization and conservative fallback behavior.

2. **Critical findings:** None.

3. **Important findings:** README coverage and revised acceptance coverage, detailed above.

4. **Minor findings:** Consolidate the duplicated collision classification.

5. **Test coverage notes:** Head suite: **70 passed, 0 failed**. Syntax and diff-whitespace checks passed. Additional temporary probes passed for explicit GitHub SSH, non-GitHub equality, transport/case differences, and unrecognized origins. These probes do not replace committed regression tests. Full build and suite were not independently rerun.

6. **Architectural notes:** **ARCH-DRY: flag** as above. **ARCH-PURE: pass** — Git lookup occurs during collection; collision decisions consume collected data. **ARCH-PURPOSE: pass** — the motivating slot layout is exercised and genuine collisions remain visible.

7. **Plan revision recommendation:** Add a `## Revisions` entry expanding acceptance criteria and fixtures to the complete origin-policy matrix. No implementation-scope revision is otherwise needed.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"4ee188b1-4e57-4072-a142-6a5a83652279","repository":"github.com/xianxu/ariadne","issue_id":"000403","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/c79d45ee4a30e25c","reviewed_head":"c64466ea43415a2281cd773e9aa50be67d1c6451","slot":"ariadne:1","requested_at":"2026-10-17T22:59:17.015214Z"},"phase":"finished","pid":22621,"started_at":"2026-10-17T22:59:17.109274Z","finished_at":"2026-10-17T23:00:47.037027Z","session_id":"01a14c17-4adc-7a93-976a-a15088310af0","outcome":"returned"}] -->

---

## Re-review — 2026-10-17T16:45:04-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 403 — dev-aliases: don't warn about duplicate binaries when both owners are checkouts of the same repository (every slot shell prints 7 warnings) |
| repo | ariadne |
| issue file | workshop/issues/000403-dev-aliases-don-t-warn-about-duplicate-binaries-when-both-owners-are-checkouts-of-the-same-repository-every-slot-shell-prints-7-warnings.md |
| boundary | whole-issue close |
| milestone | — |
| window | 75b21892013595a0101c044ec83446353b789415..95b82627c2109ab7a8f9dc0d50dddf08eb2ef959 |
| command | sdlc close --issue 403 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-17T16:45:04-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The original collision bug is fixed, and all three prior findings are addressed. The focused suite passes **75/75**. One acceptance-test gap remains: verbose diagnostics must preserve stdout, but a mutation that corrupts `--list` output still passes every test.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      README.md:345-349 now documents the diagnostic invocation and revised strict semantics, matching construct/dev-aliases.sh:267-274 and :358-360.
  - id: BR-2
    disposition: addressed
    note: |
      Issue Done when and Revisions enumerate the requested origin cases; tests at construct/scripts/test/dev-aliases.test.sh:218-273 exercise them. Removing ssh:// recognition or non-GitHub exact equality independently makes the suite fail.
  - id: BR-3
    disposition: addressed
    note: |
      construct/dev-aliases.sh:262-278 computes classification once; the same identity-count branch controls warnings and had_dup, which determines strict status.
findings:
  - id: new
    severity: Important
    family: acceptance-contract-coverage
    title: |
      Verbose stdout preservation is accepted but untested
    detail: |
      construct/scripts/test/dev-aliases.test.sh:216 discards stdout from the only verbose invocation, so it cannot verify issue Done when :46's promise that --verbose preserves --list output. Adding unexpected stdout in the verbose branch still produces 75 passed, 0 failed. This is the 2nd finding in family acceptance-contract-coverage. Apply the rule that diagnostic flags preserve machine-readable stdout while reporting on stderr: compare complete list output with and without --verbose, both normally and under --strict, and assert diagnostics separately. The sole affected fixture is :216-217.
```

1. **Strengths**

   - The slot fixture verifies silent same-origin handling, successful strict status, and selection of the substrate checkout.
   - Distinct, absent, and unrecognized origins retain strict collision failures.
   - Warning and strict decisions share one classification; owner selection remains independent.
   - README and atlas cover the new diagnostic surface.

2. **Critical findings:** None.

3. **Important findings:** Add stdout-preservation coverage described above. No runtime defect was observed.

4. **Minor findings:** None.

5. **Test coverage notes**

   - Head suite: **75 passed, 0 failed**.
   - Head tests against the base script: **70 passed, 5 failed**.
   - Removing GitHub `ssh://` recognition or non-GitHub equality each causes a targeted failure.
   - Verbose stdout-corruption mutation incorrectly stays green.
   - Shell syntax and diff-whitespace checks pass. Full build and full suite were not independently rerun. Temporary probes were removed; tracked files remain unchanged.

6. **Architectural notes**

   - **ARCH-DRY: pass** — collision classification has one source.
   - **ARCH-PURE: pass** — Git IO occurs during collection; classification operates on collected data.
   - **ARCH-PURPOSE: pass** — the motivating slot layout is covered, and genuine collisions remain visible.
   - No Core concepts table is present.

7. **Plan revision recommendation**

   Add a `## Revisions` entry stating that diagnostic-mode fixtures must independently assert stdout, stderr, and status, including verbose list-output equivalence with strict mode on and off.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"4777a6a3-40b3-4791-aa0f-ffabbb0ae88d","repository":"github.com/xianxu/ariadne","issue_id":"000403","recipe":"small-diff-review","round":2,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/6b9584961e75a262","reviewed_head":"95b82627c2109ab7a8f9dc0d50dddf08eb2ef959","slot":"ariadne:1","requested_at":"2026-10-17T23:43:43.691335Z"},"phase":"finished","pid":43205,"started_at":"2026-10-17T23:43:43.788667Z","finished_at":"2026-10-17T23:45:04.504798Z","session_id":"01a14c3f-fb9c-7dd2-9259-9f7bdee76bcf","outcome":"returned"}] -->

---

## Re-review — 2026-10-17T17:44:30-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 403 — dev-aliases: don't warn about duplicate binaries when both owners are checkouts of the same repository (every slot shell prints 7 warnings) |
| repo | ariadne |
| issue file | workshop/issues/000403-dev-aliases-don-t-warn-about-duplicate-binaries-when-both-owners-are-checkouts-of-the-same-repository-every-slot-shell-prints-7-warnings.md |
| boundary | whole-issue close |
| milestone | — |
| window | 75b21892013595a0101c044ec83446353b789415..c4815824a6b8a59fcc673bf7a7e6e5cdd2550109 |
| command | sdlc close --issue 403 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-17T17:44:30-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The implementation satisfies the revised acceptance criteria. Same-origin checkouts remain quiet, genuine collisions retain warnings and strict failures, and owner selection stays deterministic. BR-4 now has meaningful regression coverage; no new findings surfaced.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      README.md:345-349 documents --verbose and the revised strict collision rule, matching the implementation.
  - id: BR-2
    disposition: addressed
    note: |
      Issue Done when and Revisions enumerate the origin policy; construct/scripts/test/dev-aliases.test.sh:214-289 exercises the specified cases.
  - id: BR-3
    disposition: addressed
    note: |
      construct/dev-aliases.sh:262-278 classifies collisions once; warnings and strict status consume the same identity-count decision.
  - id: BR-4
    disposition: addressed
    note: |
      construct/scripts/test/dev-aliases.test.sh:30-41 compares complete stdout and status, with separate stderr assertions, across four invocations at :230-231 and :248-249. Temporary stdout pollution makes all four equivalence checks fail; a verbose-only exit-status mutation also fails.
```

1. **Strengths**

   - The slot fixture verifies silence, strict success, and selection of the substrate checkout.
   - Origin-policy fixtures cover GitHub transport equivalence and conservative non-GitHub handling.
   - Diagnostic assertions compare output files byte for byte.
   - README and atlas document the changed surface.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   - Pinned-head focused suite: **82 passed, 0 failed**.
   - Stdout-pollution mutation: **4 failures**, covering all four diagnostic combinations.
   - Verbose-only early-exit mutation: **8 failures**.
   - Shell syntax and range whitespace checks passed.
   - Full build and repository-wide suite were not independently rerun. Temporary mutation files were removed; tracked files remain unchanged.

6. **Architectural notes**

   - **ARCH-DRY: pass** — warnings and strict status share collision classification.
   - **ARCH-PURE: pass** — Git IO occurs during collection; classification operates on collected data.
   - **ARCH-PURPOSE: pass** — the motivating slot layout works while genuine collisions remain visible.
   - No Core concepts table is present.

7. **Plan revision recommendations:** None; existing revisions match the implementation and coverage.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"41f57ea9-c50e-4a54-a77b-18310066d5aa","repository":"github.com/xianxu/ariadne","issue_id":"000403","recipe":"small-diff-review","round":3,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/3a1c81bffa59a823","reviewed_head":"c4815824a6b8a59fcc673bf7a7e6e5cdd2550109","slot":"ariadne:1","requested_at":"2026-10-18T00:42:08.774339Z"},"phase":"finished","pid":79219,"started_at":"2026-10-18T00:42:08.989902Z","finished_at":"2026-10-18T00:44:29.439031Z","session_id":"01a14c75-789a-7fb3-a8b2-2e9108437f26","outcome":"returned"}] -->
