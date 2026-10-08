# Boundary Review — ariadne#284 (milestone M3)

| field | value |
|-------|-------|
| issue | 284 — Claims: multi-claim, publish, handoff, takeover |
| repo | ariadne |
| issue file | workshop/issues/000284-claims-handoff-takeover.md |
| boundary | milestone M3 |
| milestone | M3 |
| window | af225f4abb261bac51ca3d2a12c18127698f52a4..dfd89b332db71696da0b2122d916bbd020da380c |
| command | sdlc milestone-close --issue 284 --milestone M3 |
| reviewer | claude |
| timestamp | 2026-10-07T16:09:08-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

M3 meets its Done-when as written. Handing off started work refuses a dirty tree (tracked or untracked) and the wrong branch, pushes the branch, records `release {by, branch, head}` and leaves the status alone. A claim from a second clone with another machine's identity fetches the branch and checks it out at exactly that tip. A moved tip refuses and points at `git log`. A pre-#277 `codecomplete` card is taken over by a plain claim. `move` covers open claims in the model, which repairs a moved open claim. Docs, help, the recovery catalog and the project file were all swept: no stale `--adopt` hint is left outside the retired-alias notes. The targeted suite passes (`go test ./cmd/sdlc -run 'Handoff|Takeover|ClaimSet|Adopt|VerbContract|Relocation|Unclaim|Ownership'` → ok, 62.6s).

Two ordering problems in the takeover path should be fixed before the boundary:
- **Lost card write isn't finished by a rerun.** The atlas and the catalog both say a rerun finishes it, but the rerun reports "nothing to do" and never checks out the branch.
- **The card write doesn't recheck the tip.** The pre-write checks look at the release's tip, but the card write itself doesn't confirm the tip is still the same. That can check out an old tip, and the next handoff's lease then allows force-pushing over newer commits on origin.

**1. Strengths**
- `prepareTakeover` (`cmd/sdlc/handoff.go:150-198`) runs every check that can refuse before the card is written:
  - the branch must equal the issue's own name and pass `check-ref-format`;
  - fetch and checkout pass the branch as a separate argument, never through a shell;
  - the checkout must be clean and on a resting branch;
  - a diverged local copy refuses.

  `TestTakeoverRefusals` confirms the card is byte-unchanged after each refusal.
- Handoff order (push, then card, then switch) with `errAlreadyReleased` makes the lost card write converge. `TestHandoffRerunAfterALostResponse` exercises the real lost response, not a stub.
- `adoptDecision` is folded into `claimDecision` (removed from `claimdecision.go`). The verb-contract table and `TestClaimNeverMovesStatus` now assert over `CanHoldOwner`, so the "claim takes over unowned started work" row comes from the model.
- `appendUnclaimNote` gives the open-claim and handoff note paths one writer.
- The fix to `move`'s statuses was made in the model (`issue.cue`) with a `pkg/vocab` assertion, not as a hand-coded status list in claim.

**2. Critical findings**
None.

**3. Important findings**

- **A takeover whose card write loses its response cannot be finished by rerunning** (`claim.go:214-233`, `handoff.go:231-248`). ARCH-ORDER.
  - **What happens:** `uncertainCardWrite` returns before `finishTakeover` runs. On the rerun:
    - `prepareTakeover` returns nil, because the card now has a claimant;
    - the decision gives `errAlreadyMine`;
    - `finishOwnedTakeover` needs a local `refs/heads/<branch>`, which a fresh clone or slot never created, so it returns without a word;
    - claim prints "already claimed by this workspace; nothing to do".
  - **Why it sticks:** the release head was spent by the card write, so nothing durable records the tip to finish at.
  - **Docs that are now wrong:** `atlas/workflow/issue-tracker.md` ("Reruns finish a lost card write") and the catalog's `LostResponse`.
  - **Related variant:** if `branch -f` fails while a stale ancestor branch exists locally, the rerun switches to that stale tip and prints "resumed".
  - **Family:** this is the 3rd finding in `rerun-not-idempotent`. Earlier rounds fixed instances. The rule that covers all of them: *a "rerun to finish" step must derive its target from state that outlives the effect it follows, and every effect boundary that promises a rerun needs a lost-response test.* Here, the finish needs the tip, but the card write erases it.
  - **Fix sketch:** have `finishOwnedTakeover` fetch the issue branch and create or fast-forward the local branch to the fetched tip (refusing divergence, as `prepareTakeover` does) before switching.
  - **Tests to add:** `loseResponses` around the takeover claim, then a rerun that ends on the branch at `rel.Head`; and a failed `branch -f` with a stale local ancestor.
  - **Same rule, one more site:** the handoff note's dedupe key includes today's date (`appendUnclaimNote`), so a rerun on a later day commits a second note.

