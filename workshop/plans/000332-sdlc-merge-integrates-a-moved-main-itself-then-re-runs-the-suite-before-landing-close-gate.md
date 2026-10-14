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

## Open findings

- **BR-3** [Minor] `error-swallowed-as-benign-case` localRecordsOnAncestors treats any rev-list error as "root commit"
- **BR-5** [Minor] `unbounded-subprocess` The make -p probe has no timeout; a hanging $(shell) in main's Makefile stalls merge
- **BR-8** [Critical] `mandatory-checks-cannot-be-disabled-by-overrides` A declared suite can bypass the mandatory landing check
- **BR-9** [Important] `readme-tracks-user-facing-contracts` README update appears missing for the default suite and weight setting
