# Boundary Review — ariadne#400 (whole-issue close)

| field | value |
|-------|-------|
| issue | 400 — sdlc local-check cancel: stop only your own run (an agent killed another slot's suite by PID) |
| repo | ariadne |
| issue file | workshop/issues/000400-sdlc-local-check-cancel-stop-only-your-own-run-an-agent-killed-another-slot-s-suite-by-pid.md |
| boundary | whole-issue close |
| milestone | — |
| window | 78403958cf4a69d55a1fafdd76cc95b303419fa6..7d17e9481d46572fc438a665404b73397230ee43 |
| command | sdlc close --issue 400 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-17T15:47:48-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

The lease and cancel tests pass. Now I'm reading the two-checkout process test, then I'll write up the verdict.

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

The lease mechanism is sound and well tested: one held lease per run, named by its checkout, and cancel only signals a lease that is both held and its own. `classifyRunLease` is a small pure function. The lease is locked before it becomes visible, which closes the creation race. Before signalling, cancel checks the pid's start time as read from the kernel. The lease and cancel tests and the processgroup tests pass at head. One real gap stops this from being a clean SHIP. `runLocalCheck` hands SIGTERM back to the default handler (`stopSignals()`) as soon as the check finishes. It still holds the lease while it saves the attempt, posts the GitHub status and PR comment, and removes the worktree. A cancel that lands in that window kills the process outright instead of taking the managed path, which contradicts the help text and the code comment. The other finding is a sweep the design promised and the code doesn't do: `run` was meant to sweep stale leases too, and only `cancel` does.

### 1. Strengths
- **Lease is locked before anyone can see it.** `cmd/sdlc/localcheck_lease.go:78-105` writes and locks the lease under a `.lease-*.tmp` name, then renames it into place. So cancel never sees a live run's lease unlocked. `TestHoldRunLeaseLockedBeforeVisible` stress-tests this.
- **Pure classification.** `classifyRunLease` (`localcheck_lease.go:45-55`) is a pure function with a table test covering all four verdicts (ARCH-PURE pass).
- **Start time comes from the kernel, not `ps`.** On darwin it uses sysctl, on linux `/proc/<pid>/stat`, and other platforms refuse to signal. This avoids the sandbox `fork/exec /bin/ps` failure. On darwin, an absent pid is mapped correctly to `ErrNoProcess`.
- **Real end-to-end test.** `TestLocalCheckCancelOnlyOwnRun` runs two real `runLocalCheck` child processes from two checkouts of one clone. It checks that A stops on the managed interruption path and A's check process dies, while B and its check keep running, and that B can then cancel its own run.
- **Leases live outside the pruned log directory.** They sit in `local-check-runs/`, beside the logs rather than among them, so `pruneLocalCheckLogs` can never delete a live run's lease. That was a deliberate and correct choice.

### 2. Critical findings
None.

### 3. Important findings
- **The lease outlives SIGTERM handling (ARCH-ORDER).**
  - **Where:** `cmd/sdlc/localcheck_run.go:113-121` takes the lease after `signal.NotifyContext`. But `stopSignals()` is called right after the check exits (around `localcheck_run.go:163`), and the lease is only released in the last-run defer.
  - **What happens:** everything after that point runs with the default SIGTERM disposition but a held lease: `hashAndTail`, `publishSuiteAttempt`, pruning, `GH.PostStatus` / `PRComment` (network calls), and worktree cleanup. A `local-check cancel` in that window kills the sdlc process mid-write. The status may already be posted, the temp worktree is left behind, and in the `sdlc merge` path (`cigate.go:399`) the whole merge dies.
  - **Contracts broken:** the code comment "cancel always lands on the managed interruption path" and the help text "removes its worktree, recording nothing" are both false for this phase. Cancel also reports "cancelled local check" for a run that actually completed and was recorded.
  - **Fix sketch:** release the lease at the same point signals stop being owned (call `releaseLease()` just before `stopSignals()`, keeping the defer for early returns). Or keep SIGTERM ignored/owned until the lease is released. Then add a test that cancels after the check exits but before the record is posted.
- **`run` never sweeps stale leases (ARCH-FUNERAL / plan drift).** The design says "a crashed run leaves an unlocked file that the next `cancel` or `run` sweeps". Only `cancelLocalChecks` sweeps. Stale leases and `.tmp` leftovers from SIGKILLed runs pile up until someone happens to run cancel, which is a rare, operator-triggered command. The fix is cheap: run the same sweep (unlocked `*.json` files and old `.tmp` files) in `holdRunLease` or at the start of `runLocalCheck`, ideally by factoring it out of `cancelLocalChecks`.

### 4. Minor findings
- **CLI wiring is untested.** Tests call `cancelLocalChecks` directly, so `runLocalCheckCancel` (`localcheck_cmd.go:61-76`: top-level resolution and the common-dir path) is never exercised.
- **Done-when says "slot released", but the process test doesn't check it.** It uses the `"skipped"` lock helper, so no real slot is held or freed.
- **The lease directory moved.** The design names `local-checks/runs/`; the code uses `local-check-runs/`. The Log explains why, but the Design text is stale.
- **Spec wording doesn't match the implementation.** The Spec bullet still says "signals its process group" and "records PGID". The design replaced this with SIGTERM to the run's own pid, so the Spec text should be updated to match.
- **Help text overclaims.** `local-check.md` says cancel stops runs "whether queued or running". That's true only from the moment the lease is taken; prep and worktree creation happen before it, and a run in that phase can't be found.

### 5. Test coverage notes
- Leases, the creation race, stale-lease sweeping without a signal, the start-time mismatch, and two-checkout isolation are all covered with real processes and real flock. These tests pass when run locally at head.
- Missing: a cancel that lands after the check exits (the Important bug above), the CLI entry point, and leases swept by `run`.
- `TestStartTime` checks that its own start time is stable and that an absent pid is handled, which is reasonable for a kernel-reading helper.

### 6. Architecture notes
- **ARCH-DRY:** pass. It reuses `internal/flock`, `processgroup`, and `localCheckLogDir`.
- **ARCH-PURE:** pass. `classifyRunLease` is pure, and the scan/lock/kill shell is thin.
- **ARCH-PURPOSE:** pass. Cancel exists, the tl-worker rule is phrase-guarded in `rules.tsv`, and held-by attribution through `WF_TEST_LOCK_OWNER` addresses the root cause.
- **ARCH-MOCK:** pass. Real flock and real processes are used; no external service is involved.
- **ARCH-CONSTRAINTS:** pass. The wait is bounded at 2 minutes and the scan is small.
- **ARCH-SECURE:** pass. Lease JSON comes from a shared directory and is treated as untrusted: a malformed lease held by a live run is reported and left alone, and only a held lease that is ours with a matching start time gets signalled.
- **ARCH-ORDER:** flag. The lease's lifetime and the SIGTERM-ownership lifetime are not aligned; see Important #1.
- **ARCH-FUNERAL:** flag. Leases left by crashed runs are only swept by an operator-run cancel; see Important #2.

### 7. Plan revision recommendations
- Add a `## Revisions` entry: the lease directory is `<common dir>/sdlc/local-check-runs/` (not `local-checks/runs/`), and cancel SIGTERMs the run's pid (no PGID is recorded). This brings the Spec bullet in line.
- Once fixed, record in Revisions that the lease is released when signal ownership ends, and that `run` sweeps stale leases.

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: lease-outlives-signal-ownership
    title: |
      Run lease stays held after stopSignals, so cancel during record/post/cleanup kills sdlc by default SIGTERM
    detail: |
      localcheck_run.go calls stopSignals() right after the check exits but releases the lease only in the last defer; publishSuiteAttempt, GH.PostStatus/PRComment and worktree removal run with default SIGTERM while the lease is held, so a cancel there kills the process (merge path included), leaks the worktree and may half-record, contrary to the help text and the code comment. Release the lease where signal ownership ends (or keep SIGTERM owned until release) and add a post-check cancel test.
  - id: new
    severity: Important
    kind: defect
    family: artifact-removal-path-undelivered
    title: |
      Stale leases are swept only by cancel, not by run as the design states
    detail: |
      Design: a crashed run's lease is swept by the next cancel or run. Only cancelLocalChecks sweeps, so leases and .tmp leftovers from SIGKILLed runs persist until someone runs cancel. Factor the sweep out and call it from holdRunLease/runLocalCheck.
  - id: new
    severity: Minor
    kind: defect
    family: plan-text-drift
    title: |
      Spec/Design text still says local-checks/runs/ and process-group signalling with PGID
    detail: |
      The code uses local-check-runs/ and SIGTERMs the run pid without recording a PGID; add a Revisions entry.
  - id: new
    severity: Minor
    kind: defect
    family: cli-wiring-untested
    title: |
      runLocalCheckCancel (top-level resolution plus runs dir) has no test; the slot-release Done-when is not asserted
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"13c0eaf4-0fae-4dff-930f-7f171f1c9c44","repository":"github.com/xianxu/ariadne","issue_id":"000400","recipe":"milestone-review","round":1,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/a29bae60eb39ed26","reviewed_head":"7d17e9481d46572fc438a665404b73397230ee43","slot":"ariadne:2","requested_at":"2026-10-17T22:40:28.740704Z"},"phase":"finished","pid":50590,"started_at":"2026-10-17T22:40:28.869159Z","finished_at":"2026-10-17T22:47:38.293164Z","session_id":"6a9844fa-865f-4865-a495-5134587059cd","outcome":"returned"}] -->
