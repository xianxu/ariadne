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
| DependencyEnvironment | pkg/layergraph/environment.go | new shared owner extracted from #545; main wrapper retains existing callers |
| Dependency and DependencyManifest | cmd/sdlc/internal/reviewbox/dependencies.go | new: kind, disposable path, source identity, pinned SHA and fetch observation |
| DependencyTreeBudget | cmd/sdlc/internal/reviewbox/dependencies.go | new: bounded tree metadata validation before materialization |
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

## Deadline and environment integration

Add reviewbox.CreateContext(ctx, ...) and Checkout.ResetContext(ctx); require explicit caller contexts at every checkout and confinement entry point; remove unused context-free wrappers. Pass context through confinement as well as population, authority copying and every Git effect through gitx.RunGitInContext (bounded metadata output where applicable). The main boundary uses openReviewIsolationContext(p.Context, ...), with its existing wrapper only for compatibility callers. Initial dependency setup gets one two-minute child context capped by the caller; each retry gets a fresh two-minute preparation context capped by Dispatch's remaining deadline. Add DispatchOptions.BeforeRetryContext func(context.Context) error, prefer it over the compatibility BeforeRetry callback, and supply the actual dispatch context at the retry call site. BeforeRetryContext calls ResetContext; cancellation cleans all children and never launches the next reviewer. A blocked-Git descendant regression exercises production setup and retry, asserting process exit and released clone locks.

Carry disposable bindings through DispatchOptions.DependencyEnvironment and the existing Confinement environment construction, not inherited caller variables. Use #545's reserved-key clearing and deterministic projection; add these values to Claude's generated shell env file as well as the reviewer/preparation process environment. Prompt and environment derive from the same descriptor list. A reviewer-run shell command must read the actual pinned fixture using its exported path; no test relies on a preexisting shared sibling.

## Implementation and verification

### 1. Share resolution and define descriptors

Files: cmd/sdlc/localcheck_dependency_environment.go and its tests; new cmd/sdlc/reviewdependencies.go and reviewdependencies_test.go; new reviewbox/dependencies.go and dependencies_test.go.

- [x] Write failing tests using the resolution/descriptor strategies in the conformance table; run them and retain the red evidence.
- [x] Extract resolvedDependencyRows, preserving #545 environment behavior; implement pure descriptor rendering and metadata budgets.
- [x] Run targeted dependency-environment and descriptor tests to green; commit.

### 2. Own bounded dependency clones through setup and retry

Files: reviewbox/checkout.go, dependencies.go and their tests; cmd/sdlc/reviewdependencies.go, reviewisolation.go and tests; judge/dispatch.go and retry tests.

- [x] Write failing ownership, pinning and cancellation regressions using the table's strategies; verify their red failures.
- [x] Implement contextual checkout creation/reset, child ownership, bounded preparation and protected paths under the approved modes.
- [x] Run go test ./cmd/sdlc/internal/reviewbox -count=1 and targeted dependency/isolation/retry tests to green; commit.

### 3. Expose dependency context and verify the production boundary

Files: cmd/sdlc/milestoneclose.go; reviewisolation.go; reviewisolation_test.go; reviewpreparation_live_test.go; judge/dispatch.go; reviewbox/confine.go; atlas/workflow/sdlc-binary.md and setup-and-replication.md.

- [x] Write failing production-dispatch prompt and reviewer-command environment regressions using the table; verify red before threading descriptors.
- [x] Render dependency paths, pins, freshness evidence and mode from the shared descriptor list; project its bindings into confinement and Claude shell restoration.
- [x] Run targeted tests to green and the available live OS-fence conformance fixture. Document modes, local-last-fetched pins, bounds and recovery.
- [ ] Checkpoint/push; run recorded suite via sdlc local-check run --name suite; obtain the Echo-owned live check; close through the mandatory SDLC review and land on SHIP.

## Plan conformance table

