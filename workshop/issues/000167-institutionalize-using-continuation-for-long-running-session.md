---
id: 000167
status: working
deps: []
github_issue:
created: 2026-07-07
updated: 2026-07-12
estimate_hours:
started: 2026-07-12T22:02:15-07:00
---

# institutionalize using continuation for long running session

in ariadne, we created agent neutral way to compact, called continuation. agent should be told to leverage that when context is > 60% full. this way long and complex session can be broken down to smaller chunks and when we are close to context window limit, create a hand off file automatically and restart the session.

In the metis-v2 project, agent already learned to do so after I did that once. then agent starts to generate continuation automatically, which is great. 

## Problem

## Spec

## Done when

-

## Plan

- [ ]

## Log

### 2026-07-07
