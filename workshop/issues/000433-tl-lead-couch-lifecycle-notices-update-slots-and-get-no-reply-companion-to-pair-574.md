---
id: 000433
status: open
deps: []
github_issue:
created: 2026-10-17
updated: 2026-10-17
estimate_hours:
collaboration-mode: collaboration
card_mirror: '9f129aa9e85fa144cca7563413e758ac647b92cb' # card fields mirrored from issue-cards; edit via sdlc
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

- [ ]

## Log

### 2026-10-17
