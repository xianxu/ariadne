---
id: 000167
status: working
deps: []
github_issue:
created: 2026-07-07
updated: 2026-07-12
estimate_hours: 1.08
started: 2026-07-12T22:02:15-07:00
actual_hours: 2.10
---

# institutionalize using continuation for long running session

## Problem

Ariadne has an agent-neutral `continuation` datatype for preserving the human-
meaningful state of a long-running session, and Pair's continuation writer can
durably write it and restart the session from it. Agents currently initiate that
flow only after an explicit operator request. They are not told to use it
proactively as context fills, so long sessions can reach compaction with less
room to create a good handoff. In metis-v2, an agent learned this behavior after
the operator demonstrated it once; the behavior should be institutionalized for
every Ariadne consumer instead of relearned per project.

## Spec

Add the proactive context-pressure policy to `AGENTS.base.md`, the always-loaded,
agent-neutral constitution exported to every harness entry file.

- When the harness reports that the active context is more than 60% full, the
  agent proactively checkpoints the session before starting another substantial
  unit of work. If an exact percentage is unavailable, a harness context-pressure
  or compaction warning is the equivalent trigger.
- The checkpoint completes the current atomic action, updates relevant durable
  work records, and applies the canonical `continuation` datatype. The
  constitution owns **when** to checkpoint; `construct/datatype/continuation.md`
  remains the single source for **what** to preserve and **how** to finalize it
  (`ARCH-DRY`).
- The agent invokes the available continuation writer and does not separately
  perform a restart. In Pair, the existing writer owns write, commit, push, and
  automatic restart from the continuation. The policy must not duplicate that
  mechanism or add Pair-specific implementation to Ariadne (`ARCH-PURE`).
- Keep this policy in `AGENTS.base.md`, not the continuation prototype or a
  harness-specific entry file, so every derived repository and supported agent
  sees the trigger before it needs to discover the datatype (`ARCH-PURPOSE`).
- Update the workflow atlas to map the proactive trigger without duplicating the
  datatype's section-level procedure.
- Add a regression guard that pins the constitution's threshold, route to the
  continuation datatype, and writer-owned restart boundary.

Out of scope: context-meter/model-window catalogs, automatic threshold polling
inside Pair, changes to the continuation document shape, and new restart
machinery.

## Done when

- `AGENTS.base.md` tells agents to proactively create a continuation above 60%
  context utilization, with a warning-based fallback when percentage is unknown.
- The rule routes to the canonical continuation datatype and leaves restart to
  the writer.
- The rule propagates through the existing base-layer prose composition rather
  than being copied into individual harness files.
- Automated tests guard the policy boundary, and the atlas describes the
  resulting lifecycle.

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec design=0.50 impl=0.04
item: skill-or-dispatcher design=0.06 impl=0.12
item: atlas-docs design=0.05 impl=0.04
item: milestone-review design=0.05 impl=0.12
design-buffer: 0.15
total: 1.08
```

Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md`
against `baseline-v3.1.md`. Method A only. The source was marked stale by
`sdlc estimate-source` on 2026-07-12, so the calibration is provisional.

## Plan

- [x] Add the session-continuity policy to `AGENTS.base.md`.
- [x] Replace the marker-only guard with a feature-specific weave integration guard.
- [x] Update the workflow atlas and verify base-layer composition.
- [x] Resolve the close-review findings and re-run full verification.
- [x] Harden section scoping and typed export parsing from the FIX-THEN-SHIP review.
- [x] Derive the feature test's consumer sweep from the canonical harness registry.
- [x] Require every registered harness to contain the complete scoped policy.
- [x] Pin the fallback condition and atomic/durable checkpoint ordering.

Detailed execution plan:
`workshop/plans/000167-institutionalize-using-continuation-for-long-running-session-plan.md`.

## Log

### 2026-07-07

