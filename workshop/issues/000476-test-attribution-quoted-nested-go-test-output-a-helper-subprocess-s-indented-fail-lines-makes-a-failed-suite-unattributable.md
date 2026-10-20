---
id: 000476
status: codecomplete
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'ae9cd0f81e8dffeae7ea177d9a0ff2cae25e43c2' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-19T20:57:12-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
actual_hours: 0.82
---

# Test attribution: quoted nested go test output (a helper subprocess's indented FAIL lines) makes a failed suite unattributable

## Problem

pair#712's recorded suite (attempt d7b39428f6f4d3de2182c233873dbb15, 10-19) failed in one package and sdlc classified
the run as `unattributable-failure` ("couchcmd failed without a failing test"), so no flake path applied. The log does
name the failing test at column 0:

```
--- FAIL: TestContinuationWriterPublishesExactCheckpointAcrossWorktrees (3.25s)
    --- FAIL: TestContinuationWriterPublishesExactCheckpointAcrossWorktrees/warm-reattached (1.01s)
        continuation_acceptance_test.go:137: production transport failed: exit status 1
            === RUN   TestCouchLaunchPromptTransportHelper
            ...
            --- FAIL: TestCouchLaunchPromptTransportHelper (0.12s)
            FAIL
            FAIL	github.com/xianxu/pair/cmd/internal/launcher	0.226s
```

