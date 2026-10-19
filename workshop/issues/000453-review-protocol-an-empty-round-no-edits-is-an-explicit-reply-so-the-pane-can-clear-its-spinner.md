---
id: 000453
status: working
deps: []
github_issue:
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'b3b0764e2fca8bc0279ce9b59444c9ad1f43bbd3' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-18T22:24:44-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "7b13c8a6", done: "bf1368f2"}
---

# review protocol: an empty round (no edits) is an explicit reply, so the pane can clear its spinner

## Problem

Since ariadne#437 / pair#664 retired the legacy review handoff, an agent that answers a review round with no edits writes no round artifact. In pair's review pane the awaiting spinner and the Alt+c pending block then stay up until the next send. The legacy path cleared them on an empty handoff; #437 defines no empty-round reply. Found by pair:2 closing pair#664 (2026-10-18).

## Spec

An empty round is a valid, explicit reply: the agent (xx-fix / the review skill) always writes the round reply, with an empty records list when it has no edits (with its explanation or questions in the usual place), and the protocol document says so. pair's pane treats a reply with zero records as a completed round: clear the spinner and the pending block, commit nothing. Paired pair issue for the pane side.

## Done when

- The skill writes an empty-round reply when it has no edits; a test covers it.
- The protocol target states that an empty reply completes a round.
- With pair's side landed, a live round where the agent makes no edits clears the spinner.

## Plan

- [x] Red-first test the no-edits producer and empty-success commit decisions in
  the existing reviewexport skill-contract tests, plus a read-only skill scenario.
- [x] Require an apply on every completed workbench review, including
  `{"records":[]}`. Keep explanations/questions in the session reply. Distinguish
  changed replies (commit returned Body verbatim) from empty replies (no commit),
  preserving existing hash, context, operator-lock and uncertain-result guards.
- [ ] Update the ariadne workflow map and verify focused contracts, exported
  substrate integrity, recorded suite and the automatic close review.
- [ ] Land producer instructions independently of pair#690, report to the hub.


## Log

### 2026-10-18

- Hub approved the plan and moved protocol-target wording plus the real no-edits
  pane check to pair#690 (pair:2). Sent that owner the hand-off for its Log. This
  issue remains the producer-only half and may land first.
- ARCH-DRY: reuse the existing apply seam, no parallel reply file or state.
  ARCH-ORDER: only nonempty returned Body authorizes the normal agent commit;
  an empty success completes the producer response without inventing evidence.
  No new wire fields, runtime state, resource lifetime, or hot path is introduced.

- Baseline read-only skill scenario: the agent chose "No apply or agent commit: the skill specifies no empty-round submission." Real edits still committed returned Body verbatim; old-binary empty success caused no commit or retry. The missing mandatory empty apply is the observed defect.

- Red: the new existing-table/phrase contract test failed on all missing empty
  producer rules and the absent success decision table. Green: the full
  reviewexport package passed (0.266s), including export integrity and the
  unchanged routing/recovery scenarios.
- Fresh read-only skill scenario after the update sends `{"records":[]}` through
  apply, commits no agent round, and keeps explanations in the session reply.
  The real-edit control commits returned Body verbatim; an older-binary empty
  success triggers neither commit nor retry/recovery. The baseline agent skipped
  apply, so the behavior change was observed rather than inferred.

- Final focused reviewexport and processmanual checks passed (0.234s/0.164s). Regenerating the process manual produced no semantic delta after normalizing its generated EOF blank line; no manual changes are needed. Diff whitespace check is rerun before committing.

## Revisions

### 2026-10-18 — producer scope and peer acceptance ownership

Hub approved the producer-only boundary. The original protocol-target wording
and live pane no-edits Done-when steps belong to pair#690, owned by pair:2; that
owner will log the hand-off and verify the pane after its implementation lands.
This issue proves that the shared producer always sends an empty apply and never
commits an empty reply; it may land before the peer engine/pane change. Until
then the old engine harmlessly accepts the empty apply without publishing a round.
No peer source or protocol target is edited here.
