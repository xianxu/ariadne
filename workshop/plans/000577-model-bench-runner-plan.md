---
collaboration-mode: discussion
---

# Model-bench observer implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3 for execution approach. This is full-flow work; obtain hub approval and pass `sdlc change-code` before implementation.

**Goal:** Replay one autonomous landed issue through real SDLC gates, with a fixed bench TL and reviewer, and produce a scored, attributed ledger row without publishing to production.

**Architecture:** Add `sdlc bench`, so the observer can reuse SDLC's internal tracker, suite, usage and measurement packages. Materialize the selected parent tree into fresh Git storage and use a private bare origin for real publication/CAS. Keep hidden acceptance and source pins in observer storage; the TL and worker see only the sanitized issue. `tl-lead bench` adds structured decisions and a would-land handoff to the normal TL rules.

**Tech stack:** Go, Cobra, existing Git/tracker/managed-suite/usage packages, Couch's JSON CLI, versioned JSON records and JSONL ledger. No new service or GitHub emulator.

## Scope and settled decisions

- Source: operator-approved ops `teams/devinfra/projects/model-bench/project.md`, read before claiming; project M1 only. No edits to that TL-owned project.
- One autonomous task, one worker model, one fixed TL model and one explicitly fixed reviewer model per descriptor. No fleet slot takeover, account setup, second-machine setup, report, or corpus expansion.
- M1 guarantees **prepared-repository isolation with trusted agents**. It does not prevent same-user host reads or network access. OS read/network isolation is later work, required before the bench host also holds development checkouts (hub decision 2026-10-23).
- After every run, scan the bound worker transcript/tool calls for outside reads, scorer/source paths, fetch/clone and GitHub access. Record `possibly leaked` with evidence when observed. Missing/truncated/unsupported transcript coverage remains unknown; absence of matches is never a proof of secrecy.
- The original/source remote is absent. The sole remote is a newly created private bare origin containing only sanitized replay objects. This is the approved revision of Done-when, not a production remote with publication disabled.
- Claim, start-plan, change-code, close and their real CAS remain unchanged. Only local-check **status publication** gets a validated, visibly bench-only local receipt sink. No `--no-*` gate bypass, `SkipRecord`, fake green status, or production flake erasure.
- Attach to explicitly provisioned disposable bench slots; verify their roots, agent/model/effort and session generations before dispatch. The observer never silently changes models or prepares existing fleet slots.
- Autonomous M1 has no operator looks: record an empty wait list and zero operator-wait time, labelled autonomous. A future operator-wait protocol is outside this milestone; an unexpected request stops the run with its reason rather than inventing a duration.

## Core concepts

All names below are proposed delivered symbols, not existing helpers. Existing owners are named separately under reuse.

| Name | Kind | Lives in | Status |
|---|---|---|---|
| `Task`, `PublicTask`, `ValidateTask` | PURE | `cmd/sdlc/internal/bench/task.go` | new |
| `Run`, `Event`, `Effect`, `Step` | PURE | `cmd/sdlc/internal/bench/run.go` | new |
| `WouldLandFacts`, `ValidateWouldLand` | PURE | `cmd/sdlc/internal/bench/admission.go` | new |
| `LeakFinding`, `ScanLeakEvents` | PURE | `cmd/sdlc/internal/bench/leaks.go` | new |
| `Measurements`, `Acceptance`, `LedgerRow`, `BuildRow` | PURE | `cmd/sdlc/internal/bench/result.go` | new |
| `prepareBenchRun` | INTEGRATION | `cmd/sdlc/bench_prepare.go` | new |
| `loadBenchContext`, `guardBenchCommand` | INTEGRATION | `cmd/sdlc/benchcontext.go` | new |
| `runBenchObserver`, `benchCouch` | INTEGRATION | `cmd/sdlc/bench_run.go`, `bench_couch.go` | new |
| `collectBenchMeasurements` | INTEGRATION | `cmd/sdlc/bench_measure.go` | new |
| `scoreBenchRun` | INTEGRATION | `cmd/sdlc/bench_score.go` | new |
| `benchStore`, `recordBenchRow` | INTEGRATION | `cmd/sdlc/bench_store.go` | new |
| `localRunDeps`, `postLocalCheckRecord` | INTEGRATION | `cmd/sdlc/localcheck_run.go`, `localcheck_republish.go` | modified |
| `newBenchCmd`, `runBenchDecision`, `runBenchFinish` | INTEGRATION | `cmd/tl-lead/bench.go` | new |

