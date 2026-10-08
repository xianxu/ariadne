# Boundary pushes and sdlc abandon — Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Started work is always on origin (the issue branch is pushed at every sdlc boundary and deleted remotely at merge), and ending active work as `wontfix`/`punt` goes through one verb, `sdlc abandon`, that keeps the work under an archive ref which a reopen restores.

**Architecture:** One lease push (`pushIssueBranch`, `handoff.go:205`, already used by unclaim) becomes the boundary push, called at the end of start-plan, milestone-close, close (and close's reconcile completion), and by `sdlc pr`'s durable push. Merge deletes the remote branch with a lease on the landed head. `abandon` is a convergent sequence (note commit → archive ref → card CAS → narrow main archive → branch deletion) whose card record (`abandoned{ref, head}`) lets a rerun resume and a reopen restore.

**Tech Stack:** Go (`cmd/sdlc`), git (`push --force-with-lease`, custom refs), tracker envelope records (`internal/issue`), `env.main.UpdateManyPrepared` + `gitx.TrunkWrite` for the narrow main commit; fixtures `trackerRepo`, `startedHere`, `closeReady`, `procedureFixture` + `landingFakeGH`.

---

## Non-goals

- Legacy repositories (no issue tracker): no boundary pushes (start-plan returns before its tracker steps there, `pushIssueBranch` needs a `trackerEnv`) and no `abandon`. Legacy `merge` keeps `gh pr merge --delete-branch` (`ghclient.go:149`).
- Un-archiving a reopened `done` issue: the same gap D9 closes for abandoned issues, filed as a follow-up.
- Abandon by a non-owner: claim first (`claim` takes over unowned work; `reclaim` reassigns owned work). A terminal card is not claimable, so a non-owner can't reopen another workspace's abandoned issue either; that is today's reopen rule, unchanged.
- Pruning old archive refs: they are removed only by a reopen (D10).

## Operating envelope

One extra `git push` per boundary verb (start-plan, milestone-close, close), typically 1–2 s on a warm connection. Offline, the boundary verbs still complete and warn (D2). `abandon` and the reopen restore need the remote and refuse before any effect when the fetch or first push fails.

## Design decisions

- **D1 — The boundary push is `pushIssueBranch`, unchanged.** Its lease is the last-fetched remote-tracking ref (absent → the branch must not exist remotely), never a fresh read, so a rebase followed by the next boundary force-pushes safely and a peer's unseen push is refused. One wrapper, `boundaryPush(env, stderr, verb)`, pushes the current branch when it is an issue branch (not the resting branch, not main, not detached) and reports.
- **D2 — A failed boundary push warns; it never undoes or fails the boundary.** The verb's own effect (card write, close binding) has landed; the push is durability, and the next boundary or `sdlc pr` pushes again (lesson: a projection must not add a refusal to its host verb). A stale lease is the exception worth shouting about: the warning names the remote tip and says to fetch and inspect before the next boundary. Unclaim keeps its existing hard failure (a handoff without the push is no handoff).
- **D3 — milestone-close pushes HEAD as it is.** It makes no commit (the agent commits the printed trailers), so its push carries the milestone's work and the trailer commit follows at the next boundary. The Done-when ("origin equals local HEAD after the verb") holds at the moment the verb returns.
- **D4 — `sdlc pr` (durable path) uses the same lease push** instead of a plain `push -u`, so a rebase between boundaries doesn't make `pr` fail non-fast-forward. Its upstream config write stays as is. The legacy (untracked) pr path is untouched.
- **D5 — merge deletes the remote branch with a lease on the landed head** (`push --force-with-lease=refs/heads/B:<head> <remote> :refs/heads/B`), after the local deletion in `deleteLandingBranch` (`landing.go:310`). Best-effort with a warning: the landing is done. A rerun (resume path) whose remote branch is already gone is a no-op.
- **D6 — `abandon` accepts any non-terminal status the owner holds** (`open`, `working`, `blocked`, `codecomplete`); from `open` there is no branch, so the branch steps are skipped. `--as wontfix|punt` maps to the model's `abandon`/`defer` events (the edge is checked against `construct/vocabulary/issue.cue`, not hardcoded).
- **D7 — abandon's sequence, each step detectable on rerun.** First, rerun detection: a card already terminal (`wontfix`/`punt`) whose `abandoned` record names this issue's ref means "resume at step 5". The step-1 checks below don't apply to it (the checkout may already be on its resting branch), only ownership by attribution and a clean tree do. Otherwise:
  1. Refuse unless: owner (`requireCardOwnership`), a clean tree, and, when the issue has started work, the checkout is on a non-resting, non-main branch (the one being abandoned; the name is not checked, per #285).
  2. Append `- YYYY-MM-DD: abandoned (wontfix|punt) — <reason>` under `## Log` (same-day subheading rule as unclaim's note) and commit `#N: log: abandon (<as>)`. Skipped when that commit is already the tip.
  3. Push the tip to `refs/ariadne/abandoned/NNNNNN` on the publication remote. Lease: the ref must not exist, or already equal the tip (rerun), or hold an ancestor of the tip (an orphan left by an interrupted reopen, D9, whose content the branch contains); anything else refuses.
  4. One card CAS: status → `wontfix`/`punt`, record `abandoned: {ref, head}`, claimant kept as attribution, `updated` today. A card already terminal with the same record is the rerun case.
  5. One narrow main commit `#N: issue: abandon (<as>) — archive details`, through the landing archive's own rules rather than a parallel archiver: destinations from `archiveDestination(historyDir, kind, base)` (`archivepolicy.go:13`), the details projected by `archivedDetails` (`landingarchive.go:66`) generalized from "done" to the model's terminal statuses (the landing archive only ever passes done cards, so its behavior is unchanged), and every plan artifact on main that `planArtifactBelongsToIssue` assigns to the issue moved to the plans archive too. One `TrunkWrite{Write, Delete}`; ownership re-checked in `beforePush`. Skipped when main has no live copy left.
  6. Delete the remote issue branch (lease on the tip), switch to the resting branch, delete the local branch. Each skipped when already gone; the rerun detection above is what lets a rerun from the resting branch get here.
- **D8 — `set-status` to `wontfix`/`punt` from `working`/`blocked`/`codecomplete` refuses toward `sdlc abandon --issue N --as … --reason …`.** From `open` (triage) it keeps working as today, and `--force` does not bypass the redirect (there would be no archive ref, so the work would be lost; this is ARCH "a gate the agent can skip is not a gate").
- **D9 — Reopen restores and un-archives (operator decision 2026-10-08).** The seam: `set-status`'s reopen gains effects only when the card carries an `abandoned` record; the effects live in `abandon.go` (`restoreAbandoned`), which set-status calls before its card write. Who may reopen is today's rule (the attributed owner, or an unattributed card), from a resting branch or the restored branch. Steps, each detected on rerun:
  1. Fetch the archive ref; its tip must equal the recorded head. A rerun whose branch already exists skips the fetch.
  2. Create the issue branch (`issue.BranchName`) at the head and switch to it. Skipped when the branch exists and contains the head.
  3. Merge main. A merge in progress (`MERGE_HEAD`) stops with "finish the merge, then rerun". Skipped when the branch already contains main's archive commit for the details. A code conflict stops here, before the card write, with the merge left for the agent.
  4. Move the archived copy (`history/issues/X`, which holds the abandon note) back to `workshop/issues/X`, and archived plan artifacts back to `workshop/plans/`; commit `#N: reopen: restore from refs/ariadne/abandoned/NNNNNN`. Skipped when the branch already has `workshop/issues/X` and no `history/issues/X`.
  5. Push the branch (D1).
  6. Card CAS: status → `working`, clear the record (the ordinary reopen write, plus the clear).
  7. Delete the archive ref (lease on the head). A crash between 6 and 7 leaves an orphan ref whose content the branch contains: harmless, and a later abandon of the same issue overwrites it (D7 step 3).
  The reopened branch then lands cleanly through the #285 guard (owner, and based on main's archive commit). Reopening a `done` issue keeps today's behavior (non-goal).
- **D10 — Lifecycle of the new things (ARCH-FUNERAL).** `refs/ariadne/abandoned/NNNNNN`: created by abandon, last needed by a reopen, removed by that reopen (or overwritten by a later abandon if a reopen's last step was interrupted); otherwise it lives as long as the repository, at one ref plus the branch's unique objects per abandoned issue, bounded by the number of abandoned issues. The card's `abandoned` record: cleared by the reopen, kept as attribution on a terminal card otherwise (one small record per card). Remote issue branches: now removed by merge and abandon, where before they accumulated.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `Abandoned` record (`CardAbandoned`, `SetCardAbandoned`, `ClearCardAbandoned`) | `cmd/sdlc/internal/issue/abandoned.go` | new |
| `abandonDecision` | `cmd/sdlc/abandon.go` | new |
| `abandonEvent` | `cmd/sdlc/abandon.go` | new |
| `trackerEnvelope` | `cmd/sdlc/internal/issue/handoff.go` | modified (gains `Abandoned`) |

- **Abandoned** — `{Ref, Head}` on the card's tracker envelope, validated like `Release` (ref matches `refs/ariadne/abandoned/\d{6}`, head is an OID). Unit tests in `abandoned_test.go`, colocated, no IO.
- **abandonDecision(card, as, today, ref, head)** — the card bytes after abandon: status via the model's edge (refuses a status with no `abandon`/`defer` edge), the record, `updated`; the rerun case (already terminal with the same record) returns `tracker.ErrNoChange`. Table test over the model's statuses.
- **abandonEvent(as)** — `wontfix` → `abandon`, `punt` → `defer`, via the vocabulary.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `boundaryPush` | `cmd/sdlc/boundarypush.go` | new | `pushIssueBranch` |
| `pushIssueBranch` | `cmd/sdlc/handoff.go` | modified (moved to `boundarypush.go`) | git push with lease |
| `deleteRemoteBranch` | `cmd/sdlc/boundarypush.go` | new | git push delete with lease |
| `runAbandon` / `newAbandonCmd` | `cmd/sdlc/abandon.go` | new | git, tracker CAS, main publish |
| `restoreAbandoned` | `cmd/sdlc/abandon.go` | new | git fetch/branch/merge/mv/commit/push |
| `startPlanBranch`, `publishTrackerClose`, milestone-close finalize, close reconcile completion | `startplan.go`, `closetracker.go`, `close.go`/`milestoneclose.go`, recovery reconcile | modified | + `boundaryPush` |
| `runDurablePR` | `landing.go:558` | modified | lease push |
| `deleteLandingBranch` | `landing.go:310` | modified | + `deleteRemoteBranch` |
| `checkTransitionGuards` / reopen path | `setstatus.go` | modified | redirect; restore on reopen |

All IO is exercised against real bare remotes in the existing fixtures (no mocks).

## M1 — Boundary pushes and merge cleanup

- [ ] M1 — boundary pushes at start-plan, milestone-close, close; lease push in `pr`; merge deletes the remote branch

### Task 1: `boundaryPush` and `deleteRemoteBranch`
**Files:** create `cmd/sdlc/boundarypush.go`, `boundarypush_test.go`; move `pushIssueBranch` there from `handoff.go`.
- [ ] Tests (fixture `startedHere`, real bare origin) cover the lease cases: first push, rewrite then force-with-lease, an unseen remote tip refused (origin untouched), resting branch skipped, remote delete at/off its expected head, and a repeat push after a lost response.
- [ ] Implement; `boundaryPush` returns nothing and writes `cok`/`cwarn` (D2).
- [ ] Commit `#286 M1: boundary push helper`.

### Task 2: the boundary verbs push
**Files:** `startplan.go` (~:344, also the rerun path that "re-checks and does nothing"), `closetracker.go` (`publishTrackerClose` end, ~:425), milestone-close finalize (`close.go` ~:1410 and `milestoneclose.go` ~:183, opening a tracker env), the close completion in `issue recovery reconcile`; recovery contracts in `internal/recovery/catalog.go` (start-plan, milestone-close, close, reconcile, pr).
- [ ] Test `TestBoundaryVerbsPushTheIssueBranch`: in `closeReady`'s flow assert `ls-remote` equals HEAD after start-plan, after a milestone-close, after close; then rebase onto an advanced main, reopen and close again (the #301 path: `set-status working`, then `close`, as `TestTrackerCloseSurvivesARebase` drives it) and assert origin equals the rewritten HEAD (lease force). Test that start-plan's rerun pushes when origin lacks the branch.
- [ ] Insert `boundaryPush`; update the contracts' Effects ("pushes the issue branch with a lease; a failed push warns").
- [ ] Commit `#286 M1: push the issue branch at every boundary`.

### Task 3: `pr` lease push; merge deletes the remote branch
**Files:** `landing.go` (`runDurablePR` ~:558, `deleteLandingBranch` ~:310), `tracker_e2e_test.go` or a new `landingcleanup_test.go`.
- [ ] Tests: after a rebase, `sdlc pr` publishes the rewritten branch; `TestTrackerFullSlotCycle`'s landing leaves `ls-remote --heads <branch>` empty; the merge resume path with the remote branch already deleted succeeds.
- [ ] Implement; contracts for pr/merge updated; atlas (`atlas/workflow/issue-tracker.md`, the verb table rows for start-plan, close, pr, merge).
- [ ] `make test`; `sdlc milestone-close --issue 286 --milestone M1`.

## M2 — `sdlc abandon` and reopen

- [ ] M2 — abandon verb, set-status redirect, reopen restores and un-archives

### Task 4a: archive policy for terminal statuses
**Files:** `landingarchive.go` (`archivedDetails`), `archivepolicy.go`, their tests.
- [ ] `archivedDetails` accepts the model's terminal statuses; a test pins that the landing archive's output is unchanged for done cards and that a punt card mirrors as punt. Commit.

### Task 4: `Abandoned` record
**Files:** create `internal/issue/abandoned.go`, `abandoned_test.go`; modify `internal/issue/handoff.go` (envelope field).
- [ ] Round-trip, validation (bad ref, bad OID, missing head), unknown-key preservation, clear. Commit.

### Task 5: `sdlc abandon`
**Files:** create `abandon.go`, `abandon_test.go`, `helptext/abandon.md`; register in `main.go`; recovery contract (class convergent-retry).
- [ ] `TestAbandonStartedWork` (fixture: claim, start-plan, a code commit): `abandon --as punt --reason r` → no local or remote issue branch; `refs/ariadne/abandoned/NNNNNN` on origin at the tip, which contains the Log note commit; card `punt` with the record and the claimant kept; main has `workshop/history/issues/X` (mirror says punt, Log has the note) and no `workshop/issues/X`; the checkout is on its resting branch.
- [ ] `TestAbandonOpenIssue`: no branch steps, card `wontfix`, details archived.
- [ ] Refusals table, each row asserting its own message (lesson: no shared invariant may refuse first): not owner, dirty tree, resting branch with started work, bad `--as`, empty `--reason`, terminal without the record.
- [ ] `TestAbandonRerunResumes`: interrupt after each of steps 3, 4 and 5 and inside step 6 (remote branch deleted, still on the branch; switched, local branch still present), rerun → same end state, no duplicate commits.
- [ ] `abandonDecision` table test over the model's statuses. Commit.

### Task 6: set-status redirect and reopen restore
**Files:** `setstatus.go`, `setstatus_test.go`, `helptext/set-status.md`.
- [ ] `set-status punt` on working/blocked/codecomplete refuses naming `sdlc abandon --issue N --as punt --reason …`, `--force` included; from open it proceeds.
- [ ] `TestReopenRestoresAnAbandonedIssue`: abandon (punt) as above, advance main with an unrelated commit, then `set-status working` from the resting branch → the issue branch exists at a commit containing the archived tip and main; `workshop/issues/X` holds the archived details and `history/` has no copy on the branch; origin has the branch; the archive ref is gone; the card is `working` with no record. Then `sdlc pr --dry-run` passes the transfer guard.
- [ ] Rerun variants of the reopen: interrupted after step 2, with a merge in progress (code conflict; the card stays terminal), after step 4, and after step 6 (orphan ref left; a later abandon overwrites it).
- [ ] File the follow-up: "reopening a done issue does not un-archive its details" (`sdlc issue new`). Commit.

### Task 7: docs and close
- [ ] Atlas: abandon in `atlas/workflow/issue-tracker.md` (verb table, the archive ref, reopen) and `atlas/workflow/issue-lifecycle.md` (wontfix/punt from active work go through abandon); `atlas/index.md` unchanged unless a file is added.
- [ ] `make test`; `sdlc milestone-close --issue 286 --milestone M2`; `sdlc close --issue 286`.

## Revisions

- 2026-10-08 (plan-quality round 1): added Non-goals and Operating envelope (PQ-4, PQ-7). D7 detects a rerun from the card record before step 1's branch checks, and step 6 deletes the remote branch before switching (PQ-1). D9 gained per-step rerun detection, the card write before the archive ref's deletion, and abandon's step 3 tolerates the orphan that ordering can leave (PQ-2). Step 5 reuses `archiveDestination`/`archivedDetails` generalized to terminal statuses and moves the issue's plan artifacts on main (PQ-3, new Task 4a). Re-close after a rebase goes through the #301 reopen path (PQ-8). The set-status seam and who may reopen are stated (PQ-6). Test prose compressed (PQ-5).
