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
- ARCH-CONSTRAINTS: batch path once per converged split, bounded by existing train car cap (default 8). Reuse parsed failure facts per exact SHA; one current registry read per assessment, at most two exact endpoint graphs and one path diff; no graph per test and no new full suite. Existing log limit and graph/check timeouts/cancellation apply; confirm a bounded context for graph preparation. Count 8/16 failure leaves with constant registry/graph/diff calls and linear leaf decisions. The existing O(log C) split-suite bound remains; this proposal schedules zero automatic retry suites.
- ARCH-SECURE: verify run SHA/hash and typed parser facts; untrusted/truncated logs and unknown graph/configuration cannot waive gates. Preserve exact parent/leaf semantics: a registered /30 does not explain a parent-only failure, and sibling failures remain unknown. No credentials or remote writes are introduced by assessment.
- ARCH-ORDER: only the reducer can emit a suite Hold after exact-pair assessment; test external registration at controlled barriers, cancellation, stale events, main movement and mixed failures. A permanent cross-system lease is unsuitable because tracker registrations and GitHub merges are independently writable; authority is refreshed at the decision, with final passing gates retained.
- ARCH-FUNERAL: evidence/graph maps die with the conductor; detached worktrees clean up on success/error/cancel; existing suite attempts/logs retain their existing pruning. No new durable record family, queue, daemon or automatic retrain loop.
- ARCH-EVOLVE: new phase/event/effect is in-memory. Reuse existing terminal outcome kinds and reason strings; do not rewrite old suite records or upgrade earlier failed statuses to success. No persisted schema/version change or local-gates configuration edit.

## Implementation steps after approval — one review boundary

### 1. Reproduce and specify the decision

Files: cmd/sdlc/internal/landingline/attribution_test.go, trainstate_test.go, cmd/sdlc/landingtrain_test.go or landingtrain_attribution_test.go.

- [ ] Add pure invariants: no SuiteDone convergence can emit Hold without SplitAssessed; registered/untouched-only failures cannot hold; a failed hi never becomes Results[hi]=true; lo alone can land; unknown evidence stops attribution.
- [ ] Add a production two-car regression shaped like train234: full train fails broadcast.TestHubRandomInterleavings, an independent tracker writer registers its exact parent/leaf after fail→pass evidence, the lower prefix passes, and the old implementation wrongly Holds the scribe-only second car. Assert named unattributed outcome, no culprit and no failed-prefix merge. Cover registration already present as well as arriving during split.
- [ ] Add independent-package and positive controls: touched unregistered failure still holds, changed imported dependency stays relevant, embed/testdata changes are not classified untouched, mixed failures name only the supported culprit.
- [ ] Add parser/authority controls: registered leaf versus coarsened parent, sibling unknown failure, closed/removed registration, policy off, persistent failure, unreadable log/hash mismatch/registry/graph, no true green baseline for lo=0.
- [ ] Run new named regressions RED before production changes; preserve current culprit split test as a compatibility control. Commands: go test ./cmd/sdlc -run '^(TestLandingTrainSplitRefreshesRegisteredFlake|TestLandingTrainSplitDoesNotBlameUntouchedPackage|TestLandingTrainSplitHoldsTheCulpritAndLandsThePrefix|TestTrainSuiteOutcome)$' -count=1; go test ./cmd/sdlc/internal/landingline -run '^(TestSplitAttribution|TestTrainSplit)' -count=1. New tests use those exact prefixes; expect the two new production regressions to expose a wrong Hold before the fix.

### 2. Add the pure attribution boundary

Files: cmd/sdlc/internal/landingline/attribution.go, trainstate.go, trainstate_test.go.

