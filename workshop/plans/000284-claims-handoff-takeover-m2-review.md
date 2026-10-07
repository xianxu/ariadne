# Boundary Review — ariadne#284 (milestone M2)

| field | value |
|-------|-------|
| issue | 284 — Claims: multi-claim, publish, handoff, takeover |
| repo | ariadne |
| issue file | workshop/issues/000284-claims-handoff-takeover.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | bf717a7aad28d7e715efc68508c5e5bc9d3935ba..0964427cd6cde7d092e36ae67255a59266414263 |
| command | sdlc milestone-close --issue 284 --milestone M2 |
| reviewer | claude |
| timestamp | 2026-10-07T15:10:20-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

M2 delivers what its Done-when asks for:
- `issue publish --issue` covers both first publication (through `move-detail`'s transfer) and republishing.
- The base check refuses to overwrite a main copy that moved since your base.
- Unclaiming an open claim publishes its edits, then releases the whole set in one tracker commit.
- `issue sync` in a tracker repository is now a pointer to `git commit` and `issue publish`.
- AGENTS.base §2 and §14 state the resting-branch rule.
- Task 8 (the release record, and the envelope keeping unknown keys) was pulled into M2, with a recorded Revision.

I ran the targeted tests at the head commit (`0964427c`): `internal/{issue,tracker,recovery}` plus the cmd/sdlc tests matching `Publish|Unclaim|Republish|ClaimSet|ClaimClears|LegacyIssueEntry|Recovery|Catalog`. All green.

Three things should be fixed before the boundary:
- **`unclaim --note` repeats on a rerun.** The note is written before anything that can fail, so any rerun adds it again.
- **The "main moved" refusal names a remedy that can't work** from the state that triggers it.
- **README and `atlas/workflow/sdlc-binary.md` are stale.** They still tell readers to run `issue sync`, which now refuses in tracker repositories, and don't mention `unclaim` or the new `issue publish`.

**1. Strengths**
- `republishOwned` (`cmd/sdlc/republish.go:55`) judges main once up front and again inside `prepare` on every attempt. It re-checks ownership against a fresh tracker read in `beforePush`, and it compares bodies only, so the card-mirror frontmatter doesn't cause false refusals.
- `finishPublished` on a resting branch checks that each file is unchanged since it was read before restoring it and fast-forwarding. A rerun after a lost push becomes a finish-only pass, and `TestPublishRerunAfterALostResponse` proves it with a push that actually lands.
- The release record (`internal/issue/release.go`) is validated when it is read. Malformed branch names (`..`, a leading `-`, path-shaped names, `main`) and raw machine IDs are rejected. `trackerEnvelope.Extra` keeps envelope keys this binary doesn't know through a rewrite, and `TestEnvelopeKeepsUnknownKeys` tests that.
- Good consolidation for ARCH-DRY: `fastForwardRest`, `issue.JoinRefs` and `Snapshot.validateReplacements` (where the single-card check is now just the one-element case).
- The recovery catalog gets real contracts for `issue publish`, `unclaim` and the retired `issue sync`, each with named proof tests, in place of the old Exempt entry. That keeps the ARCH-PURPOSE commitment from #283's lesson.

**2. Critical:** none.

**3. Important**
- **`cmd/sdlc/unclaim.go:184-199`: the note repeats on rerun.** The `--note` line is added to the local details before `republishOwned` and the card write. If either fails, the command says to rerun with the same arguments: a refused publish says "the claim is kept; … rerun `sdlc unclaim`", and a failed card write goes through `uncertainCardWrite`. The rerun finds the card still pending and adds the note again (`insertLogLine` doesn't dedupe), so main ends up with two "unclaimed:" lines. That breaks the ConvergentRetry class the catalog claims. Fix: skip a note that is already in the Log, or make the note-plus-publish a single step judged against main's copy. Add a test where the publish is refused, then the rerun succeeds, with exactly one note line on main.
- **`cmd/sdlc/republish.go:110`: the refusal points at a remedy that can't work.** This is the 2nd finding in family `refusal-points-at-unusable-action`. The "main changed" refusal tells you to run `sdlc claim --issue N` on a resting branch to fast-forward. That refusal fires exactly when main changed this file and the local copy has edits, and `git merge --ff-only` refuses to overwrite a dirty tracked file. So the repeat claim only warns "a local change is in the way". The rule covering both instances: **every remedy a refusal names must be runnable from the state that triggered the refusal, and a test must run it from that state.** Apply that rule here: name a remedy that works (save the edits, fast-forward, re-apply), or have publish do the re-apply itself. Then sweep the other refusals in this diff the same way: `requireOwnedToPublish`, unclaim's "claim is kept", and publish's "--commit retired".
- **README is stale.** `README.md:19-23` (Concurrent issue work) still says "`sdlc issue sync --issue N` checkpoints it; … initial details: `sdlc issue move-detail`". In a tracker repository `issue sync` now refuses, and it doesn't mention `unclaim`, `issue publish --issue`, or the resting-branch rule.
- **Atlas is stale.** `atlas/workflow/sdlc-binary.md:43-72`: the verb table has no `unclaim` row, still lists `issue publish` as "(Retiring at the #252 cutover)" with only `--commit`, and the publication-unit table has no `issue publish --issue` row. `issue-tracker.md` was updated, but the binary's own map wasn't swept (ARCH-PURPOSE).

**4. Minor**
- `cmd/sdlc/unclaim.go:155`: for a terminal status (`done`) the refusal says "releasing started work is a handoff, which comes with #284 M3". That's the wrong reason; split terminal statuses from started ones.
- `cmd/sdlc/republish.go:230`: the off-rest path of `finishPublished` overwrites the file with `it.publish` without the "unchanged since read" check the rest path does. An edit made during the push window is lost.
- `finishPublished` off a resting branch commits on whatever branch is checked out. Unclaiming `#9` from `#12`'s branch adds a `#9:` commit to `#12`'s branch, which works against the one-issue-one-branch rule (#272). Refuse, or commit only on the issue's own branch.
- `finishPublished` off a resting branch: if the details file is untracked in HEAD, `git diff --quiet HEAD` reports no difference, so nothing is committed and the branch's later merge hits an untracked-file conflict.
- `TestRepublishDecision` repeats the implementation's switch as its oracle (ARCH-ORDER). The added "never overwrite a moved main" assertion is the real invariant; the duplicated switch adds nothing.
- `TestPublishRefusals`: the `if r.originMain() == before { return }` branch is effectively dead after `peerAdd`.
- ARCH-DRY: the details path `filepath.Join(env.root, filepath.FromSlash(path.Join(dirs.Rel[0], path.Base(card.Path))))` is spelled out three times across `unclaim.go` and `republish.go`. Extract a helper.
- The base read in `republish.go:91` treats any `git show` error as "no base", which fails closed (refuse). It's safe, but the refusal text then says main moved when the real cause was a failed read.

**5. Test coverage notes**
- Covered: the publish happy path from a resting branch and from an issue branch; two issues in one main commit; refusals for an unowned card, someone else's card, and a moved main; a reclaim landing before the push; a lost response; first publication as `move-detail`; the full status × owner × release table for `unclaimDecision`; the unclaim set as one tracker commit; the unclaim lost response; a claim clearing a release.
- Missing:
  - unclaim reruns after a *refused* publish, and after a card write that failed without landing, both with `--note` (this would catch the first Important finding);
  - a test that runs the "bring main in first" remedy from the refused state;
  - the plan's open-unclaim rerun on a resting branch that also checks the fast-forward.

**6. Architectural notes for upcoming work**
- ARCH-DRY: pass, apart from the path helper above.
- ARCH-PURE: pass. `republishDecision` and `unclaimDecision` are pure; the IO is in `republishOwned` and `runUnclaim`.
- ARCH-PURPOSE: flag (README and the `sdlc-binary.md` atlas sweep).
- ARCH-MOCK: pass. The `mainPublish` and `cardsPublish` seams run over real git fixtures.
- ARCH-CONSTRAINTS: pass. Each publish adds one merge-base and one `git show` per issue.
- ARCH-SECURE: pass. The release is validated when read, unknown envelope keys are kept as data rather than trusted, and a failed base read refuses.
- ARCH-ORDER: flag. The note is written before the steps that can fail, so publish → card is not convergent. M3's handoff (note commit → push → card → switch) needs its note effect keyed so a rerun recognises it, in the same way the release does.
- ARCH-FUNERAL: pass. Any claim clears the release; an unused one is a single small record per card.

**7. Plan revision recommendations**
- Add a Revisions entry saying the M2 open-unclaim note is idempotent across reruns, and how (for example, it is skipped when that line is already in main's Log), with the test named.
- In Task 7's atlas bullet, add `sdlc-binary.md` (the verb and publication tables) and the README tracker paragraph, so the plan stops implying `issue-tracker.md` and `issue-lifecycle.md` were the whole sweep.

```findings
findings:
  - id: new
    severity: Important
    family: rerun-not-idempotent
    title: |
      unclaim --note is appended locally before publish and the card write; a rerun after either fails duplicates the Log line on main
    detail: |
      cmd/sdlc/unclaim.go:184-199 writes the note, then republishOwned, then cardsPublish. A refused publish ("the claim is kept; rerun") or a failed card write tells the operator to rerun the same command, and insertLogLine does not dedupe, so the note lands twice. Make the note convergent (skip it when the line is already present, or judge it against main's copy) and test refused-publish then rerun.
  - id: new
    severity: Important
    family: refusal-points-at-unusable-action
    title: |
      republish's main-moved refusal advises `sdlc claim` to fast-forward the rest, which git refuses in exactly that state
    detail: |
      This is the 2nd finding in this family. republish.go:110 fires when main changed this file and the local copy is edited; claim's repeat path runs fastForwardRest, whose git merge --ff-only refuses because a dirty tracked file would be overwritten, so it only warns. Rule: every remedy a refusal names must be runnable from the refused state, proven by a test that runs it from that state. Fix the remedy (save the edits, fast-forward, re-apply, or have publish do it) and sweep the diff's other refusals the same way.
  - id: new
    severity: Important
    family: docs-sweep-missing
    title: |
      README and atlas/workflow/sdlc-binary.md still describe issue sync / move-detail as the tracker path; unclaim and issue publish --issue are absent
    detail: |
      README.md:19-23 says `sdlc issue sync --issue N` checkpoints on the issue branch, but it now refuses in tracker repositories. sdlc-binary.md:43-72 has no unclaim row, lists issue publish as "Retiring at the #252 cutover", and has no issue publish --issue row.
  - id: new
    severity: Minor
    family: refusal-reason-mismatch
    title: |
      unclaim on a terminal status says "releasing started work is a handoff (M3)"
    detail: |
      unclaim.go:155 tests !IsOpen, which also catches done/terminal statuses; give terminal statuses their own reason.
  - id: new
    severity: Minor
    family: rerun-not-idempotent
    title: |
      finishPublished off-rest overwrites the details without the unchanged-since-read check, and commits on whatever branch is checked out
    detail: |
      republish.go:230 can drop an edit made during the push, can add an #A commit onto #B's issue branch (against #272), and skips the commit when the file is untracked in HEAD.
  - id: new
    severity: Minor
    family: duplicated-path-derivation
    title: |
      The local details path is spelled out three times across unclaim.go and republish.go
    detail: |
      filepath.Join(env.root, filepath.FromSlash(path.Join(dirs.Rel[0], path.Base(card.Path)))) should be one helper.
```

---

## Re-review — 2026-10-07T15:26:36-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 284 — Claims: multi-claim, publish, handoff, takeover |
| repo | ariadne |
| issue file | workshop/issues/000284-claims-handoff-takeover.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | bf717a7aad28d7e715efc68508c5e5bc9d3935ba..c2871a48ba9ce00e43d9eea578f1d8da80476a95 |
| command | sdlc milestone-close --issue 284 --milestone M2 |
| reviewer | claude |
| timestamp | 2026-10-07T15:26:36-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

I found two Important bugs in the new merge-main-in path, so the verdict is FIX-THEN-SHIP.

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

Round 2 of the M2 review covered the fix commit `c2871a48` and the whole M2 window. The fixes for BR-7, BR-9, BR-10, BR-11 and BR-12 hold, and each behavior change has a test that would fail without it. BR-8's fix works on a resting branch: publish now merges a moved main in itself, and `TestPublishBringsAMovedMainIn` runs the remedy from the refused state. Three problems remain:
- **Conflict markers can reach main.** After a conflicted merge, nothing stops a rerun from publishing a file that still has unresolved markers.
- **Edits can be lost.** If setting the edits aside fails partway, earlier files are already reverted, but the error says "the local file is unchanged".
- **BR-8's sweep wasn't done.** The rewritten refusal off a resting branch still names a remedy that can't run when the edit is uncommitted, and no test runs it.

The help text and atlas also still describe the old "main moved → refused" behavior.

**1. Strengths**
- `republishDecision` (`cmd/sdlc/republish.go:35`) stays a small pure function, and the merge-in path re-enters it instead of branching around it.
- `finishPublished` now checks that every file is unchanged since it was read, on both branch kinds, and won't commit on another issue's branch. `TestPublishFromAnotherIssuesBranchCommitsNothingThere` pins the #272 rule.
- The BR-7 test injects a failing `cardsPublish` and checks main's copy, not the local file. That's the right place to look.
- The lessons entries (`workshop/lessons.md:341-349`) state the rule behind each finding, not just the incident.

**2. Critical:** none.

**3. Important**
- **Unresolved conflict markers are published** (`republish.go:120-131`, `bringMainIn` at line ~330).
  - After a conflict, HEAD equals main, so the base equals main and a rerun gets `publishWrite`. A rerun without resolving (an agent retrying the command, or `unclaim`'s "rerun") pushes `<<<<<<< yours` blocks to main.
  - Fix: refuse to publish any body that still has conflict-marker lines, naming the file. Test: conflict, then rerun without resolving, then main is unchanged.
- **Setting edits aside isn't all-or-nothing** (`republish.go` `bringMainIn`; ARCH-ORDER).
  - The loop runs `git checkout HEAD -- path` on each moved file and returns at the first failure, after earlier files were already reverted. The error still says "the local file is unchanged".
  - The likely failure is a file that isn't in HEAD: `republishDecision` sends a missing base to the merge path, and `mergeDetails` even handles an empty base, but `checkout HEAD` can't restore a path HEAD doesn't have.
  - The restore after a failed fast-forward uses `_ = os.WriteFile`, so a failure to put the user's edits back is silent.
  - Fix: one set-aside step that removes untracked files and checks out tracked ones, and on any failure writes every touched file back from `item.read`, surfacing any write error. Test: a set of two where the second file's base is missing.
- **BR-8, not addressed: the sweep.** This is the 3rd finding in family `refusal-points-at-unusable-action`.
  - Line 123 off a resting branch says "merge main into <branch>, then rerun". With the edit uncommitted, `git merge` refuses for the same dirty-file reason as before. It also doesn't say which ref ("main" or `view.Ref()`).
  - The race refusal at line 158 ("bring main in first") and `bringMainIn`'s failure messages name no runnable step either.
  - Rule from `lessons.md`: every refusal's remedy runs from the refused state, proven by a test. Fix it as a table of the verb's refusals, each running the remedy it names, not one message at a time.
- **Help text and atlas weren't updated for the new behavior** (family `docs-sweep-missing`, 2nd).
  - `helptext/issue.md:103` still says "main moved → refused, bring main in first", and line 108 says any other branch "commits the same bytes narrowly".
  - `atlas/workflow/issue-tracker.md:245,264-268` says the same.
  - Rule: a fix round that changes a verb's behavior updates that verb's help text and atlas section in the same commit.

**4. Minor**
- The BR-7 dedup line includes today's date, so a rerun after midnight still adds a second note.
- `republish.go:262` derives the branch name inline. This is the 2nd in family `duplicated-path-derivation`; rule: one helper for each name derived from a card path. The same expression already appears at `planningbranch.go:29`, `transferguard.go:69`, `observe/assemble.go:179` and `claimant.go:190`.
- ARCH-MOCK: `mergeDetails` calls `exec.Command("git", …)` directly instead of going through the `env.git` seam.
- `bringMainIn` merges whole files including frontmatter. A conflict in the card mirror would break parsing on the rerun with no remedy named.

**5. Test coverage**
- New tests: unclaim note convergence, the moved-main merge (clean and conflicted), and publishing from another issue's branch.
- Missing:
  - rerun with unresolved markers;
  - a moved file whose base is missing at the merge base;
  - a failed fast-forward that restores the edits;
  - the off-rest refusal's remedy, run from the refused state.

**6. Architecture (each principle)**
- ARCH-DRY: minor flag (branch name).
- ARCH-PURE: pass.
- ARCH-PURPOSE: pass.
- ARCH-MOCK: minor flag (`merge-file` outside the seam).
- ARCH-CONSTRAINTS: N/A; a few small files per invocation.
- ARCH-SECURE: flag; publish reads its own half-finished (conflicted) file as valid input.
- ARCH-ORDER: flag; the set-aside → fast-forward → write-back sequence has no complete rollback.
- ARCH-FUNERAL: pass; the temp dir is removed with `defer`, and no new durable families.

**7. Plan revisions:** none needed. The M2 plan entries still match the delivered scope; the help-text and atlas drift is a docs fix, not a plan change.

```findings
dispose:
  - id: BR-7
    disposition: addressed
    note: |
      strings.Contains guard at unclaim.go:201 and TestUnclaimNoteIsConvergent fails without it; the date in the line makes a cross-midnight rerun duplicate (residual, minor)
  - id: BR-8
    disposition: not-addressed
    note: |
      resting branch fixed and tested; the sweep was not done: the off-rest refusal at republish.go:123 names a merge that is refused with an uncommitted edit, and no test runs it
  - id: BR-9
    disposition: addressed
    note: |
      README.md:19-24 and sdlc-binary.md rows for unclaim, issue publish --issue and legacy-only issue sync
  - id: BR-10
    disposition: addressed
    note: |
      CanHoldOwner case at unclaim.go:156 with its own reason; covered at unclaim_test.go:44
  - id: BR-11
    disposition: addressed
    note: |
      read-check applies on both branch kinds, commits only on the issue's own branch, untracked handled through the porcelain check; TestPublishFromAnotherIssuesBranchCommitsNothingThere
  - id: BR-12
    disposition: addressed
    note: |
      localDetail helper used by unclaim.go and republish.go
findings:
  - id: new
    severity: Important
    family: intermediate-state-consumed-unchecked
    title: |
      A rerun after a conflicted merge-in publishes unresolved conflict markers to main
    detail: |
      After bringMainIn conflicts, HEAD equals main, so base equals main and the rerun gets publishWrite with the markers in the body. Refuse a publish body that still has conflict-marker lines; test conflict then rerun without resolving leaves main unchanged.
  - id: new
    severity: Important
    family: partial-effect-not-rolled-back
    title: |
      bringMainIn's set-aside loop can revert earlier edits and then report the local file unchanged
    detail: |
      Checkout HEAD on each moved file returns at the first failure (for example a file not in HEAD, which the empty-base merge path admits) after earlier files were already reverted; the restore after a failed fast-forward discards its write error. Make set-aside all-or-nothing, writing every touched file back from item.read and surfacing any failure.
  - id: new
    severity: Important
    family: docs-sweep-missing
    title: |
      issue help text and the atlas still say main moved means refused, and any other branch commits the bytes
    detail: |
      This is the 2nd finding in this family. helptext/issue.md:103,108 and atlas/workflow/issue-tracker.md:245,264-268 predate c2871a48. Rule: a fix round that changes a verb's behavior updates that verb's help text and atlas section in the same commit.
  - id: new
    severity: Minor
    family: duplicated-path-derivation
    title: |
      republish.go:262 derives the issue branch from the details path inline again
    detail: |
      This is the 2nd finding in this family. Rule: one internal/issue helper for each name derived from a card path. The same expression is at planningbranch.go:29, transferguard.go:69, observe/assemble.go:179 and claimant.go:190.
  - id: new
    severity: Minor
    family: external-call-outside-seam
    title: |
      mergeDetails execs git merge-file directly instead of through the env git seam
```

---

## Re-review — 2026-10-07T15:38:08-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 284 — Claims: multi-claim, publish, handoff, takeover |
| repo | ariadne |
| issue file | workshop/issues/000284-claims-handoff-takeover.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | bf717a7aad28d7e715efc68508c5e5bc9d3935ba..ae8591e4e4b6de8e3392ce9fc58e6fd027cf454f |
| command | sdlc milestone-close --issue 284 --milestone M2 |
| reviewer | claude |
| timestamp | 2026-10-07T15:38:08-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

Round 3 fixed most of what it set out to fix. Running the new tests against the pre-fix code (c2871a48) in a scratch worktree confirms two fixes are real. `TestPublishRefusesConflictMarkers` (BR-13) fails without its fix, and so does `TestPublishMovedMainOnTheIssueBranchRemedy` (BR-8, issue-branch path). Help and atlas now describe the new behaviour (BR-15), `issue.BranchName` replaces the five named derivations (BR-16), and `merge-file` goes through `env.gitRaw` (BR-17). All `TestPublish|TestUnclaim|TestRepublish` tests pass at HEAD.

Three things are left:
1. **BR-14 has no regression evidence.** `TestPublishBringInPutsEditsBackOnFailure` also passes on the pre-fix code. It takes the fast-forward failure path, which the old code already wrote back.
2. **BR-8's sweep is incomplete.** Three refusals still don't name an action that works from where they fire, and the other-branch remedy has no test.
3. **The recovery catalog's `Effects` and `Repeat` text still describe the old finish behaviour.** This is the third finding in its family.

All three are cheap to fix.

1. **Strengths**
   - `republish.go:116`: the conflict-marker guard sits before `republishDecision`, so no branch type can publish markers. Its test is genuinely red without the fix.
   - `movedMainRefusal` (`republish.go:417`) splits the remedy by branch type. Its issue-branch test runs the named steps from the refused state through to a successful publish (`issuepublish_tracker_test.go:325`), which is exactly what BR-8's rule asks for.
   - Before it changes anything, `bringMainIn` (`republish.go:323-335`) copies every edit and computes every merge. The `putBack` closure collects all its failures instead of stopping at the first.
   - `issue.BranchName` (`internal/issue/filename.go:47`) is now the one derivation used by every site BR-16 named.
   - The lessons entry widens a verb's surface to help + atlas + recovery contract (`workshop/lessons.md:333`).

2. **Critical:** none.

3. **Important**
   - **BR-8 (not-addressed), family `refusal-points-at-unusable-action`.**
     - (a) A `--dry-run` on a resting branch reaches `movedMainRefusal` (`republish.go:125`). That refusal says "X is not #N's branch… publish from a resting branch", but the user is already on one.
     - (b) The in-prepare refusal at `republish.go:161` still says "bring main in first". On a resting branch the working remedy is simply "rerun".
     - (c) The other-branch remedy ("publish from a resting branch") has no test that runs it. The catalog says `TestPublishFromAnotherIssuesBranchCommitsNothingThere` proves it, but that test never moves main.
     - (d) The issue-branch remedy hardcodes `origin/main`, although the publication remote is the resting branch's upstream.
     - Fix: build every moved-main refusal through one function that takes branch type and dry-run state, and test each of its outputs by running the remedy it names.
   - **BR-14 (not-addressed), family `partial-effect-not-rolled-back`.** The cited test passes on pre-fix code. Two failures are still untested: a set-aside failure part-way through the files, and a file not in HEAD (the new `os.Remove` path). Also, a write failure after the fast-forward (`republish.go:362-365`) returns the raw error. At that point the files are already set aside, so the edits survive only in `publish-aside`, and the error doesn't say so. Route that failure through the same "copies are in …" message, and add a test that is red on c2871a48.
   - **New: the recovery catalog still describes the old finish behaviour.** `catalog.go:259` (`Effects`) still says "another branch commits the same bytes", and `catalog.go:262` (`Repeat`) still says "branch commit". Since this round, another issue's branch takes its copy back instead. This is the 3rd finding in family `docs-sweep-missing`. The rule is already in lessons (help, atlas and catalog change in the same commit as the behaviour), yet this round's commit updated `Preconditions` and missed the two lines beside it. Fix: a fix-round commit greps the old behaviour phrase across `helptext/`, `atlas/` and `catalog.go` before it commits.

4. **Minor**
   - **`silent-error-in-io-glue` (2nd in family).** At `republish.go:242` and `:350`, `inHead, _ := env.gitTest(...)` turns a probe error into "not in HEAD" and then calls `os.Remove`. `gitTest` is documented as never evidence of absence. Also, `republish.go:128` treats a `gitRaw` failure as an empty merge base. Rule: no `gitTest` result is used without its error.
   - **New family `artifact-without-removal` (ARCH-FUNERAL).** `<git-dir>/sdlc/publish-aside/` copies are written on every bring-in and never removed. Stale copies of other issues then sit beside the ones the "copies are in" message points to. Clear each copy after a successful write-back, or after a successful put-back.
   - **`duplicated-path-derivation` (3rd in family).** Rule: issue-branch names come only from `issue.BranchName`. `branchcreate.go:76,89` (the code that creates issue branches) and `closetracker.go:268` still derive inline. They are outside this window but are the same class.
   - **ARCH-DRY.** The logic "restore from HEAD, else remove" appears twice: in `finishPublished.restore` (`republish.go:241`) and in the `bringMainIn` set-aside loop (`:348`). Make it one helper.
   - `putBack` rewrites files with mode 0644 whatever their original mode was.

5. **Test coverage notes:** BR-13 and BR-8 (issue-branch) are proven red without their fixes. BR-14 is not. Other-branch moved-main, dry-run-on-rest moved-main, and a changed-during-prepare rerun on rest have no coverage. `hasConflictMarkers` is pure but has no unit test. Add cases for a lone `=======` line and for markers with no label.

6. **Architectural notes:**
   - **ARCH-DRY:** flagged (restore duplication).
   - **ARCH-PURE:** pass. `republishDecision`, `hasConflictMarkers` and `issue.BranchName` are pure.
   - **ARCH-PURPOSE:** flagged (BR-8 sweep incomplete).
   - **ARCH-MOCK:** pass. The `merge-file` call now goes through the env seam.
   - **ARCH-CONSTRAINTS:** pass. The work is per-issue and bounded.
   - **ARCH-SECURE:** pass. Local files and main's bytes are parsed before use, and markers are refused.
   - **ARCH-ORDER:** flagged (probe errors collapsed into absence; the post-fast-forward failure leaves partial progress unreported).
   - **ARCH-FUNERAL:** flagged (`publish-aside`).

7. **Plan revision recommendations:** none to the plan. The catalog `Proofs` row naming `TestPublishFromAnotherIssuesBranchCommitsNothingThere` as proof of the other-branch remedy should be corrected once a real test exists.

```findings
dispose:
  - id: BR-8
    disposition: not-addressed
    note: |
      Issue-branch remedy fixed and proven red-without-fix; but dry-run on rest gets the "not N's branch" refusal (republish.go:125), republish.go:161 still says "bring main in first", the other-branch remedy has no test that runs it (the catalog cites a test that never moves main), and origin/main is hardcoded.
  - id: BR-13
    disposition: addressed
    note: |
      republish.go:116 guard; TestPublishRefusesConflictMarkers fails on c2871a48 (verified in a scratch worktree).
  - id: BR-14
    disposition: not-addressed
    note: |
      The code looks right, but TestPublishBringInPutsEditsBackOnFailure passes on pre-fix c2871a48 (fast-forward failure path, old code wrote back too); set-aside failure part-way and not-in-HEAD paths untested; a write failure after the fast-forward (republish.go:362) returns a raw error without pointing at the aside copies.
  - id: BR-15
    disposition: addressed
    note: |
      helptext/issue.md:100-117 and atlas/workflow/issue-tracker.md:245,262-267 now match republish.go behaviour. Catalog residue raised separately.
  - id: BR-16
    disposition: addressed
    note: |
      All five named sites use issue.BranchName; siblings outside the window raised as a Minor.
  - id: BR-17
    disposition: addressed
    note: |
      mergeDetails uses env.gitRaw, which wraps the error with %w so the ExitError check still works; conflict tests exercise it.
findings:
  - id: new
    severity: Important
    family: docs-sweep-missing
    title: |
      Recovery catalog Effects/Repeat for issue publish still say "another branch commits the same bytes" / "branch commit"
    detail: |
      3rd in family. catalog.go:259,262 predate this round's change (another issue's branch now takes its copy back). The rule is already in lessons (help, atlas and catalog change in the same commit); this round's commit edited Preconditions next to these lines and missed them. Fix the instance and grep the replaced behaviour phrase across helptext/, atlas/ and catalog.go before each fix-round commit.
  - id: new
    severity: Minor
    family: silent-error-in-io-glue
    title: |
      gitTest errors discarded at republish.go:242,350 turn a failed probe into "not in HEAD" and then remove the file
    detail: |
      2nd in family. gitTest's contract says an error is never evidence of absence. republish.go:128 likewise treats a gitRaw failure as an empty merge base. Rule: no gitTest/gitRaw result is used without its error.
  - id: new
    severity: Minor
    family: artifact-without-removal
    title: |
      publish-aside copies under the git dir are never removed
    detail: |
      ARCH-FUNERAL. Written on every bring-in, keyed by basename, never cleared; stale copies of other issues sit beside the ones the "copies are in" error points at. Remove each copy after a successful write-back or put-back.
  - id: new
    severity: Minor
    family: duplicated-path-derivation
    title: |
      branchcreate.go:76,89 and closetracker.go:268 still derive the issue branch/stem inline
    detail: |
      3rd in family. Rule: issue-branch names come only from issue.BranchName; the branch creator is the most important consumer. These sites are outside the window.
  - id: new
    severity: Minor
    family: duplicated-restore-logic
    title: |
      "restore from HEAD, else remove" is implemented twice (finishPublished.restore and the bringMainIn set-aside loop)
    detail: |
      ARCH-DRY: extract one helper returning an error, which would also fix the discarded gitTest error at both sites.
```

---

## Re-review — 2026-10-07T15:49:53-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 284 — Claims: multi-claim, publish, handoff, takeover |
| repo | ariadne |
| issue file | workshop/issues/000284-claims-handoff-takeover.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | bf717a7aad28d7e715efc68508c5e5bc9d3935ba..03a0d47534c1d5a95e342355d1e13264226210e0 |
| command | sdlc milestone-close --issue 284 --milestone M2 |
| reviewer | claude |
| timestamp | 2026-10-07T15:49:53-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: medium
```

I checked the window bf717a7..03a0d47, starting with the stat and name-status recipes and then the full contents of `republish.go`, `unclaim.go`, the publish tests, the recovery catalog, `helptext/issue.md` and the atlas rows. I also ran the targeted suites: `go test ./cmd/sdlc -run 'TestPublish|TestRepublish|TestUnclaim|TestClaimClearsARelease|TestLegacyIssueEntrypoints'` and the `internal/issue`, `internal/recovery` and `internal/tracker` packages. All passed. Round 4 closes the Important findings that were still open. BR-8's remedies now work from the state where publish refused, and each has a test. BR-14's set-aside is all-or-nothing and reports when a put-back fails. BR-18's wording now matches across the catalog, help text and atlas. Of the Minor findings, BR-20, BR-21 and BR-22 are fixed in code. BR-19's rule is still broken at one site in the same file, and it doesn't block the boundary. Nothing blocks shipping.

1. **Strengths**
   - **BR-8: publish brings a moved main in itself on a resting branch.** `republish.go:129-142` hands off to `bringMainIn` (`republish.go:305-367`). The remedy on a resting branch is now the command you rerun, not a separate `git` step that git would refuse. `TestPublishBringsAMovedMainIn` runs both the clean-merge case and the conflict → resolve → rerun case from the state where publish refused.
   - **Each refusal's steps are tested from the refused state.** On the issue's own branch, `TestPublishMovedMainOnTheIssueBranchRemedy` runs the named commit, fetch and merge, then publishes. On another issue's branch, `TestPublishMovedMainFromAnotherBranchRemedy` runs the switch and the rerun, and also checks that a dry run on a resting branch names the bring-in.
   - **Set-aside is safe.** Every edit is copied under the git directory before anything is touched (`republish.go:322-334`). `putBack` writes every moved file back from `item.read` and lists every write that failed (`:335-347`). The copies are removed after a successful write-back or a safe put-back. `TestPublishBringInSetAsideIsAllOrNothing` and `TestPublishBringInWriteBackFailureNamesTheCopies` cover both paths.
   - **Probe failures are no longer read as absence.** `detailsAt` and `restoreToHead` (`republish.go:426-450`) go through `env.has` and return its error. "Restore from HEAD, else remove" now lives in one helper, used by both `finishPublished` and `bringMainIn` (BR-22).
   - **The rerun after a lost response is driven by main's state.** `republishDecision` is pure and has its own table test. When main already holds the details, the rerun only finishes the checkout (`TestPublishRerunAfterALostResponse`).

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings**
   - **`republish.go:272` swallows a status error.** `if dirty, err := env.git("status", ...); err != nil || dirty == "" { continue }` treats a failed status probe as "nothing to commit". The issue branch then silently doesn't get the published bytes. This is the BR-19 rule, so I've recorded it as BR-19 not-addressed rather than a new finding.
   - **`unclaim.go:176` hides a failed main read.** `if view, err := env.main.Snapshot(); err == nil {...}` skips the fast-forward without a warning. Same rule as above.
   - **The another-branch remedy assumes `git switch` carries the edit** (`republish.go:420`). That only holds when the resting branch's HEAD copy of the file matches this branch's copy. The test covers only the matching case. This is a narrow edge, so I'm noting it rather than raising it.

5. **Test coverage notes**
   - Failures are injected through two seams, `restoreToHead` and `writeLocal`, and the uncertain push through `mainPublish`. Together they cover set-aside failure, write-back failure, a lost push response, and a reclaim that lands before the push.
   - I didn't revert the fixes to watch the tests go red. Reading the BR-14 test, it injects a failure on the second set-aside and then asserts both files match their exact local edit. Without `putBack`, the first file would have been reverted to HEAD, so the test would fail.

6. **Architectural notes**
   - **ARCH-DRY: pass.** One restore helper, one `localDetail`, and `issue.BranchName` is used at `branchcreate.go:76,89` (BR-21).
   - **ARCH-PURE: pass.** `republishDecision`, `unclaimDecision` and `unclaimSetDecision` are pure and tested without IO. The IO glue is long but goes through seams.
   - **ARCH-PURPOSE: pass.** Help text, catalog and atlas all describe the same three finishing behaviours.
   - **ARCH-MOCK: pass.** Git goes through env seams, including `merge-file` through `env.gitRaw`.
   - **ARCH-CONSTRAINTS: pass.** The work is per-issue and bounded.
   - **ARCH-SECURE: pass.** Local details are parsed, conflict markers are refused, and ownership is re-checked against a fresh tracker read before the push.
   - **ARCH-ORDER: pass.** The verdict is an explicit enum, main is re-judged inside `prepare`, and the bring-in pass is bounded to one by the `bringIn` flag.
   - **ARCH-FUNERAL: pass.** The publish-aside copies are removed on success and on a safe put-back, and kept only when the error points the operator at them.
   - **For M3 (handoff of started work):** reuse `bringMainIn`'s copy, act, put back on failure shape rather than adding a second set-aside path.

7. **Plan revision recommendations:** none. The plan's Task 8 revision already covers what this window delivered.

```findings
dispose:
  - id: BR-8
    disposition: addressed
    note: |
      On a resting branch publish brings main in itself (bringMainIn); TestPublishBringsAMovedMainIn and both *Remedy tests run the named steps from the refused state.
  - id: BR-14
    disposition: addressed
    note: |
      putBack writes every moved file back from item.read and reports every failure; copies kept under the git dir; TestPublishBringInSetAsideIsAllOrNothing injects a 2nd-file failure.
  - id: BR-18
    disposition: addressed
    note: |
      catalog.go:259,262 now say another issue's branch takes its copy back, matching helptext/issue.md:111-116 and atlas issue-tracker.md:245,267.
  - id: BR-19
    disposition: not-addressed
    note: |
      Named sites fixed via env.has in detailsAt/restoreToHead, but republish.go:272 still turns a failed git status into "nothing to commit" silently, and unclaim.go:176 drops a Snapshot error.
  - id: BR-20
    disposition: addressed
    note: |
      removeAside runs after a successful write-back and after a safe put-back; kept only on the path whose error names them.
  - id: BR-21
    disposition: addressed
    note: |
      branchcreate.go:76,89 use issue.BranchName; closetracker.go uses issue.Stem for the plan glob, which is a stem, not a branch.
  - id: BR-22
    disposition: addressed
    note: |
      restoreToHead is the single helper returning an error, used by finishPublished and bringMainIn.
```
