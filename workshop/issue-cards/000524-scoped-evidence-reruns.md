---
id: 000524
status: open
created: 2026-10-21
updated: 2026-10-21
estimate_hours:
github_issue:
---

# Scoped future evidence and exact-package flake reruns

## Problem

Prerequisite for ariadne#463, approved by hub receipt917bf022-032a-4d84-ac6b-39075bbca2b5. Today's attempt reader treats an unsupported record as a global warning; guardSuitePass then refuses unrelated successes and saved-pass loading marks the whole store incomplete. A future ResumeLink cannot retrofit stale binaries. Existing known-flake reruns also use one name-pattern cross-product across selected packages.
