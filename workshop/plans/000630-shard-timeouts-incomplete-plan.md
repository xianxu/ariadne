---
collaboration-mode: discussion
---

# Incomplete (timeout) suites: classify, record the load, retry the unfinished packages once

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution
> approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or
> superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A suite run whose shards hit the Go test alarm, with no other kind of failure, is recorded as
`incomplete-timeout` with the host's load. It is retried once on the unfinished packages plus every package that had a
named failure, and never triggers a main-baseline diagnosis or a train split. A real failure still diagnoses as today.

**Architecture:** The shard runner gains an explicit alarm marker (it alone sees the full output). A pure classifier
over the parsed log facts decides "incomplete" and the retry's package selection. The attempt store records the
outcome under a new attempt version, which older readers already degrade to a per-head refusal, and keeps the
load in a separate evidence record. The retry reuses the `local-check resume` composite (#561): the incomplete
attempt is the parent and a package-subset rerun is the child, so posting green goes through the composite proof
`guardSuitePass` already verifies. Callers learn about a final incomplete run through a typed error, the shape
`suiteEnvironmentError` already gives the merge gate and the train.

**Tech Stack:** Go (cmd/sdlc), Python (scripts/test-shard.py), golang.org/x/sys/unix for the load average.

---

## Decisions (TL, 2026-10-25)

- **Rule:** a run is incomplete when at least one shard (or the packages job) hit the alarm, and every process that
  failed hit it. Anything else (a build or setup failure, a process that failed without the alarm, an sdlc
  check-level timeout, an unparseable log) is classified exactly as today.
- **Retry scope:** only the packages that did not finish, plus every package that had a named failure in the
  incomplete run (the operator's Done when: "the retry runs only the packages that didn't finish"; the TL added the
  failed packages). Test-level granularity is a non-goal. `cmd/sdlc` is one package across eight shards, so in
  practice its retry runs it whole while the other 79 packages that passed are skipped.
- **A passing retry posts green.** The incomplete attempt never counts as unexplained. Its named failures stay
  listed in the flake view as incomplete-run failures, so recurring ones still surface for registration.
- **The retry is incomplete again:** stop and report it with the load: no third run, no main baseline, no split.
- **The retry fails for real:** classify (known flakes included) and diagnose against main as today.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `SuiteProcess` (`SuiteLogFacts.Processes`: `Label`, `Exit`, `Alarmed`, `Running`) and the `shardAlarm` marker regex | `cmd/sdlc/teststats_parse.go` | modified |
| `incompleteEligible` / `goIdentities` / `validLoadAverage` | `cmd/sdlc/suiteincomplete.go` | new |
| `SuiteAttempt.LoadAverage` (v5 only) | `cmd/sdlc/teststats.go`, `teststats_store.go` | modified |
| `incompleteRun` / `classifyIncomplete` | `cmd/sdlc/suiteincomplete.go` | new |
| `flakes.IncompleteTimeout` kind | `cmd/sdlc/internal/flakes/classify.go` | new |
| incomplete-parent branch of `resolveResumeProof` | `cmd/sdlc/localcheck_resume_evidence.go` | modified |

- **`SuiteLogFacts.Processes` (`SuiteProcess`):** parsed from the runner's marker line (`shardAlarm`:
  `--- shard3 hit its 30m0s test alarm in <pkg>; running: TestA, TestB`) and the existing `shardN exit C S.s K
  failed` summaries. Each process records its label, exit code, the packages whose alarm stopped it (`Alarmed`) and
  the tests named as running (`Running`). The parse stays a pure
  function of the log text (ARCH-PURE), and old logs without the marker never classify as incomplete.
  - **DRY rationale:** extends the one log parser every consumer already reads (`parseSuiteLogFacts`), rather than a
    second scan in the gate.
- **`classifyIncomplete(facts) (incompleteRun, bool)`:** applies the rule above. It returns the retry selection: the
  packages with a named failure, packages that failed without a named test, and `cmd/sdlc` when the never-ran list is
  non-empty. Sorted, de-duplicated, never empty when ok.
  - **Relationships:** 1:1 with a failed attempt's facts; the selection becomes the resume child's `GoPkgs`.
  - **Future extensions:** test-level selection (a non-goal now) would widen the return type, not the rule.
- **`flakes.IncompleteTimeout` (`incomplete-timeout`):** `Lands() == false`, so it never lands a change on its own. It
  is explainable for the guard's purpose ("never unexplained"), and it is a resumable parent kind.
