---
id: 000454
status: open
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
github_issue:
started: 2026-10-18T21:03:58-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
---

# ARCH principle: persisted formats evolve additively; readers ignore unknown fields; refusal only when ignoring is unsafe

## Problem

The ARCH registry has no principle for how persisted formats evolve. ARCH-SECURE covers reading input written by an *older* version, but nothing says what an older reader should do with data from a *newer* writer. Designs default to refusing unknown records (pair#684's first draft, precedent #198). The operator (2026-10-18): "typical robust strategy would be compatibility, that old binary will just ignore new attributes ... both simpler and also more robust ... if this is not your default, I wonder if we should add [it] to architecture".
