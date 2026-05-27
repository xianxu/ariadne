---
id: 000039
status: open
deps: []
created: 2026-05-27
updated: 2026-05-27
---

# sdlc now always create a worktree

this slows down development for a single operator mode. single human has a single threaded mind, so easier to just operate on master. let's update this to be something selectable via some configuration. 

actually, extend the sdlc binary itself with an `sdlc config`, then for now a flat list of configuration values. for example. 

sdlc.start.worktree: yes | no | ask | auto

auto option will defer to the agent to pass in explicit flag, or to prompt it to decide if the issue is sufficiently complex to warrant a worktree. 

ask option will generate an AskUserQuestion tool call to ask user if a worktree is needed. 

## Done when

-

## Spec


## Plan

- [ ]

## Log

### 2026-05-27

