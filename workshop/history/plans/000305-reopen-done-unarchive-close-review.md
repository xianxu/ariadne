# Boundary Review — ariadne#305 (whole-issue close)

| field | value |
|-------|-------|
| issue | 305 — Reopening a done issue does not un-archive its details |
| repo | ariadne |
| issue file | workshop/issues/000305-reopen-done-unarchive.md |
| boundary | whole-issue close |
| milestone | — |
| window | 08afcec554b1d5cb4ff03b84f57352274df5db1d..dbe0f25e9a99988e09323c13a7cc75a72408d090 |
| command | sdlc close --issue 305 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-13T13:12:00-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Reviewing the #305 diff; next I'll run the new tests and check the revert-sensitivity of the regression tests.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The change does what the issue asks. In a tracked repository, `set-status working` on a `done` card whose details are archived on main now runs `restoreDone`. That function cuts the issue branch from main, moves the details and plans back out of the archive in one commit, and pushes the branch before the card write. A new helper, `unarchiveIssue`, now holds the un-archive step that `restoreAbandoned` used to carry inline; both callers share it. The new tests, together with the existing reopen, abandon, catalog and help tests, pass at the pinned head (`go test -run 'TestReopen|TestAbandon|TestCatalog|TestHelp|TestSetStatus|TestCheckTransitionGuards'` → ok). One thing should be fixed first. I wrote a throwaway probe test (since removed) and confirmed that the reopen never completes when this checkout still has a stale `origin/<branch>` tracking ref for the old landed branch. The error tells the operator to "rerun", and the probe shows a rerun after a plain `git fetch` fails the same way.

1. **Strengths**
   - **ARCH-DRY extraction is real.** `unarchiveIssue` (`cmd/sdlc/abandon.go:87`) is now the only un-archive code. `restoreAbandoned` calls it at `abandon.go:44` and no copy of the old loop remains.
   - **The refusal really changes nothing.** `reopendone_test.go:418-421` checks HEAD, the current branch, origin/main, the remote branch and the card after a refusal. The guard reads the archived copy before any branch or commit is made (`reopendone.go:250-262`).
   - **The end-to-end assertion is meaningful.** `assertDoneReopened` checks that the branch contains origin/main, that both files moved, that the Log entry came along, that the remote tip equals HEAD, that the tree is clean, that the card says working, that `guardTransferredDetails` passes, and that there is exactly one un-archive commit. That covers the Done-when directly.
   - **The reset path for a leftover local branch is exercised.** My probe showed the fixture leaves a local `000330-e2e` that main contains, so the `switch -C name main` branch (`reopendone.go:276-277`) runs in both tests.
   - **Rename of `reopenAbandoned` to `restoreForReopen` / `reopenedIssue`** is consistent, and `finishReopen` now skips the archive-ref drop when there is no ref (`abandon.go:137`).

2. **Critical findings:** none.

3. **Important findings**
   - **A stale tracking ref blocks the reopen permanently (ARCH-ORDER: the rerun never converges)** — `cmd/sdlc/reopendone.go:300`.
     - **Cause:** `pushIssueBranch` → `leasedBranchPush` uses the local `refs/remotes/origin/<name>` as its lease. That ref is easily stale for a landed branch, for example when origin deleted the branch (GitHub auto-delete, or `sdlc merge` run from another slot) after this checkout last fetched.
     - **Probe:** after `doneArchived` plus `writeReopenEntry`, I ran `update-ref refs/remotes/origin/000330-e2e <landed tip>`. The first reopen failed with `(stale info) … — the card is unchanged; rerun`. After `git fetch -q origin`, the rerun failed identically. Only `fetch --prune` would clear it.
     - **Fix:** in step 1, when the local branch is judged landed (main contains it), also drop `refs/remotes/<remote>/<name>` if main contains its tip too. Alternatively, refresh just that ref with a pruning fetch before the push. Add the probe above as a regression test.

