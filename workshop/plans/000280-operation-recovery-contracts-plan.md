# Operation Recovery Contracts Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Every MVP workflow verb publishes what it does when repeated, interrupted or lost. Each guarantee names the test that proves it. An agent can learn the contract and verify outcomes through read-only queries instead of re-deriving workflow semantics.

**Architecture:**
- **Registry.** One registry, `internal/recovery`, like `processmanual.GateCatalog`, holds one contract per verb:
  - its class: read-only, duplicate-safe refusal, convergent retry or non-repeatable;
  - effects;
  - evidence (`sdlc issue show N --json` fields and commands);
  - preconditions;
  - behavior on repeat;
  - lost-response action;
  - when guarantees end;
  - proofs, each a claim plus the test names that demonstrate it.
- **Help.** The registry renders into each verb's `--help` through a `{{RECOVERY}}` placeholder (`{{RECOVERY <verb>}}` where one page hosts several verbs). It also renders into one help topic, `sdlc help recovery`, as recorded in the operator's decision. That topic carries the agent guidance and an executable scheduling example, and cross-links `sdlc issue recovery reconcile`.
- **Contract tests** fail when a verb lacks an entry, a named test does not exist, a page drops the placeholder, or the example stops working.
- **Gap fixes.** The verbs that today return a bare "publication outcome uncertain" gain the same "rerun; the card decides" guidance as reclaim, plus lost-acknowledgement tests.

**Tech Stack:** Go (`cmd/sdlc`, new `internal/recovery`). Real-git tests on the existing harnesses (`closeReady`, `reclaimFixture`, `raceBuiltBinary`, `observeIssue`).

**Operator decisions (2026-10-02, issue Log):**
- A registry rendered into help, with a contract test.
- An executable scheduling example.

## Non-goals

