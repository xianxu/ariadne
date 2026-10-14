---
gate: boundary-review
issue: 364
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-14T01:09:59-07:00"
      dispatched: "2026-10-14T01:04:21-07:00"
      agent: codex
      findings:
        - id: BR-1
          severity: Critical
          title: Preparation can hide tracked-source changes from the mutation guard
          detail: cmd/sdlc/internal/judge/prepare.go:86 compares HEAD, staged entries, and Git's reported diff rather than actual tracked-file contents. Preparation running `git update-index --assume-unchanged tracked && printf changed > tracked` passes this comparison; a temporary regression exercising production Dispatch failed with `error=<nil> reviewer calls=1`. Git clean filters independently hide changed bytes from the same comparison. This violates the tracked-mutation rejection contract and permits reviewing/testing altered source as the pinned HEAD (ARCH-SECURE, ARCH-PURPOSE). Compare actual tracked contents, file types, and modes independently of Git filters and index suppression flags; add regression coverage for these hiding mechanisms.
          family: review-input-integrity
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - "n": 2
      timestamp: "2026-10-14T01:28:26-07:00"
      dispatched: "2026-10-14T01:24:19-07:00"
      agent: codex
      dispose:
        - id: BR-1
          disposition: addressed
          note: prepare.go:93 fingerprints actual filesystem state. Restoring the previous guard in a disposable pinned-head checkout made production Dispatch regressions fail for assume-unchanged, skip-worktree, clean filters, and permission changes; the corrected package passes. The disposable checkout was removed.
          round: 2
      findings:
        - id: BR-2
          severity: Important
          title: The promised unverifiable-dispute lifecycle regression is missing
          detail: workshop/plans/000364-review-environment-plan.md:198 promises a real temporary-repository lifecycle proving host evidence reaches the next judge, invalid replies cannot advance ledger/head, and a valid disposition closes the loop. boundaryresponse_test.go:44 manually constructs a disputed ledger and calls the pure validator; existing dispute lifecycle fixtures use ordinary findings. Add a production close/dispatch regression that disputes an unverifiable finding, checks the reviewer receives its evidence, rejects unanswered or note-less responses without advancement, and accepts an explicit evidence-answering disposition. Prove it fails when the dispute admission guard or ledger wiring is removed (ARCH-PURPOSE).
          family: boundary-lifecycle-coverage
          round: 2
      recipe: milestone-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#364 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-14T01:09:59-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `review-input-integrity` Preparation can hide tracked-source changes from the mutation guard
  cmd/sdlc/internal/judge/prepare.go:86 compares HEAD, staged entries, and Git's reported diff rather than actual tracked-file contents. Preparation running `git update-index --assume-unchanged tracked && printf changed > tracked` passes this comparison; a temporary regression exercising production Dispatch failed with `error=<nil> reviewer calls=1`. Git clean filters independently hide changed bytes from the same comparison. This violates the tracked-mutation rejection contract and permits reviewing/testing altered source as the pinned HEAD (ARCH-SECURE, ARCH-PURPOSE). Compare actual tracked contents, file types, and modes independently of Git filters and index suppression flags; add regression coverage for these hiding mechanisms.

## Round 2 — 2026-10-14T01:28:26-07:00 (codex) — BLOCKED

### Disposed

- BR-1 — addressed — prepare.go:93 fingerprints actual filesystem state. Restoring the previous guard in a disposable pinned-head checkout made production Dispatch regressions fail for assume-unchanged, skip-worktree, clean filters, and permission changes; the corrected package passes. The disposable checkout was removed.

### Raised

- **BR-2** [Important] `boundary-lifecycle-coverage` The promised unverifiable-dispute lifecycle regression is missing
  workshop/plans/000364-review-environment-plan.md:198 promises a real temporary-repository lifecycle proving host evidence reaches the next judge, invalid replies cannot advance ledger/head, and a valid disposition closes the loop. boundaryresponse_test.go:44 manually constructs a disputed ledger and calls the pure validator; existing dispute lifecycle fixtures use ordinary findings. Add a production close/dispatch regression that disputes an unverifiable finding, checks the reviewer receives its evidence, rejects unanswered or note-less responses without advancement, and accepts an explicit evidence-answering disposition. Prove it fails when the dispute admission guard or ledger wiring is removed (ARCH-PURPOSE).

## Open findings

- **BR-2** [Important] `boundary-lifecycle-coverage` The promised unverifiable-dispute lifecycle regression is missing
