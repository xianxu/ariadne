---
collaboration-mode: discussion
---

# Clear a landed stage holder implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3 for execution strategy. Use superpowers-executing-plans for this atomic authorization change; SDLC owns its one close review. No Mx subdivision.

**Goal:** Plain `sdlc stage --clear` from the original worker slot on rest succeeds after a confirmed landing deletes its issue branch.

**Architecture:** Keep the target's existing stage lease and lifetime lock. Extend clear authorization with one narrowly qualified landed-holder recovery path, using the current completion binding, exact merged PR and existing durable archive proof against the same freshly fetched main pin used by restoration. Absence of a branch or matching slot alone grants nothing.

**Tech Stack:** Go, workspace identity, stage lease/flock, tracker completion reader, existing landing Git/GitHub seams, private native Git repositories and stateful fake GitHub.

## Decision and scope

Recommend a proof-backed exception at clear's existing authorization boundary. A slot-only exception cannot distinguish a completed holder from unfinished work. Adding a new close token to the stage lease would not solve existing stages (staging precedes close) and introduces a format lifecycle unnecessarily. Reuse the completion/archive authority already used for post-landing cleanup (ARCH-DRY).

This is full flow: it adds an authorization route spanning identity, tracker, GitHub and archive consumers. It is one atomic change with one review boundary. No estimate before change-code's plan-quality acceptance.

Existing same-holder authorization and explicit --abandoned acknowledgment stay unchanged, including legacy records and non-issue branch holders. The new recovery route requires a current tracked done completion; repositories without that proof retain the existing explicit recovery path. No new legacy landing inference is introduced.

## Core concepts

| PURE entity/function | Lives in | Status |
| --- | --- | --- |
| `landedStageCaller` | `cmd/sdlc/stageplan.go` | New predicate over existing lease/caller facts |
| `checkClearLease`, `stageHolder` | `cmd/sdlc/stageplan.go` | Reused unchanged for ordinary authority and exact acknowledgment |

`landedStageCaller(held fleet.StageLease, caller moveSide)` admits a proof candidate only when the canonical original slot is present and matches, the caller is on its own nonempty resting branch, the lease holder is an issue matching the recorded issue branch, and its staged commit is valid. It does not grant clear authority or perform IO. Another issue branch in the original slot is not a candidate.

| INTEGRATION point | Lives in | Status | Wraps |
| --- | --- | --- | --- |
| `proveLandedStageHolder` | `cmd/sdlc/stagelanded.go` | New | Fresh current completion, exact merged PR, staged commit ancestry and archive proof |
| `runUnstage` | `cmd/sdlc/stage.go` | Modified | Candidate selection, one main pin, proof and final caller/lease checks under the target lock |
| `landedCommit`, `selectArchiveLandingIssues`, `confirmLandingArchive`, `completionCarried` | `cmd/sdlc/ghlanding.go`, `landingarchive.go`, `trackercompletion.go` | Reused unchanged | Exact immutable landing and archive generation |
| `fleet.LockStage`, `workspace.Resolve`, `fleet.ReadStageLease` | `cmd/sdlc/internal/fleet/staged.go`, `pkg/workspace/resolve.go` | Reused unchanged | Exclusive stage lifetime and canonical slot identity |

The proof uses the existing landingGH seam with the stateful fake and native Git, not a second GitHub protocol or call-count mock. No new persisted noun or format.

### Proof and ordering

1. Acquire the target's stage lock as today; re-observe target and lease there. Ordinary checkClearLease success keeps the existing route. A wrong explicit --abandoned still refuses, without falling through to landed recovery.
2. On ordinary holder mismatch, reject non-candidates before network/proof work. Require canonical same-repository identities as well as exact original-slot binding; caller must be resting, never detached or on another issue.
3. Retain current target cleanliness/leavability and asset checks. Fetch and pin main once through the existing #537 snapshot. Reuse that pin for authorization, lag, fast-forward and rebuild; no independent second main fetch.
4. Within the existing 30-second landing recovery budget, read the held issue's current tracked card freshly. Require done, a valid same-repository completion and durable landed commit confirmed on that pinned main. Missing/reopened/replaced or unreadable binding is not an issue-free success.
5. Read the recorded branch's PRs through landingGH. Match exact repository/branch, MERGED state and immutable head/base. Require the lease's staged commit to be carried by that PR's head. Use landedCommit for fresh-main reachability (normal, squash and train-car variants); do not equate raw mergeCommit with the train's integration.
6. Select archive-owned issues using the existing completionCarried policy. Explicitly require the lease's issue ID among the selected current done generation. Never use landingArchiveComplete's empty-selection success as authority. Require unambiguous matching landing evidence, then call the existing archive confirmation for that selected generation, including #512's generic done-card archive path.
7. Before the first checkout/lease removal/build, re-observe the caller and target lease under the same stage lock; a changed caller identity/branch, replaced lease or changed observed completion generation refuses. The lock serializes cooperative stage/move/clear writers for the whole operation. Remote policy is freshly observed proof, not permanent authority.
8. Enter the existing pinned-main restore/fast-forward/rebuild sequence. Both detached and already-resting retry paths use the same authorization. Dry-run observes/proves identically and changes no checkout/lease/build; preserve #537's explicitly unverified fetch-failure report.

Do not weaken archive helpers or their issue-free semantics for other consumers. Keep this exception clear-only: staging/moving from a resting caller does not inherit a landed issue's authority (ARCH-PURPOSE, ARCH-SECURE).

## Visible states and recovery

Actions: clear, dry-run clear, return to the original slot's rest, finish/reconcile landing, explicitly acknowledge abandoned holder, repair evidence/environment and retry.

| Internal state/failure | Visible state | Recovery action | Possible loss |
| --- | --- | --- | --- |
| Ordinary holder or exact abandoned acknowledgment | Existing clear behavior | Clear | Existing staged checkout is released by explicit action |
| Original slot on rest, current done binding and exact landing/archive proof | Confirmed landed-holder clear | Clear | Only the intended stage/lease ends; original work is landed |
| Another slot/issue, detached caller or non-issue/malformed recovery identity | Existing holder refusal | Return to original slot's rest or explicit abandoned acknowledgment | None |
| Working/codecomplete/reopened card, missing/replaced binding, unmerged/ambiguous PR or incomplete archive | Unverified landed-holder refusal naming the evidence | Finish/reconcile landing or repair evidence and retry | None; stage and lease retained |
| Probe failure/cancellation or caller/lease change before effect | Unverified/changed-state refusal | Repair environment/evidence and retry | None before effects |
| Dirty/ahead/diverged target or failed rebuild | Existing #537 refusal/retry | Existing named repair then clear | Existing partial restore/rebuild semantics, no reset |

No background work or automatic retrain/retry. The stage lock is the existing single resource preventing competing stage release/replacement; no observation loop claims exclusivity. GitHub/tracker interruption means unknown and preserves the lease. No new state survives the invocation (ARCH-ORDER, ARCH-FUNERAL, ARCH-EVOLVE).

## Verification strategies

| Risky function | Adversarial strategy | Mechanical guard |
| --- | --- | --- |
| `landedStageCaller` | Caller/holder identity combinations | Pure table tests of original canonical slot, rest and issue/branch consistency |
| `proveLandedStageHolder` | Unrelated/absent/reopened generation masquerading as issue-free proof | Real current tracker cards, immutable Git objects/archive transitions and stateful GH fake |
| `runUnstage` | Real landing removes caller branch, then clear must select fresh main | Native slot fixture captures HEAD, lease, build marker and exact built pin |
| `runUnstage` proof/effect boundary | Lock contention, lease replacement, cancellation and caller movement | Controlled ordering at existing IO seams; denied paths assert unchanged target and no build |

Executable tests own individual cases rather than duplicating an inventory in prose. The positive fixture stages from an issue branch, publishes a real completion, merges and archives through the production completion/archive paths, removes the branch and switches the original slot to rest, then clears without --abandoned. Cover train-car missing/raw mergeCommit variations via the existing fake seam and generic archive proof, as well as ordinary merged PR evidence. Negative fixtures independently corrupt each authority input; preserve explicit acknowledgment and ordinary stage/move controls.

This is an operator-started bounded command, not a hot loop. One fresh main fetch, one bounded PR listing (existing 100-result limit), one current tracker proof and existing 10,000-entry archive bounds; no polling/fan-out. Whole new proof is bounded by existing landingRecoveryTimeout (30s). Existing retained stage lease is the only residue and is removed by the existing successful clear path. No credentials or live service writes are added (ARCH-CONSTRAINTS, ARCH-MOCK).

## Task 1: Prove and implement authorization

**Files:** `cmd/sdlc/stageplan.go`, new `cmd/sdlc/stagelanded.go`, new `cmd/sdlc/stagelanded_test.go`, `cmd/sdlc/stage.go`.

- [x] Write the native landed-holder regression and observe the original stage-lease refusal before implementation.
- [x] Implement the pure predicate and proof strategies above, retaining meaningful red evidence for unsafe cases.
- [x] Add the narrow proof route and final identity/lease revalidation under the existing lock. Reuse the single fetched main pin and strict existing proof helpers.
- [x] Run focused new tests plus stage lease, freshness/build, archive-generation and workspace controls; commit code and issue Log.

## Task 2: Document and close

**Files:** `cmd/sdlc/helptext/stage.md`, `README.md`, `atlas/workflow/sdlc-binary.md`, issue Log.

- [x] Explain original-slot/rest and confirmed landing/archive requirements, unknown-proof refusal and exact recovery commands. Keep label/deploy ownership unchanged.
- [x] Run compiled help, targeted Go/lint/recovery-catalog checks and git diff --check. Propose isolated native production-path proof as the live gate; no :0 or operator stage touch without hub approval.
- [ ] Push the implementation; retain a recorded suite on its exact head to completion. Log its real result, then complete plan ticks and run sdlc close.

On SHIP promptly enter sdlc merge --yes; retain integration verification, confirm archive and clean resting state, and report one outcome to the hub. No manual main chase.
