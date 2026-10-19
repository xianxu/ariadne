---
id: 000394
status: done
deps: [392]
github_issue:
created: 2026-10-15
updated: 2026-10-19
estimate_hours: 1.6
card_mirror: 'b24dc0c5da23aedeac9d5c1727f4633566dc4acc' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-18T15:44:55-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
actual_hours: 3.86
---

# Issue-backed flake landing and daily triage

## Problem

An intermittent suite failure currently either blocks a landing or invites an unexplained retry. Both hide the actual test-health cost. #390's approved rule requires one passing exact-head suite while letting an already identified, separately owned flake be diagnosed and rerun on the same head.

## Spec

Keep the exact-head full-suite landing gate. When a test fails, retain the attempt and its log. If that test already has an open named flake issue, attach the failed attempt to the issue; a passing full-suite rerun on the same head may count for landing. An unknown failure blocks until classified: a reproducible branch bug is fixed on that branch; a newly established intermittent fault gets its own issue that day and only then becomes eligible for the known-flake rule. A blind passing retry never erases an unexplained failure. The TL owns a daily standing list across local and CI failures, with first-seen time, head, run/log, issue, owner and root-cause investigation or an explicit diagnostic blocker within one day. Flakes are fixed as separate root-cause issues, not by raising timeouts. Report CI failures separately from same-head local flakes; neither detector may treat absent data as success. Follow #390 §3 and approved side doc 02; make the policy an ariadne default for every SDLC repo after the ariadne/pair pilot.

## Done when

- [ ] A full suite passes on the exact landing head in every permitted landing; a known issue-backed flake's failed attempt and passing same-head rerun are both retained and linked.
- [ ] An unknown failure blocks even if a retry passes; classification tests distinguish branch bug, newly filed flake, known flake, and persistent main failure.
- [ ] The TL's daily view lists unowned or older-than-one-day intermittent failures from local attempts and CI with source coverage and issue links; CI failure classes remain distinct from confirmed same-head flakes.
- [ ] Tests use recorded attempts and stateful gate behavior to prove failure, issue filing, retry and head drift; no retry or timeout change silently turns red evidence green.

## Revisions

