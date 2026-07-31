# Boundary Review — ariadne#192 (whole-issue close)

| field | value |
|-------|-------|
| issue | 192 — SpanThroughput sums re-close rows as independent issues — blessed baseline inflated 1.4x |
| repo | ariadne |
| issue file | workshop/issues/000192-spanthroughput-sums-re-close-rows-as-independent-issues-blessed-baseline-inflated-1-4x.md |
| boundary | whole-issue close |
| milestone | — |
| window | b741c53ccf5a1aa94a132769c7f027e6d91ea69c..HEAD |
| command | sdlc close --issue 192 |
| reviewer | claude |
| timestamp | 2026-07-29T22:55:51-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

The core fix is correct and lands the issue's actual thesis: the dedupe rule now exists exactly once (`estimate.NewestPerIssue`, `ledger.go:248`), both readers derive from it, `driftSample`'s filter-then-dedupe order is preserved so `drift_test.go` is untouched, and `SpanThroughput` filters to span → dedupes → sums in the only order that's sound. `SpanMeasure` reports `Issues` *and* `RowsScanned` so the gap that hid the 1.41× is now printed, and the re-blessed baseline (80.31 h/wk / 191 issues, with a `#` provenance comment marking the prior row inflated) matches the Done-when known-answer. Nothing blocks SHIP. What's left is the shadow-sweep tail: two hand-maintained restatements of the exact wrong fact are still standing (the ledger file's own header comment and the operator-facing `project throughput` helptext), one count field is still named `Rows` while counting issues, and the one line that writes the provenance count to production (`projectthroughput.go:94`) has no test. **Caveat on confidence: Bash was unavailable in this review environment (EPERM on every invocation), so I did not execute `go test ./...` nor independently recompute the 80.31 figure — every claim below is from reading the code, tests, and both repos' files.**

## 1. Strengths

- **`drift.go:43-61` is the right refactor, not a cosmetic one.** Filtering into `eligible` *before* `NewestPerIssue`, then walking the deduped slice backward for `n`, is set-equivalent to the old backward-walk-with-`seen`: same issues, same newest row each, same newest-first order. PQ-1's warning was heeded exactly, and `drift_test.go` is unmodified in the diff — the guard the plan named actually held.
- **`throughput.go:65-82` gets the two-stage ordering right and says why.** Span-filter-then-dedupe is the only correct order (a 07-29 re-close must not mask the in-span 07-15 measurement), and the comment states the counterfactual rather than just the rule.
- **`RowsScanned` alongside `Issues` is the structurally better fix than just renaming.** The gap between them *is* the duplication, and `projectthroughput.go:102-103` prints both — a recurrence becomes visible instead of silent. That's a real improvement over what the Spec asked for ("report issues, or report both").
- **`TestSpanThroughput_ReCloseEqualsLastRowAlone` (throughput_test.go:164) is the property, stated independently of the numbers**, and it is genuinely mutation-killing: delete the `NewestPerIssue` call and it reads 7.04 vs 3.71.
- **The backward-compat reasoning on `ThroughputBaseline.Rows` (throughput.go:105-112) checks out.** I verified no consumer reads it for math — `internal/project.ComputeForecast` takes only `HoursPerWeek`/`Ceiling` — so keeping the positional 6-column schema and documenting the meaning change is correct, and the `#` comment in `throughput-baseline.tsv:5-10` makes the incomparability legible at the data.

## 2. Critical findings

None.

## 3. Important findings

**I1 — `brain/data/life/42shots/velocity/calibration-ledger.tsv:1-2`: the ledger's own header comment still asserts the wrong fact (ARCH-PURPOSE).**
```
# Calibration ledger (ariadne#117) — estimate↔actual data points, one row per
# closed issue, appended automatically by `sdlc close`.
```
This is the same claim PQ-2 blocked on in `SKILL.md:93`, at a strictly more prominent surface: the first two lines of the file any reader (human or agent) opens. It is not generated — `close.go:915-916` writes only `estimate.Header()` on creation — so it's a hand-maintained doc, and `workshop/history/plans/000118-...-plan.md:579,591` establishes that this header comment is treated as a doc surface that gets updated with the model. The shadow-sweep the ARCH-PURPOSE lens asks for finds this consumer still restating the model instead of deriving from it. *Fix:* rewrite lines 1-2 to say one row per **CLOSE**, re-closes legal, repeat rows are partial sums, readers must dedupe per issue keeping the newest (`estimate.NewestPerIssue`).

**I2 — `cmd/sdlc/helptext/project.md:46-54`: the operator-facing help for the changed command was not updated (Docs update gate).**
Line 47 still reads "measured focused hours/week, **summed from the calibration ledger**" with no mention of the per-issue dedupe, and line 54 documents `--bless` as appending `{span, hours_per_week, rows, ceiling}` with `rows` unqualified. `atlas/workflow/sdlc-binary.md:211-216` was updated for precisely these two facts (including "`rows` counts distinct ISSUES since #192; earlier rows counted raw ledger lines and are not comparable") — the helptext is where an operator actually reads them, and it now contradicts the atlas by omission. The plan's docs row named "**three** surfaces"; `sdlc project throughput --help` is a fourth it missed. *Fix:* qualify line 47 ("summed per issue — the ledger is per-CLOSE, so a re-closed issue counts once at its newest measurement") and annotate `rows` on line 54 as distinct issues since #192.

**I3 — `cmd/sdlc/internal/estimate/throughput.go:33-36`: `UntrustedRows` counts issues but is named `Rows`.**
The whole issue exists because a count named `rows` was read as issues; the fix renames `Rows`→`Issues`/`RowsScanned` for exactly that reason and then leaves the third counter misnamed, documented rather than fixed. A future reader computing `UntrustedRows / RowsScanned` gets a nonsense ratio with no compile error. Behavior and the printed string are correct today, which is why this is Important and not Critical. *Fix:* rename to `UntrustedIssues`; one call site (`projectthroughput.go:104-105`) and one assertion pair (`throughput_test.go:42,197-202`).

**I4 — `cmd/sdlc/projectthroughput.go:94`: the line that writes the provenance count into production is untested.**
`Rows: measure.Issues` is the exact seam the defect shipped through — the persisted count was raw rows. Swap it to `measure.RowsScanned` today and every test in the repo still passes: `projectthroughput_test.go`'s `seedBrainLedger` fixture (lines 21-24) has no re-close, and no test asserts `rows[0].Rows` at all. *Fix:* add a re-close row to a fixture variant (e.g. a second `a#1` at a later date), assert the persisted `rows[0].Rows` equals the distinct-issue count, and assert the stdout line contains "issues from" with both numbers.

**I5 — `cmd/sdlc/internal/estimate/ledger.go:248`: `NewestPerIssue` has no direct test for its two load-bearing contract clauses.**
The function carries a 23-line contract (pre-filtered input; blank `Issue` keyed positionally and never collapsed; order-agnostic last-wins) and zero direct tests. Both clauses are covered *indirectly* — `drift_test.go:14,44,60` blank-issue fixtures would fail if the positional key were dropped — but the failure surfaces as "expected drift warning" three files away, not as a contract violation at the helper. Given this is now the single shared rule two consumers depend on, and the plan itself identified the blank-issue clause as the thing that nearly broke drift semantics, it deserves ~15 lines of direct coverage. *Fix:* `TestNewestPerIssue_KeepsLastPerIssue`, `_BlankIssueRowsStayDistinct`, `_EmptyInput`.

**I6 — issue `## Log` (lines 230-233) is empty while nine plan rows are ticked.**
The Log has a bare `### 2026-07-29` heading and no entries, yet the checklist claims a known-answer check against a live 318-row ledger and two writes into a peer repo. The only record that the 80.31/191 recount was actually run by the fixed binary lives in `brain/.../throughput-baseline.tsv:5-10`. Per §5 the proof belongs in the issue. *Fix:* log the known-answer numbers (110.60/280 → 80.31/191 over 2026-06-22..07-19), that `drift_test.go` passed unmodified, and the brain-side file state.

## 4. Minor findings

- `ledger.go:250,256-263` — the `order` slice is redundant: `sort.Ints(idx)` already makes the result deterministic, so iterating `newest` directly gives the same output. Three lines of state for nothing (Simplicity First).
- `throughput_test.go:177` — exact float equality (`a.HoursPerWeek != b.HoursPerWeek`) is safe only because both branches sum in identical order; it will get brittle the moment the fixture gains a row. Prefer the epsilon form used elsewhere in the file.
- `brain/.../estimate-logic-v2.md:132` — also says "one row per closed issue". Superseded algorithm-history doc, so low stakes, but it's the same wrong fact.
- `brain/.../calibration-findings.md` — the living analysis doc that *will* do future aggregate work over the ledger carries no dedupe note; a one-liner pointing at the SKILL.md:93 rule would close the loop the plan opened.
- The two brain edits need a manual commit in brain (the close gate never auto-commits there, per AGENTS.md §1) — worth confirming before declaring the production half landed. I could not check working-tree state.
- `throughput.go:14` — "one row per issue close" reads correctly as per-close, but sits three lines above a paragraph correcting exactly that phrasing; tightening it to "per CLOSE" would remove the ambiguity.

## 5. Test coverage notes

The re-close defect itself is well covered: `TestSpanThroughput_CountsReClosedIssueOnce` pins the numbers, `_ReCloseEqualsLastRowAlone` pins the property, `_UntrustedCountedOverDedupedSet` pins PQ-3's mixed-denominator hazard with a trusted-newest/untrusted-older fixture — that last one is a good test, it distinguishes "counted over deduped set" from "counted over any set". `TestSpanThroughput_28DaySpan` asserting both `Issues` and `RowsScanned` on a no-re-close fixture is the right call: it makes the coincidence explicit instead of implicit. `TestDrift_DedupesRepeatedIssueRows` (4 distinct of 5 rows, n=5 → no warn) still exercises the shared path.

Gaps, in order of what they'd catch: I4 (the persisted provenance count — the production-facing line, zero coverage), I5 (the shared helper's contract, indirect only). Also uncovered, and worth one cheap test since it's the ordering the code comment calls load-bearing: an out-of-span re-close (later date, outside `to`) must not mask the in-span row. Today nothing fails if the filter and dedupe stages are swapped in `SpanThroughput` — both re-close fixtures sit entirely inside the span.

