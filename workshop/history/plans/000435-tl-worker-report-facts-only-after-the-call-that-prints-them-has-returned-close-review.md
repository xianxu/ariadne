# Boundary Review — ariadne#435 (whole-issue close)

| field | value |
|-------|-------|
| issue | 435 — tl-worker: report facts only after the call that prints them has returned |
| repo | ariadne |
| issue file | workshop/issues/000435-tl-worker-report-facts-only-after-the-call-that-prints-them-has-returned.md |
| boundary | whole-issue close |
| milestone | — |
| window | b6a22b0bff4b6f37deb3046f3c2faf92552784dc..743550c0d70662d74d683e262ed653b3d699adda |
| command | sdlc close --issue 435 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-17T21:02:00-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The eight-line change satisfies the issue’s Spec and Done when. The reporting rule reaches command output through the existing embedded help, and its regression guards fail without the change. No blocking findings.

1. **Strengths**
   - `cmd/tl-worker/help.md:99–102` covers fact categories, returned-call ordering, parallel-batch prohibition, and copying verified output.
   - `cmd/tl-worker/testdata/rules.tsv:20–23` reuses the existing test machinery for both invocation forms.
   - Worker rules remain single-sourced; the skill stub and lead help point to the worker binary.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**
   - Passed: `go test ./cmd/tl-worker ./cmd/tl-lead ./pkg/tlrules -count=1`.
   - In a temporary pinned-head copy, replacing only help with the base version produced all eight expected failures: four guards in each invocation form. Temporary files were removed.
   - Built help inspection and `git diff --check` passed. The full suite was not independently rerun.

6. **Architectural notes**
   - **ARCH-DRY — pass:** existing embedded source and shared rule-table infrastructure reused.
   - **ARCH-PURE — pass:** no business logic or IO responsibilities added.
   - **ARCH-PURPOSE — pass:** delivers the full requested help contract.
   - Docs gate: no new architectural surface, command, flag, or configuration key requiring atlas or README changes.

7. **Plan revision recommendations:** None.

```findings
findings: []
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"a7928d4c-b271-4c6a-8efc-4736a53652f3","repository":"github.com/xianxu/ariadne","issue_id":"000435","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/13aef2689887149d","reviewed_head":"743550c0d70662d74d683e262ed653b3d699adda","slot":"ariadne:1","requested_at":"2026-10-18T04:01:10.750588Z"},"phase":"finished","pid":75670,"started_at":"2026-10-18T04:01:10.865047Z","finished_at":"2026-10-18T04:01:59.686102Z","session_id":"01a14d2b-b02c-7f31-b241-83aa76e83fb7","outcome":"returned"}] -->
