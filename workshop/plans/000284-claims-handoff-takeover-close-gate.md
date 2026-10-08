---
gate: boundary-review
issue: 284
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-07T14:35:04-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Minor
          title: refreshAfterClaim swallows rev-parse errors and misreports merge-base failures as "commits main lacks"
          detail: 'cmd/sdlc/claim.go refreshAfterClaim: surface the actual git error in the warning instead of returning silently or naming the wrong cause.'
          family: silent-error-in-io-glue
          round: 1
        - id: BR-2
          severity: Minor
          title: claimRefs, claimArg and cardsMessage each re-implement the CLIRef join; ChangeCards inlines validateReplacement's budget math
          detail: 'ARCH-DRY: one ref-join helper parameterized by prefix/separator, and a cumulative validateReplacements in tracker/reader.go.'
          family: duplicated-id-rendering
          round: 1
        - id: BR-3
          severity: Minor
          title: tracker.ChangeCards does not dedupe ids; relies on callers (claimIssues) to do so
          family: api-trusts-caller-normalization
          round: 1
        - id: BR-4
          severity: Minor
          title: a started unowned card in a set is refused toward --adopt, which itself refuses sets
          detail: Message should say to claim that issue alone with --adopt.
          family: refusal-points-at-unusable-action
          round: 1
        - id: BR-5
          severity: Minor
          title: 'project file ticks M1 with actual/closed before milestone-close measures it; M lines not nested under the #284 line'
          family: hand-recorded-actuals
          round: 1
        - id: BR-6
          severity: Minor
          title: no test drives a claim set through a lost publication response and a rerun
          family: test-gap-lost-response-set
          round: 1
      boundary: M1
      recipe: milestone-review
      blocked: false
    - "n": 2
      timestamp: "2026-10-07T15:10:20-07:00"
      agent: claude
      findings:
        - id: BR-7
          severity: Important
          title: unclaim --note is appended locally before publish and the card write; a rerun after either fails duplicates the Log line on main
          detail: cmd/sdlc/unclaim.go:184-199 writes the note, then republishOwned, then cardsPublish. A refused publish ("the claim is kept; rerun") or a failed card write tells the operator to rerun the same command, and insertLogLine does not dedupe, so the note lands twice. Make the note convergent (skip it when the line is already present, or judge it against main's copy) and test refused-publish then rerun.
          family: rerun-not-idempotent
          round: 2
        - id: BR-8
          severity: Important
          title: republish's main-moved refusal advises `sdlc claim` to fast-forward the rest, which git refuses in exactly that state
          detail: 'This is the 2nd finding in this family. republish.go:110 fires when main changed this file and the local copy is edited; claim''s repeat path runs fastForwardRest, whose git merge --ff-only refuses because a dirty tracked file would be overwritten, so it only warns. Rule: every remedy a refusal names must be runnable from the refused state, proven by a test that runs it from that state. Fix the remedy (save the edits, fast-forward, re-apply, or have publish do it) and sweep the diff''s other refusals the same way.'
          family: refusal-points-at-unusable-action
          round: 2
        - id: BR-9
          severity: Important
          title: README and atlas/workflow/sdlc-binary.md still describe issue sync / move-detail as the tracker path; unclaim and issue publish --issue are absent
          detail: 'README.md:19-23 says `sdlc issue sync --issue N` checkpoints on the issue branch, but it now refuses in tracker repositories. sdlc-binary.md:43-72 has no unclaim row, lists issue publish as "Retiring at the #252 cutover", and has no issue publish --issue row.'
          family: docs-sweep-missing
          round: 2
        - id: BR-10
          severity: Minor
          title: unclaim on a terminal status says "releasing started work is a handoff (M3)"
          detail: unclaim.go:155 tests !IsOpen, which also catches done/terminal statuses; give terminal statuses their own reason.
          family: refusal-reason-mismatch
          round: 2
        - id: BR-11
          severity: Minor
          title: finishPublished off-rest overwrites the details without the unchanged-since-read check, and commits on whatever branch is checked out
          detail: 'republish.go:230 can drop an edit made during the push, can add an #A commit onto #B''s issue branch (against #272), and skips the commit when the file is untracked in HEAD.'
          family: rerun-not-idempotent
          round: 2
        - id: BR-12
          severity: Minor
          title: The local details path is spelled out three times across unclaim.go and republish.go
          detail: filepath.Join(env.root, filepath.FromSlash(path.Join(dirs.Rel[0], path.Base(card.Path)))) should be one helper.
          family: duplicated-path-derivation
          round: 2
      boundary: M2
      recipe: milestone-review
      blocked: true
    - "n": 3
      timestamp: "2026-10-07T15:26:36-07:00"
      agent: claude
      dispose:
        - id: BR-7
          disposition: addressed
          note: strings.Contains guard at unclaim.go:201 and TestUnclaimNoteIsConvergent fails without it; the date in the line makes a cross-midnight rerun duplicate (residual, minor)
          round: 3
        - id: BR-8
          disposition: not-addressed
          note: 'resting branch fixed and tested; the sweep was not done: the off-rest refusal at republish.go:123 names a merge that is refused with an uncommitted edit, and no test runs it'
          round: 3
        - id: BR-9
          disposition: addressed
          note: README.md:19-24 and sdlc-binary.md rows for unclaim, issue publish --issue and legacy-only issue sync
          round: 3
        - id: BR-10
          disposition: addressed
          note: CanHoldOwner case at unclaim.go:156 with its own reason; covered at unclaim_test.go:44
          round: 3
        - id: BR-11
          disposition: addressed
          note: read-check applies on both branch kinds, commits only on the issue's own branch, untracked handled through the porcelain check; TestPublishFromAnotherIssuesBranchCommitsNothingThere
          round: 3
        - id: BR-12
          disposition: addressed
          note: localDetail helper used by unclaim.go and republish.go
          round: 3
      findings:
        - id: BR-13
          severity: Important
          title: A rerun after a conflicted merge-in publishes unresolved conflict markers to main
          detail: After bringMainIn conflicts, HEAD equals main, so base equals main and the rerun gets publishWrite with the markers in the body. Refuse a publish body that still has conflict-marker lines; test conflict then rerun without resolving leaves main unchanged.
          family: intermediate-state-consumed-unchecked
          round: 3
        - id: BR-14
          severity: Important
          title: bringMainIn's set-aside loop can revert earlier edits and then report the local file unchanged
          detail: Checkout HEAD on each moved file returns at the first failure (for example a file not in HEAD, which the empty-base merge path admits) after earlier files were already reverted; the restore after a failed fast-forward discards its write error. Make set-aside all-or-nothing, writing every touched file back from item.read and surfacing any failure.
          family: partial-effect-not-rolled-back
          round: 3
        - id: BR-15
          severity: Important
          title: issue help text and the atlas still say main moved means refused, and any other branch commits the bytes
          detail: 'This is the 2nd finding in this family. helptext/issue.md:103,108 and atlas/workflow/issue-tracker.md:245,264-268 predate c2871a48. Rule: a fix round that changes a verb''s behavior updates that verb''s help text and atlas section in the same commit.'
          family: docs-sweep-missing
          round: 3
        - id: BR-16
          severity: Minor
          title: republish.go:262 derives the issue branch from the details path inline again
          detail: 'This is the 2nd finding in this family. Rule: one internal/issue helper for each name derived from a card path. The same expression is at planningbranch.go:29, transferguard.go:69, observe/assemble.go:179 and claimant.go:190.'
          family: duplicated-path-derivation
          round: 3
        - id: BR-17
          severity: Minor
          title: mergeDetails execs git merge-file directly instead of through the env git seam
          family: external-call-outside-seam
          round: 3
      boundary: M2
      recipe: milestone-review
      blocked: true
    - "n": 4
      timestamp: "2026-10-07T15:38:08-07:00"
      agent: claude
      dispose:
        - id: BR-8
          disposition: not-addressed
          note: Issue-branch remedy fixed and proven red-without-fix; but dry-run on rest gets the "not N's branch" refusal (republish.go:125), republish.go:161 still says "bring main in first", the other-branch remedy has no test that runs it (the catalog cites a test that never moves main), and origin/main is hardcoded.
          round: 4
        - id: BR-13
          disposition: addressed
          note: republish.go:116 guard; TestPublishRefusesConflictMarkers fails on c2871a48 (verified in a scratch worktree).
          round: 4
        - id: BR-14
          disposition: not-addressed
          note: The code looks right, but TestPublishBringInPutsEditsBackOnFailure passes on pre-fix c2871a48 (fast-forward failure path, old code wrote back too); set-aside failure part-way and not-in-HEAD paths untested; a write failure after the fast-forward (republish.go:362) returns a raw error without pointing at the aside copies.
          round: 4
        - id: BR-15
          disposition: addressed
          note: helptext/issue.md:100-117 and atlas/workflow/issue-tracker.md:245,262-267 now match republish.go behaviour. Catalog residue raised separately.
          round: 4
        - id: BR-16
          disposition: addressed
          note: All five named sites use issue.BranchName; siblings outside the window raised as a Minor.
          round: 4
        - id: BR-17
          disposition: addressed
          note: mergeDetails uses env.gitRaw, which wraps the error with %w so the ExitError check still works; conflict tests exercise it.
          round: 4
      findings:
        - id: BR-18
          severity: Important
          title: Recovery catalog Effects/Repeat for issue publish still say "another branch commits the same bytes" / "branch commit"
          detail: 3rd in family. catalog.go:259,262 predate this round's change (another issue's branch now takes its copy back). The rule is already in lessons (help, atlas and catalog change in the same commit); this round's commit edited Preconditions next to these lines and missed them. Fix the instance and grep the replaced behaviour phrase across helptext/, atlas/ and catalog.go before each fix-round commit.
          family: docs-sweep-missing
          round: 4
        - id: BR-19
          severity: Minor
          title: gitTest errors discarded at republish.go:242,350 turn a failed probe into "not in HEAD" and then remove the file
          detail: '2nd in family. gitTest''s contract says an error is never evidence of absence. republish.go:128 likewise treats a gitRaw failure as an empty merge base. Rule: no gitTest/gitRaw result is used without its error.'
          family: silent-error-in-io-glue
          round: 4
        - id: BR-20
          severity: Minor
          title: publish-aside copies under the git dir are never removed
          detail: ARCH-FUNERAL. Written on every bring-in, keyed by basename, never cleared; stale copies of other issues sit beside the ones the "copies are in" error points at. Remove each copy after a successful write-back or put-back.
          family: artifact-without-removal
          round: 4
        - id: BR-21
          severity: Minor
          title: branchcreate.go:76,89 and closetracker.go:268 still derive the issue branch/stem inline
          detail: '3rd in family. Rule: issue-branch names come only from issue.BranchName; the branch creator is the most important consumer. These sites are outside the window.'
          family: duplicated-path-derivation
          round: 4
        - id: BR-22
          severity: Minor
          title: '"restore from HEAD, else remove" is implemented twice (finishPublished.restore and the bringMainIn set-aside loop)'
          detail: 'ARCH-DRY: extract one helper returning an error, which would also fix the discarded gitTest error at both sites.'
          family: duplicated-restore-logic
          round: 4
      boundary: M2
      recipe: milestone-review
      blocked: true
    - "n": 5
      timestamp: "2026-10-07T15:49:53-07:00"
      agent: claude
      dispose:
        - id: BR-8
          disposition: addressed
          note: On a resting branch publish brings main in itself (bringMainIn); TestPublishBringsAMovedMainIn and both *Remedy tests run the named steps from the refused state.
          round: 5
        - id: BR-14
          disposition: addressed
          note: putBack writes every moved file back from item.read and reports every failure; copies kept under the git dir; TestPublishBringInSetAsideIsAllOrNothing injects a 2nd-file failure.
          round: 5
        - id: BR-18
          disposition: addressed
          note: catalog.go:259,262 now say another issue's branch takes its copy back, matching helptext/issue.md:111-116 and atlas issue-tracker.md:245,267.
          round: 5
        - id: BR-19
          disposition: not-addressed
          note: Named sites fixed via env.has in detailsAt/restoreToHead, but republish.go:272 still turns a failed git status into "nothing to commit" silently, and unclaim.go:176 drops a Snapshot error.
          round: 5
        - id: BR-20
          disposition: addressed
          note: removeAside runs after a successful write-back and after a safe put-back; kept only on the path whose error names them.
          round: 5
        - id: BR-21
          disposition: addressed
          note: branchcreate.go:76,89 use issue.BranchName; closetracker.go uses issue.Stem for the plan glob, which is a stem, not a branch.
          round: 5
        - id: BR-22
          disposition: addressed
          note: restoreToHead is the single helper returning an error, used by finishPublished and bringMainIn.
          round: 5
      boundary: M2
      recipe: milestone-review
      blocked: false
    - "n": 6
      timestamp: "2026-10-07T16:09:08-07:00"
      agent: claude
      findings:
        - id: BR-23
          severity: Important
          title: Takeover whose card write loses its response cannot be finished by rerunning claim
          detail: 'uncertainCardWrite returns before finishTakeover runs. On rerun, prepareTakeover returns nil (the card is owned) and the decision gives errAlreadyMine. finishOwnedTakeover needs a local refs/heads/<branch> that a fresh clone never created, so it returns without a word ("nothing to do"). The release head was already spent, so nothing records the tip to finish at. The atlas and catalog LostResponse claim a rerun finishes this. Variant: if branch -f fails with a stale ancestor branch present, the rerun switches to the stale tip and prints resumed. 3rd in the family. Rule: a rerun-to-finish step must derive its target from state that outlives the effect it follows, and every effect boundary that promises a rerun needs a lost-response test. Fix: finishOwnedTakeover fetches the issue branch and creates or fast-forwards it (refusing divergence) before switching; add loseResponses tests. The same rule covers the handoff note''s date-keyed dedupe.'
          family: rerun-not-idempotent
          round: 6
        - id: BR-24
          severity: Important
          title: Takeover CAS does not recheck the release head prepareTakeover observed; the handoff lease cannot catch it
          detail: 'The single-issue claimSetDecision accepts any unowned card after a re-read. A re-release at a new head H2 between prepareTakeover (saw H1) and the CAS lands the claim, and finishTakeover sets the branch at stale H1. The next handoff''s force-with-lease uses a freshly read ls-remote tip, which is effectively a force push that overwrites H2 on origin. Fix: pass the expected release head into the decision and refuse on mismatch; lease the push on the last-fetched remote-tracking ref; add a beforePush interleaving test.'
          family: stale-observation-not-rechecked-in-cas
          round: 6
        - id: BR-25
          severity: Minor
          title: finishHandoff and finishOwnedTakeover return early on errors (CardRelease, rev-parse, a dirty tree) without saying why
          detail: The operator gets "nothing to release" or "nothing to do" with no hint about why the checkout did not move. A one-line cwarn per early return fixes it.
          family: silent-error-in-io-glue
          round: 6
        - id: BR-26
          severity: Minor
          title: Three finish helpers repeat the same clean-check, switch and warn sequence
          family: duplicated-restore-logic
          round: 6
      boundary: M3
      recipe: milestone-review
      blocked: true
    - "n": 7
      timestamp: "2026-10-07T21:11:08-07:00"
      agent: claude
      dispose:
        - id: BR-23
          disposition: addressed
          note: finishOwnedTakeover fetches the remote branch and fast-forwards; TestTakeoverLostResponseRerunResumes fails without it (old code needed a local ref); note dedupe spans dates.
          round: 7
        - id: BR-24
          disposition: addressed
          note: decide rechecks release branch/head (claim.go:188); lease uses last-fetched tracking ref; both tests fail with the fix removed.
          round: 7
        - id: BR-25
          disposition: not-addressed
          note: Mostly fixed, but finishOwnedTakeover's fetch failure (handoff.go:269-271) still returns silently, conflating a network error with "no branch".
          round: 7
        - id: BR-26
          disposition: addressed
          note: switchClean (handoff.go:142) replaces the three switch sequences.
          round: 7
      findings:
        - id: BR-27
          severity: Important
          title: finishOwnedTakeover's diverged-local-copy refusal has no test; removing it lets branch -f overwrite local commits
          detail: '2nd in family. Rule: every refusal branch on a rerun-to-finish path gets a fixture, not just the happy resume. Add a lost-response rerun test where the peer holds a diverged local issue branch; assert warn, branch unchanged, no switch.'
          family: test-gap-lost-response-set
          round: 7
        - id: BR-28
          severity: Minor
          title: Handoff and takeover warnings name the wrong cause (ahead treated as diverged, every push error treated as a lease miss)
          detail: '2nd in family. Rule: a refusal names the condition actually observed and is classified before wording. Sites: handoff.go:278 (local ahead of remote, i.e. the owner''s unpushed work, and merge-base errors, both reported as needing reconcile); handoff.go:181 (every push error reworded as a lease mismatch); handoff.go:270 (fetch error read as "no branch").'
          family: refusal-reason-mismatch
          round: 7
        - id: BR-29
          severity: Minor
          title: The remote-tracking ref string is built three times in handoff.go and in at least three other files
          detail: '4th in family. Rule: each derived ref or path name has one constructor. Expose a remoteTrackingRef(remote, branch) (gitx/trunkfile.go:194 already has one) and use it in handoff.go:170,222,268, transferguard.go:112 and landing.go:117.'
          family: duplicated-path-derivation
          round: 7
        - id: BR-30
          severity: Minor
          title: Note dedupe matches the same text at any date, so a later separate handoff's identical note is silently dropped
          detail: '4th in family. Rule: a convergence key must identify this attempt, not just its content. Here, look only in the branch''s unpushed tail (commits after the tracking ref) for the note commit, instead of searching the whole Log for the text.'
          family: rerun-not-idempotent
          round: 7
      boundary: M3
      recipe: milestone-review
      blocked: true
    - "n": 8
      timestamp: "2026-10-07T21:21:23-07:00"
      agent: claude
      dispose:
        - id: BR-25
          disposition: addressed
          note: finishHandoff warns on an unreadable release and a HEAD read failure; finishOwnedTakeover warns on ls-remote, fetch, rev-parse and merge-base failures; dirty trees warn via switchClean. Remaining silent returns are genuine nothing-to-do cases.
          round: 8
        - id: BR-27
          disposition: addressed
          note: TestTakeoverRerunWithALocalCopy/diverged asserts the warning, the local ref unchanged and no switch; removing the refusal at handoff.go:348 makes branch -f overwrite the ref and the test fails.
          round: 8
        - id: BR-28
          disposition: addressed
          note: Ahead and diverged are now distinct (handoff.go:342-348), merge-base errors have their own message, the push error is classified on stale info (handoff.go:215, pinned by TestHandoffLeaseRefusesAnUnseenRemoteTip), and ls-remote separates a missing branch from a failed fetch.
          round: 8
        - id: BR-29
          disposition: not-addressed
          note: 'Constructor added and used at the named sites, but siblings remain: landing.go:86 (next to mainRef in the same file), observe.go:117 and :181 (same file as the converted :237), gitx/publicationtarget.go:56.'
          round: 8
        - id: BR-30
          disposition: not-addressed
          note: The commit-keyed check fixes the across-midnight rerun, but the handoff path still runs appendUnclaimNote's same-day text match, so a second separate handoff the same day with the same note is dropped; let the handoff rely on handoffNoteCommitted alone.
          round: 8
      findings:
        - id: BR-31
          severity: Minor
          title: A failed check for the local branch is read as "absent" and followed by branch -f (handoff.go:270, :323-324)
          detail: 'This is the 4th finding in family silent-error-in-io-glue. Rule: every existence check in the handoff glue returns three outcomes (present, absent, error) through env.gitTest, and an error stops with a warning; it is never folded into "absent". Apply it to both rev-parse --verify sites; the earlier-round sites already follow it.'
          family: silent-error-in-io-glue
          round: 8
      boundary: M3
      recipe: milestone-review
      blocked: false
