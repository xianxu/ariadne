---
id: 000420
status: open
deps: []
github_issue:
created: 2026-10-16
updated: 2026-10-16
estimate_hours:
severity: medium
card_mirror: '74a82fbea998203f3b0afc860f66c8301b20e894' # card fields mirrored from issue-cards; edit via sdlc
---

# Actual: queued suite status polling bridges idle time into focused-hours samples

## Problem

While closing #416, automatic actual adoption reported 2.40h for the window 2026-10-16 19:36:21–21:58:36 America/Los_Angeles, with a dominant 142.1m/99% attribution segment. Most of that window was waiting for machine test slots, with assistant status updates and write_stdin polls about every 45–60 seconds to obey the communication rule.

The recorded suite attempts retain exact queue evidence under the shared git directory's sdlc/test-stats/attempts/:
- d5ecb880e188a182590c4815cf749537.json: queue_seconds=3055.535157833; wall_seconds=532.971970042; suite SHA 005587765f51a3baf610fd3f70b1bed72f0b1a65.
- 0bbae3af1b6112bb7cb13acbc505a1ca.json: queue_seconds=3895.523716167; wall_seconds=635.788244625; suite SHA 1b31e6fc1f288e278b1f1c69fc921726c03cd142.

Together these are 115.85 minutes queued and 19.48 minutes executing suites inside a roughly 142-minute close window. Some real work (the staged live check) occurred concurrently with the first queue wait, so blindly subtracting whole queue windows would also be wrong. No replacement hours were typed; the tool's adopted value was retained pending a calibration-policy decision.

## Spec

Investigate how native task spans and frequent assistant/tool polling turn queue idle time into focused dev-hours. Make the metric's treatment of queue waits explicit, preserve genuine overlapping work, and keep raw elapsed/queue evidence separate from focused effort. Repair or exclude affected samples through the existing generation-bound correct-actual path with honest provenance, never guessed numeric replacements.

## Done when

- A deterministic native-transcript regression reproduces a long queued suite with periodic tool polls/status updates and meaningful work overlapping part of the wait.
- The measurement distinguishes idle queue time from genuine overlapping work according to the documented metric, with retained evidence that supports replay and correction.
- #416's adopted actual is evaluated by the TL/operator and corrected or explicitly excluded if the sample is invalid; no hand-typed hour estimate is used.

## Plan

- [ ]

## Log

- Found while closing #416; independent of its queue-helper lifecycle fix. The close prints active-time-v3-engagement-v1 and 2 admitted sessions, with the complete attribution segment covering the wait-heavy window.
