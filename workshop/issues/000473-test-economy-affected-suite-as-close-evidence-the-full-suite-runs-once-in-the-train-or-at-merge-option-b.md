---
id: 000473
status: codecomplete
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-20
estimate_hours: 2.37
collaboration-mode: collaboration
card_mirror: '24bd4ee3e5c151a3fe6d2b639fe518151ab92b01' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-19T20:26:49-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
actual_hours: 2.13
---

# test-economy: affected-suite as close evidence; the full suite runs once, in the train or at merge (option B')

## Problem

A change pays the full suite twice today: once on its branch (close evidence), and again at landing when main has moved
with code (the integrated suite), or in a train (the train suite). Measured 10-19 (`sdlc test-stats`, 7 days): pair suite
12.5 min median, slot wait 11 min median / 35 min p90, ready → landed 54 min median / 2.3 h p90; ariadne's suite takes both
slots. The machine is suite-bound. Discussion and the operator's decision: ops side doc
`teams/devinfra/tl-status-xian-xu.side/2026-10-19-11-landing-schedule.md` (option B′, approved 10-19).

## Spec

- **Close evidence becomes the affected suite:** the packages the change touches plus every package that depends on them
  (exact from `go list -deps` over the repo; non-Go parts by declared path rules in `.sdlc/local-checks.json`, e.g. Lua tests
  when `*.lua` changed; anything unclassifiable → full suite). A new local-check name (e.g. `affected`) records
  `sdlc-local/affected` on the head; `close` accepts it where it accepts `suite` today. The full suite still works as
  evidence (superset).
