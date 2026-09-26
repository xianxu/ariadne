# Issue cards on a tracker branch Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development for bounded tasks or superpowers-executing-plans for session-dependent integration. Steps use checkbox syntax for tracking.

**Goal:** Make cards authoritative on `issue-tracker`, keep details with their authoring/code branch, and permit claim only after initial details land on main, without copying branch commits or stranding issue bookkeeping on resting branches.

**Architecture:** A shared card model composes authoritative metadata with branch-local details. A tracker repository reuses Git snapshot/CAS transactions; explicit creation, transfer, close and landing transitions preserve partial progress. Existing readers and alternate writers move together at a frozen, one-pass cutover.

**Tech Stack:** Go, Cobra, Git object/ref plumbing, existing Git/GitHub process seams and stateful fakes, CUE vocabulary, Markdown artifacts.

**Status:** Operator approved implementation on 2026-09-25. M1 accepted through its boundary gate (FIX-THEN-SHIP; minor artifact whitespace corrected). M2 implemented in the issue worktree and submitted to its boundary review. M3–M4 are not implemented. No production migration or consumer activation has begun.

## Core concepts

### Pure entities

This inventory distinguishes delivered M1 foundation from planned later work.
“Planned” is not a claim of presence or modification in the M1 review window.

| Name | Kind | Lives in | Delivery status |
|---|---|---|---|
| Card / owned-field schema | PURE | `cmd/sdlc/internal/issue/card.go`, `construct/vocabulary/issue.cue` | M1: new / modified, delivered |
| DetailMirror | PURE | `cmd/sdlc/internal/issue/mirror.go` | M1: new, delivered |
| Creation / transfer transitions | PURE | `cmd/sdlc/internal/tracker/creation.go`, `transfer.go` | M1: new, delivered |
| Close generation / landing transitions | PURE | `cmd/sdlc/internal/tracker/completion.go` | M1: new, delivered |
| Typed recovery receipt / shared transition engine | PURE | `cmd/sdlc/internal/tracker/receipt.go` | M1: new, delivered |
| Migration manifest | PURE | `cmd/sdlc/internal/tracker/migration.go` | Planned M4: file absent, not delivered |
| Activity event selection | PURE | `cmd/sdlc/internal/activetime/commit.go` | Planned M3: existing file unchanged in M1 |

Card owns ID, status, started, created/updated dates, estimate/actual hours, GitHub linkage and canonical title. Its Problem is the original report. Details retain editable Problem, Spec, Done when, Estimate explanation, Plan, Log, Revisions, deps, target, flow and review anchors. Define ownership once; unknown detail fields remain untouched, unknown tracker schema versions refuse. No new lifecycle statuses. Read `vocabulary` skill and the complete CUE model before changing the model.

A DetailMirror is a projection of one card blob, not an independent metadata source. Store the exact last mirrored card blob OID in details. Compare locally mirrored fields against that immutable blob first: changed local owned fields refuse, untouched old projections refresh from latest card. Refuse missing/unavailable baseline instead of guessing. Preserve absent-versus-empty field semantics and YAML scalar types. Problem is excluded from mirror enforcement. Include title in the canonical projection and add a supported title setter; details may independently revise Problem.

Each card has one stable ID/path; one or more detail checkouts can mirror it, but claim serializes work on an eligible issue. Close generation binds the ID to reviewed HEAD, close-evidence commit, and repository identity; status alone cannot establish ownership of a PR. Pure transitions produce declared effects; callers cannot directly set transaction stages. These entities centralize ownership and remove repeated local-file status interpretation (ARCH-DRY, ARCH-PURE).

### Integration points

