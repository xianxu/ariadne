# Boundary Review — ariadne#286 (milestone M2)

| field | value |
|-------|-------|
| issue | 286 — Boundary pushes and sdlc abandon |
| repo | ariadne |
| issue file | workshop/issues/000286-boundary-push-abandon.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 3feb1add521deeb64b7b2bed6b12a70c0e622f7c..34b67707cbf33f7ce6edfc8513d6ea1deb52999a |
| command | sdlc milestone-close --issue 286 --milestone M2 |
| reviewer | claude |
| timestamp | 2026-10-08T17:03:41-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

M2 delivers what the Spec asks for. `sdlc abandon` keeps the branch tip under `refs/ariadne/abandoned/NNNNNN`, makes the card terminal with a validated `{ref, branch, head}` record, archives the details on main and deletes the branch locally and on origin. `set-status` refuses `wontfix`/`punt` for started work, `--force` included. Reopening restores the branch, merges main, un-archives the details and lands through the transfer guard. I ran the targeted tests at HEAD `34b67707` and they pass: `go test ./cmd/sdlc -run 'Abandon|Reopen|SetStatusRedirects|ArchivedDetails'` and the `internal/issue` record test. Two Important findings stop this from being a clean SHIP:
- **Plan files on main:** abandon archives main's copy of the issue's plan files instead of the kept tip's copy. A reopen then silently replaces any plan edits made on the branch, and no test covers plan files at all.
- **README:** it has no entry for the new verb.

**Prior finding BR-1:** addressed. The plan has no `file.go:NNN` anchors left (grep found none), and the Revisions entry records the change.

### 1. Strengths
- **Rerun detection runs before the branch checks** (`abandon.go:124-162`). `TestAbandonRerunResumes` checks four real interruption points, including a rerun from the resting branch, and asserts there is exactly one abandon note.
- **Leases follow what is actually on the remote** (`abandon.go:215-239`, `:352-372`):
  - The archive-ref push accepts the ref absent, already at the tip, or at an ancestor of the tip.
  - The remote branch is deleted only when the kept tip contains it.
  - The local branch is deleted only when it is still at the recorded head.
- **The card record is checked when written and when read back** (`internal/issue/abandoned.go:31-48`): all fields or none, the ref pattern, the branch name rules and a full object ID. A reopen also checks the fetched ref against the recorded head (`abandon.go:427-431`). That covers ARCH-SECURE.
- **One projection is reused:** abandon's archive goes through `archivedDetails` (now any terminal status), `archiveDestination` and `planArtifactBelongsToIssue`, with no parallel archiver (ARCH-DRY).
- **The redirect sits inside `statusDecision`, before the `--force` branch** (`setstatus.go:139-143`), so it really cannot be waived. Both the forced and unforced cases are tested.

### 2. Critical findings
None.

### 3. Important findings
- **Plan files are archived from main, not from the kept tip** (`abandon.go:281-289`, `:490`, `:562`).
  - For started work, the details come from the kept tip (`abandonedDetails`), but plan files are copied from main's view (`p.Content`).
  - If the issue's plan is on main and was edited on the branch, main's archive holds the stale copy.
  - On reopen, the merge shows a modify/delete conflict on the plan. `resolveArchiveConflicts` resolves it with `git rm`, which takes main's side. Step 4 then removes the live copy and moves the stale archived one back, with a comment that is wrong for plans: "the archived copy is newer".
  - Result: the branch's plan edits silently disappear from the working tree. They survive only in git history.
  - Rule: for started work, everything abandon archives comes from the kept tip.
  - Fix: read plan contents from `rec.Head`'s tree when present there, falling back to main's copy. Add a test where the plan is on main, edited on the branch, abandoned and reopened, asserting the branch's plan content survives.
- **README update appears missing for `sdlc abandon` and the set-status redirect.** README's "Concurrent issue work" section lists the lifecycle verbs (claim, unclaim, reclaim, start-plan, move, issue show) but has no mention of abandon, the archive ref or the reopen restore.

### 4. Minor findings
- **The resume path skips the ownership check** (`abandon.go:128-162`). Plan D7 says a resume still checks "ownership by attribution and a clean tree"; the code checks only a clean tree. A foreign workspace can run the branch deletion. The archive-ref containment check limits the damage.
- **Reopen and abandon read their directories from different sources.** `restoreAbandoned` takes `WF_HISTORY_DIR`/`WF_PLANS_DIR` from the environment (`abandon.go:451-452`). `abandon` takes `--history-dir`/`--plans-dir` flags. A custom flag value at abandon is not seen by the reopen.
- **The plan's Core concepts table no longer matches the code:**
  - `ClearCardAbandoned` does not exist (clearing is `SetCardAbandoned(nil)`).
  - `abandonDecision(card, as, today, ref, head)` is actually `(card, as, today, rec)`.
  - The record is described as `{Ref, Head}` but has `Branch` too.
  - `abandonEvent` is a hard-coded map, not "via the vocabulary".
  - The rubric calls this Critical; I rate it Minor because the behavior is correct and only the names drifted.
- **3rd plan-vs-tests gap in this issue (family `test-claim-exceeds-test`).**
  - Task 6 lists reopen rerun variants "interrupted after step 2" and "after step 6". Neither has a test.
  - The after-step-6 case also does not complete. After a lost response on the card write, the rerun takes `runCardUpdate`'s "already has that" path, and `finishReopen` never runs. The mirror commit and the archive-ref deletion are skipped, and the details mirror on the branch stays out of date.
  - Rule: every test variant the plan lists maps either to a named test in the contract's Proofs or to a Revisions entry that drops it.
