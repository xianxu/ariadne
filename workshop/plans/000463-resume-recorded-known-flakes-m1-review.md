# Boundary Review — ariadne#463 (milestone M1)

| field | value |
|-------|-------|
| issue | 463 — local-check: resume recorded known-flake failures with targeted tests |
| repo | ariadne |
| issue file | workshop/issues/000463-resume-recorded-known-flakes.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 6fd7dc750557b3f51bdefc5859ecd8f76027a9ab..3817e5ce721cb875a1b7d154ff93329516268551 |
| command | sdlc milestone-close --issue 463 --milestone M1 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-21T11:08:34-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

M1 delivers a well-separated proof model, strict decoding, and carefully tested locking and retirement ordering. Two verified gaps block completion: reconciliation receipts cannot pass through the publisher, and interrupted publication leaves temporary files without cleanup. The checkout remains unchanged.

```findings
findings:
  - id: new
    severity: Critical
    kind: defect
    family: proof-publication-reachability
    title: |
      The publisher rejects every reconciliation receipt with an existing link
    detail: |
      cmd/sdlc/localcheck_resume_store.go:239 calls s.proofError before reaching receipt publication. Every valid receipt requires an existing resume link, but teststats_evidence.go:177 rejects that linked parent, making the receipt branch unreachable. A temporary regression publishing a link followed by its valid receipt fails with “immutable execution fence requires validation.” Use publication-specific validation that permits the required link while preserving unsupported/conflicting-evidence checks, and add a regression through the real publisher. This blocks the promised terminal recovery path (ARCH-PURPOSE).
  - id: new
    severity: Important
    kind: defect
    family: durable-staging-cleanup
    title: |
      Crashed publication leaves .resume-* files permanently
    detail: |
      cmd/sdlc/localcheck_resume_store.go:281 creates .resume-* files and relies exclusively on deferred removal. Process death bypasses that defer; teststats_store.go:588 recognizes only .attempt-* staging files and then skips non-JSON files. A temporary regression confirms that an abandoned 40-day-old .resume-* file survives production pruning. Add cleanup under the directory mutex, which excludes active publishers, and test interrupted publication cleanup. These files also escape the JSON record/byte capacity accounting (ARCH-FUNERAL).
```

### 1. Strengths

- The proof resolver and transition function remain pure; their tests require no filesystem, network, or subprocess mocks.
- Strict decoding covers missing/null fields, duplicate keys, filename identity, nested owner fields, and opaque future fields.
- Real subprocess barriers exercise the inverse lock ordering. Retirement fault tests check parent-first durability and terminal-witness retention.
- Atlas changes accurately distinguish M1’s core from M2’s pending CLI activation.

### 2. Critical findings

- **Receipt publication is unreachable** at `cmd/sdlc/localcheck_resume_store.go:239`. Replace the success-reader guard with validation appropriate to publication; retain the existing binding and conflict checks.

### 3. Important findings

- **Abandoned staging files have no lifecycle end** at `cmd/sdlc/localcheck_resume_store.go:281`. Extend pruning to collect inactive `.resume-*` files safely under the directory lock.

### 4. Minor findings

None.

### 5. Test coverage notes

Validation completed:

- Focused resume tests: passed.
- Focused resume tests under race detection: passed.
- Broader `Resume|SuiteAttempt|SuiteEvidence` selection: passed, 104.093s.
- Pinned-range `git diff --check`: passed.

Both temporary overlay regressions failed as described above. Existing receipt tests decode receipts or write fixtures directly; they never exercise successful receipt publication, explaining the green suite despite the unreachable branch. Temporary review files were removed.

### 6. Architectural notes for upcoming work

| Principle | Result |
|---|---|
| **ARCH-DRY** | Pass: reuses the existing classifier, revalidation, JSON validation, and flock seam. |
| **ARCH-PURE** | Pass: proof/transition logic is separated from storage effects. |
| **ARCH-PURPOSE** | **Flag:** terminal receipt publication does not work through the delivered seam. |
| **ARCH-MOCK** | Pass for M1: private filesystem fixtures, injected retention effects, and controlled subprocesses exercise the actual boundary. |
| **ARCH-CONSTRAINTS** | Pass for reviewed M1 behavior: bounded snapshot/record reads and linear index tests are present; staging residue is covered separately below. |
| **ARCH-SECURE** | Pass: typed validation, digest binding, inode checks, and private test roots are present; no new credentials. |
| **ARCH-ORDER** | Pass within M1 scope: explicit transitions and controlled lock interleavings are tested. M2 must route production execution through the model. |
| **ARCH-FUNERAL** | **Flag:** `.resume-*` staging files lack crash cleanup. |
| **ARCH-EVOLVE** | Pass: immutable raw records preserve unknown bytes; known fields remain strict, with an explicit safety rationale for the v3 boundary. |

The current entity paths and PURE/INTEGRATION split match the revised M1 scope. Atlas documentation is updated. No README update is required yet because this range introduces no runnable user-facing command or configuration.