---

# Gate ledger — ariadne#284 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-07T14:35:04-07:00 (claude) — passed

### Raised

- **BR-1** [Minor] `silent-error-in-io-glue` refreshAfterClaim swallows rev-parse errors and misreports merge-base failures as "commits main lacks"
  cmd/sdlc/claim.go refreshAfterClaim: surface the actual git error in the warning instead of returning silently or naming the wrong cause.
- **BR-2** [Minor] `duplicated-id-rendering` claimRefs, claimArg and cardsMessage each re-implement the CLIRef join; ChangeCards inlines validateReplacement's budget math
  ARCH-DRY: one ref-join helper parameterized by prefix/separator, and a cumulative validateReplacements in tracker/reader.go.
- **BR-3** [Minor] `api-trusts-caller-normalization` tracker.ChangeCards does not dedupe ids; relies on callers (claimIssues) to do so
- **BR-4** [Minor] `refusal-points-at-unusable-action` a started unowned card in a set is refused toward --adopt, which itself refuses sets
  Message should say to claim that issue alone with --adopt.
- **BR-5** [Minor] `hand-recorded-actuals` project file ticks M1 with actual/closed before milestone-close measures it; M lines not nested under the #284 line
- **BR-6** [Minor] `test-gap-lost-response-set` no test drives a claim set through a lost publication response and a rerun

