# Superseded #244 design — preserved 2026-09-23

This is provenance only. The operator replaced the receipt/ownership design with status claims and explicitly selected Git commits. The canonical `000244-slots-v2-concurrent-workflows-plan.md` governs implementation. The earlier gate verdicts and estimate apply only to this superseded design.

# Concurrent issue workflows implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3 for execution strategy. Use bounded implementation agents where context is explicit; the SDLC binary owns milestone reviews. Use test-driven development and verification-before-completion.

**Goal:** Concurrent worktrees and independent dependency clones reserve and publish issues without lost updates, while external reviews leave unrelated repository operations available.

**Architecture:** Extend the existing remote-trunk transaction with issue-specific expected-content checks and durable publication intents. Keep local-only issue sync unchanged. Run external reviews between short locked prepare/finalize phases, validating their full input set before recording results.

**Tech Stack:** Go, Cobra, Git object plumbing, local bare Git remotes, filesystem receipts, existing judge/gatestate machinery.

**State:** Approved by the operator; `sdlc change-code` cleared plan-quality and estimate-quality on 2026-09-23. Implementation is on `000244-slots-v2-concurrent-workflows`.

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

Claims require fresh remote evidence. An open issue can be claimed with a new locally persisted claim identity; an already claimed issue is idempotent only for the same locally held identity. A competing worktree/clone refuses even if its desired body is identical. A copied branch does not carry the private receipt. Legacy working issues without claim identity remain editable via explicit sync; `claim` cannot assert ownership from their status alone and must explain that limitation. No automatic takeover is added. Generic body publication must preserve the authoritative claim identity; only the dedicated claim transition may create/change it for an open issue. Copying or editing metadata cannot make a body sync transfer a reservation. `--no-start` and unfiltered legacy claim remain body-publication operations, not ownership acquisition.

New creation regards every occupied ID as taken, including identical slugs and archived records, unless the pending operation proves this is its own retry. Reuse existing bounded reallocation and identity rewriting. Preserve all unresolved local candidates until remote outcome is known; never delete evidence merely because a push returned an error.

### Durable intent and recovery

Use one versioned receipt per issue/publication operation in a worktree-private Git subdirectory, keyed by a digest of repository identity, remote target, and configured issue namespace. Store expected/desired blob identities and bytes needed for recovery, claim identity, and attempted commit identity. Do not persist credential-bearing remote URLs; compare a target digest. Strict parsing rejects unknown versions, invalid paths/OIDs, truncated records, and mismatched targets.

Persist pending intent atomically before a push; record the built commit before dispatching that push. Extend the existing trunk seam to expose this evidence without creating another retry loop. Fresh remote evidence that the attempted commit is current or an ancestor confirms the operation. An error without such evidence is an unknown outcome, not proof of failure. Reconcile on retry before reallocating, deleting candidates, or starting a replacement intent. If newer changes follow the confirmed commit, retain the confirmed baseline and apply the normal conflict check to further edits.

A conflict found before dispatch is confirmed nonapplication. A push explicitly rejected by the server is also confirmed nonapplication for that attempt; parse the Git push porcelain rejection result, not arbitrary stderr prose or a nonzero exit code. Persist this distinction and the conflicting observed revision. A transport error or incomplete response remains unknown even when a subsequent read does not contain the attempted commit. Unknown intents must reconcile before replacement, and exhaustion preserves them with an actionable diagnostic.

For a confirmed-nonapplied body intent, an ordinary retry remains a conflict until the local checkout has explicitly integrated the recorded conflicting revision (prove ancestry) and supplied the resolved desired body. Then atomically supersede the rejected intent with a new expected/desired snapshot. This is explicit Git reconciliation, not merely fetching. If the remote advanced again, compare the new expected state and refuse again as needed. A foreign claim conflict never becomes a successful takeover through this recovery path. Creation may reallocate after a proven rejected attempt using a fresh ID-space read; it retains original content and candidate evidence until final publication is confirmed. A successful acknowledged attempt later removed by a remote rewind is a conflict, not permission to replay ownership silently.