| Name | Kind | Lives in | Delivery status | Wraps |
|---|---|---|---|---|
| Tracker repository / card snapshot reader | INTEGRATION | `cmd/sdlc/internal/tracker/repository.go`, `reader.go` | M1: new, delivered; card inventory only | `gitx.TrunkFile`, pinned Git refs/trees |
| Git snapshot/CAS and bootstrap | INTEGRATION | `cmd/sdlc/internal/gitx/trunkfile.go`, `updatemany.go`, `snapshot.go`, `refbootstrap.go` | M1: modified / new, delivered | Git subprocess boundary |
| Bounded output / shared process-group cleanup | INTEGRATION | `cmd/sdlc/internal/gitx/boundedoutput.go`, `cmd/sdlc/internal/processgroup/` | M1: new, delivered; judge wrappers updated | subprocess IO and cancellation |
| Detail transfer adapter | INTEGRATION | `cmd/sdlc/issuemovedetail.go`, `internal/tracker/transferop.go` | M2: new, delivered | tracker, main publication, index/worktree and recovery refs |
| Receipt driver / candidate steps / recovery refs | INTEGRATION | `internal/tracker/drive.go`, `createop.go`, `store.go`, `internal/gitx/candidate.go`, `recoveryref.go` | M2: new, delivered | receipt engine, Git candidate publication, local refs |
| Transfer guard | INTEGRATION | `cmd/sdlc/transferguard.go` | M2: new, delivered | `git merge-tree`, tracker handoff records |
| Composed issue reader | INTEGRATION | `cmd/sdlc/issuerecord.go` | Planned M3: absent, not delivered (M2 verbs read card and details directly) | M1 card reader plus selected detail location |
| Completion adapter | INTEGRATION | `cmd/sdlc/trackercompletion.go` | Planned M3: absent, not delivered | close evidence, existing GitHub landing identity and archive transaction |
| Migration command | INTEGRATION | `cmd/sdlc/issuemigrate.go` | Planned M4: absent, not delivered | repository inventory, bootstrap, mirrors, cutover marker |
| Publication fake | INTEGRATION test double | `cmd/sdlc/internal/gitx/commitpublication_fake_test.go` | M1: modified, delivered | immutable objects, multiple refs, rejection/lost acknowledgement |

Reuse the common-dir repository lock. Correct the existing cancellation gap: `TrunkFile` calls `runGitIn`, which supplies `context.Background()` (`trunkfile.go:60`), and `UpdateMany` accepts no context (`updatemany.go:68`). Add `NewTrunkFileContext(ctx, dir, remote, branch)` and store a non-nil context; an instance runner routes all fetch/read/tree-build/push/confirmation through `runGitInContext`. Keep the old constructor as a compatibility adapter only for unmigrated callers. Tracker commands pass Cobra command context through repository/lock/preparation boundaries. Cancellation never becomes predicate absence or confirmed rejection. Bound process-group termination and pipe draining to five seconds using existing command cancellation helpers. Cancelled pushes retain Unconfirmed receipts; recovery uses a later invocation, never a detached worker. TrunkView inherits its owner's context. Linked worktrees share refs and locks; independent clones use remote CAS. Use remote-tracking refs and object reads rather than a hidden worktree. Read-only card snapshots may be materialized for navigation, never editing authority.

Use the configured publication remote/main identity, never an invented `origin`. Fetch `refs/heads/issue-tracker` explicitly in ordinary and CI clones. Reuse `UpdateMany` and `TrunkView`; do not use `PublishCommit` for cards. Bootstrap an orphan tracker history with a format manifest and imported cards, using expected-absence CAS. Normal writes operate only on an initialized tracker.

## Transaction contracts

### Creation and branch residency

`issue new` prepares local content, reserves the card on fresh tracker, and materializes details for the confirmed ID. A failed local write after reservation leaves a recoverable incomplete creation, never a second allocation on retry. Persist operation identity plus candidate/card snapshot before effects; identical content is not proof of ownership.

On a feature branch, details are ordinary local files and may be committed with ongoing work. On a resting checkout, creation leaves local details uncommitted until `move-detail` publishes them or the operator prepares an authoring branch. SDLC must not create resting-branch checkpoint commits. `start-plan`, after a valid claim, prepares an issue branch before authoring/committing design when invoked on rest; `change-code` reuses that branch. Preparation fetches/pins main, requires a clean resting checkout whose HEAD is an ancestor of that pinned main, then creates the issue branch at pinned main without moving the resting ref. Verify the matching details at that base before switching. Dirty/ahead/divergent rest refuses with preservation/reconciliation instructions; never drop local commits. A different active issue branch refuses planning for this issue with a next action to use another checkout. This resolves stale-slot readiness and the old sync-before-branch divergence without silently switching away from unrelated work.

Claim reads a fresh tracker card and a pinned fresh main tree, verifies the matching ordinary details file exists, then conditionally updates the card from open to working. Revalidate relevant readiness before publication; card-only/local-only details refuse. Absence of deps for an uncreated issue is unknown, not an empty list.

### `move-detail`: ordinary Git history, no rewritten commits

