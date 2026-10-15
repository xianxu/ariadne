---
id: 000394
status: open
created: 2026-10-15
updated: 2026-10-15
estimate_hours:
github_issue:
---

# Issue-backed flake landing and daily triage

## Problem

An intermittent suite failure currently either blocks a landing or invites an unexplained retry. Both hide the actual test-health cost. #390's approved rule requires one passing exact-head suite while letting an already identified, separately owned flake be diagnosed and rerun on the same head.
