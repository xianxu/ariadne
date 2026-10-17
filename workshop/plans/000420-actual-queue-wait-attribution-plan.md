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

`QueueInterval` is a half-open measured wait with attempt ID, owning slot, source binding and exact recorded seconds. `QueueAttribution` removes the union of queued idle from one source's activity runs; it preserves work intervals with explicit transcript timestamps and never subtracts the same wait twice. Multiple attempts belong to one measurement, and one attempt belongs to at most one pinned engagement. `MeasurementQueueWait` records the selected facts and record digest so replay cannot silently adopt a different set after the 30-day attempt retention expires (ARCH-DRY, ARCH-PURE, ARCH-ORDER).

| Integration point | Lives in | Status | Wraps |
|---|---|---|---|
| Queue selection | `cmd/sdlc/actualinputs.go`, `cmd/sdlc/actualqueue.go` | new | existing bounded `readSuiteAttempts`, common Git dir, pinned tracker engagements |
| Transcript work evidence | `cmd/sdlc/internal/activetime/event.go` | modified | existing native Claude/Codex/Qoder JSONL prefix loader |
| Actual computation | `cmd/sdlc/actualmeasurement.go` | modified | `activetime.Compute`, then existing `splitActual` |

Selection reads only validated published attempt records, with a bounded reader. It admits `check=suite` on this repository and an owning slot during a matching engagement; it clips to the measurement window and rejects ambiguous cross-issue or duplicate evidence rather than guessing. The suite record supplies the duration. For old records, `started_at - queue_seconds` through `started_at` is the reproducible interval convention, but `started_at` follows post-lock setup: this shifts the window late, leaves that setup-sized interval at the head of the actual wait counted, and excludes the same amount of setup near the run. Test that bias explicitly; do not claim exact window endpoints for legacy data. New records retain the actual queue start and acquisition timestamps. Missing queue data stays unknown and emits a bounded warning, never a fabricated zero. The existing reader is the sole validation boundary (ARCH-SECURE, ARCH-DRY).

The transcript parser identifies structural tool call/result spans. Poll/status-only calls are not work; stage/build/edit/test commands and subagent execution are work only for their timestamped interval. The engine first derives its ordinary source-scoped runs, then subtracts each queue interval from the waiting source except the union of such work spans. Other admitted sources remain unchanged. A queue interval without a unique waiting source or trustworthy record cannot silently change hours: report a provenance error or leave the sample excluded. Do not infer work from prose content or add arbitrary minutes around a tool call (ARCH-PURPOSE, ARCH-PURE).

No new long-lived files: the recorder already creates attempt JSON and prunes it at 30 days. Only a compact selected-evidence array grows in the existing bounded measurement snapshot, which remains attached to its completed issue/correction and is archived by the current flow (ARCH-FUNERAL, ARCH-CONSTRAINTS).

## Tasks

### 1. Pin queue evidence for fresh measurement and replay

**Files:** `cmd/sdlc/actualqueue.go` (new), `cmd/sdlc/actualinputs.go`, `cmd/sdlc/internal/gatestate/measurement.go`, corresponding `_test.go` files.

- [ ] Write failing tests for an owned ariadne:5 wait, another slot/repository, a wait crossing claim/release, duplicate attempts, malformed/missing durations, and a retained snapshot replay after deleting the source record. A legacy fixture must expose the late-window bias when setup separates acquisition from `started_at`; a new-format fixture must use the actual acquisition endpoints.
- [ ] Run `go test ./cmd/sdlc -run 'TestActualQueue|TestOwnedActual' -count=1` and confirm the new cases fail for the intended reason.
- [ ] Implement selection and snapshot validation using the existing suite-attempt decoder and pinned engagement owners; retain the attempt identity, digest, interval and exact seconds. Old snapshots without queue evidence keep their recorded historical semantics; new snapshots declare the queue-aware method.
- [ ] Make replay use only the saved queue evidence, not ambient attempt files; run the targeted tests green and commit.

### 2. Subtract idle queue time in the native engine

**Files:** `cmd/sdlc/internal/activetime/event.go`, `compute.go`, `queue.go` (new), native fixture/test files.

- [ ] Write a native JSONL regression with one long suite wait, 45–60 second `write_stdin` polls/status turns, a timestamped stage command during the wait, and an independently active source. Assert the queue-only stretch contributes zero while both real-work intervals survive; assert no queue record preserves old behavior.
- [ ] Run `go test ./cmd/sdlc/internal/activetime -run 'TestQueue' -count=1` red.
- [ ] Add structural work-span extraction for the supported transcript tool-call forms. Subtract the union of each source's measured queue intervals from its attributed runs, restore only the intersection with timestamped work, and keep separate sources independent. Run targeted tests green; commit.

### 3. Wire adoption, correction and proof

**Files:** `cmd/sdlc/actualmeasurement.go`, `cmd/sdlc/actualcorrection_test.go`, `cmd/sdlc/actualwiring_test.go`, `cmd/sdlc/helptext/active-time.md`, `atlas/workflow/` map.

- [ ] Add failing command-level tests: `sdlc actual`, omitted-`--actual` close and `correct-actual` use the same queue-aware hours and replay the same retained evidence after the attempt files are gone. Check build/converge/tail cuts and no negative actual.
- [ ] Wire queue evidence through `activetime.Options` and `MeasurementInputs`; show raw queue seconds, excluded idle seconds and retained work seconds in bounded diagnostics without changing the raw test-stats record. Update help and atlas; run targeted tests green.
- [ ] Run `git diff --check`, targeted `go test`, `make -k test`, then `sdlc local-check run --name suite` on the pushed code head; close through `sdlc close --issue 420 --verified '<evidence>'` and honor its fresh-context review.
- [ ] Re-measure #416 with `sdlc actual --issue 416` using current validated attempts, report the measured correction or a specific evidence gap to the TL/operator, and use the generation-bound `sdlc issue correct-actual` flow only with their decision. Do not type a replacement number from the worked example.

## Revisions

- 2026-10-17 — Initial full-flow design following the TL's interval/exact-evidence policy and the existing retained-replay contract.
- 2026-10-17 — TL approved the plan and corrected the legacy-time assumption: `started_at` is after post-lock setup, so using it as the queue endpoint shifts the old interval late. Keep the explicit legacy convention and test its bias; store both real queue endpoints in new records. No other design change.
