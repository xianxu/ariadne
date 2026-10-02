# Boundary Review — ariadne#288 (milestone M1)

| field | value |
|-------|-------|
| issue | 288 — Bulk read-only claim observation |
| repo | ariadne |
| issue file | workshop/issues/000288-bulk-read-only-claim-observation.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | afe9485ae10316d9847ce29e82eedb3c1bcb0e84..ff6bab32bc0bc801645b102d47d74121fc52a00d |
| command | sdlc milestone-close --issue 288 --milestone M1 |
| reviewer | claude |
| timestamp | 2026-10-02T14:04:49-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

M1 does what it set out to do. `parseSnapshot` now keeps a card it cannot parse as an `UnreadableCard` instead of failing the whole read. Structural errors (mode, path, filename, duplicate ID, manifest, size limits) still fail the read. Every write-path lookup listed in the plan now goes through `Snapshot.Require`, which names the parse error. `transferguard` fails closed on any unreadable card. A quarantined card keeps its ID and path, so `issue new` cannot reuse them. The tracker, observe and fleet packages pass, and so do the targeted `cmd/sdlc` tests (`MalformedCard|Transfer|Close_Milestone|LintIDs|Observe`).

Two things should be fixed before shipping, both cheap:
- **Silent skip in the landing path.** The plan says the `trackercompletion.go:39` skip is unreachable because `transferguard` refuses first. That is false for `completeLandingPR` (called from `landing.go:456`) and for `settleLandedCompletions` (called from issue recovery). Neither runs the guard, so an unreadable codecomplete card is skipped with no message.
- **Untested reader branches.** Several of the new `CardErr` branches have no test that exercises them.

1. **Strengths**
   - `reader.go:158-171`: the line between the tracker's structure and one card's content is drawn cleanly. The duplicate check runs before parsing and uses `taken()`, so an unreadable copy still collides with a readable one. This is tested ("duplicate id, one unreadable").
   - `Require` versus `Card` is a good API split. Compare-and-swap sites keep `Card`'s readable-only contract, where an unreadable current card correctly reads as `ErrCardChanged`. Verbs get the precise cause.
   - `records.go:53-58`: for a tracked record with no card, `Field()` returns unknown rather than the mirror's value, and the pure test `TestComposeRecordsCarriesUnreadableCards` covers it.
   - `claimable()` and `validateAddition` count unreadable cards toward the ID, path and entry-count budgets. `wireBytes` already includes every blob, so the size budgets stay exact.
   - The end-to-end test (`malformedcard_test.go`) uses the real stateful tracker fixture and checks behavior across many views: claim of the readable card succeeds, claim of the bad card refuses, `issue show --json` reports unknown, `issue list` and drift report it unreadable, `transferguard` refuses, and `issue new` allocates past it.

2. **Critical findings:** none.

3. **Important findings**
   - **`cmd/sdlc/trackercompletion.go:39`** — `ownedCompletions` silently skips `rec.Card == nil`, which now includes cards with `CardErr`. The plan calls this unreachable for publishing, but it is reached by:
     - `completeLandingPR` at `landing.go:456`, which runs after the PR has merged;
     - `settleLandedCompletions` at `issuerecovery.go:105` and `publishgate.go:349`.

     If a card became unreadable after its PR was opened, the landing archives anyway and leaves the card codecomplete with no message. This is the "an unreadable card reads as absent" class that M1 exists to close.
     
     **Fix:** have `ownedCompletions` return an error for any `rec.CardErr` (it may be a completion this landing owns), or at least warn. Add a test, and correct the plan table row.
   - **Missing tests for the per-site `CardErr` branches.** These branches have no test:
     - `close.go:515` (die)
     - `actual.go:173`
     - `push.go:561` (`historyFileIsTerminal`)
     - `projectstatus.go:286`
     - `issuefiles.go:85` (`overlayCardStatus`)
     - `fleet/issues.go:47` (`LookupRepoIssues`)
     - `issue.go:614` (text `issue show`)
     
     The plan's end-to-end step also lists an `issue set-status` refusal that the test does not assert. These are hand-written branches, and a regression would turn "unknown" back into "absent".
     
     **Fix:** in the same fixture, extend `TestOneMalformedCardDoesNotBlockOthers` with a few direct calls: `historyFileIsTerminal`, `lookupIssueMeta`, `fleet.LookupRepoIssues`, and the setter.

