# Reconcile merges done outside sdlc — Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** An issue branch merged into main outside sdlc (the GitHub web button) is reported by `sdlc state` and, when its close evidence is on main, finished by `reconcile` or the next `sdlc merge` (card done, details and plans archived, remote branch deleted); without the evidence, sdlc names the next action and changes nothing.

**Architecture:** One pure classifier, `externalMergeVerdict`, over facts observed per started issue (card status, completion binding on main, pushed branch merged into main, ownership). `sdlc state` renders its findings read-only. A finisher, `finishExternalMerge`, reuses `settleLandedCompletions` for the card, and a narrow main archive commit extracted from `abandon` (`archiveIssueOnMain`) for the details and plans, then deletes the remote branch with the #286 lease. `reconcile --issue N` and `sdlc merge` (after its own landing) call it.

**Tech Stack:** Go (`cmd/sdlc`), git ancestry against fetched refs only (no GitHub API), fixtures `procedureFixture`/`closeReady` and the #286 boundary-push remote.

---

## Non-goals

- Squash or rebase merges done outside sdlc: they leave neither the branch tip nor the evidence commit on main, so local ancestry can't see them. Detecting them needs the GitHub API (PR state), which the Spec defers with the required check.
- Writes from `sdlc state`: it stays a read (it reports, never refuses or mutates); the writing verbs finish the bookkeeping.
- Legacy repositories (no tracker): no cards to reconcile.

## Design decisions

- **D1 — "Merged outside sdlc" is ancestry on fetched refs.** For each card with a started, non-terminal status (working, blocked, codecomplete): the issue branch's pushed tip (`refs/remotes/<remote>/<details stem>`, which #286's boundary pushes keep current; the local branch as fallback) has commits beyond its base and is contained in fresh main. A codecomplete card whose completion binding's evidence commit is on main counts as landed whatever the branch. `sdlc state` adds no fetch: it judges the cards it already read (the command's cached records) against main as last fetched (Revisions).
- **D2 — The verdict is pure** over `{status, evidenceOnMain, branchMerged, ownership}`:
  - evidence on main → **settle** (finish the bookkeeping);
  - branch merged, no evidence on main → **close needed**: from the owner's slot, `sdlc close --issue N` on a branch from main (the #285 guard lets a renamed close branch land); unowned → **claim first**;
  - otherwise → nothing.
- **D3 — `sdlc state` reports, it never writes.** Each verdict becomes a `DriftFinding` (warn) naming the exact next action: `sdlc issue recovery reconcile --issue N` for settle, `sdlc close --issue N` or `sdlc claim --issue N` otherwise. It replaces the commit-subject close-off heuristic only for issues D1 classifies (the heuristic still covers branchless legacy work).
- **D4 — The finisher is repository-wide and convergent** (Revisions): card → done stays `settleLandedCompletions`; then `finishLandedLeftovers` archives every done card whose details are still live on main (`archiveIssueOnMain`, extracted from `abandon.go`; main's copy mirrored to the done card), and deletes every done card's issue branch still on the remote once main contains it (one `ls-remote --heads`, so a failed delete is retried next run). The local branch is left for its slot (it may be checked out there).
- **D5 — Callers.** `reconcile` runs the finisher after its existing settle, and prints the next action for the named issue if it was merged without a close. `sdlc merge` runs it at the end of its own landing (the Spec's "next merge"), with its own `--issues-dir/--plans-dir/--history-dir`, warning, never failing the landing. Ownership: settling needs none (the evidence binds the close, as `settleLandedCompletions` already assumes); the archive's `beforePush` keeps the existing check that the card's owner, if any, is this workspace or the card is done.
- **D6 — Lifecycle (ARCH-FUNERAL).** Creates nothing durable: one archive commit per reconciled issue, and it removes a remote branch that would otherwise linger.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `externalMergeVerdict` / `mergeFacts` | `cmd/sdlc/externalmerge.go` | new |

- **externalMergeVerdict(f mergeFacts) mergeVerdict** — settle / closeNeeded / claimNeeded / none, table-tested over every combination in `externalmerge_test.go`, no IO.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `externalMerges(env, rs, mainTip)` | `cmd/sdlc/externalmerge.go` | new | `ownedCompletions` (evidence on main), git ancestry, `ownership` |
| `finishLandedLeftovers(env, stderr, dirs)` | `cmd/sdlc/externalmerge.go` | new | `archiveIssueOnMain`, `deleteRemoteBranch` (after `settleLandedCompletions`) |
| `reportUnclosedMerge` | `cmd/sdlc/externalmerge.go` | new | `externalMerges` |
| `archiveIssueOnMain` | `cmd/sdlc/archivemain.go` | new (extracted from `archiveAbandoned`) | `mainPublish` |
| `archiveAbandoned` | `cmd/sdlc/abandon.go` | modified (calls `archiveIssueOnMain`) | — |
| `detectDrift` / `runState` | `cmd/sdlc/state.go` | modified | + external-merge findings |
| `runRecoveryReconcile` | `cmd/sdlc/issuerecovery.go` | modified | + finisher |
| `runDurableMerge` | `cmd/sdlc/landing.go` | modified | + finisher for settle verdicts |

## Tasks

### Task 1: pure verdict
- [ ] `TestExternalMergeVerdict` over every combination of the facts; implement; commit.

### Task 2: extract `archiveIssueOnMain`
- [ ] Move the main archive commit out of `archiveAbandoned` (params: the final details or nil for main's copy, extra plan contents, an optional Log line); `archiveAbandoned` calls it. The abandon tests stay green unchanged (pure refactor). Commit.

### Task 3: detection and `sdlc state`
- [ ] Fixture (`externalMergeFixture`): a tracker repo with an owned, started issue whose branch is pushed and merged into origin main with `git merge --no-ff` from another clone (the web button), card not done. Variants: closed (codecomplete, evidence on main), not closed, unowned.
- [ ] `TestStateReportsExternalMerges`: `sdlc state` (text and `--json`) shows a warn finding per variant with its next action, and changes nothing (tracker and main unchanged).
- [ ] Implement `externalMerges` and the findings; commit.

### Task 4: the finisher, via reconcile and merge
- [ ] `TestReconcileFinishesAnExternalMerge`: closed variant → `reconcile --issue N` → card done, details and plans archived on main mirroring done, remote branch gone; a rerun changes nothing. Not-closed variant → reconcile prints the close action and changes nothing.
- [ ] `TestMergeFinishesEarlierExternalMerges`: a closed external merge, then an ordinary `sdlc merge` of another issue → both done and archived.
- [ ] Implement; contracts for reconcile and merge; help (`state`, `recovery`); atlas (`issue-tracker.md`, `issue-lifecycle.md`); commit.

### Close
- [ ] `make test`; `sdlc close --issue 287`.

## Revisions

- 2026-10-08 (close review round 1, REWORK): the finisher is repository-wide (`finishLandedLeftovers`), not per issue (`finishExternalMerge`): a landing settles outside merges inside `completeLandingPR`, before its own archive, so the settled IDs aren't available afterwards. "Every done card whose details are still live on main, after the landing archived its own" is the same set and converges (BR-1). Its branch sweep lists the remote's heads once, so a failed delete is retried (BR-3). `sdlc state` adds no fetch: it reuses the command's cached records and main as last fetched (BR-2). Evidence on main comes from `ownedCompletions`, one source with the settle (BR-5). The branch tip is the newer of local and remote-tracking (BR-4). Merge passes its own dirs (`archiveDirs`, BR-6). The settle finding names the interrupted-merge case too (BR-8). The close-off guess is suppressed by a finding kind, not a message prefix (BR-9).
