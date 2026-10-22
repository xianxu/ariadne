---
collaboration-mode: discussion
---

# New repository tracker initialization plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development for bounded tests and superpowers-executing-plans for the warm-context orchestration. Steps use checkbox syntax for tracking.

**Goal:** Make tracker initialization the default documented final step of new non-brain repository setup, while existing legacy repositories keep their behavior and receive migration guidance.

**Architecture:** Add explicit `sdlc issue init` as a narrow empty-inventory entry to the existing migration/bootstrap owner. Reuse its remote CAS, exact-root verification and cutover marker; finish by fast-forwarding the clean main checkout. Do not infer newness during ordinary weave compilation or publish from CI/refresh.

**Tech Stack:** Go/Cobra, existing tracker migration and Git trunk APIs, real local Git/bare remotes, existing command fixtures.

**Status:** Proposed; TL must approve the explicit setup step and opt-out boundary before change-code. This is planned as full flow because it adds an initialization/publication action; actual size and estimate follow the gate.

## Design decision

There is no distinct repository creator in this tree. The current setup is `weave link` then `weave compile`; bootstrap.sh merely invokes compile, and refresh shares its implementation. Compile supports non-Git roots and CI; refresh explicitly preserves leaf HEAD/index/branch/origin-main. An unconditional hook there would unexpectedly publish and could auto-migrate legacy repos. Missing scaffold or a first managed-index write cannot prove a new repo, and loses eligibility after partial setup.

Recommend an explicit `sdlc issue init` in the canonical new-repository setup sequence, after committing scaffolding and configuring/pushing main. Existing adoption/restore instructions remain distinct. Opt-out is choosing the existing adoption/compile path without issue init (or declining issue scaffolding); no new preference format or pending-setup marker. A brain marker always suppresses initialization. This is an explicit-init interpretation of the issue's allowed equivalent init step; seek TL confirmation before implementation.

Alternatives rejected: auto-init during compile/refresh conflates composition with remote publication and breaks existing contracts. Auto-init on first issue new changes the legacy verb's authority and cannot distinguish an intentionally legacy repo. A new top-level repository creation orchestrator duplicates setup and expands this issue unnecessarily.

## Core concepts

### Pure entities

| Name | Lives in | Status |
| --- | --- | --- |
| EmptyInitEligible | cmd/sdlc/internal/tracker/initialization.go | new |
| InitState / InitEvent / InitEffect / InitTransition | cmd/sdlc/internal/tracker/initialization.go | new |
| LegacySetupHint | cmd/sdlc/internal/tracker/initialization.go | new |
| MigrationInput / MigrationManifest / PlanTrackerMigration | cmd/sdlc/internal/tracker/migration.go | reused unchanged |

Eligibility accepts only an empty migration input: no active/archived details, branch issue edits, or dirty issue paths and no migration refusals. It cannot bless a nonempty repository by checking only card count. The input belongs to one pinned repository/main inventory; callers supply already observed facts. Unit tests require no IO. Existing completed tracker verification precedes this check, so rerunning initialization after later issue creation is a no-op when local main already carries the matching cutover.

### Integration points

| Name | Lives in | Status | Wraps |
| --- | --- | --- | --- |
| issue init command / runIssueInit | cmd/sdlc/issueinit.go | new | explicit setup action and checkout completion |
| issue subcommand registration | cmd/sdlc/issue.go | modified | Cobra command surface |
| Existing migration/bootstrap owner | cmd/sdlc/issuemigrate.go | reused or minimally shared | inventory, tracker bootstrap, main cutover |
| Existing tracker/trunk publication | cmd/sdlc/internal/tracker, cmd/sdlc/internal/gitx | reused unchanged | Git remote CAS and cutover identity |
| Legacy migration hint | cmd/sdlc/legacymode.go | modified | one shared hint writer, no mutation |
| Legacy dispatches | cmd/sdlc/issue.go, cmd/sdlc/startplan.go | modified | hint at actual legacy entry |

The shell owns all Git/fs IO. Use existing local bare-remote fixtures/stateful repositories; no GitHub credentials, network service or function-call-only fake needed. Git is exercised through existing Git seams against real temporary repositories. Existing repository transaction lock spans initialization; no new lock or daemon. No weave tool lookup or compile hook is introduced.

## Contract and ordering

