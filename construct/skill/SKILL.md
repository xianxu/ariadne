---
name: construct
description: Use when managing the AI substrate — importing, adapting, upgrading, or rolling back skills and constitution files. Invoke for /construct adjust, /construct promote, /construct import, /construct upgrade, /construct diff, /construct status, /construct rollback.
---

# The Construct — AI Substrate Management

Manages the AI substrate of this repo: skills, constitution files (AGENTS.md, CLAUDE.md). Imports external skills, adapts them via semantic intent transcripts, versions rendered output, and re-applies intents onto new upstream versions.

**Core principle:** Store the intent, not the patch. Re-apply behavioral intent onto each new upstream version.

## Scope: Personal vs Repo

Every managed skill has a **scope** that determines where it deploys:

- **personal** → `~/.claude/skills/<skill>/` — available across all projects on this machine
- **repo** → `<target>/.claude/skills/<skill>/` — available only in the target repo

For repo-scoped skills, the **target** is a relative path from ariadne's root to the target repo. Examples:
- `.` → ariadne itself
- `../parley.nvim` → sibling repo
- `../../other/project` → anywhere reachable by relative path

Scope and target are declared in the intent transcript frontmatter and recorded in the manifest. The same upstream skill can have different intents for different target repos — brainstorming adapted for parley.nvim's `workshop/` layout is different from brainstorming adapted for a future project.

## Directory Layout

```
construct/                                    # top-level, ariadne's AI substrate workspace
  manifest.md                                 # index of managed artifacts
  sources/<source>/<version>/                 # frozen upstream snapshots
  intents/<source>/<skill>/                   # per-repo intent transcripts
    personal.md                               # personal-scoped intent (if any)
    <repo-name>.md                            # repo-scoped intent (one per target repo)
  intents/constitution/                       # evolution tracking for AGENTS.md, CLAUDE.md
  intents/local/<skill>/                      # intents for locally-created skills
  staging/                                    # render target before promotion (gitignored)
  versions/NNNN[-slug]/                       # last 10 snapshots of rendered state
  current                                     # marker: active version number
  rollback.sh                                 # non-AI emergency revert

# Live skill locations (determined by scope + target):
~/.claude/skills/<skill>/                     # personal-scoped
<target>/.claude/skills/<skill>/              # repo-scoped (target is relative path)
```

**Intent file naming:** defaults to the repo name (last path component of `--to`). The frontmatter `target:` field is the source of truth for where it deploys — the filename is just a readable identifier. Override with `--as` if needed (e.g., two repos with the same name at different paths).

Examples:
- `--to .` → `ariadne.md`
- `--to ../parley.nvim` → `parley.nvim.md`
- `--to ../../work/nexhealth-api` → `nexhealth-api.md`
- `--to personal` → `personal.md`
- `../../other/project` → `other-project`

## Commands

### `/construct adjust <source>:<skill> [--to <relpath>] [--as <slug>]`

The primary authoring command. Starts a conversation to capture what the user wants to change.

- `<source>:<skill>` — e.g., `superpowers:brainstorming`
- `--to <relpath>` — target repo, relative from ariadne root. Defaults to `.` (ariadne itself). Use `--to personal` for personal scope.
- `--as <name>` — (optional) override the intent filename. Rarely needed since it defaults to the repo name.
- The intent filename defaults to the **repo name** (last path component): `../parley.nvim` → `parley.nvim.md`, `../../work/nexhealth-api` → `nexhealth-api.md`, `.` → `ariadne.md`, `personal` → `personal.md`.
- Examples:
  - `/construct adjust superpowers:brainstorming --to ../parley.nvim` → `intents/superpowers/brainstorming/parley.nvim.md`
  - `/construct adjust superpowers:brainstorming --to personal` → `intents/superpowers/brainstorming/personal.md`
  - `/construct adjust superpowers:brainstorming` → `intents/superpowers/brainstorming/ariadne.md`

**For imported skills (skill exists in construct/sources/):**
1. Resolve intent filename: use `--as` if provided, otherwise the repo name (last path component of `--to`)
2. Read the current source files from `construct/sources/<source>/<version>/`
3. Read existing intent transcript from `construct/intents/<source>/<skill>/<repo-name>.md` (if any)
4. If target is a repo path, read the target repo's AGENTS.md and CLAUDE.md for context on its conventions
5. Discuss desired changes with the user — understand behavioral intent
6. Apply the full intent (existing transcript + new conversation) to the source
7. Write rendered output to `construct/staging/skills/<skill>/`
8. Show diff vs. current live version in the target location
9. Extract verify clauses from the conversation
10. Dispatch verify subagent (fresh context) with rendered output + verify clauses
11. If verify fails → re-render with failure context (max 3 attempts)
12. If verify passes → append new conversation to intent transcript, report ready
13. Tell user: "Staging is ready. Run `/construct promote <skill> --to <relpath>` to go live."

**For new skills (skill does not exist yet):**
1. Ask user: **personal** or **repo** (with `--to` path)?
2. Discuss what the skill should do — the conversation is the generative spec
3. Write rendered SKILL.md to `construct/staging/skills/<skill>/`
4. Show the rendered output
5. Extract verify clauses, dispatch verify subagent
6. If verify passes → save conversation as intent transcript at `construct/intents/local/<skill>/<repo-name>.md`
7. Tell user: "Staging is ready. Run `/construct promote <skill> --to <relpath>` to go live."

