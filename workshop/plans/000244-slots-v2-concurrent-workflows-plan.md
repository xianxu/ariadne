# Concurrent issue workflows implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3 for execution strategy. Use bounded implementation agents where context is explicit; the SDLC binary owns milestone reviews. Use test-driven development and verification-before-completion.

**Goal:** Concurrent worktrees and independent dependency clones reserve and publish issues without lost updates, while external reviews leave unrelated repository operations available.

**Architecture:** Extend the existing remote-trunk transaction with issue-specific expected-content checks and durable publication intents. Keep local-only issue sync unchanged. Run external reviews between short locked prepare/finalize phases, validating their full input set before recording results.

**Tech Stack:** Go, Cobra, Git object plumbing, local bare Git remotes, filesystem receipts, existing judge/gatestate machinery.

**State:** Proposed; implementation awaits operator approval and `sdlc change-code`. Estimates follow the plan-quality gate.

## Contract and audit

The agreed scope and Done when in `workshop/issues/000244-slots-v2-concurrent-workflows.md` govern this plan. #243 provides nested workspace identity and environment-local peer lookup. No Couch changes, automatic ownership transfer, recursive publication, slot refresh, or new merge policy belong here.

Live-code audit on 2026-09-23 found:

- `claim.go` routes main through a whole-branch push; other branches use `synctrunk.go` and `gitx.TrunkFile.UpdateMany`.
- Claim writes status/timestamps without ownership identity. Equal bytes cannot distinguish two claimants.
- Trunk retries recheck ID collisions but replace same-path bodies; they also reread local bytes on each retry.
- First allocation treats an occupied identical filename as the same issue. Equal-slug creators can collide silently.
- `change-code` holds the shared Git-common-directory lock across plan and estimate reviews.
- Close/milestone-close release the lock already, but persist review sidecars/ledgers before validating prepared state. Their snapshot omits branch identity and ledger inputs.
- Merge/push have deterministic publication gates, not external LLM review waits. Preserve their serialization; do not introduce an unnecessary review abstraction there.

### Alternatives

1. **Recommended: expected-content publication through existing trunk transactions.** Small issue-specific state makes independent clones safe and preserves existing command boundaries.
2. Git merging issue bodies would reduce some conflicts, but cannot establish reservation ownership and risks silently combining contradictory workflow records.
3. A shared coordination service would centralize reservations but add an operational dependency; Git already supplies the authoritative publication boundary.

## Core concepts

| Pure entity | Lives in | Status |
|---|---|---|
| PublicationIntent and publication transition function | `cmd/sdlc/issuepublication.go` | new |
| ClaimIdentity | `cmd/sdlc/issuepublication.go`, `construct/vocabulary/issue.cue` | new |
| PublicationConflict | `cmd/sdlc/issuepublication.go` | new |
| PreparedReview and artifact comparison | `cmd/sdlc/reviewstate.go` | new |

PublicationIntent owns one immutable target and a sorted set of expected/desired path states. A path state distinguishes absent from a present empty blob. It includes operation kind (create, claim, update), issue identity, unique operation ID, and any claim identity. Conflict decisions are pure. A claim identity is an opaque random identifier, not a secret or an authentication credential; it distinguishes this worktree's intent from another's. Existing issue status values remain unchanged.

PreparedReview owns the exact data read by a review plus its Git/worktree identity and destinations. Share artifact capture/comparison; retain each gate's existing acceptance policy, including close's intentional allowance for unrelated descendant documentation commits.

| Integration point | Lives in | Status | Wraps |
|---|---|---|---|
| Issue publication dispatch | `cmd/sdlc/claim.go`, `synctrunk.go`, `issue.go`, `issuecollision.go` | modified | local records and trunk transactions |
| Worktree-private receipt store | `cmd/sdlc/publicationreceipt.go` | new | atomic files under private Git directory |
| Trunk transaction evidence | `cmd/sdlc/internal/gitx/updatemany.go`, `trunkfile.go` | modified | fetch, commit construction, push, reconciliation |
| Review prepare/dispatch/finalize | `cmd/sdlc/changecode.go`, `close.go`, `milestoneclose.go`, `reviewstate.go` | modified | repo lock, judge process, gate ledgers |

