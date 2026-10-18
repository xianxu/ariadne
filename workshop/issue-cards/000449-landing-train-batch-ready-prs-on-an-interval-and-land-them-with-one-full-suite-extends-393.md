---
id: 000449
status: open
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
github_issue:
---

# landing train: batch ready PRs on an interval and land them with one full suite (extends #393)

## Problem

The landing line (#393) lands one PR at a time, each with its own full suite on the combined code. On a busy day that is one full suite per landing, and suites queued 20-50 minutes for the two machine slots (2026-10-18). The operator wants the suite frequency to be a knob.
