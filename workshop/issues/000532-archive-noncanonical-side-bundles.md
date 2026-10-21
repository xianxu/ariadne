---
id: 000532
status: working
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-21
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'eec450bea3f5a87732c8265a62097a127b4868bd' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-21T16:36:06-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "8de6ade2", done: "234df8ba"}
---

# Archive preflight misses noncanonical side bundles

## Problem

pair#768 landed in PR521 (https://github.com/xianxu/pair/pull/521; reported merge 2026-10-21 14:37:41 PDT), but post-merge SDLC cleanup refuses with `active or conflicting archive artifact workshop/plans/000768-darwin-pty-final-output-plan.md.side/shared-owner-draft.md`. The card is done and the issue and parent plan are archived, while this tracked side document remains active. The documented retry `sdlc merge --branch 000768-darwin-pty-final-output --yes` repeats the refusal. Reported `sdlc issue recovery reconcile --issue 768` finds no unfinished operations and deleted the remote branch without archiving this leftover. Successful landing therefore leaves an incomplete artifact bundle and no successful documented cleanup retry.

## Spec

Prevent new parent-only archives: close and the pinned-head landing preflight refuse an owned noncanonical `<document>.md.side` folder before review/archive/remote merge effects. Name both the offending folder and canonical `<document>.side` rename; the operator preserves its contents and updates links before retry. Canonical `.side` remains unchanged. Use one pure naming policy and thin filesystem/Git readers (ARCH-DRY, ARCH-PURE). Scope to the closing/landing issue family and configured active issue/plan roots, so historical leftovers for other issues do not block unrelated work.

The operator/TL chose explicit one-off Git moves for the two existing Pair leftovers (#768 and #757), preserving relative placement beneath history/plans and bytes, with their issue Logs recording the repair. No new recovery command, transition type, supplementary proof, automatic normalization, or weakening of historical proof is introduced. This issue ships prevention; the one-off repair is coordinated with the owning hub after the plan approval.

## Done when

- Close and pre-merge preflight refuse the exact incident path `workshop/plans/000768-darwin-pty-final-output-plan.md.side/shared-owner-draft.md`, naming `workshop/plans/000768-darwin-pty-final-output-plan.side`; no review, archive or merge publication effect occurs on refusal.
- The shared policy covers issue and plan document bundles, nested members, tracked and untracked close inputs, configured roots, and all selected issue families. Canonical `.side` passes; unrelated old families do not block a new landing.
- Immutable archive planning refuses the malformed family rather than publishing a parent-only archive. Ordinary canonical archive/proof behavior remains passing.
- Regression tests preserve the existing strict old-PR proof after a one-off move; independently selected future landings still succeed. No claim is made that the repair retroactively makes an old archive commit complete.
- README/atlas document canonical naming, corrective rename before close, and the explicit one-off repair policy. The Pair repair records exact source/destination bytes and issue Logs through its ordinary owning workflow, without bypassing that repository's gates.

## Plan

- [ ] Add one pure issue-scoped noncanonical-side naming validator in cmd/sdlc/archivepolicy.go, with adversarial table tests; thin close filesystem reader sees tracked/untracked top-level folders without following symlinks. Read each configured root once (O(entries)).
- [ ] Wire close before review/state changes; wire pinned-head landing preflight (including train readiness) before remote merge. Thread configured plan roots as needed. Reuse the validator inside immutable archive planning as defense in depth. Tests enter the real close, publish-gate and planner seams, with no publication on refusal; canonical controls remain valid.
- [ ] Verify a historical parent-only archive followed by an explicit byte-preserving one-off side move still fails old proof, while a separately selected future landing succeeds. Record the boundary and coordinate the two real Pair repairs with ops:0; do not mutate Pair :0 or another slot's tree.
- [ ] Document policy, targeted regressions and counted N/2N reader budgets; commit/push, run the recorded suite, close through fresh review, then land and report.

## Log

### 2026-10-21 — filed from pair#768 cleanup failure

Git verification in pair slot2: HEAD/reviewed source `8b7a72b86f618ab0b06f15e966570ad44a4e1d0d`; merge `68fac7f0bf4312610f1348601fe4f05765d865f0` has that source as its second parent (Git commit timestamp 14:37:40 -0700). Observed origin/main `1ed39dfdbbe0182b72c4b3a62b5504b2e725ce5d` archives the parent issue and plan. Both HEAD and origin/main contain the active exact side path above as regular mode 100644 blob `670ed082db7429077580eb805edcbcc3907a8ecc`; neither contains its history-side counterpart. Archived issue blob: `5451b5f38eeec15da8be9969cf4b2d4b2413f1d4`; archived parent-plan blob: `d321e0d969632c5818acb92916dab2c1d77efcdc`. No canonical plan.side counterpart was present. No tests, repair, checkout mutation, or cleanup replay performed during filing; retry/reconcile results above are supplied incident evidence.

### 2026-10-21 — investigation

SideDir strips `.md`; landingSideRelated accepts broader plan-ID `.side` family members. The snapshot therefore sees `.md.side` but planLandingSideMoves omits it, and only post-commit confirmLandingArchiveCommit rejects the leftover. Bounded read-only agent survey confirmed the close/publish gate currently has no side naming guard. No implementation or repair performed.

## Revisions

### 2026-10-21 — TL selects prevention and explicit one-off repair

Per ops:0 delivery 841eb335-f3f0-4a1d-b817-91d0c7bb9176, the earlier scope requiring automatic normal-retry repair is superseded. Refuse noncanonical names at close and preflight, naming the canonical rename. The two existing Pair leftovers use explicit ordinary Git moves plus issue Log notes; no new recovery mechanism or transition. Old proof remains strict: its re-derived expected set excludes the one-off history entries and cannot retroactively attest them, whereas future selected issue families exclude those unrelated entries. This distinction was reported to the TL before code.

Superseded acceptance included automatic idempotent cleanup retry from a done-card, already-archived parent without operation receipt, plus uncertainty/provenance repair checks. That requirement is removed by the decision above rather than implemented as a new proof-chain mechanism. Original motivating incident and source evidence remain in Problem and Log.
