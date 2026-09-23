---
id: 000242
status: working
deps: []
github_issue:
created: 2026-09-22
updated: 2026-09-22
estimate_hours: 2.65
started: 2026-09-22T22:20:59-07:00
flow: {kind: full, provenance: operator}
---

# Slots v2: workspace identity

## Problem

Slot directory names differ from repository names. Commands must distinguish the repository, its primary checkout, and the current workspace without treating pair-slot1 as a new repo.

## Spec

Project: `pair/workshop/projects/couch-slots-v2.md`. Fresh task derived from the current v2 contract; historical task bodies are not prerequisites or implementation plans.

Define one shared workspace resolver consumed by SDLC and exposed for Couch provisioning. A slot is a durable Git worktree, independent of its active branch or thread. For a primary /workspace/pair, :1 lives at /workspace/worktree/pair-slot1 and rests on main-slot1; :0 is the primary and rests on main. Within a known repo, :N is shorthand for repo:N. Resolve nested working directories and canonical Git worktree membership; directory spelling alone must not confer ownership. Recognize slots while an issue branch is checked out.

Audit SDLC consumers of repo basename, primary path, sibling peer paths, fleet identity, project discovery, and the calibration brain path. Preserve ordinary non-slot worktree behavior. ARCH-DRY: derive identity once rather than duplicating path parsers across commands. Decide the smallest shared machine-readable contract during design; no new role or issue-ownership model.

## Done when

- Primary, two slots, nested cwd, and an ordinary feature worktree resolve to the correct repo and workspace identities.
- State output identifies repo address, current branch, and resting branch; :0 and qualified addresses resolve consistently.
- Misleading paths, wrong Git membership, and an occupied/mismatched resting branch produce explicit errors without mutation.
- Affected issue/project/calibration paths use the intended repositories from a slot; regression tests cover production consumers.

## Estimate

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only.* Calibration is marked stale; treat the result as provisional.

