---
id: 000590
status: codecomplete
deps: []
github_issue:
created: 2026-10-24
updated: 2026-10-24
estimate_hours:
card_mirror: 'd47b761c2ddebff63b260983728c42b56abb6297' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-24T04:36:31-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
actual_hours: 1.20
---

# Qualify recursive JSON lint against independent oracles and DOM materializers

## Problem

Ariadne pkg/golint json-recursive-scanner reports proven false positives in Pair while adopting 5a17f789619d for pair#838. The matcher treats recursive Token + More + any string-key map as duplicate-validator authority. It flags both an independent differential-test oracle and a standalone DOM materializer. Its advice names cmd/sdlc/internal/jsonstrict, which downstream Go modules cannot import.

## Spec

Qualify/narrow the matcher using the two counterexamples without weakening actual duplicate-validator detection or adding broad filename/package exemptions. Clarify downstream-accessible remediation. Preserve the independent oracle and standalone sample boundaries; consumers must not delete independent checks or add unwanted production dependencies to satisfy a rule.

## Done when

- Fixtures reproducing both false positives are accepted, while current true duplicate-key scanner fixtures remain rejected.
- Diagnostics recommend an accessible shared authority or accurately describe ownership instead of an internal Go import inaccessible downstream.
- Counted/existing matcher tests and required checks pass; landed revision sent to pair:3 so exactly two temporary per-site allows in pair#861 can be removed.

## Plan

Durable plan: [JSON scanner roles](../plans/000590-json-lint-role-counterexamples-plan.md). Approved by ops:0 on 2026-10-24; enter change-code before implementation.

- [x] Approved role boundaries; captured both Pair counterexamples before implementation.
- [x] Implement positive materializer/oracle role qualification with unchanged true-positive controls, accessible remediation and counted scan tests.
- [x] Replay real captures through Check and CheckSources and update atlas/lessons guidance.

Delivery gates remain: recorded pre-close result, mandatory close review, landing, then notify pair:3 to remove its two temporary allows. No full-suite pass is claimed while queued.

## Log

Pair#861 qualification found51 new catalog findings. Three actual duplicate validators will be consolidated; these two are different roles:

1. cmd/internal/strictjson/decode_equivalence_test.go legacyScanValue (line46), used exclusively by legacyRejectDuplicateKeys and FuzzRejectDuplicateKeysEquivalent. It is the independent legacy Token oracle for an optimized byte scanner; replacing it with the implementation under test makes the differential test tautological. Semantics include exact key equality, invalid Unicode replacement and float overflow.
2. examples/mock-to-real/internal/jsonvalue/value.go value(d *json.Decoder, depth int) (any,error) (line28), a generic recursive DOM materializer returning map/array/scalar values with json.Number, exact keys and depth64. Nested module example.com/mock-to-real deliberately has no production Pair dependency (see catalogue/catalogue.go concluding boundary note).

Pair retains the complete observed lint log in its #838 evidence and filed adoption #861. TL explicitly authorized reasoned per-site lint:allow only at these two declarations, naming this upstream issue, pending its correction. These are temporary proven-false-positive annotations, not blanket waivers. Please notify pair:3 on landing.

### 2026-10-24 — claimed, investigated and planned
- 2026-10-24: closed — Ops approved reduced scope in final issue/plan Revisions: original matcher restored unchanged, no role-proof exemptions; only consumer-owned diagnostic and sanctioned reasoned per-site allows ship. Exact Pair captures report without their allows and pass with them through Check and CheckSources; existing validator tests unchanged. Full golint passed1.293s and race10.099s including catalog; atlas and lessons updated; pair:3 receipt submitted. Fresh recorded pre-close lease-792232057 queued; wait for pass before landing. (done-when: contract moved since change-code; judged by the close review); review verdict: SHIP
- 2026-10-24: flow upgraded quick → full — an earlier round of this close already ran the full review

Claimed and started from fresh main on 000590-json-lint-role-counterexamples, leaving #589's clean pushed branch intact. Its recorded pre-close suite remains retained in session31958, lease-355567331; no restart/cancel. Read current matcher and both real Pair counterexamples at Git HEAD 50a6a163f9551bd4e5a7ebfa5746eba0e7d83eda. Root cause: Token+More+self-recursion+any string map is treated as duplicate-validator authority, regardless of returned values or differential-test purpose. No peer edits, new suppression, implementation or runtime JSON policy changes made. Proposed shared bounded AST role qualification and module-owned diagnostic remediation; sent to hub before code.

### 2026-10-24 — approval

Ops:0 approved the proposed role boundaries (delivery 5cd939f8-dc28-440c-8b78-efec55cf79a4). Plan report receipt b87c22ef-0eb5-4cb5-a657-b92caf348a5e is submitted. #589 retained suite acquired capacity after1053s and is now running in its detached checkout; no restart or cancellation.

