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

## Open findings

- **BR-1** [Minor] `status-read-single-parser` Remaining porcelain readers (merge porcelainPaths, push parsePorcelainStatus) still bypass ParseStatusZ
- **BR-2** [Minor] `exec-plumbing-duplication` trackerEnv.statusEntries duplicates gitEnv's exec/stderr/error plumbing
- **BR-3** [Minor] `import-grouping` fleet/facts.go puts the gitx import in the stdlib import group
