# SDLC usage implementation plan

> **For agentic workers:** Consult AGENTS.md Section3 for bounded delegation; otherwise use superpowers-executing-plans. Steps use checkboxes. SDLC owns plan-quality and the single close review.

**Goal:** Report local tokens and burn by slot/agent/model, whole-system issue cost including reviewers, and optional empirical account-meter exhaustion.

**Architecture:** Normalize response usage before deduplication, time filtering and role/ownership attribution. Reuse #335 ownership/source lineage, extend existing gate evidence with explicit reviewer bindings, and keep aggregation/calibration pure. Filesystem/Git/CLI persistence stays in adapters.

**Tech stack:** Existing Go/Cobra, encoding/json, time, bounded Git/fleet/transcript and fake reviewer seams. No provider API, price table or agent UI interaction.

## Evidence and approach

Read-only inventory:2725 Codex JSONL files/4.87GB,5140 Claude files/2.08GB; largest rollout276MB. Bounded recorded samples found modern token_usage_record plus legacy cumulative token_count, and repeated/revised Claude assistant usage. Label output **reported local response usage**, not complete provider quota consumption.

Recommend one normalized usage package and shared attribution adapters. Parsing independently in commands duplicates counting/coverage rules. A persistent transcript index reduces IO but adds invalidation and retention machinery; defer until measured streaming latency warrants it. No cache writes in the initial read-only command.

Hub receipt efaa1145-7a99-4c4d-91bc-88b4d1f5858c requires worker tokens, reviewer tokens with rounds, and total. Unbound sources remain visible in a global unknown-slot bucket. Meter conversion needs two compatible observations, never one guessed credits/token rate.

## Core concepts

### Pure entities

Paths below are relative to cmd/sdlc/. Status is relative to the issue's pinned main base.

| Name | Lives in | Status |
|---|---|---|
| Counters, response identity, coverage | internal/usage/types.go | new |
| Codex response/cumulative transitions | internal/usage/codex.go | new |
| Claude response/aggregate and counter decoding | internal/usage/claude.go; decodeCounters in internal/usage/read.go | new |
| Dedup/window/slot/model/hour aggregation | internal/usage/aggregate.go | new |
| Meter validation, transition and projection | internal/usage/calibration.go | new |
| Reviewer Intent/Record/Event, Validate, Step and Advance | internal/reviewusage/model.go | new |
| Reviewer CLI identity protocol decoding | internal/reviewusage/identity.go | new |
| Accounting artifact validation and quarantine decoding | reviewer_accounting.go; internal/reviewusage/store.go (ParseAccountingEvidence) | new; pure decoding functions alongside IO adapters |
| Issue ownership interval derivation | internal/tracker/engagement.go | reused unchanged |
| Session/source lineage eligibility | internal/transcripts/provenance.go | reused unchanged |
| Immutable Codex executing identity | internal/transcripts/codex_identity.go | new; pure shared parser |
| Slot, worker-window and role selection | usage_inputs.go; usage_reviewers.go; usage.go | new; pure selection functions share files with IO adapters |
| Optional reviewer records on gate ledgers/rounds | internal/gatestate/ledger.go | modified |

Counters use checked int64 arithmetic and dimension presence. Canonical input includes cache reads/writes; output includes reasoning. Unsupported dimensions are null. Normalize and globally deduplicate responses before windows or attribution; aggregate observations diagnose missing spend rather than supplying duplicate totals. Response model and immutable session/dispatch facts determine attribution.

### Integration points

