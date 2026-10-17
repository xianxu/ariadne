---
id: 000368
status: wontfix
deps: []
github_issue:
created: 2026-10-14
updated: 2026-10-17
estimate_hours:
card_mirror: '4a485d5f5aeceaf8b4870373a660e19c0aa62fb6' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-14T09:58:17-07:00

---

# Cold consumer CI cannot prove committed generated vocabulary ownership across substrate updates

## Problem

parley.nvim#326: main merge-check run 38601556765 at 62a35b2aeb4a849cee910062a35ea02f0b5dba29 fails during bootstrap: preserving authored replacement at construct/generated/vocabulary/issue.json. CI installs published weave and clones current substrate main, while Parley ships a committed vocabulary export for standalone use. Its ignored construct/generated/weave ownership inventory is absent in a fresh checkout. A staged output differing from that committed export therefore has no previous ownership proof and is treated as authored. On a warm checkout the inventory permits displacement/regeneration. This means upstream vocabulary changes can break unchanged consumer CI.

## Spec

Define a reproducible consumer generation contract that preserves authored-file safety and standalone shipped vocabulary. Evaluate pinning substrate/tool revisions with committed generated provenance, or validating committed exports against their recorded source before advancing generation. Do not blanket overwrite occupied generated paths or weaken authored-replacement protection. Upstream vocabulary changes should carry explicit consumer regeneration/propagation when consumers track that source.

## Done when

- A cold clone of a consumer with a committed vocabulary export boots at its recorded substrate revision, even after upstream vocabulary changes.
- Explicit substrate advancement regenerates consumer exports with reviewable drift.
- Authored/tampered generated replacements remain protected; automated coverage exercises cold and warm checkouts, upstream changes, and local edits.

-

## Plan

- [ ]

## Log

### 2026-10-14
- 2026-10-17: closed wontfix, superseded by ariadne#422 (TL decision, ops:0). #422 implements this spec: consumers declare the generated outputs they ship (`internal ship PATH` in their construct/base.manifest), weave writes a committed digest sidecar as provenance, and a cold checkout regenerates a declared output only when its bytes match the sidecar, so hand edits stay refused.
