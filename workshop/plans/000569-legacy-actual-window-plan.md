---
collaboration-mode: discussion
---

# Legacy disjointness authority re-plan

> **For agentic workers:** Follow AGENTS.md Section 3. This is retained-context work; SDLC owns the fresh-context review at close.

**Goal:** Keep the approved labelled disjointness fix without letting modern or unreadable source evidence become a legacy temporal shortcut.

**Architecture:** Preserve the shared immutable generation reader and strict completion wrapper. Replace the selected-close-only authority check with one bounded source-ledger history check pinned to the pre-cutover source main. No current or historical actual, completion, review verdict or persisted format changes.

**Tech stack:** Go, immutable Git reads through the existing ownedActualGit/archiveReadContext seams, gatestate.ParseSidecar, native temporary Git/tracker fixtures.

**Status:** Proposed at the third design delta; awaiting hub approval. No implementation of this re-plan has begun.

## Why re-plan

Round-two BR-1 proves the selected-close guard misses modern or unreadable ledgers saved at archival or in a later ledger-only commit before cutover. The failure class is source authority changing independently of issue details, not a missing special case for one timestamp.

The simplest alternative is to inspect the close, archive and source-main snapshots. That covers the review's eight examples but can miss authority added and removed between those snapshots. Recommend checking the complete bounded relevant ledger history instead. Do not add another purpose flag, persisted state, fallback or cache. A lifetime lock cannot prove absence in historical Git objects; immutable pins plus complete bounded history are the authority.

## Core concepts

### Pure entities

| Name | Lives in | Status |
| --- | --- | --- |
| Legacy source-ledger eligibility predicate | cmd/sdlc/legacycompletion.go | new |
| legacyWindowDisjoint | cmd/sdlc/actualexclusion.go | unchanged |

The eligibility predicate consumes a strictly parsed ledger for the matching issue and boundary-review gate. Any Reviewed marker or ActualInputs value disallows bound-only use; invalid structure is an error, never legacy absence. Existing pure strict-before comparison and warning semantics remain unchanged. One immutable source generation owns many ledger observations; no new reusable domain type is required.

### Integration points

| Name | Lives in | Status | Wraps |
| --- | --- | --- | --- |
| resolveLegacyGeneration archive-bound branch | cmd/sdlc/legacycompletion.go | modified | existing source and generation validation |
| Bounded source-ledger authority reader | cmd/sdlc/legacycompletion.go | new | immutable Git history and strict sidecar parsing |
| ownedLedgerAt | cmd/sdlc/actualinputs.go | reused | snapshot inventory and identity validation |
| loadAdoptedExclusions | cmd/sdlc/actualexclusion.go | unchanged | bound selection and visible label |

The new reader enumerates relevant issue boundary-ledger history once at the immutable source-main pin, covering configured active plans and canonical archived plans. Preserve deletions/renames and merge history rather than following only the current surviving path. Check close, terminal and source-main inventories for absent/unique/ambiguous paths using the existing snapshot policy; historical relevant blobs are decoded once per distinct immutable blob. Invalid modes, identity, ambiguous evidence, read failure, truncated history or overflow refuse the shortcut. This is read-only. Native real-Git fixtures are the stateful dependency model and exercise the same production Git seam; no GitHub dependency is added.

## Visible states

Actions: rerun measurement; repair the named evidence.

| Internal state / failure | Visible state | Recovery action | Possible loss |
| --- | --- | --- | --- |
| Exact old generation, complete classic ledger history, end strictly before earliest candidate | Measured with disjointness label | Continue normal close | None; old actual/review unchanged |
| Modern, malformed, ambiguous or incomplete evidence; equality or overlap | Cannot measure, named evidence | Repair the named evidence, then rerun | None; no new adopted hours |

The existing legacy-only captured protocol-error ledger remains eligible for temporal relevance, not review validity. The strict repair/show/replay path remains strict. No running work, lease, retained artifact family or new cleanup obligation is created.

