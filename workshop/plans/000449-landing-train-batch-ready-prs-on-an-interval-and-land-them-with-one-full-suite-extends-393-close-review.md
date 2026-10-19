# Boundary Review — ariadne#449 (whole-issue close)

| field | value |
|-------|-------|
| issue | 449 — landing train: batch ready PRs on an interval and land them with one full suite (extends #393) |
| repo | ariadne |
| issue file | workshop/issues/000449-landing-train-batch-ready-prs-on-an-interval-and-land-them-with-one-full-suite-extends-393.md |
| boundary | whole-issue close |
| milestone | — |
| window | e279f713f66ebde2d11e8662e05d4e481369d34c..74171bfb3611d7c69ee0def1dbb914d7c2b33136 |
| command | sdlc close --issue 449 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-19T09:17:09-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Checking the remaining open findings (plan events list, branch deletes, atlas), then the new M3 code.

Open findings mostly addressed; BR-23 still has an unleased `gh pr close --delete-branch` path. Running the test suite now.

Vet clean and `landingline` passes under `-race`; the main-package tests are still running. Meanwhile I'm reviewing the M3 bisect and split code.

The main-package run hit Go's 10-minute timeout in `TestLandingLineTwoLandingsIntegrateInTheirOwnTurns`, a test that predates this branch. Running it alone to see whether it hangs.

It passes alone in 21s, so the timeout was most likely the total run time of my broad filter rather than a hang. Running the train/car tests with a longer timeout.

That also timed out, this time with stacks in the train test harness (`trainSetup.merge`). Finding which test hangs.

All ten train end-to-end tests pass (142s total). Running the remaining unit-level train/car tests.

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

The train delivers what the issue asks for. Cars vouch while they wait, a conductor batches them onto one train PR and runs one recorded suite, and a failure is bisected: the culprit is held, the clean prefix lands and the cars behind it are released. Red CI and a red main refuse the train, and a grant departs alone. All of this runs through the pure `Transition` state machine. Six of the seven open Minors are fixed with regression evidence. BR-23 is only partly fixed: when a train PR was opened, closing it still deletes `landing-train/<id>` through `gh pr close --delete-branch`, which is not leased and skips `deleteTrainRef`. The atlas, though, says every train-branch delete goes through that one helper. It is a Minor, so it does not block. Confidence is medium because I could not finish a full `cmd/sdlc` package run in the time I had (details under Test coverage).

**1. Strengths**
- `cmd/sdlc/internal/landingline/trainstate.go:171-190`: when an abort happens during the merge, the train now records merge-unconfirmed (`PhaseUnconfirmed`) instead of abandoned, and the PR is left for the next conductor to reclaim. This closes the uncertain-outcome family (BR-21), and `trainstate_test.go:325` pins it.
- `line.go:240-247`: the departure ledger is read after `head.lock` is taken, inside the line mutex. The schedule is still read outside the mutex (BR-14).
- `bisect.go`: a small, pure bisect with clear bounds that stays sound when results contradict each other. `split()` handles a culprit at the first car (`lo==0` → abandoned, `ClosePR`).
- `landingcar_test.go:424`: the seam contract now globs `landing*.go` instead of checking a hand-kept list (BR-29). `productionCarGates` now takes `ci` as a parameter (BR-22).
- `landingtrain.go:552-563`: `deleteTrainRef` checks the `landing-train/` prefix and a valid tip, runs under the repo lock and uses a lease.

**2. Critical findings:** none.

**3. Important findings:** none.

**4. Minor findings**
- BR-23 is still open: the opened-PR close path (`ghlanding.go:244`, `--delete-branch`) deletes the train branch without a lease or the repo lock, and the atlas line "Every train-branch delete goes through one helper" (`atlas/workflow/sdlc-binary.md`) is therefore false. Suggested fix: drop `--delete-branch` from `CloseTrainPR` and have `closePR()` always call `c.deleteTrainBranch()` after closing.

**5. Test coverage notes**
- `go vet` is clean, and `landingline` passes with `-race`.
- All 10 `TestLandingTrain*`/`TestLandingCar*` end-to-end tests pass (142s).
- The 13 unit-level train/car tests pass with `-race`.
- Two broad-filter runs of the main package (`-run 'Landing|Train|Car|Merge'` and `-run 'Train|Car|Conduct|…'`) hit the 550–600s timeout. The goroutine stacks dumped at the timeout ended in `TestLandingLineTwoLandingsIntegrateInTheirOwnTurns` (a test that predates this branch) and in `trainSetup.merge`. Run on its own, that older test passes in 21s, and the train tests pass as a group. That points to slow total run time, not a hang, but I could not prove it. The implementor should confirm `go test ./cmd/sdlc` finishes under CI's timeout.

**6. Architecture**
- ARCH-DRY: pass, except the delete path in BR-23.
- ARCH-PURE: pass. Transition and Bisect are pure; the conductor executes effects.
- ARCH-PURPOSE: pass. The M3 scope (splits, conflicts, red CI, grants, report) is delivered.
- ARCH-MOCK: pass. The `trainGH` seam, with the glob contract test.
- ARCH-CONSTRAINTS: pass. A bisect costs at most ⌈log₂ n⌉ extra suites, and reintegrations are capped by `MaxReintegrations`.
- ARCH-SECURE: pass. Readiness is cross-checked against the ticket's branch, and deletes are prefix-guarded.
- ARCH-ORDER: pass. The abort goes through `Transition`, and unknown merges are probed.
- ARCH-FUNERAL: pass with a note. Train branches and leftover PRs are swept by the next conductor; the only gap is the unleased close-path delete above.
- ARCH-EVOLVE: pass. Outcome and event records are additive.

**7. Plan revision recommendations:** none. The plan's line 210 now lists `Aborted` in the full event set.

```findings
dispose:
  - id: BR-14
    disposition: addressed
    note: |
      line.go:240-247 reads Events() inside withMutex after head.lock is acquired; only the schedule rule is built outside.
  - id: BR-15
    disposition: addressed
    note: |
      landingqueue.go:37-40 doc and plan line 205 now describe the 30s cached schedule and the in-mutex ledger read.
  - id: BR-21
    disposition: addressed
    note: |
      trainstate.go:181-186 maps Aborted from PhaseMerging to PhaseUnconfirmed with no ClosePR; pinned by trainstate_test.go:325.
  - id: BR-22
    disposition: addressed
    note: |
      productionCarGates(r, ci ghCaller, ...) takes the seam (landingcar.go:90); glob-based TestTrainCodeTakesItsGitHubSeam passes.
  - id: BR-23
    disposition: not-addressed
    note: |
      deleteTrainRef is the leased helper, but the opened-PR close path still deletes via unleased gh pr close --delete-branch (ghlanding.go:244), contradicting the atlas's one-helper claim.
  - id: BR-24
    disposition: addressed
    note: |
      atlas/workflow/sdlc-binary.md conductor section now covers branch cleanup, leftover-branch sweep, and merged-without-a-word resume.
  - id: BR-25
    disposition: addressed
    note: |
      Plan line 210 lists the full event set including Aborted, matching trainstate.go:113-122.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"419ddf89-913b-4340-88b8-f996621aca3f","repository":"github.com/xianxu/ariadne","issue_id":"000449","recipe":"milestone-review","round":8,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/dea1dca8d8f2007b","reviewed_head":"74171bfb3611d7c69ee0def1dbb914d7c2b33136","slot":"ariadne:2","requested_at":"2026-10-19T15:52:57.472418Z"},"phase":"finished","pid":28029,"started_at":"2026-10-19T15:52:57.570684Z","finished_at":"2026-10-19T16:17:08.384877Z","session_id":"2946c7c0-22f2-4b17-a6d8-158e61b68f27","outcome":"returned"}] -->
