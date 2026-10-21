---
collaboration-mode: discussion
---

# Reopened car publish obligation implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy). Execute this single atomic change in the current session; sdlc close owns its one fresh-context review. Steps use checkbox syntax.

**Goal:** A carried close remains a publish obligation after external reopen; its waiting car becomes not-ready with a named reason on the next pass.

**Architecture:** Separate status-independent binding ownership from publish eligibility. Reuse the existing repository/evidence/current-token ownership rule in one shared pre-publish guard; leave completion/settlement/archive status selection intact. No new persisted state or branch-name inference.

**Tech Stack:** Go, current tracker records scope, real temporary Git repositories and bare tracker remotes; existing local reviewer and GitHub fakes.

## Core concepts

| Pure entity | Lives in | Status |
|---|---|---|
| Publish eligibility of a carried completion | cmd/sdlc/trackercompletion.go | new |

The existing Completion binding in cmd/sdlc/internal/issue/handoff.go remains unchanged. The new pure decision receives issue ID and current status after ownership has been established. codecomplete and done remain eligible for their existing paths; any other current status returns an issue/status/re-close refusal. It creates no completion, and it cannot turn an unrelated card into an obligation. Colocated table tests cover every current model status; the vocabulary is construct/vocabulary/issue.cue, unchanged.

| Integration point | Lives in | Status | Wraps |
|---|---|---|---|
| Status-independent carried-completion predicate | cmd/sdlc/trackercompletion.go | modified | Existing evidenceInHead, closeTokenCarried and trackerEnv Git boundary |
| Shared pre-publish eligibility guard | cmd/sdlc/trackercompletion.go | new | Fresh scoped tracker records and carried-completion predicate |
| Ordinary publish | cmd/sdlc/publishgate.go | modified | Guard before owned publish anchors |
| Immutable landing publish | cmd/sdlc/landinggate.go | modified | Same guard with pinned PR head/base before selection |

Each binding belongs to one issue and names one repository/current close generation; one branch may carry many bindings. The predicate retains the existing missing-object handling, base exclusion and current-token fallback after rebase. Existing ownedCompletions delegates to that predicate after its current status filter; its callers' behavior stays unchanged. The publish guard checks ineligible statuses against that same ownership relation before any best-anchor ranking. No second ownership model, status mirror, watcher or durable car field is introduced.

## Scope and alternatives

Recommended: shared publish-only guard plus extracted ownership predicate. This directly enforces the reported invariant across waiting-car and ordinary publication while preserving post-landing behavior.

Rejected: moving a refusal directly into ownedCompletions, because settlement and archive call it and must keep ignoring reopened work. Rejected: refusing every empty selection or using branch names/detail diffs, because genuine issue-free branches, multiple issues and independently published details are legitimate.

The reported case preserves a valid current binding. Missing/replaced/stripped bindings and the existing difference between ordinary refuseUnownedCompletions and immutable landing are not redefined by this fix; the old token must never claim a newer binding. Preserve and run existing supersession controls. If regressions expose a need for historical ownership beyond the current binding, stop and bring that scope change to the hub rather than adding a second ownership mechanism.

## Actions, states and ordering

Actions: wait for readiness refresh; re-close and push the corrected current branch; repair unreadable authority and retry; inspect an unconfirmed outcome.

| Internal state / failure mode | Visible state | Recovery action | Possible loss |
|---|---|---|---|
| Car carries current close; card is codecomplete; other gates pass | ready | wait for landing | none added |
| Same binding remains carried; external actor changes status to working, blocked, open, wontfix or punt | not-ready: issue and status named | re-close and push current work | no branch work or queue state deleted |
| Corrected current close is pushed and carried | ready if existing gates pass | wait for landing | older close stays superseded |
| Another carried issue is eligible but one is reopened | not-ready: reopened issue named | re-close and push that work | no review obligation silently dropped |
| No bound close is carried; unrelated working cards exist | existing issue-free behavior | wait under existing gates | none added |
| Fresh tracker/Git authority cannot be read, is malformed, or cancellation arrives | not-ready with existing read error | repair authority and retry | no success inferred from uncertainty |
| Already landed or unrelated binding is only on base/not carried | not owned by this publish | existing behavior | none added |

A readiness pass already owns one coherent tracker generation; external edits during it appear on the next pass. This fix adds no background work. Reopen → next pass → refusal → re-close/push → next pass → ready is the deterministic event sequence under test. The guard is read-only; it cannot atomically prevent a remote edit after the last observation, and existing post-merge generation CAS remains the final write protection. A new cross-system lease is unnecessary for this per-pass correctness bug and would require changing tracker writers and the GitHub merge protocol.

## Operating envelope and principles

