# Boundary Review — ariadne#476 (whole-issue close)

| field | value |
|-------|-------|
| issue | 476 — Test attribution: quoted nested go test output (a helper subprocess's indented FAIL lines) makes a failed suite unattributable |
| repo | ariadne |
| issue file | workshop/issues/000476-test-attribution-quoted-nested-go-test-output-a-helper-subprocess-s-indented-fail-lines-makes-a-failed-suite-unattributable.md |
| boundary | whole-issue close |
| milestone | — |
| window | 0956e5e1c71bd6734f195929fc151e0c752dbc14..74500f04dd8acac5aad8ebaf9cc81267f5a2b2b0 |
| command | sdlc close --issue 476 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-19T21:13:01-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The motivating fixture and focused tests pass, and package trailers, timing facts, and JSON controls now preserve indentation correctly. One blocking gap remains: quoted helper failures sharing the outer test’s name prefix still become attributed subtests.

Done-when: current — its quoted-output exclusion and genuine-descendant preservation clauses cover this change, including the remaining gap.

```findings
findings:
  - id: new
    severity: Critical
    kind: defect
    family: suite-evidence-provenance
    title: |
      Shared-name helper quotations still become real failing subtests
    detail: |
      cmd/sdlc/suitefailures.go:62–63 admits every indented FAIL whose name starts with outerTest+"/", regardless of its quotation context. Compiling the pinned scanner with an outer TestOuter, real child TestOuter/real, and a twelve-space quoted "--- FAIL: TestOuter/helper" produces both example.com/outer.TestOuter/real and example.com/outer.TestOuter/helper. The invented leaf can block a registered retry; without a real child, it can also replace the outer failure as the leaf judged by the registry. ARCH-PURPOSE: quoted output must remain inert even when names overlap. The family has one admission predicate, affecting plain output, sharded summary context, CRLF output, and historical-log readers through the shared scanner. Preserve structural indentation/ancestor context rather than using name prefix alone, and add regression cases for overlapping helper names in these modes.
```

1. **Strengths**
   - Centralized provenance check protects package attribution, unattributed failures, and timing facts.
   - The integration regression exercises the actual registered outer retry.
   - README and atlas updates document the changed behavior and lint guard.

2. **Critical findings**
   - The finding above targets `cmd/sdlc/suitefailures.go:62`. The current fixture uses distinct helper names, so it misses this collision.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage**
   - Focused quoted-output, retry, traversal-count, existing parser/subtest/classifier tests, `pkg/golint`, and `internal/flakes` passed.
   - The overlapping-name reproduction compiled the pinned production scanner in temporary scratch space; the checkout remains unchanged.
   - Full-suite verification was not performed.

6. **Architecture**
   - **ARCH-DRY: pass** — shared scanner and provenance helper.
   - **ARCH-PURE: pass** — parsing remains deterministic; IO stays outside it.
   - **ARCH-PURPOSE: flag** — overlapping names defeat quoted-output isolation.
   - **ARCH-EVOLVE: pass** — no persisted format changes.

7. **Plan revision recommendation**
   - Add a `## Revisions` entry documenting structural provenance handling and the plain/sharded/CRLF overlapping-name regression matrix.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"7cda764e-5187-463e-af6c-58cd83575115","repository":"github.com/xianxu/ariadne","issue_id":"000476","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/7a8d63d3c22ef16d","reviewed_head":"74500f04dd8acac5aad8ebaf9cc81267f5a2b2b0","slot":"ariadne:3","requested_at":"2026-10-20T04:10:59.508699Z"},"phase":"finished","pid":98824,"started_at":"2026-10-20T04:10:59.683475Z","finished_at":"2026-10-20T04:13:00.105894Z","session_id":"01a15781-64ae-7422-a86f-670669da8a69","outcome":"returned"}] -->

---

## Re-review — 2026-10-19T21:21:42-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 476 — Test attribution: quoted nested go test output (a helper subprocess's indented FAIL lines) makes a failed suite unattributable |
| repo | ariadne |
| issue file | workshop/issues/000476-test-attribution-quoted-nested-go-test-output-a-helper-subprocess-s-indented-fail-lines-makes-a-failed-suite-unattributable.md |
| boundary | whole-issue close |
| milestone | — |
| window | 0956e5e1c71bd6734f195929fc151e0c752dbc14..36450af3e8a779480309930b7d6bb1516aa93fdf |
| command | sdlc close --issue 476 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-19T21:21:42-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The original twelve-space reproducer is fixed, and its regressions demonstrably fail with the pre-correction scanner. However, the same admission predicate still accepts malformed mixed indentation as real subtest evidence, contrary to the plan’s explicit rejection rule. BR-1 remains open for this variant.

