---
scope: repo
target: .
---

# Intent: superpowers → ariadne

## Conversation 1 (2026-04-15): Initial adaptation for ariadne conventions

User: Follow the same practice as parley.nvim adaptation, but keep Visual Companion intact (may do web development later).

AI: Adapting all 14 skills with these changes:

**Universal changes:**
- Namespace flattening: all `superpowers:` cross-references → `superpowers-`
- All paths `docs/superpowers/specs/` and `docs/superpowers/plans/` → ariadne conventions

**Key skill-specific adaptations:**
- **brainstorming**: Spec output → `## Spec` section of the issue file (`issues/<slug>.md`). Detailed design plans → `docs/plans/<slug>-plan.md`. Visual Companion kept intact.
- **writing-plans**: Plan output → `docs/plans/<slug>-plan.md`. Execution handoff defers to AGENTS.md Section 3 instead of mandating subagent-driven-development.
- **subagent-driven-development**: Added IMPORTANT note to consult AGENTS.md Section 3 before using. Removed "REQUIRED" as default execution path.
- **executing-plans**: Changed unconditional subagent mandate to conditional guidance per AGENTS.md Section 3.
- **using-superpowers**: Added ariadne note that AGENTS.md Section 3 governs subagent strategy.
- **using-git-worktrees**: Removed `~/.config/superpowers/worktrees/` global path option.
- **requesting-code-review**: Fixed example paths to `docs/plans/`.
- **All other skills**: Cross-ref flattening only, copied as-is.

### Verify
- No `superpowers:` cross-references in ANY file
- No references to `docs/superpowers/specs/` or `docs/superpowers/plans/` in any file
- subagent-driven-development defers to AGENTS.md Section 3, not mandated as default
- executing-plans conditionally suggests subagent-driven-development, not unconditionally
- using-superpowers includes ariadne note about AGENTS.md Section 3
- using-git-worktrees does not reference `~/.config/superpowers/worktrees/`
- writing-plans saves to `docs/plans/`
- requesting-code-review example uses `docs/plans/`
- All 14 skill frontmatter names use `superpowers-` prefix
- All cross-references use `superpowers-` format
- brainstorming spec output goes to `## Spec` section of issue file, NOT separate file
- brainstorming Visual Companion section is PRESENT (not removed)
- spec-document-reviewer-prompt.md references the issue file for spec location

## Conversation 2 (2026-04-21): Align paths with updated AGENTS.md

User: AGENTS.md now uses `workshop/plans/` instead of `docs/plans/` for detailed designs. The directory structure section was updated — `workshop/` is the active workspace, `docs/` is for durable artifacts. All plan references in skills should use `workshop/plans/` consistently.

AI: Updating all skills that reference `docs/plans/` to use `workshop/plans/` instead:
- **brainstorming**: `docs/plans/<slug>-plan.md` → `workshop/plans/<slug>-plan.md`
- **writing-plans**: `docs/plans/<slug>-plan.md` → `workshop/plans/<slug>-plan.md`
- **subagent-driven-development**: `docs/plans/` → `workshop/plans/`
- **requesting-code-review**: `docs/plans/` → `workshop/plans/`

Also updating issue file paths to use `workshop/issues/` prefix consistently where referenced.

### Verify
- No references to `docs/plans/` in any rendered skill
- All plan paths use `workshop/plans/`
- Issue file references use `workshop/issues/` where applicable
- All previous verify clauses from Conversation 1 still hold (except paths updated from `docs/plans/` to `workshop/plans/`)

## Conversation 3 (2026-05-26): Modes prose + gist vocabulary + default-acceptance markers

User: Three additions to the brainstorming skill, all tied to the same intuition — we want the skill to produce data that can compound across sessions, not just resolve the immediate brainstorm.

(1) Brainstorming modes prose. Brainstorming has multiple shapes: crystallization (idea exists, fuzzy, converge), feasibility (hypothesis open, probe with cheap tests), domain-learning (terrain unknown, build vocabulary + prior art + smallest demo). Agent infers which is in play from operator's framing — do NOT ask. Capable models do this naturally; prose makes the behavior visible and overridable.

(2) Gist vocabulary for clarifying questions. Each clarifying question gets tagged inline with a short gist marker — `(gist: scope)`, `(gist: alternative)`, etc. Starter vocabulary: scope, alternative, success-criteria, constraint, assumption, prior-art. Suggestion not enforcement — agents introduce new tags when nothing fits; new tags feed back into the vocabulary later.

(3) Default-acceptance markers. When AskUserQuestion is used with a recommended option, after the user answers, emit an HTML-comment marker capturing gist + chose-default vs. chose-other. Format `<\!-- brainstorm-log: gist=X chose=default|other option=Y -->`. As the first line of the agent's next message. Grep-able; accumulates without separate logging infra.

