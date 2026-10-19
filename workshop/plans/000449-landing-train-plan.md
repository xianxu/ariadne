# Landing Train Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Land every admission-ready PR on an interval (default hourly) as one train: one integration, one full suite, one landing commit. Split on an unexplained failure, hold the culprit, and land the clean prefix.

**Architecture:** The #393 landing line keeps its flock tickets and `head.lock`. A departure schedule now gates admission: a ticket is admitted only once the train is due, or when it holds a fleet-blocking grant. The admitted head becomes the conductor. It boards every waiter that has vouched readiness for its exact branch head. It builds an ephemeral train branch (main, then each car's exact head merged in admission order), opens a train PR, runs one recorded suite and CI on it, and merges that PR. GitHub then marks every car's PR merged, because each car's head is reachable from main. Each car's own `sdlc merge` process sees its PR merged and runs today's resume path: cards, archive, return to rest, branch delete. A one-car train is today's single landing, unchanged.

**Tech Stack:** Go (cmd/sdlc, internal/landingline), git, gh, the existing local-check suite runner.

---

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `Schedule` (interval, max cars), `DepartureDue` | `cmd/sdlc/internal/landingline/train.go` | new |
| `Readiness` (vouched head, PR, alone flag, reason) | `cmd/sdlc/internal/landingline/train.go` | new |
| `Board` (choose the cars from ordered tickets plus readiness) | `cmd/sdlc/internal/landingline/train.go` | new |
| `TrainState`, `TrainEvent`, `Transition` | `cmd/sdlc/internal/landingline/trainstate.go` | new |
| `Bisect` (next prefix to test; culprit from results) | `cmd/sdlc/internal/landingline/bisect.go` | new |
| `Event` kinds `train`, `car` | `cmd/sdlc/internal/landingline/log.go` | modified (additive) |
| `Ticket` (unchanged; readiness lives in a sidecar) | `cmd/sdlc/internal/landingline/order.go` | unchanged |

