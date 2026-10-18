---
id: 000451
status: open
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
github_issue:
---

# test economy: CI failures in the daily flake ledger (follow-up of #394)

## Problem

#394's `sdlc flake list` shows the daily flake view from local suite attempts only; it prints CI as `not covered`. #390 §3 and #394's Done-when 3 want CI failures in the TL's daily view too, kept as their own classes (branch regression, infrastructure/fixture fault, confirmed same-head flake), never merged into the local flake rate.