4. **Minor findings**
   - **A lost finish isn't recognised on rerun** (`setstatus.go:393`). If the card write landed but `finishReopen`'s mirror commit or push did not, a rerun sees `working` and calls `unfinishedReopen`. That returns nil because there is no archive ref, so the mirror commit and push are never finished. The catalog's "a lost response is settled by rerunning" holds only for the card. Either recognise the done case (on the issue branch, details live, mirror dirty or unpushed) or narrow the catalog wording.
   - **Error after a branch switch** (`reopendone.go:279` → `293`). When a local branch exists that main does not contain (e.g. a squash-landed branch), the code switches to it and only then fails with "predates main's archive". The operator is left on that branch. Check `archiveCommit` against the tip before switching.
   - **Untested refusals:** foreign owner (`:225`), wrong branch (`:228`), dirty tree besides the details (`:247`), and the `--force` skip. Each is a cheap table case on `doneArchived`.
   - **Stricter than its message:** `archiveCommit` is `rev-list -1 main -- archivedRel` (`:286`). That is the *latest* commit to touch the archived copy, not the archiving commit. The check is still correct, but "predates main's archive" can be misleading if a peer edited the history copy later.
   - **ARCH-PURE:** `restoreDone` mixes step decisions (landed or not, which file the details are in) with git IO. That matches the existing `restoreAbandoned` style. A pure planning step could be split out if a third reopen kind ever arrives.

5. **Test coverage notes**
   - The two new tests pin real behaviour against real git, with a local bare origin.
   - The happy path, the guard refusal and the failed-push rerun (from the issue branch, un-archiving once) are covered.
   - Not covered: a stale or remote-present tracking ref for the old branch (the Important finding), the lost-finish rerun, and the refusal branches listed above.

6. **Architectural notes (ARCH-* walk)**
   - **ARCH-DRY: pass.** The shared `unarchiveIssue` is used by both callers, and the guard, ownership, fast-forward and push helpers are reused rather than re-implemented.
   - **ARCH-PURE: minor flag** (above).
   - **ARCH-PURPOSE: pass.** The Done-when (details and plans live on the reopened branch, and the next landing passes the transfer guard) is delivered and asserted by `guardTransferredDetails`. Help, the recovery catalog and both atlas pages derive from it.
   - **ARCH-MOCK: pass.** Tests use the established temp-repo plus bare-origin fake. No new external dependency.
   - **ARCH-CONSTRAINTS: pass / N/A.** This is a one-shot CLI with a bounded number of git calls.
   - **ARCH-SECURE: pass.** The archived details are parsed, and a malformed file produces a visible refusal naming the path (`:257`).
   - **ARCH-ORDER: flag.** Step recognition on rerun is sound for push failures, but the stale-lease case never converges and the lost finish isn't recognised.
   - **ARCH-FUNERAL: pass.** Nothing durable is created beyond the issue branch, which its landing removes. No archive ref is involved.
   - **Docs gate:** `atlas/workflow/issue-tracker.md` and `issue-lifecycle.md` are updated, and so is `helptext/set-status.md`. No new user-facing command or flag, so no README change is needed.

7. **Plan revision recommendations**
   - The plan still matches the code.
   - If the lost-finish case is left as is, add a `## Revisions` entry saying a done reopen's mirror commit and push are best-effort and not resumed by a rerun.
   - Add a revision noting the stale tracking-ref handling once it is fixed.

```findings
findings:
  - id: new
    severity: Important
    family: rerun-must-converge
    title: |
      Done reopen never converges when origin/<branch> tracking ref is stale for the landed branch
    detail: |
      reopendone.go:300 pushes via leasedBranchPush, which leases on the local remote-tracking ref; if origin deleted the landed branch after this checkout fetched, the push fails with stale info every time, and a plain fetch (no prune) does not clear it, though the error says rerun. Verified by a probe (update-ref refs/remotes/origin/000330-e2e to the landed tip, then reopen twice with a fetch between). Fix: when the local branch is judged landed, also drop the tracking ref if main contains it (or prune-fetch that ref) before pushing; add the probe as a regression test.
  - id: new
    severity: Minor
    family: rerun-must-converge
    title: |
      A lost finish (mirror commit/push) after a done reopen's card write is not resumed by rerun
    detail: |
      restoreForReopen sees status working and calls unfinishedReopen, which needs an archive ref the done path never has, so finishReopen never runs again; the catalog's "a lost response is settled by rerunning" covers only the card here.
  - id: new
    severity: Minor
    family: refuse-before-mutate
    title: |
      restoreDone switches to an unlanded local branch before refusing it as predating main's archive
    detail: |
      reopendone.go:279 switches, then :293 refuses, leaving the operator on the old branch; check archiveCommit against the tip before switching.
  - id: new
    severity: Minor
    family: refusal-paths-untested
    title: |
      Foreign-owner, wrong-branch, dirty-tree and --force paths of restoreDone have no tests
```

