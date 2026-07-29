# Boundary Review — ariadne#187 (whole-issue close)

| field | value |
|-------|-------|
| issue | 187 — tune the change-code gate: stateful plan review, estimate after plan, churn metric |
| repo | ariadne |
| issue file | workshop/issues/000187-tune-change-code-gate.md |
| boundary | whole-issue close |
| milestone | — |
| window | af0193263534b3298af0352d021ef85049983f2e..HEAD |
| command | sdlc close --issue 187 |
| reviewer | claude |
| timestamp | 2026-07-29T15:33:22-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

M2 delivers what the Spec asked for and the replay genuinely answers the "did we weaken review" question — 2 rounds vs 6, with both must-survive findings landing in round 1 and a bug class moved from close review to plan review. `churn` and `gatestate` are properly pure (every convergence, classification and ratio test runs on in-memory strings, no mocks); the cost report is unconditional and tested against the three conditions that gate the ledger row; the ten ledger columns are appended with legacy rows and a 19-column malformed row both pinned. What keeps this from SHIP is four cheap gaps, all of the same shape — a consumer sweep that stopped one surface short. The most consequential: `appendCalibrationRow` writes the TSV header **only when the file is empty** (`close.go:915`), so every live ledger in the fleet keeps its 10-column header while new rows carry 20 columns. Separately, the new `finding` CUE noun was added without extending `construct/vocabulary/vet_test.sh` — the model gate that both prior nouns (#147 verdict, #180 project) explicitly extended, and the one place the "concrete blocks reach the export" invariant `finding.cue`'s whole design rests on is defended. **Caveat: Bash is unavailable in this session (harness EPERM on `~/.claude/session-env`), so I could not run `go test` / `go vet` / `vet_test.sh`; every finding below is from reading the files.**

## 1. Strengths

- **The unconditional-vs-gated split is exactly right, and tested at the seam that would have hidden it.** `close.go:784-805` emits the cost report in `applyClose` and passes `m` *into* `appendCalibrationRow` rather than recomputing; `TestChurnLinePrintsWhenLedgerRowIsSkipped` (`closemetrics_test.go:18`) drives all three conditions (milestone / `--no-actual` / no brain). That is the trap round-2 of the plan gate flagged, closed with the test that proves it.
- **`churnForWindow` distinguishes absent from broken, and both are pinned.** `TestChurnForWindowEmptyBase` and `TestChurnForWindowBadBaseErrors` (`churnreport_test.go:65,83`) are the two halves of PQ-2's finding; using `gitx.RunGit` over `gitx.Capture` is the difference between a warning and a silent zero, and the test asserts the error rather than the doc claiming it.
- **`ARCH-PURE` holds under the new package too.** `internal/churn` takes raw numstat text and returns a `Report`; the git invocations live at `churnreport.go`'s seam. `TestSummarizeEmptyWindowIsNotNaN` (`report_test.go:36`) checks with `math.IsNaN` rather than `!= 0`, and says why — a NaN would have passed the naive assertion.
- **`TotalInsertions` deliberately does not collapse by path**, with `TestTotalInsertionsCountsRepeatedPaths` pinning it (`report_test.go:88`). Deduplicating would have erased the one signal the metric exists for, and this is the guard that stops a future tidy-up from doing it.
- **The schema-vs-model tension is resolved honestly rather than papered over.** `churnreport.go:63-72` states plainly why column names *cannot* be model-derived, and `TestLedgerCoversEveryClosingDisposition` (`closemetrics_test.go:198`) fails vacuously-safe (it `t.Fatal`s if the model declares no closing dispositions). That is the right shape for "schema commitment + drift guard".
- **The carry-forward is live, not documented.** `code-review.md`'s plan-gate section reached me verbatim in this prompt — I read `000187-…-plan-gate.md` because the binary told me to. A3's demotion is genuinely safe.

## 2. Critical findings

None.

## 3. Important findings

**I1 — `cmd/sdlc/close.go:914-922`: existing calibration ledgers keep the 10-column header while new rows carry 20 columns; nothing upgrades it.**

```go
if len(existing) == 0 {
    buf.WriteString(estimate.Header() + "\n")
} else {
    buf.Write(existing)   // header untouched, forever
}
```

`estimate.Header()` grew from 10 to 20 columns (`internal/estimate/ledger.go:65-67`), but the header is written **only on file creation**. Every calibration ledger already in existence — this repo's sibling `brain/data/life/42shots/velocity/calibration-ledger.tsv`, and every downstream repo's — was created under the old binary, so from this issue forward it accumulates 20-column rows under a 10-column header. `ParseRows` is unaffected (it skips the header by `strings.HasPrefix(line, "issue\t")`, `ledger.go:93`), so Go consumers are fine; the breakage is for the ledger's *stated primary* consumer.

Failure scenario: the atlas designates the ledger row as **authoritative** for these metrics and the printed lines as the "human mirror" (`atlas/workflow/ledger-landscape.md`, the #187 block). An analyst loads it with `pandas.read_csv(path, sep='\t')` — pandas raises `ParserError: Too many columns specified: expected 10 and found 20` (or, with `header=None`, silently mislabels). A spreadsheet import leaves `churn_prod … gate_open` as ten unlabeled columns. The ten new metrics are unreadable in exactly the file the atlas says to read them from, on every ledger except a brand-new one — and no test covers appending to a legacy-header ledger (`close_ledger_test.go` writes only to fresh paths).

Fix sketch: in `appendCalibrationRow`, when `existing` is non-empty and its first line is a *known prefix* of `estimate.Header()` (i.e. a legacy header), replace that first line with `estimate.Header()` and leave every data row byte-identical. Idempotent, and it preserves the "never clobber prior rows" rule the function already states at `:910`. Add the missing test: seed a ledger with the 10-column header + one legacy row, append, assert the header is now 20 columns **and** the legacy row is unchanged.

**I2 — `construct/vocabulary/vet_test.sh`: the new `finding` noun is outside the model gate every other noun sits inside.**

The script enumerates nouns explicitly — `issue.cue` (`:9-17`), `verdict.cue` (`:21-23`), `project.cue` (`:28-37`) — and #187 added `construct/vocabulary/finding.cue` without a block. Two things follow:

- `cue vet finding.cue` never runs in the gate. A constraint error surfaces only when someone runs `make vocab-embed`, where `go generate`'s `vocabulary export --noun finding > finding.json` truncates the file on failure and `mustLoadFinding` then **panics on every `sdlc` invocation** (`pkg/vocab/finding.go:31`).
- More pointedly, the script's own header calls the export assertion "load-bearing: CUE `#`-definitions don't `cue export`, so this guards the concrete data sdlc actually consumes." `finding.cue`'s design depends on precisely that distinction — its header says so ("Only concrete data reaches the exported JSON — CUE definitions (#) do not export") and `#Severity`/`#Disposition`/`#Finding`/`#Dispose` are all `#`-defs. The one noun most dependent on the invariant is the one not checked for it. Both prior nouns extended this file as part of their own issue (`#147`: "mirror `construct/vocabulary/vet_test.sh`"; `#180` Task Step 3).

Note the Go side is not fully unguarded: `TestFindingSeverityOrder` (`pkg/vocab/finding_test.go:118`) asserts `Severities() == [Critical Important Minor]` with a length check, which would fail on an empty export. Worth knowing because `TestFindingConformance` would **pass vacuously** on an empty model — all its loops range over model-derived slices, and the two negative assertions (`IsSeverity("Catastrophic")`, `IsDisposition("deferred")`) are satisfied by an empty model.

Fix sketch: append a `finding` block mirroring the verdict one — `cue vet finding.cue`, then `grep -q '"categories"'`, `'"dispositions"'`, `'"hardBlocking"'` on the export. Cheap, and it is where the next reader will look.

**I3 — `atlas/workflow/vocabulary.md`: the noun registry page was not updated for `finding.cue` (Docs update gate).**

That page's entire job is enumerating the nouns: `issue.cue` (`:13`), `project.cue` (`:19`), `verdict.cue` (`:38`), `pensive.cue` (`:167`), plus `vet_test.sh` described at `:48` as covering "the valid issue/project models". A brand-new noun plus its Go binding landed in this window with no entry, so a reader asking "what nouns exist?" gets a stale answer from the page designed to answer it. `atlas/workflow/gate-state.md` documents the finding *semantics* well — this is specifically the registry gap.

Fix sketch: add a `finding.cue` bullet after the `verdict.cue` one, on the same shape (categories as the single concrete source; `#Severity`/`#Disposition` derived; the consumers that derive — prompt via `RenderBlockInstruction`, `ParseFindingsBlock`, `Decide`, `DispositionCounts`; the second realization of the `agent-binary-handoff-schema` target). Update `:48`'s "valid issue/project models" once I2 lands. `finding.json` needs no Makefile work — `vocab-embed` is generic over nouns (`Makefile.workflow:869-886`), which is correct and worth keeping.

**I4 — `cmd/sdlc/changecode.go:500-510`: the `ApplyChecked` protocol-error path discards the judge's valid new findings, and the next round's prompt then claims they were all disposed.**

When `ApplyChecked` rejects a round (a disposition naming an id we never issued — the exact case it exists to catch), the persisted round carries `ProtocolError` and **no findings**; `rr.New` is dropped on the floor even though every finding in it already passed `ParseFindingsBlock`'s severity + non-blank-title validation.

Failure scenario: the judge raises 3 real Critical/Important findings *and* hallucinates one disposition id. The gate blocks (correct), but the three findings are never recorded. On re-run, `RenderPriorFindings` reports `1 prior round(s)` and then prints `(none — every prior finding has been disposed)` — a positive claim that is false: nothing was ever recorded, let alone disposed. The gate's memory asserts a clean slate it does not have, in the artifact whose sole purpose is not losing findings, and `gate_addressed`/`gate_open` under-report the round for the close-time metric. Untested — `TestPlanQualityFallsBackWithoutFindingsBlock` covers the *no-block* path, nothing covers the *invalid-disposition* path.

Fix sketch: two independent halves, either of which helps and both of which are cheap. (a) In the `aerr` branch, persist `gatestate.AssignIDs(ledger, gatestate.RoundReport{New: rr.New}, …)`'s findings alongside the `ProtocolError`, dropping only the invalid dispositions — the findings are already validated. (b) In `gatestate/prompt.go`, make the `len(open) == 0` message conditional on some finding having ever been raised, so a findings-less history reads "no findings have been recorded yet" rather than "every prior finding has been disposed". Add a test asserting a hallucinated-id round still records the findings it legitimately raised.

## 4. Minor findings

- **Carry-forward PQ-3 (Minor, still open).** `workshop/plans/000187-tune-change-code-gate-plan.md:2440` still reads "the calibration-ledger schema row gains the **seven** columns". `:2388` is correct ("TEN"), and the issue's M2 Plan row (`:229`) is now correct — this is the last survivor, and the M1 Revisions entry already claims it fixed. Fix in place to "ten".
- `cmd/sdlc/changecode.go:3-12` — package doc still says "Composes four gates" and lists the **pre-B1** order (estimate #2, plan-quality #3). There are now five gates plus branching, plan-quality second. First thing a reader of the most-changed file sees; flagged at M1 and still unfixed.
- `cmd/sdlc/changecode.go:141` — the step comments jump `// 3. Run the gate sequence` → `// 5. Branching strategy`; step 4 vanished with the consolidation.
- `planGateContent` (`changecode.go:606-614`) scans for `## ` line-wise without fence awareness, so a line beginning `## Estimate` **inside a fenced block** in the issue body would strip real content out of the pass-through hash — meaning an edit to that content would not re-dispatch the judge. This matches existing house behavior (`issue.SectionBody`'s regex isn't fence-aware either), so it's consistency rather than novelty — but the repo does have `issue.SplitFences` if it's ever worth tightening.
- `gatestate/ledger.go`'s `nextIDSeq` names its accumulator `max`, shadowing the builtin — while `Decide` in the same package deliberately spells out `roundCap` "so it doesn't shadow the builtin". Pick one convention.
- `churn.ClassifyPath` — a cross-directory rename renders in numstat as `{atlas => docs}/x.md`, whose first segment is `{atlas => docs}`, so it buckets to `CodeProd`. The segment-not-substring rule is right; this is the one input shape it doesn't cover.
- `cmd/sdlc/planreview.go`'s `readPlanGateLedger` — with `--name` and no `--issue`, `f.Issue` is 0, so identity repair stamps `issue: 0` and `Render` emits `# Gate ledger — ariadne#0 (plan-quality)`. Carried from the M1 review; still live.
- `gatestate.DispositionCounts` seeds `counts` from `AllDispositions()`, so the returned map carries a `not-addressed` key that is never incremented (those findings go to `open`). Harmless — `closeMetrics` reads only two keys — but a caller iterating `counts` sees a permanent zero bucket.

## 5. Test coverage notes

- **I could not execute anything.** No `go test ./...`, `go vet ./...`, or `sh construct/vocabulary/vet_test.sh`. The close's `--verified` evidence should carry an actual green run, including the two fuzz targets.
- **Verified statically instead, where it was cheap:** the B2 semantic sweep is genuinely clean — running `estimate.{0,80}start-plan|start-plan.{0,80}estimate` across the tree excluding `workshop/**` returns only allowlisted files (`estimate-source.md`, `startplan.go`, `startplan_test.go`, `estimatetiming_test.go`, `sdlc-binary.md:355`, all no-timing-claim or asserted). `TestEstimateTimingConsistency` should pass, and `TestEstimateTimingStatedPositively`'s `strings.Index` ordering check on the `sdlc-binary.md` gate row is satisfied (`plan-quality` precedes `estimate (#113)`). `TestResolvePlansDirDefaults`'s `t.Setenv(…, "")` is safe because `envOr` (`term.go:114`) treats empty as unset.
- **Gaps named above:** no legacy-header append test (I1), no `vet_test.sh` coverage for `finding.cue` (I2), no test for the invalid-disposition protocol path (I4).
- `TestCloseLedgerRowCarriesCostMetrics` is the right end-to-end shape — real commits, real ledger, and it asserts the operator line and the TSV row agree rather than checking them separately. Its rework assertion derives `wantRework` from the reported `ChurnWorkshop` instead of hard-coding the fixture's line count, which is the correct call.
- `runChangeCode`'s gate loop still has no in-process coverage (`exitWithCode` → `os.Exit`, unlike the swappable `die`). Deferring the seam change was the right judgement for a FIX-THEN-SHIP commit, and `TestForceAckMatchesGateCatalog` covers the message contract that was the actual risk — but the `--force` continuation branch at `changecode.go:134-138` remains untested. Worth its own issue rather than another deferral note.

## 6. Architectural notes for upcoming work

- **ARCH-DRY — pass.** The extractions in this window are the right ones: `resolvePlansDir` as the single plans-dir resolution (and the two inline `envOr` calls at `close.go:1012,1045` switched to it), metrics passed into `appendCalibrationRow` rather than recomputed, `boundaryWindowBase` reused rather than re-derived, `contains` promoted from the test file into `pkg/vocab/lifecycle.go`, `stubJudgeSeq` extending the existing `judge.Run` seam, and `commitFile` reusing `mkArtifact` + `closeRepo`. `DispositionCounts` is now model-derived (M1's I4, confirmed fixed at `gatestate/ledger.go`).
- **ARCH-PURE — pass, and it is the diff's strongest property.** `internal/churn` and `internal/gatestate` hold no fs, clock or subprocess; `churnreport.go` and `planreview.go` are the two seams. `closeMetrics` is the only place the two meet, and it returns a value type with pure formatters hanging off it. No change recommended.
- **ARCH-PURPOSE — flag (I1, I2, I3).** Shadow-sweep on the finding model: prompt ✓, parser ✓, `Decide` ✓, `DispositionCounts` ✓, `code-review.md` severities ✓ (drift-tested), ledger columns ✓ (literals, but guarded). Sweep on the *noun*: `vet_test.sh` ✗, `atlas/workflow/vocabulary.md` ✗. Sweep on the *ledger format*: writer ✓, reader ✓, header ✗. The recurring pattern across both M1 and M2 boundaries is the same — the derivation lands, and one consumer of the changed thing is left behind. For #183, the concrete carry-forward: treat the **artifact's own header/schema line** and the **noun registry** as consumers, not documentation. Both misses here are files that describe the thing rather than use it, which is exactly why the sweep skipped them.
- **ARCH-MOCK — pass.** No new external dependency. `git` is exercised for real in a disposable repo (`closeRepo` + `commitFile` + `headSHA`), which is stronger than a fake for a binary this well-specified; production and test flow share `gitx.RunGit`. Risk 5's live conformance for the ` ```findings ` fence is discharged for `claude` by the replay (both rounds `protocol_error=""`, five dispositions round-tripped with notes intact); `codex`/`gemini` remain covered only by the prose fallback, with a stated fail-closed trigger (<5% drop the fallback / >20% simplify the schema) and — now — `gate_rounds` + persisted `protocol_error` rounds in the ledger to actually measure it. That is the correct posture and the measurement is now wired, which is what M2 owed.
- **For #183:** `Ledger.Gate`/`IDPrefix`-as-data has held up across one consumer; the thing to settle before the second lands is `ContentHash`'s *scope*. `planGateContent` shows the hashed content is a per-gate decision (what that gate reviews), not a universal `issue+plan` — consider a per-gate content extractor supplied by the shell rather than a second hash field.

## 7. Plan revision recommendations

- **Fix PQ-3 in place** at `workshop/plans/000187-tune-change-code-gate-plan.md:2440`: "seven" → "ten". It is the only surviving instance and the M1 Revisions entry already claims it closed, so leaving it makes two artifacts disagree about a third.
- **Add a `## Revisions` entry recording the M2-as-shipped deltas**, so the plan stops claiming what the code doesn't deliver:
  - *Task 1 shipped incomplete.* `construct/vocabulary/vet_test.sh` was never extended for the new noun, and `atlas/workflow/vocabulary.md` — the noun registry the page-per-noun convention lives in — was not updated. Task 1's file list named neither, and both prior nouns (#147, #180) treated `vet_test.sh` as a mandatory step of adding a noun. Record them as the task's missing surfaces plus the export assertions (`categories` / `dispositions` / `hardBlocking`).
  - *Task 13's ledger append is write-only for existing files.* The plan's column-append discipline correctly protects the *reader*; it does not state that `appendCalibrationRow` writes the header only on creation, so the ten new columns are unlabeled in every pre-existing ledger. Record the header-upgrade requirement and the missing legacy-header append test.
  - *Task 8's protocol-error path discards validated findings.* The plan specifies persisting a findings-less round for the *no-block* case (correctly, with the `len(Rounds)` reasoning) and reuses that shape for the `ApplyChecked` failure, where `rr.New` has already been validated. Record that the two cases differ and that the second should keep its findings.