### 2026-07-12
- 2026-07-12: closed — Complete fan-out contract verified: every plan.TargetAll.EntryFiles() output must contain the full scoped Session Continuity policy; go test ./... -count=1 passes, issue validation passes, and git diff --check is clean. Mutation evidence covers wrong direction, moved markers, commented/broken export; full-string consumer assertions now also reject partial composition.; review verdict: FIX-THEN-SHIP
- 2026-07-12: closed — Final review remediation verified: TestSessionContinuityPolicyFansOutToEveryHarness passes while deriving its consumer sweep from plan.TargetAll.EntryFiles(); go test ./... -count=1 passes in the writable workspace; issue validation and git diff --check pass. Earlier mutation checks also proved wrong-direction, moved-marker, commented-export, and broken-export regressions fail.; review verdict: FIX-THEN-SHIP
- 2026-07-12: closed — Post-review hardening verified: go test ./cmd/weave -run TestSessionContinuityPolicyFansOutToEveryHarness -count=1 and go test ./... -count=1 pass; issue validation and git diff --check pass. Mutation checks proved the guard fails when the canonical route is changed inside Session Continuity and when the AGENTS.base.md export is only commented out. The test now scopes policy assertions to the owning section and parses active manifest intents.; review verdict: FIX-THEN-SHIP
- 2026-07-12: closed — TDD remediation proved wrong-direction and broken-export mutants fail; go test ./cmd/weave -run TestSessionContinuityPolicyFansOutToEveryHarness -count=1, go test ./cmd/weave/... -count=1, and go test ./... -count=1 passed; sdlc issue validate passed; git diff --check clean; the integration guard pins semantic threshold, live base export, and all Claude/Codex/Gemini consumers.; review verdict: FIX-THEN-SHIP

Claimed the issue and inspected the continuation datatype plus Pair's existing
writer/restart path. Design decision: `AGENTS.base.md` owns the proactive trigger;
the datatype continues to own the handoff procedure, and Pair's writer continues
to own restart. This avoids duplicating either contract (`ARCH-DRY`, `ARCH-PURE`)
while delivering the agent-neutral behavior the issue asks for (`ARCH-PURPOSE`).

The first `sdlc change-code` attempt correctly refused because the shorter plan
filename was invisible to its exact issue-basename resolver. Renamed the plan to
the canonical discoverable path and retained the same reviewed scope.

Implemented the constitution-owned trigger with TDD: the focused policy guard
failed on the missing section, then passed after `AGENTS.base.md` gained the 60%
trigger, both fallback signals, the datatype route, and writer-owned restart
boundary. Added the referential atlas mapping without changing the continuation
prototype or Pair. Verified with the focused datatype test, all datatype tests,
`go test ./cmd/weave/... -count=1`, `go test ./... -count=1`, `git diff --check`,
and a scoped diff/name-only audit showing exactly the three planned files.

The first close review returned REWORK. It verified the policy and ownership
boundaries, then found that the plan incorrectly classified a filesystem-backed
contract test as PURE, the threshold assertion accepted reversed semantics, and
the feature test did not directly prove the three harness consumers derive from
the exported base fragment. Re-planned to treat the policy as an INTEGRATION
contract and exercise the real weave fan-out with exact semantic assertions.

Resolved the REWORK findings by replacing the `cmd/datatype` marker test with an
end-to-end `cmd/weave` integration test. The guard now pins the direction and
checkpoint boundary, checks the live export row, and composes the real base
fragment into all three harness entry files. Wrong-direction and broken-export
mutants both failed before restoration; the focused guard, weave suite, full Go
suite, issue validation, and whitespace check then passed. Recorded the reusable
classification/contract-testing rule in `workshop/lessons.md`.

The second close review returned FIX-THEN-SHIP and advanced the issue to
`codecomplete`, but found two Important test gaps: an unanchored manifest
substring accepted a commented-out export, and marker checks were not scoped to
the Session Continuity section. Reopened to `working` through `sdlc issue
set-status` to fix both before shipping.

Resolved the FIX-THEN-SHIP findings: the guard now slices only the Session
Continuity section, parses the live base manifest into typed intents, and builds
its fixture from the validated active export. A moved-marker mutant and a
commented-export mutant both failed before restoration; the focused test passed
afterward. Extended the #167 lesson with section-scoping and structured-source
parsing rules.

The final close review confirmed the behavior and documentation, then identified
one remaining `ARCH-DRY`/`ARCH-PURPOSE` weakness: the feature test repeated the
current harness filenames. Reopened the issue and changed the sweep to derive
from `plan.TargetAll.EntryFiles()`, so adding a future registered harness also
extends this policy contract automatically.

The subsequent close review confirmed the registry derivation but found that
each generated consumer was checked for only the heading and threshold. Reopened
again and strengthened the fan-out assertion to require the complete scoped
policy, including datatype routing and writer-owned restart semantics, in every
registered harness (`ARCH-PURPOSE`).

The next review validated complete propagation and identified two source-policy
semantics that still lacked independent mutation guards. Reopened and added exact
markers for the percentage-unavailable fallback, finishing the atomic action,
updating durable state first, and the successful-durable-write restart
precondition.
