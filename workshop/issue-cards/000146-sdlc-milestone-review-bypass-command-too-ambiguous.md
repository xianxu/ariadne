---
id: '000146'
status: done
started: 2026-07-05T16:28:52-07:00
created: 2026-06-30
updated: 2026-07-05
estimate_hours: 0.53
actual_hours: 1.35
---

# sdlc milestone review bypass command too ambiguous

## Problem

agent used `sdlc close --milestone Mx`, instead of `sdlc milestone-close`. `sdlc close --milestone Mx` looks too innocent. there should be some text like "force" etc. to indicate it is skipping normal path. 

let's start with checking what does th `sdlc close` verb do, and then systematically update escape patches to have word force in it, e.g. `sdlc force-close --milestone Mx`. I'm not sure about the actual contour of the verbs, let's do some investigation first.
