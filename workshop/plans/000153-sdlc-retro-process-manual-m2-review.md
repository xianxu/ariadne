# Boundary Review — ariadne#153 (milestone M2)

| field | value |
|-------|-------|
| issue | 153 — sdlc retro process manual |
| repo | ariadne |
| issue file | workshop/issues/000153-sdlc-retro-process-manual.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | b554b5f1334f7ec7bf7255eaed7df0070c83ba8d^..HEAD |
| command | sdlc milestone-close --issue 153 --milestone M2 |
| reviewer | claude |
| timestamp | 2026-07-01T15:22:27-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

I have everything I need. The privacy guard works (error + exit 1 + no file written) but is untested. Let me compose the review.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

**Summary.** This boundary (M1+M2 combined, branch-point→HEAD) delivers a clean, well-tested feature: a new read-only `sdlc process-manual` verb that unrolls every always-on injection source into one linked markdown manual (M1), plus the M2 refactor that lifts the 7 agent-emitting judge prompts out of `fmt.Sprintf` literals into embedded, readable `judge/prompts/*.md` templates. The load-bearing M2 constraint — **byte-fidelity of the review prompts** — is airtight: I verified the golden was captured from the *pre-refactor* code (commit `180eeca` still had all 7 `fmt.Sprintf`, only test+testdata added), the golden files were *never re-captured* afterward, and `TestBuildPrompt_Golden` passes today. `go build ./...` + `go vet ./cmd/sdlc/...` clean; `internal/judge` and `internal/processmanual` pass on a forced fresh run; and the committed `atlas/process-manual.md` is byte-identical to current binary output (deterministic regeneration, not stale). Nothing blocks SHIP. The two findings — a missing regression test for the documented memory-privacy footgun, and one small ARCH-DRY duplication — are non-blocking. (The `cmd/sdlc` package-level test "FAIL" I hit is a red herring: lock-using tests like `TestSetStatusAlias` block on the real-repo lock that the parent `sdlc milestone-close` legitimately holds while this review runs — not a defect in the diff.)

### 1. Strengths

- **Byte-fidelity done right (M2's whole risk).** Golden captured from original, frozen (only `180eeca` touched `testdata/golden/`), and the plan's `-update-golden` footgun is called out in a load-bearing comment (`golden_test.go`). This is exactly how a "must not change a byte" refactor should be gated.
- **The DRY win is real.** 7 bespoke `fmt.Sprintf` arg lists collapse to one `strings.Replacer` token table (`prompts.go:53-70`); `BuildPrompt` drops from a 7-arm switch to 5 lines (`prompts.go:184-189`); `orDefault` is reused for the empty-`IssueRef`/`PlanContent` fallbacks rather than re-inlining `if s==""`. The single-pass-Replacer reasoning (a `{{…}}`-bearing diff inserted literally) is correct and documented (`prompts.go:51`).
- **Textbook pure-core/thin-shell (ARCH-PURE).** `InjectionSource`, `renderManual`, `firstParagraph`, `hasHeadingLine`, `claudeProjectSlug`, `judgeSources`, `promptTemplate/Substitutions` are pure; every IO collector takes an injected seam (`fs.FS`, dir, `$HOME`) and is tested against `fstest.MapFS`/temp dirs with no mocks (`collect_test.go`, `memory_test.go`).
- **Privacy default + guard.** Memories redacted by default; `TestMemorySources_RedactedByDefault` pins the no-leak invariant against a planted `SECRET`; the `--include-memory`+`--out` refusal works (verified: error, exit 1, no file written).
- **`process-manual` is correctly read-only** (never calls `markMutatingCommand`), matching the `arch-principles`/`estimate-source` idiom.

### 2. Critical findings

None.

### 3. Important findings

- **Missing regression test for the memory-privacy refusal — `cmd/sdlc/processmanual.go:56`.** The `--include-memory` + `--out` refusal is safety-critical: the issue `## Log` (2026-07-01 "--full flag + memory privacy fix") records that tracking the manual once *did* bake private `~/.claude/projects/…` paths + memory contents into this base-layer, downstream-propagating repo — an incident caught only manually before push. The guard works today (I exercised it), but nothing pins it; a future refactor of `runProcessManual` could silently drop the branch and re-open the footgun with no failing test. `TestMemorySources_RedactedByDefault` covers the *default redaction* in the collector, not this cmd-level `--out` refusal. Fix: add a table test on `runProcessManual` asserting `includeMemory && outPath != ""` returns an error and writes no file (mirror the manual check I ran).

### 4. Minor findings

- **ARCH-DRY: the "complete injected category set" is duplicated.** `catalogCategories()` (`processmanual/collect.go:50`) and `goldenCategories()` (`judge/golden_test.go:30`) are byte-identical (`append(AllCategories(), EstimateQuality)`), and `TestJudgeSources_CoversEveryCategoryIncludingEstimate` hardcodes the same 8-element list a third time. One fact, three restatements. Consider exposing it once from the judge package (e.g. `judge.AllInjectedCategories()`) so both the manual and the golden derive from the source — otherwise a future change-code-only category (like `EstimateQuality`) silently drops from the manual, and no test catches it. Minor because categories change rarely.
- **`renderManual` fencing assumes bodies never contain `~~~` — `processmanual/source.go`.** The `~~~` fence is deliberately chosen to survive the backtick fences inside judge prompts (current content verified clean), but a body that itself contains a `~~~` line would break out of the fence. No current trigger; worth a CommonMark-style "fence longer than any run in the body" only if arbitrary external bodies ever get inlined.

### 5. Test coverage notes

Strong. Golden pins all 7 rendered prompts byte-for-byte; `TestBuildPrompt_EmptyIssueRefFallback` covers the `<unknown>` branch that `goldenInput` doesn't; the pre-existing `TestBuildPrompt_PlanQuality_NoSeparatePlan` still covers the `(no separate plan file)` fallback; `TestCollect_SpansEveryKind` is a genuine ARCH-PURPOSE shadow-sweep over all six kinds; `TestJudgeSources_FullVsGist` pins the gist-vs-`--full` distinction via the ARCH-registry marker. The one gap is the Important finding above.

### 6. Architectural notes for upcoming work

- The `InjectionSource` record is intentionally shaped to gain `Fired`/`Order` fields for M3 (dynamic session reconstruction), and this catalog is M3's stated baseline — good forward design. Consolidating the category list (Minor #1) matters more once M3 consumes it, since M3's "which fired" diff needs the *complete* set to be single-sourced.
- The naming freed by the M1 rename (`retro`→`process-manual`) is clean — no stale `internal/retro`/`NewRetroCmd` references remain — leaving `retro` available for the M3 verb as the log intends.

### 7. Plan revision recommendations

None. The M2 plan (`000153-…-m2-plan.md`) Core-concepts table matches the code exactly: all 4 pure entities + the two integration points exist at their stated paths with the stated status, and the shadow-check ("all 7 `.md` exist, each renders byte-identically") is satisfied. The main plan's M2→M3 renumber note is consistent with the issue. No `## Revisions` entry needed.
