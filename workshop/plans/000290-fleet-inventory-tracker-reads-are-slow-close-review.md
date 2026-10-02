# Boundary Review — ariadne#290 (whole-issue close)

| field | value |
|-------|-------|
| issue | 290 — Fleet inventory tracker reads are slow |
| repo | ariadne |
| issue file | workshop/issues/000290-fleet-inventory-tracker-reads-are-slow.md |
| boundary | whole-issue close |
| milestone | — |
| window | f16d418421e0ab1d8e4d6f5a1890888793305a69..63ea1624334860720a75598a64047606465d88a8 |
| command | sdlc close --issue 290 |
| reviewer | claude |
| timestamp | 2026-10-02T16:10:00-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

**Verdict: fix then ship.** All three planned changes are in the code and work. Inventory skips the tracker fetch when nothing changed, tracker fetches pass `--no-auto-maintenance`, and reads run concurrently (at most 8) through a per-repository cache with a deadline on each read. I ran the five new tests with `-race`, plus `go vet` and `gofmt -l`; all are clean. The measured result, 17.5s down to 5.9s with byte-identical JSON, meets the "under 9s" target. Nothing is blocking. Four Important gaps should be closed before close:

- **Timeout behaviour differs from the plan, with no Revisions entry.** The plan says a timed-out read is not cached and falls back to stale. The code caches the timeout, and it reports unknown, not stale.
- **The plan's Core concepts table is wrong.** It names an entity that doesn't exist at the stated path.
- **Two Done-when tests are missing.** Nothing compares concurrent output against a sequential run, and nothing checks that a hanging remote leaves other repositories unaffected.
- **`atlas/workflow/issue-tracker.md` is stale.** It still says every snapshot fetches, which the fetch skip changes for every caller.

1. **Strengths**
   - `cmd/sdlc/internal/fleet/issues.go:93-131`: the cache loads each repository once, using a `done` channel per repository, so callers of one repository never wait on another's load. The load function is a parameter, so the cache tests need no IO. `TestRecordsCacheIsPerKeyOnce` proves loads overlap by holding them open on a gate, not by timing.
   - `cmd/sdlc/internal/gitx/snapshot.go:22-30`: the fetch skip is exact, not a cache. It only applies when the local tracking ref equals the tip `ls-remote` just reported, which is the same answer a fetch would give.
   - `parseLsRemoteTip` (`candidate.go:163`) only accepts a line for exactly the requested ref, and checks the ID with `parseObjectID` before it can become a git argument. That is correct handling of a remote response (ARCH-SECURE).
   - `warmRecords` waits for every goroutine it starts before returning, and the semaphore caps how many run at once (ARCH-ORDER, ARCH-CONSTRAINTS).
   - `TestSnapshotSkipsTheFetchWhenTheTipIsUnchanged` uses real git and counts fetches through `GIT_TRACE`. It covers three cases: unchanged tip, moved tip, and a probe's tip being used only once.

2. **Critical:** none.

3. **Important**
   - **Plan says the opposite of the code on timeouts.** The plan's Deadline decision (plan:52-56) says "A read that misses it is not cached … stale from the last fetch, else unknown."
     - The code keeps timed-out entries (`issues.go:116`).
     - A timeout always reports unknown. The local fallback in `Presence`/`LocalSnapshot` runs on the same `TrunkFile` context that has already expired.
     - The Log admits the second point, but the plan was never revised.
     - Either run the local fallback outside the deadline, so a repository with a fetched tracker reports stale, or add a `## Revisions` entry saying a timeout is cached and reports unknown.
   - **Core concepts table names the wrong entity.** It lists `warmRepoRecords` in `inventory.go`; the code has `warmRecords` in `issues.go`. It also calls `parseLsRemoteTip` "unit-tested", but it has no direct test; it is only exercised through the snapshot test. I rated this Important rather than Critical because both entities exist and only their names and locations are wrong.
   - **Done-when tests are missing.**
     - "Output identical to a sequential run": no test compares the warmed `CollectInventory` against a sequential one.
     - "Other repositories are unaffected": `TestHangingRemoteDegradesWithinTheDeadline` uses only one repository, although the plan's deadline task asked for a second one that reads `present`.
     - Both were checked by hand only, per the Log.
   - **Atlas is stale.** `atlas/workflow/issue-tracker.md:125,141` still describe one tracker fetch per read. The plan listed this file for an update, but only `sdlc-binary.md` was changed. The fetch skip lives in `gitx`, so it affects every `Snapshot` caller, including verbs like `claim` and `close`, not only inventory.

