# Boundary Review — ariadne#561 (whole-issue close)

| field | value |
|-------|-------|
| issue | 561 — Landing trains: cars board on affected tests; full suite once per train; measure escapes and bisect cost |
| repo | ariadne |
| issue file | workshop/issues/000561-landing-trains-cars-board-on-affected-tests-full-suite-once-per-train-measure-escapes-and-bisect-cost.md |
| boundary | whole-issue close |
| milestone | — |
| window | 6318fc7bc700173acdcea67f3993d326f8947845..d58219e00d87c953bcbc3b6ec7ed345a1bad9185 |
| command | sdlc close --issue 561 |
| reviewer | claude (anthropic; author anthropic,openai) |
| isolation | fenced |
| timestamp | 2026-10-23T05:28:03-07:00 |
| verdict | SHIP |

## Review

Checking the open findings now: I'll revert one recovery-command site and confirm the guard test fails, then run the relevant test packages.

```verdict
verdict: SHIP
confidence: high
```

All three open findings are fixed and backed by tests that fail without the fix. The changes since round 4 are small: the status filter rendering moved into the pure `trainFilterLines` with a test that checks it line by line, and the last commit adds a lesson. The guard test that keeps local-check commands in one place goes red when I revert a fix: I restored the hand-built command at `localcheck_reclassify.go:70` and `TestLocalCheckRunCommandIsSingleSourced` failed. After restoring the file, the targeted `cmd/sdlc` tests and the whole `internal/landingline` package (including the fuzz seeds) pass. Nothing new blocks shipping.

1. **Strengths**
   - `cmd/sdlc/cigate.go:262`: `localCheckRunCommand` is now the only place that builds a user-facing check command. It adds `--boarding` for `affected` and handles the optional `--rev`. All five call sites use it, and a guard test that scans the source enforces this. That covers the whole class of bug, not just the one site a finding named.
   - `cmd/sdlc/landingcmd.go:161`: `trainFilterLines` is pure and tested exactly (`TestTrainFilterLines`). It also replaced the git-backed `shortSHA` with the pure `abbrevSHA`, which keeps git out of the status rendering.
   - The prior-round fixes still hold: strict measurement decoding, fuzz targets checked against a reference fold, and the run kind passed explicitly.

2. **Critical:** none.
3. **Important:** none.
4. **Minor:** The status test fixture shows `car #0` in its escape line (`landingcmd_test.go:220`). That is a fixture artifact; a real issue number would read better.
5. **Test coverage notes:**
   - The guard regex catches hand-built forms that use `%s`/`%v` or `" +` concatenation. A literal `--name suite` is allowed on purpose.
   - Every status line built from the fold now has a test, which satisfies the rule for the `derived-surface-untested` family.
6. **Architecture:** 
   - Pass: ARCH-DRY, ARCH-PURE, ARCH-PURPOSE, ARCH-SECURE, ARCH-ORDER and ARCH-EVOLVE.
   - ARCH-SECURE: decoding is strict, which was the BR-11 fix.
   - ARCH-MOCK, ARCH-CONSTRAINTS and ARCH-FUNERAL: no new surface since round 4, so they are unchanged from that pass. The new lesson entry follows the existing append-only lessons convention.
7. **Plan revisions:** none. The round-3 Revisions entry still matches the code.

```findings
dispose:
  - id: BR-2
    disposition: addressed
    note: |
      cigate.go:240/712 use localCheckCommand; TestLocalCheckRunCommandIsSingleSourced scans non-test sources and rejects hand-built forms, so restoring the old string goes red.
  - id: BR-5
    disposition: addressed
    note: |
      suiteenvironment.go:45 and localcheck_reclassify.go:70 call localCheckRunCommand(check, rev); verified reverting reclassify.go:70 turns TestLocalCheckRunCommandIsSingleSourced red.
  - id: BR-13
    disposition: addressed
    note: |
      Rendering extracted to pure trainFilterLines (landingcmd.go:161); TestTrainFilterLines asserts summary, escape and overflow lines exactly; status test also asserts the escape line.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"ba02a67d-1cd4-484d-9522-5abf43a50450","repository":"github.com/xianxu/ariadne","issue_id":"000561","recipe":"milestone-review","round":5,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/97fc2cc0acbfa234","reviewed_head":"d58219e00d87c953bcbc3b6ec7ed345a1bad9185","slot":"ariadne:2","requested_at":"2026-10-23T12:27:03.494487Z"},"phase":"finished","pid":31454,"started_at":"2026-10-23T12:27:03.643398Z","finished_at":"2026-10-23T12:28:02.907198Z","session_id":"5b1f1d8a-4e75-4ce7-a4ee-8ec566f24059","outcome":"returned"}] -->
