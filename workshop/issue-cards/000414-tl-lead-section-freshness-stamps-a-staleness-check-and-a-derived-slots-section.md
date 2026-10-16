---
id: 000414
status: open
created: 2026-10-16
updated: 2026-10-16
estimate_hours:
github_issue:
---

# tl-lead: section freshness stamps, a staleness check, and a derived Slots section

## Problem

The TL session file (`teams/<team>/tl-status-<operator>.md` in ops, written only through `tl-lead status`) goes stale:
on 2026-10-16 its Threads section was ~20 hours old while Needs you stayed current. Nothing tells the TL a section is
stale: updating Threads is a synthesis no single event asks for, and writes blocked by the operator's lock get queued
and forgotten. Operator, 10-16: move this into the hard shell; long-term, only Operator instructions (chat) and Needs
you need to be prose in the doc; the rest becomes command output, a "flexible cockpit", some of it eventually a pure
operations dashboard.
