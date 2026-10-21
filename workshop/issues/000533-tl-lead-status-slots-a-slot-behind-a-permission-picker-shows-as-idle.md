---
id: 000533
status: open
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-21
estimate_hours:
collaboration-mode: collaboration
card_mirror: '3941af601f8e5f93a63fc98f47ab1b37a2af48a5' # card fields mirrored from issue-cards; edit via sdlc
---

# tl-lead status slots: a slot behind a permission picker shows as idle

## Problem

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
  is unchanged.

## Plan

- [ ]

## Log

### 2026-10-21
