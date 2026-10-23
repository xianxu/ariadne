# Boundary Review — ariadne#561 (milestone M2)

| field | value |
|-------|-------|
| issue | 561 — Landing trains: cars board on affected tests; full suite once per train; measure escapes and bisect cost |
| repo | ariadne |
| issue file | workshop/issues/000561-landing-trains-cars-board-on-affected-tests-full-suite-once-per-train-measure-escapes-and-bisect-cost.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 07bdeb79d37bc188b3f65b96c5cbfdaf4be8617e..1b6dd98e497c24edbbfe8a05b53e8690b168d8e2 |
| command | sdlc milestone-close --issue 561 --milestone M2 |
| reviewer | claude (anthropic; author anthropic,openai) |
| isolation | fenced |
| timestamp | 2026-10-23T02:10:55-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Reviewing the #561 M2 diff — the core code is read; next I'm going through the new measurement module and the conductor wiring.

Still reviewing — checking how the weekly snapshot reads landing events, then the train test fixtures.

Targeted tests pass. Now checking status/weekly test coverage and the trainmetrics tests.

Cross-checking the plan's Core concepts table against the code.

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

M2 delivers what the issue asks for, and all of it works. Each car's boarding proof is now taken from the same CI gate decision that vouched it. A pure measurement tracker records escapes and every whole-suite run. One pure fold and one formatter feed both `sdlc landing status` and the weekly `train_filter`. I ran the targeted suites at the pinned head and they pass: `go test ./cmd/sdlc/internal/landingline -count=1` ok, and `go test ./cmd/sdlc -run 'Test.*(Affected|CarBoards|CarFullSuite|LandingTrain|SplitAssessment|ReadSplitEvidence|TrainMetric|TrainMeasure|MeasureExecution|Weekly|LandingStatus|CarProof|LocalCheckRunCommand|CIGate)' -timeout 20m -count=1` ok (284s). Build and vet are clean. I found no correctness bug that blocks shipping. Three things should be fixed first:
- **README:** the paragraph on what `sdlc landing status` shows doesn't mention the new train-filter line, which is the Done-when surface.
- **Tests:** the plan promised fuzz tests and a native interrupted-append test; neither exists.
- **Plan table:** the Core concepts table names locations and changes that don't match the code.

## 1. Strengths
- **Proof comes from the gate's own decision.** `ciVerdict.Accepted` is filled in by `decideCIGate` itself (`cmd/sdlc/cigate.go:136-109`), and `localRecordsOnAncestors` now records the nearest docs-only ancestor. No second GitHub lookup is made. `TestCarProofNamesTheAcceptedRecord` checks a record on the head, one on a docs-only ancestor, a suite record winning over affected, and a waived gate.
- **The proof binding survives older binaries safely.** `Vouch` stamps the proof with the vouch time (`line.go:375-379`). `AcceptedProof` requires that stamp to match, and `carryProof` keeps unknown nested fields only when check, head and source are unchanged. When an older sdlc rewrites the vouch, the proof reads as unknown and the vouch stays valid. `TestVouchProofBinding` and `TestBoardingProofCodec` pin this, including duplicate, null and trailing-data rejection.
- **The measurement tracker really has one owner.** Its state is unexported, `AdvanceMeasurement` is the only transition, incompleteness is sticky, and a finalized tracker is frozen. `TestAdvanceMeasurementRandomSequences` checks these properties with a seeded generator, so a failing ordering can be reproduced.
- **The native escape test compares against real runs.** `TestLandingTrainMeasuresAnAffectedEscape` (`landingtrain_metrics_test.go:22`) checks the measured attempt IDs against the attempts actually saved in the test-stats store, not against scheduled effects.
- **One escape classification for blame and measurement.** `ClassifySplit` returns `Escapes`, and nothing parses human refusal text.

## 2. Critical findings
None.

