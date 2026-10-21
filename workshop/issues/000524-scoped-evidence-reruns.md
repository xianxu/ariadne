---
id: 000524
status: working
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-21
estimate_hours: 2.82
collaboration-mode: collaboration
card_mirror: '20febc2959000ea7c3eeed24e9e4034844429ec9' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-21T04:42:56-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
---

# Scoped future evidence and exact-package flake reruns

## Problem

Prerequisite for ariadne#463, approved by hub receipt917bf022-032a-4d84-ac6b-39075bbca2b5. Today's attempt reader treats an unsupported record as a global warning; guardSuitePass then refuses unrelated successes and saved-pass loading marks the whole store incomplete. A future ResumeLink cannot retrofit stale binaries. Existing known-flake reruns also use one name-pattern cross-product across selected packages.

## Spec

Reader-first landing only: make valid unsupported/v3 attempt evidence refuse its own proof as “needs a newer sdlc”, without global read warnings or whole-snapshot incompleteness for unrelated records/checks. Scope using a strictly validated identity envelope; malformed or ambiguous identity remains unknown, never silently ignored. Preserve existing failure guards for affected evidence.

Build exact-package known-flake rerun batches, retaining #466's whole-top-level-test rule and siblings but never running a failed name in another package where it did not fail. Reuse one managed weight-1 lease, existing aggregate targeted budget, ownership/cancellation/process cleanup and attempt linkage. The automatic path uses the shared batch.

No ResumeLink writer or resume activation. Atlas must name the old-reader limitation and the mandatory rollout barrier: after this issue lands, the hub upgrades every slot's actual sdlc; this worker verifies paths/revisions before ariadne#463 phase2 may write links. Never hide links from older readers.

## Done when

- A mixed private store with one valid future/v3 envelope and unrelated ordinary attempts permits unrelated proof; selecting the affected proof refuses “needs a newer sdlc”. Malformed/ambiguous identity never becomes passing evidence.
- Shared store consumers preserve scoped uncertainty without discarding supported records or weakening same-head failure guards. The real pre-change reader control documents why fleet rollout is required.
- A two-package same-name-decoy fixture proves no cross-package over-selection; whole-parent subtest behavior stays intact. Exactly one managed lease and aggregate timeout/cancel behavior are verified.
- No production v3/link writer or resume command is enabled. Atlas records separate landings and fleet binary verification before activation.
- Targeted tests, private executable proof, central recorded suite, close review and normal landing pass; hub receives rollout instructions.

## Estimate

Produced via brain/data/life/42shots/velocity/estimate-logic-v3.1.md against baseline-v3.1.md. Method A only; the source is flagged stale, so provisional. Standard encoding/json/jsonstrict, existing attempt/Git/GH/queue/process seams and stateful fakes cover library needs; no novel stack or new API discovery.

