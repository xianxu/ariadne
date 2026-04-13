# Pitch Deck: Ariadne, the Collaborative AI Harness

**Status:** Draft v1
**Date:** 2026-04-11
**Purpose:** Seed-stage funding pitch deck outline

---

## Slide 1: Ariadne

**[Collaborative AI Harness]**

*"Life takes 42 shots."*

One-liner: "AI runs the loops. Humans steer. AI learns."

---

## Slide 2: Problem

**AI is powerful but unsustainable to use well.**

- Real problems don't get solved in one shot: human language is imprecise, domain understanding is patchy, market conditions change. They shift as you build and learn new potentials and new constraints. Workflow of knowledge worker is a living thing.
- Effective AI use requires **multi-shot to convergence** — narrowing infinite possibilities to a concrete outcome. It's a search problem with a stochastic machine.

---

## Slide 3: Insight

**LLMs are a new stochastic OS. The native application layer hasn't been built.**

- LLMs are the next platform — a stochastic OS that can execute cognitive tasks, **plausibly**. The application layer race starts now.
- The AI-native app looks fundamentally different:
  - AI runs the process loops, humans steer
  - The human role is managerial: providing both **the structure** (how loops work) and **the content** (judgment, context, choices)
  - Workflows (aka **the structure**) evolve through use — living structures, not static templates
  - And a natural way for humans to define what is **correct**, not only what's **plausible** (aka vibe coding)

---

## Slide 4: Solution — A Co-Designed Harness

**The gap isn't intelligence. It's the ability to steer, sustainably.**

Think AlphaZero — it works because it learns a board evaluation function through self-play. In knowledge work, there aren't enough training examples. **The human is the evaluation function.** Our job is to make that sustainable.

Ariadne unifies what today's tools keep separate — the AI loop, the domain state, and the workflow knowledge — into one co-designable environment:

- **Repo as world state.** Every "shot" is a state change, not a chat reply. Every human edit is an evaluation signal.
- **Structured checkpoints.** Surface decisions where human judgment matters most, at the right altitude.
- **Living workflow.** Workflow patterns the human can inspect, edit, and evolve. The system adapts through use; the human architects the structure.
- **Persistent, editable history.** Transparent and replayable. Change a past decision, re-converge from there.

---

## Slide 5: What We've Built

**The founder has been building and using the harness daily.**

- A working Nvim-based environment where AI runs tool loops, the repo holds all state, and the human steers through editable transcripts, living workflows, branching conversations, and persistent memory.
- The entire workflow — brainstorming, architecture, coding, issue tracking, spec writing — runs through the harness. This pitch deck was created in it.
- Key insight from dogfooding: the leverage isn't in any single AI response. It's in the accumulation — skills that remember how you work, state that persists across sessions, and the ability to re-converge when requirements change.

*(Demo: live walkthrough of a multi-shot convergence loop — from vague goal to concrete implementation, showing human steering at each checkpoint)*

---

## Slide 6: Roadmap — From Nvim to Universal

**Phase 1: Developer harness (now)**

- An opinionated Neovim-based environment where repo is the world state, markdown manages workflow, and AI loops are steered through skills, issues, and specs — all co-designed by the founder.
- Prove the thesis by building the company itself with the harness. If it can't do that, it can't do anything.
- This will be usable for other tech founders.

**Phase 2: Standalone cockpit**

- Same convergence engine, same living workflow — but multiple interfaces beyond the Neovim. IDEs, Web, desktop, collaborative.
- Target: technical founders and small dev teams who want the paradigm without living in Nvim.

**Phase 3: Universal harness**

- The convergence engine becomes a platform. Domain-specific skill ecosystems: legal, consulting, product, operations.
- Other companies build on top. The accumulated knowledge of **how humans steer AI to convergence** becomes the moat.

---

## Slide 7: Why Now

- **Capability just arrived.** LLMs can now reliably execute multi-step cognitive tasks, as evident in coding agents' capabilities. Two years ago they couldn't. The stochastic OS just booted up.
- **Tool use is maturing.** Models call tools, execute code, interact with APIs — they're "loop executors" now, not just text generators.
- **The economics are already there.** If an AI loop can do the work, it's cheaper than a human. The question was never cost — it was capability. Capability just arrived.
- **The window is open.** The industry is in its "faster horse" phase. Once someone demonstrates the native paradigm, the shift will be obvious in hindsight.

---

## Slide 8: Market

- **Wedge:** Programmers and technical founders. They feel the pain most, can evaluate the product, and software is the most mature domain for AI tool use.
- **Expansion:** All knowledge work — consulting, legal, product, operations. Any domain where complex multi-step work requires iterative refinement and human judgment.
- **This is not a tool market — it's a platform shift.** We won't capture the whole market, but a slice would have handsome reward.

---

## Slide 9: Competition

**vs. OpenClaw** (open-source AI agent, 247K GitHub stars):
OpenClaw is a task executor — "do this for me." It automates well-defined tasks (send email, run test, manage files). The collaborative AI harness solves a different problem: convergence on ill-defined outcomes when you **don't** know exactly what you want. OpenClaw is the faster horse. We're building cars.

**vs. Copilots (GitHub Copilot, Cursor, etc.):**
Code-level assistants bolted onto existing IDE workflows. They help humans execute faster. We restructure the loop so AI executes and humans steer.

**vs. Agent frameworks (LangChain, CrewAI, etc.):**
Infrastructure for developers building AI apps. We're building the end-user application — the cockpit — not the plumbing.

**vs. Claude Code / Codex CLI:**
Operate at the code plane. We operate at the decision and convergence plane — a higher knowledge level. We heavily leverage coding plane agents.

---

## Slide 10: Business Model

TKTK

---

## Slide 11: Team

- **[Founder Name]** — Engineering leadership at Meta, Twitter, Microsoft. Builder and manager at scale.
- Dual perspective that produces the insight: deep understanding of both the technical capability of AI *and* the organizational processes it replaces — and the technical details of *how* to replace them. What visibility and safeguards are needed. How to create virtuous loops where human oversight improves AI execution, which earns more autonomy.
- Already building and using the harness daily. Not theorizing — operating.

---

## Slide 12: The Ask

- Raising $1M seed round.
- Use of funds: keying founding team.
- Goal: demonstrate the paradigm with paying dev teams within 1 year.

---

## Key Thesis Summary

1. LLMs are a new stochastic OS
2. Frontier lab handles AI intelligence — we make it work in knowledge economy, sustainably
3. The native application layer for this OS hasn't been built
4. We're building it: a knowledge OS with living, evolving workflows
5. This changes not just productivity but how organizations are structured
6. Start with dev teams, expand to all knowledge work, become the platform
