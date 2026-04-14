---
scope: repo
target: ../parley.nvim
---

# Intent: superpowers/brainstorming → parley.nvim

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
