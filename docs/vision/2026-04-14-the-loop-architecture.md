# The Loop Architecture

**Date:** 2026-04-14
**Status:** Pensive — connecting the dots

---

A realization that's been building: everything I've been tinkering with — parley, ariadne, construct, the writing style guide, the founder letter — they're all instances of the same primitive:

**prompt + context → artifact**

The difference is just what goes in and what comes out. And the system is the set of loops that connect these transformations, with memory observing from the background.

## The Primitive

Every AI interaction is a transformation. You give it some context (prior chats, documents, code, specs), apply a prompt (which may be a reusable skill), and get an artifact (a document, code, an issue, a memory). The artifact becomes context for the next transformation. That's the loop.

The human's role: steer which transformations happen, verify the outputs are correct (not just plausible), and evolve the prompts/skills over time.

## The Loops

### 1. Thinking Loop
**parley chats → thought documents**

Chat is the reasoning process. You explore, brainstorm, refine. At some point, the thinking crystallizes into a document — the "fruit" of the chat tree. The document carries lineage: which chats produced it, what the reasoning path was.

Example from today: a conversation about NexHealth's AI transformation → the founder letter → the companion doc on what AI-first means.

### 2. Voice Loop
**thought document + style guide → personalized document**

Raw thinking gets adapted to voice. The style guide (itself an artifact of observing past writing) is the prompt context. The transformation: take this draft, make it sound like me.

Example from today: the founder letter drafted collaboratively, then rewritten using the style guide extracted from blog posts.

### 3. Building Loop
**parley chats → issues → specs → code**

The most concrete loop. Thinking produces issues. Issues produce specs and plans. Specs drive AI execution (claude code). Code gets verified by tests. Results flow back as updated specs and new issues.

Example: parley.nvim's entire development — 700+ commits through this loop.

### 4. Memory Loop
**past chats, docs, artifacts → memory files**

A background loop that observes the others. What did we decide? What's the user's role? What feedback was given? Memory persists across sessions, so future loops have richer context.

This is the loop that compounds. Each session makes the next session better.

### 5. Skill Evolution Loop
**usage patterns + feedback → updated skills**

Meta-loop. Observing how the other loops work (or don't), and evolving the prompts/skills that drive them. The Construct is the tool for this loop — adapt skills based on accumulated intent, verify the adaptations, deploy.

Example from today: adapting superpowers skills for parley's conventions, namespace flattening, the whole construct design.

## The Substrate

All loops run on the same substrate:

- **Markdown files in repos** — the universal state format. Chats, docs, issues, specs, skills, memory — all markdown, all version-controlled, all human-readable.
- **Two interfaces:**
  - **Parley** — the thinking interface. Chat transcripts, tree-of-conversations, review markers, document lineage. Where you reason.
  - **Claude Code** — the acting interface. Tool calls, file manipulation, code generation, agentic execution. Where things get built.
- **The repo as world state.** Parley and Claude Code share state through the filesystem. Links in markdown are the connections. No special plumbing needed — `ls`, `cat`, `git` work on everything.
- **Skills** — reusable prompt patterns that encode how each loop's transformations work. Managed by the Construct, adapted per-repo, evolved through use.
- **Memory** — persistent observations that enrich context across sessions.

## What Connects Them

The loops aren't separate systems. They feed each other:

```
Memory observes all loops
     ↓ enriches context for
Thinking (parley) → produces → Thought docs
     ↓ feeds                        ↓ personalized by
Building (claude code)          Voice loop (style + skill)
     ↓ produces                     ↓ produces
Code, specs, issues             Letters, blog posts, pitches
     ↓ observed by                  ↓ observed by
Memory                          Memory
     ↓ informs                      ↓ informs
Skill evolution                 Skill evolution
     ↓ improves                     ↓ improves
All loops                       All loops
```

The links between loops are just references in markdown. A thought doc links to the chat that produced it (frontmatter lineage). An issue links to the spec. A spec links to the tests. Memory links to the conversations it observed. Skills link to the intents that shaped them.

No special infrastructure. Just files, links, and two interfaces that know how to work with them.

## The Insight About Parley

Parley doesn't need to be an execution engine. It doesn't need to replicate Claude Code's tool-calling. Parley is where you **think** — the transcript interface. Claude Code is where you **act** — the execution engine. The repo is the shared state between them.

What parley does need:
- **Chat-to-document lineage** — connect the reasoning to its fruits
- **Review markers** — lightweight verification at the document level
- **Tree navigation** — explore and branch the reasoning process
- **Notes/documents as first-class** — not just chats, but the artifacts they produce, linked back to their source

What Claude Code does:
- **Execute transformations** — write code, manipulate files, run tests
- **Manage skills** — the Construct, adapt/promote/upgrade
- **Ground in reality** — tests, linting, verification

Two tools, one shared filesystem, markdown as the universal format.

## What This Means for Ariadne

Ariadne is the **skill workbench** — the place where the skill evolution loop runs. It doesn't produce code or documents directly. It produces the *machinery* that makes all the other loops work better: adapted skills, workflow configurations, bootstrapping templates.

The three repos, clean roles:
- **Ariadne** — the brain. Skills, construct, memory, personal configuration.
- **Parley** — the mouth/ear. Chat interface, transcript editing, document lineage.
- **Target repos** (parley.nvim, nexhealth, future projects) — the hands. Where artifacts get built.

## The DAG of Prompts

Zooming all the way out: the entire system is a **DAG of prompt-driven transformations**. Each node takes inputs (context + prompt) and produces outputs (artifacts). The edges are the references between artifacts. Memory and skills are the nodes that influence many other nodes — they're the high-fan-out vertices in the graph.

A mind map on steroids. Except the map is executable, the nodes produce real artifacts, and the whole thing evolves through use.

## The Substrate Layer

The loops don't run themselves. There's a substrate that manages the transitions between them — the plumbing that makes each loop's transformations discoverable, repeatable, and evolvable.

**Construct** manages the skill evolution loop — how prompt patterns get adapted, versioned, and deployed. But it's really managing a piece of every loop, because skills *are* the encoded transformations that drive each loop.

**Parley** manages the thinking and voice loops — how reasoning flows through chats, how documents emerge as fruits, how review markers drive refinement. Its chat tree, document lineage, and review system are the substrate for the thinking side.

Together, they're not applications in the traditional sense. They're **substrate for managing a DAG of AI-driven transformations.** The loops are the user-visible workflows. The substrate is what makes those workflows inspectable, editable, and improvable — not a black box, but a transparent, co-designable system.
