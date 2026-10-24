---
id: 000590
status: working
deps: []
github_issue:
created: 2026-10-24
updated: 2026-10-24
estimate_hours:
card_mirror: '064f21c688ed662bfb605f35899afee7fb1e6605' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-24T04:36:31-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "671f3cea", done: "1474947d"}
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
- [ ] Replay real captures, update guidance, recorded pre-close/close/land, then notify pair:3 to remove its two temporary allows.

## Log

Pair#861 qualification found51 new catalog findings. Three actual duplicate validators will be consolidated; these two are different roles:

1. cmd/internal/strictjson/decode_equivalence_test.go legacyScanValue (line46), used exclusively by legacyRejectDuplicateKeys and FuzzRejectDuplicateKeysEquivalent. It is the independent legacy Token oracle for an optimized byte scanner; replacing it with the implementation under test makes the differential test tautological. Semantics include exact key equality, invalid Unicode replacement and float overflow.
2. examples/mock-to-real/internal/jsonvalue/value.go value(d *json.Decoder, depth int) (any,error) (line28), a generic recursive DOM materializer returning map/array/scalar values with json.Number, exact keys and depth64. Nested module example.com/mock-to-real deliberately has no production Pair dependency (see catalogue/catalogue.go concluding boundary note).

Pair retains the complete observed lint log in its #838 evidence and filed adoption #861. TL explicitly authorized reasoned per-site lint:allow only at these two declarations, naming this upstream issue, pending its correction. These are temporary proven-false-positive annotations, not blanket waivers. Please notify pair:3 on landing.

### 2026-10-24 — claimed, investigated and planned

Claimed and started from fresh main on 000590-json-lint-role-counterexamples, leaving #589's clean pushed branch intact. Its recorded pre-close suite remains retained in session31958, lease-355567331; no restart/cancel. Read current matcher and both real Pair counterexamples at Git HEAD 50a6a163f9551bd4e5a7ebfa5746eba0e7d83eda. Root cause: Token+More+self-recursion+any string map is treated as duplicate-validator authority, regardless of returned values or differential-test purpose. No peer edits, new suppression, implementation or runtime JSON policy changes made. Proposed shared bounded AST role qualification and module-owned diagnostic remediation; sent to hub before code.

### 2026-10-24 — approval

Ops:0 approved the proposed role boundaries (delivery 5cd939f8-dc28-440c-8b78-efec55cf79a4). Plan report receipt b87c22ef-0eb5-4cb5-a657-b92caf348a5e is submitted. #589 retained suite acquired capacity after1053s and is now running in its detached checkout; no restart or cancellation.

### 2026-10-24 — first implementation checkpoint

change-code completed successfully. Captured immutable Pair source files (original allows retained in captures; replay strips only the two relevant annotations). Observed both real captures fail before implementing; both now pass Check and CheckSources. Existing scanner/method/closure fixture expectations remain unchanged and pass. Negative controls retain findings for self-comparison, different input, absent comparison, non-testing fuzz receiver, ordinary caller and escaped helper. Materializer controls caught and fixed overwritten map/slice values and container aliases/address/call escapes; no alternate production dependency or waiver annotation added. A red nested-closure validator control caught a candidate regression; bounded lexical feature propagation now preserves it.

One lexical AST summary visits each node once:8/16 generated helper groups have421/765 AST nodes, with217/433 additional callback-node visits and62/118 graph visits; shared helper paths are visited with memoized membership rather than rescanned per root. Focused TestJSON/TestRules currently pass. Updated remediation names the consumer module's own shared strict decoder; inaccessible internal import advice is gone. Full pkg/race/catalog checks, atlas/lessons, pre-close and boundary review remain. #589's original detached suite continues running and native allocation qualification remains held.

### 2026-10-24 — role controls and docs complete

Full golint package passed1.246s and race8.741s at the first checkpoint. Additional red controls then caught rebound or address-exposed closure self bindings, including parenthesized assignment/address forms; shared summary binding evidence now invalidates those roles and reuses assignedRoot for normalization. Existing package-wide scanners/closures/methods remain checked; no new blanket exemption. Atlas and lessons explain role qualification, conservative limits and consumer-owned remediation. Final full golint race run passed9.822s, including the tracked catalog and all role controls; then push and record pre-close on the final head. #589 suite remains retained separately.
