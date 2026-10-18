# Boundary Review — ariadne#394 (milestone M1)

| field | value |
|-------|-------|
| issue | 394 — Issue-backed flake landing and daily triage |
| repo | ariadne |
| issue file | workshop/issues/000394-issue-backed-flake-landing.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 046e0ebb03d7c78bf46c8511fa6d6084d42137f4..1a2e64e114bcf3f97920592e4a4f9350be51c5ca |
| command | sdlc milestone-close --issue 394 --milestone M1 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-18T16:45:21-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

I've finished reading the diff. Now I'm checking the card-update retry path for a possible lost update, then I'll run the tests.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The M1 code mostly matches the plan. `internal/flakes` has no IO. `Classify` covers every row of the state table and fails closed: an unreadable registry, an unnamed failure or an unregistered failure all block. The test-stats "confirmed" rule now calls the shared `flakes.Confirmed`, and its ordering and conditions are the same as the loop it replaced. The v1/v2 decoder refuses a v1 record that carries v2 fields. I ran the targeted tests on the pinned head and they are green: `internal/flakes`, `internal/issue`, `pkg/vocab`, and in `cmd/sdlc` `-run 'TestFlake|TestUnattributed|TestDecodeSuiteAttempt|TestSuiteLogFacts|…'`. Four Important findings stop a clean SHIP:
- **Test-name validation:** it lets a name through that becomes a `go test` flag in the rerun command (confirmed with a probe test).
- **`flake register` lost update:** it builds the new list from an earlier read than the one its compare-and-swap checks, so a concurrent registration can be silently dropped.
- **`flake list` is short of D3 and Done-when 3:** it omits owner, age and log.
- **README:** it does not document the new `sdlc flake` command.

None of these is a correctness bug in a path that is live today, because the rerun and the gate arrive in M2. All are cheap to fix.

### 1. Strengths
- `cmd/sdlc/internal/flakes/classify.go:714-769`: one pure decision function. The rule-off, unattributable and registry-unusable checks run before any registry lookup, so absent data never reads as "explained".
- `cmd/sdlc/teststats.go:148`: the extraction to `flakes.Confirmed` keeps the old sort order and conditions (`Exit==0 && !TimedOut`, not legacy, not baseline). It is a real single source of truth, not a copy.
- `cmd/sdlc/internal/issue/card.go:1201-1226`: `flake_tests` is checked as a card field at parse time (flow list, 1 to `MaxRerun` entries, distinct, exact names). The vocabulary row in `issue.cue` drives the ownership list, and `pkg/vocab/card_test.go` pins it.
- `cmd/sdlc/teststats_store.go:141-164`: the decoder accepts only v1 or v2, refuses v2 keys inside a v1 record, and checks `outcome`, `rerun_of` and `known_flakes`. Untrusted store files become typed values.
- `cmd/sdlc/flake_cmd_test.go:580`: `TestFlakeRegistryFromTrackerCards` runs against real tracker cards, covering an open card, a done card and a setter round trip, rather than only the in-memory fake.

### 2. Critical findings
None.

### 3. Important findings
- **I1 (ARCH-SECURE): a test name can inject a flag into the rerun command.** `cmd/sdlc/internal/flakes/registry.go:1065`. The import-path pattern `^[A-Za-z0-9_.~-]+(/…)*$` accepts a segment that starts with `-`, and also `.` and `..`.
  - A probe test on the pinned head showed `RerunCommand(["-exec.TestA"])` returns `[go test -count=1 -run ^(TestA)$ -exec]`, and `./x.TestB` and `../../y.TestC` also pass.
  - The card check uses the same `ValidTest`, and any slot can write the field. So a card listing `[-exec.TestA, ./evil.TestB]` would make M2's rerun run `go test … -exec ./evil`.
  - Fix: reject segments that start with `-` or `.`, and require the first segment to look like a module path. Add the hostile cases to `TestRerunCommand` and `TestBuildRegistry`.
- **I2 (ARCH-ORDER): `flake register` can silently drop a concurrent registration.** `cmd/sdlc/flake_cmd.go:285-294` and `:139-143`.
  - The merged list (`mergeTests(c.Tests, tests)`) and the `MaxRerun` check come from the earlier Fresh registry read.
  - The `decide` callback ignores the `card` that `runCardUpdate` re-snapshots and writes the list it was given. The compare-and-swap only guards that second snapshot.
  - If another slot adds a test between the two reads, its test is overwritten, and both commands report success.
  - Fix: inside `decide`, read `flake_tests` from `card.Raw`, merge into it, then check the cap. Add a test where the card changes between the registry read and the set.
