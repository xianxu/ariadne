---
id: 000385
status: open
deps: []
github_issue:
created: 2026-10-15
updated: 2026-10-15
estimate_hours:
card_mirror: 'ef4dfc15cda6fbb9f8c89914a96a3ad69b4e4af3' # card fields mirrored from issue-cards; edit via sdlc
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

## Spec

- Add Qoder as a transcript source behind the same seam the Claude and Codex readers use: discovery under
  `~/.qoder/projects/<slug>/`, the cwd from `workspace-directories`, events with timestamps, model and effort from
  `runtime-config`.
- Feed it to active-time-v3 (`sdlc actual`, close adoption), `agentid.Current` (`sdlc agent-id`), and `sdlc usage`
  (token counters if the format reports them; otherwise null, never zero).
- Fixtures from a real Qoder transcript (trimmed, no secrets).

## Done when

- `sdlc actual` on a Qoder-worked issue measures nonzero hours from its transcript; close adopts them.
- `sdlc agent-id` in a Qoder session prints `Agent: qoder/<model>/<effort or ->`.
- `sdlc usage` lists Qoder slots (with unsupported dimensions as null).
- Tests over the fixture for each.

## Plan

- [ ]

## Log

### 2026-10-15
