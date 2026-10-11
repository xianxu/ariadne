# Boundary Review — ariadne#317 (whole-issue close)

| field | value |
|-------|-------|
| issue | 317 — active-time: every ref in a commit subject claims an equal share, so citations and merge subjects take the issue's time |
| repo | ariadne |
| issue file | workshop/issues/000317-active-time-every-ref-in-a-commit-subject-claims-an-equal-share-so-citations-and-merge-subjects-take-the-issue-s-time.md |
| boundary | whole-issue close |
| milestone | — |
| window | 8b9052040674db7b431ad3db3844837e0c561366..57bb1c283b62ec2b90139fada28f40d2d98a2f83 |
| command | sdlc close --issue 317 |
| reviewer | claude |
| timestamp | 2026-10-10T17:14:29-07:00 |
| verdict | SHIP |

## Review

Still reviewing: checking that the commit subjects sdlc writes itself still start with the issue ref.

```verdict
verdict: SHIP
confidence: high
```

The diff does what the issue asks. `issueref.Lead` defines a lead grammar with a closed list of verbs, and `Commit.Issues` now holds only a subject's lead refs, so only those claim time. `Commit.Refs` keeps every local ref so cited issues stay in the mention scope. In scoped mode, a main commit counts only when its lead names the issue, and the branch's own merges are dropped. The two regression tests I traced fail on the pre-fix code and pass after it:
- **Integration merge:** in `merged mid-work, subject names the issue`, the old code kept the `#9: merge main (#5 landed)` merge as a boundary at 00:35.
- **Citation:** in `TestCitationsDoNotClaim`, the old code kept the main commit `#5: … (#9)` as a boundary and gave #5 a share.

`go test` passes for `activetime` and `issueref`. Nothing blocks the close. Two doc comments slightly overstate the code (see Minor).

1. **Strengths**
   - The lead extractor lives in one place, `issueref.Lead` / `LeadLocalNums` (`ref.go:76-99`). It reuses `Find`, so the merge-PR masking from #254 carries over, and the private `localNums` is shared with `LocalNums` (ARCH-DRY: pass).
   - The verb list is closed on purpose and its comment says why. The test table (`ref_test.go:168`) includes the negative cases that matter: an open verb, prose before the ref, and a foreign lead whose citation must not claim.
   - `branchCommits` gets own commits and merges from one `rev-list --parents` call, instead of adding a second git call.
   - I checked the subjects sdlc writes itself. Tracker commits (`#N: tracker: …`), issue-new commits (`#N: issue: …`) and close commits all lead with the ref, so the claim and close anchors still work as boundaries.
   - `TestAttributionGolden` was changed on purpose (`#10, #8: second`), and the comment says why the old subject now encodes the old rule.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - **Merge claim is too broad.** The help text says "With --branch-point, merge commits are not boundaries", and the `Scope` doc says "a merge is never a scoped boundary". The code only drops merges inside `BranchPoint..HEAD`. A main-side merge outside that range whose lead names the issue is still kept. PR merges are masked, so this is mostly theoretical, but the claim should say "the branch's merges". The atlas wording ("On an issue branch…") is already accurate.
   - **Atlas verb list is incomplete.** It lists `close`/`file`/`issue`/`plan`, but the regex also accepts `closes`.

5. **Test coverage**
   - The table test covers every subject shape the Spec names, including the no-lead subject (it stays a neutral boundary).
   - The integration test checks both the boundary list and `Compute` hours (pre == post, #9 = 50 min, #5 = 0) across the not-integrated, rebased, merged-at-end and both merged-mid-work cases. So the "same before and after merging main" clause is tested both ways.
   - Not tested: a `Revert "#N: …"` subject. It now becomes neutral, which is arguably right.

6. **Architecture**
   - **ARCH-DRY: pass.** There is one lead extractor, and no other subject-lead regex exists in `cmd/sdlc` or `pkg`.
   - **ARCH-PURE: pass.** `Lead` is pure and tested without IO. The git IO stays in `branchCommits` / `loadWindowCommits`, which are tested through the `gitRun` shim and real temp repos.
   - **ARCH-PURPOSE: pass.** Both places that read claimants were updated: boundaries and attribution read `Issues`, and `WindowIssues` reads `Refs`. The re-measurement and calibration deltas the Done-when asks for are recorded in the Log.

7. **Plan revisions:** none. The plan matches the code.

```findings
findings:
  - id: new
    severity: Minor
    family: doc-claim-matches-code
    title: |
      Help text and Scope doc say merges are never scoped boundaries; only BranchPoint..HEAD merges are dropped
    detail: |
      helptext/active-time.md "With --branch-point, merge commits are not boundaries" and commit.go Scope doc "a merge is never a scoped boundary". A main-side merge outside the branch range whose lead names the issue is still kept. Reword both to "the branch's own merges" (as the atlas does), or drop merges by parent count for every window commit. Same family: the atlas verb list omits `closes`, which leadRE accepts.
```
