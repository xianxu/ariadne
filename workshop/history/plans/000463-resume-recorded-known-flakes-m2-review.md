# Boundary Review — ariadne#463 (milestone M2)

| field | value |
|-------|-------|
| issue | 463 — local-check: resume recorded known-flake failures with targeted tests |
| repo | ariadne |
| issue file | workshop/issues/000463-resume-recorded-known-flakes.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 7e27fef542a65b60b574c8836bc0d520b346361b..615fcd68453a0f37523fede1f6108f15ad6a4f59 |
| command | sdlc milestone-close --issue 463 --milestone M2 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-21T11:41:15-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The resume and reconciliation flows satisfy the main acceptance checks, including real targeted execution, immutable original evidence, current-policy validation, and refusal of later green results after interruption. One reproduced lifecycle defect blocks this boundary: a crash after child publication permanently strands the locked temporary checkout.

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: durable-staging-cleanup
    title: |
      Completed-child crashes permanently strand locked resume worktrees
    detail: |
      cmd/sdlc/localcheck_resume.go:113 relies on deferred worktree retirement, but child publication at line 196 calls publishSuiteAttempt, which deletes the inflight owner binding (teststats_store.go:426). Killing the CLI during status publication therefore leaves a locked checkout without that binding. Repeated resume returns through line 89 without cleanup; reconcile rejects committed children at line 237; the stale-worktree sweeper skips locked inputs. A temporary production-CLI regression reproduced this: after killing at the post barrier, successful repeated resume and a sweep with the checkout aged 48 hours still retained it. ARCH-FUNERAL: this is the 2nd finding in family durable-staging-cleanup. Do NOT fix only this instance: state and enforce the rule that cleanup authority outlives every protected artifact, enumerate pre-link, pending, committed-child and reconciled windows, and provide crash-safe retirement for each without granting another execution or pass.
```

1. **Strengths**

   - Resume reuses the existing exact-package batch and flake classifier; the private CLI test confirmed one original suite execution and one targeted rerun per required test.
   - Saved reuse, explicit republish, and status retries resolve current linked evidence. Tests cover registry changes, altered logs, conflicting children, and pending sibling parents.
   - Original record/log bytes and full-suite duration remain intact. README, help, and atlas document the new commands and conservative interruption behavior.

2. **Critical findings**

   None.

3. **Important findings**

   The lifecycle finding above applies at [localcheck_resume.go:113](/Volumes/TestScratch/review-checkouts/8e0428314e1b709d/cmd/sdlc/localcheck_resume.go:113). Preserve discoverable cleanup authority until retirement is confirmed, and make recovery handle completed-child residue without changing test proof. Add deterministic crash barriers across the complete artifact lifecycle.

4. **Minor findings**

   None.

5. **Test coverage notes**

   - Targeted resume/saved-attempt/republish/guard tests passed: **124.938s**.
   - Opt-in production CLI conformance passed both completion and crash/reconcile scenarios: **18.761s**.
   - Temporary overlay test `TestReviewResumeCompletedCrashLeaksWorktree` failed as expected: **11.042s**, demonstrating the stranded locked checkout.
   - `git diff --check` passed; the checkout remained unchanged. No full suite was run.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared classifier, batch executor, and linked-proof resolver.
   - **ARCH-PURE — pass:** proof and transition tests use deterministic inputs without IO; orchestration stays in integration code.
   - **ARCH-PURPOSE — pass:** saved failure recovery executes targeted tests and reaches real status publication; contextual success readers share validation.
   - **ARCH-MOCK — pass:** private Git/tracker state and the existing GH seam exercise production command paths.
   - **ARCH-CONSTRAINTS — pass:** existing limits, aggregate timeout, weight-1 lease, and counted linear projection tests remain enforced.
   - **ARCH-SECURE — pass:** evidence identities/digests, owner incarnation, authority, and cleanup bindings receive explicit validation.
   - **ARCH-ORDER — pass for execution authority:** immutable reservation and pure transitions prevent duplicate execution; uncertain completion cannot become a pass.
   - **ARCH-FUNERAL — flag:** cleanup authority disappears before the protected checkout retires.
   - **ARCH-EVOLVE — pass:** original records remain immutable; resume variants retain raw unknown fields and strict known-field validation.

7. **Plan revision recommendations**

   Append a `## Revisions` entry defining protected-environment retirement across pre-link, pending, committed-child, and reconciled states. Include the environment integration entity, its durable cleanup authority, and deterministic crash/retry tests. The current terminal-receipt cleanup path does not cover the complete lifecycle.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"a6f9d8aa-9cb7-4bd2-bd2c-9abd8f33c9bd","repository":"github.com/xianxu/ariadne","issue_id":"000463","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/8e0428314e1b709d","reviewed_head":"615fcd68453a0f37523fede1f6108f15ad6a4f59","slot":"ariadne:3","requested_at":"2026-10-21T18:37:50.925332Z"},"phase":"finished","pid":64878,"started_at":"2026-10-21T18:37:51.032273Z","finished_at":"2026-10-21T18:41:14.718736Z","session_id":"01a15fc1-6160-7231-9013-f4cb22100432","outcome":"returned"}] -->

---

