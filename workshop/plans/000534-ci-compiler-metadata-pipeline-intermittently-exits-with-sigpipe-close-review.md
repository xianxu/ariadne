# Boundary Review — ariadne#534 (whole-issue close)

| field | value |
|-------|-------|
| issue | 534 — CI compiler metadata pipeline intermittently exits with SIGPIPE |
| repo | ariadne |
| issue file | workshop/issues/000534-ci-compiler-metadata-pipeline-intermittently-exits-with-sigpipe.md |
| boundary | whole-issue close |
| milestone | — |
| window | 0366e7210823bfd8606770d42cf394270ee6e4c1..c267bbefd9e4e337aebf13a2a63392c2e0ea0e8f |
| command | sdlc close --issue 534 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-21T18:51:43-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned change satisfies the issue’s Spec and Done-when requirements. Both shared metadata checks consume their full input while preserving fail-closed validation. Targeted tests pass, and restoring quiet grep reproduces exit 141. No blocking findings.

```findings
{}
```

1. **Strengths**
   - Both compiler slots reuse the corrected checks in `.github/workflows/merge-check.yml:90–91`.
   - `scripts/test/portable-ci.test.sh:28` extracts actual workflow checks and covers repeated valid input plus wrong, dirty, missing-field, and empty metadata.
   - `pkg/shelllint/metadata.go:8` reuses existing lexical helpers; its unit tests require no IO.
   - README and atlas updates accurately describe the fix and guard.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes:** Passed `go test ./pkg/shelllint`, `bash scripts/test/portable-ci.test.sh`, and `git diff --check`. An in-memory quiet-grep mutation of the extracted large-input checks exited 141; pinned-head checks exited 0. Existing identity-failure and old-PATH exclusion controls passed. The full suite was not rerun during this review.

6. **Architectural notes**
   - **ARCH-DRY — pass:** Shared build script and existing lexer remain authoritative.
   - **ARCH-PURE — pass:** Detection is pure; repository scanning stays in the test boundary.
   - **ARCH-PURPOSE — pass:** Enumeration found both production metadata pipelines, both corrected; workflow propagation remains declared through weave’s manifest.
   - **ARCH-EVOLVE — pass:** No persisted format changes.

7. **Plan revision recommendations:** None.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"fe9dcbe1-2817-4036-93ee-0346e373c321","repository":"github.com/xianxu/ariadne","issue_id":"000534","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/815a1356b0ea9227","reviewed_head":"c267bbefd9e4e337aebf13a2a63392c2e0ea0e8f","slot":"ariadne:2","requested_at":"2026-10-22T01:50:21.273853Z"},"phase":"finished","pid":66815,"started_at":"2026-10-22T01:50:21.473701Z","finished_at":"2026-10-22T01:51:43.150179Z","session_id":"01a1614d-5adf-7d02-8e0a-1049f6cb143c","outcome":"returned"}] -->
