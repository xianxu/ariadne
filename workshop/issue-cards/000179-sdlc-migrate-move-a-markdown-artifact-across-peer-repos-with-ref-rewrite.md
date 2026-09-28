---
id: '000179'
status: done
started: 2026-07-15T11:49:25-07:00
created: 2026-07-15
updated: 2026-07-15
estimate_hours: 1.16
actual_hours: 0.24
---

# sdlc migrate: move a markdown artifact across peer repos with ref rewrite

## Problem

#171's direction (peer-repo `repo#id` addressing; artifact residency = soft
center-of-gravity default) makes moving a markdown artifact between repos a
NORMAL operation — a project file follows the work, an SDLC artifact leaves
brain. But a move today is a hand job with a silent correctness trap: **bare
`#NNN` refs inside the file are repo-relative.** Moved verbatim, they
re-resolve against the destination repo's issue numbering — pointing at
unrelated issues without any error. The rewrite rules are fixed patterns (the
formal ref grammar `sdlc resolve` already owns), so this belongs in the
binary, not in agent judgment.
