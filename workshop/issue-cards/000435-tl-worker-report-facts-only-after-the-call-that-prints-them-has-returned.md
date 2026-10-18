---
id: 000435
status: open
created: 2026-10-17
updated: 2026-10-17
estimate_hours:
github_issue:
---

# tl-worker: report facts only after the call that prints them has returned

## Problem

On 2026-10-17 parley.nvim:4 reported a landing with a merge SHA that never existed. It sent the report in the same parallel tool batch as the git call that prints the SHA, so the report was written before the output existed and the SHA was typed, not read.
