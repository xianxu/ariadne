---
id: 000616
status: codecomplete
deps: [ariadne#598]
github_issue:
created: 2026-10-24
updated: 2026-10-25
estimate_hours: 1.96
card_mirror: 'c59147eacec4176c14c965d8b6871551d3bea399' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-24T23:10:54-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
actual_hours: 1.44
---

# Affected planning: traceability input and missed-mapping ledger

## Problem

ariadne#598 M1 gave affected.Plan a curated, declared ownership mapping: `owners` rules, `full` excepts and sidecar conventions, compiled into one `affected.Mapping` read from main. The operator's direction, relayed by ops:0 on 2026-10-24, asked for two more parts, which were moved out of #598 so M1 could land for pair#853:

1. Parley-style traceability files (`atlas/traceability.yaml`: `atlas: {key: {code: [...], tests: [...]}}`) as an input format: "change the toolbar, run these".
2. Missed mappings: whenever the integrated full suite at merge catches a failure the affected selection would not have run, record it and list it for curation.

The operator said go (relayed by ops:0 on 2026-10-24), with the plan already approved as #598's.

## Spec

The design is approved as part of #598's durable plan: `workshop/history/plans/000598-affected-declared-input-ownership-plan.md` (sections M2 and M3, Part C, the ledger's lifecycle and ordering). In summary:

- **Traceability input.** A new `traceability` member in affected.json lists repo paths of traceability files, read from main's pinned rev through the existing `loadAffectedMapping` / `gitx.TreeFilesSelected`.
  - Each entry compiles to one `ownerRule{paths: code, as: tests}`; `Compile` takes a sources struct instead of growing a parameter.
  - A Go test path selects its package; a non-Go test selects the group whose `paths` match it.
  - A test path that claims nothing goes full, naming the entry.
  - Parsing is strict (yaml.v3, already a dependency), with parley's real file as a golden.
- **Missed mappings.** Hook into `diagnoseSuiteFailure` (`cmd/sdlc/suitebaseline.go`) after `introduced` is classified, only when the main baseline passed.
  - Plan the branch's delta with main's config and mapping, plus the head graph.
  - For a non-full selection, record each introduced Go-test failure whose package was not selected. The package comes from `flakes.SplitTest`/`ParseName`; bare names are not misses.
  - Do the same in the train attribution path for a culprit car whose affected record missed the package.
  - The ledger lives in the common git dir: versioned, unknown members preserved, flock plus atomic rename, deduped by (package, changed-path set), capped at 200.
  - `sdlc affected missed list|dismiss`; `list` drops entries that main's mapping now covers.
  - Recording never changes a gate's outcome.

## Done when

- A traceability file declared in affected.json and read from main makes a change to an entry's `code` path select what its `tests` paths select: a Go test file selects its package, and a non-Go test selects the group whose `paths` match it. An entry whose test path claims nothing goes full, naming the entry. Parsing is strict, and parley's real `atlas/traceability.yaml` parses as a golden.
- When the merged-head suite fails at merge and main's baseline passes, each introduced Go-test failure whose package the branch's affected selection would not have run is recorded as a missed mapping. So is a train failure attributed to a car whose affected record missed the package.
- Nothing is recorded for a red main baseline, a full selection or a bare test name, and recording never changes the gate's outcome.
- `sdlc affected missed list` shows open entries and drops those that main's mapping now covers; `dismiss` removes one. The ledger is deduped, capped and locked, and keeps unknown members on rewrite.

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: smaller-go-module      design=0.1  impl=0.2
item: smaller-go-module      design=0.05 impl=0.1
item: smaller-go-module      design=0.1  impl=0.24
item: smaller-go-module      design=0.1  impl=0.28
item: smaller-go-module      design=0.05 impl=0.16
item: atlas-docs             design=0.05 impl=0.06
item: milestone-review       design=0.0  impl=0.32
design-buffer: 0.15
total: 1.96
```

The rows are:
- the traceability parser, config member and `Sources`;
- the codec move to jsonstrict;
- `Missed` and the pure ledger;
- the store, the two hooks and the `affectedSelectionAt` extraction;
- the `affected missed` CLI;
- the atlas;
- two boundary reviews.

Familiarity is 1.0 because #598 M1 built the mapping these extend.

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only.*

## Plan

Durable plan: `workshop/plans/000616-affected-traceability-missed-mappings-plan.md`.

- [x] M1 — traceability input: `traceability` member, `ParseTraceability` (strict yaml.Node walk, parley golden), `Compile(cfg, Sources)`, loader reads from main; the shared strict JSON object codec moves to `internal/jsonstrict`.
- [x] M2 — missed-mapping ledger: `Missed`, `MissedLedger` (pure, preserving), the locked store in the common git dir, the merge-diagnosis and train-split hooks via the extracted `affectedSelectionAt`, `sdlc affected missed list|dismiss`.

## Log



- 2026-10-25: closed — Close-review Minors fixed in bfc0039e1: Record counts one hit per observation per entry (TestMissedLedgerOneSightingPerObservation), list --json writes its pruning notes to stderr (TestAffectedMissedListAndDismiss asserts it), list marks a truncated entry and the atlas documents that pruning plans only the 20 stored paths. go test ./cmd/sdlc/internal/affected and the TestAffectedMissed*, TestRecordSplitMisses, TestDiagnoseSuiteFailureRecordsMissedMappings, TestMissedMergeRecorderRealGit runs pass; full suite green on 2f2c69c5 before the fix. (fixed-then-ship); review verdict: SHIP
- 2026-10-25: closed — M2 (missed-mapping ledger): pre-close suite green on pushed code head 2f2c69c5 (sdlc-local/suite, 8 shards, 2507 cmd/sdlc tests, 0 failed); 1202e2c2f only ticks the plan row. Targeted: affected (TestMissed, TestMissedLedgerSequence, TestMissedLedgerPathCap, TestMissedLedgerCodec, FuzzDecodeMissedLedger seeds, traceability golden/refusals/fuzz), jsonstrict, gitx TestTreeFilesAt, TestAffectedMissedListAndDismiss, TestDiagnoseSuiteFailureRecordsMissedMappings, TestMissedMergeRecorderRealGit, TestMainMissedCoverageRealGit, TestLandingTrainSplit* x3 and TestRecordSplitMisses, TestAffectedDeclaredTraceability*. The earlier pre-close on 62b3d29c failed on BR-16 (fixed on main by #620, now integrated) and a load-timed train barrier (passes 3/3 alone).; review verdict: SHIP
### 2026-10-24
- 2026-10-24: closed M1 — targeted: go test ./cmd/sdlc/internal/affected ./cmd/sdlc/internal/gitx ./cmd/sdlc/internal/jsonstrict ./cmd/sdlc/internal/landingline pass; go test ./cmd/sdlc -run Affected|SplitImpact|Attribution pass (180s), incl. real-git traceability-from-main, missing and malformed on main in ordinary and boarding modes; parley golden (ee938b5a) parses; FuzzParseTraceability 20s and FuzzDecodeObject 15s clean; BenchmarkPlanDeclared with 65-entry traceability 32 ms/op; drop-cost guard passes with the moved codec; review verdict: SHIP

Split from ariadne#598 by ops:0's direction: M1 landed alone so pair#853 could proceed. These were #598's M2 and M3. The approved design is in #598's durable plan, and M1's review notes for them are in that plan's Revisions (for example, `Compile` taking a sources struct).

Claimed and started after #598 landed (PR 435, 2001041f0). The durable plan restates #598's M2/M3 against the code as it now stands, with these refinements:
- `Missed` also skips selections that have groups, matching `splitImpact`'s uncertainty rule.
- The train hook reuses `ImpactUntouched` rather than re-planning.
- `affectedSelectionAt` is extracted from `splitImpact`, so the merge hook and `list` share it (ARCH-DRY).
- landingline's strict, preserving JSON codec moves to `internal/jsonstrict` so the ledger can reuse it.
- `record` sweeps stale temp files.

- M1 built:
  - the codec move (7a4910cec);
  - the traceability input (cb8e256d7), with the parley golden (ee938b5a), a refusal table and a fuzz target (20s, clean);
  - `Compile(cfg, Sources)`;
  - `gitx.TreeFilesAt`;
  - real-git loader tests, both modes.
- Benchmark with a 65-entry traceability file: 32 ms/op, unchanged from M1 of #598.
- Estimate gate: INFO. The reviewer thought the store/hooks and CLI rows look light; the estimate is kept at 1.96 and the actual will say.
- 2026-10-25: close review SHIP (window 8f3eec5e..1202e2c2) with three Minors, fixed in the same round: Record counts one hit per observation per entry (TestMissedLedgerOneSightingPerObservation); list --json writes its notes to stderr; list marks a truncated entry and the atlas says pruning plans only the stored 20 paths.
