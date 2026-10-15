---
id: 000367
status: working
deps: []
github_issue:
created: 2026-10-14
updated: 2026-10-14
estimate_hours:
card_mirror: '8b088f0b6decee523ca05cc6e74ac83dcd190f4a' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-14T20:40:03-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "9f75bd42", done: "66bf61b7"}
---

# Commits record the coding agent, model and effort in one standard trailer (today: 6+ inconsistent Co-Authored-By spellings, many commits none)

## Problem

The operator wants token velocity for planning: which agent/model did how much work (2026-10-14). Commits are the natural record, but over the last 3 days of origin/main: pair has 506 `Claude Opus 5.5`, 290 `GPT-6`, 14 `Codex`, 10 `GPT-6 Codex`, 9 `OpenAI Codex` trailers and 1,718 commits with none (merges, tracker and sdlc-generated commits, some agent commits); ariadne has 238 Claude, 47 `GPT-6`, 9 `GPT-6-Sol`, 8 `GPT-6.1-sol`, 7 `GPT-6.1`, 750 none. Model names are free text and effort is never recorded, so commits can't be joined to `sdlc usage` (ariadne#365) or the bench.

## Spec

One machine-readable trailer on every agent-authored commit, e.g. `Agent: codex/gpt-6.1-sol/medium` (agent/model/effort), alongside the human-facing Co-Authored-By. sdlc-generated commits (publish, tracker, close, merge) stamp it from the invoking session's agent where known, `Agent: sdlc` otherwise. Agents learn it from AGENTS.md/the commit convention; a commit-msg check (or close-gate check over the window) flags agent commits without it. `sdlc usage` (#365) and `sdlc actual` can then group by it.

## Done when

- The convention is in AGENTS.md's commit section and the sdlc commit helpers emit it; a check reports agent commits missing it in an issue's window.
- `git log --format='%(trailers:key=Agent)'` over a new issue's window shows one consistent value per commit.

## Plan

Durable plan: `workshop/plans/000367-agent-trailer-plan.md` (format coordinated with #365; pure `agentid` package, stamped commit sites + sweep test, `sdlc agent-id`, close report, AGENTS.base.md §12).

- [ ] agentid pure core (Identity, Stamp, MissingAgent)
- [ ] transcript readers + Current()
- [ ] sdlc agent-id verb
- [ ] stamp commit sites + sweep test
- [ ] close report of window commits without Agent:
- [ ] AGENTS.base.md §12 + atlas
- [ ] verify: full suite, window trailers

## Log

### 2026-10-14
