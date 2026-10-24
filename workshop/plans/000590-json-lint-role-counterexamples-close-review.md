# Boundary Review — ariadne#590 (whole-issue close)

| field | value |
|-------|-------|
| issue | 590 — Qualify recursive JSON lint against independent oracles and DOM materializers |
| repo | ariadne |
| issue file | workshop/issues/000590-json-lint-role-counterexamples.md |
| boundary | whole-issue close |
| milestone | — |
| window | b7eb6073c2482053f8237c481d05b9a7db092e8b..3543b227e091a856e526fa7d9374a0982f4c5921 |
| command | sdlc close --issue 590 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-24T05:13:23-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

Both captured false positives are accepted, existing tests pass, and remediation now respects downstream module ownership. However, both new exemptions can accept invalidated value-flow evidence, weakening validator detection contrary to the Spec. This blocks shipping.

1. **Strengths**

   - Exact Pair captures run through both `Check` and `CheckSources` (`pkg/golint/json_scanner_test.go:103`).
   - Role decisions remain AST-only, with a shared summary and counted traversal checks.
   - The diagnostic removes the inaccessible internal-import recommendation; atlas guidance documents the intended conservative boundary.

2. **Critical findings**

```findings
findings:
  - id: new
    severity: Critical
    kind: defect
    family: role-evidence-provenance
    title: |
      Role exemptions survive mutations that invalidate their evidence
    detail: |
      ARCH-PURPOSE: pkg/golint/json_scanner.go:250 and :299 inspect only the candidate's lexical assignments and uses. Starting from the accepted map fixture, inserting func(){m=map[string]any{}}() before return still produces zero findings; nested closure escapes through mutate(m) or mutate(&m) are also ignored. The returned container no longer has proven recursive-child flow, contradicting the promised conservative handling of overwritten/escaping containers.
      The same provenance gap affects oracle qualification at :395 and :436: modifying the captured callback to rebind raw=nil between scanner calls still produces zero findings. Overwriting both error results before comparison through a nested closure, pointer aliases, or parenthesized assignments also produces zero findings, although ordinary direct assignments correctly restore the finding. Sweep both role predicates for binding writes, address exposure and captures, reusing the shared summary; require stable comparison inputs/results and container/child evidence. Add regression cases through Check and CheckSources that fail on this pinned head.
```

   Fix sketch: use shared binding evidence across lexical scopes and normalize assignment targets consistently. Where mutation or escape makes provenance uncertain, retain the finding rather than attempting general alias analysis.

3. **Important findings**

   None.

4. **Minor findings**

   None.

5. **Test coverage notes**

   Passed:

   - `go test ./pkg/golint -count=1`
   - `go test -race ./pkg/golint -count=1`
   - `git diff --check` over the pinned range.

   Temporary external harnesses reproduced every bypass described above against the exported `Check` API. Existing controls cover direct overwrites and visibly different input expressions, but miss mutation through captures, aliases, parentheses, and rebinding between calls. Repository files remain unchanged.

6. **Architectural notes**

   - **ARCH-DRY — pass:** one summary and shared rule serve both checking entry points.
   - **ARCH-PURE — pass:** declared pure entities exist and consume AST data; fixture IO stays in replay setup.
   - **ARCH-PURPOSE — flag:** invalidated evidence establishes exemptions, violating the conservative role contract.
   - **ARCH-MOCK — pass:** no new external runtime dependency or service seam.
   - **ARCH-CONSTRAINTS — pass:** shared traversal and counted 8/16-group tests support the declared bounded approach.
   - **ARCH-SECURE — pass:** parsing remains at the existing boundary; no new credentials or external effects.
   - **ARCH-ORDER — pass:** summaries hold no state between invocations.
   - **ARCH-FUNERAL — pass:** committed captures are fixed-size evidence; summaries die with invocation scope.
   - **ARCH-EVOLVE — pass:** no persisted-format changes.

   Atlas was updated. No new command, flag, configuration key, or usage step requires a README update. Landing and downstream notification remain explicitly pending.

7. **Plan revision recommendations**

   Add a `## Revisions` entry requiring stable provenance for every binding used to establish either role: containers, recursive child values, comparison results, and scanner inputs. Enumerate direct, parenthesized, address-mediated, and captured mutations, and require conservative rejection plus regression coverage for that class.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"6f861dd9-ca59-4562-93f1-3698b9927c04","repository":"github.com/xianxu/ariadne","issue_id":"000590","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/388c734f8cbff34d","reviewed_head":"3543b227e091a856e526fa7d9374a0982f4c5921","slot":"ariadne:3","requested_at":"2026-10-24T12:10:37.99285Z"},"phase":"finished","pid":49466,"started_at":"2026-10-24T12:10:38.116004Z","finished_at":"2026-10-24T12:13:22.734086Z","session_id":"01a16dd1-f3f8-7af3-ba1e-efc600dbc082","outcome":"returned"}] -->

