---
id: 000310
status: open
deps: []
github_issue:
created: 2026-10-09
updated: 2026-10-09
estimate_hours:
card_mirror: '097855c3a3a0bfe93096cc9878b1a06b1225dbda' # card fields mirrored from issue-cards; edit via sdlc
---

# Choose the boundary reviewer's agent per repo, and model read-only

## Problem

The boundary reviewer's agent is chosen by `ResolveAgentCLI` (`cmd/sdlc/internal/judge/agent_resolver.go:34-51`). The order is `--agent`, then `AGENT_CMD`, then `PAIR_AGENT`, then codex/claude env detection, then a claude default. That is effectively "same as the main agent".

#129 (`workshop/history/issues/000129-default-sdlc-judges-to-current-agent.md:24-33`) proposed `same`, `other` (claude reviewed by codex and vice versa) and `explicit` pairs. Only `same` was built. There is no per-repo config.

"Read-only" isn't modeled either (`dispatch.go` `BuildArgs` ~130-155):
- claude gets `--allowedTools` but also `--permission-mode bypassPermissions`, with Bash in the list;
- codex runs `exec --full-auto`;
- gemini runs `--yolo`.

Reviewers have mutated the checkout, e.g. left it detached, so close refused (evidence `a1-15e9a:1453`).

## Spec

- Per-repo config for the reviewer agent: `same` (default; the most robust, since if the main agent works, its CLI works), `other`, or `explicit` pairs. Later possibly per review category.
- A per-agent mapping to its real read-only mode, passed by the dispatcher. No `bypassPermissions`/`--full-auto`/`--yolo` for reviewers.
- Companion to #204, the isolated checkout. Isolation is the agent-agnostic way to make read-only *true*. The operator notes a fresh worktree may be heavy (deps clone plus `weave compile`), so measure the setup cost. Fallback: a detached worktree reusing the slot's built deps.

## Done when

- A repo can configure `other`. A claude-driven close is then reviewed by codex and vice versa, recorded in the ledger.
- A reviewer that attempts a write fails on read-only, not via prose.
- `sdlc close --help` documents the config.

## Plan

## Log

### 2026-10-09
Filed from the robustness evidence review (ariadne-robustness-1). Revives #129's unbuilt modes; pairs with #204.
