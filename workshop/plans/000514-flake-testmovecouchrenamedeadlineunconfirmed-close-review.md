# Boundary Review — ariadne#514 (whole-issue close)

| field | value |
|-------|-------|
| issue | 514 — flake: TestMoveCouchRenameDeadlineUnconfirmed |
| repo | ariadne |
| issue file | workshop/issues/000514-flake-testmovecouchrenamedeadlineunconfirmed.md |
| boundary | whole-issue close |
| milestone | — |
| window | 6d7f67070527dec2a1569c43f4cb53d658629fe6..8d799067dcad5f717b586c067ebf575c778a1bf1 |
| command | sdlc close --issue 514 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-20T18:58:16-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The diff fixes the startup race through explicit readiness and caller cancellation, preserves the production timeout, and extends the existing lint guard. No blocking defects found. The full-suite landing requirement remains outstanding; this review does not establish that result.

Done-when: current — the revision changes verification timing without changing the root-cause acceptance requirement.

```findings
{}
```

1. **Strengths**
   - Rename tests exercise cancellation before and after the simulated effect (`cmd/tl-lead/status_move_timeout_test.go:111`).
   - Production retains the migration-only five-second bound (`cmd/tl-lead/sessionsync.go:97`).
   - Lint shares duration parsing and covers lowercase fields, exactly one second, map literals, and non-test files (`pkg/golint/rules_test.go:155`).
   - Judge retry cancellation now fires after the retry begins (`cmd/sdlc/internal/judge/dispatchstream_test.go:137`).

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage:** Passed the full `tl-lead`, `judge`, `golint`, `testgate`, and `acquire` packages; five race-enabled Couch runs; three race-enabled judge retry runs; and diff whitespace checks. No whole-module suite run during this review.

6. **Architecture**
   - **ARCH-DRY: pass** — extends the existing lint rule and reuses duration parsing.
   - **ARCH-PURE: pass** — lint logic remains pure; subprocess handling stays in the existing IO seam.
   - **ARCH-PURPOSE: pass** — addresses rename ordering and sibling startup assumptions.
   - **ARCH-EVOLVE: pass** — no persisted-format changes.
   
   Atlas documents the lint extension. No new command, flag, or configuration surface requires a README update.

7. **Plan revisions:** None required.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"9a4bfa0c-f0c7-445f-8e43-e92ed21f4634","repository":"github.com/xianxu/ariadne","issue_id":"000514","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/26ea354ae9eec3c5","reviewed_head":"8d799067dcad5f717b586c067ebf575c778a1bf1","slot":"ariadne:4","requested_at":"2026-10-21T01:56:21.415946Z"},"phase":"finished","pid":40789,"started_at":"2026-10-21T01:56:21.50575Z","finished_at":"2026-10-21T01:58:15.70581Z","session_id":"01a15c2c-7c7b-7cf3-8ebd-bba417cf8f72","outcome":"returned"}] -->

---

## Re-review — 2026-10-20T21:28:42-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 514 — flake: TestMoveCouchRenameDeadlineUnconfirmed |
| repo | ariadne |
| issue file | workshop/issues/000514-flake-testmovecouchrenamedeadlineunconfirmed.md |
| boundary | whole-issue close |
| milestone | — |
| window | 5b5a2c65f8420934fb428e694498b9aea30fdff4..211ff2523eaf7e831ba5524b0099c0e866538dd7 |
| command | sdlc close --issue 514 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-20T21:28:42-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The interdiff removes the healthy subprocess’s startup deadline while preserving the real stalled recovery probe and elapsed bound. No blocking defects found. The revision records a fixture correction within the existing startup-race scope. No prior findings require disposition.

Done-when: current — the correction supports the existing root-cause requirement without changing acceptance.

```findings
{}
```