- ARCH-DRY: extract and reuse the exact current ownership predicate for guard and ownedCompletions; reuse current per-pass records scope.
- ARCH-PURE: issue/status eligibility is pure; existing tracker/Git shell supplies ownership facts. No new external command seam.
- ARCH-PURPOSE: both publication entry points and every carried issue are checked; issue-free branches remain supported.
- ARCH-MOCK: regression uses real portable temporary Git/tracker repositories, independent tracker writer and existing landingFakeGH/stubJudge; no real hub or production remote is mutated. Current Git semantics are already exercised by real Git.
- ARCH-CONSTRAINTS: readiness is a once-per-minute batch scan, not a per-byte path. For N cards, one additional in-memory pass and at most existing bounded ancestry/token reads for K candidate ineligible bindings; O(N + K) ownership invocations with the same history-dependent Git work as current selection. One tracker generation/fetch per pass remains. Use real-Git trace counts at 8 and 16 unrelated/candidate records to reject superlinear ownership probes; no extra remote refresh per issue. Missing/cancelled authority refuses, with existing command bounds.
- ARCH-SECURE: only typed tracker records and configured repository identity grant ownership; branch labels/detail text never do. Malformed relevant authority and Git errors retain fail-closed handling. No credentials or real user state touched by tests.
- ARCH-ORDER: current binding and pinned head/base determine ownership before status; all carried status checks precede covering-anchor ranking. Reclose supersedes rather than revives old authority.
- ARCH-FUNERAL: creates no durable runtime records, logs or processes. Predicate results die with each readiness pass; test repositories use t.TempDir.
- ARCH-EVOLVE: no persisted format changes, status additions or binding rewrites. Unknown-field preservation remains owned by existing card APIs.

## Implementation steps — one close boundary

### 1. Capture the production regression first

Files: cmd/sdlc/landingcar_records_test.go, cmd/sdlc/trackercompletion_test.go (or a focused trackerpublish_eligibility_test.go).

- [ ] Build from closeReady and the existing real close plus productionCarGates fixture. Assert ready before reopening.
- [ ] Use an independent clone/tracker writer to change the card to working while preserving its completion; call carReadiness again with the same outer command context. Assert not-ready, issue ID and working in Reason, and no code/card mutations by the gate. The old implementation must fail this assertion as recorded in the issue's Problem.
- [ ] Re-close corrected work through the real close path, push, and assert readiness recovers on the current generation. Also cover a reopen after evidence was rebased and after details were independently published.
- [ ] Add ordinary runPublishGate coverage for the same carried reopened binding, plus a two-issue case where a valid newer anchor cannot hide it.
- [ ] Add controls: genuinely issue-free branch; unrelated working binding absent from head; old binding already on base; current eligible/done archive retry; foreign repository; no new generation claimed by an old token; settlement ignores reopened work.
- [ ] Run the new named tests alone and capture the expected behavioral RED. Do not fix fixtures and call that the bug reproduction.

### 2. Apply the shared guard

Files: cmd/sdlc/trackercompletion.go, cmd/sdlc/publishgate.go, cmd/sdlc/landinggate.go.

- [ ] Extract the body of ownedCompletions' repository/evidence/base/current-token matching without semantic changes. Its conceptual signature is completionCarried(env, binding, cardPath, head, base) (bool, error); retain error propagation and cancellation.
- [ ] Add the pure eligibility decision after an ownership match. Example refusal: “#499 is working (reopened after its close); re-run sdlc close --issue 499 --verified '<evidence>', push, then retry publication.” Do not call reopened work a valid completion or create an anchor for it.
- [ ] Add one guard over the current records. Eligible cards continue through existing selection; ineligible cards with a carried valid current binding refuse. Existing malformed-card/binding refusals must not become silent skips.
- [ ] Call that guard from runPublishGate and runLandingPublishGate before their existing anchor selection. Pass their existing base/head and records scope; do not fetch separately per card or modify completeLandingPR/settleLandedCompletions/archive eligibility.
- [ ] Run new regressions and ownership controls GREEN; review the diff for accidental status or token semantic changes. Commit implementation and its red/green Log evidence.

### 3. Verification and documentation after approval

Files: atlas/workflow/sdlc-binary.md, issue Log.

- [ ] Map the publish-only reopened-card refusal, binding authority and re-close/push recovery in the existing landing-train/publish documentation.
- [ ] Run go test ./cmd/sdlc -run '^(TestCarReadiness|TestLandingPublish|TestRebasedClose|TestOlderCloseToken|TestSquashedClose|TestPublishFlip)' -count=1, plus the exact new test names and relevant tracker unit tests.
- [ ] Preserve one-snapshot-per-pass coverage with TestCarReadinessPublishChecksShareGeneration and the counted 8/16 case. Compare unaffected settlement/archive tests.
- [ ] Supply the hub an isolated production-gate proof using the real temporary tracker and existing fake GitHub. Ask the hub about any separate live staging; do not touch :0 on inference.
- [ ] Push the issue branch, run sdlc local-check run --name suite, and keep the session alive while its recorded run is queued/running. Close through sdlc close with measured actuals and verification; its mandatory fresh review is the review boundary.
- [ ] On SHIP enter sdlc merge --yes promptly, retaining the authoritative integration check. Report the confirmed PR merge and clean resting branch to the hub.

## Planning evidence and approval

Existing ownership/readiness controls passed together before edits: TestCarReadinessObservesExternalReclose, TestOlderCloseTokenNeverClaimsANewerClose, TestRebasedCloseIsStillOwnedAndGated and TestSquashedCloseRefusesLoudly, 42.504s. This is baseline protection, not a new reproduction or a full-suite claim. The issue preserves the original deterministic red observation from #499.

Await hub approval of the new refusal and ownership boundary before change-code or implementation. No estimate is entered until the binary determines the flow and, if full, completes plan-quality review.