The remote write must enforce the exact observed old ref, including remote rewind races; use an explicit expected-old-ref lease, and build the proposed commit as a child of that observed ref. This is not permission to replace intervening work. Retain the existing three-attempt contention bound; auth/query errors must not become retries reporting success.

| State/event | Decision/effect |
|---|---|
| Unprepared + requested operation | capture expected/desired state, persist pending intent |
| Pending + remote matches expected | build commit, persist attempt evidence, attempt conditional push |
| Pending + unrelated remote change | rebuild on fresh parent within retry bound |
| Pending + same-record conflict / foreign claim before push | record confirmed-nonapplied conflict and observed revision; preserve local bytes |
| Pending + explicit server rejection | record confirmed nonapplication; fresh allocation/retry allowed under record preconditions |
| Rejected body intent + proven local integration of conflicting revision | atomically supersede with resolved desired snapshot and integrated baseline |
| Rejected creation + fresh occupied-ID evidence | reallocate within the same bounded operation, retaining candidates |
| Pending + push acknowledged or attempted commit proven reachable | confirm receipt, then perform safe local candidate cleanup |
| Pending + failed push and unavailable confirmation | retain unknown outcome; actionable retry |
| Confirmed + same operation retry | verify fresh authority; idempotent success if still applicable |
| Confirmed + new local edit | new intent using confirmed baseline |
| Any + malformed state / target mismatch | refuse without publication or deletion |

**Invariants:** one winner per claim; unique published creation IDs; no replacement of changed records; no ownership gained by copying Git history; no success inferred from an unavailable probe; no cleanup before confirmed outcome. The pure transition function owns authoritative state changes; IO executes its declared effects and feeds outcomes back.

**Lifetime and bounds:** atomic temp files are removed on completion/error. Confirmed receipts are superseded, not appended. Retire a receipt after fresh evidence of terminal archival and no pending operation; worktree removal also removes its private Git directory. Pending state is never age-deleted. Initial conservative operating limits (design assumptions, not measured requirements): 16 MiB per pending intent, 4096 receipt entries and 256 MiB total per worktree. Check projected usage before writing; refuse new operations at the bound with recovery instructions, while allowing reconciliation/retirement of existing entries. Test each limit. Change a limit only with recorded fixture evidence and a plan revision. No background service or accumulating transcript store is introduced.

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

- [ ] Write failing `TestPublicationTransitionSequences` and `TestPublicationConflictingRecords` using the named-function strategies below.
- [ ] Implement typed decisions and bounded receipt IO; follow `TestPublicationReceiptRejectsMalformed` and `TestPublicationReceiptInterruptedWrite` below.
- [ ] Run `go test ./cmd/sdlc -run 'Test(Publication|ClaimIdentity|PublicationReceipt)' -count=1`; demonstrate meaningful red then green. Validate vocabulary across existing issue records if its schema changes.

### Task 2 — Wire the existing transaction and every issue publisher

Files: `claim.go`, `synctrunk.go`, `issue.go`, `issuecollision.go`, `internal/gitx/updatemany.go`, associated tests and new `publication_concurrency_test.go`.

- [ ] Add failing `TestPublicationRemoteInterleavings` through the shared production seam using its named adversarial strategy below.
- [ ] Route main and non-main publication through the same guarded path; keep `NoPush` local-only. Preserve explicit already-committed-body publication.
- [ ] Ensure claim and creation carry distinct pending identities and cannot adopt a foreign reservation. Update old tests that intentionally pinned last-writer-wins or offline ownership success.
- [ ] Add real bare-remote fixtures with two independent clones and two linked worktrees. Use process barriers, not timing sleeps. Independent clones rendezvous after remote observation to force CAS contention. Linked worktrees rendezvous before lock acquisition; retain the production common-directory lock and assert serialization plus the second claimant's fresh-authority refusal. Never bypass locking to manufacture a shared-worktree interleaving.
- [ ] Implement `TestPublicationClaimContenders`, `TestPublicationAllocationContenders`, `TestPublicationConflictRecovery` and `TestPublicationMainScope` using the named strategies below.
- [ ] Run `go test ./cmd/sdlc/internal/gitx/... ./cmd/sdlc -run 'Test(Trunk|UpdateMany|Publication|Claim|Sync|IssueSync|RunIssueNew|AllocateIssueID)' -count=1` and inspect every failure.
- [ ] Update `README.md`, `cmd/sdlc/helptext/claim.md`, `helptext/issue.md`, `atlas/workflow/issue-sync.md`, `atlas/workflow/issue-lifecycle.md`; remove the known last-writer-wins/offline-success claims and explain local-main divergence/recovery. Keep atlas index links current.
- [ ] Commit explicit paths; close M1 via `sdlc milestone-close --issue 244 --milestone M1 --verified '<actual evidence>'`. Update the Pair project checkpoint.

