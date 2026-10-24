---
id: 000591
status: working
deps: []
github_issue:
created: 2026-10-24
updated: 2026-10-24
estimate_hours:
card_mirror: '275f4e047d9de9f085cce14ff4c87ade378f3b74' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-24T01:44:54-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "2d5a43fb", done: "ec3016d9"}
---

# Actual attribution refuses normal concurrent claim held through merge and resting train checkout

## Problem

Follow-up to ariadne#580, reported by ops:0 from pair:4 closing pair#856. Actual now refuses session 062e2184-2117-4c26-b929-ceec3b5a8030 in pair-slot4: concurrent claims have unknown, resting or ambiguous active branch during 2026-10-23T21:35:40-07:00 to 2026-10-24T04:36:48Z. The slot legitimately held pair#850 while merging (its branch, then resting checkout during the train) and pair#856. Earlier close rounds measured 2.72 h and 3.33 h. The hub authorized --no-actual for that close meanwhile.

## Spec

Approved boundary (ops:0, 2026-10-24): keep witnessed own/foreign issue attribution; when concurrent reservations prevent ownership-only planning, exclude witnessed resting intervals as unassigned instead of refusing the complete sample. Unknown, detached and ambiguous branch history still refuses. Never infer an owner from a neighboring branch or status. Output names excluded branch-window minutes and the source/interval, explicitly distinguished from focused activity minutes. Adopted snapshots remain unchanged; #564 code stays held.

Use one shared pure resting-branch predicate for the existing main/main-slotN convention. Add a resting-excluded audit kind under snapshot v3: prior readers already reject unknown kinds rather than reinterpret them, so no schema bump or silent compatibility fallback. Validate that its branch is resting and that excluded notes cannot justify selected windows. Unknown members/types/duplicate handling remains unchanged.

## Done when

- A captured fixture of the reported merge/train resting transition reproduces the refusal before the fix and verifies the agreed attribution afterward.
- Concurrent-claim selection remains disjoint, and genuinely missing/ambiguous branch evidence and retained-snapshot corruption retain explicit safe behavior.
- Targeted regression/race checks, recorded pre-close and boundary review pass; semantics and limits are documented.

## Plan

- [x] Investigate evidence; receive approval for the changed resting-interval boundary.
- [x] Add recorded native reflog/refusal fixtures and red selector, saved-proof and output controls.
- [x] Implement one resting predicate, selector exclusion and saved-proof validation; disclose interval minutes through one warning formatter on fresh and replay paths.
- [ ] Verify focused/race/catalog and real-Git measurement controls, update atlas and lesson, push, run recorded pre-close, close and land.

Function-level adversarial strategy: SelectBranchActivity replays the recorded 68-second claim/start-plan window plus later own/foreign/resting transitions; selecting each issue must produce disjoint credit and exclude the same resting spans. Missing/detached/other-feature branches still refuse; a sole reservation keeps the existing ownership-only fallback. ValidateOwnedBranchProof mutations turn a resting exclusion into selected time or replace its branch with an issue/unknown branch and must refuse. MeasurementBranchWarnings reports exact source, interval and branch-window minutes identically for fresh/replay. Counted 8/16 transition tests retain linear sweep bounds. Existing native Git/tracker fake and three-harness source-lineage controls supply integration coverage; private executable proof avoids modifying any production actual.

Actions: measure; inspect evidence; explicitly close with --no-actual.

| Internal condition | Visible state | Recovery action | Possible loss |
|---|---|---|---|
| Proven own / foreign issue branch | Counted / excluded | Measure normally | No reassignment |
| Witnessed resting with concurrent claims | Excluded and unassigned, minutes named | Inspect the named interval | Resting planning effort is not counted |
| Unknown/detached/ambiguous history | Numeric actual unavailable | Inspect evidence or explicit --no-actual | Numeric calibration unavailable |
| Adopted snapshot | Original replay | Measure normally | No historical rewrite |

This is a bounded selector correction: no new IO scans, scheduler, lock or durable artifact family. The existing immutable snapshot owns audit lifetime. A lifetime lock cannot select historical work ownership; the captured immutable history remains the authority.

## Log

2026-10-24: Dispatched by ops:0 as a #580 follow-up, prioritized before #564 implementation. Filing claims the issue in ariadne:3; diagnosis only so far.


### 2026-10-24 — root cause and approved boundary

pair:4 preserved the close refusal and claim-window reflog in its plan.side/evidence/{actual-refusal.txt,head-reflog-claim-window.txt}. The raw per-worktree log is /Users/xianxu/workspace/pair/.git/worktrees/pair3/logs/HEAD (administrative pair3, logical pair:4). Reflog proves #850 -> main-slot4 at21:35:32, claim/start21:35:40, fast-forward21:35:43, main-slot4 -> #856 at21:36:48. This is intentional new-item planning while #850's suite used a detached snapshot, not train cleanup. Later #856 -> #850 at22:30:19 and #850 -> main-slot4 at00:22:29 -> #856 at00:23:20 are also explicit. Earlier measurements ran the prior engine; the source is not newly corrupt.

SelectBranchActivity currently sends every unresolved branchID through the same concurrent-reservation refusal, including Known=true main-slot4. ops:0 approved separating witnessed resting exclusion from genuinely unknown history (receipt6681ce8c-2204-4dce-99a8-2c99c1222d93), with excluded minutes visible and adopted snapshots unchanged. ARCH-DRY: share resting recognition and warning rendering across fresh/replay; no nearest-branch/status heuristic. Implementation expected inside the quick-flow shell.

## Revisions

- 2026-10-24: Hub approved the resting-excluded boundary after diagnosis, replacing #580's complete refusal for witnessed resting intervals only. Record explicit unassigned interval minutes rather than inventing ownership or backfilling old snapshots.


### 2026-10-24 — red/green and second incident

Recorded fixtures preserve native OIDs/actions/timestamps/branch names for pair#856 and pair#812, with only author identity replaced; peer-preserved refusal text is included. Pair#856 fails before the fix on the exact 68-second refusal; the minimal resting selector and new saved-proof kind also fail before implementation. The second case is the same witnessed-resting class: pair:1 confirms main-slot1 from03:01:09, claim05:01:56, fast-forward05:01:59, checkout#812 at05:02:21 (25 seconds). No Pair checkout or actual metadata changed.

Green targeted selector/proof checks passed; complete workspace/activetime/gatestate packages passed (36.704s/15.411s/0.432s). Native Git/tracker measurement, manual partition, retained replay after deleting the fixture's HEAD log, original foreign-tail controls and output/proof checks passed14.097s. The fixture's first manual run exposed a test expectation error: A's phase ends at its own last commit, so manual credit is8 minutes, not10; corrected the expectation while preserving the8-minute resting exclusion. An earlier run's hermeticity guard also observed my concurrent staging/docs edits, not a test mutation; all later cmd/sdlc test runs keep the checkout unchanged until joined. Catalog passed1.717s after staging. Focused race and exact-binary conformance precede recorded pre-close.


### 2026-10-24 — third recorded incident and focused race

pair:7 supplied pair#863's committed reflog capture: #842 -> main-slot7 at00:44:22 PDT, fast-forward00:44:32, #863 checkout00:44:38. Added the16-second interval to the same recorded selector regression; no mechanism or attribution policy change. Focused race checks passed workspace/activetime/gatestate in1.228s/1.235s/1.246s. All incident fixtures are read-only captures with metadata redaction described alongside them.
