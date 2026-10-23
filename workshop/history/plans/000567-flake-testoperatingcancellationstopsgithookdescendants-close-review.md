# Boundary Review — ariadne#567 (whole-issue close)

| field | value |
|-------|-------|
| issue | 567 — flake: TestOperatingCancellationStopsGitHookDescendants |
| repo | ariadne |
| issue file | workshop/issues/000567-flake-testoperatingcancellationstopsgithookdescendants.md |
| boundary | whole-issue close |
| milestone | — |
| window | 28c40fd5dc3485dce3ec4936485238a56743f4ab..2dddb0b59de75742cb9e288d679285c327184c4f |
| command | sdlc close --issue 567 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-23T05:35:47-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The cancellation correction matches the Spec and passed native repetition and race checks. One inexpensive gap remains in the new lint rule: valid whitespace-separated `ps` formats bypass its protection. No production cancellation code changed.

Done-when: current — it still covers the revised fixture, lint backstop, negative control, and acceptance checks.

1. **Strengths**

   - Child-origin readiness establishes that the descendant is executing before cancellation (`cmd/tl-lead/doc_operating_test.go:374`).
   - The fixture validates its private group and reuses `procgroup.Kill`, preserving the deadline and cancellation assertions (`:428–460`).
   - Lint tests cover aliases, shadowing, production exclusion, the captured original source, and linear visits.

2. **Critical findings**

   None.

3. **Important findings**

   - `pkg/golint/ps_liveness.go:59`: format parsing splits only on commas. Native commands using `"pid stat="`, `"pid, stat="`, and `"pid, state="` all returned process states during review, but none matches this predicate. Normalize supported field separators and add regression cases through both `Check` and `CheckSources`. This is an **ARCH-PURPOSE** gap in the mechanical backstop.

4. **Minor findings**

   None.

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: literal-command-format-parity
    title: |
      Recognize whitespace-separated ps state format fields
    detail: |
      pkg/golint/ps_liveness.go:56–61 splits format text only on commas and does not trim fields. Native /bin/ps accepted "pid stat=", "pid, stat=", and "pid, state=" during review, but the predicate misses all three, allowing the same snapshot oracle through the new guard (ARCH-PURPOSE). The complete affected family in this window is this shared parser across Command and CommandContext, including separate -o arguments, attached -o formats, and --format= formats containing whitespace separators or comma-adjacent whitespace. Normalize supported field separators and add cases for stat/state through both Check and CheckSources.
```

5. **Test coverage notes**

   Passed: 30 cancellation repetitions; race checks repeated twice for cancellation, neighboring repository cancellation, lint, and process groups; broader operating tests; tracked-file catalog; diff whitespace check. The committed negative-control evidence was inspected, not independently rerun. The full suite remains pending according to the tracker.

6. **Architectural notes**

   - **ARCH-DRY: pass** — shared group-lifetime semantics are reused.
   - **ARCH-PURE: pass** — lint predicates remain pure; native IO stays in fixtures.
   - **ARCH-PURPOSE: flag** — format parsing leaves equivalent literal queries unguarded.
   - **ARCH-EVOLVE: pass** — no persisted format changes.

   The atlas documents the new rule. No new user command, flag, or configuration requires a README update.

7. **Plan revision recommendations**

   Append a `## Revisions` entry recording the expanded literal-format regression inventory. The existing Done-when clauses remain applicable.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"cfb2e440-fe80-4229-9c0c-b0c7d952f81d","repository":"github.com/xianxu/ariadne","issue_id":"000567","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/b787b09c2e352652","reviewed_head":"2dddb0b59de75742cb9e288d679285c327184c4f","slot":"ariadne:3","requested_at":"2026-10-23T12:33:01.754033Z"},"phase":"finished","pid":69737,"started_at":"2026-10-23T12:33:01.855061Z","finished_at":"2026-10-23T12:35:47.327745Z","session_id":"01a168c0-18b9-7b83-847e-81de3406c869","outcome":"returned"}] -->

---

## Re-review — 2026-10-23T05:41:42-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 567 — flake: TestOperatingCancellationStopsGitHookDescendants |
| repo | ariadne |
| issue file | workshop/issues/000567-flake-testoperatingcancellationstopsgithookdescendants.md |
| boundary | whole-issue close |
| milestone | — |
| window | 28c40fd5dc3485dce3ec4936485238a56743f4ab..839fc18ae652343c0777a8a515ac943476fc32d3 |
| command | sdlc close --issue 567 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-23T05:41:42-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The implementation matches the issue’s Spec and Plan. BR-1 is addressed with regression evidence, and no new findings remain. Native cancellation checks pass, and an independent leader-only mutation confirms the descendant guard detects surviving work. The recorded full-suite pass remains a separate acceptance requirement.

Done-when: current — its clauses cover the revised format parser, cancellation fixture, negative control, and acceptance checks.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      pkg/golint/ps_liveness.go:65 now normalizes comma and whitespace separators in the shared predicate. TestPSLivenessQueriesFormatSeparators covers both constructors, separate and attached formats, and both Check entry points. All 40 cases pass at head and fail with the previous comma-only parser restored through a temporary Go overlay.
```

1. **Strengths**

   - Child-origin readiness establishes an executing descendant before cancellation (`cmd/tl-lead/doc_operating_test.go:374`).
   - Private-group validation and the existing `procgroup` terminal condition replace the snapshot oracle while preserving cancellation assertions (`:428–466`).
   - Lint coverage includes aliases, shadowing, production exclusion, captured source, format variants, and counted linear visits.

2. **Critical findings**

   None.

3. **Important findings**

   None remaining.

4. **Minor findings**

   None.

5. **Test coverage notes**

   Passed: 30 native cancellation repetitions; focused race checks repeated twice; broader operating and repository-cancellation tests; complete golint and procgroup packages; diff whitespace checks.

   Independently confirmed the leader-only mutation fails with “remains runnable.” The captured original test matches the pinned base exactly. Temporary overlays were removed; the checkout remains clean. The full suite was not independently run during this review.

6. **Architectural notes**

   - **ARCH-DRY: pass** — shared process-group semantics and one format predicate are reused.
   - **ARCH-PURE: pass** — lint logic remains pure; native IO stays in integration fixtures.
   - **ARCH-PURPOSE: pass** — the cancellation proof and previously identified format family are covered.
   - **ARCH-EVOLVE: pass** — no persisted-format changes.

   The atlas documents the new lint rule. No new user command, flag, or configuration requires a README update.

7. **Plan revision recommendations**

   None. The existing BR-1 revision describes the delivered correction.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"adc631ff-17f8-4fb3-9fd4-9b134d918ba4","repository":"github.com/xianxu/ariadne","issue_id":"000567","recipe":"small-diff-review","round":2,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/8d3a008bc9780160","reviewed_head":"839fc18ae652343c0777a8a515ac943476fc32d3","slot":"ariadne:3","requested_at":"2026-10-23T12:39:33.831684Z"},"phase":"finished","pid":86717,"started_at":"2026-10-23T12:39:33.932846Z","finished_at":"2026-10-23T12:41:40.981433Z","session_id":"01a168c6-1449-74d1-b7da-9b057e7ccabf","outcome":"returned"}] -->
