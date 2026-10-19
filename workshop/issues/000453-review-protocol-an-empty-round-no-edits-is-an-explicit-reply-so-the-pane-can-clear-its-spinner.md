---
id: 000453
status: codecomplete
deps: []
github_issue:
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
collaboration-mode: collaboration
card_mirror: '5dc739b83a6c0cf58f6d2436a624d0ec1809d058' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-18T22:24:44-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "7b13c8a6", done: "bf1368f2"}
actual_hours: 0.09
---

# review protocol: an empty round (no edits) is an explicit reply, so the pane can clear its spinner

## Problem

Since ariadne#437 / pair#664 retired the legacy review handoff, an agent that answers a review round with no edits writes no round artifact. In pair's review pane the awaiting spinner and the Alt+c pending block then stay up until the next send. The legacy path cleared them on an empty handoff; #437 defines no empty-round reply. Found by pair:2 closing pair#664 (2026-10-18).

## Spec

An empty round is a valid, explicit reply: the agent (xx-fix / the review skill) always writes the round reply, with an empty records list when it has no edits (with its explanation or questions in the usual place), and the protocol document says so. pair's pane treats a reply with zero records as a completed round: clear the spinner and the pending block, commit nothing. Paired pair issue for the pane side.

## Done when

- Every completed workbench review calls apply, including `{"records":[]}` when
  there are no edits; explanations/questions remain in the session reply.
- Empty replies never create agent commits, including older-Pair success without
  Round/Body or recovery after a lost empty-apply reply. Changed replies still
  commit returned nonempty Body verbatim under the existing identity guards.
- Red/green skill-contract tests and observed agent scenarios cover the empty,
  real-edit and older-Pair reply cases. Protocol-target wording and the real
  spinner/pending-block pane check are owned by pair#690, as approved by the hub.

## Plan

- [x] Red-first test the no-edits producer and empty-success commit decisions in
  the existing reviewexport skill-contract tests, plus a read-only skill scenario.
- [x] Require an apply on every completed workbench review, including
  `{"records":[]}`. Keep explanations/questions in the session reply. Distinguish
  changed replies (commit returned Body verbatim) from empty replies (no commit),
  preserving existing hash, context, operator-lock and uncertain-result guards.
- [x] Update the ariadne workflow map; verify focused contracts and exported
  substrate integrity, then dispatch the recorded suite before close review.
  A passing final-code suite still precedes landing.
- [x] Record the peer hand-off and prepare independent producer landing; the
  binary owns close review, landing and archive, followed by the hub report.


## Log

### 2026-10-18
- 2026-10-18: closed — Observed baseline skipped empty apply; updated fresh agent sends records [] without committing, commits nonempty returned Body for a real edit, and does not retry or commit older-Pair empty success. Contract guards red then green; full reviewexport/processmanual pass. BR-1 current Done when now matches approved producer scope; original protocol/live criteria retained in Revisions under pair#690. Final-code recorded suite remains queued and must pass before landing.; review verdict: SHIP

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

- Final-code recorded suite attempt `654d955b4578efdbe4afec235805493c` passed:
  12 shards over 1,658 cmd/sdlc tests and 70 other packages, every exit 0 and zero
  failures, wall 645.7s after a 3,402s machine-queue wait. Log SHA-256 prefix
  `9de38c524871`, recorded as sdlc-local/suite. No retries or gate waivers.

## Revisions

### 2026-10-18 — producer scope and peer acceptance ownership

Hub approved the producer-only boundary. The original protocol-target wording
and live pane no-edits Done-when steps belong to pair#690, owned by pair:2; that
owner will log the hand-off and verify the pane after its implementation lands.
This issue proves that the shared producer always sends an empty apply and never
commits an empty reply; it may land before the peer engine/pane change. Until
then the old engine harmlessly accepts the empty apply without publishing a round.
No peer source or protocol target is edited here.

### 2026-10-18 — checkpoint wording

Plan rows now distinguish completed implementation/verification preparation from
the close/landing gates that necessarily follow this checkpoint. Recorded suite
is dispatched on the pushed implementation; close may overlap its queue under
the worker contract, with an actual pass required before landing. No acceptance
step or gate is waived; peer acceptance ownership remains as approved above.

### 2026-10-18 — synchronize the current acceptance contract (BR-1)

The close review correctly found that the current Done-when list retained peer
criteria even though the approved scope revision transferred them. Updated all
three current rows to the producer's empty submission, no-empty-commit rule,
changed-reply control and regression evidence. The original cross-repo criteria
were: "The protocol target states that an empty reply completes a round" and
"With pair's side landed, a live round where the agent makes no edits clears
the spinner." They remain acceptance for pair#690, not waived or claimed here.
The hand-off is recorded above and was sent to its owner. This is acceptance
bookkeeping only; the queued final-code suite remains valid and unchanged.
