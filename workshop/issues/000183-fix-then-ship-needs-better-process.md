---
id: 000183
status: working
deps: []
github_issue:
created: 2026-07-16
updated: 2026-10-10
estimate_hours:
card_mirror: 'a4d8d27e2dcab2b891cb856f183bafb382d30f47' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-10T16:53:19-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "985aba4d", done: "02bb6d4b"}
---

# FIX-THEN-SHIP needs better process

right now, it seems when sdlc instruct agent to FIX-THEN-SHIP for non-critical things, after the fix, agent has to run milestone-close with --no-judge. this leaves audit trail that milestone-close is bypass, which is not strictly the case. I'm thinking about the following scheme, to improve the process. 

1. sdlc milestone-close (or other gate) will generate a hash based on the files it think needs to change. it keep track of that in its state. let's say one content hash per file. 

2. when the agent calls again with milestone-close --fixed-to-ship, the sdlc gate will behave differently. it would not launch the configured gates in milestone-close. rather, it will check if the files it thought need to be changed, has changed hash. we can probably just check if 75%+ files changed, if yes, let it pass. make the 75% a configurable knob, which controls which agent (coding agent, or review agent) we trust more.

## Problem

A finding fixed after a finalizing review has no proportionate gate. The publish gate refuses any code delta past the close's reviewed patch, and the only way back is a **full** re-close: `sdlc close` reviews the whole branch patch again. Agents reach for `--no-judge`, or improvise `sdlc judge milestone-review --base …`, to avoid a full review of a two-line rename (evidence pensive A6). A merge of main with a code conflict hits the same wall (#320 layer 3). And on the quick flow, `done-when-fresh` refuses whenever a review-fix round appends the `## Revisions` entry the convention asks for, so agents bypass it or pad `## Done when` (F3).

## Spec

**`sdlc close --issue N --fixed-then-ship`: the narrow re-close.** It runs after a finalized whole-issue close review (SHIP or FIX-THEN-SHIP) whose fixes, or a resolution of a conflict with main, landed after it.

- **Window:** the interdiff since that review. The reviewed head the ledger stamped on the last finalized whole-issue round is replayed onto today's main (`gitx.RebasedReviewedBase`, as #304's milestone interdiff does). The diff to HEAD is exactly the fixes plus any conflict resolution; main's own changes never enter it.
- **Review:** a new `fix-review` recipe. The reviewer disposes of each prior finding (the ledger's prior-round block), checks that the interdiff introduces no Critical or Important defect, judges every conflict resolution it lists, and treats a scope change in `## Spec`/`## Revisions`/`## Done when` as REWORK, which sends the issue to a full close. The verdict and ledger work as they do for any boundary round. A finalizing verdict publishes codecomplete bound to a new evidence commit, so the publish gate's anchor advances.
- **Refusals:**
  - No finalized whole-issue review round to anchor on → refuse, naming plain `sdlc close`.
  - Quick flow: the shell is still measured over the **branch patch**. If the fixes push the issue outside the shell, refuse, because the full review is due.
  - **The interdiff carries a code-conflict resolution** (operator condition, 10-10): refuse unless `--full-suite '<evidence>'` records a full-suite run. Only the review narrows. The evidence goes into the close's `## Log` line.
- **The atlas gate** is measured over the branch patch under `--fixed-then-ship`. The prior close already satisfied it, and a fix delta should not demand a fresh atlas edit.
- **The publish gate's refusals** (code after close, a code conflict) name `sdlc close --issue N --fixed-then-ship` as the next action, so this is #320's layer-3 hint (TL 10-10: whichever lands second wires it).
- **FIX-THEN-SHIP protocol text** (close): "fixes after the close commit → `--fixed-then-ship`" replaces "re-run close". The unstarted-receipt path (`issue recovery reconcile`) stays the default for fixes the verdict already sanctions; `--fixed-then-ship` supersedes that receipt when the agent wants the fixes checked.
- **Milestone-close is unchanged.** Fixes after a milestone close already fall in the next boundary's interdiff (#304), so the help text says that and adds no flag.

**F3: done-when-fresh becomes a reviewer check whenever a review runs.** The deterministic check cannot tell a scope reframe from a fix record; a reviewer reading the diff can (TL 10-10, consistent with "code is the eventual source of truth").

- **When a review will run:** a failing `DoneWhenFresh` no longer refuses. The review prompt says the contract moved since change-code while `## Done when` did not. It tells the reviewer to judge whether Done-when still covers the work, to raise a blocking finding if it doesn't, and to state the judgement on a `Done-when:` line, which lands in the sidecar.
- **The close log line** records `done-when: contract moved since change-code; judged by the <verdict> review`.
- **No review:** under `--no-judge`, the deterministic refusal stands (`--no-done-when-fresh` still waives it). A review that did not run never finalizes (requirement 7), so it cannot pass the check.

**Dropped from the original note:** the per-file hash and 75% knob. The interdiff review checks that the fixes address the findings, which a hash count cannot.

## Done when

- `sdlc close --fixed-then-ship` after a finalized close review dispatches the `fix-review` recipe over the interdiff since the reviewed head, never the branch patch. Tested with a main merge in between, so main's changes stay out.
- On a finalizing verdict it publishes codecomplete bound to a new evidence commit, and `sdlc merge`'s publish gate then passes. On REWORK nothing is published.
- Refusals, each tested:
  - no finalized whole-issue round;
  - the quick flow outside its shell over the branch patch;
  - a code-conflict resolution in the interdiff without `--full-suite`.
- The publish gate's code-delta and code-conflict refusals name `sdlc close --issue N --fixed-then-ship`.
- Quick flow: a contract that moved with Done-when unchanged no longer refuses when the review runs. The prompt carries the done-when note, and the log line records the judgement. Under `--no-judge` it still refuses.
- Help (close and milestone-close), the recovery catalog's close entry, and the atlas all describe the new path. The `fix-review` recipe is cataloged by `process-manual`.

## Plan

Durable plan: `workshop/plans/000183-fix-then-ship-needs-better-process-plan.md`. Single pass, one close.

- [ ] Ledger: `LatestFinalizedAt(l, boundary)`; `fix-review` rounds are not full rounds
- [ ] Window planner: `FixedThenShip` fact → interdiff since the last whole-issue review
- [ ] `fix-review` prompt recipe + category
- [ ] close: `--fixed-then-ship` / `--full-suite` flags, refusals, branch-patch atlas and shell measurement, recipe selection
- [ ] F3: done-when-fresh deferral to the review, prompt note, log line
- [ ] Publish gate refusal text, FIX-THEN-SHIP protocol text
- [ ] Help, recovery catalog, atlas, process-manual

## Log

### 2026-10-10
- Claimed at ariadne:4 (TL batch 2). TL approved (10-10): #183 owns #320's layer 3; F3 becomes a reviewer check, with the conditions recorded in ## Revisions.

## Revisions

- 2026-10-10 — Scope split with #320 (TL decision). #183 owns the narrow re-close (`--fixed-then-ship`), which is #320's layer 3. #320 keeps layers 1–2 (union merge; non-code conflicts pass). Whichever lands second points #320's code-conflict refusal at `sdlc close --fixed-then-ship`, and #183 wires it if #183 lands second. Operator condition: an interdiff carrying a code-conflict resolution still requires full-suite evidence (`--full-suite`); only the review narrows. F3 conditions: the reviewer's done-when judgement is recorded (log line and sidecar), and a review that did not run never passes it.