- **The takeover's card write doesn't recheck the release head that `prepareTakeover` saw** (`claim.go:184-186` with `claimdecision.go:98-104`).
  - **What happens:** `ChangeCards` re-reads and re-decides after a race, but single-issue `claimSetDecision` accepts any unowned card. Suppose another actor takes over, commits and hands off again at H2 between `prepareTakeover` (which saw H1) and the CAS. The decision lands, and `finishTakeover` runs `branch -f <b> H1`.
  - **Consequence:** the next handoff from this checkout calls `pushIssueBranch`, whose lease is the remote tip it just read with `ls-remote`. That is effectively a plain force push, and it overwrites H2 on origin.
  - **Rule:** an observation authorises an effect only while the CAS proves it is still current.
  - **Fix sketch:**
    - pass the expected `release.head` into the decision and refuse (or re-prepare) when the fresh card's release differs;
    - lease the handoff push on the remote-tracking ref this checkout last fetched (`--force-with-lease=<ref>` with no value), not on a fresh `ls-remote`.
  - **Test:** a `beforePush` interleaving that re-releases the card at a new head.

**4. Minor findings**
- `finishHandoff` and `finishOwnedTakeover` return early without a word on several errors (`CardRelease` error, `rev-parse` error, a dirty tree). The operator gets "nothing to release" or "nothing to do" with no hint about why the checkout didn't move.
- `ownedBy`'s default case ("with an unreadable owner") is only reachable when `MatchClaimant` returns Unknown while `has` is true. If that can't happen, it's dead code; if it can, the message should name the fix.
- `finishHandoff`, `finishOwnedTakeover` and `finishTakeover` each repeat the same "clean, then switch, then warn" sequence. It could be one small helper.

**5. Test coverage notes**
- Covered:
  - handoff happy path across two clones, with the note travelling on the branch;
  - each handoff refusal;
  - handoff lost card write;
  - each takeover refusal;
  - lost takeover switch (simulated with a manual `switch main`);
  - pre-#277 `codecomplete` takeover;
  - set refusal of a handed-off member;
  - the moved-open repair (`TestRelocationAfterMoveAndRepair`, now on an open claim).
- Missing:
  - takeover lost card write (finding 1);
  - takeover race against a re-release (finding 2);
  - a failed `branch -f` with a stale local branch;
  - handoff `--dry-run` and a rerun on a later day.

**6. Architectural notes**
- **ARCH-DRY:** pass. The note writer and the clean-tree check are shared, and `adoptDecision` was removed. The three finish helpers are a Minor.
- **ARCH-PURE:** pass. `claimDecision`, `claimSetDecision` and `unclaimDecision` are pure. `prepareTakeover` and `runHandoff` are thin IO shells.
- **ARCH-PURPOSE:** pass. Every M3 Done-when item is delivered, and the docs and catalog were swept for `--adopt`.
- **ARCH-MOCK:** pass. Tests use real-git fixtures, with a second clone standing in for a second machine.
- **ARCH-CONSTRAINTS:** pass. The added cost is one `ls-remote` or `fetch` per handoff or takeover.
- **ARCH-SECURE:** pass.
  - `release.branch` is checked against the issue's own branch and `check-ref-format`, and passed as a separate argument.
  - The fetched tip is compared with the recorded head before any effect.
- **ARCH-ORDER:** **flag**, findings 1 and 2.
  - A lost card write collapses into "already mine" with nothing left to finish against.
  - An observed release head isn't guarded at the CAS.
- **ARCH-FUNERAL:** pass. Any claim spends the release. The pushed issue branch's removal is assigned to the issue's landing or #286's `abandon`, and the unclaim help says so.
- For M4: `sdlc state` will show released-but-untaken issues. Make sure its claim-time parser counts the `claim-` token for takeovers, since takeover writes through `operationToken("claim")`, not `"takeover"` as the plan's D10 / revision 4 lists.

