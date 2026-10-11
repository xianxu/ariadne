---
id: 000204
status: codecomplete
created: 2026-08-25
updated: 2026-10-11
estimate_hours:
github_issue:
started: 2026-10-11T00:56:11-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 0.99
tracker:
    version: 1
    completion:
        token: close-05a16b1950a1
        repository: github.com/xianxu/ariadne
        reviewed_head: 388cceb8bbf6cdce4b428171d7705d3801b785cf
        evidence_commit: 3e8cadc387dfabd3616a25674aa3d191e6bd9437
---

# Run each boundary reviewer in a disposable isolated checkout

## Problem

Boundary reviewers are instructed to be read-only and leave fixes to the main
agent, but the review protocol also asks them to prove fixes by reverting or
mutating them and running tests. The process currently inherits the real
checkout as its working directory. Claude retains Bash under
`bypassPermissions`; Codex and Gemini do not consume Claude's tool allowlist.
The read-only claim is therefore prose, not an isolation boundary.

An audit of pair#146's ten most recent couch reviews found tool use in all ten
and deliberate mutation or generated artifacts in eight. Some reviewers
voluntarily used scratch worktrees or `git archive` copies, while one Codex
review created and then removed a binary in the real checkout. Safety currently
depends on each reviewer independently inventing the same containment practice.
