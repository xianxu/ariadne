# Issue cards on a tracker branch Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development for bounded tasks or superpowers-executing-plans for session-dependent integration. Steps use checkbox syntax for tracking.

**Goal:** Make cards authoritative on `issue-tracker`, keep details with their authoring/code branch, and permit claim only after initial details land on main, without copying branch commits or stranding issue bookkeeping on resting branches.

**Architecture:** A shared card model composes authoritative metadata with branch-local details. A tracker repository reuses Git snapshot/CAS transactions; explicit creation, transfer, close and landing transitions preserve partial progress. Existing readers and alternate writers move together at a frozen, one-pass cutover.

**Tech Stack:** Go, Cobra, Git object/ref plumbing, existing Git/GitHub process seams and stateful fakes, CUE vocabulary, Markdown artifacts.

**Status:** Proposed for operator approval. Spec: `workshop/issues/000252-issue-cards-tracker-ref.md`. No implementation or production migration has begun. No estimate until the plan-quality gate accepts this plan.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|---|---|---|
| Card / owned-field schema | `cmd/sdlc/internal/issue/card.go`, `construct/vocabulary/issue.cue` | new / modified |
| DetailMirror / composed issue | `cmd/sdlc/internal/issue/mirror.go` | new |
| Creation / transfer transitions | `cmd/sdlc/internal/tracker/creation.go`, `transfer.go` | new |
| Close generation / landing transitions | `cmd/sdlc/internal/tracker/completion.go` | new |
| Migration manifest | `cmd/sdlc/internal/tracker/migration.go` | new |
| Activity event selection | `cmd/sdlc/internal/activetime/commit.go` | modified |

Card owns ID, status, started, created/updated dates, estimate/actual hours, GitHub linkage and canonical title. Its Problem is the original report. Details retain editable Problem, Spec, Done when, Estimate explanation, Plan, Log, Revisions, deps, target, flow and review anchors. Define ownership once; unknown detail fields remain untouched, unknown tracker schema versions refuse. No new lifecycle statuses. Read `vocabulary` skill and the complete CUE model before changing the model.

A DetailMirror is a projection of one card blob, not an independent metadata source. Store the exact last mirrored card blob OID in details. Compare locally mirrored fields against that immutable blob first: changed local owned fields refuse, untouched old projections refresh from latest card. Refuse missing/unavailable baseline instead of guessing. Preserve absent-versus-empty field semantics and YAML scalar types. Problem is excluded from mirror enforcement. Include title in the canonical projection and add a supported title setter; details may independently revise Problem.

Each card has one stable ID/path; one or more detail checkouts can mirror it, but claim serializes work on an eligible issue. Close generation binds the ID to reviewed HEAD, close-evidence commit, and repository identity; status alone cannot establish ownership of a PR. Pure transitions produce declared effects; callers cannot directly set transaction stages. These entities centralize ownership and remove repeated local-file status interpretation (ARCH-DRY, ARCH-PURE).

### Integration points

| Name | Lives in | Status | Wraps |
|---|---|---|---|
| Tracker repository | `cmd/sdlc/internal/tracker/repository.go` | new | `gitx.TrunkFile`, pinned Git refs/trees |
| Git snapshot/CAS and bootstrap | `cmd/sdlc/internal/gitx/trunkfile.go`, `updatemany.go`, new `refbootstrap.go` | modified / new | Git subprocess boundary |
| Detail transfer adapter | `cmd/sdlc/issuemovedetail.go` | new | tracker, main publication, index/worktree and recovery refs |
| Composed issue reader | `cmd/sdlc/issuerecord.go`, `cmd/sdlc/internal/tracker/reader.go` | new | cards plus selected detail location |
| Completion adapter | `cmd/sdlc/trackercompletion.go` | new | close evidence, existing GitHub landing identity and archive transaction |
| Migration command | `cmd/sdlc/issuemigrate.go` | new | repository inventory, bootstrap, mirrors, cutover marker |
| Publication fake | `cmd/sdlc/internal/gitx/commitpublication_fake_test.go` | modified | immutable objects, multiple refs, rejection/lost acknowledgement |