- **The full suite runs once, before merging:** in the train (as built, #449), or at merge on the integrated head for repos
  without trains (as today). A car no longer needs a branch full-suite record to board; it needs the affected record +
  review + CI.
- **Safety valves:** changes to build/test infrastructure (Makefiles, go.mod, test helpers declared in config, sdlc config)
  require the full suite on the branch; a per-repo setting can turn affected-mode off.
- Measure: suite slot-minutes per landed change before and after (test-stats).

## Done when

- Fixtures: a change in one leaf package records `affected` covering that package and its dependents only; a go.mod change
  forces the full suite; close accepts `affected`; a train boards a car with `affected` only.
- One week after rollout, slot-minutes per landed change in pair drop (test-stats snapshot in the Log).

## Plan

Durable plan: `workshop/plans/000473-test-economy-affected-suite-as-close-evidence-the-full-suite-runs-once-in-the-train-or-at-merge-option-b-plan.md` (D1–D6 approved by the TL 10-19; env contract fixed).

- [x] M1 — planner and runner (`internal/affected`, `local-check run --name affected`)
- [x] M2 — car readiness accepts `affected`; flake/test-stats generalization; wording, atlas

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: greenfield-go-module   design=0.5 impl=0.32
item: smaller-go-module      design=0.2 impl=0.2
item: smaller-go-module      design=0.1 impl=0.12
item: smaller-go-module      design=0.1 impl=0.16
item: atlas-docs             design=0.05 impl=0.08
item: milestone-review       design=0.0 impl=0.2
item: milestone-review       design=0.0 impl=0.2
design-buffer: 0.15
total: 2.37
```

Items, in order:
1. `internal/affected`: config, globs, graph, planner, `go list` seam.
2. `runAffectedCheck` and the verb dispatch.
3. The car's stand-in in the CI gate.
4. Flake and test-stats generalization.
5. Wording and atlas.
6–7. Reviews at M1 and M2.

Each `impl` is 40% of the v2 range's upper end.

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only.*

## Log




- 2026-10-20: closed — Done-when fixtures: leaf change selects leaf + dependents only (TestPlan Done-when row; contract fixture TestAffectedContractFixture: SDLC_AFFECTED_GO_PKGS=leaf mid tonly, GROUPS=lua); go.mod forces and records suite (TestAffectedFullSelectionRunsTheSuite); close evidence wording names affected; a train car boards on affected only (TestCarBoardsOnAffectedMergeStillNeedsSuite, TestCarVouchesADocsOnlyBranchOnItsAffectedRecord for pair#723). M1 and M2 closed SHIP. Recorded suite queued on a3b1a8cb (overlapping per tl-worker). The 1-week pair measurement is the rollout (TL, after pair counterpart lands).; review verdict: SHIP
- 2026-10-20: closed M2 — car stand-in: TestCarVouchesDocsOnlyCommitOnASuiteRecordedHead (TL shape), TestCarVouchesADocsOnlyBranchOnItsAffectedRecord (pair#723), TestCarBoardsOnAffectedMergeStillNeedsSuite, TestCarWithBrokenAffectedConfigNeedsSuite on a real repo; decideCIGate StandIn rows; D4: TestAffectedKnownFlakeReruns, TestTestRunGuardSpansSuiteAndAffected, TestSubsetPassNeverConfirms, TestAffectedStatsAreTheirOwnRow, TestDeclaredCheckResolvesAffected; mutations fail them; help/golden ok; recorded suite queued on 22460760.; review verdict: SHIP
- 2026-10-20: closed M1 — internal/affected + jsonstrict go test green (FuzzParseGoList re-fuzzed; conformance on real go list incl. an embedded asset and a root package embedding a file in a sub-package dir, owners = both); contract fixture TestAffectedContractFixture (SDLC_AFFECTED_GO_PKGS=leaf mid tonly / GROUPS=lua; status 3 packages, groups lua); full→suite for go.mod / unclassified / unloadable graph; docs-only records without running; no-change refuses; --all split tested; targeted cmd/sdlc run ok in 1205s before the review fixes, Affected tests ok after.; review verdict: SHIP
### 2026-10-19

- Exploration: close has no mechanical suite gate (free-text `--verified`); the record is enforced by the CI gate at merge and car vouch. Checks run a fixed argv. `go list -test` over pair: 1.5 s.
- M1 build: Tasks 1–3 (`internal/affected`) by a subagent from the plan. Its interpretations, all kept:
  - `go` roots are syntax-checked at parse time.
  - `GOWORK=off` on both `go list` calls.
  - The module root comes from `go list -m`, which avoids symlink skew.
  - A Full Reason names the first trigger.
  - A trailing `**` matches its own directory.
  
  Fuzzed for 20 s each. Tasks 4–5 were done in the main session: the `affected` name is reserved in local-checks.json,
  `readAffectedConfig` sits beside `readLocalChecks`, and `runAffectedCheck` plus the verb dispatch are in. The
  contract fixture is green (fake consumer `make test-affected` writing both variables), as are full→suite (go.mod,
  unclassified, graph unloadable) and docs-only. `localRunRecord.Scope` carries the selection into the status
  description; a republished saved pass doesn't carry it.
- M1 review round 1 (FIX-THEN-SHIP):
  - **BR-1:** ownership is now exact: own dir, `testdata/`, or `go list` embed files. The ancestor-dir claim made a
    root package swallow every path. The fixture module embeds an asset, and the conformance test pins it.
  - **BR-2:** an empty change set refuses (`TestAffectedNoChangeRefuses`).
  - **BR-3:** README paragraph added. A side-quest corrects #479's stale "fully qualified" flake-list wording there.
  - **Minors:** `--all` with `--name affected` refuses again. Duplicate JSON keys refuse via the new
    `internal/jsonstrict`; the three older copies (observe, fleet, teststats_store) move onto it in #489 (filed, unclaimed), ARCH-DRY.
    D4 is inert until M2: the affected check carries `KnownFlakes`, but the flake step and success guard still key on
    `suite`. A docs-only record saves no attempt, so it can't be republished; it is cheap to re-run.
- M1 review round 2:
  - **BR-7**, the same class as BR-1: ownership is a union (`Graph.Owners`; `Embeds` maps a file to every embedder).
    The fixture module's root package now embeds `other/data.txt`, so the real `go list` pins a file with two owners.
  - **BR-4:** `splitAffectedName` is pure and tested.
  - The plan's Revisions record the ownership rule.
- M2:
  - **Task 6:** the car's stand-in. Tests cover the TL's shape and pair#723's, and a mutation (dropping the stand-in)
    fails both stand-in tests.
  - **Task 7:** D4 made stricter (plan Revisions): the guard spans both checks, and a subset pass never confirms.
    Both mutations are caught.
  - **Task 8:** wording in close's VERIFIED example, the landing and local-check help, AGENTS.base.md (base layer: it
    propagates downstream) and the atlas. The publish gate's conflict-resolution instruction keeps the full suite on
    purpose, because the resolution changes main's code too.
- M2 closed SHIP (review rounds 4–5): BR-9 and BR-10 fixed. Its advisory Minor, an empty suite row in test-stats for
  an affected-only repo, is fixed in the same round: `testStatsRows` keeps a row only when its window holds runs.
- Close review: SHIP, with 3 advisory findings, all fixed in the same round:
  - Docs are judged after ownership, so an owned or embedded doc selects its package (`TestPlan` rows and
    `TestAffectedOwnedDocRunsItsPackage`).
  - The plan's merge-base claim is revised.
  - Duplicate keys that differ only in case refuse.

  These change code after the close, so the close is re-run with `--fixed-then-ship` after the suite.

