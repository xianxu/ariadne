---
id: 000466
status: working
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
---

# flake registration needs exact subtests and narrowing

## Problem

`sdlc flake register` accepts only a top-level Go test name. While triaging #465, the failing subtest was `github.com/xianxu/ariadne/scripts/test.TestConstructScripts/plain/dev-aliases.test.sh`, but the setter only accepted the broad `github.com/xianxu/ariadne/scripts/test.TestConstructScripts`. That registration covers unrelated construct-script failures. The setter can add tests but has no unregister or narrowing operation, so a mistaken broad registration cannot be corrected through a card verb before the issue closes. This is a #394 follow-up; the temporary broad #465 registration is tolerated only until its fix lands.
