---
id: 000422
status: open
deps: []
github_issue:
created: 2026-10-16
updated: 2026-10-16
estimate_hours:
collaboration-mode: collaboration
card_mirror: '1ed9df762b592340645ffe73c80ceece615b3cab' # card fields mirrored from issue-cards; edit via sdlc
---

# A downstream repo's committed generated vocabulary goes stale on every ariadne vocabulary change and turns its CI red

## Problem

## Spec

Found 2026-10-16 by parley.nvim:3. parley.nvim commits a copy of the generated `construct/generated/vocabulary/issue.json`. When ariadne's issue vocabulary changes, weave's apply refuses to overwrite the committed copy ("preserving authored replacement"). parley's main merge-check then goes red until someone refreshes it by hand. That has happened three times: parley #321, #326, and PR 285 after ariadne#415, which turned parley's main red from 22:32 to about 23:50 and blocked every parley landing.

Preserving a committed file that differs from weave's output is the right default for authored files. The gap is a generated file that a repo tracks on purpose. Options:
1. Repos stop committing generated vocabulary; it is built at bootstrap. Find out why parley committed it first.
2. A repo declares tracked generated paths, and apply refreshes those, showing the diff, instead of preserving them.
3. An ariadne vocabulary change ships with a fleet check that names each downstream repo whose committed copy it makes stale, so the refresh lands with the change.

## Done when

- An ariadne vocabulary change can no longer turn a downstream repo's main CI red silently: a test or check reproduces the parley case and shows the chosen fix handling it.
- parley's issue.json is handled by the chosen rule, and the reason is recorded.

## Plan

- [ ]

## Log

### 2026-10-16
