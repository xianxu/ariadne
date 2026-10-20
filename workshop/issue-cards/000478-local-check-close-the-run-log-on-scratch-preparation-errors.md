---
id: 000478
status: open
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
github_issue:
severity: low
---

# local-check: close the run log on scratch preparation errors

## Problem

The third close review of ariadne#477 recorded Minor BR-3: runUnderSlot opens logf before prepareScratch, and its setup-error return bypasses the explicit Close later in the function. CLI exit eventually closes the descriptor, but the function should release its own resource on every exit.
