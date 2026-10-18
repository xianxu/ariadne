---
id: 000407
status: working
deps: []
github_issue:
created: 2026-10-15
updated: 2026-10-18
estimate_hours:
card_mirror: 'b4f74ca3aac983e1a89d5337da25aca03516793b' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-18T00:13:46-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "3e0be510", done: "eb2712c8"}
---

# A curated Go lint catalog: golangci-lint + shared ruleguard rules for known mistakes (per-test builds, sleeps, wall-clock asserts, temp dirs, strict wire decode)

## Problem

Recurring mechanical mistakes are caught by humans after they cost something: 11 pair tests `go build` the binary
per run (pair#602), sleeps and wall-clock assertions behind most of 10-15's flakes, temp dirs outside TMPDIR (parley
#342, ariadne#342), strict decoding of cross-version wire messages (pair#594). Neither pair nor ariadne runs a Go
linter; the only lint is ariadne's hand-written `pkg/shelllint`, and lessons files keep rules as unenforced prose.
Operator (2026-10-15): "we need to curate a list of 'stupid mistakes'". Part of ops project test-economy.

## Spec

- **Runner:** golangci-lint (pinned version) with one config, run by `make lint` locally and at landing; fast enough
  to never skip.
- **Off-the-shelf linters** where they fit: usetesting, thelper, paralleltest/tparallel, forbidigo (e.g.
  `time.Sleep` in tests), noctx, bodyclose, errorlint, nolintlint.
- **The curated list:** a ruleguard rules file owned by ariadne and distributed by weave to Go repos (repo-specific
  rules beside it). Seed rules: `exec.Command("go","build",...)` in `_test.go` (pair#602); `time.Sleep` / fixed
  waits in tests; wall-clock duration assertions in correctness tests; `os.MkdirTemp`/`mktemp` without TMPDIR;
  strict JSON decode of peer/wire messages.
- **Rule hygiene:** every rule's message cites its incident; every rule has flag/no-flag test cases; suppressions
  require `//nolint:<rule> // <reason>` (nolintlint); a rule that never fires in a quarter is reviewed for removal.
- **Process:** a fix for a mechanical mistake class includes its rule; tl-lead/tl-worker say so (phrase-guarded).

## Done when

- `make lint` runs golangci-lint with the shared rules in ariadne and pair; each seed rule has tests; the existing
  violations are either fixed or carry reasoned suppressions with follow-up issues; landing runs lint.

## Plan

Durable plan: `workshop/plans/000407-a-curated-go-lint-catalog-golangci-lint-shared-ruleguard-rules-for-known-mistakes-per-test-builds-sleeps-wall-clock-asserts-temp-dirs-strict-wire-decode-plan.md` (engine, seed rules, sweep, follow-ups; three
deviations from the Spec are listed there for approval).

- [x] Tasks 1–6 of the durable plan

## Log

### 2026-10-15

### 2026-10-18

- Claimed by ariadne:2; plan drafted and sent to the hub before code.
- Hub approved the plan with all three deviations (stdlib engine in the suite, pair via the TL, wire-decode deferred).
- Implemented `pkg/golint` (Check with suppression hygiene; four rules; `TestTrackedGoFilesPassTheCatalog`: 1,139
  tracked Go files, about 0.7 s). Initial scan: 20 findings, all now fixed or suppressed with a reason (see the plan's
  Revisions). Three deadline tests converted to `testfix.DeadlineOnCue`, and they pass with -count=3.
- Follow-ups filed, unclaimed: #438 one build per test binary, #439 golangci-lint (operator: install and gates), #440
  wire-message annotation plus a strict-decode rule. pair adoption goes to the hub.
- The process line lives once in the shared planning block (`pkg/tlrules/planning.md`, MISTAKE CLASSES), so
  tl-worker and tl-lead both carry it, with a phrase guard in rules.tsv.
