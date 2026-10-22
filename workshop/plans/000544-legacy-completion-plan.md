---
collaboration-mode: discussion
---

# Legacy completion preservation and repair implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3. Execute the coupled evidence/repair work in this session; use bounded read-only agents for independent audits. SDLC owns plan-quality and boundary review.

**Goal:** Preserve legacy close evidence through tracker migration and repair existing imports so later actuals measure without guessing earlier work windows.

**Architecture:** One pinned legacy-close resolver supplies both initial migration and an explicit single-card repair. Completion refers to the original evidence commit and its retained inputs; an additive legacy discriminator permits only positively proven windowless closes to establish earlier disjointness. Existing adopted-credit replay/subtraction remains unchanged.

**Tech stack:** Go, existing Git/tracker adapters, version-1 additive card envelope, real private Git fixtures and stateful publication fake.

## Scope and alternatives

Implement the three requirements together: import, repair and conservative windowless ordering. Do not correct numeric actuals, invent source sessions/windows, waive actuals, reopen/re-close historical issues, rewrite archived bodies, or change the active-time allocation engine.

Prefer explicit `sdlc issue migrate --repair-completions --issue N` over overloading `--reconcile`: the latter already means branch conversion and should keep that behavior. The repair is read-only by default and applies only with `--apply --expect DIGEST`. A single card makes publication atomic and recovery inspectable. Blanket skipping of missing completion would hide damaged modern evidence; manually adding a card binding would duplicate the verifier and provide no safe repeat.

## Core concepts

| Pure entity | Lives in | Status |
|---|---|---|
| `issue.LegacyAdoption` in the existing tracker envelope | new `cmd/sdlc/internal/issue/legacyadoption.go`, `handoff.go` | new/modified |
| `LegacyCloseEvidence` and validated binding selection | `cmd/sdlc/internal/tracker/legacycompletion.go` | new |
| Legacy import and repair transformation/digest | `cmd/sdlc/internal/tracker/migration.go`, `legacyrepair.go` | modified/new |
| Proven windowless-close temporal decision | `cmd/sdlc/actualexclusion.go` | new |

Completion remains one exact close generation per card, with the original reviewed head/evidence commit and existing repository identity. Preserve its structure. Add one optional `tracker.legacy_adoption` record for the imported numeric generation: repository, pinned pre-cutover source main, source path/blob, terminal archive/done commit, optional exact completion token, and window class (`retained`, `windowless-close`, `windowless-terminal`). Known IDs/path/enum/types stay strict; nested unknown fields survive every setter. The source blob pins adopted actual/started/identity; validate them against the current card. Derive end time from the verified close object or terminal object, never from an independently mutable timestamp.

A completion-only design cannot represent a born-archived record without fabricating a close. The single optional LegacyAdoption record covers both cases: original completion binding is set only for a proven close; terminal-only provenance leaves it absent and visibly unbound. No copied snapshot or parallel measurement ledger; original `ActualInputs` remain authoritative at the original evidence commit. A new close clears legacy provenance, and readers also reject a mismatched completion token/generation, protecting older-reader rewrites that preserve unknown fields.

`LegacyCloseEvidence` holds immutable close/archive/path/issue/actual/started/review/ledger observations gathered by the shell. Pure selection checks consistency and returns a binding or a concrete refusal. Reuse it for both consumers (ARCH-DRY). Existing active codecomplete import rules stay intact, sharing verification where their evidence model matches; they need no fabricated archive.

The temporal function receives a proven windowless legacy close/terminal end bound and earliest candidate engagement. Return disjoint only for a nonzero close/terminal bound strictly before candidate `Snapshot.Since`; equality, overlap, missing bounds and unknown proof remain missing-window. Feed this into existing `classifyAdopted`, preserving all slot/replay behavior.

