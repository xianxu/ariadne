---
name: construct
description: Use when managing the AI substrate — importing, adapting, promoting, upgrading, or rolling back skills and constitution files across repos.
---

# The Construct — AI Substrate Management

Centralized management of AI skills and constitution files across repos. Imports external skill sources, adapts them via semantic intent transcripts, and deploys to target repos. Ariadne is the control plane.

**Core principles:**
- **Store the intent, not the patch.** Re-apply behavioral intent onto each new upstream version.
- **Vendor at the source level.** Skills in a source (e.g., superpowers) are a cohesive set — vendor all, render all. Unadjusted skills get the source version as-is.
- **Lazy vendor.** Don't snapshot a source until someone first adjusts a skill from it.

## Scope: Personal vs Repo

Every managed skill has a **scope** that determines where it deploys:

- **personal** → `~/.claude/skills/<skill>/` — available across all projects on this machine
- **repo** → `<target>/.claude/skills/<skill>/` — available only in the target repo

For repo-scoped skills, the **target** is a relative path from ariadne's root to the target repo. Examples:
- `.` → ariadne itself
- `../parley.nvim` → sibling repo
- `../../other/project` → anywhere reachable by relative path

Scope and target are declared in the intent transcript frontmatter and recorded in the manifest. The same source can have different intents for different target repos — superpowers adapted for parley.nvim differs from superpowers adapted for a future project.

## Directory Layout

**CRITICAL PATH RULE:** Define `$REPO_ROOT` as the git repository root (the directory containing `.git/`). All `construct/` paths in this document resolve as `$REPO_ROOT/construct/`. NEVER resolve paths relative to this skill file's location (`.claude/skills/construct/`). Before any file operation, verify you are writing to `$REPO_ROOT/construct/`, not to `.claude/skills/construct/`.

```
$REPO_ROOT/construct/                         # top-level, ariadne's AI substrate workspace
  config.json                                 # construct configuration (localPrefix, etc.)
  manifest.md                                 # index of managed artifacts
  local/<skill>/SKILL.md                      # locally-authored skills (source of truth)
  sources/<source>/<version>/                 # frozen upstream snapshot (ALL skills from source)
    skills/<skill>/SKILL.md                   # individual skill files
  intents/<source>/                            # per-source, per-target intent transcripts
    personal.md                               # personal-scoped intent (if any)
    <repo-name>.md                            # repo-scoped intent (one per target repo)
  intents/constitution/                       # evolution tracking for AGENTS.md, CLAUDE.md
  intents/local/<skill>/                      # intents for locally-created skills
  staging/<target-slug>/                      # render target before promotion (gitignored)
    skills/<skill>/SKILL.md                   # all skills from the source, rendered for this target
  versions/NNNN[-slug]/                       # last 10 snapshots of rendered state
  current                                     # marker: active version number
  rollback.sh                                 # non-AI emergency revert

# Live skill locations (determined by scope + target):
~/.claude/skills/<skill>/                     # personal-scoped
<target>/.claude/skills/<skill>/              # repo-scoped (target is relative path)

# Local skills are symlinked with configurable prefix:
<target>/.claude/skills/{prefix}<skill>/  →  ../../construct/local/<skill>/
```

### Config File

`$REPO_ROOT/construct/config.json` holds construct-wide settings:

```json
{
  "localPrefix": "xx-"
}
```

- **`localPrefix`** — prefix applied when symlinking local skills to `.claude/skills/`. Prevents name collisions with upstream or community skills. A skill at `construct/local/voice-apply/` becomes `.claude/skills/xx-voice-apply/`.

**Intent file naming:** defaults to the repo name (last path component of `--to`). The frontmatter `target:` field is the source of truth for where it deploys — the filename is just a readable identifier. Override with `--as` if needed (e.g., two repos with the same name at different paths).

Examples:
- `--to .` → `ariadne.md`
- `--to ../parley.nvim` → `parley.nvim.md`
- `--to ../../work/nexhealth-api` → `nexhealth-api.md`
- `--to personal` → `personal.md`
- `../../other/project` → `other-project`

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


### `/construct adapt <source> --to <relpath> [--as <slug>]`

The primary authoring command. Adapts an entire source (all its skills) for a target repo via conversation.

- `<source>` — e.g., `superpowers`
- `--to <relpath>` — target repo, relative from ariadne root. Defaults to `.` (ariadne itself). Use `--to personal` for personal scope.
- `--as <name>` — (optional) override the intent filename. Rarely needed since it defaults to the repo name.
- The intent filename defaults to the **repo name** (last path component): `../parley.nvim` → `parley.nvim.md`, `.` → `ariadne.md`, `personal` → `personal.md`.
- Examples:
  - `/construct adapt superpowers --to ../parley.nvim` → `intents/superpowers/parley.nvim.md`
  - `/construct adapt superpowers --to personal` → `intents/superpowers/personal.md`
  - `/construct adapt superpowers` → `intents/superpowers/ariadne.md`

