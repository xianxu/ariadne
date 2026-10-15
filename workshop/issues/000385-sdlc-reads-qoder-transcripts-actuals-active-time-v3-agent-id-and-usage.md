---
id: 000385
status: working
deps: []
github_issue:
created: 2026-10-15
updated: 2026-10-15
estimate_hours: 3.81
card_mirror: 'fb1ff013603e865a1b162809cc46c4e5660fece3' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-15T09:33:04-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
---

# sdlc reads Qoder transcripts: actuals (active-time-v3), agent-id and usage

## Problem

The first Qoder trial (pair#558 on pair:4, 2026-10-15) hit two sdlc blind spots, both because sdlc reads agent
transcripts only from Claude (`~/.claude/projects`) and Codex (`~/.codex/sessions`):
- `sdlc close` refused with "no measurable activity": active-time-v3 found 0 events, because the session was in
  `~/.qoder/projects/<slug>/<session>.jsonl`. The worker had to close with `--no-actual`, so Qoder work drops out of
  velocity calibration.
- `sdlc agent-id` (#367) printed the no-session fallback `Agent: sdlc` although `PAIR_AGENT=qoder` was set; the worker
  wrote `Agent: qoder/auto/-` by hand. `sdlc usage` (#365) likely has the same gap.

Qoder's transcript is JSONL close to Claude Code's shape: `{"type":"user"|…,"sessionId":…,"timestamp":…,"message":…}`,
with `runtime-config` rows carrying `model` and `reasoningEffort`, and a `workspace-directories` row with the cwd.

## Spec

- Add Qoder as a transcript source behind the same seam the Claude and Codex readers use: discovery under
  `~/.qoder/projects/<slug>/`, the cwd from `workspace-directories`, events with timestamps, model and effort from
  `runtime-config`.
- Feed it to active-time-v3 (`sdlc actual`, close adoption), `agentid.Current` (`sdlc agent-id`), and `sdlc usage`
  (token counters if the format reports them; otherwise null, never zero).
- Fixtures from a real Qoder transcript (trimmed, no secrets).

## Done when

- `sdlc actual` on a Qoder-worked issue measures nonzero hours from its transcript; close adopts them.
- `sdlc agent-id` in a Qoder session prints `Agent: qoder/<model>/<effort or ->`.
- `sdlc usage` lists Qoder slots (with unsupported dimensions as null).
- Tests over the fixture for each.

## Plan

Durable design: [Qoder transcript measurements](../plans/000385-qoder-transcripts-plan.md).

- [x] Shared bounded Qoder record/metadata adapter and sanitized fixtures.
- [x] Exact per-response credits, nullable token measurements, global dedup and
  native-unit aggregation/calibration through existing usage primitives.
- [x] Recorded-workspace discovery feeding ownership-proven actuals and close
  adoption, with retained-prefix and cross-checkout regression fixtures.
- [x] Measured invoking-session binding and Qoder identity; keep observed-agent
  recognition separate from gate-reviewer CLI capability.
- [x] Native columns and coverage across fleet/slot/model/issue/hourly output;
  documentation, conformance and targeted regression checks.

Release sequence: recorded exact-head full suite passed; shared close review
to SHIP, PR and merge remain pending. These gates follow completed implementation.

## Log

### 2026-10-15

- Claimed and entered planning. Read-only inspection of the supplied Qoder
  sample found 313 assistant events but 117 distinct response IDs, each with
  one terminal usage record. UUIDs identify streaming events, not responses.
  `credits` is per response; `original_credits` must not be added again.
- All 117 billed responses have zero token placeholders. Credits total exactly
  `45.938730726457140528`, with up to 18 fractional digits. An int64 scaled to
  18 decimal places already overflows on this sample; use bounded exact decimal
  arithmetic, with tokens unavailable rather than measured zero.
- Metadata has mixed timestamp shapes: runtime-config uses epoch milliseconds;
  user/assistant activity uses ISO timestamps; workspace-directories has no
  timestamp. Repeated configuration may replay older timestamps. A shared leaf
  adapter avoids divergent interpretations and the transcripts→agentid import
  cycle (ARCH-DRY, ARCH-PURE).
- Workspace comes from recorded directories, never the project slug. Multiple
  or conflicting roots must remain ambiguous; parentUuid is message ancestry,
  not evidence of a child session. Existing ownership intervals, prefix evidence,
  reviewer exclusion, and close adoption remain the attribution authority.
- Pair source pins Qoder --session-id and exports PAIR_SESSION_ID, but exports
  that context to non-agent panes too. Asked the hub for a measured native
  invoking-session signal; do not guess a QODER_* variable or broaden ordinary
  shell detection. Observed-agent support must not accidentally advertise Qoder
  as a dispatch-capable gate reviewer.
- Next: finish the durable plan (nullable native measurements and credits,
  shared discovery/decoder, identity binding, actual/close fixtures), commit it,
  and send the plan to the hub before change-code. No implementation yet.

## Revisions

### 2026-10-15 — measured native unit and integration design

The dispatch explicitly requires Qoder credits by slot/model, not zero token
spend. Extend Spec/Done when: every usage output scope carries optional credits
and nullable tokens/rates; mixed aggregates identify measured subtotals; exact
decimal sums preserve the measured precision. Optional Qoder meter projections
use reported credits as their denominator and retain the existing two-point,
complete-evidence requirement. Usage JSON declares schema version 2. The durable
plan enumerates discovery, normalization, attribution, actual/close, identity,
reporting and calibration consumers. Identity implementation remains blocked on
measured invoking-session evidence (asked the hub), rather than guessing an
environment marker or enabling an unsupported gate-reviewer CLI.

### 2026-10-15 09:47 — hub measurement resolves identity prerequisite

The TL measured QODER_CLI on the running qodercli-1.1.71 process; its tool
shells inherit it. There is no native session-ID variable. Detect Qoder before
other native-agent signals, then choose the uniquely latest activity session
whose recorded workspace matches cwd. Ambiguity or bounded/incomplete discovery
keeps model/effort unknown, with explicit SDLC_AGENT_ID as the launch-bound
fallback. Qoder also sets CLAUDE_PLUGIN_* variables; these are not Claude
markers. The plan revision specifies bounded cwd/activity selection and the
priority/ambiguity regression matrix, superseding the earlier open question.

### 2026-10-15 — approved plan; plan-quality round 1

The hub approved the core plan and minimal native-credit meter. change-code
round 1 raised PQ-1 (recency does not prove invoking session) and PQ-2 (replace
test-case inventories with named functions and mechanical strategies). The plan
now carries a concise function-level adversarial strategy and TDD task sequence.
Asked the hub to refine identity: QODER_CLI establishes invocation; a Pair-pinned
PAIR_SESSION_ID can then bind the exact transcript only after cwd/sessionId
validation; without launch binding, competing sessions remain unknown. Awaiting
that decision before the gate rerun; no production code changed.

- Hub approved the exact identity refinement: QODER_CLI proves invocation;
  PAIR_SESSION_ID binds only after transcript sessionId/cwd validation. A missing
  or mismatched binding cannot choose a competing session by recency; native-only
  ambiguity remains qoder/-/-. Recorded the superseding plan revision for PQ-1.

### 2026-10-15 — implementation checkpoint

- Shared parser/state, discovery, native usage core and invoking identity pass
  their four internal package suites together. Parser race/seeded fuzz and
  generated state/response permutations pass. Actual/close integration passes
  through production ownership fixtures, including reclaim, cross-checkout
  parity, nonzero adoption, pinned replay and truncation refusal.
- Read-only production Decode + NormalizeResponses conformance against the
  supplied transcript passes: 117 responses, exact 45.938730726457140528 credits,
  tokens null, model auto, complete spend evidence. An external temporary Go
  overlay runs this check without adding the private sample to the checkout.
- Native measurement objects and known subtotals are nullable; missing token
  support never becomes zero. Workspace/config ambiguity preserves known credits
  and marks attribution unknown; session identity conflicts remain blocking.
  Runtime-config epochs, not message aliases or replay arrival order, supply
  response models. Credit JSON roundtrips preserve aggregate precision.
- Discovery diagnostics now separate provider-local and truly global failures;
  a candidate-cap interruption also marks later unscanned providers incomplete.
  The collector regression reproduced Qoder failures contaminating Claude's
  meter before scoped coverage fixed it. A partial-group regression reproduced
  loss of a known subtotal before using the retained measured dimensions.
- Identity uses exact validated launch binding, canonical workspace comparison,
  and rejects contradictory copies. A generated middle-conflict regression
  demonstrated prefix/tail sampling could miss poisoned metadata; identity now
  scans complete sources up to 4 MiB (16 MiB overall, 256 candidates, 2 seconds),
  returning unknown on excess. No runtime artifacts or provider-store writes.
- Outer native tests pass fleet amounts, calibration, hourly/model detail,
  provider isolation and partial subtotals. Broad usage/reviewer run reached one
  new issue-fixture failure: identical physical copies are deliberately unbound
  by existing ownership selection. Fixture now separates copied-spend dedup from
  unambiguous issue attribution. Final empty-role null regression is in progress.
- Remaining: finish root output tests, reconcile inventory/docs, merge current
  main, record exact-head full suite, shared close review to SHIP, PR and merge.

### 2026-10-15 — integrated candidate and targeted verification

- Native output, issue/unknown scopes, minimal two-point credit calibration,
  provider isolation, measured partial subtotals and empty-role null behavior
  pass focused root tests (6.790s). Broader usage/reviewer tests pass (118.657s);
  exact production boundary/planning review-accounting tests pass (7.597s).
  Output-error and short-writer behavior remains covered.
- Real-sample identity conformance through production Detector also passes:
  qoder/auto/-. Both live checks use external read-only Go overlays; no private
  JSONL is copied into this branch. Portable checked-in inline fixtures preserve
  only the observed protocol shape and invented content/identifiers.
- Late malformed owned evidence reproduced partial-adoption risk: a good owned
  source was admitted while the bad owned source lost its relevance on error.
  Discovery now retains possible workspace relevance through all failure paths;
  requested-owner uncertainty refuses rather than yielding partial hours.
  Unrelated malformed sources remain warnings. Full transcripts/race tests pass.
- Temporary build volume /Volumes/AgentScratch filled; main Data volume has
  space. No cleanup or global Go configuration changes performed. Final targeted
  checks use a fresh temporary GOTMPDIR on main Data. Asked hub to coordinate
  scratch relief or the full-suite environment; suite not started yet.

Delivered concept/file reconciliation: typed Record/SessionState live in
internal/qodertranscript/{record,session}.go with inline fixtures/property tests
in record_test.go; native arithmetic/measurements/decoder/meter live in
internal/usage/{credits,types,aggregate,qoder,calibration}.go; discovery owns
internal/transcripts/qoder.go and scoped usage diagnostics. Invoking identity
lives in internal/agentid/qoder.go plus Detect's native-marker branch. Existing
Env/Known and judge dispatch capability remain unchanged; no Qoder reviewer CLI
was added. Outer collector/renderers and issue parts live in usage.go; actual
adoption integration lives in actual_qoder_test.go. This supersedes proposed
file splits/names in the initial plan without adding another runtime owner.

- Final code candidate f509924f894f1830d5e787a13befadf5d9f3e1c1 includes
  origin/main 1e2e9bd573e7b08366688d10001e95583dd35e97 (merge reports already
  up to date). Exact-candidate root Qoder/native/partial-subtotal tests pass
  (10.794s), and all four changed internal package suites pass together.
  No root test jobs remain. Ready for the recorded full suite once shared
  scratch capacity/environment is resolved; no suite or close started yet.

- Hub confirmed operator-approved cache cleanup restored scratch capacity.
  Recorded suite started via sdlc local-check run --name suite on pushed
  02cdc490f32e7952d27ca04f9841fac4e000f8d9; waiting next in machine-wide queue.
  Inflight log: git common dir/sdlc/test-stats/inflight/
  9386958194ea2e6d5bf9b125515098a4/run.log. No other test jobs owned.

- Plan-quality round 2 CLEAN: PQ-1 and PQ-2 addressed, no new findings.

- Recorded suite on 02cdc490f32e7952d27ca04f9841fac4e000f8d9 failed after 558.4s: all runtime shards/packages passed, but TestVerbContextsReachTrackerReads found background contexts in Qoder identity and transcript selection. Thread caller contexts through both APIs and commit before a new recorded suite; do not change the guard. Failure log: git common dir/sdlc/test-stats/failed/9386958194ea2e6d5bf9b125515098a4.log.

- Context repair verified: existing TestVerbContextsReachTrackerReads (0.770s), TestEveryCommitSiteStamps plus migrate/propagation checks (14.987s), and TestQoderActualCloseOwnershipAndRetainedReplay (31.015s) pass. Identity race suite passed; transcript cancellation/race suite passed. No guard exceptions changed.

- Recorded full suite passed on 36d666d5d23b68665210a0e79d8c54cb926de3ab: 12 shards over 1493 cmd/sdlc tests plus 63 other packages, zero failures, wall 549.7s. Status sdlc-local/suite; log sha256 prefix 87ceba8033ab. Queue wait 1498s was separate from runtime. Full log: git common dir/sdlc/local-checks/36d666d5d23b68665210a0e79d8c54cb926de3ab/suite.log. Gitx context/publication regression selection also passed (144.720s).

- Close round 1 returned FIX-THEN-SHIP with no Critical findings, but was not finalized because BR-1/BR-2 remained Important. Fixed the unsupported-format class across all four token fields (including streaming, invalid types and numeric forms), preserving null tokens/known credits and blocking projection through existing coverage. Red-to-green internal regressions and full internal usage race suite pass (1.243s); production collector coverage/JSON/projection regression passes (3.656s). README/atlas now describe native credits, schema-v2 nullability and calibration, and the atlas anchor is repaired. Next: record a full suite on the fix commit, then rerun shared close for SHIP.

- Fix candidate c3045138d6284cccf9dd9c51fe9bf2522dd2883d passed recorded sdlc-local/suite: all 12 shards over 1493 cmd/sdlc tests and 63 other packages, zero failures, 467.5s runtime; log sha256 prefix 96f6b115fbcd (1613s separate queue wait). Read-only conformance also passes on the original 883-record prefix: 117 responses, exact 45.938730726457140528 credits, null tokens, complete spend coverage. The live file had grown to 1107 records/148 responses when reread, so its old whole-file fixed-count assertion was replaced only in the external conformance overlay with this original-prefix check; no fixture payload was copied.

- Close round 2 disposed BR-1/BR-2 as addressed, then raised Critical BR-3 (incomplete-evidence-adoption): a parse/read/bound failure before workspace proof was silently excluded. Confirmed production red (29.68s): with a positive 0.10h good-source baseline, malformed prefixes before owned/foreign metadata, no metadata, oversized prefixes, and unreadable candidates all still measured/adopted 0.10h. Expanded internal discovery matrix also reproduces the class. Durable plan now defines unresolved/proven-unrelated/relevant source outcomes, requiring complete exclusion proof and refusal on uncertainty. Implementation and fresh verification remain pending.

- BR-3 tracked production matrix now green (23.706s), internal transcript race suite green (8.092s), and usage/adapter/identity suites green. Consumer sweep reproduced the same class in the labelled no-tracker path: positive baseline followed by malformed Qoder evidence still measured/adopted 0.10h (legacy production red, 1.15s). Propagate conventional source-selection errors there too, reusing the strict Qoder discovery path while retaining the legacy attribution heuristic. Same-size pathname replacement was independently reproduced and checked; preserve its mtime in the regression so inode identity remains the decisive guard.

## Estimate

Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against
`baseline-v3.1.md`. Method A only. Source is labelled stale by estimate-source,
so this is provisional calibration, not a throughput promise. Derivation after
plan acceptance: issue/spec 1h design ×0.2; adapter 1h ×0.2; decimal module
1h ×0.5 library discount (math/big) ×0.2; six existing-seam extensions
(discovery, identity, response decoder, meter, reporting, actual/adoption tests)
each 0.3h ×0.2 design and 0.5h ×0.4 implementation; nullable refactor 0.5h ×0.2;
docs 0.1h ×0.2; one review 0.2h ×0.2. Implementations otherwise use v2 upper
module ranges at v3.1's 40%, plus one 0.6h ×0.4 format-discovery primitive.
Familiarity 1.0 (same stack and #365 seams), thorough-plan design buffer 15%.
Design 1.02h ×1.15 plus implementation 2.64h = 3.813h (card 3.81h).

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec design=0.20 impl=0.08
item: greenfield-go-module design=0.20 impl=0.32
item: greenfield-go-module design=0.10 impl=0.32
item: smaller-go-module design=0.36 impl=1.20
item: cross-cutting-refactor design=0.10 impl=0.20
item: atlas-docs design=0.02 impl=0.08
item: milestone-review design=0.04 impl=0.20
item: real-api-discovery design=0.00 impl=0.24
design-buffer: 0.15
total: 3.813
```
