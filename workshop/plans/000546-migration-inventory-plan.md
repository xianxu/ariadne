---
collaboration-mode: discussion
---

# Complete migration inventory implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy). Use superpowers-executing-plans for the warm-context integration; bounded pure tests or observation surveys may be delegated. Steps use checkbox syntax.

**Goal:** Prevent issue-tracker cutover from silently omitting issue evidence outside main's active root, without importing records from branches or worktrees.

**Architecture:** One complete observation of pinned Git tips and live worktrees feeds the existing main-authoritative pure migration planner. Root classification and observation are shared with initialization; init's presence-only acceptance remains separate. Related active branch reconciliation and legacy close proofs retain their current semantics; new non-main archive and unrelated-snapshot checks refuse rather than choose records to import.

**Tech stack:** Go, Cobra, existing tracker/issue/gitx packages, native Git and portable bare-repository fixtures.

## Approved contract and alternatives

The issue's Spec is the root × scope authority. TL approved it on 2026-10-22 (Couch 0fc506bf-9240-416e-b647-898eb9f309a1), then explicitly confirmed (7a42fef5-cd44-4876-b333-aff944ce8657) that byte matching applies to archives and unrelated active snapshots only. Related active records still use merge-base changes and card reconciliation; detail prose may differ. Private production CLI plus real bare-Git conformance is sufficient; no stage.

Main imports active details plus both archive roots returned by `issue.IDDirs`. Every local/current publication-remote tip and every live worktree HEAD is observed, without name exemptions. Present active IDs absent from main's population refuse even when inherited. Non-main archives must match a main archive by filename ID and blob bytes; unrelated active records must match a main active record by ID and bytes. A path move is allowed. Archive absences on an older tip are not fabricated deletions. Differing stale archive copies intentionally refuse: the operator aligns, lands or drops that evidence before migration.

The alternatives were multi-source importing (requires new authority/conflict selection) and patching only the two reproductions (still misses inherited, unrelated and detached evidence). Complete observation plus conservative refusal is the smallest complete choice for this rare operator-run command.

## Core concepts

| Pure entity | Lives in | Status |
|---|---|---|
| `MigrationObservedFile` and `MigrationTipEvidence` | `cmd/sdlc/internal/tracker/migration_inventory.go` | new |
| `MigrationEvidenceRefusals` | `cmd/sdlc/internal/tracker/migration_inventory.go` | new |
| `MigrationInput` / `PlanTrackerMigration` | `cmd/sdlc/internal/tracker/migration.go` | modified |
| `IssueInventoryPath` | `cmd/sdlc/internal/issue/inventory.go` | new, extracted from initialization |

`MigrationObservedFile` contains a validated path, active/archive role, filename ID, mode and blob OID. A blob OID in this one repository/object format identifies bytes; matching keys are `(role, ID, blob OID)`, independent of path. `MigrationTipEvidence` owns one immutable tip's metadata and its diagnostic carriers, with a typed related/unrelated ancestry result. No document body parsing is needed for presence/matching. Invalid file kinds or malformed identities refuse at the adapter boundary.

`MigrationEvidenceRefusals` builds main's ID set and published active/archive key sets once, then scans observed records once. Related active records require only an already represented ID here: their actual changes/deletions remain in `MigrationInput.Branches` for the existing reconciliation/rename rules. Archives never enter `Branches`, so they cannot count as an active rename or satisfy active-card reconciliation. Unknown active IDs and unrepresented archive/unrelated-active records yield deterministic refusals naming paths, carriers and legacy recovery.

`MigrationInput` gains typed main metadata and tip evidence. `PlanTrackerMigration` remains the sole semantic manifest owner and composes the additional refusals before sorting/digesting. `EmptyInitEligible` must reject nonempty added evidence; its defensive contract cannot silently ignore new fields. Actual initialization still uses its independent presence gate.

The shared classifier derives roots through `issue.IDDirs`, accepts only direct-child canonical issue filenames, and identifies active versus archive. It replaces initialization's local duplicate predicate without changing init's behavior. Growth is additional roots supplied by the vocabulary, not copied literal path lists.

| Integration point | Lives in | Status | Wraps |
|---|---|---|---|
| `observeIssueInventory` | `cmd/sdlc/issueinventory.go` | new, extracting existing observation |
| `publicationBranchTips` | `cmd/sdlc/issuemigrate.go` | reused | existing explicit all-heads fetch and pinned names/OIDs |
| `migrationInventory` / `branchIssueEdits` | `cmd/sdlc/issuemigrate_inventory.go` | modified/extracted | pinned Git trees, merge bases and close evidence |
| `dirtyIssuePaths` | `cmd/sdlc/issuemigrate_inventory.go` | modified/extracted | NUL Git worktree/index/status data and filesystem metadata |

