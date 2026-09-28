---
id: '000141'
status: done
started: 2026-07-01T09:49:03-07:00
created: 2026-06-29
updated: 2026-07-01
estimate_hours: 0.42
actual_hours: 0.19
---

# sdlc merge confirmation should support agent-safe defaults

## Problem

`sdlc merge` asks for final confirmation before irreversible actions:

- server-side GitHub PR merge;
- remote branch deletion;
- switching/pulling the local checkout;
- archiving completed issues to history;
- deleting the local feature branch or worktree.

That confirmation is sensible for humans in an interactive terminal. In an
agent/non-interactive run, however, the prompt defaults to "no". In pair#84,
`sdlc merge` ran the expensive pre-merge judges successfully and then aborted at
the final prompt because no interactive answer could be supplied. The operator
had to rerun with `--yes`.

The problem is not the confirmation itself. The problem is that non-interactive
contexts discover the need for `--yes` only after spending time on slow judges.