### 2026-10-24 — first implementation checkpoint

change-code completed successfully. Captured immutable Pair source files (original allows retained in captures; replay strips only the two relevant annotations). Observed both real captures fail before implementing; both now pass Check and CheckSources. Existing scanner/method/closure fixture expectations remain unchanged and pass. Negative controls retain findings for self-comparison, different input, absent comparison, non-testing fuzz receiver, ordinary caller and escaped helper. Materializer controls caught and fixed overwritten map/slice values and container aliases/address/call escapes; no alternate production dependency or waiver annotation added. A red nested-closure validator control caught a candidate regression; bounded lexical feature propagation now preserves it.

One lexical AST summary visits each node once:8/16 generated helper groups have421/765 AST nodes, with217/433 additional callback-node visits and62/118 graph visits; shared helper paths are visited with memoized membership rather than rescanned per root. Focused TestJSON/TestRules currently pass. Updated remediation names the consumer module's own shared strict decoder; inaccessible internal import advice is gone. Full pkg/race/catalog checks, atlas/lessons, pre-close and boundary review remain. #589's original detached suite continues running and native allocation qualification remains held.

### 2026-10-24 — role controls and docs complete

Full golint package passed1.246s and race8.741s at the first checkpoint. Additional red controls then caught rebound or address-exposed closure self bindings, including parenthesized assignment/address forms; shared summary binding evidence now invalidates those roles and reuses assignedRoot for normalization. Existing package-wide scanners/closures/methods remain checked; no new blanket exemption. Atlas and lessons explain role qualification, conservative limits and consumer-owned remediation. Final full golint race run passed9.822s, including the tracked catalog and all role controls; then push and record pre-close on the final head. #589 suite remains retained separately.

### 2026-10-24 — retained pre-close and review preparation

Central pre-close on0f910acca is retained as shell session8547, lease-459019387, attempt85a587db0c0e9932d4116f2040ad2d38. Detached checkout /Volumes/TestScratch/local-checks/sdlc-local-3525614321; raw log /Users/xianxu/workspace/ariadne/.git/sdlc/test-stats/inflight/85a587db0c0e9932d4116f2040ad2d38/run.log . Initially queued position5 behind other fleet work; keep it. The duplicate-suite notice names #589's different-issue lease-355567331: preserve both; these are separate issues, not replacement runs.

Actual preview succeeded on the branch through active-time-v3-branch-v1, with witnessed resting exclusions labelled; close will measure/adopt normally, no typed actual. Central binary version reports b3a1850018e1664c2dc38b21e7075a77fa748d8d, modified=false. Read-only #589 shard inspection proves active progress (fresh run/pass events, one SDLC shard completed and the packages stream completed); its outer log is buffered, not evidence of a dead runner. Native qualification for #589 stays held.

Implementation checkboxes are complete; delivery gates are listed separately so review may overlap the retained queued check without claiming a full-suite pass or landing early. No review or merge bypass.

### 2026-10-24 — first boundary review REWORK

Mandatory full review returned one Critical family, BR-1 role-evidence-provenance: both role predicates consult lexical-local evidence and can ignore invalidation through nested captures, pointer aliases or parenthesized writes; the fuzz input can be rebound between the two scanner calls. Full review and ledger are preserved alongside the plan. Fix the entire class using the existing shared summary, with four binding families (container, child, input, result) crossed with direct/parenthesized/address-mediated/captured mutation controls through both checking APIs. No dispute: the counterexamples are consistent with the code's local counts.

Canceled only obsolete queued #590 lease-459019387 with supported cancel; runner session8547 exited1 context-canceled before admission, so it spent no suite slot. Requeue on corrected code after the class fix. No other waiter/process/token changed. PR430 is open and currently reflects pre-fix code: https://github.com/xianxu/ariadne/pull/430 .

#589 suite completed (session31958 exited1) after27m22s: TestVerbContextsReachTrackerReads is own deterministic context.Background misuse, to be fixed with bounded WithoutCancel(caller) contexts; TestQueueRealHelperCompletion/test-lock-exit.test.py is independent. Exact failed-head detached rerun of the whole queue helper test passed4.504s with unchanged Git HEAD/status; evidence retained under common Git sdlc/test-stats/evidence/000589. Filed #595 unclaimed for the fixture race; native IPC qualification still held, and no IPC was allocated/removed.

### 2026-10-24 — BR-1 class fix and verification

