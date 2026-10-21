---
collaboration-mode: discussion
---

# Landing split attribution implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy). The investigation is bounded and delegated read-only; execute the integrated state/effect change in the main session. This atomic work has one sdlc close review boundary, with plain checkboxes.

**Goal:** A split never holds a car solely for a currently registered flake or a failure in a package its exact prefix delta cannot affect.

**Architecture:** Insert an explicit attribution effect between bisect convergence and Hold. The IO shell retains exact prefix run evidence and supplies fresh registry plus dependency-aware impact facts; the pure core decides whether to hold, report an unattributed failure, or stop on unknown evidence. Only existing passing evidence authorizes landing.

**Tech Stack:** Go; landingline reducer/effects; current flake parser/registry and affected graph; temporary real Git repositories and existing stateful train GitHub fake.

## Core concepts

| Pure entity | Lives in | Status |
|---|---|---|
| Split attribution facts/verdict | cmd/sdlc/internal/landingline/attribution.go | new |
| Train attribution phase/event/effect | cmd/sdlc/internal/landingline/trainstate.go | modified |

An assessment belongs to one train and one adjacent lo/hi prefix pair (hi=lo+1); hi identifies the suspect car. Per-failure facts carry the exact named leaf, current registration coverage, and whether candidate impact is established, excluded or unknown. Verdict is attributable, unattributed, or unverified, with diagnostic detail. A registered/untouched failure alone cannot produce attributable. A mixture may hold only on a specific unregistered failure the candidate can affect; other failures stay reported. Missing facts never become innocence or a pass. Pure tests need no IO mocks.

| Integration point | Lives in | Status | Wraps |
|---|---|---|---|
| Exact endpoint package impact (INTEGRATION) | cmd/sdlc/landingattribution_impact.go | new | Git diff, detached endpoint worktrees, graph preparation/listing; existing pure affected ownership/selection |
| Prefix evidence retention | cmd/sdlc/landingtrain.go | modified | Exact prefix SHA and localRunRecord; diagnostic baseline record when lo=0 |
| Attribution assessment shell | cmd/sdlc/landingtrain_attribution.go | new | Verified saved log, fresh scoped registry, Git delta and exact-head package graphs |
| Result delivery | cmd/sdlc/landingtrain.go | modified | Existing Hold/Refuse outcomes and recovery messages |

Existing flake coverage/passing authority (internal/flakes and localcheck_flakes.go) and affected graph/selection seams remain unchanged and are reused. Bisect still finds bounds; only the reducer's authority to emit Hold changes. The shell verifies that retained run SHA/log hash belong to hi; log failure leaves come from parseSuiteLogFacts, never guessed from human prose. The evidence map lives only for one conductor and is discarded at termination/reintegration; raw attempts keep existing store retention. The registry uses a new withIssueRecordsScope per attribution decision, not the long-running command cache. Read it as late as practical before returning the verdict; the result is a point-in-time observation, not a lock on remote issue edits.

## Proposed policy and alternatives

Recommended: validate attribution before Hold; if no valid culprit remains, land only the already verified lo prefix and stop unresolved cars with existing OutcomeRefused plus a non-blaming “unattributed: registered flake … / untouched package …” reason and exact retry guidance. If lo=0, nothing lands. Culprit remains unset for this path.

This deliberately uses a final refusal, not Hold: the message does not accuse a car. It also avoids Release, which Line.Wait clears automatically before retrying and could launch repeated trains against a persistent unrelated failure. The operator/worker can retry explicitly after investigating or obtaining valid same-head passing evidence. Ordinary genuine-culprit Release behavior for cars behind a held culprit remains unchanged.

Rejected: silently skipping Hold in the IO handler after split already removed a car; reducer membership would be wrong. Rejected: treating registration or unchanged sources as SuitePass; that would authorize an unverified prefix and bypass local-check evidence. Optional future optimization: re-adjudicate an old failure through a recorded, bounded targeted rerun and resume a larger prefix. That is not needed for the no-false-blame contract and is not added here; current knownFlakeStep still handles reruns during ordinary suite execution.

**TL review point:** approve stopping unresolved cars with an unattributed refusal rather than immediately retraining them. This bounds retry work without waiving any gate.

## State and actions

Actions: continue verified-prefix landing; inspect named evidence; obtain valid same-head verification using existing local-check/flake policy; rerun the printed sdlc merge --branch BRANCH --yes command after resolution; cancel the owning check by run ID if needed.