One Task owns immutable source pins and test material; its public projection cannot contain landing hashes, source paths, hidden patch bytes or scorer paths. One Run owns one Task, two exact runtime identities, its phase and accepted operation IDs. A Row is derived from one terminal Run plus evidence; it is never authored by the TL. This is the first observer pipeline, not a generalized workflow framework. Future models widen descriptor fields or transcript adapters, not the run protocol.

Thin IO shells feed typed facts to pure `Step`/admission/result functions. Fakes model a persisted Couch actor/receipt store and a real temporary run folder, including uncertain deliveries and actor replacement. Git conformance uses real temporary repositories and the existing Git runner seams; no shared objects from a production repository enter a prepared run.

## Reuse and boundaries

- Reuse `gitx.TrunkView.Files`/`ReadBlob` for pinned object reads with existing limits; write only validated relative entries, preserving ordinary modes. Refuse unsupported gitlinks and escaping symlinks. Disable lazy fetching and inherited object/config/hook alternates during observation/materialization.
- Do not call `reviewbox.CreateContext` for preparation: it clones with alternates and restores current main authority. Reuse its lifetime-lock pattern and scratch placement; use normal reviewer isolation only after the repository is sanitized.
- Reuse local bare support in `gitx.ResolvePublicationTarget`, `openTrackerAt`, tracker initialization/migration, `tracker.Drive`, `publishTrackerClose` and `boundaryPush`. Existing baseline issue storage determines normal initialization versus migration; retain the selected task ID through supported migration rather than hand-writing completion state.
- Reuse `runLocalCheck` and `postLocalCheckRecord`: preserve preflight, pushed-tip checks, trusted declaration, suite slots, attempt writer, full-suite proof, cancellation and `guardSuitePassRecord`. Add an optional receipt recorder **after** the same guard; default remains GitHub. Bench run/republish construct the recorder from validated context; unsupported bench resume/reclassification paths refuse explicitly.
- Reuse `transcripts.DiscoverEvidence`/`DiscoverUsageEvidence`, `usage.Decode`/normalization and existing ownership-based actual measurement. Bind TL and worker separately; ownership of the task alone is not proof of TL token attribution. Reviewer usage retains its existing role.
- Read gate rounds/findings from `gatestate` and completion from `observe`, not prose matching. Pin reviewer with explicit CLI/model selection; default cross-family selection would vary the benchmark. Preserve any same-family/degraded-review label.
- Couch remains the runtime owner. Consume `--actors`, `--send-to`, `--message-status`, `--operation-status`, `--peek --json --transcripts` and supported interruption receipts. Never infer successful delivery, task completion or descendant shutdown from pane text or elapsed time.

## CLI and on-disk contract

Proposed observer surface:

```text
sdlc bench prepare --task TASK.json --run-dir ABS
sdlc bench run --run-dir ABS --hub SLOT --worker SLOT --budget DURATION
sdlc bench status --run-dir ABS --json
sdlc bench score --run-dir ABS
sdlc bench record --run-dir ABS --ledger ABS
sdlc bench cancel --run-dir ABS
sdlc bench clean --run-dir ABS
tl-lead bench --help
tl-lead bench decision --run ABS --kind approve|replan|ruling|stop --reason TEXT
tl-lead bench finish --run ABS --outcome would-land|stopped --head SHA --reason TEXT
```

`Task` pins corpus version/task/type, canonical source repository, parent and landing objects, original issue blob, test-only acceptance patch/paths and literal argv, original measured hours, worker/TL/reviewer identities and repeat index. Observer validates all pins and test-only patch scope before creating a run. The descriptor is trusted operator input after typed/path/bounds validation; neither worker nor TL can nominate their own acceptance tests or overwrite it.

`RUN/observer` holds private task/patch/raw evidence; `RUN/public/work`, `RUN/public/hub` and `RUN/public/origin.git` contain only replay context. The public context records run ID, canonical work root and exact private-origin identity; no source/landing/scorer details. The clean parent tree gets a new root commit, original Spec and Done-when, empty Plan/Log, open task metadata and fresh workflow state. No original refs, reflogs, hooks, worktree links, alternates, replace refs, source URLs or later objects survive. Preserve task source SHA only in observer evidence. Do not retrofit the historic task's code with the landed fix.

The bench context is explicit and root-bound; production with no context follows existing code. Before any replay effect, check the current canonical root/upstream/private-origin identity against the public context. Retargeting to production, substituted paths and symlinks refuse. Reject `pr`, `merge`, `push`, stage and unrelated publication in a bench context. Private branch publication needed by real gates remains allowed through the validated origin.