Reuse the common-dir repository lock and existing cancellation-aware Git runner. Linked worktrees share tracker refs and lock; independent clones synchronize by remote CAS. Use a remote-tracking ref and Git object reads rather than a checked-out hidden worktree: there is no mutable tracker checkout to clean, collide with or discover. `sdlc resolve` can materialize a read-only card snapshot for navigation; its path is not an editing authority. This chooses the simpler of hidden-worktree vs ref-backed access without changing the agreed normal `issue-tracker` branch.

Use the configured publication remote/main identity, never an invented `origin`. Fetch `refs/heads/issue-tracker` explicitly in ordinary and CI clones. Reuse `UpdateMany` and `TrunkView`; do not use `PublishCommit` for cards. Bootstrap an orphan tracker history with a format manifest and imported cards, using expected-absence CAS. Normal writes operate only on an initialized tracker.

## Transaction contracts

### Creation and branch residency

`issue new` prepares local content, reserves the card on fresh tracker, and materializes details for the confirmed ID. A failed local write after reservation leaves a recoverable incomplete creation, never a second allocation on retry. Persist operation identity plus candidate/card snapshot before effects; identical content is not proof of ownership.

On a feature branch, details are ordinary local files and may be committed with ongoing work. On a resting checkout, creation leaves local details uncommitted until `move-detail` publishes them or the operator prepares an authoring branch. SDLC must not create resting-branch checkpoint commits. `start-plan`, after a valid claim, prepares an issue branch before authoring/committing design when invoked on rest; `change-code` reuses that branch. A different active issue branch refuses planning for this issue with a next action to use another checkout. This resolves the old sync-before-branch divergence without silently switching away from unrelated work.

Claim reads a fresh tracker card and a pinned fresh main tree, verifies the matching ordinary details file exists, then conditionally updates the card from open to working. Revalidate relevant readiness before publication; card-only/local-only details refuse. Absence of deps for an uncreated issue is unknown, not an empty list.

### `move-detail`: ordinary Git history, no rewritten commits

For an initial file absent at the source branch/main merge bases, publish the current details as a new main-native commit D; do not merge D into the source branch. Remove the source file with a narrow branch commit R after confirmed publication. Relative to the absent baseline, add then remove is net-zero: merging R into main retains D and any later edits E. A real-Git experiment during planning confirmed both properties. Joining D into the source before deleting instead causes deletion or modify/delete conflict and is prohibited.

Preconditions: no Git operation in progress; exact repository/remote/main/source HEAD pinned; destination absent; ordinary source file or proven absence; source absent at every merge base; no staged/unstaged split for the source. Preserve unrelated index/worktree entries. Read/hash source and index before publication and revalidate before removal; changed source remains untouched with recovery instructions. If source has no tracked version, remove only its verified bytes and do not create a deletion commit. On rest, require HEAD to fast-forward to publication (no local-only commits); reconcile the verified source while fast-forwarding, with no separate deletion commit. It then appears as the newly published destination copy.

Record an operation receipt before publication, pinning source bytes and identity in a recovery ref. A retry may accept an existing destination only with its exact operation provenance and confirmed ancestry; independent existing details refuse. No-source mode derives initial details from the card. A network timeout is an unknown outcome, not failure; do not remove the source while unknown. A new card/blob version after preparation requires revalidation, not overwriting with stale metadata.

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
- Tracker format manifest persists per repository. Mirror baseline blobs stay reachable through tracker history. Recovery refs retain snapshots only until confirmed finalization, except transfer merge-protection receipts which live until source branch landing/abandonment; inspect/report abandoned receipts through a maintenance command with explicit branch identity checks, never age-delete uncertain operations. No background daemon.
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

### Task 1: Card, mirror and vocabulary contract

Files: create `cmd/sdlc/internal/issue/card.go`, `card_test.go`, `mirror.go`, `mirror_test.go`; modify `construct/vocabulary/issue.cue`, `pkg/vocab/vocab.go`, generated `pkg/vocab/issue.json`, `pkg/vocab/vocab_test.go`.

