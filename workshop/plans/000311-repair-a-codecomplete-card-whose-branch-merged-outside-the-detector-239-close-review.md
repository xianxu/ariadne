# Boundary Review — ariadne#311 (whole-issue close)

| field | value |
|-------|-------|
| issue | 311 — Repair a codecomplete card whose branch merged outside the detector (#239) |
| repo | ariadne |
| issue file | workshop/issues/000311-repair-a-codecomplete-card-whose-branch-merged-outside-the-detector-239.md |
| boundary | whole-issue close |
| milestone | — |
| window | 612c7ad98089baa6365f3fa3df5a12f56566eba7..4063fc10f67b847e5ec952df8b8773a6ce76e998 |
| command | sdlc close --issue 311 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-12T18:38:56-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Checking for duplicated logic against existing helpers (`trackercompletion.go`, `gitx/window.go`), then running the tests.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

**Summary.** The `--landed` repair matches the design in the Spec. A commit given with `--landed` must be on main. Its `C^1..C` range must contain a commit whose subject cites the issue. Only a codecomplete card with no close binding moves to `done`, and the existing reconcile sweep then archives it. The card write goes through `ChangeCard`, which re-reads the card on every retry, so a rerun does nothing. Every refusal leaves the remote unchanged, and the tests check that. The targeted suite (`Landed|CitesIssue|DoneUnbound|Recovery|ExternalMerge|Help`) passes against the pinned head. Two cheap fixes remain. First, `citesIssue` adds a second copy of the "subject opens with #N" rule, which `gitx.subjectAnchorRE` explicitly claims as its single source. Second, the first Done-when item (#239's card reads `done`) is not met inside this window. It was moved to a post-landing step, and the issue does not say so in its Done-when.

1. **Strengths**
   - `doneUnboundCard` (`cmd/sdlc/externalmerge.go:373`) is a pure transition with no IO. It has the same shape as the existing `doneCard` (`trackercompletion.go:156`): `done` returns ErrNoChange, `codecomplete` proceeds, anything else refuses.
   - All git-side checks (commit resolves, commit is on main, range cites the issue) run before any tracker write. `TestReconcileLandedRefuses` checks for every case that `originMain` and the card bytes are unchanged.
   - The fixture uses a renamed branch that is merged with `--no-ff` and is not the card's slug. It also adds a `#3300:` decoy on main. Together these pin both the #239 shape and the digit-boundary rule.
   - The Spec's ARCH reasoning holds up. A git-only check works with a deleted branch and with a squash merge, and nothing new is created except the one card write.

2. **Critical:** none.

3. **Important**
   - **ARCH-DRY: `citesIssue` duplicates `subjectAnchorRE`** (`cmd/sdlc/externalmerge.go:366`, against `internal/gitx/window.go:264`). The gitx helper's comment calls itself "the single source" for the anchor rule. The new regex restates that rule and changes it (it accepts zero-padding). It also skips the bookkeeping-subject filter that `ShippedWorkOnMain` applies, so a range holding only `#N: plan:` commits would be accepted as landed work. Fix: export an anchor helper from gitx and call it here, adding padding tolerance in gitx if it is wanted. Also decide whether bookkeeping subjects should count.
   - **ARCH-PURPOSE: Done-when item 1 is unmet.** "#239's card reads `done`, and `sdlc state` no longer reports it" is the stated purpose. The Log only records the repair as a post-landing step (`sdlc issue recovery reconcile --issue 239 --landed b066449c`). Either run the repair before close (the branch binary can do it) or revise Done-when/Plan to say explicitly that #239's repair happens after landing. That way the close verdict does not claim a result that has not happened.

4. **Minor**
   - The bound check runs before the status switch in `doneUnboundCard`. So a card that is already `done` and still bound gets an error instead of the "done is a no-op" result the Spec promises. Check status first, or at least handle `done` first.
   - Any `rev-parse` error, such as a repo or IO failure, is reported as "no such commit here (fetch main first)", which hides the real cause (`externalmerge.go:318`).
   - `citesIssue` recompiles its regex on every call. This is not a hot path, so it only matters if the DRY fix above does not replace it.
   - Only the pure `doneUnboundCard` test covers a bound card. No end-to-end test asserts that a bound refusal through `runRecoveryReconcileLanded` leaves the remote unchanged. The card is protected by `ChangeCard` returning the mutate error, so the risk is low.

5. **Test coverage notes.** Covered: the renamed and deleted branch path to done, archived, and gone from state; idempotent rerun; refusals for not-on-main, no citing commit, unknown commit, and a card that is not closed; a bound card (pure test); and subject matching (padding, `#3300`, mid-subject and PR-title references). Not covered: a squash or single-commit `--landed` (the C^1..C range is just the commit itself), and a done card that is still bound.

6. **Architectural notes**
   - ARCH-DRY: flagged above.
   - ARCH-PURE: pass. The transition is pure, and `repairLanded` is a thin IO shell.
   - ARCH-PURPOSE: flagged (Done-when deferral).
   - ARCH-MOCK: pass. The tests use real temp git repos and a peer clone through the existing `trackerRepo` and `testfix` seam, as the rest of the package does.
   - ARCH-CONSTRAINTS: pass. `git log` is bounded by one merge's range.
   - ARCH-SECURE: pass. Operator input goes to git as a single argument. Only the resolved SHA is used after that, and parse failures produce visible errors.
   - ARCH-ORDER: pass. There is one conditional card commit that is re-derived on each retry. If the process is interrupted between the card write and the archive, rerunning the sweep recovers it, and the cache is invalidated before the sweep.
   - ARCH-FUNERAL: pass. Nothing new is durable beyond the one card write.

7. **Plan revision recommendations**
   - Add a `## Revisions` entry: "Done-when item 1 (#239 reads done) is a post-landing step: run `sdlc issue recovery reconcile --issue 239 --landed b066449c` with the merged binary. The close does not attest it." Alternatively, run the repair before close and drop this entry.

```findings
findings:
  - id: new
    severity: Important
    family: single-source-helper-bypassed
    title: |
      citesIssue re-implements gitx.subjectAnchorRE, the declared single source for the subject-opens-with-#N anchor
    detail: |
      externalmerge.go:366 restates the anchor regex (adding zero-padding) instead of reusing internal/gitx/window.go:264, and skips the bookkeeping-subject filter ShippedWorkOnMain applies, so a range of only `#N: plan:` commits passes. Export a gitx anchor helper and call it.
  - id: new
    severity: Important
    family: done-when-deferred-unrecorded
    title: |
      Done-when item 1 (#239's card reads done, state clean) is not met in the window; moved post-landing only in the Log
    detail: |
      Either run `reconcile --issue 239 --landed b066449c` before close or add a Revisions entry amending Done-when, so the close does not attest an outcome that has not happened.
  - id: new
    severity: Minor
    family: noop-precedence
    title: |
      doneUnboundCard checks the binding before status, so a bound done card errors instead of being a no-op
  - id: new
    severity: Minor
    family: error-cause-masked
    title: |
      repairLanded maps any rev-parse error to "no such commit here", hiding git/IO failures
```
