# Issue cards on a tracker branch Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development for bounded tasks or superpowers-executing-plans for session-dependent integration. Steps use checkbox syntax for tracking.

**Goal:** Make cards authoritative on `issue-tracker`, keep details with their authoring/code branch, and permit claim only after initial details land on main, without copying branch commits or stranding issue bookkeeping on resting branches.

**Architecture:** A shared card model composes authoritative metadata with branch-local details. A tracker repository reuses Git snapshot/CAS transactions; explicit creation, transfer, close and landing transitions preserve partial progress. Existing readers and alternate writers move together at a frozen, one-pass cutover.

**Tech Stack:** Go, Cobra, Git object/ref plumbing, existing Git/GitHub process seams and stateful fakes, CUE vocabulary, Markdown artifacts.

**Status:** Operator approved implementation on 2026-09-25. M1 accepted through its boundary gate (FIX-THEN-SHIP; minor artifact whitespace corrected). M2 accepted through its boundary gate (SHIP after rounds 3–6). M3 accepted through its boundary gate (SHIP after rounds 7–9). M4 accepted through its boundary gate (SHIP after rounds 10–11). Production migration has not begun. No production migration or consumer activation has begun.

## Core concepts

### Pure entities

This inventory distinguishes delivered M1 foundation from planned later work.
“Planned” is not a claim of presence or modification in the M1 review window.

