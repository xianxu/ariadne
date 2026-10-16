# Archive-aware issue publication implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy). Continue the warm-context debugging in this session via superpowers-executing-plans. Steps use checkbox syntax.

**Goal:** Publish edits at the issue's current details location and prevent any initial-publication or recovery path from creating an active twin.

**Architecture:** Resolve issue identity across the canonical active and archive directories on a pinned main view. Preserve existing republish checks and use the existing TransferOp.BeforePublish hook to reject an archive in the candidate base before first publication, retry, or recovery writes main. No new transaction state, receipt format or background mechanism.

**Tech Stack:** Go, existing gitx.TrunkView and tracker transfer state machine, real Git checkouts with isolated bare origins.

## Core concepts

| Name | Kind | Lives in | Status |
|------|------|----------|--------|
| requireInitialDetails | PURE | cmd/sdlc/issuepublish.go | new |
| publishedDetailPath | INTEGRATION | cmd/sdlc/issuepublish.go | new |
| republishOwnedOnce | INTEGRATION | cmd/sdlc/republish.go | modified |
| newIssueTransferOp.BeforePublish | INTEGRATION | cmd/sdlc/issuemovedetail.go | modified |

The pure admission helper uses the vocabulary's terminal set, not a new enum. Add a direct table test without repository IO. The path resolver wraps existing pinned Git reads; it reuses the canonical filename parser and family classifier. Both publication flows consume that one resolver (ARCH-DRY, ARCH-PURE, ARCH-PURPOSE).

The external seam is existing gitx/tracker Git IO. Reuse trackerRepo, spinOffOnCodeBranch, ReceiptSpec and StepTransfer fixtures: an isolated bare origin is the stateful remote, and real Git exercises the same protocol as production (ARCH-MOCK). No new external service or production configuration.

## Existing evidence

- Original archive, terminal-missing-details and twin cases reproduced red.
- Both archive layouts, mixed-set atomic publication, republish concurrency and local-only archive refusal now pass.
- Recorded full suite on the initial correction passed 1,513 SDLC tests plus 66 packages.
- Close review BR-1 identifies transfer/recovery revalidation still missing; BR-2 identifies local filename filtering; BR-3 identifies incomplete acceptance criteria.

## Tasks

### 1. Complete acceptance and red regressions

Files: cmd/sdlc/issuepublish_tracker_test.go, cmd/sdlc/issuemovedetail_test.go, workshop/issues/000408-*.md.

- [x] Append Done-when clauses for all review findings and the local-only archive refusal, preserving the original criteria and adding a Revision.
- [x] runTrackedPublish: adversarial ID-prefixed non-detail entries → TestPublishIgnoresNonDetailEntries uses real Git and asserts publication ignores them while preserving their bytes.
- [x] newIssueTransferOp.BeforePublish: peer archives introduced across publication/retry/receipt-replay boundaries → TestInitialPublicationArchiveRevalidation uses real Git and asserts archive preservation, refusal, and source retention.
- [x] Run only these new regressions with go test ./cmd/sdlc -run 'Test(PublishIgnoresNonDetailEntries|InitialPublicationArchiveRevalidation)$' -count=1 -timeout 5m. Confirm failures identify the missing guards, then checkpoint.

### 2. Apply the shared publication guard

Files: cmd/sdlc/issuepublish.go, cmd/sdlc/issuemovedetail.go.

- [x] Filter local scan entries through issue.ParseFilename (via existing issueFilename) before classifyFamily, including both the duplicate and local-only archive decisions.
- [x] Extend newIssueTransferOp.BeforePublish: resolve directories from spec.DestinationPath and configured history, call publishedDetailPath on the supplied candidate-base view, refuse any existing details path, then run existing asset admission. The same hook is reconstructed by recovery and invoked before every actual main push.
- [x] Preserve confirmed-publication recovery: already reachable candidates finish without republishing; moved main triggers existing retry/reprepare, and the next candidate's base is checked before push. Never delete or rewrite an existing archive.
- [x] Add TestInitialDetailsTerminalAdmission as a direct pure test using validated card values for open versus terminal categories.
- [x] Run focused publication, transfer and recovery tests in a stable worktree; no edits during direct cmd/sdlc runs because of its hermeticity guard. Commit the fixes and review ledgers.

### 3. Verify and land

Files: atlas/workflow/issue-tracker.md and issue Log/Plan/Revisions.

- [x] Update Atlas to include transfer/retry/recovery location revalidation and filename grammar.
- [x] Push with sdlc pr; run sdlc local-check run --name suite on the pushed fix head. Require a recorded pass; no gate overrides.
- [ ] Run sdlc close --issue 408 with evidence disposing BR-1/BR-2/BR-3. Resolve any remaining findings; publish on SHIP through sdlc merge --yes and report PR #269 to the hub.

## Architecture and operating envelope

ARCH-ORDER: no new state machine. Existing transfer states and receipts remain authoritative. For unconfirmed main writes, probe reachability before retry; for moved main, reprepare and revalidate against the new candidate base. Tests control the archive's appearance before prepare, after a candidate is saved, and during recovery. A lock cannot replace location reconciliation across branch/main views; existing CAS and repo locks stay intact.

ARCH-SECURE: directory resolution remains repository-contained, snapshot reads use literal paths, and local candidate names must satisfy the existing Markdown filename grammar. Invalid or failed reads refuse instead of becoming absence.

ARCH-CONSTRAINTS: CLI batch path, existing three-directory metadata inventory; blob reads restricted to selected IDs. No daemon or added concurrency. Focused tests while iterating, one recorded full suite on the final fix head.

ARCH-FUNERAL: no new durable data family. Existing receipts, local-check logs and issue/review artifacts retain their existing removal/archive lifecycles.

## Revisions

### 2026-10-16 — review-driven full-flow checkpoint

The first correction fit the 100-line quick shell. Completing the common transfer/recovery guard and filename filtering is expected to exceed it. This plan captures the same no-twin invariant and the additional review evidence before further implementation; no independent feature is added.

### 2026-10-16 — PQ-1 and red evidence

Compressed the test strategy into one line per risky function; the executable tests own the scenario matrix. No behavior or mechanism changed. Red run: filename discovery fails all six cases; unguarded transfer publishes all six archive races (48.956s). The terminal-admission helper passes its pure category test.

### 2026-10-16 — correction verified locally

The new regression group passed in 84.718s. The broader Test(Publish|MoveDetail|AssetRecovery|InitialPublicationArchiveRevalidation|InitialDetailsTerminalAdmission) run passed in 207.171s with a stable worktree, including existing confirmed-publication recovery. The production correction remains one resolver and one existing publication hook, with no receipt/state changes.

### 2026-10-16 — full-suite evidence

Recorded suite on the pushed correction passed: 1,516 SDLC tests across 12 shards plus 66 packages, zero failures, 701.4s wall. Close review is the remaining boundary before landing PR #269.

### 2026-10-16 — BR-4 proposal, awaiting TL decision

Close round 2 disposed BR-1/BR-2/BR-3, but reproduced stale terminal eligibility after preflight. This is the third mechanism correction, so implementation is stopped under the TL's explicit limit. Proposed rule: every unconfirmed initial publication requires a current nonterminal card AND no ID-wide details. Reuse BeforePublish: read the current tracker card, apply requireInitialDetails, then retain location/asset checks. The existing reachability-first recovery behavior is unchanged. No new hook, receipt or state.

If approved, add terminal transition regressions at the existing first/retry/saved-receipt boundaries using the stateful Git fixture; assert refusal, unchanged main and source retention. Run red/green, the broader group and recorded full suite, then re-close. No code for BR-4 has been written.
