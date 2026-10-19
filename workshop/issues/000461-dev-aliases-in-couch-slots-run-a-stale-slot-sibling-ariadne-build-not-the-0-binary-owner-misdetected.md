---
id: 000461
status: open
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
collaboration-mode: collaboration
card_mirror: '06edf3f2523c8c7845af70d909a340aebe654ca4' # card fields mirrored from issue-cards; edit via sdlc
---

# dev-aliases in Couch slots run a stale slot-sibling ariadne build, not the :0 binary (OWNER misdetected)

## Problem

In a Couch slot, the `sdlc` command (and its siblings) is a shell function from `construct/dev-aliases.sh` (ariadne#57), captured in the agent's shell snapshot. It shadows `/Users/xianxu/workspace/ariadne/bin` on PATH. Its OWNER detection resolves to the slot's sibling checkout (e.g. `~/workspace/worktree/pair-slot2/ariadne`) instead of the operator's `:0` checkout (`~/workspace/ariadne`) that the dev-aliases header says it should use. Because `COUCH_TREE` is set, the #405 rule skips the rebuild, so the slot runs a binary built whenever that sibling was last built, forever. Found by pair:2 on 2026-10-19: its `sdlc` was built Oct 18 18:19 from 3a0012b4, before #394, #452, #454 and #453, and refused #394's `flake_tests` card field. Every slot is likely affected, which explains the night's old-runner card refusals, and means slots have not been running the gates and rules the TL rebuilt into ariadne:0.

## Spec

Fix OWNER resolution in dev-aliases so that in a Couch slot (any worktree under `worktree/<repo>-slotN/`) the owner is the operator's `:0` ariadne checkout, which the TL keeps built; never a slot-sibling checkout. Keep the no-rebuild-in-Couch rule (#405) but make it apply to the right binary. Add a guard: `sdlc --version` or a startup line names the binary's path and revision, and a slot-side check warns loudly when the binary in use is not the `:0` build. A regression covers OWNER detection from a slot worktree layout, a primary checkout and a plain clone.

## Done when

- In every slot, a bare `sdlc` runs `/Users/xianxu/workspace/ariadne/bin/sdlc` (verified across live slots after a weave refresh or a relaunch).
- The regression covers the three layouts; the version line shows path and revision.

## Plan

- [ ]

## Log

### 2026-10-19