| Name | Kind | Lives in | Delivery status |
|---|---|---|---|
| Card / owned-field schema | PURE | `cmd/sdlc/internal/issue/card.go`, `construct/vocabulary/issue.cue` | M1: new / modified, delivered |
| DetailMirror | PURE | `cmd/sdlc/internal/issue/mirror.go` | M1: new, delivered |
| Creation / transfer transitions | PURE | `cmd/sdlc/internal/tracker/creation.go`, `transfer.go` | M1: new, delivered |
| Close generation transition (evidence, codecomplete) | PURE | `cmd/sdlc/internal/tracker/completion.go`, `completeop.go` (`CodecompleteCard`) | M1: new; M3: trimmed to two stages, delivered |
| Typed recovery receipt / shared transition engine | PURE | `cmd/sdlc/internal/tracker/receipt.go` | M1: new, delivered |
| Handoff envelope / card-derived details | PURE | `cmd/sdlc/internal/issue/handoff.go` | M2: new, delivered |
| Card field/title mutators | PURE | `cmd/sdlc/internal/issue/cardset.go` | M2: new, delivered |
| Migration plan / manifest | PURE | `cmd/sdlc/internal/tracker/migration.go` (`PlanTrackerMigration`) | M4: new, delivered |
| Migration card derivation | PURE | `cmd/sdlc/internal/issue/migrate.go` (`MigrateActiveDetails`, `ArchivedCard`, `ReconcileLegacyDetails`) | M4: new, delivered |
| Composed issue records join | PURE | `cmd/sdlc/internal/tracker/records.go` (`composeRecords`, `IssueRecord.Field`) | M3: new, delivered |

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
| Receipt driver / candidate steps / recovery refs | INTEGRATION | `internal/tracker/drive.go`, `createop.go`, `store.go`, `candidates.go`, `internal/gitx/candidate.go`, `recoveryref.go`, `objects.go` | M2: new, delivered | receipt engine, Git candidate publication, recovery refs in the common Git dir |
| Publication target | INTEGRATION | `internal/gitx/publicationtarget.go` | M2: new, delivered | resting branch upstream config |
| Planning branch / tracker environment | INTEGRATION | `cmd/sdlc/planningbranch.go`, `trackerenv.go` (checkout/target glue) | M2: new, delivered | workspace identity, branch creation |
| Card setters / recovery verbs | INTEGRATION | `cmd/sdlc/cardsetters.go`, `issuerecovery.go` | M2: new, delivered | tracker CAS updates, receipt resume |
| Transfer guard | INTEGRATION | `cmd/sdlc/transferguard.go` | M2: new, delivered | `git merge-tree`, tracker handoff records |
| Composed issue reader | INTEGRATION | `cmd/sdlc/issuerecord.go` | M3: new, delivered | M1 card reader plus selected detail location |
| Records loader / repository opener | INTEGRATION | `cmd/sdlc/internal/tracker/records.go` (`LoadRecords`, `RepositoryForCheckout`), `cmd/sdlc/issuerecord.go` (per-command scope) | M3: new, delivered | tracker snapshot, details files, workspace identity |
| Activity commit loader | INTEGRATION | `cmd/sdlc/internal/activetime/commit.go` (`loadWindowCommits`), `internal/gitx/window.go` (`CommitWindow`) | M3: modified (tracker ref beside HEAD), delivered | git log over HEAD and the tracker ref |
| Close adapter | INTEGRATION | `cmd/sdlc/closetracker.go`, `internal/tracker/completeop.go` | M3: new, delivered | evidence commit (temporary index, branch CAS), card codecomplete |
| Completion adapter | INTEGRATION | `cmd/sdlc/trackercompletion.go` | M3: new, delivered | binding selection, done CAS, existing landing identity and archive transaction |
| Migration command | INTEGRATION | `cmd/sdlc/issuemigrate.go` (`migrationInventory`, `applyTrackerBootstrap`, `applyMainCutover`, `runMigrateReconcile`) | M4: new, delivered | repository inventory, bootstrap, mirrors, cutover marker |
| Cutover marker and guard | INTEGRATION | `cmd/sdlc/internal/tracker/cutover.go` (`Repository.GuardCutover`, `RefuseLegacyDetails`, `CutOver`; `ParseCutoverMarker` pure) | M4: new, delivered | checkout marker file, `gitx.TrunkFile.HasRoot` |
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
- Tracker format manifest persists per repository. Mirror baseline blobs stay reachable through tracker history. Each transferred card retains one initial-handoff record, bounded to one per initial creation; its immutable history supports fresh-clone verification even after source-clone loss. Recovery refs live in the repository's common Git directory, so every linked worktree sees every receipt; a receipt's remaining local effects (materializing details, removing a handed-off source) run only in the worktree on its source branch, and `recovery list` marks receipts owned elsewhere. They retain source snapshots until confirmed finalization; `sdlc issue recovery list` reports remaining local operations and `sdlc issue recovery reconcile --issue N` resumes or cleans confirmed completed operations through the same transition model. Never age-delete uncertain operations. No background daemon.
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
| `TrunkFile.Bootstrap` (planned `BootstrapTracker`), `Repository.UpdateCard` | Reproducibly interleave two writers and uncertain transport results against stateful Git; one authoritative winner, unique IDs, no lost updates. |
| `TrunkFile.run`, `runIssueNew` cancellation path | Cancel from the production Cobra command while a controlled child Git process blocks; process group and pipes end within five seconds and uncertain publication retains recovery evidence. |
| `StepCreation`, `StepTransfer`, `StepCompletion` | Generate event sequences, including interruption at each declared effect; assert no destructive action without confirmed ownership and no accepted stale generation. |
| `runClaim` | Drive concurrent commands against independent snapshots; only a fresh-main-ready open card can be reserved once. |
| `preparePlanningBranch` (planned `PreparePlanningBranch`) | Generate rest/main ancestry and dirty-state combinations; accepted branch contains eligible details and never changes resting refs. |
| `checkTransferredPaths` (planned `CheckTransferredPaths`) | Generate DAGs across publication and later owner changes; reject any prospective merge losing authoritative transferred content. |
| `LoadRecords` (planned `ReadIssueRecord`), `LookupRepoIssues` | Vary card and detail generations independently; source selection respects field ownership and missing detail information remains unknown. |
| `loadWindowCommits` (planned `SelectActivityEvents`), `computeActual` | Generate overlapping selected histories; each OID counted once and claim engagement survives off-branch storage. |
| `publishTrackerClose` (planned `FinalizeTrackerClose`), `ownedCompletions` (planned `SelectCompletedIssues`) | Mutate bound repo/head/evidence between review and effect; refuse stale ownership rather than completing unrelated work. |
| `PlanTrackerMigration`, `applyTrackerBootstrap`/`applyMainCutover` (planned `ApplyTrackerMigration`) | Generate legacy populations and interrupt each phase; no lost IDs/evidence and no activation from incomplete or contradictory inputs. |

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