Global summary evidence now guards container, child, input and result bindings against writes, address exposure and cross-lexical captures. Additional observed-red controls cover parenthesized/var aliases, indexed byte increments and unknown-call escapes. Children require proven storage uses; input uses must be wholly accounted for by witnessed comparison calls, with actual testing.T diagnostics observational. Multiple witnessed pairs sharing one input stay accepted. No general alias solver, waiver or dependency change. Both Check and CheckSources exercise the mutation table; the counted test also bounds the added binding-use pass.

Full golint package passed1.862s before the extra escape controls; after the shared-use proof, the full race package passed10.218s, including source replays and catalog. Final normal/race verification follows below before a new pushed-head recorded pre-close and fresh boundary review. Ops:0 confirmed #595 duplicates ariadne:5's #592 fix; leave #595 unclaimed and let that owner close it. #589 remains on its separate branch with its context-guard fix pending and native qualification held.

Final verification of the corrected code: go test ./pkg/golint -count=1 passed1.554s; go test -race ./pkg/golint -count=1 passed9.466s; git diff --check is clean. The mutation table has30 cases through each API. Push this head, then start a fresh central pre-close and normal close review; the canceled old suite is not reused.

Corrected pushed code head d01e50714ed892ac456e6a2f58bdb98f9894a4f5 now has a fresh recorded pre-close retained in session70048, lease-4188861440, attempt718961db0f09bad533fe7197f37d40de. Initially position6 for both suite slots; detached checkout /Volumes/TestScratch/local-checks/sdlc-local-3725451576. Preserve this waiter while the mandatory review runs. Raw log is common Git sdlc/test-stats/inflight/718961db0f09bad533fe7197f37d40de/run.log. No passing full-suite proof yet; wait for the result before merge.

### 2026-10-24 — round2 REWORK; third-delta stop

Fresh close review on c5f9d98cf00c351ad254cd4696e391e03053301a returned REWORK: BR-1 remains open for later indexed child overwrite/projected escape and RangeStmt result rebinding. Reviewer independently passed normal/race package checks and reproduced all three supplemental counterexamples through both APIs. The class-fix completion claim is withdrawn; source replays passing does not establish this stronger provenance property.

Stopped before a third mechanism change and wrote a re-plan proposal in the durable plan's Revisions: one exhaustive value-binding write classifier plus a positive closed set of container write/read uses. Patching the demonstrated key alone fails for dynamic keys/aliases; a whole-program alias model is too large. Await hub approval before further tests or code. Canceled only obsolete queued #590 lease-4188861440 through local-check cancel; it had not admitted. No other waiter changed. #589's context fix and refreshed private executable conformance are separately committed/pushed; its native qualification remains held.

## Revisions

### 2026-10-24 — approved smaller scope

Ops:0 delivery c777b3db-4e38-45a8-88c5-111dc2569b14 supersedes the original Spec, Done when and role-inference Plan. Proving value roles statically has grown disproportionately for two real false positives. Ship only the useful diagnostic correction plus sanctioned reasoned per-site annotations; restore the matcher unchanged and remove every automatic role-proof exemption. Historical logs and plan designs remain decision evidence, not the current acceptance contract.

Current Spec: recommend the consumer's own shared strict JSON authority, preserving semantics. A proven standalone value-producing materializer or independent differential fuzz oracle keeps a reasoned per-site existing json-recursive-scanner allow. Pair's two #861 allows stay permanently; no removal follow-up.

Current Done when: (1) original matcher unchanged, no automatic exemptions; (2) diagnostic names accessible consumer authority and reasoned site exception; (3) exact captures require their annotations through both checking APIs, existing true positives unchanged; (4) atlas/lessons document this boundary, pair:3 notified; (5) recorded pre-close and mandatory full review pass before landing.

Implementation: tests observed red against the exemption code: unannotated captures were hidden and annotated captures were rejected as unused allows. Restored the original matcher and removed role inference/provenance files. The pending third-delta classifier proposal is rejected and superseded. Fresh verification and recorded suite follow. No real IPC or other machine operation was performed.

Reduced-scope verification: full golint package passed1.293s, full race package10.099s, all existing true-positive tests/catalog included. Git diff against the base confirms the recursiveJSONScanners implementation is restored unchanged; only the catalog message differs in rules.go. Pair notification receipt ee031e46-8814-4517-a5a0-59c1133fa202 is submitted. Next: fresh pushed-head pre-close and full close review, with earlier role-proof findings disposed by complete mechanism removal.

Reduced-scope code head809d52f360048acdb63188523f50a3e40b3210ff is pushed. Fresh central pre-close retained in session14119, lease-792232057, attempta3d70b64a7af502f08374d438c1d5bf1, detached checkout /Volumes/TestScratch/local-checks/sdlc-local-230647860. Initial position3; keep the waiter. Review may overlap, but landing waits for its pass. PR430 title/body now describe only the final diagnostic/site-allow scope.
