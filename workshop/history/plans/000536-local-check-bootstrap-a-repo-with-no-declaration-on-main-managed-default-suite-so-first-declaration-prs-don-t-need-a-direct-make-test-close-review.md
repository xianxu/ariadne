# Boundary Review — ariadne#536 (whole-issue close)

| field | value |
|-------|-------|
| issue | 536 — local-check: bootstrap a repo with no declaration on main (managed default suite), so first-declaration PRs don't need a direct make test |
| repo | ariadne |
| issue file | workshop/issues/000536-local-check-bootstrap-a-repo-with-no-declaration-on-main-managed-default-suite-so-first-declaration-prs-don-t-need-a-direct-make-test.md |
| boundary | whole-issue close |
| milestone | — |
| window | 9944c67f12e581bddc96fb624449df9b9a91895a..790e3102dcc056ea10a9dda097006eb4ecc62c6c |
| command | sdlc close --issue 536 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-22T02:26:40-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned range implements the issue’s narrowed bootstrap contract: main remains authoritative, eligible subjects run the required managed default suite, and head declarations gain no policy authority. Recovery and train-prefix paths use the same derivation. No blocking defects or documentation gaps found.

```findings
{}
```

1. **Strengths**

   - `cmd/sdlc/localchecks_bootstrap.go:9` centralizes eligibility while preserving probe errors and cancellation.
   - `cmd/sdlc/localchecks_bootstrap_test.go:17` exercises the captured echo fixture through managed execution, exact-head publication and CI admission.
   - `cmd/sdlc/landingtrain.go:438` derives policy separately for each tested prefix and refuses missing-target prefixes.
   - Recovery tests verify saved/original subjects remain authoritative after the caller changes HEAD.
   - README, CLI help and atlas document eligibility, delayed declaration activation and recovery.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   All four targeted runs passed: bootstrap regressions (21.676s), CI/default/command controls (27.134s), landing/probe/ancestor controls (41.673s), and pure parsing/CI decisions (2.958s). `git diff --check` passed. The captured base Makefile’s Git blob hash matches its documented provenance. The full repository suite was not rerun during this review.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared derivation reuses existing constructors and probes.
   - **ARCH-PURE — pass:** new work is correctly classified as integration glue; pure policy constructors remain unchanged.
   - **ARCH-PURPOSE — pass:** execution, admission, recovery and train-prefix consumers are covered.
   - **ARCH-MOCK — pass:** existing dependency seams, stateful GitHub fake and private native repositories support verification.
   - **ARCH-CONSTRAINTS — pass:** existing probe deadlines remain; CI facts are cached across polling.
   - **ARCH-SECURE — pass:** pinned subjects and strict main-policy reads preserve authority boundaries.
   - **ARCH-ORDER — pass:** prefix selection is recomputed, and absent baseline evidence remains unverified.
   - **ARCH-FUNERAL — pass:** no new durable artifact family; existing managed cleanup owns execution resources.
   - **ARCH-EVOLVE — pass:** no persisted-format changes.

   Core-concept locations and classifications match the implementation.

7. **Plan revision recommendations:** None.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"dc6e9e44-fad8-483b-a779-d8600869c945","repository":"github.com/xianxu/ariadne","issue_id":"000536","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/3f370c22e9e8ada4","reviewed_head":"790e3102dcc056ea10a9dda097006eb4ecc62c6c","slot":"ariadne:2","requested_at":"2026-10-22T09:23:52.017105Z"},"phase":"finished","pid":15170,"started_at":"2026-10-22T09:23:52.170412Z","finished_at":"2026-10-22T09:26:40.140982Z","session_id":"01a162ec-8ecb-7982-a82e-4fd1ed99cf01","outcome":"returned"}] -->
