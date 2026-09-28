---
id: '000110'
status: done
created: 2026-06-16
updated: 2026-06-16
estimate_hours: 5
actual_hours: 0.89
---

# Port active-time-v3.py to a native Go package in sdlc

## Problem

`active-time-v3.py` is the engine behind `sdlc actual` (segment-anchored
per-issue dev-hour attribution, #68). But the Go binary reaches it by
*subprocessing python3*: `actual.go` resolves the script under
`construct/local/issues/`, runs `python3 active-time-v3.py …`, parses the
human-formatted stdout with a regex (`parseV3PrimaryHours`), and keys off its
exit codes (0/2/3) via `classifyV3`. That's three layers of fragile glue around
a runtime dependency:

- a **python3 dependency** on the actuals path (the `actualNoScript` "python3
  missing" branch exists only because of this);
- **stdout-as-API** — the measured hours are recovered by regex over a table
  meant for humans;
- **script resolution** — `resolveActualScript` + `substrateChain` exist so
  derivatives can find the owner's copy of a *script*, a problem that vanishes
  once the logic compiles into the binary.

This is the same consolidation `weave` did to `setup.sh`: collapse a shelled-out
script into the Go binary that already owns the workflow. (`ARCH-DRY` — one
implementation; `ARCH-PURE` — the attribution math is pure and belongs in
testable Go, not behind a subprocess boundary.)
