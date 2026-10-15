---
id: 000385
status: working
deps: []
github_issue:
created: 2026-10-15
updated: 2026-10-15
estimate_hours:
card_mirror: 'b55d53d636d237f850c177b43e12df3daea83023' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-15T09:33:04-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
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

- Claimed and entered planning. Read-only inspection of the supplied Qoder
  sample found 313 assistant events but 117 distinct response IDs, each with
  one terminal usage record. UUIDs identify streaming events, not responses.
  `credits` is per response; `original_credits` must not be added again.
- All 117 billed responses have zero token placeholders. Credits total exactly
  `45.938730726457140528`, with up to 18 fractional digits. An int64 scaled to
  18 decimal places already overflows on this sample; use bounded exact decimal
  arithmetic, with tokens unavailable rather than measured zero.
- Metadata has mixed timestamp shapes: runtime-config uses epoch milliseconds;
  user/assistant activity uses ISO timestamps; workspace-directories has no
  timestamp. Repeated configuration may replay older timestamps. A shared leaf
  adapter avoids divergent interpretations and the transcripts→agentid import
  cycle (ARCH-DRY, ARCH-PURE).
- Workspace comes from recorded directories, never the project slug. Multiple
  or conflicting roots must remain ambiguous; parentUuid is message ancestry,
  not evidence of a child session. Existing ownership intervals, prefix evidence,
  reviewer exclusion, and close adoption remain the attribution authority.
- Pair source pins Qoder --session-id and exports PAIR_SESSION_ID, but exports
  that context to non-agent panes too. Asked the hub for a measured native
  invoking-session signal; do not guess a QODER_* variable or broaden ordinary
  shell detection. Observed-agent support must not accidentally advertise Qoder
  as a dispatch-capable gate reviewer.
- Next: finish the durable plan (nullable native measurements and credits,
  shared discovery/decoder, identity binding, actual/close fixtures), commit it,
  and send the plan to the hub before change-code. No implementation yet.
