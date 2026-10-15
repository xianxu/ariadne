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

- [x] Normalize recorded usage and deduplicate before attribution.
- [x] Reuse ownership/membership and persist explicit reviewer bindings.
- [x] Add reporting, calibrated projection and docs.
- [x] Verify recorded suite and acceptance evidence.

Then publish through shared sdlc close (SHIP), PR and merge.

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

### 2026-10-14 — implementation checkpoint

change-code accepted the plan (CLEAN/high, round3) and derived8.881h estimate
(INFO/high advisory: receipt/cleanup integration and final ship checks are tight).
Implementation uses three bounded agents for normalization, source discovery,
and reviewer receipts; main owns calibration and command/gate integration.

Internal usage/transcripts tests pass. Four counting and four discovery mutations
were caught; duplicate normalization fuzz ran81,292 cases. Calibration regressions
cover one/two points, refill, conflict, stale reading, incomplete/zero burn,
cancellation and preserving the private atomic file. Root attribution tests
cover exclusive worker/reviewer union and conflicting binding copies. Root tests
remain pending until the final dispatch adapter freezes.

Candidate builds and a real local --since1h --json scan completed successfully:
7931 candidates,88 recently selected sources (the first smoke double-counted the
source counter and prompted a discovered_sources/decoded sources distinction).
It reported legacy reviewer gaps and no invented forecast. Additional integration
checks caught unbound brain-root admission, conflicting metadata restoration,
missing projection-window completeness and mutation detection; fixed before
root verification. Source stat identity now pins replacement/mtime evidence;
growth alone remains an allowed stat-sized append.

Reviewer accounting copies are optional round/sidecar fields. Estimate advisory
attempts are retained in an optional ledger-level accounting field, avoiding a
fake findings round. Cleanup requires an exact typed record in a committed
HEAD/main artifact; uncommitted or partial copies cannot collect private refs.
Token ownership and actual-time share ownershipSourceWindows; actual keeps its
independent verification-tail policy. usage reads a local tracker snapshot and
has no fetch, cache or provider interaction.

### 2026-10-14 — targeted integration and live conformance

Focused root usage/attribution/accounting tests pass (12.785s). The earlier
ownership resolver regression selection passed except two new fixture mistakes:
a fabricated unavailable reviewed SHA and an offline local remote losing its
symlink spelling. Fixed the fixtures to use the real Git head and preserve the
physical remote identity; production ownership safeguards stayed intact.

Both real CLI identity checks now pass (Claude5.20s, Codex8.09s;13.460s total).
Initial conformance found current Codex's versioned banner/optional stdin notice,
Claude's canonical physical cwd, and leading queue-operation records without
cwd. Added observed-format and queue→user regressions; no argv/permission change.
The live check uses the same usage discovery adapter, including underscore paths.

Private review admission now records its real PID/host and uses a10s bounded
lock wait; a dead-holder regression reproduced and fixed PID0 non-reclamation.
Unbound fleet evidence has its own recent window and is excluded from issue
worker/reviewer totals. Spend completeness and attribution completeness are
explicit JSON booleans; old reviewer bindings and unresolved aggregates remain
visible gaps. Projection checks its own burn/calibration evidence, not just the
issue detail window. Structured modelUsage/totalTokens counters diagnose gaps
without adding aggregate spend a second time.

### 2026-10-14 — resumed acceptance evidence

Recorded sdlc-local/suite passed on
081417175e5b47f9dd779f1747fe275f2957aab3:1389 cmd/sdlc tests in12 shards,
61 other packages, zero failures,765.9s (12m46s), log hash prefixaff116cf6257.
All code is in that tested head; the acceptance checkpoint is docs-only.

