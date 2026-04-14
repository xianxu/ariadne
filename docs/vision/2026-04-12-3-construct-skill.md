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

### 2. Intents Are Conversation Transcripts

The intent for a skill is not a distilled spec — it's the **conversation transcript** of human-AI dialogue that produced the adaptation.

**Why transcripts beat specs:** A distilled spec says "change section 3.2 line 42" — that breaks when upstream restructures. A transcript says "we wanted to stop the skill from opening a browser because we're terminal-only." That intent survives any restructuring because it's about *behavior*, not *location*. The AI reading the transcript understands the *why* and figures out the *where* in any version of the source.

**The authoring process:** `/construct adapt` starts a conversation. You discuss what to change, iterate until the output is right. The conversation itself gets saved as the intent file. There is no separate "write the intent file" step — the conversation *is* the intent.

**Verify clauses** emerge naturally during conversation ("make sure there's no mention of browser anywhere") and are collected at the end of the transcript. A separate agent (subagent with no prior context) checks rendered output against them. This prevents confirmation bias — fresh eyes, always.

**Unification:** This is the same mechanism for all cases:
- Adapting an upstream skill → transcript of conversation about what to change
- Creating a new skill from scratch → transcript describing what the skill should do (the intent *is* the generative spec)
- Evolving an existing skill → new conversation appended to the transcript

**Transcript growth:** As intents accumulate, transcripts may get long. The natural escape is decomposing the skill into sub-skills (see Future Direction below).

### 3. Rollback.sh — The Only Script

The Construct uses only one standalone script: `rollback.sh`, the non-AI emergency revert. All other mechanical operations (fetch, place, swap, verify structure) are described as steps in the Construct SKILL.md and executed by the AI using standard tools.

This distinction matters because the Construct modifies the very substrate it runs on. If the AI layer breaks, `rollback.sh` still works for recovery — it's the one piece that must function when everything else is broken.

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
construct/                          # top-level, the repo's AI substrate workspace
  manifest.md                       # index of all managed artifacts
  sources/superpowers/v5.0.2/       # frozen upstream snapshots
  intents/superpowers/              # conversation transcripts (the intents)
  intents/constitution/             # evolution tracking for AGENTS.md, CLAUDE.md
  intents/local/                    # intents for locally-created skills
  staging/                          # render target before promotion (gitignored)
  versions/0001/                    # snapshots of rendered state (last 10)
  current                           # marker: active version number
  rollback.sh                       # non-AI emergency revert

.claude/skills/
  construct/SKILL.md                # the meta-skill (interface to the system)
  brainstorming/SKILL.md            # rendered output (live)
  my-local-skill/SKILL.md           # local-origin, no upstream
```

`construct/` is a **top-level directory**, on par with `atlas/` and `workshop/` — a first-class concept in the AI-first repo. The Construct skill at `.claude/skills/construct/SKILL.md` is the interface; `construct/` is the workspace. Claude Code scans `.claude/skills/` for skills; everything in `construct/` is invisible to skill discovery.

## Operations

- **`/construct adjust <skill>`** — The primary command. Starts a conversation to capture desired changes. For imported skills, adapts behavior. For new names, creates a new skill (bootstraps directory structure). Renders to staging, shows diff vs. live, runs verification. Does NOT auto-promote.
- **`/construct promote <skill>`** — Promotes staged version to live. Snapshots into versions, copies to `.claude/skills/`, updates manifest. Explicit user action, always.
- **`/construct import <source>`** — Fetch upstream, snapshot into sources, create skeleton intent file.
- **`/construct upgrade <source>`** — Fetch new upstream version, re-render all skills from that source using existing intents, show diffs. Does NOT auto-promote.
- **`/construct diff [<skill>] [<version-a>] [<version-b>]`** — Show diffs using `git diff --no-index` between staging, live, or any version. Familiar git-compatible output format.
- **`/construct status`** — Show manifest: what's managed, active version, source versions, staging state.
- **`/construct rollback <version>`** — Revert to a previous version via rollback.sh.

## Render Flow (within `/construct adjust`)

1. Conversation with user to capture desired changes
2. Read source files from `substrate/sources/` + existing intent transcript
3. AI applies behavioral intent to source → renders to `substrate/staging/`
4. Show diff: `git diff --no-index` between live and staging
5. Extract verify clauses, dispatch verify subagent (fresh context, no bias)
6. If fail → re-render with failure context (max 3 attempts, then abandon)
7. If pass → save conversation as intent transcript, staging is ready
8. User reviews diff + verify results, calls `/construct promote` to go live

Because the AI reads behavioral intent (not positional specs), rendering is robust to upstream restructuring.

## Failure and Recovery

When a version fails and is rolled back:
1. Revert immediately via `rollback.sh` (non-AI)
2. Failed version stays in `versions/` as evidence
3. Investigate root cause: upstream problem, bad intent, Construct bug, or intent conflicts with upstream
4. Fix root cause, re-render → new version supersedes the failed one
5. Failed version pruned when outside the 10-version window

Intent is always "ahead of" materialized skills. Fixing and re-rendering is the normal flow.

## Future Direction: Constitution Decomposition

As the AI substrate grows, AGENTS.md becomes too monolithic — too large for context windows, too tangled for humans to reason about which parts apply when. The trajectory:

- **Today:** AGENTS.md is a monolith covering workflow, task management, design principles, directory structure, verification, etc.
- **Future:** AGENTS.md becomes a slim bootloader (identity, principles, directory structure) pointing to contextual skills loaded on demand.

Each section becomes a skill: workflow-orchestration, task-management, code-quality, verification, atlas-maintenance, construct itself. These are importable, adaptable via intents, distributable individually. A new repo picks what it needs — "I want your verification workflow but not your atlas system."

The Construct's model already supports this. No structural changes needed — constitutional rules are just skills that define repo-wide behavior.

## Connection to Ariadne

This is foundational infrastructure for the AI-first hermetic repo model described in the Phases doc. Every Ariadne repo — whether a solo founder's workspace or a 10-person team's monorepo — will need to manage its AI substrate. The Construct is how skills become living artifacts that evolve through use while maintaining lineage to their origins.

The same pattern extends to MCP servers, connectors, and other AI infrastructure — but v1 focuses on skills and constitution files.

## Implementation

Bootstrapping in parley.nvim (the current operating repo). Design and implementation tracked in:
- Issue: `workshop/issues/000101-a-system-to-update-local-skills-and-handle-merge-with-upstream.md`
- Plan: `workshop/plans/000101-construct-plan.md`

Will migrate to ariadne repo at cutover.
