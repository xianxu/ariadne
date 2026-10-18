---
id: 000409
status: codecomplete
created: 2026-10-16
updated: 2026-10-18
estimate_hours: 1.17
github_issue:
started: 2026-10-18T05:16:36-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 1.65
tracker:
    version: 1
    completion:
        token: close-3dbf9b699b06
        repository: github.com/xianxu/ariadne
        reviewed_head: ead57057e46998084852812ab4da05659271872c
        evidence_commit: 02bec7e2933dbe81a1db48ad166fd9181105dbc9
---

# Dynamic review policy: choose the reviewer model by author family and change kind (cross-family by default), extending #310

## Problem

ariadne#310 made the reviewer agent a static per-repo policy. The operator (2026-10-16): having different models
review is useful robustness; the policy should be dynamic, choosing the reviewer by which model family designed and
coded the change, and by condition. Evidence from 10-15/16: Codex-authored designs under safety requirements
over-engineered (pair#556/#584, #587, #571), while Claude's reviews and redesigns caught it; Codex reviews caught real
bugs in Claude-authored code too.
