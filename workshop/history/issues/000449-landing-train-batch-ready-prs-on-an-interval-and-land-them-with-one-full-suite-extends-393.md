---
id: 000449
status: done
deps: [ariadne#394]
github_issue:
created: 2026-10-18
updated: 2026-10-19
estimate_hours: 5.47
collaboration-mode: collaboration
card_mirror: 'd2550d5e700cb234a1c16c9f39027d9f3d197faa' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-19T00:24:20-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
actual_hours: 3.92
---

# landing train: batch ready PRs on an interval and land them with one full suite (extends #393)

## Problem

The landing line (#393) lands one PR at a time, each with its own full suite on the combined code. On a busy day that is one full suite per landing, and suites queued 20-50 minutes for the two machine slots (2026-10-18). The operator wants the suite frequency to be a knob.

## Spec

A landing train, extending #393: on a configurable interval (default hourly), take every PR that is admission-ready (review SHIP, branch suite recorded), merge them in admission order onto main in one integration, run one full suite, and land them all if it passes. On failure, find the culprit by splitting (bisect the batch) and land the clean prefix; a known, issue-backed flake (#394) gets its one rerun rather than splitting. A fleet-blocking grant lands alone and immediately, outside the train. Report per batch: PRs, suite result, splits.

## Done when

- With N ready PRs, one train runs one integration suite and lands all N when it passes (test with fakes).
- A failing batch is split and the culprit PR is reported and held while the rest land.
- The interval is configurable; a fleet-blocking grant bypasses the train.
- Depends on #394 (known-flake landing).

## Plan

Durable plan: `workshop/plans/000449-landing-train-plan.md` (state table, Q1–Q7 for the hub).

- [x] M1 — Pure core and readiness: schedule, readiness sidecar, boarding, train state machine, bisect, additive events, landing status, `.sdlc/landing.json`; interval 0 equals today.
- [x] M2 — Conductor: train branch and PR, one suite, CI, merge; cars resume on merged; GitHub fake merges by reachability.
- [x] M3 — Split on unexplained failure, known-flake rerun, grant bypass, conflict and CI-red handling, per-train report, live conformance, docs.

## Estimate

Derived against estimate-logic-v3.1 (calibration doc flagged stale). Familiarity 0.9: the landing line, the suite runner and #394's flake step are warm from the last item. Milestone reviews are priced higher than the default, because #394's converge took 1.66h over 10 rounds.

```estimate
model: estimate-logic-v3.1
familiarity: 0.9
item: greenfield-go-module   design=0.4 impl=0.5
item: greenfield-go-module   design=0.4 impl=0.8
item: smaller-go-module      design=0.2 impl=0.4
item: smaller-go-module      design=0.1 impl=0.3
item: smaller-go-module      design=0.2 impl=0.4
item: atlas-docs             design=0.0 impl=0.3
item: milestone-review       design=0.0 impl=0.5
item: milestone-review       design=0.0 impl=0.5
item: milestone-review       design=0.0 impl=0.5
design-buffer: 0.3
total: 5.47
```

Items in order:
1. The train core: schedule, boarding, state machine and bisect.
2. The conductor shell: train branch, PR, suite, CI and merge.
3. Car readiness and the resolved landed commit.
4. The fake that merges by reachability, plus the readiness sidecar.
5. Split, grant bypass and conflict holds.
6. Docs.
7. Three milestone reviews.

## Log

### 2026-10-18

### 2026-10-19
- 2026-10-19: closed — M1-M3 closed (M3 SHIP). Full suite: sdlc local-check run --name suite exit 0 in 10m28s, recorded as sdlc-local/suite on 74171bfb. Done-when: TestLandingTrainLandsEveryCarOnOneSuite (3 ready PRs, one train, one integration suite, all land; train/null/head mergeCommit variants); TestLandingTrainSplitHoldsTheCulpritAndLandsThePrefix (culprit held and reported, rest land/ride next); interval configurable via .sdlc/landing.json (default off), TestLandingTrainGrantDepartsAloneWhenNoTrainIsDue (grant bypasses the train); known flake is a train pass (TestTrainSuiteOutcome).; review verdict: FIX-THEN-SHIP
- 2026-10-19: closed M3 — Round 1 fixes (b1ab3a4ce): README train paragraph swept against the landing help; GrantsNow at every grant site; Board passes over a granted waiter (tested); seam test globs every landing file; landingline ok; go test -run Landing|Merge|Integrat|CIGate|Local|Archive|Tracker|CloseMirror|Train ./cmd/sdlc/ ok in 2510s; review verdict: SHIP
- 2026-10-19: closed M2 — Round 1 fixes (a395730be): TestLandingCarMergedWithoutAWordFinishes (BR-16), TestLandingTrainLaggingListingIsItsOwnMerge (BR-17), train-branch deletion and leftover sweep asserted in the three-car and orphan tests (BR-18), TestTrainAbortReleasesEveryRidingCar plus Aborted in the random-history invariants; train tests under -race; go test -run Landing|Merge|Integrat|CIGate|Local|Archive|Tracker|CloseMirror|Train ./cmd/sdlc/ ok in 1802s; review verdict: SHIP
- 2026-10-19: closed M1 — Round 2 fixes (96d1a9d0): schedule source mutex plus 30s cache, read before the line mutex (TestLandingScheduleSourceIsSharedAndCached under -race, 8 goroutines, 1 read); status through Readiness.For (tested); Passed() single audit rule (tested); landingline under -race ok; go test -run Landing|Merge|Integrat|CIGate|Local|Archive|Tracker ./cmd/sdlc/ ok in 2094s; review verdict: FIX-THEN-SHIP

- Claimed and started from ariadne:2. Explored the #393 landing line (flock tickets, head.lock, events.jsonl; no daemon in sdlc, so the schedule rides on admission). Wrote the durable plan; sent its summary and Q1–Q7 to the hub before code.
- Plan-quality: 3 rounds. PQ-1: `.sdlc/*` IS local-gates guarded, corrected to the hub, so the pilot needs the operator label. PQ-2: the sidecar moved to `ready/`. PQ-3: every `MergeOID` consumer goes through one landed commit (M2). Estimate 5.47h.
- M1 built:
  - the pure core (4c755312): a mutation of split is caught by the random-history invariants;
  - the car side, config, admission and status (5918c6fd).
- Found while reading for M2: `parseLandingPRs` refuses a MERGED PR without `mergeCommit` (ghlanding.go). That is a fifth `MergeOID` consumer beyond PQ-3's list, and it moves into the landed-commit rule in M2.
- #460 audit (ariadne:1): the readiness sidecar keeps unknown members across a same-ticket rewrite, with strict known members and no duplicate keys, plus regressions.
- Verification: the landingline package; the new cmd tests under -race; `go test -run 'Landing|Merge|Integrat|CIGate|Local|Archive|Tracker' ./cmd/sdlc/` ok in 1922s, so interval 0 leaves landing unchanged.
- M1 review round 1, FIX-THEN-SHIP. All nine findings fixed in the same round:
  - BR-1: the `Handle` mutex. Vouching ends quietly on `ErrNotWaiting`, with a -race regression.
  - BR-2: `Built` must account for every car; the generator now produces bad reports.
  - BR-3: a vouch must name its ticket's branch (`Readiness.For`), and `Car.Branch` comes from the ticket.
  - BR-4: a plan revision.
  - BR-5: README.
  - Minors:
    - admission rule built once per attempt;
    - the grant rule lives only in admission;
    - one `landingMainRef`;
    - the schedule is re-read each attempt, and an unreadable one is serial;
    - skipped tickets are in the displaced audit.
- M1 round-3 Minors fixed (29052514): the departure ledger is read under the line mutex after head.lock, and the docs now match the 30s cache.
- M2 progress:
  - landedCommit/LandedOID for every merged-PR consumer, plus the decoder (2fdafe42);
  - outcomes/<seq> car channel and the shared sidecar codec (279dcf4c).
- M2 conductor design, as coded next:
  - **Who conducts:** the admitted head conducts when Board yields ≥2 cars. One car is today's path, plus a train event so the schedule paces.
  - **Before boarding:** each car's remote tip and open PR at its vouched head are re-read; a mismatch stays waiting.
  - **Build:** a detached worktree from fetched main, `merge --no-ff` of each car head with Landing-Train/Landing-PR trailers. The instance and duplicate-ID gates run per prefix; a conflict or gate failure means that car is a Conflict.
  - **Publish:** push to `landing-train/<id>`; open the train PR once; wait for GitHub to register the head.
  - **Suite:** the local-check runner on the prefix head, with #394 inside it. On the first full-train failure, one unrecorded main baseline decides MainRed.
  - **CI and merge:** `waitCIGate` on the train PR. Main is re-fetched before the merge; if it moved, Reintegrate (merge main into the head) and NeedSuite = code surface in the delta. The merge is `LandingMerge` pinned to the head, then confirmed on GitHub.
  - **Outcomes:** written with `Line.Settle`, and car events go to events.jsonl.
  - **The conductor's own car:** it continues into today's post-merge cleanup once its PR shows MERGED.
  - **Cars:** a landed car waits for its PR to show MERGED, then takes the resume cleanup path.
  - **Abort:** any conductor error releases every unsettled car and closes the train PR.
  - **Reclaim:** at start, any open `landing-train/*` PR is closed, since its conductor is dead (head.lock was free).
- M2 built:
  - the conductor (930e8490), then robustness and the remaining cases (5b4ddd66);
  - TestLandingTrainLandsEveryCarOnOneSuite: three cars, one suite, one merge, under each mergeCommit variant;
  - plus a train of one, the orphan reclaim with a lost merge reply, and the docs-only reintegration.
- Found while testing:
  - the conductor's fetches raced the landed cars' cleanup on shared refs, so ref writes now take the repo lock;
  - the suite declaration has to come from the exact main the train is built on, not the stale tracking ref;
  - the grep test caught the probe reading the raw mergeCommit.
- M2 closed SHIP (bf4a625f) after round 1:
  - BR-16 Critical: a car whose PR merged with no word now resumes;
  - BR-17: a lagging listing is the train's own merge;
  - BR-18: train branches are deleted and swept.
  - The close-review Minors are fixed as rules (f2ff9bd74):
    - merge-unconfirmed;
    - one leased delete helper;
    - a seam test;
    - a test that keeps the plan naming every train event and effect.
- M3 built:
  - the grant holder lands alone;
  - e2e split (culprit held, prefix landed, remainder released and admitted next);
  - conflict hold;
  - red train CI refuses the whole train;
  - grant departs while no train is due;
  - suite-outcome mapping (known flake is a pass, main red);
  - `landing status` lists the last trains;
  - help text for landing and merge, and the atlas.

### Pilot plan (for the hub, after landing)

1. **Rollout:** the hub rebuilds :0 (and `weave refresh` in the slots). Trains stay off everywhere: no `.sdlc/landing.json` exists.
2. **Turn on in pair:** a pair PR adds `.sdlc/landing.json` with `{"train_interval_minutes": 60}`. It is a local-gates path, so the operator verifies it with the `local-gates-verified` label.
3. **Conformance (Q7):** the first pair train with at least two cars, low-risk where possible, is the live check of merged-by-reachability. For each car record:
   - `gh pr view N --json state,mergeCommit,headRefOid`;
   - the train PR's merge commit;
   - the train's `sdlc landing status` line.

   If GitHub's `mergeCommit` for a car is not the train merge, set the fake's default variant to what GitHub reports (resume already accepts null, the head and the train commit).
4. **Watch:** suites per landed issue (test-stats) and the p90 ready-to-landed delay against serial. An interval change is another operator-verified PR.
- Close review FIX-THEN-SHIP:
  - The one open Minor, BR-23, is fixed: an opened train PR is closed without gh's unleased `--delete-branch`, and its branch is then deleted through `deleteTrainRef` (leased, under the repo lock), as the reclaim's closed leftovers are.
  - The fake keeps a closed PR's branch, as GitHub does, so the red-CI and orphan tests now prove the leased delete.
  - The reviewer's open question, whether `go test ./cmd/sdlc` finishes in time, is answered by the full recorded suite: exit 0 in 10m28s on 74171bfb.
