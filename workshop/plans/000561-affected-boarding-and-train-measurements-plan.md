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
| `AggregateTrainMeasurements`, `FormatTrainFilterStats` | `cmd/sdlc/internal/landingline/trainmetrics.go` | new |
| `DefaultMaxCars`, `TrainRampReviewWindow`, `FormatTrainRampPolicy` | `cmd/sdlc/internal/landingline/train.go`, new `cmd/sdlc/internal/landingline/trainpolicy.go` | modified / new |
| `WeeklyRepoStats` | `cmd/sdlc/weeklystats.go` | modified |

- `affectedExecutionMode` has ordinary and boarding values. Selection remains owned by `affected.Plan`; the mode decides only whether Full executes suite or refuses. This is the smallest alternative to changing selector authority (ARCH-DRY).
- `BoardingProof` records the accepted check kind (affected or suite), car head, source head for permitted ancestor evidence, status context and observation binding. `ciVerdict` exposes the accepted source from its existing decision, not a second GitHub lookup. A readiness owns at most one suite-requirement proof; a departed measurement copies each original car's proof. Full-suite proof is valid readiness but outside the affected-filter denominator. Unknown proof affects measurement coverage, never retroactively waives or adds a gate.
- Readiness's optional proof must match its head and observation stamp. Older writers can preserve unknown JSON while changing readiness: a mismatched stamp makes the measurement proof unknown. Preserve existing `Extra` members and the readiness envelope version; optional data cannot make a previously invalid readiness valid.
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

## M1 — Explicit boarding preparation and cautious onboarding

Files: `cmd/sdlc/localcheck_cmd.go`, `localcheck_affected.go`, `localcheck_affected_test.go`, `localcheck_graphprepare_test.go`, `landingcar_affected_test.go`, `landingconfig.go`, `landingconfig_test.go`, `landingcmd.go`, `internal/landingline/train.go`, `train_test.go`, new `internal/landingline/trainpolicy.go` and test, and `atlas/workflow/sdlc-binary.md` (all code paths under `cmd/sdlc/`).

- [ ] Write failing `TestAffectedBoardingNeverEscalatesToSuite` using the existing detached execution fixture. Assert exact command counts and no passing publication for Full and unavailable selection; preserve ordinary fallback controls, partial execution and docs proof. Add CLI flag-combination tests and exact recovery text guards.
- [ ] Run `go test ./cmd/sdlc -run 'TestAffected(Boarding|FullSelection)|Test.*Boarding.*Flag' -count=1`; observe the new regression fail before implementation.
- [ ] Implement the explicit mode and plumb it only through affected execution. Ensure any early graph failure or cancellation takes the same conservative path. Refusal tests assert the exact full-suite/retry commands and the guarded configuration prerequisite for serial landing.
- [ ] Write failing default-cap/config/Board tests: omitted cap two, explicit caps unchanged, disabled trains unchanged, existing priority/alone rules unchanged. Implement the central cap and seven-day manual-review rule; render that rule in landing help/status from its single owner.
- [ ] Prove two passing affected cars depart together and execute one initial full suite in `TestLandingTrainLandsEveryCarOnOneSuite`; prove a red full integration suite still blocks/splits. Preserve `TestCarBoardsOnAffectedMergeStillNeedsSuite` and `TestCarFullSuiteFallbackReadiness`.
- [ ] Run `go test ./cmd/sdlc/internal/landingline -count=1` and targeted native CLI/car/train tests; update atlas and issue Log with evidence, commit, then `sdlc milestone-close --issue 561 --milestone M1 --verified '<returned evidence>'`. The binary owns the fresh-context review.

## M2 — Escape and split-cost evidence through status and weekly

Files: `cigate.go`, `landingcar.go`, `localchecks.go`, `localcheck_run.go`, `landingtrain.go`, `landingtrain_attribution.go`, new `landingtrain_metrics.go` and test, `internal/landingline/train.go`, `attribution.go`, `log.go`, new `trainmetrics.go` and test, `landingcmd.go`, `weeklystats.go`, `weeklystats_cmd.go`, their tests, and `atlas/workflow/sdlc-binary.md` (all `cmd/sdlc/` unless noted).

