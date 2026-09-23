# Concurrent issue workflows implementation plan

> **For agentic workers:** Consult AGENTS.md §3 for delegation. Use TDD; the SDLC binary owns milestone reviews.

**Goal:** Safely reserve work, publish explicitly chosen workflow changes, and run reviews concurrently across all slots and independent clones.

**Architecture:** `origin/main` is the issue-status authority. Claim conditionally changes an open issue to working. The agent chooses a coherent documentation commit; SDLC applies its patch to fresh remote main and conditionally publishes it. External reviews run outside the local transaction lock and revalidate before saving results.

**Tech Stack:** Go, Cobra, Git commit/merge plumbing, existing SDLC locks and judge adapters.

**State:** Scope approved in the operator discussion on 2026-09-23. This revision supersedes v1. Implementation is paused; prior plan/estimate gates do not establish clearance for this changed design. Re-run the design gate and rederive the estimate before resuming code work.

## Agreed contract

### 1. Claims use status on origin/main

`working` means the issue is taken, whether work happens in another slot or on another machine. No local-tree scan can replace that authority. On every claim attempt:

1. Fetch remote main and read the issue at that exact commit.
2. Require open status using the existing vocabulary; missing, malformed or non-open records refuse.
3. Construct only the status/start-date update on those remote bytes, preserving the rest of the record.
4. Publish only if the remote ref still equals the observed commit. On contention, fetch and check status again; retry at most three times.

Two callers observing open cannot both win: the second conditional write fails, and the reread sees working. Repeating a claim on a working issue refuses even for the original worker. Continuing already-started work does not require reclaiming it. If a network failure makes the result uncertain, say so and direct the operator to inspect origin; never invent ownership or undo a potentially successful reservation.

Only update local status metadata after confirmed publication, preserving local body edits. If the local file changed meanwhile, preserve it and report the published claim and reconciliation needed. No claim_id, owner token, receipt store, takeover, or branch-derived ownership is added. Existing status/lifecycle vocabulary remains unchanged.

### 2. The agent selects the publication unit

The agent/operator determines which changes belong together and makes a coherent local Git commit. It may contain the issue, its separate durable plan, and deliberately chosen related project/design documents in the same repository. SDLC does not infer related files from Markdown prose or automatically gather every file with the issue number.

The proposed explicit surface is `sdlc issue publish --commit <commit>`. It resolves one non-merge commit with exactly one parent. Its entire change set is the publication unit; no silent path filtering. The initial eligible set is ordinary Markdown workflow records under configured issues/plans/history roots and `workshop/projects/`. A mixed code/document commit refuses with instructions to split it; code and code-coupled atlas changes retain the normal reviewed PR path. Root commits, merge commits, symlinks and submodules refuse rather than guessing a parent or copying executable content. This eligibility check enforces the workflow/code boundary, not semantic association between documents.

`issue sync --issue N` remains the deterministic local-only convenience for the issue file itself. It does not silently start including a separate plan. To publish a plan alongside an issue, the agent explicitly stages both and selects their commit. Explicit publication may include several issue files when the agent intentionally groups them. Files in another repository require that repository's own commit/publication.

### 3. Publish the change, not a replacement snapshot

Resolve the selected source commit and parent once. Fetch fresh remote main, apply their Git change with three-way merge semantics in a private temporary index/object transaction, then construct the result as a commit whose parent is the fetched remote tip. The source parent supplies the before-image; there is no separate baseline database.

Only publish if the remote ref still equals the observed tip, using an explicit expected-old-ref lease. Every proposed commit remains a child of that tip; this does not authorize overwriting intervening commits or accepting a rewind silently. On a remote race, repeat the merge against the new tip within three attempts. Unrelated changes survive; compatible edits to the same document may merge; an actual Git conflict refuses the whole selected commit and names the paths. Preserve the source commit, caller index, working files and branch on every outcome.

Record the source SHA in a `Source-Commit:` trailer in the resulting Git commit. This is ordinary Git provenance, not a separate receipt system. Before applying, check whether the source commit itself or a publication carrying that source SHA is reachable from fresh remote main. A confirmed prior publication is already applied, including after later edits or a deliberate revert; retry must not undo that later work. History-query failure or timeout refuses rather than assuming absence. A new deliberate reapplication requires a new source commit. A byte-identical empty result may report no change; it must not acquire a claim or claim a new remote commit was created.

For multiple sequential source commits, publish the prerequisites in order; do not infer or ship their other ancestors. If a chosen commit depends on an earlier unpublished edit, report a merge conflict or missing prerequisite rather than widening the selected set. If push acknowledgment is lost, fresh Git reachability/provenance can confirm publication; otherwise return an uncertain outcome with the source/candidate SHAs and preserve the source. There is no automatic durable recovery daemon or private write-ahead journal.