1. Resolve the explicit calling repository. Use canonical gitx.IsBrainRepo before any tracker mutation; brain setup reports skipped and creates no tracker. Missing/declined workshop/issues similarly skips rather than creating unwanted scaffolding. Ordinary compile/refresh never invoke this command implicitly.
2. Require the repository's clean resting main checkout, committed HEAD, configured publication target and reachable pushed main. Non-Git, unborn, no remote, unpushed main, dirty or non-main branch refuse before writes with the specific prerequisite to fix. Do not initialize Git, create a remote, commit source files or change branch automatically.
3. Open the existing unguarded migrate environment. If main already carries a valid matching tracker root, validate existing state; return if local checkout already incorporates that marker, otherwise finish the safe fast-forward below. A malformed marker, unreachable remote or mismatched tracker root is an error, never evidence of legacy absence.
4. Inventory through migrationInventory and apply the pure empty-init predicate. Existing issues/history/branch work means refuse with `sdlc issue migrate`; do not bypass migration's reviewed digest path for nonempty repositories.
5. Derive the ordinary empty MigrationManifest; call applyTrackerBootstrap and applyMainCutover, sharing their expected-absence CAS and exact-root checks. No new manifest version, root convention, branch name or alternative card authority.
6. Obtain the confirmed main cutover commit, verify local main remains clean and is its ancestor, then `git merge --ff-only` that exact commit. Never reset, stash or overwrite edits. Failure retains both published phases and names the safe explicit retry. Init success means the calling checkout can immediately use tracker-backed issue new and start-plan.

An offline result cannot report successful initialization. The existing initial commit and Git registration stay owned by setup, not this command. Unknown push acknowledgment preserves known partial state; retry observes root/marker identity rather than inventing completion. Concurrent initializers serialize locally; cross-clone races retain the existing remote CAS. A concurrent later issue write must not be overwritten or make init re-create the tracker.

## States, actions and lifecycle

Actions: complete Git prerequisites; rerun issue init; use reviewed issue migrate for nonempty legacy inventory; opt out by continuing compile/adoption without init.

| Observation / interruption window | Visible state | Recovery action | Possible loss |
| --- | --- | --- | --- |
| Brain or declined issue scaffold | initialization skipped | continue setup without tracker | none |
| Missing Git/main/remote or dirty/non-main caller | prerequisite refusal, unchanged repository | complete named prerequisite, rerun init | none |
| Nonempty legacy inventory | needs reviewed migration | issue migrate | none |
| Valid initialized tracker and current checkout | already initialized | continue setup | none |
| Before tracker CAS | not initialized or outcome unconfirmed | rerun init | none; original evidence retained |
| Tracker published, main marker absent | initialization incomplete | rerun init, verify same empty root | none; remote branch retained |
| Main marker published, local checkout stale | initialized remotely; checkout completion pending | rerun init after clean main | none; no forced checkout mutation |
| Marker/remote root inconsistent | evidence refusal | restore/correct named authority, rerun | none; no deletion or inferred success |

The pure eligibility/phase decision declares permitted effects; the existing migration/trunk owner executes publication. Test every interruption boundary through observed state changes, including a lost push response and concurrent initialization; failed reads never become absence. Init introduces no new durable artifact family: existing tracker root, main marker, temporary Git index/trunk stages and recovery evidence retain their current lifetimes. Opt-out has no new persisted config. Tracker history is existing authoritative issue history, not an init-specific growing ledger.

## Operating envelope and architectural review

This is an explicit batch setup command, never per keystroke or ordinary compile. Reuse bounded Git calls/output limits and caller cancellation. One pinned main inventory plus existing branch/worktree migration scan; no extra scan per issue. Initial repo has zero issue files. Count 8/16 unrelated files/branches to verify init does not spawn a new process per file; preserve existing migration limits and refusals. Network latency is the existing trunk/publication budget; an unreachable origin refuses visibly. No new CPU, memory or concurrency worker pool.

ARCH-DRY/PURE: share migration publication and make only the stricter empty eligibility pure. ARCH-PURPOSE: real setup fixture must proceed from init through issue new and start-plan. ARCH-MOCK: real portable Git/bare remote models actual state transitions and failures. ARCH-CONSTRAINTS: explicit bounded batch action; no compile hot-path cost. ARCH-SECURE: no credentials created/logged; all remote/marker/root reads remain typed and validated, and brains/opt-out are checked before writes. ARCH-ORDER: retain partial publication, test before/after both CAS boundaries and checkout completion; no rollback or inferred absence. ARCH-FUNERAL: reuse existing artifact owners/lifetimes. ARCH-EVOLVE: no existing format change; reused cutover version/root schema and unknown-field preservation stay intact.

