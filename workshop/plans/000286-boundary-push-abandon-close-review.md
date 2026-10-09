# Boundary Review — ariadne#286 (whole-issue close)

| field | value |
|-------|-------|
| issue | 286 — Boundary pushes and sdlc abandon |
| repo | ariadne |
| issue file | workshop/issues/000286-boundary-push-abandon.md |
| boundary | whole-issue close |
| milestone | — |
| window | 3f92e76e4940e04b23721d548148deab1874e130..27fda9eefc3b42f5849d0122fc0d1433acd4b0d4 |
| command | sdlc close --issue 286 |
| reviewer | claude |
| timestamp | 2026-10-08T17:22:48-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

I ran the stat and name-status recipes on the pinned window (37 files, +2921/−48), then read the code in full for the new and changed surface: `abandon.go`, `boundarypush.go`, the `setstatus.go` redirect and reopen path, and the round-4 fix commit `7584eb96` together with its test diff. The issue's Spec and Done-when are delivered:
- `start-plan`, `milestone-close` (both paths), `close` and reconcile's close completion push the branch. `unclaim` already pushed through the same `leasedBranchPush`.
- `merge` deletes the remote branch.
- `abandon` keeps the work, writes the card, archives the details and drops the branch, and each step is recognised again on a rerun.
- A reopen restores the branch and un-archives the details.

Only BR-9 is still open. The ownership check on abandon's resume path is now in the code (`cmd/sdlc/abandon.go:163-169`), but no test refuses a foreign resume; `grep "was abandoned by"` finds no test. Under the rule that a behavior fix counts only if a test fails without it, it is not addressed. It is Minor and cheap to close. Nothing blocks.

**Strengths**
- `boundarypush.go:59-74`: the lease is the last-fetched remote-tracking ref, never a fresh read, and the tracking ref is updated after the push. So a rebase followed by the next boundary force-pushes safely on a lease, which matches Done-when 1.
- `boundarypush.go:103` `remoteRefTip` is now the single read of a remote ref. `abandon.go:225`, `:389` and `setstatus.go:418` all reuse it, which settles BR-14 (ARCH-DRY).
- `pushArchiveRef` (`abandon.go:224-244`) and `dropAbandonedBranch` (`:386-424`) only overwrite or delete what the kept tip contains. Every destructive step is leased and refuses when the remote has moved.
- `statusDecision` (`setstatus.go:139-155`) keeps both guards in the pure core, and `--force` waives neither: the redirect from started work to `abandon`, and the BR-13 rule that a kept branch leaves terminal status only by reopening.
- `restoreAbandoned` auto-resolves a conflict only when every conflicted path is the issue's own details or plans (`resolveArchiveConflicts`). Any other conflict is left in progress with the card unchanged.

**Critical:** none.

**Important:** none.

**Minor**
- BR-9 (re-disposed below): the resume ownership guard at `abandon.go:165-169` has no regression test. To fix it, abandon issue 9 partway, rewrite the card's claimant to another workspace, rerun `abandon9`, and assert the "only that workspace finishes it" error.
- `setstatus.go:388-392`: every `set-status working` on an issue that is already working, from an issue branch, now makes an `ls-remote` call through `unfinishedReopen`. It is harmless but adds latency. Gating it on the recorded abandoned ref being absent and terminal history being present is optional.

**Test coverage notes**
- BR-7, BR-12 and BR-13 each have a targeted test in `7584eb96` (`TestAbandonKeepsTheBranchsPlans`, `TestReopenRerunVariants`, `TestAbandonedWorkLeavesOnlyByReopen`). The tests assert the observable state on origin and in HEAD, not the calls made.
- The only gap left is the foreign-resume case (BR-9).

**Architecture**
- ARCH-DRY pass: there is one leased push, one leased delete and one remote-ref read, all over `gitFn`.
- ARCH-PURE pass: `abandonDecision` and `statusDecision` are pure; the IO stays in `runAbandon` and `restoreAbandoned`.
- ARCH-PURPOSE pass: every boundary verb in the Spec pushes, and reopening is delivered too rather than deferred. The done-reopen un-archive gap is filed as #305, which is a separable extension.
- ARCH-MOCK pass: the tests run against real local git remotes through `testfix`.
- ARCH-CONSTRAINTS pass: boundary pushes are capped at 2 minutes and the process group is killed on timeout. Abandon's own pushes are unbounded, but they hard-fail the way unclaim's do, which is consistent.
- ARCH-SECURE pass: the card's abandoned record is parsed through `issue.CardAbandoned` (all or none). The kept tip is checked against the fetched ref before a branch is recreated.
- ARCH-ORDER pass: the steps converge, each one is recognised again on a rerun, and four interruption points are tested. A resume is recognised from the card (recorded record plus terminal status), not from flags.
- ARCH-FUNERAL pass: a reopen deletes the archive ref, a later abandon overwrites a leftover one, and `merge` and `abandon` delete the remote branch. A `wontfix` ref is kept on purpose, one per abandoned issue, which is an acceptable bound.

**Plan revisions:** none. The table was corrected under BR-11 and still matches the code.

```findings
dispose:
  - id: BR-9
    disposition: not-addressed
    note: |
      The ownership guard on the resume path exists (abandon.go:164-169), but no test refuses a foreign resume, so nothing fails if the guard is removed.
```
