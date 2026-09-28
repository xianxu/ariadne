---
id: '000166'
status: done
started: 2026-07-07T16:43:47-07:00
created: 2026-07-07
updated: 2026-07-07
estimate_hours: 2.05
actual_hours: 1.04
---

# sdlc git lock is too long

## Problem

When `sdlc` runs a long review action, it holds `.git/sdlc.lock` for the entire
duration. That blocks unrelated `sdlc` commands that need git state, even while
the long-running step is waiting on external review/model work rather than
mutating local repo state.
