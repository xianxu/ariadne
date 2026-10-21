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