## Chunk 2: Review concurrency and acceptance (M2)

### Task 3 — Reuse prepared review state and unlock external planning reviews

Files: new `reviewstate.go`, `reviewstate_test.go`; modify `changecode.go`, `close.go`, `milestoneclose.go`, `repolock.go` only as needed; existing plan/boundary ledger adapters and tests.

- [ ] Write failing `TestPreparedReviewReadSet` using the comparison and no-write oracle below.
- [ ] Extract artifact capture/comparison. Convert change-code to manual prepare/dispatch/finalize locking and preserve ordering/pass-through/waivers.
- [ ] Validate close/milestone-close before sidecar/ledger writes for every result. Include boundary ledger and seed plan ledger in the prepared read set.
- [ ] Implement `TestReviewConcurrencySchedules` through real command subprocesses for each external review caller, following the barrier strategy below.
- [ ] Run the prepared-review read-set and interruption transition tests below, preserving existing gate-specific anchor policy.
- [ ] Run `go test ./cmd/sdlc -run 'Test(PreparedReview|ChangeCode|Close|Milestone|PlanQuality|EstimateQuality|ReviewConcurrency)' -count=1` and confirm red/green evidence for new tests.

### Task 4 — End-to-end dependency workflow and publication

Files: `publication_concurrency_test.go`, new `review_concurrency_test.go`; `atlas/workflow/sdlc-binary.md`, `atlas/workflow/gate-state.md`, embedded help, `README.md`, and the Pair couch-slots-v2 project.

- [ ] Implement `TestPublicationDependencyEnvironment` with the temporary nested-clone CLI strategy below.
- [ ] Run `go test ./pkg/workspace/... ./cmd/sdlc/... -count=1 -skip '^TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory$'`; the skip is the pre-existing #210 missing historical plan, not a #244 exemption. Run `go vet ./pkg/workspace/... ./cmd/sdlc/...` and scoped `git diff --check`.
- [ ] Document concurrency/refusal/retry behavior and observed test evidence. Record discovered review lessons. Update Pair project scope/checkpoint without marking the whole project complete.
- [ ] Commit; run `sdlc milestone-close --issue 244 --milestone M2 --verified '<actual evidence>'`, then the issue close gate. Fix blocking review findings before continuing.
- [ ] Publish through `sdlc pr` and `sdlc merge`, and verify archive/project links and repository state. No publication is claimed by this plan.

## Revisions

### 2026-09-23 — Explicit receipt operating bounds

Replaced deferred limit selection with concrete initial limits and recovery behavior so implementation has a testable storage envelope. These are conservative design assumptions, subject to measured revision.

### 2026-09-23 — Fresh-eyes review corrections

Separated confirmed nonapplication from unknown push outcomes; specified explicit-integration recovery and safe allocation retry. Distinguished independent-clone CAS barriers from linked-worktree lock serialization. Added ownership-metadata preservation to generic body sync. These correct review findings without changing project scope.

### 2026-09-23 — Plan-quality gate refinements (PQ-1 through PQ-3)

The following concrete contracts refine the corresponding sections above. They do not widen the approved scope.

#### Persisted formats and compatibility (PQ-2)

