---
id: 000476
status: codecomplete
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
github_issue:
started: 2026-10-19T20:57:12-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 0.82
tracker:
    version: 1
    completion:
        token: close-4a87696b9fc3
        repository: github.com/xianxu/ariadne
        reviewed_head: 9a2711a7220ccda98c24c67acad41caac4695998
        evidence_commit: 0393ba6368c7eed2caefc862e802432b7168b820
---

# Test attribution: quoted nested go test output (a helper subprocess's indented FAIL lines) makes a failed suite unattributable

## Problem

pair#712's recorded suite (attempt d7b39428f6f4d3de2182c233873dbb15, 10-19) failed in one package and sdlc classified
the run as `unattributable-failure` ("couchcmd failed without a failing test"), so no flake path applied. The log does
name the failing test at column 0:

```
--- FAIL: TestContinuationWriterPublishesExactCheckpointAcrossWorktrees (3.25s)
    --- FAIL: TestContinuationWriterPublishesExactCheckpointAcrossWorktrees/warm-reattached (1.01s)
        continuation_acceptance_test.go:137: production transport failed: exit status 1
            === RUN   TestCouchLaunchPromptTransportHelper
            ...
            --- FAIL: TestCouchLaunchPromptTransportHelper (0.12s)
            FAIL
            FAIL	github.com/xianxu/pair/cmd/internal/launcher	0.226s
```

The test runs a nested `go test` (a helper process) and quotes its output, indented, in the failure message. The quoted
nested `--- FAIL` / `FAIL <pkg>` lines most likely confuse the attribution parser, e.g. it attributes to the nested
package/test, or treats the quoted `FAIL` as package end. Interleaved non-test stderr ("couch: terminal: presenter
unavailable") precedes the FAIL block too. Saved log: pair `.git/sdlc/test-stats/failed/d7b39428….log` lines 807–820.
The flake itself is pair#715 (registered).
