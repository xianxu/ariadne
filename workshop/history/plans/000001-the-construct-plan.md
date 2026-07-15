# The Construct — AI Substrate Management System

**Issue:** #000001
**Date:** 2026-04-12
**Status:** Working
**Originated:** parley.nvim #000101, migrated to ariadne for privacy

---

## Problem

AI-first repos depend on a substrate of markdown artifacts (skills, constitution files) that instruct AI behavior. These artifacts come from upstream sources (e.g., superpowers plugin) and get customized locally, per-repo. When upstream updates, local customizations must be preserved. Textual merge of English documents produces plausible but potentially incoherent results — there's no compiler to catch semantic inconsistency.

Additionally, different repos need different adaptations of the same skill. Brainstorming for parley.nvim (which uses `workshop/plans/`) differs from brainstorming for a future project. A centralized control plane is needed.

## Core Insight

**Store the intent, not the patch. Re-apply intent onto each new upstream version.**

Rather than merging text, the Construct records *why* each local change was made (as conversation transcripts), then re-applies those intents from scratch onto new upstream versions. The AI reads the whole new version and produces a coherent document.

## What It Manages

- **Skills** — imported from upstream (e.g., superpowers) or created locally, deployed to target repos or personal scope
- **Constitution files** — AGENTS.md, CLAUDE.md — local-origin, versioned and rollback-able

## Architecture

### Ariadne as Control Plane

Ariadne is the centralized place to manage skills across repos. All construct operations run from ariadne. Skills get deployed to target repos via relative paths.

- Sources, intents, staging, versions all live in `ariadne/construct/`
- Target repos referenced by relative path from ariadne root (e.g., `../parley.nvim`)
- Plugin cache (`~/.claude/plugins/cache/`) is the primary source for upstream skills

### Scope

- **Personal** → `~/.claude/skills/<skill>/` — available in all repos
- **Repo** → `<target>/.claude/skills/<skill>/` — target-specific

### Source-Level Coherence

Skills in a source (e.g., superpowers) are a cohesive set. Always:
- **Vendor all** — snapshot all skills from a source together
- **Render all** — render all skills for a target together (adjusted ones get intents, rest copied as-is)
- **Promote all** — deploy the full set to maintain cross-skill consistency

### Lazy Vendor

Don't snapshot a source until first adjust. Lookup order:
1. Already vendored: `$REPO_ROOT/construct/sources/<source>/`
2. Plugin cache: `~/.claude/plugins/cache/**/<source>/*/skills/`
3. Target repo: `<target>/.claude/skills/`
4. Personal: `~/.claude/skills/`

### Intents as Transcripts

Intent files are conversation transcripts — the authoritative record of human-AI dialogue. Per-skill, per-target-repo.

**Why transcripts, not specs:** A distilled spec says "change section 3.2 line 42" — breaks when upstream restructures. A transcript says "we wanted to stop the skill from opening a browser because we're terminal-only" — survives any restructuring because it's about *behavior*, not *location*.

File location: `$REPO_ROOT/construct/intents/<source>/<skill>/<repo-name>.md`

Frontmatter records scope and target path:
```yaml
---
scope: repo
target: ../parley.nvim
---
```

### Rendering and Verification

1. AI reads source files + intent transcript → produces rendered output
2. Verify subagent (fresh context, no confirmation bias) checks against verify clauses
3. Max 3 attempts; on failure, abandon and keep previous version

### Versioning and Rollback

- Last 10 snapshots in `$REPO_ROOT/construct/versions/NNNN/`
- `rollback.sh` — non-AI, standalone, works when everything else is broken
- Handles both personal and repo-scoped skills

## Directory Structure

```
$REPO_ROOT/construct/                         # ariadne's AI substrate workspace
  manifest.md                                 # index of managed artifacts
  sources/<source>/<version>/                 # frozen upstream snapshot (ALL skills)
    skills/<skill>/SKILL.md
  intents/<source>/<skill>/                   # per-skill, per-repo transcripts
    <repo-name>.md                            # e.g., parley.nvim.md
    personal.md                               # personal-scoped intent
  intents/constitution/                       # evolution tracking for AGENTS.md, CLAUDE.md
  intents/local/<skill>/                      # locally-created skills
  staging/<target-slug>/                      # render target before promotion
    skills/<skill>/SKILL.md
  versions/NNNN[-slug]/                       # last 10 snapshots
  current                                     # active version marker
  rollback.sh                                 # non-AI emergency revert
  skill/SKILL.md                              # source of truth for construct skill

$REPO_ROOT/.claude/skills/construct/SKILL.md  # live copy (auto-synced from above)
```

## Commands

| Command | Purpose |
|---------|---------|
| `/construct adjust <source>:<skill> --to <relpath>` | Adapt a skill for a target repo via conversation |
| `/construct promote <source> --to <relpath>` | Deploy all staged skills from source to target |
| `/construct import <source>` | Explicitly vendor an upstream source |
| `/construct upgrade <source>` | Fetch new upstream, re-render all for all targets |
| `/construct diff <skill> --to <relpath> [--vs source\|live]` | Show diffs |
| `/construct status` | Show managed artifacts state |
| `/construct rollback <version>` | Emergency revert |

## Failure Modes and Recovery

| Failure | Cause | Recovery |
|---------|-------|----------|
| Bad render | Intent unclear or conflicting | `rollback.sh NNNN`, fix intent, re-render |
| Upstream incompatible | Major upstream restructure | `rollback.sh NNNN`, review changes, update intents |
| Construct bug | Skill itself has a defect | `rollback.sh NNNN` restores everything including Construct |
| Path confusion | Construct writes to wrong dir | $REPO_ROOT anchoring prevents this |
| Verify false positive | Verify clauses too loose | Rollback, tighten verify clauses |
| Render exhaustion | 3 attempts all fail | Staging cleared, previous version stays active |

## Future Direction

- **Constitution decomposition** — AGENTS.md → slim bootloader pointing to contextual skills
- **Multi-repo skill synchronization** — upgrade across all targets in one command
- **Automatic upstream change detection** — currently manual fetch
