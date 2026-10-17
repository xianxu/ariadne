---
id: 000415
status: codecomplete
created: 2026-10-16
updated: 2026-10-16
estimate_hours:
github_issue:
started: 2026-10-16T18:46:26-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 0.21
tracker:
    version: 1
    completion:
        token: close-fbcd5624116b
        repository: github.com/xianxu/ariadne
        reviewed_head: 614b8f11f0eb8bd986d57eac0df1963831c3eaf9
        evidence_commit: 65c5c6e9f0035d4c4fce633ec0dfe8745b04dab3
---

# S2 M2: every managed document type declares a default collaboration-mode

## Problem

S2 M1 (ariadne#412) gave each section of a reviewed document a collaboration mode (discussion, collaboration,
chat), with discussion as the implicit default when a document has no `collaboration-mode:` key. The operator wants
every ariadne-managed document type to declare its default explicitly, so a document's mode is visible at the top
and nothing new relies on the implicit default. The design is ops
`teams/devinfra/projects/universal-agent-input-interface/design.side/01-s2-section-modes.md` §6 M2 and side doc 04.
Past documents are not migrated (operator decision, QA files included).