Done-when: current — quoted-output exclusion and preservation of genuine descendants still cover the revised implementation.

```findings
dispose:
  - id: BR-1
    disposition: not-addressed
    note: |
      The original plain/CRLF/sharded regressions and quoted-leaf gate test pass at HEAD and fail with the pre-correction scanner. However, cmd/sdlc/suitefailures.go:19 counts only initial ASCII spaces, while :66 subsequently erases all whitespace with strings.Fields. Compiling the pinned production scanner with "--- FAIL: TestOuter", followed by four spaces plus a tab plus "--- FAIL: TestOuter/helper", and a genuine outer package trailer produces example.com/outer.TestOuter/helper as a failing subtest; four spaces plus U+00A0 does likewise. The plan explicitly says tab indentation cannot change the stack. ARCH-PURPOSE / ARCH-SECURE: the structural provenance rule remains incomplete. This is another instance in family suite-evidence-provenance; fix the rule across the complete leading-whitespace prefix, rather than special-casing either character, and add mixed-indentation regressions across the existing modes.
```

1. **Strengths**

   - Package controls, JSON events, and timing facts now share the raw-line provenance check.
   - Observed ancestry preserves siblings, grandchildren, and literal-slash names.
   - The integration tests verify both the registered outer retry and refusal of a quoted-only registration.
   - README and atlas updates describe the changed behavior and lint guard.

2. **Critical findings**

   - **BR-1 remains open:** `cmd/sdlc/suitefailures.go:19–23` accepts mixed whitespace because it measures only the initial spaces. Validate the entire indentation prefix before admitting a descendant. The current tab test at `suitefailures_nested_test.go:107` covers a leading tab, but misses spaces followed by a tab.

3. **Important findings**

   - None separately raised.

4. **Minor findings**

   - None.

5. **Test coverage notes**

   - Focused quoted-output, ancestry, retry, traversal, parser, subtest, saved-log, and shard-report checks passed.
   - `pkg/golint` and `internal/flakes` passed.
   - A temporary Go overlay using the pre-correction scanner made all three BR-1 regression tests fail.
   - A separate scratch reproduction confirmed mixed-indentation admission at HEAD.
   - No full suite was run. Scratch artifacts were removed; the checkout remains clean.

6. **Architectural notes**

   - **ARCH-DRY: pass** — shared scanner and control predicate.
   - **ARCH-PURE: pass** — deterministic parsing remains separate from IO.
   - **ARCH-PURPOSE: flag** — malformed indentation can still invent leaf evidence.
   - **ARCH-MOCK: pass** — existing injected integration seams remain intact.
   - **ARCH-CONSTRAINTS: pass** — indexed ancestry and traversal-count checks support the declared linear scan.
   - **ARCH-SECURE: flag** — whitespace normalization admits input outside the declared structural grammar.
   - **ARCH-ORDER: pass** — ancestor transitions, sibling truncation, and resets are explicit.
   - **ARCH-FUNERAL: pass** — new parser state dies with the scan; no durable runtime artifacts.
   - **ARCH-EVOLVE: pass** — persisted formats remain unchanged.

7. **Plan revision recommendations**

   - Add a `## Revisions` entry specifying that the complete indentation prefix must contain exactly four ASCII spaces per level, with regression coverage for mixed whitespace in plain, CRLF, and sharded contexts.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"2ec11678-3790-4414-89b9-28217a31547f","repository":"github.com/xianxu/ariadne","issue_id":"000476","recipe":"milestone-review","round":2,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/523c163bd1ab2932","reviewed_head":"36450af3e8a779480309930b7d6bb1516aa93fdf","slot":"ariadne:3","requested_at":"2026-10-20T04:19:26.60432Z"},"phase":"finished","pid":33849,"started_at":"2026-10-20T04:19:26.695946Z","finished_at":"2026-10-20T04:21:41.140325Z","session_id":"01a15789-2073-7982-b1ea-57a1006bd554","outcome":"returned"}] -->