### 4. Same behavior in every checkout

Use this path in `:0`, numbered slots, ordinary feature worktrees and private dependency clones. Local main receives no whole-branch-push shortcut. Publication does not switch, reset, fast-forward or otherwise modify any checked-out branch, including :0. The agent explicitly integrates remote changes through normal Git when appropriate.

`issue new` retains its narrow automated reservation: check the fresh remote ID space inside the conditional publication attempt. Every occupied ID, including the same slug, is occupied; creation may select another free ID before it becomes an externally referenced identity. Preserve local creation content if publication is unconfirmed. Claims and updates never silently renumber existing issues.

### 5. Short review lock scopes

Retain local serialization for multi-step mutations. Linked worktrees share the common-directory lock; independent clones do not. The remote conditional write handles contention across both.

For plan-quality, estimate-quality, close and milestone-close: lock and capture inputs → unlock and run the reviewer → reacquire and validate → persist the result and continue. Capture branch/worktree identity, review anchor, issue/optional plan (including absence), the relevant gate ledger and first-round seed ledger, plus prepared project edits. Check before writing any authoritative sidecar, ledger round, cache, status or branch change, for every verdict.

Preserve close's existing allowance for unrelated documentation-only descendant commits. Changed read-set bytes, branch switches at the same SHA, changed code/history or concurrent ledger writes make the response stale. Staleness is not waivable through --force. No change to deterministic merge/push gates is needed solely for AI-review waits.

## Core concepts and seams

| Pure concept | File | Status |
|---|---|---|
| ClaimDecision: open/status/error → attempt or refusal | `cmd/sdlc/claimdecision.go` | new |
| CommitSelection: fixed source, parent, eligible changed paths | `cmd/sdlc/commitpublication.go` | new |
| PublicationStep: selected/merged/result → retry, success, conflict or uncertain | `cmd/sdlc/commitpublication.go` | new |
| PreparedReview: captured read set and legal review phase | `cmd/sdlc/reviewstate.go` | new |

| Integration | File | Status | External boundary |
|---|---|---|---|
| Claim/create publication | `claim.go`, `issue.go`, `synctrunk.go`, `issuecollision.go` | modified | fresh remote status/ID space |
| Selected-commit publication | `cmd/sdlc/issuepublish.go`, `internal/gitx/commitpublication.go` | new | Git commit resolution, three-way merge, provenance |
| Conditional trunk write | `internal/gitx/updatemany.go` | modified | exact-old-ref push |
| Review execution | `changecode.go`, `close.go`, `milestoneclose.go`, `internal/judge/dispatch.go` | modified | locks, subprocess, ledgers |

Reuse the existing Git runner and trunk commit construction. One publication implementation serves all checkout kinds (ARCH-DRY). Keep selection/status/result decisions pure (ARCH-PURE). The Git adapter uses a stateful fake modeling commits, trees, refs, conflicts and push outcomes, with temporary bare-Git conformance in normal tests (ARCH-MOCK).

### Existing caller migration

- `claim --issue N` becomes only fresh status reservation; local body changes are not swept into the claim. Legacy unfiltered claim and --no-start must direct users to explicit body publication instead of retaining a blind overwrite route.
- `issue sync` stays local. For compatibility, `issue sync --push` can publish the exact single-issue commit it just created. If no new commit was created, require the explicit publish command/SHA; do not guess a range or substitute the entire current file.
- `change-code` may publish the exact narrow issue commit it creates (for example its flow metadata update) through the same adapter. It must not infer a broader design package. If prior local commits need publication, tell the agent to select/publish those commits explicitly.
- `issue new` uses its separately defined reservation transaction; its collision policy is not ordinary Markdown merging.
- Audit `Makefile.workflow` compatibility wrappers, help/README/atlas, and every `syncIssuesToMain` caller. Retire the old whole-file replacement and whole-local-main push paths after migration.

### Minimal transition rules

| Operation/event | Result |
|---|---|
| Claim sees open / conditional push succeeds | confirmed reservation |
| Claim sees non-open, including a repeated own call | actionable refusal |
| Conditional push loses to another publication | reread and re-evaluate, at most three attempts |
| Selected commit is already reachable/published | confirmed already-applied result; do not replay |
| Three-way merge conflicts or selection is ineligible | refuse without touching caller state |
| Push/confirmation cannot establish outcome | uncertain; preserve source and show inspection information |
| Reviewer finishes and read set still matches | persist result under lock |
| Reviewer finishes but inputs/ledger/branch changed | stale refusal before persistence |
| Review cancels/times out or relock fails | stop without persistence; terminate and reap child on normal return |

