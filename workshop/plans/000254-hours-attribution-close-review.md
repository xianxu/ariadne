# Boundary Review — ariadne#254 (whole-issue close)

| field | value |
|-------|-------|
| issue | 254 — Hours attribution: ignore GitHub PR numbers; window starts at the claim |
| repo | ariadne |
| issue file | workshop/issues/000254-hours-attribution.md |
| boundary | whole-issue close |
| milestone | — |
| window | fe6c2e2de63ec7f0545f5cba8a0cbfa97c9ad3e4..ed83e9f8f32f53c89a49e6793b73e1f1f48add0a |
| command | sdlc close --issue 254 |
| reviewer | claude |
| timestamp | 2026-10-09T21:18:43-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

This change does what the revised Spec asks, and it does it in one place. The mask lives in `issueref.Find`, and every attribution consumer reads through `Find`. That includes commit boundaries and peers (`activetime/commit.go:100` via `LocalNums`), foreign-ref warnings (`activetime/compute.go:153`) and transcript mentions (`activetime/util.go:61` via `CountLocal`). So one edit covers the whole class. The `Revisions` entry gives a sound, evidence-backed reason for dropping the squash-suffix rule, and Done-when 1 was rewritten to match. Spec 2 (claim-anchored start) is honestly marked as landed in #270. The re-measure table covers Done-when 3. The issueref and activetime tests pass, and `go vet` is clean. Nothing blocks SHIP; only minor notes remain.

1. **Strengths**
   - `cmd/sdlc/internal/issueref/ref.go:63-69`: masking at the single shared scanner means no consumer can be missed (ARCH-DRY pass, ARCH-PURPOSE shadow-sweep pass). The function stays pure, with no IO (ARCH-PURE pass).
   - `ref_test.go:139-161`: one table test checks both `LocalNums` and `CountLocal`. It covers the negative cases that matter: a merge subject that also cites an issue, a trailing ` (#N)` that must stay a ref, and `PR #106` outside the merge lead.
   - `actual.go:113`: the window label now names the resolved start. That fixes a real mislabel: the first `#N` commit is often the filing, which falls outside the claim-anchored window.
   - The lessons entry about baseline binaries records a process trap that the code cannot enforce, which is the right kind of lesson.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - No test pins the new `Window` label format. `actual_test.go:81` and `close_adopt_test.go` still build SHA-style literals by hand (`"abc12345 → HEAD"`), so those fixtures no longer match what `computeActual` produces.
   - After this change, `firstSHA` in `computeActual` is only used as the empty-window check. Rename it or add a comment so the next reader doesn't assume it feeds the label.

5. **Test coverage:** the masking cases cover every clause of Done-when 1: merge, plain `#N` and trailing ` (#N)`. Done-when 2 is pinned by `TestWindowStart` from #270. Done-when 3 is a manual re-measure recorded in the Log.

6. **Architectural notes:** `migrate.go` uses `ScanRE` directly, not `Find`, so it does not get the mask. That is correct, because migrate rewrites refs in prose rather than attributing them. Keep that split if more masks are added.

7. **Plan revisions:** none. The plan matches the code.

```findings
findings:
  - id: new
    severity: Minor
    family: label-format-unpinned
    title: |
      New `<start ISO> → HEAD` window label has no test, and the fixtures still use SHA-style literals
    detail: |
      actual.go:113 changed the label format, but actual_test.go:81 and close_adopt_test.go:38,55,80,119 still build "abc12345 → HEAD"-shaped fixtures, and no test asserts that computeActual emits the ISO start.
  - id: new
    severity: Minor
    family: vestigial-variable
    title: |
      firstSHA in computeActual now serves only as the empty-window check
    detail: |
      The label no longer uses it. A rename (or a comment) would stop readers assuming it still feeds the label.
```
