# Boundary Review — ariadne#153 (whole-issue close)

| field | value |
|-------|-------|
| issue | 153 — sdlc retro process manual |
| repo | ariadne |
| issue file | workshop/issues/000153-sdlc-retro-process-manual.md |
| boundary | whole-issue close |
| milestone | — |
| window | fe7b76e6a51c77aee9385dba24e6d8271c453690..HEAD |
| command | sdlc close --issue 153 |
| reviewer | claude |
| timestamp | 2026-07-01T15:55:40-07:00 |
| verdict | SHIP |

## Review

I have everything I need. I independently verified the load-bearing claims rather than trusting the commit messages: the byte-fidelity golden was genuinely captured pre-refactor (commit `180eeca` still had all 7 `fmt.Sprintf` and only added test+testdata), was never re-captured after, and passes today; build/vet/tests are green; the committed manual is byte-identical to a fresh regeneration; the rename left no stale `retro` references; and the memory-privacy guard is real and pinned.

```verdict
verdict: SHIP
confidence: high
```

**Summary.** This whole-issue close (M1+M2, branch-point→HEAD) delivers a clean, well-tested feature: a read-only `sdlc process-manual` verb that unrolls every always-on injection source into one linked markdown manual (M1), plus the M2 refactor that lifts the 7 agent-emitting judge prompts out of `fmt.Sprintf` literals into embedded, readable `judge/prompts/*.md` templates. The load-bearing M2 risk — **byte-fidelity of the fresh-context review prompts** — is airtight and I verified it independently via git archaeology (golden captured at `180eeca` while `prompts.go` still had 7 `fmt.Sprintf`; goldens never re-captured; `TestBuildPrompt_Golden` green). The M2 milestone review's two actionable findings were both addressed (privacy-refusal regression test added; category list consolidated into `judge.AllInjectedCategories()`). My independent pass surfaced no Critical or Important findings — only pre-existing/deferred Minors. Nothing blocks SHIP.

### 1. Strengths