- [ ] Add table-driven tests for split/compose preserving all detail prose, deps/target/flow, edited Problem; absent/empty metadata, title edit, stale untouched projection, hand edit, unknown/missing baseline, duplicate keys and malformed schema.
- [ ] Run `go test ./cmd/sdlc/internal/issue ./pkg/vocab -count=1` and record intended failures before implementation.
- [ ] Add card discovery/ownership alongside existing `discovery.home` (which remains details); implement pure projection comparison and refresh. Expose typed validation errors with setter next actions.
- [ ] Regenerate using the repository vocabulary generation path; run the tests above and `git diff --check`; commit explicit paths as `#252 M1: model: split card authority from details`.

### Task 2: Ref-backed repository and transaction evidence

Files: create `cmd/sdlc/internal/tracker/{repository,reader,creation,transfer,completion}.go` and colocated tests; create `cmd/sdlc/internal/gitx/refbootstrap.go` and tests; modify `trunkfile.go`, `updatemany.go`, `commitpublication_fake_test.go` only where existing APIs lack outcome/evidence access.

- [ ] Write fake-sequence and bare-Git tests for bootstrap winner/loser, two cards concurrently, two same-ID/different-slug filings, competing claims, server ref-lock and client stale-lease rejection, lost ack, malformed refs and cancellation.
- [ ] Run `go test ./cmd/sdlc/internal/gitx ./cmd/sdlc/internal/tracker -count=1`; verify new tests fail for the intended missing behavior.
- [ ] Implement snapshot reads, atomic expected-version card writes and typed operation receipts. Preserve existing CAS semantics; generalize the stateful fake to main and tracker refs, not duplicate it per command.
- [ ] Repeat tests, add cardinality/IO benchmarks, and document tracker model in new `atlas/workflow/issue-tracker.md` with `atlas/index.md` link. Commit, then `sdlc milestone-close --issue 252 --milestone M1` with actual test evidence per help. Foundation remains unactivated pending cutover.

## Chunk 2: Creation and handoff — M2

### Task 3: Commands, setters and early design branch

Files: modify `cmd/sdlc/{issue,issueids,claim,claimdecision,setstatus,startplan,changecode,changecode_flow}.go`; add `issuerecord.go`, `issuemetadata.go` and tests; update existing `issue_test.go`, `claimremote_test.go`, `startplan_test.go`, `changecode_test.go`.

- [ ] Write command-level tests for new creating both records, recovery after card publication/local write failure, card-only claim refusal, published-details claim success and two-thread claim race.
- [ ] Add tests that start-plan prepares a branch from rest, reuses the correct branch and refuses unrelated active work; no issue bookkeeping commit moves a resting ref.
- [ ] Run focused tests with `go test ./cmd/sdlc -run 'Test(Issue|Claim|StartPlan|ChangeCode)' -count=1`; record red results.
- [ ] Route through the composed reader and tracker. Add `sdlc issue set --issue N --field FIELD --value VALUE` for editable card metadata (title, github_issue, estimate_hours); ID/dates/started/status/actuals remain owned by their specific lifecycle verbs. Preserve Estimate explanation validation in details.
- [ ] Remove sync/publish behavior from implementation entry; keep ordinary explicit-path local checkpoints on the issue branch. Run focused tests and commit.

### Task 4: Initial-details transfer and merge protection

Files: create `cmd/sdlc/issuemovedetail.go`, `issuemovedetail_test.go`, `transferguard.go`, `transferguard_test.go`; modify `issue.go`, `pr.go`, `publishgate.go`, `merge.go`, `push.go`.

- [ ] Build real-Git fixtures for A→B→R versus A→D→E, asserting main gets only details, code stays unshipped, source removal is scoped, and eventual merge preserves E exactly. Include main-rest fast-forward and no-local-file modes.
- [ ] Add controlled failures after receipt, after remote publish/before acknowledgement, before removal, after removal/before receipt finalization; retry must not overwrite an independent destination or edited source.
- [ ] Add later-main-merge, source-rebase, archival-by-new-owner and squash-landing cases. Refuse unsafe ancestry and verify transferred-path guard catches deliberate deletion/overwrite mutations.
- [ ] Implement the pinned transfer and merge checks described above. Run `go test ./cmd/sdlc -run 'Test(MoveDetail|Transfer)' -count=1`, then affected PR/merge/push suites; no production remotes in tests.
- [ ] Update tracker atlas and README command discovery; commit and close M2 through the SDLC gate.