- [x] Write failing model-based tests for `TrunkFile.Bootstrap` (planned `BootstrapTracker`), `Repository.UpdateCard`, `StepCreation`, `StepTransfer`, `StepCompletion`; run matching sequences against the stateful fake and disposable real Git.
- [x] Run `go test ./cmd/sdlc/internal/gitx ./cmd/sdlc/internal/tracker -count=1`; verify new tests fail for the intended missing behavior.
- [x] Implement snapshot reads, atomic expected-version card writes and typed operation receipts. Preserve existing CAS semantics; generalize the stateful fake to main and tracker refs, not duplicate it per command.
- [x] Repeat tests, add cardinality/IO benchmarks, and document tracker model in new `atlas/workflow/issue-tracker.md` with `atlas/index.md` link. Committed as `2144608`; verification evidence is recorded in the issue's M1 review-submission Log (including the baseline CLI-suite failure and measured performance limits).

**M1 accepted:** `sdlc milestone-close --issue 252 --milestone M1` round 2
disposed BR-1/BR-2 and returned FIX-THEN-SHIP with one minor artifact-whitespace
finding (BR-3), corrected in the boundary commit. The issue's M1 checkbox is
the acceptance record. Foundation remains unactivated pending cutover.

## Chunk 2: Creation and handoff — M2

### Task 3: Commands, setters and early design branch

Files (revised at M2): modify `cmd/sdlc/{issue,claim,setstatus,startplan,changecode}.go`; add `trackerenv.go`, `planningbranch.go`, `cardsetters.go`, `internal/issue/cardset.go` and tests (`issuenew_test.go`, `planningbranch_test.go`, `changecode_tracker_test.go`); update `claimremote_test.go`, `startplan_test.go`, `setstatus_test.go`. `issuerecord.go` moves to M3; `issuemetadata.go` was superseded by `cardsetters.go`.