For an initial file absent at the source branch/main merge bases, publish the current details as a new main-native commit D; do not merge D into the source branch. Remove the source file with a narrow branch commit R after confirmed publication. Relative to the absent baseline, add then remove is net-zero: merging R into main retains D and any later edits E. A real-Git experiment during planning confirmed both properties. Joining D into the source before deleting instead causes deletion or modify/delete conflict and is prohibited.

Preconditions: no Git operation in progress; exact repository/remote/main/source HEAD pinned; destination absent; ordinary source file or proven absence; source absent at every merge base; no staged/unstaged split for the source. Preserve unrelated index/worktree entries. Read/hash source and index before publication and revalidate before removal; changed source remains untouched with recovery instructions. If source has no tracked version, remove only its verified bytes and do not create a deletion commit. On rest, require HEAD to fast-forward to publication (no local-only commits); reconcile the verified source while fast-forwarding, with no separate deletion commit. It then appears as the newly published destination copy.

Record an operation receipt before publication, pinning source bytes and identity in a recovery ref. Also publish a single initial-handoff record on the card before main publication: operation token, repository identity, source branch/base/head, source path/blob and intended destination. Record the main publication OID after confirmation, before local removal. These internal transaction fields are not editable mirrored user metadata. Include operation provenance in publication/removal commit trailers. A fresh clone discovers relevant handoffs from tracker records plus source ancestry, not private refs or branch name alone. Rebased source histories that cannot prove the association refuse until an explicit reconciliation records the new branch binding. A retry may accept an existing destination only with its exact operation provenance and confirmed ancestry; independent existing details refuse. No-source mode derives initial details from the card. A network timeout is an unknown outcome, not failure; do not remove the source while unknown. A new card/blob version after preparation requires revalidation, not overwriting with stale metadata.

After transfer, an ordinary merge from main may naturally bring the destination file into the original branch. Never auto-delete it again. At close/PR/push/merge, inspect outstanding transfer receipts and the prospective merge tree: it must preserve the latest destination content for each handed-off ID (including the archived path if its new owner already completed it). Refuse deletion, overwrite or unresolved merge. This guard is about transferred issues only; edits by the new claimed owner use its own issue branch. Receipts are branch-associated and removed only after confirmed landing/abandonment of that source branch. Test direct merge, merge-from-main, squash landing and source rebasing explicitly.

### Completion and recovery

Close first validates current mirror/card generation and review read set. Commit branch-local evidence (Log, review anchors/sidecars, project/atlas work) before publishing codecomplete on the card. Bind the card completion record to the exact reviewed HEAD and evidence commit; a self-referential commit OID is avoided by writing it only to the subsequent tracker transaction. Remote failure leaves evidence and a retryable pending transition. Never emit a successful close before confirmed card publication.

Merge/push select owned issues using the completion binding, not a codecomplete edit in PR ancestry. Confirm the merged PR identity/base/head/merge OID (or exact direct-main reachability), publish done for those generations, then archive their details/plans using the existing archive transaction. Treat done-write and archive as separately recoverable steps. A reopened/different generation refuses stale completion. Reconciliation on the next mutating SDLC command resumes a proven pending operation; reads report pending work without unexpectedly mutating remote state. Existing review snapshots must include card blob identity; unrelated card changes do not invalidate a review.

| State | Event | Effect / next state |
|---|---|---|
| Prepared | publish succeeds and ownership confirmed | Published; permit local finalization |
| Prepared | confirmed ref race | re-read, revalidate, bounded retry; allocation recomputes ID |
| Prepared | lost acknowledgement / cancellation | Unconfirmed; preserve receipt/source, probe on retry |
| Unconfirmed | candidate/provenance confirmed reachable | Published; resume once |
| Unconfirmed | probe fails | stay Unconfirmed; no cleanup |
| Published | source or generation changed | retain evidence, refuse destructive finalization |
| Published | matching finalization succeeds | Finalized; release temporary source recovery material |
| Codecomplete | exact reviewed generation lands | publish Done, then archive |
| Codecomplete | new review/code/reopen generation | refuse stale landing completion |
| Done | archive interrupted | resume owned archive; never revert card to codecomplete |

Implement separate tagged creation/transfer/completion state types using this common outcome vocabulary, not independent booleans. Process death, cancellation, concurrent writers and delayed responses are explicit test events. No worker continues beyond command cancellation/return. Reuse existing three-attempt publication bound. ARCH-ORDER.

