---
id: 000242
status: open
deps: []
github_issue:
created: 2026-09-22
updated: 2026-09-22
estimate_hours:
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

- [ ] Inventory live identity/path consumers and specify the shared resolver contract.
- [ ] Add Git fixture tests and implement the resolver and affected SDLC consumers.
- [ ] Document the contract for Couch and verify primary/non-slot compatibility.

## Log

### 2026-09-22 — fresh v2 task

Created from the agreed workspace/UI contract and the request for a clean task breakdown. Implementation has not started; estimates follow design approval.
