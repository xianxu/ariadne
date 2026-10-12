---
gate: boundary-review
issue: 191
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-12T07:44:55-07:00"
      dispatched: "2026-10-12T07:42:32-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Minor
          title: term.go comment says exitWithCode "stays the one process-exit seam", but judge.go:200 still calls os.Exit(1) directly
          detail: 'judge.go:200 exits directly on a judge failure, so the die cleanups are skipped and expectExit cannot swap it. Every other direct os.Exit is a top-level or signal exit: main.go:70, activetime.go:245, repolock.go:242/244. Either scope the comment or route judge.go:200 through exitWithCode.'
          family: doc-claim-unverified-against-code
          round: 1
      recipe: small-diff-review
      isolation: checkout
      reviewed: e2ffaca01aba4238a63fc4b8476663263521a812
      blocked: false
---

# Gate ledger — ariadne#191 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-12T07:44:55-07:00 (claude) — passed

### Raised

- **BR-1** [Minor] `doc-claim-unverified-against-code` term.go comment says exitWithCode "stays the one process-exit seam", but judge.go:200 still calls os.Exit(1) directly
  judge.go:200 exits directly on a judge failure, so the die cleanups are skipped and expectExit cannot swap it. Every other direct os.Exit is a top-level or signal exit: main.go:70, activetime.go:245, repolock.go:242/244. Either scope the comment or route judge.go:200 through exitWithCode.

## Open findings

- **BR-1** [Minor] `doc-claim-unverified-against-code` term.go comment says exitWithCode "stays the one process-exit seam", but judge.go:200 still calls os.Exit(1) directly
