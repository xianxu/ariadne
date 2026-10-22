---
collaboration-mode: discussion
---

# Review dependency checkouts implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3 for execution choices. Use superpowers-executing-plans for the session-warm orchestration; bounded pure implementation or exploration may be delegated.

**Goal:** A consumer's boundary reviewer can inspect its declared dependencies without finding or modifying live peer checkouts.

**Architecture:** Read the reviewed commit's direct construct/deps declaration, resolve its local sources through the same rules as #545, and pin each source's refs/remotes/origin/main once. Create disposable sibling clones under the existing reviewbox lifecycle; list their paths, immutable SHAs, freshness evidence and isolation mode in the review prompt. An available OS fence protects them; without it, mutations affect only disposable copies and retries start pristine.

**Tech Stack:** Go, existing layergraph parser/resolver, reviewbox, gitx bounded process runner, Seatbelt/bubblewrap, real disposable Git fixtures.

## Core concepts

| Pure entity | Lives in | Status |
|---|---|---|
| Dependency bindings | cmd/sdlc/localcheck_dependency_environment.go | modified: shared resolved-row projection, retaining #545 conflict/deduplication semantics |
| ReviewDependency descriptor and prompt projection | cmd/sdlc/internal/reviewbox/dependencies.go | new: kind, disposable path, source identity, pinned SHA and fetch observation |
| Dependency tree budget | cmd/sdlc/internal/reviewbox/dependencies.go | new: bounded tree metadata validation before materialization |
| Review isolation | cmd/sdlc/internal/reviewbox/fence.go | modified: carries dependency descriptors and protects clone/source paths |

| Integration point | Lives in | Status | Wraps |
|---|---|---|---|
| Resolved dependency rows | cmd/sdlc/localcheck_dependency_environment.go | modified | shared layergraph declaration reader/resolver; filesystem |
| Review dependency preparation | cmd/sdlc/reviewdependencies.go | new | local Git refs, FETCH_HEAD observation, reviewbox child clone creation |
| Checkout ownership | cmd/sdlc/internal/reviewbox/checkout.go | modified | existing Create/Reset/Close/Sweep and flock lifetime |
| Boundary setup and prompt | cmd/sdlc/reviewisolation.go; cmd/sdlc/milestoneclose.go | modified | consumer close, milestone-close and dry-run entry points |

Descriptors have no persisted schema: they exist only for one review. Pure policy tests use values; IO tests use the existing real-Git testfix repository as a stateful portable dependency backend, plus existing judge.Run injection for reviewer output. No new network service or acquisition mechanism.

## Decisions and alternatives

1. Reuse Create for separately locked sibling clones, retained by the primary Checkout. A free root/ariadne directory is unsafe because Sweep would remove it while another review runs. A new bundle-directory layout could use one lock, but would change existing checkout paths and consumers unnecessarily (ARCH-DRY, ARCH-FUNERAL).
2. Use only direct declared rows, matching #545's SDLC_DEP_SUBSTRATE and SDLC_DEP_DATA naming, deduplication and same-kind conflict refusal. No transitive exports or new revision grammar. Substrate resolution reuses layergraph's primary fallback; data resolves its mount against the initiating checkout. Never interpret the source URL as a command or fetch it.
3. Pin local refs/remotes/origin/main, never a dirty worktree, local feature HEAD or a guessed local main. Missing/unreadable sources or that ref fail setup with a named dependency and recovery. Operator-selected source revisions/pin syntax are outside this issue; the current grammar has none.
4. Preserve checkout-only fallback, per TL approval. A fenced review gets protected dependency clones and source worktrees/common Git dirs; otherwise the prompt explicitly says disposable, writable, discarded. Do not claim read-only enforcement without an OS fence.
5. Print exact SHA and fetch-age evidence. FETCH_HEAD mtime alone is not proof when another remote/branch was fetched. Report an observed matching origin/main fetch entry and its timestamp/age only when it can be tied to that pinned SHA; otherwise say fetch age unknown and why. A commit author date or ref-change date is not a fetch timestamp. No fetch is performed.

## State, actions and recovery

Actions: repair local dependencies and rerun the boundary; allow normal automatic retry; rerun after interrupted setup.

