---
id: 000385
status: done
created: 2026-10-15
updated: 2026-10-15
estimate_hours: 3.81
github_issue:
started: 2026-10-15T09:33:04-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 5.74
tracker:
    version: 1
    completion:
        token: close-8e3653a25e72
        repository: github.com/xianxu/ariadne
        reviewed_head: c907a4d0ba566153263b2334826a9eeb5779f1dd
        evidence_commit: 79c3927be1388a855e0d0b1ca15e3b85b3c74181
        landed_commit: c82c88ba71acc77227fdcb033e8ac9bf4f6aca6a
---

# sdlc reads Qoder transcripts: actuals (active-time-v3), agent-id and usage

## Problem

The first Qoder trial (pair#558 on pair:4, 2026-10-15) hit two sdlc blind spots, both because sdlc reads agent
transcripts only from Claude (`~/.claude/projects`) and Codex (`~/.codex/sessions`):
- `sdlc close` refused with "no measurable activity": active-time-v3 found 0 events, because the session was in
  `~/.qoder/projects/<slug>/<session>.jsonl`. The worker had to close with `--no-actual`, so Qoder work drops out of
  velocity calibration.
- `sdlc agent-id` (#367) printed the no-session fallback `Agent: sdlc` although `PAIR_AGENT=qoder` was set; the worker
  wrote `Agent: qoder/auto/-` by hand. `sdlc usage` (#365) likely has the same gap.

Qoder's transcript is JSONL close to Claude Code's shape: `{"type":"user"|…,"sessionId":…,"timestamp":…,"message":…}`,
with `runtime-config` rows carrying `model` and `reasoningEffort`, and a `workspace-directories` row with the cwd.
