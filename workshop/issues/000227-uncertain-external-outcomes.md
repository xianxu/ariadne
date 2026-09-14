---
id: 000227
status: working
deps: []
github_issue:
created: 2026-09-14
updated: 2026-09-14
estimate_hours:
started: 2026-09-14T15:54:50-07:00
---

# Model uncertain external outcomes

## Problem

ARCH-ORDER names ordering and enforcement, but leaves external uncertainty implicit. A timeout or failed observation can be mistaken for confirmed failure or absence, permitting unsafe retries or cleanup.

## Spec

Add the user-approved knowledge/uncertainty contract to ARCH-ORDER in `cmd/sdlc/internal/judge/architecture.md`: distinguish desired state, observed state and operation outcome where relevant; external operations define confirmed success, confirmed failure and unconfirmed outcome, each with permitted actions and reconciliation. Timeout or missing acknowledgment does not prove an effect did not occur. Retain partial-progress evidence. Pure decisions can request probes, but probe failure remains an event and an observation is not permanent authority.

Make planning name evidence and bounded reconciliation/safe retry rules. Make review flag fabricated certainty, unsafe retries/rollback, and tests omitting late completion or failed probes. Existing ARCH-PURE/ARCH-ORDER own the distinction between deterministic decisions and uncertain effects; ARCH-DRY/ARCH-PURPOSE keep all prompt consumers deriving from one registry. No new runtime machinery, FSM framework, external IO dependency, data parser or durable artifact family (ARCH-MOCK/CONSTRAINTS/SECURE/FUNERAL); this change strengthens policy only.

## Done when

- The canonical registry explicitly states uncertainty and its planning/review obligations.
- Full judge tests, CLI architecture/start-plan delivery tests, inspected generated prompt changes and built CLI/source comparison pass.
- Mandatory boundary review accepts the policy and its derived consumers.

## Plan

- [x] Add uncertainty guidance and relevant at-plan/at-review checks to ARCH-ORDER.
- [x] Update derived prompt snapshots and verify their only change is the registry replacement; check atlas references for contradictory descriptions.
- [x] Run full judge tests, CLI delivery tests, built CLI/source comparison and git diff --check.

After verification, commit and close through SDLC, then publish via PR and merge.

## Log

### 2026-09-14

User approved adding the uncertainty/reconciliation guidance following the Pair audit. Created and claimed #227 first. Bounded policy wording update with an issue-local plan; separate planning review and estimate gates waived using their precise flags because the design was just approved in conversation. Mandatory close review remains enabled. Verification exercises BuildPrompt, runArchPrinciples and runStartPlan; full-registry assertions and exact snapshot replacement detect missing, partial or unintended delivery changes. No new tests that merely duplicate prose.

Verification: `go test ./cmd/sdlc/internal/judge -count=1` and `go test ./cmd/sdlc -run 'TestRunArchPrinciples|TestArchPrinciplesCmd|TestRunStartPlan|TestStartPlanCmd' -count=1` passed. All four golden files changed only by exact replacement of the old embedded registry with the new one. `go build -o bin/sdlc ./cmd/sdlc` passed and a subprocess assertion verified the complete registry in `bin/sdlc arch-principles`. `git diff --check` passed. Atlas ordering/ownership descriptions remain consistent and point to the registry; no duplicate clause needed.
