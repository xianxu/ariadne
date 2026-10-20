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
