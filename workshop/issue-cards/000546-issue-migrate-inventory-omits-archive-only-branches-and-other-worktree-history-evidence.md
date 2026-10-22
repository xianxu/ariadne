---
id: 000546
status: open
created: 2026-10-21
updated: 2026-10-21
estimate_hours:
github_issue:
---

# issue migrate inventory omits archive-only branches and other-worktree history evidence

## Problem

Pre-existing issue migrate inventory omits archived issue evidence outside pinned main. Discovered by #543 close review round 2 (BR-2, authoritative-inventory-completeness); TL explicitly moved importer completeness out of new-repo init's scope.

Reproductions against real Git/bare origin: (1) clone B creates and pushes a branch containing workshop/history/issues/000001-old.md while main has no issues; clone A migrationInventory scans branch edits only under workshop/issues and plans an empty tracker. (2) another linked live worktree contains uncommitted workshop/history/issues/000001-old.md; dirtyIssuePaths scans only workshop/issues, again allowing an empty plan. The same roots/scope rule must examine flat legacy history and detached/inherited branch/worktree evidence, rather than patching only these two sites.