| Internal state / failure | Visible state | Recovery action | Possible loss |
|---|---|---|---|
| No declaration | Existing review behavior, no dependency block | Normal review | None |
| Valid local sources and pinned commits | Prompt lists SDLC_DEP_* paths, SHAs, age evidence and isolation mode | Normal review | Disposable edits only |
| Malformed/conflicting declaration, missing source/ref, size budget exceeded | Setup fails naming dependency and cause; reviewer does not launch | Repair dependency/declaration, rerun boundary | No accepted review |
| Partial clone or cancellation | Setup fails; all created clones closed | Rerun boundary | Disposable preparation only |
| Missing verdict and retry | Same dependency SHAs, pristine clone contents | Automatic existing retry | First attempt's disposable writes |
| Reviewer exits, fails, times out or is cancelled | Existing boundary outcome; primary and children removed | Follow existing boundary recovery | Disposable writes |
| Owner process dies | Kernel releases all clone locks; next Sweep reclaims them | Rerun boundary | Disposable writes |
| FETCH_HEAD cannot prove age | Prompt says unknown age; review continues at exact pin | Reviewer discounts staleness; owner may refresh outside review | No fabricated freshness |

Child handles attach to the primary owner immediately after creation. Every error unwinds already-created children. Their individual locks prevent another sweep from collecting live siblings; Close and Reset own every child and remain idempotent. Retry keeps the original descriptor/pin rather than resampling origin/main. No background child outlives preparation or dispatch (ARCH-ORDER).

## Operating envelope and trust

This is a local batch operation per review, not a per-event hot path. Declarations retain layergraph's existing byte bound; cap parsed direct rows at 64 before resolution. Same-kind conflicts cap distinct exported clones at two. Parse each row once, resolve each unique source once, create each unique clone once; counted duplicate-row fixtures at 8 and 16 assert linear resolution and one clone per distinct source (ARCH-CONSTRAINTS).

Proposed initial safety ceilings: 100,000 tracked entries and 512 MiB summed blob bytes per dependency, 1 GiB total materialized dependency blobs; 2 minutes total dependency preparation plus the existing short process-group shutdown grace. These are explicit initial safety limits, not measured latency promises. Preflight ls-tree metadata with bounded output before checkout, rejecting unsupported gitlinks rather than silently presenting an incomplete dependency. Use shared local object storage, with private refs/config/index and inert remotes as existing Create does. No history copy or dependency acquisition.

Use the existing gitx contextual process runner for dependency commands so cancellation owns Git descendants. Disable inherited global/system Git checkout filters, hooks and templates for new dependency clones; pin literal commit IDs and do not honor replacement refs. Preserve normal tracked file modes/symlinks, but never copy live untracked or modified source files. Paths originate in the bounded declaration/resolver, prompt rendering quotes them as data, and fence paths are canonical. No remote URLs or credentials need appear in the prompt. The reviewed repository remains writable for review tests; dependency source worktrees/common dirs are protected when a fence exists (ARCH-SECURE).

The clone's files and lock are owned by existing Create/Close/Sweep. No new durable cache, lease format or tracker field is introduced (ARCH-EVOLVE). Preparation uses a single deadline; reset rebuilds at the same pins with the same limits. Large dependencies fail before reviewer launch rather than unexpectedly filling the RAM disk. Infrastructure errors keep their original cause.

## Implementation and verification

### 1. Share declaration resolution; define descriptors and pure rendering

Files: cmd/sdlc/localcheck_dependency_environment.go and its tests; new cmd/sdlc/reviewdependencies.go and reviewdependencies_test.go; new reviewbox/dependencies.go and dependencies_test.go.

- [ ] Add red-first tests: a pinned consumer declaration wins over dirty caller edits; substrate primary fallback and data mount resolve exactly as #545; equal paths dedupe and conflicts refuse; missing/malformed declaration controls remain unchanged.
- [ ] Extract the minimal resolved-row helper and preserve local-check environment behavior byte-for-byte.
- [ ] Add pure tests for kind/path/SHA/freshness/mode rendering, path quoting, row/count and tree-byte limits, malformed metadata and unknown age. Reject over-bound data before any clone.
- [ ] Implement descriptors and tree metadata policy without filesystem IO in the pure functions (ARCH-PURE).
- [ ] Run targeted dependency-environment and new descriptor tests; commit.