- **`resolveResumeProof`, incomplete parent:** today's parent filter refuses a parent whose facts are unattributed.
  The new branch accepts a parent whose outcome is `incomplete-timeout` and whose re-parsed facts classify as
  incomplete. Its selection is `classifyIncomplete(parentFacts).Packages`, and it is eligible without `NeedRerun`. The
  child must pass for the composite to pass. Every link, receipt, and digest check is reused unchanged.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| shard alarm marker | `scripts/test-shard.py` | modified | go test2json output |
| `hostLoadAverage` | `cmd/sdlc/hostload_{darwin,linux,other}.go` | new | `vm.loadavg` sysctl / `/proc/loadavg` / an error elsewhere |
| `suiteIncompleteError` | `cmd/sdlc/suiteincomplete.go` | new | merge gate, train conductor |
| automatic retry | `cmd/sdlc/localcheck_run.go` | modified | `resumeSaved` |

- **Alarm marker:** `report()` sees each process's full output. A `panic: test timed out after D` line in a process's
  output (test2json puts it on the package key) prints `--- <label> hit its <D> test alarm; running: <names>`, with
  the names taken from the `running tests:` list. It is additive: today's parser ignores an unknown `---` line.
  `scripts/test-shard.test.py` covers it.
- **`hostLoadAverage() ([3]float64, error)`:** darwin uses `unix.SysctlRaw("vm.loadavg")` (`struct loadavg`: three
  `fixpt_t` values and `fscale`); linux parses `/proc/loadavg`. It is injected into `localRunDeps` as
  `loadAverage func() ([3]float64, error)`, and tests set a fixed value.
- **Incomplete evidence record:** `{version 1, id, sha, check, load [1,5,15], alarmed [labels], running [names],
  selection [pkgs]}`. It is written beside the attempt, in its own directory, because the attempt store's readers
  refuse any unknown key (ARCH-EVOLVE, below). It is decoded with `jsonstrict.DecodeObject`, so unknown members
  survive a rewrite, and a missing or invalid record only drops the load from the report.
- **`suiteIncompleteError{Check, Head, LogPath, Load, Alarmed}`:** returned when the run (after its one retry) is
  still incomplete. `diagnoseSuiteFailure` already returns early on a non-`localCheckFailure`, so no baseline runs.
  The train's `c.run` error path already reports "could not run" and neither splits nor runs `mainIsRed`. Its message
  names the load, the alarmed shards, and the rerun command.
- **Automatic retry:** in `runLocalCheck`, after the attempt is saved with outcome `incomplete-timeout`, a recorded
  suite or affected run that has not yet retried calls `resumeSaved` with the incomplete parent. It runs a fresh
  queue ticket, which is its normal turn behind current waiters, as `TestSuiteRerunQueuesForTheSlot` already shows. A
  main-baseline run (`SkipRecord`) never retries and never records: it returns the typed error, so a baseline that is
  itself incomplete is reported as unverified, not as main-red.

### ARCH notes

- **ARCH-EVOLVE:** an attempt with the new outcome is written as **version 5**. Older readers fail
  `decodeSuiteAttempt` on the version. Its keys are all in their allow-list, so `decodeUnsupportedProof` records it
  as an unsupported proof scoped to that head. An old binary refuses passes on that head only, which is safe: it can
  never read the attempt as a pass. Version 5 carries exactly v2's fields; only the outcome value is new. The load
  sits in the separate record because a new attempt key would make old readers warn on the whole store, and the
  guard then refuses every head. Tests: an old-reader fixture (a v4-max reader) sees a v5 incomplete attempt as
  unsupported for that subject only, and a v5 round-trip.
- **ARCH-ORDER:** the run is a short sequence: run → incomplete (saved) → retry → {passed, failed, incomplete}. At
  most one retry, enforced by a flag in `localRunDeps` that the retry sets and the resume path never clears.
  Interruptions:
  - **Process death between the save and the retry:** leaves an incomplete attempt and no child. The next
    `sdlc merge` or `local-check run` sees the head unpassed and runs fresh, the safe direction. `local-check resume`
    can also retry it by hand.
  - **Cancellation during the retry:** handled by `resumeSaved`'s existing reconciliation receipts.
  The event most likely to be mishandled is a retry whose own log is incomplete. It is classified by the same rule
  and stops there.
- **ARCH-FUNERAL:** creates nothing durable beyond the attempt itself: `load_average` rides on the v5 attempt, which
  the store's existing 30-day attempt pruning removes (see Revisions; the separate record was dropped).
