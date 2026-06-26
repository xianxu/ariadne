---
id: 000134
status: working
deps: []
github_issue:
created: 2026-06-26
updated: 2026-06-26
estimate_hours: 3.9
started: 2026-06-26T15:57:18-07:00
---

# Make actuals and estimates agent-robust

## Problem

`sdlc actual` and `sdlc`'s estimate guidance were originally dogfooded mostly
through Claude Code. That left two agent-portability gaps:

- actual measurement assumed Claude's transcript location/shape and a narrow
  issue-subject convention, so Codex work could look like "no measurable
  activity" even when the session transcript and commits existed.
- estimation logic is split correctly between shared method and repo-local/user
  calibration, but discovery is implicit. The shared `estimate-logic-v2` lives in
  brain, while the current operational grammar is in `cmd/sdlc/helptext/estimate.md`;
  an agent can satisfy the block syntax without realizing it should read the
  calibrated local estimator.

Commit `f62d099` fixed the immediate Codex transcript parser and `<area>: #N`
commit-window failure. This issue is for making that support robust instead of a
one-agent patch.

## Spec

- `sdlc actual` must treat transcript discovery as a harness abstraction, not a
  Claude-only path convention.
- Codex support must survive realistic transcript variants:
  - date-sharded `~/.codex/sessions/YYYY/MM/DD/*.jsonl` files,
  - `session_meta.payload.cwd` source selection,
  - `response_item` user/assistant messages,
  - tool/function calls whose payload text can mention issues,
  - missing or malformed records without failing the whole measurement.
- The commit-window ownership matcher must accept the documented commit
  convention (`<area>: #N ...`) while rejecting loose references such as
  `docs: mention #N`.
- Estimator discovery must make the shared-vs-local split explicit:
  - shared method/grammar remains single-sourced in `sdlc`,
  - repo/user calibration lives in the appropriate local/brain artifact,
  - `start-plan`, `change-code`, or a dedicated estimate surface points agents at
    the exact estimator/calibration source they should read.
- When the estimator source is missing, inaccessible, or stale, the tool should
  fail loudly or print a next-action, not silently let agents rely on memory.

## Done when

- `sdlc actual` has fixture coverage for Claude and Codex transcript sources,
  including malformed/irrelevant Codex sessions.
- A real Codex-authored issue can be measured without hand-passing transcript
  paths.
- Commit-window tests cover both canonical `#N ...` and `<area>: #N ...` subjects.
- Estimate guidance names the repo-local calibration source and the shared
  method source in one discoverable command output.
- Atlas/helptext documents the harness abstraction and estimator-source contract.

## Estimate

Derived against `estimate-logic-v2` (brain `…/velocity/estimate-logic-v2.md`),
low-end design per the Step-3 spec-quality discount (the durable plan is fully
specced). Detail: `workshop/plans/000134-make-actuals-and-estimates-agent-robust-plan.md`.

```estimate
model: estimate-logic-v2
familiarity: 1.0
item: smaller-go-module     design=0.2 impl=0.5
item: smaller-go-module     design=0.0 impl=0.3
item: atlas-docs            design=0.0 impl=0.2
item: milestone-review      design=0.0 impl=0.4
item: greenfield-go-module  design=0.4 impl=0.7
item: smaller-go-module     design=0.1 impl=0.3
item: atlas-docs            design=0.0 impl=0.2
item: milestone-review      design=0.0 impl=0.4
design-buffer: 0.30
total: 3.9
```

## Plan

Full detail (TDD steps, exact paths, code) in
`workshop/plans/000134-make-actuals-and-estimates-agent-robust-plan.md`.

- [x] **M1 — Measurement robustness** — extract `internal/transcripts` (Harness
      registry + pure `Select`), move Claude/Codex selection behind it, add Codex
      robustness fixtures (malformed / no-`session_meta` / empty), pin the
      commit-window matcher (`#N` + `<area>: #N` accept, `docs: mention #N`
      reject), rewire `actual.go`, document the harness contract in atlas/helptext.
      Closes via `sdlc milestone-close --issue 134 --milestone M1`.
- [ ] **M2 — Estimator-source pointer** — `sdlc estimate-source` pull command +
      `start-plan`/`change-code` pushes naming the shared method (sdlc grammar +
      `vocab.Models()`) and the repo-local calibration (`<brain>/…/velocity/<model>.md`,
      `$WF_ESTIMATOR_SRC` override); DRY the brain path out of `close.go`; loud-fail
      on missing source in the pull, warn-and-continue in the gates; document the
      estimator-source contract. Closes via the full-issue `sdlc close`.

## Log

### 2026-06-26

- Filed after Codex dogfooding found that `sdlc actual` did not measure Codex
  sessions until `f62d099`, and that estimate-logic discovery depends on
  operator memory of the brain-local calibration path.
- Planned (durable plan + `start-plan` arch lens). Scope split into two review
  boundaries: M1 measurement-robustness (transcript-harness abstraction), M2
  estimator-source pointer. Confirmed `f62d099` already made Codex *work*; this
  issue makes it robust + discoverable (ARCH-PURPOSE). `sdlc claim` flipped status
  locally but the origin push failed (SSH remote outside the sandbox network
  allowlist) — claim committed locally, not yet broadcast.
- `sdlc change-code`: estimate reconciled (Σdesign 0.7×1.30 + Σimpl 3.0 = 3.91 ≈
  3.9); plan-quality judge **INFO**, estimate-quality judge **INFO** (both
  non-blocking). Branch `000134-…` created in-place. Folded the judge's two
  structural refinements into the plan: (1) ARCH-DRY — `estimate.VelocityPath` is
  the single-source brain-path builder in the lower layer (both `close.go` and the
  new `estimate-source` call it; no package-main mirror); (2) ARCH-PURE — split a
  pure `codexCWDFromBytes([]byte)` tested directly on bytes from the `os.ReadFile`
  seam. Done-when #2 (Codex-real) will cite `f62d099`'s parley.nvim #144 → 1.17h
  evidence at close.
- Concurrency note: #129 (default judges to current agent) and #130 (model
  lifecycle in cue) are in-flight on neighbor surfaces; this issue only *reads*
  `judge.ArchitectureBlock`/`estimate.Models()` and makes a one-line `close.go`
  edit, so collision risk is low.
