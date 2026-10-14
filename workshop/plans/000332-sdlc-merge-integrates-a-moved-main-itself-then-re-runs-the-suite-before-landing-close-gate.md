---
gate: boundary-review
issue: 332
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-13T22:34:25-07:00"
      dispatched: "2026-10-13T22:28:12-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Important
          title: The docs-only check uses git diff --name-only with renames on, so code renamed to .md looks docs-only
          detail: 'cigate.go:473 lists only the new name of a renamed file (reproduced: f.go to f.md lists just f.md), so an old suite record still counts after code is removed. Use --no-renames -z, as claim.go:673 and issuemigrate.go:257 do, and add a rename test.'
          family: docs-only-classification-must-see-every-changed-path
          round: 1
        - id: BR-2
          severity: Important
          title: parseCommitStatuses and localRecordsOnAncestors have no direct tests
          detail: 'The newest-first dedupe in parseCommitStatuses is untested, and so are the walk''s limits: merge base, root commit, stopping once every name is found, and the 30-commit cap.'
          family: new-parser-needs-direct-tests
          round: 1
        - id: BR-3
          severity: Minor
          title: localRecordsOnAncestors treats any rev-list error as "root commit"
          family: error-swallowed-as-benign-case
          round: 1
        - id: BR-4
          severity: Minor
          title: '^test: also matches a prerequisite-only "Not a target" entry in make -p output'
          family: make-database-parse-precision
          round: 1
        - id: BR-5
          severity: Minor
          title: The make -p probe has no timeout; a hanging $(shell) in main's Makefile stalls merge
          family: unbounded-subprocess
          round: 1
        - id: BR-6
          severity: Minor
          title: Makefile probe plus effectiveLocalChecks is repeated in cigate.go and localcheck_cmd.go
          family: shared-helper-for-repeated-derivation
          round: 1
        - id: BR-7
          severity: Minor
          title: Plan names suiteAncestorRecords and makeTestFacts(root); code has localRecordsOnAncestors and makeTestFacts(git, rev); weight 0 means default, not refused
          family: plan-table-matches-code
          round: 1
      boundary: M1
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - "n": 2
      timestamp: "2026-10-13T23:02:31-07:00"
      dispatched: "2026-10-13T22:58:03-07:00"
      agent: codex
      dispose:
        - id: BR-1
          disposition: addressed
          note: ChangedPathArgs supplies --no-renames -z to the identified classifiers. Removing --no-renames made both TestLocalRecordsOnAncestors and TestDiffNamesListsBothSidesOfARename fail.
          round: 2
        - id: BR-2
          disposition: addressed
          note: TestParseCommitStatuses directly tests newest-first deduplication and malformed input. TestLocalRecordsOnAncestors exercises merge-base exclusion, early termination, the commit cap, root commits, and code changes; these tests pass.
          round: 2
        - id: BR-3
          disposition: not-addressed
          note: cigate.go:481 now treats every rev-parse failure as a root commit, moving the uncertainty rather than resolving it. No regression test injects either rev-parse or rev-list failure. Positively establish parentlessness and propagate operational errors.
          round: 2
        - id: BR-4
          disposition: addressed
          note: parseMakeTestFacts excludes prerequisite-only entries. TestParseMakeTestFactsSkipsNotATarget passes and fails when that exclusion is removed.
          round: 2
        - id: BR-5
          disposition: not-addressed
          note: 'localchecks_make.go:79-84 adds CommandContext but no bounded pipe wait or descendant cleanup. A temporary reproduction with a 50 ms deadline and $(shell sleep 1) returned after 1.04 seconds: Output waits for the descendant holding stderr open. No committed timeout regression exists. Bound subprocess lifetime and pipe draining, and test both.'
          round: 2
        - id: BR-6
          disposition: addressed
          note: withDefaultSuite is now the shared derivation called by runCIGate and runLocalCheckCmd. Both production paths have passing integration coverage; this consolidation preserves behavior.
          round: 2
        - id: BR-7
          disposition: addressed
          note: Plan R2's “Names as delivered” passage at line 234 explicitly corrects both function names/signatures and zero-weight semantics. These match cigate.go:472, localchecks_make.go:51, and localchecks.go:115.
          round: 2
      findings:
        - id: BR-8
          severity: Critical
          title: A declared suite can bypass the mandatory landing check
          detail: 'localchecks_make.go:38 returns a declared suite unchanged; localchecks.go:143 likewise preserves Required=false. A main-side declaration containing {"name":"suite","command":["true"]} therefore makes runCIGate accept green CI without any suite record. A temporary integration regression reproduced this acceptance. ARCH-PURPOSE: preserve command/weight overrides while requiring the suite, or reject optional suite declarations; cover omitted and explicit-false required fields.'
          family: mandatory-checks-cannot-be-disabled-by-overrides
          round: 2
        - id: BR-9
          severity: Important
          title: README update appears missing for the default suite and weight setting
          detail: README.md:318-324 still describes only explicitly declared required checks. This range adds automatic make-test enforcement, the close-suite invocation, ancestor-record reuse, and the weight configuration key without updating README. Update its local-check guidance alongside the existing atlas and embedded-help changes.
          family: readme-tracks-user-facing-contracts
          round: 2
      boundary: M1
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - "n": 3
      timestamp: "2026-10-13T23:15:59-07:00"
      dispatched: "2026-10-13T23:10:43-07:00"
      agent: codex
      dispose:
        - id: BR-1
          disposition: addressed
          note: ChangedPathArgs disables rename detection; the rename regressions passed.
          round: 3
        - id: BR-2
          disposition: addressed
          note: Direct status-parser and ancestor-walk tests exist and passed.
          round: 3
        - id: BR-3
          disposition: addressed
          note: Parent and ancestor read failures propagate. Both regression cases failed when error swallowing was restored in a disposable checkout.
          round: 3
        - id: BR-4
          disposition: addressed
          note: parseMakeTestFacts excludes prerequisite-only entries; its direct regression passed.
          round: 3
        - id: BR-5
          disposition: addressed
          note: The probe has a deadline and bounded pipe draining. Removing the deadline made the hanging-shell regression fail.
          round: 3
        - id: BR-6
          disposition: addressed
          note: Both the CI gate and local-check command derive checks through withDefaultSuite.
          round: 3
        - id: BR-7
          disposition: addressed
          note: Plan revision R2 records the delivered names, main-revision probe signature, and weight-zero default.
          round: 3
        - id: BR-8
          disposition: addressed
          note: effectiveLocalChecks forces declared suites to required while preserving execution settings. Removing enforcement failed both omitted/false integration cases and the pure regression.
          round: 3
        - id: BR-9
          disposition: addressed
          note: README.md:326-332 documents the close command, default suite, mandatory overrides, weight and ancestor reuse, matching localchecks.go, localchecks_make.go and cigate.go.
          round: 3
      findings:
        - id: BR-10
          severity: Important
          title: Probe cleanup tests accept natural child exit instead of proving termination
          detail: 'cmd/sdlc/localchecks_make_process_test.go:71 waits two minutes for children that sleep only 60 seconds. Removing process-group configuration and both termination calls leaves both probe cases passing; the corresponding runner tests correctly fail. ARCH-ORDER: this is the 2nd finding in family unbounded-subprocess. State and apply the rule across probe/runner timeout and parent-exit cases: the cleanup assertion must expire before the child can finish naturally. Verify each cleanup mutation goes red.'
          family: unbounded-subprocess
          round: 3
        - id: BR-11
          severity: Important
          title: The new commit-status dependency has no live conformance check
          detail: cmd/sdlc/cigate.go:312-339 introduces a status endpoint and newest-first interpretation used to admit ancestor evidence. Tests exercise handwritten JSON and ciFakeGH; TestLiveJobRunConformance covers only JobRun. ARCH-MOCK requires a read-only opt-in check through realGH.CommitStatuses against independently known status history, plus a documented execution cadence.
          family: external-fakes-require-live-conformance
          round: 3
      boundary: M1
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - "n": 4
      timestamp: "2026-10-13T23:22:41-07:00"
      dispatched: "2026-10-13T23:19:52-07:00"
      agent: codex
      dispose:
        - id: BR-1
          disposition: addressed
          note: ChangedPathArgs supplies --no-renames -z to classification callers; rename and ancestor regressions pass.
          round: 4
        - id: BR-2
          disposition: addressed
          note: Direct status-parser and ancestor-walk tests exist and pass.
          round: 4
        - id: BR-3
          disposition: addressed
          note: Parent and ancestor-read failures propagate; error and root-commit regressions pass.
          round: 4
        - id: BR-4
          disposition: addressed
          note: The parser excludes prerequisite-only “Not a target” entries; its direct regression passes.
          round: 4
        - id: BR-5
          disposition: addressed
          note: Probe timeout, bounded pipe draining and process-group cleanup are implemented; lifetime regressions pass.
          round: 4
        - id: BR-6
          disposition: addressed
          note: withDefaultSuite is shared by the CI gate and local-check command.
          round: 4
        - id: BR-7
          disposition: addressed
          note: Plan revision R2 records the delivered helper names, signatures and zero-weight default, matching the implementation.
          round: 4
        - id: BR-8
          disposition: addressed
          note: effectiveLocalChecks forces a declared suite to required while preserving execution settings; pure and gate regressions pass.
          round: 4
        - id: BR-9
          disposition: addressed
          note: README.md:326 documents invocation, mandatory suite, main-derived default, weight and ancestor reuse, matching the implementation.
          round: 4
        - id: BR-10
          disposition: addressed
          note: Both probe assertions now expire after one second. Removing group configuration and termination in a disposable copy makes all four probe/runner lifetime cases fail on surviving children.
          round: 4
        - id: BR-11
          disposition: addressed
          note: Live conformance passes through realGH.CommitStatuses against four pinned historical events; changing reduction to oldest-wins makes it fail. Atlas documents execution cadence.
          round: 4
      boundary: M1
      recipe: milestone-review
      isolation: fenced
      reviewed: 5599f19d3e92e5e60af36d6f2b78076750a0562c
      blocked: false
    - "n": 5
      timestamp: "2026-10-13T23:48:50-07:00"
      dispatched: "2026-10-13T23:46:42-07:00"
      agent: codex
      boundary: M2
      recipe: milestone-review
      isolation: fenced
      blocked: true
      protocol_error: no valid findings block
    - "n": 6
      timestamp: "2026-10-14T00:01:36-07:00"
      dispatched: "2026-10-13T23:58:36-07:00"
      agent: codex
      findings:
        - id: BR-12
          severity: Important
          title: CI polling can exceed its advertised deadline and accept a late green result
          detail: 'cmd/sdlc/cigate_wait.go:86-93 passes the original context through a complete gate evaluation and checks elapsed time only for pending/unregistered results. ARCH-CONSTRAINTS: enforce the remaining budget on polling IO, separately from suite runtime, and test slow reads and green completion after expiry.'
          family: polling-deadlines-cover-inflight-work
          round: 6
        - id: BR-13
          severity: Important
          title: Integration merge and push subprocesses have no deadline or cancellation
          detail: 'cmd/sdlc/landingintegrate.go:199,246 uses the unbounded execGitRunner in cmd/sdlc/runner.go:37-44. This is the 3rd finding in family unbounded-subprocess: state the rule and sweep the integration call graph, including probes and cleanup, rather than fixing only these instances. ARCH-CONSTRAINTS and ARCH-ORDER require bounded process and descendant lifetimes plus tests for cancellation and uncertain push completion.'
          family: unbounded-subprocess
          round: 6
        - id: BR-14
          severity: Important
          title: Integration state remains mutable outside the pure transition model
          detail: 'cmd/sdlc/landing.go:430-435 updates head, PR, attempts and resume state directly; cmd/sdlc/landingintegrate.go:43 only selects actions, and its lease-refused event is test-only. ARCH-ORDER: introduce an owned state/event transition model returning next state and effects, route production outcomes through it, and revise the plan to describe the enforced boundary.'
          family: authoritative-state-transitions-owned-by-core
          round: 6
      boundary: M2
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - "n": 7
      timestamp: "2026-10-14T01:02:22-07:00"
      dispatched: "2026-10-14T00:53:58-07:00"
      agent: codex
      dispose:
        - id: BR-12
          disposition: not-addressed
          note: Git/GitHub read deadlines and late-green rejection are corrected, with mutation-sensitive tests. However, prepareCIGateFacts at cmd/sdlc/cigate_wait.go:145 reaches probeMakeTestFacts, which creates a background context at cmd/sdlc/localchecks_make.go:84. A real Makefile probe took 1.07 seconds with only 40 ms remaining in the CI budget. Propagate the remaining budget through the complete facts-probe path, retaining separate bounded cleanup.
          round: 7
        - id: BR-13
          disposition: not-addressed
          note: 'Merge/push deadlines, descendant cleanup, and uncertain-push recovery have meaningful regression evidence. The integration call-graph sweep remains incomplete: publish replay calls gitx.RunGit, whose runner uses context.Background at cmd/sdlc/internal/gitx/window.go:28; runEnv does likewise at rebasedpatch.go:17. A cancelled runLandingPublishGate still spent 1.06 seconds in its first Git read. Ownership also reaches archiveRead at landingarchive.go:126, which invokes the unbounded landingArchiveRunner directly. Apply cancellation and command ceilings across the entire enumerated graph.'
          round: 7
        - id: BR-14
          disposition: addressed
          note: Production integration outcomes now pass through transitionIntegration in cmd/sdlc/integrationstate.go, owning head, lease base, PR, resume state, phase, and attempts. Removing retry-count advancement made both state regression tests fail. Production retry, recovery, and cancelled-push tests passed; Plan revision R6 describes the delivered model.
          round: 7
      boundary: M2
      recipe: milestone-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#332 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-13T22:34:25-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `docs-only-classification-must-see-every-changed-path` The docs-only check uses git diff --name-only with renames on, so code renamed to .md looks docs-only
  cigate.go:473 lists only the new name of a renamed file (reproduced: f.go to f.md lists just f.md), so an old suite record still counts after code is removed. Use --no-renames -z, as claim.go:673 and issuemigrate.go:257 do, and add a rename test.
