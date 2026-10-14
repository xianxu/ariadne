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
---

# Gate ledger — ariadne#364 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-14T01:09:59-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `review-input-integrity` Preparation can hide tracked-source changes from the mutation guard
  cmd/sdlc/internal/judge/prepare.go:86 compares HEAD, staged entries, and Git's reported diff rather than actual tracked-file contents. Preparation running `git update-index --assume-unchanged tracked && printf changed > tracked` passes this comparison; a temporary regression exercising production Dispatch failed with `error=<nil> reviewer calls=1`. Git clean filters independently hide changed bytes from the same comparison. This violates the tracked-mutation rejection contract and permits reviewing/testing altered source as the pinned HEAD (ARCH-SECURE, ARCH-PURPOSE). Compare actual tracked contents, file types, and modes independently of Git filters and index suppression flags; add regression coverage for these hiding mechanisms.

## Open findings

- **BR-1** [Critical] `review-input-integrity` Preparation can hide tracked-source changes from the mutation guard
