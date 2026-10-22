# #548 — Attribute suite waits to their issue heads

## Contract and scope

A suite's immutable tested head determines which issue may use its queue wait. Slot and wall-clock overlap then establish the owned portion for that issue; they cannot confer head membership. Echo0 legitimately holds two claims, so another issue's suite cannot poison or subtract from the measured issue. Preserve existing exact endpoint, source-slot, release/reclaim and correction rules. No new persisted fields or automatic changes to adopted hours.

## Core concepts

| Pure entity/function | Location | Change | Adversarial strategy and guard |
|---|---|---|---|
| QueueHeadMembership | cmd/sdlc/actualqueue_head.go | new | Known local issue set or unknown, pinned to exact OID. Local lead refs reuse issueref.LeadLocalNums, not a second numeric regex; citations and foreign qualifiers never confer membership. |
| selectQueueHeadMembership | same | new | Neutral integration/metadata commits inherit only their first-parent issue lineage; a definite different lead stops traversal. Real sibling branches and merged main controls reject reachability-as-ownership. Missing/truncated/ambiguous proof is unknown, never foreign. |
| filterIssueQueueAttempts | cmd/sdlc/actualqueue_filter.go | new | Drop positively foreign heads before endpoint/engagement checks. Relevant or unknown heads keep strict evidence requirements; permutation and duplicate/head bindings cannot transfer credit. |
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

- [x] Reproduce Echo's two-claims/same-slot/foreign-queue shape in a real Git/tracker measurement fixture, including failed attempts and the exact0-matching-engagements symptom.
- [x] Implement the pure membership/filter contract and bounded reader; wire once before owned queue selection. Preserve #544's prefix clipping and retained replay after integrating that landed main.
- [x] Verify the named function contracts below, including counted growth and existing owned-wait controls.
- [ ] Run targeted tests and recorded full suite; coordinate Echo-owned actual9/actual3 live previews; update atlas/help, close and land normally.

## Decisions and ARCH lenses

ARCH-PURPOSE: fix the foreign-run refusal without letting one issue adopt another's wait. ARCH-DRY/PURE: reuse reference grammar and existing interval selector, add one pure membership filter with a thin immutable reader. ARCH-ORDER/ERROR: head membership precedes slot/time ownership and missing proof stays explicit. ARCH-CONSTRAINTS/SCALE: bound complete traversal and count unique node reads. ARCH-EVOLVE: no persisted format change or historical numeric rewrite. ARCH-SIMPLE: the smallest alternative, exact equality with today's code head, drops earlier legitimate runs of the same issue and fails after normal commits/integration; pure timestamp/slot filtering reproduces this bug. A new claimed owner or hand-edited hours would invent authority and is rejected.

## Log

2026-10-22: Planning from main909bf77db0b84bfe73a4b494bec775c40d47b78e in an independent ordinary clone under ariadne1, while #544's recorded suite remains alive in the canonical checkout. First linked worktree was refused by the environment dependency gate and removed untouched; the independent clone satisfies that same guard. start-plan started the claimed card; its initial push refused stale remote-tracking information inherited from the local clone source. Fetched/pruned origin and confirmed no remote issue548 branch before the design checkpoint. No implementation admitted or changed.

Read issue.cue; no status/lifecycle or tracker record format change. Read-only Echo evidence: attempt fcf15650129203dfd1408728d625962b has SHA fc679c0745bbdd09896163978d3f24d7516b4564, slot echo0, queue2026-10-22T03:44:08.339787Z–04:10:52.119521Z; that exact Git commit leads with `#3: add generated actor and tool contracts`. The failure on actual9 is therefore positively foreign head membership, not a missing queue endpoint. No Echo state changed.

2026-10-22: Read-only seam audit while #544's fixed-head suite runs: shared `issueref.LeadLocalNums` accepts local qualified/area/verb/chained leads and excludes citations. Ordinary automatic integration is `Merge <mainOID> into <branch> (sdlc merge, #332)` plus `Integrated-Main: <mainOID>`; validate exactly two parents and trailer equals parent two, follow parent one only. A train reintegration is a neutral node inside the composition prefix: `Merge <mainOID> into landing train <id> (sdlc merge, #449)` plus Integrated-Main and Landing-Train; it adds no car and follows its first parent into same-train composition. Exact composition nodes use `Landing train <id>: #<issue> (PR #<pr>)` plus matching Landing-Train/Landing-PR trailers. A different train or main base ends that membership set. Arbitrary non-lead text is unknown, not evidence of neutrality or foreign membership. Existing producer paths are landingintegrate.go:187 and landingtrain.go:317/628; avoid modifying either producer. Real fixture seams include ownedResolverFixture/ownedFixtureAt/ownedFixtureCard and newTrainSetup/pushCar/integrationAdvance. TestLandingTrainCodeReintegrationDoesNotSplitOnStalePrefixes records the main-advance composition shape. No #548 code changed before #544 lands.

## Revisions

### 2026-10-22 — concrete bounded-reader and pure-reducer seam

Read-only preparation while #544's integrated suite runs clarified the approved thin Git reader. `gitRunner` has no stdin capability and `boundedLandingGitRunner` does not cap output bytes. Reuse the private bounded stdin/output/process implementation in `internal/gitx/boundedoutput.go`, exposed through a narrow immutable commit-batch query, following snapshot.go's exact cat-file batch framing validation. Do not broaden the generic runner or create shell stdin plumbing. Require requested OID/type/size/frame/trailing-output agreement, per-object and aggregate caps, cancellation and complete results. `landingOIDValid` is the existing main-package full40/64 lowercase nonzero OID validator.

