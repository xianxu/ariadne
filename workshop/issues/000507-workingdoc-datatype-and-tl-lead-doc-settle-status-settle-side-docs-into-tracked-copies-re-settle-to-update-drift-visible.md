---
id: 000507
status: working
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-21
estimate_hours:
collaboration-mode: collaboration
card_mirror: '9a1ff9f50c63a2689a0dad7de3de665069b68ad8' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-21T12:19:30-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "b287a525", done: "f73e0fa9"}
---

# workingdoc datatype and tl-lead doc settle/status: settle side docs into tracked copies, re-settle to update, drift visible

## Problem

Side docs (the operator–TL co-edited pages beside the session file) have become the main design medium, but they're
untracked, so settled decisions can be lost, and projects can't link them durably. Operator, 10-20: make a
`workingdoc` type; copy a settled side doc into a tracked place; projects link it as background. Concern raised:
once copied, the copy freezes while discussion continues in the side doc.

## Spec

- `construct/datatype/workingdoc.md`: a living, co-edited page on one topic, with owner, topic and a
  `settled-into: <path> @ <commit>` pointer.
- `tl-lead doc settle <side-doc> [--to PATH]`: copy with review markers stripped to `workshop/workingdoc/<date>-<slug>.md`
  (default) or the given path (e.g. a project's `design.md`), commit it, stamp the side doc's pointer. **Re-settling
  overwrites and commits the copy**, so git history is the copy's version history.
- `tl-lead doc status`: side docs edited since their last settle, so drift is visible.

## Done when

- The datatype, both verbs and their tests exist; `tl-lead --help` documents settling.

## Plan

- [ ] Implement the workingdoc prototype, marker/refusal projection and pure settlement model, with regression tests.
- [ ] Add guarded selected-copy publication and retry behavior through the existing product trunk publisher and ops main publication, with stateful fixtures.
- [ ] Wire doc settle/status, update help/maps, verify the suite, close, arrange the hub-owned live check, then merge.

Durable implementation plan: [workingdoc settlement](../plans/000507-workingdoc-settle-plan.md). Full flow; awaiting TL approval before change-code. One implementation close boundary; no Mx tags.

## Log

### 2026-10-20

### 2026-10-21

- Claimed and started from main after #514's completed merge/cleanup. Existing reusable seams: SessionSync Peek/Apply and whole-file CAS, sidedoc link/index rebasing, and sdlc trunk publication. No Go marker-cleaning implementation exists; the source of marker semantics is construct/local/fix/review-convention.md. Bounded read-only exploration confirmed the file-aware fake gap and the need for a Git blob baseline rather than mtime.
- TL clarified publication and cleaning (receipt 5ea7ef7f-3e12-4473-8369-91aa89a8425c): product repos use trunk publish; ops operates directly on primary main, inferred from the repository. Unresolved commentary/proposals and malformed markers refuse with diagnostics; settling is not implicit acceptance. Literal code and anchors survive. These decisions are recorded in the durable plan, alongside explicit partial-publication outcomes and retries.
- Plan proposes a pinned per-path baseline and JSON receipt extension to the existing trunk publisher because its current merge-base-only comparison refuses repeat settlement from a stale hub checkout. This design delta remains for TL review; no implementation started. ARCH-DRY: reuse one publisher and one projection; ARCH-STATE: commit/push and Couch stamping have explicit separate confirmation boundaries.
- TL approved the plan and publisher baseline/receipt extension (7f2905de-b94c-4cc8-bb53-582cc8aeb5b7). Pinned full flow explicitly because the pre-code diff alone inferred quick flow. Plan-quality round 1 found PQ-1: mutable destination reread could publish unvalidated bytes, and PQ-2: testing needs named functions/strategies. TL approved the immutable payload guard/private ops Git index refinement (18ce25ca-d968-4481-8f80-879f39ab36ac). Added a plan Revisions section recording exact payload authority and current function test contracts; default trunk behavior must stay unchanged. No implementation started before gate clearance.

## Revisions

### 2026-10-21 — publication ownership and unresolved review items

The initial Spec's shorthand "review markers stripped" does not authorize deleting an unanswered operator item or accepting a pending proposal. Only already-accepted clean text can settle; remaining commentary, proposals and malformed markers refuse. Product copies publish through the trunk path; ops copies commit/push directly on primary main under its operating charter. The original working page remains live and receives its confirmed path/commit pointer through Couch Apply. Re-settling updates the same tracked path and Git retains its history. Done-when includes fixtures for these decisions, interrupted publication/stamping and content-based drift; a hub-owned live check precedes landing.
