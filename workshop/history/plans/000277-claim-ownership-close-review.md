# Boundary Review — ariadne#277 (whole-issue close)

| field | value |
|-------|-------|
| issue | 277 — Record claimant ownership atomically with issue reservation |
| repo | ariadne |
| issue file | workshop/issues/000277-claim-ownership.md |
| boundary | whole-issue close |
| milestone | — |
| window | bb6221d84ecd5ba27a1336cb78515ecdb2e5d602..f48f8df3b25f5a26bd41e3cf180ca49d0a117cd5 |
| command | sdlc close --issue 277 |
| reviewer | claude |
| timestamp | 2026-10-01T15:28:01-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The two findings still open from earlier rounds are both fixed in the code at head. BR-8: `TestSlotLabelOnlyWhereSlotsExist` now checks the plain primary first, at `cmd/sdlc/claimant_test.go:42`, before it creates the malformed `ariadne-slotX` directory. BR-21: `startplan.go:296` and `changecode.go:347` now pass the card they already hold to `requireCardOwnership`. Change-code also refreshes the mirror from that same card through the new `refreshMirrorFrom`, so ownership and the mirror read one card version. The code builds, the `internal/issue` and `pkg/vocab` packages pass, and the targeted `cmd/sdlc` tests pass (`Ownership|Claimant|SlotLabel|VerbContract|Relocat|Adopt|SetStatus|Move|HelpFlags|Claim`, 212s). `gofmt` is clean on every file in the diff; its only hit, `reviewsidecar.go`, is outside this window. One new Minor finding: the milestone-close path still reads the card twice. Nothing blocks shipping.

1. **Strengths**
   - `requireCardOwnership` is split cleanly from `requireIssueOwnership` (`cmd/sdlc/claimant.go:145-185`). Every verb that already holds a card now uses the card form.
   - When the relocation probe fails, the error is now reported and wraps the cause (`claimant.go:177`), so it is no longer collapsed into a plain foreign-owner refusal.
   - `refreshMirrorFrom` pulls out the shared tail of `refreshMirror` instead of copying it (ARCH-DRY).
   - The ownership help text comes from one constant, `ownershipGateHelp`, through a template placeholder.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - `cmd/sdlc/close.go:510-520`: milestone close reads the status from `loadIssueRecords`, a `tracker.IssueRecord`, then opens the tracker again and snapshots it a second time for `requireIssueOwnership`. Status and owner can therefore come from two different card versions. This is the 3rd finding in family `repeated-tracker-snapshot`.

5. **Test coverage:** the verb contract table, the adopt race, the relocation success and Unknown branches, and the slot-label ordering are all covered and green. The milestone-close double read has no test that observes two card versions; it is a narrow window.

6. **Architecture**
   - ARCH-DRY: pass.
   - ARCH-PURE: pass. `MatchClaimant` and `RelocationAllowed` are pure; the IO is confined to `relocatable` and `ownership`.
   - ARCH-PURPOSE: pass. All four continuation verbs and `set-status` are gated.
   - ARCH-MOCK: pass. Tests use the stateful tracker repo fixtures.
   - ARCH-CONSTRAINTS: pass, apart from the Minor double read.
   - ARCH-SECURE: pass. The machine ID is fingerprinted and never stored raw.
   - ARCH-ORDER: pass, with one note. The ownership verdict reads one card version everywhere except milestone close (the Minor above).
   - ARCH-FUNERAL: pass. The relocation record is removed when nothing moved (fixed in earlier rounds).

7. **Plan revisions:** none needed.

```findings
dispose:
  - id: BR-8
    disposition: addressed
    note: |
      claimant_test.go:42 asserts the plain primary before writing worktree/ariadne-slotX at :45.
  - id: BR-21
    disposition: addressed
    note: |
      startplan.go:296 and changecode.go:347 call requireCardOwnership on the held card; changecode refreshes the mirror from the same card via refreshMirrorFrom.
findings:
  - id: new
    severity: Minor
    family: repeated-tracker-snapshot
    title: |
      Milestone close reads the status from loadIssueRecords, then opens and snapshots the tracker again for requireIssueOwnership
    detail: |
      3rd in family. Rule: each verb judges status and ownership from one card read; requireIssueOwnership only where no card or record is held. close.go:510-520 holds an IssueRecord (rs.Get) but opens the tracker again. Fix: reach the raw card through the record or a shared snapshot so both judgments read one version. Prevalence: start-plan, change-code, milestone close; the first two are now fixed.
```

