---
id: 000476
status: working
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'a06508f8cf07d1c434497bbd81d5fd2df2b8e142' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-19T20:57:12-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "c747a2d8", done: "f1c16f93"}
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

Preserve line provenance before parsing suite control records. Package outcomes,
shard summaries, JSON events and top-level test failures must start at column zero;
indented quoted output cannot change package attribution or add suite outcomes.
Real indented Go subtest failure lines remain evidence when they descend from the
current outer failing test (including the test-shard summary context). Unrelated
stderr remains inert. Keep the existing top-level persisted failures and transient
subtest fields, registry policy and rerun protocol unchanged (ARCH-DRY, ARCH-PURE).

Prefer this correction at the existing parsers over changing test producers or
introducing a new log format: historical saved logs need the same repair. Merely
ignoring all indented lines would erase the real subtests introduced by #466.

Actions: run the registered targeted retry; repair the named failure; inspect the
raw log and repair/rerun an unattributable failure.

| Input | Visible result | Recovery | Possible loss |
| --- | --- | --- | --- |
| Outer named failure with quoted helper output | Existing exact outer test/leaf | Registered retry if covered, otherwise repair that test | None; raw log unchanged |
| Genuine unrelated package/build/timeout failure | Existing blocking failure | Repair the reported failure | None |
| Missing outer name or package evidence | Existing unattributable failure | Inspect raw log, repair and rerun | No invented attribution |

No durable state, format, artifact family, new refusal or external dependency is
added. Parsing stays bounded by existing log/line caps and a fixed number of
linear scans plus name sorting. Growth drivers are bytes, lines and distinct
names; indentation checks inspect each line once, with no look-back or rescans.

## Done when

- A fixture built from this log attributes the failure to
  `couchcmd.TestContinuationWriterPublishesExactCheckpointAcrossWorktrees/warm-reattached` (indented quoted nested
  output ignored), so the registered-flake rerun applies.
- Interleaved non-test stderr lines before a `--- FAIL` don't break attribution (same fixture).
- Plain and sharded fixtures ignore quoted nested package/build/timeout/JSON
  outcomes, preserve real descendant subtests and keep genuine unowned failures
  blocking. The real local-check path takes only the registered outer retry.

## Plan

- [ ] Reproduce the issue fixture red; add plain, sharded and conservative controls.
- [ ] Correct existing failure/fact scanners before whitespace normalization; add
  the narrow lint guard against erasing line provenance in these scanners.
- [ ] Run focused parser/classifier and local-check regression checks, then the
  recorded suite, close review and normal landing gates.

## Log

### 2026-10-19

- Claimed after #477 landed. Traced `failingNames` (Fields) and both suite-facts
  loops (TrimSpace): quoted nested FAIL package trailers consume pending outer
  failures and manufacture unrelated package outcomes. The issue's captured
  snippet supplies the regression; no live sandbox or peer checkout changes.

- 10-19 19:30 (TL): second case, pair#689 at pair:6. Integrated suite 40e6dab11e0c6019301bd5767e687689 failed only
  at the same `couchcmd.TestContinuationWriterPublishesExactCheckpointAcrossWorktrees/warm-reattached` (nested launcher
  helper TempDir cleanup), classified unattributable-failure again, refused landing, and queued a main baseline. Two
  slots' landings blocked by this today; raise priority if #715's root fix (pair:3) doesn't land first.
