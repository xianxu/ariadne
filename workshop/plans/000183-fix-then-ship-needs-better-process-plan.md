# `--fixed-then-ship` Narrow Re-close Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `sdlc close --issue N --fixed-then-ship` reviews only the interdiff since the last finalized whole-issue review (post-close fixes plus any conflict resolution with main), with a fix-verification recipe, so neither a post-close fix nor a code-conflict merge forces a full re-close. Separately, the quick flow's `done-when-fresh` defers to the review whenever one runs (F3).

**Architecture:** The pieces are mostly #304's. The ledger already stamps `Reviewed` (the head a finalized round read) per round, and `gitx.RebasedReviewedBase` replays a reviewed head onto today's main. Today only milestones use that interdiff. `--fixed-then-ship` adds one fact to the pure window planner (`windowFacts.FixedThenShip`), which plans the interdiff since the last finalized **whole-issue** round. close then dispatches a new `fix-review` recipe over that window. Finalization is unchanged: the evidence commit and the card binding are the same code, so the publish anchor advances. Two gates keep their meaning by being measured on the **branch patch** under the flag: the atlas check and the quick-flow shell. F3 is a policy change in `checkQuickDoneWhen` plus one prompt note.

**Tech Stack:** Go (`cmd/sdlc`), embedded markdown prompts (`internal/judge/prompts`), golden tests.

---

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `LatestFinalizedAt` | `cmd/sdlc/internal/gatestate/reviewed.go` | new |
| `windowFacts.FixedThenShip` / `planReviewWindow` | `cmd/sdlc/reviewwindowplan.go` | modified |
| `fullRoundIn` | `cmd/sdlc/closeflow.go` | modified |
| `judge.FixReview` category + `prompts/fix-review.md` | `cmd/sdlc/internal/judge/prompts.go` | new |
| `PromptInput.ReviewNotes` / `{{REVIEW_NOTES}}` | `cmd/sdlc/internal/judge/prompts.go`, `boundary-tail.md` | new |
| `fixedThenShipRefusal` | `cmd/sdlc/closefixed.go` | new |
| `classifyPublishDelta` | `cmd/sdlc/publishgate.go` | modified |
| `formatFixThenShipProtocol` | `cmd/sdlc/close.go` | modified |
| `doneWhenDecision` | `cmd/sdlc/closeflow.go` | new |

