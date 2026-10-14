---
id: 000359
status: open
deps: []
github_issue:
created: 2026-10-13
updated: 2026-10-13
estimate_hours:
card_mirror: '38c4c530f468729d80b57c11e9b9ba865a52ec76' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-13T18:30:59-07:00
---

# weave Tools inherits MAKEFLAGS: under make -k a failing tools:: recipe no longer stops the build (TestToolsDoubleColonFailurePropagates fails under make -k)

## Problem

## Spec

## Done when

-

## Plan

- [ ]

## Log

### 2026-10-13

## Revisions

- 2026-10-13 — duplicate of ariadne#358 (weave's tools run passes -S so an inherited make -k can't change it), which
  ariadne:5 is working. Close this one when #358 lands.
