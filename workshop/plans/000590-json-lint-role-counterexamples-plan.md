# Qualify recursive JSON scanner roles — Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Accept Pair's independent differential oracle and standalone DOM materializer while continuing to flag redundant recursive duplicate validators.

**Architecture:** Keep the existing candidate matcher, then apply narrow syntax-based role qualification from one per-file AST summary. No new suppression annotation, dependency, compiler/type checker or broad test-file/package exemption. Unknown roles remain findings, with remediation describing the consumer module's own shared authority.

**Tech Stack:** Go AST; existing pkg/golint Check / CheckSources and catalog tests.

## Findings and proposed boundary

The current recursiveJSONScanners matcher in pkg/golint/rules.go equates any recursive Token + More + map[string]... function with duplicate-validator authority. It neither follows what the map produces nor sees why a test helper exists. Read both live Pair counterexamples, including their current temporary allows, from pair-slot3 at Git HEAD 50a6a163f9551bd4e5a7ebfa5746eba0e7d83eda. No peer files changed.

1. cmd/internal/strictjson/decode_equivalence_test.go: legacyScanValue returns only error; its wrapper legacyRejectDuplicateKeys is used by FuzzRejectDuplicateKeysEquivalent to compare acceptance against a distinct optimized scanner. This role must remain independent; replacing it with the implementation under test invalidates the oracle.
2. examples/mock-to-real/internal/jsonvalue/value.go: value returns (any,error), builds map[string]any and []any from recursively decoded children and returns them. Duplicate rejection is part of materialization; routing through Pair would violate the standalone nested module's ownership.

Recommend role qualification rather than (a) exempting all *_test.go or packages, which hides real duplicate validators, or (b) replacing the two allows with a new role annotation, which merely makes the exception another per-site claim. A general interprocedural/type/alias analysis is too much mechanism for this syntactic catalog. Support the witnessed local forms, retain findings for unresolved/escaping/ambiguous shapes, and state that limit.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|---|---|---|
| recursiveJSONScanners | pkg/golint/json_scanner.go (move from rules.go) | modified |
| jsonScannerSummary | pkg/golint/json_scanner.go | new |
| jsonMaterializerRole | pkg/golint/json_scanner.go | new |
| jsonDifferentialOracleRoles | pkg/golint/json_scanner.go | new |

The summary is owned by one rule invocation, indexes declarations/local calls/returns once, and dies on return. ARCH-PURE: all role decisions consume AST only. ARCH-DRY: Check and CheckSources use the same rule and role predicates; no downstream copy or new parallel scanner authority.

- recursiveJSONScanners retains current direct function/method/assigned-closure candidate detection and existing owner behavior. It reports a candidate unless a specific positive role is established. Existing positive fixtures remain unchanged and still report; a function name, filename or package name alone never establishes a new role.
- jsonMaterializerRole requires positive value flow: a local container is populated with recursive child results and that container escapes through a value return. Returning (any,error), declaring a map, returning nil, or keeping a key set alone is insufficient. Cover map assignment and slice append forms present in the capture; aliases/unresolved flow remain findings. A schema-specific nonrecursive decoder stays outside the existing candidate shape as today.
- jsonDifferentialOracleRoles operates only on test-source declarations reached from an actual testing.F.Fuzz callback. Recognize the captured acceptance comparison between error results of two distinct decoder calls on the same input. Follow the local wrapper/helper path to the recursive oracle; reject a self-comparison or an oracle whose observed uses escape the differential helper tree. No exemption merely because a function is in a test or its name includes oracle/legacy. Bare local names are resolved against local declarations and bindings; ambiguous/shadowed selectors or aliases do not establish this role. This is a bounded syntactic recognition of the observed differential idiom, not a proof of arbitrary program equivalence. Broader call/alias forms remain explicit findings; don't grow a compiler to recognize them.

### Integration points

| Name | Lives in | Status | Wraps |
|---|---|---|---|
| Catalog diagnostic | pkg/golint/rules.go | modified | existing lint finding rendering |
| Check / CheckSources | pkg/golint/check.go | unchanged | parsed caller-supplied sources |
| Pair source replay fixtures | pkg/golint/testdata/json-scanner-roles/ | new | recorded peer source bytes |

No new external service, subprocess, machine setting, persisted format or resource lifetime is introduced. Capture source bytes and provenance at design/implementation time; tests read committed fixtures, never the live Pair checkout. The new fixture family is fixed-size committed regression evidence, not per-run accumulation (ARCH-FUNERAL). ARCH-EVOLVE: no persisted schema changes or semantic changes to existing JSON decoders.

Change the diagnostic to recommend consolidating redundant validation behind the **consumer module's shared strict JSON authority**. Ariadne's internal jsonstrict can be identified as its implementation/reference, never as an import available to downstream modules. Do not impose Pair's exact-key/Unicode/float/number semantics on another owner.

