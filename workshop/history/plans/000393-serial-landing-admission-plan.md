# Serial Landing Line Implementation Plan

> **For agentic workers:** See AGENTS.md Section 3 (Subagent Strategy) to choose an approach: superpowers-subagent-driven-development (if subagents suit the work per AGENTS.md) or superpowers-executing-plans. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** At most one `sdlc merge` per repository integrates and lands at a time. Each landing runs one integration suite. A logged, expiring, :0-granted fleet-blocking class goes ahead of normal waiters.

**Architecture:** `runDurableMerge` takes a ticket in a per-repo landing line and holds it until it returns. The line lives in `<git-common-dir>/sdlc/landing/`, beside #370's repo lock, and every slot of a repo shares that directory. A ticket is a file that the merge process holds an flock(2) on, so the process's death releases it and the next scan prunes it. A pure `NextHead` picks who goes next among the live tickets: fleet-blocking first, then normal, FIFO within each class. A head is never preempted. The landing's recovery is unchanged: `integrationstate.go` plus Git and GitHub state (the `Integrated-Main` trailer, merged-PR resume, convergent retry after a lost merge response). The line adds ordering only. This replaces #390's eight-state durable queue (TL approval, 2026-10-16; #404's lock-first rule).

**Tech Stack:** Go (`syscall.Flock`, darwin and linux), cobra, the existing `landingFakeGH` and procedure fixtures.

---

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `Ticket` | `cmd/sdlc/internal/landingline/order.go` | new |
| `Grant` | `cmd/sdlc/internal/landingline/order.go` | new |
| `Order` | `cmd/sdlc/internal/landingline/order.go` | new |
| `NextHead` | `cmd/sdlc/internal/landingline/order.go` | new |
| `ValidGrants` | `cmd/sdlc/internal/landingline/order.go` | new |
| `Authorize` | `cmd/sdlc/internal/landingline/order.go` | new |
| `Delays` | `cmd/sdlc/internal/landingline/log.go` | new |
| `KeepGrants`, `KeepEvents` | `cmd/sdlc/internal/landingline/retention.go` | new |

