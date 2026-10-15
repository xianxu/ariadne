# Boundary Review — ariadne#383 (whole-issue close)

| field | value |
|-------|-------|
| issue | 383 — sdlc merge holds the machine suite slot until the process exits, through GitHub polling, the PR merge and archiving |
| repo | ariadne |
| issue file | workshop/issues/000383-sdlc-merge-holds-the-machine-suite-slot-until-the-process-exits-through-github-polling-the-pr-merge-and-archiving.md |
| boundary | whole-issue close |
| milestone | — |
| window | ef5e0bf028c33b130d93b7014be47d310812fc15..18ce936d5a3e35dbec0885ef61c9fc80fd5be496 |
| command | sdlc close --issue 383 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-14T23:21:10-07:00 |
| verdict | SHIP |

## Review

Diff and helper read; running the new tests in the checkout now.

```verdict
verdict: SHIP
confidence: high
```

This change does what #383 asked for. `acquireSuiteSlot` now starts a separate `cat` process for each run and passes its pid to the lock helper as the process to watch. `runLocalCheck` frees the slot right after the suite's process group is terminated, so the log hashing, the status post and everything `sdlc merge` does next run without the slot. A `defer` plus a `sync.Once` cover early returns and repeat calls. If `sdlc` dies, the kernel closes the pipe, `cat` exits and the slot is freed. The per-process `suiteLockOwners` cache is gone, so a second run in the same process now waits its turn instead of passing straight through. I ran the two new tests and the existing reentry tests at head, and they pass. I then put back the old `localcheck_run.go` (only adjusting it to compile against the new signature), and both new tests fail: the waiter during the GitHub phase gets `timeout`, and run 2 doesn't wait in line. Nothing blocks shipping; there are two Minor notes.

1. **Strengths**
   - `cmd/sdlc/localcheck_run.go:201-241`: the holder process's three states (started, holding, released) match the plan's ARCH-ORDER table. Every answer from the helper that isn't `acquired` frees the holder at once, and the error path frees it before returning.
   - `localcheck_run.go:112-114`: the slot is freed after `processgroup.Terminate`, so leftover suite children can't run outside the slot.
   - Go opens the pipe's write end close-on-exec, so neither the suite nor the forked watcher in the Python helper inherits it. The crash-safety claim holds.
   - The tests use the real `test-lock.py` and real flocks. The fixture Makefile (`localcheck_reentry_test.go:27`) checks inside every run that the nested make sees `reentrant`, which also covers "run 2's nested make sees its own holder".
   - `acquireSuiteSlot` has one caller (`localcheck_run.go:81`), so there's no other site where the slot could be held past the run.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - The Done-when clause says the slot is free "while `sdlc merge` polls GitHub and merges". The test probes this only inside `runLocalCheck`'s `PostStatus`. Since there's a single caller, every later merge step comes after that point, so the claim holds by structure; it just isn't measured in `merge` itself.
   - On Linux, `wait_for_exit` checks for the holder's exit every 0.5 s rather than using kqueue. So "frees the slot before the status post" (`atlas/workflow/sdlc-binary.md:1743`, and the comment at `localcheck_run.go:112`) is true on macOS but up to about 0.5 s late on Linux. This is a wording issue, not a correctness one.

5. **Test coverage**
   - Both new tests go red without the fix (checked by reverting it).
   - The "slot is held during the suite" assertion added by the TL is covered: the probe inside the suite gets `timeout`.
   - Waiting in line on a rerun is covered with an outside holder present (`errUnboundedSuite`) and without one (passes, nested reentrant).
   - The existing tests `TestSuiteBaselineRetainsRealLockOwner` and `TestRepeatedLocalChecksRetainRealLockOwner` still pass.

6. **Architecture**
   - **ARCH-DRY: pass.** The removed cache was the duplicated ownership state. `withoutEnv` is a small helper used only in tests, and I found no existing equivalent.
   - **ARCH-PURE: pass.** The code is entirely in the IO layer (process and lock lifetime). There's no business logic to pull out, and real-helper tests are the right way to test it.
   - **ARCH-PURPOSE: pass.** All three Spec bullets are delivered, with nothing deferred. Docs were updated in both atlas files the plan names. A README update isn't needed because there's no new user-facing command, flag or config key.

7. **Plan revisions:** none; the plan matches the code.

