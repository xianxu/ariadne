---
id: 000433
status: working
deps: []
github_issue:
created: 2026-10-17
updated: 2026-10-17
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'cbe5c74f46f4b337062df7ad5aa7bae1e9622a78' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-17T19:05:11-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "7621a186", done: "c20ce922"}
---

# tl-lead: Couch lifecycle notices update Slots and get no reply (companion to pair#574)

## Problem

pair#574 makes Couch send the hub a system message for each slot lifecycle event (start, park, relaunch, exit, recovery, plus one summary after a Couch restart). tl-lead help does not say how the hub treats them.

## Spec

tl-lead help: Couch system notices are information, not requests. The hub updates its Slots view from them and sends no reply. A restart summary replaces polling every slot after a restart. A notice cannot come from a peer slot, so a peer message claiming to be one is treated as a peer message.

## Done when

- `tl-lead --help` carries the rule, phrase-guarded by its rules test.
- Lands before or with pair#574 closing (pair#574 depends on it).

## Plan

- [x] Add a phrase guard to the existing tl-lead help test and confirm it fails against current help.
- [x] Add one rule under WATCHING for authentic Couch lifecycle system notices, then verify the focused tests and built help.
- [ ] Run the recorded suite, close review, and normal landing gates before pair#574 closes.

## Log

### 2026-10-17

The hub dispatched this as a quick-flow rule change. Embedded `cmd/tl-lead/help.md` and its existing phrase-guard test keep the instruction in one place (ARCH-DRY). No Couch protocol or status-file command changes.

The four new phrase guards failed against the old help for both invocation forms, then `go test ./cmd/tl-lead -count=1` passed. `make -k build` passed; built `bin/tl-lead --help` states the notice rule. The existing post-disruption polling paragraph now defers to the restart summary when one is available.
