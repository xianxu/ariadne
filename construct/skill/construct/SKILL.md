---
name: xx-construct
description: Use when managing the AI substrate — importing, adapting, promoting, upgrading, or rolling back skills and constitution files across repos.
---

# The Construct — AI Substrate Management

Centralized management of AI skills and constitution files for ariadne. Imports external skill sources, adapts them via semantic intent transcripts, and deploys to ariadne's own `.claude/skills/`. Derivative repos inherit the adapted skills through `construct/adapted/` (wired by `construct/base.manifest`), so they never run `/construct adapt` themselves.

**Core principles:**
- **Ariadne adapts; derivatives inherit.** Adaptation is single-target: it always renders for ariadne. Downstream repos pick up ariadne's `construct/adapted/` verbatim via the base layer.
- **Store the intent, not the patch.** Re-apply behavioral intent onto each new upstream version.
- **Vendor at the source level.** Skills in a source (e.g., superpowers) are a cohesive set — vendor all, render all. Unadjusted skills get the source version as-is.
- **Lazy vendor.** Don't snapshot a source until someone first adjusts a skill from it.

## Directory Layout

**CRITICAL PATH RULE:** Define `$REPO_ROOT` as the git repository root (the directory containing `.git/`). All `construct/` paths in this document resolve as `$REPO_ROOT/construct/`. NEVER resolve paths relative to this skill file's location (`.claude/skills/construct/`). Before any file operation, verify you are writing to `$REPO_ROOT/construct/`, not to `.claude/skills/construct/`.

```
$REPO_ROOT/construct/                         # top-level, ariadne's AI substrate workspace
  config.json                                 # construct configuration (localPrefix, etc.)
  manifest.md                                 # index of managed artifacts
  local/<skill>/SKILL.md                      # locally-authored skills (source of truth)
  sources/<source>/<version>/                 # frozen upstream snapshot (ALL skills from source)
    skills/<skill>/SKILL.md                   # individual skill files
  intents/<source>.md                         # one intent transcript per source (covers all its skills)
  intents/constitution/                       # evolution tracking for AGENTS.md, CLAUDE.md
  intents/local/<skill>/                      # intents for locally-created skills
  staging/skills/<skill>/                     # render target before promotion (gitignored)
  adapted/<skill>/                            # promoted output; derivatives symlink to this directory
  versions/NNNN[-slug]/                       # last 10 snapshots of rendered state
  current                                     # marker: active version number
  rollback.sh                                 # non-AI emergency revert

# Live skill location:
$REPO_ROOT/.claude/skills/<skill>/            # deployed adapted skill (ariadne's own)

# Local skills are symlinked with configurable prefix:
$REPO_ROOT/.claude/skills/{prefix}<skill>/  →  ../../construct/local/<skill>/
```

### Config File

`$REPO_ROOT/construct/config.json` holds construct-wide settings:

```json
{
  "localPrefix": "xx-"
}
```

- **`localPrefix`** — prefix applied when symlinking local skills to `.claude/skills/`. Prevents name collisions with upstream or community skills. A skill at `construct/local/voice-apply/` becomes `.claude/skills/xx-voice-apply/`.

## Commands

### `/construct local`

Lists all local skills and their symlink status.

1. Read `localPrefix` from `$REPO_ROOT/construct/config.json`
2. Scan `$REPO_ROOT/construct/local/` for skill directories (each must contain `SKILL.md`)
3. For each skill, check if symlink exists at `.claude/skills/{prefix}<skill>/`
4. Report status table:

```
construct/local/voice-apply       →  .claude/skills/xx-voice-apply       ✓
construct/local/voice-gen         →  .claude/skills/xx-voice-gen         ✓
construct/local/skill-gen         →  .claude/skills/xx-skill-gen         ✗ (missing)
```

### `/construct new <name>`

Creates a new local skill. Replaces the standalone `xx-skill-gen` skill.

1. Read `localPrefix` from `$REPO_ROOT/construct/config.json`
2. Validate `<name>` doesn't already exist in `$REPO_ROOT/construct/local/`
3. Ask user: "What should this skill do? Give me a brief description of its purpose, when it should trigger, and what behavior it should produce."
4. If the description is ambiguous, invoke `superpowers-brainstorming` to explore intent
5. Scaffold `$REPO_ROOT/construct/local/<name>/SKILL.md` following `superpowers-writing-skills` conventions:
   - YAML frontmatter with `name: {prefix}<name>` and `description` (starts with "Use when...")
   - Process section, rules section
