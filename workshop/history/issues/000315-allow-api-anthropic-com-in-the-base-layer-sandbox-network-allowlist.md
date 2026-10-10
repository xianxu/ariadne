---
id: 000315
status: done
deps: []
github_issue:
created: 2026-10-09
updated: 2026-10-09
estimate_hours:
card_mirror: '7cd9140e737747e732835899841f738b8e6c828e' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-09T20:30:47-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "81d2119f", done: "e41d3db1"}
actual_hours: 0.02
---

# Allow api.anthropic.com in the base-layer sandbox network allowlist

## Problem

The sdlc boundary reviewer (`claude -p`) runs inside the project sandbox, whose network allowlist (`.claude/settings.ariadne.json` → `sandbox.network.allowedDomains`) lacks `api.anthropic.com`. Its first run in a session often fails with `ERR_PROXY_TUNNEL` / "403 Connection blocked by network allowlist". The gate then records a round with no findings, and agents re-run outside the sandbox. This keeps happening:
- #304's plan-quality round 1 on 2026-10-09;
- xianxu.dev#4, stuck about 8 days, then closed with `--no-judge`;
- ariadne:1 and ariadne:2 closes (evidence D2 in `workshop/pensive/2026-10-09-01-pensive-sdlc-robustness-evidence.md`).

#300 M2 makes the gate *report* this honestly ("review did not run"). This issue removes the cause.

## Spec

- Add `api.anthropic.com` to `sandbox.network.allowedDomains` in the base-layer fragment `.claude/settings.ariadne.json`, which weave merges under each repo's `settings.local.json` into `.claude/settings.json`. That way it survives regeneration and reaches every adopting repo.

## Done when

- After `weave compile`, `.claude/settings.json` in ariadne lists `api.anthropic.com` under `sandbox.network.allowedDomains`.
- The weave settings merge test still passes.

## Plan

- [x] Add the domain to `.claude/settings.ariadne.json`
- [x] `weave compile` and confirm the merged `.claude/settings.json`

## Log

### 2026-10-09
- 2026-10-09: closed — weave compile merged api.anthropic.com into .claude/settings.json (grep count 1); settings.ariadne.json parses; settingsx tests pass. --no-atlas: one allowlist entry, no new architectural surface; review verdict: SHIP

Operator decision: do it in ariadne:1, the TL slot, as an exception to the implementation-in-:2–:4 rule, because it's a one-line base-layer change. Stopgap for D2 until #300 lands; part of ariadne-robustness-1's judge-robustness phase.

Built: added `api.anthropic.com` to `.claude/settings.ariadne.json`. `weave compile` (67 actions) merges it into `.claude/settings.json`, which is untracked and generated; verified with `grep`. The `settingsx` tests pass.