### 2. Attach bounded disposable dependency clones to review lifetime

Files: reviewbox/checkout.go, dependencies.go and their tests; cmd/sdlc/reviewdependencies.go and tests; reviewisolation.go and reviewisolation_test.go.

- [ ] Red-first real-Git fixture: consumer depends on a source with distinct origin/main, local HEAD and dirty bytes. Assert only origin/main content appears, exact SHA is retained, source snapshots are unchanged and no network command runs.
- [ ] Red-first lifecycle matrix: successful Close, partial second-clone failure, cancellation, automatic Reset after deliberately changed dependency contents, live Sweep preservation and dead-owner Sweep reclamation. Assert all dependency handles release, including cleanup errors.
- [ ] Implement bounded preflight and local clones through existing Create; add ownership for child handles and clean partial failure. Keep same absolute sibling paths across retries.
- [ ] Expand available fence protection to dependency clones plus source worktrees/common dirs. Preserve no-fence admission and label its disposable mode explicitly.
- [ ] Test FETCH_HEAD matches/mismatches, unavailable/corrupt/future timestamp and pin movement between attempts; never invent an age or refresh a pin.
- [ ] Run go test ./cmd/sdlc/internal/reviewbox -count=1 and targeted cmd/sdlc dependency/isolation tests; commit.

### 3. Supply dependency context to every boundary review and document it

Files: cmd/sdlc/milestoneclose.go; reviewisolation.go; reviewisolation_test.go; reviewpreparation_live_test.go; atlas/workflow/sdlc-binary.md and setup-and-replication.md.

- [ ] Red-first production-boundary tests capture the dispatched prompt for both milestone and issue-close recipes. A reviewer can read the actual dependency fixture/template at the named SDLC_DEP_SUBSTRATE path; no-deps prompts/behavior remain unchanged. Include substrate and data names, duplicate/conflict controls and dry-run behavior (no clones created).
- [ ] Thread descriptors through Isolation and append a deterministic dependency manifest to the existing review-window prompt. Names are documented path bindings for inspection; no live caller path is advertised as a review dependency. Include exact revision, freshness evidence and read-only-fenced versus disposable-write mode.
- [ ] Exercise Claude and Codex argv wrappers and preparation/retry entry points. Under real Seatbelt or bubblewrap, verify reads succeed and file/ref writes to clones and sources fail. Under FenceNone, verify writes affect only disposable clones and reset removes them.
- [ ] Add atlas documentation of local-last-fetched pins, unknown fetch age, modes, bounds, failure recovery and lifecycle. Record relevant lesson only if review finds one; no lexical lint can establish this ownership/fence contract.
- [ ] Run targeted tests once green. Checkpoint/push; run recorded suite via sdlc local-check run --name suite; close through the mandatory SDLC review; land on SHIP.

## Plan conformance table

| Risky production function | Adversarial strategy | Independent guard |
|---|---|---|
| resolvedDependencyRows | Declaration differs between reviewed head and dirty caller; two aliases/conflicting rows | #545 environment regressions and counted unique-resolution fixture |
| prepareReviewDependencies | Wrong local branch, stale/missing main, unrelated FETCH_HEAD, oversized tree, cancellation | Real source/clone bytes and ref snapshot; no network command sentinel |
| Checkout.Reset/Close/Sweep | Mutation, partial create, retry, owner death and concurrent sweep | Lock ownership plus absence/pristine-content assertions through real Git |
| openReviewIsolation | Source alternate path and dependency clone writes, fenced and unfenced | Kernel write denial where available; unchanged real source snapshots otherwise |
| boundaryReviewOptionsFor | Prompt omits or misnames dependency; alternate recipe/agent/preparation | Production-dispatch capture and actual file read using printed paths |

## Live conformance proposal

Build the branch binary. Ask ops-echo:0 to run its next boundary review from Echo using that binary and report the dependency manifest: ariadne path, exact origin/main SHA, age evidence, mode, and successful fixture/template inspection. This runs the ordinary disposable reviewer lifecycle, with no forced outage or live dependency write. Hub decides whether that owner-run check is required pre-landing; deterministic real-Git/fence fixtures provide the reproducible proof.