---

## Re-review — 2026-10-01T15:38:41-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 277 — Record claimant ownership atomically with issue reservation |
| repo | ariadne |
| issue file | workshop/issues/000277-claim-ownership.md |
| boundary | whole-issue close |
| milestone | — |
| window | bb6221d84ecd5ba27a1336cb78515ecdb2e5d602..4d449496bb6c32fc2c4e758e750ab4b1b858380d |
| command | sdlc close --issue 277 |
| reviewer | claude |
| timestamp | 2026-10-01T15:38:41-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

**Verdict: SHIP.** This round fixed the open finding, BR-22. In a tracker-era milestone close, `computeClose` used to load the card record to read the status, then open and snapshot the tracker again for the ownership check. It now passes the card it already loaded (`rec.Card`) to `requireCardOwnership`. The old `requireIssueOwnership` is gone, so one gate function takes the card it is given. Start-plan, change-code and milestone close each now judge status and ownership on a single card read.

Whole-issue close is the one sibling left. It still reads status from `loadIssueRecords` and ownership from a second snapshot inside `prepareTrackerClose`. That makes the Log line "Every verb judges status and owner on one card read" slightly overstated. It is Minor and does not block the gate.

What I ran:
- **Build and vet:** both clean.
- **gofmt:** clean on the changed files. The only flagged file, `cmd/sdlc/reviewsidecar.go`, is not in this window.
- **Tests:** the ownership, claimant, verb-contract, relocation, set-status and move tests passed. `TestMoveDetailWithoutLocalDetailsDerivesThemFromTheCard` failed once while removing its temp directory, then passed 3 out of 3 when re-run alone. It looks like a cleanup flake under parallel load, not something this diff introduced.

1. **Strengths**
   - `cmd/sdlc/close.go:498-519`: milestone close judges ownership on the card version it read the status from. The non-nil guarantee holds because the code dies at line 506 when `rec.Card == nil`.
   - `cmd/sdlc/claimant.go:141-145`: one gate function instead of two near-duplicates, which removes the snapshot-or-card split (ARCH-DRY).
   - The atlas entry and the help-template comment were renamed in the same commit, so the docs stay in step with the code.

2. **Critical:** none.

3. **Important:** none.

4. **Minor** (one item)
   - **This is the 4th finding in family `repeated-tracker-snapshot`.**
     - **Rule:** a verb's status check and its ownership check must read the same card. A precondition helper should take the card from its caller, not snapshot again on its own.
     - **Remaining instance:** whole-issue close. `close.go:500-508` reads status via `loadIssueRecords`. `prepareTrackerClose` (`closetracker.go:330-338`) snapshots again for ownership and for the card it publishes.
     - **Fix:** have `prepareTrackerClose` take the already-loaded `tracker.Record`, or take `currentStatus` from `trackerPrep.card` when `mode == "issue"`.
     - **Prevalence:** 4 verbs had the pattern; 3 are fixed.

5. **Test coverage**
   - BR-22 is a refactor that keeps behaviour the same, so no new test is needed.
   - The existing milestone-close ownership tests still pass and still go through the gate.

6. **Architecture**
   - ARCH-DRY, ARCH-PURE, ARCH-MOCK, ARCH-CONSTRAINTS, ARCH-SECURE and ARCH-FUNERAL: pass, nothing new in this round.
   - ARCH-ORDER: pass. The time window between two card reads is now gone in 3 of the 4 verbs.
   - ARCH-PURPOSE: flagged. The one-read rule is met in every verb except whole-issue close; that is the Minor above.

7. **Plan revisions:** none. The Log claim should either be narrowed to milestone close, or made true by the Minor fix.

