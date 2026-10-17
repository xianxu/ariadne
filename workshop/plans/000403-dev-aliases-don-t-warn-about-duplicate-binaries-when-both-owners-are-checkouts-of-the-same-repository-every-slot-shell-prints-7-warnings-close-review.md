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
