# Boundary Review — ariadne#454 (whole-issue close)

| field | value |
|-------|-------|
| issue | 454 — ARCH principle: persisted formats evolve additively; readers ignore unknown fields; refusal only when ignoring is unsafe |
| repo | ariadne |
| issue file | workshop/issues/000454-arch-principle-persisted-formats-evolve-additively-readers-ignore-unknown-fields-refusal-only-when-ignoring-is-unsafe.md |
| boundary | whole-issue close |
| milestone | — |
| window | 08fcf21d27f34f22d50ad384b21a77aff5f4b356..27fb3a3534794bd5ca441ae3070c0fb3cd268c61 |
| command | sdlc close --issue 454 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-18T21:14:25-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The principle matches the Spec and reaches the existing registry consumers. All three live judge fixtures passed, confirming both the refusal finding and the two accepted controls. Two existing flow tests fail because they still expect eight principles; fix these before landing.

Done-when: current — registry delivery and observed plan-quality citation still cover the implemented scope.

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: registry-consumers-derive-from-source
    title: |
      Two flow tests retain the old registry count and now fail
    detail: |
      cmd/sdlc/closeflow_test.go:109 and cmd/sdlc/flow_e2e_test.go:91 both require “each of the 8 entries”; the rendered full registry now contains nine. Independently running TestCloseQuickCrossingShellUpgrades and TestQuickAndFullFlowsSameVerbSequence reproduced both failures. These are all remaining Go assertions of the old count. Under ARCH-DRY, derive the expected header from the registry or shared renderer while retaining the full-versus-quick selection checks, then rerun both tests.
  - id: new
    severity: Minor
    kind: defect
    family: registry-documentation-stays-current
    title: |
      Existing atlas descriptions still show the previous registry
    detail: |
      atlas/workflow/architecture-principles.md:9–10 omits EVOLVE from the current quick-flow list; atlas/workflow/sdlc-binary.md:1625–1629 claims exactly one test restates the marker set despite the remaining count assertions and quick-set tripwire. Generated prompt blocks in atlas/process-manual.md:15, :538, :1000, :1362 and :1850 also omit EVOLVE. Update the descriptions and regenerate the manual; these enumerate the stale registry descriptions found in the reviewed tree.
```

1. **Strengths**
   - `architecture.md:274` includes nested preservation, strict validation, and the unsafe-ignore exception.
   - Existing renderers deliver the entry without another runtime source of truth.
   - Live fixtures exercise actual dispatcher responses, including additive and safety controls.
   - README and atlas updates satisfy the docs update gate.

2. **Critical findings:** None.

3. **Important findings:** Two failing flow tests, detailed above.

4. **Minor findings:** Stale atlas descriptions, detailed above.

5. **Test coverage:** Complete offline judge package passed; focused registry, command, start-plan, and golden checks passed. Live conformance passed all three fixtures. Both identified flow tests failed independently.

6. **Architecture:** ARCH-DRY flags the duplicated count assertions. ARCH-PURE passes: rendering remains pure and live IO stays in the integration test. ARCH-PURPOSE passes: required delivery and observed judge behavior are implemented.

7. **Plan revisions:** No scope revision needed. After fixing the assertions, record both passing flow tests in verification evidence.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"0a68e972-6727-4bfd-806f-ce1d748b597f","repository":"github.com/xianxu/ariadne","issue_id":"000454","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/efbc0e7c14d1c3e5","reviewed_head":"27fb3a3534794bd5ca441ae3070c0fb3cd268c61","slot":"ariadne:3","requested_at":"2026-10-19T04:12:37.041855Z"},"phase":"finished","pid":14425,"started_at":"2026-10-19T04:12:37.260429Z","finished_at":"2026-10-19T04:14:24.431924Z","session_id":"01a1525c-865f-7c30-8a40-fe5927825cae","outcome":"returned"}] -->