Capture gate invocations/outcomes at the SDLC command execution boundary only when context validates, so every refusal (including pre-review refusal) has one structured event. Propagate failures unchanged. Audit failures prevent further effects and appear as measurement uncertainty; never claim a complete count from lossy transcript tails. TL decisions use one shared record protocol, with actor/run/generation checks, and never grant suite/close authority themselves.

## Ordering, ownership and visible failure states

Actions: prepare; run once; inspect status; cancel; score; record; clean. Each run has one lifetime lease, held by the observer across dispatch/watch; another runner or cleanup cannot acquire it. A separate short store transaction lock serializes atomic state writes and append-only proposals from TL commands. Those proposals are not authoritative state: the sole observer validates their actor/run/generation and feeds them through `Step`. This is a local file protocol, not a new service. Only the pure model authorizes external effects.

| Internal state / event | Visible state | Recovery action | Possible loss |
|---|---|---|---|
| invalid descriptor or preparation failure before publication | not prepared | correct input, prepare a fresh run | partial owned scratch is removed |
| complete sanitized seed | prepared | run once | none |
| dispatch recorded, acknowledgment missing | dispatch unconfirmed | inspect the exact receipt; no redispatch | no new attempt until identity is known |
| matching dispatch acknowledged | running | status or cancel | budget consumed; transitions retained |
| TL reports would-land, proofs absent/stale | running / evidence required | repair within the same budget | no false successful terminal state |
| matching head, full suite and SHIP completion verified | would land | score immutable candidate | no real landing occurs |
| stop, budget expiry or unexpected operator look | stopped | inspect and record partial evidence | unfinished work retained, no score invented |
| actor replaced, supervisor lost or shutdown unconfirmed | needs reconciliation | status, then supported runtime cancellation/confirmation | no automatic scoring/cleanup or second dispatch |
| scoring patch conflicts or test tool fails to start | acceptance unavailable | inspect/retry score on the same frozen candidate | failure is not a test fail rate |
| tests complete with failures | scored | record result | benchmark can record a poor candidate |
| ledger append acknowledged or exact duplicate row found | recorded | clean disposable trees when safe | bounded raw evidence retained |
| partial/conflicting ledger tail or duplicate identity | recording refused | inspect/repair named ledger bytes | never append a duplicate or silently truncate |

`Step` covers begin, confirmed/unconfirmed dispatch, observation, TL proposal, proof result, cancellation, timeout, late outcome, score result and record result. Late events require exact run/session/operation identity and cannot resurrect a terminal run. No automatic retry of runtime launch. On an observer crash, a second invocation reports the retained phase and receipt; it may observe/reconcile, not launch another worker. Supported interruption must be confirmed before destructive cleanup. If Couch cannot prove shutdown, retain the trees and report that limit; never kill a PID found by global `ps`.

## Measurement and scoring

The observer pins transcript file identity, bounded byte prefix, digest, runtime session, cwd and clock interval per role. Decode each source once for usage, activity and leak evidence; preserve decoder coverage and unbound/unsupported values as unknown. Record gross wall time independently of measured focused hours. Compare to the immutable original's hours without rewriting historical actuals. Review rounds/findings and TL decisions retain source record IDs, not merely totals.

Leak detection runs after all terminal outcomes, including timeout/stopped. Normalize supported tool call events, canonicalize known path operands relative to recorded cwd and detect access to protected source/scorer roots; conservatively flag fetch/clone/GitHub operations and suspicious opaque commands. Allow necessary private-origin workflow fetches only when their target is proven local; label that exclusion explicitly. Structured and shell tools, indirect/wrapped commands, unsupported event formats and incomplete sources get adversarial tests. This detector is diagnostic: `no observed leak` is not `proved clean`.

Freeze the candidate commit once admitted. Build a separate scoring copy from its tree, apply the observer-pinned test-only patch without invoking patch-controlled hooks, then run literal acceptance argv through the existing managed batch/resource machinery. Worker and TL never receive hidden output. Validate hidden tests on the sanitized parent (expected red) and landed tree (expected green) when selecting the corpus fixture; invalid or vacuous acceptance cannot produce a score. Where native structured case events exist, report passed/failed/total. Otherwise record command-unit success with case pass rate explicitly unavailable, never fabricate per-test precision from an exit code.