| Name | Lives in | Status | Wraps |
|---|---|---|---|
| Usage source enumeration/streaming | internal/transcripts/usage_discovery.go | new | JSONL/meta files and harness registry; actual's root-only discovery unchanged |
| Bounded normalized-record reader | internal/usage/read.go (Decode) | new | io.Reader, cancellation and line limits; delegates to pure record transitions |
| Actual root metadata adapter | internal/transcripts/discovery.go | modified | same pure Codex identity parser; root-only discovery scope unchanged |
| Command, collection and rendering | usage.go | new | Cobra/clock, fleet/source adapters, pure aggregation and rendering |
| Shared issue source selection | actualinputs.go; actual.go; usage_inputs.go | modified/modified/new | pinned tracker history and ownershipSourceWindows |
| Reviewer attempt capture | internal/judge/dispatch.go; internal/judge/stream.go | modified | existing Run process/onStart/return seam |
| Accounting receipt store | internal/reviewusage/store.go; internal/gitx/recoveryref.go | new/modified | bounded CAS Git RecoveryData, admission lock and cleanup |
| Reviewer attachment, artifact persistence/read and binding | reviewer_accounting.go; usage_reviewers.go; boundaryledger.go; changecode.go; milestoneclose.go; reviewsidecar.go | new/new/modified/modified/modified/modified | workspace/Git identity, ledger/sidecar IO and pure typed binding |
| Meter store/process probe | internal/usage/meterstore.go; internal/usage/meterprocess.go | new | private versioned atomic file, context-aware lock and process liveness |
| Registration/docs | main.go; helptext/usage.md; README.md; atlas/workflow/sdlc-binary.md | modified/new/modified/modified | embedded help and operator docs |

Reuse fleet.GitReader/stateful Git fake and portable real repositories. Transcript integration fixtures are temp folders with sanitized recorded JSONL and metadata, not function-call mocks. Gate capture uses existing fake agent Run tests plus recorded protocol preambles. Meter tests use portable state directories, never real account observations.

## Counting contracts

### Codex

Prefer token_usage_record per-response usage keyed globally by agent+response_id. Resolve executing thread_id and match turn_context.model by exact turn_id. Child session_id can name the root: use the first session_meta.id/thread_id as file identity; copied later metadata cannot replace it.

Do not add cumulative thread/turn counters, token_count or compacted.latest_token_usage_record to modern response spend. Embedded compaction evidence dedups by response identity; if its original is absent, admit explicit usage once with a reconstruction diagnostic. Modern mixed files do not get supplemented by cumulative totals; expose uncovered segments.

Legacy-only files difference cumulative total_token_usage in file order. Read pre-window baseline before filtering. Equal snapshots add zero; missing initial baseline leaves an explicitly unmeasured segment. Decreases/reset or inconsistent dimensions re-establish the observed baseline and mark the lost segment. Never sum last_token_usage or fabricate a reset value. Accepted deltas use observation time and observed model; missing/conflicting model stays unknown.

### Claude

Key by message.id, requiring consistent requestId/model. Record UUID/block index is not spending identity. Prefer consistent terminal usage; within one response, monotone output revisions replace partial counts rather than add. Conflicting terminal counters are unmeasured. Use the selected completion timestamp for hour attribution. Partial-only responses remain provisional reported counts and mark incomplete coverage.

Child discovery is opt-in; actual's root-only default remains. Logical identity is parent session plus agentId. Adjacent .meta.json toolUseId binds to explicit parent Task/Agent dispatch ID/time, then inherits parent slot/engagement through that proof. Missing/cyclic/conflicting lineage is unknown. Task duration does not supply tokens. cost-state/modelUsage and task-notification aggregates only diagnose coverage; do not add them to response totals or invent missing splits/timestamps.

Primary samples:5379 Claude usage records represented2534 response IDs; partial child responses and aggregate-only model activity remain gaps. Codex20 rollouts had5812 response records and22 identical compaction copies. Sanitized Codex evidence is /tmp/ariadne365-codex-evidence.json. Claude cache semantics: https://platform.claude.com/docs/en/build-with-claude/prompt-caching.

## Attribution and issue scope

Default fleet mode maps root sessions by verified checkout/home-slot membership, including dependencies; bound descendants inherit source home slot. Ordinary/deleted/ambiguous roots are unknown. Enumerate relevant local sources beyond matched fleet cwds so observed unbound spend appears in unknown-slot groups. Use membership-only fleet lookup seams rather than default tracker-fetch/claim warming; enumerate once per report.

