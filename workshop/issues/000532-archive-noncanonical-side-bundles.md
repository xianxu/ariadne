---
id: 000532
status: codecomplete
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-21
estimate_hours:
collaboration-mode: collaboration
card_mirror: '86007046f4e98555b1bde5d5ed8239a068290e3c' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-21T16:36:06-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "8de6ade2", done: "234df8ba"}
actual_hours: 0.25
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

- [x] Add one pure issue-scoped noncanonical-side naming validator in cmd/sdlc/archivepolicy.go, with adversarial table tests; thin close filesystem reader sees tracked/untracked top-level folders without following symlinks. Read each configured root once (O(entries)).
- [x] Wire close before review/state changes; wire pinned-head landing preflight (including train readiness) before remote merge. Thread configured plan roots as needed. Reuse the validator inside immutable archive planning as defense in depth. Tests enter the real close, publish-gate and planner seams, with no publication on refusal; canonical controls remain valid.
- [x] Verify a historical parent-only archive followed by an explicit byte-preserving one-off side move still fails old proof, while a separately selected future landing succeeds. Record the boundary and coordinate the two real Pair repairs with ops:0; do not mutate Pair :0 or another slot's tree.
- [x] Document policy, targeted regressions and counted N/2N reader budgets; commit/push, run the recorded suite and close through fresh review.

Landing and reporting follow through the binary after this verified checklist; they are not claimed complete here.

## Log

### 2026-10-21 — filed from pair#768 cleanup failure
- 2026-10-21: closed — RED/GREEN reproduced the exact .md.side leak; production close, pinned-head/train and archive-planning regressions plus canonical and historical-repair/future-landing controls passed (119.613s); full review-entry refusal test passed (14.350s); golint and sidedoc pass. TL accepted private real-Git acceptance with no stage. Recorded full suite lease-1938211911 on pushed d6309d22f5cc7bf4911d0eba4f97abc0c7a1a6d8 is queued, not passing: review overlaps under tl-worker TESTS; the last Plan item stays pending and landing must await the pass. --no-project: TL owns project updates. (done-when: contract moved since change-code; judged by the close review); review verdict: SHIP

Git verification in pair slot2: HEAD/reviewed source `8b7a72b86f618ab0b06f15e966570ad44a4e1d0d`; merge `68fac7f0bf4312610f1348601fe4f05765d865f0` has that source as its second parent (Git commit timestamp 14:37:40 -0700). Observed origin/main `1ed39dfdbbe0182b72c4b3a62b5504b2e725ce5d` archives the parent issue and plan. Both HEAD and origin/main contain the active exact side path above as regular mode 100644 blob `670ed082db7429077580eb805edcbcc3907a8ecc`; neither contains its history-side counterpart. Archived issue blob: `5451b5f38eeec15da8be9969cf4b2d4b2413f1d4`; archived parent-plan blob: `d321e0d969632c5818acb92916dab2c1d77efcdc`. No canonical plan.side counterpart was present. No tests, repair, checkout mutation, or cleanup replay performed during filing; retry/reconcile results above are supplied incident evidence.

### 2026-10-21 — investigation

SideDir strips `.md`; landingSideRelated accepts broader plan-ID `.side` family members. The snapshot therefore sees `.md.side` but planLandingSideMoves omits it, and only post-commit confirmLandingArchiveCommit rejects the leftover. Bounded read-only agent survey confirmed the close/publish gate currently has no side naming guard. No implementation or repair performed.

### 2026-10-21 — implementation and targeted verification

TL approved the short prevention-only plan (ff2d785e-1979-42dd-9dc1-dd1e3ece13b0); pair:4 and pair:2 own the separate real repairs. Shared validateArchiveSideNames checks selected IDs in one pass; checkCloseSideNames reads each active root once, including untracked directory names, without following symlinks. Close invokes it before review/state changes; runLandingPublishGate reads the pinned PR tree with the existing bounded archive tree reader, passes configured plan roots through train readiness and integration replay, and the immutable planner applies the same policy before creating writes. No old archive proof is weakened.

RED: close accepted tracked/untracked incident folders; pinned-head preflight returned nil; remote archive committed the parent only. The first exploratory RED run overlapped a delegated test file appearing and tripped the root hermeticity guard; it is not acceptance evidence. Repeated with stable source status, the same three behavioral failures reproduced. GREEN: those three regressions passed in 8.787s. Broader production-seam tests passed in 119.613s: pure naming matrix, configured absolute roots and permission errors, counted 8/16-entry inventories (exactly two directory reads), mutable checkout deletion cannot hide a pinned folder, train publish gate, canonical rename control, canonical archives, publication eligibility/anchors/cancellation, and real private Git historical repair/future-landing proof. No full suite result claimed yet.

