# Issue-scoped Actual Measurement Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3. Delegate bounded pure parsers when useful; retain measurement integration in the main session. One atomic issue close owns the review.

**Goal:** Make close and actual agree across moves and release/resume, and correct both documented bad records using measured evidence.

**Architecture:** Derive one pinned measurement input from issue ownership and completion history. Reuse the harness registry, active-time engine, split logic, tracker CAS/receipts, and calibration supersession. Preserve the gap cap and intended parallel-agent accounting.

**Tech Stack:** Go, Git tracker history, Claude/Codex JSONL metadata, TSV calibration ledger.

## Evidence

- pair#454 close `3d8523e9f59fb59120447e3e4f35fee1e35b9d24` recorded 0.05h. Raw Claude session `51e5e28c-0ea9-4824-a11e-8bfda24922ae`, line 1320, explicitly runs close from pair:0 while its session remains rooted in pair:1. Retained close.log prints Oct11 20:08:38 PDT as the start. A current-engine replay over 20:08:38–22:22 selects 3.3 minutes from pair:0+brain versus 118.6 minutes from pair:1 alone. This proves source loss; it does not establish the prior 2.57h as the corrected total.
- pair#514 initial claim `dcecfc402b4cc3c87bb921d461a19b3ad4447ed1` stamps Oct12 18:20:35 PDT. Release `d228d955b71cd14789b8b1364df7ed18553e5b06` at 18:21:14 removes the owner but retains started. Re-claim `3f2bc6490edbe22d7a7538b6da909152f5c6769d` on Oct13 18:29:07 retains it. Close `86bd8e8074a4d55f931b34b8aeb7b96da8a2c61b` records 23.76h. Codex session `01a13646-a9b3-7443-a36e-43f301879fd6`, lines 8454/8474, captures close and actual both reporting 23.76h.
- Independent capped-gap event accounting reconstructs 22.226980h Codex + 1.528758h Claude = 23.755738h → 23.76h. Selected sessions performed pair#483/#513/#516/#524/#533/#538 during the released period. Intervals alone still admit 3.573 min from an old #533 child crossing the new claim; source lineage matters. Suspected copied histories were independent sibling agents, with no shared full response records or call IDs.
- Both callers share computeActualFor, but select the caller's exact cwd. Explicit brain overrides also differ. The printed right edge says HEAD while the engine uses the last matching issue commit; current branch scope changes after landing.

## Alternatives

1. **Recommended:** derive engagement intervals and source provenance from pinned history. Preserve early shaping and moved-session activity; make incomplete evidence explicit.
2. Reset started and admit every worktree: loses genuine earlier work and admits unrelated concurrent activity.
3. Add warnings only: reveals suspicious totals but leaves attribution wrong.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|---|---|---|
| OwnershipEvent / EngagementWindow derivation | `cmd/sdlc/internal/tracker/engagement.go` | new |
| Session provenance / eligibility | `cmd/sdlc/internal/transcripts/provenance.go` | new |
| Actual input/provenance value | `cmd/sdlc/actual.go` | modified |
| Activity interval clipping | `cmd/sdlc/internal/activetime/event.go` | modified |
| Estimate plausibility comparator | `cmd/sdlc/actualplausibility.go` | new |

An issue owns zero or more engagement windows. Claim opens, release closes, and re-claim opens another. Relocate/reclaim transfers preserve genuine preceding work. Started remains first-engagement metadata, not a continuous attribution window. Derive transitions from card snapshots, including legacy history without operation trailers; reuse existing tracker operation constants and parsing.

Only recorded participant roots for the verified same repository are eligible. A persistent source-rooted main session remains eligible after a move during the continuous engagement, because shell cd does not change session metadata. Descendants inherit their dispatch engagement; an old child cannot join a newly claimed issue merely because its cwd and timestamps overlap. Claude Agent spans inherit parent task provenance; Codex thread_spawn metadata identifies parent/dispatch context. Preserve legitimate parallel agents.

Missing/inconsistent ownership or lineage produces an explicit quality limitation, not confident inclusion of every same-cwd source. Legacy no-tracker measurements retain a labeled fallback. Ambiguous required evidence prevents automatic adoption/correction with an actionable diagnostic. Preserve the existing gap and mention-weighting rules.

### Integration points

| Name | Lives in | Status | Wraps |
|---|---|---|---|
| Pinned input resolver | `cmd/sdlc/actualinputs.go` | new | workspace identity, refs/card history |
| Harness source discovery | `cmd/sdlc/internal/transcripts/{claude,codex,transcripts}.go` | modified | files and lineage metadata |
| Explicit measurement code ref | `cmd/sdlc/internal/gitx/window.go`, `cmd/sdlc/internal/activetime/commit.go` | modified | Git log/ranges |
| Shared consumer wiring | `cmd/sdlc/{actual,close,milestoneclose}.go` | modified | resolver and existing engine |
| Measured correction command | `cmd/sdlc/actualcorrection.go` | new | tracker CAS/receipt and calibration |

Use existing Git/tracker seams and real portable temporary repositories with file-backed Claude/Codex fixtures. One snapshot names code HEAD, tracker HEAD, branch scope, effective start/end, engagements, and selected/excluded source provenance. Close measures its pinned review input; a completed issue resolves completion.reviewed_head and its finalized ledger instead of the caller's later main HEAD. Verify scope/split parity through landing. Respect explicit brain-dir identically in every consumer.