### Publication policy

All publishing issue verbs use the trunk transaction regardless of local branch. Plain `issue sync` still commits locally without network. Publication commits only the selected issue paths and required ownership metadata; it does not push unrelated local commits, switch branches, reset worktrees, or move a checked-out main. Report the published commit and any local branch divergence explicitly. Existing local issue edits remain durable via normal sync.

Snapshot intended bytes once under the local repository lock. Every attempt reads a fresh remote tree and compares touched records to the intent's expected state. Unrelated records are preserved by constructing the commit atop that tree. A conflicting same-record change refuses with issue/path, expected and observed revision, and instructions to reconcile explicitly and retry; preserve local content. Apply the same rule to deletion and both sides of a rename. Do not merge arbitrary Markdown automatically.

Initial expected state comes from provable shared Git history, never from simply accepting the newest fetched blob. A worktree's confirmed receipt supersedes that initial baseline after remote-only publication, allowing successive edits without forcing a merge of trunk into the feature branch. Missing/ambiguous ancestry or an unreadable receipt refuses rather than guessing. A baseline can advance only through confirmed publication or explicit Git integration containing the observed revision; fetching alone does not accept a conflicting edit.

Claims require fresh remote evidence. An open issue can be claimed with a new locally persisted claim identity; an already claimed issue is idempotent only for the same locally held identity. A competing worktree/clone refuses even if its desired body is identical. A copied branch does not carry the private receipt. Legacy working issues without claim identity remain editable via explicit sync; `claim` cannot assert ownership from their status alone and must explain that limitation. No automatic takeover is added. `--no-start` and unfiltered legacy claim remain body-publication operations, not ownership acquisition.

New creation regards every occupied ID as taken, including identical slugs and archived records, unless the pending operation proves this is its own retry. Reuse existing bounded reallocation and identity rewriting. Preserve all unresolved local candidates until remote outcome is known; never delete evidence merely because a push returned an error.

### Durable intent and recovery

Use one versioned receipt per issue/publication operation in a worktree-private Git subdirectory, keyed by a digest of repository identity, remote target, and configured issue namespace. Store expected/desired blob identities and bytes needed for recovery, claim identity, and attempted commit identity. Do not persist credential-bearing remote URLs; compare a target digest. Strict parsing rejects unknown versions, invalid paths/OIDs, truncated records, and mismatched targets.

Persist pending intent atomically before a push; record the built commit before dispatching that push. Extend the existing trunk seam to expose this evidence without creating another retry loop. Fresh remote evidence that the attempted commit is current or an ancestor confirms the operation. An error without such evidence is an unknown outcome, not proof of failure. Reconcile on retry before reallocating, deleting candidates, or starting a replacement intent. If newer changes follow the confirmed commit, retain the confirmed baseline and apply the normal conflict check to further edits.

The remote write must enforce the exact observed old ref, including remote rewind races; use an explicit expected-old-ref lease, and build the proposed commit as a child of that observed ref. This is not permission to replace intervening work. Retain the existing three-attempt contention bound; auth/query errors must not become retries reporting success.

| State/event | Decision/effect |
|---|---|
| Unprepared + requested operation | capture expected/desired state, persist pending intent |
| Pending + remote matches expected | build commit, persist attempt evidence, attempt conditional push |
| Pending + unrelated remote change | rebuild on fresh parent within retry bound |
| Pending + same-record conflict / foreign claim | refuse; preserve intent and local bytes |
| Pending + push acknowledged or attempted commit proven reachable | confirm receipt, then perform safe local candidate cleanup |
| Pending + failed push and unavailable confirmation | retain unknown outcome; actionable retry |
| Confirmed + same operation retry | verify fresh authority; idempotent success if still applicable |
| Confirmed + new local edit | new intent using confirmed baseline |
| Any + malformed state / target mismatch | refuse without publication or deletion |