## Enforced transition owner

`tracker.InitTransition(state InitState, event InitEvent) (InitState, InitEffect, error)` is the only phase mutation. InitState's fields and phase enum are package-private; callers construct it through NewInitialization and cannot advance it directly. Events/effects are tagged types. The returned state records confirmed root/commit identity, not just a phase boolean. The shell may only execute returned effects and feed their outcomes back; it never chooses a publication from raw marker observations.

| State | Event | Next state / returned effect |
| --- | --- | --- |
| new | Begin | inspecting / InspectInitialization |
| inspecting | Inspected(skip or prerequisite/refusal/unknown) | terminal skipped/refused; no mutation |
| inspecting | Inspected(confirmed complete/current) | done; no mutation |
| inspecting | Inspected(verified empty/uninitialized) | bootstrapping / BootstrapTracker |
| inspecting | Inspected(matching tracker-only root) | publishing / PublishCutover(root, expected main) |
| inspecting | Inspected(matching remote cutover, stale clean checkout) | completing / FastForwardCheckout(commit) |
| bootstrapping | BootstrapConfirmed(root) | publishing / PublishCutover(root, expected main) |
| publishing | CutoverConfirmed(commit) | completing / FastForwardCheckout(commit) |
| completing | CheckoutConfirmed | done; no effect |
| any active phase | effect failed, canceled or unconfirmed | terminal retry/refused, preserve confirmed identities; no next effect |
| wrong phase or terminal | duplicate/out-of-order event | reject without effect |

The in-memory machine is one invocation; retries reconstruct observations from existing durable root/marker/checkout authority. No new on-disk state is introduced. The IO adapter `initializationEffects` owns Inspect/Bootstrap/Publish/FastForward; its methods wrap existing migration/Git operations. A stateful fake and the real bare-Git adapter use the same effect dispatcher, providing a deterministic barrier before and after each external effect. Lost acknowledgments stop this invocation; a new invocation observes whether publication occurred. Cross-clone CAS races and later card creation are sequenced at these barriers. As with migration, setup requires other legacy SDLC writers to remain quiescent until initialization completes; the command does not claim to fence uncooperative old binaries or ordinary editors in another clone. Pinned-main CAS, exact-root checks, clean-checkout revalidation and the existing local repo lock remain the mechanical guards.

## Advisory legacy inventory

`legacySetupAdvice` resolves configured issue/history roots and uses `issue.LocalPathsByID`; it does not call the migration scanner, fetch or rescan remote state. `tracker.LegacySetupHint` chooses text from known-local-empty, local-nonempty or unknown observations. Empty-local text says issue init and qualifies the evidence: init verifies the complete inventory before publication. Nonempty text says migrate. Unknown reads are never called empty, and missing/declined scaffolds and brains receive no advice. The helper runs before legacy issue creation, and in runStartPlan beside its existing repositoryTracked result so bare start-plan is included. Advice changes neither legacy effects nor issue-new stdout.

## Tasks and verification strategies

### 1. Implement initialization through the pure model

Files: new cmd/sdlc/issueinit.go, cmd/sdlc/issueinit_test.go, cmd/sdlc/internal/tracker/initialization.go and initialization_test.go; register in cmd/sdlc/issue.go; share only necessary migration helpers from cmd/sdlc/issuemigrate.go.

- [x] Write pure decision/sequence tests and actual command regression RED, then implement the model and thin effect adapter.
- [x] Run the production init -> issue new -> claim/start-plan fixture against a disposable bare remote; commit after GREEN.

### 2. Wire legacy advice, recovery contract and default setup procedure

Files: cmd/sdlc/legacymode.go, cmd/sdlc/issue.go, cmd/sdlc/startplan.go and their tests; cmd/sdlc/internal/recovery/catalog.go (existing recovery owner), command help; README.md and atlas/workflow/{setup-and-replication,sdlc-binary}.md. Sweep operator/TL-facing setup guidance as required by the approval.

- [x] Write legacy advice regressions RED, implement one shared advisory helper and the pure text decision, and pass tests.
- [x] Add issue init to the recovery catalog, with verified retry semantics, help and proof references; run TestRecoveryContractsAreProven.
- [x] Make every new-repository/bootstrap creation procedure end with issue init after pushed initial main; preserve adoption/restore/refresh semantics and document opt-out/brain handling.