## Round 2 — 2026-10-07T15:10:20-07:00 (claude) — BLOCKED

### Raised

- **BR-7** [Important] `rerun-not-idempotent` unclaim --note is appended locally before publish and the card write; a rerun after either fails duplicates the Log line on main
  cmd/sdlc/unclaim.go:184-199 writes the note, then republishOwned, then cardsPublish. A refused publish ("the claim is kept; rerun") or a failed card write tells the operator to rerun the same command, and insertLogLine does not dedupe, so the note lands twice. Make the note convergent (skip it when the line is already present, or judge it against main's copy) and test refused-publish then rerun.
- **BR-8** [Important] `refusal-points-at-unusable-action` republish's main-moved refusal advises `sdlc claim` to fast-forward the rest, which git refuses in exactly that state
  This is the 2nd finding in this family. republish.go:110 fires when main changed this file and the local copy is edited; claim's repeat path runs fastForwardRest, whose git merge --ff-only refuses because a dirty tracked file would be overwritten, so it only warns. Rule: every remedy a refusal names must be runnable from the refused state, proven by a test that runs it from that state. Fix the remedy (save the edits, fast-forward, re-apply, or have publish do it) and sweep the diff's other refusals the same way.
- **BR-9** [Important] `docs-sweep-missing` README and atlas/workflow/sdlc-binary.md still describe issue sync / move-detail as the tracker path; unclaim and issue publish --issue are absent
  README.md:19-23 says `sdlc issue sync --issue N` checkpoints on the issue branch, but it now refuses in tracker repositories. sdlc-binary.md:43-72 has no unclaim row, lists issue publish as "Retiring at the #252 cutover", and has no issue publish --issue row.
- **BR-10** [Minor] `refusal-reason-mismatch` unclaim on a terminal status says "releasing started work is a handoff (M3)"
  unclaim.go:155 tests !IsOpen, which also catches done/terminal statuses; give terminal statuses their own reason.
- **BR-11** [Minor] `rerun-not-idempotent` finishPublished off-rest overwrites the details without the unchanged-since-read check, and commits on whatever branch is checked out
  republish.go:230 can drop an edit made during the push, can add an #A commit onto #B's issue branch (against #272), and skips the commit when the file is untracked in HEAD.
- **BR-12** [Minor] `duplicated-path-derivation` The local details path is spelled out three times across unclaim.go and republish.go
  filepath.Join(env.root, filepath.FromSlash(path.Join(dirs.Rel[0], path.Base(card.Path)))) should be one helper.

## Round 3 — 2026-10-07T15:26:36-07:00 (claude) — BLOCKED

### Disposed

- BR-7 — addressed — strings.Contains guard at unclaim.go:201 and TestUnclaimNoteIsConvergent fails without it; the date in the line makes a cross-midnight rerun duplicate (residual, minor)
- BR-8 — not-addressed — resting branch fixed and tested; the sweep was not done: the off-rest refusal at republish.go:123 names a merge that is refused with an uncommitted edit, and no test runs it
- BR-9 — addressed — README.md:19-24 and sdlc-binary.md rows for unclaim, issue publish --issue and legacy-only issue sync
- BR-10 — addressed — CanHoldOwner case at unclaim.go:156 with its own reason; covered at unclaim_test.go:44
- BR-11 — addressed — read-check applies on both branch kinds, commits only on the issue's own branch, untracked handled through the porcelain check; TestPublishFromAnotherIssuesBranchCommitsNothingThere
- BR-12 — addressed — localDetail helper used by unclaim.go and republish.go

### Raised

- **BR-13** [Important] `intermediate-state-consumed-unchecked` A rerun after a conflicted merge-in publishes unresolved conflict markers to main
  After bringMainIn conflicts, HEAD equals main, so base equals main and the rerun gets publishWrite with the markers in the body. Refuse a publish body that still has conflict-marker lines; test conflict then rerun without resolving leaves main unchanged.
- **BR-14** [Important] `partial-effect-not-rolled-back` bringMainIn's set-aside loop can revert earlier edits and then report the local file unchanged
  Checkout HEAD on each moved file returns at the first failure (for example a file not in HEAD, which the empty-base merge path admits) after earlier files were already reverted; the restore after a failed fast-forward discards its write error. Make set-aside all-or-nothing, writing every touched file back from item.read and surfacing any failure.
- **BR-15** [Important] `docs-sweep-missing` issue help text and the atlas still say main moved means refused, and any other branch commits the bytes
  This is the 2nd finding in this family. helptext/issue.md:103,108 and atlas/workflow/issue-tracker.md:245,264-268 predate c2871a48. Rule: a fix round that changes a verb's behavior updates that verb's help text and atlas section in the same commit.
- **BR-16** [Minor] `duplicated-path-derivation` republish.go:262 derives the issue branch from the details path inline again
  This is the 2nd finding in this family. Rule: one internal/issue helper for each name derived from a card path. The same expression is at planningbranch.go:29, transferguard.go:69, observe/assemble.go:179 and claimant.go:190.
- **BR-17** [Minor] `external-call-outside-seam` mergeDetails execs git merge-file directly instead of through the env git seam

## Round 4 — 2026-10-07T15:38:08-07:00 (claude) — BLOCKED

### Disposed

- BR-8 — not-addressed — Issue-branch remedy fixed and proven red-without-fix; but dry-run on rest gets the "not N's branch" refusal (republish.go:125), republish.go:161 still says "bring main in first", the other-branch remedy has no test that runs it (the catalog cites a test that never moves main), and origin/main is hardcoded.
- BR-13 — addressed — republish.go:116 guard; TestPublishRefusesConflictMarkers fails on c2871a48 (verified in a scratch worktree).
- BR-14 — not-addressed — The code looks right, but TestPublishBringInPutsEditsBackOnFailure passes on pre-fix c2871a48 (fast-forward failure path, old code wrote back too); set-aside failure part-way and not-in-HEAD paths untested; a write failure after the fast-forward (republish.go:362) returns a raw error without pointing at the aside copies.
- BR-15 — addressed — helptext/issue.md:100-117 and atlas/workflow/issue-tracker.md:245,262-267 now match republish.go behaviour. Catalog residue raised separately.
- BR-16 — addressed — All five named sites use issue.BranchName; siblings outside the window raised as a Minor.
- BR-17 — addressed — mergeDetails uses env.gitRaw, which wraps the error with %w so the ExitError check still works; conflict tests exercise it.

### Raised

- **BR-18** [Important] `docs-sweep-missing` Recovery catalog Effects/Repeat for issue publish still say "another branch commits the same bytes" / "branch commit"
  3rd in family. catalog.go:259,262 predate this round's change (another issue's branch now takes its copy back). The rule is already in lessons (help, atlas and catalog change in the same commit); this round's commit edited Preconditions next to these lines and missed them. Fix the instance and grep the replaced behaviour phrase across helptext/, atlas/ and catalog.go before each fix-round commit.
