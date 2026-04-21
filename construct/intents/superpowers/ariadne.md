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