- [x] Test `runIssueNew` and `runClaim` through the production command seam with controlled publication schedules; assert one reserved identity/claim and no readiness from local-only details.
- [x] Test `preparePlanningBranch` (planned `PreparePlanningBranch`) with generated checkout/ref relationships; assert fresh-main details are present and resting refs/unrelated work are unchanged.
- [x] Run focused tests with `go test ./cmd/sdlc -run 'Test(Issue|Claim|StartPlan|ChangeCode)' -count=1`; record red results.
- [x] Route through the tracker (composed reader deferred to M3 with its consumers). Card setters are `sdlc issue set-title` / `set-estimate` / `set-github` (the delivered CUE model's setter names; see Revisions); ID/dates/started/status/actuals remain owned by their specific lifecycle verbs. Estimate explanation validation stays in details.
- [x] Remove sync/publish behavior from implementation entry; keep ordinary explicit-path local checkpoints on the issue branch. Run focused tests and commit.

### Task 4: Initial-details transfer and merge protection

Files: create `cmd/sdlc/issuemovedetail.go`, `issuemovedetail_test.go`, `transferguard.go`, `transferguard_test.go`; modify `issue.go`, `pr.go`, `publishgate.go`, `merge.go`, `push.go`.

- [x] Build real-Git fixtures for A→B→R versus A→D→E, asserting main gets only details, code stays unshipped, source removal is scoped, and eventual merge preserves E exactly. Include main-rest fast-forward and no-local-file modes.
- [x] Test `StepTransfer` with reproducible interruption/event sequences; source removal requires confirmed owned publication and an unchanged source fingerprint.
- [x] Property-test `checkTransferredPaths` (planned `CheckTransferredPaths`) over generated branch/main DAGs and fresh-clone replay; prospective merge preserves current destination state or refuses without mutation.
- [x] Implement the pinned transfer and merge checks described above, including `issue recovery list/reconcile` registration and tests in `issuemovedetail_test.go`. Run `go test ./cmd/sdlc -run 'Test(MoveDetail|Transfer|IssueRecovery)' -count=1`, then affected PR/merge/push suites; no production remotes in tests.
- [x] Update tracker atlas and README command discovery; commit.
- M2 acceptance is its `sdlc milestone-close` boundary review, recorded by the issue's M2 checkbox (not ticked here).

## Chunk 3: All readers, close and landing — M3

### Task 5: Composed readers and activity evidence

Files (revised at M3 from `git diff --name-status`): create `internal/tracker/records.go`, `cmd/sdlc/issuerecord.go` and tests; modify `state.go`, `issue.go`, `startplan.go`, `repoguard.go`, `pr.go`, `internal/fleet/issues.go`, `projectstatus.go`, `projectclose.go`, `projectretro.go`, `project.go`, `projectsetstatus.go`, `projectforecast.go`, `actual.go`, `internal/activetime/{commit,compute}.go`, `internal/gitx/window.go`, `main.go` (records scope). `resolve.go`, `validategate.go` and `internal/project/metadata.go` needed no change (navigation / open schema / generic decode).

- [x] Test `LoadRecords` (planned `ReadIssueRecord`) and `LookupRepoIssues` with independently varied authoritative snapshots and projections; card metadata wins while detail-owned values remain intact.
- [x] For cross-issue/dependency reads use pinned main details; for the active issue use its checked-out details. Preserve archive navigation and allow card-only inspection without fabricating an editable file.
- [x] Test `loadWindowCommits` (planned `SelectActivityEvents`) and `computeActual` over generated code/tracker histories; assert selected-source-only membership, deduplicated OIDs and preserved engagement intervals.
- [x] Implement and run `go test ./cmd/sdlc ./cmd/sdlc/internal/fleet ./cmd/sdlc/internal/project ./cmd/sdlc/internal/activetime -count=1`; commit after the new assertions pass.

### Task 6: Close generation, exact landing and recovery

Files (revised at M3 from `git diff --name-status`): create `cmd/sdlc/{trackercompletion,closetracker}.go`, `internal/tracker/completeop.go` and tests; modify `close.go`, `milestoneclose.go` (shared trailers), `publishgate.go`, `landing.go`, `landinggate.go`, `landingarchive.go`, `merge.go`, `push.go`, `issuefiles.go`, `issuerecovery.go`, `internal/tracker/{receipt,completion,store,candidates,drive}.go`, `internal/issue/handoff.go`. `reviewstate.go` needed no change: review snapshots deliberately do not pin the card (see M3 Revisions).

- [x] Test `StepCompletion` and `publishTrackerClose` (planned `FinalizeTrackerClose`) with reproducible review/remote-generation interruptions; stale verdicts and uncertain writes never become accepted completion.
- [x] Test `ownedCompletions` (planned `SelectCompletedIssues`) against unrelated ancestry and altered generation bindings; only proven reviewed/landed generations can be selected, independent of mirrored status.
- [x] Implement explicit completion bindings and recovery through existing landing identity/review guards; retain exact reviewed-head checks and scoped archive behavior.
- [x] Run `go test ./cmd/sdlc -run 'Test(Close|Milestone|Review|Publish|Landing|Merge|Push|Archive)' -count=1`; update `atlas/workflow/{issue-lifecycle,ledger-landscape,sdlc-binary}.md`; commit.
- M3 acceptance is its `sdlc milestone-close` boundary review, recorded by the issue's M3 checkbox.

## Chunk 4: Cutover tooling, instructions and end-to-end proof — M4

### Task 7: One-pass migration and obsolete surface removal

Files: create `cmd/sdlc/issuemigrate.go`, `issuemigrate_test.go`, `internal/tracker/migration.go` and tests; modify `issue.go`, `repoguard.go`, `validategate.go`, `issuelintids.go`, portable scripts/Makefile/instructions from the inventory; delete retired `issuesync`/`issuepublish` command code only after checking remaining callers.

- [x] Add `sdlc issue migrate --dry-run` and `--apply` with an exact inventory manifest: active/archive IDs, duplicates, source OIDs, all worktrees, active branches, legacy local-only issue commits, schema version and proposed tracker root. Default dry-run never changes refs/files.
- [x] Test `PlanTrackerMigration` and `applyTrackerBootstrap`/`applyMainCutover` (planned `ApplyTrackerMigration`) against generated legacy populations and interrupted phases; preserve every used ID and valid completion, refusing ambiguous activation. Use synthetic history fixtures, not live workshop/history.
- [x] Import existing codecomplete generations only after reconstructing their binding from the legacy close-transition commit, accepted review evidence and pinned PR head (or direct-main landing target), using existing publish-gate validation. Migration refuses an unprovable binding before activation, with the exact issue/anchor and next action to reopen/reclose under the old workflow during the freeze. Never silently omit the issue, invent acceptance, or change its status. Prove the valid imported PR can next merge to done/archive; prove the invalid case blocks cutover until repaired. Include these checks in the operator cutover manifest.
- [x] Define cutover sequence: freeze all writers; inventory and checkpoint every checkout; refuse unresolved legacy divergence; publish tracker snapshot; commit/publish ordinary main mirror conversion with format marker; reconcile each inventoried branch against its captured old card projection; verify and activate new writers. No automatic discard/reset/rebase of user work. A late/unknown checkout must reconcile or refuse before mutation.
- [x] New binary refuses writes when cutover marker and tracker generation disagree. Old binaries cannot be made to honor a new marker: the operational freeze must include aliases, scripts and already-running agents until all entrypoints are upgraded. Retry resumes a matching manifest; after tracker writes resume never roll it back to the import snapshot.
- [x] Remove/delegate shell/Python writers and `issue sync`/generic copied issue publication entrypoints. Update CI fetch/ID checks and all active agent guidance, including adapted skills that prescribe sync. Use xx-construct guidance for substrate edits and preserve generated-source boundaries.
- [x] Run `go test ./cmd/sdlc/... ./pkg/vocab/... -count=1`, vocabulary generation/conformance checks, and existing shell/Python checks covering modified portable entrypoints. Run the consumer sweep and classify remaining matches in the issue Log.

### Task 8: Full slot cycle and rollout package

Files: create `cmd/sdlc/tracker_e2e_test.go`; update `README.md`, `atlas/workflow/{issue-tracker,issue-sync,workspace-branching,artifact-hierarchy,ci-merge-check,base-layer}.md`, `atlas/index.md`; add `atlas/workflow/issue-tracker-migration.md`.

- [x] Exercise two cloned slots and a bare remote: new → details publication → claim → branch/design → change-code → close → PR/merge → archive → resting refresh. Include a spin-off handoff while original code stays unshipped, a competing claim and later edits to the transferred issue.
- [x] Assert main is unchanged by card-only operations, no copied source-commit publication, card states current across slots, original PR has no issue-file conflict, and refreshed rest is exactly zero ahead/behind. Add fresh CI clone without tracker ref and disconnected-read/mutation cases.
- [x] Run full Go suites above plus focused real-Git race/conformance tests; publish measured 10k-card benchmark results and verify process counts remain bounded. Update atlas at this boundary, not after rollout.
- [x] Write exact operator cutover/recovery procedure, including owner-binary build and generated instruction propagation. Inventory participating downstream repos read-only under their local instructions before scheduling any freeze. Product-owned consumers needing changes require issues and review in their owning repos; do not call ariadne tests proof of peer integration.
- [x] Record M4 verification and close its review boundary. Actual fleet migration is a separately coordinated operational step using the tested command; do not freeze another live session or migrate peer state without that coordination. Issue closure requires the migration acceptance criterion to be satisfied, not merely the command to exist.

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

### 2026-09-25 — M2 boundary review round 3 (REWORK) corrections

Reason: the M2 review found a foreign-worktree recovery hazard, octal-parsed
`--issue` hints, an unguarded interrupted-handoff window, stale verb help and
(second in family) inventory drift. Delta:
- Local effects are bound to the receipt's source branch in both adapters, and
  reconcile refuses another checkout up front (`ErrForeignCheckout`).
- Every `--issue` hint renders through `issue.CLIRef`; a source guard test keeps
  padded IDs out of hints.
- The transfer guard also protects handoffs without `main_commit` once main holds
  the details; handoff destinations are validated and a malformed card's refusal
  says it blocks publication repo-wide.
- Help for claim, issue, set-status, start-plan, change-code and root now
  describes the tracker workflow; `issue sync`/`publish` are marked as retiring.
- Inventory rule (applied here): every non-test file added in a boundary window
  (`git diff --name-status BASE HEAD | grep '^A' | grep -v _test`) has a Core
  concepts row or is named as glue, and each Task "Files:" list is revised to
  what exists.
- Also: guard runs under the verb's context; uncertain stops report the last Git
  error; set-status reports malformed details; move-detail validates/refreshes
  the mirror before publishing. `start-plan --issue` now needs the tracker and
  network while pre-tracker details keep change-code's legacy path — acceptable
  under the freeze-and-cutover plan, removed at M4.

### 2026-09-25 — M3 implementation

Reason: executing Chunk 3 showed a checkout-local receipt cannot carry a close
to its landing (another clone, days later, or GitHub). Delta:
- Completion = evidence commit + codecomplete with a `tracker.completion`
  binding on the card; the M1 engine's landing/done/archive stages were removed
  as unreachable. Landing, done and archive re-derive from the binding; recovery
  is "codecomplete whose evidence is on main" (idempotent, token-checked).
- Close commits its own evidence for tracker-era issues (it used to leave the
  commit to the agent). The receipt carries the evidence message and path list
  so FIX-THEN-SHIP (stored unstarted; fixes first) and recovery can rebuild it.
- Tracked details are archived byte-for-byte (card = status authority), keeping
  the durable archive's retry proof deterministic; no pre-archive mirror refresh.
- Readers compose through `tracker.LoadRecords`; the repository derives from the
  issues directory; read-only views label stale reads; missing halves are
  unknown (card-only project deps ⇒ blocked). Active-time reads the tracker ref
  beside HEAD. Planned names `ReadIssueRecord`/`SelectActivityEvents`/
  `FinalizeTrackerClose`/`SelectCompletedIssues` shipped as `LoadRecords`/
  `loadWindowCommits(extraRefs)`/`publishTrackerClose`/`ownedCompletions`.
- Review snapshots do not pin the card blob: unrelated card changes do not
  invalidate a review, and a status change is refused when the card is published.
- Files added in the window (enumerated): `closetracker.go`,
  `internal/tracker/completeop.go`, `internal/tracker/records.go`,
  `issuerecord.go`, `trackercompletion.go` — each has a Core concepts row.

### 2026-09-25 — M3 boundary review round 7 (REWORK) corrections

Reason: the review found squash/rebase landings never completing, a
resumable stale close, late preconditions, ambient contexts, repeated fetches,
a wedging FIX-THEN-SHIP, and (3rd in family) inventory kinds/files drift. Delta:
- Landings complete closes by the PR's confirmed identity (slot: `MergeOID`;
  non-durable merge: selection before the server-side merge); ancestry settling
  remains for pushes. Proven through runMerge for merge/squash/rebase.