- **BR-2** [Important] `new-parser-needs-direct-tests` parseCommitStatuses and localRecordsOnAncestors have no direct tests
  The newest-first dedupe in parseCommitStatuses is untested, and so are the walk's limits: merge base, root commit, stopping once every name is found, and the 30-commit cap.
- **BR-3** [Minor] `error-swallowed-as-benign-case` localRecordsOnAncestors treats any rev-list error as "root commit"
- **BR-4** [Minor] `make-database-parse-precision` ^test: also matches a prerequisite-only "Not a target" entry in make -p output
- **BR-5** [Minor] `unbounded-subprocess` The make -p probe has no timeout; a hanging $(shell) in main's Makefile stalls merge
- **BR-6** [Minor] `shared-helper-for-repeated-derivation` Makefile probe plus effectiveLocalChecks is repeated in cigate.go and localcheck_cmd.go
- **BR-7** [Minor] `plan-table-matches-code` Plan names suiteAncestorRecords and makeTestFacts(root); code has localRecordsOnAncestors and makeTestFacts(git, rev); weight 0 means default, not refused

## Round 2 — 2026-10-13T23:02:31-07:00 (codex) — BLOCKED

### Disposed

- BR-1 — addressed — ChangedPathArgs supplies --no-renames -z to the identified classifiers. Removing --no-renames made both TestLocalRecordsOnAncestors and TestDiffNamesListsBothSidesOfARename fail.
- BR-2 — addressed — TestParseCommitStatuses directly tests newest-first deduplication and malformed input. TestLocalRecordsOnAncestors exercises merge-base exclusion, early termination, the commit cap, root commits, and code changes; these tests pass.
- BR-3 — not-addressed — cigate.go:481 now treats every rev-parse failure as a root commit, moving the uncertainty rather than resolving it. No regression test injects either rev-parse or rev-list failure. Positively establish parentlessness and propagate operational errors.
- BR-4 — addressed — parseMakeTestFacts excludes prerequisite-only entries. TestParseMakeTestFactsSkipsNotATarget passes and fails when that exclusion is removed.
- BR-5 — not-addressed — localchecks_make.go:79-84 adds CommandContext but no bounded pipe wait or descendant cleanup. A temporary reproduction with a 50 ms deadline and $(shell sleep 1) returned after 1.04 seconds: Output waits for the descendant holding stderr open. No committed timeout regression exists. Bound subprocess lifetime and pipe draining, and test both.
- BR-6 — addressed — withDefaultSuite is now the shared derivation called by runCIGate and runLocalCheckCmd. Both production paths have passing integration coverage; this consolidation preserves behavior.
- BR-7 — addressed — Plan R2's “Names as delivered” passage at line 234 explicitly corrects both function names/signatures and zero-weight semantics. These match cigate.go:472, localchecks_make.go:51, and localchecks.go:115.

