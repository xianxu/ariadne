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
