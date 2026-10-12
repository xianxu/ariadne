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
- `couch --send-to hub` from a non-hub slot reaches the hub (a live check, recorded in the Log).
- Deferred, owned by the TL after the merge and a weave refresh: a slot started by the operator, given a small task with no instructions about reporting, reports its outcome to the hub.
- The guard test pins `gh *`.

## Plan

- [x] Constitution §15 Couch Slots in `AGENTS.base.md`: hub address, report outcomes, ask the hub, plain leading command (no `$(` or backticks in arguments).
- [x] `construct/local/couch/SKILL.md` points at §15; atlas (`sandbox.md`, `index.md`) updated.
- [x] Guard test pins `gh *` (green at tip; red when the entry is reverted to bare `gh`).
- [x] Pair issue filed: pair#463 (`couch --send-to hub`, resolved live), from pair-slot1 and published on pair main.
- [x] Live check, part 1: `couch --send-to hub` from ariadne:1 resolved to ops:0 (receipt 5525cb43, after Couch restarted onto pair#463 at 08:31).
- [ ] Deferred to the TL: an operator-started slot reports on its own (see Revisions).

## Revisions

- 2026-10-12: the Done-when live check is split (TL decision, option A). Part 1, the hub target resolving, is verified here. Part 2, an operator-started slot reporting on its own, needs §15 on main and a weave-refreshed slot. Simulating it here with a nested headless agent was refused by the auto-mode classifier ("Create Unsafe Agents"), and per the TL that refusal was not worked around. The TL owns part 2 after the merge.

## Log

### 2026-10-12
- Hub address: Couch already knows its hub (pair `cmd/internal/couchcmd/message_service.go` `hubWorktree`, the tree Couch started in; used for the TL's 64-message allowance and waiting notices) but exposes it to no slot: no env var, no query. Decision: pair resolves a reserved `couch --send-to hub` target live in the supervisor. A `COUCH_HUB` env var was rejected, since it goes stale when a detached thread is resumed under a Couch started elsewhere. The rule falls back to the dispatching slot until pair ships it.
- Discovery: backticks inside a quoted `couch --send-to` message keep the call sandboxed, like `$(` (#334). Two sends to ops:0 failed this way (one with backticks, one with a literal dollar-paren in prose); neither was delivered (receipt unavailable). The rule bans both.
- Asked the TL (ops:0) which pair checkout files the pair issue, and whether the live check moves into it.
- TL (ops:0) answered: file from pair-slot1, hold #336 open with a dep on the pair issue, and close with the live check after pair ships, so the rule never names a target that doesn't exist yet. Filed pair#463 and set deps.
- pair#463 landed (pair PR #254). The live check waits for the operator to restart Couch onto a build with it; the TL messages ariadne:1 after the relaunch. Parked until then.
- TL evidence (ariadne:2 still sandboxed with a clean message). Measured with no-op probes (`couch --send-to nosuchrepo:9`, `couch --message-status <x>`), one character class at a time. Inside a quoted argument, only substitution openers defeat the exemption: `$(`, `$((`, backticks, `<(`, `>(`. These all pass: `;`, `(`, `#`, `,`, newline, `$VAR`, `${VAR}`, `&|<>`, globs, quotes, backslash. Around the command, any pipe, `;` chain or redirect to a file defeats it, and `2>&1` alone passes. Root cause for ariadne:2 (transcript 4c3ef887, 14:48Z): a trailing `2>&1 | tail -3`, not the message. Its earlier failures were `$(cat <<EOF …)` (10-11) and the same trailing pipe. §15 now states the measured rule. Recommended a pair option to read the message from a file, since agents write markdown code spans (backticks) by habit.
- Filed pair#486 (`couch --send-to … --message-file PATH`) at the TL's request. §15 names it once it ships, as a follow-up and not a dep of #336.
- Couch restarted onto pair#463 (08:31). `couch --send-to hub` from ariadne:1 resolved to ops:0 and was queued (5525cb43). Ran `weave compile` in this tree, so §15 is in its CLAUDE.md. My first `weave compile 2>&1 | tail -5` ran sandboxed and failed to write settings.json, a live instance of the rule's trailing-pipe case; the bare rerun applied cleanly.
