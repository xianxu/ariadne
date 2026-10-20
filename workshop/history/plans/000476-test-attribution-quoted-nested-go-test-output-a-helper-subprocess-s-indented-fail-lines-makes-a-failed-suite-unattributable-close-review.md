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

---

## Re-review — 2026-10-19T21:29:44-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 476 — Test attribution: quoted nested go test output (a helper subprocess's indented FAIL lines) makes a failed suite unattributable |
| repo | ariadne |
| issue file | workshop/issues/000476-test-attribution-quoted-nested-go-test-output-a-helper-subprocess-s-indented-fail-lines-makes-a-failed-suite-unattributable.md |
| boundary | whole-issue close |
| milestone | — |
| window | 0956e5e1c71bd6734f195929fc151e0c752dbc14..33d378ebf5cd67465a256df0972cc5673e02f4b3 |
| command | sdlc close --issue 476 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-19T21:29:44-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The whitespace classifier and column-zero control checks improve attribution, and focused tests pass. BR-1 remains reproducible with actual Go test output: parent logging after a failed child leaves stale ancestry that admits a quoted helper failure as a real leaf.

```findings
dispose:
  - id: BR-1
    disposition: not-addressed
    note: |
      cmd/sdlc/suitefailures.go:83–88 retains child ancestry across a parent's subsequent log records. Reproduced using actual go test output: TestOuter runs a failing child, then logs "nested helper output:\n--- FAIL: TestOuter/child/helper (0.01s)". Go prints the parent's log at four spaces and its quoted continuation at eight; the pinned scanner returns both TestOuter/child and the invented TestOuter/child/helper. Leaves then replaces the genuine child with that quoted helper for registry classification. The added regressions cover skipped levels and mixed whitespace but omit this valid output ordering. ARCH-PURPOSE, ARCH-SECURE and ARCH-ORDER remain unsatisfied for family suite-evidence-provenance.
```

Done-when: current — the existing requirement that quoted output remain inert while preserving genuine descendants covers this behavior.

1. **Strengths**

   - One shared classifier validates the complete whitespace prefix before tokenization.
   - Package outcomes, JSON events and timing records consistently require column-zero provenance.
   - Integration tests exercise actual local-check classification and retry selection.
   - README and atlas updates cover the changed attribution contract and lint rule.

2. **Critical findings**

   **BR-1 remains open**, at [suitefailures.go:83](/Volumes/TestScratch/review-checkouts/eb090cf82e2405d6/cmd/sdlc/suitefailures.go:83). The reproduced output contains:

   ```text
   --- FAIL: TestOuter (0.00s)
       --- FAIL: TestOuter/child (0.00s)
           outer_test.go:3: real failure
       outer_test.go:3: nested helper output:
           --- FAIL: TestOuter/child/helper (0.01s)
   ```

   The scanner admits both child names. Registry `Leaves` consequently selects only the invented helper. A helper-only registration can explain the genuine child failure.

   Define structural context transitions for ordinary log records as well as failure records, so returning to parent output invalidates stale descendant context. Sweep parent logging before and after failed children, nested returns and sibling transitions across plain, CRLF, sharded and historical readers. Add a regression asserting that helper-only registration cannot explain this failure.

3. **Important findings**

   None additional.

4. **Minor findings**

   None.

5. **Test coverage notes**

   Passed focused quotation, ancestry, whitespace, retry, traversal, existing parser and saved-log subtest checks, plus `pkg/golint` and `internal/flakes`. A temporary isolated module generated the failing example with real `go test`; the pinned scanner admitted its quotation. Temporary files were removed; the repository remains unchanged. No full-suite claim.

6. **Architectural notes**

   - **ARCH-DRY: pass** — shared scanner and prefix classifier.
   - **ARCH-PURE: pass** — listed pure entities contain no IO; integration fixtures use existing seams.
   - **ARCH-PURPOSE: flag** — BR-1 still violates quotation inertness.
   - **ARCH-MOCK: pass** — no new external dependency; existing integration doubles reused.
   - **ARCH-CONSTRAINTS: pass** — linear prefix inspection and indexed ancestry operations.
   - **ARCH-SECURE: flag** — quoted subprocess text still becomes authoritative evidence.
   - **ARCH-ORDER: flag** — ordinary log records fail to retire stale ancestry.
   - **ARCH-FUNERAL: pass** — transient state ends with the scan; no new runtime artifact family.
   - **ARCH-EVOLVE: pass** — persisted formats remain unchanged.