Authoritative issue frontmatter gains optional `claim_id: "<32 lowercase hex characters>"`, generated from 16 cryptographically random bytes and validated in `construct/vocabulary/issue.cue`. Absence means legacy/unclaimed ownership, not an empty token. Null, empty, malformed or duplicate fields refuse a claim/publication operation. Existing status vocabulary is unchanged. Open issues acquire a fresh ID on claim; matching local receipt plus matching remote ID permits an active claim retry. A terminal issue cannot be started by retrying claim. Legacy active issues keep their existing record and can use body sync, but claim refuses to invent an owner. No bulk migration occurs. Older binaries do not enforce the new contract; documentation requires updated SDLC for all concurrent writers. This is cooperative workflow coordination, not authorization against malicious Git writers.

Resolve the private Git directory with Git (`git rev-parse --absolute-git-dir`) from the verified repository root; never use `--git-common-dir` for receipts. Thus a linked worktree stores under its own `.git/worktrees/<name>/`, and an ordinary clone under its `.git/`. Store records at `<privateGitDir>/sdlc/publications/v1/<targetDigest>/<operationID>.json`. targetDigest is SHA-256 of a length-framed tuple of resolved remote URL, full target branch ref, and canonical repo-relative issues/history namespaces. Remote URLs are hashed in memory, never persisted or echoed. operationID is an independent random 32-hex identifier. Scanning is bounded by the receipt limits, with exactly one latest confirmed baseline for each issue/path set; new confirmed receipts retire superseded ones for that identity. Ownership continuity is copied only within the same private receipt store, never inferred from a copied branch.

Version 1 JSON (field names are the contract; `bytes` is standard JSON base64 encoding of byte arrays):

```json
{
  "version": 1,
  "target": "<64-hex digest>",
  "operation_id": "<32-hex ID>",
  "kind": "create|claim|update",
  "phase": "pending|rejected|confirmed",
  "issue_id": 244,
  "claim_id": "<32-hex ID or omitted for unowned legacy>",
  "baseline_commit": "<Git OID>",
  "conflict_commit": "<Git OID; rejected conflict only>",
  "paths": [{"path": "workshop/issues/000244-example.md", "expected": {"present": true, "bytes": "..."}, "desired": {"present": true, "bytes": "..."}}],
  "attempts": [{"base": "<Git OID>", "commit": "<Git OID>", "outcome": "unconfirmed|rejected|confirmed"}]
}
```

Use typed discriminants and validate the legal combinations before constructing the pure state. Absent path states carry no bytes; empty present files are distinct. Paths must be unique, canonical relative paths under configured issue namespaces; reject traversal, symlink escapes, duplicate JSON keys, unknown keys/versions, invalid Git OID lengths, and records outside byte/count bounds. JSON decoding is size-limited before allocation. Attempts are capped at three per operation; unresolved attempts prevent replacement. The unique operation ID is also a `Publication-Intent:` trailer on each constructed commit, preventing identical create/claim payloads from producing indistinguishable commit identities.

Write mode 0600 records in mode 0700 directories through same-directory temporary files, fsync the file, rename, and sync the directory. On platforms where directory sync is unsupported, return an explicit durability error rather than acknowledging persistence. One local repository transaction owns store mutations. Existing worktree-private receipts survive in-place branch switches; a new linked worktree does not inherit ownership. Partial temp files are ignored as authority and removed only when no live local operation holds the repository lock. Recovery compares typed target/identity and fresh remote evidence before any cleanup.

#### Named functions and adversarial strategies (PQ-1)

These are planned implementation entry points, not claims that the functions already exist. Tests exercise decisions through their production callers as well as directly.