The observer reuses #543's enumeration, root validation and metadata selection. Extract reusable mechanics, not init's acceptance policy. Read each distinct OID once, retaining all carrier labels. Snapshot the worktree list once with `gitx.ParseWorktrees`; skip only explicit bare/prunable records. Detached HEADs participate as first-class immutable tips. Migration consumes all observed tips rather than the name-filtered `migrationBranchRefs` list. Keep the old helper only if an existing semantic consumer still requires it; do not introduce a second branch enumerator.

Git remains behind `trackerEnv` / `gitx` contextual runners. Immutable trees use `TrunkView.FilesSelected`; select metadata before reading bodies, and batch bodies only for main imports and actual active changes. Existing `gitx` publication stateful fake retains bootstrap/CAS/lost-ack tests. Native folder-backed repositories with bare remotes model the observation interactions, including two-clone remote changes and worktree/index state; private production-CLI conformance exercises the same boundary. No GitHub service is needed for migration acceptance.

## Observation and worktree details

1. Pin main using the existing snapshot. Call `publicationBranchTips` once; failed fetch or refreshed-main mismatch refuses. Pin each live worktree HEAD from the same parsed worktree-list result.
2. Validate root components and candidate file kinds, then collect direct-child metadata for all three roots. Share extraction with init. Missing roots are empty; malformed/unreadable roots are unknown and refuse. Symlink roots or candidate documents cannot be followed as authoritative data.
3. Import selected main documents in bulk. Both archive roots feed the existing archived duplicate-selection policy. Keep archived documents unchanged and active conversions unchanged.
4. For related non-main tips, retain actual merge-base active changes/deletions, not a main-versus-tip diff. Obtain changed paths once and resolve their blobs from bulk snapshot data, avoiding one subprocess per file. Feed only active changes to `Branches`. Use the same pinned tips for close anchors. Unrelated tips supply presence/matching evidence rather than an invented merge base.
5. In each live worktree, collect NUL-delimited cached index entries, tracked/untracked status and relevant index flags across all roots. Filter by the same direct-child rule. Inspect direct filesystem entries against index membership so ignored documents are refused too. A staged deletion remains visible even if its physical file is gone. Rename sources and destinations both retain their identity.
6. Preserve Git's content/attribute normalization for ordinary tracked dirtiness. Refuse conflicted index entries and relevant assume-unchanged/skip-worktree entries rather than claiming a trustworthy clean result. Non-issue assets/side bundles remain excluded. Return every actionable dirty issue path, sorted/deduplicated.

## State, concurrency and recovery

Actions: resolve named legacy evidence, rerun dry run, freeze legacy writers, apply the reviewed digest, repeat the existing apply after interruption, reconcile pre-cutover branches after completion.

| Internal state / failure mode | Visible state | Recovery action | Possible loss |
|---|---|---|---|
| Complete observations and no semantic refusal | Reviewed dry-run digest | Freeze writers and apply that digest | None; no publication yet |
| Unpublished non-main evidence or dirty issue document | Refusal naming carrier/path and remedy | Resolve the named legacy evidence, rerun dry run | None automatically; operator discard is explicit |
| Fetch/tree/index/worktree observation unknown, malformed, cancelled or over limit | Observation error; no publication | Resolve observation, rerun dry run | None |
| Main moved during observation or digest differs at apply | Refusal; reviewed plan obsolete | Rerun and review dry run | None |
| Existing exact tracker bootstrap published, main cutover incomplete | Existing retryable migration state | Repeat apply under existing exact-root proof | Existing partial publication retained |
| Valid root and main marker already complete | Existing already-migrated result | Reconcile remaining pre-cutover branches | None |

No new state persists between invocations. `PlanTrackerMigration` owns the new finite semantic decisions; existing bootstrap/cutover authorities own all publication transitions. This change does not add a boolean phase constellation, lock, receipt or automatic cleanup. No worker survives invocation return.

The most dangerous event is evidence changing while a multi-clone migration is observed. Existing explicit freeze of legacy writers remains required; a local repo lock cannot hold editors or other machines. Main CAS protects cutover identity, and apply recollects inventory before checking the reviewed digest. This is not a distributed lock or historical-all-commits scan. Only current tips plus current worktree HEAD/index/filesystem are covered; other machines must run their own inventory before the shared freeze. Reflogs, unreachable objects and other remotes remain outside the current publication-remote contract.

## Operating envelope and architecture checks

