---
id: 000283
status: working
deps: [ariadne#277, ariadne#278]
github_issue:
created: 2026-10-02
updated: 2026-10-02
estimate_hours:
card_mirror: '203c1590f4dd363293c66e52efd6f3e409bd8479' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-02T10:51:37-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
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

- The lifecycle model and help text state that the claimant is the owner and the lock, and that status is lifecycle only. "Owner" no longer means a branch anywhere in sdlc.
- `sdlc claim` creates the issue branch in the claiming checkout, and refuses per #272.
- Unclaim flips the card `working → open` and clears the claimant. It deletes a branch with nothing unlanded, publishes details-only edits before deleting, and refuses when the branch carries code.
- The transfer guard accepts landings from the recorded owner's checkout, whatever the branch is named. A fixture reproduces pair#365 (close after the PR merged, on a renamed branch), and `sdlc pr` → `sdlc merge` then flips the card to done.
- `claim --adopt` records an owner on a pre-#277 `codecomplete` issue.
- A reopened issue gets a fresh issue-branch name and lands through the owner-based guard.
- A merge done outside sdlc is detected: a fixture with a GitHub-merged issue branch and a card not `done` is reported by sdlc, and finished when the close evidence is on main.

## Plan

- [ ] Claim, run start-plan, settle the open questions and the child-issue split before designing the implementation.

## Log

### 2026-10-02

Filed from a pair session that landed pair#365 by hand after `sdlc pr`, `reclaim` and `claim --adopt` all refused. Root cause traced to `transferguard.go` deriving the owner from the details basename. Operator review the same day: the claimant should be the owner and the lock (status-as-lock predates the claimant), and the scope is a rethink of the verbs and guards around the claimant. The issue is likely to split. Details left uncommitted on this resting branch, not moved.

Design discussion (claimed in ariadne:2). Decided with the operator:
- Unclaim is allowed from `blocked` as well as `working`; not from `codecomplete`.
- Stale locks: explicit `reclaim` with a reason, plus `sdlc state` showing claim age. No leases or heartbeats.
- Transfer guard condition: a details change lands only from the owner's checkout AND from a branch whose merge-base contains the latest published version (the handoff `MainCommit`). The content condition is what stops stale overwrites; branch names stop mattering.
- **Claim does not create the branch; `start-plan` does.** Pre-implementation editing (follow-ups filed from a console, one brainstorm revising several related issues) must not need one branch per issue. A slot may hold several claims at once: claim the set, edit their details, publish narrowly to main (published can't be unpublished), unclaim.
- Claimant = slot on a machine (the agent's durable workspace), operator = supervising human. This divergence from Jira-style person-assignment is intentional; an atlas terminology entry is still to be written.
