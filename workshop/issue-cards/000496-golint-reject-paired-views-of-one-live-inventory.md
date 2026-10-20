---
id: 000496
status: open
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
github_issue:
---

# golint: reject paired views of one live inventory

## Problem

pair#711 reproduced a fleet messaging failure with two individually valid native inventories: Zellij --short observed a selected session; --no-formatting observed the same selected session plus an unrelated new one. Comparing their name sets refused the healthy sender/recipient. The direct collector called both sequentially; the poller called both concurrently. One validated full reply already contains the needed names and states.
