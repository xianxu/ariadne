---
collaboration-mode: discussion
---

# Retired boundary accounting implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy). Execute with superpowers-executing-plans; use bounded agents only for independent exploration. Steps use checkbox syntax.

**Goal:** A live milestone following a retired, never-finalized review boundary measures its actual work instead of silently returning zero.

**Architecture:** Derive fresh measurement cuts from the exact code pin's canonical issue Plan and boundary ledger. Retired unfinalized boundaries cannot claim later work; finalized boundaries and executable open milestones retain their existing semantics. Store the resulting ordinary cuts in the existing snapshot, so replay remains immutable.

**Tech Stack:** Go, existing activetime/gatestate/markdoc packages, isolated native Git and transcript fixtures.

## Decision for hub approval

Recommend the issue's permitted last-closed-boundary fallback: M3 owns post-M1 work when M2 has been removed from executable milestone rows without finalization. The time spent on the discarded approach remains issue work and is included in the next increment; the diagnostic names this accounting choice. This is not a SHIP claim or retirement of review findings.

Alternatives: parse a prose supersession date or first M3 commit (fragile and excludes design/untagged work), or add a new supersession command and persisted event (larger lifecycle contract). Neither is required for the stated last-closed-boundary acceptance option. Mere selection of M3 does not retire an executable M2.

## Core concepts

| Pure entity | Lives in | Status |
|---|---|---|
| Fresh split-cut selection | cmd/sdlc/actualsplit.go | new |
| Effective phase diagnostic | cmd/sdlc/actualsplit.go | new |
| BoundaryRound / Timeline | cmd/sdlc/internal/activetime/split.go | existing, unchanged |
| MeasurementCut / MeasurementInputs | cmd/sdlc/internal/gatestate/measurement.go | existing, unchanged |

The selector accepts parsed canonical milestone membership and usable rounds; returns selected rounds plus retired boundary names. Group by boundary before deciding: any finalizing round preserves the entire boundary. Reuse cutTimes, splitRounds and issue.PlanItemsBody / MilestonesInPlanOrder. No review finding, cap or gate ledger is rewritten. The phase diagnostic derives from the same Timeline used for hours, not from a second window algorithm.

| Integration | Lives in | Status | Wraps |
|---|---|---|---|
| Pinned issue Plan resolver | cmd/sdlc/actualplan.go | new | existing ownedActualGit and exact Git tree |
| Fresh owned input construction | cmd/sdlc/actualinputs.go | modified | pinned code/tracker and retained snapshots |
| Legacy measurement path | cmd/sdlc/actual.go | modified | legacy Git/transcript discovery |
| Measurement rendering | cmd/sdlc/actual.go, actualmeasurement.go | modified | human-readable actual/close evidence |

Resolve exactly one issue details file at the code pin from configured active issue root or archive root; reuse existing discovery helpers where applicable. Read only that pin, never later HEAD or an uncommitted editor buffer. Match exact issue identity, reject ambiguous matches. A canonical Plan is required only when membership is needed to resolve a named unfinalized boundary; ordinary no-round/single whole-issue measurements retain their current contract. An explicitly requested milestone must remain executable, unless replaying its retained or finalized historical boundary.

## State and evidence table

| Evidence at pinned code generation | Fresh measurement behavior |
|---|---|
| Boundary has finalizing round | Keep all its rounds, including re-close history |
| Unfinalized named boundary remains executable | Keep its existing open convergence semantics |
| Unfinalized named boundary absent from valid canonical Plan | Exclude its rounds from fresh cuts; name it as retired for accounting |
| Whole-issue boundary or no review rounds | Existing semantics |
| Required Plan unreadable, ambiguous, missing or malformed | Refuse measurement; identify pinned path and restore/commit/retry action |
| Requested live milestone absent | Refuse; restore the executable milestone or select its replacement |
| Retained snapshot exists | Replay saved Cuts without re-filtering against current or historical Plan |

A still-executable open boundary that consumes the requested later milestone must produce a visible diagnostic naming that boundary; do not reinterpret it as retired merely to produce positive hours. Zero genuine activity stays zero. This pure evidence selection has no asynchronous effects or retry state; cancellation/read failure exits before adoption. The existing close transaction remains the sole writer.

## Operating envelope and compatibility

ARCH-CONSTRAINTS: on-demand CLI accounting, not a UI hot path. One pinned issue-body read per new derivation; selection O(rounds + plan bytes), bounded by existing artifact/snapshot limits. No per-round Git subprocess or history scan; native tests count resolver calls at N and 2N round inputs. No new concurrency, network, credential or live-fleet access.

ARCH-SECURE: version-pinned repository text is input, not an instruction. Parse with existing fence-aware Plan parser, distinguish absent from empty, verify identity and uniqueness; do not interpret arbitrary superseded prose. A dirty working copy cannot override the code pin. ARCH-ORDER: explicit table above governs retirement; finalized history wins over current membership. Reader cancellation cannot trigger writes.

