---
gate: plan-quality
issue: 206
id_prefix: PQ
rounds:
    - "n": 1
      timestamp: "2026-09-02T11:45:19-07:00"
      agent: claude
      findings:
        - id: PQ-1
          severity: Important
          title: 'New sync commit subject anchors #N, violating the documented issue-sync exclusion in gitx'
          detail: |-
            cmd/sdlc/internal/gitx/window.go:131-136 documents that issue-sync commits are excluded from
            shipped-work detection "for free" because their subject carries no #N, so bookkeepingVerbs has
            no entry for them. The Spec's "#206: issue: sync spec/plan" anchors #206 and "issue" is not a
            bookkeeping verb, so IsShippedWorkSubject returns true — feeding detectDrift's close-off check
            (state.go:130), branchStartByIssue's window base (milestoneclose.go:325), and activetime's
            run claiming (internal/activetime/compute.go:144). Decide the intended effect, add the
            bookkeepingVerbs entry, and pin the subject with a window_test.go table row.
          family: commit-subject-anchor-contract
          round: 1
        - id: PQ-2
          severity: Important
          title: No-push default is in Done-when but has no Plan step, and the obvious field polarity breaks callers
          detail: |-
            Both arms push unconditionally (claim.go:181, claim.go:313); Plan step 2 threads only a
            commit-message parameter. A `Push bool` on claimFlags zero-values to false, silently disabling
            the reservation broadcast at issue.go:311, which constructs the flags as a struct literal.
            Specify the parameter as NoPush (zero value keeps today's push) and give it its own step.
          family: donewhen-without-plan-step
          round: 1
        - id: PQ-3
          severity: Important
          title: '"Local, cheap, publishes nothing" is false on the syncOnBranch arm'
          detail: |-
            syncOnBranch refuses on a dirty main worktree (claim.go:224-231), runs git pull --rebase origin
            main (claim.go:242) — a network op that fails offline — and commits into main, not the branch
            being designed on, leaving the source worktree dirty. Suppressing the push does not make the
            branch path local or durable-on-branch. State which branch receives the durable commit.
          family: dispatch-arm-property-unverified
          round: 1
        - id: PQ-4
          severity: Important
          title: Replacing commitUntrackedIssueFile turns a deliberate push warning into a fatal change-code stop
          detail: |-
            branchcreate.go:127-131 warns on push failure by explicit design; syncOnMain returns it as an
            error and changecode.go:170-172 die()s, so an offline change-code now hard-stops before branch
            creation. The helper also exists so the branch starts tracked, which the branch-arm replacement
            does not preserve. State the intended posture and the on-main assumption.
          family: error-posture-on-helper-swap
          round: 1
        - id: PQ-5
          severity: Important
          title: Both recorded exclusions rest on claims the code contradicts
          detail: |-
            push.go:99-106 refuses only on untracked files; push.go:109-111 auto-commits tracked dirty
            content, so the cited "refuses on a non-empty porcelain" is wrong. sdlc merge checks the branch
            worktree while its archive commit runs in the main worktree (merge.go:547-551), and assessDirty
            exempts tracker paths (merge.go:155-166) — the same defect as syncOnBranch, with no precheck.
            Correct the record so the enumeration does not close a real gap.
          family: enumeration-exclusion-rationale
          round: 1
        - id: PQ-6
          severity: Minor
          title: Done-when names the swept-index regression test only for syncOnMain, not syncOnBranch
          detail: |-
            The Revision pulled syncOnBranch into piece 1, but the test obligation was not extended.
            The branch arm needs a two-worktree fixture; say whether it gets its own regression test.
          family: class-fix-without-class-test
          round: 1
        - id: PQ-7
          severity: Minor
          title: Name the real-git fixture for the pathspec test; an argv-recording gitRunner fake cannot prove it
          detail: |-
            testfix.Repo (hermeticrepo_test.go:14-19) is the real-git fixture. A fake gitRunner would
            assert the argv and pass even if the pathspec semantics were wrong, which is the thing
            under test (ARCH-MOCK).
          family: fake-cannot-observe-behavior
          round: 1
      blocked: true
    - "n": 2
      timestamp: "2026-09-02T11:51:32-07:00"
      agent: claude
      dispose:
        - id: PQ-1
          disposition: addressed
          note: Spec adds issue-sync to bookkeepingVerbs, corrects the parenthetical, pins a window_test.go row; plan step 4.
          round: 2
        - id: PQ-2
          disposition: addressed
          note: Spelled NoPush with the issue.go:311 struct-literal rationale, and given its own plan step.
          round: 2
        - id: PQ-3
          disposition: addressed
          note: New section names syncInPlace as the no-push arm and states the commit lands on the invoking branch.
          round: 2
        - id: PQ-4
          disposition: addressed
          note: Posture paragraph states warn-and-continue on both failures and how the tracked-start property is preserved.
          round: 2
        - id: PQ-5
          disposition: addressed
          note: Revision corrects the record with verified file:line and pulls both archive commits into scope.
          round: 2
        - id: PQ-6
          disposition: addressed
          note: syncViaMainWorktree gets its own two-worktree regression test in Done-when and plan step 1.
          round: 2
        - id: PQ-7
          disposition: addressed
          note: Done-when preamble mandates testfix.Repo and rejects the argv-recording fake by name.
          round: 2
      blocked: false
content_hash: f6e50287bbf84d88014f98d5b66d72ce296c3bb3b794250d0b46b81f977a5e22
---

# Gate ledger — ariadne#206 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-09-02T11:45:19-07:00 (claude) — BLOCKED

### Raised

- **PQ-1** [Important] `commit-subject-anchor-contract` New sync commit subject anchors #N, violating the documented issue-sync exclusion in gitx
  cmd/sdlc/internal/gitx/window.go:131-136 documents that issue-sync commits are excluded from
  shipped-work detection "for free" because their subject carries no #N, so bookkeepingVerbs has
  no entry for them. The Spec's "#206: issue: sync spec/plan" anchors #206 and "issue" is not a
  bookkeeping verb, so IsShippedWorkSubject returns true — feeding detectDrift's close-off check
  (state.go:130), branchStartByIssue's window base (milestoneclose.go:325), and activetime's
  run claiming (internal/activetime/compute.go:144). Decide the intended effect, add the
  bookkeepingVerbs entry, and pin the subject with a window_test.go table row.
- **PQ-2** [Important] `donewhen-without-plan-step` No-push default is in Done-when but has no Plan step, and the obvious field polarity breaks callers
  Both arms push unconditionally (claim.go:181, claim.go:313); Plan step 2 threads only a
  commit-message parameter. A `Push bool` on claimFlags zero-values to false, silently disabling
  the reservation broadcast at issue.go:311, which constructs the flags as a struct literal.
  Specify the parameter as NoPush (zero value keeps today's push) and give it its own step.
- **PQ-3** [Important] `dispatch-arm-property-unverified` "Local, cheap, publishes nothing" is false on the syncOnBranch arm
  syncOnBranch refuses on a dirty main worktree (claim.go:224-231), runs git pull --rebase origin
  main (claim.go:242) — a network op that fails offline — and commits into main, not the branch
  being designed on, leaving the source worktree dirty. Suppressing the push does not make the
  branch path local or durable-on-branch. State which branch receives the durable commit.
- **PQ-4** [Important] `error-posture-on-helper-swap` Replacing commitUntrackedIssueFile turns a deliberate push warning into a fatal change-code stop
  branchcreate.go:127-131 warns on push failure by explicit design; syncOnMain returns it as an
  error and changecode.go:170-172 die()s, so an offline change-code now hard-stops before branch
  creation. The helper also exists so the branch starts tracked, which the branch-arm replacement
  does not preserve. State the intended posture and the on-main assumption.
- **PQ-5** [Important] `enumeration-exclusion-rationale` Both recorded exclusions rest on claims the code contradicts
  push.go:99-106 refuses only on untracked files; push.go:109-111 auto-commits tracked dirty
  content, so the cited "refuses on a non-empty porcelain" is wrong. sdlc merge checks the branch
  worktree while its archive commit runs in the main worktree (merge.go:547-551), and assessDirty
  exempts tracker paths (merge.go:155-166) — the same defect as syncOnBranch, with no precheck.
  Correct the record so the enumeration does not close a real gap.
- **PQ-6** [Minor] `class-fix-without-class-test` Done-when names the swept-index regression test only for syncOnMain, not syncOnBranch
  The Revision pulled syncOnBranch into piece 1, but the test obligation was not extended.
  The branch arm needs a two-worktree fixture; say whether it gets its own regression test.
- **PQ-7** [Minor] `fake-cannot-observe-behavior` Name the real-git fixture for the pathspec test; an argv-recording gitRunner fake cannot prove it
  testfix.Repo (hermeticrepo_test.go:14-19) is the real-git fixture. A fake gitRunner would
  assert the argv and pass even if the pathspec semantics were wrong, which is the thing
  under test (ARCH-MOCK).

## Round 2 — 2026-09-02T11:51:32-07:00 (claude) — passed

### Disposed

- PQ-1 — addressed — Spec adds issue-sync to bookkeepingVerbs, corrects the parenthetical, pins a window_test.go row; plan step 4.
- PQ-2 — addressed — Spelled NoPush with the issue.go:311 struct-literal rationale, and given its own plan step.
- PQ-3 — addressed — New section names syncInPlace as the no-push arm and states the commit lands on the invoking branch.
- PQ-4 — addressed — Posture paragraph states warn-and-continue on both failures and how the tracked-start property is preserved.
- PQ-5 — addressed — Revision corrects the record with verified file:line and pulls both archive commits into scope.
- PQ-6 — addressed — syncViaMainWorktree gets its own two-worktree regression test in Done-when and plan step 1.
- PQ-7 — addressed — Done-when preamble mandates testfix.Repo and rejects the argv-recording fake by name.

## Open findings

(none — every finding has been disposed)