---

## Re-review — 2026-10-13T13:43:24-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 305 — Reopening a done issue does not un-archive its details |
| repo | ariadne |
| issue file | workshop/issues/000305-reopen-done-unarchive.md |
| boundary | whole-issue close |
| milestone | — |
| window | 08afcec554b1d5cb4ff03b84f57352274df5db1d..9ada9339305dc258994e5011703965425a727c17 |
| command | sdlc close --issue 305 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-13T13:43:24-07:00 |
| verdict | REWORK |

## Review

Running the reopen tests and reading the test file next.

Tests pass (361s). Now mutation-checking the BR-1/BR-2/BR-3 fixes by reverting each and re-running its regression test.

All three mutations go red, so the fixes are reachable and the tests actually pin them. Checking the docs changes and a few edge paths next.

Docs look fine. One path worries me: when the operator is already *on* a landed issue branch, `restoreDone` skips the `switch -C … main` reset. I'm writing a probe test for that case.

```verdict
verdict: REWORK
confidence: high
```

All four findings from round 1 are fixed, and I checked each fix by reverting it: the BR-1, BR-2 and BR-3 regression tests all fail without their fix. BR-4's refusal and `--force` tests are now in `TestReopenDoneRefusals`. The full set of reopen, set-status, abandon, catalog and help tests passes on head (361s). One new bug blocks shipping. If you run `set-status working` while still checked out on the issue's *landed* branch (for example, the slot left behind after a GitHub-button merge and a reconcile), `restoreDone` never resets the branch to main. It reports "details un-archived … and pushed", but nothing was un-archived. It pushes the old pre-landing tip and flips the card to working. That leaves the branch in exactly the state #305 exists to prevent: the details are live on a branch that does not contain main's archive. I confirmed this with a probe test, shown under Critical findings below.

**1. Strengths**
- `unarchiveIssue` (`cmd/sdlc/abandon.go:455`) is a clean extraction. Abandon and done reopens share one un-archive commit (ARCH-DRY), and it does nothing once the files are already out.
- `leaseOnRemoteBranch` (`cmd/sdlc/reopendone.go:136`) fixes BR-1 properly. It reads the remote with `ls-remote`, drops or refreshes the tracking ref, and refuses a foreign tip. It keeps the "lease on the last-fetched copy" rule for every other caller.
- BR-3 is now ordered correctly: the "predates main's archive" check runs before any `switch` (`reopendone.go:94-101`). The mutation test shows the refusal fires before anything changes.
- `assertDoneReopened` checks the whole end state: based on main, details and plans live, Log entry kept, pushed, clean tree, card working, transfer guard passes, and exactly one un-archive commit. These checks would catch the new bug if a test started from the issue branch.
- `unfinishedDoneReopen` only recognises a branch where the un-archive commit, plus at most the reopen's own mirror commit, is on top. Later work never gets swept into the finish.

**2. Critical findings**
- **Reopen from the checked-out landed branch skips the cut from main** (`cmd/sdlc/reopendone.go:87-112`; ARCH-ORDER, ARCH-PURPOSE)
  - When `env.branch == name`, `tip` is `"HEAD"`. If HEAD is contained in main, `landed` is true. The `switch -C name main` only runs `if env.branch != name`, so the branch is never reset to main.
  - On that old tip, HEAD has no archived copy, so `unarchiveIssue` does nothing. The old landed tip is pushed, the card goes to working, and the ok line claims an un-archive happened.
  - Probe: `doneArchived` → `switch 000330-e2e` → write the Log line at the live path → reopen. Result: err=nil, card working, `merge-base --is-ancestor origin/main HEAD` false, no un-archive commit.
  - The state model has a cell (on the branch × landed) that the code doesn't handle.
  - Fix sketch, either one:
    - Refuse up front, before the fast-forward or any other change: "#N: <branch> has landed; switch to <rest> and rerun".
    - Or treat this cell like the rest case: switch to main first, then cut the branch.
  - Either way, `success-without-postcondition` is the rule: before reporting success, check that HEAD contains `archiveCommit` and has the live details. Then any unenumerated cell fails loudly instead of claiming an un-archive.
  - Add this probe as a case in `TestReopenDoneRefusals` (or as a success test using `assertDoneReopened`).

