---
issue: 000140
title: sdlc boundary review should show progress while waiting
created: 2026-07-01
---

# Plan — boundary review shows live progress while waiting (#140)

## Problem restated

`sdlc close` / `sdlc milestone-close` dispatch a fresh-context boundary review
that can run for minutes. Today the flow is:

```
dispatchBoundaryReview()  → cinfo "dispatching boundary review (…HEAD) via claude …"
                          → judge.Dispatch(ctx, opts)   ← BLOCKS, silent
                          → fmt.Fprint(stdout, output)  ← everything at the end
```

`judge.Dispatch` calls the `Run` seam, which is `exec …CombinedOutput()` — one
blocking call, zero output until the child exits. During the wait the only way
to tell "alive" from "wedged" was to `ps` the process tree for a `claude -p`
child (literally what the operator did in pair#84).

## Design

Make `judge.Dispatch` emit a periodic heartbeat to `opts.Stderr` while the agent
subprocess runs, showing **elapsed time + agent name + child PID**. This is the
automated form of the operator's manual `ps` workaround.

Three deliberate design choices:

1. **Heartbeat lives inside `judge.Dispatch`, not a boundary-only wrapper.**
   `DispatchOptions` already carries `Stdout`/`Stderr` — set by *every* caller
   (close, milestone-close, changecode, preflight/judge) but **currently unused
   by `Dispatch`**. That dead field is the intended progress seam. Putting the
   heartbeat in `Dispatch` (gated on `opts.Stderr != nil`) is the one-path,
   **ARCH-DRY** choice: the boundary review — the issue's target — gets it for
   free (it already passes `Stderr`), and the plan-quality / pre-merge judges get
   it as a bonus with zero extra wiring. A boundary-only wrapper would duplicate
   the ticker logic the day changecode wants the same thing.

2. **PID is surfaced by evolving the `Run` seam**, not bolted on with global
   state. `Run` becomes:
   ```go
   var Run = func(ctx context.Context, onStart func(pid int), name string, args ...string) ([]byte, error)
   ```
   Production reimplements the exec as `Start()` → `onStart(cmd.Process.Pid)` →
   `Wait()` (combined stdout+stderr into a mutex-guarded buffer — same bytes
   `CombinedOutput` would return, same exit-code semantics), so the PID is known
   *at launch*, exactly when the heartbeat needs it. A callback (not a return
   value) is required because the return value only arrives *after* the wait —
   too late to be live. One seam, not two parallel exec paths (ARCH-DRY; cf. the
   "N walkers drift" lesson).

3. **No byte-count / log-tail signal.** Investigated per the Spec: `claude -p`
   buffers its result and emits at the very end, so a live byte counter would sit
   at "0 B" for the whole wait and *look* stalled — a misleading signal, worse
   than none. No provider (claude/codex/gemini) returns an invocation id that
   lets `sdlc` reliably correlate a per-invocation log file, so log-tailing can't
   be done reliably either. Done-when #3 is conditional ("*If* a provider log can
   be located reliably") — we document the negative finding in `## Log` and ship
   the reliable signal: elapsed + agent + PID. (Recorded here so the boundary
   reviewer sees the reasoning, not a silent omission.)

### Pure core / thin IO split (ARCH-PURE)

- **Pure:** `heartbeatLine(elapsed time.Duration, agent string, pid int) string`
  — the exact wording (elapsed, agent, pid; a "pid pending" branch for the window
  before `onStart` fires). Unit-tested directly, no IO.
- **Thin IO:** the ticker loop inside `Dispatch` — run `Run` in a goroutine,
  `select` on `done` vs a ticker, write `heartbeatLine(...)` to `opts.Stderr`
  each tick, apply the *unchanged* exit-code policy to the captured result.
- **Injected clock:** package vars `heartbeatInterval` (default `30s`),
  `newHeartbeatTicker` (default `time.NewTicker`-backed), `sinceStart` (default
  `time.Since`) so the ticker is deterministic under test — no real sleeping.
- **PID hand-off:** `Dispatch` builds the `onStart` closure that stores the pid
  in an `atomic.Int64` the ticker reads; `Run` invokes it at `Start`.

### What stays identical (Done-when #4)

`Dispatch` still returns the full combined output string; `Classify` /
`ParseVerdict` / the sidecar all consume it unchanged. When `opts.Stderr == nil`
or the run finishes before the first tick (all current fast tests), **zero**
heartbeat lines are written — existing stdout/verdict assertions are untouched.

## Plan

- [ ] Evolve the `Run` seam to `func(ctx, onStart func(pid int), name, args...) ([]byte, error)`;
      reimplement production exec as Start → `onStart(pid)` → Wait into a
      mutex-guarded combined buffer, preserving PATH augmentation (#138) and the
      `*exec.ExitError`-swallow / launch-fail-error policy. Add a `lockedBuffer`.
- [ ] Add pure `heartbeatLine(elapsed, agent, pid)` + colocated unit test
      (elapsed/agent/pid wording; pid==0 "pending" branch).
- [ ] Add the ticker loop to `Dispatch` (gated on `opts.Stderr != nil`), with the
      injected clock package-vars; thread `opts.Agent` + pid into `heartbeatLine`.
      Keep the exit-code policy verbatim on the captured `(out, runErr)`.
- [ ] Update the 12 `judge.Run` stub sites to the new signature (add the ignored
      `onStart` param). Mechanical; no behavior change.
- [ ] Add a `Dispatch` heartbeat integration test: a fake long-running `Run` that
      calls `onStart(4242)` then blocks until released, a fake ticker firing N
      times → assert N heartbeat lines carrying elapsed + agent + pid 4242 on
      `Stderr`, final output returned intact, exit-code policy unchanged.
- [ ] Investigate provider logs during a live run; record the negative finding
      (no reliable per-invocation log path) in the issue `## Log`.
- [ ] Live-verify: run a real `sdlc close`/`milestone-close` (or `sdlc judge`)
      against a slow review and observe heartbeat lines with a real PID; confirm
      final verdict parsing + sidecar unchanged.

## Verification

- `go test ./cmd/sdlc/...` green (updated stubs + new heartbeat tests).
- One live boundary review: heartbeat lines appear on stderr every 30s with a
  real PID; `Review-Verdict:` trailer + sidecar identical to before.

## ARCH notes

- **ARCH-DRY:** heartbeat in the single `Dispatch`, reusing the already-present
  `opts.Stderr`; one evolved `Run` seam instead of a second exec path.
- **ARCH-PURE:** `heartbeatLine` pure + unit-tested; ticker/clock injected; the
  only IO is the thin `select` loop.
- **ARCH-PURPOSE:** delivers the stated purpose (live progress during the wait)
  with the reliable signal, and honestly documents why the conditional
  byte/log-tail signals are omitted rather than shipping a misleading "0 B".
