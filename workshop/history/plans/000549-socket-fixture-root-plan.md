---
collaboration-mode: discussion
---

# Socket fixture root lint implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3 for execution strategy. SDLC owns the fresh-context review boundary. This work remains in the current issue branch; do not start implementation before TL approval and change-code.

**Goal:** Reject inherited temporary roots that visibly reach Unix socket paths or Pair's isolated socket parent, without rejecting unrelated temporary data.

**Architecture:** Add one test-only AST rule to the shared Go lint catalog. Build a command-local dependency graph of local path values, propagate inherited-temp provenance once, and inspect recognized socket sinks. Existing Check/CheckSources own parsing, findings and suppressions; no type checker, consumer import, external analyzer or persistent state is added.

**Tech stack:** Go standard-library AST/parser, existing pkg/golint catalog, private native Unix listener fixtures.

## Alternatives and decision

1. Recommended: shared catalog rule with exact source import/API recognition. Existing downstream catalog users receive it when they update Ariadne; import aliases and shadowed names are resolved structurally.
2. Consumer-configured sink registration adds an API and adoption step for the one known namespace API. Defer that axis until a second custom API needs it; standard net APIs and the existing Pair import path are enough here.
3. Function-wide TempDir prohibition is smaller but rejects the motivating corrected fixture. Whole-program SSA would recognize opaque helper returns but exceeds this bounded mechanical rule's purpose and dependencies.

## Core concepts

All named functions below are proposed implementation names, not claims they already exist.

| Pure entity / risky function | Lives in | Status | Adversarial strategy and independent guard |
|---|---|---|---|
| socketFixtureRoots | pkg/golint/socket_root.go | new | Replay recorded unsafe/corrected source through actual Check and CheckSources; assert finding locations, not only counts. Catalog registration and TestAtlasNamesEveryRule enforce discovery. |
| analyzeSocketRoots | pkg/golint/socket_root.go | new | Actual AST-to-graph analysis entry used by the rule and counted N/2N tests; graph propagation and sink controls exercise it directly. |
| socketRootGraph / propagateInheritedRoots / unparen | pkg/golint/socket_root.go | new | Aliases, multi-result MkdirTemp, branch writes, shadowing and cycles; explicit graph and shared-type normalization counters at N and 2N reject repeated whole-graph or parenthesis rescans. |
| socketRootSink | pkg/golint/socket_root.go | new | Exact import identity, renamed imports, unrelated same-name methods, TCP controls and independent isolation parent/hashed namespace; API matrix tests pin actual length-sensitive argument selection. |
| Catalog | pkg/golint/rules.go | modified | Shared production catalog rejects original misuse; existing tmp-literal ordinary-file controls remain unchanged. |

One graph owns all value provenance for a parsed file. Nodes represent AST expressions/local bindings; edges mean a result can contain a parent's inherited path. Reuse localName/stringLit and existing parser object identities. Do not extract the different temp-sweep safety proof into a generic framework merely because both inspect assignments.

| Integration seam | Lives in | Status | Wraps and verification |
|---|---|---|---|
| Check / CheckSources and tracked-file catalog test | pkg/golint/check.go, repo_test.go | existing, reused | Source bytes -> parsed AST -> findings; no new IO API. Exercise both entry points and the tracked repository sweep. |
| TestSocketFixtureShortRootWithLongTMPDIR | pkg/golint/socket_root_test.go | new test only | Real private directories and net.ListenUnix; long inherited TMPDIR, TestScratch preference when mounted, short /tmp fallback and close/remove verification. No Couch supervisor, live socket, production store or private fleet harness. |
| Recorded source fixture | pkg/golint/testdata/socket-root/ | new | Retain original failing pattern, corrected Pair source excerpt and reported real error with provenance. Replay source fixtures in ordinary rule tests. Obtain the original capture from the owner if available; do not present reconstructed bytes as a verbatim capture. |

## Supported analysis boundary