- **Byte-fidelity gated correctly (M2's whole risk).** `golden_test.go:36-52` captures from the pre-refactor builder, the file's load-bearing `⛔ never re-run -update-golden` comment prevents baking in drift, and the frozen goldens are the acceptance net. I confirmed `180eeca` touched only test+testdata and that `git log 180eeca..HEAD -- testdata/golden/` is empty. This is exactly how a "must not change a byte" refactor should be gated.
- **Real DRY win (ARCH-DRY pass).** 7 bespoke `fmt.Sprintf` arg lists collapse to one `strings.Replacer` token table (`prompts.go:53-70`); `BuildPrompt` drops from a 7-arm switch to 5 lines (`prompts.go:194-199`); `AllInjectedCategories()` (`prompts.go:104-106`) now single-sources the complete injected set so both the manual and the golden derive from it; `orDefault` is reused for the empty-`IssueRef`/`PlanContent` fallbacks. The single-pass-Replacer reasoning (a `{{…}}`-bearing diff inserted literally) is correct and documented (`prompts.go:51`).
- **Textbook pure-core / thin-shell (ARCH-PURE pass).** `InjectionSource`, `renderManual`, `hasHeadingLine`, `firstParagraph`, `claudeProjectSlug`, `judgeSources`, `promptTemplate/Substitutions` are pure; every IO collector takes an injected seam (`fs.FS`, dir, `$HOME`) and tests against `fstest.MapFS`/temp dirs with no mocks. `Collect` is the one thin aggregator; `processmanual.go` is thin cobra glue mirroring `NewEstimateSourceCmd`.
- **Purpose fully delivered (ARCH-PURPOSE pass).** Shadow-sweep confirmed: the manual spans all 8 judge categories (incl. `estimate-quality` that `AllCategories()` omits), help text, skills, lessons, AGENTS chain, memories (`TestCollect_SpansEveryKind`). For M2, `prompts/*.md` are the *sole* source (all `fmt.Sprintf` deleted, `fmt` import dropped), enforced byte-for-byte by the golden — not a hand-restatement — and `process-manual` links straight to them (`collect.go` `judgeLink`).
- **Privacy default + guard, both pinned.** Memories redacted by default (`memory.go`); `TestMemorySources_RedactedByDefault` plants a `SECRET` and asserts no leak; the `--include-memory`+`--out` refusal (`processmanual.go:56`) is now pinned by `TestRunProcessManual_IncludeMemoryWithOutRefused`. I re-ran it: green.
- **Deterministic committed manual.** `go run ./cmd/sdlc process-manual --out atlas/_pmtest.md` is byte-identical to the tracked `atlas/process-manual.md` — a genuine regeneration, not a stale snapshot.

### 2. Critical findings

None.

### 3. Important findings

None.

### 4. Minor findings

- **No freshness guard on the tracked `atlas/process-manual.md`.** It's deliberately committed (`046e954`) as a browsable map, and the Spec sells the manual as "cheap, deterministic, **always accurate**" — yet nothing fails when a prompt/skill/helptext/lessons/AGENTS edit lands without a regen, so the committed copy *can* silently rot even though regeneration is deterministic. The repo already has the pattern to fix this (the golden test). Consider a small test that regenerates and diffs (skip when outside a repo), or drop the tracked copy and regenerate on demand. (Low-harm — links still resolve — hence Minor.)
- **`~~~` fence assumes no inlined body contains a `~~~` line** (`source.go` `renderManual`). Verified none does today (grepped all prompt/skill/helptext/lessons sources). Already noted+deferred in the M2 review; revisit only if arbitrary external bodies ever get inlined (CommonMark "fence longer than any run in the body").
- **`hasHeadingLine` detects only ATX headings, not Setext** (`source.go:110`). A Setext-underlined body wouldn't be fenced, but it can't hijack the `##`/`###` navigation outline, so no correctness impact — note only.

### 5. Test coverage notes

Strong and targeted at the bugs this diff could ship: golden pins all 7 rendered prompts byte-for-byte; `TestBuildPrompt_EmptyIssueRefFallback` covers the `<unknown>` branch `goldenInput` doesn't; `TestCollect_SpansEveryKind` is a genuine ARCH-PURPOSE shadow-sweep over all six kinds; `TestJudgeSources_CoversEveryCategoryIncludingEstimate` keeps a *deliberately hardcoded* 8-element list as the drop-catching pin; `TestJudgeSources_FullVsGist` pins gist-vs-`--full` via the ARCH-registry marker; `TestRenderManual_FencesHeadingBearingBodies` and `TestRenderManual_AbsoluteAndEmptyLinks` cover the two rendering bugs found during the smoke run; memory privacy is covered at both the collector and cmd layers. The only untested property is the "gist vs `--full` outline is byte-identical" invariant (documented in the Log, not pinned) — acceptable.

### 6. Architectural notes for upcoming work

- `InjectionSource` is intentionally shaped to gain `Fired`/`Order` fields for the dynamic session-reconstruction pass (now **ariadne#157**), and this static catalog is that work's stated baseline — good forward design. The `AllInjectedCategories()` consolidation matters more once #157 consumes the catalog, since its "which fired" diff needs the complete injected set single-sourced.
- `cmd/sdlc` propagates downstream as a whole tool resolved by location (build-in-owner, not per-file manifest), so the new `internal/processmanual` package, `processmanual.go`, and `prompts/*.md` ship additively with no manifest edit required — confirmed against `construct/base.manifest`. The change is additive + read-only, low downstream risk.

### 7. Plan revision recommendations

None. Both plan files' Core-concepts tables match the code exactly — every pure entity and integration point exists at its stated path with the stated status, the M1 plan's `## Revisions` already records the `retro`→`process-manual` rename and the gist/fencing deltas, and the M2→M3 renumber is consistent with the issue (M3 split to #157). No `## Revisions` entry needed.