- **BR-19** [Minor] `silent-error-in-io-glue` gitTest errors discarded at republish.go:242,350 turn a failed probe into "not in HEAD" and then remove the file
  2nd in family. gitTest's contract says an error is never evidence of absence. republish.go:128 likewise treats a gitRaw failure as an empty merge base. Rule: no gitTest/gitRaw result is used without its error.
- **BR-20** [Minor] `artifact-without-removal` publish-aside copies under the git dir are never removed
  ARCH-FUNERAL. Written on every bring-in, keyed by basename, never cleared; stale copies of other issues sit beside the ones the "copies are in" error points at. Remove each copy after a successful write-back or put-back.
- **BR-21** [Minor] `duplicated-path-derivation` branchcreate.go:76,89 and closetracker.go:268 still derive the issue branch/stem inline
  3rd in family. Rule: issue-branch names come only from issue.BranchName; the branch creator is the most important consumer. These sites are outside the window.
- **BR-22** [Minor] `duplicated-restore-logic` "restore from HEAD, else remove" is implemented twice (finishPublished.restore and the bringMainIn set-aside loop)
  ARCH-DRY: extract one helper returning an error, which would also fix the discarded gitTest error at both sites.

## Round 5 — 2026-10-07T15:49:53-07:00 (claude) — passed

