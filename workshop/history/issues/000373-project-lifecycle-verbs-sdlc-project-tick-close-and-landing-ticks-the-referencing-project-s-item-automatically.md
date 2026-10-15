---
id: 000373
status: done
deps: []
github_issue:
created: 2026-10-14
updated: 2026-10-15
estimate_hours:
card_mirror: 'f82e3a918d74a3b575a11e026c87cb2dc7261a54' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-15T01:00:50-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
actual_hours: 2.60
---

# Project lifecycle verbs: sdlc project tick/close, and landing ticks the referencing project's item automatically

## Problem

Projects are hand-maintained files (`workshop/projects/…`, ops `teams/*/projects/*/project.md`). Nothing ticks a project's task when its issue lands, and there's no verb to close a project: on 2026-10-14 ariadne-robustness-1 showed 8 landed MVP items unticked. The close gate finds referencing projects but ticking/closing stays manual. Retro item (d); pairs with the operator's rule that the TL status file is mutual focus and its projects must stay current.

## Spec

`sdlc project tick --issue N` (idempotent; finds every referencing project fleet-wide, as the close gate already does) and run it automatically at landing (`sdlc merge`, after archive) for the landed issue. `sdlc project status PROJECT` prints done/open against mvp_scope and done_when. `sdlc project close PROJECT --verified` checks every MVP item is done and records the close (archive per the project datatype). Writes to another repo's project file go through that repo's normal publish path.

## Done when

- Landing an issue ticks it in every referencing project (test across two repos).
- `project status` and `project close` work on ariadne-robustness-1's real file (dry-run).

## Plan

Durable plan: `workshop/plans/000373-project-lifecycle-verbs-sdlc-project-tick-close-and-landing-ticks-the-referencing-project-s-item-automatically-plan.md`.

- [x] Pure core: `TickLandedIssue`, `MVPProgress` (mvp-landed guard on it)
- [x] `tickLanded` + `sdlc project tick` (origin publish via TrunkFile, TL-owned allowed)
- [x] merge post-archive hook
- [x] `project status` MVP + done_when
- [x] `project close` --verified, MVP re-check, --dry-run, symlinked record
- [x] docs: helptext, project datatype ownership, atlas
- [x] verify on ariadne-robustness-1 (dry-run), full suite

## Log

### 2026-10-14

### 2026-10-15
- 2026-10-15: closed — Recorded suite sdlc-local/suite green on 0aa64f88 (branch with main merged, after review rounds 1 and 2). Done-when covered: TestLandingTicksPeerProject and TestRunMerge_LegacyLandingTicksPeerProject (cross-repo tick on origin, and via tickOutsideLock the tick runs after the merge lock, mutation-checked on both paths), TestTickOnLanding_TLOwnedIsOptIn, TestTickOnLanding_FailureWarnsWithRetry, TestProjectTick_PublishesEveryCitingProjectOnOrigin, TestProjectStatusShowsMVPAndDoneWhen, TestProjectCloseVerifiedAndMVPGates, TestProjectCloseDryRunReportsAndWritesNothing, TestProjectCloseLinkedRecord (commit-point order, mutation-checked), TestProjectCloseLinkedRecordRollsBack, TestProjectCloseGatesMatchTheCatalog (no-retro and no-mvp refusal and ack vs GateCatalog); unit TestTickLandedIssue/TestMVPProgress/TestLandingTick. Read-only dry runs on ariadne-robustness-1: status shows mvp 44/44 + done_when; close --dry-run refuses (executing, no retro) and would close the real teams/ file in place.; review verdict: SHIP
- 2026-10-15: flow upgraded quick → full — 607 added lines in code files (limit 100); an earlier round of this close already ran the full review
- TL approved points 2–4 of the plan; point 1 (landing ticks TL-owned projects) is the operator's call, so TL-owned projects are opt-in via `landing_tick: true` (default `project.TLLandingTickDefault = false`) and merge prints `sdlc project tick --issue R --pr N` for the TL instead. Flipping the default is one line.
- Landed issue comes from the branch's id prefix (#272: one issue, one branch), so a resumed landing still ticks; the tick is idempotent.
- Real-file dry runs: `project status` on ariadne-robustness-1 prints `mvp: 44/44 landed`; `project close --dry-run` refuses on status executing and the missing retro and would close the real teams/ file in place; `project tick --issue ariadne#240 --pr 241 --dry-run` → no change (row names PR #241); `--issue ariadne#376` → would publish to ops's origin.
- Bug found and fixed: `project close` on a linked ops record renamed the link and orphaned teams/…/project.md.
- Mutation checks: disabling the landing_tick policy, the durable-merge hook call, and the linked-close branch each fail their tests.
- Boundary review round 1, FIX-THEN-SHIP. Fixed all three Importants: README project section (close `--verified`/`--dry-run`, `project tick`, `landing_tick`); `cmd/tl-lead/help.md` (TL runs the printed `sdlc project tick`, `landing_tick: true` opt-in, the linked close in place); the landing tick moved out of the merge lock on both paths with a 2m deadline (ARCH-CONSTRAINTS). Minors: linked close rolls back (rename is the commit point; `TestProjectCloseLinkedRecordRollsBack`); `--no-mvp` plus a gatesig row; dry-run "Would skip" exits non-zero; `mentionsPR` without a per-row regex. Two discovery routes: no change, since merge's issue is always the current repo (`identity.Repo` = the resolved repo) and both share `DiscoverInRoots` over the same roots.
- Boundary review round 2, FIX-THEN-SHIP. BR-1, BR-2, BR-5 and BR-6 were addressed. BR-3 and BR-4 needed regression tests, now added and mutation-checked:
  - `tickOutsideLock` wraps the merge lock and asserts origin is unticked as each locked section ends. Moving the tick back inside fails the durable test and the legacy in-place test.
  - The linked-close rename stub asserts the live link is gone at the commit point. The old order fails it.
  - A failed live-link restore is now joined into the error.
  - waiver-catalog-coverage: `TestProjectCloseGatesMatchTheCatalog` matches both project close rows (`no-retro`, `no-mvp`) against the emitted refusal (`assertGatesigAttributes`) and acknowledgement (new `assertGatesigAcknowledges`). The issue-close rows already have their refusal pins in close_atlasskip and closereview.
