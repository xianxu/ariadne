# Pensive: AI Workspace for the Masses

**Date:** 2026-04-16
**Status:** Thinking out loud
**Related:** [Exoskeletons for Knowledge Workers](2026-04-16-01-pensive-exoskeletons-for-knowledge-workers.md), [The Loop Architecture](2026-04-14-02-the-loop-architecture.md)

---

## The Gap

The patterns I've built (artifact management, context centralization, many-shot convergence, skill compounding, inline review) are universal. A scientist iterating on a paper, a lawyer reviewing contracts, a consultant building a deliverable: they all do the same thing. Iterate on documents, accumulate context, build on prior work.

The barrier is the interface, not the concept. Neovim + Claude Code + markdown repos is the bleeding edge for people who think in terminals. The "AI architect" persona. To bring this to everyone else, the concepts need to shed the tooling.

## What It Looks Like

**The workspace is a folder, not an app.** Like Obsidian but AI-native. Documents, notes, references, drafts all live in one folder you can see and touch. No database. No cloud-only lock-in. The folder IS the state. Git runs underneath for history, but the user never types `git`.

**The thinking space is a conversation sidebar.** Not a chatbot popup. A persistent, branchable conversation tied to whatever document you're working on. A scientist researching a paper opens a conversation, explores three hypotheses, branches at each one, and the winning branch becomes a draft section. The conversation history stays linked to the document it produced (lineage).

**Inline review is native.** Highlight a paragraph, type a comment. The AI revises. If it has a question, it appears as an inline annotation you respond to. Looks like Google Docs comments but the other party is AI, and it actually rewrites the text. No special syntax.

**Skills are "workflows" with friendly names.** A lawyer doesn't create a "skill." They say "every time I review a contract, check these 12 clauses and flag risks in this format." The system remembers. Next contract, one click. A scientist says "when I read a paper, extract methods, key findings, and open questions into my notes template." That becomes a workflow. Over time, these compound.

**The critic is a button.** Finish a draft, press "Critique." AI reads it in the context of everything else in your workspace (your prior papers, your style, your field's conventions) and leaves inline comments. Address the ones you care about. Same 🤖-> / ㊷[] loop, no terminal.

**Context proximity is automatic.** When working on a grant proposal, the system knows your CV, prior publications, lab data, and funder guidelines are all relevant. Not because you configured MCP bridges to Google Drive and Zotero and the university grants portal. Because they're all in the folder. A simple import step ("bring this PDF into my workspace") converts external artifacts into local, AI-readable format.

## The Hard Part

Not the features. The behavior change. Moving from "I keep my stuff in 10 different apps" to "I keep my stuff in one place." The technology is straightforward. The habit is not.

Every SaaS tool has a gravity well. People don't use Notion because it's the best place for docs. They use it because their docs are already there. Breaking that gravity requires a workflow that's so obviously better it justifies the migration cost.

## The Wedge

Not "here's a general-purpose AI workspace." That pitch is too abstract. Instead: one specific workflow, fully realized, that's so obviously better people adopt the workspace to get it.

Candidates:
- "The best way to write a research paper" (scientists)
- "The best way to prepare for a deposition" (lawyers)
- "The best way to build a client deliverable" (consultants)
- "The best way to do due diligence" (investors)

One workflow. Fully realized. Then they discover the compounding.

## Open Questions

- Which wedge workflow has the most pain and the shortest time-to-value?
- How much of the "folder as workspace" can coexist with existing tools vs. requiring full migration?
- Is the product a desktop app, a web app, or an Obsidian/VS Code plugin that graduates into its own thing?
- Can the skill/workflow system be genuinely self-service for non-technical users, or does it always need a power-user to set up?
