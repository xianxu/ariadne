---
type: project
name: "claimant-ownership"
goal: "Re-derive sdlc ownership from the claimant: the slot that holds a claim is the owner and the lock, and status is lifecycle only."
done_when: "Claim/unclaim/takeover/abandon run on the claimant lock with status as lifecycle only; the transfer guard decides by owner + based-on-latest; a pair#365-shaped landing and a merge done outside sdlc both finish through sdlc without hand edits."
status: defined
operator: Xian Xu
mvp_scope: [ariadne#283, ariadne#284, ariadne#285, ariadne#286, ariadne#287]
created: 2026-10-02
updated: 2026-10-02
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

- [ ] Model: claimant lock, status as lifecycle, slot/operator terminology [ariadne#283]
- [ ] Claims: multi-claim, publish, handoff, takeover [ariadne#284]
- [ ] Transfer guard: owner plus based-on-latest [ariadne#285]
- [ ] Boundary pushes and sdlc abandon [ariadne#286]
- [ ] Reconcile merges done outside sdlc [ariadne#287]

## Log

### 2026-10-02

Promoted from #283 after a four-round design discussion with the operator (decisions in #283's Log).