Decomposition follows the accepted plan, not a target total. Task 1 is one cross-cutting refactor (design 0.5 × 0.2; impl 0.5 × 0.4). Task 2 is two greenfield single concerns: address/classification and Git snapshot validation with fake/conformance (each design 1.0 × 0.2; impl 0.8 × 0.4). Existing fleet Git normalization/porcelain parsing are reused in Task 1; no external library supplies the new Ariadne slot conventions, so no further greenfield library discount. Task 3 is a smaller Go extension (design 0.2 × 0.2; impl 0.4 × 0.4). Task 4 splits into artifact/project and calibration/remaining consumers, two cross-cutting refactors (each design 0.6 × 0.2; impl 0.5 × 0.4). Task 5 adds docs (design 0.1 × 0.2; impl 0.2 × 0.4) and one boundary review (design 0.2 × 0.2; impl 0.5 × 0.4). Familiar Go/Git/Cobra stack: familiarity 1.0. Thorough approved plan: design discount 0.2 and buffer 15%. No cross-repo implementation or new external API.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: cross-cutting-refactor design=0.10 impl=0.20
item: greenfield-go-module design=0.20 impl=0.32
item: greenfield-go-module design=0.20 impl=0.32
item: smaller-go-module design=0.04 impl=0.16
item: cross-cutting-refactor design=0.12 impl=0.20
item: cross-cutting-refactor design=0.12 impl=0.20
item: atlas-docs design=0.02 impl=0.08
item: milestone-review design=0.04 impl=0.20
design-buffer: 0.15
total: 2.65
```

Design 0.84 × 1.15 + implementation 1.68 = 2.646 hours, rounded to 2.65.

## Plan

Task outline only; settle implementation design through start-plan before change-code.

Engineering proposal: [durable implementation plan](../plans/000242-slots-v2-workspace-identity-plan.md). Operator approved on 2026-09-22; change-code gates are in progress.

- [x] Inventory live identity/path consumers and specify the shared resolver contract.
- [x] Add Git fixture tests and implement the resolver and affected SDLC consumers.
- [x] Document the contract for Couch and verify primary/non-slot compatibility.

## Log

### 2026-09-22 — fresh v2 task

Created from the agreed workspace/UI contract and the request for a clean task breakdown. Implementation has not started; estimates follow design approval.

### 2026-09-22 — claimed and engineering design started

Loaded the v2 project, claimed this issue, and ran start-plan. Read-only consumer audit found existing fleet NormalizeVantage and gitx.ParseWorktrees suitable for promotion into a shared package (ARCH-DRY). The durable proposal covers slot/address validation, a JSON CLI for Couch, and all identified SDLC identity/path consumers. Local checkout paths remain distinct from repository identity. Unrelated edits to #230 and #240 are preserved. Design review and operator approval precede change-code; estimate remains unset until plan-quality acceptance.

### 2026-09-22 — design review result

Fresh-context spec/plan review approved with no blocking findings. Incorporated its clarifications for unborn HEAD, primary readiness, active-time repo qualifiers and peer-write coverage. `sdlc issue validate --issue 242` passed. Awaiting operator approval of the durable plan before change-code, as required by AGENTS.md §2.

### 2026-09-22 — approved; plan-quality refinement

Operator approved the durable plan. The first change-code plan-quality review accepted architecture/scope and raised PQ-1: replace prose test-case lists with function-level adversarial strategies. Updated the plan across that class; rerunning the gate before estimating or implementing.

### 2026-09-22 — implementation and regression evidence

Implemented the shared pkg/workspace resolver and parser promotion, JSON workspace CLI/state, current-checkout project overlay, canonical artifact/review/actual labels, default brain/project paths, migration guard and ordinary worktree placement. Focused production-command suites passed. Shared/fleet/gitx suites and parser/address fuzz runs passed; real resolver samples at 1/10/100 worktrees were 95.796/121.384/154.675ms with 7/9/9 Git reads. Resolver final checks cover selected path/common-dir, HEAD, branch and resting ref; occupancy remains observational and future mutation commands must revalidate under their own lock (ARCH-ORDER).

The first full suite exposed legacy non-Git fixtures, corrected to real Git repositories; macOS logical/physical temporary paths also needed canonical fixture identity for fault injection. The unchanged TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory fails on its missing live plan path, already tracked by #210; HEAD lacks that path and this task does not change that test. Final serialized suite will exclude only that named baseline failure. No history files read. Unrelated #230/#240 edits remain intact.

### 2026-09-22 — integrated verification checkpoint

Passed `go test ./pkg/workspace/... ./cmd/sdlc/... -count=1 -skip '^TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory$'` (cmd/sdlc 271.462s, all packages pass). The sole exclusion is the unchanged baseline #210 missing-plan test, documented above. `go vet ./pkg/workspace/... ./cmd/sdlc/...`, `git diff --check`, and `go build -o /tmp/ariadne-242-sdlc ./cmd/sdlc` passed. Built-binary workspace/state JSON smoke checks passed on this checkout; mutation/error scenarios used temporary Git fixtures. Implementation and documentation are complete; mandatory close review and publish remain.

### 2026-09-22 — close review round 1

REWORK: BR-1 found malformed all-zero HEAD observations bypassed OID validation by becoming unborn null; BR-2 found missing README command usage. Fixing the malformed-evidence class with red/green regression coverage and adding README workspace/state discovery. The issue remains working; no gate bypass. Lessons and the durable plan record both corrections.

### 2026-09-22 — review fixes verified

BR-1 class sweep now validates every observed non-bare HEAD and the selected resting ref in Classify before interpreting zero sentinels; direct callers share the same rule. Malformed-observation regression failed before the fix (0.432s), then focused tests passed (0.387s); full workspace suite passed (9.518s). Workspace/state/project/close/resolve consumer regressions passed (10.608s), and Go vet passed across workspace and SDLC. BR-2 README now documents workspace address examples, JSON purpose and state integration. Both findings are ready for gate disposition; no publish yet.

## Revisions

### 2026-09-22 — first engineering proposal

Reason: execution request starts the first v2 task. Delta: added the durable plan link and audit log while retaining the original task outline and completion contract. No implementation or lifecycle semantics changed.
