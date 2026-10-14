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

## Open findings

- **BR-1** [Important] `docs-only-classification-must-see-every-changed-path` The docs-only check uses git diff --name-only with renames on, so code renamed to .md looks docs-only
- **BR-2** [Important] `new-parser-needs-direct-tests` parseCommitStatuses and localRecordsOnAncestors have no direct tests
- **BR-3** [Minor] `error-swallowed-as-benign-case` localRecordsOnAncestors treats any rev-list error as "root commit"
- **BR-4** [Minor] `make-database-parse-precision` ^test: also matches a prerequisite-only "Not a target" entry in make -p output
- **BR-5** [Minor] `unbounded-subprocess` The make -p probe has no timeout; a hanging $(shell) in main's Makefile stalls merge
- **BR-6** [Minor] `shared-helper-for-repeated-derivation` Makefile probe plus effectiveLocalChecks is repeated in cigate.go and localcheck_cmd.go
- **BR-7** [Minor] `plan-table-matches-code` Plan names suiteAncestorRecords and makeTestFacts(root); code has localRecordsOnAncestors and makeTestFacts(git, rev); weight 0 means default, not refused