## Re-review — 2026-10-21T12:00:53-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 463 — local-check: resume recorded known-flake failures with targeted tests |
| repo | ariadne |
| issue file | workshop/issues/000463-resume-recorded-known-flakes.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 7e27fef542a65b60b574c8836bc0d520b346361b..fdc51f18f686c205e282a6470b8d8079c31874a1 |
| command | sdlc milestone-close --issue 463 --milestone M2 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-21T12:00:53-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The targeted resume flow, proof validation, and original-evidence preservation work. BR-3’s original status-post crash window is fixed and has a meaningful regression test. However, cleanup retries still discard unconfirmed durability obligations, so the broader cleanup-authority invariant remains incomplete.

```findings
dispose:
  - id: BR-3
    disposition: not-addressed
    note: |
      The original window is fixed: localcheck_resume.go:192 retires inputs before child publication; TestResumeCrashCommittedChildStatusPost passes and fails when retirement is moved back to a defer. The durable-staging-cleanup rule remains incomplete at localcheck_resume_environment.go:302-305: after worktree removal succeeds but its reply is lost, or registration-directory sync fails at lines 475-482, retry observes missing registration and releases the owner marker without syncing the registration namespace. Likewise, failed owner-release sync at line 506 is bypassed by the missing-marker return at lines 283-286. A disposable fault-injection test exercising production cleanup failed all three cases: retry returned nil without the required directory sync. ARCH-FUNERAL/ARCH-ORDER: visible absence is not durable retirement; a subsequent crash can restore locked registration after its cleanup binding was durably removed. Preserve and discharge each uncertain deletion's sync obligation before releasing authority. This remains BR-3 in family durable-staging-cleanup, rather than a new finding.
```

1. **Strengths**

   - Child publication now follows confirmed environment retirement at [localcheck_resume.go:192](/Volumes/TestScratch/review-checkouts/6b6f15e151e8280b/cmd/sdlc/localcheck_resume.go:192).
   - Private CLI tests verify exact-package execution, immutable original bytes, repeat without execution, and terminal reconciliation.
   - Saved reuse, republish, and status retries share current proof validation; pending sibling evidence blocks aggregate success.
   - README, help, and atlas cover the commands and recovery limitations.

2. **Critical findings**

   None.

3. **Important findings**

   **BR-3 remains open — cleanup retry skips durability confirmation.** At [localcheck_resume_environment.go:302](/Volumes/TestScratch/review-checkouts/6b6f15e151e8280b/cmd/sdlc/localcheck_resume_environment.go:302), missing registration immediately permits owner release. This bypasses a previously failed or never-attempted registration-directory sync. Missing-marker retries similarly bypass owner-release sync.

   Preserve enough validated cleanup context to sync the surviving directory that owns each absent entry before reporting completion or releasing authority. Cover lost replies, failed syncs, and repeated retries across the entire deletion sequence.

4. **Minor findings**

   None.

5. **Test coverage notes**

   - Resume tests passed: **17.188s**.
   - Opt-in private CLI/crash tests passed: **160.080s**.
   - Legacy saved-proof/republish/flake checks passed: **59.699s**.
   - Selected resume race tests passed: **15.790s**.
   - Moving retirement back to deferred cleanup made the status-post regression fail with a retained locked checkout.
   - Disposable retry-durability checks failed for lost removal reply, registration-sync failure, and owner-release-sync failure.
   - `git diff --check` passed. No full suite run. Disposable files were removed; pinned checkout unchanged.

6. **Architectural notes**

   | Principle | Result |
   | --- | --- |
   | ARCH-DRY | Pass — shared classifier, executor, resolver, and cleanup core. |
   | ARCH-PURE | Pass — deterministic proof/transition core; IO stays in integration code. |
   | ARCH-PURPOSE | Pass for targeted recovery and success-reader integration; lifecycle completion remains under BR-3. |
   | ARCH-MOCK | Pass — private stateful Git/tracker/GH fixtures exercise production paths. |
   | ARCH-CONSTRAINTS | Pass — bounded evidence, aggregate timeout, one lease, and linear-work checks. |
   | ARCH-SECURE | Pass — explicit identity, digest, incarnation, authority, and reverse-binding checks. |
   | ARCH-ORDER | Flag — retry converts observed absence into confirmed durable cleanup. |
   | ARCH-FUNERAL | Flag — cleanup authority can disappear before retirement is durable. |
   | ARCH-EVOLVE | Pass — immutable raw evidence preserves unknown fields; known fields remain strict. |

   The revised entity tables match the implemented pure and integration seams.

7. **Plan revision recommendations**

   Append a `## Revisions` entry defining retry obligations for every unlink/removal and directory-sync pair, including already-absent entries. Require deterministic cleanup fault tests asserting durability before authority release.

   The existing approved revision explicitly requires returning to the TL/operator for another lifecycle gap; this remaining BR-3 defect meets that condition.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"9d6e8b46-3158-471a-a3df-cc236f8d0879","repository":"github.com/xianxu/ariadne","issue_id":"000463","recipe":"milestone-review","round":2,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/6b6f15e151e8280b","reviewed_head":"fdc51f18f686c205e282a6470b8d8079c31874a1","slot":"ariadne:3","requested_at":"2026-10-21T18:56:29.529913Z"},"phase":"finished","pid":15149,"started_at":"2026-10-21T18:56:29.768124Z","finished_at":"2026-10-21T19:00:52.633639Z","session_id":"01a15fd2-7482-7081-8ccb-d3ae453789e0","outcome":"returned"}] -->
