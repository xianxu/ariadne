---
id: 000140
status: working
deps: []
github_issue:
created: 2026-06-29
updated: 2026-07-01
estimate_hours:
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

- [ ] Boundary review dispatch emits periodic progress while the judge is still
      running.
- [ ] The progress output includes elapsed time and enough process/log identity
      to distinguish "alive" from "stalled".
- [ ] If a provider log can be located reliably, `sdlc` prints the path or tails
      a compact progress signal.
- [ ] Existing final review output and `Review-Verdict:` parsing remain
      unchanged.
- [ ] Tests cover the heartbeat behavior with a fake long-running judge.

## Plan

- [ ] Inspect `cmd/sdlc/internal/judge/dispatch.go` and boundary review callers.
- [ ] Determine whether supported agents expose useful progress logs.
- [ ] Add a dispatch progress callback or heartbeat wrapper around judge waits.
- [ ] Wire heartbeat output into `close` and `milestone-close`.
- [ ] Keep stdout/stderr/result parsing stable for existing verdict extraction.
- [ ] Add fake-process tests for heartbeat intervals and final output ordering.

## Log

### 2026-06-29

- Created from pair#84 dogfooding: close boundary reviews were valid but silent
  for minutes, requiring manual `ps` inspection to confirm the review child was
  still alive.
