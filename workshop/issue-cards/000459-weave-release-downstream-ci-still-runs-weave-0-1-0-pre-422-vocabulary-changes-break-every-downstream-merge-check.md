---
id: 000459
status: done
created: 2026-10-19
updated: 2026-10-19
estimate_hours: 2.72
github_issue:
started: 2026-10-19T01:11:31-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 4.88
tracker:
    version: 1
    completion:
        token: close-f73b861424c4
        repository: github.com/xianxu/ariadne
        reviewed_head: e4e9f4fd8392af3490ad86f53ddb885e512776ec
        evidence_commit: 5bfd02bc61f1606da8f3dc4a53ab5bd038b4a693
        landed_commit: 401e1ebc001e25a4f574d7bda04a015489796e55
---

# weave release: downstream CI still runs weave 0.1.0 (pre-#422); vocabulary changes break every downstream merge-check

## Problem

Downstream repos' merge-checks install the published weave release 0.1.0, which predates ariadne#422 (ship and sidecar ownership). On 2026-10-19 ariadne#394 changed construct/vocabulary/issue.cue; weave 0.1.0 then treated parley's regenerated construct/generated/vocabulary/issue.json as an "authored replacement" and failed every parley PR's merge-check (parley.nvim#396 regenerates to unblock). Any later vocabulary or schema change will trip the old released compiler the same way, in every downstream repo whose CI pins the release.
