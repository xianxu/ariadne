---
id: 000131
status: working
deps: []
created: 2026-06-25
updated: 2026-06-25
started: 2026-06-25T16:48:25-07:00
---

# create a model to estimate context window per agent based on scrollback line count

this provides a good signal how we want to organize a session, including guiding user to alt+shift+n for new session. 

and we can display this inferred percentage behind the agent string in the agent pane frame

claude (80%) [cwd]

on tricky thing though is an agent may have many different models with different context window size... so this may not work well. in practice, typically one agent family would have similar context window size, at least in the context of a coding agent. 

## Done when

-

## Spec


## Plan

- [ ]

## Log

### 2026-06-25