Issue mode factors the ownership/source phase beneath actual's full resolver. Reuse EngagementHistory, DeriveEngagements, ownedSourceWindows and EligibleSources. Do not require a code commit for planning token measurement, and do not reuse actual's pruned contributing SourceIDs as the complete token source set. Worker responses must occur in admitted ownership intervals; descendants remain bound to original dispatch engagement, including moves/reclaims/brain sources.

Active right edge is captured report time. Completed worker right edge is the exact completion evidence commit time, including work up to close, not today's HEAD or stale started. Show actual active time with its own provenance/window because that metric deliberately excludes final verification tail. Explicit --since intersects issue intervals and reviewer windows; omitted --since means the whole issue, including old bench runs.

Reviewer cost requires typed dispatch binding: repository/issue/recipe/round/attempt, agent, recorded root, session identity when available, dispatch/finish timestamps, reviewed head and dispatching slot. Retries are distinct; bound descendants inherit reviewer role. Boundary roots are unique disposable roots; in-place plan/estimate reviewers require exact CLI-emitted session ID. No cwd+time guess for in-place review.

Current ledger/sidecar types do not retain roots, so add optional evidence and preserve legacy rendering when absent. Capture Claude session_id from stream-json init/result and Codex UUID from the bounded startup metadata preamble BEFORE prompt/transcript text, keeping argv/permissions unchanged. Unknown protocol identities produce binding diagnostics. Unique boundary roots can still bind without session ID; in-place absence cannot.

Persist all launched attempts, including protocol failure/retry/cancellation, through existing gate evidence paths without advancing verdict/findings state. Preparation without reviewer launch creates no charged attempt. Historical missing bindings are counted limitations; their unbound spend remains globally visible. #365's own prelanding reviews may remain legacy/unbound because it must use the shared gate binary; never install candidate gate code to hide that gap.

## CLI and JSON

- sdlc usage --since 6h, default24h; RFC3339 also accepted. Capture now once; half-open [since,until). --burn-window defaults3h and accepts1h..3h. Rates divide the selected recent window's elapsed WALL hours, not active-time hours. UTC hour buckets identify partial duration.
- sdlc usage --issue N prints worker, reviewer with rounds/attempts, measured total, actual active time and coverage; agent/model breakdown remains. Missing evidence cannot look like a complete benchmark total.
- --json returns schema_version1, scope/window, stable-sorted groups, hourly buckets, recent rate/window, issue role/round totals and optional active time, meter projections and coverage. Integers for counters; null for unsupported dimensions/projections; explicit completeness/status. No prompts, message bodies, raw stderr or secrets.
- sdlc usage calibrate --agent codex --remaining 29276 --unit credits records one timestamped operator observation. Known parser agents; credits/percent units (percent0..100), finite nonnegative values. A meter increase starts a new epoch with diagnostic; units never mix.

Coverage counts sources/records, accepted responses, duplicates, revisions, provisional responses, malformed/skipped records, unsupported formats, unknown slot/model, missing reviewer binding, baseline/reset gaps, truncated tails and unavailable sources. Retain counts and at most3 bounded examples per class.

## Meter state and projection

At least two strictly ordered same-agent/unit observations. Measure that agent's local reported tokens across ALL known/unknown-slot sources between observations. Empirical units_per_token=(remaining_before-remaining_after)/reported_tokens_between; units/hour=units_per_token*recent_agent_token_burn; remaining/units_hour yields projected exhaustion. Label it a local observed-mix estimate, not provider pricing or guaranteed quota deadline.

Single point, stale pair (>7days), meter reset/nondecrease, zero token denominator/current burn, or incomplete required local evidence yields explicit status and null exhaustion. Provisional/aggregate discrepancy coverage remains visible and prevents confident projection. Issue-only burn cannot project an entire account; issue mode marks that projection not applicable.

Pure transitions: empty+point→one; one+later lower same-unit point→pair; pair+compatible later point→latest pair; exact repeated point→no-op; earlier/conflicting same-time→refuse; increased meter→new one-point epoch. No stale/conflicting overwrite.

