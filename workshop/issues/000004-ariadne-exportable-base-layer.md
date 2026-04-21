---
id: 000004
status: done
deps: []
created: 2026-04-20
updated: 2026-04-20
---

# Ariadne Exportable Base Layer

## Problem

Ariadne has two identities — the workshop (where the system evolves) and the export (portable fragments other repos consume). These need clear boundaries, a manifest, and a setup mechanism.

## Spec

Disentangle ariadne-the-workshop from ariadne-the-export. Produce a base layer that can bootstrap new repos with the "ariadne way of working."

### What's portable (base layer)

- `AGENTS.md` — constitution, workflow orchestration
- `CLAUDE.md` — entry point referencing AGENTS.md
- `.claude/settings.ariadne.json` — base settings with merge semantics
- `.claude/skills/` — superpowers skills, `/fix`, `/construct`, `construct/local/` skills
- `Makefile.ariadne` — workflow targets (`make issue`, `make worktree`, `make push`, `make merge`)
- `scripts/ariadne/` — scripts supporting the Makefile system
- `workshop/` — directory structure convention (issues, plans, history, vision, staging)
- `atlas/workflow/` — documentation of the base workflow system
- `setup.sh` — bootstrapper that wires fragments into target repos

### What stays local (workshop only)

- `docs/vision/` content
- `workshop/` content (issues, plans, history)
- `atlas/` content beyond workflow basics
- Personal skills not in construct/local/
- `data/`, repo-specific scripts

### Settings merge convention (from parley pattern)

- **Merge keys** (additive): `allowedTools`, `permissions`, `hooks`
- **Replace keys** (local wins): `model`, feature flags, scalar values
- Local file: `settings.local.json` overlays on `settings.ariadne.json`

### Boundary enforcement

- `base.manifest` — lists all portable paths, source of truth
- `setup.sh` reads manifest to know what to symlink/scaffold
- AGENTS.md rule: "base-layer changes require considering downstream impact"

### Consuming repo structure

```
target-repo/
  AGENTS.md -> ../ariadne/AGENTS.md   (symlink)
  AGENTS.local.md                      (repo-specific extensions)
  CLAUDE.md -> ../ariadne/CLAUDE.md    (symlink)
  Makefile                             (includes ../ariadne/Makefile.ariadne)
  .claude/settings.ariadne.json -> ... (symlink)
  .claude/settings.local.json          (repo overrides)
  .claude/skills/ -> ...               (symlinks to ariadne skills)
  skills.local/                        (repo-only skills)
  workshop/
  atlas/
```

## Plan

1. [x] Create `base.manifest` listing all portable paths
2. [ ] Restructure scripts into `scripts/ariadne/` namespace — DEFERRED: scripts already work via symlink, namespace change is cosmetic
3. [x] Create `settings.ariadne.json` with merge logic in setup.sh
4. [ ] Rename/restructure Makefile into includable `Makefile.ariadne` — DEFERRED: `Makefile.workflow` already works as includable, no rename needed
5. [x] Write `atlas/workflow/` documenting the AGENTS.md system
6. [x] Write `setup.sh` that reads manifest and scaffolds a target repo
7. [x] Test by bootstrapping a fresh repo (temp dir — works, idempotent)

## Decisions made during implementation

- Scripts stay as `scripts/` (not `scripts/ariadne/`) — symlinked as-is, namespace would break relative sourcing
- Makefile stays as `Makefile.workflow` (not `Makefile.ariadne`) — already designed as includable
- setup.sh IS idempotent — re-running is a no-op
- Settings merge uses `$merge_keys` in the JSON to declare which keys combine vs replace
- `AGENTS.md` now references `@AGENTS.local.md` at the bottom for repo extensions
- Base layer governance section added to AGENTS.md

## Open Questions

- Versioning: does consuming repo pin an ariadne commit or track HEAD? (for now: track HEAD, git pull in ariadne updates everything)
- Skill name conflicts: simple override-by-name for now

## Log

- 2026-04-20: Created from brainstorming session. See `docs/vision/2026-04-20-02-pensive-ariadne-exportable-fragments.md` for original thinking.
- 2026-04-20: Implemented base layer — base.manifest, settings.ariadne.json, setup.sh, atlas/workflow/ docs, AGENTS.local.md pattern. Tested successfully against temp repo.