1. **Strengths**
   - Initial observation is deterministic and uses the existing `CouchActors` seam (`cmd/tl-lead/actors_bound_test.go:23`).
   - The test retains recovery-output ordering, elapsed-bound, and attempted-slot assertions, and now verifies exactly two observations (`cmd/tl-lead/actors_bound_test.go:28–40`).

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**
   - Passed five focused normal runs and five race-enabled runs.
   - Passed the complete `cmd/tl-lead` and `pkg/golint` packages, including the tracked-file catalog.
   - Full module suite not rerun. Whitespace check reported one trailing-space line in the added historical review artifact.

6. **Architectural notes**
   - **ARCH-DRY: pass** — reuses the existing actor interface and subprocess implementation.
   - **ARCH-PURE: pass** — IO remains behind the injected seam; the stateful adapter is test-only.
   - **ARCH-PURPOSE: pass** — eliminates startup dependence while retaining the bounded recovery observation.
   - **ARCH-EVOLVE: pass** — no persisted-format changes.
   - No new architectural or user-facing surface requires atlas or README changes.

7. **Plan revision recommendations:** None.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"5d39f85d-fc3e-46b5-a407-af3b260d046d","repository":"github.com/xianxu/ariadne","issue_id":"000514","recipe":"fix-review","round":2,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/b4142df65d5ef879","reviewed_head":"211ff2523eaf7e831ba5524b0099c0e866538dd7","slot":"ariadne:4","requested_at":"2026-10-21T04:27:29.450914Z"},"phase":"finished","pid":19586,"started_at":"2026-10-21T04:27:29.592644Z","finished_at":"2026-10-21T04:28:41.13564Z","session_id":"01a15cb6-dd27-7e82-a287-f0f0e1fbee3d","outcome":"returned"}] -->

---

## Re-review — 2026-10-21T08:47:59-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 514 — flake: TestMoveCouchRenameDeadlineUnconfirmed |
| repo | ariadne |
| issue file | workshop/issues/000514-flake-testmovecouchrenamedeadlineunconfirmed.md |
| boundary | whole-issue close |
| milestone | — |
| window | 944ce7899af5978a1557b83b3ac93b8c4ad4d176..ed14a3fc97dc91e10b245f00516fce367dbc308a |
| command | sdlc close --issue 514 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-21T08:47:59-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned interdiff contains documentation and review records only. The atlas conflict resolution preserves both sides’ intent and matches the code. No prior findings require disposition, and no new defects were found.

Done-when: current — this window leaves Spec, Revisions, and acceptance unchanged; the documentation resolution supports the existing startup-race fix.

```findings
{}
```

1. **Strengths**
   - `atlas/workflow/sdlc-binary.md:2253` retains main’s landing-completion rule.
   - `atlas/workflow/sdlc-binary.md:2255` retains #514’s assignment and struct-literal deadline coverage, matching `pkg/golint/rules.go:357`.
   - The #514 lesson is relocated without changing its content.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes:** `go test ./pkg/golint -count=1` passed, including atlas/catalog consistency and tracked-source checks. `git diff --check` passed. No executable changes require mutation testing; the full suite was not rerun.

6. **Architectural notes**
   - **ARCH-DRY: pass** — conflict resolution removes duplicated catalog prose.
   - **ARCH-PURE: pass** — no executable logic or IO changes.
   - **ARCH-PURPOSE: pass** — both documented rule intents survive integration.
   - **ARCH-EVOLVE: pass** — no persisted-format or reader/writer changes.
   - No new surface requires additional atlas or README updates.

7. **Plan revision recommendations:** None.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"d2ddfe62-48b7-4ae2-a8fe-f2bbba9b6660","repository":"github.com/xianxu/ariadne","issue_id":"000514","recipe":"fix-review","round":3,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/3c2c9e78d0e30fd0","reviewed_head":"ed14a3fc97dc91e10b245f00516fce367dbc308a","slot":"ariadne:4","requested_at":"2026-10-21T15:47:14.981622Z"},"phase":"finished","pid":53260,"started_at":"2026-10-21T15:47:15.117439Z","finished_at":"2026-10-21T15:47:59.323364Z","session_id":"01a15f25-3182-7381-869e-a6621e85c8ce","outcome":"returned"}] -->