Store versioned bounded observations at XDG_STATE_HOME/sdlc/usage-meter.json, fallback ~/.local/state/sdlc. At most two per known agent/unit,64KiB total. Lock via existing context-aware seam; same-dir unique temp,0600, sync/close, rename, deferred cleanup. Cancellation/failure preserves prior state. Latest-pair replacement bounds residue; no credentials/API keys. usage is read-only; calibrate alone writes meter state.

## Operating envelope and lifecycle

Batch CLI, not UI keystroke path. Initial target≤60s cold on the measured7GB store, one reader at a time; verify with real smoke, do not claim speed beforehand. Stat-old files may be skipped for recent windows only where mtime proves no newer append; selected files still need pre-window baseline/model metadata. Historical issues retain needed older sources.

Enforce caller cancellation and60s report deadline; max16384 candidates,4MiB JSONL line,250000 accepted response identities, checked arithmetic. Bound exceedance means incomplete coverage and no projection; cancellation exits nonzero. Capture stat-sized read prefix, exclude concurrent appends, report truncation/mutation. Reuse metadata parsing without hashing all7GB to discover sources; hash selected prefixes only where provenance needs it. No silent recent-file or timestamp guess.

No usage cache. In-memory counters die with command. Meter latest pairs bound file residue. Reviewer metadata is permanent≤2KiB/attempt, at most two normal attempts/round; existing issue archival carries it and Git retains history. Validate writer bound. ARCH-DRY shares ownership/membership/normalization; ARCH-PURE/ORDER types transitions; ARCH-SECURE bounds untrusted evidence; ARCH-MOCK uses portable state/Git/fake agents; ARCH-PURPOSE includes reviewers/unknown spend; ARCH-CONSTRAINTS/FUNERAL govern work and residue.

## Verification by named risky function

| Delivered function/seam | Adversarial class and mechanical proof |
|---|---|
| codexState.step / response / finish | modern+compaction copies, forks, stale/duplicate legacy totals, reset/no baseline; sequence/fuzz and mutation summing last snapshots |
| decodeClaude / NormalizeResponses | repeated blocks, monotone partial→terminal, conflicting finals, absent dims; recorded fixtures, property and mutation adding block usage |
| AggregateUsage | dedup across files before boundary selection, model switches, hour edges, overflow, unknown groups; pure properties and mutations wrong denominator/drop unknown |
| resolveUsageAuthority / usageWorkerWindows / usageSessionSlots | moved roots, release/reclaim, dependencies, brain/Claude children, planning-only/completed issue; real Git/temp source conformance, mutation bypass owner or revive old dispatch |
| attachReviewerAccounting / reviewusage.Step / Validate / bindUsageReviewers | exact identity, retries, cancellation, failed protocol, in-place ambiguity; production dispatch→persistence→usage regression, mutate omitted binding persistence |
| ObserveMeter / ProjectMeter | reorder/duplicate/unit mismatch/reset, one-point/zero/stale/partial; transition sequences and mutations one-point rate or wrong token interval |
| RecordMeter | competing writers, interruption/truncated state/temp cleanup; portable state directory tests proving no lost pair/prior-file corruption |

## Tasks — one atomic close boundary

### Task1: normalized core

Files: new internal/usage/{types,codex,claude,aggregate}.go, colocated tests and sanitized testdata/{codex,claude}.jsonl.

- [x] Write recorded-format and counter/revision sequence regressions; run go test ./cmd/sdlc/internal/usage -count=1 and observe RED.
- [x] Implement typed transitions, checked counters, dedup-before-window and hourly/burn aggregation.
- [x] Run properties/fuzz for duplicates/model/reset/partial/overflow/boundaries; kill last-snapshot and block-summing mutations; commit evidence.

### Task2: attribution and gate evidence