- **Schedule / DepartureDue(now, lastDeparture, sched, line)** returns due, plus the reason, when one of these holds:
  - the interval is 0 (serial, today's behavior);
  - `now >= lastDeparture + interval`;
  - the head ticket holds a valid grant;
  - the previous train left ready cars behind (a cap overflow or a split remainder), which makes the next departure immediate.

  `lastDeparture` is read from the `train` events in events.jsonl, so the schedule survives process death and needs no daemon (ARCH-PURE: the clock comes in as an argument).
  - **Relationships:** 1 Schedule per repo (config); N departures in events.
  - **DRY:** reuses `Order`/`ValidGrants` for class and grant; adds no second ordering.
- **Readiness**: a car vouches for `{Seq, Branch, Head SHA, PR, Alone, Reason, At}`. `Alone` is set when the car touches local-gates paths (the guard would fail the train PR), when a grant applies, or when the repo's train is off.
  - **DRY:** the car computes readiness with today's predicates in its own process: `runPublishGate` (reviewed anchor unchanged), the `sdlc-local/suite` record on the head (`loadSavedAttempt` and commit statuses), and `decideCIGate` on its PR head.
- **Board(ordered tickets, readiness, sched)**:
  - takes the head plus every following ticket whose readiness is current, not Alone, and live, up to `MaxCars` (default 8);
  - stops at the first Alone car, which rides in the next train by itself;
  - returns cars, left-behind, and why.
- **TrainState machine (ARCH-ORDER)**. States: `Forming`, `Integrating`, `Testing`, `Bisecting`, `Landing`, then the terminal states `Landed`, `Held`, `Abandoned`. Events: `CarsBoarded`, `CarConflict(i)`, `CarGone(i)`, `TipMoved(i)`, `SuitePass(sha)`, `SuiteFail(sha, kind)`, `FlakeRerunPass`, `MainMoved`, `CIGreen`, `CIRed`, `Merged(oid)`, `MergeUnknown`, `ConductorDead`. Effects are values the IO shell executes:
  - `BuildTrain`, `PushTrain`, `OpenTrainPR`, `RunSuite(prefix k)`, `WaitCI`, `MergeTrainPR`;
  - `HoldCar(i, reason)`, `ReleaseCars(rest)`, `CloseTrainPR`, `LogTrain`.

  The table is in the State design section. Sequence tests drive `Transition` directly.
- **Bisect(results map[k]pass/fail, n)**: the cars are the prefixes `train_1..train_n` (`train_0` = main is known good). It returns the next k to test, or `culprit = k` with `lo = k-1` once `lo+1 == hi`. At most ⌈log2 n⌉ extra suites (3 for 8 cars).

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `trainReadiness` sidecar `ready/<seq>` | `cmd/sdlc/internal/landingline/line.go` | new | file under the line mutex (`line.lock`) |
| `conductTrain` | `cmd/sdlc/landingtrain.go` | new | git, gh, local-check runner |
| `buildTrainBranch` | `cmd/sdlc/landingtrain.go` | new | `git merge --no-ff` in a detached worktree |
| `ghLanding.CreateTrainPR` / `CloseTrainPR` | `cmd/sdlc/ghlanding.go` | new methods | `gh pr create/close` |
| `landingConfig` (`.sdlc/landing.json`) | `cmd/sdlc/landingconfig.go` | new | file read |
| `runDurableMerge` (car side: vouch, wait, resume on merged) | `cmd/sdlc/landing.go` | modified | |
| `landingFakeGH` (+ train PRs; PRs merged by reachability) | `cmd/sdlc/landing_test.go` | modified (fake) | |

- **conductTrain** is the thin shell. It loops `Transition`, then executes the effects. Its suite is `runLocalCheck` on the train head (#394's known-flake step included, so a known flake gets its rerun and is never split). CI uses `waitCIGate` on the train PR. The merge is `gh pr merge --merge --match-head-commit`. Main moving reuses today's integration loop (`integrationLimit`, docs-only reuse through `localRecordsOnAncestors`).
- **landingFakeGH** gains state:
  - train PRs;
  - a merge that marks every open PR whose head is an ancestor of the merge commit as MERGED. GitHub documents the state but not which `mergeCommit` such a PR reports, so the fake's `mergeCommit` for a reachability-merged PR is a parameter: null, the car's own head, or the train's merge commit. Every car-resume test runs all three variants (PQ-3).
- **Car resume** (`landing.go:581` today requires `MergeOID` ⊑ main) accepts a MERGED PR when either its `MergeOID` or its `HeadOID` is an ancestor of the fetched main. Both are equally strong evidence that the car's exact reviewed head is on main. A MERGED PR with neither reachable still refuses. The rule lives in one pure resolver, `landedCommit(pr, isAncestor) (oid, ok)`. It returns `MergeOID` when that is a valid OID on main, otherwise `HeadOID` when that is on main, otherwise not-ok. Its result is stored once as the new field `landingPR.LandedOID`. **Every** `MergeOID` consumer on the landing and resume path reads `LandedOID` instead. The full enumeration (`grep -n MergeOID cmd/sdlc/*.go`, non-test):
  - `landing.go:581`, the post-merge reachability check;
  - `landingarchive.go:162`, the archive's PR validity check, which today refuses a null `MergeOID`;
  - `landingarchive.go:634`, the archive's ancestor-of-tip check;
  - `trackercompletion.go:249`, the card's landed commit.

  `ghlanding.go:20/123` stays the decode site of the raw value. The single-landing path goes through the same resolver, so its behavior is unchanged: `MergeOID` is always on main there (ARCH-DRY, one definition). Tests run the fake's three `mergeCommit` variants (null, the car's head, the train's merge commit) end to end through resume, archive and card completion, not only through the gate. A test greps that no non-test file outside ghlanding.go reads `MergeOID` (ARCH-PURPOSE: the class, not the site).
- **Live conformance (ARCH-MOCK):** per the hub (Q7, Revisions), there is no scratch repo. The first pilot train's real evidence is recorded as the conformance check: each car PR's state, `mergeCommit` and `headRefOid` after the train merge. If GitHub's `mergeCommit` turns out to be a variant other than the fake's default, the fake's default changes to match. Resume already accepts all three.

## Operating envelope (ARCH-CONSTRAINTS)

- **Workload:** batch landing, minutes to an hour. N cars per train ≤ MaxCars (8). Suites per train: 1, plus ≤ ⌈log2 N⌉ on failure, plus the #394 flake reruns (targeted).
- **Latency:** a ready PR waits at most `interval` (default 60m) plus one train (suite ~12m plus CI ~3m). Interval 0 restores today's serial landing.
- **Disk:** one detached worktree per train, removed at the end. No new machine-wide suite slots: a train is one suite holder, as today.
- **Overload:** more than MaxCars ready → the overflow departs immediately after. Over the 20m CI budget → `Abandoned` with the train PR closed; the cars return to waiting.

## Lifecycle of created things (ARCH-FUNERAL)

- **The train branch `landing-train/<seq>` and the train PR:** the conductor creates them. The merge deletes the branch (as today). On `Abandoned` the conductor closes the PR and deletes the branch. If the conductor dies, the next conductor's reclaim closes and deletes any train PR whose conductor's head.lock is gone. Count ≤ 1 live per repo.
- **The `ready/<seq>` sidecar** lives in its own directory, never in `tickets/`. `scan` (line.go:261) flock-tests and deletes every unlocked file in `tickets/`, so a sidecar there would vanish on the next poll (PQ-2). The car writes it under the line mutex (`line.lock`), by tmp then rename. `scan` sweeps `ready/<seq>` when `tickets/<seq>` is absent: the ticket's owner died or left, and the car also removes it on exit. ≤ 1 per live ticket. Test: a sidecar survives scans while its ticket is held and is swept after the holder is killed.
- **`train`/`car` events:** appended to events.jsonl and swept by the existing 30-day `KeepEvents`. A few hundred bytes per train.
- **The train worktree:** removed on every exit path, including signals, through the same deferred cleanup `runLocalCheck` uses.

## Persisted formats (ARCH-EVOLVE)

- events.jsonl gains the kinds `train` and `car` and the optional fields `Train`, `Cars`, `Result`, `Splits`, `Culprit`. Additive: older readers ignore unknown kinds (`Delays` reads only `enqueue`/`admit`/`landed`). A test proves an old-shape reader keeps working and `KeepEvents` keeps the new fields.
- The readiness sidecar is a new file with `version: 1`. An unreadable or hand-edited sidecar counts as not ready, never as ready (ARCH-SECURE: typed parse at the boundary; the vouched SHA is re-verified against the remote tip at departure).
- `.sdlc/landing.json` `{train_interval_minutes, train_max_cars}`. Absent → defaults (Q3). Unknown keys preserved, out-of-range values refused with the path named.

## Trust boundaries (ARCH-SECURE)

- **Car readiness** comes from another process. The conductor trusts it only as a claim. It re-reads the car's remote branch tip and refuses a mismatch (`TipMoved`). It never merges a SHA that was not vouched for.
- **Branch names and SHAs** reach git as argv elements, validated through the existing ref and SHA parsers.
- **No credentials:** gh uses its existing auth.

## State design table (tl-worker)

**Actions** a user or automation can take:
- start or repeat `sdlc merge --yes` (board / resume);
- Ctrl-C or kill a car's merge;
- kill the conductor;
- `sdlc landing grant/revoke`;
- `sdlc landing status`;
- push or publish to main concurrently;
- edit `.sdlc/landing.json`.

**Visible states**, shown by `sdlc landing status` and the car's wait line:
- W1 waiting, not ready (why)
- W2 waiting, ready, train at HH:MM
- T1 on train T (integrating/testing/bisecting, car i of n)
- L landed (cleanup running or pending)
- H held (culprit or conflict; reason, log)
- R refused (today's refusals)

| Internal state / failure mode | Visible | Recovery action | Possible loss |
|---|---|---|---|
| Car enqueued, gates not yet passing (no suite record, CI pending, review stale) | W1 + reason | fix and keep waiting (the car re-vouches each poll) | none |
| Car ready, train not due | W2 | wait, or grant for fleet-blocking | latency ≤ interval |
| Car ready but touches local-gates paths | W2 "rides alone" | none; it departs as a one-car train | latency |
| Conductor admitted, boards cars | T1 | — | — |
| Car's remote tip ≠ vouched head at departure (pushed after vouching) | W1 "head moved" | car re-vouches; next train | none |
| Car's merge conflicts with the train | H "conflicts with car j / main" | integrate main manually, re-close, `sdlc merge` (today's conflict recovery) | none |
| Suite passes, CI green, merge confirmed | L for all cars | cars run cleanup themselves | none |
| Suite fails on a known flake only; #394 rerun passes | T1 → L | — | none |
| Suite fails unexplained → bisect finds culprit k | H for car k; L for 1..k-1; W2 (immediate) for k+1..n | culprit fixes its branch; the rest ride the next train | suite time |
| Persistent main failure (baseline fails too) | R for all, "main is red" | fix main | suite time |
| Car process dies before departure | its ticket vanishes; not boarded | rerun `sdlc merge` | none |
| Car process dies after boarding | the train continues (removing it would invalidate the tested tree); L | rerun `sdlc merge` in that slot to finish cleanup (resume path) | cleanup delayed |
| Conductor dies mid-train (before the merge) | cars see no head; the next head reclaims: closes the orphan train PR, deletes its branch | automatic at the next admission | suite time |
| Conductor dies after `gh pr merge` was sent, reply lost (MergeUnknown) | the next head (or a car's resume) reads GitHub: every car whose PR is MERGED → L; else re-form | automatic | none |
| Main moves during the train (docs publish) | T1 | re-integrate (≤3, today's loop); docs-only reuse keeps the suite | CI rerun time |
| Main moves with code during the train (a legacy-path push) | T1 → new suite on the re-integrated head | automatic, capped at 3, then R | suite time |
| CI red on the train PR | R "train CI red", naming the check, and the train PR closed; the cars return to W1 (no CI bisect) | the operator or TL reads the check; the cars re-vouch | latency |
| Grant ticket arrives while the train is forming | the grant waits for the running train (never preempts, as #393), then departs alone immediately | — | ≤ one train |

Covered: every action above, process death on both sides, lost merge reply, main moving, conflicts, unexplained and known-flake failures. Not covered: a split on CI failures (CI failures are rare and not test-attributed; refusing the train is simpler), and cross-repo trains (one line per repo, as #393).

## Milestones

- [ ] M1 — Pure core and readiness:
  - `Schedule`/`DepartureDue`, `Readiness` sidecar, `Board`, `TrainState`/`Transition`, `Bisect`, and the additive event kinds;
  - the car's vouching from today's predicates;
  - `sdlc landing status` shows W1/W2 and the next departure;
  - `.sdlc/landing.json`.

  Interval 0 → behavior identical to today (the existing landing tests pass unchanged). Tests:
  - table tests for `DepartureDue` and `Board`;
  - `Transition` sequence tests over the state table rows;
  - `Bisect` at N and 2N, counting suites ≤ ⌈log2 N⌉ (the HOT-PATH counted test);
  - sidecar parse refusals.
- [ ] M2 — Conductor:
  - `conductTrain`: build the train branch, open the train PR, one suite, CI, merge;
  - cars resume on merged;
  - `landingFakeGH` with merge-by-reachability.

  Done-when test: with N=3 ready cars (fakes, real git), one train runs one integration suite and lands all three. Also tested: a one-car train takes today's path; main moving re-integrates with docs-only reuse; a dead conductor's orphan train PR is reclaimed; a lost merge reply resolves from GitHub state.
- [ ] M3 — Split, grants, conformance, docs:
  - bisect on an unexplained failure: culprit held, prefix landed, remainder departs immediately;
  - the known flake reruns instead of splitting;
  - a grant bypasses the train;
  - conflict hold, and train CI red;
  - the per-train report (events plus status: PRs, suite result, splits);
  - live conformance of merge-by-reachability (Q7);
  - atlas `sdlc-binary.md` landing-line section, `helptext/landing.md`, `helptext/merge.md`, README;
  - pilot plan for the hub.

## Questions for the hub (user-facing tradeoffs)

- **Q1 Landing primitive.** I recommend an ephemeral train PR merged once: GitHub marks every car merged by reachability, CI runs on the exact integrated tree before landing, and main moves in one commit. Alternatives:
  - (b) Push the train straight to main: no train PR, but CI runs only after landing.
  - (c) Rewrite each car's branch to its prefix and merge N PRs: N CI waits, and main passes through untested intermediate states.
- **Q2 Readiness before admission.** #393's PQ-1 said a merge observes nothing before admission. Cars now vouch for their head while waiting, and the conductor re-verifies the tip at departure. This keeps each car's gates in its own process.
- **Q3 Default.** The spec says hourly. I suggest shipping with the file absent meaning 60m, plus a pilot. The hub may prefer 0 (off) until the pilot.
- **Q4** Local-gates-path PRs and grant holders ride alone.
- **Q5** The split remainder departs immediately instead of waiting a full interval.
- **Q6** CI red on a train refuses the train (no CI bisect).
- **Q7** Live conformance needs a scratch GitHub repo, which is outward-facing. Alternatively, the evidence comes from the ariadne pilot's first train, with the fake's rule written from GitHub's docs until then.

## Revisions

### 2026-10-19 — hub answers (delivery 14ac7039)

- Q1: train PR merged once. Q2: cars vouch while waiting; the conductor re-verifies the tip at departure. Q4, Q5, Q6: as proposed (CI bisect is a later follow-up if needed).
- Q3: **the default is 0 (off)**. With `.sdlc/landing.json` absent, or `train_interval_minutes` 0, landing is today's serial line. The pilot sets 60 in pair's `.sdlc/landing.json`. **Correction (PQ-1):** `.sdlc/*` *is* in local-gates-guard's protected list (`.github/workflows/local-gates-guard.yml:69`, base-layer seeded), so the pilot PR, and every later interval change, is a local-gates change. The operator verifies it with the `local-gates-verified` label. The config stays in `.sdlc/` on purpose: the interval decides how often the authoritative suite runs, which is exactly the class of setting the guard exists for.
- Q7: no scratch GitHub repo. The fake models GitHub's documented merged-by-reachability behavior. The first pilot train's real evidence (each car's PR state and `mergeCommit` after the train merge) is recorded as the conformance check, with low-risk cars where possible. M3's conformance task becomes "record the pilot train's evidence" and moves after landing, with the hub dispatching the pilot.