**7. Plan revision recommendations**
- Add a Revisions entry: "Takeover writes with `operationToken("claim")`, not `"takeover"`. M4's claim-time parser keys on `claim-`." The plan says the opposite today.
- Add a Revisions entry for the finding-1 fix: the owner's repeat claim on started work refetches and fast-forwards the issue branch, so a lost takeover card write finishes. The catalog's `LostResponse` and the atlas then match the behaviour.

```findings
findings:
  - id: new
    severity: Important
    family: rerun-not-idempotent
    title: |
      Takeover whose card write loses its response cannot be finished by rerunning claim
    detail: |
      uncertainCardWrite returns before finishTakeover runs. On rerun, prepareTakeover returns nil (the card is owned) and the decision gives errAlreadyMine. finishOwnedTakeover needs a local refs/heads/<branch> that a fresh clone never created, so it returns without a word ("nothing to do"). The release head was already spent, so nothing records the tip to finish at. The atlas and catalog LostResponse claim a rerun finishes this. Variant: if branch -f fails with a stale ancestor branch present, the rerun switches to the stale tip and prints resumed. 3rd in the family. Rule: a rerun-to-finish step must derive its target from state that outlives the effect it follows, and every effect boundary that promises a rerun needs a lost-response test. Fix: finishOwnedTakeover fetches the issue branch and creates or fast-forwards it (refusing divergence) before switching; add loseResponses tests. The same rule covers the handoff note's date-keyed dedupe.
  - id: new
    severity: Important
    family: stale-observation-not-rechecked-in-cas
    title: |
      Takeover CAS does not recheck the release head prepareTakeover observed; the handoff lease cannot catch it
    detail: |
      The single-issue claimSetDecision accepts any unowned card after a re-read. A re-release at a new head H2 between prepareTakeover (saw H1) and the CAS lands the claim, and finishTakeover sets the branch at stale H1. The next handoff's force-with-lease uses a freshly read ls-remote tip, which is effectively a force push that overwrites H2 on origin. Fix: pass the expected release head into the decision and refuse on mismatch; lease the push on the last-fetched remote-tracking ref; add a beforePush interleaving test.
  - id: new
    severity: Minor
    family: silent-error-in-io-glue
    title: |
      finishHandoff and finishOwnedTakeover return early on errors (CardRelease, rev-parse, a dirty tree) without saying why
    detail: |
      The operator gets "nothing to release" or "nothing to do" with no hint about why the checkout did not move. A one-line cwarn per early return fixes it.
  - id: new
    severity: Minor
    family: duplicated-restore-logic
    title: |
      Three finish helpers repeat the same clean-check, switch and warn sequence
```

---

## Re-review — 2026-10-07T21:11:07-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 284 — Claims: multi-claim, publish, handoff, takeover |
| repo | ariadne |
| issue file | workshop/issues/000284-claims-handoff-takeover.md |
| boundary | milestone M3 |
| milestone | M3 |
| window | af225f4abb261bac51ca3d2a12c18127698f52a4..5fedc44c0f70fa08eb936fc1202052c1e1903b25 |
| command | sdlc milestone-close --issue 284 --milestone M3 |
| reviewer | claude |
| timestamp | 2026-10-07T21:11:07-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

Both Important round-1 findings are fixed, and each fix has a test that would fail without it. **BR-23:** a rerun now resumes from the remote branch, so it no longer depends on the spent release. **BR-24:** the takeover's card write rechecks the release head, and the handoff push leases on the last-fetched tracking ref. The targeted tests pass at HEAD: `go test ./cmd/sdlc/ -run 'Handoff|Takeover|UnclaimNote|ClaimSet|VerbContract|Recovery|Catalog'` → ok in 63s. Nothing remaining is Critical. One Important gap stays open. The new guard in `finishOwnedTakeover` against a diverged local copy has no test, even though removing it would let `branch -f` overwrite local commits. The rest are Minor: wrong or missing warning messages, one more repeat of the tracking-ref string, and a note dedupe that is now too broad.

**Strengths**
- `claim.go:188-196`: the takeover's decide closure compares the release on fresh card bytes against the branch and head that `prepareTakeover` fetched. This is a correct compare-and-swap recheck. `TestTakeoverRefusesAReReleaseBeforeTheWrite` injects the re-release between attempts through the real `cardsPublish` seam, so the test controls the ordering rather than relying on one lucky run.
- `handoff.go:169-185`: `pushIssueBranch` leases on `refs/remotes/<remote>/<branch>` and updates it after a successful push. An empty lease means "the branch must not exist yet". `TestHandoffLeaseRefusesAnUnseenRemoteTip` covers this end to end with a rogue push.
- `handoff.go:262-284`: the rerun now gets its tip from the remote branch, which survives the lost write, instead of from the spent release. The lesson added to `workshop/lessons.md:353` states that rule for the general case.
- `switchClean` (`handoff.go:142`) folds three copies of the switch sequence into one, and it now says why it did not switch.
- The recovery catalog's proofs table and the help, contract and atlas text changed in the same commit as the behaviour.