Files: internal/transcripts/usage_discovery.go/tests (reuse provenance.go unchanged); usage_inputs.go/tests and actualinputs.go/tests; internal/reviewusage/{model,identity,store}.go/tests; internal/gatestate/ledger.go and ledger/parser tests; internal/judge/{dispatch,stream}.go/tests; reviewer_accounting.go/tests, usage_reviewers.go, boundaryledger.go, changecode.go, milestoneclose.go, reviewsidecar.go and command tests.

- [x] Add metadata/fork/Claude-meta child, unknown cwd, copied history, partial-tail, append/truncation fixtures; observe RED.
- [x] Add production gate-dispatch→binding→usage regression for boundary/plan/estimate; include retry/failure/cancellation with no verdict advancement.
- [x] Implement shared selection and optional bindings; preserve existing actual/gate/default discovery semantics.
- [x] Verify moves/reclaims/dependency/brain/old reviewer/planning-only/completed edges; kill missing-binding, eligibility-bypass and child-doublecount mutations; commit.

### Task3: command, meter and docs

Files: usage.go, usage_command_test.go, usage_attribution_test.go; internal/usage/{calibration,meterstore}.go/tests; main.go, helptext/usage.md, README.md, atlas/workflow/sdlc-binary.md.

- [x] Write human/JSON command and pure meter transition/projection tests plus portable store failure/concurrency; observe RED.
- [x] Implement flags, optional empirical projection and bounded observations, keeping usage read-only.
- [x] Verify per-issue role/round totals, null dimensions, skipped coverage, zero/stale/reset behavior and no transcript/UI mutation; kill one-point/rate/unknown/reviewer omission mutations.
- [x] Document JSON semantics, coverage limits, supported schemas, calibration location and reset/unit behavior.
- [x] Targeted package/command checks plus recorded-fixture conformance; real read-only --since6h/--issue smoke, record duration and limitations; commit.

### Task4: acceptance and landing

- [x] Push exact code HEAD; sdlc local-check run --name suite; inspect passing record. All direct make invocations use make -k.
- [ ] Shared sdlc close with fixture/mutation/smoke evidence; fix findings by class to final SHIP.
Publication follows via sdlc pr and sdlc merge --yes with green CI/exact-code suite; report PR and git-rev-parse hashes to hub. Shared rebuild follows landing.

No estimate before change-code plan-quality acceptance. No Mx tags: one atomic feature/close boundary. Hub approves durable plan before implementation.

## Revisions

### 2026-10-14 — approved fleet glance first

Hub receipt4479fade-4e09-47d1-be7d-806e76f8d261 approves1157b761 and requires the
first report surface to be the TL's fleet glance: one line per agent, tokens in
the last hour, burn rate (label its1–3h denominator), and calibrated projection
or explicit non-projecting status. JSON adds stable agent_glance rows with an
explicit fleet scope and1h window, ahead of detailed groups. This remains a
fleet measure even alongside an issue detail; do not project whole-account burn
from issue-only tokens. Source selection retains recent1h/3h evidence needed
for this glance independently of an explicitly narrower detail --since.

### 2026-10-14 — plan-quality interruption and role refinements

PQ-1: accounting cannot depend on gate finalization. planningreview.go's dispatch
returns before ledger writes after stale/cancelled validation; close does likewise.
Use the existing gitx.RecoveryStore CAS/reachability machinery with a closed
review-usage namespace, not a parallel file journal or acceptance gate. Add a
thin internal/reviewusage receipt store and pure attempt transition model; gate
rounds/sidecars retain optional copies of that same typed binding. This receipt is
necessary specifically to preserve dispatch attribution when normal gate writes
are forbidden. It is nonsemantic and never contributes a gate round or verdict.

Attempt transitions are explicit: Requested(intent/root/issue/recipe/round/attempt)
→ Started(on actual child launch, PID/time) → Finished(on return with exact CLI
identity if available, finish time and outcome). Exact event retries are idempotent;
identity rewrite/out-of-order confirmation refuses. A Requested intent may have
launched before process death, so absence of Started confirmation is uncertain,
not proof of no spend. Persist Requested with CAS BEFORE launching; failure to
save prevents launch. Capture Started via Run's onStart, and finish via the shared
sync/heartbeat return seam; both use this model. CLI failure/protocol retry are
accounted even though no acceptance result is published. Each retry gets its own
attempt ID; no overwrite of the earlier attempt.

