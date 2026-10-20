---
id: 000488
status: open
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
github_issue:
severity: medium
---

# Duplicate-ID guard counts numbered .side directories as issue records

## Problem

PR338's merge-check failed in scripts/merge-checks.d/40-duplicate-issue-id.sh (GitHub run 39388400345) after adding a legitimate numbered issue side folder. The checker reported the issue .md and its matching .side directory as two claimants of #487. The .side contained only a measured sandbox transcript, not a second issue record. The local-gates guard passed. #424 explicitly supports document side folders and archives them with the owning document.

The CI diagnostic named workshop/issues/000487-make-testscratchfallbackoverridesinheritedtemp-independent-of-host-sandbox-temp-permissions.md and the same stem ending .side as separate issue paths. #487 avoids the bug by relocating the unchanged transcript into its numbered workshop/plans artifact family; it does not bypass or modify the guard.
