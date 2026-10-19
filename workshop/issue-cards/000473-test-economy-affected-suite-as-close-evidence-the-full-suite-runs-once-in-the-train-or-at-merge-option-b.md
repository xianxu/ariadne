---
id: 000473
status: open
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
github_issue:
---

# test-economy: affected-suite as close evidence; the full suite runs once, in the train or at merge (option B')

## Problem

A change pays the full suite twice today: once on its branch (close evidence), and again at landing when main has moved
with code (the integrated suite), or in a train (the train suite). Measured 10-19 (`sdlc test-stats`, 7 days): pair suite
12.5 min median, slot wait 11 min median / 35 min p90, ready → landed 54 min median / 2.3 h p90; ariadne's suite takes both
slots. The machine is suite-bound. Discussion and the operator's decision: ops side doc
`teams/devinfra/tl-status-xian-xu.side/2026-10-19-11-landing-schedule.md` (option B′, approved 10-19).