Receipt writes use bounded Git CAS in the common dir, outside worktree artifacts
and semantic snapshots. They do not acquire/release the SDLC repo lock while an
external reviewer runs, and do not change plans, ledgers, findings, round caps or
pass-through. Terminal persistence gets a fresh bounded cleanup context (2s), not
an already-cancelled dispatch context. If a terminal write fails, retain known
partial receipt and expose incomplete attempt coverage; never authorize another
review or replay a provider call. Hard process death leaves Requested/Started:
usage is read-only and counts only explicit session/unique disposable-root proof,
reporting unfinished status and unknown coverage where in-place ID was not captured.
It never guesses a finish time/session or declares success. Normal gate artifact
copies happen only after their existing safe reacquire/validation.

Ref retention: each attempt payload≤2KiB; max4096 pending accounting refs per repo
(~8MiB payload). Before new dispatch, cleanup compare-deletes only records whose
identical binding is proven committed and reachable in the issue's ledger/sidecar
on a branch/main; uncommitted or stale receipts stay. A full store refuses BEFORE
launch with the retained refs and publication/cleanup next action. Git keeps small
historical object overhead; no automatic deletion of uncertain attempts. Add
namespace isolation/cap/CAS tests. Tests drive cancellation, stale validation,
lock reacquire failure, death-equivalent Requested/Started and retry: binding
retention must hold while findings/round count/reviewed anchor remain unchanged.

PQ-2: role selection is exclusive, after response dedup. A single valid explicit
reviewer binding takes precedence over ownership-based worker eligibility,
including in-place plan/estimate roots. A reviewer root plus bound descendants
inherits that role; subtract it from worker candidates. Conflicting issue/round
or session bindings put the response in unknown attribution with a counted
conflict, never both roles. Only otherwise can shared ownership admit a worker.
For every complete report, worker+reviewer equals the union of accepted issue
response identities exactly once. Production ResolveUsageSources→AggregateUsage
regression must include an owned in-place reviewer and fail when precedence is
removed; do not rely only on a pure role selector test.

PQ-3: opt-in live identity conformance test (WF_LIVE_USAGE_CONFORMANCE=1) compares
supported Claude/Codex Run startup identities with a recorded transcript identity
through the same adapter. Run on supported CLI upgrades and before a benchmark
batch; no background scheduler. Offline tests use sanitized real preambles and
stateful fake processes. Missing/changed protocol remains coverage and cannot
silently create a guessed binding. The usage command never launches an agent.

Core concept additions: PURE Intent/Record/Event/Step/Advance in
internal/reviewusage/model.go; INTEGRATION Store in
internal/reviewusage/store.go, wrapping gitx.RecoveryStore. Existing judge and
round/sidecar adapters consume that single typed evidence. Task2 owns these files
and gitx/recoveryref.go closed-namespace extension with colocated tests.

### 2026-10-14 20:40 — resume checkpoint

Task1–3 are complete per the issue's implementation, conformance, mutation and
smoke checkpoints; mark their execution rows accordingly. Task4 is the remaining
acceptance/landing work after the operator freeze. Main integrated without
conflicts, and root/internal targeted regressions pass on the merged code.
The issue checklist now separates verified implementation from the subsequent
SHIP/PR/merge operations, preserving their order without premature completion.

### 2026-10-14 — recorded acceptance evidence

Task4 suite step passed on081417175e5b47f9dd779f1747fe275f2957aab3 (1389 root
tests/61 packages; zero failures). Live Claude/Codex conformance and refreshed
fleet/issue smokes pass. Publication is tracked by the eventual landing record
rather than a pre-close checkbox that would prematurely claim a merged PR.


