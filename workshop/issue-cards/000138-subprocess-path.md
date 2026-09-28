---
id: '000138'
status: done
started: 2026-06-29T22:09:01-07:00
created: 2026-06-26
updated: 2026-06-29
estimate_hours: 0.27
actual_hours: 0.49
---

# sdlc subprocess path resolution

## Problem

Fresh agent subprocesses invoked by `sdlc` can fail to find `sdlc` on `PATH`.
During pair#81 retro, the main Codex shell had to discover
`/Users/xianxu/workspace/ariadne/bin/sdlc`, and fresh review subprocesses that
started from a narrower shell environment attempted `sdlc --help` and hit
`command not found`.

User shell configuration can work around this locally, but SDLC-spawned agents
should not depend on every harness loading the same interactive zsh startup
files.
