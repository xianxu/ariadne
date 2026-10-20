---
id: 000476
status: open
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'd5738647f01ef3a07b294ee4da23ec04be57249f' # card fields mirrored from issue-cards; edit via sdlc
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

## Spec

## Done when

- A fixture built from this log attributes the failure to
  `couchcmd.TestContinuationWriterPublishesExactCheckpointAcrossWorktrees/warm-reattached` (indented quoted nested
  output ignored), so the registered-flake rerun applies.
- Interleaved non-test stderr lines before a `--- FAIL` don't break attribution (same fixture).

## Plan

- [ ]

## Log

### 2026-10-19

- 10-19 19:30 (TL): second case, pair#689 at pair:6. Integrated suite 40e6dab11e0c6019301bd5767e687689 failed only
  at the same `couchcmd.TestContinuationWriterPublishesExactCheckpointAcrossWorktrees/warm-reattached` (nested launcher
  helper TempDir cleanup), classified unattributable-failure again, refused landing, and queued a main baseline. Two
  slots' landings blocked by this today; raise priority if #715's root fix (pair:3) doesn't land first.
