---
id: 000391
status: open
created: 2026-10-15
updated: 2026-10-15
estimate_hours:
github_issue:
---

# sdlc actuals fail machine-wide on a Qoder subagent transcript ('incomplete qoder session metadata') after #385

## Problem

After ariadne#385 (Qoder transcripts in sdlc) landed, `sdlc close` refuses to compute actuals on this machine:
"can't auto-compute, source ~/.qoder/projects/-Users-xianxu-workspace-pair/6d3388c7-…/subagents/agent-aExplore-….jsonl:
incomplete qoder session metadata" (ariadne:4 closing #387, 2026-10-15). A Qoder **subagent** transcript (under a
session's `subagents/` directory) lacks the session metadata the new reader requires, and one unreadable file makes
the whole measurement fail. Since every close/actual scans the transcript roots, this likely blocks **every**
`sdlc close` and `sdlc actual` on any machine with Qoder subagent files.