- **LatestFinalizedAt(l, boundary)**: the newest round AT `boundary` carrying `Reviewed` (the head and the round's verdict-independent identity). `LatestReviewed` answers "outside a boundary" for milestones; this answers "the last finalized review of this boundary" for the whole-issue re-close. Both are one loop over `l.Rounds`, and the shared predicate (`r.Reviewed != ""`) stays inline: two 6-line functions, not an abstraction (ARCH-DRY judged: different questions).
- **windowFacts.FixedThenShip**: when set (with `Milestone == ""`), the planner takes the interdiff case on `Reviewed`/`Rebased` from the whole-issue boundary. If no finalized round exists, it returns `windowNone` with `Note` naming the reason, and close turns that into a refusal. `Conflicted` flows through unchanged; it is the code-conflict signal.
- **fullRoundIn**: a `fix-review` round is not a full review. It reviewed an interdiff, so it must not make a quick issue's later close "sticky full" (#231), and it must not count as one either.
- **FixReview**: `fix-review.md` = `{{CODE_REVIEW_BODY}}` orientation + a fix-verification focus section + `{{ARCH_BLOCK}}` (quick-flow markers only: the window is a fix delta) + `{{BOUNDARY_TAIL}}`. `ReviewMarkers(FixReview)` returns `QuickMarkers()`. Added to `AllInjectedCategories` so process-manual and the goldens pin it. Not in `AllCategories` (it isn't a standalone `sdlc judge` check).
- **ReviewNotes**: free text rendered into the boundary tail before "Review window:". It carries (a) the conflict-resolution list and (b) the F3 done-when note. Empty → the line renders as nothing (goldens for existing recipes gain only the empty-token position, checked byte-for-byte).
- **fixedThenShipRefusal(facts)**: pure. Inputs: the planned window, whether the quick-flow decision upgraded, whether any `Conflicted` path has code surface, and the `--full-suite` value. Returns the refusal text, or "". It is unit-tested over every combination.
- **classifyPublishDelta**: both refusal arms (code delta, conflict) name `sdlc close --issue <N> --fixed-then-ship`. The conflict arm adds `--full-suite '<evidence>'` when a conflicted path has code surface. Unresolvable keeps the plain re-close (there is no reviewed head to diff from).
- **doneWhenDecision(stale, judgeRuns, skip)**: pure → `{refuse | defer-to-review | waived | fresh}`. Stale with the review running → defer; stale with `--no-judge` → refuse unless waived.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `gatherWindowFacts` | `cmd/sdlc/reviewwindowplan.go` | modified | git + ledger read |
| `computeClose` / `runCloseWithReview(Locked)` | `cmd/sdlc/close.go` | modified | close orchestration |
| `checkQuickDoneWhen` / `closeFlowStep` | `cmd/sdlc/closeflow.go` | modified | flow record |
| `boundaryReviewDispatchOptions` | `cmd/sdlc/milestoneclose.go` | modified | prompt build |

- **gatherWindowFacts** gains a `fixedThenShip bool` argument (threaded through `planBoundaryWindow` and `resolveReviewWindow`). It reads the ledger at boundary `""` via `LatestFinalizedAt`. Tests use real git in temp repos (the existing `reviewwindowplan_test.go` / `milestoneclose_test.go` fixtures), not mocks.
- **computeClose**: under `--fixed-then-ship` it plans two windows, the interdiff (review) and the branch patch (atlas gate + quick-flow shell). It applies `fixedThenShipRefusal` before any review. The log line carries `fixed-then-ship` and the `--full-suite` evidence. ARCH-FUNERAL: creates nothing durable beyond what a close already writes (one ledger round, one sidecar, one evidence commit), all with existing archive paths.

## Tasks

### Task 1: ledger + recipe accounting
**Files:** `internal/gatestate/reviewed.go` (+`_test.go`), `closeflow.go` (+ test)
- [ ] Failing test: `LatestFinalizedAt` returns the newest `Reviewed` round at `""`, skipping milestone rounds, `BoundaryAll`, and unstamped (REWORK) rounds; `ok=false` on none.
- [ ] Implement: loop newest-first, `r.Boundary == boundary && r.Reviewed != ""`.
- [ ] Failing test: `fullRoundIn` over a single `fix-review` round → false.
- [ ] Implement: exclude `judge.FixReview` alongside `SmallDiffReview`.
- [ ] `go test ./cmd/sdlc/internal/gatestate/ -run LatestFinalized` and `go test ./cmd/sdlc/ -run FullRound`.

### Task 2: window planner
**Files:** `reviewwindowplan.go`, `reviewwindowplan_test.go`, `milestoneclose.go` (`resolveReviewWindow`, `boundaryWindowBase` signatures)
- [ ] Failing table tests on `planReviewWindow`: FixedThenShip + Reviewed + Rebased → `windowInterdiff`, Base=S, HumanBase=H_r, Conflicted carried; FixedThenShip with no Reviewed → `windowNone`, Note "no finalized close review to fix against"; FixedThenShip + RebaseErr → `windowNone` with the error as Note (never the branch-patch fallback: a fix review over the whole branch would be mis-scoped).
- [ ] Implement: in `planReviewWindow`, before `if f.Milestone == ""`, a `case f.FixedThenShip:` arm. In `gatherWindowFacts`, when `fixedThenShip`, read `LatestFinalizedAt(l, "")` and replay it as the milestone arm does (factor the replay into `replayReviewed(f *windowFacts, sha)`, shared by both arms; ARCH-DRY).
- [ ] `formatReviewWindow`: interdiff `since` reads "the close review" when `PriorBoundary == ""`.
- [ ] Integration test (real git): branch, close-review round stamped at H_r, fix commit, merge of a main commit → window diff names only the fix's paths.

### Task 3: `fix-review` recipe
**Files:** `internal/judge/prompts.go`, `prompts/fix-review.md`, `boundary-tail.md`, `golden_test.go` + `testdata`, `cmd/sdlc/processmanual` catalog if it enumerates categories
- [ ] Add `{{REVIEW_NOTES}}` to `boundary-tail.md` immediately before `Review window:`; `PromptInput.ReviewNotes`; regenerate goldens and confirm existing recipes change only by the empty token.
- [ ] Write `fix-review.md`. Its focus section, verbatim intent: this window is ONLY what changed since a finalized close review; (1) dispose every prior finding; a finding whose fix is absent is `not-addressed`; (2) raise new findings only for defects the interdiff introduces; (3) every conflict resolution listed under review notes is part of the patch, and you judge it for correctness against both sides; (4) if the interdiff changes `## Spec`/`## Revisions`/`## Done when` beyond recording the fix, that is a scope change: verdict REWORK, with the finding "scope changed after close; run a full `sdlc close`".
- [ ] `FixReview` const, `Label`, `ReviewMarkers` → quick markers, `AllInjectedCategories`.
- [ ] `go test ./cmd/sdlc/internal/judge/ -run 'Golden|Prompt'`.

### Task 4: close wiring
**Files:** `close.go`, new `closefixed.go` (+`closefixed_test.go`), `milestoneclose.go` (dispatch passes `ReviewNotes`)
- [ ] Flags: `--fixed-then-ship` (bool), `--full-suite '<evidence>'` (string). `--fixed-then-ship` with `--no-judge` → refuse (the flag IS a review). `--full-suite` without `--fixed-then-ship` → refuse.
- [ ] `fixedThenShipRefusal` unit tests: no window; quick upgrade; code conflict without suite; non-code conflict passes; code conflict with suite passes.
- [ ] `computeClose`: under the flag, plan the review window with `fixedThenShip=true`; atlas + `closeFlowStep` measure the branch-patch window (`planBoundaryWindow(..., false)`). Apply the refusal. Log line: `closed (fixed-then-ship[, full suite: <evidence>]) — <verified>`.
- [ ] `runCloseWithReview`/`Locked`: base/head from the fixed window; `Category = judge.FixReview`; `ReviewNotes` = conflict list (+ F3 note from Task 5). The dry-run shows the same.
- [ ] Integration test (tracker fake used by `closetracker_test.go`): SHIP close → fix commit → `close --fixed-then-ship` with a fake agent SHIP → card codecomplete bound to a new evidence commit; `runPublishGate` passes. REWORK → card unchanged.
- [ ] Integration test: after a FIX-THEN-SHIP close (unstarted receipt), `--fixed-then-ship` supersedes the receipt.

### Task 5: F3
**Files:** `closeflow.go` (+ test), `internal/flow/donewhen.go` (message only), `close.go` log line
- [ ] `doneWhenDecision` table test (stale × judge runs × waived).
- [ ] `checkQuickDoneWhen` returns the decision; on defer it emits an info line and closeResult carries `doneWhenNote`. That becomes `ReviewNotes` text: "The contract (`## Spec`/`## Revisions`) changed since change-code while `## Done when` did not. Judge whether Done-when still covers what this diff builds; if not, raise an Important finding. State it on one line: `Done-when: current — <why>` or `Done-when: stale — <why>`." It is also appended to the close log line as `done-when: contract moved since change-code; judged by the <verdict> review`, at finalize (so REWORK writes nothing, #139).
- [ ] Test: quick issue, Spec revised, Done-when unchanged, fake agent SHIP → close finalizes, log line carries the judgement; same with `--no-judge` → refuses.

### Task 6: refusal + protocol text
**Files:** `publishgate.go` (+ test), `close.go` `formatFixThenShipProtocol` (+ test), `internal/processmanual/gatesig.go` if a pinned refusal substring changes
- [ ] `classifyPublishDelta` tests for the new next action in both arms. Keep the `publishGateRefusal` prefix (gatesig keys on it).
- [ ] Protocol: "Only if fixes must land AFTER the close commit: `sdlc close --issue N --fixed-then-ship` reviews just those fixes and advances the anchor (doc-only deltas pass on their own)."

### Task 7: docs surface
**Files:** `helptext/close.md`, `helptext/milestone-close.md`, `internal/recovery/catalog.go` close entry (+ named proof test), `atlas/workflow/sdlc-binary.md`, `atlas/workflow/pre-merge-checks.md` (publish gate)
- [ ] close help: a `FIXED-THEN-SHIP (#183)` section (window, recipe, refusals, `--full-suite`); rewrite POST-VERDICT and quick-flow Done-when paragraphs for F3.
- [ ] milestone-close help: fixes after a milestone close are covered by the next boundary; no `--fixed-then-ship` there.
- [ ] Catalog: close Effects mention `--fixed-then-ship`; add the Task 4 test names as Proofs.
- [ ] Atlas sections updated; `atlas/index.md` unchanged (no new file).

### Verification
- [ ] Targeted `go test ./cmd/sdlc/... -run '<new tests>'` per task; `make test` once before close.
- [ ] Live check: this issue's own close runs plain `sdlc close`. If its review returns FIX-THEN-SHIP, the post-fix step uses the new `--fixed-then-ship` path from the freshly built binary (`go build -o $TMPDIR/sdlc ./cmd/sdlc`).
