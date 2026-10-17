---
collaboration-mode: discussion
---

# Queue wait attribution implementation plan

> **For agentic workers:** Consult AGENTS.md §3 for execution. Use test-first steps, commit each verified unit, and let `sdlc change-code` and `sdlc close` own the plan and code reviews.

**Goal:** Stop recorded machine-suite queue waits from becoming focused development hours while preserving work actually performed during those waits.

**Architecture:** Resolve validated suite attempts against the issue's pinned ownership history, retain the selected queue evidence with the measurement inputs, and pass source-scoped queue intervals to the native active-time engine. The pure engine subtracts their overlap with attributed activity, except for independently timestamped work intervals. The existing phase split runs on the corrected segments, so build, converge and tail all use one measurement.

**Tech stack:** Go; `cmd/sdlc` measurement and suite-attempt readers; `internal/activetime`; native JSONL transcript fixtures; `gatestate.MeasurementInputs` retained by close/correction.

## Core concepts

| Pure entity | Lives in | Status |
|---|---|---|
| `QueueInterval` | `cmd/sdlc/internal/activetime/queue.go` | new |
| `QueueAttribution` | `cmd/sdlc/internal/activetime/queue.go` | new |
| `MeasurementQueueWait` | `cmd/sdlc/internal/gatestate/measurement.go` | new |

`QueueInterval` is a half-open measured wait with attempt ID, owning slot, source binding and exact recorded seconds. `subtractQueueIdle(runs, waits, workSpans)` removes the union of queued idle from one source's activity runs; it preserves only paired, non-wait work intervals and never subtracts the same wait twice. Its property tests use overlapping, duplicated, nested and clipped intervals, asserting nonnegative minutes and conservation: removed + retained = original queue-overlapped active. Multiple attempts belong to one measurement, and one attempt belongs to at most one pinned engagement. `MeasurementQueueWait` records the selected facts and record digest so replay cannot silently adopt a different set after the 30-day attempt retention expires (ARCH-DRY, ARCH-PURE, ARCH-ORDER).

| Integration point | Lives in | Status | Wraps |
|---|---|---|---|
| Queue selection | `cmd/sdlc/actualinputs.go`, `cmd/sdlc/actualqueue.go` | new | existing bounded `readSuiteAttempts`, common Git dir, pinned tracker engagements |
| Transcript work evidence | `cmd/sdlc/internal/activetime/event.go` | modified | existing native Claude/Codex/Qoder JSONL prefix loader |
| Actual computation | `cmd/sdlc/actualmeasurement.go` | modified | `activetime.Compute`, then existing `splitActual` |
| Queue producer | `cmd/sdlc/localcheck_run.go`, `teststats.go`, `teststats_store.go` | modified | machine lock acquisition and attempt JSON |

Selection reads only validated published attempt records, with a bounded reader. It admits `check=suite` on this repository and an owning slot during a matching engagement; it clips to the measurement window and rejects ambiguous cross-issue or duplicate evidence rather than guessing. The suite record supplies the duration. For old records, `started_at - queue_seconds` through `started_at` is the reproducible interval convention, but `started_at` follows post-lock setup: this shifts the window late, leaves that setup-sized interval at the head of the actual wait counted, and excludes the same amount of setup near the run. Test that bias explicitly; do not claim exact window endpoints for legacy data. New records retain the actual queue start and acquisition timestamps. Missing queue data stays unknown and emits a bounded warning, never a fabricated zero. The existing reader is the sole validation boundary (ARCH-SECURE, ARCH-DRY).

`workSpansFromTranscript` pairs structural tool calls with their final results by call ID; for yielded execution it follows the returned session ID through polls until a final result. An unfinished, ambiguous or unrecognized sequence is not work evidence. The waiting `sdlc local-check run` or `sdlc merge` invocation is classified as **wait**, however long its enclosing call/result span is; `write_stdin`/status polls are also wait. A small explicit allowlist of independently timestamped, mutating work (starting with `sdlc stage --build/--clear` and edits) can restore time inside a wait; unknown commands cannot. `bindQueueSource` requires a unique admitted owning-session wait call covering the queue interval and the attempt's slot/engagement. The engine derives ordinary source-scoped runs, then subtracts each queue interval from that source except the union of paired allowlisted work and task spans. Other admitted sources remain unchanged. No unique binding means a provenance refusal for correction, not a guessed subtraction. Parser sequence tests include wrappers, partial/yielded calls, repeated polls, missing final result, duplicate IDs and an enclosing suite call that must never restore its own wait (ARCH-PURPOSE, ARCH-ORDER).

`SuiteAttempt` keeps schema version 1 for compatibility and adds optional `queued_at` and `acquired_at` UTC timestamps as an all-or-none pair. The recorder takes `queued_at` at the existing `queueStart` and `acquired_at` immediately after `acquireSuiteSlot` returns, deriving `queue_seconds` from the same two instants. `decodeSuiteAttempt` admits old records with neither endpoint and rejects only-one-endpoint, reversed/future endpoints, `acquired_at > started_at`, or a timestamp difference inconsistent with `queue_seconds` beyond serialization precision. The reader prefers the exact pair; the documented late-shift convention applies only to old records. A stateful recorder test checks the queued/acquired/run-start ordering and a decoder test checks malformed endpoints (ARCH-SECURE).