### 7. Plan revision recommendations

Add a `## Revisions` entry defining the staging-file lifecycle: publication creates `.resume-*`; normal completion removes them; pruning under the directory mutex removes crash leftovers. Include the interrupted-publication regression in M1’s retention coverage.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"ffcd2dfa-1b1b-408e-899a-52af58613306","repository":"github.com/xianxu/ariadne","issue_id":"000463","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/035c66f5d43c7223","reviewed_head":"3817e5ce721cb875a1b7d154ff93329516268551","slot":"ariadne:3","requested_at":"2026-10-21T18:05:30.678299Z"},"phase":"finished","pid":49121,"started_at":"2026-10-21T18:05:30.777277Z","finished_at":"2026-10-21T18:08:33.95971Z","session_id":"01a15fa3-c627-7d42-aa18-a489ef653585","outcome":"returned"}] -->

---

## Re-review — 2026-10-21T11:17:22-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 463 — local-check: resume recorded known-flake failures with targeted tests |
| repo | ariadne |
| issue file | workshop/issues/000463-resume-recorded-known-flakes.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 6fd7dc750557b3f51bdefc5859ecd8f76027a9ab..6caff439fe18afe6ac4ce3920f5e009137effa53 |
| command | sdlc milestone-close --issue 463 --milestone M1 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-21T11:17:22-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned range satisfies the revised M1 scope. Both prior findings are addressed with regression tests that fail when their fixes are removed. No new blocking findings emerged. CLI activation and composite success-reader integration remain explicitly assigned to M2.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      localcheck_resume_store.go:249 now checks unsupported authority separately before validating the supported graph. TestResumeStorePublishesTerminalReceipt passes on head and fails with “immutable execution fence requires validation” when the old guard is restored.
  - id: BR-2
    disposition: addressed
    note: |
      pruneResumeStaging runs under the directory mutex during publication and pruning. TestResumeStoreStagingLifetime and TestResumeStoreProcessStagingCrashCleanup pass on head; both fail with retained crash residue when cleanup is removed. The disposable checkout was removed.
```

1. **Strengths**

   - Strict decoding rejects missing/null fields, duplicate keys, invalid identities and filename mismatches (`cmd/sdlc/localcheck_resume_evidence.go:111`).
   - Publication preserves raw bytes and uses exclusive hard-link publication (`cmd/sdlc/localcheck_resume_store.go:294`).
   - Retirement durably removes the selectable parent before its fence (`cmd/sdlc/localcheck_resume_retention.go:287`).
   - Real subprocess barriers exercise lock contention and publication/pruning ordering.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   Focused resume/evidence/attempt tests passed (6.050s), resume race tests passed (11.014s), and expanded suite/statistics compatibility tests passed (66.790s). Both prior fixes passed negative-control mutation checks. `git diff --check` passed; the reviewed checkout remains unchanged. The full repository suite was not run.

6. **Architectural notes for upcoming work**

   | Principle | M1 assessment |
   |---|---|
   | ARCH-DRY | Pass — reuses the classifier, strict JSON machinery and shared snapshot. |
   | ARCH-PURE | Pass — proof and transition logic have direct tests without IO. |
   | ARCH-PURPOSE | Pass — current M1 deliverables exist; execution and reader integration are explicitly M2. |
   | ARCH-MOCK | Pass — private filesystem/process fixtures and injected retention effects exercise the actual seams. |
   | ARCH-CONSTRAINTS | Pass — bounded reads, capacity refusal and linear index checks are present. |
   | ARCH-SECURE | Pass — typed validation, digest checks and inode verification guard authority. |
   | ARCH-ORDER | Pass — explicit transitions, deterministic interleavings and retirement fault tests cover M1 sequencing. |
   | ARCH-FUNERAL | Pass — staging now has crash cleanup; terminal bundles have ordered retirement. |
   | ARCH-EVOLVE | Pass — immutable publication preserves unknown bytes; version refusal has a stated authority-safety rationale. |

   The revised entity tables match M1’s paths and responsibilities. Atlas documents the new surface. No README change is required because this range introduces no user-facing command, flag or configuration. M2 must route execution through the transition model and all success consumers through the shared resolver.

7. **Plan revision recommendations:** None; the appended revisions describe the implemented corrections.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"ed78b1e7-f8aa-4ec2-9329-f7ae0aa8e741","repository":"github.com/xianxu/ariadne","issue_id":"000463","recipe":"milestone-review","round":2,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/2d752b076d705416","reviewed_head":"6caff439fe18afe6ac4ce3920f5e009137effa53","slot":"ariadne:3","requested_at":"2026-10-21T18:13:55.953519Z"},"phase":"finished","pid":92559,"started_at":"2026-10-21T18:13:56.048213Z","finished_at":"2026-10-21T18:17:21.535553Z","session_id":"01a15fab-7be6-7393-8ebd-cff2c80cf1cb","outcome":"returned"}] -->