```findings
findings:
  - id: new
    severity: Minor
    kind: defect
    family: done-when-measured-at-proxy
    title: |
      The "slot is free while merge polls and merges" clause is measured only at runLocalCheck's PostStatus
    detail: |
      TestSuiteSlotCoversOnlyTheRun probes inside PostStatus, not inside sdlc merge's CI poll or PR merge. It holds by structure because acquireSuiteSlot has a single caller, localcheck_run.go:81, and the slot is freed before that function returns. This is the only instance in the window.
  - id: new
    severity: Minor
    kind: defect
    family: doc-claim-platform-timing
    title: |
      Docs say the slot is freed "before the status post"; on Linux the watcher frees it up to about 0.5s later
    detail: |
      test-lock.py wait_for_exit uses kqueue on macOS but checks kill(pid,0) every 0.5s elsewhere. Instances: atlas/workflow/sdlc-binary.md:1743 and the comment at cmd/sdlc/localcheck_run.go:112. Consider saying the release is started before the status post.
```

## Review preparation and protocol diagnostics

Quoted values are diagnostic data, not review verdicts.
Protocol attempt 1
reason: "missing valid findings block"
output: "Reading the diff and the helper is done; next I'll check the callers and run the new tests.\n\n```verdict\nverdict: SHIP\nconfidence: high\n```\n\n**Summary.** The change does what the issue asks. `acquireSuiteSlot` no longer passes `os.Getpid()` as `--watch`. Each run now starts its own holder, a `cat` process reading a pipe from `sdlc` in its own process group, and passes that pid instead. `runLocalCheck` releases the holder right after the suite's process group is terminated. A `defer` also releases it on every early return.\n\nAll three callers get the new behaviour because each one goes through `runLocalCheck`: `cigate.go:399` (merge), `localcheck_cmd.go:117` and `suitebaseline.go:27`. `suiteLockOwners` and `suiteLockScope` are fully removed and nothing else references them.\n\nI checked the change in three ways:\n- **New tests pass at head.** `go vet` is clean, and the two new tests pass along with the existing `RealLockOwner` reentry tests.\n- **Fix reverted, tests fail.** With `localcheck_run.go` put back to base, both new tests fail for the expected reasons. In `TestSuiteSlotCoversOnlyTheRun`, the GitHub-phase waiter answers `timeout`. `TestSuiteRerunQueuesForTheSlot` also fails. I restored the file afterwards and the working tree is clean.\n- **Lock helper checked.** I read `scripts/test-lock.py`. The watcher's `wait_for_exit` uses kqueue `NOTE_EXIT` on macOS and polls every 0.5 s elsewhere. On the `reentrant` path, the holder is always a fresh pid, so it can never match a slot record. That means `reentrant` really does mean \"nested under an inherited `WF_TEST_LOCK_HELD_BY`\", so dropping the old `return pid` fallback is correct.\n\nNothing blocks SHIP.\n\n1. **Strengths**\n   - `localcheck_run.go:201-242`: the holder lifecycle is small and complete. `sync.Once` makes release idempotent. Every non-`acquired` outcome (`reentrant`, `skipped`, timeout, helper error) releases at once, and the `defer` at `:85` covers early returns between acquire and run.\n   - The write end of the pipe comes from `os.Pipe` and is close-on-exec. So neither the suite nor the helper inherits it, and if `sdlc` dies, the kernel closes it and the slot is freed. There is no change to `test-lock.py`, which matches the Spec (\"no change to admission order, slot counts or the suite command\").\n   - `localcheck_reentry_test.go:150-198`: the tests use the real helper and real flocks, not mocks. The `PROBE_OUT` recipe also covers the TL's added requirement that no waiter is admitted while the suite runs. The fixture Makefile's first line insists on `reentrant`, which checks that `WF_TEST_LOCK_HELD_BY` is the holder's pid.\n   - The release happens before `hashAndTail`, `PostStatus` and `PRComment` (`:112-114`), so the `blockingPostGH` probe correctly marks where the slot is free.\n\n2. **Critical:** none.\n\n3. **Important:** none.\n\n4. **Minor:** none worth a finding. Two notes:\n   - On Linux the watcher polls, so the slot is freed up to about 0.5 s after `release()` rather than immediately. That's harmless.\n   - If the suite starts a daemon that escapes its process group (via `setsid`), that daemon now outlives the slot. Before this change it was covered until `sdlc` exited. That's out of scope for this issue.\n\n5. **Test coverage notes**\n   - Done-when 1 is tested at the `runLocalCheck` boundary (a blocked `PostStatus`), not through `sdlc merge` end to end. That's enough, because every step of merge after the suite (CI poll, re-integration, PR merge, archiving) runs after `runLocalCheck` returns, and by then the slot is released.\n   - Done-when 2, both cases:\n     - **Rerun while an outsider holds the slot:** it queues and times out.\n     - **Rerun once the outsider has left:** it acquires, and its nested make sees `reentrant`.\n   - The baseline-diagnosis rerun is covered by the existing `TestSuiteBaselineRetainsRealLockOwner`, which still passes.\n   - Done-when 3: both atlas docs (`ci-merge-check.md`, `sdlc-binary.md`) were updated, and their claims match the code. These are the pipe/EOF crash path, the rerun taking a new ticket, and `HELD_BY` being the holder's pid. No README change is needed because there is no new flag or command.\n\n6. **Architectural notes**\n   - **ARCH-DRY: pass.** The holder logic exists in one place. `withoutEnv`, a test-only helper, has no existing equivalent in `cmd/`.\n   - **ARCH-PURE: pass.** `acquireSuiteSlot` is a thin IO layer by nature (it starts processes and handles the lock). No business logic was mixed into it.\n   - **ARCH-PURPOSE: pass.** I checked every caller of `runLocalCheck` (merge via cigate, `local-check run`, baseline diagnosis), and all of them now get the run-only slot. Nothing purpose-critical was deferred.\n   - One trade-off to know about, which the Spec accepts: a merge whose suite fails now gives up its place in line before the baseline-diagnosis rerun.\n\n7. **Plan revision recommendations:** none. The Plan's checklist and Log match the code.\n\n```findings\n```"

