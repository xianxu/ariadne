# Qoder transcript measurements Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make recorded Qoder work visible to usage, focused actuals/close adoption, and invoking-agent identity, preserving native credits and unknown token dimensions.

**Architecture:** A bounded, pure Qoder record adapter supplies the existing transcript, usage, and agent-id seams. Usage gains nullable native measurements and exact credits through its existing global response normalization and attribution path; actuals retain #335's ownership/prefix model. Session recognition remains separate from reviewer CLI dispatch capability.

**Tech Stack:** Go, encoding/json, bounded exact decimal arithmetic using math/big internally, local JSONL stores, existing portable Git/home fixture harnesses.

---

## Evidence and scope

Read-only inspection of the operator-supplied pair:4 sample found 883 rows,
313 assistant events, and 117 response IDs. Each response has one terminal
usage row. Top-level UUIDs identify streaming events; message.id identifies
the response. usage.request_id is distinct evidence and is not a session ID.
Credits are per-response, not cumulative: their exact sum is
45.938730726457140528. All four token counters are zero placeholders on
positively billed responses. original_credits equals credits in this sample,
but is not another charge. The declared model is auto; do not infer a backend.

workspace-directories supplies an untimed list of roots and sessionId.
runtime-config supplies epoch-millisecond timestamps, model, and nullable
reasoningEffort. User/assistant events use ISO timestamps. Resumption can replay
older configuration records. parentUuid is message ancestry, not session lineage.

One atomic issue, one close/SHIP boundary. Read transcripts only: do not invoke
Qoder, upload transcript content, mutate its store, alter the review sandbox,
or add Qoder reviewer dispatch in this issue. Qoder-native credit burn is in
scope. Extend the existing optional meter to a credit denominator only through
its two-observation contract; do not invent token-equivalent spend.

**Unresolved before identity implementation:** Pair pins --session-id and exports
PAIR_AGENT/PAIR_SESSION_ID, but those also reach non-agent panes. The hub has
been asked to measure a native Qoder tool-shell marker or approve an explicit
launch-bound identity contract. Do not invent a QODER_* variable or select a
recent transcript by cwd. Record the measured binding and its tests here as a
revision before implementing Task 4. Independent tasks do not depend on it.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| QoderRecord / SessionState | cmd/sdlc/internal/qodertranscript/record.go; session.go | new |
| CreditAmount | cmd/sdlc/internal/usage/credits.go | new |
| Response / Group / Hour / Report native measurements | cmd/sdlc/internal/usage/types.go | modified |
| NormalizeResponses / AggregateUsage | cmd/sdlc/internal/usage/aggregate.go | modified |
| Qoder response decoder | cmd/sdlc/internal/usage/qoder.go | new |
| Projection credit denominator | cmd/sdlc/internal/usage/calibration.go | modified |
| Observed agent recognition | cmd/sdlc/internal/agentid/env.go; agentid.go | modified |

- **QoderRecord / SessionState:** one typed interpretation of identity,
  workspace, activity timestamps, and model/effort configuration. One session
  owns many records and config observations. No transcript IO or usage-package
  dependency in this leaf package: transcripts already imports agentid, so
  agentid must not import transcripts. Usage converts raw numeric credit
  evidence into CreditAmount. Metadata/model extensions belong in this adapter,
  rather than three consumer-specific JSON decoders (ARCH-DRY, ARCH-PURE).
- **CreditAmount:** immutable canonical decimal value, constructed only by a
  bounded JSON-number parser. Exact addition, comparison, validation, and JSON
  number emission; no float accumulation and no scaled-int64 overflow. A response
  optionally owns one credit amount. This is native reported credits, not a
  pricing model or general currency framework. Approximate rates are calculated
  at the presentation/calibration boundary after exact sums.
