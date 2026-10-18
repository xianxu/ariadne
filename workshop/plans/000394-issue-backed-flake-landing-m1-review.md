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