- 2026-10-18 (hub D3): Done-when 3's CI failure classes move to #451. Here the daily view covers local attempts, and CI is shown as `not covered` until #451 lands, never as zero.

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: greenfield-go-module   design=0.2 impl=0.22
item: smaller-go-module      design=0.05 impl=0.14
item: smaller-go-module      design=0.05 impl=0.2
item: smaller-go-module      design=0.05 impl=0.14
item: atlas-docs             design=0.0 impl=0.08
item: milestone-review       design=0.0 impl=0.14
item: milestone-review       design=0.0 impl=0.14
item: milestone-review       design=0.0 impl=0.14
design-buffer: 0.15
total: 1.6
```

Derivation (*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only.*). The design is settled by the hub's D1–D4 and two plan-quality rounds, so a ×0.2 spec discount applies to the greenfield design (0.5–2h becomes about 0.2). impl= is 40% of the v2 midpoints:
- `internal/flakes` (registry, classify, evidence, rerun) is greenfield: 0.55 → 0.22.
- The card field, vocabulary and `flake register`/`list`: 0.35 → 0.14.
- The runLocalCheck gate, the `runUnderSlot` extraction and the v2 attempts, the riskiest piece, take the top of the range: 0.5 → 0.2.
- The success-post guard, republish and the `Unattributed` parse: 0.35 → 0.14.
- tl-worker help with its phrase guards, the help text and the atlas: 0.08.
- Three milestone reviews.

The calibration doc is marked stale (#127), so these numbers are provisional.

## Plan

Durable plan: `workshop/plans/000394-issue-backed-flake-landing-plan.md`, with the state table. It awaits the hub's decisions D1–D4 before code.

- [x] M1 — Registry and classifier (pure) plus `sdlc flake register`/`list`.
- [x] M2 — The gate: a targeted known-flake rerun in `runLocalCheck` (local-check run and merge), both attempts linked, the policy knob.
- [x] M3 — No blind rerun (the success-post guard, republish after registration), the rules in tl-worker/help/atlas, a pair pilot. **Pilot and rollout check: after landing (Revision 2026-10-18 M3 round 1; owners hub and ariadne:2), not done in M3.**

## Log

### 2026-10-15

Filed from #390's operator-approved flake/landing decision. The weekly snapshot #392 supplies the baseline and durable measurement view.

### 2026-10-18
- 2026-10-18: closed — Integrated origin/main (merge cf0284d2): resolved the localcheck_run.go conflict with #452 by threading suitePrioritySubject through runUnderSlot (the suite passes subject, the known-flake rerun passes g.subject). Targeted go test ./cmd/sdlc/... ok (423s); full suite green on cf0284d2. (fixed-then-ship; full suite: sdlc local-check run --name suite: exit 0 in 12m17s, 12 shards over 1667 cmd/sdlc tests and 71 other packages, recorded as sdlc-local/suite on cf0284d2); review verdict: SHIP
- 2026-10-18: closed — Suite green on bda6cc0e (sdlc-local/suite, 13m39s); the previous attempt on 7d3d7ab0 found two branch defects, fixed (flake register recovery contract; only a vanished record is skipped). M1 (registry as flake_tests card field, pure classifier/evidence/rerun, v2 attempts, Unattributed parse, flake register/list), M2 (targeted rerun gate in runLocalCheck for local-check run and merge, revalidation at post/reuse/republish, suite-rerun record kind, policy knob) and M3 (no blind rerun: guardSuitePass fail-closed; shared isSuiteRun/isFlakeEvidence; tl-worker rule) each closed SHIP. Pilot and :0 rollout moved to after landing (plan Revision, owners hub and ariadne:2).; review verdict: SHIP
- 2026-10-18: closed M3 — M3 round-1: BR-21 pilot and propagation moved to after landing with owners (plan Revision; M3 line marked); guard recovery wording per kind (flakes.Explainable; unattributable case asserts no register advice); pruned-mid-read records skipped; weekly/stats readers through isSuiteRun; test renamed TestRunLocalCheckUnreadableStoreRefusesTheSuitePass; atlas headroom note. Gate, local-check, republish, weekly and stats tests pass (342s).; review verdict: FIX-THEN-SHIP
- 2026-10-18: closed M2 — M2 round-2: reruns saved as check suite-rerun (decoder ties it to rerun_of; reserved in declarations), so every suite-attempt reader excludes them by construction and reuse comes from the parent; regression test: an inconclusive rerun is never reused as the head pass; BR-16 timeout bound tested; test-stats counts a parent known-flake landing as confirmed; decoder and declaration cases added. flakes, helptext and targeted cmd/sdlc (gate, local-check, CI-gate, landing, merge, stats, weekly, actual, queue) pass in 1699s.; review verdict: SHIP
- 2026-10-18: closed M1 — M1 round-2: mergeFlakeTests is the one pure decision production and the fake share; TestAddFlakeTestsMergesOntoTheCard on a real card (existing tests survive, cap counts them, closed refuses) and the real-card registry test asserts the old test survives; replacing the merge with tests := add fails four tests. loadFlakeRegistry returns cards; suiteStoreDir dedupes store setup. Flake tests pass.; review verdict: SHIP

- Dispatched by the hub (ops:0): the registry of known flakes is tied to their issues, and a rerun of just the registered tests replaces the whole suite. #449 builds on this. Plan drafted; it surfaces D1 (the targeted rerun amends #390's approved full-rerun rule), D2 (the accept policy), D3 (the daily view's scope: CI to a follow-up) and D4 (the registry lives in issue frontmatter).
- The hub decided D1–D4 as proposed: the targeted rerun counts (the operator's 10-18 ask); `accept`/`off` are built with the guarded config left at its default; CI in the daily view goes to a follow-up; the registry is issue frontmatter. Side doc 02 got a Revision note.
- Plan-quality round 1 answered in the plan's Revisions. PQ-4 changes the registry's storage from details frontmatter to a card field, with `sdlc flake register` as its setter, with no owner check and read via Fresh tracker records; the hub is told of the delta. PQ-1: the rerun takes its own slot, lease and signals through the shared `runUnderSlot`. PQ-2: classification happens in runLocalCheck; the main baseline stays merge's unrecorded diagnosis. PQ-3: the decoder reads v1 and v2, and the success guard fails closed.
- The hub approved the card-field registry (`flake_tests`, setter `sdlc flake register`, compare-and-swap with no owner check, read with the Fresh status read). Rollout: one rebuild of :0 after landing, before anything can write the field.
- M1 implemented:
  - `internal/flakes` (`BuildRegistry`, `Classify` over every state-table row, `Confirmed`/`SameHeadFailPass`/`Unexplained`, `RerunCommand`).
  - test-stats' confirmed-flake rule now calls `flakes.Confirmed` (ARCH-DRY).
  - Attempt records are version 2 (`outcome`, `known_flakes`, `rerun_of`); the decoder reads v1 and v2, and a v1 record carrying a v2 field refuses.
  - `SuiteLogFacts.Unattributed` (a package-level failure without a test of its own, a shard with no failing test, tests that never ran, a build or setup failure, a timeout, a name without a package).
  - The `flake_tests` card field (vocabulary kind `test-list`, at most 5 exact names).
  - `sdlc flake register` and `sdlc flake list`.
  - Tests: TestBuildRegistry, TestClassify, TestEvidence, TestRerunCommand, TestUnattributedFailures, TestParseCardFlakeTests, the v2 decoder cases, TestFlakeRegisterNeedsSameHeadEvidence, TestFlakeRegisterRules, TestFlakeList, and TestFlakeRegistryFromTrackerCards (real tracker cards). The cmd/sdlc tests on this surface pass (749s, targeted).
- M1 boundary review round 1 (FIX-THEN-SHIP): BR-1 (a test name can't become a go test flag or a relative path), BR-2 (the merge happens on the compare-and-swap card; racing registrations kept), BR-3 (list shows owner, age and log), BR-4 (README), and the three Minors (evidence text, closed issue refused, Unexplained policy) fixed in the plan's Revisions with tests.
- M1 boundary review round 2: BR-2 stayed open. The production merge was right but untested, since the race test's fake copied it. Fix: one pure `mergeFlakeTests` that both production and the fake call; `TestAddFlakeTestsMergesOntoTheCard` on a real card (existing tests survive, the cap counts them, a closed card refuses); the real-card registry test asserts the old test survives. Replacing the merge with `tests := add` now fails four tests. Minor (ARCH-DRY): `loadFlakeRegistry` returns the cards, and `suiteStoreDir` replaces the duplicated store setup. lessons.md has the rule.
- M1 closed SHIP (3 rounds). Two advisories were fixed in M2's first commit: one owner per test is enforced where the registry is read (register help, atlas), and the v2 attempt schema is noted as a breaking change for binaries from before #394 (atlas).
- M2 implemented:
  - `runUnderSlot` was extracted from `runLocalCheck` (slot, lease and signals; #383 release; the suite's behavior is unchanged and its local-check, republish, CI-gate, lease and cancel tests pass).
  - `known_flakes` on the suite declaration is parsed strictly and honored by the suite only.
  - `knownFlakeStep`: classify; on `NeedRerun`, one targeted rerun (weight 1) saved as a v2 attempt `rerun_of` the failed one; classify again. The outcome and issues go on `localRunRecord` (passed() honors a known-flake landing; the status reads "pass · known flake #N rerun").
  - Republish carries a saved attempt's outcome.
  - `gateFlakeRegistry(root)` is wired into `local-check run`, the durable landing and the legacy merge.
  - `diagnoseSuiteFailure` labels persistent main failures through `flakes.Classify`.
  - TestKnownFlakeGate covers: rerun lands, rerun fails, unknown (no rerun), accept, registry unavailable, unattributable, an interrupted rerun posting nothing, and a baseline that never classifies.
  - Help: local-check, merge and flake register. Atlas updated.
- 2026-10-18: the hub's quiet hold for the operator's disk trace. I stopped the targeted run on request and will rerun it when the hold lifts.
- M2 boundary review round 1 (REWORK) answered in the plan's Revisions. BR-11, the Critical: a timed-out suite could land on a partial log; failure signals the log can't name are now unattributed, and a rerun needs a PASS line per test. BR-12: revalidation at the post, at reuse and at republish. BR-13: reruns are evidence, not suite runs, with their logs immutable per ID. BR-14: docs. BR-15: tests. The three Minors are fixed too.
- M2 boundary review round 2 (REWORK): BR-19, the Critical (merge reused an inconclusive rerun as the head's pass), fixed at the representation. Reruns are check `suite-rerun`, tied to `rerun_of` by the decoder and reserved in declarations, so no suite reader can take one for a suite run. Reuse comes from the parent's outcome. BR-16: the timeout bound is tested.
- M2 closed SHIP (3 rounds). Its advisory, the third in the reader family, is fixed at the store: `isSuiteRun` and `isFlakeEvidence` are the two shared predicates. Test-stats, `flake list` and `flake register` read evidence through the second, so a passing rerun proves fail→pass even when revalidation downgraded its parent (TestSuiteStatsRerunEvidenceSurvivesADowngradedParent).
- M3 implemented:
  - `guardSuitePass` in `postLocalCheckRecord` (every run, reuse and republish) refuses a suite success while a failed attempt on the head is unexplained. It fails closed on an unreadable store, and it names the attempts and the recovery (register, then republish).
  - Five existing tests encoded the old loophole or the old "stats are optional" stance; they now assert the refusal. The behavior change: a pass on a head with an unexplained failure, or on an unreadable store, is recorded but not posted.
  - New gate cases: a blind rerun is refused, then register plus republish lands; an unattributable failure stays unexplained.
  - The tl-worker help's flake rule now describes the new mechanism, citing the operator's 10-18 ask; phrase guards updated.
  - Help, README and atlas cover the guard.
- M3 boundary review round 1 (FIX-THEN-SHIP): BR-21 (the pilot and propagation move to after landing, with owners, in the plan's Revisions; the M3 line now says so). Minors: per-kind recovery wording, the pruned-record race, readers routed through `isSuiteRun`, a test renamed.
- M3 closed SHIP (2 rounds). Its advisory: merge help names the no-blind-rerun guard. The recorded suite on 7d3d7ab0 found two branch defects, both fixed here: (1) `flake register` lacked a recovery contract; it now has one (NonRepeatable; with `--issue` a repeat writes nothing, proved by TestFlakeRegistryFromTrackerCards); (2) skipping a record missing at open also skipped a dangling symlink, which must stay unreadable (TestOwnedActualQueueCoverageRefusesUnknownRecords). A record is skipped only when `Lstat` also finds nothing.
