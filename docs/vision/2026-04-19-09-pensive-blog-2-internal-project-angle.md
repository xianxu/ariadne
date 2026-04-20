# Pensive: Blog 2 — The Internal Project as Proof Point

**Date:** 2026-04-19
**Status:** Thinking out loud
**Series:** AI Workflow Blog Series (2 of 3)
**Related:** [Blog 2 pensive](2026-04-16-06-pensive-blog-2-super-knowledge-worker.md), [Blog 1 published](tale-of-two-harness)

---

## The New Material: An Internal Project

Started an internal project that follows the "single filesystem" pattern from post 1, but applied to an operational context — not just personal thinking, but a system that bridges knowledge work and production.

### The Three Layers

1. **Mirror and clean.** Sync ticketing systems (Jira, Intercom) into markdown files in directories. Raw data becomes discoverable, searchable, connectable. The filesystem becomes the single source of truth that no individual SaaS tool provides.

2. **Claude Code as extended REPL.** Use AI to generate all kinds of artifacts:
   - Sync code (full and incremental) for those ticketing systems
   - Daily digests based on new changes — for human consumption, but also as a "last update index" for the agent itself
   - Correlations between tickets and code
   - One-off analytics that can be ad-hoc, then lifted into scripts, then into testable subsystems
   - All the intermediate artifacts (specs, tests, code) that lead to each piece

3. **Incremental architecture.** The codebase builds up organically — forming boundaries, modules, structure — designed to be legible to both humans and machines.

### The Virtuous Cycle

The pattern: maintain a set of artifacts tied together by AI, or by script/code written by AI. Each layer builds on the one below it. Ad-hoc query → script → tested subsystem → architectural boundary.

### The Tricky Part

How do you organize this "hierarchy of artifacts"? And how do you create the right "tests" at each level — tests that both:
- Gain confidence in each transformation (does the sync actually capture what matters?)
- Catch issues that matter, when they matter (not too early, not too late)

This is the open design question.

---

## The Angle Problem for Post 2

The existing pensive frames post 2 as "the super knowledge worker" — personal workflow, exoskeleton metaphor, day-in-the-life. But the internal project complicates this in an interesting way:

**The knowledge worker's output eventually needs to become production system.**

The boundary between "thinking/exploring" and "operating/producing" isn't clean. The internal project shows the continuum:
- Start with ad-hoc exploration (knowledge work)
- Artifacts accumulate and harden (the in-between)
- Eventually they become reliable, tested, scheduled systems (production)

**Possible angles for post 2:**

- **Angle A: Pure knowledge worker story.** Keep post 2 personal and aspirational. The internal project becomes a brief example. Production concerns go to post 3.

- **Angle B: The artifact hierarchy.** Post 2 is about how knowledge work *hardens into* production systems. The continuum from ad-hoc → script → subsystem → architecture. The key insight: the shared filesystem pattern doesn't just help you think — it gives your thinking a path to becoming real.

- **Angle C: Both, sequenced.** First half is the personal workflow (exoskeleton, day-in-the-life). Second half introduces the internal project as "what happens when you apply this pattern to a team problem" — the bridge from knowledge work to production.

**Unresolved:** Is the audience for post 2 still developers? Or is it broadening to "anyone who does knowledge work"? The internal project material pulls it back toward a technical audience. The exoskeleton metaphor pushes it toward a broader one.
