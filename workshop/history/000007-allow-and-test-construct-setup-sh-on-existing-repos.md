---
id: 000007
status: done
deps: []
created: 2026-04-21
updated: 2026-05-03
---

# Allow and test construct/setup.sh on existing repos

## Done when

- `/construct adopt ../parley.nvim` migrates an existing repo to ariadne's base layer
- After adoption, `./ariadne-refresh.sh` (setup.sh) is purely repair/refresh — no decisions needed
- Repo-specific content preserved in extension points (AGENTS.local.md, settings.local.json, Makefile.local)

## Spec

### Problem

`setup.sh` works great for fresh repos but skips existing regular files silently. For repos that already have `AGENTS.md`, `CLAUDE.md`, `Makefile`, `.claude/settings.json`, etc., we need a nuanced one-time migration that:

1. Reads `construct/base.manifest` as the source of truth for what ariadne manages
2. For each manifest entry, checks if the target repo has an existing file
3. Extracts repo-specific content into the appropriate extension point
4. Backs up originals, then lets `setup.sh` create symlinks

### Extension Points

| Ariadne manages | Repo-specific goes in |
|---|---|
| `AGENTS.md` (symlink) | `AGENTS.local.md` |
| `CLAUDE.md` (symlink) | N/A (CLAUDE.md just references AGENTS.md) |
| `.claude/settings.json` (merge) | `.claude/settings.local.json` |
| `Makefile` (template) | `Makefile.local` |
| `Makefile.workflow` (symlink) | N/A |
| Skills (symlinks) | Kept if not conflicting, otherwise noted |

### Flow

`/construct adopt <path>` (e.g., `/construct adopt ../parley.nvim`):

1. Parse `construct/base.manifest` — get the list of managed files/actions
2. For each `symlink` entry where target repo has a regular file:
   - Read the existing file
   - Read ariadne's version
   - AI judges what's repo-specific vs what ariadne replaces
   - Write repo-specific content to the extension point
3. For existing `Makefile`:
   - Extract repo-specific targets to `Makefile.local`
   - Replace with ariadne's template (includes Makefile.workflow + Makefile.local)
4. For existing `.claude/settings.json`:
   - Diff against `settings.ariadne.json`
   - Write differences to `settings.local.json`
5. Back up all originals to `.ariadne-backup/`
6. Run `setup.sh` to create all symlinks and scaffolding
7. Show summary of what was migrated where

### Post-adoption

After `/construct adopt`, `setup.sh` becomes purely mechanical:
- Repairs broken symlinks
- Creates missing scaffolding
- Merges settings (ariadne base + local overrides)
- No decisions, no skipping — everything is either a symlink or in an extension point

## Plan

- [x] Write spec (this file)
- [x] Add `/construct adopt` command to construct skill SKILL.md
- [x] Test on `../parley.nvim`

## Log

### 2026-04-21

- Analyzed setup.sh behavior with existing files — skips regular files silently
- Identified extension points: AGENTS.local.md, settings.local.json, Makefile.local
- Designed `/construct adopt` as the nuanced one-time migration, leaving setup.sh as repair/refresh

### 2026-05-03

- `/construct adopt` shipped in `.claude/skills/construct/SKILL.md`; confirmed tested on `../parley.nvim`. Closing.