Keep the pure parser/reducer IO-free: facts contain OID, parents and raw message; the parsed step describes direct, neutral, train-car, train-integration or unknown, its issue set/train/PR and required first parent. A reducer receives already-read facts and resolved parent membership. The integration walker owns batching, dependency worklist and command-local fact/step/result memoization; read each unique initial head then only unread first-parent frontier nodes, resolving each node once. Same-train composition uses the existing maxTrainCars16 bound (landingconfig.go); neutral traversal also has a complete-history bound/sentinel. The bounded train size keeps prefix copies constant-sized. Count unique node reads on shared-tail N/2N graphs, not just distinct head fetches. Different-train or nontrain base ends composition without inheriting base membership; a same-train integration requires the matching composed prefix and adds no car. Malformed purported train records and unreadable parents cannot masquerade as boundaries. No producer change or persisted field is added. Implementation still starts only after #544 lands.

Named function contracts for the admitted implementation:

| Function/entity | Kind/location | Adversarial strategy and guard |
|---|---|---|
| `gitx.ReadCommitObjects` / batch parser | INTEGRATION reader / PURE parser, new `internal/gitx/commits.go` | Real private Git and malformed batch frames verify exact requested OID/commit type/length/trailing bytes and existing cancellation/output caps. Shared bounded input/process runner is the independent IO guard. |
| `parseQueueHeadObject`, `parseQueueHead` | PURE, `actualqueue_head.go` | Parse immutable raw headers/message, reuse issue lead grammar, mutate parent/trailer/subject identities and citations. Direct, neutral and purported-train cases must not fall through to unrelated authority. |
| `selectQueueHeadMembership` | PURE reducer, `actualqueue_head.go` | Explicit parent-step/result values prove first-parent neutrality and same-train prefix union only, with16-car cap and unknown/malformed controls. No Git callback or mutable ref is an input. |
| `readQueueHeadMemberships` | INTEGRATION, `actualqueue_head_git.go` | Stateful graph reader and real-Git fixtures cover missing/cancelled/overflow results. Count unique fact reads and reducer resolutions at N/2N, including shared parent tails. |
| `filterIssueQueueAttempts` | PURE, `actualqueue_filter.go` | Permute own/foreign/unknown results; foreign skips before owned-evidence validation, while relevant unknown and duplicate identity stay strict. Existing owned queue selector supplies independent endpoint/engagement controls. |
| `attachOwnedQueueEvidence` | INTEGRATION, existing `actualinputs.go` | Real two-claims/same-slot fixture records A's failed suite then measures B; the old0-matching-engagement symptom must go red before wiring. Own wait, prefix clipping and historical replay controls retain existing measurement authority. |

### 2026-10-22 — plan-quality immutable authority and resource envelope

PQ-1 (`immutable-object-read-authority`): every new commit-object read invokes `git --no-replace-objects cat-file --batch --buffer`. The option belongs to the narrow reader, not caller convention. A real replacement-ref fixture substitutes another issue's commit and verifies that the original object still determines membership. All attribution parents come from those raw objects; no replacement-sensitive secondary history queries.

PQ-2 (`explicit-command-resource-envelope`): one command admits at most 10,000 distinct commit objects, 32 MiB aggregate raw commit content, and 256 first-parent edges per head. Bound each object to 1 MiB and each batch to 32 MiB including framing, with at most 128 requested objects per batch. These are safety ceilings: the node ceiling matches the store's 10,000-attempt input ceiling, aggregate/object caps reuse the existing Git snapshot safety envelope, and 256 edges allows many metadata/integration commits while refusing an unexpectedly deep lineage. Parsed facts replace raw objects after parsing; retained messages/parents plus accounting never exceed the aggregate raw-byte budget, and per-node metadata is bounded by the node ceiling. Check command-global admission before retaining each newly read object, deduplicate shared parents, and track derived lineage depth even when a parent was memoized by another head. On exhaustion refuse with `queue head attribution exceeds <node/byte/depth> safety budget`, naming the relevant attempt and SHA at the caller; restore bounded proof or seek an explicit limit revision, never use partial membership. Tests through `readQueueHeadMemberships` exercise each boundary at limit and limit+1, cross-batch byte accumulation, shared-parent depth, and N/2N unique-node counts. Smaller injected budgets make those tests fast; production defaults are asserted separately. Re-measure these ceilings against recorded suite graphs before raising them.

PQ-3 (`concise-function-test-strategies`): the verification checklist now points to the function-contract table; concrete cases live in executable tests.

### 2026-10-22 — implementation proof details

The command checks the lineage ceiling during discovery before requesting deeper unseen parents, and again over memoized dependency depths before returning membership. Neutral descendants share immutable issue sets; only train union allocates. This keeps derived membership storage proportional to parsed direct leads plus bounded train nodes, rather than multiplying a large lead list by every descendant. A conservative per-attempt probe reuses the existing owned selector only to rule out disjoint evidence; warnings/errors from that probe are deferred until immutable membership is known. Duplicate IDs are validated before foreign exclusion. The attachment receives the canonical issue ID explicitly. Adversarial resource regressions went red before these fixes.

### 2026-10-22 — filter module location (BR-1)

The pure attempt relevance probe and final membership filter live together in `cmd/sdlc/actualqueue_filter.go`, separate from raw commit parsing and membership reduction in `actualqueue_head.go`. Corrected both function-location tables to the implemented split; the pure/IO contract and approved behavior are unchanged. Audited all table locations against definitions, not only the first mismatched row.