- **Native measurements:** Response.Tokens becomes optional; Credits is a
  separate optional CreditAmount. Group/Hour/Report totals and token burn are
  optional too. Known token/credit subtotals are explicitly labelled alongside
  completeness, so a mixed-agent subtotal cannot masquerade as a complete total.
  Existing cache/reasoning subset semantics remain unchanged. Response equality,
  monotonic provisional revisions, validation, and checked aggregation cover
  both measurements. One response is normalized once before any window/slot/
  issue selection; the same selected response feeds both columns.
- **Projection:** retain token-denominator projections for Claude/Codex and add
  a credit-denominator path for Qoder. One projection uses exactly one reported
  native denominator; expose units_per_token or units_per_credit accordingly.
  No token conversion from credits. Existing observation ordering, refill/reset,
  freshness, complete-evidence and two-observation requirements remain authority.
- **Observed agent recognition:** identity parsing can recognize qoder without
  claiming its CLI is dispatchable. Keep reviewer capability validation in the
  judge seam; explicit reviewer choices and existing policies must retain their
  documented behavior. No fallback may silently attempt an unsupported Qoder CLI.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| QoderHarness | cmd/sdlc/internal/transcripts/qoder.go; defaults.go | new | ~/.qoder/projects/*/*.jsonl |
| Evidence/usage metadata integration | cmd/sdlc/internal/transcripts/discovery.go; usage_discovery.go | modified | pinned local transcript prefixes |
| Usage Decode dispatch | cmd/sdlc/internal/usage/read.go | modified | bounded JSONL reader |
| Usage collector / renderers | cmd/sdlc/usage.go; usage_reviewers.go | modified | fleet/issue selection, stdout |
| Qoder invoking-session identity | cmd/sdlc/internal/agentid/detect.go; transcript.go | modified | measured env binding and exact session file |
| Reviewer capability boundary | cmd/sdlc/internal/judge/agent_resolver.go | modified | existing supported reviewer CLIs |
| Meter storage | cmd/sdlc/internal/usage/meterstore.go; meterprocess.go | reused | existing bounded private observation store |
| Actual/close adoption | cmd/sdlc/actualinputs.go; actualmeasurement.go; close_actualdev_test.go | reused | ownership, retained prefixes, measured adoption |

- Harnesses accept injected temporary roots. Discovery uses recorded workspace,
  never reverse-decoding a slug into authority; missing or ambiguous metadata is
  visible uncertainty. Usage retains readable unbound responses in unknown-slot
  coverage; automatic actuals admit only ownership-proven source roots.
- The filesystem is the consumed external surface. Tests use a portable
  temporary home/project folder with persisted JSONL append/replay/truncation
  states, through production readers. Reuse real temporary Git repositories for
  ownership/close fixtures; no provider executable is called. Live conformance is
  an explicit read-only check against the supplied sample after implementation
  and when the Qoder format changes, never a default test's dependence on home.
- Invoking identity reads only an exact, measured session binding. Keep
  SDLC_AGENT_ID precedence and no-session shell fallback; never choose the first
  glob match when identities disagree. Model and effort are metadata, not an
  inference from zero counters or a routing alias.
- Collector/renderer changes enumerate all measurement consumers: fleet glance,
  groups by slot/agent/model, hourly rows, issue worker/reviewer/total lines,
  unbound rows, review spend validation, meter windows, JSON and text output.
  Preserve reviewer exclusion and role precedence; this issue adds no Qoder
  review launch receipt producer.

## Measurement contract

| Input / scope | Tokens | Credits | Coverage / rate |
|---------------|--------|---------|-----------------|
| Qoder sample terminal usage | null | exact reported credits | token rate null; credit rate measured |
| Qoder nonterminal streaming rows | no response charge | no response charge | unfinished response stays provisional/incomplete |
| Claude/Codex measured response | existing counters | null | existing token rate; credit rate unknown |
| Mixed-agent aggregate | null if any included response lacks tokens | null if any included response lacks credits | retain explicitly labelled measured subtotals for both units |
| Native-unit complete empty window | zero in the supported unit | zero in the supported unit | unsupported units remain null, including empty Qoder glance tokens |
| Malformed/missing terminal credits | never fabricated zero | unknown; retain other independently valid evidence | visible spend gap blocks credit-based projection |

Unsupported token dimensions on Qoder are a unit capability, not missing
reported credits. Credit completeness and token completeness must be queried
separately; unknown slot/model does not erase known native spend. Existing
malformed/conflict/truncated/unresolved-reviewer diagnostics remain conservative
for affected projections. Partial counters never become a complete denominator.

Qoder token fields remain null for this documented format even if a future row
contains nonzero fields: without a tested semantic contract it is unsupported,
not silently adopted as Claude accounting. Keep a visible format diagnostic for
unrecognized counter semantics. credits is the authoritative reported field;
original_credits is not added. billable=false or contradictory pricing metadata
is reported as uncertain charge semantics, retaining reported evidence but
blocking a billing projection rather than inventing a refund or charge.

Usage JSON moves to schema_version 2 for nullable totals/rates and credits.
Known Claude/Codex token values preserve their numeric meanings. Unknowns emit
JSON null and a dash with a coverage label in text. The first output remains one
fleet-glance row per agent with last-hour native amounts, native burn, and a
projection or labelled reason. Detail rows have parallel tokens/credits columns;
issue worker/reviewer/total and hourly rows carry both. Document the schema change.

## Ordering, input bounds, and lifetime

SessionState owns transitions; callers cannot mutate authoritative identity or
config state directly. Tests drive its transition function with deterministic
record sequences:

| State + event | Next state / effect |
|---------------|---------------------|
| unresolved + consistent session/workspace records | resolve once sufficient proof exists; usage before resolution remains buffered within bounds or visibly incomplete |
| resolved + identical replay metadata | same identity, no extra activity or charge |
| resolved + conflicting session/root identity | ambiguous/poisoned identity; later good rows cannot silently repair it |
| valid config + older replayed config | retain newer applicable config; do not rewind model |
| configs + event timestamp | select latest valid config at/before event; preserve distinct model epochs |
| same-time incompatible configs | explicit ambiguity, no arbitrary filesystem/arrival-order choice |
| unfinished response + terminal usage | one final reported charge |
| final response + identical final/copy | globally deduplicated |
| final response + incompatible final | conflict; exclude contested charge, record gap |
| prefix read + concurrent append | consume only pinned complete prefix; later run sees append |
| read + replacement/truncation/cancel | visible source uncertainty or cancellation; no successful complete denominator |

Identity/workspace metadata may be repeated and untimed. Resolve the pinned
prefix's metadata before attributing its events; any pre-identity handling must
be explicit and shared by discovery and decoding, not accidentally account a
copied prefix. Ambiguous multi-root sessions are not attributed to an arbitrary
root. parentUuid never creates session lineage.

This is a local batch CLI on Darwin/Linux sharing the developer machine. Reuse
existing 60-second usage context, 16,384-candidate discovery cap, 4-MiB line cap,
250,000-response cap, and bounded diagnostic examples. Identity remains a bounded
exact-session read (existing 1-MiB tail budget plus a bounded metadata prefix if
needed). Enforce bounds before decoding/copying large records and check context
at the command success boundary. Bound numeric lexical input to 128 bytes,
64 significant digits and 36 fractional digits; bound aggregation results too
(e.g. 80 significant digits), with explicit coverage/error on excess. Parse
exponents without allocating powers from unbounded input. This accommodates the
measured 18-digit fractions and >64-bit sums without a floating-point shortcut.
Test exact limits, huge exponents, overflow/precision refusal and cancellation.

No network, background tasks, provider calls, credential creation, or persistent
discovery cache. JSONL content is untrusted: reject malformed/contradictory rows
visibly; never include message/tool content, runtime credential fields, or raw
records in diagnostics/fixtures. Local-store paths are labels, not ownership
proof. Added fixtures are sanitized invented identifiers/paths plus measured
field shapes and nonsecret numeric values. Meter observations reuse the existing
bounded store and replacement lifecycle (two per agent/unit); adding Qoder does
not append an unbounded history. Actual evidence uses existing snapshot retention
and cleanup; do not create a second Qoder snapshot family (ARCH-FUNERAL).

## Implementation tasks

### Task 1: shared typed Qoder records and portable fixtures

**Create:** cmd/sdlc/internal/qodertranscript/{record.go,session.go,record_test.go,session_test.go,testdata/session.jsonl}.

- [ ] Write failing pure tests for mixed timestamp types, stable/repeated versus
  contradictory session/workspace records, missing/multiple roots, stale config
  replay, null effort, model changes and message ancestry without child lineage.
- [ ] Run go test ./cmd/sdlc/internal/qodertranscript -count=1; confirm the new
  behavior fails before implementing the shared adapter/transition model.
- [ ] Implement whitelisted typed records and encapsulated transition state.
  Ignore unknown payload fields without copying them into state or diagnostics.
- [ ] Exercise streamed/terminal rows and malformed/torn input in the portable
  fixture. Run the package tests and commit the passing adapter/fixture.

### Task 2: exact credits and optional native measurements

**Create:** cmd/sdlc/internal/usage/{credits.go,credits_test.go,qoder.go,qoder_test.go,testdata/qoder.jsonl}.
**Modify:** cmd/sdlc/internal/usage/{types.go,aggregate.go,read.go,usage_test.go,calibration.go,calibration_test.go}; affected existing decoder tests.

- [ ] Write failing tests: exact 45.938730726457140528 sum; 18-digit fractions
  beyond scaled-int64 range; credits versus original_credits; null/negative/
  malformed/oversized numbers; billable ambiguity; provisional→final and
  duplicate/conflicting finals across copied files, outside requested windows.
- [ ] Run go test ./cmd/sdlc/internal/usage -run 'Qoder|Credit|Native|Mixed' -count=1
  and record the intended failures.
- [ ] Implement bounded immutable CreditAmount and optional token/credit response
  measurements. Qoder terminal usage yields Tokens:nil, exact Credits and native
  response identity; streaming UUIDs never multiply charges. Refactor existing
  normalization/aggregation once, preserving Claude/Codex cumulative-gap checks.
- [ ] Test the full measurement matrix, per-model changes, mixed/unit-incomplete
  totals, hour/role selection, zero-response supported dimensions, cancellation
  and bounds. Sweep all Response.Tokens and total/rate consumers for nil safety.
- [ ] Extend existing meter projection for --agent qoder with reported-credit
  denominator; two compatible observations are mandatory. Test one-observation
  insufficient, missing credits, stale/future/refill/zero-burn intervals, unit
  selection, mixed-agent isolation and unchanged token-agent results.
- [ ] Run the entire internal/usage package once and commit the coherent core.

### Task 3: discovery, active-time and close adoption

**Create:** cmd/sdlc/internal/transcripts/{qoder.go,qoder_test.go,qoder_integration_test.go}.
**Modify:** cmd/sdlc/internal/transcripts/{defaults.go,discovery.go,usage_discovery.go,discovery_test.go,usage_discovery_test.go}; cmd/sdlc/{actualwiring_test.go,close_actualdev_test.go}.

- [ ] Add failing production-path tests for a temporary Qoder home: discovery
  uses workspace metadata despite misleading slug; absent/ambiguous roots cannot
  leak to actuals; usage keeps readable unknown-slot charges; numeric config rows
  are not malformed activity; housekeeping does not inflate focused time.
- [ ] Run go test ./cmd/sdlc/internal/transcripts -run Qoder -count=1; verify red.
- [ ] Register QoderHarness and reuse the leaf metadata decoder in both discovery
  paths. Preserve pinned length/stat/hash and ownership selectors. Return files
  selected by recorded workspace, not whole unrelated project directories.
- [ ] Drive real ownership fixtures through resolve/measurement, sdlc actual,
  close measured adoption and retained-prefix replay. Cover another checkout
  closing the owner's issue, reclaim intervals, unrelated sessions, nonzero
  activity, ambiguity and a subsequently extended/removed live source. Reuse
  existing actual/close test seams; do not special-case Qoder hours in close.
- [ ] Run go test ./cmd/sdlc -run 'Qoder.*(Actual|Close)|Actual.*Qoder|Close.*Qoder' -count=1
  plus the transcripts package, then commit verified wiring.

### Task 4: exact invoking identity without enabling unsupported reviewers

**Modify:** cmd/sdlc/internal/agentid/{env.go,detect.go,transcript.go,agentid.go,detect_test.go,agentid_test.go}; cmd/sdlc/internal/judge/{agent_resolver.go,agent_resolver_test.go} (verify existing test filename before editing).

- [ ] Resolve the measured native session binding with the hub; append a plan
  revision with signal provenance, exact lookup policy and ambiguity behavior.
  This step blocks identity implementation if the evidence is still missing.
- [ ] Write failing tests for invoking Qoder→qoder/auto/-, configured effort,
  config replay/model updates, exact session mismatch, multiple matching copies,
  absent/truncated/oversized metadata, SDLC_AGENT_ID precedence, Pair-only shell
  context, and unchanged Claude/Codex/Gemini identities.
- [ ] Separate observed-agent recognition from dispatch capability. Keep Known's
  existing reviewer contract or give reviewer checks an explicit capability
  predicate before extending identity recognition. Assert all reviewer policy
  modes/overrides do not select an unsupported executable as a side effect.
- [ ] Wire the measured signal and exact session reader to the shared adapter.
  Unknown model/effort stay dashes; auto is retained as recorded. Run
  go test ./cmd/sdlc/internal/agentid ./cmd/sdlc/internal/judge -run 'Qoder|Detect|Agent|Reviewer' -count=1,
  then commit verified identity behavior.

### Task 5: collector, native columns, documentation, and SHIP

**Modify:** cmd/sdlc/{usage.go,usage_reviewers.go,usage_command_test.go,usage_output_errors_test.go}; cmd/sdlc/helptext/usage.md; atlas/workflow/sdlc-binary.md; atlas/index.md; this plan and the issue.
**Create:** cmd/sdlc/usage_qoder_test.go.

- [ ] Write failing collector/CLI tests using a temporary home, fleet and issue
  ownership fixture. Assert Qoder appears in fleet glance and slot/model detail;
  tokens and token burn are JSON null; credits are exact; worker/reviewer/total,
  hourly and unbound scopes preserve both native units and exclusion semantics.
- [ ] Enumerate hardcoded provider lists in collection, parser coverage, meter
  windows and glance. Derive supported measurement agents from one usage-owned
  registry rather than scattering a third name across loops. Add Qoder read
  dispatch; leave gate CLI capability independent.
- [ ] Render schema v2 plus parallel credits/token columns; retain labelled
  coverage and existing first-output glance. Check output errors/short writes,
  late cancellation and durable-calibration summary errors still propagate.
- [ ] Run go test ./cmd/sdlc -run 'Usage.*Qoder|Qoder.*Usage|Usage.*(Output|Cancellation|Attribution)' -count=1
  and all changed internal packages. Keep the checkout/HEAD frozen during
  production-root fixture tests. Update atlas/help with format, null, exact
  native credits, calibration, unknown-model and ownership contracts.
- [ ] Read-only conformance: run the candidate reader against the supplied real
  sample, with no transcript content emitted. Compare normalized response count
  117 and exact credits 45.938730726457140528, Qoder token null, auto/null effort
  and recorded workspace. Record command/evidence hashes, not raw records.
- [ ] Reconcile plan entity/file inventory with the delivered diff; record test
  results in the issue. Merge origin/main and resolve before final verification.
- [ ] Run sdlc local-check run --name suite on the exact candidate commit in the
  background, inspect its complete result, then sdlc close --issue 385 --verified
  with targeted and full-suite evidence. Use shared sdlc, measured actual adoption,
  and the binary's one fresh-context boundary review; fix/dispute findings until
  SHIP. Direct make invocations use make -k per standing rule.
- [ ] Open/land with sdlc pr and sdlc merge, verify real hashes via git rev-parse,
  and report one landed outcome with PR and commit hashes to the hub. Honor moved
  main/exact-suite requirements; no self-applied local-gates label or CI bypass.

## Architecture cross-check

- ARCH-DRY: one Qoder metadata interpretation, one global normalization and
  attribution path, existing actual input ownership and meter storage reused.
- ARCH-PURE: metadata transitions, decimal arithmetic, normalization, projection
  are pure; filesystem/env/clock/output remain thin injected seams.
- ARCH-PURPOSE: all three requested consumers plus credits in every report scope;
  identity binding is an explicit blocking decision, not an unfulfilled follow-up.
- ARCH-MOCK: portable JSONL home plus real temporary Git fixtures model the
  consumed filesystem protocol; no provider calls. Read-only live conformance is
  explicit and default tests never touch the real account store.
- ARCH-CONSTRAINTS: local batch budgets, numeric precision, candidate/line/response
  caps, identity bounds and visible overload/cancellation are named above.
- ARCH-SECURE: typed whitelisted untrusted JSON, no payload/credential diagnostics,
  fixture sanitization, exact invoking identity and ownership proof.
- ARCH-ORDER: encapsulated session/config transitions plus global response
  duplicate/final/conflict states, pinned-prefix concurrency and sequence tests.
- ARCH-FUNERAL: no new runtime cache or sidecar family; existing bounded meter
  replacement and actual snapshot retention own all persistent measurement data.

## Revisions

### 2026-10-15 09:47 — measured Qoder detection and transcript selection

Supersedes the unresolved binding prerequisite and exact-session-environment
assumption above. The hub measured variable names on pair:4's running
qodercli-1.1.71 process: QODER_CLI is a native invoking-stack marker; no native
session-ID environment variable exists. Qoder also exports CLAUDE_PLUGIN_ROOT
and CLAUDE_PLUGIN_DATA, which must never establish Claude identity. No variable
values or private environment dumps are needed or retained.

Task 4's binding decision is resolved: after a valid SDLC_AGENT_ID override,
detect Qoder from nonempty QODER_CLI before other native-agent markers. Neither
PAIR_AGENT nor PAIR_SESSION_ID establishes Qoder invocation. Resolve candidates
under ~/.qoder/projects/*/*.jsonl by recorded workspace-directories matching the
invoking cwd (inject cwd/clock at the Detector IO seam). Rank by actual ISO
user/assistant activity timestamps, not file modification time, directory slug,
or replayed runtime-config timestamps. Select the uniquely latest active session;
read its model/effort through the shared adapter. No candidate, tied distinct
sessions, conflicting metadata, future/invalid activity or incomplete scanning
returns a guessed model: keep qoder/-/- and document explicit launch-bound
SDLC_AGENT_ID as the ambiguity fallback. Duplicate copies of the same identity
must agree rather than selecting the first glob match.

Identity discovery is a bounded local read: at most 256 candidate files,
16 MiB total bytes and a 2-second context, with the existing 1-MiB per-session
read budget shared between metadata prefix and activity/config tail. These are
initial operator-facing batch budgets, not measured latency claims. If any bound
prevents proving the latest candidate, degrade to unknown model/effort rather
than assuming unseen files are inactive. No goroutine or background scan
outlives Detect. Usage/actual retain their existing larger batch budgets.

Extend identity tests with QODER_CLI plus competing Claude/Codex/Pair signals;
CLAUDE_PLUGIN_* alone; ordinary Pair panes; wrong cwd despite newer mtime;
uniquely latest activity; timestamp ties; repeated copies; invalid/future rows;
replayed old config; injected cancellation and each scan bound. Test Qoder
observed identity without making the judge accept/dispatch a Qoder reviewer.
