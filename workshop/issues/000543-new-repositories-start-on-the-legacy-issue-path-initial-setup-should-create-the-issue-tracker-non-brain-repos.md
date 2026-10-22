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

- [x] Add an explicit empty-only issue init step reusing tracker migration, with brain/opt-out exclusion, prerequisite refusals, safe retry and checkout completion.
- [x] Add the shared legacy migrate hint at issue new and start-plan; document the default new-repository sequence separately from adoption/refresh.
- [x] Verify pure eligibility and real-Git sequence/failure/concurrency fixtures plus private CLI conformance; submit the recorded pushed-head suite and whole-issue close review. Land through sdlc only after a passing suite and SHIP.


## Log

### 2026-10-21

- Claimed and entered planning from refreshed main. Read-only survey confirms weave link only declares substrate/seeds a manifest, compile is shared with bootstrap/CI/refresh, and no separate repo-creation command exists. Hooking implicit publication into compile would change existing non-Git/refresh contracts; scaffold absence is not proof of newness and loses retry intent after partial setup. Recommend an explicit issue init final step in the canonical new-repo runbook, reusing migration's exact-root/two-phase publication. Pending TL approval: this interpretation of the allowed equivalent init step, and opt-out as omission/declined issue scaffolding rather than a new config format. No implementation changes.

- TL approved explicit issue init, full flow, no implicit weave publication, and private production CLI/bare-Git conformance with no stage (Couch 87ebe800-5c28-4e19-b11b-5bf5c2c1453a). Added the condition that canonical operator/TL new-repo/bootstrap procedures end with init, and empty legacy inventories hint init rather than migrate. No code changed before approval.

- Plan-quality round 1 requested an enforceable named transition owner (PQ-1), function-level test strategy instead of prose inventories (PQ-2), and carried Minor recovery catalog integration (PQ-3). Verified catalog obligation in recovery_contract_test.go. Revised durable plan names opaque InitState, InitEvent, InitEffect and InitTransition, deterministic effect barriers, shared local-only advisory hint selection versus full publication inventory, and recovery proofs. These formalize approved behavior; no implementation started.

- Plan-quality rounds 2 and 3 were clean; change-code full-flow passed after estimate-quality INFO. Estimate's implementation entries include the real-Git/private-CLI verification, with the dedicated review row covering the final boundary. The tight allowance is provisional, not a reason to reduce acceptance; measured actuals will record the work.
- Implemented the opaque pure initialization protocol and effect dispatcher, reusing migration bootstrap/CAS/root verification and pinned cutover introduction. No persistent phase artifact or implicit weave publication. Added recovery catalog proofs and inventory-sensitive local advice; absent scaffold plus existing history suppresses advice.
- Verification: pure tracker package passed (24.990s); targeted init/recovery/hint/migration/recovery-contract set passed (cmd/sdlc 90.944s). Stateful real-Git barriers cover every before/after effect, lost acknowledgments, cancellation, dirty checkout, history/branch evidence, conflicting markers, competing clones, prerequisites and later card preservation. Initial failures were fixture-only trailing-newline OID comparisons and an empty fixture commit; corrected before GREEN. A declined-scaffold-with-history regression first failed, then the scaffold check fixed it and the complete targeted set passed.
- Counted inventory guard: 8/16 unrelated files used 94/94 Git processes; 8/16 branches used 126/158, retaining fixed file cost and linear branch scanning.
- Private built production CLI against a real disposable bare origin passed init -> first issue new (HEAD/index/worktree unchanged) -> repeated init after later issue -> claim -> start-plan, producing 000001-first-real-cli-work. Evidence retained locally at /Volumes/TestScratch/ariadne-543-cli-8gw8_mh5/commands.json. No stage per TL approval.
- Procedure sweep found TL-facing team creation instructions in cmd/tl-lead/help.md and team_cmd.go; updated their delivered next steps under the approved scope. New ordered-procedure regression failed before the update; TestTeamNew and TestLeadHelp passed afterward (0.628s). README and both setup/bootstrap atlas entry points now finish new-repository setup with explicit init. No other repo-creation procedure was found. Global GOCACHE is the main-SSD default; no private cache created.

- Go lint and recovery package checks passed (1.096s / 0.133s). Implementation committed as 8d0ac0109 and PR385 opened. Recorded suite submitted on that pushed head, run lease-1773336371 / attempt dbec417e9238668ff9df807e468f0597, currently queued (not test evidence yet). Worker policy permits close review to overlap this queue; no landing until the final code has its recorded pass. Plan's acceptance checkbox records submission, not a suite pass or review verdict.

- Close round 1 REWORK: BR-1 Critical (authoritative-inventory-completeness). Reviewer reproduced a remote-only legacy issue branch created since the initializer's last fetch, missed by shared migrationBranchRefs's cached refs. Verified the shared scanner does not refresh all heads. Explicitly canceled lease-1773336371 before execution; canceled output confirms no suite result. Appended a bounded correction to the durable plan and sent it to TL; production code held pending approval. Existing code and targeted successes do not dispose BR-1.

- TL approved BR-1 correction (Couch 258bb7aa-6771-4df9-afb1-9784ce0a3803). Implemented shared all-heads observation and one names/OIDs snapshot; both issue and close-anchor consumers read the pins. Main mismatch and failed refresh refuse; no-prune-tags/refmap suppression constrain effects to the intended remote-head cache mapping. All eight before-fix regressions produced intended RED, then passed (11.666s). Broad targeted set passed (106.994s), including branch movement after enumeration, fetch-scope preservation and counted growth checks. Private rebuilt production CLI passed remote-only refusal via init/migrate and fresh init -> no-checkout issue -> repeat init -> claim/start-plan; local evidence /Volumes/TestScratch/ariadne-543-br1-cli-b2g2r2ln/commands.json. BR-1 disposition awaits the next gate review; old suite canceled, fresh suite not yet submitted.

- Close round 2 REWORK: BR-1 addressed; BR-2 Critical repeats authoritative-inventory-completeness because branch/dirty-worktree scans omit archives. Canceled queued lease-2849438579, with no executed suite/pass. Source audit additionally identified flat legacy history, inherited/unrelated branch evidence, detached clean worktree heads and ignored candidates as cells the complete rule must address. Appended a whole-inventory root × scope re-plan and held production changes for TL approval; this replaces site-by-site repairs.

- TL stopped importer widening and directed a conservative init presence gate (Couch 5119a959-0128-4420-89ee-9a91fb0b08d9). Superseded the prior whole-rule importer proposal: init refuses recognized evidence anywhere across three roots and all branch/worktree/index/filesystem scopes, importing/merging nothing. Pre-existing migrate completeness gaps filed separately as unclaimed #546; its card/details published with no checkout write. Existing BR-1 fresh/pinned helper remains; no new migration import/refusal policy in #543. Revised durable plan is authorized to proceed.

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