**3. Important findings**
- None.

**4. Minor findings**
- **The guard's "Nothing was changed" can be false** (`reopendone.go:47-49` vs `:75`). The resting branch is fast-forwarded before the dirty-tree check and the Log guard run. The spec orders it that way on purpose, since the archived copy has to be in the checkout before the entry can be written there. But when rest was behind main, the refusal still says "Nothing was changed".
  - **This is the 2nd finding in family `refuse-before-mutate`.** The rule: every refusal is decided before the first change, except a change the refusal's own remedy depends on. When that happens, the refusal says so ("rest fast-forwarded to main; nothing else changed").
- Dry-run of a done reopen still goes through the old path (`restoreForReopen` returns nil on `DryRun`), so it reads a live details path that doesn't exist. This predates #305; noting it only.

**5. Test coverage notes**
- BR-1 is red without `leaseOnRemoteBranch` ("stale info … rerun"). BR-2 is red without `unfinishedDoneReopen` ("origin's branch "" is not HEAD"). BR-3/BR-4 are red without the predates check: the switch aborts on the dirty archived copy.
- Missing: a reopen started from the issue branch itself, either landed (the Critical above) or this reopen's own branch on a first run.
- Missing: a refusal after the fast-forward when rest is behind main.

**6. Architectural notes**
- **ARCH-DRY:** pass. The un-archive and the mirror subject are each defined once.
- **ARCH-PURE:** flag (advisory). The landed / merged / refuse decision in `restoreDone` is mixed in with git calls. A pure `decideDoneReopen(onBranch, tipInMain, tipHasArchive, …)` would have made the missing cell visible and testable without git.
- **ARCH-PURPOSE:** flag. The Critical bug defeats the issue's Done-when on a reachable path.
- **ARCH-MOCK:** pass. Tests run real git against a portable bare origin in a temp dir.
- **ARCH-CONSTRAINTS:** N/A. One `ls-remote` and fetch per reopen is a cold path.
- **ARCH-SECURE:** pass. A malformed archived copy is refused visibly (`reopendone.go:68-71`).
- **ARCH-ORDER:** flag. Rerun recognition rests on git evidence, which is acceptable, but the (current branch × landed × archived) cells aren't written down, and one is wrong.
- **ARCH-FUNERAL:** pass. The only durable thing created is the issue branch, which its landing removes.

**7. Plan revision recommendations**
- Add a `## Revisions` entry naming the reopen-from-the-landed-issue-branch case and the behaviour chosen for it (refuse, or switch to main first). Also add the postcondition check before the ok line.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      leaseOnRemoteBranch refreshes/drops the tracking ref; TestReopenDoneWithAStaleTrackingRef goes red with the call removed (stale info).
  - id: BR-2
    disposition: addressed
    note: |
      unfinishedDoneReopen wired after unfinishedReopen; TestReopenDoneRerunFinishesALostFinish goes red without it.
  - id: BR-3
    disposition: addressed
    note: |
      predates-archive check now precedes the switch (reopendone.go:94-101); refusal case red when the check is disabled.
  - id: BR-4
    disposition: addressed
    note: |
      TestReopenDoneRefusals covers foreign owner, other branch, dirty tree, stale local branch, and --force.
findings:
  - id: new
    severity: Critical
    family: success-without-postcondition
    title: |
      Reopen run while on the landed issue branch never cuts from main, un-archives nothing, yet pushes and flips the card
    detail: |
      reopendone.go:87-112 resets to main only when env.branch != name; on the landed branch HEAD lacks the archive, unarchiveIssue no-ops, the old tip is pushed and the ok line lies. Probe confirmed (not based on main, no un-archive commit, card working). Refuse or switch to main first, before any mutation, and verify HEAD contains archiveCommit with live details before reporting success; add the probe as a test.
  - id: new
    severity: Minor
    family: refuse-before-mutate
    title: |
      Log-guard refusal says "Nothing was changed" after the resting branch was fast-forwarded
    detail: |
      2nd in family. Rule: decide every refusal before the first mutation except a mutation the refusal's remedy needs, and then name it in the refusal. fastForwardRest (reopendone.go:47) precedes the guard (:75) by design.