| Integration point | Lives in | Status | Wraps |
|---|---|---|---|
| Bounded immutable legacy evidence resolver | `cmd/sdlc/legacycompletion.go` | new | Existing contextual Git runner and ledger readers |
| Migration inventory / manifest | `cmd/sdlc/issuemigrate.go` | modified | Pinned main, active/archive issue paths and tracker bootstrap |
| Single-card repair driver | `cmd/sdlc/issuemigrate_repair.go` | new | Cutover marker/root and `tracker.Repository.UpdateCardWithTrailers` CAS |
| Actual exclusion / observations | `cmd/sdlc/actualexclusion.go`, `actualinputs.go`, issue-show observation rendering | modified | Pinned legacy proof and retained snapshot reader |

The Git seam is the existing contextual runner/TrunkFile. Reuse private bare-repository fixtures for actual commit ancestry, archive renames and tracker cutover. Reuse the existing stateful publication fault adapter for concurrent writers and lost replies. No GitHub API dependency or new service is introduced. Native Git conformance is exercised by those fixtures; the approved live follow-up is an exact dry-run/apply/repeat repair of Echo #1 and read-only actual #3 in its owner's checkout, coordinated through the hub.

## Evidence contract

1. Pin source main before cutover. For repair, resolve the actual cutover commit and its parent from the marker's tracker root; prove it is the unique cutover for this repository. Read the imported card at that immutable root, not a current checkout's historic-looking details.
2. For the selected archived generation, traverse the active and archive paths within that pinned history. Require a real archive transition and issue identity/started/actual agreement. A filename or commit subject alone is insufficient. Respect renamed paths and configured roots; do not infer an ID solely from arbitrary history.
3. When close evidence exists, resolve the latest review-bearing successful close of the final archived generation. Validate SHIP, unique Review-Window trailers, full resolved reviewed OID/ancestry and boundary ledger agreement using existing review-window/ledger helpers. A subsequent metadata edit retaining codecomplete is not a close. Multiple legitimate re-closes are supported by ordering within one final generation; incomparable closes, reopen/ID-reuse ambiguity or mismatched final actual refuse.
4. Keep the original evidence commit. If its retained measurement is valid, mark retained and let `completedMeasurementSnapshot` discover it. If retained inputs genuinely do not exist, mark windowless; absent old ledger is distinct from malformed/ambiguous ledger. Bad saved inputs, foreign repository, wrong issue/head or failed reads refuse, never downgrade to windowless. If no review-bearing close ever existed, retain the import unbound with an explicit warning and `windowless-terminal` provenance only when the terminal Git commit and source archived generation are positively proven. A record born archived/done uses that actual done commit; a genuine archive uses its archive transition. Do not turn mere history absence or truncation into terminal proof. Repair of an unprovable close refuses; terminal-only imports are handled by migration provenance rather than a fake completion.
5. At exclusion time, retained inputs and validated corrections have priority. Only explicit, revalidated windowless legacy provenance can use the close or terminal object's committer time as a conservative end bound, according to its class. Prove the bound belongs to this archived adopted generation and is reachable in pinned code history. Compare against the earliest admitted engagement, covering park/reclaim history. Missing modern completion is never legacy by inference or token prefix.

## Repair authority and visible states

Actions: inspect dry run, apply exact digest, rerun inspection after uncertainty, restore evidence or ask owner to resolve a refusal.

| Internal state / failure | Visible state | Recovery | Possible loss |
|---|---|---|---|
| Imported done card lacks binding; exact legacy proof exists | repairable, close/window class/digest shown | apply printed exact command | none; original actual preserved |
| Born-archived/done with no close but proven terminal generation | unbound legacy, terminal bound shown, warning | continue migration; later actual skips only if strictly earlier | no close or window is fabricated |
| Valid exact completion already present | already complete; no write | no action | none |
| Root/status/actual/generation differs, correction exists, or proof is ambiguous/damaged | refused with named evidence | owner resolves evidence then reruns dry run | none; no write |
| Tracker card changed after dry run | stale plan/card; refused | rerun dry run | none; concurrent edit preserved |
| Push reply unknown | outcome unknown, exact retry command | rerun inspection/repair; observe remote binding first | remote may already contain repair; no duplicate semantic effect |
| Windowless close or unbound terminal strictly predates earliest engagement | disjoint legacy adoption | continue actual | none; no guessed subtraction |
| Windowless close overlaps/equal/unknown | trustworthy window unavailable | restore retained evidence or resolve with owner | measurement unavailable; numeric history preserved |

