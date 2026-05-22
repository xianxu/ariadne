---
type: pensive
date: 2026-05-22
topic: durable-target-pattern
mode: eureka
description: Target is a new artifact type above projects in the dependency graph. Narrative intent, under-specified by design, durable across sessions. Projects manage execution that flows from a target; issues are leaf work units that may reference a target directly. Human accumulates the full target document; agent reads the delta. 🤖{} convention for agent contributions; chat holds provisional thinking; target file holds commitments.
references: [/Users/xianxu/workspace/nous/AGENTS.md, /Users/xianxu/workspace/brain/memory/feedback_iterative_intent.md, /Users/xianxu/workspace/brain/memory/feedback_under_specify.md]
---

# Pensive: the durable target pattern

## The asymmetry

There's a productivity primitive in how human and agent should read a shared intent document differently:

- **Human reads the full document.** Continuity matters — re-reading the whole thing keeps the through-line in mind, makes contradictions visible, lets situational context resurface.
- **Agent reads the delta.** What changed since last session is what needs acting on. The agent's omniscient attention to all word-pairs makes diff-reading cheap.

Most spec-driven workflows force the human to do the diff work themselves — changelogs, tickets, "what's new in v2." The durable target pattern offloads that to the agent and lets the human just keep accumulating intent into one document.

## Target is a separate datatype

Earlier draft of this pensive treated targets as a posture for root issues. After more thinking, that conflates two different things. Targets want to be a *separate* datatype, sitting above projects and issues in the dependency graph:

- **Target** = intent. Narrative prose. Durable across sessions. Under-specified by design. *What we want and what for.*
- **Project** = execution shell. Tracks the work that flows from a target — milestones, timeline, assignees, status. *When and by whom.*
- **Issue** = work unit. May reference a target directly (for small work) or live under a project (for larger initiatives). *What to do.*

Why three types? Because they change at different cadences. Intent shifts when understanding evolves — slow, narrative-driven, occasional. Execution churns constantly — checkboxes flip, milestones slip, assignments move. Issues open and close on a fast clock. Forcing them into one file means either intent gets buried under execution-noise diffs, or execution updates pollute a document you wanted to re-read in 30 seconds. Three types let each evolve at its native rate.

The dependency graph: target ←referenced-by── project ──manages──→ issues. Each issue may additionally reference a target directly when it's not part of a managed project.

## The `## Problem`-extraction promotion path

Issue files today combine: problem (why), spec (what), plan (how), log (history). The `## Problem` section has always been doing target work — it's the narrative why, separable from the spec/plan/log execution.

For small issues this is fine — the issue *is* the target, embedded inline. The promotion path comes when the same problem ends up referenced from multiple places: a second issue addressing a different facet, a project tracking many issues against it, or a parley exploring it. At that point, extract `## Problem` into a standalone target file; both issues and the project link to it via `target:` frontmatter.

Most issues never need promotion. The promotion gesture is cheap, and the cost of premature promotion (a target with one referencer) is small. Wait until the references are real.

## Under-specification as the discipline

What differentiates targets from PRDs is the under-specification posture. PRDs over-specify by reflex — sub-features, success metrics, edge cases. A target stays narrative: what we want and what for, with the *what for* getting most of the words. Specifics emerge later, in projects and issues, and only when the agent (or the operator's next read of the target) reveals a gap that needs to be filled.

The rule: **only get more specific when the agent fills the gap wrong.** A target that says "create a shared brain infrastructure and user interface" doesn't need to enumerate features unless the agent's natural decomposition produces the wrong ones. Trust the substrate to derive specifics; let the target stay durable.

This makes targets git-friendly in a particular way: most edits to a target are *insertions* (refined narrative, new sub-paragraph), not *modifications* (changed wording). Modifications mean intent shifted; insertions mean intent crystallized further. Both are legitimate, but the ratio tells you something about how stable a target is.

## Lifecycle

Projects have an execution state machine: open → working → done. Targets have an intent journal:

- `active` — current commitment.
- `achieved` — the world now looks the way the target wanted.
- `split` — single target broke into two because they turned out to be separable commitments. Both halves get their own files; original gets a pointer to the splits.
- `deferred` — still want it, but not now.
- `abandoned` — no longer want it. Kept for context; doesn't get deleted.

A target getting split as understanding evolves is git-natural. The original isn't "closed" — its content becomes the seed for the two children.

## Schema sketch

```yaml
---
type: target
slug: shared-brain-infrastructure-and-ui
status: active
created: 2026-05-22
---

# Target: <title>

<narrative — 1-3 paragraphs of "what we want and what for">

## Why now
<motivation — 1-2 paragraphs>

## What this is NOT
<rule-outs — under-specifying by negation, narrowing without enumerating>

## Open questions
<things we're explicitly choosing not to commit on yet>
```

Projects + issues gain `target: <slug>` in their frontmatter. Grep by slug = the dependency graph; you can see at a glance which execution work is advancing which intent.

## Conventions for agent collaboration on target files

**`🤖{...}` for agent contributions.** Inline markers for agent-proposed additions or refinements. Accept/reject gestures are one-edit-class operations:

- Accept: delete `🤖{` and `}`, leave the text in place.
- Reject: delete the whole `🤖{...}` block.
- Edit: change the content inside the braces.

Each is git-natural — the diff IS the contract change, atomically reviewable.

**Chat is provisional; target is commitment.** Provisional thinking ("I'm thinking aloud," "open question," "what if we did X") stays in chat. After back-and-forth converges, the agent updates the relevant section of the target file with the new commitment. The file is the contract; the chat is the working memory. Agents diff the target across sessions to know what's actionable.

**Where it lives.** In each project repo's `workshop/targets/` (mirroring the existing `workshop/issues/`, `workshop/plans/`, `workshop/parley/` layout). The *convention* (this pensive's claim) belongs in ariadne's AGENTS.md §1 artifact hierarchy.

## Why this matters

The pattern is the explicit form of disciplines already implicit in this system: iterative intent ("42 shots, not one"), under-specification, brain as durable substrate, single-threaded human attention. Targets solve the "I needed to repeat myself" problem by making intent durable across sessions. Agents read deltas; humans want continuity; the convention gives each side what it needs without forcing either to compromise.

Three artifact types — target (intent), project (execution shell), issue (work unit) — three cadences, three audiences. Each stays slim because the others absorb their natural content.

## Open questions

- Do targets need a status workflow tool (analog of `make close-issue`), or is the journal-shaped status field enough?
- Should `## What this is NOT` be a standard section, or operator-discretion? The under-specification discipline benefits from explicit negation, but mandating it might create empty boilerplate.
- How do targets evolve when their referencing projects/issues teach us the target was wrong? Direct edit, or a "supersedes" pointer to a successor target?
- Does the 🤖{} convention extend to other agent surfaces — code review, atlas drafts, project files? Likely yes; the convention is cheap and the gesture set is the same.
- Where does the parley → target transition happen? Parley is exploration; target is commitment. Some parley chats clearly want to be promoted to targets when the operator's intent crystallizes. A `## Crystallization` section in parley that becomes the target's body, perhaps.