### 2026-10-14 21:15 — close round1 classes (BR-1 through BR-4)

BR-3 reconciles the Core concepts tables and task paths above with the pinned
implementation: provenance.go and ownership derivation are reused unchanged;
reviewusage/model.go owns the pure attempt model, reviewusage/store.go owns IO,
and usage.go owns rendering. No gatestate/reviewer.go or usage_render.go exists.
This replaces planned locations with delivered ownership (ARCH-PURE/ARCH-DRY).

BR-1 requires reconciliation of every applicable modern/legacy cumulative
observation, including mixed format and reset/high-water evidence, before
spend_complete can authorize a meter denominator. Snapshots remain coverage-only.
BR-2 requires one validated artifact decoder for ledger and comment readers;
invalid or unreadable evidence must degrade coverage, and usable reviewer
identity must quarantine that session from ordinary worker attribution.

BR-4 reopens Task2's production integration row: existing decoder, direct judge,
serialization and pure attribution tests do not prove production attachment.
Add real gate attachment→fake CLI→receipt/artifact→usage-loading regressions for
boundary/plan/estimate, including interruption without semantic advancement;
removing production attachment must fail them. Rerun the recorded suite after
these code changes, then close to SHIP. Previous passing evidence remains an
accurate record for its earlier head, not evidence for the pending fixes.


### 2026-10-14 21:16 — round1 regression evidence

BR-2 reporting now uses Store.AccountingEvidence for private receipts, retaining
partial identity plus errors while strict gate-facing List/Load remain unchanged.
A shared artifact decoder validates ledger/comment records and cleanup accepts
only valid exact records. Invalid identity is quarantine evidence, not attribution
authority. Private/ledger/comment version99, strict shape/duplicate-key failures,
malformed artifacts, invalid timestamps/root evidence and cleanup guards pass;
root targeted2.369s and reviewusage package3.113s after observed RED.

BR-4 production tests now exist in reviewer_accounting_integration_test.go:
boundary dispatch verifies disposable identity and exact sidecar copy; production
change-code verifies plan/estimate receipts and exact ledger copies through usage
loading and worker0/reviewer20/total20 with one round/attempt. Cancellation after
observed Started retains Finished/canceled accounting without semantic ledger or
HEAD movement. All three tests pass across three runs15.261s. Independent Go
overlays omitting boundary/plan/estimate attachment each fail the intended
missing-binding assertion (1.179s/2.389s/2.590s). Initial fixture agent-selection
and duplicate-ID errors were corrected; every fallback CLI is now a failing local
guard. Task2's reopened production row is complete on this evidence.


### 2026-10-14 21:57 — complete cumulative-evidence class

A final BR-1 sibling sweep found turn_token_usage in the recorded Codex fixture
but absent from reconciliation. An external production-decoder probe proved that
one turn's missing spend stayed complete and another turn's responses could mask
it. Decode optional turn totals, retain per-turn dimension maxima, and compare
with response-ID-deduplicated sums from the same turn, never thread-wide spend.
Invalid present turn totals degrade coverage; absent/null totals are neutral.
Eight defective cases failed before the fix; four controls already passed; all12
new cases and the usage package pass0.165s after the fix. This closes the named
class across legacy, thread and turn cumulative evidence without adding charges.

The f3bfa3ed suite waited1959s and then ran until superseded by this concrete
coverage fix; it was cancelled, not passed. Parent local-check SIGINT exited130
but left Make/Python running; verified own disposable-root cwd and process group
81248, then terminated that group and verified it was gone. A fresh suite must
record the corrected code. Refreshed fleet/issue smokes on the previous fix head
completed25.02s/23.03s and correctly exposed coverage/projection limits.


### 2026-10-14 22:49 — corrected acceptance pass