## Operating envelope and durable residue

- CLI operations, not keystroke paths. Initial design budget: one tracker fetch per command, one tree listing per snapshot, batched blob reads, no fetch per card. Exercise 10,000 cards and 100 active details in benchmarks; target under one second local snapshot processing on a developer machine, report hardware/results rather than claiming it now. Network uses existing command cancellation/timeouts; stale reads must be visibly labelled and cannot authorize writes. ARCH-CONSTRAINTS.
- Cards persist for repository lifetime: one small record per issue plus Git history. Derive max ID, retain terminal cards, and measure scan/storage cost in the benchmark; no counter file. Historical migration seeds every used ID. This is the recommended allocation choice adopted by this proposed plan.
- Tracker format manifest persists per repository. Mirror baseline blobs stay reachable through tracker history. Each transferred card retains one initial-handoff record, bounded to one per initial creation; its immutable history supports fresh-clone verification even after source-clone loss. Recovery refs retain source snapshots until confirmed finalization; `sdlc issue recovery list` reports remaining local operations and `sdlc issue recovery reconcile --issue N` resumes or cleans confirmed completed operations through the same transition model. Never age-delete uncertain operations. No background daemon.
- Inputs are untrusted: validate IDs, OIDs, ordinary file modes, repository-relative roots, duplicate IDs/YAML keys, schema versions, receipt size/shape and exact repository identity before effects. Reject symlinks and malformed trees; missing and unreadable differ. Fixtures isolate HOME/config/remotes and disable hooks/signing; no new credentials. ARCH-SECURE, ARCH-FUNERAL.

## Consumer inventory and ownership

| Consumer | Required source / change |
|---|---|
| `issue.go`, `issueids.go`, `claim.go`, `claimdecision.go`, `setstatus.go` | card authority, confirmed creation, main readiness, explicit metadata setters |
| `startplan.go`, `changecode.go`, `changecode_flow.go` | mirror validation, early design branch, ordinary checkpoint commits; no copied publication |
| `close.go`, `reviewstate.go`, `milestoneclose.go`, `publishgate.go` | composed inputs and card-bound review snapshots/completion |
| `merge.go`, `push.go`, `landing.go`, `landingarchive.go`, `archivepolicy.go` | exact completion binding, transfer guard, done/archive recovery |
| `state.go`, `issuefiles.go`, `internal/fleet/issues.go` | current cards, explicit detail availability; no local mirror status authority |
| `projectstatus.go`, `internal/project/metadata.go`, `projectclose.go` | status/hours from card; deps and target from selected details |
| `resolve.go`, `pr.go`, `repoguard.go`, `validategate.go` | details navigation remains; card-only visibility, GH linkage, distinct card/detail schemas |
| `actual.go`, `internal/gitx/window.go`, `internal/activetime/commit.go`, `compute.go` | selected code history plus tracker claim/close events, deduplicated by OID |
| `Makefile.workflow`, `scripts/close-issue.py`, `scripts/merge-checks.d/40-duplicate-issue-id.sh` | delegate writers to SDLC; tracker ID validation; remove shell allocation/status fallbacks |
| `scripts/parallel-checks.sh`, `scripts/pre-merge-checks.sh` | correct detail discovery and card-aware gate input |
| `AGENTS.base.md`, `construct/local/issues/SKILL.md`, `construct/datatype/project.md` | new ownership/checkpoint guidance, no manual hours editing |

Complete the sweep with `rg -n 'issue sync|issue publish|PublishCommit|estimate_hours|actual_hours|workshop/issues' cmd/sdlc scripts Makefile.workflow AGENTS.base.md construct/local construct/adapted construct/datatype`, classifying each match as body/history text, authoritative reader, writer or generated consumer. Do not mechanically replace historical descriptions. Remove live compatibility writers, not merely their README links. ARCH-PURPOSE.

## Chunk 1: Model and tracker foundation — M1

### Function-level adversarial test contract (all milestones)