6. Create symlink: `.claude/skills/{prefix}<name>/` → `../../construct/local/<name>/`
7. Present the generated skill for review
8. Skill is immediately live via symlink — no promote step needed


### `/construct adapt <source>`

The primary authoring command. Adapts an entire source (all its skills) for ariadne via conversation. The adaptation is ariadne-only — downstream repos inherit verbatim via `construct/adapted/`.

- `<source>` — e.g., `superpowers`

**Flow:**
1. **Lazy vendor:** If `$REPO_ROOT/construct/sources/<source>/` doesn't exist yet, find and snapshot the entire source:
   - Look in plugin cache: `~/.claude/plugins/cache/**/<source>/*/skills/`
   - Look in ariadne's live skills: `$REPO_ROOT/.claude/skills/`
   - Look in personal: `~/.claude/skills/`
   - Snapshot ALL skills from the source into `$REPO_ROOT/construct/sources/<source>/<version>/`
2. Read ALL source skills from `$REPO_ROOT/construct/sources/<source>/<version>/skills/`
3. Read existing intent transcript from `$REPO_ROOT/construct/intents/<source>.md` (if any)
4. Discuss desired changes with the user — the conversation covers all skills in the source. User may have opinions on some skills (brainstorming, tdd) and leave others as-is. All of this goes in one intent transcript.
5. **Render ALL skills from this source:**
   - Apply the full intent (existing transcript + new conversation) to each skill
   - Skills the user didn't mention in the conversation → copy source as-is
   - **Namespace flattening:** rename each skill directory to `<source>-<skill>` (e.g., `superpowers-brainstorming`, `superpowers-tdd`)
   - **Rewrite cross-references:** scan all rendered SKILL.md files and replace `/<source>:` with `/<source>-` in internal references (e.g., `/superpowers:writing-plans` → `/superpowers-writing-plans`)
   - Write all rendered skills to `$REPO_ROOT/construct/staging/skills/`
6. Show diff for adapted skills vs. current live version at `$REPO_ROOT/.claude/skills/`
7. Extract verify clauses from the conversation
8. Dispatch verify subagent (fresh context) with rendered output + verify clauses
9. If verify fails → re-render with failure context (max 3 attempts)
10. If verify passes → append new conversation to intent transcript, report ready
11. Tell user: "Staging is ready. Run `/construct promote <source>` to deploy all skills."

**For new local skills (no upstream source):**
Use `/construct new <name>` instead. Local skills live in `construct/local/` and are symlinked to `.claude/skills/` with the configured prefix. They don't go through the staging/promote pipeline — they're live immediately via symlink.

### `/construct promote <source>`

Promotes ALL staged skills from a source. Deploys as a cohesive set to both ariadne's live skills directory and the `construct/adapted/` tree that derivative repos inherit. Explicit user action, never automatic.

1. Check that `$REPO_ROOT/construct/staging/skills/` exists and has content
2. Determine next version number: scan `$REPO_ROOT/construct/versions/` for highest NNNN, increment
3. Snapshot current live state into `$REPO_ROOT/construct/versions/NNNN/`:
   - Copy all skills being deployed
   - Write manifest to `versions/NNNN/manifest.md`
4. **Delete-then-copy:** First, remove all `<source>-*` skill directories from both `$REPO_ROOT/.claude/skills/` and `$REPO_ROOT/construct/adapted/` (e.g., `rm -rf .claude/skills/superpowers-*/ construct/adapted/superpowers-*/`). Then copy ALL staged skills into both locations. This ensures removed or renamed skills don't linger, and keeps the inheritable `adapted/` tree in lockstep with ariadne's live skills.
5. **Disable source plugin in ariadne:** Set the source plugin to `false` in `$REPO_ROOT/.claude/settings.json` under `enabledPlugins`. This prevents the global plugin's skills from shadowing the adapted local versions. Example for superpowers:
   ```json
   { "enabledPlugins": { "superpowers@claude-plugins-official": false } }
   ```
   If the file already exists, merge into the existing `enabledPlugins` object.
6. Update `$REPO_ROOT/construct/current` with the new version number
7. Clear `$REPO_ROOT/construct/staging/`
8. Prune versions beyond the last 10 (by numeric prefix, oldest first)
9. Update `$REPO_ROOT/construct/manifest.md`

