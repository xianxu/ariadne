---
id: 000452
status: open
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
github_issue:
---

# suite queue priority: a fleet-blocking landing grant also jumps the machine suite queue

## Problem

A fleet-blocking landing grant (sdlc landing grant) orders the merge line but not the machine-wide suite queue (WF_TEST_SUITES, two slots). On 2026-10-18 the urgent pair#672 and pair#676 each waited behind ordinary suites for a machine slot even with the grant; the weekly snapshot shows a median queue wait of 11 to 20 minutes and a 90th percentile over 30 minutes.