- Evidence is pinned as blobs in the receipt; an empty evidence commit is legal.
  One live close per issue (re-close supersedes an unstarted one); every resume
  proves no newer close generation holds the card; tracker preconditions run in
  computeClose before review.
- Verb contexts reach all tracker reads (source guard); one composed view per
  repository per command (records scope, invalidated by card writes).
- Inventory rule, applied here: re-derive each entity's kind from its shipped
  code and test IO (pure `composeRecords` extracted; loaders are INTEGRATION),
  and revise every Task "Files:" list against `git diff --name-status`.


### 2026-09-25 — M3 boundary review round 8 (FIX-THEN-SHIP) corrections

Reason: an abandoned worktree branch completed its cards (BR-29); a deferred
evidence commit reverted files a fix commit had edited after the close (BR-30);
fleet inventory still minted its own context; and (4th in family) ticked rows
named planned identifiers that never shipped. Delta:
- Non-durable merge completes owned cards only on a confirmed landing (a PR
  merged now or found merged on resume); abandoning the branch completes nothing.
- A deferred evidence commit replays a pinned file only where HEAD still holds
  the reviewed commit's version (`EvidenceEntry.Replays`, base derived from
  `ReviewedHEAD`); a later committed edit is newer and survives.
- `CollectInventory`/`LookupRepoIssues` take the caller's context; the context
  source guard parses every `internal/` package too. `guardIssueNotDone` reads
  Fresh (it authorizes a write).
