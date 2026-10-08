# Boundary pushes and sdlc abandon — Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Started work is always on origin (the issue branch is pushed at every sdlc boundary and deleted remotely at merge), and ending active work as `wontfix`/`punt` goes through one verb, `sdlc abandon`, that keeps the work under an archive ref which a reopen restores.

**Architecture:** One lease push (`pushIssueBranch`, `handoff.go:205`, already used by unclaim) becomes the boundary push, called at the end of start-plan, milestone-close, close (and close's reconcile completion), and by `sdlc pr`'s durable push. Merge deletes the remote branch with a lease on the landed head. `abandon` is a convergent sequence (note commit → archive ref → card CAS → narrow main archive → branch deletion) whose card record (`abandoned{ref, head}`) lets a rerun resume and a reopen restore.

**Tech Stack:** Go (`cmd/sdlc`), git (`push --force-with-lease`, custom refs), tracker envelope records (`internal/issue`), `env.main.UpdateManyPrepared` + `gitx.TrunkWrite` for the narrow main commit; fixtures `trackerRepo`, `startedHere`, `closeReady`, `procedureFixture` + `landingFakeGH`.

---

## Design decisions

- **D1 — The boundary push is `pushIssueBranch`, unchanged.** Its lease is the last-fetched remote-tracking ref (absent → the branch must not exist remotely), never a fresh read, so a rebase followed by the next boundary force-pushes safely and a peer's unseen push is refused. One wrapper, `boundaryPush(env, stderr, verb)`, pushes the current branch when it is an issue branch (not the resting branch, not main, not detached) and reports.
- **D2 — A failed boundary push warns; it never undoes or fails the boundary.** The verb's own effect (card write, close binding) has landed; the push is durability, and the next boundary or `sdlc pr` pushes again (lesson: a projection must not add a refusal to its host verb). A stale lease is the exception worth shouting about: the warning names the remote tip and says to fetch and inspect before the next boundary. Unclaim keeps its existing hard failure (a handoff without the push is no handoff).
- **D3 — milestone-close pushes HEAD as it is.** It makes no commit (the agent commits the printed trailers), so its push carries the milestone's work and the trailer commit follows at the next boundary. The Done-when ("origin equals local HEAD after the verb") holds at the moment the verb returns.
- **D4 — `sdlc pr` (durable path) uses the same lease push** instead of a plain `push -u`, so a rebase between boundaries doesn't make `pr` fail non-fast-forward. Its upstream config write stays as is. The legacy (untracked) pr path is untouched.
- **D5 — merge deletes the remote branch with a lease on the landed head** (`push --force-with-lease=refs/heads/B:<head> <remote> :refs/heads/B`), after the local deletion in `deleteLandingBranch` (`landing.go:310`). Best-effort with a warning: the landing is done. A rerun (resume path) whose remote branch is already gone is a no-op.
- **D6 — `abandon` accepts any non-terminal status the owner holds** (`open`, `working`, `blocked`, `codecomplete`); from `open` there is no branch, so the branch steps are skipped. `--as wontfix|punt` maps to the model's `abandon`/`defer` events (the edge is checked against `construct/vocabulary/issue.cue`, not hardcoded).
- **D7 — abandon's sequence, each step detectable on rerun:**
  1. Refuse unless: owner (`requireCardOwnership`), a clean tree, and, when the issue has started work, the checkout is on a non-resting, non-main branch (the one being abandoned; the name is not checked, per #285).
  2. Append `- YYYY-MM-DD: abandoned (wontfix|punt) — <reason>` under `## Log` (same-day subheading rule as unclaim's note) and commit `#N: log: abandon (<as>)`. Skipped when that commit is already the tip.
  3. Push the tip to `refs/ariadne/abandoned/NNNNNN` on the publication remote, lease "must not exist" (or already equal to the tip on a rerun).
  4. One card CAS: status → `wontfix`/`punt`, record `abandoned: {ref, head}`, claimant kept as attribution, `updated` today. A card already terminal with the same record is the rerun case.
  5. One narrow main commit `#N: issue: abandon (<as>) — archive details`: write the final details (mirror refreshed to the terminal card) at `workshop/history/issues/X` and delete `workshop/issues/X` (`TrunkWrite{Write, Delete}`), ownership re-checked in `beforePush`. Skipped when main already has the archived copy and no live one.
  6. Switch to the resting branch, delete the local branch, delete the remote issue branch (lease on the tip). Each skipped when already gone.
- **D8 — `set-status` to `wontfix`/`punt` from `working`/`blocked`/`codecomplete` refuses toward `sdlc abandon --issue N --as … --reason …`.** From `open` (triage) it keeps working as today, and `--force` does not bypass the redirect (there would be no archive ref, so the work would be lost; this is ARCH "a gate the agent can skip is not a gate").
- **D9 — Reopen restores and un-archives (operator decision 2026-10-08).** `set-status working` on a `wontfix`/`punt` card that carries an `abandoned` record, from the owner's resting branch: fetch the archive ref; check its tip equals the recorded head; create the issue branch (`issue.BranchName`) there; merge main; move `workshop/history/issues/X` back to `workshop/issues/X` (taking the archived copy, which holds the abandon note and is newer than the branch's); commit `#N: reopen: restore from refs/ariadne/abandoned/NNNNNN`; push the branch (D1); delete the archive ref (lease on the head); clear the record in the same card CAS that reopens. A code conflict in the merge stops before the card write with the merge left for the agent and a rerun as the next action. The reopened branch then lands cleanly through the #285 guard (owner, and based on main's archive commit). Reopening a `done` issue keeps today's behavior; its un-archive gap is filed as a follow-up.
- **D10 — Lifecycle of the new things (ARCH-FUNERAL).** `refs/ariadne/abandoned/NNNNNN`: created by abandon, last needed by a reopen, removed by that reopen; otherwise it lives as long as the repository, at one ref plus the branch's unique objects per abandoned issue, bounded by the number of abandoned issues. The card's `abandoned` record: cleared by the reopen, kept as attribution on a terminal card otherwise (one small record per card). Remote issue branches: now removed by merge and abandon, where before they accumulated.

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
- [ ] Tests (fixture `startedHere`): pushes a new branch (origin equals HEAD); after `commit --amend` (a rewrite) the next push force-pushes with lease; an unseen remote tip (another clone pushed) refuses with the stale-lease warning text and leaves origin untouched; on the resting branch it pushes nothing; `deleteRemoteBranch` removes the branch at its expected head, refuses a moved one, and is a no-op when it is gone. A lost push response: pushing again when origin already has the tip but the tracking ref is stale succeeds (pins the survey's open question).
- [ ] Implement; `boundaryPush` returns nothing and writes `cok`/`cwarn` (D2).
- [ ] Commit `#286 M1: boundary push helper`.

### Task 2: the boundary verbs push
**Files:** `startplan.go` (~:344, also the rerun path that "re-checks and does nothing"), `closetracker.go` (`publishTrackerClose` end, ~:425), milestone-close finalize (`close.go` ~:1410 and `milestoneclose.go` ~:183, opening a tracker env), the close completion in `issue recovery reconcile`; recovery contracts in `internal/recovery/catalog.go` (start-plan, milestone-close, close, reconcile, pr).
- [ ] Test `TestBoundaryVerbsPushTheIssueBranch`: in `closeReady`'s flow assert `ls-remote` equals HEAD after start-plan, after a milestone-close, after close; then rebase onto an advanced main, run close again (re-close) and assert origin equals the rewritten HEAD (lease force). Test that start-plan's rerun pushes when origin lacks the branch.
- [ ] Insert `boundaryPush`; update the contracts' Effects ("pushes the issue branch with a lease; a failed push warns").
- [ ] Commit `#286 M1: push the issue branch at every boundary`.

### Task 3: `pr` lease push; merge deletes the remote branch
**Files:** `landing.go` (`runDurablePR` ~:558, `deleteLandingBranch` ~:310), `tracker_e2e_test.go` or a new `landingcleanup_test.go`.
- [ ] Tests: after a rebase, `sdlc pr` publishes the rewritten branch; `TestTrackerFullSlotCycle`'s landing leaves `ls-remote --heads <branch>` empty; the merge resume path with the remote branch already deleted succeeds.
- [ ] Implement; contracts for pr/merge updated; atlas (`atlas/workflow/issue-tracker.md`, the verb table rows for start-plan, close, pr, merge).
- [ ] `make test`; `sdlc milestone-close --issue 286 --milestone M1`.

## M2 — `sdlc abandon` and reopen

- [ ] M2 — abandon verb, set-status redirect, reopen restores and un-archives

### Task 4: `Abandoned` record
**Files:** create `internal/issue/abandoned.go`, `abandoned_test.go`; modify `internal/issue/handoff.go` (envelope field).
- [ ] Round-trip, validation (bad ref, bad OID, missing head), unknown-key preservation, clear. Commit.

### Task 5: `sdlc abandon`
**Files:** create `abandon.go`, `abandon_test.go`, `helptext/abandon.md`; register in `main.go`; recovery contract (class convergent-retry).
- [ ] `TestAbandonStartedWork` (fixture: claim, start-plan, a code commit): `abandon --as punt --reason r` → no local or remote issue branch; `refs/ariadne/abandoned/NNNNNN` on origin at the tip, which contains the Log note commit; card `punt` with the record and the claimant kept; main has `workshop/history/issues/X` (mirror says punt, Log has the note) and no `workshop/issues/X`; the checkout is on its resting branch.
- [ ] `TestAbandonOpenIssue`: no branch steps, card `wontfix`, details archived.
- [ ] Refusals, each asserting its own message: not owner; dirty tree; on the resting branch with started work; `--as` other than wontfix/punt; empty `--reason`; already terminal without the record.
- [ ] `TestAbandonRerunResumes`: interrupt after each of steps 3, 4 and 5 (hook or seam), rerun → same end state, no duplicate commits.
- [ ] `abandonDecision` table test over the model's statuses. Commit.

### Task 6: set-status redirect and reopen restore
**Files:** `setstatus.go`, `setstatus_test.go`, `helptext/set-status.md`.
- [ ] `set-status punt` on working/blocked/codecomplete refuses naming `sdlc abandon --issue N --as punt --reason …`, `--force` included; from open it proceeds.
- [ ] `TestReopenRestoresAnAbandonedIssue`: abandon (punt) as above, advance main with an unrelated commit, then `set-status working` from the resting branch → the issue branch exists at a commit containing the archived tip and main; `workshop/issues/X` holds the archived details and `history/` has no copy on the branch; origin has the branch; the archive ref is gone; the card is `working` with no record. Then `sdlc pr --dry-run` passes the transfer guard.
- [ ] A code-conflict variant stops before the card write with the merge in progress and a rerun as the next action.
- [ ] File the follow-up: "reopening a done issue does not un-archive its details" (`sdlc issue new`). Commit.

### Task 7: docs and close
- [ ] Atlas: abandon in `atlas/workflow/issue-tracker.md` (verb table, the archive ref, reopen) and `atlas/workflow/issue-lifecycle.md` (wontfix/punt from active work go through abandon); `atlas/index.md` unchanged unless a file is added.
- [ ] `make test`; `sdlc milestone-close --issue 286 --milestone M2`; `sdlc close --issue 286`.
