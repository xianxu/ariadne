# Founding Context

**Date:** 2026-04-11

## Core Thesis

LLMs are a new stochastic OS. The native application layer hasn't been built. Everyone is building faster horses — bolting chatbots onto existing tools. The real opportunity is building the car: a knowledge OS where AI runs process loops and humans steer.

## Key Insights

1. **Multi-shot convergence is the real interaction model.** One-shot doesn't work for real problems. Human language is imprecise, domain understanding is patchy, requirements shift as you learn.

2. **Human is the evaluation function.** Like AlphaZero needs a board evaluation function, knowledge work needs human judgment to distinguish correct from merely plausible. The system must make this role sustainable.

3. **Structure and content are both human-provided.** The human provides the structure (how AI loops work — living skills/workflows) and the content (domain judgment, context, choices). Neither AI nor human evolves the structure alone.

4. **Co-designability is the key.** Today's tools keep the agentic loop, domain state, and workflow knowledge in separate planes. They must be unified into one co-designable environment.

5. **This changes organizations, not just productivity.** The widespread usage of tools like this will change how companies are structured — replacing layers of coordination, delegation, and review with AI loops steered by humans.

## Product Progression

- **Phase 1:** Nvim-based developer harness (bootstrap). Repo as world state, markdown manages workflow, AI loops steered through skills/issues/specs. Dogfood by building the company with the harness.
- **Phase 2:** Standalone cockpit. Multiple interfaces (IDE, web, desktop). Target: technical founders, small dev teams.
- **Phase 3:** Universal harness / platform. Domain-specific skill ecosystems. Other companies build on top.

## Bootstrapping Strategy

1. Fork parley.nvim (closed source) as the starting point
2. Use and upgrade parley enough to create the product
3. Use the product to create the product (recursive bootstrapping)
4. If the harness can't accelerate building its own company, the thesis is wrong

## 3-Month Focus (April–July 2026)

1. **Skill system** — living skills that evolve through usage. Demonstrate before/after.
2. **One non-coding use case** — prove it's not just a dev tool (fundraising, market research, writing).
3. **One external user** — one technical founder using it and giving feedback.

## Founder Background

Engineering leadership at Meta, Twitter, Microsoft. Builder and manager at scale. Dual perspective: deep understanding of both technical AI capability and organizational processes — and the technical details of how to replace those processes with AI loops, including visibility, safeguards, and virtuous feedback loops.

## Lineage

Parley.nvim (~25K lines Lua, 60+ modules) is the bootstrapping tool. Key capabilities carried forward:
- Markdown-as-state (transparency, editability, version control)
- Tree-of-conversations (branching exploration)
- Client-side tool orchestration
- Memory and summarization
- Exchange model architecture (clean single-source-of-truth for state)
- Repo mode (marker-file detection, auto-directories)
- Issue-driven development with constitution hooks

Parley is NOT part of the public pitch. It's the personal bootstrapping tool.
