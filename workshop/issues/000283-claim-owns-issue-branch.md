---
id: 000283
status: working
deps: [ariadne#277, ariadne#278]
github_issue:
created: 2026-10-02
updated: 2026-10-07
estimate_hours: 3.48
card_mirror: 'cdb17dbbfbd10598b39347388f7094dfd48273f7' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-02T10:51:37-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
actual_hours: 2.11
---

# Re-derive ownership from the claimant

## Problem

#277 added a recorded owner to issues (the claimant), but the rest of sdlc still reasons about ownership through older proxies:

- **Status as the lock.** `open` is claimable and `working` is not. This predates the claimant. It leaves `blocked` and `codecomplete` issues effectively unowned, although someone still has to finish them.
- **Branch name as owner.** The transfer guard (`cmd/sdlc/transferguard.go`) calls the branch named after the details file the "owner": `owner := strings.TrimSuffix(path.Base(h.Destination), ".md")`. It ignores the claimant. Once a PR from that name merges, #148 retires the name, and nothing can legitimately edit the handed-off details again, not even the issue's own close.
- **No release.** There is no unclaim. The lifecycle model has no `working → open` edge, so "claim to edit details, then release" leaves the card `working` and a branch behind.

Observed on pair#365. PR #193 merged on GitHub (not `sdlc merge`) at 17:28:09. `sdlc close` finished a minute later on a hand-made `000365-message-lifecycle-close` branch and recorded `evidence_commit 64666d92`. Then three verbs refused:

- `sdlc pr`: "landing would change handed-off issue details".
- `reclaim`: no recorded owner (pre-#277).
- `claim --adopt`: the issue is codecomplete.

The close was landed by hand; the card is stuck at `codecomplete`.

## Spec

Captured for operator review; no implementation is authorized by this issue creation. **Scope is a rethink**: with the claimant in place, several verbs and guards need to be re-derived from it. Expect this issue to be split into child issues, or promoted to a project, during design.

### Principle

The **claimant is the owner and the lock**. **Status is lifecycle only.**

- An owner is a `Claimant` (`cmd/sdlc/internal/issue/claimant.go`): operator, machine fingerprint and name, optional workspace label (`repo:N`), canonical worktree path, and repository. Couch slots are optional, so the checkout path is what identifies it.
- Ownership persists through `working`, `blocked` and `codecomplete`, and ends at release or a terminal status.
- Terminology: say **"issue branch"** for the derived branch and **"owner"** only for the claimant. The guard's current "owner = branch" wording goes.

### Intended lifecycle (operator statement, 2026-10-02, amended)

1. An issue can be created on any branch; its card is on `issue-tracker` immediately.
2. Details are worked out over several rounds, so they are not published early. They stay on the filing feature branch and land with it, or stay untracked on a resting branch.
3. `move-detail` publishes the details, making the issue real and claimable.
4. Once published, only the owner may change the details. Claim records the owner with an atomic compare-and-swap on the card, so only one process can hold it.
5. Claim creates the local issue branch, named from the issue.
6. The issue branch can move between local slots (`sdlc move`), and the recorded owner moves with it.
7. An issue can be unclaimed: the owner is cleared and the card goes back to `open`.
8. Editing published details requires being the owner.

### Surfaces to re-derive from the claimant

- **claim:** creates the issue branch in the claiming checkout, from main. It refuses where `start-plan` does today (#272: checkout on another issue's unlanded branch). `start-plan` keeps the design-principles delivery and becomes idempotent about the branch.
- **unclaim (new):** add a `working → open` release edge to `construct/vocabulary/issue.cue`. The verb clears the claimant by compare-and-swap.
  - No unlanded work on the branch: delete the branch.
  - Only edits to this issue's own details: publish them narrowly to main first, so a claim-to-edit cycle never loses work.
  - Code on the branch: refuse with the next action.
- **transfer guard:** decide edits by owner, not by branch name. A landing is the owner's when it runs from the recorded owner's checkout. The guard runs in a checkout and compares it with the card, because commits don't carry the claimant. A retired or renamed branch no longer strands the owner.
- **claim --adopt:** also covers pre-#277 issues that are `blocked` or `codecomplete`, so stuck issues like pair#365 can get an owner and finish through `sdlc merge`.
- **reopen:** `codecomplete → working` and `done → working` reopen an issue whose branch name #148 has retired. Derive a fresh issue-branch name (e.g. a suffix); the owner-based guard accepts it.
- **move / reclaim:** re-check that `sdlc move` and #278 `reclaim` update the owner consistently with the above. `RelocationAllowed` already exists.

### Open questions

- **Merges outside sdlc.** *Decided 2026-10-02: detect after the fact; no dependency on GitHub CI for now.* A GitHub merge (web button or `gh pr merge`) while the card is not `codecomplete` started pair#365. Have a routine sdlc read, such as `sdlc state` or the next `merge`, notice "issue branch merged but card not `done`". It should then finish the bookkeeping when the close evidence is present (flip to `done`, archive), or print the next action when it isn't. A required GitHub check is deferred.
- **Publishing on unclaim.** Should the details-only publish happen automatically, or need a flag?
- **Splitting.** Settle the child-issue split. Candidates: (a) owner-as-lock in the model and status vocabulary, (b) claim creates the branch, plus unclaim, (c) owner-based transfer guard, plus adopt and reopen naming, (d) detecting merges outside sdlc.

## Done when

- `construct/vocabulary/issue.cue` states the owner is the lock and status is lifecycle only. Claim and unclaim change the owner, never the status. `open`-with-owner and active-without-owner are legal states, and their meaning is documented.
- `claim` no longer changes status. `start-plan` requires the claim, sets `working`, and creates the issue branch from main (the #272 refusal stays there).
- Unclaim from `working`/`blocked`/`codecomplete` is legal in the model (the verb itself is #284).
- `start-plan` carries uncommitted edits confined to the claimed issue's own details onto the new branch; any other dirty tracked file still refuses.
- `change-code` refuses an owned but unstarted (`open`) issue, pointing at `start-plan`; `reclaim` accepts an owned `open` card.
- A close survives a rebase: after a rebase onto main, reopening and re-closing the issue lands. The card's earlier close, whose reviewed commit the rebase rewrote off the branch, no longer blocks it; a stale receipt from before the rebase is still refused.
- The atlas terminology defines slot = executor/owner and operator = supervisor, and no help text uses "owner" for a branch or a person.
- Tests pin the legal (status, owner) combinations and the claim/start-plan split.

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec             design=1.2 impl=0.08
item: smaller-go-module      design=0.2 impl=0.16
item: smaller-go-module      design=0.1 impl=0.14
item: smaller-go-module      design=0.2 impl=0.2
item: smaller-go-module      design=0.05 impl=0.1
item: cross-cutting-refactor design=0.1 impl=0.2
item: atlas-docs             design=0.1 impl=0.08
item: milestone-review       design=0.0 impl=0.14
item: milestone-review       design=0.0 impl=0.14
design-buffer: 0.15
total: 3.48
```

Items, in plan order:
- the four-round design, the project and the plan;
- Task 1, the vocab axis;
- Task 2, claim;
- Task 3, start-plan plus the dirty carry;
- Task 4, change-code and reclaim;
- Task 5, the e2e fixture sweep;
- Task 6, docs;
- the plan-quality and close reviews.

The design buffer is +15% because a thorough plan doc exists.

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only.*

## Plan

- [x] Claim, run start-plan, settle the open questions and the child-issue split before designing the implementation.
- [x] Design the model change; durable plan at `workshop/plans/000283-claimant-lock-model-plan.md` (full flow: >100 code lines).
- [x] Model: `ownership` axis in issue.cue + pkg/vocab; open→working is `start` (plan Task 1)
- [x] claim records the owner only (Task 2)
- [x] start-plan starts the lifecycle; carries own details edits (Task 3)
- [x] change-code refuses unstarted; reclaim derives holdable statuses (Task 4)
- [x] e2e fixtures: claim then start-plan (Task 5)
- [x] Help text, atlas, terminology (Task 6)

## Log

### 2026-10-02
- 2026-10-02: closed — make test green (898 cmd/sdlc tests + 19 pkgs; processgroup /bin/ps test sandbox-only, passes unsandboxed); round-2 Minor fixed: start-plan admits only the start edge source/target (open, working), asserted over the model status x owner product in TestStartDecision; earlier evidence stands: owned guard on start, TestShapeUnderClaimThenStart e2e, TestClaimNeverMovesStatus, make vocab-embed clean; review verdict: SHIP
- 2026-10-02: closed — make test green after the minor fixes (898 cmd/sdlc tests + 19 pkgs; processgroup /bin/ps test is sandbox-only and passes unsandboxed); owned guard on start enforced in set-status with tests (unowned refuses, --force waives, held starts); swept every test calling set-status working; prior evidence: TestShapeUnderClaimThenStart e2e, TestClaimNeverMovesStatus and TestStartDecision over the model status x owner product, make vocab-embed clean; review verdict: SHIP
- 2026-10-02: closed — make test green (898 cmd/sdlc tests + 19 pkgs; processgroup ps test sandbox-only, passes unsandboxed); make vocab-embed clean; TestShapeUnderClaimThenStart e2e: claim leaves card open+owner, shaping edit carried, start-plan starts card keeping stamp, rerun no-op; TestClaimNeverMovesStatus + TestStartDecision over the model status x owner product; review verdict: SHIP

Filed from a pair session that landed pair#365 by hand after `sdlc pr`, `reclaim` and `claim --adopt` all refused. Root cause traced to `transferguard.go` deriving the owner from the details basename. Operator review the same day: the claimant should be the owner and the lock (status-as-lock predates the claimant), and the scope is a rethink of the verbs and guards around the claimant. The issue is likely to split. Details left uncommitted on this resting branch, not moved.

Design discussion (claimed in ariadne:2). Decided with the operator:
- Unclaim is allowed from `blocked` as well as `working`; not from `codecomplete`.
- Stale locks: explicit `reclaim` with a reason, plus `sdlc state` showing claim age. No leases or heartbeats.
- Transfer guard condition: a details change lands only from the owner's checkout AND from a branch whose merge-base contains the latest published version (the handoff `MainCommit`). The content condition is what stops stale overwrites; branch names stop mattering.
- **Claim does not create the branch; `start-plan` does.** Pre-implementation editing (follow-ups filed from a console, one brainstorm revising several related issues) must not need one branch per issue. A slot may hold several claims at once: claim the set, edit their details, publish narrowly to main (published can't be unpublished), unclaim.
- Claimant = slot on a machine (the agent's durable workspace), operator = supervising human. This divergence from Jira-style person-assignment is intentional; an atlas terminology entry is still to be written.

Second round:
- **Claim never touches status, and neither does unclaim.** Unclaiming started work is a handoff, not abandonment: the card keeps `working`/`blocked`/`codecomplete`, the owner is cleared, and the issue is open for takeover. Active-with-no-owner is a legal state. That replaces the planned `→ open` edge, and `claim --adopt` folds into an ordinary claim of an active, unowned card.
- Handoff design (to settle): unclaim refuses on a dirty tree, pushes the issue branch, and records `{branch, head}` on the card plus a handoff note in `## Log`. A takeover claim fetches the branch, checks the head matches, and checks it out in a clean slot. `reclaim` (with a reason) is the forced path when the owner can't unclaim. `move` is the same-operator local fast path.
- Shaping time: the window anchors at `start-plan`, and each claim also counts the full shaping segments of the commits it rode in. Agents may record pre-claim shaping time as an estimate; drop that part if it proves too imprecise.
- Order: model first. Doing it properly, not fast-tracking the guard for pair#365.
- Merges outside sdlc: detect and reconcile. There must always be an escape hatch.

Third round:
- The handoff note goes in `## Log`, unstructured. The next agent needs the raw information, not a schema.
- Unclaim pushes automatically.
- `move` stays its own verb: it relocates between worktrees on one machine, which share an object store, so it needs no push. Unclaim + claim is the cross-machine or cross-operator path, and it goes through origin.
- Who resolves guard refusals: restoring main's version is mechanical, so sdlc does it. A semantic merge of details prose is the owner slot's LLM's job. Under the lock, stale-base conflicts only happen across a handoff, and the claim-time refresh from main prevents most of them.
- Insight: the lock plus #272 (no stacking) means an owned issue branch has a single writer, so the owner can force-push it with lease. Pushing more often doesn't cost the ability to rebase. The push policy is still open.

Fourth round:
- Push policy: push the issue branch at every sdlc boundary (`start-plan`, `milestone-close`, `close`, `unclaim`). `merge` deletes the remote issue branch.
- Claiming several issues at once is atomic: one tracker commit compare-and-swaps all the cards, and on a race it re-reads, re-checks every card and retries.
- `sdlc issue publish --issue a,b,c` republishes owned details edits narrowly to main. `move-detail` stays the first publication.
- No `abandon` verb exists; `wontfix`/`punt` go through `set-status`, which leaves the branch, the remote branch and the owner behind. Proposed: `sdlc abandon` (sketch in the reply; to settle).

Implementation (plan Tasks 1–4 and 6), one commit each:
- vocab axis `62d27b83`;
- claim `ad040794`;
- start-plan `9ff1d044`;
- change-code/reclaim `69537645`;
- docs `6c294eb0`.

Notes:
- `claimDecision` keeps its legacy arm keyed on a nil identity (`legacymode.go:275`), which matches the model's scope gloss.
- `adoptDecision` now takes any active status, so a pre-#277 `codecomplete` card can be adopted.
- AGENTS.md/CLAUDE.md are weave outputs; the source edit is `AGENTS.base.md`. The local entry files refresh on the next `weave compile`.
- Left for #285: the transfer guard's code still names the branch `owner` (`transferguard.go:60`).
- Plan-quality ran three rounds. Round 1 hit a sandbox block on the reviewer's API host and recorded no findings; round 2 raised six Minors, all folded into the plan's Revisions; round 3 dispositioned them.

Verification (Task 7):
- The sharded `make test` (898 cmd/sdlc tests + 19 packages) is green. The one exception is `processgroup.TestCancellationKillsDescendants`, which the sandbox breaks by denying `/bin/ps`; it passes outside the sandbox.
- `make vocab-embed` is clean.
- The e2e `TestShapeUnderClaimThenStart` is the smoke test: claim leaves the card open with an owner, a shaping edit on rest is carried, start-plan starts the card keeping the claim's stamp, and a re-run is a no-op.

Two finds while running the suite:
- start-plan had dropped `requireCardOwnership`'s relocation hint for held open cards; fixed.
- side-quest: `TestClose_MilestoneRefusesWithRedirect` died under `make test`, because `Makefile.workflow` exports a cwd-relative `WF_ISSUES_DIR`. It predates #283 (a main baseline run directly through `scripts/test-shard.py` skipped make), and the test now clears the variable.

Close review round 1 (SHIP, five Minors), reopened to fix in the same round:
1. The stale `setstatus.go` comment about claim as an open→working lock is rewritten.
2. The `owned` guard on `start` is now enforced: `set-status working` on an unowned open card refuses toward claim, so starting never skips claim's details-on-main check. `--force` waives it, like every named guard; a foreign owner is still never waivable. Tests are updated to match.
3. The contention warning still counts only `working`. This is intended (plan revision 3); the shaping-claim views are #284's.
4. The relocation hint for a moved open, held card is recorded in #284's Log, since move semantics for open claims belong there.
5. Imports regrouped.

Close review round 2 (SHIP; all five round-1 findings dispositioned, one new Minor), reopened again to fix it: `startDecision` had gated on the ownership axis (`CanHoldOwner`), which admitted held `blocked` and `codecomplete` cards. A codecomplete card could have got a fresh branch from main. Admission is now a lifecycle question: the `start` edge's source or target (open or working), as before #283's working-only gate plus open. The product test asserts this.

### 2026-10-07
- 2026-10-07: closed — make test green after rebase + fixes (910 cmd/sdlc tests + 20 pkgs; processgroup /bin/ps test is sandbox-only and passes unsandboxed); BR-8 fixed (change-code contract states its own refusal, proved by TestChangeCodeRefusesUnstartedClaim); scope revision: a close survives a rebase (TestNewestCloseAfterARebase over the rewritten-binding table, TestCloseAncestorOfTreatsAnUnknownCommitAsNoAncestor on real git); this close landing on the rebased branch is the end-to-end proof
- 2026-10-07: closed — rebased onto origin/main; BR-7 fixed: recovery catalog contracts for claim/start-plan/reclaim/set-status and the scheduling example follow the lock model, with named proofs (recovery contract tests pass); make test green (909 cmd/sdlc + 20 pkgs; processgroup /bin/ps sandbox-only); fleet claims derive from CanHoldOwner; earlier evidence stands
- 2026-10-07: closed — rebased onto origin/main; make test green (909 cmd/sdlc tests + 20 pkgs; processgroup /bin/ps test is sandbox-only and passes unsandboxed); make vocab-embed clean; fleet claims now derive from CanHoldOwner (fleet unit + TestFleetInventoryPlacesClaims), scheduling example expects open after claim (TestSchedulingExampleRuns); earlier evidence stands: TestShapeUnderClaimThenStart, TestClaimNeverMovesStatus, TestStartDecision over status x owner

Rebased onto origin/main (114 commits). Conflicts:
- `workshop/lessons.md`: both sides appended; kept both.
- `startplan.go`: imports; merged.

A conflicted commit's `#283:` subject was dropped as a comment line; restored from the pre-rebase tag.

Code from main that treated an active status as the claim now derives from the ownership axis:
- #288's fleet claims (`internal/fleet/claims.go` `mine`/`validate`), help text included;
- #280's scheduling example, which expects `open` after claim.

Five tests from main were updated to the new contract. `make test` is green (909 + 20 packages; processgroup is sandbox-only). Reopened to re-close, so a fresh review covers the rebased window.

Close after the rebase:
- Round 4 (FIX-THEN-SHIP, BR-7): the recovery catalog still described the old claim and start-plan; fixed.
- Round 5 (FIX-THEN-SHIP, BR-8): my BR-7 edit's `str.replace` copied start-plan's proof rows into change-code's entry too; fixed (`6636b1a1`).

Blocked landing: `sdlc issue recovery reconcile` releases every new receipt with "the card records a newer close than this receipt (close-a09782625336 reviewed d8f753db)". The cause is `tracker/completeop.go` `newestClose`, which judges the card's prior close superseded only if its reviewed HEAD is an ancestor of the new one. The rebase rewrote history, so the pre-rebase reviewed HEAD is no ancestor of anything, and the stale binding wins forever. Reopen (`set-status working`) does not clear the binding. This is a general gap from #252 M3: any rebase after a close permanently blocks re-closing. The operator chose to fold the fix into #283 (see Revisions).

## Revisions

### 2026-10-02 — promoted to project `claimant-ownership`; #283 narrowed to the model

Reason: four rounds of design with the operator (see Log) produced five separable pieces. #283 keeps the model; #284 (claims, handoff, takeover), #285 (transfer guard), #286 (boundary pushes, abandon) and #287 (reconcile merges done outside sdlc) carry the rest. The Spec above is the pre-design capture; the Log and the project PRD supersede it where they differ (notably: claim does not create the branch, and unclaim does not change status).

Delta, the Done when before narrowing:

> - The lifecycle model and help text state that the claimant is the owner and the lock, and that status is lifecycle only. "Owner" no longer means a branch anywhere in sdlc.
> - `sdlc claim` creates the issue branch in the claiming checkout, and refuses per #272.
> - Unclaim flips the card `working → open` and clears the claimant. It deletes a branch with nothing unlanded, publishes details-only edits before deleting, and refuses when the branch carries code.
> - The transfer guard accepts landings from the recorded owner's checkout, whatever the branch is named. A fixture reproduces pair#365 (close after the PR merged, on a renamed branch), and `sdlc pr` → `sdlc merge` then flips the card to done.
> - `claim --adopt` records an owner on a pre-#277 `codecomplete` issue.
> - A reopened issue gets a fresh issue-branch name and lands through the owner-based guard.
> - A merge done outside sdlc is detected: a fixture with a GitHub-merged issue branch and a card not `done` is reported by sdlc, and finished when the close evidence is on main.

### 2026-10-02 — Done when: two additions from the durable plan

Reason: writing the plan surfaced two consequences of splitting claim from start. Shaping under a claim on a resting branch must be able to reach `start-plan` (the friction hit in round 1), and a verb past planning must not run on an unstarted card. Delta: added the `start-plan` carry bullet and the `change-code`/`reclaim` bullet. Also decided in the plan (D1): `claim` keeps stamping `started`, so shaping time counts, per the operator's round-2 call; precise multi-claim attribution goes to #284.

### 2026-10-07 — scope: a close survives a rebase

Reason: after rebasing onto main, #283 could not land. Every close and reconcile was released against the card's pre-rebase close (`completeop.go` `newestClose` judged supersession by ancestry only). This is a general sdlc gap, since rebasing onto origin/main is routine. The operator chose a surgical fix inside #283 over a separate issue, because sdlc itself blocked landing.

Delta:
- `newestClose` also supersedes a binding the branch no longer contains, when the receipt's own review is on the branch.
- The two completion-op callers use `closeAncestorOf`, which treats an unknown reviewed commit (rewritten away and never fetched) as preceding nothing.
- The close recovery contract states the rule.
- Done when gains the bullet above.

