# Boundary Review — ariadne#394 (milestone M3)

| field | value |
|-------|-------|
| issue | 394 — Issue-backed flake landing and daily triage |
| repo | ariadne |
| issue file | workshop/issues/000394-issue-backed-flake-landing.md |
| boundary | milestone M3 |
| milestone | M3 |
| window | 7a3f91e459f922f7bc82fb738ed3894add7c8c31..d46c4e3163b0c2f0bebf1fcc8169826f466f4606 |
| command | sdlc milestone-close --issue 394 --milestone M3 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-18T20:08:46-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Reviewing the M3 diff now; next I'm checking the plan's M3 items and the guard logic.

Still reviewing: sweeping every path that posts a suite success status, and checking the plan's "closed issue" test.

Running the targeted tests now, then reverting the guard to confirm the new tests fail without it.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

M3 does what it set out to do in code. `guardSuitePass` in `cmd/sdlc/localcheck_republish.go:137` closes the "blind rerun turns green" gap, and it sits at the only place a local-check status gets posted: `postLocalCheckRecord`. Every route goes through it: `runLocalCheck`, `republishSavedPass` for merge/CI-gate reuse, and `local-check republish`. If the store can't be read in full, it refuses. I checked that it's needed: with the guard switched off (`if false && …`), 7 subtests in 5 test functions fail, including both new M3 cases. With it back, the targeted suites, `internal/flakes` and `tl-worker` all pass. One thing stops a straight SHIP: the issue ticks M3 as done, including "a pair pilot", but nothing in the Log, Plan or Revisions shows the pilot happened or was moved. The Spec says the policy only becomes the default for every repo *after* that pilot. Everything else is Minor.

### 1. Strengths
- **One choke point.** A sweep of `PostStatus(` finds only `postLocalCheckRecord` (`localcheck_republish.go:185-202`). Nothing can post a suite success around the guard.
- **Fails closed, with a way out.** A store that can't be read whole, a missing registry, or a registry error all refuse the post. The message names the attempt IDs and gives the exact `sdlc flake register` and `republish` commands (`localcheck_republish.go:171-180`).
- **The decision itself is pure.** It reuses `flakes.Unexplained` and its `neverExplained` set (`internal/flakes/evidence.go:66-86`). So a failure that is unattributable or was judged and rejected can't be excused by registering a flake later, and a test covers this (`localcheck_flakes_test.go:270`).
- **Old tests updated honestly.** The five tests that encoded the old loophole now assert the refusal instead of being deleted. The "post failed, NOT recorded" path is still covered at `localcheck_run_test.go:145`.
- **One evidence rule.** `isFlakeEvidence` (`teststats.go:103`) is now used by `flake register`, `flake list` and test-stats, so a passing rerun still counts as evidence when its parent was later downgraded (`TestSuiteStatsRerunEvidenceSurvivesADowngradedParent`).

### 2. Critical findings
None.

### 3. Important findings
- **The pilot is ticked but neither done nor deferred** (`workshop/issues/000394-issue-backed-flake-landing.md:75`; plan line 104). The plan's M3 lists "Pilot: a pair run on a real registered flake, with the hub", and the Spec makes the all-repo default depend on it. The M3 Log entry lists no pilot, and no Revision defers it. The plan's line 103, "Base-layer propagation is checked", also has nothing behind it. **This is the 2nd finding in family `deferred-purpose-subset`.** The rule it breaks: every Plan or Done-when item that a boundary ticks needs either a delivered artifact (code, a test, or a Log entry with evidence) or a dated `## Revisions` entry that moves it with an owner. Fix: since the rollout needs a rebuild of :0 after landing, add a Revision that moves the pilot and the propagation check to after landing, as a named step before close. Also either untick M3's pilot clause or mark it pending, and go through the M3 bullets one by one against this rule.

