---
id: 000407
status: done
created: 2026-10-15
updated: 2026-10-18
estimate_hours:
github_issue:
started: 2026-10-18T00:13:46-07:00
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
        token: close-370eb5394958
        repository: github.com/xianxu/ariadne
        reviewed_head: e8a9c84116b1901d7f6e56d741816c74143d4f25
        evidence_commit: a40bf5369187a8f13c70ea9b522a74a78c93a4ed
        landed_commit: dd1a2913c9cfd1493e30be2eb104cd09021c2236
---

# A curated Go lint catalog: golangci-lint + shared ruleguard rules for known mistakes (per-test builds, sleeps, wall-clock asserts, temp dirs, strict wire decode)

## Problem

Recurring mechanical mistakes are caught by humans after they cost something: 11 pair tests `go build` the binary
per run (pair#602), sleeps and wall-clock assertions behind most of 10-15's flakes, temp dirs outside TMPDIR (parley
#342, ariadne#342), strict decoding of cross-version wire messages (pair#594). Neither pair nor ariadne runs a Go
linter; the only lint is ariadne's hand-written `pkg/shelllint`, and lessons files keep rules as unenforced prose.
Operator (2026-10-15): "we need to curate a list of 'stupid mistakes'". Part of ops project test-economy.