| Internal state / failure mode | Visible state | Recovery action | Possible loss |
|---|---|---|---|
| Bisect has adjacent lo pass / hi fail | assessing attribution | wait for bounded assessment | none |
| Exact unregistered failure can be affected by suspect delta | held with test/package evidence | fix/re-close/push and retry merge | suspect does not land; existing split behavior |
| Every failure is currently registered or provably untouched | unattributed; unresolved cars not landed, no culprit | inspect evidence, obtain valid verification, retry printed merge | they wait for explicit retry; verified lo can still land |
| Registered failure still fails under current rule, or rule is off | unattributed registration, passing gate remains blocked | investigate/fix, then retry | no failed head lands |
| Mixed registered and real touched unregistered failure | held only for the real touched failure | fix/re-close/push and retry | no known flake alone is blamed |
| Log/hash/registry/graph unavailable; unnamed/coarsened evidence cannot establish attribution | attribution unverified; unresolved cars stopped | repair named evidence and retry | no suspect falsely held; no unverified prefix lands |
| Cancellation/process error while assessing | existing abort cleanup, no new Hold | inspect retained train/PR/check evidence before retry | no new remote merge request sent |
| Main moves after assessment | existing reintegration and gate path | wait for required new verification | stale prefix evidence is invalidated |

Use a new ephemeral attribution phase, AssessSplit effect, and SplitAssessed event naming exact lo/hi. Reject mismatched, duplicate, late and out-of-phase events without mutation. A terminal train never resumes. Existing Aborted/Refuse/ClosePR machinery owns failure cleanup. No new persisted outcome enum or ticket schema is needed.

## Exact package comparison

Use actual adjacent prefix SHAs: lo is heads[lo-1], or the retained main baseline when lo=0; hi is heads[hi-1]. Do not compare the suspect branch to today's main or the train's first failing head indiscriminately.

Obtain changed paths and package ownership/dependencies at both endpoints; union impact so removal/move/import changes cannot disappear from the proof. Reuse affected.Graph.Owners/Dependents and affected.Plan for test imports, testdata and embedded inputs. Include configured full-selection/build/module/generator inputs conservatively. Only a fully known selection excluding a failing package proves untouched. Unknown/unclassified/global selection never proves innocence; it yields unverified attribution unless another exact unregistered failure independently establishes the suspect connection. Missing/mismatched module/package/graph likewise stops attribution.

For lo=0 retain the actual baseline record/facts: c.baseline=false currently means “not every failing train test failed on main”, not necessarily a globally passing baseline. Do not upgrade that bool to a passing-prefix proof; if suitable baseline evidence is missing, report unverified rather than holding the first car from an assumption.

## Architectural and operating envelope

- ARCH-DRY: one assessment owner before every suite-culprit Hold, including single-car and first-car cases; use existing flake coverage and affected ownership. Conflict/instance-gate holds are a separate cause and remain unchanged.
- ARCH-PURE: reducer/verdict tests are pure; IO shell verifies logs, registry and exact-head graph facts.
- ARCH-PURPOSE: pre/during-split registration and untouched packages are defended without passing a failed prefix. All suite-hold consumers use the same boundary.
- ARCH-MOCK: real temporary Git/tracker repositories plus landingFakeGH model train execution and remote registry updates. Barriers inject registration between failed full suite and passing split result. No live Pair/Hubs or production tracker mutations in the regression.
- ARCH-CONSTRAINTS: batch path once per converged split, bounded by existing train car cap (default 8). Reuse parsed failure facts per exact SHA; one current registry read per assessment, at most two exact endpoint graphs and one path diff; no graph per test and no new full suite. Attribution owns one context.WithTimeout(c.ctx, splitAttributionTimeout), initially 2 minutes, covering log reads/verification, graph preparation, both go list calls, Git diff and registry IO. This is an explicit diagnostic-budget choice: no new test suite runs, only two endpoint graph probes, and failure to finish safely stops without blame. It is not a measured universal graph guarantee; slow/cold/generated trees may return unverified. The outer deadline always dominates each preparation's own timeout. Cleanup has a separate at-most-5-second context (existing bounded cleanup seam), preserving the original caller's cancellation semantics; no child remains running at return. Re-measure endpoint graph preparation/list latency in the isolated production proof before widening this budget; keep the enforced initial bound until a scope revision is approved. Count 8/16 failure leaves with constant registry/graph/diff calls and linear leaf decisions. The existing O(log C) split-suite bound remains; this proposal schedules zero automatic retry suites.
- ARCH-SECURE: verify run SHA/hash and typed parser facts; untrusted/truncated logs and unknown graph/configuration cannot waive gates. Preserve exact parent/leaf semantics: a registered /30 does not explain a parent-only failure, and sibling failures remain unknown. No credentials or remote writes are introduced by assessment.
- ARCH-ORDER: only the reducer can emit a suite Hold after exact-pair assessment; test external registration at controlled barriers, cancellation, stale events, main movement and mixed failures. A permanent cross-system lease is unsuitable because tracker registrations and GitHub merges are independently writable; authority is refreshed at the decision, with final passing gates retained.
- ARCH-FUNERAL: evidence/graph maps die with the conductor; detached worktrees clean up on success/error/cancel; existing suite attempts/logs retain their existing pruning. No new durable record family, queue, daemon or automatic retrain loop.
- ARCH-EVOLVE: new phase/event/effect is in-memory. Reuse existing terminal outcome kinds and reason strings; do not rewrite old suite records or upgrade earlier failed statuses to success. No persisted schema/version change or local-gates configuration edit.