Fresh live identity conformance passed12.513s (Claude4.40s, Codex7.96s).
Candidate six-hour fleet JSON completed19.69s; issue365 JSON23.85s. Fleet
glance reports both agents first, with no invented projection when observations
are absent. Issue reports65,302,026 worker tokens, zero bound reviewer tokens,
the same total, and measured active time5.2947516667h with its own provenance.
Three pre-feature reviewer gaps remain explicitly labelled; shared old gates
cannot be retroactively guessed into the new binding schema. Reports are local
/tmp/ariadne365-resume-fleet.json and /tmp/ariadne365-resume-issue.json.

Next shared close to SHIP, PR and merge.


### 2026-10-14 21:12 — boundary round1 REWORK

Shared close reviewed2723e1f5892053ae932f5c86bf14818454ae62ad, measured5.70h,
and returned three Critical/one Important findings. BR-1 reproduced mixed modern
100/20 and legacy500/100 snapshots falsely complete; preserve per-dimension
high-water evidence across both formats and resets, never add snapshots to
response totals. New regression failed7/8 cases before fix, plus a copied-parent
ambiguity case failed before removing its blanket exemption; all9 cases and
usage package pass0.174s after fix. BR-2 invalid reviewer evidence and BR-4
production attachment regressions are being addressed separately. BR-3 concept
map now names delivered paths/status and PURE/IO ownership; Task2 production
integration checkbox reopened until the actual wiring tests pass. Prior suite
pass remains evidence for08141717 only; rerun after these code fixes.


### 2026-10-14 21:16 — round1 fixes verified

BR-2 now validates and reports invalid ledger/comment/private accounting, retains
usable identity solely for quarantine, and never authorizes cleanup from invalid
records. Targeted regressions2.369s and reviewusage package3.113s pass after RED.
BR-4 production boundary/plan/estimate dispatch→receipt/artifact→usage tests pass
three repetitions15.261s; independent omission mutations fail each production
attachment (1.179s/2.389s/2.590s). Cancellation preserves Finished/canceled receipt
without semantic advancement. Initial fixture mistakes briefly allowed fallback
real reviewers; both were interrupted/reaped, and explicit agent selection plus
failing fallback executables now prevent any provider call. BR-3 reconciled plan
paths/status/PURE ownership. All four classes are addressed pending fresh suite
and shared close review; no earlier suite is claimed for changed code.


### 2026-10-14 21:57 — BR-1 turn-counter sibling

Production-decoder overlay proved turn_token_usage was ignored: one turn's
missing spend appeared complete, including when another turn hid its deficit.
Per-turn maxima now reconcile only against that turn's deduplicated responses;
invalid counters are visible and no cumulative totals are charged. Eight RED
cases/four controls; all12 cases and full usage package pass0.165s after fix.
Superseded f3bfa3ed suite was cancelled after1959s queue wait and part of its run,
not adopted as passing evidence. Verified and stopped its orphan Make process
group81248 after local-check parent SIGINT exited130; no owned suite remains.
Fresh recorded suite follows this fix. Fleet/issue smoke25.02s/23.03s on the prior
head retained explicit coverage and null uncalibrated projections.


### 2026-10-14 22:49 — corrected suite passes

Recorded sdlc-local/suite passes on3679928075e6e17e5448ffb355fe3324886ee2ab:
1396 cmd/sdlc tests plus61 packages, zero failures,649.8s/10m50s;2392s queue wait.
All12 shards/packages exit0, log SHA256 prefixec5f77327432. Final fleet/issue
smokes pass40.92s/52.30s under load, with explicit coverage and null uncalibrated
projections. Live conformance12.513s remains valid for unchanged CLI identity
protocol. Plan function labels now match delivered seams. Shared close rerun is
next; no candidate gate binary grades its own landing.


### 2026-10-14 22:56 — round2 REWORK, prior classes disposed

Shared close verified/disposed all four original findings and reproduced BR-5
(discovery accepts Codex thread_id fallback but decoder does not; copied parent
spend admitted) and BR-6 (issue/repository filters hide other issue reviewer
identity before role selection). Fix the shared identity and exclusion-evidence
classes with production regressions; preserve all scan bounds and explicit
coverage. Fresh recorded suite follows behavior changes. No disputed finding or
waived guard; continue to SHIP.


