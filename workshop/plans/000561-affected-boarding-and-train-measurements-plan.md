---
collaboration-mode: discussion
---

# Affected boarding and train measurements implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development for bounded independent tasks, or superpowers-executing-plans for work depending on current context. Steps use checkbox syntax for tracking.

**Goal:** Let cars prepare for a train without launching individual full suites, preserve the train's full-suite gate, and give the TL evidence for a cautious two-car rollout.

**Architecture:** Keep main-authoritative selection and the existing train reducer. Add an explicit affected boarding execution mode, preserve which check actually satisfied readiness, and collect bounded measurements alongside existing landing events. One pure aggregation and formatting path serves status and weekly artifacts; measurements never grant landing authority.

**Tech stack:** Go, Cobra, existing affected selector, local-check runner, landing-line JSONL storage, native Git fixtures and stateful fake GitHub.

## Approved scope and evidence

The operator's issue and the hub's 2026-10-23 edge decisions govern this plan. Cars already accept affected commit statuses in `carCIGateFacts`; the unwanted full runs occur in `runAffectedCheck` when selection returns Full, including graph/preparation failures. Therefore do not weaken or duplicate readiness, broaden Full into all known packages, or invoke a native filter with unknown paths.

Add `sdlc local-check run --name affected --boarding`. In this mode a Full/unavailable selection refuses before command execution or passing-status publication. Its error identifies the selection reason and the affected-map/preparation repair, and prints `sdlc local-check run --name suite --rev SHA` followed by `sdlc merge --yes` as the full-proof alternative. Serial landing requires the TL's guarded main configuration change setting `train_interval_minutes` to zero first; name that prerequisite explicitly. There is no per-call serial flag, and this issue adds none. Ordinary affected execution retains its full-suite fallback. Already-passing full-suite evidence remains valid for boarding. Partial selections and proven docs-only selections retain existing behavior. Validate the boarding flag with affected alone; reject suite, mixed names and all-check requests rather than silently ignoring it.

The default cap becomes two when `train_max_cars` is omitted, including every new train adopter. Explicit caps remain unchanged; absent or zero interval still disables trains. One central policy states: review a complete seven-day measurement window before manually raising the repo's cap; show escape counts, denominator, unknown coverage and split costs; never raise automatically. No numeric quality threshold is invented. The guarded main configuration remains the only cap/interval authority. A TL can require full suites again by restoring serial landing through that guarded configuration; this issue neither edits consumer configuration nor labels its own PR.

“Once per train” means one initial full integration suite for a healthy train. Existing main-baseline checks, failing-prefix probes, registered-flake recovery and code reintegration checks remain necessary and are measured separately. No gate, attribution rule, integration freshness rule or bounded retry rule is waived.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|---|---|---|
| `affectedExecutionMode` | `cmd/sdlc/localcheck_affected.go` | new |
| `BoardingProof` | `cmd/sdlc/internal/landingline/train.go` | new |
| `ciVerdict` | `cmd/sdlc/cigate.go` | modified |
| `Readiness` | `cmd/sdlc/internal/landingline/train.go` | modified |
| `SplitAttribution` | `cmd/sdlc/internal/landingline/attribution.go` | modified |
| `TrainMeasurement`, `TrainFilterStats` | `cmd/sdlc/internal/landingline/trainmetrics.go` | new |
| `MeasurementTracker`, `MeasurementEvent`, `AdvanceMeasurement` | `cmd/sdlc/internal/landingline/trainmetrics.go` | new |
| `AggregateTrainMeasurements`, `FormatTrainFilterStats` | `cmd/sdlc/internal/landingline/trainmetrics.go` | new |
| `DefaultMaxCars`, `TrainRampReviewWindow`, `FormatTrainRampPolicy` | `cmd/sdlc/internal/landingline/train.go`, new `cmd/sdlc/internal/landingline/trainpolicy.go` | modified / new |
| `WeeklyRepoStats` | `cmd/sdlc/weeklystats.go` | modified |