### 3. Acceptance and landing (one whole-issue boundary)

- [ ] Run private production CLI conformance against a disposable bare remote (TL approved, no stage), then the recorded pushed-head suite and close review.
- [ ] On SHIP and completed branch pass enter sdlc merge --yes, retain its integration run, verify merge/card/archive and report to hub. TL owns project updates and natural fleet follow-through.

| Risky function | Adversarial strategy / mechanical guard |
| --- | --- |
| EmptyInitEligible | Table over arbitrary MigrationInput contents and refusals; any nonempty or unknown input cannot authorize init. Pure predicate, no IO. |
| InitTransition | Generated event sequences with duplicate/out-of-order/cancel/unconfirmed events; opaque state and returned effects prevent skipped publication phases and further effects after uncertainty. |
| InspectInitialization | Contradictory repo/marker/root observations in a stateful backend; typed observed disposition and existing cutover parser/root verification reject fabricated empty or complete state. |
| runIssueInit / initializationEffects | Real bare-Git fixture plus controlled before/after-effect failures and concurrent clones; compare refs, card blobs, index/HEAD and working bytes across retries, with remote CAS and ff-only as guards. |
| LegacySetupHint / legacySetupAdvice | Local discovery fakes/temporary roots with empty, nonempty and unreadable observations; conservative text policy, no remote calls, pre-creation timing and unchanged stdout/effects are asserted through both production verbs. |
| Initialization inventory cost | Count Git calls at 8/16 unrelated files/branches through the existing runner; retain linear branch scan and fixed metadata reads, reject new per-file processes. |
| Recovery contract / command help | TestRecoveryContractsAreProven derives the mutating command obligation and validates proof names plus generated help. |

Targeted command: `go test ./cmd/sdlc/internal/tracker ./cmd/sdlc -run 'Test(IssueInit|EmptyTrackerInit|Initialization|LegacySetupHint|RecoveryContractsAreProven)' -count=1 -timeout 10m`. Expected GREEN; each behavior change must first produce its intended RED. The private production-CLI scenario is the full new-repository procedure, followed by the first issue and planning branch, rather than a mocked bootstrap-only call. All individual cases belong in executable tests, not this plan.

## Revisions

### 2026-10-21 — TL-approved default setup and inventory-sensitive hints

TL approved the explicit command (Couch 87ebe800-5c28-4e19-b11b-5bf5c2c1453a), full flow, and private production CLI plus real bare-Git conformance with no stage. Ordinary weave link/compile/refresh must not publish. To make initialization the default, every documented new-repo/bootstrap procedure in README and atlas, including operator/TL-facing creation guidance discovered in the implementation sweep, ends with issue init. The legacy hint is inventory-sensitive: empty inventory says issue init, nonempty says issue migrate. Brain/declined scaffold skip unchanged. This replaces Task 2's unconditional migrate wording; tests must prove both hints without changing legacy effects or the issue-new stdout path contract. Reuse one empty-inventory policy for initialization and hint selection; hint evidence is advisory, never publication authority, and unreadable/unknown inventory must not be claimed empty. The first hint observes pre-creation inventory (before legacy issue new creates its first details file).

### 2026-10-21 — plan-quality round 1 correction

PQ-1: named the opaque InitState/InitEvent/InitEffect model and InitTransition owner, explicit state/event/effect table, stateful effect dispatcher and deterministic publication boundaries. This formalizes the approved phases without changing authority or adding persisted state. PQ-2: replaced procedural case inventories with one named function/adversarial strategy/mechanical guard row per risky function. PQ-3 (Minor): included the required mutating-command recovery catalog/help/proof integration and TestRecoveryContractsAreProven. Advisory hint discovery uses the existing local-only scanner and labels local-empty evidence; authoritative init retains full migration inventory. No code changes before the gate passes.

### 2026-10-21 — implementation checkpoint after approval and change-code

The proposed status above is historical: TL approved the design and the full-flow gate passed after clean plan-quality rounds 2/3 and estimate acceptance. Implementation and targeted/private-CLI acceptance are complete; the recorded suite and whole-issue review remain. The approved procedure sweep also found tl-lead team-new help and its printed operator next steps; both now give base adoption/compile, committed pushed main, then explicit init, with a regression over their order. No new publication authority or mechanism was added.