4. **Minor findings**
   - `issuefiles.go:85`: the plan table says "status/card shown as unreadable", but the code fails the whole scan. Fail-closed is defensible here, because every caller is a publish path that `transferguard` already refuses. Record it as a revision.
   - `transferguard.go:49`: only the first unreadable card is named. If several are bad, the operator repairs them one per round trip.
   - `state.go:60`: the new `unreadable` JSON field on `IssueState` changes the `state --json` output. It is additive, but check whether a versioned schema contract covers it.
   - `closetracker.go:130` (`retryCloseMirror`) warns and returns. That is correct, but untested.

5. **Test coverage notes**
   - Pure tests (`TestParseSnapshotQuarantine`, `TestCreateRefusesAnUnreadableID`, `TestComposeRecordsCarriesUnreadableCards`) cover the snapshot and composition well, with no IO.
   - There is no pure `observe` assemble test for `CardErr`; only the end-to-end test covers it. Adding one is cheap.
   - The hermetic fix to `TestClose_MilestoneRefusesWithRedirect` is the right shape (a scratch SDLC repo), but its trigger was never isolated, as the Log itself says.

6. **Architecture (one line per principle)**
   - **ARCH-DRY:** pass. Every write-path lookup was folded into `Require`, and `taken()` is shared by parse and create. The repeated `errors.Is(ErrNoCard)` / `else if err` shape in claim, move and closetracker reflects different policies, so it is acceptable.
   - **ARCH-PURE:** pass. Quarantine lives in pure `parseSnapshot` and `composeRecords`, and the IO shell is unchanged.
   - **ARCH-PURPOSE:** flag (the first Important finding). The audit's claim of "never absent" has one reachable silent-skip site. The rest of the reader enumeration was swept.
   - **ARCH-MOCK:** pass. The end-to-end test uses the real stateful tracker fixture.
   - **ARCH-CONSTRAINTS:** pass. There are no new reads; quarantine adds an O(1) map entry per bad card, and budgets include unreadable entries.
   - **ARCH-SECURE:** pass. Card bytes are untrusted, parse failure is reported visibly rather than replaced with a made-up value, and nothing touches secrets.
   - **ARCH-ORDER:** pass. The snapshot holds no state between events; CAS sites treat an unreadable current card as changed.
   - **ARCH-FUNERAL:** pass. Nothing durable is created; quarantined cards exist only in memory.
   - **For M2:** `LookupRepoClaims` must turn `Records` entries with `CardErr` into `claims_state: partial` (or unknown), never drop them. This is the same class as the first Important finding.

7. **Plan revision recommendations**
   - Revise the `trackercompletion.go:39` row: it is reachable from landing and recovery, and the new behavior is to refuse or warn naming the card.
   - Revise the `issuefiles.go:85` row: it fails the scan (fail-closed for publish callers) rather than showing the status as unreadable.

```findings
findings:
  - id: new
    severity: Important
    family: unreadable-card-read-as-absent
    title: |
      ownedCompletions silently skips unreadable cards on landing/recovery paths that transferguard does not guard
    detail: |
      trackercompletion.go:39 skips rec.Card == nil (now including CardErr). completeLandingPR (landing.go:456, post-merge) and settleLandedCompletions (issuerecovery.go:105, publishgate.go:349) do not run transferguard, so a codecomplete card that became unreadable is never completed and no one is told; the plan's "unreachable" claim is false. Refuse or warn naming the card, add a test, and revise the plan row.
  - id: new
    severity: Important
    family: per-site-branch-untested
    title: |
      Per-reader CardErr branches (close, actual, push, projectstatus, issuefiles, fleet LookupRepoIssues, issue show text, set-status) have no test
    detail: |
      Only claim, observe, listIssueStates/drift, transferguard and issue new are exercised by TestOneMalformedCardDoesNotBlockOthers; the plan's e2e also promised an issue set-status refusal. Extend the same fixture with direct calls to these sites.
  - id: new
    severity: Minor
    family: plan-table-drift
    title: |
      overlayCardStatus fails the whole scan rather than showing the card unreadable as the plan table states
    detail: |
      Fail-closed is defensible since every caller is a publish path transferguard refuses; record it as a plan revision.
  - id: new
    severity: Minor
    family: error-names-one-of-many
    title: |
      transferguard names only the first unreadable card
    detail: |
      With several malformed cards the operator repairs them one per round trip; list every ID.
```

---

## Re-review — 2026-10-02T14:11:17-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 288 — Bulk read-only claim observation |
| repo | ariadne |
| issue file | workshop/issues/000288-bulk-read-only-claim-observation.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | afe9485ae10316d9847ce29e82eedb3c1bcb0e84..26cede27b5943ee731f8cfbecc5adaad867307da |
| command | sdlc milestone-close --issue 288 --milestone M1 |
| reviewer | claude |
| timestamp | 2026-10-02T14:11:17-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