Review lifetime retains the prior bounded policy: default 30 minutes, `WF_REVIEW_TIMEOUT` configurable from 1 second to 2 hours; invalid settings refuse before dispatch. Cancellation allows at most five seconds before kill/wait; bound pipe draining. Uncatchable parent death cannot promise portable child reaping, but cannot finalize a result; a later command starts from durable Git/ledger state. No asynchronous writer outlives its command.

Publication is a bounded CLI transaction. Three retries are sufficient for the intended handful of concurrent slots. Bound provenance queries to 30 seconds; timeout is a refusal. Temporary indexes/files are private and removed on normal success/error; durable source/result commits have ordinary repository retention. No new persistent store, claim vocabulary, migration of existing issues or cleanup service is required (ARCH-FUNERAL). Treat source refs, paths and Git responses as untrusted inputs; resolve exact commit OIDs, separate absent from failed queries, and never interpolate them into shell code (ARCH-SECURE/ORDER).

### Production test map and operating envelope

The following is the exhaustive risky-function inventory for this change. New names are implementation contracts; if a helper is renamed, update this map in the same commit. Each row names the adversarial strategy and the mutation that must make its test fail; scenario names in Tasks 1–3 refer to these production functions, not test-only models.

| Production functions | Adversarial strategy / mechanical guard |
|---|---|
| `claimDecision`, `runClaim` | Table-drive every vocabulary status and malformed remote bytes; rendezvous two real CLIs after observing open. Removing the fresh-status check or conditional ref must make exactly-one-winner assertions fail. |
| `decideCollision`, `syncViaTrunk`, `runIssueNew` | Occupy the proposed ID with identical and different slugs between observations; fail confirmation after push. Assert unique retained records and preservation of uncertain local content; treating equal slug as free must fail. |
| `TrunkFile.UpdateMany`, shared `pushExpected` | Stateful ref advancement/rewind and lost-ACK schedules plus real bare-Git hooks. Assert no stale overwrite and no prepare rerun after confirmed own candidate publication; ordinary unleased push must fail rewind regression. |
| `TrunkFile.SelectCommit`, `validateCommitSelection` | Real root/merge/regular commits and all changed path/mode classes, including rename endpoints. Assert complete-set eligibility before effects; ignoring one ineligible path must fail. |
| `TrunkFile.PublishCommit` | Model refs/trees/provenance across conflict, rejection, unknown acknowledgment, successful retry, later edit and revert. Real Git confirms compatible merge and caller-state preservation. Removing exact trailer/ancestry checks must replay the revert and fail. |
| `runIssuePublish`, `syncIssuesToMain`, `syncIssue` | CLI matrix across primary/worktree/clone, isolated HOME, bare remotes, staged unrelated code and older unselected ancestors. Compare remote trees and local branch/index/dirty files. Restoring main's whole-branch shortcut or no-new-commit inference must fail. |
| `capturePreparedReview`, `PreparedReview.validate` | Change each read-set member independently, including missing→present plan, branch identity and ledger generation. Exact read-set oracle rejects each; retain separate docs-only descendant acceptance test. Omitting any captured member must fail. |
| `runPlanQualityJudge`, `runEstimateQualityJudge`, `reviewThenFinalizeLocked`, `finalizeBoundaryReview` | Barrier-controlled reviewer process pauses each review kind while unrelated commands finish; mutate inputs before every verdict (success and failure). Assert no stale ledger/sidecar/status/branch writes. Moving persistence ahead of validation or retaining lock during dispatch must fail. |
| `judge.Dispatch`, reviewer process runner | Real controlled child ignores graceful cancel and holds output pipes; short deadline/cancel cases assert bounded return, child reaping and no authority writes. Removing kill/wait or bounded drain must fail timeout test. |

ARCH-CONSTRAINTS: this is an interactive developer CLI, not a keystroke/UI or server path. Domain-informed workload assumption: a handful (roughly 2–8) of concurrent slots, small Markdown commits and an ordinary developer Git repository. No daemon, fan-out, memory-resident repository cache, or throughput promise is introduced. Excess contention exhausts three exact-ref attempts and returns an actionable retry error rather than spinning. Re-measure with the deterministic race suite and wall-clock CLI timings if fleet concurrency grows.