### 2026-10-14 23:02 — round2 fixes verified

BR-5 shared immutable Codex parser now serves both discovery paths and Decode;
20 cross-consumer cases cover aliases, conflicts and malformed/empty first
metadata followed by copied headers. Old decoder mutation reproduces132 vs12
with false completeness (0.166s). Usage/transcripts packages pass0.321s/0.382s;
existing actual/provenance ownership regressions pass103.888s with frozen tree.
BR-6 loads all bounded fleet reviewer identities before issue role selection;
real committed cleanup parity for ledger/comment, cross-repo exclusion, invalid
identity and scan-bound regressions pass5.067s; broad targeted tests pass21.340s.
Both new classes have RED/GREEN evidence. Integrate main and run fresh recorded
suite before shared close; maintain all previous dispositions.


### 2026-10-14 23:11 — scope diagnostics by repository and issue

Final read-only issue smoke passes50.60s; fresh live identity conformance passes
14.543s. Smoke gap count4→14 exposed same-number foreign-repository legacy rounds
entering selected diagnostics. Real-Git RED proved2 vs1; explicit repo-qualified
scope now retains all exclusion identities with correctly selected gap counts.
Distinct origins/reversed roots/same-repo alternate worktree/fleet/error-cache
controls pass4.719s, broad usage/production/accounting45.998s. Queued b9d5da55
suite cancelled130 before running; own parent/acquire child verified gone.


### 2026-10-14 23:16 — cancellation/output regressions

Late-actual cancellation was represented as a status while collectUsage returned
success. Final collector/command boundaries now return context errors and retain
output errors. Bounded buffers propagate text/JSON/calibrate short-write failures;
already-saved meter observations are preserved with explicit recorded-before-
summary-failed diagnostics. Four production regressions pass count3/21.872s;
reverted-fix overlay fails all four7.808s. No provider calls or transition changes.
Commit this final code before a fresh recorded suite.


### 2026-10-14 23:59 — final suite passes

Recorded sdlc-local/suite passes on0ba4bc396785b7cdbd7a21cbffa66f474ccadb37:
1408 cmd/sdlc tests plus61 packages, zero failures,589.4s/9m50s,1943s queue wait.
All12 shards/packages exit0, log SHA256 prefix1adc88ca710d. Final issue smoke
passes48.18s with correct five own legacy gaps; latest live identity conformance
passes14.543s. Run shared close on this code with only documentation bookkeeping
since the recorded suite.

Standalone final --since6h fleet JSON smoke passes37.71s; final fleet/issue
timings37.71s/48.18s on the passing suite code.


### 2026-10-15 00:12 — round3 identity ordering REWORK

Shared close verified BR-1/2/3/4/6 and retained only BR-5. A production
discovery→Decode overlay admitted120 parent tokens before a late child metadata
header, plus12 child tokens, with complete spend coverage; absent metadata also
reported complete132. The plan now enumerates early/late/absent/invalid identity,
both aliases, copied headers and ordinary/compacted records. Resolve identity
before accounting or report uncertainty; absent executing identity must prevent
complete spend coverage. Implement and verify the whole class before fresh suite
and shared close. No passing gate or landing is claimed.


### 2026-10-15 00:20 — identity ordering class verified

BR-5 now rejects ordinary/compacted usage and cumulative snapshots before a
valid executing identity, before counter, baseline, high-water, timestamp or
compaction bookkeeping. Missing identity always leaves spend incomplete, even
for empty/legacy-only sources; valid identity without cwd remains spend-complete.
The production discovery→Decode matrix covers both aliases, early/late headers,
absent/invalid headers, copied later headers and owned/foreign records:39 identity
cases. Explicit headers were added only to individual independent fixtures.