---

## Re-review — 2026-10-24T05:43:07-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 590 — Qualify recursive JSON lint against independent oracles and DOM materializers |
| repo | ariadne |
| issue file | workshop/issues/000590-json-lint-role-counterexamples.md |
| boundary | whole-issue close |
| milestone | — |
| window | b7eb6073c2482053f8237c481d05b9a7db092e8b..c5f9d98cf00c351ad254cd4696e391e03053301a |
| command | sdlc close --issue 590 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-24T05:43:07-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The captured counterexamples pass, existing package tests pass, and downstream remediation is improved. However, BR-1 remains incomplete: both checking APIs still grant exemptions after mutations invalidate the evidence supporting them.

```findings
dispose:
  - id: BR-1
    disposition: not-addressed
    note: |
      ARCH-PURPOSE: the added regressions demonstrably fail against the pre-fix implementation, but the provenance class remains open. At pkg/golint/json_scanner.go:369-418, adding m["key"]=nil after the accepted fixture's recursive-child storage still produces zero findings through Check and CheckSources; stored remains true and every indexed container use is considered safe. Passing m["key"] to an unknown mutate helper likewise remains accepted. At :63-119 and :611, range assignments are absent from binding-write evidence: inserting for _, oldErr = range []error{nil} {}; for _, newErr = range []error{nil} {} before the captured oracle comparison still produces zero findings through both APIs, although neither result retains scanner provenance. Extend the shared evidence rule to all binding-write syntax and to overwrites/escapes of stored children. Enumerate these cases and add regressions; do not patch only the demonstrated expressions.
```

1. **Strengths**

   - Exact Pair captures are replayed through both `Check` and `CheckSources`.
   - The diagnostic now recommends the consumer module’s shared decoder.
   - The shared AST summary and counted traversal tests provide useful structural and cost controls.
   - Existing provenance regressions are meaningful: reverting the implementation in a disposable copy makes them fail behaviorally.

2. **Critical findings**

   **BR-1 remains open.** Two independently reproduced gaps:
   
   - [json_scanner.go:369](/Volumes/TestScratch/review-checkouts/2e1c466d7cdc4564/pkg/golint/json_scanner.go:369): stored-child evidence survives indexed overwrite and projected-value escape. Conservatively invalidate uncertain storage evidence.
   - [json_scanner.go:63](/Volumes/TestScratch/review-checkouts/2e1c466d7cdc4564/pkg/golint/json_scanner.go:63): the write summary omits `RangeStmt` assignments, allowing overwritten comparison results to qualify an oracle. Include every applicable binding-write form.

   These maintain the existing `role-evidence-provenance` finding; no duplicate finding is raised.

3. **Important findings**

   None additional.

4. **Minor findings**

   None.

5. **Test coverage notes**

   - `go test ./pkg/golint -count=1` passed.
   - `go test -race ./pkg/golint -count=1` passed.
   - Three supplemental counterexamples failed their expected-finding assertions through **both APIs**: map-element overwrite, projected child escape, and range-rebound error results.
   - Disposable test files were removed; the reviewed checkout remains unchanged.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared summary and checking paths.
   - **ARCH-PURE — pass:** role decisions consume AST data without IO.
   - **ARCH-PURPOSE — flag:** BR-1’s conservative provenance contract remains incomplete.
   - **ARCH-MOCK — pass:** no new external dependency.
   - **ARCH-CONSTRAINTS — pass:** counted tests exercise declared traversal bounds.
   - **ARCH-SECURE — flag:** BR-1 permits invalidated source evidence to retain authority.
   - **ARCH-ORDER — pass:** no state persists between invocations.
   - **ARCH-FUNERAL — pass:** summary lifetime is scoped; fixtures are fixed repository artifacts.
   - **ARCH-EVOLVE — pass:** no persisted format changes.

   Atlas documentation was updated. No new command, flag, or configuration surface requires a README update. Core-concept entities exist at their stated locations.

7. **Plan revision recommendations**

   Add a `## Revisions` entry expanding BR-1’s enumeration to all Go binding-write forms, including range assignments, and to mutation or escape of children after storage. Keep the class-fix completion claim pending until these regressions pass through both APIs.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"b2a38413-e32c-4705-80c8-1acbe319fa35","repository":"github.com/xianxu/ariadne","issue_id":"000590","recipe":"milestone-review","round":2,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/2e1c466d7cdc4564","reviewed_head":"c5f9d98cf00c351ad254cd4696e391e03053301a","slot":"ariadne:3","requested_at":"2026-10-24T12:39:27.635974Z"},"phase":"finished","pid":31111,"started_at":"2026-10-24T12:39:27.7397Z","finished_at":"2026-10-24T12:43:06.561437Z","session_id":"01a16dec-5972-71a3-bcbf-c60cf7a4a299","outcome":"returned"}] -->
