---
id: 000471
status: working
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'af347e54e50e5d70a9590e3a1686455254fef604' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-19T13:17:35-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "f4858913", done: "e6005be1"}
---

# abandon: --issue N acted on the current (other issue's) branch and deleted it on origin

## Problem

2026-10-19, pair:1 (pair repo): `sdlc abandon --issue 694 --as punt --reason …` was run from the checkout of a *different*
issue's branch (#695). It acted on the current branch instead of #694's:
- committed #694's abandon log on top of #695's tip;
- saved that commit as `refs/ariadne/abandoned/000694` on origin;
- **deleted the #695 branch locally and on origin**;
- left #694's real branch untouched.

pair:1 repaired it by hand (recreated #695 at its tip, force-with-lease'd the archive ref to #694's real tip 45baebbbd, deleted
#694's branch). Nothing was lost only because the slot noticed. An abandon from the wrong checkout can delete another issue's
unmerged work on origin.

## Spec

`sdlc abandon --issue N` operates on issue N's branch, never on "the current branch" by assumption:
- If the checkout is on N's branch: today's behavior.
- If the checkout is on another issue's branch (or any branch that isn't N's): refuse, naming both, unless N's branch can be
  acted on without touching the checkout (archive its tip from the ref, delete it); prefer refusing in the first version.
- If N has no branch (planned but never started): abandon the card only.
- Never delete a branch whose tip isn't the archived commit. Check this before any push or delete (compare-and-swap on the ref).

Add the same guard to any other verb that takes `--issue N` but acts on HEAD's branch (audit and list them in the plan).

## Done when

- Fixture: on #B's branch, `abandon --issue A` refuses and changes nothing (local and origin refs unchanged).
- Fixture: on A's branch, today's behavior is unchanged.
- The audit of `--issue N` verbs that act on HEAD is in the plan, each either guarded or shown safe.

## Plan

- [x] Write and approve the durable plan at `workshop/plans/000471-abandon-issue-n-acted-on-the-current-other-issue-s-branch-and-deleted-it-on-origin-plan.md`, including the `--issue`/HEAD audit.
- [x] Add red real-Git regressions for wrong-branch abandon, close, milestone-close and attach; preserve open-issue and interrupted-abandon recovery cases.
- [x] Bind started work to the card-derived issue branch before any effect. For remote deletion, make the remote branch tip equal the archived commit with leased updates, then delete it with a lease.
- [x] Run targeted tests, `make -k build`, and the recorded full suite on the final pushed head.
- [ ] Complete the close review and landing gates.

## Log

### 2026-10-19

- Claimed and started in ariadne:1. The incident is a branch identity failure: `abandon` accepts any non-rest branch as the named issue's branch and stores that name in its terminal record. Planning the shared guard and remote-ref checks before code.
- The wrong-branch abandon, close, milestone-close and attach fixtures failed against the old behavior before the guards were added. The real-Git fixtures now check that an A-on-B command leaves both local and origin refs, the archive ref, cards and main unchanged. Terminal retries reject another branch or a persisted branch name from another issue; an open issue remains card-only.
- Remote retirement now proves the archive ref holds the recorded head, advances an older remote branch to that head with a lease, then deletes with an exact-head lease. A pre-receive hook checks the deleted old tip equals the archive ref; a divergent remote tip refuses before the note or archive push and names both tips.
- Focused abandon/close/attach group passed (`go test ./cmd/sdlc -run '^(TestAbandon|TestTrackerCloseRefusesAnotherIssuesBranch|TestMilestoneCloseRefusesAnotherIssuesBranch|TestIssueAttach)' -count=1 -timeout 12m`). `make -k build`, `make sdlc-build`, the recovery package test, and `git diff --check` passed. The full recorded suite and close review remain.
- PR #328's first full suite exposed the deliberate #285/#365 recovery: a renamed close branch at the canonical issue tip must still close. The hub approved the narrower proof: exact HEAD equality with the local canonical branch tip, or a fresh remote tip if local is absent; another head refuses. Abandon and attach still require the canonical name. Focused recovery and different-head tests pass. The same suite had the known `dev-aliases.test.sh` #465 flake; its fix is landing separately.
- Live smoke used branch binary `99fe1ce6770d1c26f7c826cd6b86264cf18fe860` in a disposable local clone with a separate bare origin. From #475's branch, `abandon --issue 474` named both branches and refused; every local and bare-origin branch/archive ref was byte-for-byte unchanged. The hub denied ariadne:0 staging because that binary serves the fleet, so no shared binary was replaced.
- The final pushed code head `d2536fa3bf909382af7cb032cfd715c18d905ea0` passed the recorded full suite (`sdlc-local/suite`, 12 shards plus 71 other packages, no failures, log SHA-256 prefix `c7c70f8e0e89`). An earlier run on this same head was interrupted by the hub's operator-requested RAM-disk reformat and recorded no result; the unchanged rerun passed.
- Close review round 1 returned REWORK: a post-checkout hook can move the local branch between the early tip check and unconditional `git branch -D`, deleting an unarchived tip; README also omits the new checkout requirements. The revised plan pins local deletion to the recorded tip atomically and adds the reviewer’s race regression before the fix.
- The post-checkout-hook regression failed against the old deletion (`abandon` returned success despite the moved ref), then passed with `git update-ref -d` conditioned on the recorded archived head. The broader abandon/close/attach group passed (149.390s); README now names the checkout rules. The other durable-landing local branch deletion in `landing.go` already uses expected-tip `update-ref -d`; the legacy non-durable `merge.go` cleanup still uses `branch -D` after a merged PR, a separate merge-cleanup lifecycle outside this named-issue abandon change.
- A second red real-Git fixture showed that `update-ref -d` alone deletes a branch checked out in another linked worktree. Reusing durable landing's occupancy check before the CAS preserves the old refusal; both the occupied-branch and moved-tip fixtures pass together.
- The final code head `38abeb176619d0cb5914a2184a1659eb281b9cf2` passed `make -k build`, `make sdlc-build`, the focused real-Git group (213.191s), `git diff --check` and the recorded full suite (`sdlc-local/suite`, 12 shards over 1,726 `cmd/sdlc` tests plus 71 packages, no failures, log SHA-256 prefix `2115980134ac`).
