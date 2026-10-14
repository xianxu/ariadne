# Boundary Review — ariadne#364 (whole-issue close)

| field | value |
|-------|-------|
| issue | 364 — Boundary review environment: allow read-only process listing, prepare the checkout (declared prep step), and REWORK must cite an answerable finding |
| repo | ariadne |
| issue file | workshop/issues/000364-boundary-review-environment-allow-read-only-process-listing-prepare-the-checkout-declared-prep-step-and-rework-must-cite-an-answerable-finding.md |
| boundary | whole-issue close |
| milestone | — |
| window | 85135ce32b676c60628649a96c93446ac238a11c..fe02fd92b39babca787cd32ce9172afedfc38ac2 |
| command | sdlc close --issue 364 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-14T01:09:59-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The preparation lifecycle, answerable-response validation, and documentation largely match the issue and plan. One reproduced correctness bug blocks shipping: preparation can change tracked source without triggering the mutation guard, allowing review against contents that differ from the pinned HEAD.

1. **Strengths**

- Preparation declarations come from pinned regular Git blobs, with strict parsing.
- Retry admission reuses the existing two-attempt loop and resets/reprepares before retry.
- Unverifiable disputes require explicit evidence-answering dispositions.
- README and atlas updates cover the new configuration, limits, and failure behavior.

2. **Critical findings**

```findings
findings:
  - id: new
    severity: Critical
    family: review-input-integrity
    title: |
      Preparation can hide tracked-source changes from the mutation guard
    detail: |
      cmd/sdlc/internal/judge/prepare.go:86 compares HEAD, staged entries, and Git's reported diff rather than actual tracked-file contents. Preparation running `git update-index --assume-unchanged tracked && printf changed > tracked` passes this comparison; a temporary regression exercising production Dispatch failed with `error=<nil> reviewer calls=1`. Git clean filters independently hide changed bytes from the same comparison. This violates the tracked-mutation rejection contract and permits reviewing/testing altered source as the pinned HEAD (ARCH-SECURE, ARCH-PURPOSE). Compare actual tracked contents, file types, and modes independently of Git filters and index suppression flags; add regression coverage for these hiding mechanisms.
```

3. **Important findings**

None.

4. **Minor findings**

None.

5. **Test coverage notes**

- Reviewprepare, gatestate, reviewbox, judge, and vocabulary package tests passed.
- Broader boundary/preparation/sidecar/dispute tests passed in 153.6 seconds.
- The temporary production-dispatch regression failed as expected, demonstrating the uncovered bug. Temporary files were removed; repository status remains clean.
- Live Seatbelt conformance **skipped** because `sandbox_apply` was refused. Direct `/bin/ps` was also denied here. Live downstream and installed-agent checks were not rerun; the full repository suite was not rerun.

6. **Architectural notes**

- **ARCH-DRY — pass:** shared attempt loop, runner, parsers, and ledger authority.
- **ARCH-PURE — pass:** parsing, admission, and rendering remain independent of IO.
- **ARCH-PURPOSE — flag:** tracked-source preservation is incompletely enforced.
- **ARCH-MOCK — pass:** stateful temporary repositories and executable fixtures exercise production seams; live-platform verification has the limitation above.
- **ARCH-CONSTRAINTS — pass:** bounded attempts, preparation timeout, output capture, and diagnostic envelopes.
- **ARCH-SECURE — flag:** mutable Git reporting settings can conceal changed review inputs.
- **ARCH-ORDER — pass:** reset, confinement preparation, build preparation, review, and admission are explicitly sequenced.
- **ARCH-FUNERAL — pass:** generated files share checkout cleanup; diagnostic retention and per-dispatch growth are documented and tested.

7. **Plan revision recommendations**

Add a `## Revisions` entry specifying that tracked-state preservation compares actual file contents and modes independently of Git normalization and suppression settings. Extend the verification strategy to cover assume-unchanged, skip-worktree, and clean-filter cases.