- `affectedExecutionMode` has ordinary and boarding values. Selection remains owned by `affected.Plan`; the mode decides only whether Full executes suite or refuses. This is the smallest alternative to changing selector authority (ARCH-DRY).
- `BoardingProof` records the accepted check kind (affected or suite), car head, source head for permitted ancestor evidence, status context and observation binding. `ciVerdict` exposes the accepted source from its existing decision, not a second GitHub lookup. A readiness owns at most one suite-requirement proof; a departed measurement copies each original car's proof. Full-suite proof is valid readiness but outside the affected-filter denominator. Unknown proof affects measurement coverage, never retroactively waives or adds a gate.
- Readiness's optional proof must match its head and observation stamp. Older writers can preserve unknown JSON while changing readiness: a mismatched stamp makes the measurement proof unknown. Preserve existing `Extra` members and the readiness envelope version; optional data cannot make a previously invalid readiness valid. `BoardingProof.UnmarshalJSON` uses existing `decodeStrict`, and `BoardingProof.MarshalJSON` uses `encodeWithExtra` with its own `Extra` map: the nested object rejects duplicate known/unknown keys, null known values, wrong types and trailing data, and preserves future nested members. Keep its shape flat to avoid unowned deeper codecs. `Readiness` registers this custom-decoded object in `readinessFields`. Re-vouch merges unknown proof members only when the existing head/source/check binding is unchanged; a new binding starts a new proof and does not carry old attached metadata. An older proof-aware rewrite must preserve newer unknown members at both levels, without letting `Extra` override known facts. Apply the same strict, preserving object contract to every new structured payload that a writer reads and rewrites; raw immutable event retention continues to preserve bytes without reconstruction.
- `SplitAttribution` retains all qualifying, exact, unregistered affected failure identities alongside its existing kind/reason. The same pure classification owns blame and measurement inputs. Do not parse human refusal strings or count registered flakes, untouched packages, unknown impact, stale/missing evidence or main-red outcomes as escapes (ARCH-PURE).
- A `TrainMeasurement` owns original car identities, their boarding proofs, coverage outcomes, qualifying failure identities, and execution costs for one train. A car boarding is resolved when a verified passing prefix covers it or existing exact adjacent-prefix attribution confirms it as culprit. Released/unverified cars are unresolved. One car with several qualifying failures contributes one escaped boarding and retains each exact failure identity. Repeated trains are separate exposures; repeated records of one train are not.
- Escape rate is `escaped affected-pass car-boardings / resolved affected-pass car-boardings`; zero denominator means unknown, not zero percent. Show unresolved boardings and legacy/unmeasured trains beside the fraction to prevent a biased fraction appearing complete. An escape describes a gate miss, not proof that an individual test was omitted or that an unregistered intermittent failure is deterministic. No claim about train interaction causality beyond existing attribution is added.
- `TrainFilterStats` is a pure fold over a supplied time window, with one formatter for status and weekly summary text. Relationships are one train to many cars/attempts and one repo/window to many trains. Future check kinds or measurement versions widen this typed payload, not the landing state machine.

### Integration points

| Name | Lives in | Status | Wraps |
|---|---|---|---|
| `runLocalCheckCmd`, `runAffectedCheck` | `cmd/sdlc/localcheck_cmd.go`, `cmd/sdlc/localcheck_affected.go` | modified | CLI, Git selection and detached execution |
| `carReadiness`, `productionCarGates` | `cmd/sdlc/landingcar.go` | modified | existing prepared CI decision and readiness publication |
| `localRunRecord`, `runLocalCheck` | `cmd/sdlc/localchecks.go`, `cmd/sdlc/localcheck_run.go` | modified | existing managed execution and persisted attempt identity/timing |
| `conductor` measurement ownership | `cmd/sdlc/landingtrain.go`, new `cmd/sdlc/landingtrain_metrics.go` | modified / new | initial, baseline, prefix and reintegration executions |
| `assessSplit` | `cmd/sdlc/landingtrain_attribution.go` | modified | hash-verified logs, dependency impact and fresh registry |
| `Event` measurement payload and record kind | `cmd/sdlc/internal/landingline/log.go` | modified | existing locked JSONL log and retention |
| `runLandingStatus` | `cmd/sdlc/landingcmd.go` | modified | repository landing events and main configuration |
| `runWeeklySnapshot` | `cmd/sdlc/weeklystats_cmd.go` | modified | existing event evidence and immutable weekly artifact |

