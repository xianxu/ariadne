---
scope: repo
target: ../parley.nvim
---

# Intent: superpowers → parley.nvim

## Conversation 1 (2026-04-13): Initial adaptation for parley.nvim conventions

User: Remove Visual Companion entirely (it's browser-based, not relevant for a Neovim plugin). No need to add parley chat integration — if parley chat is used, it will be explicitly referenced in the issue file itself.

AI: Removing the entire Visual Companion section, its checklist step, process flow node, and all supporting files (visual-companion.md, scripts/*). Updating paths from superpowers conventions to parley.nvim conventions: design docs go to `workshop/plans/<issue-slug>-plan.md` instead of `docs/superpowers/specs/`. Adding an "Update atlas" step to the checklist and process flow to match parley.nvim's use of `atlas/` for high-level documentation.

### Verify
- No references to `docs/superpowers/specs/` in rendered output
- No mention of "Visual Companion", "browser", or "mockup"
- No visual companion scripts staged (frame-template.html, helper.js, server.js, start-server.sh, stop-server.sh, visual-companion.md)
- Design doc path references `workshop/plans/`
- Spec document reviewer prompt references `workshop/plans/`
- Checklist includes atlas update step
- Process flow does NOT include Visual Companion nodes
- HARD-GATE section is preserved
- Spec review loop is preserved
- User review gate is preserved

## Conversation 2 (2026-04-14): Adapt all 14 skills for parley.nvim

User: Adapt all superpowers skills. User works both on main and feature branches depending on feature size. For subagent strategy, follow AGENTS.md — remove conflicting instructions.

AI: Adapting all 14 skills with these changes applied across the board:

**Universal changes:**
- Namespace flattening: all `superpowers:` cross-references → `superpowers-` (in SKILL.md files AND supporting files)
- All paths `docs/superpowers/specs/` and `docs/superpowers/plans/` → `workshop/plans/`
- Remove Visual Companion references throughout

**Key skill-specific adaptations:**
- **brainstorming**: Existing intent preserved (remove Visual Companion, workshop/plans/, atlas update step)
- **writing-plans**: Plan output → `workshop/plans/<issue-slug>-plan.md`. Execution handoff defers to AGENTS.md Section 3 instead of mandating subagent-driven-development
- **subagent-driven-development**: Added IMPORTANT note to consult AGENTS.md Section 3 before using. Removed "REQUIRED" as default execution path. Updated "When to Use" flowchart to check if context is capturable as prompt
- **executing-plans**: Changed "use subagent-driven-development instead" to conditional guidance per AGENTS.md Section 3
- **using-superpowers**: Added parley.nvim note that AGENTS.md Section 3 governs subagent strategy, overriding skills that mandate subagent-driven-development
- **using-git-worktrees**: Removed `~/.config/superpowers/worktrees/` global path option. Added Lua/Neovim `make` detection to setup
- **requesting-code-review**: Fixed example path to `workshop/plans/`
- **finishing-a-development-branch, dispatching-parallel-agents, systematic-debugging, test-driven-development, receiving-code-review, verification-before-completion, writing-skills**: Cross-ref flattening + as-is (generic, no path-dependent content)

### Verify
- No `superpowers:` cross-references in ANY file (SKILL.md or supporting files)
- No references to `docs/superpowers/specs/` or `docs/superpowers/plans/` in any file
- subagent-driven-development defers to AGENTS.md Section 3, not mandated as default
- executing-plans conditionally suggests subagent-driven-development, not unconditionally
- using-superpowers includes parley.nvim note about AGENTS.md Section 3
- using-git-worktrees does not reference `~/.config/superpowers/worktrees/`
- writing-plans saves to `workshop/plans/`
- requesting-code-review example uses `workshop/plans/`
- All 14 skill frontmatter names use `superpowers-` prefix
- All cross-references use `superpowers-` format
