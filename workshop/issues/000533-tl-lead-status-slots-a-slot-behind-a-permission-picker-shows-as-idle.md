---
id: 000533
status: blocked
deps: [pair#786]
github_issue:
created: 2026-10-21
updated: 2026-10-21
estimate_hours:
collaboration-mode: collaboration
card_mirror: '86d3d44499447511dbc114aeedf5b8fd767e1898' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-21T16:31:41-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "a0724001", done: "02f13960"}
---

# tl-lead status slots: a slot behind a permission picker shows as idle

## Problem

The renderer treats Couch actors Resting as prompt readiness. Couch defines it as resting Git branch, so a running or picker-blocked agent can be called idle.

## Spec

From ops-echo:0's TL (10-21): `tl-lead status slots` listed echo:1 as **"codex idle"** for about an hour while it was
actually blocked behind a permission picker (a `git switch -c` approval prompt). Couch's own wait notices already
detect this state ("waiting N min behind a picker: …"), so the information exists; the Slots section loses it. Same
class seen on ops:0's side today (pair:4 behind an `rm -rf` approval prompt).

Fix: Slots shows **"blocked: picker"** (with the prompt's first line, truncated) for a slot whose screen is a picker,
using the same detection Couch's wait notices use; never "idle". A slot behind a picker counts as needing the
operator.

## Done when

- A fixture slot behind a picker renders as "blocked: picker" with the prompt excerpt, not idle; a genuinely idle slot
  keeps its agent identity. Resting alone must never claim runtime idleness.
- Claude, Codex and Qoder labels retain their agent identity on both resting and issue branches.
- Picker consumption uses the authoritative additive API from pair#786, with captured-response coverage and shared disposable-Couch conformance; no terminal detector is copied into Ariadne.

## Plan

- [x] Remove the false idle inference in pure agentLabel, document the distinction, and test all three agent kinds before implementation.
- [ ] Checkpoint and park pending pair#786; work #532 separately from main.
- [ ] Consume pair#786 Picker (boolean or unknown) and PickerPrompt (bounded first line only while true), then verify, run recorded suite and close/land.

## Log

### 2026-10-21

Root cause traced through cmd/tl-lead/actors.go and slots.go to Pair couchmessage.Candidate and Broker.Actors: Resting comes from restingView (Git), not SlotLiveness.Settled or Picker. A read-only live actors reply confirms no picker fields. No live fleet was driven or modified.

TL approved filing the producer dependency and correcting the false idle label now, then parking for #532. Filed pair#786 unclaimed via a branched sibling because Pair :0 is detached; creation wrote no checkout. Contract: Picker boolean or null/absent when unavailable; PickerPrompt bounded recognized first line only while true, same wrapper incarnation; Resting/Known unchanged.

The immediate correction uses pure agentLabel and existing decodeActors, with no new IO, persistent state or lifecycle. Tests enumerate agent kinds and Resting values; existing no-actor/read-error controls remain. This semantic provenance error needs behavioral tests rather than a lexical rule incapable of proving field meaning (ARCH-PURE, ARCH-PURPOSE). Picker detection remains single-sourced in Couch (ARCH-DRY).

## Revisions

### 2026-10-21 — producer dependency and immediate correction

On ops:0 instruction, defer picker consumption until pair#786 exposes the registry signal. Remove independently incorrect idle labels now. Original acceptance said genuinely idle slots are unchanged; current acceptance preserves agent identity while removing the unsupported idle inference. Full blocked: picker acceptance remains pending, not silently reduced to this correction.

2026-10-21: TDD regression failed for all three agent kinds on Resting=true (fabricated ·idle), then passed after removing that inference. Full cmd/tl-lead package passed (21.793 s). Existing gathering/rendering expectations updated; help, README and atlas distinguish branch placement from runtime readiness. No full recorded suite or close claimed: picker acceptance is pending pair#786. Branch checkpoint is parked for the independent #532 dispatch.
