---
id: 000463
status: open
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
github_issue:
---

# local-check: resume recorded known-flake failures with targeted tests

## Problem

ariadne#461's recorded suite attempt fa09ccf7eb6fd0862fadf1d94e865944 failed
only TestSuiteQueuePriorityRealHelper when its private provider launch returned
Operation not permitted. The exact same commit passed that targeted test without
changes, and the test was registered on ariadne#462. Existing local-check run
can invoke the known-flake step only while processing a fresh full suite;
republish reposts an already-saved failure unchanged. There is no supported
way to finish this failed attempt with just the now-registered targeted tests,
so the hub approved another full suite for #461. This is a #394 follow-up.