- [ ] Define typed failure facts and attributable/unattributed/unverified verdicts, with per-failure reasons. Pure classification must use coverage facts produced by existing flakes helpers rather than duplicate their matching grammar.
- [ ] When Bisect converges, enter attribution phase and request AssessSplit{lo,hi}. Store/validate the exact expected pair; keep existing Results immutable until an actual new suite result.
- [ ] On a matching attributable SplitAssessed, call the existing split/hold path. Otherwise Refuse unresolved cars with the unattributed/unverified recovery reason, retain only verified lo for final landing checks, or close with no landing for lo=0. Do not emit Release for this new unresolved case.
- [ ] Extend illegal/duplicate/stale event, random-sequence and membership tests. The existing conflict-hold path and genuine culprit remainder-release behavior remain covered.

### 3. Supply fresh exact evidence from the conductor

Files: cmd/sdlc/landingtrain_attribution.go, landingtrain.go, landingtrain_attribution_test.go.

- [ ] Retain local-run records/facts keyed by prefix and exact SHA, including baseline evidence needed for prefix0. Clear invalid evidence on code reintegration.
- [ ] Execute AssessSplit through a bounded cancellable shell: verify saved run/log identity, parse exact facts, derive endpoint impact using existing graph/config seams, then read the current registry with a new scope and classify once.
- [ ] Deliver a SplitAssessed event for that exact pair. Keep existing knownFlakeStep/Classify/Revalidate/Unexplained passing authority intact; no synthetic success posting or saved-attempt rewriting.
- [ ] Print “unattributed: registered flake …” with issue reference, or “unattributed: untouched package …” with package/prefix scope. Unverified errors name the missing evidence. Print exact merge recovery for affected cars through existing result reporting.
- [ ] Run deterministic production regressions GREEN; assert stateful GitHub fake never merges failed hi and all cars receive an outcome. Demonstrate one terminal outcome per unresolved car rather than a retrain loop.
- [ ] Verify at 8/16 leaves that registry, graph and diff calls stay constant per assessment and decisions scale linearly. Logs/graphs never escape their bounded lifetime.

### 4. Documentation and verification

Files: atlas/workflow/sdlc-binary.md, issue Log and this plan.

- [ ] Document attribution phase, AssessSplit/SplitAssessed names (existing atlas drift test), diagnostics, explicit retry, and unchanged passing-evidence contract. Update existing train sketch; no new atlas file needed.
- [ ] Run go test ./cmd/sdlc/internal/landingline ./cmd/sdlc/internal/flakes ./cmd/sdlc/internal/affected -count=1, and targeted root-package train/split/flake/landing checks with exact names. Add semantic regressions rather than an inapplicable static lint rule; run pkg/golint.
- [ ] Build own committed binary and present the isolated train234 production-path proof to the hub for its live-gate decision; do not touch :0 without authorization.

The binary owns the remaining lifecycle: after approval run change-code's full-flow plan-quality gate, then derive actual estimate only when requested; implement, push, recorded suite, close's mandatory fresh review, and merge. Keep continuation/restart outside any owned queued/running check. If close is reviewed while a suite remains queued, avoid a redundant merge-triggered check by waiting for the original result before resuming landing if that occurs; cancel only the redundant owned run by exact ID after proving equivalence.

## Planning evidence

Read-only inspection and a bounded independent read agree on the boolean-only attribution gap and scope-cache freshness requirement. Existing pure suites pass: landingline 7.021s, flakes 0.140s, affected 0.396s. No new failing regression or implementation exists yet. The original live incident is recorded in the issue Spec, not claimed as a locally reproduced test.

## Approval required

TL to approve the explicit no-auto-retrain policy for unattributed cases and the dependency-aware untouched proof. This new landing-gate state/refusal is outside self-approved quick fixes. No code or gate configuration changes before approval.

## Approval recorded — 2026-10-21

TL approved the review point and the complete plan via submitted Couch 538eeecd-6365-4b44-8729-b5a2d048e632: final non-blaming unattributed refusal with exact retry, verified-prefix-only landing, fresh dependency-aware assessment phase, and train234's registered broadcast/scribe-only regression. The Approval required section above is the proposal's historical request; this record admits the full-flow change-code gate, not gate bypasses.
