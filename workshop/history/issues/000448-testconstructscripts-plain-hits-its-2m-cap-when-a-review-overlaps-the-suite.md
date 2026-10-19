---
id: 000448
status: done
deps: []
github_issue:
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
severity: low
collaboration-mode: collaboration
card_mirror: 'e6657f2c4911af8a3e98a9a333652bce6ad3144a' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-18T19:54:59-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "19ed9822", done: "828e3cbf"}
actual_hours: 0.24
---

# TestConstructScripts/plain hits its 2m cap when a review overlaps the suite

## Problem

`TestConstructScripts/plain` (`scripts/test/scripts_test.go`) runs all of `construct/scripts/test` under one `runnerDeadline` of 2m per environment. Under load the plain subtest hits that cap and fails, unrelated to the change being tested.

Evidence (ariadne:2, #441, 2026-10-18):
- The recorded suite on 208ed53a (`sdlc local-check run --name suite`) failed with `construct/scripts/test did not finish within 2m0s`. The last script reported ok was `bootstrap-transitive.test.sh`. The whole run took 14m7s against a usual 8-12m, because a `sdlc milestone-close` boundary review ran alongside it, which the #442 overlap rule permits. The failed log is kept at `/Users/xianxu/workspace/ariadne/.git/sdlc/test-stats/failed/dd86d2996958c8e19e84dd8b9c49ee61.log`.
- The branch touched nothing under `construct/scripts` or `scripts/test` (only prose in `construct/local/fix/SKILL.md`).
- On the same tree, `go test ./scripts/test/ -run TestConstructScripts -count=1` alone passed in 230.9s for three environments, about 77s each against the 120s cap, so the headroom is only about 1.5x.
- The next two recorded suites (b7c033ab, 389adfdc), run with nothing overlapping, passed.

## Spec

The canonical shell runner owns selection and exclusions. Add a NUL-delimited
`--list` mode that lists selected paths without executing them. The Go harness
uses that selection and runs each script as a named subtest, creating a fresh
unchanged `testgate.HangCeiling` (two minutes) for each command. Preserve process
group cancellation and pipe-drain bounds; preserve the default CI batch runner.
No timeout increases and no changes to the scripts under test.

ARCH-DRY: selection remains in one runner. ARCH-PURE: selection is distinct from
execution. ARCH-PURPOSE: a completed script no longer spends the next script's
hang budget. ARCH-CONSTRAINTS: six sequential scripts across three environments,
with shared-machine load; bound a hung script, not cumulative healthy progress.
ARCH-MOCK/SECURE: private fixture scripts, no external service or credentials.
ARCH-ORDER: no durable state between invocations; cancellation owns one process
group until it exits. ARCH-FUNERAL: fixture files use test-owned temporary dirs;
children are killed on cancellation and output draining stays bounded.

Alternatives: increasing the aggregate cap violates the dispatch and merely
moves the load cliff; implementing a shell watchdog duplicates Go's existing
process-group cancellation. Use the existing Go command seam per selected path.

## Done when

- Each selected script in each environment receives its own unchanged two-minute
  hang bound, preserving cancellation of its process group. A recorded suite
  under the machine's normal concurrent workload passes and records timings.
- Go and CI share the canonical selection and exclusions. Fixture regressions
  cover listing without execution, unusual path names, exclusions, empty input,
  and default runner success/failure behavior.
- `pkg/golint` rejects Go tests executing the known construct batch runner except
  for listing, citing this issue. Tests demonstrate the original call is caught.

## Plan

- [x] Add regressions for canonical listing and the original aggregate call;
  run red before implementation.
- [x] Add listing and per-script Go subtests; retain the existing per-command
  ceiling and process-group cleanup. Run targeted regressions and the catalog.
- [x] Start the recorded full suite; close review may overlap its queue under
  the worker contract. A recorded final-head pass remains required before landing.

## Log

### 2026-10-18
- 2026-10-18: closed — Budget independence regression drives all three real environment loops: nine distinct live deadline contexts and prior cancellation; shared-context mutation fails deterministically. README documents NUL listing without execution. Complete targeted scripts/golint packages pass (53.690s/0.794s), including 18 real script combinations. Corrected-head recorded full suite is queued and must pass before landing; prior queued attempt cancelled before tests. Same 2m ceiling and process-group cleanup. (done-when: contract moved since change-code; judged by the close review); review verdict: SHIP

- Claimed and started in ariadne:3. Hub dispatch: diagnose first, no timeout bumps.
- Read both retained failure logs: the first reports bootstrap and clone passing
  before timing out; the #452 ancestor reports bootstrap, clone, dev-aliases and
  portable passing. The original Problem's first last-success summary was
  incomplete. No one script consistently hangs.
- Unchanged individual plain-environment measurements (seconds): bootstrap
  0.746; clone 8.129; dev-aliases 19.205; portable 0.758; sandbox 0.814; vm-log
  0.042. All passed. These are a fresh baseline, not a loaded-suite guarantee.
- Posted the quick-flow plan to the hub, proceeding unless it objects. A bounded
  independent read-only diagnosis confirmed cumulative scope and recommended
  listing plus per-script execution. The narrow lint protects this known batch
  entrypoint; it does not claim to infer arbitrary shell loops from Go syntax.

### 2026-10-18 — recorded loaded construct-script measurements

- Runtime head: `25bf575191a753a127daf2bd1833cd5b1e27a6ec` (resolved with Git).
  Attempt `71aab835839cdd344d0e88c13262deec` acquired both machine suite slots
  after 1,238s queued. The sharded full suite is still running; its construct
  component has passed all eighteen combinations (252s total).
- Timings below were read from this attempt's live test2json stream, identified
  by its own detached checkout path, before the sharder removes that stream.
  Values are seconds; each individual command retains its two-minute ceiling.

| Script | Plain | No-color CI | Forced-color terminal |
|---|---:|---:|---:|
| bootstrap-transitive | 1.11 | 0.84 | 0.58 |
| clone-data-deps | 2.19 | 2.60 | 1.76 |
| dev-aliases | 87.52 | 72.50 | 63.89 |
| portable-makefile | 1.95 | 1.59 | 1.71 |
| sandbox-sync-set | 4.26 | 2.82 | 5.52 |
| vm-log | 0.28 | 0.16 | 0.28 |
| Whole environment | 97.47 | 80.65 | 73.89 |

The longest leaf took 87.52s versus 19.205s in the initial standalone sample.
These are measured outcomes under the parallel suite workload, not a claim that
all future contention is bounded by this sample. No timeout value changed.

### 2026-10-18 — recorded full suite passed

- Attempt `71aab835839cdd344d0e88c13262deec` passed on the runtime head above:
  1,657 cmd/sdlc tests across twelve shards and seventy other packages, all zero
  failures. Runtime 716.0s (11m56s); packages job 401.0s. Recorded status:
  `sdlc-local/suite`, log SHA256 prefix `e28953535f0a`.
- Log: `.git/sdlc/local-checks/25bf575191a753a127daf2bd1833cd5b1e27a6ec/suite.log`.
  The only commits after that tested runtime head are close/bookkeeping prose.
  Both round-one findings were disposed in round two, which returned SHIP.
  The final landing still uses the normal integration and CI gates.

## Revisions

### 2026-10-18 — dispatch constraint and diagnosed scope

Removed the cap-increase option from Done when: the hub explicitly requires no
timeout bumps. Preserve two minutes for each independent script, and verify
canonical selection plus guard the original aggregate invocation.

### 2026-10-18 — implementation and targeted verification

- Hub approved the plan. Red: the selection fixtures executed the batch instead
  of listing, and both original direct command shapes escaped the lint catalog.
- Added NUL listing in the canonical runner and per-script named Go subtests;
  each creates the same two-minute context and preserves the existing group
  SIGKILL and five-second pipe-drain bound. No tested script changed.
- `go test ./scripts/test ./pkg/golint -count=1 -v -timeout 10m` passed: scripts
  package 154.083s; all eighteen real script/environment combinations passed;
  lint package 0.847s including its scan of 1,195 tracked Go files. The fixture
  first exposed macOS physical-path spelling; expectations now resolve symlinks.
- The narrow lint's red cases catch the original aggregate invocation; positive
  and negative cases cover path-qualified commands, listing, leaf scripts,
  non-test code and a justified suppression. Selection fixtures cover exclusions,
  spaces/newlines, no execution during listing, batch success/failure, unknown
  mode and empty discovery.

### 2026-10-18 — recorded suite queued

- Recorded full-suite attempt `e21e1003c50385972f78fd0f450ea4d1` is queued with
  normal FIFO priority, initially position four behind existing machine work.
  The recorded log is under `.git/sdlc/test-stats/inflight/` until completion.
- All targeted checks pass; the recorded full suite has not passed yet. The
  worker contract permits close review to overlap a queued recorded suite;
  landing still requires its final-head pass, and code fixes requeue it.
- Plan bookkeeping now names dispatch of the recorded suite as the completed
  implementation action; close and landing remain their ordinary gates.

### 2026-10-18 — boundary round one: deterministic budget regression

BR-1 requires a regression for the central lifetime contract across all three
color environments. Inject only command execution into the existing harness;
keep fresh context creation and cancellation in its per-script caller. A fake
command observes each context, verifies the prior context has ended before the
next invocation, and checks independent live contexts for every selected path.
Run the production environment loop with this seam, without waits or shortened
timeouts. Deliberately share a context to demonstrate the regression fails.
BR-2 requires README usage for the sole new flag: non-executing NUL listing
versus default batch execution. The first recorded attempt was cancelled while
queued, before tests ran; requeue the final corrected head.

### 2026-10-18 — round-one findings addressed

- BR-1: `TestConstructScriptBudgetsIndependent` drives the production three-
  environment loop through an injected command seam. It checks nine distinct
  live deadline contexts (selection plus two scripts per environment), cancellation
  before the next command and after the last, and execution in every color policy.
  Mutation red: one shared context produced one budget instead of nine, with
  uncancelled prior budgets in every environment. Restoring the unchanged
  per-command context passed; no fixed sleeps, timeout overrides or clock races.
- BR-2: README describes NUL-delimited non-executing `--list` and default batch
  execution; atlas already explains the per-script flow. Lessons now require
  independent-budget regression proof and helper-flag usage documentation.
- Fresh complete targeted run: `go test ./scripts/test ./pkg/golint -count=1
  -timeout 10m` passed, scripts 53.690s and golint 0.794s. The first full attempt
  ran no tests: its normal queue was cancelled before this code correction.