This is a rare batch CLI on the shared Mac, not a startup or UI path. Growth drivers are R ref names, T distinct pinned tips, W live worktrees, E tree entries, I issue documents/bytes, C active changes and A legacy close anchors. One remote fetch and ref enumeration; metadata reads O(T × roots), O(E) parsing, bulk issue reads O(I), dirty reads O(W) subprocess groups plus O(local entries), and keyed evidence comparison O(I + total observed issue records). Existing legacy-anchor history work remains O(A × T); this plan does not claim history traversal constant-time. No per-file Git process in the new snapshot/delta path.

Use existing `gitx` snapshot entry, blob and aggregate-output limits. Apply equivalent bounded reads/parsing to the new local observations; over-budget input refuses with the scope named rather than truncating into absence. Do not add an arbitrary fleet-wide tip cap. Keep processing serial, deduplicate equal OIDs, release each tip's bulk bodies after extracting the needed values, and retain metadata only. No latency SLA beyond bounded context cancellation; native tests count subprocesses at 8/16 documents, tips and worktrees. Changing unrelated asset volume must not cause blob reads.

- **ARCH-DRY:** one complete enumerator and root/path grammar; separate consumer policy, no second init scanner.
- **ARCH-PURE:** pure evidence matcher and existing manifest planner; adapters only collect validated facts.
- **ARCH-PURPOSE:** executable root × scope matrix covers the class, including inherited/unrelated/detached and name-excluded carriers.
- **ARCH-MOCK:** existing Git runner seam and publication stateful fake; portable native Git dependency fixtures and private CLI conformance cover actual observation behavior.
- **ARCH-CONSTRAINTS:** counted growth guard and existing bounded Git reads; no recursive filesystem traversal of side bundles.
- **ARCH-SECURE:** raw Git/index/worktree data is untrusted, parsed with NUL grammar and validated paths/modes/OIDs; unknown never becomes empty. No secrets created or logged; all tests use private origins.
- **ARCH-ORDER:** no new durable protocol; apply still reobserves then checks the digest before existing CAS effects. Failures/cancellation authorize no new publication.
- **ARCH-FUNERAL:** no new artifacts outside existing issue/plan/test fixtures; private test repositories are owned by test/temp cleanup, and the normal merge archives the plan/issue. Existing tracker/marker lifetime is unchanged.

## Implementation tasks — one atomic close boundary

### 1. Pure inventory policy and defensive consumers

Files: create `cmd/sdlc/internal/tracker/migration_inventory.go` and `_test.go`; modify `migration.go`, `initialization.go` and their tests. Extract shared classification to `cmd/sdlc/internal/issue/inventory.go` with colocated tests.

- [x] Write pure failing tests using the `IssueInventoryPath`, `MigrationEvidenceRefusals` and `PlanTrackerMigration` strategies below. Preserve the existing active-removal and close-binding oracles.
- [x] Run `go test ./cmd/sdlc/internal/tracker ./cmd/sdlc/internal/issue -run 'TestMigration|TestEmptyTrackerInit|TestIssueInventory' -count=1`; confirm new cases fail for the intended missing observation/policy.
- [x] Implement the typed evidence and keyed matcher; compose with the existing planner and reject added nonempty evidence in `EmptyInitEligible`. Run the same tests green and commit.

### 2. Shared complete observation and main-only import

Files: create `cmd/sdlc/issueinventory.go`, `cmd/sdlc/issuemigrate_inventory.go` and targeted tests; modify `issuemigrate.go` and `issueinit_inventory.go`. Extend existing inventory/init tests where helper extraction changes names, preserving their behavioral oracles.

- [x] Add real-Git failing regressions for both reviewer reproductions and exercise the complete observation strategies below. Assert the resulting imported population reserves the published IDs.
- [x] Exercise `dirtyIssuePaths` against adversarial native worktree/index/filesystem state, including a normalization-aware clean control. Keep the finite root × scope inventory in executable tests rather than repeating it in this plan.
- [x] Run `go test ./cmd/sdlc -run 'TestIssueMigrate|TestMigrationInventory|TestIssueInit' -count=1 -timeout 20m` to establish intended failures, then implement the shared observer and adapter, bulk active-delta reads and complete worktree dirty checks. No repository edits while this root-package test process runs.
- [x] Rerun targeted tests green. Preserve the existing remote-refresh/main-pin/close-pin/rename/reconcile/bootstrap retry tests. Commit.

### 3. Cost, integration contract, documentation and acceptance

Files: migration inventory tests; `atlas/workflow/issue-tracker-migration.md`; README migration discovery paragraph if its current guidance changes; issue Log and `workshop/lessons.md`.

