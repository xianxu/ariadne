---
gate: boundary-review
issue: 207
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-09-11T11:04:04-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Critical
          title: issue new prints the pre-reallocation path on stdout; that file was deleted by finish()
          detail: |-
            issue.go:361 prints `shown`, computed before the sync; syncIssuesToMain returns no
            final path. Reproduced end-to-end: stdout workshop/issues/000002-mine.md, which is
            gone; only 000003-mine.md exists. Surface the final path and print it.
          family: stale-machine-output
          round: 1
        - id: BR-2
          severity: Critical
          title: 'per-attempt rc reset forgets earlier candidates: orphan duplicate survives, retry re-reallocates'
          detail: |-
            synctrunk.go:123 drops rc.written, so finish() cannot remove rejected attempts'
            candidates, and unionIDSpace counts them. Reproduced: 700->701->702 with 701 free,
            and an untracked full-duplicate 000701-mine.md left for a bare `claim` to publish.
          family: realloc-ordering-contract
          round: 1
        - id: BR-3
          severity: Important
          title: two re-allocations in one pass mispair rc; finish() deletes a just-published file
          detail: |-
            BR-3(a) only half-fixed. Reproduced: 000701-a.md published then removed, 000700-b.md
            (colliding) kept. rc is a single pair; the requested two-file test was not written.
          family: realloc-ordering-contract
          round: 1
        - id: BR-4
          severity: Important
          title: Spec step 5 exhaustion message and the unrelated-retry Done-when are undelivered/untested
          detail: |-
            updatemany.go:106 emits a generic contention error and rc/Foreign are dropped on
            error. No test covers "unrelated-path retry succeeds without re-allocating"; C2
            shows it currently re-allocates.
          family: donewhen-not-delivered
          round: 1
        - id: BR-5
          severity: Important
          title: concurrent same-issue trunk edits (incl. stale double-claim) are now silently overwritten
          detail: |-
            The merge-base check detected these; the Spec deleted it on a wrong premise and
            claim.md's CONFLICT BEHAVIOR dropped the promise without stating last-writer-wins.
            Document the loss and file a lost-update follow-up.
          family: deleted-guard-unreplaced
          round: 1
        - id: BR-6
          severity: Important
          title: unionIDSpace silently swallows a failed local scan on a false CAS justification
          detail: |-
            synctrunk.go:204. The CAS cannot see local-only ids, so it will not reject. It is a
            second union with a different failure policy from allocateIssueID, and nextFreeID is
            a second allocation rule beside issue.NextID (ARCH-DRY).
          family: free-id-space-incomplete
          round: 1
        - id: BR-7
          severity: Important
          title: atlas/workflow/issue-sync.md still documents the deleted worktree route (BR-7 claimed addressed)
          detail: |-
            issue-sync.md:20-42; also changecode.go:257 and sdlc-binary.md:69 describe copying
            into the main worktree, and sdlc-binary.md:94 has a spliced sentence.
          family: stale-user-facing-docs
          round: 1
        - id: BR-8
          severity: Important
          title: UpdateMany/commitSetAndPush duplicate Update/commitAndPush; Update has no production consumer
          detail: |-
            Retry-policy tests (declined hook, exhaustion, offline, mode) cover only the dead
            path. Re-express Update as an adapter over UpdateMany.
          family: parallel-implementation
          round: 1
        - id: BR-9
          severity: Important
          title: a git-quoted path reads as not-exist, becomes a Delete, and the arm prints synced
          detail: |-
            Reproduced with 000300-café.md: nothing published, `synced` emitted. Classify deletion
            from git status (-z, --name-status) and error on an unexplained not-exist read.
          family: failed-read-as-benign-state
          round: 1
        - id: BR-10
          severity: Minor
          title: FailedPublishKeepsTheOriginalFile stays green with the BR-2 fix reverted
          detail: |-
            No collision, so rc is nil. Only the real-git test pins BR-2, which contradicts the
            Log's "by fake and by real git".
          family: tautological-test
          round: 1
        - id: BR-11
          severity: Minor
          title: TrunkView.Ref returns the mutable tracking-ref name rather than the resolved base SHA
          family: snapshot-not-pinned
          round: 1
        - id: BR-12
          severity: Minor
          title: refIDSpace ls-tree is cwd-relative, so the collision guard is blind from a subdirectory
          detail: |-
            Pre-existing, and the resolveIDDirs comment claims otherwise. Fix with --full-tree or
            by reading through TrunkView in the repo root.
          family: cwd-relative-git-read
          round: 1
        - id: BR-13
          severity: Minor
          title: issue new fallback after a failed realloc publish commits the old id and gives refusing advice
          family: realloc-ordering-contract
          round: 1
        - id: BR-14
          severity: Minor
          title: idFrontmatterRE matches any line-start id:, not only the frontmatter block
          family: identity-rewrite-scope
          round: 1
        - id: BR-15
          severity: Minor
          title: decideCollision ignores two files in the same publish set sharing one id
          family: free-id-space-incomplete
          round: 1
      blocked: true
    - "n": 2
      timestamp: "2026-09-11T11:46:11-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: Reverting the stdout path fix fails TestRunIssueNew_PrintsThePathItActuallyPublished.
          round: 2
        - id: BR-2
          disposition: not-addressed
          note: 'Resetting res.candidates per attempt leaves every test green (no test runs finish() after a two-attempt re-allocation), and the retry still re-allocates over its own candidate: reproduced 700 to 702 with 701 free, because unionIDSpace counts it.'
          round: 2
        - id: BR-3
          disposition: addressed
          note: Collapsing reallocs to a single record fails TestSyncViaTrunk_TwoReallocationsInOnePublish.
          round: 2
        - id: BR-4
          disposition: addressed
          note: The exhaustion message is pinned and the unrelated-retry test exists; accumulating collisions across attempts is unpinned (resetting them per attempt leaves the suite green).
          round: 2
        - id: BR-5
          disposition: addressed
          note: 'claim.md and issue-sync.md state last-writer-wins and #222 exists; neither doc names #222, and the Spec premise is uncorrected.'
          round: 2
        - id: BR-6
          disposition: addressed
          note: The swallowed local scan is pinned by RefusesWhenTheLocalIDScanFails; nextFreeID (first free id above ours) still diverges from NextID (highest + 1), with low reachability.
          round: 2
        - id: BR-7
          disposition: not-addressed
          note: changecode.go:257-260 and the broken sentence at sdlc-binary.md:94-97 are unchanged; unswept siblings at issue.go:270-273, issue.go:354-357, sdlc-binary.md:104-108 and issue_test.go:202-204.
          round: 2
        - id: BR-8
          disposition: addressed
          note: Update is an adapter over UpdateMany, commitAndPush is gone, and trunkfile_test.go is unchanged and green.
          round: 2
        - id: BR-9
          disposition: addressed
          note: -z on all three queries is pinned by real-git and stub tests; the not-on-disk-and-not-tracked guard (synctrunk.go:211-219) is never exercised.
          round: 2
        - id: BR-10
          disposition: addressed
          note: Calling finish() on the failure path fails both the fake and the real-git test.
          round: 2
        - id: BR-11
          disposition: not-addressed
          note: The code pins the base SHA, but pointing the view back at trackingRef() leaves every test green; no test moves the ref inside prepare.
          round: 2
        - id: BR-12
          disposition: addressed
          note: Removing --full-tree fails TestRefIDSpace_IsNotBlindFromASubdirectory; the same family recurs elsewhere as a new finding.
          round: 2
        - id: BR-13
          disposition: addressed
          note: Both the advice branch and the errIDTaken wrapping are caught when reverted.
          round: 2
        - id: BR-14
          disposition: not-addressed
          note: 'Removing the frontmatter scoping leaves every test green; add a frontmatter block without id: plus a prose id: line, which must refuse.'
          round: 2
        - id: BR-15
          disposition: addressed
          note: Disabling the same-id check fails TestSyncViaTrunk_RefusesTwoFilesClaimingOneID.
          round: 2
      findings:
        - id: BR-16
          severity: Important
          title: issue new, claim and issue sync publish nothing from a subdirectory and report ok
          detail: |-
            This is the 2nd finding in family cwd-relative-git-read. Reproduced with the HEAD
            binary from docs/sub on a feature branch against a bare origin: "[ok] No issue
            changes to sync.", exit 0, the file left untracked, and origin unchanged. The id is
            never reserved, which is the failure in #207's own Problem. Also reproduced at base
            on main, so both arms had it. execGitRunner.Git runs in the process cwd
            (runner.go:36). The changedIssueFiles pathspecs and the issueFilesForID glob resolve
            against the cwd, and ls-files returns cwd-relative paths that prepare then joins
            onto the repo root. issue.go:270-273 asserts the opposite.
            RULE: every git call on the issue-sync and id path runs with its working directory
            pinned to the repo top level. Build the gitRunner rooted there, the way
            NewTrunkFile refuses an empty dir, and join every path that reaches os.* onto that
            same root. A per-call --full-tree is an instance of the rule, not the rule.
            Measured prevalence: 2 sites patched one at a time (refIDSpace, pathTrackedAtHEAD);
            at least 5 unpatched on the same path (three changedIssueFiles queries at
            claim.go:348-351, the issueFilesForID glob, the syncPathspec add/commit). Pin it
            with a test that publishes with cwd set to a subdirectory.
          family: cwd-relative-git-read
          round: 2
      blocked: true
