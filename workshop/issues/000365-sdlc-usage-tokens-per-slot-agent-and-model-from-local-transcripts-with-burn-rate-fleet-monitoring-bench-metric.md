---
id: 000365
status: working
deps: []
github_issue:
created: 2026-10-14
updated: 2026-10-14
estimate_hours: 8.881
card_mirror: '4807cc301e82494819858a46b67e1bcb8f0e2ad4' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-14T07:39:58-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
---

# sdlc usage: tokens per slot, agent and model from local transcripts, with burn rate (fleet monitoring + bench metric)

## Problem

The fleet has no view of token spend. On 2026-10-14 the TL learned only by driving `/status` into an idle Codex slot that Codex's weekly limit was exhausted and workers were drawing credits, and by `/usage` in a Claude slot that Claude was at 94%. Driving commands into agents is intrusive and impossible mid-turn. The operator wants token monitoring with burn rates before routing work to other providers (e.g. Qoder's off-peak window), and the upcoming bench (ops workshop/pensive/2026-10-13-02-pensive-model-eval-by-replay.md) needs tokens per run.

The agents' own transcripts already record usage: Claude's session JSONL carries `usage` (input, output, cache read/write) per response; Codex's rollout JSONL carries `token_count` events. ariadne#335 just made transcript attribution per owning slot reliable.

## Spec

`sdlc usage` (read-only, no agent interaction), reusing #335's transcript discovery and attribution:
- Totals per slot, agent and model over a window (`--since`, default 24h), split input / output / cache read / cache write where the agent reports them; per hour, so a **burn rate** (tokens/hour over the last 1–3 h) falls out.
- `--issue N`: tokens for an issue's ownership intervals (the bench metric), alongside `sdlc actual`'s active time.
- `--json` for the TL's session file and the bench.
- Optional calibration: the TL records an observed account meter (e.g. Codex credits, Claude weekly %) with a timestamp (`sdlc usage calibrate --agent codex --remaining 29276 --unit credits`), and usage projects time-to-exhaustion from the measured burn rate. No scraping of agent UIs.
- Unknown or unparseable transcript formats are reported (counts of skipped records), never guessed.

## Done when

- `sdlc usage --since 6h` lists tokens per slot/agent/model with an hourly burn rate, from real Claude and Codex transcripts (tests on recorded transcript fixtures for both formats).
- `sdlc usage --issue N` reports tokens within the issue's ownership intervals, consistent with `sdlc actual`'s attribution.
- A calibration plus burn rate yields a time-to-exhaustion projection; `--json` output is documented.

## Plan

Durable plan: workshop/plans/000365-usage-plan.md.

- [ ] Normalize recorded usage and deduplicate before attribution.
- [ ] Reuse ownership/membership and persist explicit reviewer bindings.
- [ ] Add reporting, calibrated projection and docs.
- [ ] Verify recorded suite, SHIP, PR and merge.

## Log

### 2026-10-14

### 2026-10-14 — claimed and initial measurements

Claim/start-plan completed in ariadne:3. Read-only store inventory:2725 Codex
JSONL files/4.87GB,5140 Claude files/2.08GB; largest Codex rollout276MB. Default
usage needs bounded streaming and recent-file pruning rather than unbounded
whole-store reads. Actual source discovery is deliberately root-only for Claude;
token accounting must retain explicitly-bound child billing without duplicating
parent responses. Fleet membership collection must avoid default tracker fetches.

Local format investigations found Codex0.162.1 token_usage_record (response identity
and cumulative thread/turn counters), including compaction copies; legacy token_count
snapshots cannot be summed. Claude repeats message-level usage across content blocks
and may revise a streamed response. Response identity/dedup and original model are
required before attribution. Unknown dimensions/formats must remain explicit.

Asked hub for the bench's gate-reviewer attribution policy: #335's strict owner/root
and bound-descendant admission excludes standalone disposable-checkout reviewer
roots, so global unknown-slot spend and per-issue coverage must name that limitation
unless an explicit review-session binding is required. Meter conversion needs two
aligned observations in the same agent/unit; a single remaining value cannot imply
credits per token. No implementation code or account observations written yet.