Ordinary `sdlc actual` on a completed issue continues to replay the retained measurement generation unchanged, including old snapshots without queue evidence. Add `actualCorrectionCandidate` as an explicit measurement purpose for `sdlc issue correct-actual`: it is bound to that completion's code head, tracker head, source prefixes/windows and phase cuts, but selects currently available validated queue attempts and pins them into a new queue-aware snapshot. It refuses if those records or source bindings are incomplete; it cannot silently reinterpret an old adopted value. After a successful correction, ordinary historical reads replay the corrected snapshot. #416 therefore gets a fresh generation-bound correction measurement, not a changed historical display; if its attempts or transcript evidence cannot support it, the TL/operator excludes the sample through the existing correction verb (ARCH-PURPOSE, ARCH-ORDER).

No new long-lived files: the recorder already creates attempt JSON and prunes it at 30 days. Only a compact selected-evidence array grows in the existing bounded measurement snapshot, which remains attached to its completed issue/correction and is archived by the current flow (ARCH-FUNERAL, ARCH-CONSTRAINTS).

## Tasks

### 1. Pin queue evidence for fresh measurement and replay

**Files:** `cmd/sdlc/actualqueue.go` (new), `cmd/sdlc/actualinputs.go`, `cmd/sdlc/internal/gatestate/measurement.go`, `cmd/sdlc/localcheck_run.go`, `teststats.go`, `teststats_store.go`, corresponding `_test.go` files.

- [ ] Write failing `TestSelectOwnedQueueAttempts` with a foreign slot, interval crossing release/reclaim, duplicate record, and an incomplete record; the mechanical guard is an exact slot plus pinned engagement intersection, never the current checkout's owner. Write `TestSuiteAttemptQueueEndpoints` with only-one/reversed/inconsistent endpoints and a recorder fixture that delays post-acquisition setup. A legacy fixture must expose its late-window bias; a new-format fixture must use the actual acquisition endpoints.
- [ ] Run `go test ./cmd/sdlc -run 'TestActualQueue|TestOwnedActual' -count=1` and confirm the new cases fail for the intended reason.
- [ ] Implement `selectOwnedQueueAttempts` and `MeasurementQueueWait.Validate` using the existing suite-attempt decoder and pinned engagement owners; retain attempt identity, digest, interval and exact seconds. Add the endpoint writer and compatible strict decoder. Old snapshots without queue evidence keep their recorded historical semantics; new snapshots declare the queue-aware method.
- [ ] Make replay use only the saved queue evidence, not ambient attempt files; run the targeted tests green and commit.

### 2. Subtract idle queue time in the native engine

**Files:** `cmd/sdlc/internal/activetime/event.go`, `compute.go`, `queue.go` (new), native fixture/test files.

- [ ] Write `TestQueueNativeTranscript` with one long suite wait, 45–60 second `write_stdin` polls/status turns, a timestamped stage command during the wait, and an independently active source. Assert the queue-only stretch contributes zero while both real-work intervals survive; assert no queue record preserves old behavior. Write `TestWorkSpansFromTranscriptSequence` with a yielded call, polls and final result plus ambiguous/incomplete wrappers; write `TestSubtractQueueIdleIntervals` with overlapping waits and work to prove the conservation invariant.
- [ ] Run `go test ./cmd/sdlc/internal/activetime -run 'TestQueue' -count=1` red.
- [ ] Implement `workSpansFromTranscript`, `bindQueueSource` and `subtractQueueIdle` for supported native transcript forms. The classifier denies a suite/merge invocation and all poll-only operations before considering any allowlisted work, including through wrappers and yielded executions. Run targeted tests green; commit.

### 3. Wire adoption, correction and proof

**Files:** `cmd/sdlc/actualmeasurement.go`, `cmd/sdlc/actualcorrection_test.go`, `cmd/sdlc/actualwiring_test.go`, `cmd/sdlc/helptext/active-time.md`, `atlas/workflow/` map.

- [ ] Add failing `TestActualQueueCorrectionReplay`: a completed issue's ordinary historical read stays unchanged, `correct-actual` uses `actualCorrectionCandidate` against the existing generation, and the new retained snapshot replays the same hours after attempt files disappear. Add `TestActualQueuePhaseSplit` for build/converge/tail and nonnegative hours; a fresh omitted-`--actual` close must use the same queue-aware engine.
- [ ] Wire queue evidence through `activetime.Options` and `MeasurementInputs`; show raw queue seconds, excluded idle seconds and retained work seconds in bounded diagnostics without changing the raw test-stats record. Update help and atlas; run targeted tests green.
- [ ] Run `git diff --check`, targeted `go test`, `make -k test`, then `sdlc local-check run --name suite` on the pushed code head; close through `sdlc close --issue 420 --verified '<evidence>'` and honor its fresh-context review.
- [ ] Preview #416's queue-aware correction with the generation-bound `sdlc issue correct-actual` dry-run, report the measured value or a specific evidence gap to the TL/operator, and use the correction or exclusion flow only with their decision. Do not type a replacement number from the worked example; ordinary `sdlc actual --issue 416` remains its historical replay until correction succeeds.

## Revisions

- 2026-10-17 — Initial full-flow design following the TL's interval/exact-evidence policy and the existing retained-replay contract.
- 2026-10-17 — TL approved the plan and corrected the legacy-time assumption: `started_at` is after post-lock setup, so using it as the queue endpoint shifts the old interval late. Keep the explicit legacy convention and test its bias; store both real queue endpoints in new records. No other design change.
- 2026-10-17 — Plan-quality PQ-1–4 required exact work-span exclusion (including yielded suite calls), a distinct correction-candidate purpose, the recorder/schema contract, and named adversarial test boundaries. Added those before implementation; the same issue scope and TL queue policy remain.