## Chunk 3: All readers, close and landing — M3

### Task 5: Composed readers and activity evidence

Files: modify every reader in the inventory, with existing colocated tests; add `cmd/sdlc/issuerecord_test.go`, `internal/tracker/reader_test.go`.

- [ ] Write fixtures with intentionally stale detail status/hours and current card values; assert list/state/fleet/projects/GH use cards while deps/target/plan use selected details. Missing unpublished details are visible as incomplete, not silently omitted.
- [ ] For cross-issue/dependency reads use pinned main details; for the active issue use its checked-out details. Preserve archive navigation and allow card-only inspection without fabricating an editable file.
- [ ] Extend actual-time tests for tracker-only claim/design intervals, code+tracker events, concurrent issues, reopened episodes and duplicate reachable OIDs. Select only HEAD plus the pinned tracker history, not `--all`; retain the existing attribution algorithm and explicit telemetry-gap result.
- [ ] Implement and run `go test ./cmd/sdlc ./cmd/sdlc/internal/fleet ./cmd/sdlc/internal/project ./cmd/sdlc/internal/activetime -count=1`; commit after the new assertions pass.

### Task 6: Close generation, exact landing and recovery

Files: create `cmd/sdlc/trackercompletion.go`, `trackercompletion_test.go`; modify `close.go`, `reviewstate.go`, `milestoneclose.go`, `publishgate.go`, `landing.go`, `landingarchive.go`, `merge.go`, `push.go` and their tests.

- [ ] Add sequences covering stale card during review, close evidence committed/card publish failed, two independent completions, PR merged/done ack lost, done/archive failed, retry after source checkout deletion and reopen before stale recovery.
- [ ] Prove archive selection no longer depends on a branch-local codecomplete edit and never archives unrelated issues merely mentioned in a PR. Test direct-main push as well as merge/squash PR landing.
- [ ] Implement explicit completion bindings and recovery through existing landing identity/review guards; retain exact reviewed-head checks and scoped archive behavior.
- [ ] Run `go test ./cmd/sdlc -run 'Test(Close|Milestone|Review|Publish|Landing|Merge|Push|Archive)' -count=1`; update `atlas/workflow/{issue-lifecycle,ledger-landscape,sdlc-binary}.md`, commit and close M3.

## Chunk 4: Cutover tooling, instructions and end-to-end proof — M4

### Task 7: One-pass migration and obsolete surface removal

Files: create `cmd/sdlc/issuemigrate.go`, `issuemigrate_test.go`, `internal/tracker/migration.go` and tests; modify `issue.go`, `repoguard.go`, `validategate.go`, `issuelintids.go`, portable scripts/Makefile/instructions from the inventory; delete retired `issuesync`/`issuepublish` command code only after checking remaining callers.

- [ ] Add `sdlc issue migrate --dry-run` and `--apply` with an exact inventory manifest: active/archive IDs, duplicates, source OIDs, all worktrees, active branches, legacy local-only issue commits, schema version and proposed tracker root. Default dry-run never changes refs/files.
- [ ] Write migration fixtures for old flat/nested archives, terminal active files, dirty/unsynced branches, duplicate IDs, competing bootstrap, crash at every phase and incompatible format. Do not read live workshop/history during development; fixtures provide coverage.
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

The user approved the issue's behavior; this document supplies the still-new Git transfer, mirror provenance, completion binding and migration mechanics. Review this durable plan before `change-code`, per AGENTS.md §2. Each M1–M4 row is a real mandatory SDLC review boundary; do not add a second ad-hoc code reviewer at those boundaries. A fresh plan-document review precedes operator approval. After approval, run `sdlc change-code --issue 252 --worktree=yes`; address its plan-quality/estimate next actions without bypassing them. Use an isolated checkout so unactivated base-layer changes do not become live in primary downstream consumers.