Eligibility requires membership at the marker's tracker root, root completion absent, current done/actual/started still the same imported generation, no correction and no intervening reopen/re-close adoption generation in complete tracker history. Existing completion is preserved, never replaced. Title/owner/body/unknown metadata survive. Apply re-reads and validates the exact expected card blob, then one CAS write provides the fence against a concurrent lifecycle change. No absence-by-PID or checkout scan authorizes a write.

Digest binds repository/root, pinned source evidence, current card path/blob and exact replacement bytes. Print the candidate and deterministic operation token before push. Lost acknowledgment re-enters through the same explicit command: an observed matching binding is complete; absent unchanged card may safely apply; a moved generation refuses. Reuse typed uncertain-publication errors and current Git semantics, with no blind retry, new receipt family or direct ref writes. `--reconcile` and normal cutover flags remain unambiguous and mutually validated.

## Bounds and lifecycle

This is an explicit migration/repair path, not a hot path. Bound per-issue path history using the existing 10,000-entry limit plus overflow sentinel, existing Git blob/output/context budgets, and complete bounded tracker history. A truncated or failed history is a refusal. Each issue/path history is scanned once; candidate examination is O(H) and all migration issues O(NH), with memoized object/ledger reads inside the command. Count history scans at N and 2N and assert linear growth/no per-candidate full-history rescan. Apply changes one card. Actual's ordinary modern path gains no legacy-history scan; only explicitly marked legacy candidates invoke proof, cached per measurement.

No durable family is introduced: provenance adds a bounded number of scalars to one existing completion per imported card; Git retains its original evidence and one repair commit per repaired card. Card history is already permanent and repair is idempotent, so retries add no commits. Temporary maps/observations die with command scope; no background jobs or source snapshot copies survive.

## Implementation and verification

### Named risky functions and test strategies

| Function | Adversarial strategy and independent guard |
|---|---|
| `issue.LegacyAdoption.validate` / `SetCardLegacyAdoption` / `SetCardCompletion` | Generate malformed/newer envelope structures; strict type/duplicate validation and existing unknown-field round trips must reject invalid records and preserve future data. |
| `tracker.SelectLegacyClose` | Permute conflicting generation/review/archive observations; pure property tests prove selection invariant under input order and a missing-evidence control cannot produce proof. |
| `resolveLegacyCompletion` | Replay actual immutable Git close/archive histories through the contextual runner; production fixtures and verifier-removal mutations prove metadata edits cannot manufacture a reviewed close. |
| `tracker.PlanLegacyCompletionRepair` | Model stale or changed imported generations; pure transformation tests assert no output outside eligibility and byte preservation for unrelated fields. |
| `tracker.LegacyCompletionRepairDigest` | Mutate each authority/input/replacement component independently; every semantic plan change must alter the digest, reordered canonical inputs must not. |
| `applyLegacyCompletionRepair` | Stateful publication schedules race CAS and lose replies; ref/card snapshots prove refusal preserves concurrent work and repeated confirmation adds no commit. |
| `legacyWindowDisjoint` | Property/table tests over timestamp order and missing proof; only strict earlier proven windowless closes may skip, independent of zone representation. |
| `loadAdoptedExclusions` | Run actual measurement over retained and windowless imported generations with pinned engagement history; park/reclaim and provenance mutations must preserve strict refusals and valid replay subtraction. |

### 1. Import proof and format

Files: `internal/issue/handoff.go` and codec tests; new `internal/issue/legacyadoption.go` and tests; new `internal/tracker/legacycompletion.go` and tests; `internal/tracker/migration.go`, `issuemigrate.go`; new `legacycompletion.go` and real-Git tests.

- [ ] Write named pure/codec regressions above and production close → archive → migration → issue-show → later-actual and born-archived unbound → later-actual regressions; observe red before import wiring.
- [ ] Implement shared proof and archived import, retaining existing active codecomplete controls and original snapshot references. Run targeted tests and mutation guards.