truncated: false
retained-bytes: 0000005222

---

## Re-review — 2026-10-14T23:22:45-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 383 — sdlc merge holds the machine suite slot until the process exits, through GitHub polling, the PR merge and archiving |
| repo | ariadne |
| issue file | workshop/issues/000383-sdlc-merge-holds-the-machine-suite-slot-until-the-process-exits-through-github-polling-the-pr-merge-and-archiving.md |
| boundary | whole-issue close |
| milestone | — |
| window | 551d94ed2823dcf2856fed196c50192bc9c90c23..d7e7e9272d35731033477332b25470bd92ef26e7 |
| command | sdlc close --issue 383 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-14T23:22:45-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

This review covers only the interdiff since the close review. It changes two lines of wording in `atlas/workflow/sdlc-binary.md`, rewrites one comment in `cmd/sdlc/localcheck_run.go`, adds the generated close-gate and close-review files under `workshop/plans/`, and edits the issue file's frontmatter and Log. No executable code changed. BR-2 is fixed at both places the prior round named. The new wording matches `scripts/test-lock.py:198-214`, where `wait_for_exit` returns at once through kqueue `NOTE_EXIT` and otherwise checks `kill(pid,0)` every 0.5 s. BR-1 was left unchanged on purpose, and the reason holds: `acquireSuiteSlot` still has exactly one caller (`cmd/sdlc/localcheck_run.go:81`), and the slot is released before `runLocalCheck` returns. So every later merge step runs without the slot. The issue-file edits only record the close: status set to codecomplete, `actual_hours`, and Log lines. `## Spec`, `## Revisions` and `## Done when` are unchanged, so the scope has not changed.

1. **Strengths**
   - `cmd/sdlc/localcheck_run.go:112-115`: the comment now says `release()` *starts* the release, and the watcher frees the slot when the holder exits. That is accurate on every platform.
   - `atlas/workflow/sdlc-binary.md:1742-1744`: the timing is stated for each watcher, kqueue and the 0.5 s poll. This was the fix the prior finding suggested.
   - I checked the rest of the class: `atlas/workflow/ci-merge-check.md:149-150` only says the holder "is released when the suite ends". It makes no claim about ordering against the status post, so nothing else needs rewording. A search for "before the status post" in `atlas/` and `cmd/` found nothing.

2. **Critical:** none.

3. **Important:** none.

4. **Minor:** none new. One trivial nit, not raised as a finding: the atlas text says "macOS's kqueue", but the helper uses kqueue wherever `select.kqueue` exists, which includes the BSDs. The Go comment's wording, "with kqueue", is more precise.

5. **Test coverage:** the interdiff changes only comments, prose and generated files, so no regression test is needed. The existing behavioural tests (`TestSuiteSlotCoversOnlyTheRun`, `TestSuiteRerunQueuesForTheSlot`) still cover the behaviour the docs describe.

6. **Architecture (interdiff only)**
   - **ARCH-DRY: pass.** Only wording changed; nothing is duplicated.
   - **ARCH-PURE: pass.** No code changed.
   - **ARCH-PURPOSE: pass.** The fix covers the whole doc-timing class: both sites named, plus the sweep of `ci-merge-check.md`.

7. **Plan revisions:** none.

```findings
dispose:
  - id: BR-1
    disposition: withdrawn
    note: |
      Verified by structure. acquireSuiteSlot's only caller is localcheck_run.go:81, and the slot is released before runLocalCheck returns, so the CI poll and PR merge always run without the slot. The prior round said the same. Not worth an end-to-end merge test.
  - id: BR-2
    disposition: addressed
    note: |
      Fixed at both sites: sdlc-binary.md:1742-1744 and localcheck_run.go:112-115 now state the release timing per watcher, matching test-lock.py:198-214. ci-merge-check.md:149 makes no ordering claim.
```
