---
type: project
name: "claimant-ownership"
goal: "Re-derive sdlc ownership from the claimant: the slot that holds a claim is the owner and the lock, and status is lifecycle only."
done_when: "Claim/unclaim/takeover/abandon run on the claimant lock with status as lifecycle only; the transfer guard decides by owner + based-on-latest; a pair#365-shaped landing and a merge done outside sdlc both finish through sdlc without hand edits."
status: defined
operator: Xian Xu
mvp_scope: [ariadne#283, ariadne#284, ariadne#285, ariadne#286, ariadne#287, ariadne#301]
created: 2026-10-02
updated: 2026-10-07
sources: [workshop/issues/000283-claim-owns-issue-branch.md]
---

# claimant-ownership

#277 recorded a claimant on every card, but sdlc still reasons about ownership through status (`open` = claimable) and branch names (the transfer guard's "owner"). This project makes the claimant the single owner and lock, ahead of higher concurrency: several slots per operator, and several operators on several machines. **Not in MVP:** leases or heartbeats for stale claims (explicit `reclaim` covers that), and a required GitHub check against merges done outside sdlc (we detect and reconcile instead).

## PRD

**Who owns work.** A claim belongs to a *slot*: a workspace on a machine (repository + machine + worktree), the durable home of one agent's thread of work. It outlives any session: compaction, restarts, a different CLI. The *operator* on the claim is the supervising human, the escalation and grounding point, and is descriptive, not a lock. This intentionally departs from Jira-style person assignment, which combines executor and supervisor in one assignee.

**Model.** The owner is the lock; status is lifecycle only. Claim and unclaim change the owner, never the status. `open` with an owner is legal ("being shaped"). An active issue with no owner is available for takeover. `start-plan` starts the lifecycle: it sets `working`, creates the issue branch from main, and pushes it.

**Shaping before implementation.** A slot may hold several claims at once, claimed atomically in one tracker compare-and-swap commit. It edits their details on any branch, republishes them narrowly to main (`issue publish --issue a,b,c`; published can't be unpublished), and unclaims. A claim refreshes local details from main. `issue sync` commits claimed details locally on any branch.

**Handoff.** Unclaiming started work is a handoff:
- Unclaim refuses on a dirty tree, writes a note in `## Log`, pushes the branch, and records `{branch, head}` on the card. Status is unchanged.
- A takeover claim fetches the branch, checks the head and checks it out in a clean slot.
- `reclaim` (with a reason) is the forced path.
- `move` is the same-machine shortcut, with no push.

**Pushing.** The issue branch is pushed at every sdlc boundary. The lock plus no-stacking (#272) make the owner the only writer, so force-with-lease keeps rebasing safe. `merge` deletes the remote branch. `abandon` keeps the work under an archive ref.

**Transfer guard.** It protects only published details files. A landing may change one only from the owner's checkout and only when based on main's latest version; otherwise the merged result must equal main's version. sdlc restores main's version mechanically; semantic merges go to the owner's LLM.

**Escape hatch.** Merges done outside sdlc are detected (`sdlc state`) and reconciled.

## Estimate

Estimated per child at its `change-code`.

## Breakdown

Order: model first, done properly, with no fast-track for pair#365. #287 is independent and can run in parallel once its details are on main.

- [x] Model: claimant lock, status as lifecycle, slot/operator terminology [ariadne#283]
- [x] Pin the rebase-aware close rule: e2e test and atlas [ariadne#301]
- [ ] Claims: multi-claim, publish, handoff, takeover [ariadne#284]
  - [x] Atomic multi-claim and refresh on claim [ariadne#284 M1]
  - [ ] `issue publish`, open unclaim, `issue sync` retirement [ariadne#284 M2]
  - [ ] Handoff and takeover [ariadne#284 M3]
  - [ ] `sdlc state` owner, claim age, views [ariadne#284 M4]
- [ ] Transfer guard: owner plus based-on-latest [ariadne#285]
- [ ] Boundary pushes and sdlc abandon [ariadne#286]
- [ ] Reconcile merges done outside sdlc [ariadne#287]

<a id="ariadne-284-m1"></a>
### ariadne#284 M1 — Atomic multi-claim and refresh on claim

**est:** ~1.1h (M1's share of #284's 6.1h)
**actual:** 1.15h
**closed:** 2026-10-07

What shipped:
- `tracker.ChangeCards` writes N cards in one tracker commit and re-decides every card over the bytes each attempt reads; the single-card `UpdateCard` could only refuse a changed card.
- `claim --issue a,b,c` uses it through a `cardsPublish` seam and a pure `claimSetDecision`. A peer claiming a member fails the set whole; a benign change to a member is re-decided and the set lands.
- After the card lands, claim fast-forwards a resting branch to main, or names each issue whose details body differs from main's.

Decision worth keeping: compare details *bodies*, never whole files, because claim's mirror refresh rewrites the frontmatter by design.

<a id="ariadne-284-m2"></a>
### ariadne#284 M2 — `issue publish`, open unclaim, `issue sync` retirement

**est:** ~1.3h (M2's share of #284's 6.1h)
**closed:** 2026-10-07
**actual:** 0.47h

What shipped:
- `issue publish --issue a,b` replaces both of the old publication paths: first publication (`move-detail`, no owner needed) and the owner's later edits.
  - A republish goes in one narrow main commit, judged on details bodies against the checkout's merge base, so it never overwrites a main copy that moved.
  - Ownership is re-checked before the push.
  - Afterwards a resting branch fast-forwards and an issue branch commits the same bytes.
- `unclaim` releases open claims after publishing their edits.
- The `release` record (who let go) and an envelope that keeps unknown keys were pulled forward from M3, because unclaim's rerun depends on them.
- The audit retired `issue sync` in tracker repositories: it was only `git commit -- <details>`. AGENTS.base gained the resting-branch rule.

<a id="ariadne-284-m3"></a>
### ariadne#284 M3 — Handoff and takeover

**est:** ~1.5h (M3's share of #284's 6.1h, after Task 8 moved into M2)
**closed:** 2026-10-07
**actual:** 2.75h

What shipped:
- Unclaiming started work is a handoff: from the issue's branch with a clean tree, it commits the note, pushes the branch with a lease, and records the release's branch and tip.
- A plain claim elsewhere takes it over. Every check runs before the card write: branch name, clean resting branch, fetched tip equal to the recorded head, no diverged local copy. Then it checks out exactly that tip.
- `--adopt` folds into claim, since unowned started work, released or claimed before #277, is just taken over.
- In the model, `move` covers open claims, which closed #283's moved-open repair.

Decision worth keeping: the handoff returns the releasing checkout to rest, because a branch checked out in one worktree can't be checked out in another on the same machine.

## Log

### 2026-10-02

Promoted from #283 after a four-round design discussion with the operator (decisions in #283's Log).

### 2026-10-07 — scope: add #301

#283 landed (PR #161), but landing it needed a scope revision: a close now survives a rebase (`completeop.go` `newestClose` treats a binding that the rebase rewrote off the branch as replaced). Its final review left two advisory Minors: no end-to-end test pins the close/reconcile wiring, and the atlas lacks the rule. Filed as #301 and added to scope at the operator's request. It's next in line, ahead of #284.

[ariadne#284 M1]: #ariadne-284-m1
[ariadne#284 M2]: #ariadne-284-m2
[ariadne#284 M3]: #ariadne-284-m3
