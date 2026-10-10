---
id: 000315
status: open
deps: []
github_issue:
created: 2026-10-09
updated: 2026-10-09
estimate_hours:
card_mirror: '1273ae2fdc6c155bec391670fb55868fb6816d4d' # card fields mirrored from issue-cards; edit via sdlc
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

- [ ] Add the domain to `.claude/settings.ariadne.json`
- [ ] `weave compile` and confirm the merged `.claude/settings.json`

## Log

### 2026-10-09

Operator decision: do it in ariadne:1, the TL slot, as an exception to the implementation-in-:2–:4 rule, because it's a one-line base-layer change. Stopgap for D2 until #300 lands; part of ariadne-robustness-1's judge-robustness phase.
