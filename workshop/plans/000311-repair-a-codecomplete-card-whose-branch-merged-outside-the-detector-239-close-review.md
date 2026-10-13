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

---

## Re-review — 2026-10-12T18:44:05-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 311 — Repair a codecomplete card whose branch merged outside the detector (#239) |
| repo | ariadne |
| issue file | workshop/issues/000311-repair-a-codecomplete-card-whose-branch-merged-outside-the-detector-239.md |
| boundary | whole-issue close |
| milestone | — |
| window | 612c7ad98089baa6365f3fa3df5a12f56566eba7..e71649e74406111e46462d76af9be657173b7cba |
| command | sdlc close --issue 311 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-12T18:44:05-07:00 |
| verdict | SHIP |

## Review

Targeted tests pass (141s). I checked the prior findings against the code and am writing up the verdict.

```verdict
verdict: SHIP
confidence: high
```

This is ready to ship. The round-1 findings are fixed in code. `citesIssue` now delegates to the existing exported `gitx.IsShippedWorkSubject`, so there is one anchor rule and bookkeeping-only ranges no longer count. Done-when item 1 is formally moved to after landing in a `## Revisions` entry. `doneUnboundCard` checks status before the binding, and `rev-parse` failures are now told apart from a missing commit. The targeted run passed in the pinned checkout: `go test -run 'Landed|CitesIssue|DoneUnbound|Recovery|Help|ExternalMerge'` returned `ok` in 141.8s. The only open item is Minor: BR-4's fix has no regression test.

Done-when: current. Item 2 (a test for a card whose branch name differs from its slug) is delivered by `landedFixture` / `TestReconcileLandedRepairsAnUnboundCard`. The 2026-10-12 Revisions entry moves item 1 (#239 repaired) to after landing, because the new flag only exists in the merged binary. That gives a recorded reason and a concrete command (`reconcile --issue 239 --landed b066449c`).

## 1. Strengths
- `cmd/sdlc/externalmerge.go:368`: `citesIssue` now uses the shared classifier in gitx (`window.go:190`). The test table (`landedrepair_test.go:159`) pins `#330: close` → false and `sdlc: #330 implement` → true, which are the convention's edge cases.
- `doneUnboundCard` is a pure `[]byte → []byte` transition, tested without IO (`TestDoneUnboundCardRefusesABoundCard`). The IO shell `repairLanded` does the checks in order: commit exists, commit is on main, its range cites the issue, then the card write.
- `TestReconcileLandedRefuses` checks that every refusal leaves both `origin/main` and the card unchanged, not just that an error comes back.
- The end-to-end test reproduces #239's actual shape: a renamed branch, merged `--no-ff`, then gone. It also runs the command a second time and asserts nothing changes.
- The atlas (`issue-tracker.md:302`) and the `reconcile` long help both document the new flag.

## 2. Critical findings
None.

## 3. Important findings
None.

## 4. Minor findings
- BR-4: the fix (`gitTest`, which separates exit 1 from other failures) is correct, but no test drives a git failure other than exit 1 through `repairLanded`.
- `repairLanded` runs `rev-parse` twice (first `gitTest --verify`, then `git rev-parse`). Capturing the SHA from the first call would save one subprocess. Cosmetic.

## 5. Test coverage notes
The refusal table covers: not on main, a bookkeeping-only range, an unrelated commit, an unknown commit, and a card that isn't closed. The bound-card and done-bound-card cases are unit tests. A git error other than exit 1 is not covered (BR-4).

## 6. Architecture
- **ARCH-DRY: pass.** The duplicate regex is gone and the gitx helper is reused.
- **ARCH-PURE: pass.** `doneUnboundCard` and `citesIssue` are pure; the clock is read in the shell.
- **ARCH-PURPOSE: pass.** The mechanism is delivered, and repairing #239 itself is a recorded post-landing step.
- **ARCH-MOCK: pass.** Tests run against temp-repo git, which is the repo's existing convention, and no new external dependency is added.
- **ARCH-CONSTRAINTS: N/A.** It is a one-shot operator command over a single range.
- **ARCH-SECURE: pass.** The commit ref is passed as an argv element, an unknown ref is refused visibly, and a git failure now surfaces as an error.
- **ARCH-ORDER: pass.** The write goes through `ChangeCard` with an operation token, `ErrNoChange` makes reruns idempotent, and the rerun is tested.
- **ARCH-FUNERAL: pass.** It creates nothing durable except one card write, and the existing sweep archives the details.

## 7. Plan revision recommendations
None. The Revisions entry already matches the code and the post-landing step.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      citesIssue now calls exported gitx.IsShippedWorkSubject (window.go:190), the anchor plus bookkeeping filter; tests pin the "#330: close" refusal and the "only bookkeeping" range refusal.
  - id: BR-2
    disposition: addressed
    note: |
      A 2026-10-12 Revisions entry moves item 1 to after landing, with the exact repair command; item 2 is delivered in the window.
  - id: BR-3
    disposition: addressed
    note: |
      The status switch now runs before CardCompletion; TestDoneUnboundCardRefusesABoundCard asserts ErrNoChange for a bound done card, which would fail under the old order.
  - id: BR-4
    disposition: not-addressed
    note: |
      The code fix (gitTest separates exit 1 from other errors) is correct, but no regression test drives a git failure other than exit 1 through repairLanded. Minor, non-blocking.
```