The runtime naming validator is the mechanical enforcement at close/preflight/archive planning. A Go AST lint cannot establish which dynamically enumerated artifact family belongs to a selected issue, nor prevent an agent from naming a Markdown folder incorrectly; no unrelated source-pattern lint is invented.

### 2026-10-21 — second case, pair#757 (reported by pair:4)

pair#757 landed via PR532 (https://github.com/xianxu/pair/pull/532), reported GitHub MERGED at 2026-10-21T22:12:21Z with merge commit dfb44839f7a8c4006690e6d05c1a3e994b31194d. Post-merge cleanup refused `active or conflicting archive artifact workshop/plans/000757-mock-to-real-plan.md.side/browser-progress-negative-control.log`. The directory also holds browser-pass.log. Both logs remain unchanged; parent plan is archived at workshop/history/plans/000757-mock-to-real-plan.md. One documented already-merged recovery, `sdlc merge --branch 000757-mock-to-real-app-template-per-screen-real-or-mock-dev-router-refusing-production-build-fixtures-contract-tests-product-shaping --yes`, repeated the exact refusal. Landed branch preserved; canonical slot switched to main-slot4. TL authorized manual movement of the unchanged two-file bundle into canonical history plan.side as a separate repair. No lifecycle/CI/review bypass; all product gates had passed, including the registered pair#778 same-head test rerun. This block records supplied peer evidence; this worker has not independently rerun the old landing or performed the repair.

2026-10-21: Full runCloseWithReview entrypoint regression also passed (14.350s), proving malformed tracked/untracked folders stop before the judge. TL accepted private real-Git production-seam acceptance and explicitly waived stage (6499451a-ee94-4c54-8c7d-9febf6a7ed3f). Recorded full suite is the next check; not yet passing. No cache override or machine change is required.

### 2026-10-21 — recorded branch suite passed

The original lease-1938211911 survived the Couch restart without a replacement runner. It acquired the full-suite slots after 4114s in FIFO and passed on pushed code head d6309d22f5cc7bf4911d0eba4f97abc0c7a1a6d8: all eight shards, 2045 cmd/sdlc tests and 75 other packages, zero failures; wall 723.6s (12m4s). Saved log: .git/sdlc/local-checks/d6309d22f5cc7bf4911d0eba4f97abc0c7a1a6d8/suite.log; reported digest prefix 37e0c9039557; sdlc-local/suite recorded on that head.

Whole-issue close returned SHIP with no findings and adopted measured 0.25h, evidence eb95a774a2e110fdc66e413d9356ad6992507b24. The reviewer independently passed focused and broader archive/publish regressions. The subsequent required main integration 8823171c88a21e828046174c6c223b4ffb25d02e resolved only the published issue's mirror conflict, retaining the sdlc-generated codecomplete card fields. Its diff from the tested head contains only issue/review bookkeeping. PR378 has green server checks. Landing has not yet occurred; sdlc merge owns current-main integration and any authoritative integrated suite.

## Revisions

### 2026-10-21 — TL selects prevention and explicit one-off repair

Per ops:0 delivery 841eb335-f3f0-4a1d-b817-91d0c7bb9176, the earlier scope requiring automatic normal-retry repair is superseded. Refuse noncanonical names at close and preflight, naming the canonical rename. The two existing Pair leftovers use explicit ordinary Git moves plus issue Log notes; no new recovery mechanism or transition. Old proof remains strict: its re-derived expected set excludes the one-off history entries and cannot retroactively attest them, whereas future selected issue families exclude those unrelated entries. This distinction was reported to the TL before code.

Superseded acceptance included automatic idempotent cleanup retry from a done-card, already-archived parent without operation receipt, plus uncertainty/provenance repair checks. That requirement is removed by the decision above rather than implemented as a new proof-chain mechanism. Original motivating incident and source evidence remain in Problem and Log.

### 2026-10-21 — checklist bookkeeping clarification

The original last checkbox also included future landing/reporting. Its verified documentation/test/close portion is now checked after the recorded pass; landing/reporting remain explicit binary-gated follow-through rather than a premature completion claim. No scope, mechanism or acceptance criterion changed.
