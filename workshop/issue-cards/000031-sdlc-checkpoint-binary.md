---
id: '000031'
status: done
created: 2026-05-25
updated: 2026-05-26
actual_hours: 18.8
---

# sdlc checkpoint binary

## Problem

Two related problems, captured in `docs/vision/2026-05-25-01-pensive-sdlc-checkpoint-binary.md`:

1. **Progressive-disclosure leak.** Markdown-index skills lose context-budget control when agents glob `**/*.md`. The disclosure boundary is conventional, not enforced.
2. **Workflow drift defense lives in Make.** Ariadne's checkpoint guards (close-issue, fetch, worktree, push, pull-request, merge, issue-sync, check-*) live as Makefile targets with weak arg parsing, env-export shims, no embedded help, and no shared discoverability surface.

The lift: a single Go binary `sdlc` that collects existing checkpoint guards as subcommands, ships its own guidance via `sdlc --help` / `sdlc <verb> --help`, exposes workflow state via `sdlc state`, and names the anti-collusion judge pattern as a first-class primitive (`sdlc judge ...`).
