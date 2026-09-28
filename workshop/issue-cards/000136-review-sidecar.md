---
id: '000136'
status: done
started: 2026-06-29T15:42:51-07:00
created: 2026-06-26
updated: 2026-06-29
estimate_hours: 0.59
actual_hours: 0.48
---

# sdlc boundary review sidecar

## Problem

`sdlc` boundary reviews can produce more output than an interactive agent can
reliably keep in scrollback. During pair#81 retro, pair#72 boundary-review
output was effectively a transient terminal artifact: the close/milestone gate
could make a decision, but the agent did not have a stable file to reopen for
details after truncation or context compaction.

Boundary reviews are workflow evidence. They should be persisted as first-class
sidecar artifacts alongside the issue/plan record instead of existing only as
TTY output and commit trailers.
