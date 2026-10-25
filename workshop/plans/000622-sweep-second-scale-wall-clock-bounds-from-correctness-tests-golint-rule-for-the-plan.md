---
collaboration-mode: discussion
---

# Wall-clock bounds sweep and golint rule Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** No correctness test asserts a second-scale wall-clock bound or polls against a short deadline; a golint rule keeps it that way. Closes ariadne#625 with it.

**Architecture:** The existing `test-tight-wallclock` rule (`pkg/golint/rules.go`) widens from sub-second bounds to any bound below `testgate.HangCeiling`, plus polled `time.Now().Add(d)` deadlines. Each flagged test is rewritten to assert its event or outcome, with `HangCeiling` as the only hang guard. Documented contracts keep a reviewed `//lint:allow`. M1 lands the rule with a transitional per-file pending list; M2 empties the list and deletes it.

**Tech Stack:** Go, `go/ast` (golint), `pkg/testgate`.

---

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `tightWallclock` | `pkg/golint/rules.go` | modified |
| `hangBound` | `pkg/golint/rules.go` | new |
| `wallclockPending` | `pkg/golint/rules.go` | new (M1), deleted (M2) |

- **tightWallclock** — flags (1) an upper bound on elapsed time, `time.Since(x)` or a variable assigned from it, and (2) a `time.Now().Add(d)` deadline assigned to a variable that is then polled with `time.Now().After/Before(v)` or `time.Until(v)`, unless the bound is a hang bound. A timestamp fixture built with `time.Now().Add` is never polled, so it is not flagged. Unit tests in `rules_test.go` (AST only, no IO).
  - **DRY rationale:** extends the one rule that already owns this class (10-15 flakes) instead of a sibling rule; reuses `constDuration`.
- **hangBound** — true for an expression naming `HangCeiling`, or a constant duration ≥ `testgate.HangCeiling`. `testgate` is the single source of the ceiling (ARCH-DRY).
- **wallclockPending** — M1 only: the set of files whose findings are still to be swept, so the repo test (`TestTrackedGoFilesPassTheCatalog`) passes while M2 is in progress. M2 removes every entry and the mechanism.
  - **Future extensions:** none; it exists for one milestone (ARCH-FUNERAL: created in M1, deleted in M2).

### Integration points

None new. The tests under change already use real processes, git and the fakes they had; the sweep changes their oracle, not their seams (ARCH-MOCK unchanged).

### Principles