| Function / existing caller | Test entry point | Adversarial strategy and oracle |
|---|---|---|
| `stepPublication` pure transition | `TestPublicationTransitionSequences` | Enumerate bounded sequences of conflict, rejected/unknown push, restart and confirmation; assert ownership uniqueness, no replacement without expected-state equality, and no cleanup while unknown. |
| `decideRecordUpdate` / `syncViaTrunkWithRealloc` | `TestPublicationConflictingRecords` | Mutate expected/remote/desired independently, including presence, deletion and rename pairs; changed same-record input refuses while unrelated records survive. |
| `parsePublicationReceipt` | `TestPublicationReceiptRejectsMalformed` | Table/fuzz over truncation, duplicate/unknown fields, invalid IDs, traversal and oversized input; every malformed record refuses without IO effects. |
| `writePublicationReceipt` / receipt loader | `TestPublicationReceiptInterruptedWrite` | Inject failures before fsync, rename and directory sync; restart reads either complete prior/new state, never partial authority. |
| `TrunkFile.UpdateMany` / conditional push adapter | `TestPublicationRemoteInterleavings` | Stateful remote fake schedules unrelated write, same-record write, rewind and lost acknowledgment; exact-old-ref mismatch never overwrites and unknown outcome retains intent. Real bare Git conformance repeats each supported push outcome. |
| `runClaim` / shared publishing dispatch | `TestPublicationClaimContenders` | Clones barrier after observation, linked worktrees barrier before lock; equal timestamps/content still yield exactly one owner. Repeat winner succeeds; copied branch loses. |
| `runIssueNew` / `decideCollision` | `TestPublicationAllocationContenders` | Same/different slugs and bytes at one candidate ID; rejection followed by reallocation produces two unique reservations, with no lost originals. |
| `runIssueSync` / baseline selection | `TestPublicationConflictRecovery` | Publish twice without moving local HEAD, inject third-party edit, integrate explicitly and retry; unresolved intent cannot bypass conflict and reconciled intent cannot stay stuck. |
| `syncIssuesToMain` | `TestPublicationMainScope` | Stage/commit unrelated code before issue publication; remote contains only selected issue updates and local code/index remain untouched. |
| `comparePreparedReview` | `TestPreparedReviewReadSet` | Change each input separately, including absent→present, branch-only switch, ledger-only change, code and docs descendants; verify gate-specific acceptance and no authoritative writes on stale input. |
| `stepPreparedReview` / manual command wrappers | `TestReviewConcurrencySchedules` | Barrier-controlled reviewer pauses, cancellation, failed relock and competing finalization; unrelated work completes, stale/cancelled results cannot advance ledgers or branches, child is waited on before normal return. |
| actual nested-clone CLI workflow | `TestPublicationDependencyEnvironment` | Drive creation/claim/local-sync/explicit-push from private dependency cwd with unreachable-remote and conflicting-clone episodes; assert normal namespace and preservation of local edits. |

#### Review transition model (PQ-3)

`stepPreparedReview` owns the phase and emits effects; command shells cannot persist result state before it reaches validated. PreparedReview carries immutable input bytes, branch/worktree identity, review anchor and ledger generation. No durable in-flight review record is added.

| State/event | Next state and effects |
|---|---|
| Idle + start | acquire local lock; capture prepared input set |
| Prepared + dispatch | release lock; run one reviewer with command context |
| Reviewing + completed response | reviewer process is waited/reaped; reacquire local lock |
| Reviewing + cancellation/dispatch error | cancel reviewer and wait/reap before normal return; no result persistence |
| Response + relock failure/cancellation | refuse; no sidecar, ledger, status, flow or branch write |
| Response + reacquired lock, changed read set/identity | stale refusal for every verdict; print output, no authoritative persistence |
| Response + reacquired lock, unchanged read set | validated; parse/persist existing gate result under lock |
| Validated + rejected verdict | retain valid review findings using existing ledger rules; no implementation/close transition |
| Validated + accepted verdict | continue deterministic gates/finalization while locked; another external wait requires another prepared snapshot |
| Any + parent process death | no later parent finalization exists; existing dead-holder recovery reclaims local lock; rerun starts a fresh review from durable ledger state |