## 3. Important findings
- **README not updated for the new status line.** `README.md:729-731` still lists what `sdlc landing status` shows as "the setting, the next departure, each waiting merge's readiness, and the last trains". The new seven-day train-filter line (escape rate, coverage, split cost) is missing, and so is the weekly `train_filter` field. Help text and atlas were updated. Fix: add one sentence there and point to `sdlc landing --help`.
- **Promised test strategies were not delivered (ARCH-PURPOSE).** The plan's function-strategy table names:
  - fuzz tests for the `BoardingProof` codec, `DecodeReadiness` and `Line.Vouch`;
  - a fuzz test of `AggregateTrainMeasurements` against a simple reference fold;
  - an injected interrupted append inside a native train, showing the landing outcome doesn't change.
  
  None of these exist: there is no `Fuzz*` for proof or aggregation. The aggregation is covered by one example test. `TestTrainMeasurementAppendFailure` drives a bare `conductor` with no train. Fix: add a `FuzzTrainMeasurementAggregate` with a reference fold and a `FuzzBoardingProof` round trip. Run the existing native train with an unwritable events path and assert the landing verdicts are unchanged.
- **The Core concepts table doesn't match the code.**
  - It lists `BoardingProof` in `internal/landingline/train.go`; it is in `trainmetrics.go:30`.
  - It lists `assessSplit` in `landingtrain_attribution.go` as "modified"; that file isn't in the diff, because `Escapes` flows through the `ClassifySplit` return value.
  
  This doesn't change behaviour. I'm rating it Important rather than Critical because it's file placement within one package. Fix it with a Revisions entry (section 7).

## 4. Minor findings
- **In-flight trains count as incomplete.** A running train has a start record and no final, so status shows it as incomplete. A weekly snapshot taken while a train is running freezes it as incomplete forever. Consider labelling "in flight" separately.
- **Escape names aren't shown anywhere readable.** The plan's visible-state row promises "Escape, named car/head/tests". The names are in the log, but status and weekly only show counts.
- **Run kind passes through a mutable field.** `c.suiteKind` is set before `execute` and read inside `runSuite` (`landingtrain.go:162-171`). Pass it as an argument instead.
- **Reused saved passes may count as runs.** `newLocalCheckRunner` can republish a saved pass instead of running. That would be measured as an executed run with no attempt ID and unknown cost. This is unlikely on freshly composed train heads.
- **Lenient decode of untrusted log data (ARCH-SECURE).** `aggregateTrainMeasurements` decodes the log payload with plain `json.Unmarshal`, so duplicate keys and `"cars": null` are accepted. The payload is evidence only, so the impact is low.
- **Duplicate test names can double-count.** `ClassifySplit` `Escapes` and `EscapeFailures` don't deduplicate a test that is reported twice.
- **Oversized records lose their cars.** When `EncodeMeasurement` replaces an over-limit record with its minimal fallback, the car list is dropped. The record is marked incomplete, but affected boardings for that train are undercounted rather than shown as unresolved.
- **"No suite" policy resolves cars without a run.** When a repo's policy declares no suite command, `runSuite` reports a pass without running anything. That pass still marks cars as resolved, inflating the escape-rate denominator.
- **Filter hidden when the setting is unreadable.** If the train setting can't be read (`serr != nil`), the filter line is hidden even when the window contains trains.

## 5. Test coverage notes
- What's covered well: the tracker lifecycle (example, sticky, random-sequence and bounds tests); fold semantics (duplicates, conflicts, unsupported versions, the half-open window, legacy trains); the N/2N linear work tests; the native two-car escape; weekly immutability of a pre-change artifact; and the status line.
- Gaps: no fuzz tests (above); no cancellation or startup failure driven through the real `runLocalCheck` (only synthetic `localRunRecord`s); and no reintegration-suite or drop-cancelled-suite classification test for `ExecReintegration`.

## 6. Architecture (each principle)
- **ARCH-DRY: pass.** `localCheckRunCommand` gives every run command one source, enforced by a test. One fold and one formatter serve both status and weekly.
- **ARCH-PURE: pass.** Lifecycle, fold, formatting and proof binding are pure. `landingtrain_metrics.go` is a thin layer that only appends.
- **ARCH-PURPOSE: flag.** The promised test strategies were not delivered, and the README surface is missing (section 3).
- **ARCH-MOCK: pass.** The tests reuse the native bare remotes, the stateful fake GitHub and a deterministic fake suite.
- **ARCH-CONSTRAINTS: pass.** Each record is bounded (16/128/256 entries, 256 KiB), and the fold's linear work is tested at N and 2N.
- **ARCH-SECURE: minor flag.** The log payload is decoded leniently (section 4). The proof binding and `fullOID` check are good.
- **ARCH-ORDER: pass, one note.** All transitions go through `AdvanceMeasurement`. A dead process leaves a start with no final, which reads as incomplete. The `suiteKind` side channel is the only state outside the model.
- **ARCH-FUNERAL: pass.** `train_measure` events fall under the existing 30-day `KeepEvents` sweep, and the weekly field is small and bounded.
- **ARCH-EVOLVE: pass.** `train_filter` is an optional addition and old artifacts are never rewritten. The nested proof codec is strict and preserves unknown fields. The raw log payload is never rewritten.

