---
id: 000593
status: working
deps: []
github_issue:
created: 2026-10-24
updated: 2026-10-24
estimate_hours:
severity: high
card_mirror: '49757d89b05f3b8b548bbf06982ea695c3e917d2' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-24T04:40:48-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
---

# Measure close windows whose branch-attribution snapshot exceeds 64 KiB instead of refusing

## Problem

Since ariadne#580 (branch attribution, landed 2026-10-24 01:26), a close on an
ordinary single-session issue can refuse its measured actual because the
retained measurement snapshot outgrows its 64 KiB bound:

```
can't auto-compute (measurement branch proof: measurement snapshot does not
serialize: json: error calling MarshalJSON for type *gatestate.MeasurementInputs:
measurement exceeds 65536 bytes).
```

Hit by ariadne#586's close (ariadne:2, 2026-10-24), which then closed with
`--no-actual` on the TL's authorization. The same window measures 1.25h on a
pre-#580 build, so this is a #580 regression, separate from the
resting-branch refusal ariadne#591 fixes.

Measured on #586's window (2026-10-23T22:52 → 2026-10-24T02:57), with an
instrumented build of main@8f3e340e:

- whole snapshot: 67,203 bytes against `SnapshotLimitBytes` = 65,536.
- `sources`: 75 admitted transcript sources, 35,090 bytes (about 470 bytes
  each: harness, id, slot, cwd, digest, started, bytes, plus `windows` and
  `dispatch_windows`). Within `SnapshotLimitSources` = 128.
- `branches` (new in #580): one HEAD-log evidence record holding 150
  intervals, 30,643 bytes. Each source gets two intervals (75 `branch`, 75
  `ownership-only`). Each interval repeats the source id, two RFC 3339 times and
  the 112-character branch name, about 200 bytes each.
- everything else (engagements, queue_waits, cuts): under 1 KB.

So #580 roughly doubled the snapshot's per-source cost (about 470 bytes of
source plus about 400 bytes of branch intervals). A window passes 64 KiB at
about 75 sources, well under the 128-source bound the snapshot already admits.
Long-lived slot sessions with several review dispatches reach that.

## Spec

Measure big windows rather than refuse them. The snapshot must stay bounded
and replay-exact, and must not drop evidence silently. Candidate shapes, to be
settled in design:

- **Compact proof:** don't repeat per-interval data the record already holds.
  For example, intern branch names and source ids per evidence record, or
  merge adjacent intervals of one source and kind. The replayed measurement
  stays identical.
- **Honest cap:** bound the branch evidence (as #547 bounds optional train-wait
  coverage to 16 intervals and 8 KiB) so the sources bound, not the byte bound,
  is the binding limit. Whatever overflows is still measured and reported,
  never silently dropped.

Either way, the byte bound and the source bound must agree. No input admitted
by `SnapshotLimitSources` should fail serialization.

## Done when

- A measurement window like #586's (75 sources, one branch record with 150
  intervals) measures and is retained under the snapshot bound. A regression
  test pins that shape.
- A window at `SnapshotLimitSources` with worst-case branch evidence either
  serializes within the bound, or refuses up front with an error that names the
  bound and the counts. It never fails late inside MarshalJSON.
- Replay of an existing adopted v3 snapshot is unchanged.

## Plan

Durable plan: workshop/plans/000593-measure-close-windows-whose-branch-attribution-snapshot-exceeds-64-kib-instead-plan.md (D1 to D3, approved by the TL on 2026-10-24).

- [ ] Write failing tests: round-trip, bad index, the #586 shape, 128 sources, the counted refusal, v3 replay, the older-reader refusal, unknown members
- [ ] Add the v4 wire types and compactBranches / expandBranches
- [ ] Add the version-aware bound, the structural cap and the counted refusal, and swap the codec's v4 branches member
- [ ] Make the writer emit version 4
- [ ] Update the docs: the sdlc actual help
- [ ] Re-measure #586 with the built binary

## Log

### 2026-10-24

- Claimed and planned from ariadne:2 after #586 landed. The TL approved the plan over Couch. Version 4 alone marks the compact form; the branch method names are unchanged, because version names the encoding and method names the interpretation.