- [ ] Write red pure tests for accepted check provenance and head/stamp binding, then implement optional `BoardingProof` through the existing decision/publication path. Test older readiness rewrites and unknown fields, valid full-suite substitution, and permitted ancestor proof; no extra GH read.
- [ ] Write red attribution tests retaining every qualifying failure while preserving all existing kind/reason decisions. Implement structured output once in `ClassifySplit`; propagate it through `assessSplit` without changing reducer authority.
- [ ] Write red pure measurement/fold tests for the defined denominator, partial prefix resolution, duplicate/conflicting records, missing/old/new payloads, window boundaries and bounded overflow. Implement the types, linear fold and shared formatter. Run `go test ./cmd/sdlc/internal/landingline -run 'Test.*(TrainMetric|TrainFilter|TrainFailureMeasurement|ClassifySplit)' -count=1`.
- [ ] Extend managed execution records with existing timing/attempt provenance. In the train owner instrument the initial suite, baseline, split probes and reintegration; record start/final measurements using the existing log. Verify actual executions, pre-start cancellation and unknown startup/cleanup outcomes against owned managed-run fixtures; never infer execution count from `Splits`.
- [ ] Extend the native train fixture so two affected-pass cars enter, the initial full suite finds an exact unregistered affected failure, the adjacent prefix proves attribution, and persisted events name that escaped car/head/test. Assert one escaped/resolved exposure, unresolved trailing cars where applicable, actual baseline/probe counts, queue/run timing and elapsed split time. Cover registered-flake/untouched train-234 behavior and stale-prefix reintegration as controls: no false escape.
- [ ] Add start-only crash, metric-write failure, invalid proof binding, overflow and duplicate final-event replays. Confirm landing verdicts stay unchanged and summaries expose missing coverage. Reuse existing raw-retention/unknown-field sweep tests.
- [ ] Wire status's recent seven-day filter summary/cap/ramp guidance and the weekly window's additive `train_filter` data/text to the same fold/formatter. Test old immutable v2 rereads, path-only stdout, malformed/missing data, and parity on the same event window. Update atlas with numerator/denominator and known coverage limits.
- [ ] Run targeted tests: `go test ./cmd/sdlc -run 'Test.*(Affected|CarBoards|CarFullSuite|LandingTrain|SplitAssessment|ReadSplitEvidence|TrainMetric|Weekly|LandingStatus)' -timeout 20m -count=1`; run the linear counted tests at N/2N.
- [ ] Commit and publish through `sdlc pr`; run the required recorded check on the pushed head using the repo's declared affected policy if present, otherwise `sdlc local-check run --name suite`. Isolated real CLI/bare-remote/fake-suite proof is the proposed live proof; ask the hub whether it suffices before staging anything. Close M2 and issue through the binary's review gates with returned evidence, then `sdlc merge --yes` on SHIP and report the PR. No live :0 or consumer mutation without hub authorization.

## Alternatives considered

Broadening Full or trusting a native filter on unknown paths would manufacture coverage; rejected by the hub. Removing all suite fallback would regress ordinary close/serial merge; mode selection isolates the approved exception. Counting `Train.Splits` would count cancelled effects and miss baseline/reintegration distinctions; actual managed execution owns those facts. Measuring only final successful trains would hide crash/unresolved coverage; start/final records and explicit unknowns prevent that claim. Automatically ramping from a numeric threshold would invent operator policy; centralized guidance plus guarded manual configuration preserves authority.

## Revisions

- 2026-10-23: Initial full-flow design. Hub confirmed explicit boarding refusal, valid existing suite proof, omitted-cap default two, preservation of explicit caps and seven-day manual measured review. Full plan still awaiting hub approval; no implementation or estimate yet.