- **Ticket** — `{Seq int64, Issue int, Branch string, PID int, Enqueued time.Time}` (no PR: it is selected after admission, and the `admit` event records it). `Issue` is the normalised number from `issue.IDFromFilename(branch)` (cmd/sdlc/internal/issue/scaffold.go:41), the parse `tickOnLanding` already uses, so `0391` and `391` agree, a live waiter or the head. `Seq` comes from a per-line counter, so FIFO is a total order that clock skew can't break.
  - **Relationships:** N tickets per line, and at most one per process.
  - **DRY rationale:** this is the first Go line. scripts/test-lock.py keeps the suite's FIFO line, but the landing line needs sdlc's grants, and the waiter is the Go merge process itself (no helper process). The two lines share their design (flock liveness, pruning under a mutex), not code. The atlas notes this.
  - **Future extensions:** batching (#393 revision): `NextHead` widens to `NextBatch` and returns several tickets.
- **Grant** — `{Issue int, Grantor Grantor, Reason string, Granted, Expires time.Time, Revoked bool}`, where `Grantor` is `{Address string, Slot *int, Name string}`. One entry in `grants.jsonl`.
- **ValidGrants(entries, now) map[int]Grant** — the latest entry per issue counts if it is not revoked, not expired, has a reason, lasts at most 24 h, and records a grantor whose slot is nil or 0. Authority is checked when a grant is read as well as when it is written, so a hand-appended entry from a worker slot confers nothing. The log is the only source of grants, so an unlogged grant can't exist. **Trust model:** this is cooperative. It guards against agents' mistakes on one machine, not against an adversary. A process that forges the grantor fields can still grant, and the atlas says so.
- **Order(tickets, grants) []Ticket** — the admission order: fleet-blocking tickets first, then normal ones, FIFO by Seq within each class. Both the waiting position and `NextHead` derive from it.
- **NextHead(tickets []Ticket, grants map[int]Grant) (head Ticket, displaced []Ticket, ok bool)** — Order's first ticket. It takes no `now`, because `ValidGrants` already applied the clock. `displaced` lists the normal tickets with a lower `Seq` than a fleet-blocking head, for the audit event. A head that is already admitted is not passed in (see Line), so nothing can preempt it.
- **Authorize(callerSlot, treeSlot *int) error** (the write-side check; `ValidGrants` repeats the slot rule on read) — a grant is refused when the caller's workspace or `$COUCH_TREE`'s workspace resolves to a numbered slot (N ≥ 1). The operator's shell and :0 slots, including the TL's hub ops:0, may grant.
- **Delays(events) []float64** — enqueue-to-landed seconds over the `landed` events whose enqueue is still logged. `sdlc landing status` feeds them to test-stats' `suiteQuantiles`, so there is one nearest-rank implementation.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `Line` | `cmd/sdlc/internal/landingline/line.go` | new | flock(2) on files in `<common>/sdlc/landing/` |
| `flock.Exclusive` | `cmd/sdlc/internal/flock/flock_unix.go` | new (lifted) | `syscall.Flock`, moved from reviewbox's `lockExclusive` |
| `Log` | `cmd/sdlc/internal/landingline/log.go` | new | append-only `events.jsonl` and `grants.jsonl` |
| `runDurableMerge` | `cmd/sdlc/landing.go` | modified | confirms, takes a turn when an open PR exists, observes everything after admission |
| `landingTurn`, `enterLandingLine`, `openLandingLine`, `landingLineDir`, `lineLog`, `resumeOnly` | `cmd/sdlc/landingqueue.go` | new | the merge's glue to the line: production options, events, the turn's outcome, ending a landed grant |
| `sdlc landing` (`newLandingCmd`, `landingDeps`, `runLandingGrant`, `runLandingStatus`, `openRepoLine`) | `cmd/sdlc/landingcmd.go`, help `cmd/sdlc/helptext/landing.md` | new | `status`, `grant`, `revoke`; injected clock, caller and `$COUCH_TREE` |

- **Line** (as built) — `Open(dir, Options{Now, Sleep, Poll, Grants, OnWait, OnAdmit})`; production uses `time.Now`, a context-aware sleep, 2 s, and the grants log's `ValidGrants`.
  - `Enqueue(t) (*Handle, error)` takes the next number from the `seq` counter, which is untrusted input. It writes and flock-locks `tickets/<%020d seq>` (no extension) under a temporary name, then renames it into place, all under the allocation mutex (`line.lock`).
  - `Wait(ctx, h)` repeats `tryAdmit`. Under the mutex, that scans the live tickets (an unheld ticket is removed; a held one counts even when its JSON doesn't parse, and its number comes from its file name). If `head.lock` is free and `NextHead` names h, it takes `head.lock`, writes h's ticket into it, and deletes h's ticket. The head therefore holds only `head.lock`.
  - `Status() (head *Ticket, waiting []Ticket, err error)`.
  - `Release` drops whichever flock the handle holds. Process death releases every flock.
  - **Injected into (as built):** `runDurableMerge` through `enterLandingLine` (landingqueue.go), which opens the repository's real line in its git common dir. The test seams are `landingLinePoll` and `landingNow`, plus a head the test holds in the fixture repository's own line (`heldLine` in landingqueue_test.go). There is no `landingLineOpen` variable.
  - **ARCH-ORDER:** the line holds no state between events beyond the live flocks. Each process is in one of `waiting` (its ticket held) or `head` (`head.lock` held, ticket gone), and every exit ends in released. The race to watch is two waiters scanning at once. The allocation mutex serializes the scans, and `head.lock` is the only admission token. Process death mid-transition releases everything. A crash between taking `head.lock` and logging `admit` loses only an audit line.
- **Log** (as built) — methods on Line.
  - `Record(Event)` and `AddGrant(Grant)` append one JSON line with `O_APPEND` under the line mutex. Each first runs the daily sweep (a `.swept` stamp), which applies the pure `KeepEvents` and `KeepGrants` (`retention.go`) and counts the malformed lines it drops in the stamp.
  - `Events()` and `GrantEntries()` read without the mutex. A torn or malformed line is skipped and counted, never trusted, and a line of any length reads (ARCH-SECURE).
- **ARCH-FUNERAL:** a ticket is removed on Release, or by the next scan if its owner is dead. `head.lock` and `line.lock` are fixed files, two per repo, that are never deleted. Events and grants are swept by the pure retention rules (30 days, matching test-stats' retention; a sweep never changes which grants are valid). Size: about 200 B per event and 3 events per landing.
- **ARCH-CONSTRAINTS:** a scan takes O(waiters), a handful of files every 2 s per waiter. It is local IO only, and the wait is the queue's purpose. A waiter has no deadline of its own beyond its context; Ctrl-C releases it. The head's runtime is bounded by merge's existing gate and CI timeouts.
- **ARCH-MOCK:** Git and GitHub reuse the existing stateful `landingFakeGH` on real temp repos, and the live read-only `TestLandingGHLiveConformance`. The line itself is real files in a temp directory, so it needs no fake. Process death is tested with a helper process (the test binary re-executed).

## Milestones

- [ ] M1 — the landingline package: pure ordering, grants, stats, and Line and Log over real files, all tested.
- [ ] M2 — `runDurableMerge` holds the line; `sdlc landing status|grant|revoke`; integration tests on `landingFakeGH`; atlas.

### Task 1 (M1): pure ordering

**Files:** create `cmd/sdlc/internal/landingline/order.go` and `order_test.go`.

- [ ] Write the table tests first:
  - FIFO by Seq;
  - fleet-blocking overtakes older normal tickets and reports them as displaced;
  - FIFO among fleet-blocking tickets;
  - an expired, revoked, grantor-less or over-24 h grant confers nothing;
  - the latest entry per issue wins;
  - an empty line returns `ok=false`;
  - `Authorize`: nil/nil allowed, 0 allowed, a numbered caller refused, a numbered COUCH_TREE refused.
- [ ] Run `go test ./cmd/sdlc/internal/landingline/ -run 'TestNextHead|TestValidGrants|TestAuthorize'`. Expect FAIL (undefined).
- [ ] Implement:

```go
type Class int
const (Normal Class = iota; FleetBlocking)

// Superseded as built (see Revisions): NextHead(ts, grants) takes no now, and
// derives from Order.
func NextHead(ts []Ticket, grants map[int]Grant, now time.Time) (Ticket, []Ticket, bool) {
	if len(ts) == 0 { return Ticket{}, nil, false }
	sorted := slices.SortedFunc(slices.Values(ts), func(a, b Ticket) int { return cmp.Compare(a.Seq, b.Seq) })
	for _, t := range sorted {
		if _, ok := grants[t.Issue]; ok {
			var displaced []Ticket
			for _, o := range sorted { if o.Seq < t.Seq { displaced = append(displaced, o) } }
			return t, displaced, true
		}
	}
	return sorted[0], nil, true
}
```

  `ValidGrants` folds the entries in log order (the last one per issue wins), then filters with `!Revoked && Grantor.Address != "" && (Grantor.Slot == nil || *Grantor.Slot == 0) && Reason != "" && now.Before(Expires) && Expires.Sub(Granted) <= 24*time.Hour`. Pass it `now` from the caller. Never call `time.Now` in this file.
- [ ] Run the tests again. Expect PASS.
- [ ] Commit `#393 M1: landingline: pure next-head ordering and grant validity`.

### Task 2 (M1): Line over real files

**Files:**
- Create: `cmd/sdlc/internal/flock/flock_unix.go` and `flock_other.go`, lifted from `cmd/sdlc/internal/reviewbox/lock_unix.go` and `lock_other.go`.
- Modify: reviewbox, which now calls `flock.Exclusive`; delete its two lock files (ARCH-DRY).
- Create: `cmd/sdlc/internal/landingline/line.go` and `line_test.go`.

- [ ] Lift the helper first and run `go test ./cmd/sdlc/internal/reviewbox/` to confirm reviewbox is unchanged.

- [ ] Write the tests:
  - two handles in one process: the second `Wait` blocks until the first `Release`. Use a context deadline and an injected sleep that counts polls;
  - a dead waiter: a helper process (`TestHelperHoldTicket`, run with `os.Args[0] -test.run=TestHelperHoldTicket` and an env var naming the directory) enqueues and blocks. Kill it with `SIGKILL`; the next scan prunes its ticket and the survivor is admitted;
  - a dead head: the helper takes `head.lock` and is killed; the next waiter is admitted;
  - a grant reorders: three tickets, grant the third, release the head; the third is admitted next;
  - a hand-appended grant whose grantor slot is 3 confers nothing (the read-side check);
  - the head is not preempted: with the head admitted, a later grant doesn't affect `head.lock`, and `Wait` for the granted ticket returns only after `Release`.
- [ ] Run `go test ./cmd/sdlc/internal/landingline/ -run TestLine`. Expect FAIL.
- [ ] Implement `Open`, `Enqueue`, `Wait`, `Release` and `Status() (head *Ticket, waiting []Ticket, err error)` (as built) as described in Core concepts. Seq is a counter file `seq` read and incremented under `line.lock`. Ticket writes are atomic (`os.CreateTemp` in the directory, then `os.Rename`). A Ticket whose JSON doesn't parse but whose lock is held counts as live with Seq taken from its file name, so it is never silently skipped.
- [ ] Run the tests again. Expect PASS.
- [ ] Commit `#393 M1: landingline: flock-held tickets and head admission`.

### Task 3 (M1): Log, sweep, Delays

**Files:** create `cmd/sdlc/internal/landingline/log.go` and `log_test.go`.

- [ ] Write the tests:
  - append and read back;
  - a malformed line is skipped and counted;
  - Sweep drops entries older than 30 days and keeps the rest (superseded by round 2's pure retention rules and their invariant test; see Revisions);
  - Sweep runs at most once a day (the stamp);
  - `Delays` over a known event set: landed seconds only, and none for a swept enqueue. The p50 and p90 are tested in Task 5 through `suiteQuantiles`.
- [ ] Run, expect FAIL, implement, run, expect PASS.
- [ ] Commit `#393 M1: landingline: events and grants logs with 30-day sweep`.
- [ ] Run `sdlc milestone-close --issue 393 --milestone M1`.

### Task 4 (M2): merge holds the line

**Files:** modify `cmd/sdlc/landing.go` (`runDurableMerge`, after `gh` is obtained (around :353) and **before** the `head` read, `LandingPRs` and `selectIntegrationPR`); test in `cmd/sdlc/landingqueue_test.go` (as built).

**Order:** the interactive confirmation (`mergePrompter.Ask`, landing.go:372) moves before the enqueue, because its text depends on nothing it selects. An operator's prompt never holds the line's head (plan-quality minor).

**Rule (from plan-quality PQ-1, class: stale observation after a wait):** every remote or GitHub observation the landing acts on is taken *after* admission, never before the wait. That covers the branch head, the PRs, PR selection and integration resume. Before admission, only local branch selection and checkout readiness run (`landingIssueBranch` and `landingCheckoutReady`); both concern the waiting process's own slot, and `revalidateLanding` checks them again under the repo lock. One cheap pre-read decides whether to queue at all. As built after the M2 reviews, the run takes no ticket only when that read shows a merged PR and no open one (`resumeOnly`). The selected PR is then held to the rule that a merge without a turn acts only on an already-merged PR (see Revisions).

- [ ] Write the tests on `landingFixture`:
  - **holds and waits:** a test-held handle on the fixture's line (the git common dir of `roots[0]`) is the head. `runMerge` runs in a goroutine; assert that `gh.merges == 0` and no integration push happens while the line is held. Advance remote main (a commit pushed to the bare remote, standing in for the head's landing), then release. The merge integrates once against the new main, posts `sdlc-local/suite` exactly once (count it through `afterStatus`), and lands. `gh.merges == 1`;
  - **already-merged:** a PR with `State: "MERGED"` takes no ticket, so the line stays empty during the run;
  - **duplicate while waiting (PQ-1):** a second invocation for the same PR waits while the test, standing in for the original, lands it (`gh.pr.State = "MERGED"` plus a remote main advance), then releases. The admitted duplicate re-reads the PR, takes the MERGED path, and finishes cleanup without a second merge (`gh.merges` unchanged) and without the "push selected branch" refusal;
  - **pushed while waiting:** the branch gets a new commit on the remote and the PR head moves during the wait. After admission the landing selects the new head, integrates it once and lands it;
  - **refusal releases:** a gate refusal (`gh.checks` red) returns an error, and the line is empty afterwards;
  - **events:** enqueue, admit and landed are logged with PR, issue and Seq.
- [ ] Run `go test ./cmd/sdlc/ -run TestLandingLine`. Expect FAIL.
- [ ] Implement:

```go
// after gh is obtained, before head/LandingPRs/selectIntegrationPR:
if pre, err := gh.LandingPRs(ctx, t.Repo, branch); err != nil {
	return err
} else if !resumeOnly(pre) { // as built (M2 review I-A); first drafted as !anyMerged(pre)
	common, err := landingGit(r, t.Root, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil { return err }
	// as built: enterLandingLine/landingTurn in landingqueue.go hold this, and
	// the skip test is resumeOnly, and the selected PR is held to the no-turn rule (M2 reviews).
	line, err := landingLineOpen(filepath.Join(common, "sdlc", "landing"))
	if err != nil { return err }
	h, err := line.Enqueue(landingline.Ticket{Issue: issue.IDFromFilename(branch), Branch: branch, PID: os.Getpid()})
	if err != nil { return err }
	defer h.Release()
	if err := line.Wait(ctx, h); err != nil { return err }
}
// the existing head read, LandingPRs and selectIntegrationPR follow unchanged,
// so everything the landing acts on is observed after admission.
```

  The ticket records the branch and issue; the PR number goes into the `admit` event once selection has run. `Wait` reads grants through the Log on each poll. It prints a position line (`position N in line; head: PR #M`) only when the position changes, like test-lock's waiting lines.
- [ ] Run the tests again, then the existing `go test ./cmd/sdlc/ -run 'TestLanding'` to confirm nothing regressed.
- [ ] Commit `#393 M2: merge holds the repository landing line for its lifetime`.

### Task 5 (M2): `sdlc landing` commands

**Files:** create `cmd/sdlc/landingcmd.go` and `landingcmd_test.go`; modify `cmd/sdlc/main.go` (`add(NewLandingCmd(), "landing", ...)` next to `merge`).

- [ ] Write the tests:
  - `status` prints head, waiters, live grants, and n/p50/p90 of the delay;
  - `grant --repo DIR --issue N --reason R --hours 4` from a :0 fixture root appends a grant;
  - the same grant from a numbered-slot fixture root, or with `COUCH_TREE` pointing at one, refuses and writes nothing;
  - `--hours 25` refuses;
  - `revoke` appends a revoked entry, and the issue loses priority;
  - landing an issue appends a revoked entry for its grant (Task 4's landed path).
- [ ] Run, expect FAIL, implement, run, expect PASS. The grantor is the resolved workspace address (`workspace.Resolve`), or `operator shell` outside any workspace, plus `git config user.name`.
- [ ] Commit `#393 M2: sdlc landing status, grant and revoke`.

### Task 6 (M2): docs and the full suite

- [ ] Atlas: add the landing line to `atlas/workflow/sdlc-binary.md` (the merge section): its purpose, storage, ordering, grants and their cooperative trust model, retention, and how it relates to test-lock's suite line. Note that #370's repo lock beside it is mkdir-based while the line uses flock, so the atlas doesn't describe them as one mechanism. Mention it in `sdlc merge --help`'s description and RECOVERY ("a waiting merge holds only a ticket; Ctrl-C or death leaves the line"). Link from `atlas/index.md` if a new page is added.
- [ ] Run the full `make test` in the background.
- [ ] Run `sdlc close --issue 393`.
- [ ] The PR touches merge policy, so local-gates-guard applies. Stop there for the operator's `local-gates-verified` label (TL instruction) and report to the hub.

## Verification against Done when

1. One integration per repo: Task 4 (holds and waits, one suite), and Task 2 (one head). Different repos use different git common dirs, so their lines are separate by construction; Task 2 tests two line directories, and neither blocks the other. Exact-head evidence is unchanged (cigate).
2. Drift, failure, death, duplicates and uncertainty: death at each point is in Task 2 (helper SIGKILL as waiter and as head). A duplicate invocation queues behind the original, which is another process with its own ticket. After admission it re-reads the PR and resumes as MERGED (Task 4's duplicate-while-waiting test, PQ-1). A refusal releases the line (Task 4). Drift and uncertain outcomes keep the existing tests (TestLandingRefusesUncertainty, TestLandingInterruptedIntegration).
3. Priority: Tasks 1, 2 and 5 cover granted overtaking, FIFO, no preemption, expired, unauthorized and unlogged grants, and the displacement audit.
4. Couch down: the line never involves Couch. The git common dir is the only location (TL decision 2). Every test runs without Couch.
5. Fakes and conformance: `landingFakeGH` and the existing live check. **Pilot data:** `events.jsonl` (delay p90) and test-stats (suites per landed issue). The seven-day before/after report is the TL's, recorded in test-economy (see the issue's Revisions).

## Revisions

### 2026-10-16 — M1 as built, and the M1 review (REWORK) answered

**Reason:** building M1 simplified four entities, and the M1 boundary review found that the table no longer matched the code. It also found two retention and validity bugs in grants.

**Delta:**
- **`Order` added:** the pure admission order; `NextHead` and the waiting position derive from it.
- **`NextHead` drops `now`:** `ValidGrants` already applied the clock.
- **`Ticket` drops `PR`:** the PR is selected after admission (PQ-1), so the `admit` event records it.
- **`DelayStats` (stats.go) → `Delays` (log.go):** raw seconds that M2 feeds to test-stats' `suiteQuantiles`, keeping one nearest-rank implementation (ARCH-DRY).
- **The admitted head leaves the ticket set** and records itself in `head.lock`, so an admitted head is never passed to `NextHead`.
- **`OnAdmit` runs outside the line mutex:** flock is per open file, so a hook that logs under it would deadlock.

**Review fixes:**
- `ValidGrants` requires `Granted` to be no later than now, and the sweep drops future-dated entries, so a forged future grant confers nothing and doesn't live forever.
- `AddGrant` stamps a zero `Granted`, and a grant's retention keys on the later of `Granted` and `Expires`, so a revoke outlives the grant it revokes.
- A ticket's number comes from its file name, never its JSON.
- The sweep counts the malformed lines it drops in its stamp.
- `Wait` on a released or vanished ticket returns an error.
- The line tests step admission through `tryAdmit`, with `Wait`'s sleep injected, so each interleaving is the test's choice rather than the scheduler's (ARCH-ORDER).

### 2026-10-16 — M1 review round 2: retention is a pure rule with a tested invariant

**Reason:** round 1 aged a grant by the later of its grant and expiry times, and also dropped anything dated in the future. Every grant still in force has its expiry ahead of now, so a daily sweep deleted it. The retention decision was buried in the IO routine, where no pure test reached it.

**Delta:**
- Retention is two pure functions in `retention.go`. `sweep` only applies them.
  - `KeepEvents` keeps entries dated within the last 30 days and not in the future.
  - `KeepGrants` ages by `Granted` only. It keeps an issue's fresh entries while the issue's latest entry is fresh. Otherwise it drops the issue's whole history, so an older entry can't become the latest.
- **Invariant:** `ValidGrants(kept, now) == ValidGrants(all, now)`, and nothing outside retention is kept. `TestKeepGrantsPreservesValidity` checks it over 2000 random logs with out-of-order, old and future entries.
- A revoke outlives its grant because a revoke's `Granted` is the time of the revoke, which is later than the grant's.
- The ticket counter is untrusted input. A malformed `seq` refuses rather than resetting, and a counter set back can't overwrite a live ticket.

### 2026-10-16 — M2 as built, and the M2 review (FIX-THEN-SHIP) answered

**Reason:** M2's boundary review found a way to land outside the line (I-1). It also found that the Core concepts tables still described M2 as planned, not as built (I-2, the third `plan-table-drift` finding). The rule is now: every entity and seam the diff adds appears in the tables at its built path, and every as-built deviation gets a dated entry here in the same milestone.

**Delta:**
- **Skipping the line:** a merge skips the line only when its branch has no open PR, so all it can do is resume a merged PR's cleanup. Before, it skipped whenever any PR on the branch had merged, so a reused branch name with an old merged PR and a new open one landed outside the line. `TestLandingLineReusedBranchStillQueues` pins this, and a mutation check confirms it. What follows admission for a reused name is the landing's existing rule: cleanup refuses it as ambiguous, which is unchanged.
- **The merge's glue lives in `landingqueue.go`:** `enterLandingLine` opens the repository's line with production options and takes the turn before any observation. `landingTurn.leave(err)` logs `landed`, `refused` or `resumed`, ends a landed issue's valid grant, and releases.
- **`resumed`:** an admitted merge that finds its PR already landed by another invocation logs `resumed`, which `Delays` doesn't count, so one landing isn't counted twice.
- **Event and grant write failures** warn once on stderr. The landing goes on, unaudited, rather than silently losing pilot data.
- **One clock:** `landingNow` serves merge and the commands.
- **Commands:** `sdlc landing` is in `landingcmd.go`, with its help in `helptext/landing.md` (the `add(cmd, name)` convention) and merge's help in `helptext/merge.md` (step 0 and RECOVERY). Status prints grants sorted by issue.
- **Tests** are in `landingqueue_test.go` and `landingcmd_test.go`. The fake GitHub gained `older`, earlier PRs on the same branch name.

### 2026-10-16 — M2 review round 2: the no-turn rule is enforced on the selected PR

**Reason:** round 1's skip decision (no open PR) came from the first GitHub read, while the landing acts on a second read. A lagging listing could hide an open PR at the first read and show it at the second, so it landed outside the line. This was the second finding in `gate-skip-predicate-divergence`.

**Delta:**
- **Rule:** a merge without a turn acts only on an already-merged PR.
- **Skipping the line:** `resumeOnly` skips it only when the first listing shows a merged PR and no open one. An empty listing queues.
- **Enforcement:** after selection, a merge without a turn that selected a PR that isn't merged refuses, having changed nothing. A mutation check confirms the test fails without this.
- **Logging:** the merge's logging is a `lineLog` owned by the opened line (no process-global map). `leave` warns when it can't read grants to end a landed issue's priority.
- **Docs:** README and the atlas command table now cover `sdlc landing`.