### 4. Minor findings
- **The refusal can point at a recovery that can't work.** It always says "If it is intermittent… `sdlc flake register`". For unattributable, failed-twice, rule-off or inconclusive failures, registering never helps. Choose the wording from `neverExplained[a.Kind]`.
- **One test name now says the opposite of what it checks.** `TestRunLocalCheckStatsPublicationFailureStillPostsStatus` now asserts the status is *not* posted. Rename it.
- **The atlas overstates the predicate rule.** It says "Store readers use `isSuiteRun` … or `isFlakeEvidence`", but `weeklyattribution.go:144/327/333`, `weeklystats.go:98`, `weeklystats_inputs.go:235` and `teststats_cmd.go:140` still compare `Check == suiteCheckName` directly. They behave the same, but they're separate copies of the rule (ARCH-DRY).
- **A record-limit warning now blocks posts, not just stats.** `readSuiteAttempts` warns when it hits the record limit (10000) or byte budget (64 MiB), and the guard refuses on any warning. The writer is bounded only by the 30-day prune. I measured the live ariadne store: 155 records, 620K, so there is roughly 60× headroom. Still, it's a reader cap with no matching cap on the writer (ARCH-FUNERAL). A record removed by another slot's prune between ReadDir and Open also raises a warning, so that race can refuse a valid pass.
- **The merge help (`helptext/merge.md:145`) doesn't mention the guard**, though the plan's M3 names "merge help". The atlas update went to `ci-merge-check.md`, not the "sdlc-binary test economy" section the plan names. The content is fine either way; the plan text is what's off.

### 5. Test coverage notes
- With the guard switched off, these fail: the 2 new `TestKnownFlakeGate` subtests, `ReusesSavedPass/saved_failure_is_rerun`, 2 `SavedPassEdges` subtests, `RecordsEveryAttemptBeforeStatusPost`, and `StatsPublicationFailure…`. So the guard's behaviour is pinned.
- The plan's "a closed issue → unknown" case is covered by construction rather than tested directly. A closed issue drops out of `BuildRegistry`, and the guard's empty-registry case is the first half of the M3 subtest. That's acceptable.
- No test covers two slots on the same SHA, where B's guard reads the store before A has saved its failure (see 6).

### 6. Architectural notes
- ARCH-DRY: flag (Minor), for the leftover direct check comparisons above.
- ARCH-PURE: pass. The IO stays in the guard; the decision is `flakes.Unexplained`.
- ARCH-PURPOSE: flag (Important), for the pilot and propagation.
- ARCH-MOCK: pass. Uses the stateful `ciFakeGH`, real temp stores and git repos.
- ARCH-CONSTRAINTS: pass. Reading the whole store once per suite post is cheap at the measured size.
- ARCH-SECURE: pass. Store records go through the strict decoder and refuse when they don't parse; the registry is read from origin/main.
- ARCH-ORDER: pass, with a note. The guard rereads fresh state at the moment it posts. The state table's "two slots on one SHA" row assumes A's failure post replaces B's success. If B posts success before A saves, and A's failure post then fails every retry, success stays up over an unexplained failure. That's narrow; I'm noting it, not flagging it.
- ARCH-FUNERAL: Minor flag, for the reader cap that now blocks posts.

### 7. Plan revision recommendations
- Add: "2026-10-18 (M3 close): the pair pilot and the base-layer propagation check move to after landing (they need the :0 rebuild); owner: hub; the all-repo default waits for them."
- Correct the M3 docs bullet to say the atlas update landed in `ci-merge-check.md`, and either add the guard to the merge help or drop "merge help" from the bullet.