```

---

## Re-review — 2026-10-13T14:12:30-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 305 — Reopening a done issue does not un-archive its details |
| repo | ariadne |
| issue file | workshop/issues/000305-reopen-done-unarchive.md |
| boundary | whole-issue close |
| milestone | — |
| window | 08afcec554b1d5cb4ff03b84f57352274df5db1d..983ab20e0424c72a4d2d3b0c2e6713a54ee7e272 |
| command | sdlc close --issue 305 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-13T14:12:30-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Tests pass. Now I'm re-running the probe without the Log-entry write, which was getting in the fast-forward's way.

Confirmed: the dirty-tree refusal fires after the fast-forward and doesn't mention it. Next I'll revert the BR-5 refusal and check that its test goes red.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The BR-5 fix holds where it matters. `restoreDone` now makes every branch decision before the first mutation (`cmd/sdlc/reopendone.go:46-79`), and a reopen run while on the landed branch is refused with no changes made. The new "the landed branch itself" case in `TestReopenDoneRefusals` covers it: I deleted the `case env.branch == name` refusal and the test went red. Two problems remain, neither blocking. First, the "postcondition" added for BR-5 (`:136-143`) checks nothing the landed branch didn't already satisfy. With the refusal removed, the old bug came back in full: the old tip was pushed and both ok lines printed. The postcondition never fired. Second, the BR-6 rule was applied only to the Log-guard refusal. The dirty-tree refusal still fires after the resting branch has been fast-forwarded and doesn't mention it (confirmed by probe).

**1. Strengths**
- `reopendone.go:50-79`: the branch decision is computed once, up front (`holds`/`landed`/`reset`), and the switch at `:121-131` only acts on it. This is the right shape for the refuse-before-mutate family.
- `reopendone.go:160-179`: `leaseOnRemoteBranch` re-reads origin before it leases, so it doesn't rely on a stale tracking ref (BR-1 stays fixed).
- `TestReopenDoneRefusals` checks one shared invariant (branch, HEAD, remote tip and card all unchanged) across every refusal case, not a per-case restatement of the code.
- `TestReopenDoneGuardRefusalNamesTheFastForward` checks both the wording and that the fast-forward really happened.
- The recovery catalog proof (`internal/recovery/catalog.go:227`) lists the new tests.

**2. Critical findings**
None.

**3. Important findings**
- **`reopendone.go:136-143`: the postcondition can't fail on the case it was added for (ARCH-PURPOSE).** This is the 2nd finding in family `success-without-postcondition`. It checks that HEAD has the live details and no archived copy. A landed branch that predates main's archive already has both properties, so the check passes on exactly the wrong starting point BR-5 described. I verified this by removing the `:74-75` refusal: the reopen pushed the landed tip and printed "un-archived … and pushed". The rule to apply: a postcondition must test the step's effect, which the starting state couldn't already satisfy. Here that means requiring `env.ancestorOf(archiveCommit, "HEAD")` as well as the live/archived checks. Pull it into a small helper and test it directly with a landed-branch HEAD, since the refusal above makes it unreachable from the command.

**4. Minor findings**
- BR-6 is not addressed (see the dispositions block). The dirty-tree refusal (`:104-106`) and the malformed-details refusal (`:112-115`) come after `fastForwardRest` (`:86`) and don't mention the move. Probe: rest at `HEAD~1`, `cmd/a.go` dirty → HEAD moved and the refusal said nothing about it. The dirty-tree check doesn't need the fast-forward, so the fix is to move it above `:81`, exempting both `archivedRel` and the live path. Any refusal that has to stay after the fast-forward should get the `moved` suffix through one shared helper.
- `reopendone.go:80-92`: `moved` is set whenever `before != view.Ref()`, including when `fastForwardRest` returned early because HEAD already equalled the target. That's harmless, but the condition repeats what `fastForwardRest` already knows (`claim.go:288`). Returning a "moved" bool from it would keep that in one place (ARCH-DRY, nit).

**5. Test coverage notes**
- I ran `go test -run 'TestReopen|TestCheckTransition' ./cmd/sdlc`; it passes (188s).
- BR-5 regression: reverting the fix makes `TestReopenDoneRefusals/the_landed_branch_itself` fail. I restored the file afterwards; the checkout is clean.
- No test can reach the postcondition, and none puts a dirty tree on a resting branch behind main. Both gaps above would ship unnoticed without them.

**6. Architectural notes**
- ARCH-DRY: pass, apart from the `moved` nit.
- ARCH-PURE: pass for this boundary. The decision logic still runs through live git calls; `restoreDone`'s decision block (`:50-79`) could later become a pure function of (tips, ancestry) → action.
- ARCH-PURPOSE: flagged — the hollow postcondition, and BR-6 swept as an instance rather than the class.
- ARCH-MOCK: pass. Tests run against real local git repos with a local origin, not production state.
- ARCH-CONSTRAINTS: N/A — a one-shot CLI with a few git calls.
- ARCH-SECURE: pass. The details file read is parsed and a malformed one is refused (`:112-115`).
- ARCH-ORDER: partially met. Rerun recognition (`unfinishedDoneReopen`) and decide-then-act are explicit, but the order "fast-forward, then refuse" is still implicit. The fix above makes it explicit.
- ARCH-FUNERAL: pass. Nothing durable is added beyond the issue branch, whose lifecycle is the existing land/abandon flow.
- Docs gate: atlas (`issue-tracker.md`, `issue-lifecycle.md`) and the `set-status` help text cover the done reopen, and no new flags were added, so a README update isn't needed.

**7. Plan revision recommendations**
- Add a Revisions entry saying the BR-5 postcondition requires HEAD to descend from main's archive commit, not just to have the live details.
- Make `workshop/lessons.md` (the BR-6 sentence) say "every refusal after a mutation names it", not just the one guard.

```findings
dispose:
  - id: BR-5
    disposition: addressed
    note: |
      The landed-branch refusal at reopendone.go:74-75 is decided before any mutation; reverting it turns TestReopenDoneRefusals/the_landed_branch_itself red. Its postcondition half is hollow and is raised as a new finding.
  - id: BR-6
    disposition: not-addressed
    note: |
      Only the Log guard names the move. The dirty-tree refusal (:104) and the malformed refusal (:114) still fire after fastForwardRest (:86) without naming it (probe: rest at HEAD~1 with cmd/a.go dirty, HEAD moved, refusal silent). Move the dirty check before :81; route post-ff refusals through one helper that appends moved.
