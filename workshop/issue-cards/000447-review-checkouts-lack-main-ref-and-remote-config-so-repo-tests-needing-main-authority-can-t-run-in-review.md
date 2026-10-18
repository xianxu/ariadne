---
id: 000447
status: open
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
github_issue:
---

# review checkouts lack main ref and remote config, so repo tests needing main authority can't run in review

## Problem

The boundary review runs in a disposable checkout that lacks the main branch ref, its upstream and remote configuration. Repository tests that need main authority (pair's Go lint catalog finds tracked files and the merge base against main) correctly refuse there, so the reviewer cannot run them and raises an "unverifiable" finding. It hit pair#665 M1 (BR-4) and pair#666 (BR-1) on 2026-10-18; each slot had to capture host evidence and dispute it.
