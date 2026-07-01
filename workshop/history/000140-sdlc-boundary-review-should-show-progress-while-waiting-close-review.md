# Boundary Review — ariadne#140 (whole-issue close)

| field | value |
|-------|-------|
| issue | 140 — sdlc boundary review should show progress while waiting |
| repo | ariadne |
| issue file | workshop/issues/000140-sdlc-boundary-review-should-show-progress-while-waiting.md |
| boundary | whole-issue close |
| milestone | — |
| window | eb229a1e468d20b34a4dc2fab758b3904ad0eeeb..HEAD |
| command | sdlc close --issue 140 |
| reviewer | claude |
| timestamp | 2026-07-01T09:07:21-07:00 |
| verdict | SHIP |

## Review

I have everything I need. Build is clean, the `internal/judge` package and all #140-attributable `cmd/sdlc` tests pass under `-race`, `gofmt`/`vet` are clean, and the full-suite hang is fully explained as a pre-existing, out-of-scope environmental artifact (a repolock-acquiring test blocked by the live `sdlc close --issue 140` holding `.git/sdlc.lock`).

```verdict
verdict: SHIP
confidence: high
```

**Summary.** This is a clean, well-architected, well-tested change that delivers exactly the issue's stated purpose: `judge.Dispatch` now emits a 30s heartbeat (elapsed + agent + child PID) to `opts.Stderr` while the agent subprocess runs, and both `sdlc close` and `sdlc milestone-close` funnel through `dispatchBoundaryReview` → `Dispatch` with `Stderr` set, so both get the signal for free. The load-bearing correctness claim — that reimplementing `Run` as `Start→onStart→Wait` into a single shared `*bytes.Buffer` is byte-faithful to `CombinedOutput` — is correct: `os/exec`'s `stderr()` uses `interfaceEqual(Stderr, Stdout)` to share one child fd and one copy goroutine, so there are no concurrent buffer writes (no lock needed), and this is confirmed empirically by the `-race` run passing. The exit-code policy is preserved verbatim in the extracted `classifyRunResult`, so verdict parsing / sidecar are untouched, and the fast path (`opts.Stderr == nil`) stays synchronous. Nothing blocks SHIP.

**1. Strengths**
- `Run` seam evolution (`dispatch.go:87`): the `onStart(pid)` callback is the right shape — the PID must surface *at launch* to be live, and a return value arrives too late. Single evolved seam, not two parallel exec paths (ARCH-DRY).
- `classifyRunResult` (`dispatch.go:210`) cleanly deduplicates the exit-code policy across the sync and heartbeat paths, and the launch-failure diagnostic (#138) is preserved intact.
- `heartbeatLine` (`heartbeat.go:32`) is genuinely pure and harness-agnostic — all three fields come from `sdlc` wrapping the child, never from child output — so it reads identically for claude/codex/gemini. Clean pure/IO split (ARCH-PURE).
- Excellent test coverage: `TestRun_RealSubprocess` exercises the real exec path end-to-end (real PID, combined streams, `*exec.ExitError` on non-zero exit) — this is the test that actually catches IO bugs; `TestDispatch_HeartbeatWhileWaiting` uses a synchronous tick→observe handshake to avoid select-race flakiness; `TestDispatch_NoStderrNoHeartbeat` guards the fast path.
- Honest negative-finding documentation for the conditional Done-when #3 (no reliable per-invocation provider log; `claude -p` buffers so a byte counter reads "0 B" and misleads) rather than shipping a misleading signal (ARCH-PURPOSE).

**2. Critical findings** — none.

**3. Important findings** — none.

**4. Minor findings**
- First-beat latency: with a 30s ticker, a review under 30s emits zero heartbeats and there's a 30s silent window before the first beat. Acceptable given the `"dispatching boundary review …"` line precedes `Dispatch` and the issue's concern was multi-minute silence — noting only as a future tuning knob, not a defect.
- The in-`Dispatch` "pid pending" window (a tick firing before `onStart` stores the PID) is covered only at the pure `heartbeatLine` level, not in the integration test — fine, since `Start()` is effectively instantaneous in production.

**5. Test coverage notes**
- `internal/judge` passes under `-race`; the 12 `judge.Run` stub-site migrations are mechanical and all present (grep-confirmed — no site missed). The only production caller of `Run` is `Dispatch` (two call sites, sync + goroutine).
- Full-suite caveat (NOT a #140 issue): `go test ./cmd/sdlc/...` hangs to timeout when run concurrently with a live `sdlc` command, because `TestSetStatusAlias_BothPathsMutate` (and peers) call `buildRoot().Execute()`, which acquires the **real cwd-based repo transaction lock** rather than an isolated one. During this review the lock is held by `pid 82204: sdlc close --issue 140`. Pre-existing, outside the #140 window. Run the suite without a concurrent lock holder and it's green.

**6. Architectural notes for upcoming work**
- The heartbeat is now available to *every* `Dispatch` caller that sets `Stderr` (changecode plan-quality, preflight, `sdlc judge`), not just the boundary review — an intentional ARCH-DRY bonus. If any of those ever wants a *different* cadence or to suppress it, that'll need a per-call knob; today it's global via `heartbeatInterval`. Fine for now.
- Test-hygiene backlog (pre-existing, worth a separate issue): command-tree tests that go through `buildRoot().Execute()` acquire the real repo lock, making them non-hermetic and un-parallelizable across processes. Isolating the repo lock to a temp dir under test would remove the class of "suite hangs when a real sdlc is running" flakes I hit here.

**7. Plan revision recommendations** — none. The plan matches the code: the `Run`-seam evolution, pure `heartbeatLine` + injected clock, `classifyRunResult` extraction, 12 stub migrations, integration test, negative-finding investigation, and live-verify are all delivered as described (and the plan already records the dropped-mutex refinement from the plan-quality judge). The atlas gate is satisfied — `atlas/workflow/sdlc-binary.md` gained the "Dispatch progress heartbeat (#140)" paragraph covering the new surface.
