# Boundary Review — ariadne#524 (whole-issue close)

| field | value |
|-------|-------|
| issue | 524 — Scoped future evidence and exact-package flake reruns |
| repo | ariadne |
| issue file | workshop/issues/000524-scoped-evidence-reruns.md |
| boundary | whole-issue close |
| milestone | — |
| window | 63ebb3385af267d64084921c6e8d1150a188c734..30b4964360436b238aadaed6ce5a7f261622952e |
| command | sdlc close --issue 524 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-21T08:25:20-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned range fulfills the scoped-reader and exact-package rerun contracts. All production evidence consumers use the new snapshot, regression controls detect the prior defects, and README/atlas cover the new surfaces and rollout barrier. No new findings.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Parent-index removal in a disposable pinned-head clone makes TestScopedSavedProofConsumers and TestScopedLinkedProofUsesParentCheck fail. Both pass on the reviewed implementation.
  - id: BR-2
    disposition: addressed
    note: |
      Ran scripts/verify-pre524-reader.sh successfully against pinned base 63ebb338: the retained fixture demonstrates global warnings, saved-proof incompleteness, and publication refusal in the actual old reader.
  - id: BR-3
    disposition: addressed
    note: |
      README.md:768 documents the compatibility command and fleet rollout barrier; scripts/verify-pre524-reader.sh implements the described pinned-source control.
  - id: BR-4
    disposition: addressed
    note: |
      README.md:771 and atlas/workflow/sdlc-binary.md:2653 assign ownership and removal of the printed private clone directory after inspection. The script creates that documented directory.
  - id: BR-5
    disposition: addressed
    note: |
      Real Go fixtures pass on the reviewed code. Independently restoring restrictive incidental-name validation or inferred slash-ancestor requirements makes the corresponding production-gate regression fail.
```

1. **Strengths**

   - Strict identity decoding preserves opaque future bytes while keeping malformed evidence visibly incomplete (`cmd/sdlc/teststats_evidence.go:152`).
   - Parent indexing consistently handles both cross-check links and inverse parent scope (`cmd/sdlc/teststats_evidence.go:54`).
   - Exact-package commands preserve whole-parent execution and require package-bound proof (`cmd/sdlc/internal/flakes/rerun.go:17`).
   - Batch execution reuses existing ownership and cleanup seams, stops after failure/cancellation, and shares one aggregate timeout (`cmd/sdlc/localcheck_run.go:530`).

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   Independently passed:

   - Flakes package tests and race tests.
   - Focused reader, authority, batch, and real-package regressions.
   - Broader known-flake, statistics, actuals, weekly, and runner regressions.
   - Scoped-reader and managed-batch race checks.
   - Private production-CLI conformance and real old-reader compatibility control.
   - `git diff --check`.

   Disposable mutation controls failed as expected and were removed afterward. No reviewed source was changed.

   GitHub independently reports `sdlc-local/suite: success` on `2750b1ed88b29c30a74bb781578b2ba59150f7c1`; the subsequent range to reviewed HEAD changes only issue/review/plan documents. I did not rerun the full suite.

6. **Architectural notes**

   | Principle | Result |
   |---|---|
   | ARCH-DRY | Pass — shared snapshot projections and batch execution seam. |
   | ARCH-PURE | Pass — identity, selection, and PASS logic remain directly testable without IO. |
   | ARCH-PURPOSE | Pass — all production reader consumers migrated; automatic reruns use exact-package batches. |
   | ARCH-MOCK | Pass — existing injected dependencies, stateful fake GitHub, and private real-process fixtures exercise production paths. |
   | ARCH-CONSTRAINTS | Pass — bounded reads and one aggregate timeout; counted scan/group tests cover growth. |
   | ARCH-SECURE | Pass — strict envelope validation, literal argv, and command diagnostics separated from proof logs. |
   | ARCH-ORDER | Pass — sequential process ownership, fail-fast behavior, and controlled cancellation tests. |
   | ARCH-FUNERAL | Pass — existing attempt lifecycle retained; private verification clones have documented removal ownership. |
   | ARCH-EVOLVE | Pass — unsupported bodies remain opaque and untouched; malformed identities remain incomplete. |

   Fleet binary path/revision verification remains the documented prerequisite for later #463 activation.

7. **Plan revision recommendations:** None. The concept table and completed implementation agree.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"f6195487-8476-4457-9df5-d344d0e52dbb","repository":"github.com/xianxu/ariadne","issue_id":"000524","recipe":"milestone-review","round":6,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/2818c18fef246e1b","reviewed_head":"30b4964360436b238aadaed6ce5a7f261622952e","slot":"ariadne:3","requested_at":"2026-10-21T15:20:26.962733Z"},"phase":"finished","pid":64272,"started_at":"2026-10-21T15:20:27.134987Z","finished_at":"2026-10-21T15:25:19.784501Z","session_id":"01a15f0c-a905-7ad0-a8bc-085923af76aa","outcome":"returned"}] -->