Recorded sdlc-local/suite succeeds on3679928075e6e17e5448ffb355fe3324886ee2ab:
1396 cmd/sdlc tests/61 other packages, zero failures,649.8s (10m50s), after2392s
queue wait. Log SHA256 prefixec5f77327432, hostMac.localdomain. All12 shards and
packages exit0. Final read-only fleet/issue smokes on this code pass40.92s/52.30s
under load; both stay within60s, expose incomplete coverage, and keep absent-meter
projections null. Earlier live CLI identity conformance remains applicable: no
identity protocol or provider argv changed in the review fixes.

BR-3's concept-map reconciliation also includes the reporting-only accounting
decoder and replaces conceptual function labels in the verification table with
delivered functions/seams. These are documentation-only corrections; the passing
suite's code is unchanged. Shared close remains the final review authority.


### 2026-10-14 22:56 — round2 identity and role-exclusion classes

Shared review verified/disposed BR-1 through BR-4, then reproduced two new
Critical findings. BR-5 requires a single immutable first Codex session-metadata
identity contract across discovery and decoding: id/thread_id aliases, conflicting
aliases, copied later metadata, and malformed/empty first metadata. Share a pure
parser; first evidence cannot be repaired from later copied history (ARCH-DRY).

BR-6 requires all relevant reviewer identity evidence before exclusive role
selection. File/repository issue filters must not hide a different issue's known
in-place reviewer and turn it into this issue's worker. Load bounded fleet
identity evidence, then filter included reviewer totals after role selection;
verify ledger/comment/private receipt parity before and after committed cleanup.
Keep selected-issue gap coverage distinct from unrelated historical gaps where
possible, and keep scan bounds/unreadable evidence explicit (ARCH-PURPOSE).

Task2's verification scope now includes production discovery→Decode identity
conformance and production loader→role attribution with cross-issue/cross-repo
persistence/cleanup. The passing36799280 suite remains true for its code, but
these behavior fixes require a fresh recorded suite before the next close.


### 2026-10-14 23:02 — round2 fixes verified

BR-5 adds pure transcripts.CodexExecutingID, shared by actual discovery, usage
discovery and usage decoding. It accepts id/thread_id fallback, requires equality
when both are present, ignores ancestor session_id, and locks the first metadata
result including errors. Malformed JSON before identity cannot be repaired by a
later copied header. Twenty production cross-consumer cases pass; restoring the
previous decoder/reader reproduces discovered child with132 tokens and complete
coverage instead of12 (mutation fails0.166s). Usage/transcripts packages pass
0.321s/0.382s; existing actual/provenance/ownership regressions pass103.888s after
freezing edits (the earlier attempt's guard correctly detected concurrent edits).

BR-6 retains all bounded artifact/private identities across fleet repositories,
current root first. Exclusive role selection precedes issue inclusion. Legacy
identity-gap diagnostics remain issue-scoped; unreadable possible exclusions and
the4096-artifact bound stay explicit. Real cleanup lifecycle tests fail before
fix for ledger/comment and cross-repository exclusion, then pass5.067s; broad
usage, production dispatch, binding and accounting regressions pass21.340s. No
aggregate counter or ownership interval behavior was changed. Fresh suite follows
current-main integration; previous pass is not reused for changed code.


### 2026-10-14 23:11 — repository-qualified diagnostic scope

The final issue smoke exposed a BR-6 diagnostic sibling: after all fleet identity
loading, legacy gaps rose4→14 because unrelated repositories' issue365 matched
an integer-only scope. A real-Git overlay with distinct origins proved2 gaps
instead of1 for the selected issue. Use explicit usageReviewerScope{Repository,
Issue} and shared read-only resolveUsageRepository (extracted unchanged from
ownership resolution), never positional root guesses. All exclusion identities
remain loaded; only legacy-gap diagnostics are repository/issue qualified.
Lazy root identity failures remain explicit and resolve once per root. Tests
cover distinct origins, reversed root order, fleet mode, alternate worktrees,
missing scope identity and cached failure; focused4.719s/broad45.998s pass after
RED. Superseded b9d5da55 suite was still queued and cancelled130; verified parent
68943 and its own acquire-child73249 were gone. No passing result is claimed for
that head. Fresh suite follows final corrected diagnostic code.
