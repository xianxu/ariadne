---
id: 000096
status: wontfix
created: 2026-06-14
updated: 2026-10-09
estimate_hours:
github_issue:
started: 2026-10-09T18:40:42-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
tracker:
    version: 1
    abandoned: {}
---

# Peer-repo work legibility: make cross-repo base-layer edits visible at the gate

## Problem

We co-develop the **base layer (ariadne)** alongside example apps, with multiple
work threads spanning peer repos at once (e.g. brain: kaggle + a workflow thread;
ariadne: weave/colima; parley; pair). The pain:

- **ariadne is a shared mutable singleton.** The `sdlc` shell function hardcodes
  the path — it rebuilds-and-runs `/Users/xianxu/workspace/ariadne/bin/sdlc` from
  *that absolute path* on every call, regardless of which repo you're in. And
  `brain/construct/{adapted,datatype,local,config.json,dev-aliases.sh}` are **live
  symlinks** into `../../ariadne/construct/`. So an in-place edit to ariadne is
  globally live across every peer — and changes the behavior of the workflow tool
  every peer runs — the moment it's saved.
- **Worktrees solve leaf work, not base-layer work.** An app is a leaf — fork it,
  nothing points back, worktrees just work. The base layer is the root everyone
  points at, by *absolute path* (the `sdlc` function) and *relative symlink*
  (`../../ariadne/`). A worktree is invisible to both, so we branch ariadne
  **in-place** — and concurrent threads on one working tree tangle.
- **Concrete failure mode.** A session in brain edits ariadne's files *through the
  symlink* while on a brain branch. brain's git sees nothing (it tracks the
  *link*, not the target). The change lands dirty in **ariadne** with no issue, no
  branch, no claim → unattributed → "unrelated changes show up," and both agent
  and human get confused about who changed what.