### Disposed

- BR-8 — addressed — On a resting branch publish brings main in itself (bringMainIn); TestPublishBringsAMovedMainIn and both *Remedy tests run the named steps from the refused state.
- BR-14 — addressed — putBack writes every moved file back from item.read and reports every failure; copies kept under the git dir; TestPublishBringInSetAsideIsAllOrNothing injects a 2nd-file failure.
- BR-18 — addressed — catalog.go:259,262 now say another issue's branch takes its copy back, matching helptext/issue.md:111-116 and atlas issue-tracker.md:245,267.
- BR-19 — not-addressed — Named sites fixed via env.has in detailsAt/restoreToHead, but republish.go:272 still turns a failed git status into "nothing to commit" silently, and unclaim.go:176 drops a Snapshot error.
- BR-20 — addressed — removeAside runs after a successful write-back and after a safe put-back; kept only on the path whose error names them.
- BR-21 — addressed — branchcreate.go:76,89 use issue.BranchName; closetracker.go uses issue.Stem for the plan glob, which is a stem, not a branch.
- BR-22 — addressed — restoreToHead is the single helper returning an error, used by finishPublished and bringMainIn.

## Round 6 — 2026-10-07T16:09:08-07:00 (claude) — BLOCKED

### Raised

- **BR-23** [Important] `rerun-not-idempotent` Takeover whose card write loses its response cannot be finished by rerunning claim
  uncertainCardWrite returns before finishTakeover runs. On rerun, prepareTakeover returns nil (the card is owned) and the decision gives errAlreadyMine. finishOwnedTakeover needs a local refs/heads/<branch> that a fresh clone never created, so it returns without a word ("nothing to do"). The release head was already spent, so nothing records the tip to finish at. The atlas and catalog LostResponse claim a rerun finishes this. Variant: if branch -f fails with a stale ancestor branch present, the rerun switches to the stale tip and prints resumed. 3rd in the family. Rule: a rerun-to-finish step must derive its target from state that outlives the effect it follows, and every effect boundary that promises a rerun needs a lost-response test. Fix: finishOwnedTakeover fetches the issue branch and creates or fast-forwards it (refusing divergence) before switching; add loseResponses tests. The same rule covers the handoff note's date-keyed dedupe.
