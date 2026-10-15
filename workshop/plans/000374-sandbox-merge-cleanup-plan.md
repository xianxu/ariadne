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
- Existing config is the retry marker. A done card plus absent local ref and no checkout authorizes removal. Recreated refs retain their metadata. Failed probes never authorize deletion.
- Denial after ref deletion warns and succeeds. Duplicate runs observe the existing merge/archive without repeating publication. A later writable merge/reconcile removes metadata. Cancellation before deletion preserves the ref; cancellation after it leaves discoverable metadata.
- Batch CLI cleanup uses the existing repo lock and command context. Inventory is the finite done-card/config population. Query candidate keys once; avoid Git subprocesses for cards without matching config. Existing bounded Git runners remain in use. No background work is introduced.
- No new artifact family (ARCH-FUNERAL): residue is a pre-existing branch section, normally two upstream fields per prior/manual issue. Cleanup is its last consumer; a writable sweep removes it. Persistent config protection retains only that residue and warns.

## Tasks

### Producer correction

- [x] Add real-Git tests in cmd/sdlc/issueupstream_test.go: publication/legacy PR preserve remote tips without upstream config; creation/reopen suppress automatic inheritance; main tracking remains valid. Observe RED.
- [x] Remove issue push -u in boundarypush.go/pr.go; add --no-track at legacy creation/restoration sites in branchcreate.go/abandon.go. Preserve existing durable --no-track. Observe GREEN.

### Compatibility cleanup

- [x] Add branchconfig_test.go regressions: persistent config.lock does not fail a full merge; retry does not repeat merge/archive; writable retry removes metadata. Reconcile preserves live/recreated refs and unrelated config, then removes obsolete metadata. Observe RED.
- [x] Keep landing.go ref guards blocking. Route config cleanup through branchconfig.go and retry it from externalmerge.go using done cards and existing sections. Report path, deferred status and retry owner. Observe combined GREEN.
- [ ] Adapt existing interruption tests for nonfatal config cleanup; add native sandbox and occupancy/probe-failure controls. Mutation-check new guards.

### Documentation and acceptance

- [ ] Update helptext/pr.md, merge.md, root.md, recovery/catalog.go, atlas/workflow/workspace-branching.md and README.md for explicit issue publication and automatic compatibility cleanup.
- [ ] Run focused publication/creation/landing/reconcile/integrate regressions, then recorded full suite on pushed integrated HEAD using sdlc local-check run --name suite.
- [ ] Fulfill original default-agent-sandbox scratch landing acceptance with the hub; record exact environment/effects. Custom-profile reproduction cannot silently replace that live criterion.
- [ ] Close through binary-owned review; fix findings, merge current main, land via sdlc merge --yes with all validations, report actual hashes.

## Verification commands

New cases: go test ./cmd/sdlc -run '^(TestLandingDefersBlockedConfigCleanup|TestReconcileRetriesObsoleteBranchConfig|TestIssuePublicationDoesNotSetUpstream|TestIssueCreationDoesNotInheritUpstream|TestIssueReopenDoesNotSetUpstream|TestLegacyPRDoesNotSetIssueUpstream)$' -count=1 -timeout 5m. Baseline fails for unwanted upstream config, fatal config lock, unretried residue; implemented head passes.

Compatibility: go test ./cmd/sdlc -run '^(TestLanding.*|TestReconcile.*|TestBoundary.*|TestDeleteRemoteBranch.*|TestCreateInPlaceBranch.*|TestCreateWorktreeBranch.*|TestIntegrateGate.*|TestReopen.*)$' -count=1 -timeout 15m. Expected PASS with no shared-checkout edits during the run.