## Contracts and limits

- Verify participant paths against canonical repository identity; exclude unrelated slots/repos and review checkouts. Deleted roots, malformed history, missing telemetry, and stale reads remain visible uncertainty.
- Print actual start/end, code/tracker refs, scope, participant roots, excluded interval/source reasons, source/event counts, and quality. Detailed segment accounting stays in active-time.
- Whole-issue adoption warns when actual/estimate < 0.1 or > 10 with at least 0.5h difference. These are review heuristics, not proof of error. Do not compare milestone increments with a whole-issue estimate. Incomplete required evidence refuses auto-adoption independently of ratio; preserve existing explicit judgment escapes.
- Reuse the 5,000-record history ceiling and stop when relevant issue history is resolved. Hitting it means unknown, not empty. Honor cancellation; stream metadata before large payloads. No background work outlives measurement, and no persistent transcript cache is added.
- No new status/lifecycle enum or sandbox. Update modeled actual_hours writer ownership only if correction introduces a writer. Diagnostics return/print; bounded correction evidence uses existing tracker receipts/history, never published raw transcripts.

## Measured correction

Propose `sdlc issue correct-actual --issue N --reason TEXT`, with dry-run preview. It computes the value, accepts no hand-typed hours, requires a completed close generation, and preserves status/reviewed head/landing/original close evidence. Record old/new actual, input/algorithm provenance, and reason, bound to that generation.

Pin tracker generation and calibration inputs before effects. Tracker CAS and calibration append are separate effects: use existing receipts/recovery, persist confirmed progress, and make retries converge without remeasuring or appending duplicates. Generation change refuses; uncertain publication reconciles remote evidence. Reuse append-only calibration supersession/NewestPerIssue; sweep all readers so invalid superseded rows cannot re-enter through pre-filtering. Refresh an archived card mirror through the existing narrow publication path, never another live slot's tree.

After #335 lands and the shared binary is rebuilt, inspect dry-run segment evidence and correct pair#454/#514 through the owning verb. Use fixed measurements, not 2.57h or a reset-window approximation. If historical provenance cannot be established, explicitly exclude the invalid sample and report what evidence repair needs.

## Tasks and verification

- [ ] **Reproduce:** portable histories for move :1→:0→:1 with a persistent source-rooted session, and claim→release→other work→claim with an old child crossing the new claim. State expected minutes independently, preserve early shaping, and observe RED before production changes.
- [ ] **History/source rules:** implement pure transitions and eligibility. Sequence/property tests exercise repeated claims, transfers/releases, missing/truncated history, and lineage. Individually mutation-check release clipping, moved-source retention, and old-child exclusion.
- [ ] **Consumer parity:** drive actual, whole close, milestone close and explicit brain-dir through production temporary-repo entry points. Compare hours/provenance for identical pinned inputs before/after movement and landing; concurrent unrelated slots stay excluded. Mutation-check each consumer's wiring separately.
- [ ] **Adoption:** pure ratio boundary tests and production omitted-actual warning/refusal tests. Update README, actual/active-time/close/milestone help and `atlas/workflow/sdlc-binary.md`.
- [ ] **Correction:** test dry-run/no writes, generation races, interruption after each effect, retry/no duplicate row, and unchanged close binding. Update help/README/atlas and modeled field ownership; verify effective calibration/split readers use corrected evidence.
- [ ] **Ship:** targeted checks then committed/frozen-tree background `make -k test`; close on SHIP, PR/CI/merge, real-hash report to hub.
- [ ] **Repair data:** after landing, fixed-engine replay and owning-verb corrections for pair#454/#514; card/archive/effective calibration agree. Peer code/config remains separately owned.

Named risky-function strategies: engagement derivation uses sequence/property tests; provenance eligibility uses adversarial metadata and dispatch-time intervals; event clipping preserves interval union and rejects released gaps; input resolution uses actual moves/landing in real repositories; adoption is pure boundary testing plus production guards; correction uses stateful Git/receipt interruption tests. Historical opt-in replay supplies conformance, distinct from synthetic expected-minute fixtures.

Targeted commands: `go test ./cmd/sdlc/internal/tracker ./cmd/sdlc/internal/transcripts ./cmd/sdlc/internal/activetime -count=1`; `go test ./cmd/sdlc -run 'TestActual.*(Move|Release|Provenance)|TestClose.*Actual|TestActualCorrection' -count=1`. Full: `make -k test`; avoid the serial whole cmd/sdlc suite.

## Architecture lenses

ARCH-DRY: one resolver/engine/correction authority. ARCH-PURE: typed event/window/eligibility/comparison rules. ARCH-PURPOSE: both failures, all consumers, actual data repair. ARCH-MOCK: file-backed repos/transcripts and opt-in historical conformance. ARCH-CONSTRAINTS: bounded issue history and metadata-first reads. ARCH-SECURE: verified identity/provenance and explicit ambiguity. ARCH-ORDER: modeled ownership transitions, pinned generations, interrupted/repeated effects. ARCH-FUNERAL: no cache family; bounded receipts/correction rows use existing retention and Git history.

## Approval checkpoint

Proposed to the hub; no implementation or estimate yet. Enter change-code after direction approval and derive the estimate after plan-quality. No Mx tags: one atomic implementation review boundary.
