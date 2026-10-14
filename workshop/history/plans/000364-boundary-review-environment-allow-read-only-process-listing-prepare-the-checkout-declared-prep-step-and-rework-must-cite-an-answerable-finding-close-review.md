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

---

## Re-review — 2026-10-14T01:28:26-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 364 — Boundary review environment: allow read-only process listing, prepare the checkout (declared prep step), and REWORK must cite an answerable finding |
| repo | ariadne |
| issue file | workshop/issues/000364-boundary-review-environment-allow-read-only-process-listing-prepare-the-checkout-declared-prep-step-and-rework-must-cite-an-answerable-finding.md |
| boundary | whole-issue close |
| milestone | — |
| window | 85135ce32b676c60628649a96c93446ac238a11c..8d30689ac48e03acf61382fb2617c0856a04ccd6 |
| command | sdlc close --issue 364 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-14T01:28:26-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

BR-1 is addressed with independently verified regression evidence. Preparation, bounded retries, response admission, and documentation are implemented coherently. One Important gap remains: the plan promises a production lifecycle regression for disputed unverifiable findings, but the checked-in tests exercise that requirement only through the pure validator.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      prepare.go:93 fingerprints actual filesystem state. Restoring the previous guard in a disposable pinned-head checkout made production Dispatch regressions fail for assume-unchanged, skip-worktree, clean filters, and permission changes; the corrected package passes. The disposable checkout was removed.
findings:
  - id: new
    severity: Important
    family: boundary-lifecycle-coverage
    title: |
      The promised unverifiable-dispute lifecycle regression is missing
    detail: |
      workshop/plans/000364-review-environment-plan.md:198 promises a real temporary-repository lifecycle proving host evidence reaches the next judge, invalid replies cannot advance ledger/head, and a valid disposition closes the loop. boundaryresponse_test.go:44 manually constructs a disputed ledger and calls the pure validator; existing dispute lifecycle fixtures use ordinary findings. Add a production close/dispatch regression that disputes an unverifiable finding, checks the reviewer receives its evidence, rejects unanswered or note-less responses without advancement, and accepts an explicit evidence-answering disposition. Prove it fails when the dispute admission guard or ledger wiring is removed (ARCH-PURPOSE).
```

1. **Strengths**

   - The tracked-state guard handles Git suppression flags, normalization, file modes, symlinks, and gitlinks.
   - Preparation and review share process cleanup, confinement, and one bounded attempt loop.
   - Diagnostics remain separate from semantic output and have enforced retention limits.
   - README, atlas, help, and prompt contracts document the new surface.

2. **Critical findings:** None remaining.

3. **Important findings:** The lifecycle coverage gap above. Extend the existing temporary-repository fixtures rather than adding another harness.

4. **Minor findings:** None.

5. **Test coverage notes:** Judge, reviewprepare, gatestate, reviewbox, and vocab suites passed. Targeted boundary, preparation, sidecar, admission, and dispute tests passed in 83.6 seconds. BR-1’s regression failed with the previous guard restored. Live Darwin conformance skipped because the parent sandbox rejects `sandbox_apply`; downstream and installed-agent opt-in checks were not run.

6. **Architectural notes**

   - **ARCH-DRY — pass:** Shared runner, retry loop, and ledger authority.
   - **ARCH-PURE — pass:** Parsing, rendering, and admission consume values without IO.
   - **ARCH-PURPOSE — flag:** Complete the promised dispute lifecycle proof.
   - **ARCH-MOCK — pass:** Stateful temporary repositories and executable fixtures exercise production seams.
   - **ARCH-CONSTRAINTS — pass:** Attempts, deadlines, output, and shutdown are bounded.
   - **ARCH-SECURE — pass:** Pinned regular declarations and direct filesystem fingerprints protect review inputs.
   - **ARCH-ORDER — pass:** Reset, confinement, preparation, and admission ordering is explicit and tested.
   - **ARCH-FUNERAL — pass:** Checkout cleanup owns generated assets; permanent diagnostic retention has an explicit per-dispatch cost.

7. **Plan revision recommendations:** Append a `## Revisions` entry recording the missing lifecycle proof and its regression strategy; mark that verification complete only after the production test and mutation check pass.

