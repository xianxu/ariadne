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

---

## Re-review — 2026-10-17T16:43:09-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 400 — sdlc local-check cancel: stop only your own run (an agent killed another slot's suite by PID) |
| repo | ariadne |
| issue file | workshop/issues/000400-sdlc-local-check-cancel-stop-only-your-own-run-an-agent-killed-another-slot-s-suite-by-pid.md |
| boundary | whole-issue close |
| milestone | — |
| window | 78403958cf4a69d55a1fafdd76cc95b303419fa6..5575e0caf791d4f856c756e6304ef1ee8935f58c |
| command | sdlc close --issue 400 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-17T16:43:09-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

I'm reading the fix commit and its tests now. Next I'll run them and revert each fix to confirm the tests fail without it.

Tests pass. Now I'll revert each fix and run the lifetime test.

Both fixes go red when reverted. Next I'm checking whether README covers `local-check`, and how `kill` handles a pid of 0 or below.

README documents `local-check run` but has no mention of `cancel`. I'm writing up the verdict now.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

Both round-1 defects are fixed, and each fix has a regression test that I confirmed goes red when the fix is reverted. BR-1: the lease is now released right after `release()` and before `stopSignals()` (`cmd/sdlc/localcheck_run.go:165`). Removing that call makes `TestRunLocalCheckLeaseLifetime` fail with "leases held while recording". BR-2: every run now calls `sweepStaleRunLeases` before it takes its lease (`localcheck_run.go:115`). Removing that call leaves `dead.json` behind and the same test fails. BR-3 is covered by the new Revisions entry, and BR-4 by B's cancel, which now goes through `runLocalCheckCancel` from its own checkout. The lease and cancel tests pass (`go test -run 'Lease|ClassifyRunLease|CancelSweeps|CancelChecks|LocalCheckCancel|LocalCheckSignal'`: ok in 296 s), and so do the processgroup tests. One cheap Important gap is left: README.md documents `local-check run` but not the new `cancel` subcommand.

Done-when: current — the Revisions entry changes paths, drops the PGID and moves where the lease ends. "Cancel one, the other keeps running, slot released, lease gone, stale lease swept" still describes what is built and tested.

1. **Strengths**
   - The lease is published already locked (write and lock under `.lease-*.tmp`, then rename; `localcheck_lease.go:79-108`). This removes the "visible but unlocked reads as stale" race, and `TestHoldRunLeaseLockedBeforeVisible` stress-tests it.
   - Ownership is held, not observed. A kernel-released flock plus a kernel start-time match (`processgroup.StartTime`, read via sysctl or `/proc`, never `ps`) replaces hand-picked PIDs.
   - `classifyRunLease` is a small pure function with table tests. The shell around it is thin.
   - The lease now spans exactly the stretch where SIGTERM is owned, and `leaseProbeGH` checks that from inside `PostStatus`.
   - The two-checkout process test uses real `runLocalCheck` children, real flock and real signals.

2. **Critical:** none.

3. **Important**
   - README.md (around lines 464-473) is missing the new `sdlc local-check cancel` subcommand. Its local-check paragraph covers only `run`. Fix: add one sentence, e.g. "stop your own run with `sdlc local-check cancel [--name N]`, never by a PID from `ps`."

4. **Minor**
   - **Residual race:** cancel finds the lock held, checks the start time, then calls `Kill`. If the run releases its lease and calls `stopSignals` in that gap, the default SIGTERM action kills sdlc mid-record or mid-merge. The window is microseconds. It is the inverse of BR-1 rather than a repeat of it, so I'm noting it, not raising it. The residual note in the Spec should name it next to the PID-reuse residual.
   - **ARCH-SECURE:** `rec.PID` comes from a directory every slot shares, and nothing rejects `pid <= 1`. `syscall.Kill(0 or -1, …)` would signal a process group or every process. The start-time match and the held-lock-plus-root check make this hard to reach, but reject `PID <= 1` when parsing.
   - `cancelLocalChecks` still has its own `leaseStale` removal branch right after `sweepStaleRunLeases`. It is harmless (it covers leases released between the two scans).

5. **Test coverage notes**
   - Both BR-1 and BR-2 fail without their fix (verified).
   - Slot release on cancel isn't asserted directly. But cancel returns only once the lease is released, which now happens after `release()`, so a successful cancel implies the slot was freed. That is acceptable.
   - No test can control the ordering of the cancel-versus-run-end race (ARCH-ORDER). It is a single-interleaving gap, noted above.

