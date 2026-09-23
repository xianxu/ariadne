---
id: 000242
status: working
deps: []
github_issue:
created: 2026-09-22
updated: 2026-09-22
estimate_hours:
started: 2026-09-22T22:20:59-07:00
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

## Plan

Task outline only; settle implementation design through start-plan before change-code.

Engineering proposal: [durable implementation plan](../plans/000242-slots-v2-workspace-identity-plan.md). Operator approved on 2026-09-22; change-code gates are in progress.

- [ ] Inventory live identity/path consumers and specify the shared resolver contract.
- [ ] Add Git fixture tests and implement the resolver and affected SDLC consumers.
- [ ] Document the contract for Couch and verify primary/non-slot compatibility.

## Log

### 2026-09-22 — fresh v2 task

Created from the agreed workspace/UI contract and the request for a clean task breakdown. Implementation has not started; estimates follow design approval.

### 2026-09-22 — claimed and engineering design started

Loaded the v2 project, claimed this issue, and ran start-plan. Read-only consumer audit found existing fleet NormalizeVantage and gitx.ParseWorktrees suitable for promotion into a shared package (ARCH-DRY). The durable proposal covers slot/address validation, a JSON CLI for Couch, and all identified SDLC identity/path consumers. Local checkout paths remain distinct from repository identity. Unrelated edits to #230 and #240 are preserved. Design review and operator approval precede change-code; estimate remains unset until plan-quality acceptance.

### 2026-09-22 — design review result

Fresh-context spec/plan review approved with no blocking findings. Incorporated its clarifications for unborn HEAD, primary readiness, active-time repo qualifiers and peer-write coverage. `sdlc issue validate --issue 242` passed. Awaiting operator approval of the durable plan before change-code, as required by AGENTS.md §2.

### 2026-09-22 — approved; plan-quality refinement

Operator approved the durable plan. The first change-code plan-quality review accepted architecture/scope and raised PQ-1: replace prose test-case lists with function-level adversarial strategies. Updated the plan across that class; rerunning the gate before estimating or implementing.

## Revisions

### 2026-09-22 — first engineering proposal

Reason: execution request starts the first v2 task. Delta: added the durable plan link and audit log while retaining the original task outline and completion contract. No implementation or lifecycle semantics changed.
