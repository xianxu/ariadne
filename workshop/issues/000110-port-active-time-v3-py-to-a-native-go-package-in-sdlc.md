---
id: 000110
status: working
deps: []
github_issue:
created: 2026-06-16
updated: 2026-06-16
estimate_hours: 5
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

## Spec

Port the v3 logic into a native Go package and call it directly; keep the
standalone/manual-inspection affordance as an `sdlc active-time` subcommand.

- **New package `cmd/sdlc/internal/activetime`** holding:
  - *pure* core: `activeMinutes` (15-min gap-truncated sum), `attributeSegment`
    (commit-weight + mention split), segment construction (prefix / per-commit /
    suffix boundaries, prefix detection, exact-instant anchor matching).
  - *IO* edges: transcript `.jsonl` event loading (`walkSessionEvents`,
    `loadEvents`) and `git log` window loading (`loadCommits`).
  - a `Compute(opts)` entrypoint returning a **structured result** (per-issue
    totals, per-segment rows, a status discriminating measured / telemetry-gap
    (commits-but-0-events) / empty-window / no-commits-fallback).
- **`actual.go` calls `activetime.Compute` directly** — delete `v3Runner`,
  `resolveActualScript`/`actualScriptRel`, `parseV3PrimaryHours`, `classifyV3`,
  and the python3 `LookPath`. `actualNoScript` collapses (the only remaining
  "can't compute" reasons are no-window / internal error). The
  brain+repo dir-selection heuristic (`selectActualDirs`) and the structured
  status→message mapping (`printActual`) stay.
- **`sdlc active-time` subcommand** exposes the full original CLI (`--dir`
  repeatable, `--git-repo`, `--since/--until`, `--issue`, `--commit-weight`,
  `--prefix-commit-weight`, `--threshold-min`, `--include-assistant`) and renders
  the per-segment table + per-issue totals for manual inspection. Preserve the
  #68 loud-fail contract as exit codes (2 misinvoke / 3 telemetry-gap / 0
  measured-or-empty) so the standalone behavior is unchanged.
- **Migrate the explainers** off the python command: `sdlc close`'s missing-
  `--actual` text, `construct/local/issues/SKILL.md`, and `helptext/actual.md` /
  `close.md` / `milestone-close.md` reference `sdlc active-time …` instead of
  `python3 …/active-time-v3.py …`.
- **Delete** `active-time-v3.py` and `test_active_time_v3.py`; port the #68 guard
  tests + the attribution-math cases into Go tests.
- **Atlas**: update `atlas/workflow/sdlc-binary.md` (and ledger-landscape if it
  names the script) for the new surface.

**Correctness bar — numeric parity.** Before deleting the Python, run both over
the same real commit window (a recently-closed issue) and confirm per-issue
hours match to the rounding `sdlc actual` reports. The float math
(`active_minutes`, the commit/mention split) and the subtle segment logic
(boundary dedup by instant, anchor match by timestamp equality, prefix
detection) must reproduce exactly — a drift here silently re-bases velocity
calibration.

Out of scope: the #092 fat-segment over-attribution fix (separate behavior
change; this port preserves current behavior verbatim).

## Done when

- `sdlc actual --issue N` produces the same hours it does today (parity-checked
  against the Python over ≥1 real window) with **no python3 process spawned**.
- `sdlc active-time <flags>` reproduces the Python script's table + per-issue
  totals and its 2/3/0 exit-code contract (the ported guard tests pass).
- `active-time-v3.py` + `test_active_time_v3.py` are deleted; nothing in the repo
  (code, helptext, SKILL.md, atlas) still points at them.
- `go build ./... && go test ./cmd/sdlc/...` green; `go vet` clean.

## Plan

- [ ] M1 — `internal/activetime` package: pure core (`activeMinutes`,
      `attributeSegment`, segment build) + IO loaders (events, commits) +
      `Compute()` structured entrypoint. Go unit tests for the math and the #68
      guards. Parity-check `Compute` output vs. the Python over a real window.
- [ ] M2 — Integrate + retire: `actual.go` calls `activetime.Compute` (drop the
      subprocess/script-resolution/stdout-parse machinery); wire `sdlc
      active-time` subcommand + `helptext/active-time.md`; migrate close.go /
      SKILL.md / helptext explainers; delete the two `.py` files; update atlas.

## Log

### 2026-06-16