| Risky production function | Adversarial strategy | Independent guard |
|---|---|---|
| resolvedDependencyRows | Declaration differs between reviewed head and dirty caller; two aliases/conflicting rows | #545 environment regressions and counted unique-resolution fixture |
| prepareReviewDependencies and tree budget | Adversarial local source/ref/freshness and tree metadata | Real pinned clone bytes and source snapshot; bounded-output/budget guard and no-network sentinel |
| CreateContext/ResetContext/Close/Sweep | Interrupt or mutate lifecycle at effect boundaries | Blocked-Git descendant termination, lock release and pristine-content/absence assertions |
| openReviewIsolation | Source alternate path and dependency clone writes, fenced and unfenced | Kernel write denial where available; unchanged real source snapshots otherwise |
| boundaryReviewOptionsFor and Confinement.Env | Prompt/environment drift across recipe, retry and agent shell | Production-dispatch capture plus reviewer-run command reading the actual fixture from its exported path |

## Live conformance proposal

Build the branch binary. Ask ops-echo:0 to run its next boundary review from Echo using that binary and report the dependency manifest: ariadne path, exact origin/main SHA, age evidence, mode, and successful fixture/template inspection. This runs the ordinary disposable reviewer lifecycle, with no forced outage or live dependency write. Hub decides whether that owner-run check is required pre-landing; deterministic real-Git/fence fixtures provide the reproducible proof.

## Revisions

### 2026-10-22 — plan-quality round 1 and environment conformance

PQ-1: name the contextual Create/Reset and BeforeRetryContext seams explicitly; each retry receives a fresh bounded preparation child capped by the remaining dispatch deadline, and all Git descendants terminate before cleanup. PQ-2: remove repeated concrete test inventories, retaining the function-level adversarial strategy table and red/green execution steps; detailed cases belong in tests. PQ-3: align the issue's current acceptance criteria with the approved two isolation modes. TL also approved exporting disposable SDLC_DEP_* through confinement and Claude shell restoration, with the reviewer-command regression. Echo confirmed #10's readers all honor that binding, making its held boundary review the live proof.

### 2026-10-22 — implementation traceability and no-fetch conformance

The shared pure environment projection is extracted to pkg/layergraph/environment.go so both the existing main wrapper and judge/reviewbox consume one owner. Runtime descriptors travel as DispatchOptions.Dependencies (rather than a second environment-string representation); Confinement projects them once with that shared owner. Observational Git and dependency materialization suppress lazy fetching; observational commands deny every transport, while only the explicit local clone permits file transport. A real missing-promisor-object regression exposed implicit Git fetches and source-config writes before this guard, and passes afterward. These complete the approved no-fetch and naming contracts without new persisted state or acquisition.

### 2026-10-22 — README completion from boundary review

BR-1 adds the user-facing dependency/preparation contract to README alongside its existing boundary setup instructions, linking the detailed atlas. No runtime behavior or interface changed. Keep the in-flight recorded suite and original clean binary; close round 2 disposes this documentation finding before publication.

### 2026-10-22 — live shell failure and full-suite context contract

The first Echo live review exposed Bash-only parameter expansion in Claude restoration while the actual command shell was zsh. A real Bash/zsh/POSIX-sh matrix reproduced that failure, then verified the portable names-only scrub, including hostile and nonexported values and retry restoration. The exact peer-supplied zsh diagnostic is retained as a fixture. Also print the existing descriptor manifest in close output, using the same projection as the reviewer prompt, so owners have direct live-check evidence.

The full suite's unchanged context guard rejects the proposed compatibility wrappers that create Background contexts. Remove these unused internal production entry points, update their test callers to contextual APIs, and thread Dispatch's context through confinement and dry-run protection reads. Do not relax the guard. This corrects the already-approved cancellation contract without adding state or changing the feature's scope. The new runtime fixes require a fresh recorded suite and a narrow post-close review before landing.