### 2. Atomic repair and retries

Files: new `issuemigrate_repair.go` and tests, `issuemigrate.go` flags; existing tracker CAS adapter reused.

- [ ] Write already-migrated missing-binding production regression through dry-run/apply/repeat/later-actual and stateful publication schedules; observe red before repair wiring.
- [ ] Implement single-card eligibility/digest/CAS repair and recovery output, with the named guards proving no checkout/archive/numeric rewrite or duplicate retry effect.

### 3. Conservative exclusion and observation

Files: `actualexclusion.go` and tests; reuse helpers from `actualinputs.go`; issue-show observation rendering/tests.

- [ ] Write named temporal properties and actual-measurement production regressions; observe red for the Echo-shaped windowless close.
- [ ] Implement explicit validated legacy ordering before slot/replay requirements. Print original close and retained/windowless class without reporting a fabricated window. Run existing decision-table and retained-replay controls.

### 4. Documentation and conformance

Files: `cmd/sdlc/helptext/issue.md` or extracted migration help, `README.md`, `atlas/workflow/sdlc-binary.md` and tracker/actual map, `workshop/lessons.md`.

- [ ] Document command/recovery and conservative disjointness, add lesson, run targeted migration/issue/tracker/actual tests, relevant race, tracked-file lint and touched-package vet. Commit/push and queue one full recorded suite. SDLC owns close review.
- [ ] Give ops-echo:0 the exact built binary and Git source for their owner-run dry-run/apply/no-op-repeat/actual3 conformance; preserve returned raw output on SSD and record the pins. If they choose post-landing conformance, record that authorized timing and report any failure immediately.
- [ ] Passing final-head suite precedes landing. Close with measured actual/no-project, then run `sdlc merge --yes`; its runner owns integration and the authoritative suite. Report landed PR and Echo outcome to hub.

## Architectural check

ARCH-DRY: shared proof, existing ledger and CAS. ARCH-PURE: selection/temporal/digest separated from Git IO. ARCH-PURPOSE: import, repair and later actual all delivered. ARCH-MOCK: private real Git plus stateful publication fault model and coordinated Echo conformance. ARCH-CONSTRAINTS: explicit history/output/context bounds and counted scans. ARCH-SECURE: local repository-bound evidence, no arbitrary supplied commit/card trust. ARCH-ORDER: immutable proof plus single-card CAS, with conservative unknown outcomes. ARCH-FUNERAL: bounded fields, one idempotent repair, no new artifacts. ARCH-EVOLVE: optional scalars in version1; known types strict and unknown members preserved. Old readers ignore legacy hints and continue refusing windowless measurement, rather than granting unsafe credit.

## Approval

Pending TL review of plan, repair command shape and Echo live-conformance authority. No implementation or peer state mutation before approval.

## Revisions

### 2026-10-22 — TL approval and Echo owner

TL approved the shared verifier, explicit digest-reviewed single-card repair, validated windowless provenance and strict earliest-engagement ordering (delivery d46eb30c-7e6c-4da2-9368-129e6b4cdb7c). Live conformance is owned by ops-echo:0: they run dry run/apply/no-op repeat and actual3 in their temporary origin/main worktree and return output (delivery 904d435d-de8b-446e-8fea-489c1c90ddbe). Hand off an exact built binary before landing when ready, or notify after main contains it. No repair in Echo from this slot.

### 2026-10-22 — plan-quality test contract (PQ-1)

Named the selection, repair eligibility/digest, temporal, validation and effect functions; replaced prose case inventories with one adversarial strategy/mechanical guard per risky function, retaining real production-flow objectives. No mechanism change. Existing migration fixtures also contain born-archived done records with no historical close at all; proposed compatibility clarification: initial migration retains such unbound cards with an explicit warning, never fabricating proof, while malformed/contradictory evidence and repair of an unprovable close refuse. Their overlapping actual exclusion remains strict. Awaiting TL confirmation of that narrow distinction before code.

