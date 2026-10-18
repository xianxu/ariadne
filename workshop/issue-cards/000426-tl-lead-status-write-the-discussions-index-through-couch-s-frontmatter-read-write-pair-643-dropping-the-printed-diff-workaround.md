---
id: 000426
status: done
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
actual_hours: 0.10
tracker:
    version: 1
    completion:
        token: close-fa472f1816d2
        repository: github.com/xianxu/ariadne
        reviewed_head: 219cb76ce811ac966ab6585e1cc5ce7a7d3872be
        evidence_commit: ccc4e6338282c7104653b29bab0cf3b65eb35ec5
        landed_commit: ec57cef8839afc9e94141f36415140ea082a29e6
---

# tl-lead status: write the discussions index through couch's frontmatter read/write (pair#643), dropping the printed-diff workaround

## Problem

ariadne#424 (M2) writes the TL session file's `discussions:` index through couch's whole-file `read --all` / `write --all`, which reset the diff baseline for the whole file. To avoid consuming the operator's unread edits, `tl-lead`'s `syncIndex` decides no-ops by `Peek` and prints the pending diff (`readAllShowingUnread`) before every whole-file read: correct, but noisy. pair#643 (landed, pair 43313c59) added `couch --session-sync read --frontmatter` and `write --frontmatter … --if-hash H`, which move the baseline only for the frontmatter.
