---
id: 000453
status: open
deps: []
github_issue:
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
collaboration-mode: collaboration
card_mirror: '67812f0d8ea45f896114a9dd14fc5e25f70fb172' # card fields mirrored from issue-cards; edit via sdlc
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

- [ ]

## Log

### 2026-10-18