---

## Re-review — 2026-10-14T01:38:16-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 364 — Boundary review environment: allow read-only process listing, prepare the checkout (declared prep step), and REWORK must cite an answerable finding |
| repo | ariadne |
| issue file | workshop/issues/000364-boundary-review-environment-allow-read-only-process-listing-prepare-the-checkout-declared-prep-step-and-rework-must-cite-an-answerable-finding.md |
| boundary | whole-issue close |
| milestone | — |
| window | 85135ce32b676c60628649a96c93446ac238a11c..37c3e27aac67f707fcc2cc06ad8f6f4dd9a452e9 |
| command | sdlc close --issue 364 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-14T01:38:16-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned implementation matches the issue’s preparation, process-listing, and answerable-review contracts. BR-2 is addressed: the production lifecycle regression passes, and disabling the dispute-answer guard makes both invalid-response cases fail by incorrectly finalizing the issue. No new blocking findings emerged.

```findings
dispose:
  - id: BR-2
    disposition: addressed
    note: |
      cmd/sdlc/review_unverifiable_lifecycle_test.go:16 exercises production close/dispute/review, verifies host evidence reaches the reviewer, rejects unanswered and note-less replies without advancement, and accepts an answering withdrawal. It passes at pinned HEAD; disabling boundaryresponse.go's dispute-answer guard makes both rejection cases fail. The disposable mutation checkout was removed.
  - id: BR-1
    disposition: addressed
    note: |
      The filesystem fingerprint remains in cmd/sdlc/internal/judge/prepare.go, independently checking tracked bytes, modes, types, and symlink targets. The suppression, normalization, and path-mutation regressions pass; the prior disposition remains supported.
```

1. **Strengths**

   - Preparation and admission reuse one bounded attempt loop, with reset and preparation before retry (`cmd/sdlc/internal/judge/dispatch.go:336`).
   - Configuration comes from pinned regular Git blobs; parsing rejects ambiguous declarations (`cmd/sdlc/reviewprepare.go:17`).
   - The new lifecycle test checks durable outcomes through production commands, including ledger, issue, and reviewed-ref identity (`cmd/sdlc/review_unverifiable_lifecycle_test.go:16`).
   - README, atlas, help, and generated prompt contracts document the new surface and limits.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   Passed the judge, reviewprepare, reviewbox, gatestate, and vocabulary packages, plus targeted admission, pinned-reader, preparation, diagnostic-retention, and dispute-lifecycle tests. The independent guard mutation failed as expected.

   Live Seatbelt tests skipped because this environment rejects `sandbox_apply`. Pair-generation and installed-Codex checks are opt-in and were not independently rerun. The full suite was not rerun during this review.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared attempt loop, runner, and ledger rules.
   - **ARCH-PURE — pass:** declaration parsing, admission, and rendering remain pure; declared entities exist at their stated paths.
   - **ARCH-PURPOSE — pass:** production lifecycle coverage now completes the promised evidence-answering loop.
   - **ARCH-MOCK — pass:** temporary repositories and controlled subprocess seams exercise production boundaries; live conformance hooks exist.
   - **ARCH-CONSTRAINTS — pass:** deadlines, two-attempt budget, output caps, and descendant cleanup are enforced and tested.
   - **ARCH-SECURE — pass:** pinned configuration, direct argv, narrow executable exception, and separate diagnostic channel.
   - **ARCH-ORDER — pass:** retry sequencing and rejected-response non-advancement are exercised through production paths.
   - **ARCH-FUNERAL — pass:** generated assets follow checkout cleanup; permanent diagnostic retention and per-dispatch cost are explicit.

7. **Plan revision recommendations:** None. Remaining close and landing checkboxes appropriately remain pending.
