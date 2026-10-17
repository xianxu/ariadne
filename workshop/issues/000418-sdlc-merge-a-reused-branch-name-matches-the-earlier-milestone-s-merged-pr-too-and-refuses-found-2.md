---
id: 000418
status: working
deps: []
github_issue:
created: 2026-10-16
updated: 2026-10-16
estimate_hours:
card_mirror: 'fdde79c8b471882b44e5ad379f993acbd78f7cab' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-16T20:02:46-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: operator, spec: "1951640a", done: "575d4647"}
---

# sdlc merge: a reused branch name matches the earlier milestone's MERGED PR too and refuses ('found 2')

## Problem

## Spec

Reported by pair:2 (pair#618 M2), 2026-10-16. `sdlc merge` refuses with: "need one exact matching PR for 000618-slots-...; found 2".

How it happened:
1. M1 landed from the issue branch as PR #380, now MERGED. `sdlc merge` then deleted the branch.
2. `start-plan` recreated the same branch name for M2, and M2's PR #387 is OPEN at 9917e65f.
3. PR matching by branch name finds both PRs and refuses.

Every multi-milestone issue that lands milestone by milestone hits this.

Rule: match only PRs whose head is this branch's current head, or that are OPEN. A MERGED or CLOSED PR whose head commit is not the branch head is history, not a candidate. Two OPEN PRs for one branch still refuse.

## Done when

- With a fake GitHub holding one MERGED PR (an older head) and one OPEN PR (the current head) for the same branch, `sdlc merge` picks the OPEN one; a test covers it.
- Two OPEN PRs for the branch still refuse, naming both.
- pair#618 M2's PR #387 lands through `sdlc merge` with no manual gh merge.

## Plan

- [ ] Add fake-GitHub regressions through the real landing flow for completed older heads beside the open PR and ambiguous open PRs; confirm red.
- [ ] Reuse one head-aware candidate rule in strict selection, interrupted integration and head-registration polling (ARCH-DRY); preserve identity validation and stale-open push guidance.
- [ ] Run targeted tests, stage the binary for pair#618 live verification through the hub, record the full suite, then close and land.

## Log

### 2026-10-16

### 2026-10-16 — bounded diagnosis and plan

selectIntegrationPR's recovery fallback and syncIntegrationPR call selectLandingPR with an empty head, so every old merged PR becomes live again. Strict selection also permits a matching open PR beside another stale open one. Use one pure candidate rule across those callers: preserve every OPEN candidate (even if stale), ignore completed mismatched heads, and refuse multiple candidates with their PR numbers. Strict selection retains its existing named push remedy for one stale open PR. CLOSED PRs remain ineligible under the existing policy. No new external seam, state or lock: selection operates on the immutable GitHub observation, while the existing landing line and expected-head merge protect effects (ARCH-PURE, ARCH-DRY). The absence of competing PRs cannot be a local lock proof because GitHub owns them; bounded structured evidence and the existing expected-head operation remain authoritative.

The existing landingFakeGH and real Git/bare-origin fixture exercise production merge, integration and cleanup. Tests own the scenario matrix. Post the small plan to the hub and self-approve under tl-worker quick-flow rules.
