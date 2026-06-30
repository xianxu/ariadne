---
id: 000146
status: open
deps: []
created: 2026-06-30
updated: 2026-06-30
---

# sdlc milestone review bypass command too ambiguous

agent used `sdlc close --milestone Mx`, instead of `sdlc milestone-close`. `sdlc close --milestone Mx` looks too innocent. there should be some text like "force" etc. to indicate it is skipping normal path. 

let's start with checking what does th `sdlc close` verb do, and then systematically update escape patches to have word force in it, e.g. `sdlc force-close --milestone Mx`. I'm not sure about the actual contour of the verbs, let's do some investigation first. 

## Done when

-

## Spec


## Plan

- [ ]

## Log

### 2026-06-30