---

# Gate ledger — ariadne#207 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-09-11T11:04:04-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Critical] `stale-machine-output` issue new prints the pre-reallocation path on stdout; that file was deleted by finish()
  issue.go:361 prints `shown`, computed before the sync; syncIssuesToMain returns no
  final path. Reproduced end-to-end: stdout workshop/issues/000002-mine.md, which is
  gone; only 000003-mine.md exists. Surface the final path and print it.
- **BR-2** [Critical] `realloc-ordering-contract` per-attempt rc reset forgets earlier candidates: orphan duplicate survives, retry re-reallocates
  synctrunk.go:123 drops rc.written, so finish() cannot remove rejected attempts'
  candidates, and unionIDSpace counts them. Reproduced: 700->701->702 with 701 free,
  and an untracked full-duplicate 000701-mine.md left for a bare `claim` to publish.
- **BR-3** [Important] `realloc-ordering-contract` two re-allocations in one pass mispair rc; finish() deletes a just-published file
  BR-3(a) only half-fixed. Reproduced: 000701-a.md published then removed, 000700-b.md
  (colliding) kept. rc is a single pair; the requested two-file test was not written.
- **BR-4** [Important] `donewhen-not-delivered` Spec step 5 exhaustion message and the unrelated-retry Done-when are undelivered/untested
  updatemany.go:106 emits a generic contention error and rc/Foreign are dropped on
  error. No test covers "unrelated-path retry succeeds without re-allocating"; C2
  shows it currently re-allocates.