Reuse `affectedFixture`, native bare-remotes and `trainFakeGH`; extend their persisted checks, statuses and suite outcomes rather than replacing them with argv-only mocks. They model pushed heads, open train/car PRs, suite executions, merges and failures. No new service, daemon, lock or external API is needed. GitHub reads remain behind the existing prepared CI seam; the exact accepted result is captured while that decision is made. The existing landing owner and log mutex serialize control and append effects; observation cannot replace those lifetime resources (ARCH-DRY).

## Measurement persistence and costs

Record a version-1 optional measurement payload using a distinct measurement event kind in the existing log: one start record after departure identities are fixed and one final record on every controlled exit. Existing terminal `train` events, scheduling and admission semantics remain unchanged. Join by train ID and departure identity; reject conflicting duplicates as unknown, fold identical duplicates once. A start without final result is incomplete. A final without a valid start is incomplete. If the process dies before writing either record, any later terminal train event without measurement is legacy/unknown; do not infer unobserved work succeeded. A wholly missing event cannot be reconstructed and is an explicit coverage limit.

`AdvanceMeasurement(tracker, event)` is the sole pure transition owner. `MeasurementTracker` has unexported accumulation fields and returns an immutable `TrainMeasurement` snapshot only through finalization. Its events are departure initialization, start-record result, managed execution observation, verified-prefix resolution, confirmed attribution and terminal finalization. Only `AdvanceMeasurement` can update coverage/cost state; the conductor supplies facts and executes append effects, never changes counters directly. The lifecycle is independent of landing authority (ARCH-ORDER, ARCH-PURE):

| Current phase | Event | Next phase / effect |
|---|---|---|
| Empty | Initialize original departure identities | Starting; emit start snapshot |
| Starting | Start append acknowledged or failed | Collecting; retain acknowledged flag or sticky incomplete reason |
| Collecting | Execution observed, prefix verified or attribution confirmed | Collecting; validate identity, dedupe observations, update only justified costs/outcomes |
| Collecting | Terminal finalization, including controlled abort/cancellation | Finalized; freeze terminal snapshot, preserve unresolved coverage and sticky incomplete reasons |
| Finalized | Same finalization again | Same immutable snapshot; no second effect |
| Not finalized | Event illegal for phase or contradictory fact | Retain valid facts, mark sticky incomplete; no invented execution/resolution |
| Finalized | Any different event | Reject with diagnostic; frozen snapshot unchanged |

A finalized tracker never resumes accumulation; later facts are rejected and cannot alter its snapshot. Start-record failure allows bounded collection but never complete coverage. Final append failure leaves the stored start incomplete and emits a warning. Death during collection has no synthetic finalization: read-side aggregation shows the unmatched start as incomplete. Sequence-based tests drive arbitrary event orderings and assert no resolved exposure without proof/resolution, no cost before execution evidence, no changes after finalization, and no complete flag after sticky uncertainty.

Both append operations use the existing owner/context and locked log. Measurement write failure warns visibly and leaves coverage unknown; it never becomes a passing test or a zero-cost success. The final record must not assert completeness if start recording failed. No separate recovery daemon or recomposition is added.

Reuse managed-run timing and attempt identity in `localRunRecord` instead of rescanning the attempt store per car. Carry queue/acquisition/start/finish evidence from `runLocalCheck`; classify startup failure and cancellation before execution as not-started or unknown, not a completed suite. Count actual managed full-suite executions, not reducer `Train.Splits` or effect dispatches. Expose initial runs, baseline runs, split probes and reintegration runs separately. Bisect extra suites are baseline plus split probes; reintegration is a separate extra cost. Sum measured run and queue seconds separately; also measure elapsed split wall time from first failed initial outcome through split settlement/terminal exit. Registered targeted reruns are not whole-suite executions; their time remains included in elapsed recovery wall time and is labelled accordingly. Missing timestamps leave the relevant cost unknown.

Payload limits: at most 16 original cars, 128 execution summaries and 256 exact failure identities, with bounded identity/reason strings and at most 256 KiB encoded per record. Reaching a bound marks overflow/incomplete coverage and records bounded counts, never silently drops evidence while claiming completeness. Store hashes/identities, not log bodies. Existing raw-line sweep preserves unknown members and its 30-day retention applies. This change adds no absolute fleet-log size cap; it bounds new per-train writes. A seven-day weekly window fits retention, but missing/malformed/expired evidence must remain visible. Status reads once and weekly uses its existing bounded event reader.

