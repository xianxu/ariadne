---
id: 000134
status: open
deps: []
github_issue:
created: 2026-06-26
updated: 2026-06-26
estimate_hours:
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

## Plan

- [ ] Audit transcript-source selection for Claude, Codex, and future harness
      extension points.
- [ ] Add regression fixtures for Codex transcript variants and malformed input.
- [ ] Harden commit-window ownership matching around documented subject forms.
- [ ] Design and implement an estimator-source pointer/config surfaced during
      planning and estimate validation.
- [ ] Update `atlas/` and helptext with the resulting contracts.
- [ ] Verify with `go test ./cmd/sdlc/...` and at least one real Codex `sdlc
      actual --issue <N>` run.

## Log

### 2026-06-26

- Filed after Codex dogfooding found that `sdlc actual` did not measure Codex
  sessions until `f62d099`, and that estimate-logic discovery depends on
  operator memory of the brain-local calibration path.
