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
