---
id: 000461
status: open
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
github_issue:
---

# dev-aliases in Couch slots run a stale slot-sibling ariadne build, not the :0 binary (OWNER misdetected)

## Problem

In a Couch slot, the `sdlc` command (and its siblings) is a shell function from `construct/dev-aliases.sh` (ariadne#57), captured in the agent's shell snapshot. It shadows `/Users/xianxu/workspace/ariadne/bin` on PATH. Its OWNER detection resolves to the slot's sibling checkout (e.g. `~/workspace/worktree/pair-slot2/ariadne`) instead of the operator's `:0` checkout (`~/workspace/ariadne`) that the dev-aliases header says it should use. Because `COUCH_TREE` is set, the #405 rule skips the rebuild, so the slot runs a binary built whenever that sibling was last built, forever. Found by pair:2 on 2026-10-19: its `sdlc` was built Oct 18 18:19 from 3a0012b4, before #394, #452, #454 and #453, and refused #394's `flake_tests` card field. Every slot is likely affected, which explains the night's old-runner card refusals, and means slots have not been running the gates and rules the TL rebuilt into ariadne:0.