- **BR-5** [Important] `deleted-guard-unreplaced` concurrent same-issue trunk edits (incl. stale double-claim) are now silently overwritten
  The merge-base check detected these; the Spec deleted it on a wrong premise and
  claim.md's CONFLICT BEHAVIOR dropped the promise without stating last-writer-wins.
  Document the loss and file a lost-update follow-up.
- **BR-6** [Important] `free-id-space-incomplete` unionIDSpace silently swallows a failed local scan on a false CAS justification
  synctrunk.go:204. The CAS cannot see local-only ids, so it will not reject. It is a
  second union with a different failure policy from allocateIssueID, and nextFreeID is
  a second allocation rule beside issue.NextID (ARCH-DRY).
- **BR-7** [Important] `stale-user-facing-docs` atlas/workflow/issue-sync.md still documents the deleted worktree route (BR-7 claimed addressed)
  issue-sync.md:20-42; also changecode.go:257 and sdlc-binary.md:69 describe copying
  into the main worktree, and sdlc-binary.md:94 has a spliced sentence.
- **BR-8** [Important] `parallel-implementation` UpdateMany/commitSetAndPush duplicate Update/commitAndPush; Update has no production consumer
  Retry-policy tests (declined hook, exhaustion, offline, mode) cover only the dead
  path. Re-express Update as an adapter over UpdateMany.
- **BR-9** [Important] `failed-read-as-benign-state` a git-quoted path reads as not-exist, becomes a Delete, and the arm prints synced
  Reproduced with 000300-café.md: nothing published, `synced` emitted. Classify deletion
  from git status (-z, --name-status) and error on an unexplained not-exist read.
- **BR-10** [Minor] `tautological-test` FailedPublishKeepsTheOriginalFile stays green with the BR-2 fix reverted
  No collision, so rc is nil. Only the real-git test pins BR-2, which contradicts the
  Log's "by fake and by real git".
- **BR-11** [Minor] `snapshot-not-pinned` TrunkView.Ref returns the mutable tracking-ref name rather than the resolved base SHA
- **BR-12** [Minor] `cwd-relative-git-read` refIDSpace ls-tree is cwd-relative, so the collision guard is blind from a subdirectory
  Pre-existing, and the resolveIDDirs comment claims otherwise. Fix with --full-tree or
  by reading through TrunkView in the repo root.
- **BR-13** [Minor] `realloc-ordering-contract` issue new fallback after a failed realloc publish commits the old id and gives refusing advice
- **BR-14** [Minor] `identity-rewrite-scope` idFrontmatterRE matches any line-start id:, not only the frontmatter block
- **BR-15** [Minor] `free-id-space-incomplete` decideCollision ignores two files in the same publish set sharing one id