### Raised

- **BR-8** [Critical] `mandatory-checks-cannot-be-disabled-by-overrides` A declared suite can bypass the mandatory landing check
  localchecks_make.go:38 returns a declared suite unchanged; localchecks.go:143 likewise preserves Required=false. A main-side declaration containing {"name":"suite","command":["true"]} therefore makes runCIGate accept green CI without any suite record. A temporary integration regression reproduced this acceptance. ARCH-PURPOSE: preserve command/weight overrides while requiring the suite, or reject optional suite declarations; cover omitted and explicit-false required fields.
- **BR-9** [Important] `readme-tracks-user-facing-contracts` README update appears missing for the default suite and weight setting
  README.md:318-324 still describes only explicitly declared required checks. This range adds automatic make-test enforcement, the close-suite invocation, ancestor-record reuse, and the weight configuration key without updating README. Update its local-check guidance alongside the existing atlas and embedded-help changes.

## Round 3 — 2026-10-13T23:15:59-07:00 (codex) — BLOCKED

### Disposed

- BR-1 — addressed — ChangedPathArgs disables rename detection; the rename regressions passed.
- BR-2 — addressed — Direct status-parser and ancestor-walk tests exist and passed.
- BR-3 — addressed — Parent and ancestor read failures propagate. Both regression cases failed when error swallowing was restored in a disposable checkout.
- BR-4 — addressed — parseMakeTestFacts excludes prerequisite-only entries; its direct regression passed.
- BR-5 — addressed — The probe has a deadline and bounded pipe draining. Removing the deadline made the hanging-shell regression fail.
- BR-6 — addressed — Both the CI gate and local-check command derive checks through withDefaultSuite.
- BR-7 — addressed — Plan revision R2 records the delivered names, main-revision probe signature, and weight-zero default.
- BR-8 — addressed — effectiveLocalChecks forces declared suites to required while preserving execution settings. Removing enforcement failed both omitted/false integration cases and the pure regression.
- BR-9 — addressed — README.md:326-332 documents the close command, default suite, mandatory overrides, weight and ancestor reuse, matching localchecks.go, localchecks_make.go and cigate.go.