- **ARCH-CONSTRAINTS:** the batch path. One full suite is about 20 to 30 min. The retry costs only the unfinished
  and failed packages (in the #616 case, `cmd/sdlc` alone and not the 79 other packages: about 15 to 20 min, against
  a full suite plus a main baseline today). It is bounded to one retry.
- **ARCH-SECURE:** both the log and the evidence record are inputs from earlier processes. The classification
  re-parses the retained log, whose digest the attempt binds, rather than trusting the evidence record's selection.
  The record carries only the load and report fields.
- **ARCH-MOCK:** `go test` and test2json are exercised through the real shard runner in `scripts/test-shard.test.py`
  fixtures. The load seam has fixed values in tests.

## M1: classify, record, report, skip diagnosis (no retry yet)

### Task 1: the alarm marker in the shard runner
**Files:** Modify `scripts/test-shard.py` (`report`). Test: `scripts/test-shard.test.py`.
- [ ] Failing test: a fake shard JSON stream with `panic: test timed out after 30m0s` and a `running tests:` block
  prints `--- shardN hit its 30m0s test alarm; running: TestA, TestB`; a stream without the panic prints no marker.
- [ ] Implement in `report()`; run `python3 scripts/test-shard.test.py`.

### Task 2: parse the marker and the per-process summary
**Files:** Modify `cmd/sdlc/teststats_parse.go`. Test: `cmd/sdlc/teststats_parse_test.go`.
- [ ] Failing test `TestParseSuiteLogFactsAlarms`: the #616 log shape, with markers added, gives 9 processes, 8
  alarmed, `packages` clean. An old log without markers gives zero alarms.
- [ ] Add `Processes []SuiteProcess{Label, Exit, Alarmed}` and `Alarmed []string` to `SuiteLogFacts`.

### Task 3: the classifier and the kind
**Files:** Create `cmd/sdlc/suiteincomplete.go`; modify `cmd/sdlc/internal/flakes/classify.go`. Tests:
`cmd/sdlc/suiteincomplete_test.go`, `internal/flakes/flakes_test.go`.
- [ ] Table test `TestClassifyIncomplete`:
  - every failed process alarmed → incomplete, with the selection;
  - one failed shard not alarmed → not incomplete;
  - a build failure → not;
  - no alarm → not;
  - a failure in the packages job alarmed → its package is selected;
  - a never-ran list → `cmd/sdlc` is selected.
- [ ] Add `IncompleteTimeout` with `Lands()==false`. `Explainable` is unchanged for the other kinds.

### Task 4: record it
**Files:** Modify `cmd/sdlc/localcheck_run.go`, `cmd/sdlc/teststats_store.go`; create the `hostload_*.go` files.
Tests: `teststats_store_test.go`, `localcheck_typed_oldreader_test.go`, `localcheck_run_test.go`.
- [ ] Failing test `TestRunLocalCheckRecordsIncompleteTimeout`, a fake suite whose log has the markers. It checks:
  - the outcome is `incomplete-timeout`, at v5;
  - the evidence record holds the injected load;
  - `knownFlakeStep` is not called;
  - stderr names the load and the alarmed shards.
- [ ] v5 decode and round-trip test; old-reader test: a reader capped at v4 reports it unsupported for that head
  only.
- [ ] Implement the classification before `knownFlakeStep`, the v5 write, the evidence record, its pruning, and the
  `hostLoadAverage` seam.

### Task 5: no diagnosis, no split, guard and flake view
**Files:** Modify `localcheck_run.go` (typed error), `localcheck_republish.go` (guard), `flake_cmd.go` (view). Tests:
`suiteenvironment_test.go`-style `TestSuiteIncompleteBaselineIsNotRun`, `landingtrain_test.go`
`TestLandingTrainIncompleteSuiteDoesNotSplit`, `localcheck_republish_test.go`, `flake_cmd_test.go`.
- [ ] Failing tests:
  - the merge gate returns the typed error unchanged, and no baseline `runLocalCheck` runs;
  - the train reports "could not run", with no split and no `mainIsRed`;
  - a later pass on the same head is not refused because of an incomplete attempt;
  - the flake view lists its named failures under "incomplete-run failures".
- [ ] A real failure in a non-alarmed shard still diagnoses: the existing `TestTypedSuiteBaselineUsesSharedFacts`
  stays green, and a new case covers a mixed run.
- [ ] Atlas: `atlas/workflow/ci-merge-check.md` (the outcome and its consumers).
- [ ] `sdlc milestone-close --issue 630 --milestone M1`.