Derivative repos pick up the new versions automatically the next time they refresh: their `construct/adapted/` is a symlink (or vendored copy) of ariadne's `construct/adapted/`, and their `.claude/skills/superpowers-*` symlinks point through it.

### `/construct import <source>`

Explicitly fetches and snapshots an entire upstream source. Optional — `adapt` triggers lazy import on first use. Useful for pre-staging or inspecting a source before adjusting.

1. Identify source location:
   - For Claude Code plugins: `~/.claude/plugins/cache/<marketplace>/<plugin>/<version>/skills/`
   - For git repos: clone to temp, extract skills directory
2. Create `$REPO_ROOT/construct/sources/<source>/<version>/` and copy ALL skill files from the source
3. Report: source name, version, list of all skills found

### `/construct upgrade <source>`

Fetches a new upstream version and re-renders all skills from that source.

1. Fetch new version (same as import, into a new version directory under `$REPO_ROOT/construct/sources/<source>/`)
2. Render ALL skills from the source:
   - Skills covered by `$REPO_ROOT/construct/intents/<source>.md` → re-apply the intent transcript to the new source version
   - Skills not mentioned in the intent → copy new source as-is
3. Write to `$REPO_ROOT/construct/staging/skills/`
4. Show diff vs. current live version at `$REPO_ROOT/.claude/skills/`
5. Dispatch verify subagent for each adjusted skill
6. Report which skills passed verification, which failed
7. User calls `/construct promote <source>` to deploy

### `/construct diff <skill> [--vs source|live] [<version-a>] [<version-b>]`

Shows diffs using `git diff --no-index` for familiar output format.

- `/construct diff <skill>` — staging vs. live (default)
- `/construct diff <skill> --vs source` — staging vs. source (shows what the intent changed)
- `/construct diff <skill> <version>` — version vs. live
- `/construct diff <skill> <version-a> <version-b>` — between two versions

`--vs` controls what staging is compared against:
- `--vs live` (default) — compare staging to what's currently deployed at `$REPO_ROOT/.claude/skills/<skill>/`
- `--vs source` — compare staging to the vendored source, showing exactly what intents changed

Paths:
- Live: `$REPO_ROOT/.claude/skills/<skill>/`
- Source: `$REPO_ROOT/construct/sources/<source>/<version>/skills/<skill>/`
- Staging: `$REPO_ROOT/construct/staging/skills/<skill>/`

### `/construct status`

Shows the current state of all managed artifacts.

1. Read `$REPO_ROOT/construct/manifest.md`
2. Read `$REPO_ROOT/construct/current`
3. Check if anything is in `$REPO_ROOT/construct/staging/`
4. Report: managed skills, their sources, active version, staging state

### `/construct rollback <version>`

Emergency revert to a previous version.

1. Run `$REPO_ROOT/construct/rollback.sh <version>`
2. Report what was restored

For listing available versions: `$REPO_ROOT/construct/rollback.sh --list`

## Intent Transcript Format

Intent files are conversation transcripts — the authoritative record of human-AI dialogue that produced the adaptation. They live at `$REPO_ROOT/construct/intents/<source>.md` — one file per source. A single intent file covers ALL skills from that source.

```markdown
# Intent: superpowers → ariadne

## Conversation 1 (2026-04-15): Initial adaptation for ariadne conventions

User: For brainstorming — spec output should land in the `## Spec` section
of the issue file, and detailed designs go to `workshop/plans/<slug>-plan.md`.

AI: Updating brainstorming accordingly...

User: For writing-plans — same target, `workshop/plans/<slug>-plan.md`.

AI: Updating writing-plans accordingly...

User: Leave all other skills as-is for now.

### Verify
- All spec write instructions point to the issue file's `## Spec` section
- writing-plans output goes to `workshop/plans/`
- No references to `docs/superpowers/specs/` in any rendered skill

## Conversation 2 (2026-04-21): Path realignment

User: AGENTS.md now uses `workshop/plans/` instead of `docs/plans/`...

### Verify
- All plan references point to `workshop/plans/`
```

New conversations are appended. Verify clauses accumulate. The full transcript is given to the AI during rendering — it understands the behavioral intent and applies it to whatever version of the source is current.

## Verify Subagent

When dispatching the verify subagent, use this prompt structure:

```
You are verifying a rendered skill. You have NO prior context about how
this skill was created. Read the output files and check each verify clause
independently.