ARCH-EVOLVE/FUNERAL: creates no new durable family, fields or interpretation of saved cuts. Selected cuts already express the intended timeline under old readers, so no schema/method-version bump is needed. Old snapshots retain old results; this issue does not silently rewrite historical calibration or repair pair#737's adopted sample. Existing artifact archive/snapshot bounds and cleanup apply.

ARCH-DRY/PURE: one selector feeds fresh tracker and legacy paths; BuildTimeline and Split remain the pure accounting engine. ARCH-MOCK: use existing native private Git/transcript harnesses, with portable HOME and no real fleet state. ARCH-PURPOSE: cover actual preview, milestone-close/close fresh candidates and historical no-snapshot reconstruction; saved replay/correction retains its original inputs. Tests demonstrate both sides rather than only positive M3 output.

## Function-level verification contracts

| Function | Adversarial strategy and mechanical guard |
|---|---|
| selectFreshSplitRounds | Generate arbitrary boundary round order and canonical membership; assert finalized-history preservation, executable-open preservation, retired-only omission and idempotence in pure property tests. |
| pinnedActualPlan | Supply adversarial private Git trees and mutable checkout state; native pinned-read tests enforce exact generation/identity/unique Plan authority and count a constant observation budget at N/2N ledger sizes. |
| actualPhaseDiagnostic | Generate selected and blocked/empty phase sets; pure tests require visible effective bounds and no invented contributing-session claim. |
| resolveOwnedActualInputs | Native tracker/Git fixture changes plan generations around fresh/replay selection; assert new ordinary Cuts, pinned authority and refusal before adoption. |
| computeActualFor / computeOwnedActualFor | Replay sanitized pair#737 before M3 review using portable transcript fixtures; exact positive M3 totals and credit conservation are the oracle. |
| resolveOwnedReplay / ownedReplaySources | Save filtered and historical unfiltered Cuts, then mutate later state; assert each replays its own original hours/credits. |
| printActualMeasurement | Render production measurement results and assert semantic bounds/code pin/source-count diagnostics, including visible unresolved executable boundaries. |

## Tasks — one atomic review boundary

### 1. Capture the failure and pure contract

Files: cmd/sdlc/actualsplit_test.go, new cmd/sdlc/actualplan_test.go.

- [x] Write the sanitized pair#737 production-path regression before its first M3 review; run go test ./cmd/sdlc -run 'Test.*RetiredBoundary' -count=1 and observe the existing zero-hour failure.
- [x] Add the named pure contracts above as executable property tests; use the state table as the invariant source.

### 2. Select fresh cuts from the pinned Plan

Files: new cmd/sdlc/actualplan.go; cmd/sdlc/actualsplit.go, actualinputs.go, actual.go.

- [x] Implement selectFreshSplitRounds and pinnedActualPlan with the authority and failure contracts above, reusing existing parsers and Git seams.
- [x] Connect both fresh measurement paths and no-snapshot historical reconstruction. Keep saved replay outside fresh selection.
- [x] Execute the named native resolver and input-selection strategies; confirm red-to-green production regression and constant resolver observation budget. Commit the coherent implementation with its tests.

### 3. Explain and verify the result

Files: cmd/sdlc/actual.go, actualmeasurement.go, actualmeasurement_test.go, actualinputs_test.go, helptext/actual.md, atlas/workflow/sdlc-binary.md.

- [x] Implement actualPhaseDiagnostic and expose it through actual/close rendering with the existing provenance diagnostics.
- [x] Execute the named compute, replay and rendering contracts using disposable Claude and Codex sources, with the available Qoder registry control.
- [x] Document the fallback and retained-history limitation in help and atlas.
- [x] Run go test ./cmd/sdlc/internal/activetime ./cmd/sdlc/internal/gatestate -count=1 and targeted cmd/sdlc measurement tests. Push, run sdlc local-check run --name suite and retain the runner through completion.

Follow-through (pending): close with measured verification and mandatory review; on SHIP land through sdlc merge --yes.

## Proof and review point

Request hub approval of canonical Plan membership as the fresh accounting authority and post-last-finalized fallback, plus isolated production-path replay as the live proof. No pair:0 stage or calibration correction is proposed. If membership proves insufficient to distinguish intended retirement in the production fixtures, stop and return to design rather than adding prose heuristics or a new lifecycle silently.

## Revisions

### 2026-10-22 — PQ-1: executable function contracts

Named every new risky function and the existing production entry points, replacing procedural test-case inventories with one adversarial strategy and mechanical guard per function. The evidence state table remains the accounting contract. No scope or policy change.

### 2026-10-22 — completion bookkeeping

Separate the pending close/landing procedure from completed verification, without marking a future gate as already passed. All implementation tasks and the corrected recorded suite are complete; mandatory close and landing remain pending.