- No new recovery mechanism. Card CAS, receipts, `issue recovery reconcile`, landing observations and #279 observations stay as they are. The contracts *describe and test* them.
- No exactly-once claims for arbitrary agent instructions. The page states that claim reserves an issue, and that later code edits are not made idempotent by it.
- Legacy (non-slot) `pr` stays non-repeatable. The contract says so (re-run with `sdlc pr` only after checking for an open PR) instead of this issue changing its behavior.
- No Couch, scheduler or skill. pair#362 consumes this page.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `recovery.Contract`, `Class`, `Proof` | `cmd/sdlc/internal/recovery/contracts.go` | new |
| `recovery.Catalog` (the per-verb entries) | `cmd/sdlc/internal/recovery/catalog.go` | new |
| `recovery.Section(verb)`, `recovery.Page()`, `recovery.JSON()` | `cmd/sdlc/internal/recovery/render.go` | new |
| `recovery.Example` (the scheduling example's steps) | `cmd/sdlc/internal/recovery/example.go` | new |

- **Contract** fields: `Verb`, `Class`, `Effects`, `Evidence`, `Preconditions`, `Repeat`, `LostResponse`, `Ends`, and `Proofs []Proof{Claim, Tests []string}`.
  - A claim with no test renders as **unproven**. That covers the issue's "unproven guarantees are explicitly unknown".
  - **Class** is a closed set, validated: `read-only`, `duplicate-safe-refusal`, `convergent-retry`, `non-repeatable`.
- **The required verb set is derived, not listed.** It is every command in the cobra tree annotated `markMutatingCommand`, plus the workflow verbs that manage their own lock or take none (start-plan, change-code, milestone-close, close), plus `issue show` (the evidence query).
  - Each one needs a catalog entry, or an entry in `recovery.Exempt` with a reason. Examples: the project verbs, `migrate`, `issue migrate`, `fetch`, and the retired `issue publish`.
  - So a new mutating verb fails the contract test until someone decides its contract. The MVP entries include push, `issue sync`, `issue move-detail` and `issue new` alongside claim (with `--adopt` and relocation repair), reclaim, start-plan, change-code, milestone-close, close (with `issue recovery reconcile`), pr, merge, `issue set-status`, move and `issue show`.
  - Each entry's text and test names come from the explorer map in the issue Log, rechecked against code while writing.
- **Example** is a list of steps. Each step names an actor (coordinator or recipient), a command, the evidence to read (dot-paths into the observation JSON), the expected value, and what to do otherwise. It is both rendered into the page and executed by a test.
- **DRY rationale:** one source for the contract text, the help, the JSON and the test links, as the gate catalog is for gate flags.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `{{RECOVERY}}` placeholder | `cmd/sdlc/main.go renderLong` | modified | `recovery.Section(name)` |
| `sdlc help recovery` (a help topic) | `cmd/sdlc/recoverycmd.go`, `helptext/recovery.md` | new | the registry |
| `cardPublish` (one publication seam for single-card CAS verbs) | `cmd/sdlc/cardpublish.go` | new | `tracker.Repository.UpdateCardWithTrailers` |
| uncertain-publication guidance | `cmd/sdlc/claim.go`, `cardsetters.go` (set-status) | modified | `gitx.ErrPublicationUncertain` |

- **Uncertain-publication guidance.** One helper wraps `ErrPublicationUncertain` from a single-card CAS verb with "the change may or may not have published; rerun the same command — the card decides". It is used by claim, adopt, relocation, set-status **and reclaim** (replacing its inline message), single-sourced instead of written into each verb.
- **`cardPublish`.** It is the one injection seam those verbs publish through, so a lost-acknowledgement test injects once rather than through a package var per verb. `reclaimEffect` becomes a thin caller of it. A caller-guard test, like `TestReclaimIsOnlyOperatorInvoked`, pins its callers.

### Lifecycle and ordering

The registry is static data: it creates nothing durable, and adds no runtime state or ordering. The executable example runs only in tests, against disposable fixtures.

## M1 — The registry, its rendering, and the gaps it exposes

- [ ] `internal/recovery`: the types, the class set and validation. Pure tests check that every class is valid and every entry has an evidence line and a lost-response line.
- [ ] Write the catalog entries for all MVP verbs, rechecking each claim against the code.
- [ ] `{{RECOVERY}}` in each verb's helptext, plus `recovery.Section`. `TestNoCommandLongHasSurvivingPlaceholder` keeps passing.
- [ ] **Contract test `TestRecoveryContractsAreProven`:**
  - every command in the derived required set has exactly one entry or one exemption with a reason; there are no stale entries or exemptions;
  - every named test exists as a `func TestX` somewhere under `cmd/sdlc` (AST scan);
  - every verb page renders its section.
  - Mutation check: rename a referenced test and the contract test must go red.
- [ ] **Gap fixes:** the uncertain-publication helper and the `cardPublish` seam, used by claim, adopt, relocation, set-status and reclaim. New lost-acknowledgement tests:
  - claim: an injected uncertain error after the effect lands, then a rerun gives "already claimed";
  - set-status: the same shape, the rerun gives "already has that status".
- [ ] **Done-when coverage.** Each case class names its proof:
  - **race:** `TestClaimRaceHasExactlyOneWinner`, `TestAdoptRaceHasExactlyOneWinner`, `TestReclaimRaceHasExactlyOneWinner`.
  - **duplicate:** `TestVerbContractTable`, `TestClaimDryRunOwnerRepeatWritesNothing`, `TestRunMerge_ResumeMergedPR_FinishesCleanup`, and the move refusals.
  - **lost acknowledgement:** the two new tests, `TestReclaimStaleAndLostResponse`, `TestUpdateMany_LostAcknowledgmentIsUncertainWithoutReplay`, and `TestReconcileRetriesAnInterruptedCloseMirror`.
  - **reclaimed generation:** `TestReclaimTransfersResponsibility`, where the old owner's gates are refused.
  - **re-closed generation:** `TestPublishFlipLeavesAReclosedGenerationAlone`, `TestTrackerReCloseSupersedesAnUnstartedClose`, `TestNewestCloseRefusesAnOlderReviewThanTheCardsClose`.
  - **New tests where none exists:**
    - `TestCloseRerunAfterShipStartsANewGeneration`: a second close rebinds the card to a new token; the first evidence commit stays in history.
    - `TestLandingLeavesAReopenedCardAlone`: a card reopened after its close is not completed by a later landing of that close.
- [ ] M1 milestone-close.

## M2 — The recovery page and the executable scheduling example

- [ ] `sdlc help recovery`, a cobra help topic (no Run, so it collides with nothing; `sdlc issue recovery` keeps its name and both cross-link), with `helptext/recovery.md`. The page holds:
  - the class definitions;
  - the agent guidance: verify effects through read-only queries; roughly 30 s is a revisit heuristic, not proof of loss; retry only under the verb's contract; unknown or stale is not negative evidence; claim reserves, it does not make later edits idempotent;
  - the per-verb table;
  - the example.

- [ ] **Executable example** `TestSchedulingExampleRuns`, run on a real-git fixture with a coordinator checkout and a recipient worktree:
  - the recipient claims; the coordinator verifies with `issue show --json` (`card.status`, `assignment.relation`);
  - the recipient runs start-plan and change-code; the coordinator sees `branch` and `workspaces` activity and `checkpoints.flow`;
  - the recipient closes; the coordinator sees the `close` review verdict and `completion`;
  - with an unreachable tracker, the coordinator sees `stale` and the example says not to conclude the claim was lost.

  The test executes the steps from `recovery.Example` itself, so the rendered example and the tested one are the same data.
- [ ] Docs: atlas (a `workflow/recovery-contracts.md` page plus the index), a README pointer, and the process manual. `TestEveryFlagAppearsInItsHelp` covers the new command.
- [ ] Close.

## Revisions

- 2026-10-02 — plan-quality round 1 (PQ-1 to PQ-3 Important, three Minor).
  - **PQ-1:** the required verb set is derived from the command tree
    (mutating annotation, plus self-locking workflow verbs, plus
    `issue show`), with reasoned exemptions. push, issue sync, move-detail
    and issue new were added.
  - **PQ-2:** the surface is `sdlc help recovery`, a help topic, as the
    operator decided. There is no `recovery` verb to collide with
    `issue recovery`; the two cross-link. `--json` was dropped (YAGNI): the
    consumers (pair#362 and pair#367 skills) read help text.
  - **PQ-3:** each Done-when case class names its proof. Two new tests cover
    the close re-run after SHIP and a reopened card at landing.
  - **Minors:** reclaim adopts the uncertain helper; one `cardPublish` seam
    with a caller guard; `{{RECOVERY <verb>}}` for pages hosting several
    verbs.

- 2026-10-02 — M1 implementation.
  - **Change:** no `{{RECOVERY}}` placeholder. One `attachRecoveryContracts`
    pass in `buildRoot` appends each contracted verb's section to its help.
    That reaches every page, embedded or inline (the issue subcommands' Longs
    are inline), and none can forget it. The contract test asserts the
    section on every contracted verb.
  - **Coverage:** the derived required set is 26 commands. Every one
    without an `Exempt` reason has a contract. The top-level `set-status`
    alias and the issue card setters share one contract with
    `issue set-status`.

