---
id: 000410
status: working
deps: []
github_issue:
created: 2026-10-16
updated: 2026-10-17
estimate_hours:
card_mirror: '7eab5a1cf3d499623a8849c46eb540adadf0ecea' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-17T19:34:20-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "8f2289a4", done: "81a913b1"}
---

# tl-worker/tl-lead: a state design table for stateful plans (actions, failure modes to few visible states, one recovery each, explicit loss, tradeoff to TL/operator)

## Problem

Over-engineered state designs (pair#556/#584, #587) and under-engineered recovery have one root: plans for state that
users or automation can change don't state their action set, their failure modes, the recovery for each, or what state
can be lost. The operator (2026-10-16): no model is immune; adjust the prompt. A design that changes such state must
present the tradeoff explicitly, to the TL and often the operator, because it is almost always a tradeoff.

## Spec

tl-worker's planning rules (and tl-lead's plan review) gain a **state design table** for any plan that adds or changes
state a user or another system can act on:
1. **The actions:** what the user or automation can do (relaunch, resume, reboot, cancel, edit, retry…).
2. **The failure modes:** internal states, honestly including combinations, mapped onto a **small set of visible
   states**.
3. **Each visible state's one recovery action,** chosen from the action list.
4. **What can be lost** in each recovery (state, work, history), stated explicitly.
5. **The tradeoff summary** for the TL (and the operator when user-facing): what's covered, what isn't, and why.
The TL brings user-facing tradeoffs to the operator before code. Phrase-guarded in the help tests.

## Done when

- `tl-worker --help` and `tl-lead --help` carry the table and the escalation rule; help tests guard the phrases.

## Plan

- [x] Add phrase guards first in both help tests for the shared state-design table, one worked example, and TL/operator escalation; confirm they fail against current help.
- [x] Add a compact table template and example to `pkg/tlrules/planning.md`, which both binaries embed via `{{PLANNING}}`. Require action enumeration, reachable internal and failure combinations mapped to few visible states, one listed recovery per visible state, possible loss, and covered/uncovered tradeoffs sent to TL (operator for user-facing effects).
- [ ] Run focused tests and verify both built help outputs; record the exact-head suite, close review, and merge through the normal gates.
- [x] Side quest from the hub: test and correct `tl-lead` WATCHING's event list to pair#574's actual start, park, relaunch, recovery, needs-relaunch, and restart-summary notices, excluding exit until pair#657. Commit this correction separately.

## Log

### 2026-10-16

2026-10-17: The shared `pkg/tlrules/planning.md` is the one rule source for both help surfaces (ARCH-DRY); `cmd/tl-worker/testdata/rules.tsv` and `cmd/tl-lead/main_test.go` guard their rendered outputs. A separate help paragraph per role would drift. The small worked example will be a check process that dies after starting: visible state “needs retry,” recovery “rerun the exact head,” possible loss “partial output if not saved.” The added #433 event-list correction is tracked as a side quest and will get its own commit.

The hub approved the plan with the table held to about ten lines plus the example so workers will read it.

Both new help-output rule sets failed against the old shared planning prose, then passed after the ten-line table and example were added. The side-quest test separately failed on the premature exit wording, then passed after the event list and atlas were corrected. `go test ./cmd/tl-lead ./cmd/tl-worker ./pkg/tlrules -count=1` passed.

## Side quests

- Correct pair#574 lifecycle notice list before exit ships in pair#657 — commit 3bbce8666ac8edd79c22855d2de5cefe3e9b0396.