```findings
dispose:
  - id: BR-22
    disposition: addressed
    note: |
      close.go:508,518 judges ownership on rec.Card from the status read; requireIssueOwnership removed; start-plan/change-code already single-read.
findings:
  - id: new
    severity: Minor
    family: repeated-tracker-snapshot
    title: |
      Whole-issue tracker close still reads status (loadIssueRecords) and ownership (prepareTrackerClose snapshot) from two card reads
    detail: |
      4th in family. Rule: status and ownership must be judged on the same card, so a precondition helper takes the caller's card instead of snapshotting again. Remaining instance: close.go:500-508 vs closetracker.go:330-338. Fix: pass the loaded tracker.Record into prepareTrackerClose, or take currentStatus from trackerPrep.card when mode == "issue". Prevalence: 4 verbs had it, 3 fixed. The Log line "Every verb judges status and owner on one card read" overstates this until it is fixed.
```

---

## Re-review — 2026-10-01T15:49:01-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 277 — Record claimant ownership atomically with issue reservation |
| repo | ariadne |
| issue file | workshop/issues/000277-claim-ownership.md |
| boundary | whole-issue close |
| milestone | — |
| window | bb6221d84ecd5ba27a1336cb78515ecdb2e5d602..d130b35c36108a69cfc375487325df8ad72c55e4 |
| command | sdlc close --issue 277 |
| reviewer | claude |
| timestamp | 2026-10-01T15:49:01-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

BR-23, the only open finding, is fixed in d130b35c. A whole-issue tracker close used to read the card's status through `loadIssueRecords` and then judge ownership on a second card read inside `prepareTrackerClose`. It now runs `prepareTrackerClose` first and takes `currentStatus` from that same card (`trackerPrep.card.Card.Frontmatter`), at `cmd/sdlc/close.go:499-509`. The `loadIssueRecords` read is left only for milestone close, which already judges ownership on the record it loaded (`close.go:527`). I checked every caller of `requireCardOwnership`: `changecode.go:347`, `close.go:527`, `closetracker.go:338` and `startplan.go:296`. Each one passes a card the caller already holds. None of them takes a second snapshot, so the rule is now enforced everywhere, not just at the sites earlier findings named. I ran the stat and name-status recipes on the window (46 files, +3790/−82). The new commit only reorders the close path and doesn't add any new user-facing surface.

1. **Strengths**
   - `close.go:499-509`: one card read now gives status, the ownership verdict and the tracker base for the receipt (`trackerRef`). All three judge the same card version.
   - `requireCardOwnership` (`claimant.go:145`) is the single ownership gate. Every caller hands it the card it already holds (ARCH-DRY).
   - The fix still runs before the review and before any write, so the BR-22 ordering of preconditions is kept.
   - The Log entry admits the earlier claim was overstated, and it records that this round swept the whole class of verbs rather than one site.

2. **Critical:** none.

3. **Important:** none.

4. **Minor:** none new.

5. **Test coverage notes:** I found no test written specifically to pin "whole-issue close does exactly one card read". The change is a pure reordering with no new branch, and the existing tests for close ownership and status still exercise the path. That's acceptable for a Minor follow-through, and I'm not raising it.

6. **Architectural notes**
   - ARCH-DRY passes.
   - ARCH-PURE passes; the gate is a pure check on a card that is passed in.
   - ARCH-PURPOSE passes; all eight verbs in the class are covered.
   - ARCH-MOCK has no change in this round.
   - ARCH-CONSTRAINTS passes; one fewer tracker read on close.
   - ARCH-SECURE has no change.
   - ARCH-ORDER passes. Status and owner now come from the same card version, which removes a time-of-check/time-of-use gap between the two reads.
   - ARCH-FUNERAL has nothing new that persists.

7. **Plan revisions:** none needed.

```findings
dispose:
  - id: BR-23
    disposition: addressed
    note: |
      close.go:499-509 now takes currentStatus from prepareTrackerClose's card (same read that runs requireCardOwnership and supplies trackerRef); all requireCardOwnership callers pass a caller-held card.
```
