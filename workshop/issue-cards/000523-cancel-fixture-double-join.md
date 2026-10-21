---
id: 000523
status: open
created: 2026-10-21
updated: 2026-10-21
estimate_hours:
github_issue:
---

# Native cancellation fixture waits four minutes after successful joins

## Problem

TestLocalCheckCancelOnlyOwnRun spends about four minutes waiting in successful cleanup after both cancellation assertions already joined their native children. This inflates the machine-wide suite wall time without exercising more behavior.
