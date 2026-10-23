---
id: 000549
status: codecomplete
deps: []
github_issue:
created: 2026-10-22
updated: 2026-10-22
estimate_hours: 0.882
severity: medium
collaboration-mode: collaboration
card_mirror: 'e1c9dba0e192d6c05bd2fa9f43b00f6a75419ed3' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-22T13:51:39-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
actual_hours: 0.99
---

# Lint inherited temp roots used for Unix socket fixtures

## Problem

pair#792 close review BR-3 found an inverse API misuse that the shared tmp-literal rule cannot catch. A private Unix-socket RPC test used os.MkdirTemp("", "ro-") for COUCH_ISOLATED_ROOT, then derived its store with filepath.Join(root, "store") and passed that namespace to couchmessage.SocketPath. Empty-parent MkdirTemp correctly follows TMPDIR for ordinary files, but a long inherited TMPDIR makes the derived Unix socket exceed Darwin's 103-byte limit.

The failing source was cmd/internal/testcouch/agentfixture/resources_test.go in Pair. Reproduction under an existing long TMPDIR failed before starting the server with: isolated message socket exceeds 103 bytes; choose a shorter COUCH_ISOLATED_ROOT. This is a fixture-construction bug, not a production timeout or protocol failure.

Pair's BR-3 fix uses a short, canonical private socket root, prefers /Volumes/TestScratch when mounted, falls back to explicitly short /tmp, and registers cleanup immediately. Its RPC regression now also sets a long TMPDIR internally. The same external long-TMPDIR test passed with race detection in 1.229 s; the resource-observer test selection passed in 0.166 s. No production/native behavior or bound changed.

## Spec

Add a mechanical socket-fixture-root rule class to the shared Go lint machinery, or an equivalent reusable matcher that a consumer can register for its namespace APIs. Detect inherited-temp roots when the actual root/path value reaches a Unix-socket namespace or socket sink. Cover local assignments and aliases, filepath.Join-derived namespace paths, and imported API aliases; make the supported analysis boundary explicit.

Unsafe producers include os.MkdirTemp("", ...), os.TempDir(), and t.TempDir() when their resulting path is used as an unbounded socket root. Pair's concrete sink is couchmessage.SocketPath with COUCH_ISOLATED_ROOT selecting the private socket parent. Generic Unix socket APIs should use the same mistake class where their path flow can be established. Do not import consumer runtime packages into the linter just to recognize source syntax; choose shared configuration or source-level API recognition during design.

Do not ban TempDir across an entire function or file that uses sockets. The corrected Pair regression legitimately uses t.TempDir for unrelated data that constructs a long TMPDIR while its actual socket root comes from the tested short-root helper. That case must remain clean. Keep ordinary temporary files on TMPDIR/TestScratch and preserve bounded explicit-short-root exceptions with their rationale and cleanup; do not invert the existing tmp-literal policy globally.

This is an independent, unclaimed follow-up to pair#792 BR-3. Filing it does not claim that its catalog change must land before Pair's issue; that sequencing decision belongs to the hubs.

## Done when

- The original inherited-parent namespace flow is mechanically rejected, including local alias and filepath.Join variants.
- Direct inherited-temp socket paths and the supported Unix-socket/consumer namespace APIs have focused positive tests.
- The short-root fixture and unrelated TempDir usage in the same function are negative tests; no function-wide or file-wide ban produces a false positive.
- A long-TMPDIR regression verifies that the safe private socket fixture still starts the server, preserves TestScratch preference when usable, and cleans up completely without changing socket or transport bounds.
- The rule is in the shared catalog or exposed for consumer registration, with an explicit scope and guidance on the short-root exception to tmp-literal.

## Estimate

Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only; calibration is marked stale/provisional by estimate-source.