### Raised

- **BR-10** [Important] `unbounded-subprocess` Probe cleanup tests accept natural child exit instead of proving termination
  cmd/sdlc/localchecks_make_process_test.go:71 waits two minutes for children that sleep only 60 seconds. Removing process-group configuration and both termination calls leaves both probe cases passing; the corresponding runner tests correctly fail. ARCH-ORDER: this is the 2nd finding in family unbounded-subprocess. State and apply the rule across probe/runner timeout and parent-exit cases: the cleanup assertion must expire before the child can finish naturally. Verify each cleanup mutation goes red.
- **BR-11** [Important] `external-fakes-require-live-conformance` The new commit-status dependency has no live conformance check
  cmd/sdlc/cigate.go:312-339 introduces a status endpoint and newest-first interpretation used to admit ancestor evidence. Tests exercise handwritten JSON and ciFakeGH; TestLiveJobRunConformance covers only JobRun. ARCH-MOCK requires a read-only opt-in check through realGH.CommitStatuses against independently known status history, plus a documented execution cadence.

## Round 4 — 2026-10-13T23:22:41-07:00 (codex) — passed

### Disposed

- BR-1 — addressed — ChangedPathArgs supplies --no-renames -z to classification callers; rename and ancestor regressions pass.
- BR-2 — addressed — Direct status-parser and ancestor-walk tests exist and pass.
- BR-3 — addressed — Parent and ancestor-read failures propagate; error and root-commit regressions pass.
- BR-4 — addressed — The parser excludes prerequisite-only “Not a target” entries; its direct regression passes.
- BR-5 — addressed — Probe timeout, bounded pipe draining and process-group cleanup are implemented; lifetime regressions pass.
- BR-6 — addressed — withDefaultSuite is shared by the CI gate and local-check command.
- BR-7 — addressed — Plan revision R2 records the delivered helper names, signatures and zero-weight default, matching the implementation.
- BR-8 — addressed — effectiveLocalChecks forces a declared suite to required while preserving execution settings; pure and gate regressions pass.
- BR-9 — addressed — README.md:326 documents invocation, mandatory suite, main-derived default, weight and ancestor reuse, matching the implementation.
- BR-10 — addressed — Both probe assertions now expire after one second. Removing group configuration and termination in a disposable copy makes all four probe/runner lifetime cases fail on surviving children.
- BR-11 — addressed — Live conformance passes through realGH.CommitStatuses against four pinned historical events; changing reduction to oldest-wins makes it fail. Atlas documents execution cadence.