6. **Architectural notes**
   - ARCH-DRY: pass. The sweep is now one shared helper.
   - ARCH-PURE: pass. `classifyRunLease` is pure, and the shell is thin.
   - ARCH-PURPOSE: pass. The verb, the lease, the attribution via `WF_TEST_LOCK_OWNER` and the rule are all delivered. The merge path shares `runLocalCheck`.
   - ARCH-MOCK: pass. Tests use real flock and processes, and GitHub goes through the existing fake.
   - ARCH-CONSTRAINTS: pass. The wait is bounded at 2 minutes, and a scan touches a handful of files.
   - ARCH-SECURE: pass, apart from the pid-bound Minor. Foreign leases are never signalled, and a malformed held lease is left alone and reported.
   - ARCH-ORDER: pass, with the Minor above. The lease nests strictly inside signal ownership.
   - ARCH-FUNERAL: pass. A lease is removed on release, every run and cancel sweeps dead ones, and leftover `.tmp` files are removed after 24 h.

7. **Plan revisions:** none beyond adding the cancel/run-end race to the Spec's residual note.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      releaseLease() now precedes stopSignals() (localcheck_run.go:165); reverting it makes TestRunLocalCheckLeaseLifetime fail with "leases held while recording".
  - id: BR-2
    disposition: addressed
    note: |
      runLocalCheck calls sweepStaleRunLeases before holdRunLease; reverting it leaves dead.json and TestRunLocalCheckLeaseLifetime fails.
  - id: BR-3
    disposition: addressed
    note: |
      The Revisions entry (2026-10-17) records local-check-runs/ and pid-only SIGTERM with no PGID, matching localCheckRunsDir and cancelLocalChecks.
  - id: BR-4
    disposition: addressed
    note: |
      TestLocalCheckCancelOnlyOwnRun now cancels B through runLocalCheckCancel after t.Chdir(rootB); cancel returning implies release() ran, since the lease now ends after it.
findings:
  - id: new
    severity: Important
    kind: defect
    family: user-surface-docs-gate
    title: |
      README update appears missing for the new `sdlc local-check cancel` subcommand
    detail: |
      README.md's local-check paragraph (around lines 464-473) documents only `local-check run`; the new user-typed `cancel` verb, the way to stop your own run instead of killing a PID from ps, is absent. Add one sentence there.
  - id: new
    severity: Minor
    kind: defect
    family: untrusted-lease-fields-unvalidated
    title: |
      Cancel passes the shared-directory lease's pid to syscall.Kill without bounding it
    detail: |
      rec.PID is read from a directory every slot shares; pid 0 or -1 would signal a process group or every process. The start-time and held-lock checks make this hard to reach, but reject PID <= 1 when parsing the lease (ARCH-SECURE).
  - id: new
    severity: Minor
    kind: defect
    family: lease-outlives-signal-ownership
    title: |
      A cancel that saw the lock held can SIGTERM the run just after it dropped signal ownership
    detail: |
      This is the inverse of BR-1, not a repeat: between cancel's flock probe and Kill, the run can call releaseLease and stopSignals, so the default action kills sdlc during recording or merge. The window is microseconds. Name it in the Spec's residual note beside PID reuse; no ordering seam tests it (ARCH-ORDER).
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"fa56c63f-3b2c-44ae-acf8-fcb5445e8ee8","repository":"github.com/xianxu/ariadne","issue_id":"000400","recipe":"milestone-review","round":2,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/7606c4fc25496e37","reviewed_head":"5575e0caf791d4f856c756e6304ef1ee8935f58c","slot":"ariadne:2","requested_at":"2026-10-17T23:36:02.338205Z"},"phase":"finished","pid":60911,"started_at":"2026-10-17T23:36:02.543868Z","finished_at":"2026-10-17T23:43:09.158091Z","session_id":"84330501-9939-4150-af12-4a88b2748797","outcome":"returned"}] -->

---

## Re-review — 2026-10-17T17:18:02-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 400 — sdlc local-check cancel: stop only your own run (an agent killed another slot's suite by PID) |
| repo | ariadne |
| issue file | workshop/issues/000400-sdlc-local-check-cancel-stop-only-your-own-run-an-agent-killed-another-slot-s-suite-by-pid.md |
| boundary | whole-issue close |
| milestone | — |
| window | 78403958cf4a69d55a1fafdd76cc95b303419fa6..c36adf9fb0c7378c3922916dfe8506985e6b0ace |
| command | sdlc close --issue 400 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-17T17:18:02-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Verdict is FIX-THEN-SHIP. Two of the three open findings are fixed; the third, a Minor, has a test that passes even with its fix removed.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