- Ticked rows now name the shipped identifiers, the planned names kept in
  parentheses; rule: every identifier in a ticked row greps to shipped code.

### 2026-09-26 — M4 design: concrete migration and cutover mechanics

Reason: Task 7 named the phases but not the mechanics. A read-only probe of the
fleet (11 repositories, 280 active / 860 archived details) found the strict card
parser rejects 15 active and 217 archived files (no `## Problem` — older issues
keep it as the preamble under the H1; closed without hours; no frontmatter), and
this repository holds duplicate IDs (two archived 000040s; open 000096 beside an
archived 000096). Delta (ARCH-PURE, ARCH-ORDER, ARCH-SECURE, ARCH-DRY):

- **Card derivation (pure, `internal/issue`).** Active details split with the
  existing `SplitCardWithFormat`; one without `## Problem` gets the heading
  inserted above its non-empty preamble (the modern template, content unchanged)
  and splits normally; an empty preamble, any invalid owned field, or a closed
  status without hours refuses with the file and next action. Archived details
  are never rewritten (their frontmatter stays their terminal authority, as
  `historyFileIsTerminal` already reads unmirrored files): each gets a card built
  tolerantly — valid owned fields copied, invalid ones dropped, N/A hours for
  closed work lacking them, status inferred `done` only for a file without
  frontmatter, title from the H1 or slug, Problem from the section, else the
  preamble, else a pointer to the archived file — every inference listed in the
  manifest for operator review.