Ledger key: corpus version, repo/task/type, repeat, worker model/effort, TL model/effort, reviewer model and run ID. Row includes terminal outcome, source/candidate pins, hidden-test proof and score, active/gross/operator-wait durations, all roles' token/credit coverage, review rounds/findings, gate refusals, structured TL decisions, leak status/evidence and raw record digest. No credentials or whole private transcript bodies enter the shared ledger. Append once under a ledger lock; identical existing run ID/digest is a no-op, differing content refuses.

## Operating envelope and lifecycle

- Batch workload: one run active per bench workspace, one worker plus one TL, normal suites under the existing machine queue. Required explicit budget, at most 24h; cancellation uses a separate bounded cleanup deadline. M1 target is one small autonomous issue, not corpus throughput. Capacity and retention values below are initial conservative implementation limits, to be measured by fixtures and the M1 replay.
- Parent export inherits the bounded Git snapshot contract (10,001 entries, 1 MiB per blob); enforce a 64 MiB total tree budget before writing. Unsupported repositories refuse and name the exceeded bound. No fetch/promisor fallback.
- Task/raw record: 1 MiB each; transition/decision audit at most 2,000 records/run; individual tool/transcript lines at most 1 MiB; transcript/log capture at most 256 MiB/run. Caps are enforced by writers, with visible incomplete coverage/stopped outcomes rather than fabricated totals.
- Poll every five seconds until budget expiry; observe small current facts, not repeated transcript scans or full history. Persist transitions, not every unchanged poll. Collection/scoring is one pass over N captured events/bytes and test entries; count visits at N and 2N. Parent materialization is one tree traversal and one read per selected blob; count reads at N/2N. Ledger append scans existing run IDs once under lock, bounded to 10,000 rows/160 MiB; refuse before a new run when full.
- The run lifetime lock owns disposable work/hub/origin/scoring trees. Normal confirmed terminalization plus recording permits cleanup; dead-owner sweep reclaims only stopped/confirmed runs. Unknown runtime ownership retains the folder and names reconciliation. Never remove a live or possibly live run.
- Retain raw records/logs for 30 days and at most 50 terminal runs; sweep before admission, never active/uncertain records. Compact ledger rows remain (bounded as above) for calibration. An explicit clean command uses the same owner/identity checks and leaves the row plus summary digest. No hidden indefinite transcript copies.
- New JSON formats use versioned strict required/known fields and duplicate-key rejection; nested unknown fields survive every update. Unknown unsafe schema versions refuse. Ledger append never rewrites older rows. Older-reader/newer-record round-trip fixtures cover nested extensions (ARCH-EVOLVE).

## Test strategy by risky production function

| Function | Adversarial strategy | Independent guard |
|---|---|---|
| `ValidateTask`, `prepareBenchRun` | hostile Git config/promisor objects, path escape, substituted input or later object | real Git object lookup for landing SHA fails; all public refs/objects/URLs inventoried; source snapshots unchanged |
| `loadBenchContext`, `guardBenchCommand` | root/origin/generation drift and context spoofing | production publication sentinel never called; local backend object set stays sanitized |
| `postLocalCheckRecord` | local receipt attempt bypassing existing red/flake/unknown guards | same native same-head failure guard refuses; no production GH calls; default path unchanged |
| `Step`, `runBenchObserver` | lost acknowledgment, timeout racing terminal proof, duplicate/late events, actor replacement | controllable event sequences with explicit invariants; one dispatched run, no post-terminal launch |
| `collectBenchMeasurements` | missing/truncated source, foreign session, duplicate usage or role mismatch | pinned independent counters/digests and real Claude/Codex captures; unknown never becomes zero |
| `ScanLeakEvents` | outside reads, wrappers, fetch/clone/GitHub, malformed tools and private-origin exception | recorded cross-harness fixtures and counted N/2N scans; match evidence exposed in result |
| `ValidateWouldLand`, `scoreBenchRun` | stale HEAD, absent suite/review, hidden patch collision/tool startup failure | exact-head receipt and completion binding, independent parent-red/landing-green/candidate oracle |
| `recordBenchRow`, `benchStore` | concurrent writers, partial append, lost acknowledgment, nested future fields | one row per run ID; immutable original bytes; strict parse/round-trip guards |
| `runBenchDecision`, `runBenchFinish` | wrong role/session/run and false successful proposal | normal gate evidence decides outcome; no TL write can mint a pass |

## Implementation sequence