### 2026-10-22 — terminal-only legacy evidence (TL requirement)

TL delivery 82679d7d-74d5-4320-9051-4232bdd6f97f requires unbound born-archived imports also to stop blocking later measurement when their proven archive/done commit strictly predates the earliest admitted engagement. Replanned the simplest representation: one additive LegacyAdoption envelope member for both close-backed and terminal-only generations; Completion remains unchanged and absent for unbound records. Terminal-only provenance is explicit and revalidated, not inference from missing modern completion. Added production migration/later-actual conformance for unbound imports; overlap remains strict. This supersedes the completion-scalar proposal without weakening its authority boundary.

### 2026-10-22 — old unbound imports, read-only recovery

TL delivery ec3a2adf-2b31-4b08-a330-56f170a32a64 requires existing born-archived imports to benefit too. Replan the remaining representation gap through the same proof boundary: when explicit provenance and completion are absent, the actual reader may recover a terminal bound read-only only from a done/numeric card at the actual tracker root matching the exact pre-cutover archive generation. It writes neither a card nor a completion. This supersedes an assumption that all missing-provenance imports require explicit close repair; modern missing records and overlapping bounds remain strict. A completion-only alternative fabricates closes; blanket missing-record exemption admits damaged modern records, so both fail. `legacyImportedSource` becomes the shared immutable selector for repair and read-only recovery; `verifyLegacyAdoption` owns root/source validation. Its named adversarial guard is a real already-migrated born-archive → claim → actualMeasured fixture with unchanged remote card/ref, plus modern absence and overlap controls.

The actual measurement batch now performs bounded root eligibility reads for unmarked numeric missing completions; only proven imported candidates receive a source history scan. Modern valid completions retain their existing path. Cost remains O(NH) for N relevant legacy candidates and bounded H, with command-local temporary data only. Existing source traversal is one 10000-entry-plus-sentinel walk per candidate; actual card history uses the existing5000 bound.

The issue vocabulary's CUE models the internal envelope field/version, not the existing Handoff/Completion/ActualCorrection record kinds; the new internal record follows that existing Go-owned contract rather than adding a duplicate record registry. Observation output displays all three adoption classes, including unbound terminal-only records, per TL delivery55e35115-04b6-43b8-856e-7de8e15423ef.

### 2026-10-22 — native copy-history boundary found by Echo conformance

The real pre-landing dry run revealed that Git's low similarity --follow may continue into a non-issue template before the issue existed. Keep the complete bounded scan and damaged-identity refusals. The added effect helper legacyHistoryIssueBirth proves the oldest issue path absent in every parent before truncating only non-issue template ancestry; any other issue path beyond that boundary refuses. Its adversarial strategy is a real Git copy-template/archive/close fixture plus a merge whose second parent already has the issue. The independent guard is the existing damaged-identity/recreated/incomparable-history matrix; ordinary similarity defaults failed that guard and were rejected. This corrects source traversal within the approved immutable-generation verifier, without adding authority or persisted state.

### 2026-10-22 — cutover-spanning measurement (TL scope revision)

TL delivery9668b551-4c1c-44d3-8751-52b9a5f6a69 folds Echo's live 0.03h result into this issue: pre-cutover development must not disappear. A legacy claim proves its start but not a modern workspace owner, so importing its timestamp as tracker ownership would invent authority. Proposed design: retain the existing labelled repository-plus-brain legacy heuristic for a separately bounded pre-cutover prefix, then join its attributed segments with the verified tracker suffix. Both components use the existing active-time engine and one final build/converge/tail split; no hours are guessed or typed. TL approved this revision in delivery66dc716c-80f0-41d0-97d0-db740be9ebee.

