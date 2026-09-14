---
id: 000226
status: working
deps: []
github_issue:
created: 2026-09-14
updated: 2026-09-14
estimate_hours:
started: 2026-09-14T15:37:11-07:00
---

# Enforce state-machine ownership and behavioral invariants in ARCH-ORDER

## Problem

ARCH-ORDER calls for explicit states and transitions but does not explicitly require structural and behavioral enforcement of the executable model. A diagram or bypassable FSM component can satisfy the wording without governing production state changes.

## Spec

Strengthen the existing ARCH-ORDER entry in `cmd/sdlc/internal/judge/architecture.md` with the wording approved in conversation:

- Structural enforcement: an explicit state/event/transition component in the pure core owns authoritative state changes; types, encapsulation and dependency checks prevent bypasses. The IO shell executes declared effects and returns outcomes as events.
- Behavioral enforcement: independently stated invariants are checked across applicable failure, cancellation, retry, duplicate and out-of-order event sequences using the production transition function, controllable ordering and reproducible failures. Check permitted and rejected/ignored events; transition coverage alone is insufficient.
- At plan, name the component, state/event model, invariants, enforcement mechanism and sequence-testing strategy. At review, verify production routing, prevention of bypasses and independent invariant assertions.

This is a bounded policy change to the existing embedded registry, not implementation of an FSM framework or a new static analyzer. ARCH-DRY/ARCH-PURPOSE: existing consumers must continue deriving the full registry; no copied policy. ARCH-PURE/ARCH-ORDER shape the added contract. The change adds no runtime workload, external dependency, untrusted-data boundary or durable artifact family beyond the normal issue record (ARCH-CONSTRAINTS/MOCK/SECURE/FUNERAL).

## Done when

- ARCH-ORDER explicitly requires structural and behavioral enforcement and supplies concrete at-plan/at-review checks.
- Existing prompt and CLI delivery tests pass against the changed registry, and the built arch-principles command delivers the exact registry.
- The diff remains scoped to the approved policy and its issue record; boundary review has no unresolved blocking findings.


## Plan

- [x] Update the canonical ARCH-ORDER entry, preserving existing ordering guidance.
- [x] Run judge architecture tests and CLI architecture/start-plan delivery tests; compare built CLI output with the source; run git diff --check.
After implementation verification: commit, close with mandatory boundary review and measured actuals, then publish via SDLC.

## Log

### 2026-09-14

User approved the proposed enforcement wording and requested an ariadne ticket first. Created and claimed #226 before editing policy. Single-pass policy update; issue-local plan is sufficient. Existing full-registry delivery tests cover propagation; avoid duplicating prose in new assertions.

Plan-quality passed (INFO). Its minor named-test-surface finding is addressed by the revision below. Estimate gates explicitly waived for this bounded, user-approved policy wording change; no implementation estimate is being fabricated.

## Revisions

### 2026-09-14 — Verification clarification after plan review

The verification step covers `BuildPrompt`, `runArchPrinciples`, and `runStartPlan`: existing full-registry assertions detect missing or partial policy delivery. Compare freshly built CLI output against the canonical registry as an additional executable check. No new functions or independent policy copies are introduced.

### 2026-09-14 — Separate lifecycle actions from acceptance checklist

Converted the close/publish row into a subsequent lifecycle instruction: close requires the implementation checklist complete before running, so closing itself cannot be a pre-close acceptance checkbox. Scope unchanged.

Verification: `go test ./cmd/sdlc/internal/judge -run 'TestArchitecture|TestDeferred' -count=1` passed; `go test ./cmd/sdlc -run 'TestRunArchPrinciples|TestArchPrinciplesCmd|TestRunStartPlan|TestStartPlanCmd' -count=1` passed. `go build -o bin/sdlc ./cmd/sdlc` passed; a Python subprocess assertion confirmed the entire source registry is contained in `bin/sdlc arch-principles` output. `git diff --check` passed.
