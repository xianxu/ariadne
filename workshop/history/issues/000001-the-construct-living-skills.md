---
id: 000001
status: done
deps: []
created: 2026-04-12
updated: 2026-04-13
---

# The Construct — AI Substrate Management System

A skill-based system to manage the AI substrate (skills, constitution files) across repos. Ariadne is the centralized control plane. Import external skills, adapt them via semantic intent transcripts, version rendered output, and re-apply intents onto new upstream versions instead of text-merging.

Named after The Matrix's Construct — where you load up what you need and reshape it.

## Done when

- The Construct skill exists at `.claude/skills/construct/SKILL.md` and can be invoked
- Can import an upstream skill source (e.g., superpowers), snapshot all skills from that source
- Can adjust: per-skill, per-target-repo intent transcripts via `/construct adjust <source>:<skill> --to <relpath>`
- Can render: source + intents → rendered SKILL.md for all skills in a source, verified by subagent
- Can promote: deploy full source skill set to target repo via `/construct promote <source> --to <relpath>`
- Can upgrade: fetch new upstream version, re-render with existing intents for all targets, verify, promote
- Versioning works: last 10 snapshots, non-AI rollback via `rollback.sh`
- Scope support: personal (`~/.claude/skills/`) and repo (`<target>/.claude/skills/`)
- Constitution files (AGENTS.md, CLAUDE.md) are versioned alongside skills

## Spec

**Core insight:** Store the intent, not the patch. Re-apply intent onto each new upstream version.

Textual merge of English documents produces plausible but potentially incoherent results. Instead, record *why* each local change was made (as conversation transcripts), then have AI re-apply those intents from scratch onto new upstream versions.

**What it manages:**
- Skills (imported from upstream + local-origin), deployed to target repos or personal scope
- Constitution files (AGENTS.md, CLAUDE.md) — local-origin, versioned and rollback-able

**Key concepts:**
- **Sources** — frozen upstream snapshots, vendored at source level (all skills from a source together). Lazy vendor on first adjust.
- **Intents** — per-skill, per-target-repo conversation transcripts. One file per (skill, target) pair. The conversation is the authoritative artifact.
- **Rendering** — source-level: all skills from a source rendered together for a target. Adjusted skills get intents applied, unadjusted skills copied as-is.
- **Scoping** — personal (`~/.claude/skills/`) or repo (`<target>/.claude/skills/`). Target repos referenced by relative path from ariadne root.
- **Versioning** — last 10 snapshots. Non-AI rollback via `rollback.sh`.
- **Promotion** — staging passes verification → deployed to target. Never auto-promote.

**Detailed design:** `docs/plans/000001-the-construct-plan.md`

**Live skill:** `construct/skill/SKILL.md` (source of truth, auto-synced to `.claude/skills/construct/SKILL.md`)

## Plan

- [x] Design and brainstorm (originated in parley.nvim #000101)
- [x] Write `rollback.sh` — standalone non-AI revert script
- [x] Write the Construct SKILL.md — orchestrates all operations
- [x] Bootstrap: import superpowers as first managed source
- [x] Add scope support (personal vs repo)
- [x] Add target repo support (`--to <relpath>`)
- [x] Add source-level vendoring and rendering
- [x] Add lazy vendor from plugin cache
- [x] Add `--vs source` to diff command
- [x] Add self-sync rule (construct/skill/SKILL.md → .claude/skills/construct/SKILL.md)
- [x] Fix $REPO_ROOT path resolution (prevent writing to .claude/skills/construct/)
- [x] First real test: adjust superpowers:brainstorming --to ../parley.nvim
- [x] Promote brainstorming to parley.nvim (v0001)
- [ ] Verify rollback works
- [ ] Adjust remaining superpowers skills for parley.nvim as needed
- [ ] Test upgrade flow with a new superpowers version

## Log

### 2026-04-12 (in parley.nvim)

- Brainstormed design with user
- Key decisions:
  - "Store intent, not patch" — semantic re-apply instead of text merge
  - Per-skill intents (not generic) — "spec" means different things in different skills
  - Hermetic repo model — skills vendored in, not referenced externally
  - Versioning with last 10 snapshots + non-AI rollback for safety
  - No baking period — promote on verify pass, rollback is cheap
- Design written, spec review completed

### 2026-04-13

- Migrated construct from parley.nvim to ariadne (private repo)
- Key design evolutions:
  - Intents are conversation transcripts, not distilled specs
  - `/construct adjust` + `/construct promote` — never auto-promote
  - `$REPO_ROOT` anchoring to prevent path confusion with `.claude/skills/construct/`
  - Scope support: personal (`~/.claude/skills/`) and repo (`<target>/.claude/skills/`)
  - Target repos via relative path (`--to ../parley.nvim`)
  - Intent files named by repo name (e.g., `parley.nvim.md`), frontmatter has actual target path
  - Source-level vendoring: all skills from a source vendored, rendered, and promoted together
  - Lazy vendor: don't snapshot until first adjust, look in plugin cache first
  - `--vs source` for diff: compare staging against vendored source
  - Self-sync rule: construct/skill/SKILL.md auto-copied to .claude/skills/construct/SKILL.md
- First successful test: adjusted superpowers:brainstorming for parley.nvim, promoted to v0001
- Bug found and fixed: construct was writing to `.claude/skills/construct/` instead of `construct/` at repo root — fixed with $REPO_ROOT anchoring
