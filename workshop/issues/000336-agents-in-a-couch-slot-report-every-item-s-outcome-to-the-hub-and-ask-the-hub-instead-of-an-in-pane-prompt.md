---
id: 000336
status: open
deps: []
github_issue:
created: 2026-10-12
updated: 2026-10-12
estimate_hours:
card_mirror: 'dcbd1528d76ffab56f6e9a65419c6cee7f5e0c51' # card fields mirrored from issue-cards; edit via sdlc
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

- [ ]

## Log

### 2026-10-12
