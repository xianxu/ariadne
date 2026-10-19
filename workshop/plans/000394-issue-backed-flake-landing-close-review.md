# Boundary Review — ariadne#394 (whole-issue close)

| field | value |
|-------|-------|
| issue | 394 — Issue-backed flake landing and daily triage |
| repo | ariadne |
| issue file | workshop/issues/000394-issue-backed-flake-landing.md |
| boundary | whole-issue close |
| milestone | — |
| window | 046e0ebb03d7c78bf46c8511fa6d6084d42137f4..bda6cc0efc5613bf2994d9cdca538942bb2d42fd |
| command | sdlc close --issue 394 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-18T21:50:20-07:00 |
| verdict | SHIP |

## Review

Disposing the open findings and checking the last two commits. Next I'll run the relevant tests.

Wrapping up: the fix in the latest commit is confirmed (its regression test fails without it). Doing one last check on the help text, then writing the verdict.

```verdict
verdict: SHIP
confidence: high
```

Five of the six open findings are fixed. The sixth (BR-23) is half fixed, and its remaining half is a Minor fail-closed edge case that does not block. The window after the last round has two commits. 7d3d7ab0 adds the no-blind-rerun guard to merge help. bda6cc0e adds a recovery contract for `flake register`, and changes the store reader so it only skips an attempt record that is truly gone, not a symlink whose target is missing. I checked the reader change by reverting it in this checkout: `TestOwnedActualQueueCoverageRefusesUnknownRecordsForCloseAndCorrection/unreadable` fails without it and passes with it. The `flakes`, `recovery` and `issue` packages pass. The targeted `cmd/sdlc` run of flake, suite-attempt, test-stats, local-check, republish, weekly and recovery tests also passes (535s). Nothing new rises to Important or Critical.

1. **Strengths**
   - `teststats.go:94-105`: one record-kind predicate pair. `isSuiteRun` covers run counts and saved-pass reuse; `isFlakeEvidence` (a suite run, or a passing `suite-rerun`) covers fail→pass evidence. The guard, `flake list`/`register`, aggregation and the weekly readers all go through them now.
   - The rerun's pass survives a downgraded parent, and `teststats_test.go:100` covers this.
   - `teststats_store.go:464`: an attempt record that disappears after the directory listing is skipped only if `Lstat` also confirms it is gone. A dangling symlink is still reported as unreadable, so the guard stays fail-closed. An existing test proves it.
   - `recovery/catalog.go:290`: the `flake register` contract covers both paths, with and without `--issue`. Its proofs point at real tests, including a new check that repeating a registration writes nothing (`flake_cmd_test.go:215`).
   - `flake_cmd.go:224-228` and atlas `sdlc-binary.md:49` both state the one-owner rule where it is actually enforced: when the registry is read (`BuildRegistry`).

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings**
   - BR-23 (kept open): the guard still refuses a suite pass when the reader hits its record limit (10,000) or byte budget (64 MiB). This fails closed, and there is about 60× headroom today. Cheapest fix: one sentence in the local-check help saying a store over its limits refuses suite passes until it is pruned.

5. **Test coverage notes:** the evidence for BR-20 is behavioral (`teststats_test.go:100`). The bda6cc0e reader fix fails without the change, as verified above. One test run failed when I filtered tests by name (`TestScanIssueFilesSkipsDeletedUnreadableAndMalformed`), but only because the sandbox can't run the `git remote-readonly` helper. It is not related to this change, and the full targeted run passed.

6. **Architectural notes**
   - ARCH-DRY: pass. Record-kind predicates exist only in `teststats.go`.
   - ARCH-PURE: pass. `flakes.Classify`, `Unexplained` and `Confirmed` are pure and tested without IO.
   - ARCH-PURPOSE: pass. Every store reader now goes through the predicates, and both help pages name the guard.
   - ARCH-MOCK: pass. The tracker-card setter is tested against real cards.
   - ARCH-CONSTRAINTS: pass, apart from the BR-23 reader cap noted above.
   - ARCH-SECURE: pass. Test names are argv-safe, and unreadable records fail closed.
   - ARCH-ORDER: pass. The card merge is decided on the compare-and-swap snapshot, and uniqueness across issues is enforced when the registry is read, as documented.
   - ARCH-FUNERAL: pass. Attempts are removed by the 30-day prune, and a registry entry ends when its issue closes.
   - BR-10's schema-v2 break is documented in atlas. The release notes for downstream repos should repeat it.

7. **Plan revision recommendations:** none required.

