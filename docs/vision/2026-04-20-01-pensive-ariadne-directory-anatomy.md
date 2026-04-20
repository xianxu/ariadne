# Pensive: Ariadne Directory Anatomy

**Date:** 2026-04-20
**Status:** Thinking out loud

---

Ariadne is evolving into a command center — my extended brain. The repo isn't just code; it's a structured workspace where different kinds of artifacts live in purpose-built directories. Here's how the anatomy is shaping up:

**atlas/** — The map of where things are. Not just code, but all kinds of artifacts across the workspace. Mostly automatically maintained by parley, with minimal human touch. Think of it as the index that keeps the territory navigable as it grows.

**brain/** — The new addition. Memory artifacts I actually care about. The things worth remembering across sessions and contexts — the durable knowledge layer.

**construct/** — The meta-layer: how AI works in this repo. Skills, their evolution, the machinery of adaptation. This is where the harness itself gets tuned. The `/construct` skill manages importing, adapting, and versioning skills through an intent-transcript architecture that survives upstream upgrades.

**docs/** — Brainstorms, higher-quality documents, co-authored artifacts. The human sets direction on what to create; AI helps shape it. Pensives live here in `docs/vision/`. These are fundamentally human-directed.

**skills/** — This one's unclear and may be redundant. Currently holds two custom skills (`xx-voice-apply`, `xx-voice-gen`), but the active deployed skills live in `.claude/skills/`. These probably belong either in `construct/` (as source material for the skill pipeline) or directly in `.claude/skills/`. Worth resolving — having skills in two places creates confusion about which is canonical.

**vision/** — Strategic planning and roadmapping. The initial form lived in `../parley.nvim/vision/` as a YAML-based system for tracking initiatives, dependencies, team capacity, and timelines. It exports to CSV and Graphviz, does capacity-aware scheduling, and propagates priority through dependency graphs. This subsystem is about the 1-2 year horizon: what are the big rocks and how do we sequence toward them.

**workshop/** — The working directory for coding up this repo. Active development, issue tracking, implementation plans. The forge where things get built.

## The Bigger Picture

Ariadne isn't just self-contained. The intention is hermetic-from-clone — settings in `.claude/settings.json`, not `.claude/settings.local.json` — but also reaching outward. Sibling directories like `../parley.nvim` are already in scope (sandbox write access configured). The trajectory is toward ariadne as the hub that drives surrounding repos, a central nervous system for a constellation of tools.

## Open Questions

- **skills/ vs construct/ vs .claude/skills/**: Three places where skill-related things live. What's the clean separation? Is `skills/` just a staging area, or does it have its own identity?
- **brain/ vs memory/**: How does `brain/` relate to the auto-memory system in `.claude/projects/.../memory/`? One is for the human, one is for the AI? Or should they converge?
- **vision/ migration**: Vision started in parley.nvim. Does it fully move into ariadne, or does it stay as a parley feature that ariadne consumes?