`sdlc local-check cancel` now finds runs only through a lease file that each run holds locked for as long as SIGTERM is its own. It checks that the lease names this checkout and that the pid's kernel start time still matches before it signals anything. I ran the six lease, cancel and lifetime tests at the pinned head and they pass (242 s). BR-5 (the README sentence) and BR-7 (the residual race now named in the Spec) are both done. BR-6 is not: the `rec.PID <= 1` guard at `cmd/sdlc/localcheck_lease.go:162` is in place, but I replaced it with `if false && …` in this disposable checkout, re-ran `TestCancelChecksProcessStartTime`, and it still passed. Pids -1, 0 and 1 also fail the start-time check, and both refusal messages contain "not signalled", so the test never isolates the guard. I restored the file afterwards. BR-6 is a Minor, so it does not block shipping; nothing else new came up.

Done-when: current — it still requires the two-checkout process test, the `classifyRunLease` table test, the test-lock owner case and the help phrase guards. The Revisions only change where the lease lives (a path Done-when never names) and when it ends, which `TestRunLocalCheckLeaseLifetime` covers on its own.

**1. Strengths**
- The lease is written and locked under a temporary name and only then renamed into place (`localcheck_lease.go:77-106`). So no reader ever sees a live run's lease unlocked, and the concurrent-sweep test (`TestHoldRunLeaseLockedBeforeVisible`) checks this.
- The lease is released exactly where the run gives up SIGTERM ownership (`localcheck_run.go:163-165`), and `TestRunLocalCheckLeaseLifetime` checks that no lease is held while the status posts.
- Process start time comes from the kernel (sysctl on darwin, `/proc/<pid>/stat` on linux), not from `ps`, so it works inside the agent sandbox. On unsupported platforms it refuses rather than guessing (`starttime_other.go`).
- `classifyRunLease` is pure and table-tested; the scan, lock and kill code around it stays thin.
- `TestLocalCheckCancelOnlyOwnRun` runs real `runLocalCheck` child processes from two checkouts of one clone and shows cancel stopping only its own.

**2. Critical:** none.

**3. Important:** none.

**4. Minor**
- BR-6 is still open (details in the findings block). The fix: in the pid ≤ 1 loop of `TestCancelChecksProcessStartTime`, stub the `processStartTime` variable to return `rec.PIDStarted`, so only the pid guard can refuse. Then assert the guard's own message, `names pid %d; not signalled`, not the shared phrase "not signalled". Do not run that case against real pid -1 if the guard is ever removed: `kill(-1, SIGTERM)` signals every process you own.

**5. Test coverage notes**
- Without the pid guard, the -1/0/1 loop passes anyway; with the stub above it would fail.
- The race in BR-7 has no ordering seam to test it. That is accepted and documented in the Spec.

**6. Architecture**
- ARCH-DRY: pass.
- ARCH-PURE: pass.
- ARCH-PURPOSE: pass. The tl-worker rule, the help text, the queue's held-by attribution and the verb are all delivered.
- ARCH-MOCK: pass. Tests use real processes and real flock; there is no external service.
- ARCH-CONSTRAINTS: pass. The wait is bounded at 2 minutes and each scan touches a handful of files.
- ARCH-SECURE: pass, apart from BR-6 not being proven by a test. Leases are re-checked by lock state, checkout root and start time before any signal.
- ARCH-ORDER: pass. The lock's kernel lifetime is the state, and the remaining race is named in the Spec.
- ARCH-FUNERAL: pass. A run removes its lease on release; every run and every cancel sweeps unlocked leases; temporary names are swept after the stale age.

**7. Plan revisions:** none needed.

```findings
dispose:
  - id: BR-5
    disposition: addressed
    note: |
      README.md:472-473 now names `sdlc local-check cancel` beside local-check run and warns against PIDs from ps.
  - id: BR-6
    disposition: not-addressed
    note: |
      The guard exists (localcheck_lease.go:162), but with it disabled TestCancelChecksProcessStartTime still passes, because pids -1/0/1 fail the start-time check and both messages say "not signalled". Stub processStartTime to return rec.PIDStarted in that loop and assert the guard's own message.
  - id: BR-7
    disposition: addressed
    note: |
      The issue's Design residual bullet now names the race where a cancel lands just after the run released its lease, beside PID reuse; prose-only, as the finding asked.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"753827d3-50e9-413f-b335-46424d66c350","repository":"github.com/xianxu/ariadne","issue_id":"000400","recipe":"milestone-review","round":3,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/6b373cd19d584dc1","reviewed_head":"c36adf9fb0c7378c3922916dfe8506985e6b0ace","slot":"ariadne:2","requested_at":"2026-10-18T00:12:24.950019Z"},"phase":"finished","pid":60421,"started_at":"2026-10-18T00:12:25.188093Z","finished_at":"2026-10-18T00:18:01.539417Z","session_id":"7bd0709c-7ea3-45e0-8717-8da84e2a88ed","outcome":"returned"}] -->
