# Boundary Review — ariadne#285 (whole-issue close)

| field | value |
|-------|-------|
| issue | 285 — Transfer guard: owner plus based-on-latest |
| repo | ariadne |
| issue file | workshop/issues/000285-transfer-guard-owner-latest.md |
| boundary | whole-issue close |
| milestone | — |
| window | 8d8c8f947d876942ac28e0c7aabf26f6cf07fd63..1c3faf37eff30e2e5ea8a5cf23165ce278c76b7c |
| command | sdlc close --issue 285 |
| reviewer | claude |
| timestamp | 2026-10-08T13:58:44-07:00 |
| verdict | unknown |

## Review

API Error: Couldn't connect through your proxy (ERR_PROXY_TUNNEL) — the proxy refused the tunnel: check its credentials and that it allows this host

---

## Re-review — 2026-10-08T14:02:49-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 285 — Transfer guard: owner plus based-on-latest |
| repo | ariadne |
| issue file | workshop/issues/000285-transfer-guard-owner-latest.md |
| boundary | whole-issue close |
| milestone | — |
| window | 8d8c8f947d876942ac28e0c7aabf26f6cf07fd63..1c3faf37eff30e2e5ea8a5cf23165ce278c76b7c |
| command | sdlc close --issue 285 |
| reviewer | claude |
| timestamp | 2026-10-08T14:02:49-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The work does what the issue asked for. The guard no longer looks at branch names or handoff records. It now covers any details file that main's history has touched. For each one that the prospective merge would change, it asks two questions: is this checkout the card's claimant (`ownership`)? Does HEAD contain main's last commit to that file? Each Done-when item has a real-git test. pair#365 is reproduced end to end, with the close on a renamed branch, then `pr`, then `merge`, ending with the card done. A stale filing branch is refused and `issue restore` lets it land. An owner who is behind is refused, then lands after merging main, in both the clean-merge and conflict variants. A walk over all help text checks that no "owner" means a branch. The focused tests ran green at HEAD 1c3faf37: `go test ./cmd/sdlc -run 'TransferGuard|PublishGateAndPR|DetailsVerdict|IssueRestore|CloseOnARenamedBranch|NoHelpCallsABranch|OneMalformedCard|Recovery'` took 121.8s. Only Minor findings remain.

**1. Strengths**
- **Pure verdict:** `cmd/sdlc/transferverdict.go` keeps the decision pure, and `TestDetailsVerdict` covers all 8 combinations of facts. The plan's Revisions entry for dropping the `Conflict` fact is correct: a branch based on main's last commit can't conflict on that file, and a conflicted file already shows up in the merge-tree diff.
- **One shared collector (ARCH-DRY):** `changedDetails` feeds both the guard's refusal and `issue restore`, so the two can't disagree about which files are refused.
- **Narrower failure on unreadable cards:** these now refuse only the landings that change that card's details (`transferguard.go` ~84). `malformedcard_test.go` checks both directions.
- **Interrupted restores:** the restore handles them well. `restorable` accepts a file that already holds main's version, and `TestIssueRestoreResumesAnInterruptedRestore` pins this. A recovery-catalog contract was added alongside.
- **The behind case is really tested:** per the Log, a mutation check found that every behind fixture also conflicted, so a clean-merge behind case was added. That closes the gap where the behind check was never tested on its own.

**2. Critical:** none.

**3. Important:** none.

**4. Minor**
- **`issue restore` ignores the verdict.** `cmd/sdlc/issuerestore.go` ~73 selects by ID only (`want[c.ID]`), so `verdictBehind` entries are included. An owner who runs restore while behind will quietly revert their own edit. The help text presents restore as an action for non-owners. Fix: restore only `verdictNotOwner` entries, or refuse when this checkout is the owner and point to the merge-main action.
- **Atlas paragraph runs on.** In `atlas/workflow/issue-tracker.md` ~284, the new "Transfer guard" paragraph is followed on the same line by the unrelated "`sdlc issue recovery list|reconcile` resumes…" sentence. Start a new paragraph there.
- **Reopen needs a claim.** `set-status` writes a reopen Log entry into the details file. That edit is now refused for anyone except the claimant, with a hint to claim. This matches the Spec, but it is a behavior change worth a line in the reopen help text.

**5. Test coverage notes**
- Everything runs against real git fixtures. There are no function mocks: the guard's published, owner and based facts all come from real repositories.
- The pair#365 end-to-end test runs the real publish gate (`NoJudge=false`).
- Restore has no test for an owner who is behind (see Minor 1).

**6. Architectural notes (each principle pass/flag)**

| Principle | Result | Why |
|---|---|---|
| ARCH-DRY | pass | One collector. |
| ARCH-PURE | pass | Pure verdict; thin git collector. |
| ARCH-PURPOSE | pass | Every Done-when item is delivered and branch names are fully out of the guard. |
| ARCH-MOCK | pass | Real git plus the existing gh fake behind `ghClient`. |
| ARCH-CONSTRAINTS | pass | One merge-tree and one diff, plus a rev-list and an ancestry check per changed details file. |
| ARCH-SECURE | pass | Malformed cards and claimants fail closed, scoped to their own files. |
| ARCH-ORDER | pass | An interrupted restore converges on rerun; the guard holds no state between events. |
| ARCH-FUNERAL | pass | Creates nothing durable beyond an ordinary branch commit. |

Watch for: under D4, owners whose branches landed by squash will often hit "behind". That is intended, but the merge-main next action should stay prominent.

**7. Plan revision recommendations:** none needed. The existing Revisions entry already matches the code: three facts, and no refusal for a published file without a card.

```findings
findings:
  - id: new
    severity: Minor
    family: verb-scope-wider-than-advertised
    title: |
      issue restore also reverts verdictBehind entries, discarding an owner's own edit
    detail: |
      issuerestore.go selects changedDetails by ID only; filter to verdictNotOwner or refuse in the owner's checkout with the merge-main action.
  - id: new
    severity: Minor
    family: atlas-prose-structure
    title: |
      atlas issue-tracker.md transfer-guard paragraph runs into the unrelated recovery sentence
  - id: new
    severity: Minor
    family: behavior-change-undocumented-in-help
    title: |
      a reopen Log edit to published details now needs a claim; note it in set-status reopen help
```
