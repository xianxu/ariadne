---
id: 000284
status: working
deps: [ariadne#283]
github_issue:
created: 2026-10-02
updated: 2026-10-07
estimate_hours:
card_mirror: 'ca6f62abe87b99b35c71c34f7a92cbbc3f8d11f9' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-07T13:45:08-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
---

# Claims: multi-claim, publish, handoff, takeover

## Problem

Part of project `claimant-ownership` (see its PRD; design rationale in #283's Log). Once #283 makes the claimant the lock, claims must support shaping several issues before implementation and handing started work between slots, machines and operators.

## Spec

- **Atomic multi-claim:** `sdlc claim --issue a,b,c` compare-and-swaps all the cards in one tracker commit. On a race it re-reads, re-checks every card and retries; it is all-or-nothing.
- **Refresh on claim:** claiming refreshes local details from main, so the lock never protects a stale copy.
- **`issue publish --issue a,b,c`:** narrowly republishes owned details edits to main from any branch. `move-detail` stays the first publication. On an issue branch the identical content is committed there too, so the later merge is a no-op on those files; a resting branch fast-forwards.
- **`issue sync` re-scoped:** commits the details of issues you hold a claim on, locally, on any branch (resting branches included). Its `--push` is retired in favour of `issue publish`.
- **Unclaim:** clears the owner; status is unchanged. For an unstarted (`open`) claim, unpublished details edits are published automatically. For started work it is a handoff: refuse on a dirty tree, write a handoff note in `## Log`, push the issue branch, record `{branch, head}` on the card.
- **Takeover claim:** claiming an active, unowned card fetches the branch, checks origin's head equals the recorded head, and checks it out in a clean slot on a resting branch. This replaces `claim --adopt` (pre-#277 cards are active and unowned).
- **`move` vs. `reclaim`:** `move` stays the same-machine local relocation. `reclaim` (with a reason) stays the forced takeover and gets the last pushed state.
- **`sdlc state`:** shows both views, by slot and by operator, with claim age.

## Done when

- **M1:**
  - Claiming three issues is one tracker commit.
  - A concurrent claim of one of them makes the whole set fail cleanly, with no partial claims.
  - A race on an unrelated card retries and lands.
- **M2:**
  - `sdlc issue publish` covers first publication (today's `move-detail`, which remains as an alias) and republishing an owned issue's edits.
  - A claim → edit → publish → unclaim cycle on a resting branch lands the details on main and leaves the branch identical to main.
  - Republishing over a main copy that changed since the local base refuses.
  - Unclaiming an `open` issue publishes its unpublished edits first.
  - In a tracker repository, `issue sync` points at `git commit` / `issue publish`.
  - AGENTS.base §2 states the resting-branch rule.
- **M3:**
  - Unclaiming a `working` issue refuses a dirty tree, pushes the branch, and records `{branch, head}` on the card; the status is unchanged.
  - A claim from another workspace (fixture: a second machine) checks the branch out at that head.
  - A mismatched head refuses with the next action.
  - A pre-#277 `codecomplete` card is taken over by a plain claim.
  - An older envelope rewrite keeps unknown internal keys.
  - Claim repairs a moved `open` claim.
- **M4:** `sdlc state` shows each issue's owner and claim age (from tracker history), with views by slot and by operator.

## Plan

- [x] Claim, run start-plan, and design against the Spec and the project PRD (`workshop/projects/claimant-ownership.md`); size the flow at change-code.
- [ ] M1 — atomic multi-claim and refresh on claim
- [ ] M2 — `issue publish` (first publication + republish), unclaim of an open claim, `issue sync` retirement, resting-branch rule
- [ ] M3 — handoff and takeover, envelope keeps unknown keys, moved-open repair
- [ ] M4 — `sdlc state` owner, claim age, by-slot and by-operator views

## Log

### 2026-10-02

From #283's close review: `requireCardOwnership`'s "moved here, finish with `sdlc claim`" hint (`claimant.go:162`) assumes claim finishes relocations, but `claim.go` does so only for `move`'s statuses (active). An open, held card that `sdlc move` relocated (start-plan's card write lost, then move) gets a refusal instead. Decide here whether `move` applies to open claims; then align claim's repair and the hint.

Also from #283: the model's `unclaim` lists every holdable status, `codecomplete` included, following round 2's handoff decision (which supersedes round 1's "not from codecomplete"). Confirm that when building the verb.

### 2026-10-07

Design with the operator, after a code digest:
- `tracker` has no multi-card CAS; build it on `gitx.UpdateMany`, which re-decides on every attempt.
- Only `sdlc pr` pushes an issue branch, and nothing fetches one.
- `issue sync` in a tracker repository is just `git commit -- <details>` on the issue branch (no ownership check, `--push` retired).
- Every SDLC repository in the workspace is a tracker repository. The legacy ones are the brains plus two empty scaffolds, which the operator had removed.

Decisions:
- Claim age is derived from tracker history.
- Unclaim is a normal operation: the note is optional (`--note`). The only refusal is a dirty tree for started work.
- `issue publish` replaces `move-detail` (kept as an alias): first publication needs no owner, and republishing needs the claim.
- `issue sync` is not extended; in tracker repositories it points at `git commit` / `issue publish`.
- AGENTS.base §2 gains the resting-branch rule.
- Rollout check: older binaries ignore unknown internal envelope keys on read. However, `updateEnvelope` rewrites only known keys, so an old binary's close or handoff write would drop a new handoff record. M3 makes the envelope keep unknown keys. The operator accepts "use the new binary" as the rollout step.

## Revisions

### 2026-10-07 — design after the code digest

Reason: the design session above.

Delta against the Spec:
- `issue sync` is no longer extended to any branch (it is retired to a pointer in tracker repositories).
- `issue publish` absorbs `move-detail`'s first publication.
- The unclaim note is optional, not required.
- Claim age comes from tracker history.
- Milestones M1–M4 replace the single Plan step.
- Done when is restated per milestone.

The previous Done when:

> - Claiming three issues is one tracker commit; a concurrent claim of one of them makes the whole set fail cleanly, with no partial claims.
> - A claim, edit, `issue publish`, unclaim cycle on a resting branch lands the details on main, and nothing is left behind.
> - Unclaim of a `working` issue pushes the branch and records `{branch, head}`; a claim from another worktree (fixture: a second machine) resumes at that head. A mismatched head refuses with the next action.
> - A pre-#277 `codecomplete` card is taken over by a plain claim.
> - `sdlc state` shows the slot and operator views with claim age.

