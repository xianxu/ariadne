---
id: 000426
status: codecomplete
created: 2026-10-17
updated: 2026-10-18
estimate_hours:
github_issue:
started: 2026-10-17T07:33:33-07:00
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
        token: close-552e73009b1c
        repository: github.com/xianxu/ariadne
        reviewed_head: 1278a02061304bf1509f3a016e0be5922e872539
        evidence_commit: fcfd65824b64ae008e54331bdea8453890ae2452
---

# tl-lead status: write the discussions index through couch's frontmatter read/write (pair#643), dropping the printed-diff workaround

## Problem

ariadne#424 (M2) writes the TL session file's `discussions:` index through couch's whole-file `read --all` / `write --all`, which reset the diff baseline for the whole file. To avoid consuming the operator's unread edits, `tl-lead`'s `syncIndex` decides no-ops by `Peek` and prints the pending diff (`readAllShowingUnread`) before every whole-file read: correct, but noisy. pair#643 (landed, pair 43313c59) added `couch --session-sync read --frontmatter` and `write --frontmatter … --if-hash H`, which move the baseline only for the frontmatter.
