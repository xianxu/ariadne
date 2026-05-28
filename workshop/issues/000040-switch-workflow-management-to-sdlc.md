---
id: 000040
status: working
deps: [000031, 000039]
created: 2026-05-28
updated: 2026-05-28
estimate_hours: 1.5
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

## Spec

- AGENTS.md should tell agents to invoke `sdlc` directly for workflow
  checkpoints:
  - claim work with `sdlc set-status --issue N --status working` plus
    `sdlc claim --issue N`
  - close work with `sdlc close`
  - ship with `sdlc push`, `sdlc pr`, and `sdlc merge`
- sdlc top-level help and generated skill prose should list the live verbs:
  `claim` and `change-code`, not stale `start` and `lock`.
- Workflow atlas docs should frame Makefile targets as compatibility wrappers,
  not the canonical agent path.
- The regenerated `construct/local/sdlc/SKILL.md` should match the binary's
  updated embedded helptext.

## Plan

- [ ] Update AGENTS.md workflow instructions from Makefile-era commands to
      direct `sdlc` commands.
- [ ] Update `cmd/sdlc/helptext/root.md` and `cmd/sdlc/helptext/index.md`
      so generated help/skill prose lists `claim` and `change-code`.
- [ ] Regenerate `construct/local/sdlc/SKILL.md` from `sdlc --index`.
- [ ] Update workflow atlas docs that still describe Makefile as canonical.
- [ ] Run focused tests for sdlc helptext/embedding and a full Go test pass if
      feasible.

## Log

### 2026-05-28 — issue created

Created after downstream `pair` inspection showed SDLC migration drift:
ariadne owns the SDLC binary and generated skill source, so the canonical
fix belongs here.
