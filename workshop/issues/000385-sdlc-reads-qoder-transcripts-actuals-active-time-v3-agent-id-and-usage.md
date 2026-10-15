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

Durable design: [Qoder transcript measurements](../plans/000385-qoder-transcripts-plan.md).

- [ ] Shared bounded Qoder record/metadata adapter and sanitized fixtures.
- [ ] Exact per-response credits, nullable token measurements, global dedup and
  native-unit aggregation/calibration through existing usage primitives.
- [ ] Recorded-workspace discovery feeding ownership-proven actuals and close
  adoption, with retained-prefix and cross-checkout regression fixtures.
- [ ] Measured invoking-session binding and Qoder identity; keep observed-agent
  recognition separate from gate-reviewer CLI capability.
- [ ] Native columns and coverage across fleet/slot/model/issue/hourly output;
  documentation, conformance, exact-head suite, close review and SHIP landing.

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

## Revisions

### 2026-10-15 — measured native unit and integration design

The dispatch explicitly requires Qoder credits by slot/model, not zero token
spend. Extend Spec/Done when: every usage output scope carries optional credits
and nullable tokens/rates; mixed aggregates identify measured subtotals; exact
decimal sums preserve the measured precision. Optional Qoder meter projections
use reported credits as their denominator and retain the existing two-point,
complete-evidence requirement. Usage JSON declares schema version 2. The durable
plan enumerates discovery, normalization, attribution, actual/close, identity,
reporting and calibration consumers. Identity implementation remains blocked on
measured invoking-session evidence (asked the hub), rather than guessing an
environment marker or enabling an unsupported gate-reviewer CLI.

### 2026-10-15 09:47 — hub measurement resolves identity prerequisite

The TL measured QODER_CLI on the running qodercli-1.1.71 process; its tool
shells inherit it. There is no native session-ID variable. Detect Qoder before
other native-agent signals, then choose the uniquely latest activity session
whose recorded workspace matches cwd. Ambiguity or bounded/incomplete discovery
keeps model/effort unknown, with explicit SDLC_AGENT_ID as the launch-bound
fallback. Qoder also sets CLAUDE_PLUGIN_* variables; these are not Claude
markers. The plan revision specifies bounded cwd/activity selection and the
priority/ambiguity regression matrix, superseding the earlier open question.

### 2026-10-15 — approved plan; plan-quality round 1

The hub approved the core plan and minimal native-credit meter. change-code
round 1 raised PQ-1 (recency does not prove invoking session) and PQ-2 (replace
test-case inventories with named functions and mechanical strategies). The plan
now carries a concise function-level adversarial strategy and TDD task sequence.
Asked the hub to refine identity: QODER_CLI establishes invocation; a Pair-pinned
PAIR_SESSION_ID can then bind the exact transcript only after cwd/sessionId
validation; without launch binding, competing sessions remain unknown. Awaiting
that decision before the gate rerun; no production code changed.
