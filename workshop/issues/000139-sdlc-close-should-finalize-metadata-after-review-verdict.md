---
id: 000139
status: working
deps: [ariadne#147]
target: agent-binary-handoff-schema
github_issue:
created: 2026-06-29
updated: 2026-06-30
estimate_hours: 0.63
started: 2026-06-30T12:11:03-07:00
---

# sdlc close should finalize metadata after review verdict

## Problem

`sdlc close` currently mutates the issue file before the boundary review has
returned. In pair#84, the first close attempt flipped the issue to `status:
done`, wrote `actual_hours`, and appended a "closed" log line, then the boundary
review returned `REWORK`.

That leaves the repo in an awkward intermediate state:

- the issue says `done` while the boundary says "do not cross yet";
- rerunning close requires `--no-reclose-guard`;
- repeated close attempts append stale close log lines that the operator must
  clean up manually;
- the close bookkeeping itself becomes tangled with the code-review loop.

The status flip should be the last successful act of close, not something that
happens before the gate has established that the boundary can be crossed.

## Spec

Redesign `sdlc close` as a two-phase operation:

1. Validate close form and compute the proposed close metadata in memory:
   `actual_hours`, verification line, planned status/log changes, and any atlas
   or project requirements.
2. Run the boundary review against the implementation window while the issue is
   still logically in progress. The judge should still see enough issue content
   to check plan/done-when state, but the repo should not be committed to
   `status: done` before the verdict is known.
3. Only after an acceptable verdict (`SHIP`, or `FIX-THEN-SHIP` when policy
   allows) apply the close metadata and produce the final close bookkeeping
   commit/trailer.

The final close bookkeeping commit is allowed to bypass another close review:
it records the result of the gate rather than introducing new implementation
surface. This is the "commit on its own" shape: code commits are reviewed by the
boundary review; the final close commit records status/actual/verdict and should
not recursively trigger another review.

If the boundary review returns `REWORK`, close should leave the issue in
`working` (or restore the pre-close issue file), print the verdict and next
action, and avoid appending a misleading "closed" log entry.

## Done when

- [x] A failed/REWORK boundary review does not leave the issue at `status: done`.
- [x] Rerunning close after fixing review findings does not require
      `--no-reclose-guard`.
- [x] Close appends exactly one final "closed" log line for the successful close.
- [x] The final close bookkeeping commit can carry `Review-Verdict:` trailers
      without causing a recursive review requirement.
- [x] Tests cover `SHIP`, `FIX-THEN-SHIP`, `REWORK`, judge failure, and
      `--no-judge` close flows.

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: smaller-go-module   design=0.1 impl=0.2
item: smaller-go-module   design=0.1 impl=0.2
design-buffer: 0.15
total: 0.63
```

Two extends of `cmd/sdlc/close.go` (+ `milestoneclose.go`): (1) split `runClose`
at its existing compute/write boundary into `computeClose`/`applyClose` (behavior-
preserving); (2) `closeVerdictOutcome` (derived from `vocab.Verdict()`, #147) +
`reviewThenFinalize` (three outcomes) + rewire full-issue close AND fold in
milestone-close. Design pre-resolved by the durable plan → reduced design + +15%
buffer; impl at the v3.1 40%-scaled smaller-go-module top.

Detailed design + TDD breakdown: `workshop/plans/000139-close-finalize-after-verdict-plan.md`.

## Plan

- [x] Inspect the current close mutation order in `cmd/sdlc/close.go`.
- [x] Define the acceptable verdict policy for finalizing close metadata
      (3 outcomes, derived from `vocab.Verdict()`; only REWORK reworks, unknown halts).
- [x] Introduce a close plan/result object (`closeResult`) computed before writes.
- [x] Move issue-file mutation until after boundary review success (`applyClose`
      fires only on a finalizing verdict; success messages emitted post-write).
- [x] Reviewer reads the un-mutated `working` tree (real Plan/Done-when) — no
      pre-review status flip.
- [x] Regression tests: REWORK/unknown leave `working`; rerun produces one clean
      close line, no `--no-reclose-guard`; milestone-close mirrored.

## Log

### 2026-06-29

- Created from pair#84 dogfooding: `sdlc close` flipped pair#84 to `done`
  before the boundary review returned `REWORK`, forcing re-close guard bypasses
  and manual cleanup of stale close log lines.

### 2026-06-30

- **Parked, blocked on #147.** Design + durable plan complete
  (`workshop/plans/000139-close-finalize-after-verdict-plan.md`: two-phase
  compute→review→apply for close + milestone-close, three verdict outcomes
  finalize/rework/halt). During design review the operator flagged that the
  policy's halt-on-`unknown` is unsound while `unknown` is *frequent* — the root
  cause is the unstructured verdict handoff (free-text stdout regex-parsed), which
  #147 fixes with a CUE-modeled, schema-validated structured handoff. Do #147
  first so this issue's close reads a robust verdict; then resume here.

- **Resumed + implemented (#147 merged).** `runClose` split at its existing
  compute/write boundary into read-only `computeClose` (all gates, composes new
  issue/project text into a `closeResult`, writes nothing) + `applyClose` (writes +
  calibration ledger). The success messages ("flipped → done", etc.) are collected
  in `closeResult.appliedMsgs` and emitted **only** by `applyClose` post-write — so
  a REWORK never prints a write that didn't happen (a plan-quality gate finding).
  Both full-issue close (`runCloseWithReview`) and milestone-close
  (`runMilestoneClose`) reorder to **computeClose → review-against-the-un-mutated-tree
  → finalize** via the shared `reviewThenFinalize`. `closeVerdictOutcome` derives
  from `vocab.Verdict().IsFinalizing/IsBlocking` (#147 single-source): finalizing →
  apply; REWORK → not finalized (stays `working`, non-zero, "fix + re-run", no
  `--no-reclose-guard` on rerun); unknown / dispatch-error → **halt** ("UNEXPECTED —
  stop, consult a human"). `--no-judge` finalizes, handled before dispatch so only a
  genuine dispatch error reaches the halt path. Done-when 4 is structural (close
  never commits; the review fires only inside the verb).

  Implementation delegated to a context-carrying fork (the plan + gate feedback);
  fresh-eyes reviewed the diff (the compute/apply seam, the 3-outcome dispatch, the
  `--no-judge`-before-dispatch ordering, and the REWORK/unknown/rerun + milestone
  tests). Verification: `go test ./cmd/sdlc/... ./pkg/vocab/` all pass. New
  `close_finalize_test.go`: `TestCloseVerdictOutcome` (pure);
  `TestRunCloseWithReview_{REWORK_DoesNotFinalize,Unknown_Halts,RerunAfterREWORK}`
  (REWORK asserts no `status: done`, no `closed` line, no `actual_hours`, no
  "flipped" printed; rerun finalizes with exactly one close line, no
  `--no-reclose-guard`) + the milestone REWORK/SHIP mirror. Existing close tests
  stay green (extract was behavior-preserving). This close dogfoods #139 + #147.