One greenfield single-concern provenance rule uses the existing stdlib AST/parser and lint engine (no new library or novel runtime). Raw design 1.0h receives the thorough-spec x0.2 discount; raw implementation 0.6h receives the v3.1 x0.4 scaling. Native fixture/catalog integration is one smaller-Go extension: raw design 0.2h x0.2 and implementation 0.4h x0.4. Atlas/lessons use raw design 0.1h x0.2 and implementation 0.1h x0.4. One close review uses raw design 0.1h x0.2 and implementation 0.3h x0.4. Familiarity 1.0 (existing AST/catalog/native Unix stack), thorough-plan design buffer 15%; no cross-repo propagation implementation, only optional read-only conformance.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: greenfield-go-module design=0.20 impl=0.24
item: smaller-go-module design=0.04 impl=0.16
item: atlas-docs design=0.02 impl=0.04
item: milestone-review design=0.02 impl=0.12
design-buffer: 0.15
total: 0.882
```

## Plan

- [x] Red-first recorded source replay and source-to-socket flow controls.
- [x] Implement the shared catalog rule and counted local provenance analysis.
- [x] Verify short-root Unix sockets under long TMPDIR, document the boundary, run recorded suite and close review.

Detailed design: [implementation plan](../plans/000549-socket-fixture-root-plan.md). TL-approved design implemented and reviewed.

## Log


- 2026-10-22: closed — Resolve main integration conflict by preserving both independent catalog registrations unchanged (test-socket-fixture-root and unit-child-process-group). Full pkg/golint PASS0.945s; recorded integrated-head full suite PASS2357 SDLC tests/76 packages/0 failures/14m0s, attempt785aff8da0351ad139dcbf4df040e59f. Original/corrected Pair source, alias/parenthesis/local receiver controls and native private Unix listener remain covered. Only issue/plan verification bookkeeping changed after original SHIP; no scope, sink policy, runtime or timeout change. TL explicitly holds merge retry until after tomorrow restart. (fixed-then-ship; full suite: Recorded suite PASS on integration e8a44c6b12cfa6253067517d64aafcdd289a6ec3: attempt785aff8da0351ad139dcbf4df040e59f, eight shards over2357 SDLC tests and76 packages,0 failures,14m0s; raw log SHA256af509f0a2f33554771315c0393cd102ce8ff958ba3a3f8536e0b9b49f2f452ab.); review verdict: SHIP
2026-10-22 20:29 PDT: Conflict-resolution full suite PASS on e8a44c6b12cfa6253067517d64aafcdd289a6ec3, attempt785aff8da0351ad139dcbf4df040e59f: eight shards over2357 cmd/sdlc tests and76 other packages, zero failures,839.6s/14m0s. Queue1307.0s; acquired20:14:27 PDT and finished20:28:27 PDT. Verified saved raw log SHA256af509f0a2f33554771315c0393cd102ce8ff958ba3a3f8536e0b9b49f2f452ab. This tests the real main integration retaining both independent catalog registrations; package-only check also passed0.945s. Next tonight: narrow re-close with --fixed-then-ship --full-suite using this proof. After accepted review, push its completion binding and a final checkpoint, stay idle, and hold merge retry until tomorrow after the planned restart per TL a0f370e1-b3b6-46a1-8473-e7a86d14149b. No check or merge runner remains from the completed runs; do not duplicate them.

2026-10-22 19:54 PDT: Authenticated TL instruction a0f370e1-b3b6-46a1-8473-e7a86d14149b approves running the conflict-resolution suite and narrow re-close tonight but explicitly holds merge retry until after the restart, first thing tomorrow. Retain only the suite already queued on e8a44c6b12cfa6253067517d64aafcdd289a6ec3: lease-247521553, attempt785aff8da0351ad139dcbf4df040e59f, tool20532, detached tree /Volumes/TestScratch/local-checks/sdlc-local-2427134384, initially FIFO2. The old merge86687 is finished, not a runner to retain or restart tonight. Exact next steps: read suite20532 result; on genuine PASS, record raw digest and run sdlc close --issue 549 --fixed-then-ship --full-suite with that evidence, push the resulting completion binding and final Log checkpoint. Then stay idle on this issue branch. After the planned restart/tomorrow continuation, first inspect queue/landing status, PR402 and the Log; with no duplicate runner, re-enter sdlc merge --yes once permitted and let its fresh-main integration checks finish. No new item or merge tonight. This paragraph is bookkeeping after the pinned suite head, not a code change.

2026-10-22 19:47 PDT: Retained merge86687/pid45523 reached admission and exited1 on an integration conflict in pkg/golint/rules.go against main36544319c0c14a714e70f729ff73d21e8dab0df6. It aborted its attempted merge cleanly. Manually merged that pinned origin/main as directed: the only conflict is independent catalog entries appended at the same location (#549 test-socket-fixture-root and #556 unit-child-process-group); preserve both unchanged. Full pkg/golint PASS0.945s with both rules registered. Staged diff whitespace check reports only main's exact captured cmd/sdlc/testdata/typed-flakes/parley-410-o-retry.log bytes, which must remain verbatim; no issue-owned whitespace error. Next: push this integration, run one recorded full suite on it, then narrow re-close with --fixed-then-ship --full-suite before any new merge entry. The previous runner has ended, so under the TL shutdown hold do not start a replacement merge without the hub's call. The old branch suite remains genuine passed evidence but does not substitute for this conflict-resolution suite.

2026-10-22 18:32 PDT: Branch suite completed PASS (attempt15fe86ae8c6cddc0f314b7ab9dea1737, raw digest97a59c19621727826cfbb456ec0974f33ac680f18a3e99105b9d60404697a778; eight shards/2290 SDLC tests/76 packages, zero failures,14m39s). All implementation-plan verification rows complete; close round2 SHIP remains accepted. PR402 has exactly one retained managed merge: tool session86687, pid45523, entered16:53:30 PDT. It is second queued behind #560, with #556 holding the landing turn and waiting on its integrated-head suite; no failure or duplicate runner. Multiple Couch restarts preserved this runner. TL's shutdown notice forbids new merges/items and requires a pushed checkpoint by21:45. No new merge is started: retain the existing waiter. Exact next step: observe this runner's admission and result; if admitted in time, let its managed integration suite/CI/landing finish, then verify PR402 merged, tracker done/archive and clean unchanged main-slot5. If shutdown interrupts it, first run sdlc queue status, sdlc landing status and inspect PR402 plus this process/lease state; re-enter only a genuinely lost runner, using the printed recovery command, after the TL lifts the hold. Do not rerun the completed branch suite or create duplicate checks. Existing branch code remains c677d1f53ba04b9fcbed79490981f9a0e0d7abfa; subsequent edits are bookkeeping only. No live fleet smoke or :0 stage.


- 2026-10-22: closed — BR-1/BR-2 reproduced RED through Check/CheckSources and fixed across binding/call/selector/type/key parenthesis boundaries and explicitly declared local testing receiver types. Corrected pkg/golint PASS1.588s and focused race PASS1.226s; original Pair source replay, native private listener and counted64/128 graph/shared-type tests pass. Three private Go overlays reintroducing each omission or disabling normalization memoization fail their intended assertions. Previous queued lease-540637535 cancelled before execution because code changed. One new recorded suite on pushed c677d1f53ba04b9fcbed79490981f9a0e0d7abfa is queued; no full-suite pass claimed, pass remains required before landing. --no-plan-check acknowledges only pending recorded-suite/close row, per tl-worker permission to overlap review with queued checks. Same TL-approved scope and private-native-only smoke, no live fleet or :0 stage.; review verdict: SHIP
2026-10-22: Filed from Pair's pair#792 BR-3 repair. Read-only inspection found no socket-fixture rule in Pair's pinned ariadne/pkg/golint or the current Ariadne source; tmp-literal addresses the opposite ordinary-file misuse. The Pair repair and its red/green verification remain on Pair's own issue branch. No Ariadne implementation, branch, or checkout edit is part of this filing.

2026-10-22: Claimed by ariadne:5 and entered planning after #552 landed. Read the shared AST-only catalog, existing temp-sweep provenance matcher, Pair's corrected resource fixture and actual SocketPath implementation. Important distinction: SocketPath hashes its namespace, while COUCH_ISOLATED_ROOT selects its socket parent; the rule must recognize the isolation-root setting as well as namespace flow. Pair's existing corrected fixture remains clean despite unrelated t.TempDir used to construct long TMPDIR. Source-level API recognition avoids importing Pair runtime packages. No code changed; plan submitted for TL approval. Initial Couch socket was briefly unavailable; subsequent canonical receipt lookup verified the dispatch.

2026-10-22: TL approved the design and native-private-listener-only smoke; no live fleet or :0 stage. Plan-quality round 1 found PQ-1 (Important): the proposed unconditional Pair namespace sink is a false positive because the namespace is hashed. Verified against transport.go:36-66. Corrected the plan to report inherited COUCH_ISOLATED_ROOT setting, keep generic actual Unix path sinks, and require long inherited namespace + bounded parent to remain clean. This preserves rejection of the original fixture at its isolation-root setting. PQ-2 (Minor) prose inventory compressed. Await TL confirmation of this design delta; implementation has not started.

2026-10-22: Plan-quality round 2 CLEAN, both PQ-1/PQ-2 addressed and no new findings. Derived the full-flow estimate only after this result, against the printed stale/provisional calibration; 0.882h itemized above. TL confirmation of the semantic correction remains pending; no implementation yet.

2026-10-22: TL confirmed PQ-1 correction via authenticated receipt 1f2465f0-878f-4cb3-9bca-e839477262fe. Full-flow change-code passed (plan-quality CLEAN, no blockers; estimate-quality INFO noted optimistic verification budget). Implemented only after these gates. Exact original/corrected Pair source blobs were retained from commits printed by Git, with provenance README distinguishing source capture from the issue-reported diagnostic.

2026-10-22: RED: TestSocketFixtureRootRecordedPairSource/unsafe produced no expected finding under the old catalog; every supported positive flow likewise failed while the corrected/unrelated-data controls passed. GREEN: shared test-only rule recognizes actual isolation-root/Unix path sinks and preserves hashed namespaces. Focused source/API/engine tests pass; full pkg/golint PASS 1.793s, focused race run PASS 1.199s. N64 counters AST1633/nodes650/edges648/propagated581/sinks128; N128 AST3233/nodes1290/edges1288/propagated1157/sinks256, within one-pass bounds including fan-out and cycle. Native listener opened an 82-byte socket under TestScratch despite 254-byte inherited TMPDIR; owned listener/root closed and removed. Preference/fallback selector covered without machine changes. No live fleet, :0 stage or Pair checkout edits.

2026-10-22: Independent Go overlay mutations: removing empty-parent MkdirTemp recognition made the original unsafe source regression fail (missing finding); widening sink taint to any inherited value made the corrected Pair source regression fail (false positive at isolation-root setting). Both failed for the intended behavioral assertion; overlays removed automatically, production source untouched. README discovery and atlas rule/boundary guidance updated. Recorded full suite and close review remain outstanding.

2026-10-22: PR402 opened; CI green. Original recorded suite lease-540637535 / attempt418e415016fda61a3514b5d211922f61 on41ca34e209e77a44661d746a1dbbb21ad78d29d7 remained queued, including through authenticated Couch restart. After over30 minutes queued, used tl-worker's permitted close-review overlap and explicitly acknowledged the pending acceptance row without claiming a suite pass. Round1 REWORK found BR-1 Critical (parenthesized binding/tuple-call normalization) and BR-2 Important (explicitly typed local testing receivers). No full-suite unverifiability finding. Cancelled exactly that superseded queued lease via sdlc local-check cancel; runner exited context-cancelled before execution, not a failed test or passing attempt.

2026-10-22: Both findings reproduced RED through actual Check/CheckSources, including transparent syntax across recognition boundaries and direct local testing types with unrelated/inferred controls. GREEN after shared per-file normalization and explicit local var type recognition. Shared wrapped-type N64: AST786/normalizations451, N128: AST1554/normalizations899, eliminating repeated type walks while preserving linear graph propagation. Plan and atlas clarify the remaining inferred receiver/type-alias boundary. Corrected head needs a fresh recorded suite and review disposition.

2026-10-22: Corrected full pkg/golint PASS 1.588s; focused socket tests with race PASS 1.226s. Go overlay mutations separately restore the binding-parenthesis hole and local receiver omission, making their new real-engine regression fail; disabling normalization cache makes the counted shared-type bound fail. All overlays are private and automatically removed. No other issue's runner or resources changed.

2026-10-22 15:10 PDT: Close round2 SHIP, BR-1/BR-2 addressed and zero open findings, accepted evidence40535f3d15e0d8a56be3289e97645a31ddb5b036. Corrected recorded suite remains outstanding on executable headc677d1f53ba04b9fcbed79490981f9a0e0d7abfa: lease-3953551684, attempt15fe86ae8c6cddc0f314b7ab9dea1737, retained tool session4029, detached tree /Volumes/TestScratch/local-checks/sdlc-local-1746181496. Authenticated Couch restart at15:09 did not lose it; readback FIFO15, queue complete=true, no requeue. PR402 has only issue/review bookkeeping after the tested code. Next: retain this waiter, read its actual result, record a genuine pass and finish the verification row; then enter sdlc merge --yes. Do not create a second suite for close-only bookkeeping while this attempt is pending. This is a normal issue Log checkpoint, not a continuation or agent restart.

## Revisions

2026-10-22 16:53 PDT: Corrected branch suite PASS on c677d1f53ba04b9fcbed79490981f9a0e0d7abfa, attempt15fe86ae8c6cddc0f314b7ab9dea1737: eight shards over2290 cmd/sdlc tests and76 other packages, zero failures, wall878.8s. Queued6530.9s; acquired16:37:57 PDT and finished16:52:36 PDT. Saved log SHA25697a59c19621727826cfbb456ec0974f33ac680f18a3e99105b9d60404697a778. Close round2 remains SHIP with no code changes since its anchor. Verification row is now complete; managed merge and its authoritative integration checks are next. Clarify the detailed plan's final row as pre-landing verification, leaving merge/report as subsequent workflow rather than prematurely checking future actions.

2026-10-22: Added the concrete implementation plan and analysis boundary without changing the requested acceptance criteria. Use the shared catalog with source-level API recognition; do not add a consumer runtime dependency or a general interprocedural analyzer.

2026-10-22 14:01 PDT: Clarify the existing Spec's socket-namespace wording: detect actual length-sensitive path/root flow, not arbitrary inherited bytes used only as a hash input. Pair SocketPath's namespace is hashed and must remain allowed when its parent is bounded; the original failing fixture is rejected at literal COUCH_ISOLATED_ROOT Setenv. Plan-quality PQ-1 motivated this correction; no acceptance scope, production bounds or timeout changes.