Six smaller modules cover envelope decoding, bounded snapshot/index, authority-consumer integration, exact-package planning, structured PASS collection and flake-step integration. Each uses v2 design0.2 × thorough-spec0.2 =0.04 and impl0.5 × v3.1 scaling0.4 =0.20. The managed-run refactor uses base design0.4 ×0.2 =0.08 and impl0.5 ×0.4 =0.20; private executable conformance is one smaller module at0.04/0.20. Docs use base0.1/0.2 →0.02/0.08; issue/spec0.5/0.1 →0.10/0.04. Two milestone reviews plus final close use base0.1/0.4 →0.02/0.16 each. Familiarity1.0 from recent queue/process/evidence work. Design0.54 ×1.15 buffer + impl2.20 =2.821, rounded2.82h.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec design=0.10 impl=0.04
item: smaller-go-module design=0.04 impl=0.20
item: smaller-go-module design=0.04 impl=0.20
item: smaller-go-module design=0.04 impl=0.20
item: smaller-go-module design=0.04 impl=0.20
item: smaller-go-module design=0.04 impl=0.20
item: smaller-go-module design=0.04 impl=0.20
item: cross-cutting-refactor design=0.08 impl=0.20
item: smaller-go-module design=0.04 impl=0.20
item: atlas-docs design=0.02 impl=0.08
item: milestone-review design=0.02 impl=0.16
item: milestone-review design=0.02 impl=0.16
item: milestone-review design=0.02 impl=0.16
design-buffer: 0.15
total: 2.82
```

## Plan

- [x] M1 — Typed scoped unsupported-evidence reader and all proof-consumer projections, old-reader controls, atlas rollout boundary.
- [ ] M2 — Exact-package managed rerun batch, automatic integration, process/adversarial tests and final delivery preparation.

## Log

### 2026-10-21 — Approved reader-first prerequisite
- 2026-10-21: closed M1 — Scoped reader focused regressions passed 10.218s, race 2.122s, Go lint 0.649s, shell lint 0.313s and diff check; real pre-change control retained; BR-3 README discovery and BR-4 private-clone lifecycle documented; inverse-parent RED/GREEN proves parent check authority.; review verdict: SHIP

Claimed at filing and started independently from fresh main. The approved design is recorded in [the durable plan](../plans/000524-scoped-evidence-reruns-plan.md): one typed scoped read view plus exact-package managed batches. Hub receipt917bf022-032a-4d84-ac6b-39075bbca2b5 owns the split and fleet rollout barrier. Phase1 does not create links or enable resume; #463 and #520 branches are preserved. Package-bound JSON PASS events prevent a same-name or quoted output from satisfying another package's test. Unscopable malformed data remains unknown; valid future identities are localized so unrelated records/checks stay usable.

### 2026-10-21 — Plan-quality round1

PQ-1 (function-test-strategy-contract) asks for named risky functions and adversarial strategies rather than case inventories. The plan now names decoder, authority relation, bounded snapshot, saved/post consumers, batch builder, structured PASS collector, managed batch and production flake entrypoint with one strategy/mechanical guard per function. This is test-contract refinement only; no implementation, mechanism delta or scope change.

Filed as the separately landable phase1 prerequisite from the approved ariadne#463 plan. The original #463 branch holds only committed design, with no implementation. #520 remains held for operator approval on its own branch.

### 2026-10-21 — M1 RED and implementation admission

change-code passed after plan-quality disposed PQ-1 and estimate-quality returned INFO. The derived2.82h is provisional; each component includes its adversarial tests, and review/private-conformance allocations include verification execution and gate overhead, with integration/remediation a stated uncertainty rather than excluded unattended work. M1 TestScopedSavedProofConsumers failed as intended through real saved-proof/post guards: a single valid future record made unrelated repo/head/check proofs incomplete and refused status, while related records lacked a scoped upgrade diagnostic. The existing supported-attempt decoder still rejects opaque v3 data. All fixture state is private; no production evidence was written.

### 2026-10-21 — M1 scoped reader verification

Recovered from branch/issue evidence; previous Couch screen unavailable. The initial saved-proof RED control now passes through load, guard and status publication. All production callers use SuiteEvidenceSnapshot; the old tuple helpers exist only in tests, so production cannot compile against a projection that drops unsupported authority. Strict envelope parsing preserves opaque bytes, scoped indexes avoid repeated sibling scans, flake proof excludes related unknowns, and legacy fallback cannot replace an unsupported suite record. Statistics retain real supported samples with per-record coverage notes; actuals retain supported owned measurements and expose matching-repository unsupported coverage without interpreting it as queue idle.

Focused checks passed in9.812s: envelope perturbations, indexed relation controls, bounded mixed-store reads, counted8/16-record single scans, saved/republish guards, stats, actuals and weekly coverage regressions. Malformed/unreadable evidence still refuses authority. Atlas records the old-reader limitation and fleet upgrade barrier. No v3 writer, resume CLI or production future record was added. The actuals fixture uses a file remote, so its GitHub-shaped future record is deliberately unrelated and produces no matching-repository warning.

### 2026-10-21 — M1 review1 REWORK, class correction verified

BR-1 linked-proof-authority reproduced with a future vm record linked to a supported suite parent: implicit load/reuse and direct publication wrongly passed. The new production-consumer regression failed in0.452s before correction, then passed in0.833s. The reader now resolves supported-parent edges into the subject index once, including partial reads, so explicit/implicit load, reuse, flake projection and publication share the same parent authority. No new persisted state or rollout policy.

BR-2 compatibility-control-evidence is retained as cmd/sdlc/testdata/evidence-pre524-control_test.go.txt plus opt-in sh scripts/verify-pre524-reader.sh. The script checked out pinned pre-change source in a private clone and passed TestPre524FutureRecordPoisonsUnrelatedProof in0.610s: the real old reader emitted one global invalid-record warning, marked the unrelated saved proof not-newest and refused guard/post. Private evidence checkout: /Volumes/TestScratch/tmp/ariadne-slot3/ariadne-pre524-reader.mZhsQY/repo. Equivalent new-reader TestScopedSavedProofConsumers keeps unrelated proof usable and refuses linked uncertainty. Future bytes remain untouched in both controls.

Final focused M1 regressions passed9.710s; Go lint catalog0.644s and shell lint0.308s passed. Earlier focused race run passed1.971s before the parent-index correction; the final focused race run follows. Lessons/plan Revisions record both classes; M1 is not finalized yet.

Final focused race verification of the corrected parent index passed1.942s.

### 2026-10-21 — M1 review2 corrections and durable stop

Review2 disposed BR-1/BR-2 with independent mutation and real-old-source measurements. It returned FIX-THEN-SHIP but did not finalize M1 because BR-3 (README discovery) remained open; BR-4 is Minor (private-clone cleanup lifecycle). Both are corrected in README/atlas/plan Revisions. The invoker owns each printed private compatibility clone, retained through inspection and removed afterwards; no shared-root sweep. Our original evidence clone remains at the path above through inspection.

The BR-1 class sweep added TestScopedLinkedProofUsesParentCheck: a link labelled suite but naming a supported VM parent must not block unrelated suite proof. It reproduced the false refusal in1.099s; the completed index now resolves authority solely through a visible supported parent's check, retaining declared scope only for unresolved parents. All scoped reader/decoder/budget/saved-proof/stats/actual/weekly focused tests passed10.218s, final focused race2.122s, Go lint0.649s and shell lint0.313s. git diff --check passed. No M2 execution work or future writer was started.

NEXT ACTION: run sdlc milestone-close --issue 524 --milestone M1 against this checkpoint, citing BR-3/BR-4 documentation plus the inverse-parent RED/GREEN controls; M1 still needs its final review disposition. Then implement approved M2 exact-package batches, private executable conformance, recorded central suite and delivery. After landing, verify actual fleet binary paths/revisions before #463 activation. #463 and #520 branches remain held unchanged.

No test, suite, merge or review runner remains. The retained session-context pressure checkpoint requires the canonical continuation writer, but none is exposed by this environment. Per AGENTS14 and construct/datatype/continuation.md no-writer rule, this issue Log/plan is committed and pushed, no continuation is hand-written, and work stops for Pair park/continue recovery. Previous-screen evidence is unavailable; branch/issue/gate records are the durable recovery evidence.


### 2026-10-21 — Recovery, M1 SHIP and M2 RED/GREEN

Recovered with sdlc state on the clean owned issue branch. Final M1 round3
returned SHIP, disposed BR-1/BR-2/BR-3/BR-4, and committed/pushed its boundary.
The reviewer independently ran scoped consumers, broader proof/statistics/
actuals/weekly cases, race controls, real old-source compatibility, and parent
index mutation; it reported no new findings.

M2 TestKnownFlakeExactPackageGate failed against the old runner in2.807s:
both unrelated same-name decoy sentinels were created. The exact-package
batch now preserves one managed queue/slot/lease/scratch/deadline while joining
each group before the next. Pure per-package selection and structured PASS
functions were delegated as a bounded AGENTS3 task; main retained the shared
runner and gate integration. Pure GREEN0.060s. The new argv-as-data regression
also failed against unquoted announcements (0.513s), then passed with quoted
argv. Existing raw-PASS fixture scripts were updated to explicit structured
package/test facts, keeping original failure-attribution assertions intact.

Focused normal checks passed20.911s (known-flake gates, real package decoys,
subtest/sibling rules, managed batch and descendant cleanup); pure flakes
checks passed0.063s. No full suite or M2 close is claimed yet. Next: private
executable conformance, broader focused/race + source checks, durable checkpoint
and recorded central suite, M2 review and final issue close. No v3/link writer
or resume activation was added. #463/#520 branches remain untouched.


### 2026-10-21 — M2 executable proof and pre-suite checkpoint

Private production-main conformance passed9.452s using private real Git,
tracker refs/cutover, queue, helper, scratch, lock and fake GitHub. It proves
unrelated lint republish with one future smoke record; smoke upgrade refusal
with no status; automatic two-package reruns with no same-name decoys; one
linked child plus original/logs; unchanged future bytes and expected statuses.
Fixture-only setup corrections supplied upstream, exact cutover root and card
Problem section. All fixture state is automatically removed.

The shared runner's existing environment refusal test caught command headers
in the raw log (RED0.633s). Announcements now go only to terminal diagnostics,
quoted; the strict raw-output grammar is unchanged. Broader normal checks
passed44.601s, including environment repair, runner cleanup and uncertainty
retention, saved reuse/republish, known flake/affected, queue signals, managed
batches and descendants. Focused race passed21.489s; pure race1.152s. Go lint
catalog0.699s, shell lint0.325s and both diff checks passed. The old unused
text-PASS helper is removed; all new execution proof is package-bound JSON.

PR #371 is open. This checkpoint precedes the required central recorded suite
on the pushed head; suite and M2/final close are still pending. No activation,
production future record or installed fleet binary change is authorized here.


### 2026-10-21 — Recorded suite queued; review overlap

The central recorded suite is live on pushed code head
21c1c5f3d1bae7d66adeccd4ec8ff160f604b992 (run lease-1500034090), queued
second behind two Pair holders and an earlier Ariadne suite. Server
local-gates-guard and its verdict passed on that head. Under tl-worker's
explicit review-overlap rule, M2 review may proceed with this pending record;
the suite checklist stays pending and no final delivery/acceptance is claimed.
A recorded pass on the final code still precedes whole-issue close and landing;
a code correction will requeue it. This docs-only checkpoint preserves the
queued run and its FIFO position; no restart is attempted.
