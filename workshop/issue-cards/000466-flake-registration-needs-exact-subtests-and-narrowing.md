---
id: 000466
status: codecomplete
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
github_issue:
started: 2026-10-19T10:37:08-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 1.77
tracker:
    version: 1
    completion:
        token: close-92ef073ce6e8
        repository: github.com/xianxu/ariadne
        reviewed_head: ee4d22050707b73f6c81b9a365082237d3ac687e
        evidence_commit: cc7668e7d753870e8db01ec88659aa4040c22935
---

# flake registration needs exact subtests and narrowing

## Problem

`sdlc flake register` accepts only a top-level Go test name. While triaging #465, the failing subtest was `github.com/xianxu/ariadne/scripts/test.TestConstructScripts/plain/dev-aliases.test.sh`, but the setter only accepted the broad `github.com/xianxu/ariadne/scripts/test.TestConstructScripts`. That registration covers unrelated construct-script failures. The setter can add tests but has no unregister or narrowing operation, so a mistaken broad registration cannot be corrected through a card verb before the issue closes. This is a #394 follow-up; the temporary broad #465 registration is tolerated only until its fix lands.
