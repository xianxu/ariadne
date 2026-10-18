---
id: 000437
status: open
deps: [pair#661]
github_issue:
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
collaboration-mode: collaboration
card_mirror: '2a4d2bb80826e1fc157acc22558261c8fc5c7e71' # card fields mirrored from issue-cards; edit via sdlc
---

# xx-universal-review edits through couch --session-sync apply; retire deletions: propose-strike

## Problem

Slice 4 of the universal-agent-input-interface project (design: `ops:teams/devinfra/projects/universal-agent-input-interface/design.side/10-s4-edit-path.md`, sections 3, 5 and 8). The shared review skill tells agents to edit documents with `couch --session-sync write` and to hand-write proposal markers under `deletions: propose-strike`. S4 gives edits one path, `apply`, with per-edit reasons, and one knob, `presentation: propose`, for proposals.

## Spec

Design: `ops:teams/devinfra/projects/universal-agent-input-interface/design.side/10-s4-edit-path.md` (operator-reviewed, 2026-10-17). Depends on pair's M1 (`couch --session-sync apply`, the derived round, the propose converter).
- `xx-universal-review` makes every edit to an existing document with `apply`: answering markers, fold-ins and collaboration edits. Each record is `{old, occurrence, new, explain}`, and the `explain` is written for the operator.
- An exit 6 is retried by reading again and rebuilding the records. An exit 1 `edits-invalid` is fixed and retried. Nothing is ever partial.
- `write` stays for generated content (tl-lead status's sections) and new side docs; its changes highlight via the derived round.
- The agent never hand-writes proposal markers: Couch converts under `presentation: propose`. On `propose-too-large`, the agent writes the rewrite as a side doc proposal (S2 naming), links it in place with `→ [proposal](…)`, and applies it on `🤖[fold in]` by the link (S3 fold-in).
- The reply lists the round's edits.
- `deletions: propose-strike` is retired in favor of `presentation: propose`, with the skill's prose, examples and examples checker updated.

## Done when

The skill's contract and examples use `apply` for edits, cover the exit-6 retry, the propose-too-large side doc path and fold-in, and no longer mention `deletions: propose-strike` except as retired. The examples checker passes. An agent round on a real document with the skill loaded arrives through `apply` with reasons.

## Plan

- [ ]

## Log

### 2026-10-18