## Visible states

Actions: consolidate redundant validation; retain the proven role; inspect unresolved role.

| Internal role / failure mode | Visible state | Recovery action | Possible loss |
|---|---|---|---|
| Existing scanner candidate, no positive role | Finding with module-accessible remediation | Consolidate redundant validation | None; code unchanged |
| Witnessed value-producing materializer | No finding at that declaration | Retain the proven role | No independent materializer lost |
| Witnessed differential oracle | No finding at that declaration | Retain the proven role | No independent test oracle lost |
| Ambiguous, escaping, shadowed or unsupported role flow | Existing finding remains | Inspect unresolved role | Possible conservative false positive, never automatic code replacement |

No lock/lease applies: the invariant is about immutable AST inputs, not temporal absence. No runtime decoder or lint result store is added.

## Cost contract

Growth drivers are A AST nodes, F local functions and E local call/reference edges per file. Build one shared summary, traverse each function body a bounded number of times, and use visited sets for wrapper traversal; total O(A+F+E), with no whole-file rescan per candidate. The role helpers share memoized candidate classifications. Add internal counted traversal hooks used by tests, not global metrics. At8/16 independent scanner/helper groups assert bounded node/edge visits proportional to input and reject nested repeated whole-file scans. Timing is supplementary only. Do not build transitive sets per function that multiply memory by F.

## Steps and adversarial strategies

One atomic close boundary, no Mx tags. Hold before change-code until ops:0 approves the plan.

- [x] Capture the two exact Pair source files and source provenance; strip only the temporary json-recursive-scanner allows in test input, leaving the original capture available. Add source replay through both Check and CheckSources: both currently report and must turn green. Existing true-positive table/method/closure fixtures remain rejected unchanged.
- [x] jsonScannerSummary / recursiveJSONScanners: add red role controls for ordinary test helpers, production validators, fake Token/More receivers, shadowed recursive names, assigned closures and methods. Build one per-file summary and preserve the candidate boundary. Count8/16 groups and cycles with visited edges; prove each source/body index is built once. Unknown role evidence must not erase a current positive fixture.
- [x] jsonMaterializerRole: add red/green controls for returned map and slice child values, exact Pair capture, and decoys (unused map, returned nil, only key-set bookkeeping, scalar/error result, recursive result not stored, alias uncertainty). Independent guard: a materializer can keep its duplicate-rejection logic; the linter never changes decoder semantics or dependencies.
- [x] jsonDifferentialOracleRoles: test the full captured fuzz wrapper/callback path against two distinct scanners. Negative controls compare the same scanner twice, omit comparison, pass different inputs, use a non-testing callback, add an escaping/ordinary caller, shadow a binding, or put the identical validator in production source. Positive qualification is function-specific; other candidate validators in the same test file still report.
- [x] Diagnostic tests: prohibit advice to import cmd/sdlc/internal/jsonstrict downstream; require the module-owned shared-authority recovery. Existing line-precise suppressions still report unused allows after these sites become accepted, so notify Pair to remove exactly its two annotations after landing.
- [x] Run focused role tests red before implementation, then go test ./pkg/golint -count=1, tracked-Go catalog and counted controls; add a lessons.md rule about classifying purpose from value/caller evidence rather than container syntax. Update atlas/workflow/sdlc-binary.md's existing catalog description; its index link already exists.
- [ ] Commit/push; central sdlc local-check run --pre-close, mandatory close review, then sdlc merge --yes after the recorded pass. No :0 staging or live fleet mutation: this is pure syntax plus real peer-source replay. Report the landed revision to hub and pair:3 so pair#861 removes its two temporary allows.

## Revisions

- 2026-10-24: Initial plan after reading both live Pair captures and the matcher. Choose positive structural role evidence over blanket test/package exemptions or replacement waiver annotations. Await hub approval before implementation; #589's existing recorded suite remains retained and native qualification held.

- 2026-10-24: Ops:0 approved the plan (delivery 5cd939f8-dc28-440c-8b78-efec55cf79a4): positive returned-container or witnessed differential-fuzz role only; unknown/escaping/shadowed roles remain findings; existing true positives unchanged; no waiver annotation; consumer-owned strict-authority diagnostic. Proceed to change-code.

- 2026-10-24: Implementation keeps per-lexical-function use/write evidence in the shared summary. Named/assigned closure self bindings must remain stable; container writes/escapes invalidate materialization proof. Bounded feature-bit propagation preserves nested-closure validator detection without copied transitive call sets. Reused assignedRoot for transparent write/address syntax, following #549's lesson. These implement the approved unknown/shadowed/escaping boundary; no new waiver or role mechanism.