### `/construct promote <skill> [--to <relpath>]`

Promotes the staged version to live. Explicit user action, never automatic.

- `--to` must match what was used during adjust. Read from staging metadata if omitted (staging records the target from the adjust step).

1. Check that `construct/staging/skills/<skill>/` exists
2. Resolve target directory:
   - `--to personal` → `~/.claude/skills/<skill>/`
   - `--to <relpath>` → `<relpath>/.claude/skills/<skill>/` (resolved relative to ariadne root)
   - `--to .` or omitted → `.claude/skills/<skill>/` (ariadne itself)
3. Verify the target path exists and is a valid directory (for repo targets, check it's a git repo)
4. Determine next version number: scan `construct/versions/` for highest NNNN, increment
5. Snapshot current live state of the skill being promoted into `construct/versions/NNNN/`:
   - Record: skill name, source, target, scope, previous content
   - Write manifest to `versions/NNNN/manifest.md`
6. Copy staged skill to the target directory (overwrite existing)
7. Update `construct/current` with the new version number
8. Clear `construct/staging/`
9. Prune versions beyond the last 10 (by numeric prefix, oldest first)
10. Update `construct/manifest.md`

### `/construct import <source>`

Fetches upstream skills and snapshots them.

1. Identify source location:
   - For Claude Code plugins: `~/.claude/plugins/cache/<marketplace>/<plugin>/<version>/skills/`
   - For git repos: clone to temp, extract skills directory
2. Create `construct/sources/<source>/<version>/` and copy all skill files
3. Create skeleton intent files at `construct/intents/<source>/<skill>.md` for each skill found
4. Report what was imported

### `/construct upgrade <source>`

Fetches a new upstream version and re-renders all skills from that source.

1. Fetch new version (same as import, but into a new version directory)
2. For each skill managed from this source:
   a. Read new source files
   b. Read existing intent transcript
   c. Re-render: AI reads transcript, applies behavioral intent to new source
   d. Write to `construct/staging/skills/<skill>/`
   e. Show diff vs. current live version
   f. Dispatch verify subagent
3. Report results for all skills — which passed, which failed
4. User calls `/construct promote <skill>` for each skill they want to promote

### `/construct diff <skill> [--to <relpath>] [<version-a>] [<version-b>]`

Shows diffs using `git diff --no-index` for familiar output format.

- `/construct diff <skill> --to ../parley.nvim` — staging vs. live in parley.nvim
- `/construct diff <skill> <version>` — version vs. live
- `/construct diff <skill> <version-a> <version-b>` — between two versions

Paths resolve based on scope and target:
- Live (personal): `~/.claude/skills/<skill>/`
- Live (repo): `<target>/.claude/skills/<skill>/`
- Staging: `construct/staging/skills/<skill>/`

### `/construct status`

Shows the current state of all managed artifacts.

1. Read `construct/manifest.md`
2. Read `construct/current`
3. Check if anything is in `construct/staging/`
4. Report: managed skills, their sources, active version, staging state

### `/construct rollback <version>`

Emergency revert to a previous version.

1. Run `construct/rollback.sh <version>`
2. Report what was restored

For listing available versions: `construct/rollback.sh --list`

## Intent Transcript Format

Intent files are conversation transcripts — the authoritative record of human-AI dialogue that produced the adaptation. They live at `construct/intents/<source>/<skill>/<target-slug>.md` — one file per target repo (or `personal.md` for personal scope).

```markdown
---
scope: repo                    # "personal" or "repo"
target: ../parley.nvim         # relative path from ariadne root (omit for personal)
---

# Intent: superpowers/brainstorming → parley.nvim

## Conversation 1 (2026-04-12): Initial adaptation

User: We need to change where design docs get written. Parley uses
workshop/plans/ as the execution space, not docs/superpowers/specs/.

AI: I'll change all references to the spec output path...

User: Also remove the Visual Companion section entirely...

AI: Removing the Visual Companion section...

### Verify
- No references to `docs/superpowers/specs/` in rendered output
- No mention of "Visual Companion", "browser", or "mockup"
- Checklist includes atlas update step

## Conversation 2 (2026-04-15): Add test coverage step

User: I want brainstorming to require a test plan before approval...

### Verify
- Checklist includes "Write test plan" step before "User approves design"
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
[list file paths in construct/staging/]

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

| Skill | Source | Source Version | Scope | Target | Intent File |
|-------|--------|---------------|-------|--------|-------------|
| brainstorming | superpowers | v5.0.2 | repo | ../parley.nvim | intents/superpowers/brainstorming/parley.md |

## Constitution Files

| File |
|------|
| AGENTS.md |
| CLAUDE.md |

## Local-Origin Skills

| Skill |
|-------|
| construct |
```

## Key Rules

- **Never auto-promote.** Always render to staging, show diff, let user decide.
- **Verify with fresh eyes.** Always dispatch a new subagent for verification, never verify in the current session.
- **Intents are transcripts.** The conversation is the authoritative artifact, not a distilled spec. Transcripts survive upstream restructuring because they describe behavior, not location.
- **rollback.sh is sacred.** Never modify it through the Construct's own rendering pipeline. It must always work independently.
- **Last 10 versions.** Prune at promotion time. Slugs are cosmetic, don't protect from pruning.
