---
id: 000578
status: working
deps: []
github_issue:
created: 2026-10-23
updated: 2026-10-23
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'cd759daa78c9bc120d9d82b1c3012b86124b4d7b' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-23T17:24:51-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "7f08b9d8", done: "3887c8fb"}
---

# Train repos: pre-close and boarding use the affected check, so the full suite runs once, in the train

## Problem

## Spec

**Use affected boarding in train repos, so the full suite runs once (operator, 10-23).** ariadne#473 (affected
suite) and ariadne#561 (trains board on affected) landed, but slots still run a full recorded suite before close and
boarding, because the worker guidance says so: tl-worker ("a full suite before a close"; "Run full suites through
`sdlc local-check run --name suite`"; "Run the required recorded pre-close suite") and AGENTS §5 ("run the full suite
before a close"). In pair this week, 1 affected boarding vs. 18 trains: the full suite runs twice (pre-close and in the
train) and trains save nothing. Goal: cut machine CPU/SSD per landing so development goes faster at the same cost.
Source: ops `session.side/2026-10-23-04-landing-queue-health.md`.

Change, for repositories where trains are on and `.sdlc/affected.json` exists (pair today; ariadne stays off):
- The pre-close and boarding check is `sdlc local-check run --name affected` on the pushed head; the train's
  integration suite is the full-suite gate. A full selection still records `suite` as today.
- tl-worker and AGENTS §5 say so, conditioned on the repo's landing config; other repos keep the full pre-close suite.
- If `sdlc close` (or any gate) demands a recorded full suite in such a repo, it accepts the affected record instead.
- A failure fix after a train failure still follows the existing rerun rules.

## Done when

- In a train repo with affected.json, close and boarding pass on an affected record alone, and the train's suite is
  the only full run for that landing (test).
- tl-worker's help and AGENTS §5 state the per-repo rule; repos without trains are unchanged (test or doc check).

## Plan

Map (10-23): boarding already accepts an affected record (#473/#561, `carCIGateFacts` stand-in, tested in
`landingcar_affected_test.go`), and `sdlc close` requires no recorded suite. What still sends a full suite before
close in a train repo is guidance (tl-worker help lines 27, 40, 51; AGENTS §5's "run the full suite before a close")
and two refusals that name `--name suite`: fixed-then-ship's code-conflict refusal (`closefixed.go`) and the publish
gate's (`publishgate.go`). An alone or fleet-blocking car lands serially and its landing turn runs the suite: still
one full run.

Design: the repo condition lives in one pure function, not in prose an agent must evaluate (ARCH-DRY).
- `preCloseCheck(trainsOn, affectedDeclared bool) string`: `affected` when main's `.sdlc/landing.json` has trains on
  and main declares `.sdlc/affected.json`, else `suite`. Read from origin/main with `readLandingSchedule` and
  `readAffectedConfig`; an unreadable file degrades to `suite` (ARCH-SECURE, as the car gate does).
- `sdlc local-check run --pre-close`: runs that check (ordinary mode: a full selection runs and records suite) and
  prints which one and why. Exclusive with `--name`/`--all`/`--boarding`.
- The two refusals and close's VERIFIED example name `--pre-close` (fixed-then-ship's `--full-suite` evidence then
  may be an affected record in a train repo, since the train's suite runs before landing).
- tl-worker help and AGENTS.base.md §5 say: run `sdlc local-check run --pre-close` on the pushed head before close;
  in a train repo with affected.json it is the affected check and the train's suite is the landing's full run;
  elsewhere it is the full suite, as today.

Tasks:
- [x] `preCloseCheck` table test (4 combos + unreadable config → suite); `--pre-close` resolution test in a fixture
      repo for pair's shape and ariadne's shape (no landing.json); flag exclusivity test.
- [x] Landing test: a car with only an affected record lands in a two-car train and the counted local-check runner
      runs suite once, on the prefix, never on a car head.
- [x] Refusals and VERIFIED example via `preCloseCheck`; their tests assert affected in a train repo, suite elsewhere.
- [x] tl-worker help.md + rules.tsv + main_test/drain_test phrases; AGENTS.base.md §5 (base layer; regenerate
      AGENTS.md/CLAUDE.md); atlas `sdlc-binary.md` local-check section.
- [x] Targeted tests, recorded suite (ariadne: no trains, so `--pre-close` resolves to suite), close.

ARCH-EVOLVE/FUNERAL: no persisted format, nothing durable created. ARCH-ORDER: holds no state between events; the
resolution is read once per command from main. ARCH-CONSTRAINTS: two `git show` reads per run.

## Revisions

- 2026-10-23 (implementation): the refusals name `sdlc local-check run --pre-close` in every repository rather than
  resolving affected or suite themselves. The flag resolves at run time, so the pure refusal functions take no repo
  state; their tests pin the command.

## Log

### 2026-10-23

- Implemented: `preCloseCheck`/`resolvePreCloseCheck` (`cmd/sdlc/precloseck.go`), `--pre-close` on
  `local-check run` (exclusive with --name/--all/--boarding). The code-conflict refusals and close's VERIFIED
  example name it via `preCloseRunCommand`. The train test now counts saved suite attempts (exactly one, on the
  train), on top of #561's status check. Guidance: tl-worker help/rules/tests, AGENTS(.base).md §5, and the atlas
  `ci-merge-check.md` pre-close paragraph.
- `TestLocalCheckCmdPreClose` and the train tests need the host preflight (boot identity), so they run outside the
  sandbox; inside it they fail with "operation not permitted", as the other local-check CLI tests do.
- Filed #581 (pair:3's answered-grammar report); the TL assigned it to me next.

