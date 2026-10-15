---
id: 000401
status: open
created: 2026-10-15
updated: 2026-10-15
estimate_hours:
github_issue:
---

# tl-lead/tl-worker: hot-path plans state their expected cost per driver and the counted test that defends it

## Problem

Resource regressions (pair#579's 61× writes, pair#584's 15 s decode) passed plan review because nobody asked what an
operation should cost. Ops project resource-observability decided (10-15) that hot-path changes state their expected
cost and count it.