Usage/transcripts packages pass1.045s/0.918s; stable command, attribution,
reviewer-accounting and production dispatch regressions pass105.478s. Agent
mutation runs fail as intended with the old decoder1.299s, removed snapshot
guard0.189s and removed unresolved-identity finish guard0.211s. Help/atlas now
document skipped preidentity usage and incomplete coverage. Commit and integrate
main before a fresh recorded suite; shared close remains pending.

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

### 2026-10-14 — final attribution regressions and smoke

A disposable checkout is now an explicit immutable dispatch fact, rather than a
path-name heuristic. A captured session identity outranks an earlier unfinished
root-only receipt on a reused checkout; contradictory exact identities remain
unknown. Offline legacy ownership uses a local-only tracker resolver, preserving
the labelled legacy fallback without a fetch. Calibration coverage is scoped to
each agent after global deduplication: another agent's format gaps cannot erase
a complete denominator, while the agent's own gaps still refuse projection.

Focused root regressions passed12.029s; full usage/reviewusage/judge/transcripts
package runs passed (0.164s/1.872s/14.559s/0.171s). Earlier live CLI conformance
passed13.460s. Real candidate scans on f2695117751b33038fef21c26ebd721dbd93371b
took23.94s for six-hour fleet and17.69s for issue365. Issue smoke reported
29,884,871 worker tokens, zero bound reviewer tokens and three explicit legacy
review gaps; shared pre-feature reviews cannot be guessed into the new schema.
No account observations existed, so projections were null.

The first recorded suite attempt was cancelled while still queued, before any
suite execution. It is not passing evidence. Next: commit these tested fixes,
run the recorded suite on the exact pushed head, then close/SHIP/PR/merge.

### 2026-10-14 — full-suite findings and targeted repair

Recorded suite on5f2038d83b65449224fff53a945dfd6cd7cf6421 failed after531.6s
(1357 root tests/12 shards,56 other packages). Three classes: new Git-driving
reviewusage package lacked TestMain's real-Git/testgate setup; accounting appended
after the byte-measured diagnostic envelope; private accounting refs unnecessarily
pinned reviewed code and prevented a rebased generation from being pruned.

Reproduced all failures. A new store regression failed before removing object
pins, then passed while proving the receipt still retains its reviewed SHA.
Accounting receipts contain metadata only, not code recovery authority. Moved
accounting before the measured diagnostics and added the standard package test
harness. Targeted root lifecycle/usage/accounting regressions passed50.143s;
reviewusage package passed1.330s. Next integrate current main and rerun the suite.

### 2026-10-14 — operator freeze checkpoint

Operator paused #365 to reserve machine/suite capacity for essential fixes.
Implementation and targeted repairs are committed and pushed through
a328babd198a2e5696a69c565701c872385b32d2. Replacement full suite was cancelled
while queued (owned group32406, exit130), before execution. No passing full
suite record exists for the repaired implementation. No close or PR yet.

Hub subsequently authorized only #369 cancellation-flake repair from main.
#365 remains paused; resume requires explicit unfreeze, then recorded full
suite, shared-binary close to SHIP, PR and merge.

### 2026-10-14 20:40 — unfreeze and main integration

Hub/operator authorized resume. Checked out the preserved issue branch and
merged current origin/main without conflicts at
74f8305f9fd7b9b76f1a6ae63cb7902cc9ddc59c. This includes the #369 termination
and lease-fixture repairs plus later SDLC/TL changes. Focused root usage,
accounting and prior lifecycle regressions pass61.001s; usage/reviewusage/judge/
transcripts packages pass (0.158s/1.723s/9.697s/0.160s). No code change was
needed for integration. Next recorded suite, shared close review and landing.

### 2026-10-14 20:40 — acceptance checklist clarification

The original final checkbox combined verification with SHIP/PR/merge, which
necessarily occur after close checks the implementation plan. Track verification
as the checkbox and retain the unchanged publishing sequence as prose, so a
checked plan cannot falsely claim that future landing has already happened.
Durable Task1–3 reflect the completed implementation/checkpoint evidence;
Task4 retains the pending suite/close/landing steps. Scope and Done when unchanged.
