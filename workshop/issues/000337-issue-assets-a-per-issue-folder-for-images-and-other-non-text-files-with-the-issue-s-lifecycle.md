---
id: 000337
status: open
deps: []
github_issue:
created: 2026-10-12
updated: 2026-10-12
estimate_hours:
card_mirror: 'ae37fa1a5dadae3e93aa27559a11c08601f1f417' # card fields mirrored from issue-cards; edit via sdlc
---

# issue assets: a per-issue folder for images and other non-text files, with the issue's lifecycle

## Problem

Issues are text only. Couch's DOM submission (pair#483) and the terminal note submission (pair#455) produce screenshot crops that an agent may turn into a new ticket or attach to an existing one ("this demonstrates XXX#NNN, update that ticket"). Today there's nowhere in the issue format for those images to live with the ticket and follow it through its lifecycle. The operator, 10-12: extend the ticket format to hold non-text files, following parley.nvim's chat assets.

## Spec

Unshaped; a starting sketch from parley.nvim's `lua/parley/assets.lua` (#231):
- A per-issue sidecar folder keyed by the issue's **id**, not its slug, so a rename never orphans it (e.g. `workshop/issues/assets/<id>/`).
- **The issue body is the index:** the folder holds only bytes the body links (`![](assets/<id>/<file>)`); nothing scans the folder to discover content.
- **Same lifecycle as the issue:** archive to `workshop/history/` moves the folder too; `sdlc migrate` and peer moves carry it; one writer owns create/move/delete (ARCH-DRY).
- **Git implications (the operator, 10-12):** binary files change how the repos use git: repository size and history weight, diffs, encrypted repos (gcrypt brain), publication through `sdlc issue publish` and the issue-tracker branch (#252). Decide in the design: plain git with size limits, LFS, or keeping bytes outside git with a pointer.

## Done when

- An agent can attach an image to an issue through an sdlc verb, the body links it, and the folder follows the issue through publish, archive and migrate.
- The git storage decision is recorded with its reasons.

## Plan

- [ ]

## Log

### 2026-10-12

Filed from pair#483's shaping (pair:2), at the operator's request. Not a dependency of #483: a submission is written to Couch's store first; attaching it to a ticket is step two.

## Revisions

- 2026-10-12 — the operator: issue assets are where submitted web-cast note crops (pair#483) land when notes become an issue
  (text first in the issue body; only the crops that matter as assets; recordings stay local, never S3 yet). Add a
  **privacy gate enforced by the binary**, not agent judgment: a per-repo asset policy (`allowed` | `text-only`), defaulting
  to `text-only` for a public repository (from its host visibility) and to `allowed` for a private one; the attach verb
  refuses images under `text-only` and says why; the policy and its source are printed by `sdlc issue --help`. The
  operator changes a repo's policy; agents never do.
