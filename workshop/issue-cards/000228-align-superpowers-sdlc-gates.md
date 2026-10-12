---
id: 000228
status: done
created: 2026-09-15
updated: 2026-10-12
estimate_hours:
github_issue:
started: 2026-10-12T16:34:13-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 0.08
tracker:
    version: 1
    completion:
        token: close-3959482f683b
        repository: github.com/xianxu/ariadne
        reviewed_head: dfe05cbd2be40066e4b17bcd0937287e9675fe34
        evidence_commit: 75f73d3285dc44e8af482977de58a9ece828c7f1
        landed_commit: 72b6c9cab10ef77498a4782ed717d4074a60e007
---

# Align local superpowers with SDLC review gates

## Problem

The local adaptation of the Superpowers brainstorming and writing-plans skills
requires review loops that overlap with Ariadne's SDLC gates. In particular,
brainstorming requires an automated spec-document review, and writing-plans
requires a plan-document review. `sdlc change-code` then performs the mandatory
fresh-context plan-quality review before implementation, with persistent findings
and disposition tracking.

This creates duplicate review authority and makes the intended workflow unclear.
The local skills should preserve useful design discipline without requiring agents
to pass multiple substantially overlapping automated approvals.