**Flow:**
1. Resolve intent filename: use `--as` if provided, otherwise the repo name (last path component of `--to`)
2. **Lazy vendor:** If `$REPO_ROOT/construct/sources/<source>/` doesn't exist yet, find and snapshot the entire source:
   - Look in plugin cache: `~/.claude/plugins/cache/**/<source>/*/skills/`
   - Look in target repo: `<target>/.claude/skills/`
   - Look in personal: `~/.claude/skills/`
   - Snapshot ALL skills from the source into `$REPO_ROOT/construct/sources/<source>/<version>/`
3. Read ALL source skills from `$REPO_ROOT/construct/sources/<source>/<version>/skills/`
4. Read existing intent transcript from `$REPO_ROOT/construct/intents/<source>/<repo-name>.md` (if any)
5. If target is a repo path, read the target repo's AGENTS.md and CLAUDE.md for context on its conventions
6. Discuss desired changes with the user — the conversation covers all skills in the source. User may have opinions on some skills (brainstorming, tdd) and leave others as-is. All of this goes in one intent transcript.
7. **Render ALL skills from this source for the target:**
   - Apply the full intent (existing transcript + new conversation) to each skill
   - Skills the user didn't mention in the conversation → copy source as-is
   - **Namespace flattening:** rename each skill directory to `<source>-<skill>` (e.g., `superpowers-brainstorming`, `superpowers-tdd`)
   - **Rewrite cross-references:** scan all rendered SKILL.md files and replace `/<source>:` with `/<source>-` in internal references (e.g., `/superpowers:writing-plans` → `/superpowers-writing-plans`)
   - Write all rendered skills to `$REPO_ROOT/construct/staging/<target-slug>/skills/`
8. Show diff for adapted skills vs. current live version in the target location
9. Extract verify clauses from the conversation
10. Dispatch verify subagent (fresh context) with rendered output + verify clauses
11. If verify fails → re-render with failure context (max 3 attempts)
12. If verify passes → append new conversation to intent transcript, report ready
13. Tell user: "Staging is ready. Run `/construct promote <source> --to <relpath>` to deploy all skills."

**For new local skills (no upstream source):**
Use `/construct new <name>` instead. Local skills live in `construct/local/` and are symlinked to `.claude/skills/` with the configured prefix. They don't go through the staging/promote pipeline — they're live immediately via symlink.

### `/construct promote <source> --to <relpath>`

Promotes ALL staged skills from a source to the target. Deploys as a cohesive set. Explicit user action, never automatic.

- `--to` must match what was used during adapt. Read from staging metadata if omitted.

1. Check that `$REPO_ROOT/construct/staging/<target-slug>/skills/` exists and has content
2. Resolve target directory:
   - `--to personal` → `~/.claude/skills/`
   - `--to <relpath>` → `<relpath>/.claude/skills/` (resolved relative to ariadne root)
   - `--to .` or omitted → `.claude/skills/` (ariadne itself)