### 2026-10-14 — durable plan and whole-system cost

Verified hub receipt efaa1145-7a99-4c4d-91bc-88b4d1f5858c: include reviewer tokens
with rounds separately from worker tokens and total. Existing sidecar/ledger types
do not retain roots yet; add optional typed dispatch bindings and label legacy
gaps. Modern Codex compaction/response identity and Claude duplicate/revised or
provisional usage require normalization with explicit coverage. Meter conversion
needs two observations. Durable plan ready for hub approval; no code changes yet.

## Revisions

### 2026-10-14 — whole-system bench cost

Per hub, issue cost includes explicitly bound reviewer rounds/attempts alongside
ownership-proven workers; expose separate subtotals and measured total. Missing
binding/format evidence remains coverage, never guessed spend.

### 2026-10-14 — hub plan approval

Verified receipt4479fade-4e09-47d1-be7d-806e76f8d261 approves plan1157b761.
Presentation addition: first output is one fleet-glance row per agent with last
hour tokens, labelled burn rate and calibrated projection/status; JSON exposes
the same fleet-scope glance. Proceed through change-code's owned plan-quality
gate; no estimate or implementation before its acceptance.

### 2026-10-14 — plan gate round1 refinements

Plan gate returned FAILURE with PQ-1 interrupted accounting persistence, PQ-2
exclusive roles, and Minor PQ-3 live protocol conformance cadence. Refined durable
plan with Requested→Started→Finished accounting-only receipts reusing Git CAS;
no semantic gate advancement during cancellation/staleness/death. Explicit reviewer
binding precedes worker eligibility; conflicts are unknown and measured issue
union counts each response once. Opt-in conformance runs on CLI upgrade/before
bench. No implementation or estimate yet; rerun plan-quality next.

## Estimate

Produced via brain/data/life/42shots/velocity/estimate-logic-v3.1.md against
baseline-v3.1.md. Method A only. Calibration source is flagged stale; provisional.
Read v2/v2.1/v3/v3.1 and baseline before deriving. Stdlib JSON/time and existing
Git/fleet/recovery/lock seams cover mechanics; no library replaces response
revision semantics, ownership or interrupted review attribution. Thorough approved
plan gives most domain components ×0.2 design discount and15% design buffer.
Issue-spec captures the already-performed authoring separately. Impl values are
40% of the v2 ranges; familiarity1.0 for this existing stack.

Items: issue/spec; counter schema; Codex decoder; Claude decoder; dedup/aggregation;
source discovery/lineage; shared ownership refactor; review attempt transition;
review receipt IO; gate binding integration; meter model; meter store; command/
rendering; docs; Claude/Codex live conformance each; one close review. Targeted
regressions/mutations are included in their owning components. Design3.34h +
impl5.04h; buffered total8.881h. No parallelism multiplier or typed actual.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec design=0.60 impl=0.08
item: greenfield-go-module design=0.10 impl=0.24
item: greenfield-go-module design=0.20 impl=0.32
item: greenfield-go-module design=0.20 impl=0.32
item: greenfield-go-module design=0.16 impl=0.28
item: api-integration design=0.40 impl=0.60
item: cross-cutting-refactor design=0.12 impl=0.20
item: greenfield-go-module design=0.16 impl=0.32
item: api-integration design=0.40 impl=0.60
item: api-integration design=0.40 impl=0.60
item: greenfield-go-module design=0.16 impl=0.28
item: smaller-go-module design=0.06 impl=0.20
item: api-integration design=0.30 impl=0.48
item: atlas-docs design=0.04 impl=0.08
item: real-api-discovery design=0.00 impl=0.12
item: real-api-discovery design=0.00 impl=0.12
item: milestone-review design=0.04 impl=0.20
design-buffer: 0.15
total: 8.881
```

### 2026-10-14 — plan-quality accepted

Round2 CLEAN high confidence; PQ-1/PQ-2/PQ-3 addressed, zero blockers. Derived
estimate only after that acceptance; next rerun change-code for reconciliation
and estimate advisory, then implementation.
