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
| `NextHead` | `cmd/sdlc/internal/landingline/order.go` | new |
| `ValidGrants` | `cmd/sdlc/internal/landingline/order.go` | new |
| `Authorize` | `cmd/sdlc/internal/landingline/order.go` | new |
| `DelayStats` | `cmd/sdlc/internal/landingline/stats.go` | new |

- **Ticket** — `{Seq int64, Issue int, PR int, Branch string, PID int, Enqueued time.Time}`. `Issue` is the normalised number from `issue.IDFromFilename(branch)` (cmd/sdlc/internal/issue/scaffold.go:41), the parse `tickOnLanding` already uses, so `0391` and `391` agree, a live waiter or the head. `Seq` comes from a per-line counter, so FIFO is a total order that clock skew can't break.
  - **Relationships:** N tickets per line, and at most one per process.
  - **DRY rationale:** this is the first Go line. scripts/test-lock.py keeps the suite's FIFO line, but the landing line needs sdlc's grants, and the waiter is the Go merge process itself (no helper process). The two lines share their design (flock liveness, pruning under a mutex), not code. The atlas notes this.
  - **Future extensions:** batching (#393 revision): `NextHead` widens to `NextBatch` and returns several tickets.
- **Grant** — `{Issue int, Grantor Grantor, Reason string, Granted, Expires time.Time, Revoked bool}`, where `Grantor` is `{Address string, Slot *int, Name string}`. One entry in `grants.jsonl`.
- **ValidGrants(entries, now) map[int]Grant** — the latest entry per issue counts if it is not revoked, not expired, has a reason, lasts at most 24 h, and records a grantor whose slot is nil or 0. Authority is checked when a grant is read as well as when it is written, so a hand-appended entry from a worker slot confers nothing. The log is the only source of grants, so an unlogged grant can't exist. **Trust model:** this is cooperative. It guards against agents' mistakes on one machine, not against an adversary. A process that forges the grantor fields can still grant, and the atlas says so.
- **NextHead(tickets, grants, now) (head Ticket, displaced []Ticket, ok bool)** — the oldest live fleet-blocking ticket, otherwise the oldest normal one. `displaced` lists the normal tickets with a lower `Seq` than a fleet-blocking head, for the audit event. A head that is already admitted is not passed in (see Line), so nothing can preempt it.
- **Authorize(callerSlot, treeSlot *int) error** (the write-side check; `ValidGrants` repeats the slot rule on read) — a grant is refused when the caller's workspace or `$COUCH_TREE`'s workspace resolves to a numbered slot (N ≥ 1). The operator's shell and :0 slots, including the TL's hub ops:0, may grant.
- **DelayStats(events) (n, p50, p90)** — enqueue-to-landed delay over the `landed` events, using nearest rank like test-stats.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `Line` | `cmd/sdlc/internal/landingline/line.go` | new | flock(2) on files in `<common>/sdlc/landing/` |
| `flock.Exclusive` | `cmd/sdlc/internal/flock/flock_unix.go` | new (lifted) | `syscall.Flock`, moved from reviewbox's `lockExclusive` |
| `Log` | `cmd/sdlc/internal/landingline/log.go` | new | append-only `events.jsonl` and `grants.jsonl` |
| `runDurableMerge` | `cmd/sdlc/landing.go` | modified | takes and holds the line |
| `sdlc landing` | `cmd/sdlc/landingcmd.go` | new | `status`, `grant`, `revoke` |

- **Line** — `Open(dir)`. `Enqueue(t) (*Handle, error)` creates `tickets/<seq>.json` under the allocation mutex (`line.lock`, flock), writes it atomically (temp file plus rename), and keeps the ticket's flock until `Release`. `Wait(ctx, h, grantsFn, poll)` loops: under the mutex it scans the live tickets (a ticket nobody holds is removed), drops the head file's owner if dead, and if `NextHead` names h and nobody holds `head.lock`, it takes `head.lock` and returns. It takes an injected clock and sleep, and polls every 2 s in production. `Release` drops `head.lock`, removes the ticket and closes it. Process death releases every flock.
  - **Injected into:** `runDurableMerge` through a package variable `landingLineOpen`, so tests can use a temp directory or a held line.
  - **ARCH-ORDER:** the line holds no state between events beyond the live flocks. Each process is in one of `waiting` (ticket held) or `head` (ticket and `head.lock` held), and every exit ends in released. The race to watch is two waiters scanning at once. The allocation mutex serializes the scans, and `head.lock` is the only admission token. Process death mid-transition releases everything. A crash between taking `head.lock` and logging `admit` loses only an audit line.
- **Log** — `Append(file, v)` uses `O_APPEND` with one JSON line per write, which is atomic under PIPE_BUF for our record sizes. `Sweep(file, now)` rewrites the file without entries older than 30 days, under the mutex, at most once a day (it checks a `.swept` stamp). Malformed lines are skipped and counted, never trusted (ARCH-SECURE).
- **ARCH-FUNERAL:** a ticket is removed on Release, or by the next scan if its owner is dead. `head.lock` and `line.lock` are fixed files, two per repo, that are never deleted. Events and grants are swept after 30 days, which matches test-stats' retention. Size: about 200 B per event and 3 events per landing.
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

func NextHead(ts []Ticket, grants map[string]Grant, now time.Time) (Ticket, []Ticket, bool) {
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

  `ValidGrants` folds the entries in log order (the last one per issue wins), then filters with `!Revoked && Grantor != "" && Reason != "" && now.Before(Expires) && Expires.Sub(Granted) <= 24*time.Hour`. Pass it `now` from the caller. Never call `time.Now` in this file.
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
- [ ] Implement `Open`, `Enqueue`, `Wait`, `Release` and `Status() ([]Ticket, headPID int)` as described in Core concepts. Seq is a counter file `seq` read and incremented under `line.lock`. Ticket writes are atomic (`os.CreateTemp` in the directory, then `os.Rename`). A Ticket whose JSON doesn't parse but whose lock is held counts as live with Seq taken from its file name, so it is never silently skipped.
- [ ] Run the tests again. Expect PASS.
- [ ] Commit `#393 M1: landingline: flock-held tickets and head admission`.

### Task 3 (M1): Log, sweep, DelayStats

**Files:** create `cmd/sdlc/internal/landingline/log.go`, `stats.go` and `log_test.go`.

- [ ] Write the tests:
  - append and read back;
  - a malformed line is skipped and counted;
  - Sweep drops entries older than 30 days and keeps the rest;
  - Sweep runs at most once a day (the stamp);
  - `DelayStats` p50 and p90 over a known set match test-stats' nearest rank.
- [ ] Run, expect FAIL, implement, run, expect PASS.
- [ ] Commit `#393 M1: landingline: events and grants logs with 30-day sweep`.
- [ ] Run `sdlc milestone-close --issue 393 --milestone M1`.

### Task 4 (M2): merge holds the line

**Files:** modify `cmd/sdlc/landing.go` (`runDurableMerge`, after `gh` is obtained (around :353) and **before** the `head` read, `LandingPRs` and `selectIntegrationPR`); test in `cmd/sdlc/landing_line_test.go`.

**Rule (from plan-quality PQ-1, class: stale observation after a wait):** every remote or GitHub observation the landing acts on is taken *after* admission, never before the wait. That covers the branch head, the PRs, PR selection and integration resume. Before admission, only local branch selection and checkout readiness run (`landingIssueBranch` and `landingCheckoutReady`); both concern the waiting process's own slot, and `revalidateLanding` checks them again under the repo lock. One cheap pre-read decides whether to queue at all: if `LandingPRs` already shows the branch's PR as MERGED, the run takes no ticket and resumes cleanup at once.

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
} else if !anyMerged(pre) {
	common, err := landingGit(r, t.Root, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil { return err }
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