- **Plan (pure, `internal/tracker/migration.go`, `PlanTrackerMigration`).** Input
  is an inventory: pinned main commit, repository identity, object format, active
  and archived files from main's tree, per-branch issue files that differ from
  their merge base, dirty-worktree issue paths, and codecomplete anchor evidence.
  Output is a deterministic manifest (cards, details conversions, refusals,
  inferences, duplicates, digest). One card per ID: the active file if any, else
  the newest archived file; two active files sharing an ID refuse (renumber first).
  A branch whose issue file changes card-owned fields relative to main, an issue
  present only on a branch, or dirty issue edits in a worktree refuse (publish or
  discard under the old workflow during the freeze). A codecomplete issue imports
  with a completion binding only when exactly one branch (or main) carries a
  legacy codecomplete anchor (`codecompleteAnchorCommitAt`) with only docs after
  it; binding = {token `migrate-<id>`, repository, reviewed = anchor's parent,
  evidence = anchor}; otherwise it refuses as unprovable.
- **Apply (IO, `issuemigrate.go`).** `sdlc issue migrate` is a dry run;
  `--apply --expect <digest>` re-inventories, requires the reviewed digest and no
  refusals, then: (1) bootstraps `issue-tracker` by expected-absence CAS, or
  accepts an existing tracker only when its root commit's files equal the plan's;
  (2) publishes one main commit (details conversions + the cutover marker
  `workshop/issue-tracker.json` = `{version, tracker_root}`) by CAS on the pinned
  main. Each phase is observable, so a retry resumes; main already carrying the
  marker for this tracker reports "already migrated". Never rolled back.
