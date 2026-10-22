---
id: 000543
status: working
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-21
estimate_hours: 1.83
collaboration-mode: collaboration
card_mirror: '07f774ecb2c3167c468e8aa3107b09ac5a64fbb0' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-21T22:43:36-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
---

# new repositories start on the legacy issue path: initial setup should create the issue tracker (non-brain repos)

## Problem

## Spec

The operator asked (via ops-echo:0, 10-21) why echo, created 10-18, weeks after the issue tracker landed (#252),
started as a legacy repository. Confirmed: new-repository setup never creates the tracker. `construct/base.manifest`
scaffolds `workshop/issues` (line 182) but nothing writes `workshop/issue-tracker.json` or the `issue-tracker`
branch, so every new repo starts on the retiring legacy path and only reaches the tracker by a later
`sdlc issue migrate`. Inventory of the workspace on 10-21: ducks (09-27) and ops (10-09) were migrated by hand (ducks
the same day; ops on 10-20); echo (10-18) was migrated tonight after three legacy-path surprises (stray sync commits,
start-plan not cutting a branch); **ops-echo (10-19) and emma-private (10-05) are still legacy**. The brains
(brain, brain-family, brain-private) are legacy by design and must stay so.

Fix: initial setup of a new, non-brain repository creates the tracker (the empty-repo case of `sdlc issue migrate`,
or an equivalent `init` step), so new repos start on the tracker. Brain repos (`.brain/config.md`) and repos that
opt out are left alone. Existing legacy repos get a one-line hint from `sdlc issue new` and `start-plan` pointing at
`sdlc issue migrate`.

## Done when

- Bootstrapping a fresh non-brain repo yields `workshop/issue-tracker.json` and an `issue-tracker` branch on origin, so
  `sdlc issue new` writes no checkout and `start-plan` cuts the branch (test with a bare remote).
- Bootstrapping a brain repo creates no tracker (test).
- In a legacy repo, `issue new` and `start-plan` print the migrate hint (test).

## Plan

Proposed durable plan: [new repository tracker initialization](../plans/000543-new-repo-tracker-plan.md). Await TL approval of the explicit initialization seam before change-code.

- [ ] Add an explicit empty-only issue init step reusing tracker migration, with brain/opt-out exclusion, prerequisite refusals, safe retry and checkout completion.
- [ ] Add the shared legacy migrate hint at issue new and start-plan; document the default new-repository sequence separately from adoption/refresh.
- [ ] Verify pure eligibility and real-Git sequence/failure/concurrency fixtures, private CLI conformance, recorded suite and whole-issue close review; land through sdlc.


## Log

### 2026-10-21

- Claimed and entered planning from refreshed main. Read-only survey confirms weave link only declares substrate/seeds a manifest, compile is shared with bootstrap/CI/refresh, and no separate repo-creation command exists. Hooking implicit publication into compile would change existing non-Git/refresh contracts; scaffold absence is not proof of newness and loses retry intent after partial setup. Recommend an explicit issue init final step in the canonical new-repo runbook, reusing migration's exact-root/two-phase publication. Pending TL approval: this interpretation of the allowed equivalent init step, and opt-out as omission/declined issue scaffolding rather than a new config format. No implementation changes.

- TL approved explicit issue init, full flow, no implicit weave publication, and private production CLI/bare-Git conformance with no stage (Couch 87ebe800-5c28-4e19-b11b-5bf5c2c1453a). Added the condition that canonical operator/TL new-repo/bootstrap procedures end with init, and empty legacy inventories hint init rather than migrate. No code changed before approval.

- Plan-quality round 1 requested an enforceable named transition owner (PQ-1), function-level test strategy instead of prose inventories (PQ-2), and carried Minor recovery catalog integration (PQ-3). Verified catalog obligation in recovery_contract_test.go. Revised durable plan names opaque InitState, InitEvent, InitEffect and InitTransition, deterministic effect barriers, shared local-only advisory hint selection versus full publication inventory, and recovery proofs. These formalize approved behavior; no implementation started.

## Revisions

### 2026-10-21 — approved default initialization boundary

The default new-repo procedure explicitly finishes with issue init. Legacy issue new/start-plan hint init for an observed empty inventory and migrate for existing issues; brain/declined scaffold skip. Durable plan's revision records the full condition and acceptance. Existing composition/refresh behavior stays unchanged.

## Estimate

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only.* The source is tagged stale by sdlc; amounts are provisional. Scope is one repo, including code, real-Git/private CLI verification, docs and one close review; machine queue idle time is excluded by the measured-actual engine.

Decomposition after clean plan-quality: issue/spec work uses the low issue-spec design anchor 0.5h with raw implementation 0.2h; no spec discount on the spec itself. The new pure transition module uses raw design 1.0h with thorough-spec factor 0.2 and raw implementation 0.6h. Existing migration-backed IO uses one smaller-go-module (raw design 0.3h, impl 0.5h), and hint/recovery integration another (0.2h, 0.4h); both design values take the 0.2 thorough-spec factor. Docs use raw design 0.1h/impl 0.15h, review overhead 0.1h/0.3h, again 0.2 on design. All implementation entries below already apply v3.1's 40% once. Familiarity is 1.0 for Go/Cobra/tracker/Git patterns; design buffer is 15% for the approved detailed plan.

Library check: existing migration/trunk APIs already supply CAS, root verification and cutover. The only greenfield module is the issue-specific pure transition policy; no general library supplies that domain decision, and adding a framework would not shorten it. No new API discovery or cross-repo propagation project is required. Total derives from design subtotal 0.84h and implementation subtotal 0.86h: 0.84*1.15 + 0.86 = 1.826h, rounded to 1.83h.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec design=0.50 impl=0.08
item: greenfield-go-module design=0.20 impl=0.24
item: smaller-go-module design=0.06 impl=0.20
item: smaller-go-module design=0.04 impl=0.16
item: atlas-docs design=0.02 impl=0.06
item: milestone-review design=0.02 impl=0.12
design-buffer: 0.15
total: 1.83
```