```findings
dispose:
  - id: BR-9
    disposition: addressed
    note: |
      Register help (flake_cmd.go:224-228) and atlas sdlc-binary.md:49 state the rule: uniqueness is checked at registration but enforced when the registry is read (BuildRegistry refuses an ambiguous test, which fails closed).
  - id: BR-10
    disposition: addressed
    note: |
      Atlas sdlc-binary.md:49 Schema note documents that pre-394 binaries refuse v2 attempts and the fleet must run rebuilt ariadne:0 binaries; plan line 122 matches.
  - id: BR-20
    disposition: addressed
    note: |
      isFlakeEvidence (teststats.go:103) admits a passing suite-rerun row as evidence regardless of the parent's Outcome. Used by aggregate, flake list/register and the guard; teststats_test.go:100 covers a downgraded parent.
  - id: BR-23
    disposition: not-addressed
    note: |
      Concurrent-prune half fixed in bda6cc0e (Lstat check; the actualinputs unreadable subtest fails without it). The record/byte-limit refusal remains: fail-closed, Minor; note it in local-check help.
  - id: BR-24
    disposition: addressed
    note: |
      weeklyattribution.go:144/327/333, weeklystats.go:98, weeklystats_inputs.go:235 and teststats_cmd.go:140 all call isSuiteRun now.
  - id: BR-26
    disposition: addressed
    note: |
      7d3d7ab0 merge.md:148-152 names the refusal and its recovery (flake register, then local-check republish --name suite), matching guardSuitePass in localcheck_republish.go:137.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"133673a6-7a80-4706-8a5e-64071791ce05","repository":"github.com/xianxu/ariadne","issue_id":"000394","recipe":"milestone-review","round":9,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/913af4c248182d61","reviewed_head":"bda6cc0efc5613bf2994d9cdca538942bb2d42fd","slot":"ariadne:2","requested_at":"2026-10-19T04:37:16.152728Z"},"phase":"finished","pid":21786,"started_at":"2026-10-19T04:37:16.514649Z","finished_at":"2026-10-19T04:50:14.718877Z","session_id":"525a1d20-4dde-4bc1-84e3-214166e46c0f","outcome":"returned"}] -->

---

## Re-review — 2026-10-18T23:19:19-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 394 — Issue-backed flake landing and daily triage |
| repo | ariadne |
| issue file | workshop/issues/000394-issue-backed-flake-landing.md |
| boundary | whole-issue close |
| milestone | — |
| window | 67db085211bed150086a8f929727671f52f7ce23..cf0284d281f887fbcea87863421bc1e8ace37c3c |
| command | sdlc close --issue 394 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-18T23:19:19-07:00 |
| verdict | SHIP |

## Review

Reviewing the conflict resolution; next I'll build and run the affected tests.

```verdict
verdict: SHIP
confidence: high
```

This window contains no new feature work. It has two parts. The first resolves a merge conflict in `cmd/sdlc/localcheck_run.go`: the pinned base still had literal conflict markers in `runUnderSlot`. The second moves two #394 lessons to a different spot in `workshop/lessons.md`. The resolution keeps both sides. It keeps this branch's change that made `runUnderSlot` generic, so the slot uses the `weight` parameter instead of `c.Weight`. It also keeps main's #452 queue priority, which passes `subject` to `acquireSuiteSlot`. The resolution also threads `subject` through `suiteGateRun`, so the known-flake rerun queues at the same priority as the suite run that triggered it. That change was needed: the `runUnderSlot` signature changed, and the rerun runs only on the suite path, where `subject.Issue` is kept. `go build ./...` and `go vet` are clean. `go test -run 'LocalCheck|KnownFlake|Flake|SuiteSlot|Priority|Republish' ./cmd/sdlc` passed (308s). The window changes no Spec, Revisions or Done-when, so scope is unchanged.

1. **Strengths**
   - `cmd/sdlc/localcheck_run.go:257`: the resolution calls `acquireSuiteSlot(ownedCtx, d, wt, weight, helper, subject)`. That keeps this branch's generic weight and main's priority subject. No stale `c` reference is left inside the generic helper.
   - `cmd/sdlc/localcheck_flakes.go:37,86`: the rerun gets the suite's priority subject through `suiteGateRun` rather than a second `suiteSubject` lookup (ARCH-DRY pass).
   - No conflict markers remain in the code. The `<<<<<<<` matches under `cmd/` are test fixtures and the marker detector in `republish.go`.

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings:** none new. The `lessons.md` change only moves two bullets into the section they belong to; the text is unchanged.

5. **Test coverage notes:** The resolution only threads a parameter through, and the existing local-check, slot and known-flake tests pass on it. No test checks that the rerun inherits the suite's priority subject. A cheap later addition would be a fake slot acquirer that records the subject it receives.

6. **Architectural notes**
   - **ARCH-DRY: pass.** The subject is computed once and reused.
   - **ARCH-PURE: pass.** No logic was added inside the IO code; it is parameter threading only.
   - **ARCH-PURPOSE: pass.** Both #394 and #452 still do what they set out to do after the merge.
   - **ARCH-EVOLVE: pass.** No persisted format changes in this window.

7. **Plan revision recommendations:** none.

**BR-23 disposition:** Round 9 confirmed the concurrent-prune half fixed and tested. The remaining cap refusal is a deliberate fail-closed choice, documented in two places:
- `cmd/sdlc/helptext/local-check.md:97`: "A test-stats store it cannot read whole also refuses the pass."
- `atlas/workflow/ci-merge-check.md:192-194`: says the 10,000-record / 64 MiB caps now gate landings, notes the measured headroom, and says a record pruned mid-read is skipped.

```findings
dispose:
  - id: BR-23
    disposition: addressed
    note: |
      The concurrent-prune half was fixed and tested in bda6cc0e (round 9). The remaining cap refusal is deliberately fail-closed and is now documented as asked: local-check help (helptext/local-check.md:97) and atlas/workflow/ci-merge-check.md:192-194, which names the 10,000/64 MiB caps and the measured ~60x headroom. Nothing in this interdiff changes that behavior.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"49888287-8134-4675-b949-a0356bafb724","repository":"github.com/xianxu/ariadne","issue_id":"000394","recipe":"fix-review","round":10,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/04b23ac34fc7e191","reviewed_head":"cf0284d281f887fbcea87863421bc1e8ace37c3c","slot":"ariadne:2","requested_at":"2026-10-19T06:13:17.09262Z"},"phase":"finished","pid":79053,"started_at":"2026-10-19T06:13:17.204552Z","finished_at":"2026-10-19T06:19:14.094788Z","session_id":"8e1c9fbe-2006-4f9d-9e54-c92a16c58623","outcome":"returned"}] -->
