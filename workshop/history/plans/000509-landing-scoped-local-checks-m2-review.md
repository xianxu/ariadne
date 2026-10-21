# Boundary Review — ariadne#509 (milestone M2)

| field | value |
|-------|-------|
| issue | 509 — Landing scoped local checks |
| repo | ariadne |
| issue file | workshop/issues/000509-landing-scoped-local-checks.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | e32179d2eb0ca49c42733e792334656339ca9605..d0e43f42fa53f1bf7d94137baf3ba18988103526 |
| command | sdlc milestone-close --issue 509 --milestone M2 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-21T00:36:46-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned M2 range implements the shared post-suite gate across trains, tracked single landings, and legacy merges. Selection, owner-bound evidence, terminal failure handling, and manual burn-in restrictions match the documented plan. No blocking defects found; targeted tests passed. Consumer activation and final delivery verification remain outside this milestone.

```findings
{}
```

1. **Strengths**

   - The train model explicitly requires checking before CI/merge and rejects wrong-prefix completions (`cmd/sdlc/internal/landingline/trainstate.go:253`).
   - One shared gate binds successful evidence to policy, included cars, head, and retained-log hash (`cmd/sdlc/landingchecks_gate.go:70`).
   - Launch intent prevents automatic retries after failed or unobserved completion (`cmd/sdlc/landingchecks_gate.go:107`).
   - Production tests exercise trains, single landings, good prefixes, disabled/unrelated checks, failure, legacy merges, and `--no-ci` (`cmd/sdlc/landingchecks_train_test.go:13`).
   - README, CLI help, and atlas describe the delivered surface.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   Passed:

   - `go test ./cmd/sdlc/internal/landingline ./cmd/sdlc -run 'Test(Train|LandingCheck|LandingRequirements|LandingManual)' -count=1`
   - `go test ./cmd/sdlc -run 'Test(CIGate|Car.*Landing|RunLocalCheckRunnerResource)' -count=1`
   - `git diff --check` for the pinned range.

   Inspected coverage includes later-revert selection, counted/cache reads, stale evidence, watched-doc invalidation, cancellation before launch, and refusal without a real suite. No full suite, live VM qualification, or mutation test was performed during this review.

6. **Architectural notes**

   - **ARCH-DRY — Pass:** all landing shells share the gate and existing runner.
   - **ARCH-PURE — Pass:** deterministic selection/proof logic and train transitions remain separate from IO.
   - **ARCH-PURPOSE — Pass:** the consumer sweep covers car readiness and every landing path.
   - **ARCH-MOCK — Pass:** stateful GitHub fixtures and private executable checks use production seams.
   - **ARCH-CONSTRAINTS — Pass:** cached immutable inputs and counted growth tests support the declared batch workload.
   - **ARCH-SECURE — Pass:** pinned policy, strict path evidence, and verified retained logs guard authorization.
   - **ARCH-ORDER — Pass:** explicit checking transitions control CI/merge; failures cannot enter suite bisection.
   - **ARCH-FUNERAL — Pass:** new proofs and attempt memory end with their owner; execution retains existing artifact lifecycle handling.
   - **ARCH-EVOLVE — Pass:** M2 introduces no persisted-format change or lossy rewrite.

7. **Plan revision recommendations:** None required.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"bc4018b8-9d7b-4aba-8662-bca4e0b3ace1","repository":"github.com/xianxu/ariadne","issue_id":"000509","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/c7887a4890892179","reviewed_head":"d0e43f42fa53f1bf7d94137baf3ba18988103526","slot":"ariadne:3","requested_at":"2026-10-21T07:33:59.650878Z"},"phase":"finished","pid":45293,"started_at":"2026-10-21T07:33:59.792457Z","finished_at":"2026-10-21T07:36:46.261966Z","session_id":"01a15d61-9ad2-7211-a09a-a65de485dbd0","outcome":"returned"}] -->
