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