- Test-only rule `test-socket-fixture-root`, incident ariadne#549 / pair#792. Generated files and standard reasoned suppressions follow the existing engine.
- Inherited producers: os.TempDir(), os.MkdirTemp with an empty or inherited parent, testing T/B/F/TB TempDir calls whose receiver binding/type is visible. Multi-result declarations bind only the path result. Imported package aliases are honored; shadowed package identifiers do not match imports.
- Propagate through local declarations/assignments and aliases, parentheses, string concatenation, filepath.Join, Clean, Abs and EvalSymlinks (first result), plus UnixAddr.Name composite literals and local aliases of those addresses. An explicit parent that is not inherited does not acquire taint merely from an unrelated TempDir elsewhere.
- Join all visible writes to a binding: any inherited write can taint a sink, including conditional writes. This is conservative syntax analysis, not a proof of execution order. Document and test its suppression escape for a proven overwrite.
- Generic sinks: net.Listen, ListenPacket and Dial with a literal Unix network (unix/unixgram/unixpacket as appropriate), ResolveUnixAddr, ListenUnix, ListenUnixgram and DialUnix. Read only their actual path/address arguments; TCP and unrelated arguments remain clean.
- Pair sink: os.Setenv or testing Setenv of literal COUCH_ISOLATED_ROOT, the actual length-sensitive socket parent. Pair's SocketPath hashes its namespace into a fixed-length leaf (pair/cmd/internal/couchmessage/transport.go:36-66); namespace length is not a sink. A long inherited namespace with an explicitly bounded or default parent remains clean. The recorded original regression is rejected at its inherited isolation-root Setenv even when its namespace passes through Join and aliases. Never interpret ordinary TMPDIR Setenv as a socket sink.
- Helpers across function boundaries, interface-dispatched methods, arbitrary container mutation, reflection, dynamic network/env keys and opaque return values are outside this first syntax-only boundary. Unknown provenance produces no finding, never a claim that the path is safe. Function-value aliases are outside scope unless their direct imported selector can be resolved without adding a second analysis mechanism. State this visibly in atlas/tests.
- The rule reports at the recognized sink, so suppressions remain local and actionable. A message explains short, canonical, private roots and immediate cleanup; existing tmp-literal exceptions need a reason and remain exceptional.

## Cost, state and lifecycle

Batch workload: per file, N AST nodes, E expression/binding dependency edges and S sinks. One AST index pass, a worklist where a node transitions from untainted to tainted at most once, each outgoing edge visited at most once, one sink pass: O(N+E+S) time and storage. No per-sink whole-file scans, recursive alias rescans or repeated fixed-point scans. Counters in the actual analyzer count AST/node/edge/sink visits for 64 and 128 aliases, shared fan-out and cycles; assert a constant per node/edge bound rather than elapsed time. Parser cost stays owned by Check/CheckSources.

Actions: edit the fixture or add a reasoned suppression, rerun lint.

| Internal outcome | Visible state | Recovery action | Possible loss |
|---|---|---|---|
| Proven inherited root reaches recognized sink | Finding with source location and incident | Edit fixture or justify suppression, rerun | None; source is never rewritten |
| No proven flow, including unsupported/opaque flow | No finding from this rule | Review documented boundary; rerun after edits | Possible missed misuse, explicitly not a safety proof |
| Invalid source or invalid suppression | Existing parser/suppression error | Correct source/suppression, rerun | None |

There is no durable runtime state, user cancellation protocol or external event ordering in this pure analysis. Analysis memory dies with the check. Native tests own their directory/listener lifetimes lexically, close listeners before removing roots, and verify removal even on test failure. No credential or real fleet state is consumed. No persisted format changes.

## Implementation and verification steps

### 1. Recorded red and behavior controls

Files: create pkg/golint/socket_root_test.go and testdata/socket-root source/evidence fixtures.

- [x] Obtain/record provenance of Pair's original failure and corrected source. Reproduce the original source shape in the normal Check path; expect a finding at its isolation-root/namespace sink, initially absent under the old catalog.
- [x] Add adversarial tests from the pure-function strategy table and supported-analysis contract; executable tests own the concrete cases and finding locations.
- [x] Run `go test ./pkg/golint -run 'TestSocketFixtureRoot' -count=1 -v`; retain meaningful RED before implementation.

### 2. Pure rule and catalog integration

Files: create pkg/golint/socket_root.go; modify pkg/golint/rules.go.

- [x] Implement source/API indexing, path dependency graph and one-way worklist propagation; report only recognized sink nodes with proven inherited input. Reuse the existing lint parser and suppression engine.
- [x] Run the red selection to GREEN; exercise both Check and CheckSources and all supported cases. Add counted growth tests through the actual analyzer and verify N/2N bounds, shared fan-out and cycles.
- [x] Mutation check: remove producer recognition or a required propagation edge and confirm the recorded positive regression fails; add a broad function-level taint and confirm corrected/unrelated-temp negative controls fail. Restore code before committing.