**Critical:** none.

**Important**
- `handoff.go:277-281`: the refusal branch in `finishOwnedTakeover` for a diverged local copy has no fixture. `grep` for its message finds nothing in the tests. Only `prepareTakeover`'s guard is covered, at `handoff_test.go:157`. If someone removes this guard, `finishTakeover` runs `branch -f` and silently overwrites local commits.
  - Fix: add a lost-response rerun test in which the peer has a local `000009-s09` that diverged from the remote. Assert that it warns, the branch is unchanged and nothing is switched.

**Minor**
- `handoff.go:269-281`: three problems in `finishOwnedTakeover`.
  - Any fetch error returns without a word, so a network failure looks the same as "no branch" (BR-25's residue).
  - An original owner whose local branch is simply *ahead* of the remote (their own unpushed work) gets "reconcile them first", although nothing needs reconciling. It should resume on the local branch when the remote tip is an ancestor of it.
  - A `merge-base` error is reported as divergence.
- `handoff.go:181`: every push failure (network, auth) is reworded as "not the copy this checkout last fetched".
- `handoff.go:170,222,268`: `"refs/remotes/"+remote+"/"+branch` is built three times here, and again in `transferguard.go:112`, `landing.go:117` and `gitx/trunkfile.go:194`.
- `handoff.go:39`: the dedupe now matches the same note text at any date. A later, genuinely separate handoff with the same note (for example "WIP") is silently not filed.

**Test coverage notes**
- BR-23's resume path, BR-24's compare-and-swap recheck, the push lease and the cross-date dedupe are all covered.
- Missing: the diverged-local refusal in `finishOwnedTakeover`, and an original owner's repeat claim from rest with unpushed local commits.

**Architecture**
- ARCH-DRY: flag, Minor (the tracking-ref string; `prepareTakeover` and `finishOwnedTakeover` also repeat the fetch, rev-parse and divergence check, which could be one helper).
- ARCH-PURE: pass. The git glue is thin, and the release check is a pure comparison inside decide.
- ARCH-PURPOSE: pass. The handoff and takeover Done-when items are delivered.
- ARCH-MOCK: pass. Git goes through `env.git`, tests run against real local bare origins, and card writes use the `cardsPublish` seam.
- ARCH-CONSTRAINTS: pass. One note: every repeat claim by an owner from rest on started work now fetches from the network.
- ARCH-SECURE: pass. The release branch must equal the issue's own branch name and pass `check-ref-format`.
- ARCH-ORDER: pass. The stale-observation compare-and-swap is now rechecked, and the interleaving is injectable through `before`.
- ARCH-FUNERAL: pass. The pushed branch is removed by landing (#286) or `abandon`.

**Plan revisions:** none. The code matches M3's Done-when list.

```findings
dispose:
  - id: BR-23
    disposition: addressed
    note: |
      finishOwnedTakeover fetches the remote branch and fast-forwards; TestTakeoverLostResponseRerunResumes fails without it (old code needed a local ref); note dedupe spans dates.
  - id: BR-24
    disposition: addressed
    note: |
      decide rechecks release branch/head (claim.go:188); lease uses last-fetched tracking ref; both tests fail with the fix removed.
  - id: BR-25
    disposition: not-addressed
    note: |
      Mostly fixed, but finishOwnedTakeover's fetch failure (handoff.go:269-271) still returns silently, conflating a network error with "no branch".
  - id: BR-26
    disposition: addressed
    note: |
      switchClean (handoff.go:142) replaces the three switch sequences.
findings:
  - id: new
    severity: Important
    family: test-gap-lost-response-set
    title: |
      finishOwnedTakeover's diverged-local-copy refusal has no test; removing it lets branch -f overwrite local commits
    detail: |
      2nd in family. Rule: every refusal branch on a rerun-to-finish path gets a fixture, not just the happy resume. Add a lost-response rerun test where the peer holds a diverged local issue branch; assert warn, branch unchanged, no switch.
  - id: new
    severity: Minor
    family: refusal-reason-mismatch
    title: |
      Handoff and takeover warnings name the wrong cause (ahead treated as diverged, every push error treated as a lease miss)
    detail: |
      2nd in family. Rule: a refusal names the condition actually observed and is classified before wording. Sites: handoff.go:278 (local ahead of remote, i.e. the owner's unpushed work, and merge-base errors, both reported as needing reconcile); handoff.go:181 (every push error reworded as a lease mismatch); handoff.go:270 (fetch error read as "no branch").
  - id: new
    severity: Minor
    family: duplicated-path-derivation
    title: |
      The remote-tracking ref string is built three times in handoff.go and in at least three other files
    detail: |
      4th in family. Rule: each derived ref or path name has one constructor. Expose a remoteTrackingRef(remote, branch) (gitx/trunkfile.go:194 already has one) and use it in handoff.go:170,222,268, transferguard.go:112 and landing.go:117.
  - id: new
    severity: Minor
    family: rerun-not-idempotent
    title: |
      Note dedupe matches the same text at any date, so a later separate handoff's identical note is silently dropped
    detail: |
      4th in family. Rule: a convergence key must identify this attempt, not just its content. Here, look only in the branch's unpushed tail (commits after the tracking ref) for the note commit, instead of searching the whole Log for the text.
```

---

## Re-review — 2026-10-07T21:21:23-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 284 — Claims: multi-claim, publish, handoff, takeover |
| repo | ariadne |
| issue file | workshop/issues/000284-claims-handoff-takeover.md |
| boundary | milestone M3 |
| milestone | M3 |
| window | af225f4abb261bac51ca3d2a12c18127698f52a4..aa8ddc024bb8825c7dfc32f2c1832fa546480ef7 |
| command | sdlc milestone-close --issue 284 --milestone M3 |
| reviewer | claude |
| timestamp | 2026-10-07T21:21:23-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

I recommend crossing the M3 boundary. The one open Important finding, BR-27, is fixed and has a test that fails without the fix. Three small issues remain, none of them blocking. In `finishOwnedTakeover` (`handoff.go:299-349`), a takeover rerun now sorts a local copy of the branch into one of four cases: absent, behind, ahead or diverged. Each case has its own message. `TestTakeoverRerunWithALocalCopy` covers both the diverged case (it warns and changes nothing) and the ahead case (it resumes on the local work). If the diverged refusal were removed, `finishTakeover`'s `branch -f` would overwrite the local commits, and the test would catch it by comparing the ref. The targeted test run passes (`go test -run 'Handoff|Takeover|UnclaimNote|ClaimSet|Recovery|VerbContract'` took 68s). The remaining issues:
- **BR-29:** the one-constructor sweep left three spellings behind, two of them in files this round edited.
- **BR-30:** the note dedupe still drops a second handoff's identical note when both happen on the same day.
- **New:** two places treat an error from a "does this local branch exist?" check as "no branch", and the next step is `branch -f`.

1. **Strengths**
   - `handoff.go:131-149` `handoffNoteCommitted`: it decides "is this a rerun?" from the commit that attempt made in the unpushed tail, not from the note text. That is the right way to identify an attempt, and `TestHandoffRecognisesItsNoteCommit` exercises it with a note dated before midnight.
   - `handoff.go:305-317`: listing the remote with `ls-remote` before fetching separates "the branch doesn't exist" from "the fetch failed". Before, both were silently treated as "no branch" (BR-28).
   - `handoff.go:214-218`: the push error is now checked before it is reworded. Only git's lease-rejection message gets the "last fetched" explanation, and `TestHandoffLeaseRefusesAnUnseenRemoteTip` depends on that wording.
   - The recovery catalog rows (`recovery/catalog.go:26,298`) now name the new tests, so each recovery contract stays traceable to a test.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - **BR-29 residue:**
     - `landing.go:86` spells `"refs/remotes/"+t.Remote+"/main"` even though `mainRef()` (`landing.go:118`) is defined in the same file.
     - `observe.go:117` and `observe.go:181` still hand-build the main tracking ref; line 237 of the same file was converted.
     - `gitx/publicationtarget.go:56` builds it inside the very package that owns `RemoteTrackingRef`.
     - Note: the `for-each-ref` namespace prefixes (`issuemigrate.go:222`, `planningbranch.go:116`) are prefixes, not tracking refs, so they are fine as they are.
   - **BR-30 residue:** in the handoff path, `appendUnclaimNote`'s same-day text match is now redundant with `handoffNoteCommitted`, and it still drops the note of a second, separate handoff made the same day with the same text. Fix: let the handoff skip the text match and rely on the commit key. The plain-unclaim path (`unclaim.go:195`) can keep the text match.
   - **New, silent-error-in-io-glue:** `handoff.go:270` and `handoff.go:323-324` check whether the local branch exists with `env.git("rev-parse","-q","--verify",…)` and treat any error as "absent". The next step is `branch -f`, so a transient rev-parse failure would overwrite a diverged local copy.
   - `pushIssueBranch` (`handoff.go:215`) classifies the error by matching git's English text `"stale info"`. Under a translated `LANG` the friendlier explanation silently disappears, though the error still reaches the operator. Consider forcing `LC_ALL=C` for that push, or noting the dependency.

5. **Test coverage notes**
   - BR-27: covered as described above.
   - BR-30's two branches: covered by `TestUnclaimNoteIsKeyedToTheAttempt` and `TestHandoffRecognisesItsNoteCommit`.
   - The new warnings on error paths (BR-25) have no tests. That is acceptable: those git failures are hard to set up as fixtures, and none of them is a refusal that protects state.

6. **Architectural notes**
   - **ARCH-DRY:** flagged, see BR-29 residue.
   - **ARCH-PURE:** passes for now. The behind/ahead/diverged decision is mixed into IO in both `prepareTakeover` and `finishOwnedTakeover`. A pure `classify(local, tip, isAncestor…)` would let both sites share one decision table; worth doing in M4 if that code is touched.
   - **ARCH-PURPOSE:** passes. Every M3 Done-when item is delivered. The "envelope keeps unknown keys" item landed in M2 (ee7e83fb).
   - **ARCH-MOCK:** passes. Tests run against real fixture repos and a second-machine fixture through the same `env.git` boundary as production.
   - **ARCH-CONSTRAINTS:** passes. A rerun costs one `ls-remote` plus one fetch.
   - **ARCH-SECURE:** passes. The branch named in a release is checked against the issue's own branch and against `check-ref-format` before it is fetched.
   - **ARCH-ORDER:** passes. Push, then card, then switch, with each lost step finished by a rerun, and the rerun paths are tested, including the lost response.
   - **ARCH-FUNERAL:** passes. The pushed branch's end (removed on land or abandon) is stated in the unclaim help.

7. **Plan revision recommendations:** none. The plan matches the code.

```findings
dispose:
  - id: BR-25
    disposition: addressed
    note: |
      finishHandoff warns on an unreadable release and a HEAD read failure; finishOwnedTakeover warns on ls-remote, fetch, rev-parse and merge-base failures; dirty trees warn via switchClean. Remaining silent returns are genuine nothing-to-do cases.
  - id: BR-27
    disposition: addressed
    note: |
      TestTakeoverRerunWithALocalCopy/diverged asserts the warning, the local ref unchanged and no switch; removing the refusal at handoff.go:348 makes branch -f overwrite the ref and the test fails.
  - id: BR-28
    disposition: addressed
    note: |
      Ahead and diverged are now distinct (handoff.go:342-348), merge-base errors have their own message, the push error is classified on stale info (handoff.go:215, pinned by TestHandoffLeaseRefusesAnUnseenRemoteTip), and ls-remote separates a missing branch from a failed fetch.
  - id: BR-29
    disposition: not-addressed
    note: |
      Constructor added and used at the named sites, but siblings remain: landing.go:86 (next to mainRef in the same file), observe.go:117 and :181 (same file as the converted :237), gitx/publicationtarget.go:56.
  - id: BR-30
    disposition: not-addressed
    note: |
      The commit-keyed check fixes the across-midnight rerun, but the handoff path still runs appendUnclaimNote's same-day text match, so a second separate handoff the same day with the same note is dropped; let the handoff rely on handoffNoteCommitted alone.
findings:
  - id: new
    severity: Minor
    family: silent-error-in-io-glue
    title: |
      A failed check for the local branch is read as "absent" and followed by branch -f (handoff.go:270, :323-324)
    detail: |
      This is the 4th finding in family silent-error-in-io-glue. Rule: every existence check in the handoff glue returns three outcomes (present, absent, error) through env.gitTest, and an error stops with a warning; it is never folded into "absent". Apply it to both rev-parse --verify sites; the earlier-round sites already follow it.
```