- [x] Add a counted N/2N Git-process guard for 8/16 issue documents, distinct tips and worktrees, plus unrelated asset-volume control. Assert no per-document subprocess growth, fixed per-worktree overhead and at-most-linear tip/worktree growth. Record measured counts rather than latency claims.
- [x] Test before/after remote refs and caller HEAD/index/status on all refusals: local fetch tracking refs may refresh by the existing explicit mapping, but tracker/main publication and caller checkout/index must not change. Test an apply whose evidence changes after the reviewed dry run; it must refuse before any tracker creation.
- [x] Build a private production `sdlc` binary and exercise both reproduced cases against real bare origin, then resolve their evidence onto main, run dry-run/apply, verify archived cards reserve their IDs, repeat apply idempotently, and use a normal tracker issue-new/claim workflow. Preserve a command/result artifact for the Log. No stage, per TL.
- [x] Update the atlas operator inventory/refusal procedure and verification pointers. Explain intentional stale-archive refusal, detached/worktree coverage, all three roots, main-only imports and the remaining freeze/other-machine boundary. Record the semantic lesson; existing exhaustive regression policy is the mechanical guard, not a lexical helper-name lint.
- [ ] Run focused tests plus `go test ./pkg/golint ./cmd/sdlc/internal/recovery -count=1`, inspect `git diff --check`, then push the tested branch and queue the required recorded check with `sdlc local-check run --name suite`. Close through `sdlc close` with actual evidence and the TL-owned project acknowledgment; its fresh-context review is the sole implementation review boundary.
- [ ] On final recorded branch PASS and SHIP, open/update the PR and enter `sdlc merge --yes`; let it integrate current main and run the authoritative integration suite. Verify archived card/artifacts/cleanup and report the landed PR to the hub.

## Risky-function adversarial strategies and independent guards

| Function / boundary | Adversarial strategy | Mechanical guard |
|---|---|---|
| `issue.IssueInventoryPath` | Misclassify external paths and root boundaries | Pure root × exact-parent tests, malformed-path fuzz seeds and native negative controls |
| `MigrationEvidenceRefusals` | Unrepresented or mismatched records disguised as published evidence | Pure role/ID/blob population tests and order-invariant refusal assertions |
| `PlanTrackerMigration` / `EmptyInitEligible` | New observation data bypasses semantic refusal or emptiness defense | Existing rename/prose/removal/close contracts plus added-input rejection tests |
| `observeIssueInventory` / `publicationBranchTips` | Omit a carrier or follow a moving name after pinning | Native root × scope/ref-category matrix, two-clone refresh tests and pinned-read assertions |
| `branchIssueEdits` | Fabricate edits from stale main comparisons or read unpinned data | Related stale-branch and pinned-ref regressions over actual Git histories |
| `dirtyIssuePaths` | Hidden, ambiguous or malformed worktree evidence looks clean | Native index/filesystem matrix, parser fuzz seeds, normalization control and refusal-preservation assertions |
| `migrationInventory` / `runIssueMigrate` | Failed or changed observations authorize an empty cutover | Apply-before-publication refusal checks and unchanged exact-root/marker retry tests |
| `observeIssueInventory` / `migrationInventory` | Per-file subprocesses or duplicate-tip rescans | Counted N/2N process tests and duplicate-OID controls |

## Review and execution boundary

No implementation before TL approval of this plan and `sdlc change-code` plan-quality admission. This is full flow with one close boundary, not artificial M1/M2 labels. Estimate is derived only after plan quality passes. Another design change is recorded in Revisions; repeated changes to one mechanism trigger a whole-mechanism re-plan under the worker contract.


## Revisions

### 2026-10-22 — PQ-1 function-level test strategy

Plan-quality round 1 accepted the architecture and scope, but requested exact risky-function names and compressed test prose. Named the shared classifier, complete observer, active-delta reader and dirty observer, then replaced redundant case inventories with one adversarial strategy and independent guard per function. No behavior, authority, state or scope changed. Detailed input enumeration belongs in executable tests.

### 2026-10-22 — bounded verification corrections

Preserved literal POSIX backslashes accepted by the existing filename grammar, with behavioral RED/GREEN and native unusual-filename regressions. Canonicalized the legacy path-with-spaces fixture without weakening its byte-exact quoting oracle. Native growth and uncertainty guards passed; documentation now maps the approved rules. Issue checklist rows describe implemented outcomes while binary suite/review/landing remain explicit outstanding gates. No authority, state or scope change.

### 2026-10-22 — post-close fixture correction and Minor cleanup

A recorded suite exposed an asset-conversion fixture whose stale archive copy now correctly triggers the approved earlier inventory refusal. Publish identical archive-reference bytes before the fixture fork, retain main-only asset/attribute arrival, and preserve the existing conversion/no-mutation assertion. Remove the two unused compatibility paths from BR-1 and point their tests at the complete observer. No scope, state, authority or policy changes; targeted RED/GREEN and the 124.191s regression group verify the bounded correction. Fresh suite and narrow re-close are required on the new head.