Add optional `train_filter` to weekly schema v2, containing typed counts, coverage, costs and a human `text` line produced by the same formatter as status. Preserve snapshot path-only stdout and immutable existing artifacts. Missing payload on a pre-change v2 snapshot means unmeasured; do not rewrite old weeks or reinterpret missing fields as healthy. Unknown/new metric versions are displayed as unknown, not rejected as corrupt landing authority. Tests cover the old/new codec and immutable reread contract.

## Actions and visible states

Actions: repair the named map/preparation, run ordinary suite and use supported serial landing, retry the exact head through existing SDLC commands, inspect the summary, or propose a guarded cap/interval change.

| Internal state / failure mode | Visible state | Recovery action | Possible loss |
|---|---|---|---|
| Partial or proven docs-only selection; affected succeeds | Ready on affected | Inspect train status | No new loss |
| Full selection, unknown path or graph/preparation unavailable in boarding mode | Boarding refused with reason and exact commands | Repair named map/preparation; alternative ordinary suite plus serial landing | No suite launched or false proof published |
| Existing passing suite satisfies readiness | Ready on suite; outside filter denominator | Inspect summary | No filter-quality observation |
| Train suite passes, or verified prefix lands | Resolved measured boardings | Inspect summary | No new loss |
| Exact attributable failure after affected pass | Escape, named car/head/tests and split costs | Retry exact head after fixing through existing SDLC flow | Failed integration work, retained evidence |
| Registered flake, untouched/unknown impact, main-red, stale prefix or released untested car | Unattributed/unverified with unresolved coverage | Use existing exact retry/refusal action | No fabricated blame; some filter outcomes unknown |
| Start without finish, failed metric write, overflow, unsupported version or stale proof binding | Incomplete/unmeasured, rate coverage shown | Inspect retained evidence before guarded rollout decision | Partial measurement; no landing authority gained |
| Seven-day data absent or incomplete | Ramp review lacks complete evidence | Keep cap and inspect next complete window | Ramp delayed; no automatic config mutation |

Covered: conservative boarding, valid existing suite evidence, confirmed escape identities, actual split cost, partial data and cautious default onboarding. Not covered: redesigning affected maps, new causal attribution, consumer config adoption, automatic thresholds/cap increases, or recovering vanished process measurements. These remain separate authority or operational decisions.

## Cost model and counted tests

Growth drivers are changed paths/packages during the existing selector, cars C (cap at most 16), failure identities F, execution summaries A, and retained events E. New proof capture reuses the one prepared gate decision per readiness observation; it adds no API read per test/failure. Classification visits each failure once. Measurement accumulation is O(C + F + A), bounded above; aggregation is one O(E) pass plus O(C + F + A) payload work with train-key maps, followed by bounded final formatting. No event rescans per train/car or attempt-store scans per failure.

`TestTrainMetricAggregationWorkIsLinear` supplies N and 2N trains/events and instruments event/payload visits, asserting each input is visited once apart from a constant bounded finalization pass. `TestTrainFailureMeasurementWorkIsLinear` does the same for qualifying failures and verifies registry/impact collection is not repeated for formatting. Exercise 8/16 cars and payload limits separately; elapsed-time benchmarks alone are insufficient.

## Function-based test strategy

Individual examples belong in executable tables/fuzz seeds, not duplicated case inventories here. Each risky production function has an adversarial strategy and mechanical guard:

