---
id: 000183
status: open
deps: []
github_issue:
created: 2026-07-16
updated: 2026-07-16
estimate_hours:
---

# FIX-THEN-SHIP needs better process

right now, it seems when sdlc instruct agent to FIX-THEN-SHIP for non-critical things, after the fix, agent has to run milestone-close with --no-judge. this leaves audit trail that milestone-close is bypass, which is not strictly the case. I'm thinking about the following scheme, to improve the process. 

1. sdlc milestone-close (or other gate) will generate a hash based on the files it think needs to change. it keep track of that in its state. let's say one content hash per file. 

2. when the agent calls again with milestone-close --fixed-to-ship, the sdlc gate will behave differently. it would not launch the configured gates in milestone-close. rather, it will check if the files it thought need to be changed, has changed hash. we can probably just check if 75%+ files changed, if yes, let it pass. make the 75% a configurable knob, which controls which agent (coding agent, or review agent) we trust more.

## Problem

## Spec

## Done when

-

## Plan

- [ ]

## Log

### 2026-07-16
