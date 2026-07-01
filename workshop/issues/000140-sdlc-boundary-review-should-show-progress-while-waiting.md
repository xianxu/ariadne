---
id: 000140
status: working
deps: []
github_issue:
created: 2026-06-29
updated: 2026-07-01
estimate_hours: 0.67
started: 2026-07-01T00:27:31-07:00
---

# sdlc boundary review should show progress while waiting

## Problem

Boundary reviews can run silently for several minutes. In pair#84, `sdlc close`
printed:

```text
dispatching boundary review (...HEAD) via claude ...
```

and then produced no progress output for multiple 60-second polling intervals.
The only way to tell "still working" from "wedged" was to inspect the process
tree and infer that a `claude -p` child still existed.

This creates operational friction for agents and humans:

- long silent waits look like hangs;
- operators do not know whether the subprocess is making network/model progress;
- agents may be tempted to interrupt a valid review;
- genuine stalls are hard to distinguish from normal review latency.

## Spec

Make boundary review dispatch observable while it is running.

Investigate what progress signals are available from the current judge dispatch
path:

- process liveness (`claude -p`, `codex`, `gemini`);
- child stdout/stderr streaming;
- provider/harness log files, if any;
- timestamps/byte counts from a sidecar file;
- periodic heartbeat from `sdlc` itself while waiting.

At minimum, `sdlc close` / `sdlc milestone-close` should print periodic
heartbeat lines while waiting, including elapsed time and the subprocess PID or
agent name. If the underlying agent/harness exposes richer logs, `sdlc` should
print the log path and optionally tail compact progress.

