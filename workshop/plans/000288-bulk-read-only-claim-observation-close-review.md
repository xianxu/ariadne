# Boundary Review — ariadne#288 (whole-issue close)

| field | value |
|-------|-------|
| issue | 288 — Bulk read-only claim observation |
| repo | ariadne |
| issue file | workshop/issues/000288-bulk-read-only-claim-observation.md |
| boundary | whole-issue close |
| milestone | — |
| window | afe9485ae10316d9847ce29e82eedb3c1bcb0e84..a356366940f08254cf9ad6b73e20c8de5fe82af7 |
| command | sdlc close --issue 288 |
| reviewer | claude |
| timestamp | 2026-10-02T14:42:23-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

All three open findings were checked against the code at `a3563669`. Two are fixed. The third, BR-15, is fixed in its main point but has one false clause left, and it is Minor. BR-8's fix is structural: every card-field lookup now goes through `Records.Require`, and a grep finds no `rs.Get` read left outside it. BR-16 is fixed in both the renderer and the test. BR-15's plan wording now says, correctly, that a tracker fetch on inventory is unbounded. But it says the mechanism is "the command's context cancels it", and the code shows no context is passed into either git call. That is one wrong clause in the plan; it doesn't block shipping. The targeted tests pass (`TestOneMalformedCard*`, `TestFleetInventoryPlacesClaims`, `TestFleetClaims*`), and so do the `fleet` and `tracker` packages.

1. **Strengths**
   - `Records.Require` at `cmd/sdlc/internal/tracker/records.go:117` mirrors `Snapshot.Require`. Every lookup site goes through it: `repoguard.go:104`, `pr.go:175`, `observe.go:49/97`, `close.go:514`, `actual.go:172`, `push.go:560`, `issue.go:602`, `projectstatus.go:285`, `issuefiles.go:85`. This fixes the whole family `unreadable-card-read-as-absent`, not just one site.
   - The not-done guard is proven to fail closed. The added assertion in `malformedcard_test.go:98` calls `guardIssueNotDone` directly, so the failure is no longer hidden by a later `snap.Require`.
   - `records_test.go` covers `Require` for both an unreadable card and a missing record.

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings**
   - BR-15 is still open. The plan's network-budget paragraph says "the command's context cancels it". In fact `LoadRecords` calls `repo.Presence()`, which runs `RemoteExists` (`ls-remote` with a nil context, `gitx/candidate.go:151`), and then `repo.Snapshot()`. Neither call takes a context; `ctx.Err()` is only checked after both return (`records.go:157`). Ctrl-C does stop the git child, so "until it fails or is interrupted" is accurate. Fix: delete the clause about the context.

5. **Test coverage notes**
   - BR-16 has no regression test, and none is possible: `MachinePresent` and `ClaimsPresent` are both the string `"present"`. The fix changes meaning, not behaviour, so it is accepted on inspection.
   - If you want the compiler to catch this mix-up, give the machine state and the claims state separate named string types.

6. **Architectural notes for upcoming work**
   - Inventory's fetches can't be cancelled through the context because `Presence`, `Initialized` and `Snapshot` don't take one. This matches the declared budget (ARCH-CONSTRAINTS passes as declared). Passing a context through the `tracker.Repository` git calls would be a separate improvement.
   - ARCH check for this round's changes:
     - ARCH-DRY: pass. `Require` is now the single lookup.
     - ARCH-PURE: pass.
     - ARCH-PURPOSE: pass. The fix was applied to every site, not just the one named.
     - ARCH-MOCK: pass, no new external seam.
     - ARCH-SECURE: pass. Unreadable input now fails closed.
     - ARCH-ORDER: not applicable. These changes hold no state between events; they are one-shot reads.
     - ARCH-FUNERAL: not applicable. The fixes create nothing durable; they change only reads and rendering.

7. **Plan revision recommendations**
   - Drop "the command's context cancels it" from the M2 network-budget revision, or say that inventory's git calls take no context and only an interrupt ends a stalled fetch.

```findings
dispose:
  - id: BR-8
    disposition: addressed
    note: |
      Records.Require added; all rs.Get card reads (repoguard, pr, observe, close, actual, push, issue, projectstatus, issuefiles) routed through it; guardIssueNotDone fail-closed asserted in malformedcard_test.go.
  - id: BR-15
    disposition: not-addressed
    note: |
      Plan now states unbounded (good), but asserts the command ctx cancels the fetch; Presence/RemoteExists (nil ctx) and Snapshot take no ctx, ctx.Err() is only checked after. Delete that clause.
  - id: BR-16
    disposition: addressed
    note: |
      render.go:137 and fleetclaims_test.go:99 now use MachinePresent; same string value so no behavioral test can fail, verified by inspection.
```
