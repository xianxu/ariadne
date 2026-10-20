---
id: 000486
status: open
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'c234dd9a240e395dc2a4560a5f1620d09d71c292' # card fields mirrored from issue-cards; edit via sdlc
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

## Spec

- The train checkout resolves the workflow provider exactly as local-check suite checkouts do (shared helper;
  ARCH-DRY), so `make test` works in it.
- The instance gate names each nonconforming file and the rule it breaks, in the train report and in the held car's
  outcome.
- A train abandoned for an environment reason (not a test failure) releases its cars to the serial path or the next
  train automatically, never stranding them (check what happens today and pin it).

## Done when

- A fixture repo whose Makefile.workflow is a symlink to a sibling substrate runs a 2-car train end to end (test).
- The instance-gate failure message names the file (test).
- Rerun pair's first train live; record it in test-economy.

## Plan

- [ ]

## Log

### 2026-10-19