## M2: the one retry, on the unfinished and failed packages

### Task 6: an incomplete parent is resumable with its own selection
**Files:** Modify `localcheck_resume_evidence.go`, `localcheck_resume.go`. Tests: `localcheck_resume_evidence_test.go`
(pure table), `localcheck_resume_test.go`.
- [ ] Failing pure tests:
  - an `incomplete-timeout` parent whose log still classifies incomplete is eligible, with
    `selection == classifyIncomplete(...).Packages`;
  - a tampered log digest is refused;
  - an incomplete parent's child pass makes the composite pass;
  - a child failure is classified by the normal rules.
- [ ] `resumeSaved` uses that selection as `c.Rerun`'s `GoPkgs` for a suite parent as well as an affected one.

### Task 7: the automatic retry
**Files:** Modify `localcheck_run.go`. Tests: `localcheck_run_test.go`, `localcheck_reentry_test.go`.
- [ ] Failing tests:
  - the operator's test: a timed-out attempt with some packages passed retries only the unfinished and failed ones
    (the rerun command's `SDLC_AFFECTED_GO_PKGS` names exactly them);
  - a retry that passes posts green;
  - a retry that is incomplete again returns `suiteIncompleteError` with no third run;
  - a retry with a real failure is classified and diagnosed as today;
  - a main-baseline run never retries;
  - the retry takes a fresh queue ticket.
- [ ] Implement: after the incomplete save, call `resumeSaved` once; `localRunDeps.incompleteRetried` bounds it.
- [ ] Atlas update; full `sdlc local-check run --pre-close` through the queue; `sdlc close --issue 630`.

## Verification

Targeted tests per task; `python3 scripts/test-shard.test.py`; the pre-close suite on the pushed head. A manual
check: run a fixture repo whose suite command sleeps past a 1-minute alarm in one shard, and watch one retry of that
package only, then the typed incomplete report with the load.

## Revisions

- 2026-10-25, during M1: **the load is an additive attempt field, not a separate record.** The plan said a new
  attempt key would make older readers refuse the whole store. That was wrong: `canonicalSuiteKey` accepts any
  unknown root key that isn't a case-alias of a known one (atlas: "Attempt compatibility (#460)"). An older reader
  refuses the new **outcome value**, which is why version 5 stays. With version 5, the `load_average` key is
  tolerated by the old reader's unsupported-proof path, and the old-reader test covers it. So `load_average` sits on
  the v5 attempt, valid there alone. The `incomplete/<id>.json` record, its writer and reader, and its pruning are
  dropped (ARCH-DRY: one record, one prune). The retry's selection is re-derived from the retained log in any case
  (ARCH-SECURE).
- 2026-10-25, during M1: the shard runner's marker names the package (`--- shard3 hit its 30m0s test alarm in <pkg>;
  running: …`) and prints once per alarmed package. A shard with no named failure otherwise gives no package to
  retry, and the packages job can have several binaries alarm at once.
- 2026-10-25, plan-quality round 1 (PQ-1, PQ-2, two Minors). **M2's mechanics, named.** The resume composite is the
  frame (link, owner fence, receipts, digest checks, retention, `guardSuitePass`'s composite verification). Its
  child is a known-flake rerun today, so the incomplete parent needs these new, named pieces:
  1. **Link (PQ-1).** A selection-bearing suite link is version 5, like an affected one:
     `resumeLinkVersion(check, selected bool)` returns 5 whenever a selection is fenced, else 3 for suite. An older
     sdlc reads a v5 suite link as unsupported proof of that head alone, as it already does for affected
     (`localcheck_resume_evidence.go:47-56`). `validateResumeLink`'s `(Version==5) == (Selection != nil)` rule
     (`:122`) is unchanged. `resumeEnvelope`'s `version == resumeLinkVersion(check)` (`:110`) becomes "3 or 5 for
     suite, 5 for affected". A receipt must carry its link's version, and `resolveResumeProof` checks
     `Receipt.Version == Link.Version`.
  2. **Child command (PQ-1).** For an incomplete parent, `resumeSaved` builds no `flakeRerunCommands` batch. When
     origin/main declares `.sdlc/affected.json`, the child runs the affected command under its env contract,
     `affectedEnvCommand(cfg.Command, selection.GoPkgs, nil)`. That is the operator's "through the affected-run
     machinery", with the selection being `classifyIncomplete(parentFacts).Packages`. When it does not (ariadne
     today), the child runs the declared suite command whole: still one retry, still the composite. In ariadne
     `cmd/sdlc` dominates the suite, so that costs about the same. A new `incompleteRetryCommands(c, cfg, ok, sel)`
     is the one place this is decided, and it is pure, with a table test.
  3. **Pass rule (PQ-1).** A new branch in `resolveResumeProof`, taken when `p.Outcome == incomplete-timeout`. After
     the existing child identity and digest checks (`:354`), the composite is:
     - **passed**, with outcome `flakes.IncompleteRetried` (`incomplete-retry-passed`, `Lands()==true`), when the
       child has exit 0, is not timed out, has no failures, no unattributed reasons and no warnings, and the
       child's log names no package outside the selection as failed;
     - **incomplete-timeout** when the child's own facts classify incomplete;
     - otherwise **failed**, with `Classify` over the child's facts, so known flakes, unknown failures and so on
       behave as today.
     `guardSuitePass` verifies it unchanged, through `resolvedResumeRecords` → `resolveResumeProof`: the composite
     outcome replaces the parent's (`localcheck_republish.go:232-235`), and a landing outcome is explained. A test
     has a composite pass post green and a forged child digest refuse.
  4. **Envelope (PQ-1).** An incomplete child runs with the parent check's declared weight and timeout, not
     `rerunEnvelope`'s 1 slot and 15 min (`localcheck_affected_rerun.go:36-42`). `rerunEnvelope` gains the parent
     kind as an input; the known-flake path is unchanged.
  5. **Version gates swept (PQ-2).** These are every gate on an attempt's version in the resume and retention
     paths:
     - the parent gate (`localcheck_resume_evidence.go:285`) admits v5 exactly when the outcome is
       `incomplete-timeout`, and it replaces the unattributed-facts refusal (`:296`) for that parent with
       `classifyIncomplete(parentFacts)` being ok;
     - the child gate (`:354`) and retention (`localcheck_resume_retention.go:190`) are unchanged, because the
       child is an ordinary v2 `suite-rerun` attempt;
     - `suitereclassification.go:35` (v1/2 only) correctly ignores v5;
     - the store's decoder already admits v5 (M1).
     `grep -n 'Version [!=]= [0-9]'` over `cmd/sdlc/*.go` is the sweep's enumeration, recorded in the M2 log.
  6. **Typed runners (PQ-2).** v5 is written only when `d.flakeContract == ""` (M1, `localcheck_run.go`). A typed
     (v4) suite that times out keeps today's classification, so no v5 record ever lacks a contract it should bind.
     A test pins it.
  - **Minor:** `classifyIncomplete` selects the packages named by the marker and the named failures. Nothing
    hard-codes `cmd/sdlc` (the base layer runs in other repositories).
  - **Minor:** test strategy, one line per risky function. The marker parser and `classifyIncomplete` get a table
    test plus a fuzz test over arbitrary shard output seeded with truncated and partial markers; the property is
    that it never selects a package the log does not name, and never classifies a non-alarmed failure as
    incomplete. `resolveResumeProof`'s incomplete branch gets a pure table test (pass, fail, incomplete again,
    forged digest, wrong link version).
- 2026-10-25, M1 boundary review round 1 (BR-2): the Core concepts tables now name what M1 built: `SuiteProcess`
  and `shardAlarm` (the plan's `SuiteLogFacts.Alarmed` and `suiteAlarm` were drafts), `incompleteEligible`,
  `goIdentities`, `validLoadAverage`, `SuiteAttempt.LoadAverage`, and `hostload_other.go`. The separate evidence-record
  row is removed, not marked deleted, because it never shipped. The review also found these, all M1 work, fixed in
  the same round:
  - a FreeBSD build (`hostload_other.go`);
  - republish and saved-pass reuse refusing an incomplete attempt (`savedAttemptRecord`);
  - the typed-contract guard pinned on its own (`incompleteEligible`, plus the classifier's typed-facts case);
  - a live conformance test of the real Go alarm through both runner paths;
  - and four Minors.
  The outcome-consumer sweep enumerated:
  - `savedAttemptRecord` (republish and reuse): fixed;
  - `isFlakeEvidence` (guard, flake view, statistics, flake evidence): M1;
  - `flakeAttempts` callers: all filter by `isFlakeEvidence` first;
  - `encodeReclassifiedAttempt`: now refuses v5 explicitly;
  - `suitereclassification` (v1/2 only): already ignores v5;
  - the resume gates: M2.