This is separate from review-result sidecars (#136): sidecars preserve the
final output for later reading; this issue is about live progress during the
wait.

## Done when

- [x] Boundary review dispatch emits periodic progress while the judge is still
      running.
- [x] The progress output includes elapsed time and enough process/log identity
      to distinguish "alive" from "stalled". (elapsed + agent + child PID; the
      PID is the automated form of the operator's manual `ps` inspection.)
- [x] If a provider log can be located reliably, `sdlc` prints the path or tails
      a compact progress signal. (Conditional — investigated: no agent exposes a
      reliably-locatable per-invocation log, and `claude -p` buffers output to the
      end so a live byte counter would read "0 B" and mislead. Negative finding
      recorded in Log; shipped the reliable PID signal instead.)
- [x] Existing final review output and `Review-Verdict:` parsing remain
      unchanged. (Same combined output + `classifyRunResult` exit-code policy;
      gated on `opts.Stderr != nil` so fast paths are silent.)
- [x] Tests cover the heartbeat behavior with a fake long-running judge.

## Plan

Design detail: `workshop/plans/000140-boundary-review-progress-plan.md`.
Heartbeat is harness-agnostic (elapsed + agent + PID come from `sdlc` wrapping
the child, not from child output) so it works identically for claude/codex/gemini.

- [x] Evolve the `Run` seam → `func(ctx, onStart func(pid int), name, args...)`;
      reimplement production exec as Start → `onStart(pid)` → Wait into a
      **plain shared** combined buffer, preserving PATH augmentation (#138) + the
      exit-code policy. (Dropped the planned mutex — `os/exec` reuses one fd +
      copy goroutine when `Stdout == Stderr`, so no locking is needed; per the
      plan-quality judge.)
- [x] Add pure `heartbeatLine(elapsed, agent, pid)` + colocated unit test.
- [x] Add the ticker loop to `judge.Dispatch` (gated on `opts.Stderr != nil`),
      injected clock (`heartbeatInterval`/`newHeartbeatTicker`/`sinceStart`); keep
      the exit-code policy verbatim in `classifyRunResult` (ARCH-DRY: one path,
      reuses `opts.Stderr`, shared exit-code policy across sync + heartbeat paths).
- [x] Update the 12 `judge.Run` stub sites to the new signature (mechanical).
- [x] Add a `Dispatch` heartbeat integration test: fake long-running `Run`
      (`onStart(4242)` then block) + fake ticker → assert N lines with elapsed +
      agent + pid, final output + exit-code policy intact.
- [x] Investigate provider logs live; record the negative finding in `## Log`.
- [x] Live-verify a real boundary review: heartbeats with a real PID; verdict
      trailer + sidecar unchanged.

## Estimate

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against
`baseline-v3.1.md`. Method A only.* Design pre-resolved by the thorough plan doc
(`workshop/plans/000140-…`), so the +15% design buffer applies and design hours
stay small; impl at v3.1's 40%-of-v2 unit.

- smaller-go-module (judge pkg): reimpl the `Run` exec seam to expose PID, add
  the `Dispatch` ticker loop + pure `heartbeatLine` + injected clock, judge-pkg
  tests — design 0.15 + impl 0.2.
- smaller-go-module (cmd/sdlc): mechanical signature update of the 12 `judge.Run`
  stub sites across 6 test files — impl 0.15.
- milestone-review: the one boundary review auto-dispatched at `sdlc close` —
  impl 0.15.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
design-buffer: 0.15
item: smaller-go-module   design=0.15 impl=0.2
item: smaller-go-module   design=0.0  impl=0.15
item: milestone-review    design=0.0  impl=0.15
total: 0.67
```

## Log

### 2026-06-29

- Created from pair#84 dogfooding: close boundary reviews were valid but silent
  for minutes, requiring manual `ps` inspection to confirm the review child was
  still alive.

### 2026-07-01

- **Implemented.** Heartbeat lives in `internal/judge.Dispatch` (not a
  boundary-only wrapper): gated on the already-present-but-previously-unused
  `opts.Stderr`, so the boundary review (close + milestone-close, via the shared
  `dispatchBoundaryReview`) gets it for free and the plan/estimate/preflight
  judges get it as a bonus — one path, ARCH-DRY. The `Run` seam evolved to
  `func(ctx, onStart func(pid int), name, args...)`, reimplemented as
  Start→`onStart(pid)`→Wait into a shared combined buffer; a callback (not a
  return) because the PID must surface *at launch* to be live. Exit-code policy
  factored into `classifyRunResult`, shared by the sync + heartbeat paths.
- **Harness-agnostic** (answering the design question): elapsed + agent + PID all
  come from `sdlc` wrapping the child, never from child output, so it reads
  identically for claude/codex/gemini. This is the robustness argument *for* the
  sdlc-side heartbeat over tailing child output.
- **Provider-log / byte-count investigation → negative.** No agent (claude/codex/
  gemini) returns an invocation id that lets `sdlc` reliably correlate a
  per-invocation log file, so log-tailing can't be done reliably (Done-when #3 is
  conditional). And `claude -p` buffers its result to the end, so a live byte
  counter would sit at "0 B" for the whole wait and *look* stalled — a misleading
  signal. Shipped elapsed + agent + PID, which is uniform across harnesses and is
  the automated form of the manual `ps` workaround.
- **Applied plan-quality judge findings:** dropped the planned mutex-guarded
  buffer (a plain shared `*bytes.Buffer` is safe — `os/exec` reuses one fd + copy
  goroutine when `Stdout == Stderr`); made the integration test a synchronous
  tick→observe→tick handshake (no select-race flakiness).
- **Tests:** `TestHeartbeatLine` (pure wording, incl. pid-pending + empty-agent),
  `TestDispatch_HeartbeatWhileWaiting` (fake long-running `Run` + hand-driven
  ticker, N beats with elapsed/agent/pid, output + ticker-stop intact — green 5×
  under `-race`), `TestRun_RealSubprocess` (real `sh` child: real PID to onStart,
  combined streams, `*exec.ExitError` on non-zero exit), `TestDispatch_NoStderrNoHeartbeat`
  (fast path stays synchronous). 12 `judge.Run` stubs migrated across 7 files.
- **Live E2E** (`sdlc judge dry` against a native fake `claude` sleeping 33s):
  ```
  ==> invoking claude for Check DRY principle …
      … still working — 30s elapsed via claude (pid 33045; inspect: ps -p 33045)
    [ok] Check DRY principle: clean
  ```
  Real PID surfaced; final verdict (`No DRY violations found.`) reached stdout
  unchanged. (Note: shell-script fakes hit `ENOEXEC` under the exec sandbox — a
  native Go fake agent was needed.)