## Round 5 — 2026-10-13T23:48:50-07:00 (codex) — BLOCKED

**Protocol error:** no valid findings block — this round contributed no findings.

## Round 6 — 2026-10-14T00:01:36-07:00 (codex) — BLOCKED

### Raised

- **BR-12** [Important] `polling-deadlines-cover-inflight-work` CI polling can exceed its advertised deadline and accept a late green result
  cmd/sdlc/cigate_wait.go:86-93 passes the original context through a complete gate evaluation and checks elapsed time only for pending/unregistered results. ARCH-CONSTRAINTS: enforce the remaining budget on polling IO, separately from suite runtime, and test slow reads and green completion after expiry.
- **BR-13** [Important] `unbounded-subprocess` Integration merge and push subprocesses have no deadline or cancellation
  cmd/sdlc/landingintegrate.go:199,246 uses the unbounded execGitRunner in cmd/sdlc/runner.go:37-44. This is the 3rd finding in family unbounded-subprocess: state the rule and sweep the integration call graph, including probes and cleanup, rather than fixing only these instances. ARCH-CONSTRAINTS and ARCH-ORDER require bounded process and descendant lifetimes plus tests for cancellation and uncertain push completion.
- **BR-14** [Important] `authoritative-state-transitions-owned-by-core` Integration state remains mutable outside the pure transition model
  cmd/sdlc/landing.go:430-435 updates head, PR, attempts and resume state directly; cmd/sdlc/landingintegrate.go:43 only selects actions, and its lease-refused event is test-only. ARCH-ORDER: introduce an owned state/event transition model returning next state and effects, route production outcomes through it, and revise the plan to describe the enforced boundary.