3. Verify the target path exists (for repo targets, check it's a git repo)
4. Determine next version number: scan `$REPO_ROOT/construct/versions/` for highest NNNN, increment
5. Snapshot current live state into `$REPO_ROOT/construct/versions/NNNN/`:
   - Record all skills being deployed, their source, target, scope
   - Write manifest to `versions/NNNN/manifest.md`
6. **Delete-then-copy:** First, remove all `<source>-*` skill directories from the target's `.claude/skills/` (e.g., `rm -rf <target>/.claude/skills/superpowers-*/`). Then copy ALL staged skills into the target. This ensures removed or renamed skills don't linger.
7. **Disable source plugin in target repo:** Set the source plugin to `false` in `<target>/.claude/settings.json` under `enabledPlugins`. This prevents the global plugin's skills from shadowing the adapted local versions. Example for superpowers:
   ```json
   { "enabledPlugins": { "superpowers@claude-plugins-official": false } }
   ```
   If the file already exists, merge into the existing `enabledPlugins` object. Skip this step for `--to personal`.
8. Update `$REPO_ROOT/construct/current` with the new version number
9. Clear `$REPO_ROOT/construct/staging/<target-slug>/`
10. Prune versions beyond the last 10 (by numeric prefix, oldest first)
11. Update `$REPO_ROOT/construct/manifest.md`

### `/construct import <source>`

Explicitly fetches and snapshots an entire upstream source. Optional — `adapt` triggers lazy import on first use. Useful for pre-staging or inspecting a source before adjusting.

1. Identify source location:
   - For Claude Code plugins: `~/.claude/plugins/cache/<marketplace>/<plugin>/<version>/skills/`
   - For git repos: clone to temp, extract skills directory
2. Create `$REPO_ROOT/construct/sources/<source>/<version>/` and copy ALL skill files from the source
3. Create intent directory at `$REPO_ROOT/construct/intents/<source>/`
4. Report: source name, version, list of all skills found

### `/construct upgrade <source>`

Fetches a new upstream version and re-renders all skills from that source, for each target that has intents.

1. Fetch new version (same as import, into a new version directory under `$REPO_ROOT/construct/sources/<source>/`)
2. For each target repo that has intents for skills from this source:
   a. Render ALL skills from the source for this target:
      - Skills with intents → re-apply intent transcripts to new source version
      - Skills without intents → copy new source as-is
   b. Write to `$REPO_ROOT/construct/staging/<target-slug>/skills/`
   c. Show diff vs. current live version in target
   d. Dispatch verify subagent for each adjusted skill
3. Report results per target: which skills passed verification, which failed
4. User calls `/construct promote <source> --to <relpath>` for each target they want to update

### `/construct diff <skill> [--to <relpath>] [--vs source|live] [<version-a>] [<version-b>]`

Shows diffs using `git diff --no-index` for familiar output format.

- `/construct diff <skill> --to ../parley.nvim` — staging vs. live (default)
- `/construct diff <skill> --to ../parley.nvim --vs source` — staging vs. source (shows what the intent changed)
- `/construct diff <skill> <version>` — version vs. live
- `/construct diff <skill> <version-a> <version-b>` — between two versions

`--vs` controls what staging is compared against:
- `--vs live` (default) — compare staging to what's currently deployed in the target
- `--vs source` — compare staging to the vendored source, showing exactly what intents changed

Paths resolve based on scope and target:
- Live (personal): `~/.claude/skills/<skill>/`
- Live (repo): `<target>/.claude/skills/<skill>/`
- Source: `$REPO_ROOT/construct/sources/<source>/<version>/skills/<skill>/`
- Staging: `$REPO_ROOT/construct/staging/<target-slug>/skills/<skill>/`

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

Intent files are conversation transcripts — the authoritative record of human-AI dialogue that produced the adaptation. They live at `$REPO_ROOT/construct/intents/<source>/<repo-name>.md` — one file per source+target pair (or `personal.md` for personal scope). A single intent file covers ALL skills from that source for that target.

```markdown
---
scope: repo                    # "personal" or "repo"
target: ../parley.nvim         # relative path from ariadne root (omit for personal)
---

# Intent: superpowers → parley.nvim

## Conversation 1 (2026-04-12): Initial adaptation

User: For brainstorming — change where design docs get written. Parley uses
workshop/plans/ as the execution space, not docs/superpowers/specs/.
Also remove the Visual Companion section entirely — we work in terminal-only Neovim.

AI: I'll change all spec output paths in brainstorming to workshop/plans/
and remove the Visual Companion section...

User: For writing-plans — same path change, workshop/plans/ instead of
docs/superpowers/specs/.

AI: Updating writing-plans accordingly...

User: Leave all other skills as-is for now.

### Verify
- No references to `docs/superpowers/specs/` in any rendered skill
- No mention of "Visual Companion", "browser", or "mockup" in brainstorming
- All spec write instructions point to `workshop/plans/`

## Conversation 2 (2026-04-15): Add test coverage step to brainstorming

User: I want brainstorming to require a test plan before approval...

### Verify
- Brainstorming checklist includes "Write test plan" step before "User approves design"
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
| superpowers-brainstorming | superpowers | v5.0.2 | repo | ../parley.nvim | intents/superpowers/parley.nvim.md |
| superpowers-tdd | superpowers | v5.0.2 | repo | ../parley.nvim | intents/superpowers/parley.nvim.md |
| superpowers-writing-plans | superpowers | v5.0.2 | repo | ../parley.nvim | intents/superpowers/parley.nvim.md |

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
- **rollback.sh is sacred.** Never modify it through the Construct's own rendering pipeline. It must always work independently.
- **Last 10 versions.** Prune at promotion time. Slugs are cosmetic, don't protect from pruning.
- **Source-level coherence.** Always vendor, render, and promote all skills from a source together. Unadjusted skills get the source version as-is — they're still deployed so the target gets the full, consistent set.
- **Self-sync.** The source of truth for the construct skill is `$REPO_ROOT/construct/skill/SKILL.md`. After ANY edit to that file, immediately copy it to `$REPO_ROOT/.claude/skills/construct/SKILL.md` to keep the live version in sync. This is the one skill that bootstraps itself.
- **Namespace flattening.** Plugin skills use `plugin:skill` namespacing which can't be overridden locally. When deploying, rename skill directories to `<source>-<skill>` and rewrite all internal `/<source>:` references to `/<source>-`. This makes adapted skills invocable as `/<source>-<skill>` without conflicting with the global plugin.
- **Local skills are symlinked, not copied.** Source of truth is `$REPO_ROOT/construct/local/<skill>/`. Symlinks in `.claude/skills/{prefix}<skill>/` point back to source. Edits to either location affect the same file. The prefix (default `xx-`) is configured in `construct/config.json` and prevents collisions with upstream or community skills.
- **Auto-healing symlinks.** A `SessionStart` hook runs `construct/scripts/sync-local-skills.sh` to automatically repair or create missing symlinks. No manual intervention needed.
