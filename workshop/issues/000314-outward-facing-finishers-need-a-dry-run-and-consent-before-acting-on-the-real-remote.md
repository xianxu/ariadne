---
id: 000314
status: open
deps: []
github_issue:
created: 2026-10-09
updated: 2026-10-09
estimate_hours:
card_mirror: '17e8f521119df021a4d4345c3863bd04469b2fe5' # card fields mirrored from issue-cards; edit via sdlc
---

# Outward-facing finishers need a dry run and consent before acting on the real remote

## Problem

During #287's development in ariadne:2 (2026-10-08), running `sdlc issue recovery reconcile` against the real ariadne repo invoked the branch's **unmerged** landing finisher. It deleted several old merged issue branches on GitHub, an outward-facing, irreversible action nobody had sanctioned. From the agent's own words: "running reconcile against the real repo triggered the new finisher, which deleted several old merged issue branches on GitHub — an outward-facing action I hadn't flagged" (evidence D5, `a2-1008:14599`; 4 heads left at `:14617`). No work was lost, because the branches were merged, but the step was unreviewed code acting on shared remote state.

Two factors combined:
- **Build provenance:** the `sdlc` that close/reconcile run is whatever binary the slot built from its working tree, which here was the in-progress branch. Nothing distinguishes "a dev build of this branch" from "a released binary" when deciding whether to take remote actions.
- **No preview or consent:** finishers (#286/#287: remote branch deletion, archive commits on main, card CAS) act directly. There's no dry run listing the remote mutations, and no confirmation for bulk or repo-wide effects. #287's finisher is deliberately repo-wide (it settles every landed leftover, not just the current issue).

## Spec

To be refined at design. Candidates:
- `--dry-run` on reconcile, merge and the other finisher-running verbs, printing every remote mutation (refs deleted, commits pushed, cards changed) without performing them.
- Repo-wide or bulk remote effects (more than the current issue's own branch/card) require confirmation, or `--yes`, when stdin is interactive. Non-interactive callers get a refusal naming the effects, unless they pass an explicit flag.
- A dev build (version stamp says "built from a non-main branch or dirty tree", cf. evidence G1) refuses remote-mutating finishers against the publication remote unless explicitly allowed (e.g. `SDLC_DEV_REMOTE=allow`), steering tests to fixtures.

## Done when

- `sdlc issue recovery reconcile --dry-run` lists the remote branch deletions it would perform and performs none.
- A repo-wide finisher effect beyond the current issue asks for confirmation (or refuses non-interactively without the flag).
- A binary built from an unmerged branch refuses remote-mutating finishers by default, with a test.

## Plan

## Log

### 2026-10-09

Filed at the operator's direction (chosen over a lessons.md rule) from the ariadne-robustness-1 review. ariadne:2 flagged D5 as unfiled while planning #300. Related: G1 (binary provenance and skew, batch 2), #286/#287 (the finishers).
