# The Construct — AI Substrate Management for Hermetic Repos

**Date:** 2026-04-12
**Context:** Brainstorming session on how to manage the AI substrate (skills, constitution files) in a hermetic repo. Designing a system that imports external skills, tracks local customizations as semantic intents, and re-applies them onto new upstream versions.

---

## The Problem

AI-first hermetic repos depend on a substrate of markdown artifacts — skills, constitution files (AGENTS.md, CLAUDE.md) — that instruct AI behavior. These artifacts come from upstream sources (e.g., the superpowers plugin) and get customized locally. When upstream updates, local customizations must be preserved.

Textual merge of English documents fails here. A three-way merge can produce something grammatically correct but semantically incoherent — and unlike code, there's no compiler to catch it. "Spec" means design document in brainstorming, test specification in TDD, implementation plan in writing-plans. A generic "change all spec paths" intent silently does the wrong thing in at least one of those.

## Core Insight: Store Intent, Not Patch

Rather than maintaining patches or doing text merges, record *why* each local change was made, then re-apply those intents from scratch onto each new upstream version. The AI reads the whole new version and produces a coherent document, rather than splicing two versions together.

This is the same mechanism whether you're:
- Adapting an upstream skill to local conventions
- Creating a new skill from scratch (the intent *is* the spec)
- Evolving a skill as requirements change

## The Name

The Construct — the Matrix reference. Where you load up what you need and reshape it.

## Key Design Decisions

### 1. Intents Are Per-Skill, Not Generic

Each intent file targets a specific skill with concrete directives. No shared abstraction layer across skills. "Spec" in brainstorming and "spec" in TDD are different things — generic intents would silently do the wrong thing. Accept duplication; refactor when the pattern is clear, not before.

This reflects a broader principle: there's no magic. You can't expect a single sentence and all skills magically start working. Maybe in the happy case, but not *sustainably*.

### 2. Intent File Format: What / Why / Verify

Each directive has three parts:

```markdown
## 1. Spec output location
Change: Design docs go to `workshop/plans/` not `docs/superpowers/specs/`
Why: Our repo uses workshop/ as execution space per AGENTS.md
Verify: The rendered SKILL.md contains no references to `docs/superpowers/specs/`
  and all spec write instructions point to `workshop/plans/`
```

The **verify** clauses are checked by a separate agent (subagent with no prior context) after rendering. This prevents confirmation bias — the agent that rendered the output would naturally think it did it correctly. Fresh eyes, always.

### 3. Mechanical Scripts + Semantic Skill

The Construct is two layers:

**Scripts (mechanical, tested, rigid):** fetch upstream, place rendered output, swap versions, verify structural consistency. These are deterministic and testable. They enforce invariants the AI shouldn't have to remember.

**Construct skill (semantic, AI):** orchestrate the flow, read source + intents to produce rendered output, dispatch verify subagent, create/edit intents through conversation with the user.

The scripts are the skeleton, the skill is the brain. This distinction matters because the Construct modifies the very substrate it runs on — if the AI layer breaks, the scripts still work for recovery.

### 4. Versioning with Non-AI Rollback

Last 10 snapshots of all rendered output. Each version is a complete copy. A simple shell script (`rollback.sh`) can revert to any version without AI involvement.

Version directories are named `NNNN` by default (e.g., `0001`, `0002`). Users may append a slug for annotation: `0002-broken`, `0003-pre-upstream-update`. Scripts match on the numeric prefix.

This is the safety net for the self-referential problem: the Construct manages skills including potentially itself. A bad render could brick the whole system. `rollback.sh` is the non-AI escape hatch.

### 5. No Baking Period

Render → verify → promote immediately. Rollback cost is low enough that a baking period adds complexity without proportional safety. If something goes wrong, revert and investigate.

### 6. Constitution Files Are Local-Origin (for now)

AGENTS.md and CLAUDE.md benefit from versioning and rollback but don't currently have an upstream source. They don't go through the source+intent→render pipeline. If an upstream template is adopted later, they can be promoted to managed artifacts.

### 7. Promotion, Not Merge

When upstream updates, the Construct does NOT merge old+new. It renders from scratch: new upstream source + existing intents → new rendered output. This is always semantically coherent because the AI reads the complete new version as context.

## Directory Structure

```
.claude/
  skills/
    construct/SKILL.md            # the meta-skill itself
    brainstorming/SKILL.md        # rendered output (live)
    debugging/SKILL.md            # rendered output (live)
    my-local-skill/SKILL.md       # local-origin, no upstream

  substrate/
    manifest.md                   # index of all managed artifacts + status
    sources/
      superpowers/
        v5.0.2/                   # frozen upstream snapshot
          skills/brainstorming/SKILL.md
          skills/brainstorming/visual-companion.md
          ...
        v5.1.0/                   # newer upstream version
    intents/
      superpowers/
        brainstorming.md          # per-skill intents (what/why/verify)
        debugging.md
      constitution/
        agents.md                 # evolution tracking for AGENTS.md
        claude.md
    staging/                      # render target before promotion (gitignored)
      skills/brainstorming/SKILL.md
      constitution/AGENTS.md
    versions/
      0001/                       # complete snapshot of rendered state
      0002/
      ...                         # last 10 kept, older pruned at promotion
    current                       # marker: active version number
    rollback.sh                   # non-AI emergency revert
```

Claude Code scans `.claude/skills/*/SKILL.md` one level deep. Everything in `substrate/` is invisible to skill discovery. The Construct manages the boundary.

## Operations

- **`/construct import <source>`** — fetch upstream, snapshot, create skeleton intent file
- **`/construct adapt <skill>`** — discuss changes, write/update intent file with what/why/verify
- **`/construct render [<skill>|all]`** — source + intents → rendered output → verify → promote
- **`/construct upgrade <source>`** — fetch new upstream, re-render all skills from that source
- **`/construct status`** — show manifest: what's managed, versions, sources
- **`/construct rollback <version>`** — revert to any of the last 10 versions

## Render Flow

1. Read all source files for the skill from `substrate/sources/`
2. Read intent file from `substrate/intents/`
3. AI produces all rendered output files → writes to `substrate/staging/`
4. Dispatch verify subagent: rendered output + verify clauses → pass/fail
5. If fail → re-render with failure context (max 3 attempts, then abandon)
6. If pass → snapshot to `substrate/versions/NNNN/`, copy to `.claude/skills/`, update `current`

## Failure and Recovery

When a version fails and is rolled back:
1. Revert immediately via `rollback.sh` (non-AI)
2. Failed version stays in `versions/` as evidence
3. Investigate root cause: upstream problem, bad intent, Construct bug, or intent conflicts with upstream
4. Fix root cause, re-render → new version supersedes the failed one
5. Failed version pruned when outside the 10-version window

Intent is always "ahead of" materialized skills. Fixing and re-rendering is the normal flow.

## Connection to Ariadne

This is foundational infrastructure for the AI-first hermetic repo model described in the Phases doc. Every Ariadne repo — whether a solo founder's workspace or a 10-person team's monorepo — will need to manage its AI substrate. The Construct is how skills become living artifacts that evolve through use while maintaining lineage to their origins.

The same pattern extends to MCP servers, connectors, and other AI infrastructure — but v1 focuses on skills and constitution files.

## Implementation

Bootstrapping in parley.nvim (the current operating repo). Design and implementation tracked in:
- Issue: `workshop/issues/000101-a-system-to-update-local-skills-and-handle-merge-with-upstream.md`
- Plan: `workshop/plans/000101-construct-plan.md`

Will migrate to ariadne repo at cutover.