- **BR-24** [Important] `stale-observation-not-rechecked-in-cas` Takeover CAS does not recheck the release head prepareTakeover observed; the handoff lease cannot catch it
  The single-issue claimSetDecision accepts any unowned card after a re-read. A re-release at a new head H2 between prepareTakeover (saw H1) and the CAS lands the claim, and finishTakeover sets the branch at stale H1. The next handoff's force-with-lease uses a freshly read ls-remote tip, which is effectively a force push that overwrites H2 on origin. Fix: pass the expected release head into the decision and refuse on mismatch; lease the push on the last-fetched remote-tracking ref; add a beforePush interleaving test.
- **BR-25** [Minor] `silent-error-in-io-glue` finishHandoff and finishOwnedTakeover return early on errors (CardRelease, rev-parse, a dirty tree) without saying why
  The operator gets "nothing to release" or "nothing to do" with no hint about why the checkout did not move. A one-line cwarn per early return fixes it.
- **BR-26** [Minor] `duplicated-restore-logic` Three finish helpers repeat the same clean-check, switch and warn sequence

## Round 7 — 2026-10-07T21:11:08-07:00 (claude) — BLOCKED

### Disposed

- BR-23 — addressed — finishOwnedTakeover fetches the remote branch and fast-forwards; TestTakeoverLostResponseRerunResumes fails without it (old code needed a local ref); note dedupe spans dates.
- BR-24 — addressed — decide rechecks release branch/head (claim.go:188); lease uses last-fetched tracking ref; both tests fail with the fix removed.
- BR-25 — not-addressed — Mostly fixed, but finishOwnedTakeover's fetch failure (handoff.go:269-271) still returns silently, conflating a network error with "no branch".
- BR-26 — addressed — switchClean (handoff.go:142) replaces the three switch sequences.