- `resolveLegacyActualPrefix`: read the real cutover commit and imported active card at tracker root; select the exact immutable pre-cutover source with the migration selector. Use its legacy started stamp/working transition, else its pinned first issue commit, as the existing legacy rule does. Root-absent modern issues and terminal imported issues gain no prefix. Malformed/contradictory source, pin or timestamps refuse. The prefix ends at the cutover commit; clip verified tracker source/queue intervals to begin there, so bootstrap-before-cutover cannot double-count. Retain the original tracker dispatch origin separately for child provenance.
- `MeasurementInputs` gains an optional legacy prefix, represented by a nonrecursive child of the existing legacy-v1 input shape, plus a cutover commit pin. This avoids pretending the prefix is owned or inventing a second engine. Validate one level only, source/byte limits across both, prefix bounds/commit OIDs, and no interval crossing the boundary. Primary tracker method and ownership evidence remain exact; visible output explicitly names both attribution rules and their intervals.
- `resolveLegacyPrefixSources` discovers the same repository+brain population through the existing harness registry, pins exact transcript prefixes and clips all prefix activity to its interval. `computeLegacyPrefixActivity` uses the existing legacy allocation rule without invented source slots or queue evidence. A pure segment/result join sums allocations and sorts boundaries before the existing phase split; it must preserve uncertainty from either component instead of reporting partial totals.
- Retained close/correction replay revalidates pinned cutover/source authority and replays both exact source populations/digests. Later transcript appends, unrelated tracker changes and repeated actual must not move the old prefix. No card write or fabricated completion is needed for preview. Existing exclusion uses the combined earliest admitted start; overlapping old missing evidence remains strict.

Actions remain preview/close/correct/retry. Visible states: complete mixed measurement → preview/adopt through normal close, with both rules labelled; missing or contradictory prefix/source → measurement refused with the missing proof named, repair that evidence before retry, never adopt a suffix-only value; root-absent modern issue → existing owned measurement unchanged. Persistent evidence may gain bounded prefix metadata; no source transcript bytes are copied.

Named adversarial strategies: resolver production fixture claims under legacy, migrates, adds owned work, then measures and closes with positive contributions on both sides; mutate away prefix resolution and it undercounts. Prefix validation rejects recursion, overflow, malformed pins and boundary overlap. Pure result join preserves exact sums/permutation and no double-count at the cutover. Replay test appends transcripts and advances unrelated refs yet returns the retained number; damaged prefix refuses. Controls cover modern root-absent, terminal imports, missing legacy start fallback, parked/reclaimed ownership after cutover, shared repository/brain sources and both harness kinds (Qoder where available). The owner repeats Echo actual3 and previews the other spanning issues after the new binary.

Cost: one bounded cutover/source/legacy claim traversal per qualifying issue plus one source-discovery pass for the prefix population; O(H+S+E) for bounded Git history H, sources S and events E, using existing source caps. Count discovery once at N and 2N sources and assert linear prefix processing; do not rescan per event. ARCH-ORDER keeps heuristic legacy provenance distinct from modern ownership; ARCH-DRY reuses engine and persisted input shape.

The legacy population is caller plus its verified primary Git checkout and brain, not every linked worker. This preserves the repository heuristic when preview runs from an owner-created temporary worktree. Source discovery counts one pass at N and2N candidates. The same-generation verifier now checks complete pinned tracker-card history, including restored endpoints, as the original immutable-generation contract requires.

### 2026-10-22 — close review compatibility and generation-continuity correction

Round-one close review returned REWORK with BR-1 and BR-2. TL approved this correction in delivery24b0ff95-3691-4347-82fa-aed92f08a7b0. Hybrid snapshots require an explicit new method and version that older decoders reject loudly; silently ignoring an optional prefix changes the measured total and is unsafe. Ordinary nonhybrid version1 snapshots remain supported. The legacy child remains a bounded legacy-v1 component, while the hybrid root identifies its complete interpretation. Cross-version tests use the pinned base decoder, prove the current encoding is rejected there, and keep ordinary old snapshots readable. Modern validation rejects mixed or downgraded method/version/prefix combinations. This supersedes the additive-only measurement compatibility assumption above (ARCH-EVOLVE).