**Invariants:** one winner per claim; unique published creation IDs; no replacement of changed records; no ownership gained by copying Git history; no success inferred from an unavailable probe; no cleanup before confirmed outcome. The pure transition function owns authoritative state changes; IO executes its declared effects and feeds outcomes back.

**Lifetime and bounds:** atomic temp files are removed on completion/error. Confirmed receipts are superseded, not appended. Retire a receipt after fresh evidence of terminal archival and no pending operation; worktree removal also removes its private Git directory. Pending state is never age-deleted. Impose a documented conservative receipt size/count limit and refuse new operations at the bound with recovery instructions; choose constants from repository-size fixtures during implementation, before any write. No background service or accumulating transcript store is introduced.

### Review policy

Prepare under the repository lock, dispatch without it, reacquire and validate before any authoritative write. Capture the issue, optional plan including absence, relevant gate ledger including absence, branch/worktree identity, and Git anchor. Close also captures project destinations and the plan-ledger seed used for a first boundary round. Validate all verdicts, not just SHIP. A stale result must not overwrite a sidecar, advance a ledger/cap, set a passing cache, change status, or branch. Print the response for diagnosis and require rerun; no extra durable stale-response artifact is necessary.

For `change-code`, preserve structural → plan-quality → estimate/reconcile → estimate-quality ordering, existing cached pass-through, and quick-flow gates. Split external dispatch from result persistence, not the entire command into a second workflow framework. Revalidate again before branching/publication if another unlocked wait occurred. Staleness is a concurrency failure that `--force` and judge waivers cannot waive.

Close retains unrelated docs-only descendant acceptance, but rejects changed read-set artifacts, branch switches even at the same SHA, code/history changes, and concurrent same-boundary ledger writes. Peer project writes retain their existing exact-content guard; this work does not claim a multi-repository atomic close.

## Operating envelope and test seams

This is a small number of concurrent CLI workflows, not a high-throughput service. Each publication handles the selected issue set; remote retries remain bounded at three. Existing lock timeout/cancellation semantics stay in force. Reviews can last minutes; an unrelated issue command must complete while a reviewer is deliberately paused, without releasing that reviewer. Network/git latency is not promised to be eliminated.

ARCH-PURE/DRY: use one decision model and the existing trunk adapter. ARCH-ORDER/SECURE: explicit uncertain outcomes, immutable desired content, exact record preconditions, strict receipt parsing. ARCH-MOCK: stateful fake remote/receipt adapters share production seams and model refs, blobs, conditional writes and lost acknowledgments. Real temporary bare-Git tests provide conformance on each affected test run. No test may inherit production remotes, credentials, HOME, or Git config. ARCH-FUNERAL: bounded receipts and cleaned temporary indexes; ARCH-PURPOSE: exercise both shared worktrees and independent private clones.

## Chunk 1: Publication (M1)

### Task 1 — Specify and test reservation/publication decisions

Files: new `issuepublication.go`, `issuepublication_test.go`, `publicationreceipt.go`, `publicationreceipt_test.go`; modify `construct/vocabulary/issue.cue` and its generated consumer through the existing vocabulary generation workflow.

- [ ] Write pure failing sequence tests for every transition above, including equal-byte claims, copied ownership metadata, same-slug creation, delete/update, rename/update and failed confirmation.
- [ ] Implement typed decisions and bounded versioned receipt IO. Test atomically interrupted writes, unknown schema, invalid paths, empty-vs-absent state, target mismatch and cleanup.
- [ ] Run `go test ./cmd/sdlc -run 'Test(Publication|ClaimIdentity|PublicationReceipt)' -count=1`; demonstrate meaningful red then green. Validate vocabulary across existing issue records if its schema changes.

### Task 2 — Wire the existing transaction and every issue publisher

Files: `claim.go`, `synctrunk.go`, `issue.go`, `issuecollision.go`, `internal/gitx/updatemany.go`, associated tests and new `publication_concurrency_test.go`.

