---
id: 000420
status: codecomplete
created: 2026-10-16
updated: 2026-10-17
estimate_hours: 4.92
github_issue:
severity: medium
started: 2026-10-17T00:53:13-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: N/A
tracker:
    version: 1
    completion:
        token: close-f79ace7f3dd1
        repository: github.com/xianxu/ariadne
        reviewed_head: 61af4a691cabb809460f5a08bc44f9e3ccb2812f
        evidence_commit: ce306d97ec959fe78cfa4ae487c3902fa804911e
---

# Actual: queued suite status polling bridges idle time into focused-hours samples

## Problem

While closing #416, automatic actual adoption reported 2.40h for the window 2026-10-16 19:36:21–21:58:36 America/Los_Angeles, with a dominant 142.1m/99% attribution segment. Most of that window was waiting for machine test slots, with assistant status updates and write_stdin polls about every 45–60 seconds to obey the communication rule.

The recorded suite attempts retain exact queue evidence under the shared git directory's sdlc/test-stats/attempts/:
- d5ecb880e188a182590c4815cf749537.json: queue_seconds=3055.535157833; wall_seconds=532.971970042; suite SHA 005587765f51a3baf610fd3f70b1bed72f0b1a65.
- 0bbae3af1b6112bb7cb13acbc505a1ca.json: queue_seconds=3895.523716167; wall_seconds=635.788244625; suite SHA 1b31e6fc1f288e278b1f1c69fc921726c03cd142.

Together these are 115.85 minutes queued and 19.48 minutes executing suites inside a roughly 142-minute close window. Some real work (the staged live check) occurred concurrently with the first queue wait, so blindly subtracting whole queue windows would also be wrong. No replacement hours were typed; the tool's adopted value was retained pending a calibration-policy decision.