## Operating envelope and guard

This is an on-demand measurement read. Growth drivers are relevant adopted issues A, bounded relevant ledger-history entries H, distinct ledger blobs B and plan inventory P at the three fixed snapshots. Expected work is O(A*(H+B+P)); never rescan the full plans inventory per historical commit. Reuse the existing complete-history ceiling of 10,000, refuse overflow before reading all objects, and use existing bounded Git output/sidecar parsing. No network, process/topology scan, new concurrency or durable cache. Tests at N and 2N count history walks, inventory reads and blob reads, including repeated identical blobs; timing alone is not the cost defense.

## Function-level adversarial strategies

| Boundary | Adversarial strategy | Independent guard |
| --- | --- | --- |
| New source-ledger reader | Change evidence independently of issue details, including add/remove, rename, deletion, merges and malformed snapshots | Native Git history matrix; no-skip/no-warning oracle and unchanged refs/status |
| Eligibility predicate / strict parser | Retained input, damaged input, reviewed-head-only, unreadable bytes, wrong identity, classic ledger | Table assertions; existing strict damaged-evidence controls |
| resolveLegacyGeneration | Keep issue generation valid while changing ledger authority | Public exclusion fixture and unchanged strict completion refusal |
| loadAdoptedExclusions | Candidate prefix earlier than modern ownership; strict equality and overlap | Existing real hybrid actual/output controls |
| Cost boundary | Double unrelated commits and relevant ledger changes separately | Counted N/2N calls, fixed inventory passes, unique blob reads and overflow-before-object-read assertion |

## Tasks — one atomic close boundary

- [ ] Add permanent RED controls for all four evidence forms at archive and later ledger-only write; include added-then-removed evidence and valid classic controls. Correct BR-2's fixture comment in its own branch.
- [ ] Implement the single bounded reader and shared eligibility predicate; remove the selected-close-only authority decision. Keep public policy and strict wrappers unchanged.
- [ ] Run focused native controls and cost/overflow tests normally and under race; run relevant lint. Freeze the entire checkout during cmd/sdlc tests.
- [ ] Update atlas/help/lessons only for delivered behavior; checkpoint clean code and rebuild a private CLI for read-only Pair proof. Expected result remains past151 and strict refusal on overlapping161.
- [ ] Push, run one central recorded suite on final code, obtain a fresh full close review, and enter normal central merge on SHIP. No staging or private lifecycle binary. Hub rebuilds primary after landing.

The independent no-retroactive-owner follow-up remains after #569, from fresh main. No readiness message to pair:2 before successful measurement. If this re-plan leaves the quick-flow shell, rerun change-code's inferred full-flow plan-quality/estimate gates before implementation; do not pin quick flow or waive them.

## Architecture lenses

ARCH-DRY: share generation validation, parser and snapshot identity policy. ARCH-PURE: eligibility and temporal comparison are pure; Git is the thin integration boundary. ARCH-PURPOSE: sweep independent ledger chronology, not only review examples. ARCH-MOCK: native private Git/tracker fixtures model persisted history. ARCH-CONSTRAINTS: bounded history and counted linear growth. ARCH-SECURE: untrusted historical bytes are strictly parsed and pins are immutable; no credentials. ARCH-ORDER: no state survives this read, but all historical evidence is considered before weaker authority is returned. ARCH-FUNERAL: creates no production artifact or process. ARCH-EVOLVE: changes no persisted format or replay meaning.

## Revisions

### 2026-10-23 — declined by TL; do not implement

Authenticated hub receipt74137932-2fd2-4698-8067-62ad44c54a91 declines this mechanism on cost/payoff grounds for a shrinking set of legacy-era issues. #569 ends wontfix; the first-modern-owner follow-up is dropped too. Preserve this proposal and existing work as evidence, without landing code. Affected legacy-crossing closes explicitly use --no-actual and verification citing 'legacy-era overlap; measurement unavailable by design (ariadne#569)'.