Repair and actual exclusion must invoke one complete monotone generation-history validator. Matching completion/provenance may be added once to the unchanged imported numeric generation, but may never be removed or replaced and later restored. A repair dry run must refuse any history its eventual actual consumer would reject. The real sequence regression imports an unbound card, repairs it, restores the old unbound bytes, then proves a second repair refuses without publishing; matching positive repair-to-exclusion controls remain. Completion and provenance removal/replacement variants defend the whole class (ARCH-DRY, ARCH-ORDER).

Core-concept location/name corrections: the pure repair transformation and digest are `tracker.PlanLegacyCompletionRepair` and `tracker.LegacyCompletionRepairDigest` in `cmd/sdlc/internal/tracker/legacyrepair.go`; migration selection remains in `migration.go` and `legacycompletion.go`. The delivered prefix helpers are integration functions `legacyActualPrefixAuthority`, `resolveLegacyActualPrefix`, `discoverLegacyPrefixInputs`, `discoverLegacyPrefixSources`, `bindLegacyPrefixSources`, `validateRetainedLegacyPrefix` and `measureLegacyPrefix` in the corresponding `actuallegacyprefix*.go` files. `joinLegacyActualResults` is the pure reducer in `actuallegacyprefix_measurement.go`; `clipOwnedAfterCutover` is pure in `actuallegacyprefix.go`. Earlier proposed helper names describe those same responsibilities and are superseded by these exact names.

Both findings need red-first regressions, targeted/race checks, a new pushed-head recorded suite and another binary-owned close review before landing. The prior suite remains evidence for its old head only.

Delivered correction names: `legacyGenerationHistory` in `cmd/sdlc/legacycompletion.go` is the shared integration validator called by repair, explicit adoption verification and old unbound recovery; `legacyRepairGenerationUnchanged` is deleted. Hybrid roots use version2 with `MeasurementMethodHybridV1` or `MeasurementMethodHybridQueueV1`; the bounded prefix child stays version1 legacy-v1. The frozen pre-change decoder fixture lives in `internal/gatestate/testdata/measurementv1/measurement.go` and rejects each discriminator independently. Production replay, queue promotion and correction trust tests exercise the consumers of both new methods.

### 2026-10-22 — actual publication-base authority and lossless measurement rewrites

The isolated real-tracker publication probe proved that a previously prepared repair can still publish after authority is added and removed back to identical card bytes: a card-blob CAS permits a new tracker history. TL delivery3cb0a86f-2063-4d74-8328-54865f0e6a88 approves validating the actual publication base with the same `legacyGenerationHistory` before every push attempt, including a rebased retry. The base remains protected by the existing conditional push, so a move after validation retries against a newly validated base. Tests force both an intervening restored-authority history and a competing push/retry, and retain positive unrelated-card movement. No new durable state or second validator. This is the second concurrency correction for BR-2: if the next review identifies a third repair concurrency window, stop and bring it to the TL; the approved fallback to consider is requiring tracker quiescence rather than adding more validation sites.

Round2 close review independently disposed BR-1 and BR-2 as addressed but returned BR-3 for accepted unknown measurement fields being discarded on rewrite. TL delivery6cd081e8-dc7b-4aa8-a453-af4c50404bb3 approves the family-wide rule: every accepted unknown member survives JSON and YAML decode/rewrite at MeasurementInputs root and its bounded legacy_prefix, engagements, sources, source windows, waits, protected credits and cuts. Preserve typed known-field checks, duplicate-key refusals, one-level prefix depth/cycle guards, combined byte cap and source cap. Enumerate path by record type and representation in red-first tests; cover the real `readGateLedger`/`writeGateLedger` rewrite in addition to pure codec round trips. Keep this strictly within the measurement-input family this issue introduces or rewrites; unrelated tracker evolution remains separate work.

The documented complete tracker history bound is corrected to the enforced10,000 entries plus overflow sentinel, superseding the earlier5,000 statement. The Core concepts table now directly names legacyrepair.go for the pure repair transform/digest, matching the earlier erratum. After both corrections finish, run targeted/race/lint/vet, a new recorded pushed-head suite and the next close review. Echo remains owner of live previews; #548 implementation still follows #544 landing.
