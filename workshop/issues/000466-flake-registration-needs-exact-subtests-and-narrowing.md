---
id: 000466
status: working
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
collaboration-mode: collaboration
card_mirror: '26300b29473d3824af298a470d9d52c05d11d11f' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-19T10:37:08-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
---

# flake registration needs exact subtests and narrowing

## Problem

`sdlc flake register` accepts only a top-level Go test name. While triaging #465, the failing subtest was `github.com/xianxu/ariadne/scripts/test.TestConstructScripts/plain/dev-aliases.test.sh`, but the setter only accepted the broad `github.com/xianxu/ariadne/scripts/test.TestConstructScripts`. That registration covers unrelated construct-script failures. The setter can add tests but has no unregister or narrowing operation, so a mistaken broad registration cannot be corrected through a card verb before the issue closes. This is a #394 follow-up; the temporary broad #465 registration is tolerated only until its fix lands.

## Spec

Support exact subtest names in the known-flake registry and same-head evidence/rerun path. Add a compare-and-swap card setter that removes or replaces a registered name, preserving concurrent additions and validating the resulting list. The list and landing classifier must distinguish a registered subtest from other subtests of the same top-level test; a failure in another construct script remains unknown and blocks. Keep the existing strict exact-head and ownership rules.

## Done when

- A real nested Go test fixture registers one named subtest; its same-head failure may use the known-flake rerun, while a sibling subtest failure cannot.
- A card owner can narrow or remove one registered name through the verb without editing card files; concurrent card additions survive, and a repeated setter is idempotent.
- Help and recovery contracts state the new names, writes and retry behavior; targeted tests and the recorded suite pass.

## Plan

Durable plan: `workshop/plans/000466-flake-subtests-plan.md` (names, leaf coverage, top-level rerun, unregister and narrow verbs, state table, Q1–Q3).

- [ ] Names, log subtests, classification, gate, verbs, docs; targeted tests then the recorded suite.

## Log

2026-10-19: Found during #458's full-suite triage. `flakes.ValidTest` and `SplitTest` accept only top-level names; `sdlc flake register` only adds. #465 temporarily registers the top-level `scripts/test.TestConstructScripts` while its test-cost fix is in progress.

2026-10-19 (ariadne:2): claimed and started.
- Design: subtests come from the log as an unpersisted fact, so the test-stats record format is unchanged. Classification covers each failing leaf by a registered name (the leaf or an ancestor).
- The rerun reruns the owning top-level tests and needs each known entry's exact PASS line.
- Plan sent to the hub before code.