findings:
  - id: new
    severity: Important
    family: success-without-postcondition
    title: |
      The BR-5 postcondition (reopendone.go:136-143) passes on the landed branch, so it guards nothing
    detail: |
      2nd in family. Rule: a postcondition must test the step's effect, which the starting state could not already satisfy. A landed branch already has the live details and no archived copy; with the :74 refusal removed the old tip was pushed and "un-archived … and pushed" printed. Require env.ancestorOf(archiveCommit, "HEAD") as well, in a helper tested directly.
```

---

## Re-review — 2026-10-13T14:47:02-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 305 — Reopening a done issue does not un-archive its details |
| repo | ariadne |
| issue file | workshop/issues/000305-reopen-done-unarchive.md |
| boundary | whole-issue close |
| milestone | — |
| window | c40e1b7f5ff15d1d81f5311815e8123101104046..225c7da3dc431a9162516d69ca96c27abc413d68 |
| command | sdlc close --issue 305 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-13T14:47:02-07:00 |
| verdict | SHIP |

## Review

Both fixes from the last round fail their tests when reverted, so they hold up. One last check on the remote-branch lease path, then the verdict.

```verdict
verdict: SHIP
confidence: high
```

Both open findings are fixed, and I checked each one by reverting the fix. BR-7: `checkUnarchived` now also requires that HEAD descends from main's archive commit. With that clause turned off, `TestCheckUnarchived` fails at `reopendone_test.go:255`, because the landed branch then passes the postcondition. BR-6: the dirty-tree refusal now runs before the fast-forward, and every refusal after the fast-forward names it through the `after` helper. Moving the dirty block back below the fast-forward makes `TestReopenDoneDirtyTreeRefusesBeforeTheFastForward` fail at `:277`. The targeted suite (`TestReopen|TestCheckUnarchived|TestCatalog|TestSetStatus`) passes on the pinned head, and the checkout was restored to clean after the reverts. One Minor issue remains: the lease refusal still runs after the branch switch and the un-archive commit. It doesn't block shipping.

**1. Strengths**
- `reopendone.go:46-79` makes every branch decision (holds the archive / landed / stale / landed-branch-itself) before the first change, keyed on one fact: `ancestorOf(archiveCommit, tip)`.
- `checkUnarchived` (`reopendone.go:229`) tests the effect of the step, which the wrong starting state can't already satisfy. It is tested directly on the landed branch.
- The un-archive step is shared between the abandon and done reopens (`unarchiveIssue`, `abandon.go`), so the logic lives in one place (ARCH-DRY).
- `leaseOnRemoteBranch` refreshes or drops a stale tracking ref instead of leasing the push on old local evidence. That preserves the BR-24 lease semantics.
- `dirtyBesides` is pure and covers both copies of the details.

**2. Critical:** none.

**3. Important:** none.

**4. Minor**
- `reopendone.go:147` / `:180`: the "holds work neither main nor this reopen has; inspect it before reopening" refusal fires after the switch (step 1) and the un-archive commit (step 2), and doesn't mention either. No test covers this path.
  - **This is the 3rd finding in family `refuse-before-mutate`.** The rule is already written in `workshop/lessons.md`: decide every refusal before the first change, unless the refusal's remedy needs that change, and if so name the change in the refusal.
  - Fix the rule, not this one site. List every `return …Errorf` after line 126, and for each one either decide it earlier or route it through `after` with the switch and commit named.
  - For the lease specifically: before switching, check `remote` against `{mainTip, the local tip of <name> if any}`. Before the switch, those two tips decide the same thing HEAD decides after it.

**5. Test coverage notes**
- Mutation checks confirmed for BR-6 and BR-7.
- The refusals table checks branch, HEAD, remote and card on every refusal.
- The foreign-work-on-origin lease path is the one refusal with no test.

**6. Architecture notes**
- **ARCH-DRY:** pass.
- **ARCH-PURE:** pass. `dirtyBesides` is pure; the rest is git glue.
- **ARCH-PURPOSE:** pass. The Done-when is met, and the transfer-guard landing is asserted in `assertDoneReopened`.
- **ARCH-MOCK:** pass. Tests run against a real git checkout with a local origin, which is the repo's existing fixture.
- **ARCH-CONSTRAINTS:** N/A. This is a one-shot CLI step.
- **ARCH-SECURE:** pass. The archived details are parsed with a visible error on malformed input, and git is called with argv arrays.
- **ARCH-ORDER:** pass, with the note above. Rerun convergence comes from observed git state (branch tip, commit subjects) rather than flags. The order of steps after the switch is still implicit.
- **ARCH-FUNERAL:** pass. Landing removes the issue branch, and a stale tracking ref is deleted.

**7. Plan revisions:** none.

```findings
dispose:
  - id: BR-6
    disposition: addressed
    note: |
      Dirty check moved above fastForwardRest; later refusals go through after(); TestReopenDoneDirtyTreeRefusesBeforeTheFastForward fails when the block is moved back below (verified).
  - id: BR-7
    disposition: addressed
    note: |
      checkUnarchived requires ancestorOf(archiveCommit, HEAD); TestCheckUnarchived fails on the landed branch with that clause disabled (verified).
