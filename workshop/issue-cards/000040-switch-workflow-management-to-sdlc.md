---
id: '000040'
status: done
created: 2026-05-28
updated: 2026-05-28
estimate_hours: 1.5
actual_hours: 1.0
---

# Switch workflow management instructions to sdlc

## Problem

Ariadne's SDLC binary is now the canonical workflow checkpoint surface,
but several base-layer instructions still point agents at the old Makefile
workflow (`make issue-sync`, `make close-issue`, `make push`, `make worktree`)
or expose stale `sdlc start` / `sdlc lock` helptext.

This creates two forms of drift:

- Agents read AGENTS.md and still invoke Makefile-era workflow commands.
- The generated `construct/local/sdlc/SKILL.md` carries stale subcommand
  rows even though the live Cobra registry has moved to `claim` and
  `change-code`.