## Implementation tasks — one review boundary

### 1. Regression and function-level test contracts

Files: cmd/sdlc/internal/landingline/attribution_test.go, trainstate_test.go, cmd/sdlc/landingtrain_attribution_test.go and existing landingtrain_test.go fixture.

| Risky function | Adversarial input and mechanical guard |
|---|---|
| ClassifySplit (new pure classifier) | Generate mixed registration/impact/unknown fact sets; property-check that registered/untouched facts alone never authorize Hold or SuitePass and that reasons identify supporting evidence. |
| Transition | Extend the existing random event-sequence model; assert exact-pair/phase authority, no Hold before assessment, immutable failed-prefix results, terminal idempotence/refusal, and final membership bounded by verified passing evidence. |
| splitImpact (new INTEGRATION package-impact proof helper) | Graph/path mutation properties exercise ownership, transitive/test dependency and resource relationships; removing graph/config/path knowledge may never create an untouched proof. Real Git/Go fixtures exercise the IO shell; existing pure affected graph functions own ownership and dependency selection. |
| readSplitEvidence (new IO verifier) | Corrupted/unavailable/mismatched record/log inputs in temporary stores must return unverified; no partial parse or guessed leaf creates authority. |
| conductor.assessSplit (new bounded IO effect) | Independent tracker writes at explicit train barriers and controllably stalled Git/graph/registry seams prove fresh decision-scoped authority, deadline expiry, cancellation and bounded cleanup; count graph/diff/registry calls at 8/16 leaves. |

- [x] Add TestLandingTrainSplitRefreshesRegisteredFlake replaying train234 through the real train fixture: failed broadcast.TestHubRandomInterleavings, external issue-backed registration after same-head evidence, passing lower prefix, scribe-only suspect. Assert the current false Hold, then the corrected unattributed outcome, exact retry and absence of any failed-prefix merge.
- [x] Add TestLandingTrainSplitDoesNotBlameUntouchedPackage through the same stateful fake and real Git/tracker seam, retaining the ordinary culprit-hold fixture as the positive compatibility control. Pure strategies above defend the class rather than duplicating a prose inventory.
- [x] Run the new production regressions RED before implementation: go test ./cmd/sdlc -run '^(TestLandingTrainSplitRefreshesRegisteredFlake|TestLandingTrainSplitDoesNotBlameUntouchedPackage|TestLandingTrainSplitHoldsTheCulpritAndLandsThePrefix|TestTrainSuiteOutcome)$' -count=1. Pure verification: go test ./cmd/sdlc/internal/landingline -run '^(TestSplitAttribution|TestTrainSplit)' -count=1.

### 2. Pure attribution authority

Files: cmd/sdlc/internal/landingline/attribution.go, trainstate.go, trainstate_test.go.

- [x] Implement ClassifySplit over typed current-registration and impact facts, returning attributable/unattributed/unverified plus reason; existing flakes helpers own matching grammar.
- [x] Extend Transition with attribution phase, AssessSplit{lo,hi} and SplitAssessed for that exact pair. Only attributable invokes existing split/Hold; other verdicts Refuse unresolved cars and retain verified lo, or close without landing when lo=0. No Results false-to-true rewrite and no automatic Release on this new path.
- [x] Verify the named pure strategies and preserve existing conflict-hold and genuine-culprit remainder-release behavior.

### 3. Exact evidence and bounded IO

Files: cmd/sdlc/landingtrain_attribution.go, landingtrain.go, landingtrain_attribution_test.go.

