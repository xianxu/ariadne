# Sandbox-safe merge cleanup implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3. Execute coupled cleanup in the main session; delegate bounded producer work. Steps use checkboxes.

**Goal:** Protected shared Git config cannot fail an otherwise confirmed landing.

**Architecture:** Prevent unnecessary issue upstream metadata, retain guarded ref retirement, and retry legacy metadata through the existing done-card sweep. Reuse Git runners, stateful GitHub fixtures and tracker cards; create no journal or permission exception.

**Tech Stack:** Go, Git refs/config/worktrees, Cobra, existing tracker/GitHub seams.

## Core concepts

No new pure entity: existing landing observations and done-card status retain their model. New helpers read or mutate Git and are INTEGRATION.

| Entity | Kind | Location | Status |
| --- | --- | --- | --- |
| leasedBranchPush | INTEGRATION | cmd/sdlc/boundarypush.go | modified |
| runPR | INTEGRATION | cmd/sdlc/pr.go | modified |
| createInPlaceBranch/createWorktreeBranch | INTEGRATION | cmd/sdlc/branchcreate.go | modified |
| deleteLandingBranch | INTEGRATION | cmd/sdlc/landing.go | modified |
| cleanupBranchConfig/removeObsoleteBranchConfig | INTEGRATION | cmd/sdlc/branchconfig.go | new |
| finishLandedLeftovers | INTEGRATION | cmd/sdlc/externalmerge.go | modified |

gitFn/gitRunner inject real Git; landingFakeGH models mutable PR state and merge effects. Temporary repositories prove config/ref semantics; native sandbox checks qualify protected-file behavior (ARCH-MOCK). The shared cleanup helper serves direct cleanup and the done-card sweep (ARCH-DRY).

## Invariants and operating envelope

- Issue pushes use explicit destinations/refspecs, preserve leases and remote tracking refs, and create no upstream config even with automatic setup enabled. Main/rest tracking stays unchanged.
- Ref deletion remains CAS after identity, rest and occupancy checks. Critical probe failures remain errors; only ancillary config cleanup is nonblocking.
- Existing config is the retry marker. A done card plus absent local ref and no checkout authorizes removal. SDLC callers hold the shared repository transaction lock, so another cooperating SDLC command cannot recreate the branch during cleanup. A ref/checkout observed before removal preserves its metadata; failed probes never authorize deletion. Manual concurrent Git does not take this lock: it may recreate an absent branch after these probes, whose ancillary tracking config may be cleared; cleanup never deletes that recreated ref or its commits. This is the supported ownership boundary, not a cross-resource Git transaction.
- Denial after ref deletion warns and succeeds. Duplicate runs observe the existing merge/archive without repeating publication. A later writable merge/reconcile removes metadata. Cancellation before deletion preserves the ref; cancellation after it leaves discoverable metadata.
- Batch CLI cleanup uses the existing repo lock and command context. Inventory is the finite done-card/config population. Query candidate keys once; avoid Git subprocesses for cards without matching config. Existing bounded Git runners remain in use. No background work is introduced.
- No new artifact family (ARCH-FUNERAL): residue is a pre-existing branch section, normally two upstream fields per prior/manual issue. Cleanup is its last consumer; a writable sweep removes it. Persistent config protection retains only that residue and warns.

## Function-specific verification strategies

All rows are integration tests; none misclassifies Git-backed behavior as pure.

| Risky function | Adversarial input or ordering class | Mechanical guard |
| --- | --- | --- |
| leasedBranchPush | Automatic Git setup settings and rewritten/unseen remote tips | Real scratch Git asserts config absence, exact published/tracking OIDs and existing lease refusals; restoring -u must fail |
| runPR | Publication without upstream, configured remote and legacy workspace | Real repositories and existing stateful PR seam assert selected remote head and no new branch config |
| createInPlaceBranch/createWorktreeBranch and abandoned restoration | Automatic tracking inheritance settings | Real Git asserts unchanged branch tip and no issue upstream while main config remains valid |
| deleteLandingBranch | Persistent/transient failures across irreversible landing and ref retirement | Full merge caller with stateful GitHub fake and real config.lock asserts one merge/archive, critical refusal before ref retirement, nonfatal metadata failure afterward |
| removeObsoleteBranchConfig | Ref/checkout presence, failed observations and actor ordering at config removal | Real Git plus deterministic boundary injection asserts no config write on present/failed probes; cooperating operations are serialized by the existing repo lock; a manual recreation at the final mutation boundary retains its exact ref/commit even though config is outside that guarantee |
| finishLandedLeftovers | Done-card authorization versus absent, live and unrelated metadata | Drive production reconcile over persisted cards/config; removing sweep wiring must leave a failing stale-section assertion |

## Tasks

- [x] Observe producer regressions fail on baseline, implement correction behind existing Git seams, and observe GREEN.
- [x] Observe compatibility regressions fail on baseline, implement nonblocking ancillary cleanup and retry discovery, and observe combined GREEN.
- [ ] Finish compatibility adaptation and adversarial guards using the function strategies above; mutation-check the changed admission rules.
- [ ] Update PR/merge help, recovery catalog, atlas and README to describe explicit issue publication and deferred legacy metadata cleanup with its ownership boundary.
- [ ] Run relevant focused regressions and recorded full suite on pushed integrated HEAD via sdlc local-check run --name suite.
- [ ] Fulfill original default-agent-sandbox scratch landing acceptance with the hub; record exact environment/effects. Custom-profile reproduction cannot silently replace that live criterion.
- [ ] Close through binary-owned review, fix findings, merge current main, land via sdlc merge --yes with all validations and report actual hashes.

## Verification commands

New cases: go test ./cmd/sdlc -run '^(TestLandingDefersBlockedConfigCleanup|TestReconcileRetriesObsoleteBranchConfig|TestIssuePublicationDoesNotSetUpstream|TestIssueCreationDoesNotInheritUpstream|TestIssueReopenDoesNotSetUpstream|TestLegacyPRDoesNotSetIssueUpstream)$' -count=1 -timeout 5m. Baseline fails for unwanted upstream config, fatal config lock, unretried residue; implemented head passes.

Compatibility: go test ./cmd/sdlc -run '^(TestLanding.*|TestReconcile.*|TestBoundary.*|TestDeleteRemoteBranch.*|TestCreateInPlaceBranch.*|TestCreateWorktreeBranch.*|TestIntegrateGate.*|TestReopen.*)$' -count=1 -timeout 15m. Expected PASS with no shared-checkout edits during the run.

## Revisions

### 2026-10-14 — plan review PQ-1/PQ-2

PQ-1: narrowed the former universal claim “Recreated refs retain their metadata” to the actual SDLC actor model. The shared repo transaction lock excludes cooperating SDLC recreation; a manual Git actor can change ancillary config after probes, but cleanup never deletes its recreated ref or commit. Add deterministic final-boundary injection to prove this limit rather than promising atomic config/ref transactions that Git's runner does not supply.

PQ-2: replaced scenario/edit inventories with a function-specific adversarial strategy table and mechanical guards. Tasks now track deliverables and reference that executable verification strategy.
