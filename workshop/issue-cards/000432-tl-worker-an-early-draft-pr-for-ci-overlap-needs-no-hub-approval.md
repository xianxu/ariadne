---
id: 000432
status: open
created: 2026-10-17
updated: 2026-10-17
estimate_hours:
github_issue:
---

# tl-worker: an early draft PR for CI overlap needs no hub approval

## Problem

Slots ask the hub for permission before opening an early draft PR so GitHub CI overlaps a queued local suite (pair:3 asked twice on 2026-10-17). The hub always says yes, because the suite, close review and merge gates still run before landing. Each ask costs a hub round trip while the suite queue is long.