- [ ] Add failing stateful-fake tests for immutable desired bytes, per-attempt preconditions, exact-old-ref comparison, and push-success/lost-ack recovery.
- [ ] Route main and non-main publication through the same guarded path; keep `NoPush` local-only. Preserve explicit already-committed-body publication.
- [ ] Ensure claim and creation carry distinct pending identities and cannot adopt a foreign reservation. Update old tests that intentionally pinned last-writer-wins or offline ownership success.
- [ ] Add real bare-remote fixtures with two independent clones and two linked worktrees. Use process barriers, not timing sleeps, to force both contenders past observation before either publishes.
- [ ] Prove one claim winner, two unique creation records (same and different slugs), preservation of unrelated updates, visible same-record conflicts, sequential remote-only updates from unchanged HEAD, crash/restart recovery, and selected issue publication without unrelated local-main commits.
- [ ] Run `go test ./cmd/sdlc/internal/gitx/... ./cmd/sdlc -run 'Test(Trunk|UpdateMany|Publication|Claim|Sync|IssueSync|RunIssueNew|AllocateIssueID)' -count=1` and inspect every failure.
- [ ] Update `README.md`, `cmd/sdlc/helptext/claim.md`, `helptext/issue.md`, `atlas/workflow/issue-sync.md`, `atlas/workflow/issue-lifecycle.md`; remove the known last-writer-wins/offline-success claims and explain local-main divergence/recovery. Keep atlas index links current.
- [ ] Commit explicit paths; close M1 via `sdlc milestone-close --issue 244 --milestone M1 --verified '<actual evidence>'`. Update the Pair project checkpoint.

## Chunk 2: Review concurrency and acceptance (M2)

### Task 3 — Reuse prepared review state and unlock external planning reviews

Files: new `reviewstate.go`, `reviewstate_test.go`; modify `changecode.go`, `close.go`, `milestoneclose.go`, `repolock.go` only as needed; existing plan/boundary ledger adapters and tests.

- [ ] Write failing tests for stale branch identity, issue/plan appearance or edits, gate-ledger changes, and non-finalizing stale results writing records.
- [ ] Extract artifact capture/comparison. Convert change-code to manual prepare/dispatch/finalize locking and preserve ordering/pass-through/waivers.
- [ ] Validate close/milestone-close before sidecar/ledger writes for every result. Include boundary ledger and seed plan ledger in the prepared read set.
- [ ] Add real subprocess tests using a fake reviewer with ready/release barriers: while plan, estimate, close or milestone review is paused, another worktree completes an unrelated issue operation. Concurrent same-boundary reviews must reject the stale second result without advancing the ledger.
- [ ] Retain docs-only descendant acceptance and code/diverged-anchor refusal. Test cancellation/error releases locks and reaps reviewer children.
- [ ] Run `go test ./cmd/sdlc -run 'Test(PreparedReview|ChangeCode|Close|Milestone|PlanQuality|EstimateQuality|ReviewConcurrency)' -count=1` and confirm red/green evidence for new tests.

### Task 4 — End-to-end dependency workflow and publication

Files: `publication_concurrency_test.go`, new `review_concurrency_test.go`; `atlas/workflow/sdlc-binary.md`, `atlas/workflow/gate-state.md`, embedded help, `README.md`, and the Pair couch-slots-v2 project.

- [ ] In temporary nested Pair-slot environments, drive ordinary Ariadne clones through normal CLI creation, claim, local body sync and explicit publication. Assert local edits survive conflict/offline conditions and each environment retains its Git state.
- [ ] Run `go test ./pkg/workspace/... ./cmd/sdlc/... -count=1 -skip '^TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory$'`; the skip is the pre-existing #210 missing historical plan, not a #244 exemption. Run `go vet ./pkg/workspace/... ./cmd/sdlc/...` and scoped `git diff --check`.
- [ ] Document concurrency/refusal/retry behavior and observed test evidence. Record discovered review lessons. Update Pair project scope/checkpoint without marking the whole project complete.
- [ ] Commit; run `sdlc milestone-close --issue 244 --milestone M2 --verified '<actual evidence>'`, then the issue close gate. Fix blocking review findings before continuing.
- [ ] Publish through `sdlc pr` and `sdlc merge`, and verify archive/project links and repository state. No publication is claimed by this plan.
