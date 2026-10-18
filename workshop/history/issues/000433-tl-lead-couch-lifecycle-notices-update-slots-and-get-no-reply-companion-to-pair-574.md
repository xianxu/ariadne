---
id: 000433
status: done
deps: []
github_issue:
created: 2026-10-17
updated: 2026-10-17
estimate_hours:
collaboration-mode: collaboration
card_mirror: '482735a8dace15f539d3e162dcb1c6d19615c225' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-17T19:05:11-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "7621a186", done: "c20ce922"}
actual_hours: 0.23
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
- [x] Run the recorded suite and submit the tested branch to close review and normal landing gates before pair#574 closes.

## Revisions

- 2026-10-17: The close review identified the new TL operational flow as missing from the atlas map. Add a concise pointer in `atlas/workflow/tl-skills.md`; the embedded help remains authoritative.

## Log

### 2026-10-17
- 2026-10-17: closed — Four tl-lead phrase guards failed against old help, then focused tests and make -k build passed; built help carries all notice rules; exact-head suite passed on 358d53dfffbcd2cbdb548aaf81d0cc9d7ed8b587; round-1 atlas finding BR-1 addressed in atlas/workflow/tl-skills.md. (done-when: contract moved since change-code; judged by the close review); review verdict: SHIP

The hub dispatched this as a quick-flow rule change. Embedded `cmd/tl-lead/help.md` and its existing phrase-guard test keep the instruction in one place (ARCH-DRY). No Couch protocol or status-file command changes.

The four new phrase guards failed against the old help for both invocation forms, then `go test ./cmd/tl-lead -count=1` passed. `make -k build` passed; built `bin/tl-lead --help` states the notice rule. The existing post-disruption polling paragraph now defers to the restart summary when one is available.

Opened PR #292 as a draft while the suite queued. `sdlc-local/suite` passed on 358d53dfffbcd2cbdb548aaf81d0cc9d7ed8b587: 12 shards, 1584 sdlc tests, 67 other packages, zero failures (log sha256:ded296cdc67d). No new atlas surface is introduced by the help wording.

Close review round 1 returned FIX-THEN-SHIP with Important finding BR-1: map the new lifecycle-notice procedure in atlas. Added the brief map in `atlas/workflow/tl-skills.md`, pointing to embedded help as the rule source. The branch suite still covers unchanged code.
