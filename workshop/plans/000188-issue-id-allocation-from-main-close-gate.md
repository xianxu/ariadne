---
gate: boundary-review
issue: 188
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-09-09T21:17:06-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Critical
          title: claim / issue new push trunk commits with an EMPTY subject on the feature-branch arm
          detail: |-
            syncViaTrunk passes msg straight to UpdateMany -> `commit-tree -m ""`; the
            deleted arm defaulted it via syncMessage(msg, "issue-sync: update issues
            from branch '<b>'"). runClaim (claim.go:114) and runIssueNew (issue.go:339)
            both pass "". Reproduced against a real bare origin: %s and %B are both
            empty on the pushed commit. Breaks AGENTS.md 12 and `git log --grep "^#N"`
            on the tracker's own trunk; issue.go:337's comment now states the opposite
            of what happens.
          family: sync-arm-contract-drift
          round: 1
        - id: BR-2
          severity: Critical
          title: a failed publish removes the ORIGINAL local issue file, inverting the Spec's ordering
          detail: |-
            syncViaTrunk (synctrunk.go:49-56) calls rc.finish() whenever rc != nil,
            regardless of err, and finish() os.Remove's rc.OldPath. The Spec says
            "remove the old one only after the push succeeds" and the Done-when asks
            for a test injecting a failure between the local write and the push.
            Reproduced: with the push rejected, 000700-mine.md is deleted and only the
            unpublished 000701-mine.md survives while the trunk holds nothing. The
            fakePublisher.err seam that exists to inject this is at zero call sites,
            propped up by a bare `var _ = errors.Is`.
          family: realloc-ordering-contract
          round: 1
        - id: BR-3
          severity: Critical
          title: nextFreeID searches the trunk only, so re-allocation can mint the duplicate it prevents
          detail: |-
            allocateIssueID (issueids.go:33, #213) allocates from the UNION of trunk and
            local. nextFreeID (issuecollision.go:80) sees only the trunk map. Two
            reproduced instances of the class: (a) two local files at one id both
            re-allocate in one prepare pass and both pick 000701, landing two files at
            one id in ONE commit, with rc mispaired so finish() then deletes one of
            them AFTER publishing it; (b) a local unpushed 000701 is ignored, so
            re-allocation collides with it. Feed nextFreeID trunk union local union
            ids-issued-this-pass.
          family: free-id-space-incomplete
          round: 1
        - id: BR-4
          severity: Important
          title: rc is stale state carried across retry attempts and is never invalidated
          detail: |-
            A re-allocation recorded on attempt N survives into attempt N+1 even when
            that attempt decides verdictPublish for the original path (reachable when a
            peer withdraws its colliding file, exactly how this repo's own 000207
            duplicate was resolved). Reproduced: the final set publishes
            000700-mine.md, then finish() removes it locally and keeps an unpublished
            000701-mine.md. ARCH-ORDER: reset rc per prepare invocation.
          family: realloc-ordering-contract
          round: 1
        - id: BR-5
          severity: Important
          title: syncViaTrunk emits no success confirmation and drops the `synced` stdout marker
          detail: |-
            The deleted arm printed cok(stderr, "Issues synced to main and pushed to
            origin.") and Fprintln(stdout, "synced"). syncViaTrunk prints the file list
            and then nothing, so a successful publish and a whole-set no-op are
            indistinguishable to the operator. issue.go:335 still describes the
            "synced" marker as a contract.
          family: publish-success-signal
          round: 1
        - id: BR-6
          severity: Important
          title: sdlc claim --help still documents the deleted main-worktree route
          detail: |-
            cmd/sdlc/helptext/claim.md:32-53,68 lists the worktree hunt, pull --rebase,
            merge-base conflict detection, an entire CONFLICT BEHAVIOR section, exit-1
            causes "missing main worktree, dirty main, conflicts", and the
            "issue-sync: update issues from branch '<branch>'" default the code
            dropped. This is surface a reader types.
          family: stale-user-facing-docs
          round: 1
        - id: BR-7
          severity: Important
          title: atlas is not updated inside the review window
          detail: |-
            At b7f08ec, atlas/workflow/sdlc-binary.md:89,117,149,660 still documents
            syncViaMainWorktree; the syncViaTrunk rewrite exists only as an
            uncommitted working-tree edit, outside this range.
            atlas/workflow/issue-sync.md:22-42 still describes the worktree route and
            conflict detector and is untouched even in the working tree.
          family: stale-user-facing-docs
          round: 1
        - id: BR-8
          severity: Important
          title: three Done-when bullets have no test
          detail: |-
            (a) "publish while another worktree on main is dirty or mid-rebase" — only
            the absent-worktree half is covered; (b) "a failure injected between the
            local write and the push" — no test, and the seam is unused; (c) "the
            published blob round-trips byte-identical, with attributes applied" —
            asserted only via strings.Contains.
          family: donewhen-not-delivered
          round: 1
        - id: BR-9
          severity: Minor
          title: claim.go's header renames the bullet to syncViaTrunk but keeps the deleted route's six steps
          family: stale-user-facing-docs
          round: 1
        - id: BR-10
          severity: Minor
          title: '`_ = branch` at claim.go:159 is dead; branch is already used at line 142'
          family: dead-code
          round: 1
        - id: BR-11
          severity: Minor
          title: TrunkView.Read and TrunkFile.TrackingRef are exported at zero call sites
          family: dead-code
          round: 1
        - id: BR-12
          severity: Minor
          title: commitCount's `len(strings.Fields(...))*0 +` term is dead and costs an extra git call
          family: dead-code
          round: 1
        - id: BR-13
          severity: Minor
          title: TestSyncViaTrunk_CheckRunsOnEveryAttempt asserts a counter the fake increments unconditionally
          detail: |-
            pub.runs == 2 cannot fail unless syncViaTrunk errors;
            TestUpdateMany_PathsAreDerivedPerAttempt is the load-bearing one.
          family: tautological-test
          round: 1
        - id: BR-14
          severity: Minor
          title: go test ./cmd/sdlc/ is red on TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory
          detail: |-
            workshop/plans/000200-sdlc-fleet-thread-inventory-plan.md was archived at
            dfeba9c, which precedes the review base — pre-existing, not introduced
            here, but the package cannot be cited as green close evidence without
            naming it.
          family: preexisting-red-suite
          round: 1
      blocked: true
---

# Gate ledger — ariadne#188 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-09-09T21:17:06-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Critical] `sync-arm-contract-drift` claim / issue new push trunk commits with an EMPTY subject on the feature-branch arm
  syncViaTrunk passes msg straight to UpdateMany -> `commit-tree -m ""`; the
  deleted arm defaulted it via syncMessage(msg, "issue-sync: update issues
  from branch '<b>'"). runClaim (claim.go:114) and runIssueNew (issue.go:339)
  both pass "". Reproduced against a real bare origin: %s and %B are both
  empty on the pushed commit. Breaks AGENTS.md 12 and `git log --grep "^#N"`
  on the tracker's own trunk; issue.go:337's comment now states the opposite
  of what happens.
- **BR-2** [Critical] `realloc-ordering-contract` a failed publish removes the ORIGINAL local issue file, inverting the Spec's ordering
  syncViaTrunk (synctrunk.go:49-56) calls rc.finish() whenever rc != nil,
  regardless of err, and finish() os.Remove's rc.OldPath. The Spec says
  "remove the old one only after the push succeeds" and the Done-when asks
  for a test injecting a failure between the local write and the push.
  Reproduced: with the push rejected, 000700-mine.md is deleted and only the
  unpublished 000701-mine.md survives while the trunk holds nothing. The
  fakePublisher.err seam that exists to inject this is at zero call sites,
  propped up by a bare `var _ = errors.Is`.
- **BR-3** [Critical] `free-id-space-incomplete` nextFreeID searches the trunk only, so re-allocation can mint the duplicate it prevents
  allocateIssueID (issueids.go:33, #213) allocates from the UNION of trunk and
  local. nextFreeID (issuecollision.go:80) sees only the trunk map. Two
  reproduced instances of the class: (a) two local files at one id both
  re-allocate in one prepare pass and both pick 000701, landing two files at
  one id in ONE commit, with rc mispaired so finish() then deletes one of
  them AFTER publishing it; (b) a local unpushed 000701 is ignored, so
  re-allocation collides with it. Feed nextFreeID trunk union local union
  ids-issued-this-pass.
- **BR-4** [Important] `realloc-ordering-contract` rc is stale state carried across retry attempts and is never invalidated
  A re-allocation recorded on attempt N survives into attempt N+1 even when
  that attempt decides verdictPublish for the original path (reachable when a
  peer withdraws its colliding file, exactly how this repo's own 000207
  duplicate was resolved). Reproduced: the final set publishes
  000700-mine.md, then finish() removes it locally and keeps an unpublished
  000701-mine.md. ARCH-ORDER: reset rc per prepare invocation.
- **BR-5** [Important] `publish-success-signal` syncViaTrunk emits no success confirmation and drops the `synced` stdout marker
  The deleted arm printed cok(stderr, "Issues synced to main and pushed to
  origin.") and Fprintln(stdout, "synced"). syncViaTrunk prints the file list
  and then nothing, so a successful publish and a whole-set no-op are
  indistinguishable to the operator. issue.go:335 still describes the
  "synced" marker as a contract.
- **BR-6** [Important] `stale-user-facing-docs` sdlc claim --help still documents the deleted main-worktree route
  cmd/sdlc/helptext/claim.md:32-53,68 lists the worktree hunt, pull --rebase,
  merge-base conflict detection, an entire CONFLICT BEHAVIOR section, exit-1
  causes "missing main worktree, dirty main, conflicts", and the
  "issue-sync: update issues from branch '<branch>'" default the code
  dropped. This is surface a reader types.
- **BR-7** [Important] `stale-user-facing-docs` atlas is not updated inside the review window
  At b7f08ec, atlas/workflow/sdlc-binary.md:89,117,149,660 still documents
  syncViaMainWorktree; the syncViaTrunk rewrite exists only as an
  uncommitted working-tree edit, outside this range.
  atlas/workflow/issue-sync.md:22-42 still describes the worktree route and
  conflict detector and is untouched even in the working tree.
- **BR-8** [Important] `donewhen-not-delivered` three Done-when bullets have no test
  (a) "publish while another worktree on main is dirty or mid-rebase" — only
  the absent-worktree half is covered; (b) "a failure injected between the
  local write and the push" — no test, and the seam is unused; (c) "the
  published blob round-trips byte-identical, with attributes applied" —
  asserted only via strings.Contains.
- **BR-9** [Minor] `stale-user-facing-docs` claim.go's header renames the bullet to syncViaTrunk but keeps the deleted route's six steps
- **BR-10** [Minor] `dead-code` `_ = branch` at claim.go:159 is dead; branch is already used at line 142
- **BR-11** [Minor] `dead-code` TrunkView.Read and TrunkFile.TrackingRef are exported at zero call sites
- **BR-12** [Minor] `dead-code` commitCount's `len(strings.Fields(...))*0 +` term is dead and costs an extra git call
- **BR-13** [Minor] `tautological-test` TestSyncViaTrunk_CheckRunsOnEveryAttempt asserts a counter the fake increments unconditionally
  pub.runs == 2 cannot fail unless syncViaTrunk errors;
  TestUpdateMany_PathsAreDerivedPerAttempt is the load-bearing one.
- **BR-14** [Minor] `preexisting-red-suite` go test ./cmd/sdlc/ is red on TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory
  workshop/plans/000200-sdlc-fleet-thread-inventory-plan.md was archived at
  dfeba9c, which precedes the review base — pre-existing, not introduced
  here, but the package cannot be cited as green close evidence without
  naming it.

## Open findings

- **BR-1** [Critical] `sync-arm-contract-drift` claim / issue new push trunk commits with an EMPTY subject on the feature-branch arm
- **BR-2** [Critical] `realloc-ordering-contract` a failed publish removes the ORIGINAL local issue file, inverting the Spec's ordering
- **BR-3** [Critical] `free-id-space-incomplete` nextFreeID searches the trunk only, so re-allocation can mint the duplicate it prevents
- **BR-4** [Important] `realloc-ordering-contract` rc is stale state carried across retry attempts and is never invalidated
- **BR-5** [Important] `publish-success-signal` syncViaTrunk emits no success confirmation and drops the `synced` stdout marker
- **BR-6** [Important] `stale-user-facing-docs` sdlc claim --help still documents the deleted main-worktree route
- **BR-7** [Important] `stale-user-facing-docs` atlas is not updated inside the review window
- **BR-8** [Important] `donewhen-not-delivered` three Done-when bullets have no test
- **BR-9** [Minor] `stale-user-facing-docs` claim.go's header renames the bullet to syncViaTrunk but keeps the deleted route's six steps
- **BR-10** [Minor] `dead-code` `_ = branch` at claim.go:159 is dead; branch is already used at line 142
- **BR-11** [Minor] `dead-code` TrunkView.Read and TrunkFile.TrackingRef are exported at zero call sites
- **BR-12** [Minor] `dead-code` commitCount's `len(strings.Fields(...))*0 +` term is dead and costs an extra git call
- **BR-13** [Minor] `tautological-test` TestSyncViaTrunk_CheckRunsOnEveryAttempt asserts a counter the fake increments unconditionally
- **BR-14** [Minor] `preexisting-red-suite` go test ./cmd/sdlc/ is red on TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory
