# #548 — Attribute suite waits to their issue heads

## Contract and scope

A suite's immutable tested head determines which issue may use its queue wait. Slot and wall-clock overlap then establish the owned portion for that issue; they cannot confer head membership. Echo0 legitimately holds two claims, so another issue's suite cannot poison or subtract from the measured issue. Preserve existing exact endpoint, source-slot, release/reclaim and correction rules. No new persisted fields or automatic changes to adopted hours.

## Core concepts

| Pure entity/function | Location | Change | Adversarial strategy and guard |
|---|---|---|---|
| QueueHeadMembership | cmd/sdlc/actualqueue_head.go | new | Known local issue set or unknown, pinned to exact OID. Local lead refs reuse issueref.LeadLocalNums, not a second numeric regex; citations and foreign qualifiers never confer membership. |
| selectQueueHeadMembership | same | new | Neutral integration/metadata commits inherit only their first-parent issue lineage; a definite different lead stops traversal. Real sibling branches and merged main controls reject reachability-as-ownership. Missing/truncated/ambiguous proof is unknown, never foreign. |
| filterIssueQueueAttempts | same | new | Drop positively foreign heads before endpoint/engagement checks. Relevant or unknown heads keep strict evidence requirements; permutation and duplicate/head bindings cannot transfer credit. |
| selectOwnedQueueAttempts | cmd/sdlc/actualqueue.go | unchanged contract | Existing owned interval clipping remains the final authority. Same-slot two-claim fixture reproduces the exact old refusal before filtering; own malformed evidence still refuses. |

| Integration | Location | Change | Boundary and proof |
|---|---|---|---|
| queue head reader | cmd/sdlc/actualqueue_head_git.go | new | Bounded, cancellable immutable Git reads; cache only within one command by full OID. A stateful commit graph fake covers missing objects/read failures; real Git covers integration and branch ancestry. |
| attachOwnedQueueEvidence | cmd/sdlc/actualinputs.go | modified after #544 lands | Supply explicit canonical issue ID and pinned code/repository context, classify candidate attempts before existing queue selection. No current-branch inference. Close preview/correction share this one route. |
| Echo owner live preview | ops-echo:0 temporary worktree | read-only | actual9 no longer consumes/refuses on #3 attempt; actual3 still subtracts its own exact queue. |

## Immutable head proof

Read each candidate's SuiteAttempt.SHA. A local subject lead establishes its issue set under the existing commit convention; `#3: ... (see #9)` belongs to3 alone. A neutral merge or bookkeeping head follows first-parent history to the first definite lead, with a bounded complete traversal. Do not walk second-parent main history and borrow another issue. A synthetic train head needs its explicit immutable admitted membership: consecutive same-train composition merges and their matching Landing-Train/Landing-PR metadata identify included cars, stopping before the base. Held cars and a different train ID never enter that set. Retain unknown when canonical proof cannot be established, rather than inventing membership from present branch names.

A lifetime lock cannot improve this check: the resource is the immutable Git commit identified by the saved full OID. Mutable refs are not authority and the command-local cache cannot outlive the measurement. Validate commit availability, parent identities and bounded traversal before treating absence as foreign.

The reader uses existing bounded/context-aware Git seams and shared issue-reference grammar. Reuse exact train metadata already emitted by the compositor; if that grammar needs extraction, keep producer/parser together and coordinate any landingtrain.go edit with the TL before code. No new queue state, SuiteAttempt field or schema version is planned.

## Visible states and recovery

Actions: preview, close, correct, restore evidence, retry.

| Internal state | Visible state | Recovery | Possible loss |
|---|---|---|---|
| Positively another issue's head | excluded from this issue's queue selection | preview/close normally | none; original attempt remains |
| This issue's head and complete owned wait | normal queue subtraction | preview/close normally | none |
| Relevant head with missing endpoints or ambiguous engagement | existing named queue evidence refusal | restore evidence, retry actual | no partial total adopted |
| Missing/ambiguous/truncated head proof for potentially relevant wait | named attempt ID and SHA attribution refusal | restore the immutable commit/proof, retry actual | no partial total adopted |
| Historical retained measurement | existing exact replay | explicit correction selects fresh queue evidence | no automatic rewrite of adopted hours |

Known disjoint repository/check/baseline/interval candidates can be rejected before Git reads using the current provable non-overlap rules. Missing queue starts cannot be treated as disjoint solely because acquisition occurred after the sample. An unknown head cannot be quietly classified foreign.

## Cost and bounds

Let A be potentially overlapping suite attempts, U their unique head OIDs, and V the unique first-parent commit nodes actually inspected. Deduplicate heads, read direct head facts in bounded batches, and memoize immutable lineage nodes once per command; no per-event or per-engagement Git walk. Expected O(A+V), with one bounded history budget per unique lineage and existing record/IO byte caps. Count reader/discovery calls at N and2N unique heads, and shared-parent graphs, asserting no repeated node reads. Timing is diagnostic only. Sentinel overflow refuses before a partial membership is used.

## Implementation and verification

- [ ] Reproduce Echo's two-claims/same-slot/foreign-queue shape in a real Git/tracker measurement fixture, including failed attempts and the exact0-matching-engagements symptom.
- [ ] Implement the pure membership/filter contract and bounded reader; wire once before owned queue selection. Preserve #544's prefix clipping and retained replay after integrating that landed main.
- [ ] Verify direct heads, neutral metadata/integration heads, earlier same-issue suite heads, references in citations, distinct issue sibling branches, train included/held membership, missing/cancelled/truncated objects, deduplication and N/2N counts. Existing own-wait endpoint/refusal controls remain strict.
- [ ] Run targeted tests and recorded full suite; coordinate Echo-owned actual9/actual3 live previews; update atlas/help, close and land normally.

## Decisions and ARCH lenses

ARCH-PURPOSE: fix the foreign-run refusal without letting one issue adopt another's wait. ARCH-DRY/PURE: reuse reference grammar and existing interval selector, add one pure membership filter with a thin immutable reader. ARCH-ORDER/ERROR: head membership precedes slot/time ownership and missing proof stays explicit. ARCH-CONSTRAINTS/SCALE: bound complete traversal and count unique node reads. ARCH-EVOLVE: no persisted format change or historical numeric rewrite. ARCH-SIMPLE: the smallest alternative, exact equality with today's code head, drops earlier legitimate runs of the same issue and fails after normal commits/integration; pure timestamp/slot filtering reproduces this bug. A new claimed owner or hand-edited hours would invent authority and is rejected.

## Log

2026-10-22: Planning from main909bf77db0b84bfe73a4b494bec775c40d47b78e in an independent ordinary clone under ariadne1, while #544's recorded suite remains alive in the canonical checkout. First linked worktree was refused by the environment dependency gate and removed untouched; the independent clone satisfies that same guard. start-plan started the claimed card; its initial push refused stale remote-tracking information inherited from the local clone source. Fetched/pruned origin and confirmed no remote issue548 branch before the design checkpoint. No implementation admitted or changed.

Read issue.cue; no status/lifecycle or tracker record format change. Read-only Echo evidence: attempt fcf15650129203dfd1408728d625962b has SHA fc679c0745bbdd09896163978d3f24d7516b4564, slot echo0, queue2026-10-22T03:44:08.339787Z–04:10:52.119521Z; that exact Git commit leads with `#3: add generated actor and tool contracts`. The failure on actual9 is therefore positively foreign head membership, not a missing queue endpoint. No Echo state changed.
