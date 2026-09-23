---
id: 000246
status: working
deps: [ariadne#242, ariadne#243, ariadne#244, ariadne#245]
github_issue:
created: 2026-09-22
updated: 2026-09-23
estimate_hours:
started: 2026-09-23T15:24:19-07:00
---

# Slots v2: land while retaining the workspace

## Problem

Current SDLC merge cleanup removes feature worktrees and returns in-place work to updated main. A durable slot must survive publication and return to its unchanged resting branch.

## Spec

Project: `pair/workshop/projects/couch-slots-v2.md`. Fresh task derived from the current v2 contract; historical task bodies are not prerequisites or implementation plans.

Adapt existing publication/review/merge flow for durable slots. Confirm integration into the configured remote main, safely return to the existing main-slotN, then safely remove the completed local issue branch. Do not refresh the resting branch. Preserve directory, unpublished commits, dirty files, and other workspaces. Do not operate on the primary checkout merely to complete a slot’s cleanup or archive its artifacts.

Account for squash/rebase merge evidence when deciding that branch deletion is safe; an upstream relationship or branch name is insufficient proof. Detect commits added after the integrated head. Define recoverable phases for integration completed but switch/deletion/archive cleanup interrupted. Re-running must recognize already-confirmed integration and finish safely without duplicating publication. Keep ordinary feature-worktree and :0 behavior explicit and tested. ARCH-FUNERAL: retain reusable workspaces intentionally, while completed issue branches have an evidence-backed removal path.

### Agreed scope — 2026-09-23

This section takes precedence over earlier conflicting layout or policy text.

The durable main worktree is `/workspace/worktree/<repo>-slotN/<repo>`; its parent contains ordinary dependency clones. Landing the main repository must preserve that enclosing directory and every dependency checkout, including dirty files and unpublished commits. Dependency repositories continue to use their normal per-repository SDLC publication/cleanup behavior; do not classify them as numbered worktrees or recursively delete them. Coordinated work is driven from the parent thread and ships each repository separately in dependency order when required. This version adds no automatic compare-and-merge of all dependencies and no multi-repository publication transaction.

## Done when

- A slot completes review/publication and lands; its path survives and its resting SHA/upstream are unchanged.
- Primary and another dirty active slot retain their branch, files, and refs throughout landing.
- Unpublished additions, dirty switch collisions, and uncertain integration evidence block destructive cleanup and explain recovery.
- Integration followed by interrupted cleanup can be retried safely, including the repository’s actual merge strategy.
- Issue archiving/project updates complete through existing gates without requiring a clean idle primary; ordinary worktree behavior has regression coverage.

- Landing and interrupted-cleanup retry preserve the enclosing environment and dependency clones, including unpublished issue records and code changes.
- A dependency change can be published first through its own normal SDLC flow, then the parent feature can land without recursively publishing or cleaning other repositories.

## Plan

Task outline only; settle implementation design through start-plan before change-code.

- [ ] Trace current merge/publication/archive phases and specify durable-slot recovery.
- [ ] Add integration fixtures, then adapt landing and safe cleanup.
- [ ] Verify retry, concurrent workspace preservation, and operator/agent documentation.

## Log

### 2026-09-22 — fresh v2 task

Created from the agreed workspace/UI contract and the request for a clean task breakdown. Implementation has not started; estimates follow design approval.

## Revisions

### 2026-09-23 — Retain nested environments and publish repositories separately

Reason: operator agreed nested environments, ordinary remote dependency clones and existing per-repository publication. Delta: added the authoritative scope clarification and acceptance criteria above; original task context remains as provenance. Added #243 as a prerequisite for the nested identity contract. No implementation or lifecycle-status change is claimed by this revision.
