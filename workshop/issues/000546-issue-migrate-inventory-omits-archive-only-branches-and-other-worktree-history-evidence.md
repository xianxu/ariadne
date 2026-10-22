---
id: 000546
status: open
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-21
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'dc5fc83117cb00a722d670df63678d2111b0850c' # card fields mirrored from issue-cards; edit via sdlc
---

# issue migrate inventory omits archive-only branches and other-worktree history evidence

## Problem

Pre-existing issue migrate inventory omits archived issue evidence outside pinned main. Discovered by #543 close review round 2 (BR-2, authoritative-inventory-completeness); TL explicitly moved importer completeness out of new-repo init's scope.

Reproductions against real Git/bare origin: (1) clone B creates and pushes a branch containing workshop/history/issues/000001-old.md while main has no issues; clone A migrationInventory scans branch edits only under workshop/issues and plans an empty tracker. (2) another linked live worktree contains uncommitted workshop/history/issues/000001-old.md; dirtyIssuePaths scans only workshop/issues, again allowing an empty plan. The same roots/scope rule must examine flat legacy history and detached/inherited branch/worktree evidence, rather than patching only these two sites.

## Spec

Define and review migrate's complete inventory/import/refusal rule across active, flat legacy history and history/issues, pinned main, local/current remote branches and live worktrees. Keep this distinct from issue init's conservative presence-only gate (#543): init imports nothing and refuses any issue evidence anywhere. #543's BR-1 already refreshes/pins remote head evidence; this issue owns importer gaps beyond that correction.

## Done when

- Real-Git regressions reproduce the remote archive-only branch and other-worktree dirty archive cases and prevent silent empty migration.
- A reviewed root-by-scope inventory rule states what is imported, what refuses, and how unknown observations fail closed.
- Existing active branch rename/archive and provable legacy-close migration behavior remains verified.

## Plan

- [ ]

## Log

Filed unclaimed per TL direction, Couch 5119a959-0128-4420-89ee-9a91fb0b08d9. Review reproduction was against #543 pinned head 96e8ff1cfaf539a80c5c174ce4ba7bbf83992bae; the omission predates #543 in issuemigrate.go (branch/worktree scopes use d.Home only). No importer expansion authorized in #543.
