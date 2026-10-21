# Boundary Review — ariadne#463 (milestone M3)

| field | value |
|-------|-------|
| issue | 463 — local-check: resume recorded known-flake failures with targeted tests |
| repo | ariadne |
| issue file | workshop/issues/000463-resume-recorded-known-flakes.md |
| boundary | milestone M3 |
| milestone | M3 |
| window | 1902b643cb56a439f164c7e55b1bcbc50910bb3d..3a4007c6663cf47c5969e0b223b2db4727956ba7 |
| command | sdlc milestone-close --issue 463 --milestone M3 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-21T12:34:52-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The targeted resume flow, shared proof validation, evidence-only reconciliation and permanent interrupted-evidence retention are well covered. All review test runs passed. One Important cancellation gap remains: cleanup can delete the surviving owner marker after cancellation, contrary to M3’s stop-before-subsequent-effects contract.

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: durable-staging-cleanup
    title: |
      Cancellation during registration retirement still permits owner-marker deletion
    detail: |
      cmd/sdlc/localcheck_resume_environment.go:390 calls releaseResumeEnvironmentOwner after registration removal and directory sync without rechecking ctx. If cancellation occurs during either operation and it returns success, authorize does not inspect ctx, so lines 427-430 still delete and sync the surviving owner marker. The caller subsequently reports cancellation before child publication, but the marker has already been removed. ARCH-ORDER/ARCH-FUNERAL: M3 requires cancellation to stop subsequent cleanup effects and preserve surviving inputs. This is the 3rd finding in family durable-staging-cleanup. Do NOT fix only this instance: state and sweep the rule that every destructive cleanup effect requires current, uncancelled authority. Add deterministic cancellation injection at the remove/sync boundaries and assert no subsequent mutation. Follow the plan's lifecycle-gap escalation requirement.
```

1. **Strengths**

   - Resume uses the existing exact-package batch and flake classifier, preserving original suite evidence.
   - Saved reuse and status retries validate every relevant same-head parent through the shared resolver.
   - Reconciliation and retention tests assert preserved bytes, locks and execution counts through private production CLI fixtures.

2. **Critical findings:** None.

3. **Important findings:** The cancellation gap above. Pass context through the final release boundary and enforce the cancellation rule across cleanup effects.

4. **Minor findings:** None.

5. **Test coverage notes**

   Passed focused resume/saved-proof tests, resume race tests, legacy known-flake/batch checks, private CLI crash/conformance tests, and `git diff --check`. The missing regression is cancellation during an otherwise successful cleanup effect; existing fault tests inject errors.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared classifier and proof resolver.
   - **ARCH-PURE — pass:** proof decisions remain directly testable without IO.
   - **ARCH-PURPOSE — pass:** targeted-only execution and same-head proof fencing delivered.
   - **ARCH-MOCK — pass:** private stateful tracker/GH fixtures and real Git seams exercised.
   - **ARCH-CONSTRAINTS — pass:** bounded execution and linear-growth checks retained.
   - **ARCH-SECURE — pass:** strict evidence validation and exact cleanup bindings.
   - **ARCH-ORDER — flag:** cancellation can permit a subsequent destructive effect.
   - **ARCH-FUNERAL — flag:** surviving ownership metadata can be removed after cancellation.
   - **ARCH-EVOLVE — pass:** immutable raw evidence preserves unknown fields.

   README, embedded help and atlas updates cover the new surface.

7. **Plan revision recommendations**

   Add a `## Revisions` entry specifying cancellation checks at every cleanup-effect boundary and deterministic cancellation tests, including registration retirement → owner-marker release. Preserve M2’s superseded status and follow the documented lifecycle-gap escalation.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"875a7be6-4c8d-44a0-8eba-9587cc427d74","repository":"github.com/xianxu/ariadne","issue_id":"000463","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/81ab78122cb504fe","reviewed_head":"3a4007c6663cf47c5969e0b223b2db4727956ba7","slot":"ariadne:3","requested_at":"2026-10-21T19:30:40.550044Z"},"phase":"finished","pid":1672,"started_at":"2026-10-21T19:30:40.645622Z","finished_at":"2026-10-21T19:34:51.498973Z","session_id":"01a15ff1-be8b-70a1-8622-942646e5c83e","outcome":"returned"}] -->