```findings
findings:
  - id: new
    severity: Important
    family: deferred-purpose-subset
    title: |
      M3 is checked off claiming a pair pilot and propagation check that were neither done nor deferred
    detail: |
      Issue line 75 marks M3 [x] including a pair pilot; plan lines 103-104 list the propagation check and the pilot, and the Spec makes the all-repo default conditional on that pilot. No Log entry, test, or Revision shows either. This is the 2nd finding in this family. The rule: every Plan or Done-when item a boundary checks off must cite a delivered artifact or a dated Revisions entry that moves it, with an owner. Fix: add the Revision moving the pilot and propagation check to after landing, untick the pilot clause or mark it pending, and audit the other M3 bullets against the same rule.
  - id: new
    severity: Minor
    family: unexplained-outcome-policy
    title: |
      The guard's refusal suggests registering even for failures registration can never explain
    detail: |
      guardSuitePass always says to run sdlc flake register and then republish. For neverExplained kinds (unattributable, failed twice, rule off, inconclusive), registering cannot help. Choose the recovery wording per kind.
  - id: new
    severity: Minor
    family: reader-cap-writer-unbounded
    title: |
      The store's record and byte limits now block every suite pass instead of only marking stats partial
    detail: |
      The guard refuses on any readSuiteAttempts warning, including hitting the record limit (10000) or byte budget (64 MiB). The writer is bounded only by the 30-day prune. The live ariadne store measures 155 records / 620K, so there is about 60x headroom. A concurrent prune that removes a record between ReadDir and Open also raises a warning and refuses a valid pass.
  - id: new
    severity: Minor
    family: duplicated-io-glue
    title: |
      Atlas says store readers use isSuiteRun or isFlakeEvidence; the weekly readers still compare check names directly
    detail: |
      weeklyattribution.go:144/327/333, weeklystats.go:98, weeklystats_inputs.go:235 and teststats_cmd.go:140 behave the same but repeat the predicate. Route them through isSuiteRun, or soften the atlas claim. This is the 3rd finding in this family; the rule is that the store's record-kind predicates live only in teststats.go.
  - id: new
    severity: Minor
    family: test-name-contradicts-assertion
    title: |
      TestRunLocalCheckStatsPublicationFailureStillPostsStatus now asserts the status is not posted
    detail: |
      Rename it to describe the fail-closed refusal.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"96994ef4-294f-413d-a476-b8da65fd33a1","repository":"github.com/xianxu/ariadne","issue_id":"000394","recipe":"milestone-review","round":1,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/0f367c41eb7dbc02","reviewed_head":"d46c4e3163b0c2f0bebf1fcc8169826f466f4606","slot":"ariadne:2","requested_at":"2026-10-19T03:03:58.053655Z"},"phase":"finished","pid":54804,"started_at":"2026-10-19T03:03:58.171668Z","finished_at":"2026-10-19T03:08:45.036215Z","session_id":"ab2bea49-fa9a-4bdf-8396-a60551c2b7ad","outcome":"returned"}] -->

---

## Re-review — 2026-10-18T20:19:21-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 394 — Issue-backed flake landing and daily triage |
| repo | ariadne |
| issue file | workshop/issues/000394-issue-backed-flake-landing.md |
| boundary | milestone M3 |
| milestone | M3 |
| window | 7a3f91e459f922f7bc82fb738ed3894add7c8c31..3c715596519a09b40e4eeee0b66809252f7eb6fd |
| command | sdlc milestone-close --issue 394 --milestone M3 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-18T20:19:21-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Checking the remaining direct check-name comparisons and whether the prune-race fix has a test.

The merge path checks the runner error before it reads `rec.passed()`, so a refused pass blocks the merge. Next I'm running the tests that touch the guard.

Tests pass. Now I'm reverting the per-kind wording fix to confirm its test fails without it.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

**Summary.** This round fixed what M3 needed to fix. `guardSuitePass` lives in `postLocalCheckRecord`, so every suite pass goes through it: a fresh run, a reused saved pass, or a republish. The merge and landing gates check the runner's error before they read `rec.passed()` (`cigate.go:609-614`), so a refused pass blocks the merge. BR-21's pilot and propagation work now has a dated Revision with named owners, and the issue's M3 line says it is not done in M3. The per-kind wording (BR-22) has a test that I confirmed goes red when the fix is reverted. The test rename (BR-25) is done. Two Minors are not finished. The reader sweep for BR-24 still misses `actualqueue.go:34`, and the pruned-record skip for BR-23 has no test. The merge help also never mentions the new guard. None of these block. The targeted tests pass in `internal/flakes`, `cmd/sdlc` (gate, republish, run, stats, flake) and `cmd/tl-worker`.

1. **Strengths**
   - One choke point: the guard sits in `postLocalCheckRecord` (`localcheck_republish.go:220-224`), so run, reuse and republish can't drift apart. Its error type is passed through without the misleading "publish it with republish" hint (`localcheck_run.go:287-290`, `localcheck_republish.go:232-235`).
   - It fails closed when the store or the registry can't be read, and the refusal names the attempts. The policy for which failures can never be explained is a single source, `flakes.Explainable`/`neverExplained` (`evidence.go:86-89`).
   - Test-stats now reads a passing `suite-rerun` row as evidence in its own right, not the parent's possibly downgraded outcome. `TestSuiteStatsRerunEvidenceSurvivesADowngradedParent` pins this, including that a failing rerun adds nothing.
   - The five old tests that encoded the loophole now assert the refusal and check the posted status, not just the error.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - `cmd/sdlc/helptext/merge.md:145-148`: Plan M3 says "the local-check and merge help … are updated". Merge help doesn't mention that a passing suite after an unexplained failure on the same head is refused. Merge reaches that refusal through `runLocalCheck`.
   - BR-24 is still open: `actualqueue.go:34` (`a.Check != "suite"`) and `weeklystats_inputs.go:223` (a direct `suiteRerunCheck` compare).
   - BR-23 is still open: the `fs.ErrNotExist` skip (`teststats_store.go:464-466`) has no test.
   - Plan line 104 ("Pilot: a pair run …") is not marked deferred, though the Revision says "the M3 plan line is marked accordingly". Only the issue's line is marked.

5. **Test coverage notes.** The blind-rerun refusal, register-then-republish, refusal on an unreadable store, and the unattributable kind are all covered against the real store with a fake GitHub that keeps state. Not covered: a head with two unexplained failures, one explainable and one not (the code chooses the conservative wording), and the prune race.

6. **Architectural notes**
   - ARCH-DRY: pass, apart from the BR-24 stragglers.
   - ARCH-PURE: pass. The guard is thin glue over the pure `flakes.Unexplained`/`Explainable`.
   - ARCH-PURPOSE: pass. The Spec's "a blind passing retry never erases an unexplained failure" is enforced at the single post seam. The pilot is deferred with owners.
   - ARCH-MOCK: pass. Gate tests run the real store with a fake GitHub that keeps statuses.
   - ARCH-CONSTRAINTS: pass. The reader caps that now gate landings are stated with measured headroom (155 records / 620K against 10,000 / 64 MiB).
   - ARCH-SECURE: pass. Store warnings fail closed rather than reading as success.
   - ARCH-ORDER: pass. The decision reads fresh store and registry state at post time. A concurrent prune is handled, though untested.
   - ARCH-FUNERAL: pass. M3 adds no new durable artifact, and attempts stay bounded by the 30-day prune.
   - Upcoming: enforce the reader-predicate rule mechanically, for example with a test that greps for `.Check ==`/`!=` outside `teststats.go` and the decoder. A fourth instance would then be impossible, not just unlikely.

7. **Plan revision recommendations**
   - Mark plan line 104 (Pilot) as moved by the 2026-10-18 M3 round-1 Revision.
   - Record that the merge help is updated, or move it explicitly.

```findings
dispose:
  - id: BR-21
    disposition: addressed
    note: |
      Plan Revisions 2026-10-18 M3 round 1 moves the pilot and propagation work to after landing, with owners (hub ops:0, ariadne:2), and the issue's M3 line says so; plan line 104 should be marked too (Minor).
  - id: BR-22
    disposition: addressed
    note: |
      guardSuitePass branches on flakes.Explainable; the TestKnownFlakeGate unattributable subtest asserts no "flake register" and "cannot explain it", and it fails with the branch disabled (verified).
  - id: BR-23
    disposition: not-addressed
    note: |
      The cap headroom is now in the atlas, which is fine, but the ErrNotExist skip at teststats_store.go:464 changes behavior and has no test (grep finds none); add an injected-removal case.
  - id: BR-24
    disposition: not-addressed
    note: |
      The named sites are routed through isSuiteRun, but actualqueue.go:34 still compares a.Check != "suite" and weeklystats_inputs.go:223 compares suiteRerunCheck directly, so the atlas claim is still false. The rule: no Check comparison outside teststats.go's predicates and the decoder. Enforce it with a grep test rather than fixing one more site.
  - id: BR-25
    disposition: addressed
    note: |
      Renamed TestRunLocalCheckUnreadableStoreRefusesTheSuitePass; its assertions match the name.
findings:
  - id: new
    severity: Minor
    family: deferred-purpose-subset
    title: |
      Merge help omits the no-blind-rerun guard although Plan M3 claims the merge help is updated
    detail: |
      cmd/sdlc/helptext/merge.md:145-148 still has only the M2 known-flake text, but merge reaches guardSuitePass through runLocalCheck and refuses a passing suite after an unexplained same-head failure. The rule (the same one as BR-21): every Plan item a boundary checks off cites a delivered artifact or a dated Revision. Add the refusal and its recovery (register, then republish) to merge help.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"297c1a47-76f2-4b29-96e5-4a4db60317a5","repository":"github.com/xianxu/ariadne","issue_id":"000394","recipe":"milestone-review","round":2,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/e76638a2c712254e","reviewed_head":"3c715596519a09b40e4eeee0b66809252f7eb6fd","slot":"ariadne:2","requested_at":"2026-10-19T03:16:50.92299Z"},"phase":"finished","pid":86906,"started_at":"2026-10-19T03:16:51.144367Z","finished_at":"2026-10-19T03:19:20.437334Z","session_id":"19b66bc0-ded1-4956-b677-a57be080c101","outcome":"returned"}] -->
