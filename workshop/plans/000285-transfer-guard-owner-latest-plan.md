# Transfer guard: owner plus based-on-latest — Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** The transfer guard decides a landing's change to published details by the card's owner and by whether the branch is based on main's latest version of the file, never by branch names or handoff records.

**Architecture:** `guardTransferredDetails` stays the one seam that `pr`, the publish gate (push, merge) and the durable landing call. Its body becomes: compute the prospective merge of fresh main and HEAD, collect the issue details paths it changes, and judge each with a pure verdict over three observed facts (published on main?, this checkout owns the card?, HEAD contains main's last commit to the path?). A new `sdlc issue restore --issue N` reuses the same collection to put main's version back on a non-owner branch. The handoff record is no longer read by the guard.

**Tech Stack:** Go (`cmd/sdlc`), git plumbing (`merge-tree --write-tree`, `rev-list`, `merge-base --is-ancestor`), the in-repo tracker fixtures (`newTrackerRepo`, `spinOffOnCodeBranch`, `closeReady`, `withClaimant`).

---

## Design decisions

- **D1 — Protected set = published details, not handed-off details.** A path is protected when it is an issue's details file (basename equals a tracker card's basename) and main's history has ever touched it (`git rev-list -1 <main> -- p` non-empty). That covers move-detail handoffs, `issue publish` republishes, landed issue branches and archived copies (`workshop/history/issues/…`, whose deletion from `workshop/issues/` is "the last change"). Never-published details are unguarded (a new issue's first landing). The handoff record stays on the card for move-detail's idempotency, but the guard stops reading it, so `ownerAuthored` and `validHandoffDestination` (the guard was the record's only path consumer) go.
- **D2 — "Changes" means the prospective merge differs from main.** `merge-tree --write-tree <main> HEAD`; a protected path that conflicts, or whose blob/presence in the result differs from main's, is changed. Unchanged paths pass whoever holds the branch (net-zero handoff, merged-main-and-left-alone).
- **D3 — Owner = card claimant matched against this checkout** via the existing `ownership(env, card)` (repository + machine + worktree). `OwnershipMine` only; unknown (no claimant) and another workspace are both "not owner". Branch names play no part, so a renamed close branch (pair#365) or a reopened issue lands from the owner's checkout.
- **D4 — "Based on main's latest" = ancestry of main's last commit to the path.** `c = git rev-list -1 <main> -- p`; based iff `git merge-base --is-ancestor c HEAD`. This is the Spec's literal condition ("contains the commit that last published it"). A squash- or rebase-landed owner branch is then "behind" until it merges main, which is the intended next action.
- **D5 — Verdict order: not-owner before behind.** A non-owner's refusal offers `sdlc issue restore --issue N` (mechanical) and names `sdlc claim --issue N` to keep the edit. The owner's refusal when behind (or conflicting) says to merge main and resolve the prose, then rerun.
- **D6 — Unreadable cards fail closed only where they matter.** The old guard refused every PR repo-wide while any card was unreadable (it might hide a handoff record). Ownership now comes from the card of a *changed* path, so the guard refuses only when a changed path's basename matches an unreadable card's path basename. A changed published details path with no card at all also refuses (fail closed: no owner can be established). The atlas line about repo-wide refusal is updated.
- **D7 — `issue restore` is a narrow verb, not an automatic repair.** Dropping another slot's edit silently inside `pr` would lose work; the guard refuses and the operator/agent runs the restore deliberately. Restore writes main's blob to each changed path of issue N (or removes it when main has none), commits `#N: issue: restore main's details`, and refuses a dirty path. It acts only on paths the shared collector reports as changed, so it is a no-op (and says so) when the branch already lands cleanly.

ARCH notes: ARCH-DRY — one collector (`changedDetails`) feeds both the guard and restore. ARCH-FUNERAL — creates nothing durable beyond an ordinary commit on the caller's branch; no new file family, record or ref. Simplicity — the guard loses the handoff-record branch logic, the `ownerAuthored` ancestry walk and the stacked-branch exemption (#272 forbids stacking).

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `detailsVerdict` | `cmd/sdlc/transferverdict.go` | new |
| `detailsFacts` | `cmd/sdlc/transferverdict.go` | new |

- **detailsFacts** — observed facts for one changed details path: `Published`, `Owner` (bool), `Based` (bool), `Conflict` (bool).
- **detailsVerdict(f detailsFacts) verdict** — `accept` when not published; else `notOwner` when not owner; else `behind` when conflicting or not based; else `accept`. Table-tested over the full 2⁴ product in `transferverdict_test.go`, no IO.
  - **DRY rationale:** the guard's refusal text and restore's selection both switch on this one verdict.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `changedDetails` | `cmd/sdlc/transferguard.go` | new | git merge-tree / rev-list / merge-base, tracker snapshot, `ownership` |
| `guardTransferredDetails` | `cmd/sdlc/transferguard.go` | modified | the above, formatted as a refusal |
| `runIssueRestore` / `newIssueRestoreCmd` | `cmd/sdlc/issuerestore.go` | new | git write/rm/commit on the current branch |
| `ownerAuthored`, `validHandoffDestination`, `checkTransferredPaths` | `cmd/sdlc/transferguard.go` | deleted | — |

- **changedDetails(env, mainTip) ([]changedDetail, error)** — `changedDetail{ID, Path string; Verdict verdict}` for every protected path the prospective merge changes, sorted by path. Facts come from real git against the tracker fixtures (no mocks); `claimantIdentity` is the existing in-process seam (`withClaimant`).
- **guardTransferredDetails** — keeps its signature and the `guardTransferredDetailsFn` seam; legacy repositories still return nil; a vanished tracker still errors `tracker.ErrCutover`.
- **runIssueRestore** — `sdlc issue restore --issue N[,N…]`, mutating (repo lock via `markMutatingCommand`), tracker repositories only.

## Tasks

### Task 1: Pure verdict

**Files:** Create `cmd/sdlc/transferverdict.go`, `cmd/sdlc/transferverdict_test.go`.

- [ ] Write `TestDetailsVerdict` over all 16 fact combinations: unpublished → accept; published ∧ ¬owner → notOwner; published ∧ owner ∧ (conflict ∨ ¬based) → behind; published ∧ owner ∧ based ∧ ¬conflict → accept. Run `go test ./cmd/sdlc -run TestDetailsVerdict` → FAIL (undefined).
- [ ] Implement `type verdict int` (`verdictAccept`, `verdictNotOwner`, `verdictBehind`), `detailsFacts`, `detailsVerdict`. Run → PASS.
- [ ] Commit `#285: transfer guard: pure owner/based-on-latest verdict`.

### Task 2: Collector + guard rewrite (TDD against the existing fixtures)

**Files:** Modify `cmd/sdlc/transferguard.go`, `cmd/sdlc/transferguard_test.go`.

- [ ] Rewrite `TestTransferGuardOverBranchShapes` rows for the new semantics. Each row gains `owner bool` (the fixture checkout is #8's claimant via `withClaimant` + `withOwner` on the card, else another claimant) and asserts its own refusal text (lesson: each rejection row asserts its own check):
  - net-zero source, any owner → accept; source merged main and left details alone → accept; unrelated branch → accept.
  - source re-adds its stale copy, non-owner → `notOwner` text (`sdlc issue restore --issue 8`); same as owner → `behind` text (`merge main`).
  - source deletes / rewrites details after merging main, non-owner → `notOwner`.
  - owner's checkout on a **renamed** branch (`000008-spin-off-close`) edits after merging main → accept (the pair#365 core).
  - owner's checkout: local main merges the owner branch for a direct push → accept; main edits details after that merge → accept (owner's checkout; this row flips from the old guard, by design D3).
  - non-owner checkout: main merges a non-owner rewrite → `notOwner`.
  - drop "a branch stacked on the owner's" (#272 forbids stacking).
- [ ] Add `TestTransferGuardOwnerBehindMainLatest`: owner branch from main; a direct commit to origin main edits #8's details (another slot's publish); owner edits details on the branch → refused with the merge-main text; `git merge origin/main` resolving to the owner's text → accepts.
- [ ] Replace `TestTransferGuardRefusesMalformedHandoffRepoWide` with `TestTransferGuardUnreadableCardRefusesOnlyItsDetails`: an unreadable card #9 refuses a branch changing #9's published details, and does not refuse an unrelated branch.
- [ ] Adapt `TestTransferGuardInterruptedRemovalProtectsOwnerEdits`, `TestTransferGuardFromAFreshCloneUsesOnlyTrackerRecords` (the fresh clone is a non-owner worktree → rewrite refused as `notOwner`) and `TestTransferGuardProtectsAnUnrecordedPublication` (still protected without a recorded handoff: now because main's history touched the path). Keep `TestPublishGateAndPRRefuseChangedHandedOffDetails` (production entry points).
- [ ] Run `go test ./cmd/sdlc -run 'TransferGuard|PublishGateAndPR'` → FAIL.
- [ ] Implement `changedDetails`: merge-tree result + conflicted names; index card records and unreadable cards by path basename; for each path changed vs main (`git diff --name-only <main> <result>` plus conflicts) whose basename matches a card: published probe (D1), `ownership` (D3), ancestry (D4); unreadable-match or published-without-card → error (D6). Rewrite `guardTransferredDetails` to format the first non-accept verdict per issue; delete `ownerAuthored`, `checkTransferredPaths`, `validHandoffDestination`; replace `transferRefusal` with two messages:
  - notOwner: `landing would change published details of #N (<path>) from a checkout that does not own #N. Restore main's version with `sdlc issue restore --issue N`; to keep the edit, `sdlc claim --issue N` first.`
  - behind: `landing would change #N's details (<path>) from a branch not based on main's latest version of them. Merge main (`git fetch <remote> && git merge <remote>/main`), resolve the details prose, and rerun.`
  Both wrap `errTransferredDetails`.
- [ ] Run the focused tests → PASS. Commit `#285: transfer guard: decide by owner and based-on-latest`.

### Task 3: `sdlc issue restore`

**Files:** Create `cmd/sdlc/issuerestore.go`, `cmd/sdlc/issuerestore_test.go`; modify `cmd/sdlc/issue.go` (AddCommand), `cmd/sdlc/helptext/issue.md`.

- [ ] Test `TestIssueRestoreMakesAStaleFilingBranchLand`: `handedOff` fixture, non-owner checkout, source re-adds its stale copy → guard refuses naming `sdlc issue restore --issue 8`; run `runIssueRestore(ctx, out, errs, []int{8})` → one new commit `#8: issue: restore main's details`; guard → nil; `sdlc pr --dry-run` passes the guard.
- [ ] Test restore removes a path main has archived (stale branch keeps `workshop/issues/X` after main moved it to history): restore `git rm`s it; guard → nil.
- [ ] Test restore refuses a dirty details path, and reports "nothing to restore" when #N has no changed details (no commit created).
- [ ] Implement: open tracker, `changedDetails`, filter by requested IDs, refuse dirty paths, write main's blob or remove, `git add`/`rm`, one commit per invocation naming every restored issue. Mark mutating. Help text section in `helptext/issue.md`.
- [ ] Run → PASS. Commit `#285: issue restore: put main's details back on a non-owner branch`.

### Task 4: pair#365 end to end

**Files:** Test in `cmd/sdlc/transferguard_e2e_test.go` (new), reusing `closeReady`, `stubJudge`, `executeSDLCTestCommand` and the merge e2e gh fake (`merge_e2e_test.go` harness).

- [ ] `TestPair365CloseOnRenamedBranchLands`: owned issue on its branch; push it and merge it into origin main outside sdlc (`git merge --no-ff` in a clone, push); card still `working`. In the owner checkout create `<id>-slug-close` from the old branch tip, run `sdlc close` there, then `sdlc pr` and `sdlc merge` (gh fake) → card `done`. Assert the old guard's failure mode is gone (pr not refused).
- [ ] Fix any other branch-name dependency the run surfaces on the close → pr → merge path (the 2026-10-08 survey found none besides the guard: close, `ownedCompletions`, `newestClose` and PR matching key on the claimant, ancestry and the current branch) inside this issue — it is the issue's Done-when, not a follow-up.
- [ ] Commit `#285: pin the pair#365 landing end to end`.

### Task 5: Retire "owner = branch" wording; atlas

**Files:** `cmd/sdlc/transferguard.go` (comments, identifiers), `cmd/sdlc/transferguard_test.go` (row names, `ownerBranchEdit` helper), `atlas/workflow/issue-tracker.md` (~196, ~269–273). A survey (2026-10-08) found "owner" meaning a branch *only* in `transferguard.go` (`owner := issue.BranchName(...)`, `ownerAuthored`, its comments); no helptext, embedded help or atlas file uses it — the claimant-meaning uses in `helptext/move.md`, `root.md`, `reclaim.md`, `unclaim.md`, `handoff.go`, `observe/text.go`, `recovery/catalog.go` stay.

- [ ] After Task 2, `grep -n "owner" cmd/sdlc/transferguard.go` shows only claimant meanings; rename the test helper `ownerBranchEdit` → `issueBranchEdit` and row names accordingly.
- [ ] Add `TestNoHelpCallsABranchTheOwner`: for every embedded help topic (`helptext` registry) and both guard refusal messages, assert no match for `(?i)owner'?s? (issue )?branch|owner branch`. Mutation check: temporarily put "owner's branch" in a refusal → test fails.
- [ ] Atlas: rewrite the transferguard paragraph (published details; owner + based-on-latest; `issue restore`; no handoff-record dependency) and the repo-wide unreadable-card bullet (now scoped, D6). Add the `issue restore` row to the verb table.
- [ ] `make test` green. Commit `#285: retire owner-as-branch wording; atlas`.

### Close

- [ ] `sdlc close --issue 285 --verified '<make test + the four Done-when tests>'`.

## Revisions

- 2026-10-08 (Task 2): dropped the `Conflict` fact. A mutation check showed it can't be observed: a branch based on main's last commit to a file can't conflict on it, and a conflicted file already appears in the merge-result diff (it's written with markers). The verdict is the Spec's two conditions over 8 combinations. Also dropped D6's "published path with no card refuses": without a card a path isn't recognisable as details. The unreadable-card case moved into `malformedcard_test.go`, which held the old repo-wide assertion.
