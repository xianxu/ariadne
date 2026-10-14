---
id: 000365
status: open
deps: []
github_issue:
created: 2026-10-14
updated: 2026-10-14
estimate_hours:
card_mirror: '1e60accb6ea30a9f9f6c214f73549d73d3412497' # card fields mirrored from issue-cards; edit via sdlc
---

# sdlc usage: tokens per slot, agent and model from local transcripts, with burn rate (fleet monitoring + bench metric)

## Problem

The fleet has no view of token spend. On 2026-10-14 the TL learned only by driving `/status` into an idle Codex slot that Codex's weekly limit was exhausted and workers were drawing credits, and by `/usage` in a Claude slot that Claude was at 94%. Driving commands into agents is intrusive and impossible mid-turn. The operator wants token monitoring with burn rates before routing work to other providers (e.g. Qoder's off-peak window), and the upcoming bench (ops workshop/pensive/2026-10-13-02-pensive-model-eval-by-replay.md) needs tokens per run.

The agents' own transcripts already record usage: Claude's session JSONL carries `usage` (input, output, cache read/write) per response; Codex's rollout JSONL carries `token_count` events. ariadne#335 just made transcript attribution per owning slot reliable.

## Spec

`sdlc usage` (read-only, no agent interaction), reusing #335's transcript discovery and attribution:
- Totals per slot, agent and model over a window (`--since`, default 24h), split input / output / cache read / cache write where the agent reports them; per hour, so a **burn rate** (tokens/hour over the last 1–3 h) falls out.
- `--issue N`: tokens for an issue's ownership intervals (the bench metric), alongside `sdlc actual`'s active time.
- `--json` for the TL's session file and the bench.
- Optional calibration: the TL records an observed account meter (e.g. Codex credits, Claude weekly %) with a timestamp (`sdlc usage calibrate --agent codex --remaining 29276 --unit credits`), and usage projects time-to-exhaustion from the measured burn rate. No scraping of agent UIs.
- Unknown or unparseable transcript formats are reported (counts of skipped records), never guessed.

## Done when

- `sdlc usage --since 6h` lists tokens per slot/agent/model with an hourly burn rate, from real Claude and Codex transcripts (tests on recorded transcript fixtures for both formats).
- `sdlc usage --issue N` reports tokens within the issue's ownership intervals, consistent with `sdlc actual`'s attribution.
- A calibration plus burn rate yields a time-to-exhaustion projection; `--json` output is documented.

## Plan

- [ ]

## Log

### 2026-10-14
