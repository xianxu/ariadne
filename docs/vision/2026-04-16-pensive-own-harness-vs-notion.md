# Pensive: Own Harness, Individual-First, vs. Notion

**Date:** 2026-04-16
**Status:** Thinking out loud
**Related:** [AI Workspace for the Masses](2026-04-16-pensive-ai-workspace-for-the-masses.md), [The Duo Workflow](2026-04-14-pensive-the-duo-workflow.md)

---

## The Duo Collapses to One

For coders, the duo harness (Parley for thinking, Claude Code for building) makes sense. Two modes of work, two tools, shared filesystem. But for non-coding knowledge workers (scientists, lawyers, consultants), there's no "building" step that needs a separate execution engine. Their entire workflow is thinking, drafting, refining, and publishing. One harness. The thinking space and the artifact space are the same thing.

This means the product for non-coders isn't "Parley + something." It's one integrated environment that does thinking, drafting, reviewing, and skill/workflow management in a single surface. The conversation IS the workspace.

## Why Not Notion?

Notion wants to be this harness. So does Coda, Obsidian, and probably every productivity tool with an AI feature roadmap. The question: what do they get wrong?

**Notion's problem is architectural.** Your data lives in their database. Their AI features operate on their data format, through their APIs, on their servers. You can't bring your own model. You can't run workflows locally. You can't compose tools the way you can with files. And critically: your context is locked inside Notion's silo. If your research notes are in Notion, your code is in GitHub, your papers are in Overleaf, and your data is in a Jupyter notebook, Notion's AI sees one slice. MCP bridges are the industry's answer, but they're lossy and slow.

**The real gap is ownership and composability.** Notion gives you AI-on-their-terms. The opportunity is AI-on-your-terms: your files, your models (or a choice of models), your workflows that you can inspect, edit, and evolve. The folder-as-workspace model wins here because it's fundamentally open. Git for history. Markdown for state. Any AI model can read files.

**But Notion has distribution.** Millions of users already have their stuff there. The gravity well problem. You don't beat Notion by being slightly better at AI. You beat it by being fundamentally different in a way that matters to a specific user who hits the ceiling.

## Individual-First

Start with the individual, not the team. One person organizing their thoughts, accumulating context, building workflows that compound. The product is a personal knowledge workspace.

Sharing comes later, and it's simple: publish artifacts to a shared repo. A research group shares a repo of papers, notes, and workflows. A law firm shares contract review templates. The sharing primitive is the filesystem, not a collaboration platform. Git handles conflicts. Published artifacts are just files others can read.

This is deliberately the opposite of Notion's approach (collaboration-first, individual-second). Notion optimized for teams sharing a wiki. We optimize for one person thinking clearly, then selectively sharing the fruits.

## The Wedge Sharpens

Individual user. Specific workflow. The product is not "AI Notion." It's:

"The best environment for a scientist to go from reading papers to publishing one."

Or: "The best environment for a lawyer to go from case files to a brief."

One user. One workflow. Everything they need in one place. The compounding kicks in after week one.

## Open Questions

- How do you acquire individual users without enterprise sales? Content? Community? A free tier that's genuinely useful?
- How much of Parley's architecture (Lua, Neovim-native) is reusable vs. needs rewriting for a standalone app?
- Is Electron/Tauri the right shell, or is web-first better for distribution?
- Can you offer model choice (Claude, GPT, Gemini, local models) as a differentiator against Notion's locked-in AI?
