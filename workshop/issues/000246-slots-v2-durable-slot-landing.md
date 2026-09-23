---
id: 000246
status: open
deps: [ariadne#242, ariadne#244, ariadne#245]
github_issue:
created: 2026-09-22
updated: 2026-09-22
estimate_hours:
---

# Slots v2: land while retaining the workspace

## Problem

Current SDLC merge cleanup removes feature worktrees and returns in-place work to updated main. A durable slot must survive publication and return to its unchanged resting branch.

## Spec

Project: `pair/workshop/projects/couch-slots-v2.md`. Fresh task derived from the current v2 contract; historical task bodies are not prerequisites or implementation plans.

Adapt existing publication/review/merge flow for durable slots. Confirm integration into the configured remote main, safely return to the existing main-slotN, then safely remove the completed local issue branch. Do not refresh the resting branch. Preserve directory, unpublished commits, dirty files, and other workspaces. Do not operate on the primary checkout merely to complete a slot’s cleanup or archive its artifacts.

Account for squash/rebase merge evidence when deciding that branch deletion is safe; an upstream relationship or branch name is insufficient proof. Detect commits added after the integrated head. Define recoverable phases for integration completed but switch/deletion/archive cleanup interrupted. Re-running must recognize already-confirmed integration and finish safely without duplicating publication. Keep ordinary feature-worktree and :0 behavior explicit and tested. ARCH-FUNERAL: retain reusable workspaces intentionally, while completed issue branches have an evidence-backed removal path.

## Done when

- A slot completes review/publication and lands; its path survives and its resting SHA/upstream are unchanged.
- Primary and another dirty active slot retain their branch, files, and refs throughout landing.
- Unpublished additions, dirty switch collisions, and uncertain integration evidence block destructive cleanup and explain recovery.
- Integration followed by interrupted cleanup can be retried safely, including the repository’s actual merge strategy.
- Issue archiving/project updates complete through existing gates without requiring a clean idle primary; ordinary worktree behavior has regression coverage.

## Plan

Task outline only; settle implementation design through start-plan before change-code.

- [ ] Trace current merge/publication/archive phases and specify durable-slot recovery.
- [ ] Add integration fixtures, then adapt landing and safe cleanup.
- [ ] Verify retry, concurrent workspace preservation, and operator/agent documentation.

## Log

### 2026-09-22 — fresh v2 task

Created from the agreed workspace/UI contract and the request for a clean task breakdown. Implementation has not started; estimates follow design approval.