Commit 26cede27 fixes BR-3 for real. `ownedCompletions` (`cmd/sdlc/trackercompletion.go:39`) now refuses on an unreadable non-duplicate card instead of skipping it, so every caller fails closed: landing, landing-archive, publish-gate, branch-owned and settle. Recovery (`issuerecovery.go:105`) shows the refusal as a visible `cwarn` rather than dropping it. The new assertion in `TestOneMalformedCardDoesNotBlockOthers` would go red without the fix, because the old `rec.Card == nil → continue` returned no error. The plan's reader table is now keyed by function instead of stale line numbers, and a Revisions entry records it. Transferguard joins every unreadable card. The targeted tests pass (`./`, `internal/tracker`, `internal/observe`). One reader branch is still untested: the milestone-mode `CardErr` die in `close.go:515`. That is the only thing between this boundary and SHIP, and it is cheap to fix.

1. **Strengths**
   - `trackercompletion.go:39` refuses with the card ID wrapped (`card #%s: %w`), the same way a malformed completion binding is refused. The `!rec.Duplicate` guard is correct: the primary record already carries the error.
   - `transferguard.go:49-54` uses `errors.Join` across every entry from `Unreadable()`, so all broken cards are named in one pass (BR-6).
   - `malformedcard_test.go:65-100` runs real production readers against one real fixture: `runSetStatus`, `historyFileIsTerminal`, `lookupIssueMeta`, `overlayCardStatus`, `fleet.LookupRepoIssues`, `ownedCompletions`, `actualTrackerInputs` and `runIssueShow`. Each must name the cause. There are no mocks.
   - `close_test.go:132-136` makes a test hermetic that had been taking the real checkout's lock. This is a good side fix.
   - The new rule in `workshop/lessons.md` ("unreachable because guard X" means listing every caller) covers the whole class behind BR-3, not just this instance.

2. **Critical:** none.

3. **Important**
   - **BR-4 is still open:** the `close.go:515-516` milestone-mode die (`#N: <CardErr>`) is the one `rec.CardErr` reader the shared fixture doesn't call. The package already has die-catching tests (`die_test.go`, `closeflow_test.go`), so the fix is cheap.
     - **This is the 2nd finding in family `per-site-branch-untested`.** The rule: every `rec.CardErr` branch in production code must be exercised by `TestOneMalformedCardDoesNotBlockOthers`.
     - The enumeration is mechanical: `grep -rn 'CardErr' cmd/sdlc --include='*.go' | grep -v _test` gives 11 sites. All are covered except `close.go:515`.
     - Fix: add a `close --milestone` call on #42 to the fixture under the die-catching helper, asserting the cause ("fingerprint"). Then state the grep as the coverage rule in the test's doc comment.

4. **Minor**
   - **Partial progress at landing:** `completeLandingPR` refusing on an unrelated unreadable card also stops healthy owned completions after the merge has already happened. Settle re-derivation recovers this once the card is repaired, and an `sdlc merge` is refused earlier by transferguard, so this is acceptable. Still worth one line in the atlas quarantine section: "landing completion defers until the card is repaired."
   - **BR-6 has no multi-card test:** no test uses two malformed cards. The code obviously iterates, so this is not blocking.

5. **Test coverage notes**
   - I checked by reasoning that the BR-3 regression goes red without the fix; I did not revert it in a scratch copy.
   - `observe_test.go:366` pins the `unknown` card status with derived sections blanked and still validates. This is good pure coverage.
   - Only `close.go:515` is uncovered (see Important).

6. **Architecture**
   - **ARCH-DRY: pass.** Each reader handles `CardErr` in a single local branch, with no copy-pasted helper logic.
   - **ARCH-PURE: pass.** `observe.Assemble` takes `CardErr` as pure input and is tested without IO.
   - **ARCH-PURPOSE: pass.** Every consumer reads quarantine from the single `Records.CardErr` source; the class sweep is done except for the test gap above.
   - **ARCH-MOCK: pass.** Tests run against a real git tracker fixture.
   - **ARCH-CONSTRAINTS: pass.** No new hot path.
   - **ARCH-SECURE: pass.** An untrusted card is parsed into a typed error at the boundary, and all readers degrade visibly rather than treating it as absent.
   - **ARCH-ORDER: pass.** Recovery is idempotent re-derivation (see the partial-progress Minor above).
   - **ARCH-FUNERAL: pass.** Nothing durable is added; the quarantine is derived on each read.