For each clause, report:
- PASS: with evidence (quote the relevant text or confirm absence)
- FAIL: with what you found instead

Rendered files:
[list file paths in construct/staging/skills/]

Verify clauses:
[list all verify clauses from the intent transcript]
```

The subagent must be dispatched as a fresh agent (not resumed) to prevent confirmation bias.

## Version Manifest Format

Each version's `manifest.md` records what produced it:

```markdown
# Version NNNN Manifest

Promoted: YYYY-MM-DDTHH:MM:SSZ

## Managed Skills

| Skill | Source | Source Version | Intent File |
|-------|--------|---------------|-------------|
| superpowers-brainstorming | superpowers | v5.0.2 | intents/superpowers.md |
| superpowers-tdd | superpowers | v5.0.2 | intents/superpowers.md |
| superpowers-writing-plans | superpowers | v5.0.2 | intents/superpowers.md |

## Constitution Files

| File |
|------|
| AGENTS.md |
| CLAUDE.md |

## Local-Origin Skills

| Skill | Source Dir | Symlink |
|-------|-----------|---------|
| construct | construct/skill/ | .claude/skills/construct/ (copied, not symlinked) |
| xx-datatype | construct/local/datatype/ | .claude/skills/xx-datatype/ → ../../construct/local/datatype/ |
| xx-voice-apply | construct/local/voice-apply/ | .claude/skills/xx-voice-apply/ → ../../construct/local/voice-apply/ |
| xx-voice-gen | construct/local/voice-gen/ | .claude/skills/xx-voice-gen/ → ../../construct/local/voice-gen/ |
| xx-interview-feedback | construct/local/interview-feedback/ | .claude/skills/xx-interview-feedback/ → ../../construct/local/interview-feedback/ |
| xx-skill-gen | construct/local/skill-gen/ | .claude/skills/xx-skill-gen/ → ../../construct/local/skill-gen/ |
```

## Key Rules

- **Never auto-promote.** Always render to staging, show diff, let user decide.
- **Verify with fresh eyes.** Always dispatch a new subagent for verification, never verify in the current session.
- **Intents are transcripts.** The conversation is the authoritative artifact, not a distilled spec. Transcripts survive upstream restructuring because they describe behavior, not location.
- **Ariadne adapts; derivatives inherit.** There is exactly one adaptation target: ariadne itself. Downstream repos pick up the rendered output via `construct/adapted/` (wired by `construct/base.manifest`). They never run `/construct adapt`.
- **rollback.sh is sacred.** Never modify it through the Construct's own rendering pipeline. It must always work independently.
- **Last 10 versions.** Prune at promotion time. Slugs are cosmetic, don't protect from pruning.
- **Source-level coherence.** Always vendor, render, and promote all skills from a source together. Unadjusted skills get the source version as-is — they're still deployed so ariadne gets the full, consistent set.
- **Self-sync.** The source of truth for the construct skill is `$REPO_ROOT/construct/skill/SKILL.md`. After ANY edit to that file, immediately copy it to `$REPO_ROOT/.claude/skills/construct/SKILL.md` to keep the live version in sync. This is the one skill that bootstraps itself.
- **Namespace flattening.** Plugin skills use `plugin:skill` namespacing which can't be overridden locally. When deploying, rename skill directories to `<source>-<skill>` and rewrite all internal `/<source>:` references to `/<source>-`. This makes adapted skills invocable as `/<source>-<skill>` without conflicting with the global plugin.
- **Local skills are symlinked, not copied.** Source of truth is `$REPO_ROOT/construct/local/<skill>/`. Symlinks in `.claude/skills/{prefix}<skill>/` point back to source. Edits to either location affect the same file. The prefix (default `xx-`) is configured in `construct/config.json` and prevents collisions with upstream or community skills.
- **weave renders the symlinks.** A layer's skills are declared by a `skill <dir>` intent in `construct/base.manifest`; `weave compile` (via `make weave`) lowers them to the `.claude/skills/<prefix><name>` symlinks (each pointing at the source layer's skill dir) and prunes orphaned ones. This replaced the old `SessionStart` hook running `construct/scripts/sync-local-skills.sh` (both retired in #95). No manual intervention needed.
