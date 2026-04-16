# Pensive: The Duo Workflow — Parley and Claude Code

**Date:** 2026-04-14
**Status:** Thinking out loud

---

## Two Modes of Thinking

Parley and Claude Code are not competing tools. They're two fundamentally different modes of working with AI, and trying to make one do the other's job is a mistake I almost made.

**Parley is divergent.** You don't know where you're going when you start. The name was chosen deliberately — a parley is a conversation, a negotiation. You explore, branch, change your mind, revisit. The transcript is the artifact. The value is in the wandering. You might start researching a legal question and end up brainstorming a company pitch. That's not a bug.

**Claude Code is convergent.** You have a task. Spec → plan → execute → verify → done. The repo is the artifact, not the conversation. The conversation is disposable — a means to an end. You don't go back and edit what the agent said three turns ago. You move forward.

This maps to a well-known distinction in creative thinking: divergent thinking (generate possibilities) and convergent thinking (narrow to a solution). Both are necessary. They need different interfaces.

## What Each Is Good At

**Parley's strengths:**
- Editable transcripts — rewrite the past, re-derive from there
- Tree-of-conversations — branch and explore multiple paths
- Multi-model — bounce between Claude, ChatGPT, Gemini
- Non-coding use cases — research, writing, legal review, brainstorming
- Chat as durable artifact — publish it, link to it, curate it
- Review markers (㊷[]) — inline annotation-driven editing
- Long-form document engagement — 100-page legal docs, research threads with dozens of questions

**Claude Code's strengths:**
- Agentic execution — tool calls, file manipulation, bash, tests
- Repo as world state — understands project structure, AGENTS.md, skills
- Structured workflows — skills, hooks, plans, issues
- Multi-file changes — refactor across a codebase
- Verification — run tests, check output, iterate
- Memory system — persistent across sessions

## The Natural Workflow

The duo workflow in practice:

1. **Think in parley.** Explore the problem space. Brainstorm. Research. The chat tree branches as you explore different angles. At some point, thinking crystallizes.

2. **Fruit drops.** A document, a spec, an issue, a letter. Created within the chat, linked back to its lineage. This is the bridge between divergent and convergent.

3. **Act in claude code.** Take the fruit — the spec, the issue — and execute. Claude code operates on the repo, writes code, runs tests, manages skills. It doesn't need to know about the thinking process. It just needs the artifact.

4. **Results flow back.** Updated documents, code, test results. These become context for the next parley session. The loop continues.

The repo filesystem is the shared state. No special integration needed — both tools read and write markdown files. Links in documents are the connections.

## Where the Boundary Sits

Pretty clear after today's session:

| | Parley | Claude Code |
|---|---|---|
| **Purpose** | Think, explore, reason | Execute, build, verify |
| **Conversation** | The artifact | Disposable means |
| **State** | Chat tree, transcripts | Repo files |
| **Editing** | Go back, branch, revise | Move forward |
| **Models** | Any (Claude, GPT, Gemini) | Claude only |
| **Best for** | Research, writing, brainstorming, review | Coding, refactoring, skill management, deployment |
| **Human role** | Explorer, editor, curator | Architect, verifier, steerer |

## What NOT to Build

Temptation: make parley handle tool calls and agentic execution. I've been down this road — parley has some tool-calling capability. But investing more here competes with claude code, which is better at it and improving fast.

Temptation: make claude code have editable persistent transcripts. Maybe Anthropic does this eventually. But it's not what claude code is *for*. The ephemeral conversation is a feature — it keeps focus on the task, not the history.

Instead: invest in what's unique to each, and make the handoff clean.

**Parley investments:**
- Chat-to-document lineage (issue 000104) — connect thinking to its fruits
- Visual feedback for review (issue 000105) — better ㊷[] experience
- Notes/documents as first-class alongside chats
- Tree navigation improvements

**Claude Code investments (via ariadne/construct):**
- Skill adaptation and management
- Workflow bootstrapping for new repos
- Verification infrastructure
- Memory system

**Shared investments:**
- The writing style guide — used by both (parley for voice, claude code for document generation)
- Repo conventions (AGENTS.md, workshop/ layout) — consumed by claude code, evolved through parley thinking

## The Pitch Deck Connection

The Ariadne pitch deck talks about "multi-shot convergence." That's exactly this duo:
- Parley is the multi-shot part — exploring, iterating, branching
- Claude Code is the convergence part — narrowing to a concrete outcome

The harness that connects them is the repo: markdown files, skills, specs, issues. Living state that both tools can read and write. The human steers the transitions — when to stop exploring and start building, when to step back from building and rethink.

Two tools, one filesystem, human at the wheel.