- **Forced reopens to other statuses lose the record** (`setstatus.go:150-152`). `set-status --force` from punt to open or blocked clears the `abandoned` record (any terminal → non-terminal transition) without running the restore. The archive ref is left with nothing on the card pointing to it.
- **ARCH-DRY: "read a ref's tip from the remote" is written out four times**, as `ls-remote` plus `Fields[0]`: `abandon.go:216`, `abandon.go:355`, `boundarypush.go:81`, `handoff.go:284`. It should be one `remoteRefTip(git, remote, ref)` helper.

### 5. Test coverage notes
- These tests use real bare remotes with no mocks: interrupting `cardPublish` and `mainPublish` at their real seams, and a pre-push hook that fails.
- Missing:
  - any plan file in the abandon or reopen tests;
  - a foreign-owner rerun of abandon;
  - reopen reruns after step 2 and after the card write;
  - abandon from `blocked` or `codecomplete`, which only the decision table covers.

### 6. Architectural notes
| Principle | Result |
|---|---|
| ARCH-DRY | Pass, apart from the remote-tip duplication above |
| ARCH-PURE | Pass: `abandonDecision` and the record are pure and unit-tested; `TestAbandonDecision` covers every model status |
| ARCH-PURPOSE | Pass: every Done-when item is delivered and the done-reopen gap is filed as #305; the plan-content loss weakens "reopening restores the branch at that tip" |
| ARCH-MOCK | Pass: real git and bare remotes throughout |
| ARCH-CONSTRAINTS | Pass: abandon and reopen are interactive and refuse when the remote is unreachable, with no budget beyond the envelope's |
| ARCH-SECURE | Pass |
| ARCH-ORDER | Pass with the caveats above: the sequence is detected from observed state, not an explicit state machine, but D7/D9 list the states and the interruption tests check the end state; the gaps are the lost card-write response on reopen and the ownership check skipped on resume |
| ARCH-FUNERAL | Pass: archive refs have a stated lifetime (removed by reopen, overwritten by a later abandon, otherwise bounded by the number of abandoned issues); remote issue branches are now removed |

### 7. Plan revision recommendations
- A Revisions entry fixing the Core concepts row for the record: `{Ref, Branch, Head}`, cleared with `SetCardAbandoned(nil)`, and `abandonDecision(card, as, today, rec)`.
- A Revisions entry on D7 step 5 / D9 step 4: plan files are archived from the kept tip for started work.
- Either add tests for the reopen rerun variants after step 2 and after the card write, or record their removal in Revisions.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      The plan has no file.go:NNN anchors left (grep found none); functions are named and a Revisions entry records it.
findings:
  - id: new
    severity: Important
    family: archive-source-of-truth
    title: |
      abandon archives plan files from main, not the kept tip, so a reopen replaces the branch's plan edits
    detail: |
      abandon.go:281-289 copies p.Content from main's view, while the details come from rec.Head. Reopen's resolveArchiveConflicts takes main's side with git rm, and step 4 at abandon.go:490 removes the live copy and moves the stale archived plan back. Rule: for started work, everything abandon archives comes from the kept tip. No test covers plan files in abandon or reopen.
  - id: new
    severity: Important
    family: readme-verb-surface
    title: |
      README update appears missing for sdlc abandon, the archive ref and the set-status redirect
    detail: |
      README's Concurrent issue work section lists every ownership and lifecycle verb (claim, unclaim, reclaim, move, issue show) but not abandon or the reopen restore.
  - id: new
    severity: Minor
    family: rerun-skips-precondition
    title: |
      the abandon resume path skips the ownership check that plan D7 keeps for reruns
    detail: |
      abandon.go:128-162 checks only a clean tree on resume; a foreign workspace can run dropAbandonedBranch. The containment lease limits the harm.
  - id: new
    severity: Minor
    family: config-source-divergence
    title: |
      restoreAbandoned reads WF_HISTORY_DIR/WF_PLANS_DIR from the environment while abandon takes --history-dir/--plans-dir flags
  - id: new
    severity: Minor
    family: plan-table-drift
    title: |
      the Core concepts table names ClearCardAbandoned, a 5-argument abandonDecision and a {Ref, Head} record, none of which match the code
  - id: new
    severity: Minor
    family: test-claim-exceeds-test
    title: |
      reopen rerun variants after step 2 and after the card write are untested, and the after-card-write rerun never runs finishReopen
    detail: |
      This is the 2nd finding in this family. Rule: every test variant the plan lists maps to a named test in the contract's Proofs or to a Revisions entry that drops it. After a lost response on the card write, the rerun returns "already has that" and skips the mirror commit and the archive-ref deletion.
  - id: new
    severity: Minor
    family: record-cleared-without-effect
    title: |
      a forced set-status from punt to open or blocked clears the abandoned record without restoring, leaving an archive ref nothing points to
  - id: new
    severity: Minor
    family: shared-remote-ref-read
    title: |
      reading a ref's tip from the remote via ls-remote is duplicated at abandon.go:216, abandon.go:355, boundarypush.go:81 and handoff.go:284
```