| Production function | Strategy and independent invariant |
|---|---|
| `SplitCard`, `ParseCard` | Fuzz malformed YAML/Markdown; reject ambiguous ownership and preserve all unowned bytes. |
| `RefreshMirror` | Generate baseline/local/current projections; reject changed owned local values, refresh only unchanged projections. |
| `BootstrapTracker`, `Repository.UpdateCard` | Reproducibly interleave two writers and uncertain transport results against stateful Git; one authoritative winner, unique IDs, no lost updates. |
| `TrunkFile.run`, `runIssueNew` cancellation path | Cancel from the production Cobra command while a controlled child Git process blocks; process group and pipes end within five seconds and uncertain publication retains recovery evidence. |
| `StepCreation`, `StepTransfer`, `StepCompletion` | Generate event sequences, including interruption at each declared effect; assert no destructive action without confirmed ownership and no accepted stale generation. |
| `runClaim` | Drive concurrent commands against independent snapshots; only a fresh-main-ready open card can be reserved once. |
| `PreparePlanningBranch` | Generate rest/main ancestry and dirty-state combinations; accepted branch contains eligible details and never changes resting refs. |
| `CheckTransferredPaths` | Generate DAGs across publication and later owner changes; reject any prospective merge losing authoritative transferred content. |
| `ReadIssueRecord`, `LookupRepoIssues` | Vary card and detail generations independently; source selection respects field ownership and missing detail information remains unknown. |
| `SelectActivityEvents`, `computeActual` | Generate overlapping selected histories; each OID counted once and claim engagement survives off-branch storage. |
| `FinalizeTrackerClose`, `SelectCompletedIssues` | Mutate bound repo/head/evidence between review and effect; refuse stale ownership rather than completing unrelated work. |
| `PlanTrackerMigration`, `ApplyTrackerMigration` | Generate legacy populations and interrupt each phase; no lost IDs/evidence and no activation from incomplete or contradictory inputs. |

Real-Git conformance runs with the normal Go suite on every relevant PR and
before each milestone gate. Existing GitHub stateful landing tests run on every
PR; live read-only GitHub response-schema conformance runs before rollout and
whenever the consumed gh/GitHub contract changes, using an explicitly selected
fixture PR and credentials already configured by the operator. No live mutation
is needed for that check. Record versions and fixture identity with results.

### Task 1: Card, mirror and vocabulary contract

Files: create `cmd/sdlc/internal/issue/card.go`, `card_test.go`, `mirror.go`, `mirror_test.go`; modify `construct/vocabulary/issue.cue`, `pkg/vocab/vocab.go`, generated `pkg/vocab/issue.json`, `pkg/vocab/vocab_test.go`.

- [x] Write failing property/fuzz tests for `SplitCard`, `ParseCard`, `RefreshMirror` per the function strategy contract below; assert preservation independently of serialization.
- [x] Run `go test ./cmd/sdlc/internal/issue ./pkg/vocab -count=1` and record intended failures before implementation.
- [x] Add card discovery/ownership alongside existing `discovery.home` (which remains details); implement pure projection comparison and refresh. Expose typed validation errors with setter next actions.
- [x] Regenerate using the repository vocabulary generation path; run the tests above and `git diff --check`; commit explicit paths as `#252 M1: model: split card authority from details`.

### Task 2: Ref-backed repository and transaction evidence

Files: create `cmd/sdlc/internal/tracker/{repository,reader,creation,transfer,completion}.go` and colocated tests; create `cmd/sdlc/internal/gitx/refbootstrap.go` and tests; modify `trunkfile.go`, `updatemany.go`, `commitpublication_fake_test.go` only where existing APIs lack outcome/evidence access.

- [x] Write failing model-based tests for `BootstrapTracker`, `Repository.UpdateCard`, `StepCreation`, `StepTransfer`, `StepCompletion`; run matching sequences against the stateful fake and disposable real Git.
- [x] Run `go test ./cmd/sdlc/internal/gitx ./cmd/sdlc/internal/tracker -count=1`; verify new tests fail for the intended missing behavior.
- [x] Implement snapshot reads, atomic expected-version card writes and typed operation receipts. Preserve existing CAS semantics; generalize the stateful fake to main and tracker refs, not duplicate it per command.
- [x] Repeat tests, add cardinality/IO benchmarks, and document tracker model in new `atlas/workflow/issue-tracker.md` with `atlas/index.md` link. Committed as `2144608`; verification evidence is recorded in the issue's M1 review-submission Log (including the baseline CLI-suite failure and measured performance limits).

