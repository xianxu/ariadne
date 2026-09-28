---
id: '000225'
status: done
started: 2026-09-13T12:29:20-07:00
created: 2026-09-13
updated: 2026-09-13
estimate_hours: 1.03
actual_hours: 0.10
---

# Publish a portable root Makefile as a safe seed

## Problem

parley.nvim#208 needs a standalone contributor checkout with a real root
Makefile. The inherited `symlink Makefile` declaration replaces a portable
root on the next weave. A leaf seed cannot counteract it safely: current
`applySeed` follows destination symlinks, potentially overwriting an ancestor.
Fresh clones also lack ignored Makefile.workflow, so bootstrap's handoff to
`make bootstrap` cannot work merely by making that include optional.