## 7. Plan revision recommendations
- "## Revisions — M2: `BoardingProof`, `AcceptedProof` and `carryProof` live in `internal/landingline/trainmetrics.go`, not `train.go`. `assessSplit` is unchanged, because `SplitAttribution.Escapes` reaches it through the `ClassifySplit` return value."
- If the fuzz strategies are deliberately dropped instead of added, record that and name the replacement tests (`TestAdvanceMeasurementRandomSequences`, `TestAggregateTrainMeasurements`).

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: readme-tracks-user-surface
    title: |
      README landing-status paragraph omits the new train filter line and weekly train_filter
    detail: |
      README.md:729-731 lists what `sdlc landing status` shows but not the seven-day escape-rate/split-cost line, which is the issue's Done-when surface, nor the weekly snapshot's train_filter. Help text and atlas were updated in this range; README was not.
  - id: new
    severity: Important
    kind: defect
    family: plan-test-strategy-undelivered
    title: |
      Planned fuzz and native interrupted-append strategies for proof codec and aggregation are absent
    detail: |
      The plan's strategy table requires fuzzing BoardingProof/DecodeReadiness/Vouch and fuzzing AggregateTrainMeasurements against a reference fold, plus an interrupted append in a native train with landing verdicts unchanged. No Fuzz* exists for either. TestTrainMeasurementAppendFailure uses a bare conductor with no train. Add the fuzz targets and a native train run with an unwritable event log.
  - id: new
    severity: Important
    kind: defect
    family: plan-table-tracks-code
    title: |
      Core concepts table places BoardingProof in train.go and marks assessSplit modified; code differs
    detail: |
      BoardingProof, AcceptedProof and carryProof are in internal/landingline/trainmetrics.go. landingtrain_attribution.go (assessSplit) is untouched, because Escapes flows through the ClassifySplit return value. Add a Revisions entry.
  - id: new
    severity: Minor
    kind: defect
    family: measurement-coverage-labels
    title: |
      An in-flight train reads as incomplete in status and can be frozen as incomplete in a weekly artifact
    detail: |
      A start without a final is folded as incomplete even while its conductor is still running. Consider an in-flight label for trains still under way.
  - id: new
    severity: Minor
    kind: defect
    family: measurement-coverage-labels
    title: |
      Escape identities (car/head/tests) are not shown on any human surface
    detail: |
      The plan's visible-state row promises a named car/head/tests for an escape. Status and weekly show only counts; the names live only in raw train_measure events.
  - id: new
    severity: Minor
    kind: defect
    family: lenient-decode-untrusted-log
    title: |
      Measurement payloads are decoded with plain json.Unmarshal (duplicate keys and null cars accepted)
    detail: |
      aggregateTrainMeasurements reads hand-editable log payloads leniently. Evidence-only, so impact is low; the codecs elsewhere in the package are strict.
  - id: new
    severity: Minor
    kind: defect
    family: measurement-execution-provenance
    title: |
      Run kind is passed through the mutable c.suiteKind field; a republished saved pass could be counted as an execution
    detail: |
      landingtrain.go:162-171 sets the kind before execute and runSuite reads it. newLocalCheckRunner can republish a saved pass, which would be measured as an executed run with no attempt ID. Pass the kind explicitly and skip measuring republished passes.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"f0835410-4eb2-4880-8c4f-36c6901f18c3","repository":"github.com/xianxu/ariadne","issue_id":"000561","recipe":"milestone-review","round":1,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/ebe409e06f42128e","reviewed_head":"1b6dd98e497c24edbbfe8a05b53e8690b168d8e2","slot":"ariadne:2","requested_at":"2026-10-23T09:01:12.299943Z"},"phase":"finished","pid":90799,"started_at":"2026-10-23T09:01:12.404177Z","finished_at":"2026-10-23T09:10:54.982306Z","session_id":"2f3ba45e-c27f-462b-aaff-a3123e3fc736","outcome":"returned"}] -->
