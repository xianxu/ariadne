---
id: '000244'
status: done
started: 2026-09-23T11:35:53-07:00
created: 2026-09-22
updated: 2026-09-23
estimate_hours: 17.01
actual_hours: 6.38
---

# Slots v2: concurrent issue workflows

## Problem

Two slots running SDLC concurrently must not claim the same issue silently, overwrite newer issue records, or block unrelated work throughout a long review.