## Round 2 — 2026-09-11T11:46:11-07:00 (claude) — BLOCKED

### Disposed

- BR-1 — addressed — Reverting the stdout path fix fails TestRunIssueNew_PrintsThePathItActuallyPublished.
- BR-2 — not-addressed — Resetting res.candidates per attempt leaves every test green (no test runs finish() after a two-attempt re-allocation), and the retry still re-allocates over its own candidate: reproduced 700 to 702 with 701 free, because unionIDSpace counts it.
- BR-3 — addressed — Collapsing reallocs to a single record fails TestSyncViaTrunk_TwoReallocationsInOnePublish.
- BR-4 — addressed — The exhaustion message is pinned and the unrelated-retry test exists; accumulating collisions across attempts is unpinned (resetting them per attempt leaves the suite green).
- BR-5 — addressed — claim.md and issue-sync.md state last-writer-wins and #222 exists; neither doc names #222, and the Spec premise is uncorrected.
- BR-6 — addressed — The swallowed local scan is pinned by RefusesWhenTheLocalIDScanFails; nextFreeID (first free id above ours) still diverges from NextID (highest + 1), with low reachability.
- BR-7 — not-addressed — changecode.go:257-260 and the broken sentence at sdlc-binary.md:94-97 are unchanged; unswept siblings at issue.go:270-273, issue.go:354-357, sdlc-binary.md:104-108 and issue_test.go:202-204.
- BR-8 — addressed — Update is an adapter over UpdateMany, commitAndPush is gone, and trunkfile_test.go is unchanged and green.
- BR-9 — addressed — -z on all three queries is pinned by real-git and stub tests; the not-on-disk-and-not-tracked guard (synctrunk.go:211-219) is never exercised.
- BR-10 — addressed — Calling finish() on the failure path fails both the fake and the real-git test.
- BR-11 — not-addressed — The code pins the base SHA, but pointing the view back at trackingRef() leaves every test green; no test moves the ref inside prepare.
- BR-12 — addressed — Removing --full-tree fails TestRefIDSpace_IsNotBlindFromASubdirectory; the same family recurs elsewhere as a new finding.
- BR-13 — addressed — Both the advice branch and the errIDTaken wrapping are caught when reverted.
- BR-14 — not-addressed — Removing the frontmatter scoping leaves every test green; add a frontmatter block without id: plus a prose id: line, which must refuse.
- BR-15 — addressed — Disabling the same-id check fails TestSyncViaTrunk_RefusesTwoFilesClaimingOneID.

### Raised

- **BR-16** [Important] `cwd-relative-git-read` issue new, claim and issue sync publish nothing from a subdirectory and report ok
  This is the 2nd finding in family cwd-relative-git-read. Reproduced with the HEAD
  binary from docs/sub on a feature branch against a bare origin: "[ok] No issue
  changes to sync.", exit 0, the file left untracked, and origin unchanged. The id is
  never reserved, which is the failure in #207's own Problem. Also reproduced at base
  on main, so both arms had it. execGitRunner.Git runs in the process cwd
  (runner.go:36). The changedIssueFiles pathspecs and the issueFilesForID glob resolve
  against the cwd, and ls-files returns cwd-relative paths that prepare then joins
  onto the repo root. issue.go:270-273 asserts the opposite.
  RULE: every git call on the issue-sync and id path runs with its working directory
  pinned to the repo top level. Build the gitRunner rooted there, the way
  NewTrunkFile refuses an empty dir, and join every path that reaches os.* onto that
  same root. A per-call --full-tree is an instance of the rule, not the rule.
  Measured prevalence: 2 sites patched one at a time (refIDSpace, pathTrackedAtHEAD);
  at least 5 unpatched on the same path (three changedIssueFiles queries at
  claim.go:348-351, the issueFilesForID glob, the syncPathspec add/commit). Pin it
  with a test that publishes with cwd set to a subdirectory.

## Open findings

- **BR-2** [Critical] `realloc-ordering-contract` per-attempt rc reset forgets earlier candidates: orphan duplicate survives, retry re-reallocates
- **BR-7** [Important] `stale-user-facing-docs` atlas/workflow/issue-sync.md still documents the deleted worktree route (BR-7 claimed addressed)
- **BR-11** [Minor] `snapshot-not-pinned` TrunkView.Ref returns the mutable tracking-ref name rather than the resolved base SHA
- **BR-14** [Minor] `identity-rewrite-scope` idFrontmatterRE matches any line-start id:, not only the frontmatter block
- **BR-16** [Important] `cwd-relative-git-read` issue new, claim and issue sync publish nothing from a subdirectory and report ok
