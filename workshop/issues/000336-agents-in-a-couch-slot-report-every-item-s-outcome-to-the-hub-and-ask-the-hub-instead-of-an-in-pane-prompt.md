---
id: 000336
status: working
deps: [pair#463]
github_issue:
created: 2026-10-12
updated: 2026-10-12
estimate_hours:
card_mirror: '62cb688a0b715069902000921e6863c37e558392' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-12T00:10:32-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
---

# agents in a Couch slot report every item's outcome to the hub, and ask the hub instead of an in-pane prompt

## Problem

In Couch, the operator starts some work in slots by hand and the TL (the hub slot, ops:0) dispatches the rest. Only dispatched slots are told to report to the TL; a slot the operator started by hand reports nowhere, and the operator forgets to check on it (the operator, 10-12: "it's better to have TL having a centralized view"). Slots also ask questions through the agent's in-pane question prompt, which blocks silently until Couch's 10-minute waiting notice (10-11: pair:2 on pair#456, twice).

## Spec

A constitution rule (ariadne's AGENTS.md, which weave compiles into every ariadne-style repo), agent-neutral:
- **Report every item's outcome to the hub** with `couch --send-to <hub>`: landed (with the PR), blocked (on what), waiting for the operator (on what), whoever started the work, the operator by hand or the hub by dispatch.
- **Ask the hub, not an in-pane question prompt.** The hub answers what it can and brings the operator in when it's their call. An in-pane prompt is for work the operator explicitly shapes live in that slot.
- **Call `couch` and `sdlc` as a plain leading command** (#334): not after `cd x &&`, not inside `$(…)`, not by full path, and no `$(` even quoted inside an argument; only that form matches the sandbox exemption.
- How a slot learns its hub's address: decide the simplest source (a Couch-provided environment variable or `couch` reporting it, versus a fixed default); if it needs a pair change, file it in pair and keep the rule here.

Side quest from #334's report: pin `gh *` in `TestBaseSettingsExcludedVerbsMatchArguments`, so the operator's choice can't silently regress.

## Done when

- The rule is in ariadne's AGENTS.md (and wherever the agent-facing Couch skill lives), with the hub's address resolvable from inside any slot.
- A slot started by the operator, given a small task with no instructions about reporting, reports its outcome to the hub (a live check, recorded in the Log).
- The guard test pins `gh *`.

## Plan

- [x] Constitution §15 Couch Slots in `AGENTS.base.md`: hub address, report outcomes, ask the hub, plain leading command (no `$(` or backticks in arguments).
- [x] `construct/local/couch/SKILL.md` points at §15; atlas (`sandbox.md`, `index.md`) updated.
- [x] Guard test pins `gh *` (green at tip; red when the entry is reverted to bare `gh`).
- [x] Pair issue filed: pair#463 (`couch --send-to hub`, resolved live), from pair-slot1 and published on pair main.
- [ ] Live check: an operator-started slot reports its outcome to the hub unprompted (blocked on pair#463; the TL says hold #336 open and close it with this check once pair ships).

## Log

### 2026-10-12
- Hub address: Couch already knows its hub (pair `cmd/internal/couchcmd/message_service.go` `hubWorktree`, the tree Couch started in; used for the TL's 64-message allowance and waiting notices) but exposes it to no slot: no env var, no query. Decision: pair resolves a reserved `couch --send-to hub` target live in the supervisor. A `COUCH_HUB` env var was rejected, since it goes stale when a detached thread is resumed under a Couch started elsewhere. The rule falls back to the dispatching slot until pair ships it.
- Discovery: backticks inside a quoted `couch --send-to` message keep the call sandboxed, like `$(` (#334). Two sends to ops:0 failed this way (one with backticks, one with a literal dollar-paren in prose); neither was delivered (receipt unavailable). The rule bans both.
- Asked the TL (ops:0) which pair checkout files the pair issue, and whether the live check moves into it.
- TL (ops:0) answered: file from pair-slot1, hold #336 open with a dep on the pair issue, and close with the live check after pair ships, so the rule never names a target that doesn't exist yet. Filed pair#463 and set deps.
