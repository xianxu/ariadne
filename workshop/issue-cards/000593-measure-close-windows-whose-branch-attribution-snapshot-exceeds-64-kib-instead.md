---
id: 000593
status: open
created: 2026-10-24
updated: 2026-10-24
estimate_hours:
github_issue:
severity: high
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
