# Boundary Review — ariadne#449 (milestone M2)

| field | value |
|-------|-------|
| issue | 449 — landing train: batch ready PRs on an interval and land them with one full suite (extends #393) |
| repo | ariadne |
| issue file | workshop/issues/000449-landing-train-batch-ready-prs-on-an-interval-and-land-them-with-one-full-suite-extends-393.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 4f6dfd5a183f881e6007d8f73747e8d4ff2566ab..d82c4f30c30bf529b0b3cce2ad5c26bb16d3e969 |
| command | sdlc milestone-close --issue 449 --milestone M2 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-19T05:02:37-07:00 |
| verdict | REWORK |

## Review

Now the core new file.

Reviewing the train state machine's merge/probe transitions now.

Checking README/atlas docs and the remaining concerns (waitCI's global `ghClient`, reclaim scope).

```verdict
verdict: REWORK
confidence: medium
```

M2 delivers most of what it promises. The conductor runs the effects of the pure `landingline.Transition`. One train builds one branch, runs one suite, and merges once. All three `mergeCommit` variants work end to end through resume, archive and card completion. Every `MergeOID` consumer now reads `LandedOID`, and a grep test enforces that. I ran `go test ./cmd/sdlc/internal/landingline/` and `go test -run 'TestLandingTrain|TestLandedCommit|MergeOID' ./cmd/sdlc/` at the pinned head, and both pass.

One thing blocks SHIP: the uncertain-merge path. A car whose PR is merged while it is still waiting in the line never leaves the wait when trains are on. The plan's state table says the next head or the car's resume settles this "automatically". Instead, the car waits until an operator interrupts it. That is reachable in three ways:
- a conductor dies after `gh pr merge`;
- a `Settle` call fails;
- the conductor aborts after an unconfirmed merge.

The conductor also mistakes its own successful merge for an outside move of main, which feeds the third case. Separately, a merged train branch is never deleted.

### 1. Strengths
- **Pure core, thin shell (ARCH-PURE / ARCH-ORDER).** `conductTrain` (`cmd/sdlc/landingtrain.go:100-146`) is an effect loop around the pure `Transition`. Each IO effect returns its outcome as an event.
- **One landed-commit rule (ARCH-DRY / ARCH-PURPOSE).** `landedCommit` (`cmd/sdlc/ghlanding.go:187`) and `LandedOID` cover every consumer: `landing.go:616`, `landingarchive.go:162/634`, `trackercompletion.go:249`, plus the decoder. `TestOnlyLandedCommitReadsMergeOID` (`landingcar_test.go:395`) enforces the whole class, not just one site. I confirmed by grep that no non-test file outside `ghlanding.go` reads `MergeOID`.
- **One shared codec (ARCH-EVOLVE).** `codec.go` serves both sidecars. It is strict on known members (exact names, no duplicates, no nulls) and keeps unknown members across `Settle` and `Vouch` rewrites. `TestOutcomeCodecIsStrictAndPreserving` tests this.
- **A vouch is only a claim (ARCH-SECURE).** `carStillAtHead` re-reads each car's tip and PR, and `build` checks the fetched head against the vouched one. A forged `landed` outcome still has to pass `awaitMergedByTrain` and `landedCommit` against fetched main before any archive happens.
- **A stateful fake (ARCH-MOCK).** `trainFakeGH` uses a real bare remote and merges by reachability, and all three variants run through real `sdlc merge` processes.

### 2. Critical findings
- **C1 — A car whose PR merged while it waited never leaves the line (ARCH-ORDER: uncertainty collapsed into "not ready").**
  - **Where:** `cmd/sdlc/landingcar.go:56-58`, `cmd/sdlc/landingcar.go:156-175`, `cmd/sdlc/internal/landingline/line.go` `Wait`.
  - **Mechanism:** once a car's PR is MERGED, `carReadiness` keeps returning "no open PR at …". The train admission rule admits only granted tickets, or vouched-ready ones once the train is due. `Wait` returns only on a final outcome. So the car can never be admitted and never resumes.
  - **Ways to reach it:**
    1. The conductor dies after its merge lands. Its boarded cars keep the `boarded` outcome forever. This is exactly the plan's row "Conductor dies after gh pr merge was sent … every car whose PR is MERGED → L".
    2. `Settle(landed)` fails with an error other than `ErrGone`. The warning at `landingtrain.go:495` then claims "its sdlc merge resumes from GitHub's state", which is false.
    3. The conductor's abort path releases cars whose train PR actually merged (see I1).
  - **Test gap:** `TestLandingTrainReclaimsAnOrphanAndSurvivesALostMergeReply` covers none of these. Its orphan PR has no boarded cars, and its lost reply is visible to the very next probe.
  - **Fix sketch:** make "my PR is MERGED" an exit from the wait. Either the vouching goroutine sees a merged PR for its branch and ends `Wait` with a resume, or the admission rule admits a ticket whose vouch reports a merged PR, which is today's serial resume. Add a sequence test: a conductor is killed after `LandingMerge`, and a boarded car must finish cleanup.

### 3. Important findings
- **I1 — The conductor treats its own landed merge as "main moved" (ARCH-ORDER).**
  - **Where:** `cmd/sdlc/landingtrain.go:399-418`.
  - **Sequence:**
    1. `LandingMerge` succeeds, or its reply is lost.
    2. The probe still sees OPEN because the listing lags. `awaitMergedByTrain` exists because such lag happens.
    3. `Probed{false}` leads to a second `Merge`.
    4. `fetchMain` now differs from `c.base`, because main contains the train merge.
    5. That produces `MainMoved` and `Reintegrate`.
    6. `publish` force-pushes and polls for an OPEN train PR that will never come back. It times out, the abort defer releases every car and records `abandoned`, and the conductor returns an error.
  - **Result:** all cars, merged on GitHub, are back to waiting and hit C1. The second-`Probed{false}` branch in `trainstate.go:272-276` releases them the same way.
  - **Fix sketch:** in `merge()`, before treating a moved main as `MainMoved`, check whether `c.heads[k-1]` is an ancestor of the fetched main. If it is, that is confirmed `Merged`, the same rule `landedCommit` uses. Add a fake mode where `LandingPRs` shows OPEN for one or two reads after a merge.
- **I2 — Merged train branches are never deleted (ARCH-FUNERAL).**
  - **Where:** `realGH.LandingMerge` (`cmd/sdlc/ghlanding.go:172`) runs `gh pr merge … --merge --match-head-commit` with no `--delete-branch`. After `Merged`, the conductor never deletes `landing-train/<id>`. `reclaimTrains` lists only OPEN PRs.
  - **Result:** every landed train leaves its branch on the remote. The plan says "The merge deletes the branch (as today)".
  - **Fix sketch:** after `LandCars`, call `deleteRemoteBranch(train branch, c.pushed)`, which is leased. Assert in `landTrainOfThree` that no `refs/heads/landing-train/*` is left on the remote.

### 4. Minor findings
- **ARCH-ORDER:** the abort defer sets `tr.Phase = PhaseAbandoned` directly (`landingtrain.go:111`), bypassing `Transition`. The cars it releases are not in `tr.Released`, so `NextDue` waits a full interval for cars it just called "still ready". An `Aborted` event in the state machine would fix both.
- **ARCH-MOCK:** `waitCI` passes the global `ghClient` (`landingtrain.go:385`), and `maybeConductTrain` type-asserts the global (`:591`), instead of using the injected `c.gh` seam.
- **ARCH-FUNERAL:** `reclaimTrains` runs only when two or more cars board. A branch pushed before its PR was created (the conductor died between the two) is never reclaimed.
- `maybeConductTrain` records the "alone" train event, which sets `next_due`, before the single landing runs. A refused solo landing still pushes the next departure back a full interval.

### 5. Test coverage notes
- The done-when test is solid. It covers three cars, one suite, one merge, and three `mergeCommit` variants with real git and a real repo lock.
- Missing:
  - a lagging probe after an accepted merge (I1);
  - a conductor killed after the merge, with boarded cars (C1);
  - a failed `Settle` (C1);
  - remote branch residue after a train (I2).
- The "orphan" test pushes a fake `landing-train/7` with no cars. It shows reclaim closes PRs, not the plan's dead-conductor recovery.
- The train tests observe one interleaving each. Nothing seam-injects the order of a merge reply against probe visibility.

### 6. Architectural notes for upcoming work (M3)
- **ARCH-DRY:** pass. The codec and `landedCommit` are shared, and reintegration reuses `publishGateHasCodeSurface`.
- **ARCH-PURE:** pass.
- **ARCH-PURPOSE:** pass for the PQ-3 sweep; flag for the uncertain-outcome rows of the state table (C1).
- **ARCH-MOCK:** pass, with the Minor about the global `ghClient`.
- **ARCH-CONSTRAINTS:** pass. Waits are bounded by `integrationSyncTimeout` and `trainCarWait`.
- **ARCH-SECURE:** pass.
- **ARCH-ORDER:** flag (C1, I1, and the Minor about the abort defer).
- **ARCH-FUNERAL:** flag (I2).
- **ARCH-EVOLVE:** pass. Refusing an unknown outcome version leaves the car waiting, which is the safe default.
- **For M3:** bisect, holds and CI-red refusals multiply the cars that end up "released after uncertainty". Fix C1's exit rule first so M3's paths inherit it.

### 7. Plan revision recommendations
- Add a `## Revisions` entry stating where a waiting car learns that its PR merged without an outcome, and what evidence settles an unconfirmed train merge: the train head being an ancestor of fetched main.
- The ARCH-FUNERAL line "The merge deletes the branch (as today)" should name the conductor's explicit deletion, since `gh pr merge` here does not delete it.

```findings
findings:
  - id: new
    severity: Critical
    kind: defect
    family: uncertain-outcome-collapsed
    title: |
      With trains on, a waiting car whose PR is already MERGED can never be admitted or settled and hangs
    detail: |
      carReadiness (landingcar.go:56) returns not ready for a merged PR, the train admission rule (landingcar.go:156-175) admits only granted or vouched-ready tickets, and Wait returns only on a final outcome. A conductor that dies after its merge lands, a Settle that fails (landingtrain.go:495 wrongly says the car resumes), or an abort after an unconfirmed merge all leave cars waiting until an operator interrupts. This contradicts the plan's state-table row (the next head or the car's resume settles every MERGED car). Fix: make "own PR is MERGED" an exit from Wait (resume) and add a killed-conductor-after-merge sequence test.
  - id: new
    severity: Important
    kind: defect
    family: uncertain-outcome-collapsed
    title: |
      The conductor's merge() treats its own landed train merge as an outside move of main, then aborts and releases merged cars
    detail: |
      After LandingMerge is accepted but the probe lags, Probed{false} re-sends Merge. fetchMain then differs from c.base because main contains the train merge, so the conductor reintegrates, publish times out waiting for an OPEN train PR, and the abort defer releases every car (which then hits the hang above). Fix: in merge() (landingtrain.go:399-418), treat c.heads[k-1] being an ancestor of the fetched main as Merged before emitting MainMoved; add a fake mode where the listing lags.
  - id: new
    severity: Important
    kind: defect
    family: artifact-without-removal
    title: |
      A merged train's landing-train/<id> branch is never deleted from the remote
    detail: |
      realGH.LandingMerge (ghlanding.go:172) has no --delete-branch, the conductor deletes nothing after Merged, and reclaimTrains lists only OPEN PRs, so every landed train leaves a branch on the remote. The plan says the merge deletes it. Fix: run a leased deleteRemoteBranch on the train branch after LandCars, and assert in landTrainOfThree that no landing-train ref remains.
  - id: new
    severity: Minor
    kind: defect
    family: state-machine-bypass
    title: |
      The abort defer sets tr.Phase directly, outside Transition, and its released cars miss tr.Released, so NextDue waits a full interval
    detail: |
      landingtrain.go:104-112. Model the abort as a Transition event so released cars make the next train depart at once.
  - id: new
    severity: Minor
    kind: defect
    family: seam-bypass-global-client
    title: |
      waitCI and maybeConductTrain use the global ghClient instead of the injected trainGH seam
    detail: |
      landingtrain.go:385 and landingtrain.go:591.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"67360e34-42b2-4c6f-b75d-384157d88394","repository":"github.com/xianxu/ariadne","issue_id":"000449","recipe":"milestone-review","round":1,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/728725127a45d5d1","reviewed_head":"d82c4f30c30bf529b0b3cce2ad5c26bb16d3e969","slot":"ariadne:2","requested_at":"2026-10-19T11:58:25.229904Z"},"phase":"finished","pid":17703,"started_at":"2026-10-19T11:58:25.344938Z","finished_at":"2026-10-19T12:02:37.196507Z","session_id":"4d4bb86f-80f9-42f3-a9c9-b765603e5060","outcome":"returned"}] -->

---

## Re-review — 2026-10-19T06:12:06-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 449 — landing train: batch ready PRs on an interval and land them with one full suite (extends #393) |
| repo | ariadne |
| issue file | workshop/issues/000449-landing-train-batch-ready-prs-on-an-interval-and-land-them-with-one-full-suite-extends-393.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 4f6dfd5a183f881e6007d8f73747e8d4ff2566ab..a395730bed6d00be7b996616f427d8b8594ff447 |
| command | sdlc milestone-close --issue 449 --milestone M2 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-19T06:12:06-07:00 |
| verdict | SHIP |

## Review

Let me look at the last commit (the round-1 fix) first, then the main files.

The landing test suite hit Go's 10-minute timeout inside a train test. Next I'll find out which test hangs.

Train tests pass on their own (54s). The earlier timeout may just be the wider set running past 10 minutes; I'm rerunning it verbosely to see which test was in flight.

Not a hang: lots of landing tests take 20–60s each, so the total goes past the limit. Next I'll run the test that was in flight on its own, then the rest of the set.

Everything else passes (227s). Now I'm reverting each claimed fix one at a time to check that its regression test fails without it.

BR-16's test fails without the fix. Next, BR-17 and BR-18.

Both go red without their fixes. Now checking for any remaining global-client use (BR-20) and how the merged-while-waiting resume proves the landing.

Every prior fix checks out, and none of the remaining issues are blocking. Confirming the checkout is clean, then writing the verdict.

```verdict
verdict: SHIP
confidence: high
```

All five round-1 findings are fixed, and the three behavioural ones (BR-16, BR-17, BR-18) each have a regression test that I watched fail with the fix reverted. A car whose PR is already merged now stops waiting and finishes as a resume (BR-16). The conductor now treats its own train head on main as its merge, not as main moving (BR-17). Landed train branches are deleted, and leftovers are swept (BR-18). The abort is now an `Aborted` event inside `Transition` (BR-19), and the conductor uses the seams it is given (BR-20). The window's tests all pass, run in three batches. One combined run hit Go's 10-minute limit only because of total time: about 150 landing tests at up to 60s each, and the test that was running passes alone in 3s. What's left is five Minors, none of which block.

**1. Strengths**
- The merged-while-waiting exit sits in the car's own view of GitHub (`landingcar.go:58-61`, `landingqueue.go:136-151`). So a car never depends on a conductor that may have died; that is the right owner for the fact.
- `Aborted` is a real event in the pure machine (`trainstate.go:171-183`) and is covered by the random-sequence test (`trainstate_test.go:280`). The abort path now goes through `Transition`, not around it.
- The fake gained a lag mode (`landingtrain_test.go:67-72`) that models GitHub's listing delay across calls. This is a stateful double in the ARCH-MOCK sense, not a stub.
- Branch deletions are leased on the commit last pushed or observed (`landingtrain.go:553`, `:619`), so a delete can't remove a commit the deleter never saw.

**2. Critical:** none.

**3. Important:** none.

**4. Minor**
- `landingtrain.go:110-128`: when the train is already merging, an IO failure (such as a failed probe after an unconfirmed merge) is recorded as `abandoned`, and its cars as `released`, even though the train may have landed. This is the 3rd finding in `uncertain-outcome-collapsed`. The rule: a terminal train record needs a confirming observation, and an unconfirmed outcome is recorded as unknown.
- `landingcar.go:100`: `productionCarGates` still uses the global `ghClient`. This is a sibling of BR-20 and the 2nd finding in `seam-bypass-global-client`. The rule: no code under `landing*.go` reads `ghClient` except at the point where the merge is wired up (`landing.go:360`).
- The remote branch is deleted by three hand-written pushes (`landingtrain.go:532`, `:553`, `:619`), and only `closePR`'s is unleased. This is the 3rd finding in `duplicated-rule`. The rule: there should be one leased `deleteRemoteBranch` helper.
- The atlas "Leftovers" bullet (`atlas/workflow/sdlc-binary.md:436`) doesn't mention that landed branches are deleted, that leftover branches are swept, or that a car merged while waiting resumes itself.
- Plan line 197: the list of events the code implements leaves out `Aborted`. This is the 3rd finding in `plan-core-concepts-drift`. The rule: that list mirrors every type that implements `trainEvent`/`trainEffect`.

**5. Test coverage**
I reverted each of BR-16, BR-17 and BR-18 and its test went red:

| Finding | Test | Result without the fix |
|---|---|---|
| BR-16 | `TestLandingCarMergedWithoutAWordFinishes` | merge did not finish, 121s |
| BR-17 | `TestLandingTrainLaggingListingIsItsOwnMerge` | merge did not finish, 125s |
| BR-18 | `landTrainOfThree`, all 3 variants | "train branches outlived their trains" |

- **BR-19:** checked by the pure test `TestTrainAbortReleasesEveryRidingCar`.
- **BR-20:** this is wiring only. In production the injected value is `ghClient` itself, so no test can tell the two apart. I checked it by reading the code.
- **Gap:** no test drives a probe failure after an unconfirmed merge (the first Minor).

**6. Architecture notes**
- **ARCH-DRY:** flag (third Minor).
- **ARCH-PURE:** pass.
- **ARCH-PURPOSE:** pass; the seam sibling is noted.
- **ARCH-MOCK:** pass; the fake now has a lag mode.
- **ARCH-CONSTRAINTS:** pass. A merged car is noticed within one vouch interval (1 minute).
- **ARCH-SECURE:** pass. Branch names from `ls-remote` are prefix-checked and passed as separate arguments.
- **ARCH-ORDER:** pass, apart from the merging-phase abort (first Minor).
- **ARCH-FUNERAL:** pass. Train branches now have both a removal path and a sweep.
- **ARCH-EVOLVE:** pass; no persisted format changed in this round.
- **For later:** `OpenTrainPRs` lists all open PRs up to 100 and has no overflow check. The new sweep makes a missed orphan harmless, but a `--search head:landing-train/` query would be exact.

**7. Plan revisions**
Add a "## Revisions" entry that lists `Aborted{Reason}` among the events, saying that an IO failure releases every car still riding and closes the PR.

```findings
dispose:
  - id: BR-16
    disposition: addressed
    note: |
      carReadiness reports merged and vouchWhileWaiting ends the wait; TestLandingCarMergedWithoutAWordFinishes goes red (hang, 121s) with the check reverted.
  - id: BR-17
    disposition: addressed
    note: |
      merge() treats head-on-main as Merged (landingtrain.go:428); TestLandingTrainLaggingListingIsItsOwnMerge goes red with the ancestor check reverted.
  - id: BR-18
    disposition: addressed
    note: |
      deleteTrainBranch on PhaseLanded plus the reclaim sweep; landTrainOfThree fails in all 3 variants (landing-train/2 remains) with the call reverted.
  - id: BR-19
    disposition: addressed
    note: |
      Aborted is a Transition event that appends to Released; pinned by TestTrainAbortReleasesEveryRidingCar and the random-sequence test.
  - id: BR-20
    disposition: addressed
    note: |
      waitCI uses c.ci and maybeConductTrain uses the injected lgh/ci; wiring only, checked by reading the code (the sibling at landingcar.go:100 is raised separately).
findings:
  - id: new
    severity: Minor
    kind: defect
    family: uncertain-outcome-collapsed
    title: |
      An abort while merging records the train abandoned and its cars released although the merge may have landed
    detail: |
      landingtrain.go:110-128 with trainstate.go:171-183: a probe IO error after MergeUnknown aborts into PhaseAbandoned and closes the PR. This is the 3rd finding in this family. Rule: a terminal train record needs a confirming observation; from PhaseMerging, Aborted should record merge-unknown, not abandoned.
  - id: new
    severity: Minor
    kind: defect
    family: seam-bypass-global-client
    title: |
      productionCarGates still calls the global ghClient for the car CI gate
    detail: |
      landingcar.go:100, a sibling of BR-20. Rule: no landing*.go code reads ghClient except where the merge is wired up (landing.go:360); pass the seam in.
  - id: new
    severity: Minor
    kind: defect
    family: duplicated-rule
    title: |
      Remote train-branch deletion is hand-written three times, and only closePR's is unleased
    detail: |
      landingtrain.go:532, 553 and 619. This is the 3rd finding in this family. Rule: one leased deleteRemoteBranch(branch, tip) helper for every train-branch delete.
  - id: new
    severity: Minor
    kind: defect
    family: docs-gate-atlas
    title: |
      The atlas conductor section does not describe landed-branch deletion, the leftover-branch sweep, or the merged-while-waiting resume
    detail: |
      atlas/workflow/sdlc-binary.md:431-436 still says only that open leftover PRs are closed.
  - id: new
    severity: Minor
    kind: defect
    family: plan-core-concepts-drift
    title: |
      The plan's list of events the code implements leaves out Aborted
    detail: |
      Plan line 197 lists events without Aborted (trainstate.go:102). This is the 3rd finding in this family. Rule: that list mirrors every type that implements trainEvent and trainEffect.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"b70e3700-f4b8-4293-bbb3-19baadb4369f","repository":"github.com/xianxu/ariadne","issue_id":"000449","recipe":"milestone-review","round":2,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/ac572036146ee30f","reviewed_head":"a395730bed6d00be7b996616f427d8b8594ff447","slot":"ariadne:2","requested_at":"2026-10-19T12:37:54.938341Z"},"phase":"finished","pid":75404,"started_at":"2026-10-19T12:37:55.034639Z","finished_at":"2026-10-19T13:12:05.608555Z","session_id":"5ba1e5bb-a8ef-45e1-a772-1c9e5c5bd13d","outcome":"returned"}] -->