## Round 7 — 2026-10-14T01:02:22-07:00 (codex) — BLOCKED

### Disposed

- BR-12 — not-addressed — Git/GitHub read deadlines and late-green rejection are corrected, with mutation-sensitive tests. However, prepareCIGateFacts at cmd/sdlc/cigate_wait.go:145 reaches probeMakeTestFacts, which creates a background context at cmd/sdlc/localchecks_make.go:84. A real Makefile probe took 1.07 seconds with only 40 ms remaining in the CI budget. Propagate the remaining budget through the complete facts-probe path, retaining separate bounded cleanup.
- BR-13 — not-addressed — Merge/push deadlines, descendant cleanup, and uncertain-push recovery have meaningful regression evidence. The integration call-graph sweep remains incomplete: publish replay calls gitx.RunGit, whose runner uses context.Background at cmd/sdlc/internal/gitx/window.go:28; runEnv does likewise at rebasedpatch.go:17. A cancelled runLandingPublishGate still spent 1.06 seconds in its first Git read. Ownership also reaches archiveRead at landingarchive.go:126, which invokes the unbounded landingArchiveRunner directly. Apply cancellation and command ceilings across the entire enumerated graph.
- BR-14 — addressed — Production integration outcomes now pass through transitionIntegration in cmd/sdlc/integrationstate.go, owning head, lease base, PR, resume state, phase, and attempts. Removing retry-count advancement made both state regression tests fail. Production retry, recovery, and cancelled-push tests passed; Plan revision R6 describes the delivered model.

## Open findings

- **BR-12** [Important] `polling-deadlines-cover-inflight-work` CI polling can exceed its advertised deadline and accept a late green result
- **BR-13** [Important] `unbounded-subprocess` Integration merge and push subprocesses have no deadline or cancellation