### Raised

- **BR-27** [Important] `test-gap-lost-response-set` finishOwnedTakeover's diverged-local-copy refusal has no test; removing it lets branch -f overwrite local commits
  2nd in family. Rule: every refusal branch on a rerun-to-finish path gets a fixture, not just the happy resume. Add a lost-response rerun test where the peer holds a diverged local issue branch; assert warn, branch unchanged, no switch.
- **BR-28** [Minor] `refusal-reason-mismatch` Handoff and takeover warnings name the wrong cause (ahead treated as diverged, every push error treated as a lease miss)
  2nd in family. Rule: a refusal names the condition actually observed and is classified before wording. Sites: handoff.go:278 (local ahead of remote, i.e. the owner's unpushed work, and merge-base errors, both reported as needing reconcile); handoff.go:181 (every push error reworded as a lease mismatch); handoff.go:270 (fetch error read as "no branch").
- **BR-29** [Minor] `duplicated-path-derivation` The remote-tracking ref string is built three times in handoff.go and in at least three other files
  4th in family. Rule: each derived ref or path name has one constructor. Expose a remoteTrackingRef(remote, branch) (gitx/trunkfile.go:194 already has one) and use it in handoff.go:170,222,268, transferguard.go:112 and landing.go:117.
- **BR-30** [Minor] `rerun-not-idempotent` Note dedupe matches the same text at any date, so a later separate handoff's identical note is silently dropped
  4th in family. Rule: a convergence key must identify this attempt, not just its content. Here, look only in the branch's unpushed tail (commits after the tracking ref) for the note commit, instead of searching the whole Log for the text.

## Round 8 — 2026-10-07T21:21:23-07:00 (claude) — passed

### Disposed

- BR-25 — addressed — finishHandoff warns on an unreadable release and a HEAD read failure; finishOwnedTakeover warns on ls-remote, fetch, rev-parse and merge-base failures; dirty trees warn via switchClean. Remaining silent returns are genuine nothing-to-do cases.
- BR-27 — addressed — TestTakeoverRerunWithALocalCopy/diverged asserts the warning, the local ref unchanged and no switch; removing the refusal at handoff.go:348 makes branch -f overwrite the ref and the test fails.
- BR-28 — addressed — Ahead and diverged are now distinct (handoff.go:342-348), merge-base errors have their own message, the push error is classified on stale info (handoff.go:215, pinned by TestHandoffLeaseRefusesAnUnseenRemoteTip), and ls-remote separates a missing branch from a failed fetch.
- BR-29 — not-addressed — Constructor added and used at the named sites, but siblings remain: landing.go:86 (next to mainRef in the same file), observe.go:117 and :181 (same file as the converted :237), gitx/publicationtarget.go:56.
- BR-30 — not-addressed — The commit-keyed check fixes the across-midnight rerun, but the handoff path still runs appendUnclaimNote's same-day text match, so a second separate handoff the same day with the same note is dropped; let the handoff rely on handoffNoteCommitted alone.

### Raised

- **BR-31** [Minor] `silent-error-in-io-glue` A failed check for the local branch is read as "absent" and followed by branch -f (handoff.go:270, :323-324)
  This is the 4th finding in family silent-error-in-io-glue. Rule: every existence check in the handoff glue returns three outcomes (present, absent, error) through env.gitTest, and an error stops with a warning; it is never folded into "absent". Apply it to both rev-parse --verify sites; the earlier-round sites already follow it.

## Open findings

- **BR-1** [Minor] `silent-error-in-io-glue` refreshAfterClaim swallows rev-parse errors and misreports merge-base failures as "commits main lacks"
- **BR-2** [Minor] `duplicated-id-rendering` claimRefs, claimArg and cardsMessage each re-implement the CLIRef join; ChangeCards inlines validateReplacement's budget math
- **BR-3** [Minor] `api-trusts-caller-normalization` tracker.ChangeCards does not dedupe ids; relies on callers (claimIssues) to do so
- **BR-4** [Minor] `refusal-points-at-unusable-action` a started unowned card in a set is refused toward --adopt, which itself refuses sets
- **BR-5** [Minor] `hand-recorded-actuals` project file ticks M1 with actual/closed before milestone-close measures it; M lines not nested under the #284 line
- **BR-6** [Minor] `test-gap-lost-response-set` no test drives a claim set through a lost publication response and a rerun
- **BR-19** [Minor] `silent-error-in-io-glue` gitTest errors discarded at republish.go:242,350 turn a failed probe into "not in HEAD" and then remove the file
- **BR-29** [Minor] `duplicated-path-derivation` The remote-tracking ref string is built three times in handoff.go and in at least three other files
- **BR-30** [Minor] `rerun-not-idempotent` Note dedupe matches the same text at any date, so a later separate handoff's identical note is silently dropped
- **BR-31** [Minor] `silent-error-in-io-glue` A failed check for the local branch is read as "absent" and followed by branch -f (handoff.go:270, :323-324)
