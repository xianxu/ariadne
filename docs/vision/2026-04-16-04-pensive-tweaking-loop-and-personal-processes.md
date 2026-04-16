# Pensive: The Tweaking Loop and Personal Processes

**Date:** 2026-04-16
**Status:** Thinking out loud
**Related:** [Own Harness vs Notion](2026-04-16-03-pensive-own-harness-vs-notion.md), [AI Workspace for the Masses](2026-04-16-02-pensive-ai-workspace-for-the-masses.md)

---

## The Tweaking Loop as Differentiator

The key differentiator isn't features. It's the interaction model.

Notion's AI gives you a blob of text: accept or start over. ChatGPT gives you a response: copy-paste or re-prompt. Our system lets you write `㊷[too formal here]` and the AI adjusts that paragraph. Write `㊷[this doesn't sound like me]` and it rewrites using your style guide. That's collaborative editing with AI, not AI-generates-human-accepts.

This loop is hard to replicate in a chat interface. It needs the document as the primary surface, with AI operating surgically, not wholesale. The document is the workspace. The AI is a collaborator you can direct at the paragraph level.

The higher the stakes, the more this matters. Nobody cares about tweaking a Slack message. Everyone cares about a blog post with their name on it, a legal brief, a research paper, a pitch deck. The users who value precise control over AI output are the users with high-stakes writing. That's a real market.

## Harness Visibility for Non-Coders

There's a design tension. Too visible (YAML files, terminal commands) and non-coders flee. Too invisible (black box AI) and you're Notion. The sweet spot:

- Workflows visible as cards, not files. "Interview Feedback" with a description and an edit button.
- Editing a workflow is guided. "What triggers this?" "What's the output format?" "Show me an example." Not raw YAML.
- Execution is transparent. When AI runs, you see what it read, what it wrote, what it decided. Human-readable build log, not a terminal.
- The folder is browsable. Finder-like, not `ls`. But it IS the filesystem underneath, so power users can drop to terminal anytime.

## Personal Processes: The Self-Molding Layer

Beyond document editing, there's a whole layer around organizing life and work. Morning summaries, daily goals, task tracking, project management. These feel like "stock" features with minor adjustments per user.

But here's the twist: those processes themselves can be user-authored. Not "here's our task management feature, configure these 5 settings." Instead: the user creates their own morning ritual as a process the AI follows.

Examples:
- "Every morning, scan my open issues, calendar, and recent notes. Give me a summary of what needs attention and suggest three goals for the day."
- "At end of week, review what I accomplished vs. what I planned. Highlight what slipped and why."
- "When I finish a document, run my personal review checklist: clarity, evidence for claims, tone, length."

These are personal workflows, created in the same system that creates document-editing skills. The user says what they want, the system remembers, and it runs on schedule or on trigger. Over time, these compound too. Your morning summary gets better because it knows your patterns. Your review checklist evolves as you add new checks.

The insight: project/task management isn't a separate feature. It's just another set of personal processes running on the same skill/workflow engine. No separate "project management module." Just processes the user defined, operating on artifacts in the folder.

## The Compounding Stack

Everything is the same primitive:
1. **Document-level**: ㊷[] tweaking loop. AI operates on your text surgically.
2. **Workflow-level**: Skills/workflows for recurring tasks. Interview feedback, paper reviews, contract analysis.
3. **Process-level**: Personal rituals. Morning summaries, weekly reviews, goal tracking.
4. **Meta-level**: The system that helps you create and evolve all of the above.

All four levels run on the same engine: prompt + context -> artifact, with human steering. The difference is just the scope and the trigger.

## Open Questions

- How do scheduled/triggered processes work in a local-first system? Cron-like? Or does the user just invoke them?
- Is there a library of starter processes (like app store for workflows) or does everyone build from scratch?
- How much of the task/project management space is "good enough" from existing tools that people won't switch for it alone? (Probably a lot. It's a complement to the tweaking loop, not the wedge.)