| Function | Adversarial strategy / mechanical guard |
|---|---|
| `runLocalCheckCmd`, `runAffectedCheck` | Drive malformed mode combinations and uncertain selections through the native CLI/detached fixture; command/status counters prove boarding never executes suite or publishes unsupported success, while ordinary semantics remain intact. |
| `carReadiness`, `ciVerdict.refusal` | Exercise missing-proof state through the existing stateful fake; assert every train-affected recovery command selects boarding mode and already-running checks are retained. |
| `parseLandingConfig`, `Board` | Generate omitted/explicit/invalid policy inputs; assert literal default two, exact explicit authority and unchanged single-car/grant constraints. |
| `FormatTrainRampPolicy` | Phrase guards assert the central seven-day manual-review contract and no automatic cap/enable action; help and status derive from this function. |
| `decideCIGate`, `localRecordsOnAncestors`, `carReadiness` | Vary status provenance/head movement using stateful fake GitHub and native Git; assert the accepted source comes from the same gate decision and no extra source lookup is introduced. |
| `BoardingProof.UnmarshalJSON`, `BoardingProof.MarshalJSON`, `DecodeReadiness`, `Line.Vouch` | Fuzz malformed and future nested objects; round-trip invariants preserve raw unknown members across older proof-aware rewrites, reject duplicate keys, and invalidate stale bindings. |
| `ClassifySplit`, `conductor.assessSplit` | Replay adversarial evidence with the existing train-234 and exact-log fixtures; structured failure output must preserve existing blame decisions and cannot promote uncertainty. |
| `AdvanceMeasurement` | Model-based/fuzz event sequences assert legal lifecycle transitions, idempotence, sticky incompleteness and immutable finalization; inaccessible accumulation fields prevent bypass. |
| `runLocalCheck`, train execution wrapper | Fault-inject at owned queue/process/storage boundaries; assert cost derives from returned execution/timing evidence rather than scheduled effects, with cancellation and uncertainty visible. |
| `conductTrain`, `conductor.recordTrain` | Native bare-remotes, trainFakeGH and deterministic fake suite replay a two-car escape; persist exact accepted/attributed identities and compare costs with the fake's physical invocation ledger. Inject interrupted append boundaries and require unchanged landing authority. |
| `AggregateTrainMeasurements`, payload decode | Fuzz bounded malformed/partial event streams against a simple reference fold; invalid/absent evidence cannot become a healthy zero, and deduplication/window accounting remains deterministic. Count N/2N work. |
| `FormatTrainFilterStats`, `runLandingStatus`, `runWeeklySnapshot`, `readExistingWeeklySnapshot` | Feed shared evidence windows and pre-change immutable artifacts; assert identical fractions/cost semantics, explicit unknown coverage and unchanged existing artifact bytes/stdout contract. |

## M1 — Explicit boarding preparation and cautious onboarding

Files: `cmd/sdlc/localcheck_cmd.go`, `localcheck_affected.go`, `localcheck_affected_test.go`, `localcheck_graphprepare_test.go`, `landingcar.go`, `landingcar_affected_test.go`, `landingcar_test.go`, `cigate.go`, `landingconfig.go`, `landingcmd.go`, `helptext/local-check.md`, `helptext/landing.md`, `internal/landingline/train.go`, `train_test.go`, new `internal/landingline/trainpolicy.go` and test, and `atlas/workflow/sdlc-binary.md` (all code paths under `cmd/sdlc/`). Existing configuration tests live in `landingcar_test.go`.

- [ ] Write failing `TestAffectedBoardingNeverEscalatesToSuite` using the native fixture and the function strategies above.
- [ ] Run `go test ./cmd/sdlc -run 'TestAffected(Boarding|FullSelection)|Test.*Boarding.*Flag' -count=1`; observe the new regression fail before implementation.
- [ ] Implement the explicit mode through affected execution and recommend it in every affected missing-proof recovery message. Retain the exact full-proof alternative and the guarded serial prerequisite; update help from the same policy owner.
- [ ] Write the failing policy tests, implement the default cap and central manual-review rule, and render that rule in landing help/status.
- [ ] Extend `TestLandingTrainLandsEveryCarOnOneSuite` to prove affected-only boarding and the unchanged integration gate, using the native train strategy above.
- [ ] Run `go test ./cmd/sdlc/internal/landingline -count=1` and targeted native CLI/car/train tests; update atlas and issue Log with evidence, commit, then `sdlc milestone-close --issue 561 --milestone M1 --verified '<returned evidence>'`. The binary owns the fresh-context review.

## M2 — Escape and split-cost evidence through status and weekly

Files: `cigate.go`, `landingcar.go`, `localchecks.go`, `localcheck_run.go`, `landingtrain.go`, `landingtrain_attribution.go`, new `landingtrain_metrics.go` and test, `internal/landingline/train.go`, `attribution.go`, `log.go`, new `trainmetrics.go` and test, `landingcmd.go`, `weeklystats.go`, `weeklystats_cmd.go`, their tests, and `atlas/workflow/sdlc-binary.md` (all `cmd/sdlc/` unless noted).

