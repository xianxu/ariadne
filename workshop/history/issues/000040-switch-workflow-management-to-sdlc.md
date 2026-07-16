---
id: 000040
status: done
deps: [000031, 000039]
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

## Spec

- AGENTS.md should tell agents to invoke `sdlc` directly for workflow
  checkpoints:
  - claim work with `sdlc set-status --issue N working` plus
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

- [x] Update AGENTS.md workflow instructions from Makefile-era commands to
      direct `sdlc` commands.
- [x] Update `cmd/sdlc/helptext/root.md` and `cmd/sdlc/helptext/index.md`
      so generated help/skill prose lists `claim` and `change-code`.
- [x] Regenerate `construct/local/sdlc/SKILL.md` from `sdlc --index`.
- [x] Update workflow atlas docs that still describe Makefile as canonical.
- [x] Run focused tests for sdlc helptext/embedding and a full Go test pass if
      feasible.

## Log


- 2026-05-28: closed — Updated ariadne workflow docs/help to make direct sdlc invocation canonical; regenerated construct/local/sdlc/SKILL.md; verified with go run ./cmd/sdlc --help, go test ./cmd/sdlc/..., and go test ./...
### 2026-05-28 — issue created

Created after downstream `pair` inspection showed SDLC migration drift:
ariadne owns the SDLC binary and generated skill source, so the canonical
fix belongs here.

### 2026-05-28 — base docs and helptext updated

Updated AGENTS.md, SDLC root/index helptext, workflow atlas pages, and the
xx-issues local skill so direct `sdlc` invocation is the canonical agent path.
Makefile workflow targets are now described as compatibility wrappers.

### 2026-05-28 — generated skill refreshed

Ran `go run ./cmd/sdlc --index > construct/local/sdlc/SKILL.md` so the local
SDLC skill reflects the corrected `claim` / `change-code` command table.

### 2026-05-28 — verification

Verified updated SDLC help from source with `go run ./cmd/sdlc --help`.
Ran `go test ./cmd/sdlc/...` and `go test ./...`; both passed.
