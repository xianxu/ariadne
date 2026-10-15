---
id: 000387
status: open
created: 2026-10-15
updated: 2026-10-15
estimate_hours:
github_issue:
---

# Agent: trailer says 'human' (not 'sdlc') for a commit made outside an agent session

## Problem

ariadne#367's `Agent:` trailer says `Agent: sdlc` when `sdlc` runs in a plain shell with no agent session, i.e. when
the operator makes the commit by hand. Velocity and usage numbers then credit the operator's hand work to "sdlc". The
operator (2026-10-15): "say human in that case".
