# Witnessed staging interval attribution implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Exclude witnessed stage/* time under concurrent reservations without assigning combined staging work to any issue.

**Architecture:** Extend the existing pure branch selector and its retained-proof contract. Reuse captured HEAD history, claim admission and warning rendering; no neighboring-branch inference, new capture, retrospective reassignment or new IO authority.

**Tech Stack:** Go, existing native Git reflog fixtures and measurement snapshots.

## Core concepts

| Name | Kind | Lives in | Status |
|------|------|----------|--------|
| IsStagingBranch | PURE | cmd/sdlc/internal/activetime/branch.go | new |
| SelectBranchActivity | PURE | cmd/sdlc/internal/activetime/branch.go | modified |
| MeasurementBranchInterval | PURE | cmd/sdlc/internal/gatestate/measurement_branch.go | modified |
| measurementBranchWarning / validateOwnedBranchProof | PURE | cmd/sdlc/actualbranch.go | modified |
| ParseHeadHistory | PURE | cmd/sdlc/internal/activetime/branchhistory.go | unchanged |

IsStagingBranch recognizes an attached, syntactically valid branch with literal stage/ prefix and a nonempty suffix, reusing historyBranch's validation. It is shared by live selection and retained validation (ARCH-DRY); no IDs are inferred from the suffix. The existing BranchNote/MeasurementBranchInterval gains a stage-excluded kind, distinct from resting-excluded, without changing old kinds.

| Name | Kind | Lives in | Status | Wraps |
|------|------|----------|--------|-------|
| captureOwnedBranchEvidence | INTEGRATION | cmd/sdlc/actualbranch.go | unchanged | bounded native Git HEAD records |

The existing capture supplies bytes to pure decoding and classification; warning rendering and saved-proof validation are pure too. Existing native temporary Git repositories and immutable captured records exercise the seam; no new external service or production telemetry mutation is needed (ARCH-PURE, ARCH-MOCK).

## Policy and compatibility

Apply the existing decision order: witnessed own issue selects; witnessed foreign issue excludes; a sole live reservation keeps labelled ownership-only planning; under concurrent reservations a witnessed resting branch excludes; a witnessed valid stage/* branch now excludes; otherwise refuse with the existing explicit --no-actual recovery. Unknown/detached stage-looking text never supplies a witness. Stage exclusion applies to each candidate issue independently, so combined work is unassigned to both.

Output names source, branch, interval and excluded branch-window minutes, explicitly distinguished from focused activity minutes. The peer capture must yield 172 seconds excluded (2.87 displayed minutes), while own/foreign spans around it keep their existing attribution.

ARCH-EVOLVE: add stage-excluded to v3 audit kinds. Existing strict readers refuse this previously unknown kind for that proof; no method/version bump is necessary because they cannot silently select it. Test the frozen older reader's refusal. Preserve unknown fields through existing codecs and keep duplicate/type validation strict. Existing adopted snapshots replay unchanged and are never rewritten.

ARCH-SECURE: branch authority comes only from captured, known attached HEAD evidence; literal stage text in an unknown interval or another arbitrary branch is insufficient. Saved records must validate stage-excluded against the same predicate, exclude its windows from numeric credit, and reject relabelling to credited or resting kinds. No credentials are touched.

ARCH-CONSTRAINTS: batch measurement keeps its existing bounded capture and sorted interval sweep; one constant-time classification per interval adds no subprocess, scan or durable artifact family. ARCH-ORDER: no new state between events; selection remains a deterministic transform of pinned claims, windows and history. ARCH-FUNERAL: reuse existing snapshot lifetime and archive policy; create no runtime residue. ARCH-PURPOSE: cover default actual/close and active-time paths through their shared resolver, plus saved replay and visible output.

## Implementation and adversarial verification

- [x] Add failing selector and capture-backed tests in cmd/sdlc/internal/activetime/branch_staging_test.go using pair#830's reported checkout/merge/return sequence and its 172-second boundaries. Reuse the native decoder to prove a merge preserves the stage witness. Both concurrent issue selections exclude that interval, with disjoint remaining windows. Controls: sole claim, own/foreign branch, stage-prefix near misses, malformed names, detached/unknown evidence and missing history; neighboring issue names must never supply an owner.
- [x] Implement IsStagingBranch and one stage-excluded selector arm after sole-reservation fallback, preserving resting precedence and bounded sweep behavior. Run `go test ./cmd/sdlc/internal/activetime -run 'Test.*(Stage|Staging|Resting|Branch)' -count=1` and confirm the new regression changes from refusal to exclusion.
- [x] Extend measurement_branch.go validation and actualbranch.go's warning/proof handling. Add tests in cmd/sdlc/internal/gatestate/measurement_branch_test.go and cmd/sdlc/actualbranch_resting_test.go for valid saved exclusion, exclusion overlapping credited windows, invalid branch/kind mutations, output minutes, old-reader refusal, and unchanged old adopted replay. Verify JSON/YAML unknown-member preservation using existing codec controls.
- [x] Run focused normal and race tests: activetime and gatestate packages; then cmd/sdlc with `-run 'Test.*(Branch|Resting|Staging|StageExcluded)'`. Check actual test names and enumerate them before trusting a regex run; retain the precise commands and results in the Log.
- [x] Update README.md's shared actual/active-time paragraph, cmd/sdlc/helptext/actual.md, cmd/sdlc/helptext/active-time.md and atlas/workflow/sdlc-binary.md together. Sweep existing resting-excluded consumers and stale refusal wording. Add a lessons.md line for witnessed neutral states and retained-proof parity.
- [x] Commit and push; launch `sdlc local-check run --pre-close` on the pushed head. PR 433 is open; the retained suite passed 2471 command tests and 77 packages with no failures/reruns, and close round 2 is SHIP.

Completion sequence: the binary-owned close review can overlap the detached suite; wait for its pass and SHIP before `sdlc merge --yes`. Report the Git-verified landing to the hub. No live actual adoption or peer checkout edit is part of qualification.

## Approval checkpoint

Approved by ops:0 on 2026-10-24 (receipt a7713eb0-c36e-4728-9500-4a29a1cca2e6); change-code inferred quick flow. One atomic review boundary, no Mx milestones. This appears within the quick-flow code/design limits; change-code owns flow inference.

## Revisions

2026-10-24 — Round-1 review BR-1/BR-2: include README's shared actual/close and active-time contract, which the initial documentation sweep missed. Add explicit PURE/INTEGRATION table columns and separate existing IO capture from pure decoding, rendering and proof validation. These are prose corrections; runtime policy and code are unchanged.

2026-10-24 — Approval recorded and plan renamed to change-code's canonical filename. Peer evidence supplies event descriptions and abbreviated object IDs, so the committed replay explicitly reconstructs those messages with synthetic IDs; the production Git/tracker test independently exercises staging and saved replay. Old-reader coverage distinguishes the observed pre-change v3 unknown-kind refusal (RED) from the frozen pre-v3 reader's version refusal.
