---
id: 000483
status: working
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-20
estimate_hours:
collaboration-mode: collaboration
card_mirror: '0da1c1f30094177d4b4809b09580756754ee6eb3' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-20T04:06:52-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
---

# Continuing an issue whose branch already landed: refuse with 'close it and file a follow-up' (pair#622 case)

## Problem

pair#622 (10-19): the issue's code landed (PR 406), but its live acceptance stayed open and needed a little more
work. `sdlc change-code --issue 622` insists on the canonical branch, which already merged; `--name` and `--issue`
are mutually exclusive; so there is no supported way to continue an issue whose branch has landed. The worker
improvised a fresh branch outside the gates.

## Spec

Pick one, simplest first, and make the refusal say it:
1. **Refuse with the recovery:** "#N's branch already landed; close #N with what's verified and file a follow-up
   issue for the rest (`sdlc issue new … --deps N`)." This keeps one issue, one branch.
2. Only if (1) proves too rigid in practice: a supported continuation branch (`<canonical>-2`) that the gates treat
   as the same issue.

## Done when

- `change-code` and `start-plan` on an issue whose canonical branch is merged refuse with that next action
  (test), and `sdlc issue new --deps` is shown as the path.

## Plan

Design: Spec option (1), plus a matching close admission. No new state; one predicate, two refusals, one relaxed admission.

**Landed** = the issue's canonical branch (`issue.BranchName(card)`) has a MERGED PR (`ghClient.PRMergedForBranch`, the existing seam with fakes), and the local canonical branch is absent or wholly on main. This catches squash merges, which `externalMerges` misses. The second clause keeps a mid-issue landing with new local commits on the old path. A gh query failure reads as "not landed", which keeps today's refusals (the safe direction).

1. **start-plan / change-code refuse a landed issue** with the next action: "#N's branch B already landed. Close #N with what's verified: cut a follow-up branch from main (`git switch -c NNNNNN-<suffix> origin/main`), run `sdlc close --issue N`, then `sdlc pr` and `sdlc merge`. File the rest as a follow-up: `sdlc issue new '<title>' --deps N`." Today start-plan silently re-creates B at main, and change-code points back at start-plan. This keeps one issue, one branch.
2. **close admits a follow-up branch for a landed issue.** In `requireCloseIssueBranch`, a landed issue also admits a branch whose name has the issue's id prefix (`NNNNNN-…`, the same prefix rule as `fleet.AssociateBranchIssue`). That branch must carry no other issue's unlanded commits (reuse start-plan's unlanded-base check, ARCH-DRY). The evidence commit lands through that branch's own PR, and `settleLandedCompletions` marks the card done as it does today. From rest or main, the refusal names the same recovery as (1).

For pair#622: from `000622-live-acceptance` (PR pair#474, carrying the attestation), `sdlc close --issue 622` then `sdlc merge` closes it with that evidence. Nothing in #622 needs to change.

ARCH: MOCK, the landed check reads only the existing `ghClient` and git seams, with fakes in the tests. ORDER, this holds no state between events: each verb re-derives "landed" from PR and git facts at call time. A PR that merges between two verbs is just a later observation. SECURE, a branch name with only the issue prefix is not proof on its own: the unlanded-base check stops another issue's work from riding in. EVOLVE/FUNERAL, N/A: no persisted format and nothing durable. CONSTRAINTS, one `gh pr list` per refusal check, only on the paths that would otherwise refuse or (for close) admit.

Steps:
- [ ] Pure decision `landedIssue(mergedPR bool, localTip, onMain)` plus the IO wrapper; unit table.
- [ ] start-plan and change-code refusals with the recovery text. Tests: landed→refuse naming `issue new --deps`; merged PR with unlanded local commits→old path; gh error→old path.
- [ ] close admission. Tests: landed plus prefixed follow-up branch from main→admitted; prefixed branch carrying another issue's commit→refused; non-prefixed branch→refused with the recovery; not landed→today's refusal unchanged.
- [ ] atlas: sdlc lifecycle note on closing a landed issue.

## Log

### 2026-10-19

## Log

- 10-19 23:25 (TL): the gap blocks **close** too, not just change-code: pair#622's `sdlc close` refuses ("run from
  canonical 000622-…") from the follow-up branch `000622-live-acceptance` that holds the operator's attestation (PR
  pair#474), because the canonical branch landed as PR 406. #622 is parked on this issue. Whatever (1)/(2) decides
  must also let such an issue close (e.g. close with evidence on main, or the refusal names `sdlc close --no-pr`-style
  recovery if one exists).