4. **Minor**
   - `probedTip` is only cleared by the next `Snapshot`. A `fetch`/`refreshTip` in between (for example `UpdateMany`) leaves it in place, so the "right after a probe" rule in the comment isn't enforced (ARCH-ORDER). Clearing it inside `fetch()` would close this.
   - The fetch skip also changes write verbs that call `Initialized` and then `Snapshot`. This looks semantically safe, because the push still checks the base and fetches, but the plan says "write paths untouched".
   - `snapshot_test.go:5`: the `testfix` import is mixed into the standard-library import group.
   - If a repository's spelling differs between the warm-up and the row walk, that repository silently gets a second, sequential read. This is documented, but no test catches the key mismatch.

5. **Test coverage**
   - Covered: cache loads once per repository and runs repositories in parallel; timeouts are kept and other errors retried; the concurrency limit holds; the deadline turns a hang into unknown.
   - Not covered: output identical to a sequential run, isolation from a hanging neighbour, a timeout falling back to stale, and `parseLsRemoteTip` on bad or unrelated lines.

6. **Architecture, per principle**
   - ARCH-DRY: pass.
   - ARCH-PURE: pass. The cache takes its loader as a parameter, and `parseLsRemoteTip` is pure.
   - ARCH-PURPOSE: pass.
   - ARCH-MOCK: pass. Tests use real git plus an ssh transport that hangs.
   - ARCH-CONSTRAINTS: pass. There is a limit of 8, a deadline of 15s, and a measurement.
   - ARCH-SECURE: pass.
   - ARCH-ORDER: pass, apart from the `probedTip` Minor above.
   - ARCH-FUNERAL: pass. Everything lives in memory for the process only.
   - For #289: the local per-worktree git work (about 6s) is the next thing to speed up. It can reuse `warmRecords`' semaphore pattern.

7. **Plan revisions** (`workshop/plans/000290-fleet-inventory-tracker-reads-are-slow-plan.md`)
   - Deadline: a timed-out read is cached and reports unknown, because the local fallback shares the expired context. Drop that entry if the stale fallback is implemented instead.
   - Core concepts: rename to `warmRecords` in `issues.go`, and note that `parseLsRemoteTip` is tested indirectly.
   - Fetch skip: it applies to every `TrunkFile.Snapshot` caller, not only inventory.

```findings
findings:
  - id: new
    severity: Important
    family: plan-code-drift
    title: |
      Deadline decision says timeouts are not cached and fall back to stale; code caches them and reports unknown
    detail: |
      issues.go:116 keeps DeadlineExceeded entries, and the local fallback in Presence/LocalSnapshot shares the expired TrunkFile context, so a tracked repository with a fetched tracker reports unknown instead of stale. Either run the local fallback outside the deadline, or add a plan Revisions entry (the Log notes the deviation; the plan does not).
  - id: new
    severity: Important
    family: plan-code-drift
    title: |
      Core concepts table names warmRepoRecords in inventory.go; the code has warmRecords in issues.go
    detail: |
      The table also says parseLsRemoteTip is unit-tested, but there is no direct test; it is only exercised through TestSnapshotSkipsTheFetchWhenTheTipIsUnchanged. Revise the table or add the unit test.
  - id: new
    severity: Important
    family: done-when-untested
    title: |
      Done-when items untested: concurrent output equals a sequential run; a hanging remote leaves other repositories unaffected
    detail: |
      TestHangingRemoteDegradesWithinTheDeadline uses a single repository, although the plan asked for a second one that reads present. No test compares a warmed CollectInventory against a sequential run; both checks were only done by hand per the Log.
  - id: new
    severity: Important
    family: atlas-stale-after-contract-change
    title: |
      atlas/workflow/issue-tracker.md still says every snapshot fetches; the fetch skip now applies to every Snapshot caller
    detail: |
      Lines 125 and 141 describe one tracker fetch per read. The plan listed this file for an update, but only sdlc-binary.md changed.
  - id: new
    severity: Minor
    family: implicit-cross-call-state
    title: |
      probedTip survives an intervening fetch or refreshTip, so the "right after a probe" rule is not enforced
    detail: |
      Clear probedTip inside fetch(). The skip also reaches write verbs (Initialized then Snapshot), contrary to the plan's "write paths untouched"; this looks safe because CAS still fetches, but the plan should say so.
  - id: new
    severity: Minor
    family: import-grouping
    title: |
      snapshot_test.go puts the testfix import inside the standard-library import group
```

---

## Re-review — 2026-10-02T16:19:08-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 290 — Fleet inventory tracker reads are slow |
| repo | ariadne |
| issue file | workshop/issues/000290-fleet-inventory-tracker-reads-are-slow.md |
| boundary | whole-issue close |
| milestone | — |
| window | f16d418421e0ab1d8e4d6f5a1890888793305a69..bc926309a4a9b77976be560609ca7448a1de167b |
| command | sdlc close --issue 290 |
| reviewer | claude |
| timestamp | 2026-10-02T16:19:08-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

