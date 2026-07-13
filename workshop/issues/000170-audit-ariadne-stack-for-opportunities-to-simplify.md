---
id: 000170
status: open
deps: []
github_issue:
created: 2026-07-13
updated: 2026-07-13
estimate_hours:
---

# audit ariadne stack for opportunities to simplify

I have been using ariadne based system to create several software (ariadne itself, nous and brain - the personal assistant (just started as information proxy for agents), metis based ML workbench, pair - the agent neutral development frontend, parley a harness in nvim, you-decide - voter information system, etc.). So far, it works well for those tasks, on both codex and claude code. I can freely switch between them yet using same development flow. 

On the other hand, `sdlc process-manual` shows 61 markdown artifacts in play and I know there are "introspect" based additional files, and some binary code (`sdlc`, `weave`, etc.) that form the spine. It is to say there are some complexities. Ariadne's both organically grew, based on my need, but also have some guiding principles. 

I suspect it is a good time to take a look at the current workflow, defined by the combination of those 61 markdown files, introspect knowledge files, the two binaries `sdlc`, `weave`, and likely some (probably minor) instruction files I missed. We should take a holistic audit of it, and then simplify. 

The starting point I suspect, is the following:

1. run one more introspect, which is ticket #169
2. use brain's data/project/metis-v2-experiment-algebra.md as the project, check its full history from git commits, and agents (mostly claude I think, but also check codex), to answer the following questions.
   1. the timeline of interactions and work, by main agent, or subagent. 
   2. are there slow segment that we may speed things up.
   3. are there repeatedly loaded context that we can avoid. 
   4. are there opportunity to make ariadne agent instructions more concise/precise.
   5. the size of lessons file, how fast are they growing? should we have compaction algorithm, to periodically synthesize them to control its overall size.
   6. is the introspect distilled knowledge useful. 
   7. any key mechanism I have created that I missed above?
   8. more things you think would help?

## Problem

## Spec

## Done when

-

## Plan

- [ ]

## Log

### 2026-07-13
