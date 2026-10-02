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
- **Help.** The registry renders into each verb's `--help` through a `{{RECOVERY}}` placeholder, and into one `sdlc recovery` page (`--json` for agents). That page also carries the agent guidance and an executable scheduling example.
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
- **Catalog entries** cover claim (with `--adopt` and relocation repair), reclaim, start-plan, change-code, milestone-close, close (with `issue recovery reconcile`), pr, merge, `issue set-status`, move, and `issue show` (read-only). Each entry's text and test names come from the explorer map in the issue Log, rechecked against code while writing.
- **Example** is a list of steps. Each step names an actor (coordinator or recipient), a command, the evidence to read (dot-paths into the observation JSON), the expected value, and what to do otherwise. It is both rendered into the page and executed by a test.
- **DRY rationale:** one source for the contract text, the help, the JSON and the test links, as the gate catalog is for gate flags.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `{{RECOVERY}}` placeholder | `cmd/sdlc/main.go renderLong` | modified | `recovery.Section(name)` |
| `sdlc recovery [verb] [--json]` | `cmd/sdlc/recoverycmd.go`, `helptext/recovery.md` | new | the registry |
| uncertain-publication guidance | `cmd/sdlc/claim.go`, `cardsetters.go` (set-status) | modified | `gitx.ErrPublicationUncertain` |

- **Uncertain-publication guidance.** One helper wraps `ErrPublicationUncertain` from a single-card CAS verb with "the change may or may not have published; rerun the same command — the card decides", matching reclaim. It is used by claim, adopt, relocation and set-status, and is single-sourced instead of written into each verb.

### Lifecycle and ordering

The registry is static data: it creates nothing durable, and adds no runtime state or ordering. The executable example runs only in tests, against disposable fixtures.

## M1 — The registry, its rendering, and the gaps it exposes

- [ ] `internal/recovery`: the types, the class set and validation. Pure tests check that every class is valid and every entry has an evidence line and a lost-response line.
- [ ] Write the catalog entries for all MVP verbs, rechecking each claim against the code.
- [ ] `{{RECOVERY}}` in each verb's helptext, plus `recovery.Section`. `TestNoCommandLongHasSurvivingPlaceholder` keeps passing.
- [ ] **Contract test `TestRecoveryContractsAreProven`:**
  - every MVP verb has exactly one entry;
  - every named test exists as a `func TestX` somewhere under `cmd/sdlc` (AST scan);
  - every verb page renders its section.
  - Mutation check: rename a referenced test and the contract test must go red.
- [ ] **Gap fixes:** the uncertain-publication helper, used by claim, adopt, relocation and set-status. New lost-acknowledgement tests:
  - claim: an injected uncertain error after the effect lands, then a rerun gives "already claimed";
  - set-status: the same shape, the rerun gives "already has that status".
  - Injection uses a package-var publish seam on each, as `reclaimEffect` does.
- [ ] M1 milestone-close.

## M2 — The recovery page and the executable scheduling example

- [ ] `sdlc recovery [verb] [--json]` (read-only), with `helptext/recovery.md`. The page holds:
  - the class definitions;
  - the agent guidance: verify effects through read-only queries; roughly 30 s is a revisit heuristic, not proof of loss; retry only under the verb's contract; unknown or stale is not negative evidence; claim reserves, it does not make later edits idempotent;
  - the per-verb table;
  - the example.

  `--json` emits the catalog, versioned, for agents.
- [ ] **Executable example** `TestSchedulingExampleRuns`, run on a real-git fixture with a coordinator checkout and a recipient worktree:
  - the recipient claims; the coordinator verifies with `issue show --json` (`card.status`, `assignment.relation`);
  - the recipient runs start-plan and change-code; the coordinator sees `branch` and `workspaces` activity and `checkpoints.flow`;
  - the recipient closes; the coordinator sees the `close` review verdict and `completion`;
  - with an unreachable tracker, the coordinator sees `stale` and the example says not to conclude the claim was lost.

  The test executes the steps from `recovery.Example` itself, so the rendered example and the tested one are the same data.
- [ ] Docs: atlas (a `workflow/recovery-contracts.md` page plus the index), a README pointer, and the process manual. `TestEveryFlagAppearsInItsHelp` covers the new command.
- [ ] Close.

## Revisions