Retain the existing 30-minute local lock-wait ceiling (`repolock.DefaultWaitTimeout`); acquisition/reacquisition failure performs no subsequent authoritative write and tells the operator to inspect the holder. This compatibility ceiling is not a target latency: tests require an unrelated local command to finish while a reviewer remains paused. Reviewer time is intentionally minutes; its existing 30-minute default and configurable ceiling are explicit above. The 30-second history-query deadline is a provisional interactive budget, exercised with an injected stalled Git process; large/slow repositories get an explicit refusal rather than skipped provenance. Normal Git network transport remains subject to Git/SSH timeout and user cancellation behavior; this change does not promise a total network deadline. A failed/interrupting push is uncertain until reachability confirms its candidate, never evidence of absence. Test bare local remotes deterministically; external network latency/service availability is outside the performance claim. CPU/disk work is one selected merge per attempt; no repository-wide checkout or copy. Performance beyond ordinary developer repositories is not claimed; record representative repository size and query timing in verification rather than inventing a benchmark guarantee.

Conformance cadence: run `go test ./cmd/sdlc/internal/gitx/... -count=1` against the locally installed real Git and temporary bare remotes on every publication-adapter change and Git upgrade, and in the normal CI/test suite before merging. `TestCommitPublicationMerge`, `TestCommitPublicationRetry` and UpdateMany conformance cases execute matching fake/real schedules; any mismatch is a failing test that blocks shipping until the adapter or fake is corrected. This is live binary conformance; no GitHub service or credential is needed for the Git semantics used here.

Retention: reachable publication commits intentionally remain in the repository's ordinary history for its lifetime, owned by the repository operator; rewriting that history can remove the provenance guarantee and is outside this command's contract. Each effective publication adds one commit and only changed trees/blobs, plus one OID-sized trailer, rather than a second snapshot database. Rejected temporary objects are unreachable and eligible for ordinary Git GC. The cost grows with effective publications and changed-document bytes; verification records `git count-objects -vH` and provenance lookup timing on a repeated-publication fixture. This is an explicit unbounded historical retention policy, shared with every normal Git commit, with a bounded reader that refuses on timeout; no automatic history rewrite or deletion is appropriate for workflow audit evidence.

## Chunk 1 — Claims and selected-commit publication (M1)

### Task 0: Reconcile the paused v1 draft

- [x] Retired the abandoned uncommitted draft outside the repository at `/var/folders/07/b9wcwwld4_v2w9r3hk525bm80000gn/T/ariadne-244-retired-draft-j6pdwxfl`; removed its receipt/token implementation.
- [ ] Retain useful behavior regressions: unrelated code leakage, same-record overwrite, same-slug creation collision, empty-file presence, remote rewind. Replace assertions/API sketches that depend on removed ownership receipts.
- [x] Removed v1-only production/test placeholders and restored the tracked pre-draft files. Baseline targeted Git/command tests passed before revised implementation.

### Task 1: Fresh status claims and reservation allocation

Files: `claim.go`, `claimdecision.go`, `issue.go`, `issuecollision.go`, `synctrunk.go`, `internal/gitx/updatemany.go`, related tests.

- [ ] Write `TestClaimRemoteStatusRace`: independent clones rendezvous after observing remote open; linked worktrees rendezvous before lock acquisition. Assert exactly one winner and an explicit non-open loser, with no production-lock bypass.
- [ ] Write `TestClaimNonOpenAndUncertain`: vocabulary statuses, malformed/absent records, repeat claim and lost acknowledgment; no owner inference and no local-body loss.
- [ ] Write `TestCreationOccupiedID`: same/different slugs at one candidate ID, fresh retry and interrupted publication; no overwritten reservation.
- [ ] Observe expected failures; implement pure decisions and conditional publication; rerun to green.
- [ ] Run `go test ./cmd/sdlc/internal/gitx/... ./cmd/sdlc -run 'Test(Claim|Creation|StartOnClaim|AllocateIssueID|UpdateMany|Trunk)' -count=1`.

### Task 2: Explicit commit publication and callers

Files: new `issuepublish.go`, `commitpublication.go`, `internal/gitx/commitpublication.go` and tests; modify `issue.go`, `claim.go`, `synctrunk.go`, `changecode.go`, `Makefile.workflow` as required.

