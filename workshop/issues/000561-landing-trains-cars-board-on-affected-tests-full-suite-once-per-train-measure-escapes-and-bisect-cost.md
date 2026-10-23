---
id: 000561
status: working
deps: []
github_issue:
created: 2026-10-22
updated: 2026-10-23
estimate_hours: 4.46
collaboration-mode: collaboration
card_mirror: '28881f06522dad33442dc98746c0e324f65405a8' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-23T00:12:42-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
---

# Landing trains: cars board on affected tests; full suite once per train; measure escapes and bisect cost

## Problem

## Spec

Landing trains (ariadne#449, #475) rarely batch: since 10-21 21:11 every pair train carried one car (trains 271, 277,
278). The car readiness gate requires a recorded **full suite** on each car's head before it boards ("needs recorded
local checks on its head: affected or its full-suite fallback", and in practice affected escalates to the full suite),
so each car pays a full suite, cars are rarely ready together with the queue an hour deep, and the train's single
suite saves little. pair ready→landed p50 52 min / p90 2.4 h (n=138), unchanged from the pre-train baseline.

Operator decision (10-22): **a car boards on its affected tests only** (pair declares them, pair#725; `make
test-affected`), and the train runs the **full suite once per train**, as originally designed. "Please monitor closely
the train test failure and the cost to bisect. All depends on whether we have a high-quality filter of affected tests
to catch issues early."

Scope:
- Readiness: a recorded **affected** check on the car's head suffices to board; no per-car full-suite fallback
  escalation for boarding (escalation stays for close/merge outside trains).
- The train's full suite stays the landing gate; a failing train splits (existing behaviour).
- **Measure the filter's quality** (the operator's condition): per train, record each failure the full suite found
  that the failing car's affected run passed ("escape"), the escape rate per repo over time, and the bisect cost
  (extra suite runs and wall time spent splitting). Surface it in `sdlc landing status` and a weekly line, so the TL
  can watch it and turn per-car full suites back on for a repo whose escape rate climbs.

- **Start small (operator, 10-22):** begin with trains of at most **2 cars** per repo, so a failing train splits once
  and an escape is cheap to attribute; raise the cap only after the escape rate stays low for about a week. The cap is a
  per-repo setting the TL adjusts, shown in `sdlc landing status`.

- **The default onboarding for any repo (operator, 10-22):** every repo that adopts trains starts with the small cap and ramps up by its own measured escape rate; the cap and the ramp rule live in one place, not per-repo improvisation.

## Done when

- A car with a passing affected record boards without a full suite (test); the train's full suite still gates.
- A train failure attributable to a car whose affected run passed is recorded as an escape, with the bisect cost
  (test with a fake suite).
- `sdlc landing status` shows the escape rate and bisect cost for the repo.
- The weekly artifact carries the same measurement definitions and explicit unknown coverage; existing snapshots remain immutable.
- Omitted train caps default to two, explicit caps remain unchanged, and one central seven-day measured-review rule governs manual ramp guidance without enabling trains or raising caps automatically.

## Plan

- [ ] M1 — Explicit boarding-only affected execution, conservative refusals and default two-car onboarding; retain ordinary fallback and train suite gates.
- [ ] M2 — Accepted-proof provenance, confirmed escape and actual split-cost measurement, status/weekly reporting and incomplete-coverage controls.

Durable design: [affected boarding and train measurements](../plans/000561-affected-boarding-and-train-measurements-plan.md).

## Log

### 2026-10-22

### 2026-10-23 — Planning checkpoint

- Claimed and entered planning on the issue branch. Read-only traces found that readiness already accepts affected records; `runAffectedCheck` escalates Full/graph failure into suite. Existing full integration suite, baseline, bisect attribution and reintegration remain authoritative (ARCH-DRY, ARCH-PURE).
- Hub confirmed explicit boarding-only mode: Full/unavailable selection refuses without running suite; keep selector authority and already-passing suite proof. Omitted cap defaults to two, explicit caps survive, and a central seven-day review informs guarded manual changes only. Authenticated delivery b92915a9-f0ca-4b8f-a469-406a36ebe9de was submitted.
- Proposed two review boundaries in the linked durable plan. Escape measurements distinguish confirmed qualifying failures from registered/untouched/unverified outcomes and disclose unresolved coverage. Count managed executions rather than scheduled `Train.Splits`; preserve immutable weekly v2 artifacts with additive optional measurement data.
- Checked `sdlc merge --help`: there is no per-call serial flag. Recovery text will print the exact ordinary suite/retry commands and name the guarded main configuration prerequisite (zero train interval) for serial landing; no new merge flag is proposed.
- No code edited, checks queued, PR opened or estimate supplied. Next: send the committed/pushed full plan to the hub for approval, then run `sdlc change-code --issue 561` for plan-quality and estimation before implementation. Isolated native CLI/train fixtures are the proposed live proof; no :0 or consumer settings will be touched without hub authorization.
- Hub approved both boundaries and the measurement definitions in submitted delivery 7014fc79-c66e-4c72-9189-3c38b303d775; isolated real-CLI, bare-remote and train-fake proof suffices, with no :0 change. Proceeding to `change-code`; the binary's plan-quality review and estimate gates still precede implementation.
- Plan-quality round one blocked on PQ-1 (function-named adversarial strategies), PQ-2 (explicit pure measurement transition ownership), and PQ-3 (nested proof unknown-field preservation). Refined the full classes in the plan: compressed test inventories into one strategy per risky function, added `AdvanceMeasurement` with private accumulation and legal lifecycle transitions, and specified nested strict preserving codecs plus binding-aware re-vouch behavior. Also traced missing-proof diagnostics and included boarding-mode recommendations, so normal recovery uses the new mode. No code or estimate yet; rerun plan-quality next.

## Revisions

- 2026-10-23: Refined implementation scope from the observed fallback owner, incorporated the hub's boarding/default-cap decisions, and added weekly compatibility/onboarding acceptance criteria already required by the operator's Spec. Full plan awaits approval.

## Estimate

Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only; `sdlc estimate-source` marks the calibration stale, so these numbers are provisional. Derived after plan-quality round two passed, not before.

Decomposition: one issue/spec/design effort; six smaller Go extensions (boarding CLI/execution, default-cap policy/rendering, nested proof codec, accepted CI provenance, managed-run observation, status/weekly wiring); two new single-concern pure modules (measurement lifecycle and aggregate/format); three review boundaries (M1, M2, issue close); one atlas/help maintenance effort. Existing selector, CI/ancestor gates, codec, runner, log and stateful fakes are reused. No new external API or unfamiliar stack needs discovery. No available library owns this domain-specific lifecycle/fold; standard encoding/time/maps and existing codecs cover infrastructure, so no greenfield infrastructure is charged.

For each smaller module, use v2 design 0.3 × thorough-spec 0.2 = 0.06 and impl 0.5 × v3.1 0.4 = 0.20. Each new pure module uses design 1.5 × 0.2 = 0.30 and impl 0.8 × 0.4 = 0.32. Spec authoring uses 0.75 design and 0.2 × 0.4 implementation; each review uses 0 design and 0.375 × 0.4 implementation. Atlas uses design 0.2 × 0.2 and impl 0.2 × 0.4. Familiarity 1.0 and thorough-plan design buffer 15%; design subtotal 1.75, implementation subtotal 2.45; 1.75 × 1.15 + 2.45 = 4.4625 ship-hours, rounded 4.46. Main risks are proof-preserving compatibility and interrupted measurement coverage, rather than novel API behavior.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec design=0.75 impl=0.08
item: smaller-go-module design=0.06 impl=0.20
item: smaller-go-module design=0.06 impl=0.20
item: smaller-go-module design=0.06 impl=0.20
item: smaller-go-module design=0.06 impl=0.20
item: smaller-go-module design=0.06 impl=0.20
item: smaller-go-module design=0.06 impl=0.20
item: greenfield-go-module design=0.30 impl=0.32
item: greenfield-go-module design=0.30 impl=0.32
item: milestone-review design=0 impl=0.15
item: milestone-review design=0 impl=0.15
item: milestone-review design=0 impl=0.15
item: atlas-docs design=0.04 impl=0.08
design-buffer: 0.15
total: 4.46
```