All new tests pass under `-race` (`TestRecordsCache*`, `TestWarmRecordsIsBounded`, `TestHangingRemoteDegradesOnlyItsRepository`, `TestWarmedInventoryEqualsSequential`, `TestSnapshot*`). Round-1 findings BR-1, BR-3, BR-4 and BR-6 are fixed. BR-1 was closed by revising the plan to match the code, which the finding allowed as one of its two options. BR-3 now has real tests: two tracked repositories with bare origins, and a warmed run compared against a sequential one as JSON. Two items are still partly open, and neither is a correctness bug:
- **BR-2:** the pure function `parseLsRemoteTip` is still described as "unit-tested" in the plan, but it has no direct test.
- **BR-5:** the new line that clears `probedTip` inside `fetch()` has no test that would fail without it.

Both are cheap to fix.

1. **Strengths**
   - The per-key-once `recordsCache` (`cmd/sdlc/internal/fleet/issues.go:93-133`) is small and correct. Each caller waits only on its own repository's load. A timed-out load stays cached, so the row walk never waits twice. Other errors are dropped from the cache so the next caller retries.
   - The concurrency tests prove overlap and the bound through gates and channels, not timing (`recordscache_test.go:16-59`, `:81-122`), so they don't depend on scheduling luck.
   - The fetch-skip is exact rather than a cache. It reads the probed tip only when the local tracking ref already points there, and uses it only once (`gitx/snapshot.go:22-30`). Writes still do a compare-and-swap against a fetched tip, through `refreshTip` → `fetch`.
   - `TestHangingRemoteDegradesOnlyItsRepository` uses real git with a hanging ssh transport. It checks the time bound, the healthy repository reading `present`, and the stuck repository reading `unknown` with a reason that names the deadline.
   - The help text, `atlas/workflow/sdlc-binary.md` and `atlas/workflow/issue-tracker.md` all now describe the probe-then-fetch-if-moved read.

2. **Critical:** none.

3. **Important:** none new. BR-2 is still open; see its disposition below.

4. **Minor:** none new. BR-5 is still open; see its disposition below.

5. **Test coverage notes**
   - The fetch-skip test covers four cases: unchanged tip, moved tip, a probed tip used only once, and `--no-auto-maintenance`. It does not cover the sequence probe → `refreshTip` → `Snapshot`, which is the path the BR-5 change targets.
   - `TestWarmedInventoryEqualsSequential` also asserts that its fixture actually reached tracker reads (`claims_state":"present"`). That guards against the comparison passing because both runs did nothing.

6. **Architecture**
   - **ARCH-DRY:** pass. `loadRepoRecords` is shared by production and the tests' `freshRecords`.
   - **ARCH-PURE:** pass, with the BR-2 caveat. `parseLsRemoteTip` is pure but tested only through git integration.
   - **ARCH-PURPOSE:** pass. All four Done-when items are delivered and measured (17.5s → 5.9s).
   - **ARCH-MOCK:** pass. Tests use real git with bare origins and an ssh hang stub.
   - **ARCH-CONSTRAINTS:** pass. At most 8 reads in flight, each with a 15s deadline. The warm-up is lexically bounded by `wg.Wait`.
   - **ARCH-SECURE:** pass. `ls-remote` output is validated with `parseObjectID`, and malformed output falls back to a fetch.
   - **ARCH-ORDER:** pass with a note. `probedTip` is implicit state carried between calls on `TrunkFile`. It is cleared in all three places that touch it (`RemoteExists`, `Snapshot`, `fetch`), which is acceptable for a single string. If this hand-off ever grows a second field, make it an explicit token returned by `RemoteExists` instead of hidden state on the struct.
   - **ARCH-FUNERAL:** pass. Nothing durable is created; the cache lives for the process.

7. **Plan revision recommendations**
   - Either add a direct test for `parseLsRemoteTip` or change the Core concepts text "pure, unit-tested" so it no longer claims one.
   - If the fetch-clear stays without a regression test, add a Revisions note saying it is a defensive tightening with no observable behavior change.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Plan Revisions now states that timeouts are cached and report unknown; this was one of the finding's two accepted options.
  - id: BR-2
    disposition: not-addressed
    note: |
      The name is fixed in Revisions, but the plan still calls parseLsRemoteTip "pure, unit-tested" and no direct test exists. A table test (wrong ref, malformed OID, multi-line output) is cheap.
  - id: BR-3
    disposition: addressed
    note: |
      TestWarmedInventoryEqualsSequential and TestHangingRemoteDegradesOnlyItsRepository (healthy repo reads present) pass under -race.
  - id: BR-4
    disposition: addressed
    note: |
      issue-tracker.md lines 125, 142 and 260 now describe ls-remote plus a fetch only when the tracker moved.
  - id: BR-5
    disposition: not-addressed
    note: |
      The plan wording is fixed and fetch() clears probedTip (trunkfile.go:201), but no test runs the RemoteExists, refreshTip, Snapshot sequence, so reverting the clear stays green.
  - id: BR-6
    disposition: addressed
    note: |
      The testfix import is now in its own group (snapshot_test.go:10).
```