7. **Plan revision recommendations**
   - Once the close test lands, add `close.go` to the Revisions sentence that says "each reader branch is now exercised".

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Task 4 now uses a one-line "strategy:" test bullet; Task 2's enumerated bullet describes the delivered fixture and is moot.
  - id: BR-2
    disposition: addressed
    note: |
      Reader table rewritten keyed by function name (plan lines 170-182) with a Revisions entry; matches the grep of CardErr sites.
  - id: BR-3
    disposition: addressed
    note: |
      trackercompletion.go:39 refuses on CardErr; asserted by ownedCompletions call in malformedcard_test.go, which fails under the old skip; recovery surfaces it via cwarn.
  - id: BR-4
    disposition: not-addressed
    note: |
      All reader sites now exercised except close.go:515 milestone-mode CardErr die; add it to the fixture via the die-catching helper.
  - id: BR-5
    disposition: addressed
    note: |
      Plan row now records overlayCardStatus as fail-closed, plus a Revisions entry.
  - id: BR-6
    disposition: addressed
    note: |
      transferguard.go joins every Unreadable() entry via errors.Join.
findings:
  - id: new
    severity: Minor
    family: unreadable-card-read-as-absent
    title: |
      Landing completion is all-or-nothing on any unreadable card; document the deferral
    detail: |
      2nd in family; the rule (every CardErr reader refuses or reports, never skips) now holds at all 11 grep-enumerated sites. Remaining nuance: an unrelated malformed card defers healthy post-merge completions until repair; settle re-derivation recovers. Note it in the atlas quarantine section.