- **ARCH-PURPOSE:** the class is enumerated by the rule itself (35 sites); the plan sweeps all of it, not the 6 named.
- **ARCH-CONSTRAINTS:** the operating envelope is fleet load (load average 20–50, several suites at once). Waits are bounded only by `HangCeiling` (2 min); each package's changed tests are proven with `-count=10`, one run at a time, at least once under `nice -n 19`.
- **ARCH-ORDER:** waits observe events (a barrier file, a recorded timeout, a worker's exit); where a worker can finish before the barrier, the wait fails at once instead of waiting out the ceiling.
- **ARCH-SECURE, ARCH-EVOLVE, ARCH-FUNERAL:** creates nothing durable and changes no persisted format; only test code and a lint rule change. ARCH-PURE: the rule is pure AST code.

## Non-goals

- **`select { case <-time.After(d): t.Fatal(…) }` fail-timeouts** (about 33 sites): the same class, but a different AST shape. Filed as ariadne#626, which reuses this rule's `hangBound` (TL scope call on PQ-1).
- **`context.WithTimeout(…, d)` deadlines** (about 17 sites): these usually bound the code under test, which is a production deadline being tested rather than the test's own wait; `test-shrunk-deadline` already owns the sub-second ones.
- **Lower bounds** (proof that a wait happened), **sub-second deadlines** (`test-shrunk-deadline`) and **fixed sleeps** (`test-fixed-sleep`) belong to their own rules.

## Rule test strategy (adversarial inputs)

The rule is pure AST code, so its inputs are source shapes. Seeded cases: the pre-fix #620 shape (red), a non-constant bound (`tc.timeout+2*time.Second`, red), a polled 30 s deadline in both `After` and `Before` forms (red), an unpolled `time.Now().Add(-time.Hour)` fixture timestamp (green), `HangCeiling` and `2*time.Minute` bounds (green), a deadline reassigned to `HangCeiling` before polling (green), and one name reused across functions with different meanings (`TestTightWallclockScopesNamesPerFunction`, green; it fails when variables are keyed by name). Variables are keyed by the parser's declaring object (`varKey`). An aliased `time` import is covered by `localName`.

## Site groups (from the rule, 35 sites)

- **A — polled deadlines (13):** `landingtrain_attribution_test.go:147` (#625), `localcheck_resume_conformance_test.go:808`, `merge_main_test.go:77`, `queue_interruption_test.go:122`, `landingtrain_test.go:610`, `landingintegrate_test.go:158`, `landingtrain_archive_test.go:163`, `landingqueue_test.go:53`, `publishgate_cancel_test.go:36`, `makebuild_publish_test.go:150,171`, `cmd/weave/internal/release/main_test.go:61`, `cmd/weave/internal/weavefs/runner_setup_test.go:145`, `cmd/tl-lead/doc_operating_test.go:406` (fallback when the test has no deadline). Fix: the deadline becomes `HangCeiling`; fail fast where a worker can exit before the event.
- **B — elapsed bounds proving a kill or cancel (~14):** `ghlanding_test.go:153`, `git_refresh_test.go:66`, `actors_bound_test.go:28`, `status_move_timeout_test.go:105`, `landingattribution_impact_test.go:126,145`, `boundarypush_test.go:169`, `remotebudget_test.go:99`, `integrategate_test.go:267`, `landinggit_process_test.go:164`, `recordsdeadline_test.go:121`, `golist_lifetime_test.go:30`, `queue_suite_status_test.go:144`, `localcheck_resume_retention_test.go:175`, `helpercapability_test.go:164,167`. Fix: assert the outcome (the deadline error, the killed child's exit); a fake that would outlive `HangCeiling` if unkilled, so a broken kill fails by hang guard, not by timing.
- **C — documented contract bounds (~4):** `processgroup/process_unix_test.go:67` (5 s shutdown contract), `judge/cancellation_unix_test.go:98` (`oldGrace`), `testfix/build_unix_test.go:26` (build limit), `localcheck_lease_test.go:226` (`cleanupBudget`), `net_signal_unix_test.go:72` (budget from the test's own deadline). Fix: an injected clock where cheap, else `//lint:allow test-tight-wallclock — <the contract>`.
- **D — absence by observation:** any site in A or B found, on reading, to prove that something did *not* happen within a window. Fix: an event where one exists, else `//lint:allow` naming it.

## Plan

- [ ] M1 — rule, group A and #625
  - [ ] Rule: widened `tightWallclock`, `hangBound`, polled-deadline detection; rule tests including the pre-fix #620 BR-16 shape (`err != nil || time.Since(start) > 8*time.Second`) red and `HangCeiling` bounds green.
  - [ ] `wallclockPending`: the B/C/D files, so the repo lint test passes; group A files are not on it.
  - [ ] Group A edits, #625 first (fail fast when either landing ends before the barrier).
  - [ ] Proof: `pkg/golint` tests; per package touched, the changed tests with `-count=10`, one run at a time; #625's test also under `nice -n 19`.
  - [ ] Mutation: revert #625's fix to a 0.1 s deadline under `nice -n 19` → red; restored → green.
- [ ] M2 — groups B, C, D and enforce
  - [ ] Group B edits; group C allow or injected clock; group D event or allow.
  - [ ] Delete `wallclockPending`; the repo lint test passes with the rule enforcing.
  - [ ] Proof: per package `-count=10`, one at a time; mutation for one B site (break the kill) → hang-guard failure, not a timing failure.
  - [ ] Close #625 with the issue (its registered flake unregisters on close).

## Revisions

### 2026-10-25 — plan-quality round 1

- PQ-1 (Important): `time.After` fail-timeouts and `context.WithTimeout` are outside the rule. TL chose to keep #622 at 35 sites: both are stated as non-goals, and the `time.After` shape is filed as ariadne#626.
- Minors: added Non-goals; added the rule's adversarial-input strategy. Variables are now keyed by declaring object (`varKey`), and a reassignment case plus a cross-function test were added, mutation-checked.

### 2026-10-25 — M1 boundary review round 1

- BR-1 (Important): group A loops now fail fast when their worker ends before the event: the six named sites plus three with the same gap (`makebuild_publish_test.go:150` and `localcheck_resume_conformance_test.go:808` gained a wait channel; `landingtrain_archive_test.go` now watches the conductor as well).
- BR-2 (Important), downstream impact: pair runs this catalog (`internal/repolint`, `pairCatalogFor`), and the widened rule flags 183 pair test sites. The TL filed pair#885 to sweep them and then enable `test-tight-wallclock` in `pairCatalogFor`; until it lands, whoever bumps pair's ariadne pin excludes the rule there (one line). ariadne's own enforcement is unchanged.
- Minors: `hangBound` accepts only `HangCeiling` itself (plus a non-negative constant, or times a positive integer); the inverted `v.Before(time.Now())` / `v.After(time.Now())` polls are matched; rule tests cover `time.Until` and the pending-list skip.

### 2026-10-25 — M1 boundary review round 3: rule-level re-plan (TL-approved)

- The `polled-wait-fail-fast` family repeated a third time (BR-5: `waitVouched` callers omitted a live conductor; the optional variadic argument let callers drop a worker unnoticed). Per the third-delta rule I stopped and re-planned; the TL approved the rule-level fix.
- New pure helper `testfix.AwaitCondition(t, what, cond, workers []chan error)` (`cmd/sdlc/internal/testfix/wait.go`): polls `cond`, bounded only by `HangCeiling`; the workers argument is **required** (nil only with a comment). A worker that ends first fails the wait at once, and its result is put back on the channel, so a deferred join or a later completion wait still receives it. That is why workers are `chan error`, not `<-chan error`: no site needs a nil because a cleanup owns the channel.
- `waitQueued` and `waitVouched` are rebuilt on it with `workers []chan error` required; all 51 callers pass the workers they hold that are still running. The hand-copied poll loops (`publishgate_cancel`, `makebuild_publish` ×2, `queue_interruption` ×2, `landingintegrate`, `merge_main`, `landingtrain`, `landingtrain_archive`, #625's `landingtrain_attribution`, `localcheck_resume_conformance`) are rebuilt on it (ARCH-DRY Minor). The train helpers return `chan error`, and `blockedMergeWorker` gains an `exited` channel.
- Lint Minor: the deadline matcher covers every binding (`:=`, N:N `=`, `var`), polls via `Compare` and `time.Since` as well, judges each poll by the binding it reads in source order, and accepts a bare `HangCeiling` only when the file does not declare it (a local const or variable is not the ceiling). Struct-field deadlines remain a #626 candidate.