The test runs a nested `go test` (a helper process) and quotes its output, indented, in the failure message. The quoted
nested `--- FAIL` / `FAIL <pkg>` lines most likely confuse the attribution parser, e.g. it attributes to the nested
package/test, or treats the quoted `FAIL` as package end. Interleaved non-test stderr ("couch: terminal: presenter
unavailable") precedes the FAIL block too. Saved log: pair `.git/sdlc/test-stats/failed/d7b39428….log` lines 807–820.
The flake itself is pair#715 (registered).

## Spec

Preserve line provenance before parsing suite control records. Package outcomes,
shard summaries, JSON events and top-level test failures must start at column zero;
indented quoted output cannot change package attribution or add suite outcomes.
Attribute to the deepest name proven by column-zero or structured records.
Raw indented subtest lines under a failing ancestor are ambiguous quotations;
coarsen to the proven ancestor, usually its top-level test. Structured Go JSON
failure events and explicit test-shard headers still prove exact leaves.
Unrelated stderr remains inert. Keep existing persisted failures, transient
subtest fields, registry policy and rerun protocol (ARCH-DRY, ARCH-PURE).

Prefer this correction at the existing parsers over changing test producers or
introducing a new log format: historical saved logs need the same repair.
Keep the shared prefix/Unicode classifier; remove ancestry tracking entirely.
Use the existing warning channel to explain coarse attribution and why a
subtest-only registration may not cover it. No unproven leaf can narrow a
failure into a registration that would not cover its proven ancestor.

Actions: run the registered targeted retry; repair the named failure; inspect the
raw log and repair/rerun an unattributable failure.

| Input | Visible result | Recovery | Possible loss |
| --- | --- | --- | --- |
| Outer named failure with ambiguous raw child output | Proven ancestor plus coarsening note | Retry only if that ancestor is covered; otherwise repair reported test | Precision may be lost; raw log unchanged |
| Structured exact leaf | Existing exact leaf | Existing registered retry or repair | None |
| Genuine unrelated package/build/timeout failure | Existing blocking failure | Repair the reported failure | None |
| Missing outer name or package evidence | Existing unattributable failure | Inspect raw log, repair and rerun | No invented attribution |

No durable state, format, artifact family, new refusal or external dependency is
added. Parsing stays bounded by existing log/line caps and a fixed number of
linear scans plus name sorting. Growth drivers are bytes, lines and distinct
names; indentation checks inspect each line once, with no look-back or rescans.

## Done when

- The incident fixture attributes its failure to the proven
  `couchcmd.TestContinuationWriterPublishesExactCheckpointAcrossWorktrees` ancestor,
  with no invented nested helper leaf and no false unattributable package.
- Interleaved non-test stderr lines before a `--- FAIL` don't break attribution (same fixture).
- Indented quoted package/build/timeout/JSON outcomes stay inert. Proven JSON
  and test-shard leaves retain exact matching; genuine unrelated or unidentified
  failures still block.
- The gate matches at the proven granularity via flakes.Covers: an ancestor
  registration can rerun; a subtest-only registration cannot cover a coarsened
  ancestor. The output explicitly explains this precision miss.
- Real Go parent/child/log ordering fixtures coarsen safely across plain, CRLF,
  sharded contexts and retained-log readers. Unicode grammar and N/2N visit
  controls remain.

## Plan

- [x] Reproduce the issue fixture red; add plain, sharded and conservative controls.
- [x] Correct existing failure/fact scanners before whitespace normalization; add
  the narrow lint guard against erasing line provenance in these scanners.
- [x] Run focused parser/classifier and local-check regression checks.
- [x] Complete the round-2 BR-1 whole-prefix re-plan and its regression matrix.
- Superseded: the round-3 full ancestry-lifetime proposal is dropped by the
  hub's revised goal; no implementation of that model is required.
- [x] Implement proven-granularity attribution, visible precision misses and
  real-Go/historical/structured evidence controls under the revised contract.

Workflow gates: recorded suite, close review, then normal landing. These remain
pending independently of the completed implementation checkboxes.

## Log

### 2026-10-19
- 2026-10-19: closed — Revised TL-approved goal: raw ambiguous child output coarsens to the proven parent, JSON/shard leaves remain exact, and subtest-only registration misses are visible. Real Go buffered/verbose, incident, retry, saved-pass/replay, Unicode, N/2N, parser and environment controls passed (33.889s); golint and flakes passed. Final pushed-head recorded suite is queued/running; no full-suite pass claimed. README, atlas and lessons reflect the revised contract. Re-review BR-1 against the current approved acceptance in Revisions.; review verdict: SHIP
- 2026-10-19: flow upgraded quick → full — 111 added lines in code files (limit 100); an earlier round of this close already ran the full review

- Full close round 3 kept BR-1 Critical: real Go output can return from a failed
  child to parent logging, then quote a helper at a depth that fits stale child
  ancestry. The whole-prefix correction passed, but ordinary output never
  retired ancestry. Stopped again before code, cancelled the still-queued suite,
  and appended a complete line-event/context-lifetime re-plan for hub review.

- Hub approved the complete-prefix re-plan (6844e546-4bc3-4e66-b635-e03181de0edb,
  canonical receipt verified). Full Unicode-whitespace and mixed-prefix real
  gate tests failed red before the correction. The single shared classifier now
  supplies column-zero admission and the pre-Fields indentation level; existing
  ancestors stay unchanged. Expanded controls passed (40.830s), golint (1.309s)
  and flakes (0.070s) passed. BR-1 awaits the next full review; the new recorded
  suite and landing remain pending.

- Full close round 2 kept BR-1 open: initial ASCII-space measurement plus Fields
  admits mixed space/tab or space/Unicode-whitespace prefixes. Stopped before
  the third production delta and appended a complete re-plan to the durable
  plan. The current-head suite was cancelled while still queued; no full suite
  has run on this issue. Awaiting hub approval of the single prefix classifier.

- Hub approved the structural correction (09f8538f-08a4-4e8f-aebe-0786f25ec39e,
  canonical receipt verified). BR-1's shared-name matrix, ancestry/slash controls
  and quoted-only-registration gate failed red before correction. After the
  observed-ancestor fix, expanded parser/gate/saved-evidence/environment checks
  passed (12.864s); golint (0.550s) and flakes (0.052s) passed. No full-suite pass
  yet. The earlier queued suite ended cancelled without admission or a record.

- PR #336 is open; the recorded suite is queued on the pushed code head.
  Closing review overlaps that queue under the worker rule; any code correction
  requires a new suite on the corrected head. No passing full-suite claim yet.

- Focused quoted fixture/parser/registered-retry checks passed (1.165s).
  Broader parser/local-check/known-flake/saved-evidence/environment controls passed
  (296.155s); `pkg/golint` including tracked-source and atlas guards passed
  (1.042s), and `internal/flakes` passed (0.123s). N/2N traversal checks passed.
  The full recorded suite and close review remain pending.

- TDD red reproduced exact `couchcmd failed without a failing test` through
  `runLocalCheck`; nested launcher trailers also invented package timings. After
  the scanner correction, all focused parser controls and the registered outer
  retry passed. Corrected the test expectation to #466's whole-parent retry
  contract (the registered leaf's PASS remains required).
- The new `suite-line-provenance` lint's negative fixtures failed before the
  rule, then passed with it. Its scope is the named scanners and lexical call
  ordering, not a general control-flow or alias proof.

- Claimed after #477 landed. Traced `failingNames` (Fields) and both suite-facts
  loops (TrimSpace): quoted nested FAIL package trailers consume pending outer
  failures and manufacture unrelated package outcomes. The issue's captured
  snippet supplies the regression; no live sandbox or peer checkout changes.

- 10-19 19:30 (TL): second case, pair#689 at pair:6. Integrated suite 40e6dab11e0c6019301bd5767e687689 failed only
  at the same `couchcmd.TestContinuationWriterPublishesExactCheckpointAcrossWorktrees/warm-reattached` (nested launcher
  helper TempDir cleanup), classified unattributable-failure again, refused landing, and queued a main baseline. Two
  slots' landings blocked by this today; raise priority if #715's root fix (pair:3) doesn't land first.

## Revisions

- 2026-10-19 — Goal revised by hub receipt
  c4637f8b-c20e-4bb2-99b6-9077742773b6 (canonical verified): precision is optional,
  false unattributable output and invented leaves are not. Drop the pending
  ancestry-lifetime proposal. Keep prefix/Unicode classification, trust only
  column-zero/structured failure records, and use existing Covers at the proven
  granularity. A subtest-only registration missing a coarsened ancestor is an
  accepted, visible limitation. This supersedes earlier precise-raw-subtest
  requirements and the unimplemented context model.
  Previous acceptance required the incident's exact warm-reattached leaf and
  its registered retry from raw output, plus preservation of raw real descendant
  subtests. Those clauses are intentionally replaced by ancestor attribution
  and exact leaves only from unambiguous structured evidence. Existing real
  build/timeout/missing-identity guards remain.

- 2026-10-19 — Context authority must expire on ordinary parent output, not only
  the next test/package marker. The durable plan now proposes a tagged line
  classification and complete ancestry transitions, including blank/invalid
  input and conservative historical replay. Pending hub review; no fourth
  production delta implemented. All three old-head suites were cancelled
  before admission and produced no full-suite result.

- 2026-10-19 — Third-delta stop/re-plan: one classifier must validate the entire
  leading Unicode-whitespace prefix as ASCII-space quads before Fields, and the
  column-zero predicate must derive from it. Existing ancestor transitions stay
  unchanged. The durable plan records alternatives and the complete whitespace
  regression class. No third production correction has been made.

- 2026-10-19 — Corrected code adds 101 lines outside tests/docs, so the next
  close upgrades from quick to full review. Record the already approved design
  in workshop/plans/000476-suite-output-provenance-plan.md. No new scope or
  approval request; no claim that an initial full plan-quality gate ran.

- 2026-10-19 21:14 PDT — BR-1 (suite-evidence-provenance): the prefix-only
  exception admits quoted helpers sharing the outer root. Reproduced red in
  plain, CRLF and sharded forms, with and without a real failing child. Cancelled
  the queued old-head suite before it started; no suite result was recorded.
  Proposed correction, pending hub approval: retain an ancestor-name stack
  indexed by observed four-space indentation. A real child must be exactly one
  indentation level below an admitted parent and descend from that parent's
  full name; truncate the stack on siblings. Reject skipped levels and leave
  the stack unchanged on quoted records. Column-zero outcomes and authoritative
  shard summaries reset the context; summary names remain authoritative even
  when raw output omits ancestors. No depth inference from slash counts, since
  one t.Run name may itself contain slashes. This is one structural admission
  rule for all existing text readers, not a second format or persisted state.
  Rejecting every indented line is simpler but loses real #466 leaf evidence;
  name prefix alone is the reproduced defect.
  Add shared-name quote controls across the three modes, literal-slash children,
  real sibling/grandchild transitions, and the actual gate refusing a registration
  that covers only a quoted leaf. Existing three visible outcomes and recoveries
  remain unchanged. The ancestor operation is an indexed lookup plus slice
  truncate/append per record (no ancestor rescans); preserve N/2N traversal tests.

- 2026-10-19 — Separate the implementation checklist from pending workflow gates
  (the same #477 lesson): the focused-check row is complete, while the recorded
  suite, close and landing remain explicitly pending rather than falsely ticked.

- 2026-10-19 21:07 PDT — Make the hot-path cost check executable: the existing
  failure scan takes a line iterator through `failingNameLines`, with production
  using `strings.Lines`. `TestFailingNamesLineVisits` counts exactly 4N+1 visits
  for N=8 and 16 outer failures with real children and quoted helper output.
  This adds no scan or durable state and keeps the same attribution policy.


### 2026-10-19 — Revised proven-granularity implementation

The current approved goal is implemented without ancestry state. The captured
real Go buffered/verbose fixtures and actual gate controls failed red on the
old parser, then passed after coarsening raw descendants. Structured leaves stay
exact. A JSON package-fail control also reproduced false unattributability while
names were still being collected; the derivation now runs after final names.
Focused parser, real gate, saved-pass/replay, Unicode and N/2N controls passed in
33.889s. Golint and flakes passed in 0.604s and 0.056s. Documentation now describes
the accepted visible subtest-only registration miss. No passing full suite is
claimed yet; the prior queued runs were cancelled before admission.