- **Cutover guard.** Repositories opened by commands enforce the marker: a
  checkout whose marker disagrees with the tracker (marker without tracker,
  tracker without marker, or a root not in the tracker's history) refuses with the
  next action (`git pull` on rest, `sdlc issue migrate --reconcile` on a branch).
  A legacy (unmirrored) active details file in a tracked checkout refuses the
  legacy close/change-code paths instead of writing status into it.
- **Branch reconciliation.** `sdlc issue migrate --reconcile` on a pre-cutover
  branch appends the mirror line against the imported card (after proving the
  branch's owned fields equal it) and adds the marker in one ordinary commit, so
  a later merge with main applies identical changes; divergent fields refuse.
- **Legacy writers.** In a tracked repository `issue sync`/`issue publish` and the
  Makefile/Python fallbacks refuse (marker-detected) with the tracker-era action;
  their code stays until every fleet repository has cut over, and is deleted at
  the coordinated migration step before issue close (removing it earlier would
  strand unmigrated repositories sharing the binary).

### 2026-09-26 — M4 implementation corrections

Reason: implementation and the rehearsals refined the design. Delta:
- `issue sync` is not retired: in a tracker repository it stays the checkpoint
  verb as a local commit on the issue branch, refusing its legacy behaviors
  (a resting-branch commit, `--push`). `issue publish` and the Makefile/Python
  fallbacks refuse on the marker. Code deletion waits for the fleet cutover.
- A legacy close on a branch wins over main's issue-sync copy of it (which
  also records codecomplete); main's anchor binds only a close made on main.
- Branch refusals group by (path, reason) across stacked branches and their
  remote-tracking copies (a real stack produced 53 lines for 10 problems).
- The e2e found two defects outside the migration, fixed with their class:
  close compared symlink-resolved and unresolved paths (`canonRoot` now
  resolves through a missing path's nearest ancestor), and `issue list`
  printed stale statuses bare — every PreferFresh view now labels a stale
  read, `pr` reads Fresh, and a source guard enforces it.
- Files, per `git diff --name-status` of the window. Task 7: new
  `issuemigrate.go`, `internal/tracker/{migration,cutover}.go`,
  `internal/issue/migrate.go` (+ tests, `trackedlegacy_test.go`,
  `stale_guard_test.go`); modified `issue.go`, `issuepublish.go`, `close.go`,
  `changecode.go`, `claim.go`, `issuemovedetail.go`, `issuerecord.go`,
  `trackerenv.go`, `internal/tracker/{repository,candidates,records}.go`,
  `internal/gitx/candidate.go`, readers (`state.go`, `startplan.go`, `pr.go`,
  `actual.go`, `projectstatus.go`, `internal/fleet/{issues,render}.go`),
  `closetracker.go`, `propagatebase.go`, `Makefile.workflow`,
  `scripts/close-issue.py`, `AGENTS.base.md`, `helptext/{issue,estimate}.md`,
  `construct/datatype/project.md`. `repoguard.go`, `validategate.go` and
  `issuelintids.go` needed no change. Task 8: new `tracker_e2e_test.go`,
  `atlas/workflow/issue-tracker-migration.md`; modified `README.md`,
  `atlas/index.md`, `atlas/workflow/{issue-tracker,issue-sync,
  workspace-branching,artifact-hierarchy,ci-merge-check,base-layer}.md`.
- Measured (Apple Silicon, 3 CPUs visible, a full suite running alongside):
  Git snapshot of 10,000 cards 775 ms with 2 Git processes; parse 75 ms;
  compose with 100 details 4 ms; refresh 100 mirrors 14 ms — about 0.85 s for
  the whole local read, inside the one-second envelope. Planning a
  10,000-issue migration: 294 ms.
- Read-only fleet inventory (dry runs on disposable copies): kaggle, kbench,
  metis, nous, xianxu.dev and you-decide are ready; ariadne, pair and tools
  hold stale branches editing archived issues and a few unsynced closes;
  parley.nvim holds an unlanded stack (#276–#285) whose issues exist only on
  branches. Each is a pre-cutover task for its repository, under its own
  instructions.

### 2026-09-26 — M4 boundary review round 10 (REWORK) corrections

Reason: the review found (Critical, 4th in family deferred-effect-input-drift)
that a legacy codecomplete's main conversion pinned the card from before its
close was bound, so main named a blob the tracker never held and the reconciled
branch conflicted on the mirror line; plus five Minors. Delta:
- One constructor pins a mirror, from final bytes: `MigrateActiveDetails`
  returns unmirrored details; `issue.MirrorDetails` proves the owned fields and
  pins the card after binding; reconcile uses the same constructor. The
  generated-population test asserts every conversion's baseline equals its
  final card's blob and that an unchanged branch copy reconciles to main's
  bytes; the imported-close e2e now merges migrated main without conflict,
  lands, and settles the card to done.
- A legacy close already in main's history (landed, never marked done) binds
  main's own close record; code after it on main is other work.
- `lint-ids` refuses details a range adds without a card at the same id and
  slug in a tracker repository (the consumer inventory's "tracker ID
  validation" row, now delivered, with a real-Git exit-code test).
- The Makefile/Python fallbacks detect a cut-over repository as
  `tracker.CutOver` does (marker, or a fetched tracker).
- `TrunkFile.Roots` is the one root listing; the guard caches its verified
  (root, ref) pair. Measured: 54 ms and one Git process per new tracker
  generation over a 10,000-write history; the local read stays under a second.
- The moved-main apply error names both outcomes (resume, or abandon an
  unwritten tracker).