- [x] Retain per-prefix SHA/run evidence and actual baseline evidence, invalidating it on code reintegration. Implement readSplitEvidence and splitImpact using existing parser, flake and affected owners.
- [x] Implement conductor.assessSplit as the sole owner of the 2-minute assessment context. Pass it through every descendant Git, graph-prepare, go-list and registry operation; each expiry yields unverified, never Hold. Cleanup uses the separate bounded 5-second lifetime; a stalled-dependency test must observe child exit and no remote merge before return.
- [x] Execute AssessSplit and feed its exact-pair result into the reducer. Report named registered issue/untouched package/unverified evidence and print the affected car's exact sdlc merge --branch BRANCH --yes recovery. Keep knownFlakeStep, Classify, Revalidate and Unexplained passing authority unchanged; no saved-attempt rewriting or synthetic success status.
- [x] Run the production regressions GREEN and the named IO strategy; assert one final unattributed outcome per unresolved car, no automatic retrain, and constant per-assessment IO calls with linear leaf processing.

### 4. Documentation and verification

Files: atlas/workflow/sdlc-binary.md, issue Log and this plan.

- [x] Map the attribution phase/event/effect, diagnostics, exact retry and unchanged passing-evidence contract into the existing train sketch, satisfying its event/effect drift test.
- [x] Run go test ./cmd/sdlc/internal/landingline ./cmd/sdlc/internal/flakes ./cmd/sdlc/internal/affected -count=1, targeted root-package train/split/flake/landing checks and pkg/golint. No static lint rule can infer the semantic attribution invariant; the production/model regressions enforce it.
- [x] Build the committed own binary and present the isolated train234 production-path proof plus attribution-IO timing to the hub for its live-gate decision. Leave :0 untouched without authorization.

The binary owns the remaining lifecycle: after approval run change-code's full-flow plan-quality gate, then derive actual estimate only when requested; implement, push, recorded suite, close's mandatory fresh review, and merge. Keep continuation/restart outside any owned queued/running check. If close is reviewed while a suite remains queued, avoid a redundant merge-triggered check by waiting for the original result before resuming landing if that occurs; cancel only the redundant owned run by exact ID after proving equivalence.

## Planning evidence

Read-only inspection and a bounded independent read agree on the boolean-only attribution gap and scope-cache freshness requirement. Existing pure suites pass: landingline 7.021s, flakes 0.140s, affected 0.396s. No new failing regression or implementation exists yet. The original live incident is recorded in the issue Spec, not claimed as a locally reproduced test.

## Approval required

TL to approve the explicit no-auto-retrain policy for unattributed cases and the dependency-aware untouched proof. This new landing-gate state/refusal is outside self-approved quick fixes. No code or gate configuration changes before approval.

## Approval recorded — 2026-10-21

TL approved the review point and the complete plan via submitted Couch 538eeecd-6365-4b44-8729-b5a2d048e632: final non-blaming unattributed refusal with exact retry, verified-prefix-only landing, fresh dependency-aware assessment phase, and train234's registered broadcast/scribe-only regression. The Approval required section above is the proposal's historical request; this record admits the full-flow change-code gate, not gate bypasses.

## Revisions

2026-10-21: Addressed PQ-1 by replacing the prose test-case inventory/procedural detail with named-function adversarial strategies, preserving the two motivating production regressions and exact verification commands. Addressed PQ-2 with a concrete 2-minute outer assessment context owned by conductor.assessSplit, propagated through all IO and graph preparation/listing, a 5-second cleanup bound, explicit unverified expiry and a controllably stalled dependency test. The bound is an initial diagnostic-budget choice, not an assertion that arbitrary generators finish within it; no approved attribution/landing policy changed.

2026-10-21 implementation refinement: real production testing additionally exposed stale smaller heads after code-main reintegration. Existing PhaseTesting/Reintegrated and Refuse machinery now stops a failed multi-car reintegration before bisection because no valid current-base smaller prefix exists; lower heads/evidence are invalidated. No new state/outcome type or recomposition mechanism. Production RED11.430s proved the unsafe old-prefix landing, GREEN5.34s proves no merge/hold/release, and docs-only control9.35s stays valid. Production motivating regressions were validated GREEN and by controlled restoration of the old adjacency-only decision (RED14.208s); the first behavioral reducer RED preceded implementation. The binary-owned suite/close/merge gates remain outstanding, as does the hub live-gate response; the checked implementation checklist does not claim those gates passed.

2026-10-21 BR-1: Corrected splitImpact classification and the integration concepts inventory to match its Git/worktree/preparation/Go-list IO. The implementation already keeps those effects outside ClassifySplit/Transition; this is a documentation correction with no authority or behavior change.