**M1 accepted:** `sdlc milestone-close --issue 252 --milestone M1` round 2
disposed BR-1/BR-2 and returned FIX-THEN-SHIP with one minor artifact-whitespace
finding (BR-3), corrected in the boundary commit. The issue's M1 checkbox is
the acceptance record. Foundation remains unactivated pending cutover.

## Chunk 2: Creation and handoff — M2

### Task 3: Commands, setters and early design branch

Files: modify `cmd/sdlc/{issue,issueids,claim,claimdecision,setstatus,startplan,changecode,changecode_flow}.go`; add `issuerecord.go`, `issuemetadata.go` and tests; update existing `issue_test.go`, `claimremote_test.go`, `startplan_test.go`, `changecode_test.go`.

- [x] Test `runIssueNew` and `runClaim` through the production command seam with controlled publication schedules; assert one reserved identity/claim and no readiness from local-only details.
- [x] Test `PreparePlanningBranch` with generated checkout/ref relationships; assert fresh-main details are present and resting refs/unrelated work are unchanged.
- [x] Run focused tests with `go test ./cmd/sdlc -run 'Test(Issue|Claim|StartPlan|ChangeCode)' -count=1`; record red results.
- [x] Route through the tracker (composed reader deferred to M3 with its consumers). Card setters are `sdlc issue set-title` / `set-estimate` / `set-github` (the delivered CUE model's setter names; see Revisions); ID/dates/started/status/actuals remain owned by their specific lifecycle verbs. Estimate explanation validation stays in details.
- [x] Remove sync/publish behavior from implementation entry; keep ordinary explicit-path local checkpoints on the issue branch. Run focused tests and commit.

### Task 4: Initial-details transfer and merge protection

Files: create `cmd/sdlc/issuemovedetail.go`, `issuemovedetail_test.go`, `transferguard.go`, `transferguard_test.go`; modify `issue.go`, `pr.go`, `publishgate.go`, `merge.go`, `push.go`.

- [x] Build real-Git fixtures for A→B→R versus A→D→E, asserting main gets only details, code stays unshipped, source removal is scoped, and eventual merge preserves E exactly. Include main-rest fast-forward and no-local-file modes.
- [x] Test `StepTransfer` with reproducible interruption/event sequences; source removal requires confirmed owned publication and an unchanged source fingerprint.
- [x] Property-test `CheckTransferredPaths` over generated branch/main DAGs and fresh-clone replay; prospective merge preserves current destination state or refuses without mutation.
- [x] Implement the pinned transfer and merge checks described above, including `issue recovery list/reconcile` registration and tests in `issuemovedetail_test.go`. Run `go test ./cmd/sdlc -run 'Test(MoveDetail|Transfer|IssueRecovery)' -count=1`, then affected PR/merge/push suites; no production remotes in tests.
- [x] Update tracker atlas and README command discovery; commit.
- M2 acceptance is its `sdlc milestone-close` boundary review, recorded by the issue's M2 checkbox (not ticked here).

## Chunk 3: All readers, close and landing — M3

### Task 5: Composed readers and activity evidence

Files: modify every reader in the inventory, with existing colocated tests; add `cmd/sdlc/issuerecord_test.go`, `internal/tracker/reader_test.go`.

- [ ] Test `ReadIssueRecord` and `LookupRepoIssues` with independently varied authoritative snapshots and projections; card metadata wins while detail-owned values remain intact.
- [ ] For cross-issue/dependency reads use pinned main details; for the active issue use its checked-out details. Preserve archive navigation and allow card-only inspection without fabricating an editable file.
- [ ] Test `SelectActivityEvents` and `computeActual` over generated code/tracker histories; assert selected-source-only membership, deduplicated OIDs and preserved engagement intervals.
- [ ] Implement and run `go test ./cmd/sdlc ./cmd/sdlc/internal/fleet ./cmd/sdlc/internal/project ./cmd/sdlc/internal/activetime -count=1`; commit after the new assertions pass.

### Task 6: Close generation, exact landing and recovery

Files: create `cmd/sdlc/trackercompletion.go`, `trackercompletion_test.go`; modify `close.go`, `reviewstate.go`, `milestoneclose.go`, `publishgate.go`, `landing.go`, `landingarchive.go`, `merge.go`, `push.go` and their tests.

- [ ] Test `StepCompletion` and `FinalizeTrackerClose` with reproducible review/remote-generation interruptions; stale verdicts and uncertain writes never become accepted completion.
- [ ] Test `SelectCompletedIssues` against unrelated ancestry and altered generation bindings; only proven reviewed/landed generations can be selected, independent of mirrored status.
- [ ] Implement explicit completion bindings and recovery through existing landing identity/review guards; retain exact reviewed-head checks and scoped archive behavior.
- [ ] Run `go test ./cmd/sdlc -run 'Test(Close|Milestone|Review|Publish|Landing|Merge|Push|Archive)' -count=1`; update `atlas/workflow/{issue-lifecycle,ledger-landscape,sdlc-binary}.md`, commit and close M3.

## Chunk 4: Cutover tooling, instructions and end-to-end proof — M4

### Task 7: One-pass migration and obsolete surface removal

Files: create `cmd/sdlc/issuemigrate.go`, `issuemigrate_test.go`, `internal/tracker/migration.go` and tests; modify `issue.go`, `repoguard.go`, `validategate.go`, `issuelintids.go`, portable scripts/Makefile/instructions from the inventory; delete retired `issuesync`/`issuepublish` command code only after checking remaining callers.

- [ ] Add `sdlc issue migrate --dry-run` and `--apply` with an exact inventory manifest: active/archive IDs, duplicates, source OIDs, all worktrees, active branches, legacy local-only issue commits, schema version and proposed tracker root. Default dry-run never changes refs/files.
- [ ] Test `PlanTrackerMigration` and `ApplyTrackerMigration` against generated legacy populations and interrupted phases; preserve every used ID and valid completion, refusing ambiguous activation. Use synthetic history fixtures, not live workshop/history.
- [ ] Import existing codecomplete generations only after reconstructing their binding from the legacy close-transition commit, accepted review evidence and pinned PR head (or direct-main landing target), using existing publish-gate validation. Migration refuses an unprovable binding before activation, with the exact issue/anchor and next action to reopen/reclose under the old workflow during the freeze. Never silently omit the issue, invent acceptance, or change its status. Prove the valid imported PR can next merge to done/archive; prove the invalid case blocks cutover until repaired. Include these checks in the operator cutover manifest.
- [ ] Define cutover sequence: freeze all writers; inventory and checkpoint every checkout; refuse unresolved legacy divergence; publish tracker snapshot; commit/publish ordinary main mirror conversion with format marker; reconcile each inventoried branch against its captured old card projection; verify and activate new writers. No automatic discard/reset/rebase of user work. A late/unknown checkout must reconcile or refuse before mutation.
- [ ] New binary refuses writes when cutover marker and tracker generation disagree. Old binaries cannot be made to honor a new marker: the operational freeze must include aliases, scripts and already-running agents until all entrypoints are upgraded. Retry resumes a matching manifest; after tracker writes resume never roll it back to the import snapshot.
- [ ] Remove/delegate shell/Python writers and `issue sync`/generic copied issue publication entrypoints. Update CI fetch/ID checks and all active agent guidance, including adapted skills that prescribe sync. Use xx-construct guidance for substrate edits and preserve generated-source boundaries.
- [ ] Run `go test ./cmd/sdlc/... ./pkg/vocab/... -count=1`, vocabulary generation/conformance checks, and existing shell/Python checks covering modified portable entrypoints. Run the consumer sweep and classify remaining matches in the issue Log.

### Task 8: Full slot cycle and rollout package

Files: create `cmd/sdlc/tracker_e2e_test.go`; update `README.md`, `atlas/workflow/{issue-tracker,issue-sync,workspace-branching,artifact-hierarchy,ci-merge-check,base-layer}.md`, `atlas/index.md`; add `atlas/workflow/issue-tracker-migration.md`.

- [ ] Exercise two cloned slots and a bare remote: new → details publication → claim → branch/design → change-code → close → PR/merge → archive → resting refresh. Include a spin-off handoff while original code stays unshipped, a competing claim and later edits to the transferred issue.
- [ ] Assert main is unchanged by card-only operations, no copied source-commit publication, card states current across slots, original PR has no issue-file conflict, and refreshed rest is exactly zero ahead/behind. Add fresh CI clone without tracker ref and disconnected-read/mutation cases.
- [ ] Run full Go suites above plus focused real-Git race/conformance tests; publish measured 10k-card benchmark results and verify process counts remain bounded. Update atlas at this boundary, not after rollout.
- [ ] Write exact operator cutover/recovery procedure, including owner-binary build and generated instruction propagation. Inventory participating downstream repos read-only under their local instructions before scheduling any freeze. Product-owned consumers needing changes require issues and review in their owning repos; do not call ariadne tests proof of peer integration.
- [ ] Record M4 verification and close its review boundary. Actual fleet migration is a separately coordinated operational step using the tested command; do not freeze another live session or migrate peer state without that coordination. Issue closure requires the migration acceptance criterion to be satisfied, not merely the command to exist.

## Approval and execution

The operator approved this plan on 2026-09-25. Each M1–M4 row is a mandatory SDLC review boundary; do not add a second code reviewer at those boundaries. Run `sdlc change-code --issue 252 --worktree=yes` and satisfy its plan-quality/estimate gates. Use an isolated checkout so unactivated base-layer changes do not become live in primary downstream consumers.

## Revisions

### 2026-09-25 — Fresh plan review

Reason: the first review found three missing cross-checkout/cutover contracts.
Pinned start-plan to fresh main with safe-rest preconditions; moved durable
handoff provenance onto the tracker so source-clone loss cannot disable merge
protection; required reconstruction of legacy codecomplete bindings from proven
review evidence or refusal before activation. Named the recovery command and
added corresponding fixtures. Second review approved Chunks 1–4 with no
Important/Critical findings. Operator approval remains pending.

### 2026-09-25 — Implementation approval and plan-quality refinement

The operator said to continue, approving execution. Gate round 1 identified
PQ-1 (context propagation through the legacy Git transaction seam), PQ-2
(function-level adversarial test strategies), and PQ-3 (conformance cadence).
The plan now explicitly migrates transaction execution to command contexts,
names risky production functions and their invariants, and states recurring
conformance triggers. These refinements preserve the approved behavior.

### 2026-09-25 — M1 implementation checkpoint

The implementation gate passed and the approved model is now implemented in
the isolated issue worktree. Updated execution status and Task 1 progress;
the transaction foundation remains in progress and no consumers are activated.

### 2026-09-25 — M1 boundary inventory correction (BR-1, BR-2)

Reason: the review could not distinguish project-wide planned entities from
delivered M1 entities, and the final task row combined completed verification
with the review gate itself. Swept both concept tables: every row now names its
kind and delivered/planned milestone, including absent future files and unchanged
future modifications. Split implementation evidence from acceptance: tests,
benchmarks, atlas and commit are complete; the mandatory M1 gate is still pending
and remains represented by the unchecked issue milestone. No code or scope change.

### 2026-09-25 — M1 acceptance

Round 2 accepted the corrected inventory and checklist (BR-1/BR-2 addressed).
Fixed the remaining minor trailing whitespace in the generated review artifact
before the boundary commit, as the FIX-THEN-SHIP protocol requires. Updated
current execution status; no production activation or implementation scope change.

### 2026-09-25 — M2 implementation

Reason: executing Chunk 2 exposed mechanics the plan left implicit.
- `UpdateManyPrepared` fuses prepare/push/retry, so the receipt engine could not
  own retries or resume. Added candidate steps (prepare/push/probe) and one
  `tracker.Drive` dispatcher; an unmoved-tip probe re-pushes the identical
  candidate under the same lease to settle a delayed push (ARCH-ORDER).
- Card stages of an operation re-derive their owned envelope change from the
  current card (`PrepareCardChange`) instead of CAS against the last confirmed
  blob: once details reach main another thread may claim, and the engine forbids
  changing `CardOID` on revalidation, so a fixed-blob record stage could never
  finish. This is revalidation, not stale overwrite.
- Mirror refresh never edits the resting branch (a generated test showed claim
  dirtying rest and blocking start-plan); change-code refuses rest for
  tracker-era details and points at start-plan. Mirror presence is a textual
  check so malformed frontmatter cannot bypass validation.
- Setters follow the delivered CUE model (`set-title`, `set-estimate`,
  `set-github`) rather than a generic `issue set --field`.
- `recovery reconcile` releases a creation that published nothing instead of
  re-rendering its draft; the operator reruns `issue new`.
- The transfer guard reads only tracker handoff records (fresh clones are
  protected), exempts the issue's own branch, and is not applicable without a
  tracker. Remote-less publish-gate unit fixtures stub it through a seam.
- The composed issue reader moves wholly to M3 with its consumers.
