---
id: 000465
status: codecomplete
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
github_issue:
flake_tests: [github.com/xianxu/ariadne/scripts/test.TestConstructScripts]
started: 2026-10-19T07:56:05-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 1.17
tracker:
    version: 1
    completion:
        token: close-9072cbfc8e43
        repository: github.com/xianxu/ariadne
        reviewed_head: f874054852a15da9f76d9d2fc23482a05f99a148
        evidence_commit: fb45f20e175e09b1a433ebdbfbf0d330f03b12c0
---

# dev-aliases script exceeds its per-script budget under suite load

## Problem

The recorded ariadne suite on head da8c62d2e111ef7fe86cc0c351e00d6512a5a87a failed only `scripts/test.TestConstructScripts/plain/dev-aliases.test.sh`: the per-script 2m deadline fired after about 40 passing assertions. The failed attempt is 4c96305468ac871ff567625384059cc8, log SHA-256 prefix 7714ac8e8e11, saved at `/Users/xianxu/workspace/ariadne/.git/sdlc/test-stats/failed/4c96305468ac871ff567625384059cc8.log`. All 12 cmd/sdlc shards and 1,688 tests passed. An unchanged exact-head targeted run of that subtest passed 82/82 assertions in 33.6 seconds. #461 expanded the script with real Git layouts and sentinel binary cases; that cost crosses #448's fixed two-minute per-script ceiling under fleet suite load. The #458 branch changes no construct script or runner code.