- [ ] `TestCommitSelection`: real commit topology and typed path validation; accept intentionally grouped issue/plan/project documents, refuse mixed code, root/merge commits, symlinks, malformed refs and silent partial publication.
- [ ] `TestCommitPublicationMerge`: stateful Git adapter and real bare remote; preserve unrelated changes and compatible same-file edits, expose conflicting edits/deletion/rename without changing caller files/index/branch.
- [ ] `TestCommitPublicationRetry`: force remote advancement/rewind, repeat source SHA, later edit/revert and missing acknowledgment; never overwrite newer state or replay a previously published change.
- [ ] `TestCommitPublicationSlots`: run the same production CLI in primary :0, two linked numbered worktrees and private dependency clones; preserve unrelated local commits and dirty files in every case.
- [ ] `TestCommitPublicationCallers`: local-only sync performs no network; explicit selected commit includes its complete eligible set; legacy convenience callers do not infer or publish unselected ancestors/plan files.
- [ ] Implement through one shared adapter after observing red tests. Keep existing workflow gates intact.
- [ ] Run `go test ./cmd/sdlc/internal/gitx/... ./cmd/sdlc -run 'Test(CommitPublication|CommitSelection|IssueSync|RunIssueNew|Sync|Claim)' -count=1`.
- [ ] Update README, issue/claim help, `atlas/workflow/issue-sync.md`, `issue-lifecycle.md` and the Pair project checkpoint. Document selected files, primary-slot symmetry, conflict handling and repeated-claim refusal.
- [ ] Commit explicit paths and run `sdlc milestone-close --issue 244 --milestone M1 --verified '<observed evidence>'`.

## Chunk 2 — Review lock scope and integration (M2)

### Task 3: Prepare, unlock, validate, finalize

Files: new `reviewstate.go` and tests; modify `changecode.go`, `close.go`, `milestoneclose.go`, `internal/judge/dispatch.go`, review/lock tests.

- [ ] `TestPreparedReviewReadSet`: change each captured artifact/identity independently, including file appearance and ledger generation; assert gate-specific acceptance and no persistence for stale results.
- [ ] `TestReviewConcurrencySchedules`: barrier-controlled real reviewer subprocess; unrelated issue commands finish while each external review is paused; competing same-boundary responses cannot both advance the ledger.
- [ ] `TestReviewInterruption`: injected cancellation, deadline and relock errors; no passing cache/status/branch mutation, bounded process cleanup.
- [ ] Implement shared snapshots and manual lock phases after observing failures. Preserve plan-before-estimate order, pass-through caching and existing legitimate gate waivers.
- [ ] Run `go test ./cmd/sdlc -run 'Test(PreparedReview|ReviewConcurrency|ReviewInterruption|ChangeCode|Close|Milestone|PlanQuality|EstimateQuality)' -count=1`.

### Task 4: Verify and ship

- [ ] Run a nested dependency workflow using ordinary CLI creation, claim, local issue/plan commit, explicit publication and conflict recovery; no recursive dependency publication.
- [ ] Run `go test ./pkg/workspace/... ./cmd/sdlc/... -count=1 -skip '^TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory$'` (the pre-existing #210 missing-history fixture only), `go vet ./pkg/workspace/... ./cmd/sdlc/...`, and scoped `git diff --check`.
- [ ] Update review/lock atlas/help and project evidence; record lessons. Commit and close M2, then the issue, through the SDLC gates. Address blocking findings before publication.
- [ ] Publish through `sdlc pr` and `sdlc merge`; verify archived links and final repository state.

## Revisions

### 2026-09-23 — Simplified after operator discussion

Reason: the earlier design added ownership/retry conveniences beyond the required reservation model. Delta: fresh open→working conditional claims replace tokens; explicitly selected Git commits and three-way merge replace whole-file snapshots and receipt baselines; publication includes the agent's chosen issue/plan/project files and behaves identically in :0 and other checkouts. Short review locks remain required.

The operator requested removal of the old design to avoid misleading future agents. It is available only in Git history (commit `bbd3122`). Earlier gate results apply to that superseded design only; derive a new estimate after the current plan clears its gate.

### 2026-09-23 — Remove obsolete design from the working tree

Operator authorized implementation and asked to remove the old design. Deleted the superseded design artifact and retired its unfinished draft outside the repository. Git history preserves the old decisions; only this plan is active.

### 2026-09-23 — Gate review coverage and runtime assumptions

Reason: revised gate reported executable-test-strategy and operating-envelope gaps. Delta: enumerate every risky production function with its adversarial test and mutation guard; name concurrency, lock-wait, history-query and reviewer budgets and network limitations. No ownership or receipt mechanism is reintroduced.

### 2026-09-23 — Conformance cadence and Git retention

Reason: gate follow-up requested the real-Git execution cadence and durable artifact policy. Delta: normal tests and every adapter/Git update execute matching fake/live-binary schedules; reachable publication history is intentionally retained as ordinary Git audit history, with measured fixture size/query time and ordinary GC for unreachable objects.
