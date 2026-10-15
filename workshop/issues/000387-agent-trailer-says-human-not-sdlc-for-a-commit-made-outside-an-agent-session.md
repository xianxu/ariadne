---
id: 000387
status: working
deps: []
github_issue:
created: 2026-10-15
updated: 2026-10-15
estimate_hours:
card_mirror: '49c00a61085133c15bc1550afad620f6c1ea0b0c' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-15T11:01:43-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "309edd46", done: "ab958c7e"}
---

# Agent: trailer says 'human' (not 'sdlc') for a commit made outside an agent session

## Problem

ariadne#367's `Agent:` trailer says `Agent: sdlc` when `sdlc` runs in a plain shell with no agent session, i.e. when
the operator makes the commit by hand. Velocity and usage numbers then credit the operator's hand work to "sdlc". The
operator (2026-10-15): "say human in that case".

## Spec

- The no-agent-session fallback renders `Agent: human` (agentid's fallback constant), and `Parse` accepts `human`;
  existing `sdlc` trailers in history still parse (treated as human for attribution).
- `sdlc agent-id` in a plain shell prints `Agent: human`.
- Docs (AGENTS.base.md §12, help) updated.

## Done when

- Tests: plain-shell fallback → `human`; Parse of `human` and of legacy `sdlc`; the commit-site sweep still passes.

## Plan

- [x] `agentid.Human` ("human") is the no-session value; `agentid.Legacy` ("sdlc") still parses to the same human identity
- [x] tests: round trip of `Human`, Parse of `Legacy`, detect's plain-shell cases expect `Human`
- [x] docs: `helptext/agent-id.md`, `AGENTS.base.md` §12; the detect.go override comment
- [x] full suite; land after ariadne#385, whose new detect tests expect "sdlc", switching them to `Human` at the main merge

## Log

### 2026-10-15
- 2026-10-15: closed — Recorded suite sdlc-local/suite green on 035f0447 (main with #385 and #391 merged); the only commit since is the docs-only atlas row (83cf89ad). agentid tests: TestIdentityRoundTrip renders the empty identity as Human and parses it back; Parse(Legacy "sdlc") yields the human identity; the plain-shell and no-session cases in detect_test and #385 qoder_test expect Human; the commit-site sweep (TestEveryCommitSiteStamps) passes. Docs: helptext/agent-id.md, AGENTS.base.md section 12, and the atlas agent-id row.; review verdict: SHIP
- One source for the value (ARCH-DRY): `Human` and `Legacy` constants in agentid.go; tests reference the constant, not the literal. The only readers of the value are `Parse` and the agentid tests (grep across cmd/ and pkg/), so nothing else needs to change.
- Coordination: #385 (ariadne:3) adds detect tests expecting "sdlc" for plain-shell cases. #387 lands after it and updates those cases at the main merge (told ariadne:3).