- [ ] Add pure task/public projection, run phase/effect model, admission/result shapes and strict codecs in `cmd/sdlc/internal/bench`; colocated no-IO tests red first, sequence tests for uncertainty and event ordering, N/2N operation counts.
- [ ] Add `bench_prepare.go`, `bench_store.go`, CLI registration in `cmd/sdlc/bench.go`/`main.go`, and real-Git preparation fixtures. Prove source untouched and landed-object/remote/archive absence before wiring runtime.
- [ ] Add validated bench context and the optional local receipt sink in local-check run/republish construction. Instrument bench-only command outcomes at the existing command execution boundary; reject unsupported publication/resume paths explicitly. Exercise real claim -> start-plan -> change-code -> full suite -> close against private origin; verify ordinary GitHub path unchanged. No synthetic pass.
- [ ] Add stateful Couch adapter/fake and observer loop with preprovisioned identity/model checks, acknowledged dispatch, budget, current-proof admission and cancellation/reconciliation. Use argv, bounded replies and context propagation at every subprocess seam.
- [ ] Add `tl-lead bench` embedded help, decision/finish commands and shared event protocol. Render normal stewardship rules plus bench overrides; test Claude/Codex (and Qoder fixtures if supported) without text-only assertions replacing protocol checks.
- [ ] Add pinned usage/activity collection, structured refusal/decision aggregation and leak diagnostics with existing transcript decoders. Add recorded Claude/Codex tool/event captures and unknown-coverage controls; aggregate once.
- [ ] Add hidden acceptance on separate frozen candidates, native case/command score semantics, parent-red/landing-green validation and idempotent ledger append. Run deterministic end-to-end through a stateful disposable runtime fixture using real Git and real tests; produce the raw record and ledger row in the test output/artifact.
- [ ] Update README, `cmd/sdlc/helptext/bench.md`, `cmd/tl-lead/bench.md`, `atlas/workflow/model-bench.md` and its atlas index link; add recovery catalog entries/proofs for every mutating bench verb. Record deployment limits, isolation claim and repeatable conformance commands.
- [ ] Run focused tests (`go test ./cmd/sdlc/internal/bench ./cmd/tl-lead -run Bench -count=1`, `go test ./cmd/sdlc -run 'TestBench|TestLocalCheck.*Record|TestRecoveryContractsAreProven' -count=1 -timeout 10m`) plus lint. Native service conformance uses Pair's existing `couch-conformance-bridge`, built with `make -C /Users/xianxu/workspace/pair couch-conformance-bridge`, through pipes and its complete replacement environment. Its ready line is bounded to 64 KiB; EOF then successful Wait certifies cleanup. The existing bridge provisions a steward but does not start a worker, so it proves service/receipt/lifetime compatibility, not a complete live model replay. Use the persisted runtime fake for automated end-to-end tests, clearly marking their rows synthetic and keeping them out of the real ops ledger. A true paid model replay requires the separately provisioned disposable bench hub/worker coordinated with the TL; do not extend Pair or invent another Couch launcher in this issue.
- [ ] Commit/push final source; recorded suite through `sdlc local-check run --name suite`; close with full review and measured actual. Coordinate any paid real replay/live rollout with the hub, retain one autonomous run's raw/ledger evidence, then normal PR/merge. No replay code ever merges into a product's real main.

## Architecture review

ARCH-DRY: existing tracker CAS, managed suite, measurement and reviewer owners are reused; no fake GitHub or alternate SDLC implementation. ARCH-PURE: model alone authorizes transitions/effects and builds rows. ARCH-PURPOSE: complete prepare/run/measure/score/ledger path and TL mode, not just checkout scaffolding. ARCH-MOCK: persisted Couch fake plus native conformance; real temporary Git and portable stores. ARCH-CONSTRAINTS: one host run, explicit budget, streaming scans and writer caps. ARCH-SECURE: prepared-repository guarantee, hidden observer inputs and visible diagnostic leak status. ARCH-ORDER: receipt identity, exact-head admission and no blind dispatch retry. ARCH-FUNERAL: lifetime lock, bounded evidence retention, no destructive uncertain cleanup. ARCH-EVOLVE: additive nested-field preservation and strict authority-bearing fields.

## Revisions

- 2026-10-23: hub approved prepared-repository isolation with trusted agents plus transcript/tool leak detection, deferring OS read/network isolation; the latter is required before running alongside development checkouts on the bench host.
- 2026-10-23: hub approved refining remote absence: remove original/source remote; allow only a new private bare origin with sanitized replay objects, preserving real tracker CAS and gates. Only local-check status publication gets a local bench sink.