- **I3 (ARCH-PURPOSE): `flake list` does not deliver the D3 view.** `cmd/sdlc/flake_cmd.go:341-343`. D3 asks for "registry (test, issue, owner, age)… Each item comes with its attempt id and log", and Done-when 3 asks for owner and issue links. The code prints test, issue and status only: no owner (the card's claimant), no age, and no log path on the attempt lines. Fix: add a claimant and a created/started time to `flakes.Card`, print `LogPath`, and assert all three in `TestFlakeList`.
- **I4 (docs gate): README update missing for `sdlc flake register` / `sdlc flake list`.** README.md documents `local-check` and `test-stats` (around lines 486-522) but has nothing for the new command. The atlas was updated (`atlas/workflow/sdlc-binary.md:49`).

### 4. Minor findings
- `flake register --issue N --evidence TEXT` only prints the evidence. The plan says it "is recorded verbatim", but it is kept only when a new issue is filed (`flake_cmd.go:268-297`).
- `flake register --issue N` does not check that issue N is open. Registering on a done issue reports success, but the test never enters the registry.
- If `setTests` fails after `newIssue` in the no-`--issue` path, an unclaimed flake issue with no tests is left behind.
- `flakes.Unexplained` treats `KnownFlakeTwice`, `TooManyKnownFlaky` and `RuleOff` attempts as explained once all their failures are registered. A blind full rerun after "failed twice" would therefore pass M3's guard. Decide this in M3 on purpose.
- ARCH-DRY nits:
  - The `Repo == repo && Check == suiteCheckName` filter is duplicated (`flake_cmd.go:211-216` and `:348-353`).
  - `defaultFlakeDeps.registry` repeats the body of `loadFlakeRegistry`.
- `unattributedFailures` matches the JSON form only on `"Action":"fail"` with no spaces. That is fine for go's output, but brittle.

### 5. Test coverage notes
- `TestClassify` covers every row of the state table plus the cap and the `Issues()` dedupe.
- `TestUnattributedFailures` has a fixture for each unattributable shape.
- The decoder's v1 and v2 cases are present.
- Gaps:
  - No hostile-path cases (I1).
  - No concurrent-registration case (I2).
  - No owner, age or log assertion in `TestFlakeList` (I3).
  - No test of registering on a closed issue.
- I did not revert-test anything: this is the first round, so there were no claimed fixes to check.

### 6. Architectural notes for upcoming work
- Each ARCH principle:
  - **ARCH-DRY:** pass, with the nits above.
  - **ARCH-PURE:** pass. `internal/flakes` has no IO, and the verbs use injected `flakeDeps`.
  - **ARCH-PURPOSE:** flag (I3).
  - **ARCH-MOCK:** pass. `flakeWorld` is a stateful fake, and the real-tracker test covers the seam.
  - **ARCH-CONSTRAINTS:** pass. `MaxRerun` is enforced both at the card check and in `Classify`.
  - **ARCH-SECURE:** flag (I1).
  - **ARCH-ORDER:** flag (I2).
  - **ARCH-FUNERAL:** pass. The new attempt fields ride the existing 30-day pruning, and entries end when the issue closes. The orphan issue noted above is the only leftover.
- In M2, a targeted rerun recorded as a passing attempt on the same SHA will count as `Confirmed` evidence in test-stats. Decide whether attempts with `RerunOf` set should count as fail→pass proof for registration and stats.
- `ParsePolicy` and `RerunCommand` have no production consumer yet. That is expected for M1; M2 must wire them through `runUnderSlot`.

### 7. Plan revision recommendations
- Add a "## Revisions" entry saying the test-name rule rejects path segments that start with `-` or `.`. This is part of ARCH-SECURE's "validated identifiers only" claim.
- Add a "## Revisions" entry saying `flake register` merges into the card's current `flake_tests` inside the compare-and-swap callback.

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: argv-identifier-validation
    title: |
      Test-name validation accepts '-' and '.'/'..' path segments, so RerunCommand can emit go test flags (ARCH-SECURE)
    detail: |
      registry.go importPath regex lets "-exec.TestA" and "./x.TestB" pass ValidTest. A probe test showed RerunCommand returns [go test -count=1 -run ^(TestA)$ -exec]. The card check uses the same rule and any slot can write the card. Reject segments starting with '-' or '.', and add the hostile cases to TestRerunCommand and TestBuildRegistry.
  - id: new
    severity: Important
    kind: defect
    family: cas-mutator-uses-stale-read
    title: |
      flake register computes the merged flake_tests from an earlier read than its compare-and-swap snapshot, so concurrent registrations are lost (ARCH-ORDER)
    detail: |
      flake_cmd.go:285-294 merges from the Fresh registry read, and the setTests decide callback (:139-143) ignores card.Raw. Merge from the card passed to decide, check MaxRerun there, and test a card that changes between the two reads.
  - id: new
    severity: Important
    kind: defect
    family: deferred-purpose-subset
    title: |
      flake list omits owner, age and the log that D3 and Done-when 3 require (ARCH-PURPOSE)
    detail: |
      flake_cmd.go:341-343 prints test, issue and status only. Attempt lines carry no LogPath. Add the claimant and age to flakes.Card, print the log path, and assert all three in TestFlakeList.
  - id: new
    severity: Important
    kind: defect
    family: docs-gate-readme
    title: |
      README update missing for sdlc flake register / sdlc flake list
    detail: |
      The atlas row was added, but README.md, which documents local-check and test-stats, has nothing for the new user-facing command.
  - id: new
    severity: Minor
    kind: defect
    family: evidence-recorded-verbatim
    title: |
      --evidence with --issue N is only printed, not recorded as the plan says
    detail: |
      Only the path that files a new issue keeps the evidence text, in the issue body.
  - id: new
    severity: Minor
    kind: defect
    family: register-target-state-check
    title: |
      flake register --issue N does not refuse a terminal (closed) issue
    detail: |
      It reports success, but the test never enters the registry.
  - id: new
    severity: Minor
    kind: defect
    family: unexplained-outcome-policy
    title: |
      Unexplained treats failed-twice, too-many and rule-off attempts as explained once their failures are registered
    detail: |
      A blind full rerun after "known flake failed twice" would then pass M3's guard. Decide this explicitly in M3.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"86e468cc-1abc-40f2-9b8e-6476db969e02","repository":"github.com/xianxu/ariadne","issue_id":"000394","recipe":"milestone-review","round":1,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/6f0fffb25c67a2a1","reviewed_head":"1a2e64e114bcf3f97920592e4a4f9350be51c5ca","slot":"ariadne:2","requested_at":"2026-10-18T23:42:59.604244Z"},"phase":"finished","pid":15555,"started_at":"2026-10-18T23:42:59.697805Z","finished_at":"2026-10-18T23:45:20.68225Z","session_id":"a95c51ac-f943-44e8-9dc9-1eb426ba6e61","outcome":"returned"}] -->

---

## Re-review — 2026-10-18T16:52:50-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 394 — Issue-backed flake landing and daily triage |
| repo | ariadne |
| issue file | workshop/issues/000394-issue-backed-flake-landing.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 046e0ebb03d7c78bf46c8511fa6d6084d42137f4..93687c932b8b1220194fa048d949abdab7bc99d7 |
| command | sdlc milestone-close --issue 394 --milestone M1 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-18T16:52:50-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Prior findings BR-1 and BR-7 look fixed. Now I'm checking the tests for BR-2 and BR-3, then running the suite.

Now a revert experiment: make `addFlakeTests` ignore the card's existing tests.

The revert stays green, so BR-2's production merge has no regression test. Next I'm checking BR-3's real owner and created-date paths.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The fix commit `93687c93` correctly handles six of the seven round-1 findings. One prior finding stays open: BR-2. Its production code is correct: `addFlakeTests` merges onto the card that `runCardUpdate`'s compare-and-swap snapshot hands it. But nothing tests that merge. I changed `addFlakeTests` to write only the new tests (`tests := add`, dropping the card's existing ones), and every flake test stayed green, including `TestFlakeRegisterKeepsARacingRegistration` and `TestFlakeRegistryFromTrackerCards`. That racing test runs against `flakeWorld.addTests`, a fake that has its own copy of the merge and cap logic. The real-card test only checks that the new test was added; it never checks that the old one survived. The fix is cheap: add one assertion on the real card. Nothing else blocks.

Checks run:
- `go test ./cmd/sdlc/internal/flakes ./cmd/sdlc/internal/issue ./pkg/vocab` passes.
- `go test ./cmd/sdlc -run 'Flake|TestStats|SuiteAttempt|…'` passes.
- `go vet ./cmd/sdlc/...` is clean.
- Three reverts, each restored afterwards:
  - Dropping the merge (BR-2): tests stay green.
  - Disabling the closed-issue guard (BR-6): goes red.
  - Making the decoder accept only v1: goes red.

1. **Strengths**
   - `flakes.Confirmed` (`cmd/sdlc/internal/flakes/evidence.go:28`) is now the one same-head fail→pass rule. `aggregateSuiteAttempts` calls it at `teststats.go:148` instead of keeping its own loop (ARCH-DRY delivered).
   - The import-path regex (`registry.go:20`) rejects a segment that starts with `-` or `.`. The card check (`validateTestListNode`) and `RerunCommand` share that rule, and both test tables have the hostile cases (BR-1).
   - The decoder (`teststats_store.go:141-164`) is strict about versions: a v2 field in a v1 record, an unknown outcome, a malformed known flake and a rerun that points at itself are all rejected and tested.
   - `Unexplained` now writes the policy for rejected outcomes down in code and tests it (`neverExplained`, TestEvidence), in M1 rather than M3 (BR-7).
   - The closed-issue refusal is checked on a real tracker card inside the compare-and-swap callback. The revert showed that test fails without the guard (BR-6).

2. **Critical findings:** none.

3. **Important findings:**
   - BR-2 stays open (see dispositions). Fix sketch: in `TestFlakeRegistryFromTrackerCards`, after `d.addTests(ctx, …, 51, []string{flakeT2})`, assert that `reg.Lookup(flakeT1)` still returns `000051`. Better still, test `addFlakeTests` directly on a real card that already lists tests, and on one that already has `MaxRerun` tests (the cap check is also exercised only through the fake). Have `flakeWorld.addTests` call the production `addFlakeTests`/`mergeTests` rules instead of re-implementing them (ARCH-MOCK).

4. **Minor findings:**
   - ARCH-DRY: `defaultFlakeDeps.registry` (`flake_cmd.go:159-170`) repeats `loadFlakeRegistry` (`:104-114`) line for line except for returning the cards. The `attempts` and `logPath` closures (`:130-157`) each rebuild the same `localCheckLogDir` setup. Have `loadFlakeRegistry` return the cards and pull out a `suiteStoreDir()` helper.
   - `flakeCardOwner` and the `created` lookup are only exercised through `flakeWorld`, which fills in `Owner` and `Created` itself. No real-card test checks what `flake list` prints for owner and age.

5. **Test coverage notes**
   - `TestUnattributedFailures` covers every shape of unattributable failure.
   - `TestClassify` covers every row of the state table.
   - The v1/v2 decoder fixtures are in place.
   - The gap is the one in BR-2: the fake re-implements the production compare-and-swap decision, so the race test can only pass.

6. **Architectural notes for upcoming work**
   - ARCH-DRY: flagged (Minor, above).
   - ARCH-PURE: pass. `internal/flakes` does no IO and is tested without mocks.
   - ARCH-PURPOSE: pass for M1's scope. The gate itself is M2.
   - ARCH-MOCK: flagged, via BR-2. The fake must sit behind the seam and call the production decision, not copy it.
   - ARCH-CONSTRAINTS: pass. The 5-test cap is enforced on both the card and the rerun.
   - ARCH-SECURE: pass after BR-1. `--evidence` is operator text that goes into an issue body, which is acceptable.
   - ARCH-ORDER: a racing registration on the same card is now handled by the compare-and-swap. A race across cards is not: two slots registering the same test on different issues both pass the earlier one-owner check. `BuildRegistry` then refuses as ambiguous, which fails closed (every known flake blocks until someone fixes it). M2/M3 should name this, and the recovery, in the state table.
   - ARCH-FUNERAL: pass. The temp details file is removed, and the new attempt fields ride the store's 30-day pruning.
   - Rollout: the writer now records `Version: 2` (`localcheck_run.go:190`) before M2 fills in any v2 field, so binaries built before this land will refuse new attempts from now on. The plan's ":0 rebuild" rollout covers this; just be aware it starts at M1, not M2.

7. **Plan revision recommendations**
   - Under the BR-2 revision, say that the regression evidence is a real-card setter test that asserts the existing tests survive the merge, not the `flakeWorld` race test alone.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      registry.go:20 segment regex rejects leading - and . ; hostile cases in TestBuildRegistry and TestRerunCommand; card check shares ValidTest.
  - id: BR-2
    disposition: not-addressed
    note: |
      Code is right (addFlakeTests merges onto the CAS card) but nothing tests it: changing it to tests := add keeps every flake test green, because the race test's fake re-implements the merge and the real-card test never checks that the old test survives. Add that assertion on a real card.
  - id: BR-3
    disposition: addressed
    note: |
      flake list prints owner, age and log path (flake_cmd.go:369,396,406); TestFlakeList asserts all three.
  - id: BR-4
    disposition: addressed
    note: |
      README.md:498-505 documents flake register and flake list, matching the help and code.
  - id: BR-5
    disposition: addressed
    note: |
      Plan Revisions replace 'recorded verbatim' with filed-issue body or printed for N's Log; help (flake_cmd.go:210-213) and code (:321-323) match.
  - id: BR-6
    disposition: addressed
    note: |
      addFlakeTests refuses a terminal status inside the CAS callback; TestFlakeRegistryFromTrackerCards goes red when the guard is reverted.
  - id: BR-7
    disposition: addressed
    note: |
      neverExplained in evidence.go writes the policy down; TestEvidence covers failed-twice, too-many, rule-off and unattributable.
findings:
  - id: new
    severity: Minor
    kind: defect
    family: duplicated-io-glue
    title: |
      defaultFlakeDeps.registry duplicates loadFlakeRegistry, and attempts/logPath duplicate the store-dir setup (ARCH-DRY)
    detail: |
      flake_cmd.go:159-170 repeats loadFlakeRegistry (:104-114) except for returning the cards; :130-157 builds localCheckLogDir twice. Have loadFlakeRegistry return the cards and pull out a suiteStoreDir helper.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"e48c0ccd-fa0e-42bb-bf56-99e70549daea","repository":"github.com/xianxu/ariadne","issue_id":"000394","recipe":"milestone-review","round":2,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/2ccbf0b21a1c09d2","reviewed_head":"93687c932b8b1220194fa048d949abdab7bc99d7","slot":"ariadne:2","requested_at":"2026-10-18T23:49:32.691167Z"},"phase":"finished","pid":76127,"started_at":"2026-10-18T23:49:32.922656Z","finished_at":"2026-10-18T23:52:49.010462Z","session_id":"fa33b215-2823-4fd1-b6c8-9d7016d91b53","outcome":"returned"}] -->

---

## Re-review — 2026-10-18T16:56:28-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 394 — Issue-backed flake landing and daily triage |
| repo | ariadne |
| issue file | workshop/issues/000394-issue-backed-flake-landing.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 046e0ebb03d7c78bf46c8511fa6d6084d42137f4..9baa43c83bf3abb2623fd48cdd7acec651730e05 |
| command | sdlc milestone-close --issue 394 --milestone M1 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-18T16:56:28-07:00 |
| verdict | SHIP |

## Review

Both prior fixes check out: reverting the BR-2 fix makes two tests fail. Now skimming the rest of the pure core and the plan.

```verdict
verdict: SHIP
confidence: high
```

Both open findings are fixed, and I checked the BR-2 fix by reverting it. `addFlakeTests` (`cmd/sdlc/flake_cmd.go:82-93`) now makes the merge, the 5-test cap and the closed-issue check on the card that `runCardUpdate`'s compare-and-swap snapshot passes to `decide` (`cmd/sdlc/cardsetters.go:48-68`). In the disposable checkout I made `addFlakeTests` ignore the card's existing tests. `TestAddFlakeTestsMergesOntoTheCard` and `TestFlakeRegistryFromTrackerCards` both failed, and I restored the file afterwards. The BR-8 cleanup also holds: `loadFlakeRegistry` returns the cards, and one `suiteStoreDir` helper is shared by `attempts` and `logPath`. The flake-related tests pass in `cmd/sdlc`, `internal/flakes` and `internal/issue`. Nothing blocks SHIP. Two Minor notes follow.

1. **Strengths**
   - Production and the test fake make the same decision. The fake's `addTests` calls `mergeFlakeTests` directly (`flake_cmd_test.go:62`), so the race test checks production logic rather than a copy of it.
   - `flakes.Confirmed` is now the single same-head fail→pass rule, used by both test-stats (`teststats.go:148`) and registration (`flakes.SameHeadFailPass`). This is ARCH-DRY done right.
   - `Classify` is pure and fails closed. A registry error, an unnamed failure or the rule being off all block; none is treated as an empty registry.
   - Card validation (`issue/card.go:309`) and v2 record decoding (`teststats_store.go:141-164`) parse input that comes from outside the process into checked values: names are validated, outcomes come from an enum, and counts are bounded (ARCH-SECURE).
   - The README and atlas both cover the `flake` verbs.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - The check that a test is "already registered on another issue" (`flake_cmd.go:294-304`) uses the registry read, not the card being written. Two registrations of the same test on different issues at the same time can both pass. `BuildRegistry` then refuses the whole registry, which fails closed and visibly.
   - `cmd/sdlc/internal/flakes/evidence.go:84-86`: the `// Sort orders attempts…` comment sits above `neverExplained` instead of above `Sort`.
   - The writer now records `version: 2` (`teststats_store.go:50`, `localcheck_run.go:190`). An older `sdlc` binary reading the same store will reject those records as an unsupported version. That is safe, but it is not written down as a breaking change for downstream repos.

5. **Test coverage**
   - The card merge, the cap, closed-issue refusal and a racing registration on the same card are covered, both on real tracker cards and through the fake.
   - Nothing covers a race between two different cards. That is a known limit of a per-card compare-and-swap; reading the registry is the backstop.

6. **Architecture, marker by marker**
   - **ARCH-DRY: pass.** BR-8 is resolved.
   - **ARCH-PURE: pass.** `internal/flakes` does no IO, and the verbs reach IO through `flakeDeps`.
   - **ARCH-PURPOSE: pass for M1's scope.** The gate itself is M2 work.
   - **ARCH-MOCK: pass.** `flakeWorld` keeps state across calls, and the tracker-card test runs against `newTrackerRepo`.
   - **ARCH-CONSTRAINTS: pass.** The cap is `MaxRerun` (5), and the store is read with `testStatsRecordLimit`.
   - **ARCH-SECURE: pass.** BR-1 still holds: import-path segments cannot start with `-` or `.`, and the rerun command is built as an argv list.
   - **ARCH-ORDER: pass for the card itself.** Writes go through the compare-and-swap, and a changed card gives `ErrCardChanged` and asks for a rerun. The cross-card limit is the first Minor above.
   - **ARCH-FUNERAL: pass.** A registration ends when its issue closes. The temporary details directory is removed with `defer os.RemoveAll`.
   - **For M2:** the gate must check the new `flake_tests` field at the moment it decides, not rely on an earlier list. The fail-closed `BuildRegistry` already gives it that.

7. **Plan revisions:** none needed. M1's checklist matches the code: `flakes` package, v2 attempt fields, `register` and `list`.

```findings
dispose:
  - id: BR-2
    disposition: addressed
    note: |
      addFlakeTests merges, caps and checks closed status on the card the CAS hands to decide (cardsetters.go:68); making it ignore the card's tests turns TestAddFlakeTestsMergesOntoTheCard and TestFlakeRegistryFromTrackerCards red.
  - id: BR-8
    disposition: addressed
    note: |
      registry dep now calls loadFlakeRegistry (which returns cards); suiteStoreDir (flake_cmd.go:128) is the single store-dir helper used by attempts and logPath.
findings:
  - id: new
    severity: Minor
    kind: defect
    family: cas-mutator-uses-stale-read
    title: |
      The one-owner-per-test check reads the registry, not the card under compare-and-swap, so concurrent registrations on two issues can both pass
    detail: |
      This is the 2nd finding in family cas-mutator-uses-stale-read. The rule: an invariant a mutator enforces must either be checked on the snapshot the CAS protects, or be enforced again where the data is read. The cross-card uniqueness here can only be the second kind. BuildRegistry (registry.go:71) already refuses a test named by two issues, which fails closed and visibly. The fix is to state that rule in the register help and atlas row, not to add another check at the write site. Measured prevalence: 1 cross-card site (flake_cmd.go:294-304).
  - id: new
    severity: Minor
    kind: defect
    family: breaking-change-undocumented
    title: |
      The suite-attempt writer now records version 2, which older sdlc binaries reject as an unsupported version
    detail: |
      teststats_store.go:50 and localcheck_run.go:190: an older binary reading the same store skips or flags v2 records. Note this as a breaking change for downstream repos when the release ships.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"2a2cafb7-c774-4abd-9cb6-7c16c4ebd1ac","repository":"github.com/xianxu/ariadne","issue_id":"000394","recipe":"milestone-review","round":3,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/ea18e1af9ed61283","reviewed_head":"9baa43c83bf3abb2623fd48cdd7acec651730e05","slot":"ariadne:2","requested_at":"2026-10-18T23:54:42.537406Z"},"phase":"finished","pid":27945,"started_at":"2026-10-18T23:54:42.646538Z","finished_at":"2026-10-18T23:56:18.771914Z","session_id":"dbb666a8-4eb0-499d-9bdb-08e228ed9f4e","outcome":"returned"}] -->
