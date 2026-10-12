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
    - "n": 2
      timestamp: "2026-10-12T07:46:56-07:00"
      dispatched: "2026-10-12T07:46:29-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: term.go:58-61 now scopes the claim to a verb's refusal and lists the remaining direct exits (main.go, activetime.go, repolock.go, judge.go). A grep of os.Exit( in non-test cmd/sdlc files finds exactly that set.
          round: 2
      recipe: fix-review
      isolation: checkout
      reviewed: 891711ba3f7f864161f4907434d966a65870d527
      blocked: false
---

# Gate ledger — ariadne#191 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-12T07:44:55-07:00 (claude) — passed

### Raised

- **BR-1** [Minor] `doc-claim-unverified-against-code` term.go comment says exitWithCode "stays the one process-exit seam", but judge.go:200 still calls os.Exit(1) directly
  judge.go:200 exits directly on a judge failure, so the die cleanups are skipped and expectExit cannot swap it. Every other direct os.Exit is a top-level or signal exit: main.go:70, activetime.go:245, repolock.go:242/244. Either scope the comment or route judge.go:200 through exitWithCode.

## Round 2 — 2026-10-12T07:46:56-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — term.go:58-61 now scopes the claim to a verb's refusal and lists the remaining direct exits (main.go, activetime.go, repolock.go, judge.go). A grep of os.Exit( in non-test cmd/sdlc files finds exactly that set.

## Open findings

(none — every finding has been disposed)