- [ ] Write red provenance/codec tests using the strategies above, then implement optional `BoardingProof` through the existing gate decision/publication path and strict preserving codecs.
- [ ] Write red attribution tests, implement structured qualifying output once in `ClassifySplit`, and propagate it through `assessSplit` without changing reducer authority.
- [ ] Write red lifecycle and aggregation tests, then implement `AdvanceMeasurement`, bounded payloads, linear fold and shared formatter. Run `go test ./cmd/sdlc/internal/landingline -run 'Test.*(TrainMetric|TrainFilter|TrainFailureMeasurement|ClassifySplit)' -count=1`.
- [ ] Extend managed execution records with existing timing/attempt provenance. Feed execution observations and verified resolution facts into the pure tracker; the train owner handles start/final append effects through the existing log.
- [ ] Add the native two-car escape proof and storage/process fault injection using the function strategies above; compare persisted measurements against actual fake-suite executions and retain existing landing verdicts.
- [ ] Wire status's recent seven-day filter summary and the weekly window's additive `train_filter` data/text to the same fold/formatter. Verify compatibility through the named reporting strategies; update atlas with definitions and coverage limits.
- [ ] Run targeted tests: `go test ./cmd/sdlc -run 'Test.*(Affected|CarBoards|CarFullSuite|LandingTrain|SplitAssessment|ReadSplitEvidence|TrainMetric|Weekly|LandingStatus)' -timeout 20m -count=1`; run the linear counted tests at N/2N.
- [ ] Commit and publish through `sdlc pr`; run the required recorded check on the pushed head using the repo's declared affected policy if present, otherwise `sdlc local-check run --name suite`. Isolated real CLI/bare-remote/fake-suite proof is the proposed live proof; ask the hub whether it suffices before staging anything. Close M2 and issue through the binary's review gates with returned evidence, then `sdlc merge --yes` on SHIP and report the PR. No live :0 or consumer mutation without hub authorization.

## Alternatives considered

Broadening Full or trusting a native filter on unknown paths would manufacture coverage; rejected by the hub. Removing all suite fallback would regress ordinary close/serial merge; mode selection isolates the approved exception. Counting `Train.Splits` would count cancelled effects and miss baseline/reintegration distinctions; actual managed execution owns those facts. Measuring only final successful trains would hide crash/unresolved coverage; start/final records and explicit unknowns prevent that claim. Automatically ramping from a numeric threshold would invent operator policy; centralized guidance plus guarded manual configuration preserves authority.

## Revisions

- 2026-10-23: Initial full-flow design. Hub confirmed explicit boarding refusal, valid existing suite proof, omitted-cap default two, preservation of explicit caps and seven-day manual measured review. Full plan still awaiting hub approval; no implementation or estimate yet.
- 2026-10-23: Hub approved the full plan and isolated production proof. Plan-quality round one requested function-based adversarial strategies (PQ-1), a pure measurement lifecycle owner (PQ-2), and nested proof preservation (PQ-3). Added the named strategy table, `AdvanceMeasurement` transition contract and strict preserving proof codec/re-vouch binding policy; compressed procedural case inventories. Also made train recovery recommendations use boarding mode and corrected the location of existing config tests. These refine the approved mechanism; no selector, gate or rollout authority changes.
- 2026-10-23: M2 as built (M2 review BR-8). `BoardingProof`, `AcceptedProof` and `carryProof` live in `internal/landingline/trainmetrics.go`, not `train.go`; `Readiness` only gains the optional `proof` member. `assessSplit` is unchanged: `SplitAttribution.Escapes` (sorted, deduplicated) reaches the conductor through the `ClassifySplit` return value, and the conductor reads it after the reducer commits the culprit. The run kind is computed from the reducer state that emitted `RunSuite` (`suiteRunKind`) and passed to `runSuite`; a reused saved pass is no execution, and a pass from a policy with no suite command resolves no boarding. The fold adds `Open` (a start with neither final nor terminal train event: in flight or a dead conductor) and names escapes (bounded). Function strategies delivered as `FuzzBoardingProof`, `FuzzTrainMeasurementAggregate` (against a quadratic reference fold), `TestAdvanceMeasurementRandomSequences`, and `TestLandingTrainLandsDespiteLostMeasurements` (native train with every measurement append failing).