7. **Plan revision recommendations**

   Add a `## Revisions` entry replacing “rejected records leave the stack unchanged” with a rule distinguishing inert quotations from log records that establish a return to an ancestor’s output. Enumerate the ordering cases above and require parser plus gate regressions before claiming BR-1 resolved.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"3ae9ada4-c4de-41b8-807e-2822cc2bd6c6","repository":"github.com/xianxu/ariadne","issue_id":"000476","recipe":"milestone-review","round":3,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/eb090cf82e2405d6","reviewed_head":"33d378ebf5cd67465a256df0972cc5673e02f4b3","slot":"ariadne:3","requested_at":"2026-10-20T04:28:20.845683Z"},"phase":"finished","pid":19799,"started_at":"2026-10-20T04:28:21.139185Z","finished_at":"2026-10-20T04:29:43.447489Z","session_id":"01a15791-48b2-71a0-a475-b62864cf5be6","outcome":"returned"}] -->

---

## Re-review — 2026-10-19T21:49:16-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 476 — Test attribution: quoted nested go test output (a helper subprocess's indented FAIL lines) makes a failed suite unattributable |
| repo | ariadne |
| issue file | workshop/issues/000476-test-attribution-quoted-nested-go-test-output-a-helper-subprocess-s-indented-fail-lines-makes-a-failed-suite-unattributable.md |
| boundary | whole-issue close |
| milestone | — |
| window | 0956e5e1c71bd6734f195929fc151e0c752dbc14..9a2711a7220ccda98c24c67acad41caac4695998 |
| command | sdlc close --issue 476 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-19T21:49:16-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The implementation matches the revised Spec: ambiguous raw subtests coarsen to proven ancestors, quoted controls remain inert, and structured leaves retain precision. BR-1 is addressed with meaningful regression coverage. No blocking findings remain.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      cmd/sdlc/suitefailures.go:72–76 rejects every indented record before tokenization. TestSuiteQuotedOverlappingNames covers shared-name quotations with and without real children across plain, CRLF and sharded output. It fails when the admission guard is removed in a disposable pinned-head copy. Gate and retained-log tests pass under the revised coarsening contract.
```

1. **Strengths**

   - One shared prefix classifier protects failure names, outcomes and package attribution.
   - Actual retry tests verify parent registrations succeed while quoted-leaf registrations cannot explain failures.
   - Structured names are collected before unidentified-package failures are derived ([teststats_parse.go](cmd/sdlc/teststats_parse.go:167)).
   - README and atlas document the deliberate precision loss and registration implications.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   - Focused scanner, suite facts, retry, saved-pass and environment tests passed: 48.488s.
   - Unattributed-failure, golint and flakes tests passed.
   - Removing the admission guard made the overlapping-name regression fail.
   - Full repository suite was not rerun. `diff --check` reports only intentional indented blank lines in captured Go-output fixtures.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared scanner and prefix grammar.
   - **ARCH-PURE — pass:** deterministic parsing remains separate from IO; concept-table entities exist at their stated paths.
   - **ARCH-PURPOSE — pass:** quotations remain inert across current and retained-log readers.
   - **ARCH-MOCK — pass:** integration checks reuse existing injected seams and stateful GitHub fake.
   - **ARCH-CONSTRAINTS — pass:** fixed linear scans, existing input bounds and N/2N traversal coverage.
   - **ARCH-SECURE — pass:** external output retains provenance before tokenization.
   - **ARCH-ORDER — pass:** unsafe ancestry state is removed; package attribution follows completed name collection.
   - **ARCH-FUNERAL — pass:** no new durable runtime artifacts.
   - **ARCH-EVOLVE — pass:** no persisted-format changes.

7. **Plan revision recommendations:** None. The final revision explicitly supersedes the historical ancestry proposals.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"aac2941f-51e7-4af3-a485-9957dcc4296d","repository":"github.com/xianxu/ariadne","issue_id":"000476","recipe":"milestone-review","round":4,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/5274f5cd64a1a1b7","reviewed_head":"9a2711a7220ccda98c24c67acad41caac4695998","slot":"ariadne:3","requested_at":"2026-10-20T04:47:50.164186Z"},"phase":"finished","pid":27650,"started_at":"2026-10-20T04:47:50.284985Z","finished_at":"2026-10-20T04:49:15.500211Z","session_id":"01a157a3-1f50-7763-9320-be1e00fe1ac7","outcome":"returned"}] -->
