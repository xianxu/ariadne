---
id: 000011
status: done
deps: []
created: 2026-04-27
updated: 2026-06-03
actual_hours: 0.2
---

# update AGENTS.md to explain the .claude/skills system

The goal is to make ariadne based repo working for all type of agents, not just `claude`. the only requirement is both the local harness and model support good tool call and long reasoning chain. 

## Done when

- AGENTS.md points at `.claude/skills/` as where this repo's agentskills.io
  skills live, and frames agent-agnosticism honestly (shared AGENTS.md + the
  shell-invokable `sdlc` binary, not skill discovery).

## Spec

Scope reduced from the original "explain the whole `.claude/skills` system." The
goal — work for all agents — was reached by a different route since this was
filed: (1) `AGENTS.md` is natively the cross-agent constitution (`CLAUDE.md` is
just `@AGENTS.md`; codex/OpenAI agents read `AGENTS.md` by their own convention);
(2) the workflow moved out of skill prose into the `sdlc` binary any agent shells
out to (`xx-sdlc` is a static pointer to `sdlc --help`); (3) the binary is
agent-pluggable (`dispatch.go` claude|codex|gemini via `--agent`/`$AGENT_CMD`).
So the remaining useful deliverable is just a §11 pointer: say where the skills
live and that following the SDLC doesn't depend on the claude-specific skill path.

## Plan

- [x] AGENTS.md §11 — note `.claude/skills/` holds this repo's agentskills.io
  skills (`xx-*` + `construct/adapted/`), and that agent-agnosticism rests on the
  shared `AGENTS.md` + the `sdlc` binary, not on skill discovery.

## Log

### 2026-06-03 — closed (scope-reduced)
- 2026-06-03: closed — AGENTS.md §11 now points at .claude/skills/ as where this repo agentskills.io skills live + frames agent-agnosticism on shared AGENTS.md + the sdlc binary (not skill discovery). Scope reduced: workflow already binary-first + AGENTS.md natively cross-agent. doc-only; --no-judge (2-line clarification), --no-atlas (no new surface).; review verdict: not-run
- Reviewed with operator. The original "document the whole `.claude/skills`
  system for agent-agnosticism" was overtaken by the binary-first architecture
  (workflow in `sdlc`, shared `AGENTS.md`) — codex/etc. already follow the
  constitution because they read `AGENTS.md` natively. Shipped the minimal
  remaining value: a §11 pointer (where skills live + the honest agent-agnostic
  framing). **Residual not pursued here:** skills carrying *behavior* (the `xx-*`
  ones) are still claude-namespace discovery; whether a non-claude agent needs
  that reach is a separate, narrower question — file it fresh if it becomes live.

### 2026-04-27