```

---

## Re-review — 2026-10-02T14:14:45-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 288 — Bulk read-only claim observation |
| repo | ariadne |
| issue file | workshop/issues/000288-bulk-read-only-claim-observation.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | afe9485ae10316d9847ce29e82eedb3c1bcb0e84..6b89b5584c89d06d37568d5f55606ee1aa87218a |
| command | sdlc milestone-close --issue 288 --milestone M1 |
| reviewer | claude |
| timestamp | 2026-10-02T14:14:45-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

M1 does what it set out to do. A card whose own content won't parse is now set aside as an `UnreadableCard` and no longer fails every tracker read. `Snapshot.Require` gives each verb that acts on one card a clear choice between "no card" and "unreadable card". Every reader that worked from a nil `Card` now either refuses or reports the cause. Push, PR and merge refuse repo-wide, and so do landing and recovery. A quarantined card keeps its ID and path, so they can't be handed out again. Both open findings are resolved:
- **BR-4:** the shared fixture now calls each reader directly. I removed the `projectstatus` and `fleet` fixes in a scratch copy and the test went red at those two exact checks.
- **BR-7:** the atlas now describes the landing deferral.

One Minor finding is left in the `unreadable-card-read-as-absent` family. Neither of its callers can reach it today, so it doesn't block SHIP.

**1. Strengths**
- `reader.go` separates structural failure (mode, path, duplicate ID, manifest, limits) from bad card content, which is the right line. The duplicate-ID check now counts unreadable copies (`s.taken` runs before parsing), and a test covers it ("duplicate id, one unreadable").
- `Require` with `ErrNoCard`/`ErrUnreadableCard` gives callers one consistent lookup. Claim and move keep their special "no card" handling via `errors.Is`. `Card` stays readable-only for compare-and-swap; a card that becomes unreadable mid-swap surfaces as `ErrCardChanged`, and the retry then refuses through `Require`.
- The quarantined card still counts in `MaxID`, `claimable`, `validateAddition` and the lint check in `issuelintids.go`. The test proves `issue new` allocates 000043 rather than reusing 42.
- `transferguard` now names every unreadable card via `errors.Join`, which closes BR-6.
- The `close_test.go` hermetic fix and the new `lessons.md` rule ("list the callers before claiming a guard covers them") target a real root cause.

**2. Critical findings**
None.

**3. Important findings**
None.

**4. Minor findings**
- **`cmd/sdlc/repoguard.go:104` (`guardIssueNotDone`).** This is the 3rd finding in family `unreadable-card-read-as-absent`.
  - **What's wrong:** for an unreadable card, `rec.Status()` returns `""`, so this guard passes. Its own doc says it "fails closed".
  - **Why it's masked today:** both callers (`changecode.go:130`, `startplan.go:60`) call `Require` right afterwards and refuse with the cause.
  - **The rule:** any card-owned read on an `IssueRecord` (`Card`, `Status()`, a card-owned `Field`) must check `CardErr` first. The earlier sweep grepped for `.Card == nil` and `snap.Card`, so it missed reads that go through `rs.Get(...).Status()/Field()`.
  - **Measured spread:** 10 `rs.Get` sites. Three don't check `CardErr`:
    - `repoguard.go:104` — fails open, but masked as above.
    - `pr.go:175` — safe: both callers (`pr.go:93`, `landing.go:520`) run after `guardTransferredDetailsFn`; I checked.
    - `observe.go:95` — harmless: the card is already shown as Unknown, and `status` only gates the "done" evidence.
  - **Structural fix:** add a `Records.Require(id)` that matches `Snapshot.Require` and returns `ErrUnreadableCard`. Make that the one lookup for card-owned reads (ARCH-DRY), and use it in `guardIssueNotDone`.

**5. Test coverage notes**
- `TestOneMalformedCardDoesNotBlockOthers` drives claim, observe, `issue show` (JSON and text), list/state/drift, transferguard, set-status, the archive terminal check, project status, the issue-file overlay, fleet lookup, `ownedCompletions`, actual, milestone close and `issue new`.
- These passed locally: the `tracker`, `observe` and `fleet` packages, and the targeted `cmd/sdlc` tests (51s).
- I removed two fixes in a scratch copy to confirm the test bites, and it failed at those two checks.
- `TestParseSnapshotQuarantine` covers three content-failure kinds, plus MaxID, `Card` and `Require` for all three outcomes.
- Gap: no test for `guardIssueNotDone` with an unreadable card. See Minor 1.

**6. Architectural notes**
- **ARCH-DRY:** pass. `Require` replaces about 10 copies of the "no card" message. The `Records`-level counterpart above is what's left.
- **ARCH-PURE:** pass. `parseSnapshot`, `composeRecords` and `assembleCard` stay pure and are tested without IO.
- **ARCH-PURPOSE:** pass for M1. All readers that check `CardErr` were swept; the family above names the remaining enumeration hole.
- **ARCH-MOCK:** pass. The tests use the existing tracker repo fixture (`newTrackerRepo`), with no stateless mocks.
- **ARCH-CONSTRAINTS:** pass. Entry-count and byte budgets still include unreadable cards (`validateAddition` counts them; `wireBytes` is computed before parsing).
- **ARCH-SECURE:** pass. Tracker content is treated as untrusted. Failure shows up as "unknown/unreadable" plus the cause, never as an invented absence or status. The mirror's possibly stale status is not used instead (`TestComposeRecordsCarriesUnreadableCards`).
- **ARCH-ORDER:** pass. The new state is per-read, nothing is held between events, and the compare-and-swap path is unchanged.
- **ARCH-FUNERAL:** pass. Nothing durable is created; quarantine lives only in memory per snapshot, and repair is the existing card edit.

**7. Plan revision recommendations**
None. The M1 rows of the core-concepts table (`UnreadableCard`, `Snapshot`/`Require`/`Unreadable`, the two errors, `IssueRecord.CardErr`, `observe.Inputs.CardErr`) all exist at their stated paths. The `fleet/claims.go` rows belong to M2.

```findings
dispose:
  - id: BR-4
    disposition: addressed
    note: |
      malformedcard_test.go now calls set-status, historyFileIsTerminal, lookupIssueMeta, overlayCardStatus, fleet.LookupRepoIssues, ownedCompletions, actualTrackerInputs, computeClose and runIssueShow text directly. With the projectstatus and fleet fixes removed in a scratch copy, the test fails at exactly those two checks.
  - id: BR-7
    disposition: addressed
    note: |
      atlas/workflow/issue-tracker.md quarantine section now states that landing and recovery refuse while any card is unreadable, deferring healthy completions until repair, with the next settle re-deriving them.
findings:
  - id: new
    severity: Minor
    family: unreadable-card-read-as-absent
    title: |
      guardIssueNotDone reads an unreadable card's status as not-done, contrary to its fail-closed contract
    detail: |
      3rd in family. Rule: every card-owned read on an IssueRecord (Card, Status(), card-owned Field) checks CardErr first. The prior sweep grepped .Card == nil and snap.Card, so it missed rs.Get(...).Status()/Field() reads. Of 10 rs.Get sites, 3 do not check CardErr: repoguard.go:104 (fails open, masked because start-plan and change-code call snap.Require next), pr.go:175 (safe, both callers run after transferguard), observe.go:95 (harmless, card already Unknown). Structural fix: add Records.Require(id) mirroring Snapshot.Require and route card-owned reads through it, starting with guardIssueNotDone.
```
