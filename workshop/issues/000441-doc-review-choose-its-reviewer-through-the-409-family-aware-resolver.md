---
id: 000441
status: open
deps: []
github_issue:
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
collaboration-mode: collaboration
card_mirror: '814cc9b80dfa9948ad3dde0b667339651a1b4b8e' # card fields mirrored from issue-cards; edit via sdlc
---

# doc-review: choose its reviewer through the #409 family-aware resolver

## Problem

`cmd/doc-review` hard-codes its reviewer (`DefaultAgent = AgentCodex`, "a different vendor from the usual co-authoring agent"). #409 gives sdlc one family-aware reviewer resolver (cross-family by default, visible fallback); doc-review does not use it, so a codex-authored document is reviewed by codex, and an unavailable codex has no visible fallback.

## Spec

## Done when

- doc-review chooses its reviewer through the #409 resolver (the document author family when known, else the policy order), with the same visible degradation; or a decision records why it stays fixed.

## Plan

- [ ]

## Log

### 2026-10-18