Use the Cobra command context for external dispatch instead of `context.Background()`. Bound each external review with a derived timeout context: default 30 minutes, configurable by `WF_REVIEW_TIMEOUT` as a positive Go duration from 1 second through 2 hours; invalid values refuse before dispatch. This is a conservative operating assumption, not a latency promise. Timeout follows cancellation: signal the owned process group, allow at most 5 seconds to exit, then kill and wait; cap pipe draining with `exec.Cmd.WaitDelay` so inherited pipe descriptors cannot hang return. Inject the timeout/clock in tests rather than waiting in real time. No timeout produces a passing ledger/cache entry, and retry starts a new review against fresh inputs. Normal cancellation/error paths must reap the direct reviewer and terminate its owned process group where supported; do not orphan a reviewer on a normal function return. An uncatchable parent kill cannot promise portable child reaping: review children have no authority to persist gate results, and the next command must not consume their output as a completed round. Existing immutable reviewed commit plus read-set checks remain mandatory after restart. No asynchronous goroutine may mutate gate state after its command returns. Barrier-driven tests inject every interruption above through the production shell; same-boundary concurrent results invalidate on ledger generation before persistence.

### 2026-09-23 — Plan gate round 2 corrections

Replaced duplicate task-level scenario inventories with references to the named test/function strategy table. Added a concrete per-review timeout, bounded shutdown and pipe-drain policy; retained the explicit uncatchable-parent-death limitation. This addresses remaining PQ-1/PQ-3 findings.

### 2026-09-23 — Integration contract details at implementation entry

The gate passed; preserve the approved transaction semantics with these concrete seam refinements:

- A reallocated creation uses a new immutable operation ID and optional `predecessor` (32-hex operation ID) in its receipt. Keep rejected predecessors until successor confirmation and exact-content candidate cleanup. An initial occupied-ID observation can persist an explicitly rejected, never-dispatched predecessor to preserve the original local filename through crashes. Load/use checks same-target creation lineage, confirmed nonapplication, bounded depth and absence of cycles. Never delete edited candidate content. Git `TrunkWrite.Message` may override the transaction's default subject to bind each candidate commit to its current Publication-Intent trailer.
- Unfiltered publication remains one atomic multi-issue update: `issue_id: 0` is permitted only for kind update with no local claim_id. Select baselines per path using confirmed commit ancestry; divergent ambiguous evidence refuses. Confirmed receipt retirement requires complete path-set supersession and preservation of any locally owned claim evidence. A broad body update never acquires claim ownership merely by preserving authoritative claim_id bytes. Keep local ownership provenance separate from content baselines.

These implement the existing recovery, atomicity and no-transfer contracts; they add no workflow or ownership transfer mechanism.


## Superseded estimate at scope change

## Estimate

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only.*