findings:
  - id: new
    severity: Minor
    family: refuse-before-mutate
    title: |
      Lease refusal for foreign work on origin's branch fires after the switch and un-archive commit, unnamed and untested
    detail: |
      3rd in family. Rule (lessons.md): decide every refusal before the first change, or name the change in it. Enumerate the returns after reopendone.go:126 and fix them all; the lease can be decided before the switch against mainTip and the local tip of the branch. Add a refusal test for this path.
```

---

## Re-review — 2026-10-13T15:15:35-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 305 — Reopening a done issue does not un-archive its details |
| repo | ariadne |
| issue file | workshop/issues/000305-reopen-done-unarchive.md |
| boundary | whole-issue close |
| milestone | — |
| window | c40e1b7f5ff15d1d81f5311815e8123101104046..099857eddb368f90bb0df47ec4eb979ffbff1de6 |
| command | sdlc close --issue 305 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-13T15:15:35-07:00 |
| verdict | SHIP |

## Review

I've confirmed the round-4 fix: with it reverted, the new refusal test fails. Now I'm putting the file back.

```verdict
verdict: SHIP
confidence: high
```

**Verdict: ship it.** The one open finding, BR-8, is fixed and has a test that fails without the fix. Commit 099857ed moves `leaseOnRemoteBranch` up to `cmd/sdlc/reopendone.go:80-88`. It now runs after the branch-state checks and before the dirty-tree check, the fast-forward, the switch and the un-archive commit. The lease is now checked against what this reopen will build on: main's tip when the branch is reset, otherwise the local branch's tip. It no longer checks against the HEAD left by an earlier step. A new refusal case, "foreign work on origin's branch", has a peer clone push a commit to the issue branch on origin. The test now records the remote tip before the run and requires it to be unchanged afterwards, instead of requiring it to be empty. I reverted only the `reopendone.go` part of 099857ed in the disposable checkout and ran that case: it failed with "a refused reopen changed something". With the fix restored, all six `TestReopenDone*` tests pass. Nothing else in the window blocks shipping.

1. **Strengths**
   - Every refusal in `restoreDone` (`reopendone.go:38-98`) is now decided before anything changes. That covers the foreign owner, the wrong branch, a stale branch from before the archive, the landed branch, a lease conflict and a dirty tree. Only the guard check comes after the resting branch is fast-forwarded, and its refusal says so through `after` (`reopendone.go:114`).
   - **ARCH-DRY:** `unarchiveIssue` (`abandon.go:452-490`) pulls the un-archive step out of `restoreAbandoned` so the abandon reopen and the done reopen share it. `mirrorSubject` gives the mirror commit message one source, used both when writing it and when recognising it.
   - The postcondition `checkUnarchived` (`reopendone.go:235`) checks the actual result (HEAD descends from main's archive commit and holds the details live) before anything is pushed.
   - `TestReopenDoneRefusals` is a table of refusals that all check the same thing: branch, HEAD, origin's branch tip and the card status are all unchanged.

2. **Critical:** none.
3. **Important:** none.
4. **Minor**
   - **ARCH-ORDER (informational, not raised):** the lease check runs `fetch` or `update-ref -d` on the remote-tracking ref before the dirty-tree refusal. That only refreshes a cache of origin's state, so it's harmless. The refusal test doesn't check that ref, which is reasonable.
   - When `checkUnarchived` fails (`reopendone.go:153`), the error says "nothing was pushed" but doesn't say the branch was switched and an un-archive commit was made. This is a postcondition failure after the changes, not a refusal that could have been decided earlier, so I'm not raising it.

5. **Test coverage**
   - All six `TestReopenDone*` tests pass, and none of the reopen tests failed.
   - One test failed: `TestLegacyRepoLockSetStatusMutationWaits`. Its error says the issue tracker named in `workshop/issue-tracker.json` is unreachable. This review checkout has no remotes, `legacymode_test.go` isn't changed in this window, and the error is about tracker reachability, so I believe it's an environment problem. I didn't run that test at base to confirm.
   - The catalog's list of tests backing the done-reopen claim (`catalog.go:227`) matches the tests that exist.

6. **Architecture**
   - **ARCH-DRY:** pass (shared `unarchiveIssue`, `mirrorSubject`).
   - **ARCH-PURE:** pass (`dirtyBesides` is pure; the git calls stay in the thin layer around it).
   - **ARCH-PURPOSE:** pass. The atlas, help text and recovery catalog are all updated. A README change isn't needed because no new command or flag was added.
   - **ARCH-MOCK:** pass (tests use real git against a local origin and a peer clone).
   - **ARCH-CONSTRAINTS:** N/A (an interactive command with a bounded number of git calls).
   - **ARCH-SECURE:** pass. A malformed details file is refused with a message naming it, and a foreign owner is refused even with `--force`.
   - **ARCH-ORDER:** pass. All refusals now come before the first change, and a rerun is recognised at each step (`unfinishedDoneReopen`).
   - **ARCH-FUNERAL:** pass. The only thing created is the issue branch, which the normal merge lifecycle owns. No archive ref is made for a done reopen.

7. **Plan revisions:** none.

```findings
dispose:
  - id: BR-8
    disposition: addressed
    note: |
      Lease check moved to reopendone.go:80-88, before the dirty check, fast-forward, switch and un-archive commit; the new TestReopenDoneRefusals case fails when the fix is reverted.
```