### 3. Native proof, documentation and close

Files: pkg/golint/socket_root_test.go; atlas/workflow/sdlc-binary.md; workshop/lessons.md; README.md only if its existing lint discovery pointer needs changing.

- [x] Add the private native listener regression under a real long TMPDIR, selecting a canonical short TestScratch root when available or explicitly short /tmp otherwise, immediate cleanup, successful connect/accept and verified root removal. Test selection/fallback logic independently without changing machine mounts or settings. Retain existing 103-byte socket bounds; no transport/timeout changes.
- [x] Run `go test ./pkg/golint -count=1` and the targeted native test with `-race`. Under TL's native-private-listener-only smoke decision, cross-repo coverage is the recorded original/corrected Pair source replay; do not execute Pair or use live Couch.
- [x] Document the rule, exact recognized APIs, conservative write join, unsupported helper boundary and short-root/tmp-literal rationale. Check the existing README lint discovery pointer; update atlas's derived catalog coverage and lessons.
- [x] Commit/push final code, run `sdlc local-check run --name suite`, record evidence and close with SDLC's mandatory review. Branch suite PASS (attempt15fe86ae8c6cddc0f314b7ab9dea1737); close round2 SHIP.

Following verification, enter managed merge and report its observed outcome. No separate review agent or live :0 stage is proposed for a pure source lint plus native private-resource tests; TL decides smoke.

## Architecture lenses

ARCH-DRY: one catalog rule and provenance graph, reuse parsing/import/suppression ownership; no Pair runtime dependency or duplicate consumer lint. ARCH-PURE: analyzer is pure, filesystem/listener work is test-only. ARCH-PURPOSE: cover the original isolation parent plus generic sockets, retaining corrected source and bounded-parent hashed-namespace controls. ARCH-MOCK: no new production external dependency; native kernel conformance uses private resources rather than live fleet. ARCH-CONSTRAINTS: O(N+E+S), counted growth, no subprocess per source. ARCH-SECURE: parse bytes without execution; exact imported API identity; private owned native roots only. ARCH-ORDER: no cross-event state; worklist is monotone and cycle-safe, test IO lifetimes explicit. ARCH-FUNERAL: no runtime artifacts; test listener/root cleanup verified, fixture bytes remain bounded version-controlled evidence. ARCH-EVOLVE: no persisted format or configuration changes; additive catalog entry uses existing finding/suppression contracts.

## Revisions

2026-10-22 16:53 PDT — Branch suite passed all eight shards over2290 SDLC tests and76 other packages in14m39s, zero failures. Close round2 SHIP already accepted. Separate completed pre-landing verification from subsequent merge/report workflow so a checked row does not claim a future landing. No implementation or scope change.

2026-10-22 14:01 PDT — PQ-1: correct the sink-contract class by recognizing actual path parents, not hash inputs. Remove the unconditional Pair namespace sink, retain inherited isolation-root Setenv and generic Unix path sinks, and require a bounded-parent/long-namespace negative control. All other analysis bounds and private-native verification remain as approved. PQ-2: compress repeated implementation test inventory into the function strategy/contract references. Await TL confirmation of this correction before code.

2026-10-22 14:04 PDT — TL confirmed PQ-1 correction and native-only smoke. Plan-quality CLEAN with prior findings addressed, estimate recorded and change-code passed. Implementation names the single graph-construction entry analyzeSocketRoots, now included in Core concepts; no design or scope change.

2026-10-22 — Verification follows TL's approved native-private-listener-only smoke: exact original/corrected Pair Git blobs run through the source lint, while the owned native listener proves path length and cleanup. The optional read-only Pair RPC run is superseded by this explicit smoke decision, not claimed as executed.

2026-10-22 14:48 PDT — Close BR-1/BR-2: normalize transparent parentheses consistently at binding, first-result, selector/receiver/type and literal-key recognition boundaries. The same per-file graph owns normalization memoization, with a counted shared deeply wrapped receiver type at N/2N. Recognize explicit testing receiver types from parameter fields and local var declarations; inferred receiver aliases and named type aliases remain outside the documented syntax-only boundary. This corrects promised coverage rather than adding interprocedural inference or changing sink policy. Original queued suite was owner-cancelled before execution because corrected code needs a new recorded head.
