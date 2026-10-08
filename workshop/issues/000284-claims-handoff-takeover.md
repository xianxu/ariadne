---
id: 000284
status: working
deps: [ariadne#283]
github_issue:
created: 2026-10-02
updated: 2026-10-07
estimate_hours: 6.1
card_mirror: 'f8ee27d3565872056e3a13966668f121b72e3643' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-07T13:45:08-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
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

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec             design=1.0 impl=0.08
item: smaller-go-module      design=0.2 impl=0.2
item: smaller-go-module      design=0.1 impl=0.2
item: smaller-go-module      design=0.05 impl=0.12
item: milestone-review       design=0.0 impl=0.14
item: greenfield-go-module   design=0.3 impl=0.3
item: smaller-go-module      design=0.1 impl=0.2
item: atlas-docs             design=0.1 impl=0.08
item: milestone-review       design=0.0 impl=0.14
item: smaller-go-module      design=0.05 impl=0.12
item: greenfield-go-module   design=0.3 impl=0.3
item: greenfield-go-module   design=0.3 impl=0.3
item: atlas-docs             design=0.05 impl=0.08
item: milestone-review       design=0.0 impl=0.14
item: smaller-go-module      design=0.1 impl=0.2
item: smaller-go-module      design=0.1 impl=0.2
item: milestone-review       design=0.0 impl=0.14
design-buffer: 0.15
total: 6.1
```

Items, in plan order:
- the design (digest, three rounds with the operator, plan);
- M1: `ChangeCards`, the claim set, refresh, its review;
- M2: `issue publish` (greenfield: two paths plus the base check), open unclaim, the `issue sync`/AGENTS.base text, its review;
- M3: the envelope, the handoff unclaim, the takeover claim, docs, its review;
- M4: claim times, state views, the close review.

The design buffer is +15% because a thorough plan doc exists.

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only.*

## Plan

- [x] Claim, run start-plan, and design against the Spec and the project PRD (`workshop/projects/claimant-ownership.md`); size the flow at change-code.
- [x] M1 — atomic multi-claim and refresh on claim
- [x] M2 — `issue publish` (first publication + republish), unclaim of an open claim, `issue sync` retirement, resting-branch rule
- [ ] M3 — handoff and takeover, envelope keeps unknown keys, moved-open repair
- [ ] M4 — `sdlc state` owner, claim age, by-slot and by-operator views

## Log

### 2026-10-02

From #283's close review: `requireCardOwnership`'s "moved here, finish with `sdlc claim`" hint (`claimant.go:162`) assumes claim finishes relocations, but `claim.go` does so only for `move`'s statuses (active). An open, held card that `sdlc move` relocated (start-plan's card write lost, then move) gets a refusal instead. Decide here whether `move` applies to open claims; then align claim's repair and the hint.

Also from #283: the model's `unclaim` lists every holdable status, `codecomplete` included, following round 2's handoff decision (which supersedes round 1's "not from codecomplete"). Confirm that when building the verb.

### 2026-10-07
- 2026-10-07: closed M2 — make test green (936 cmd/sdlc + 20 pkgs; processgroup /bin/ps sandbox-only). Round-4 fixes tested from their refused states: BR-8 TestPublishMovedMainFromAnotherBranchRemedy (switch to rest carries the edit, rerun publishes; dry-run on rest names the bring-in) + TestPublishMovedMainOnTheIssueBranchRemedy (remote from the publication target); BR-14 TestPublishBringInSetAsideIsAllOrNothing (fails on the pre-fix code: verified by mutation) + TestPublishBringInWriteBackFailureNamesTheCopies, copies removed once safe; BR-18 catalog Effects/Repeat swept by grep across helptext/atlas/catalog; minors: detailsAt (no probe without its error), issue.Stem/BranchName at every site. Earlier M2 evidence stands; review verdict: SHIP
- 2026-10-07: closed M1 — make test green (915 cmd/sdlc + 20 pkgs; processgroup /bin/ps sandbox-only); ChangeCards real-git tests (one commit, refusals, re-decide on named and unrelated races); claim set: TestClaimSetDecision over status x owner, TestClaimSetIsOneTrackerCommit, TestClaimSetRace (contested member fails whole, benign change re-decided); TestClaimRefreshesTheCheckout (rest ff, blocked ff warns, branch body-diff warns); catalog + help + atlas updated; review verdict: SHIP

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

M1 implemented:
- `tracker.ChangeCards` (one commit for N cards, re-deciding per attempt): `a41d5ab7`;
- the claim set via the `cardsPublish` seam and pure `claimSetDecision`: `f50523ea`;
- refresh after claim: `3656c010`.

Test-fixture slips caught while writing the tests (not code bugs):
- a decision oracle passed `"now"` as a timestamp;
- a peer "retitle" set `title`, which lives in the H1, not the frontmatter.

Design point surfaced: comparing local and main details must compare bodies only, since claim's mirror refresh rewrites the frontmatter by design.

M1 close review (SHIP, six advisory Minors), all fixed in the same round:
1. `refreshAfterClaim` surfaces rev-parse and merge-base errors instead of swallowing or misnaming them.
2. One `issue.JoinRefs` for claim's refs and rerun hint and the tracker subject. `Snapshot.validateReplacements` holds the multi-card budget check, and `validateReplacement` is its one-element case.
3. `ChangeCards` collapses duplicate IDs itself (test added).
4. An unowned started member of a set is refused toward claiming it alone with `--adopt` (sentinel `errUnownedStarted`); test added.
5. The milestone lines are nested under #284 in the project file. On the hand-recorded actual: it was the measured 1.15h, and the close rewrote the block.
6. Test of a claim set through a lost response and a rerun.

M3 review round 2 (one Important, three Minors), all fixed:
- BR-27: the takeover rerun's diverged-copy refusal is tested, and a local copy *ahead* of the remote (the owner's unpushed work) is resumed on instead of refused.
- Refusals name what was observed (ahead vs diverged; only a stale lease reads as a lease miss; a missing remote branch vs a failed fetch).
- One `gitx.RemoteTrackingRef`, swept to every site.
- The note dedupe identifies the attempt: same-day text, or a handoff's own note commit in the unpushed tail. Known residual: an *open* unclaim whose card write fails, rerun after midnight, files its note again.
- Hours: the measurement engine undercounted M3's first close (0.07h) and corrected itself by the re-close (4.18h total); recorded as measured.

