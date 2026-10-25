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
