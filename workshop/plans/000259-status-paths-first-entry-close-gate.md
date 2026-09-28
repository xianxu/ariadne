---
gate: boundary-review
issue: 259
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-09-28T09:53:57-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Minor
          title: Remaining porcelain readers (merge porcelainPaths, push parsePorcelainStatus) still bypass ParseStatusZ
          detail: merge.go:202 splits on whitespace over trimmed output; push.go:516 slices line[3:] over untrimmed non -z output. Neither drops the first entry, but both mis-handle quoted paths with spaces; the new lesson prescribes ParseStatusZ for status reads.
          family: status-read-single-parser
          round: 1
        - id: BR-2
          severity: Minor
          title: trackerEnv.statusEntries duplicates gitEnv's exec/stderr/error plumbing
          detail: trackerenv.go:88 vs gitEnv; a shared untrimmed gitRaw(dir, args) helper would keep one exec path (ARCH-DRY).
          family: exec-plumbing-duplication
          round: 1
        - id: BR-3
          severity: Minor
          title: fleet/facts.go puts the gitx import in the stdlib import group
          family: import-grouping
          round: 1
      recipe: small-diff-review
      blocked: false
    - "n": 2
      timestamp: "2026-09-28T10:28:53-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: porcelainPaths and parsePorcelainStatus deleted; merge/push/startplan read -z via gitx.ParseStatusZ; spaced-path case in TestAssessDirty; remaining --porcelain callers only test emptiness.
          round: 2
        - id: BR-2
          disposition: addressed
          note: trackerenv.go gitRaw is the single exec path; gitEnv trims it, statusEntries parses it.
          round: 2
        - id: BR-3
          disposition: addressed
          note: fleet/facts.go now puts gitx in its own group after the stdlib imports.
          round: 2
      findings:
        - id: BR-4
          severity: Minor
          title: Status reads via CombinedOutput runners mix stderr into the -z stream; start-plan swallows the parse error
          detail: '2nd in family. runner.go:37 uses CombinedOutput, so a git status warning fails ParseStatusZ opaquely in merge/push and zeroes DirtyCode in startplan.go:455. Rule: status is read stdout-only through one runner-level StatusEntries helper, replacing the 4 run+parse sites.'
          family: status-read-single-parser
          round: 2
        - id: BR-5
          severity: Minor
          title: merge_test.go puts the gitx import in the stdlib group
          detail: '2nd in family. Rule: goimports -local github.com/xianxu/ariadne enforced by lint; no goimports/gci installed. Measured prevalence about 10 files (branchcreate.go, setstatus.go, workspace.go, workspacepaths.go, several _test.go files), so fix via tooling, not per instance.'
          family: import-grouping
          round: 2
      recipe: milestone-review
      blocked: false
---

# Gate ledger — ariadne#259 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-09-28T09:53:57-07:00 (claude) — passed

### Raised

- **BR-1** [Minor] `status-read-single-parser` Remaining porcelain readers (merge porcelainPaths, push parsePorcelainStatus) still bypass ParseStatusZ
  merge.go:202 splits on whitespace over trimmed output; push.go:516 slices line[3:] over untrimmed non -z output. Neither drops the first entry, but both mis-handle quoted paths with spaces; the new lesson prescribes ParseStatusZ for status reads.
- **BR-2** [Minor] `exec-plumbing-duplication` trackerEnv.statusEntries duplicates gitEnv's exec/stderr/error plumbing
  trackerenv.go:88 vs gitEnv; a shared untrimmed gitRaw(dir, args) helper would keep one exec path (ARCH-DRY).
- **BR-3** [Minor] `import-grouping` fleet/facts.go puts the gitx import in the stdlib import group

## Round 2 — 2026-09-28T10:28:53-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — porcelainPaths and parsePorcelainStatus deleted; merge/push/startplan read -z via gitx.ParseStatusZ; spaced-path case in TestAssessDirty; remaining --porcelain callers only test emptiness.
- BR-2 — addressed — trackerenv.go gitRaw is the single exec path; gitEnv trims it, statusEntries parses it.
- BR-3 — addressed — fleet/facts.go now puts gitx in its own group after the stdlib imports.

### Raised

- **BR-4** [Minor] `status-read-single-parser` Status reads via CombinedOutput runners mix stderr into the -z stream; start-plan swallows the parse error
  2nd in family. runner.go:37 uses CombinedOutput, so a git status warning fails ParseStatusZ opaquely in merge/push and zeroes DirtyCode in startplan.go:455. Rule: status is read stdout-only through one runner-level StatusEntries helper, replacing the 4 run+parse sites.
- **BR-5** [Minor] `import-grouping` merge_test.go puts the gitx import in the stdlib group
  2nd in family. Rule: goimports -local github.com/xianxu/ariadne enforced by lint; no goimports/gci installed. Measured prevalence about 10 files (branchcreate.go, setstatus.go, workspace.go, workspacepaths.go, several _test.go files), so fix via tooling, not per instance.

## Open findings

- **BR-4** [Minor] `status-read-single-parser` Status reads via CombinedOutput runners mix stderr into the -z stream; start-plan swallows the parse error
- **BR-5** [Minor] `import-grouping` merge_test.go puts the gitx import in the stdlib group