Issue context: ariadne workshop issue #34 covers (2) and (3); (1) was deferred from an earlier exchange. All three land here.

AI: Added three sections to SKILL.md — modes section between Anti-Pattern and Checklist; gist-tagging instruction under The Process → Understanding the idea; default-acceptance marker convention in the same section.

### Verify
- A "Brainstorming Modes" section names crystallization / feasibility / domain-learning, and includes "infer from cues — do not ask" (or equivalent)
- Starter gist vocabulary listed: scope, alternative, success-criteria, constraint, assumption, prior-art
- Default-acceptance marker format shown as HTML comment with gist=, chose=, option= fields
- All previous verify clauses from Conversations 1-2 still hold

## Conversation 4 (2026-05-26): writing-plans — Core concepts section replaces File Structure

User: Planning is where DRY and PURE principles should insert themselves into how code is structured. The agent should surface key load-bearing abstractions/concepts in the code, explain the reasoning behind those choices, what's implied downstream, and what DRY/PURE considerations they embedded when designing.

The existing File Structure section names file boundaries but doesn't force entity-level conceptual thinking. Replace it with a Core concepts section that's:
- Always required (not optional based on abstraction count — the absence of an abstraction listing is the failure mode we want to catch, not just the presence of a bad one).
- Split into PURE entities (the conceptual core — almost always pure; if state or IO sneaks in, it's an integration point in disguise) and INTEGRATION points (where pure meets the world). A plan with no integration points is a smell — features almost always need side effects to be useful.
- Folds the existing file-organization heuristics into a brief block at the top.
- Each entity declares: where it lives, relationships (1:1, 1:N, ownership direction), DRY rationale (what duplication it eliminates), future extension axes.
- Integration points additionally declare: what they wrap, what pure entities they're injected into.

The pure-first ordering and the structural split (pure entities listed first, then integration) is the lens that forces the agent to think conceptually before writing code.

AI: Replaced ## File Structure with ## Core concepts. Heuristics from the old section preserved as a brief lead-in. Two sub-sections — Pure entities and Integration points — with an IssueWindow + GitRunner example demonstrating the convention.

### Verify
- A "## Core concepts" section exists between Plan Document Header / Scope Check and Bite-Sized Task Granularity
- "## File Structure" section does not exist (replaced)
- Section is marked "always required" (or equivalent)
- Has a "Pure entities" sub-section AND an "Integration points" sub-section
- The integration-absence smell is called out in the Integration points sub-section
- Pure entities sub-section names "no IO mocks" or equivalent purity-boundary criterion
- Each entity template includes: Lives in, Relationships, DRY rationale, Future extensions
- Integration points template additionally includes: Wraps, Injected into
- File-organization heuristics (clear boundaries, smaller focused files, co-locate-by-responsibility, follow existing patterns) preserved in the lead-in
- All previous verify clauses from Conversations 1-3 still hold

## Conversation 5 (2026-05-26): writing-plans — entity status + greppable table format

User: Two updates to Core concepts:

(1) Entities should declare a status — new / modified / deleted. Current shape only covers introducing entities; planning should also surface entities being modified (existing shape/contract changes) or deleted (removed in favor of a different abstraction). Coverage gap: today plans implicitly assume net-additive change.

(2) The Core concepts entity listing should be a greppable table, not just prose bullets. The milestone-review judge (sdlc judge milestone-review) cross-checks the table against the diff — for each row, verifies the entity exists at the stated path with the stated status. Prose bullets aren't reliably parseable; a markdown table is.

Both tie back to milestone-review's ability to enforce that the plan and the code agree.

Combined design: table at top (Name / Lives in / Status / [Wraps for integration]), prose detail below for each non-trivial entity (Relationships, DRY rationale, Future extensions, Injected into for integration). Table is the load-bearing surface; prose is supporting context.

AI: Restructured the Core concepts section's two sub-sections:

- Pure entities: added a markdown table (Name / Lives in / Status), status legend explaining new/modified/deleted semantics, callout noting the table is what milestone-review grep-checks against. Prose detail below for Relationships, DRY rationale, Future extensions.

- Integration points: same shape, table adds a "Wraps" column. Prose below covers Injected into + Future extensions.

Example updated to show both tables with the IssueWindow + GitRunner case, with status=new for both.

### Verify
- "## Core concepts" section still exists, still marked "always required"
- "### Pure entities" has a markdown table with columns: Name, Lives in, Status
- "### Integration points" has a markdown table with columns: Name, Lives in, Status, Wraps
- Status legend present, names all three states (new, modified, deleted) with one-line semantics each
- Callout noting that the milestone-review judge cross-checks the table against the diff
- Prose detail below tables (Relationships, DRY rationale, Future extensions for pure; Injected into, Future extensions for integration)
- Example shows tables for IssueWindow (pure, new) and GitRunner (integration, new, wraps exec.Command)
- All previous verify clauses from Conversations 1-4 still hold