Revised after estimate-quality feedback: separate concrete concerns instead of treating all receipt behavior, all callers or all fixtures as one module. Each implementation primitive includes its own unit tests; the separately itemized integration harnesses exercise cross-component processes and protocol conformance, not those unit tests. Gate reviews are three real boundaries. The calibration source is marked stale (#127), so the result is provisional focused ship-hours.

Derivation: greenfield concerns use base design 1.5 × 0.2 detailed-spec discount = 0.30 and base impl 0.8 × 0.4 AI-paired scale = 0.32. Standard-library JSON/filesystem primitives halve receipt IO/decoding design to 0.15; domain ownership/recovery/lifecycle have no library substitute. Each API integration uses base design 2.0 × 0.2 = 0.40 and base impl 1.5 × 0.4 = 0.60. These are behavioral integrations, not mechanical refactors. Smaller module uses 0.3 × 0.2 design and 0.5 × 0.4 impl. Each docs group uses 0.10 design and 0.20 × 0.4 impl. Reviews use 0 design and 0.5 × 0.4 impl. Issue authoring uses 1.0 design without double-discounting the authoring work itself, and 0.3 × 0.4 impl. Familiarity remains 1.0 (existing Go/Git stack); thorough-plan design buffer is 15%.

| Primitive | Concrete concern | Design | AI-paired implementation |
|---|---|---|---|
| issue-spec | Audit and durable design | 1.00 | 0.12 |
| greenfield-go-module | Publication transition model | 0.30 | 0.32 |
| greenfield-go-module | Ownership provenance and metadata | 0.30 | 0.32 |
| greenfield-go-module | Record preconditions and reconciliation policy | 0.30 | 0.32 |
| greenfield-go-module | Atomic receipt IO | 0.15 | 0.32 |
| greenfield-go-module | Strict receipt schema and decoding | 0.15 | 0.32 |
| greenfield-go-module | Crash recovery and predecessor lineage | 0.30 | 0.32 |
| greenfield-go-module | Receipt bounds and retirement | 0.30 | 0.32 |
| api-integration | Git evidence hooks and conditional push | 0.40 | 0.60 |
| api-integration | Unknown push outcome reconciliation | 0.40 | 0.60 |
| api-integration | Claim caller integration | 0.40 | 0.60 |
| api-integration | Create and reallocation integration | 0.40 | 0.60 |
| api-integration | Sync baseline and conflict integration | 0.40 | 0.60 |
| greenfield-go-module | Prepared review transition model | 0.30 | 0.32 |
| api-integration | Plan review prepare/finalize integration | 0.40 | 0.60 |
| api-integration | Estimate review prepare/finalize integration | 0.40 | 0.60 |
| api-integration | Close and milestone review record integration | 0.40 | 0.60 |
| greenfield-go-module | Reviewer timeout/process lifetime shell | 0.30 | 0.32 |
| api-integration | Stateful Git protocol fake and conformance | 0.40 | 0.60 |
| api-integration | Independent-clone and worktree publication harness | 0.40 | 0.60 |
| api-integration | Reviewer subprocess barrier harness | 0.40 | 0.60 |
| api-integration | Nested dependency CLI conformance | 0.40 | 0.60 |
| smaller-go-module | Vocabulary and generated contract compatibility | 0.06 | 0.20 |
| atlas-docs | Claim and issue command help | 0.10 | 0.08 |
| atlas-docs | Issue sync and lifecycle atlas | 0.10 | 0.08 |
| atlas-docs | Review/gate atlas and command help | 0.10 | 0.08 |
| atlas-docs | README and project checkpoint | 0.10 | 0.08 |
| milestone-review | M1 boundary | 0.00 | 0.20 |
| milestone-review | M2 boundary | 0.00 | 0.20 |
| milestone-review | Issue close boundary | 0.00 | 0.20 |

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec design=1.00 impl=0.12
item: greenfield-go-module design=0.30 impl=0.32
item: greenfield-go-module design=0.30 impl=0.32
item: greenfield-go-module design=0.30 impl=0.32
item: greenfield-go-module design=0.15 impl=0.32
item: greenfield-go-module design=0.15 impl=0.32
item: greenfield-go-module design=0.30 impl=0.32
item: greenfield-go-module design=0.30 impl=0.32
item: api-integration design=0.40 impl=0.60
item: api-integration design=0.40 impl=0.60
item: api-integration design=0.40 impl=0.60
item: api-integration design=0.40 impl=0.60
item: api-integration design=0.40 impl=0.60
item: greenfield-go-module design=0.30 impl=0.32
item: api-integration design=0.40 impl=0.60
item: api-integration design=0.40 impl=0.60
item: api-integration design=0.40 impl=0.60
item: greenfield-go-module design=0.30 impl=0.32
item: api-integration design=0.40 impl=0.60
item: api-integration design=0.40 impl=0.60
item: api-integration design=0.40 impl=0.60
item: api-integration design=0.40 impl=0.60
item: smaller-go-module design=0.06 impl=0.20
item: atlas-docs design=0.10 impl=0.08
item: atlas-docs design=0.10 impl=0.08
item: atlas-docs design=0.10 impl=0.08
item: atlas-docs design=0.10 impl=0.08
item: milestone-review design=0.00 impl=0.20
item: milestone-review design=0.00 impl=0.20
item: milestone-review design=0.00 impl=0.20
design-buffer: 0.15
total: 21.28
```

Design 8.66 × 1.15 + implementation 11.32 = 21.28 hours (rounded).