I did not execute the suite (no shell available), so "tests pass" is not something I can assert; the main agent should run `go test ./cmd/...` and confirm `drift_test.go` is green unmodified before recording the verdict.

## 6. Architectural notes

- **ARCH-DRY — pass.** This is the finding the issue was really about, and it's delivered: one `NewestPerIssue`, two consumers, and `grep` over `*.go` shows no second dedupe (the old `seen`/`@row:%d` block is gone from `drift.go`, and `fmt` there is now used only by `DriftVerdict`'s messages). The doc comment explaining *why* the rule is direction-agnostic is the kind of note that stops the next divergence.
- **ARCH-PURE — pass.** `NewestPerIssue`, `SpanThroughput`, and `driftSample` are deterministic and IO-free; every test in `internal/estimate` runs on in-memory fixtures with no mocks. IO stays at `projectthroughput.go`/`close.go`. The pure/shell split is exactly where it should be.
- **ARCH-PURPOSE — flag (I1, I2).** The code half of the single-source change is complete and *enforced* — both readers derive. The documentation half is not: two hand-maintained restatements of the superseded model survive (the ledger file's own header, the `project throughput` helptext), and one of them is the very sentence PQ-2 called "the exact wrong fact this defect grew from." A restatement that doesn't derive is a deferred consumer, and here the deferred consumers are docs an agent reads *before* touching the ledger.
- **ARCH-MOCK — pass.** No external binary or service enters via this diff. `projectthroughput_test.go` boots the whole bless flow against a `t.TempDir()` brain — a portable folder standing in for production storage, with production and test sharing the `--brain-dir`/`WF_*` seam. That's the posture the principle asks for.
- **Forward note.** `SpanMeasure` now carries two counts with different denominators, and this diff already produced one mixed-denominator bug (PQ-3) and one surviving misnomer (I3). If a third count lands, consider making the deduped/raw distinction structural (a small `spanCounts` value that owns both and refuses to be mixed) rather than three sibling ints held apart by comment discipline.

## 7. Plan revision recommendations

One entry, on the docs scope:

> **## Revisions — <timestamp>: docs surfaces were five, not three.** The plan's docs row named `atlas/workflow/ledger-landscape.md`, `atlas/workflow/sdlc-binary.md`, and `brain/.../velocity/SKILL.md`. The close review found two further hand-maintained restatements of the superseded per-issue model: `brain/data/life/42shots/velocity/calibration-ledger.tsv:1-2` (the ledger's own header comment — not generated; `close.go:915` writes only the column header, and #118 precedent treats this comment as a maintained surface) and `cmd/sdlc/helptext/project.md:46,54` (the operator-facing `sdlc project throughput` help, which still says "summed from the calibration ledger" and documents `rows` unqualified). Both added to the docs scope.

The Plan's technical contract for `NewestPerIssue`, the pre-filtered/positional-key clauses, the baseline-schema non-change, and the named known limit all still describe the code accurately — no revision needed there. No Core concepts table exists (simple work, §1), so there is nothing to cross-check. The plan-gate ledger's `## Open findings` is empty and all five PQ findings are genuinely addressed in the code — PQ-3 in particular is now pinned by a test, not just by a comment.
