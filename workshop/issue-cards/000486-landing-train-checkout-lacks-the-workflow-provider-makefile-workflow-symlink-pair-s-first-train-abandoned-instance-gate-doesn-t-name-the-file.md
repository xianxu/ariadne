---
id: 000486
status: working
created: 2026-10-19
updated: 2026-10-20
estimate_hours:
github_issue:
started: 2026-10-20T01:02:55-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
---

# Landing train checkout lacks the workflow provider (Makefile.workflow symlink): pair's first train abandoned; instance gate doesn't name the file

## Problem

pair's first multi-car landing train (train 167, 10-19 23:13; cars pair#720, #714, #723; train PR pair#476) was
abandoned before any test ran:

- The conductor's train checkout `/Volumes/TestScratch/tmp/pair-slot7/sdlc-train-3847377690/` has no
  `Makefile.workflow`, so the suite refused: "refusing unbounded suite because no Makefile.workflow (lstat …: no such
  file or directory)". In pair (as in every ariadne-layered repo) `Makefile.workflow` is a symlink to
  `../ariadne/Makefile.workflow`, which doesn't resolve from a disposable checkout elsewhere. local-check's suite
  checkouts already handle this (workflow-provider resolution); the train checkout doesn't.
- The instance-conformance gate reported "1 nonconforming instance" with no file name, so the holding car (pair#723)
  couldn't tell what to fix; its own issue frontmatter validates locally.

While pair's trains are on, **every pair landing is blocked** by this.
