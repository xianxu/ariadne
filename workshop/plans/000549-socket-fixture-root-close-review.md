# Boundary Review — ariadne#549 (whole-issue close)

| field | value |
|-------|-------|
| issue | 549 — Lint inherited temp roots used for Unix socket fixtures |
| repo | ariadne |
| issue file | workshop/issues/000549-socket-fixture-root.md |
| boundary | whole-issue close |
| milestone | — |
| window | cb3952669997cf4835f5ce139a4e0b1b72840f6f..41ca34e209e77a44661d746a1dbbb21ad78d29d7 |
| command | sdlc close --issue 549 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-22T14:45:27-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The rule rejects the recorded unsafe Pair fixture, preserves the corrected fixture, and passes package and focused race tests. Two gaps remain within its documented analysis boundary: parenthesized bindings lose provenance, and explicitly typed local testing receivers are ignored.

```findings
findings:
  - id: new
    severity: Critical
    kind: defect
    family: transparent-expression-provenance
    title: |
      Parentheses bypass supported assignment and multi-result propagation
    detail: |
      pkg/golint/socket_root.go:143 requires an identifier LHS, and line 150 requires an unparenthesized call RHS. Consequently, both `var root string; (root)=t.TempDir(); net.Listen("unix",root)` and `root,err:=(os.MkdirTemp("","s-")); _=err; net.Listen("unix",root)` produce no finding. Independently verified that both compile. Normalize transparent parentheses before binding identifiers and recognizing first-result calls, and add regressions through Check and CheckSources. This violates the plan's assignment/parentheses contract (ARCH-PURPOSE).
  - id: new
    severity: Important
    kind: defect
    family: visible-receiver-recognition
    title: |
      Explicitly typed local testing receivers are silently excluded
    detail: |
      pkg/golint/socket_root.go:84 accepts receiver declarations only as ast.Field. Thus `var tt *testing.T=t; net.Listen("unix",tt.TempDir())` produces no finding, despite the receiver type being directly visible and the plan promising visible testing T/B/F/TB receivers. Recognize explicit local variable types for TempDir and Setenv, with positive tests and unrelated-type controls; document any remaining receiver-alias limitation precisely (ARCH-PURPOSE).
```

1. **Strengths**

   - One monotone provenance graph handles branching writes, aliases, cycles, and fan-out without repeated whole-file scans.
   - Recorded unsafe/corrected Pair source tests assert the actual isolation-root sink, preserving hashed namespace inputs.
   - The native test successfully connects and accepts under long TMPDIR, then verifies private-root removal.
   - README and atlas describe the rule, suppression policy, and unsupported analysis boundaries.

2. **Critical findings**

   - `pkg/golint/socket_root.go:143–150`: unwrap parentheses consistently at binding boundaries. Both independently compiled reproductions currently return zero findings.

3. **Important findings**

   - `pkg/golint/socket_root.go:84`: support directly declared local testing receiver types. The current parameter-only implementation is narrower than the documented contract.

4. **Minor findings**

   None.

5. **Test coverage notes**

   - Passed `go test ./pkg/golint -count=1`.
   - Passed `go test -race ./pkg/golint -run 'TestSocketFixture' -count=1 -v`.
   - Native proof used an 81-byte socket path despite a 253-byte TMPDIR.
   - Additional probes confirmed both reported gaps. Add regression tests for them.
   - The whole-module suite was not rerun during this review.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared graph and existing parser/suppression machinery.
   - **ARCH-PURE — pass:** analysis remains deterministic; IO stays in fixture tests.
   - **ARCH-PURPOSE — flag:** documented supported syntax has the gaps above.
   - **ARCH-MOCK — pass:** no new production external dependency; native conformance uses private resources.
   - **ARCH-CONSTRAINTS — pass:** worklist structure and counted growth tests support linear analysis.
   - **ARCH-SECURE — pass:** source is parsed without execution; native resources are privately owned.
   - **ARCH-ORDER — pass:** monotone analysis has no cross-event authoritative state; listener lifetime is bounded.
   - **ARCH-FUNERAL — pass:** cleanup closes connections/listener and removes the root.
   - **ARCH-EVOLVE — pass:** no persisted format changes.

7. **Plan revision recommendations**

   Add a `## Revisions` entry recording consistent parenthesis normalization and explicit local testing receiver recognition. Clarify the remaining receiver-alias boundary and link the new regression coverage.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"e15553c4-5a33-4a7b-b632-ebfd71c54de2","repository":"github.com/xianxu/ariadne","issue_id":"000549","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/e22ccfaaec5d768e","reviewed_head":"41ca34e209e77a44661d746a1dbbb21ad78d29d7","slot":"ariadne:5","requested_at":"2026-10-22T21:42:36.101603Z"},"phase":"finished","pid":38840,"started_at":"2026-10-22T21:42:36.411579Z","finished_at":"2026-10-22T21:44:54.952988Z","session_id":"01a16590-e4c5-7c11-8123-92c0072c1ea5","outcome":"returned"}] -->
