---
gate: plan-quality
issue: 218
id_prefix: PQ
rounds:
    - "n": 1
      timestamp: "2026-09-09T11:42:04-07:00"
      agent: claude
      findings:
        - id: PQ-1
          severity: Important
          title: Done-when's only mechanical check exceeds go test's default timeout and dies before reporting
          detail: |-
            "go test ./cmd/sdlc/... green except the pre-existing ariadne#210 failure" cannot
            produce that result: the cmd/sdlc package blows the default 600s per-package timeout
            (reproduced: FAIL ... 602.502s with a goroutine dump, one goroutine in "syscall, 9
            minutes"), and a -timeout 1800s rerun was still running at 14 minutes. There is no
            Makefile target or CI workflow, so this literal command is the contract. The #210
            failure itself is real and exactly as described (fleet_plan_test.go:14, confirmed by a
            targeted -run). State the required -timeout, and say whether the long tail is expected
            slowness or a hang, or the close gate has no producible evidence.
          family: verification-not-executable
          round: 1
        - id: PQ-2
          severity: Important
          title: Plan re-anchors one "built for the queue" reference and freezes or omits its siblings
          detail: |-
            ARCH-PURPOSE, instance vs class. The plan correctly re-anchors
            atlas/workflow/sdlc-binary.md:688 but never enumerates the class it belongs to:
            gitx/trunkfile.go:452 ("fine for a queue, wrong for a general primitive that ariadne#207
            will point at"), trunkfile_test.go:15/:83/:721 plus the queue.md fixture path,
            workshop/lessons.md:6-24 whose evidence cites fakeTrunk and two TestQueueEdit_* tests
            deleted here, and workshop/issues/000207-sync-without-worktree.md:175, an OPEN issue
            reasoning from "#209's queue passes an intent-replaying transform" - an undeclared
            cross-issue interaction with the very issue the plan re-anchors onto. Done-when's
            "TrunkFile and its tests are UNCHANGED" forbids fixing the first two, so this needs a
            decision, not a discovery. Write the enumeration and dispose of it in one round.
          family: stale-anchor-sweep
          round: 1
        - id: PQ-3
          severity: Minor
          title: '"its 37 tests" matches no countable set in gitx'
          detail: |-
            cmd/sdlc/internal/gitx/trunkfile_test.go has 27 top-level Test* functions (plus one
            t.Run); the whole gitx package has 45. A Done-when clause pinned to 37 cannot be checked
            mechanically - drop the number or correct it.
          family: unbacked-count-claim
          round: 1
        - id: PQ-4
          severity: Minor
          title: No statement of whether any deleted test covered surviving behavior
          detail: |-
            A deletion's substitute for a test-strategy line is the coverage-loss claim, and a green
            suite cannot supply it because a deleted test never fails. I ran the check and it comes
            out clean: TestQueueCmd_WriteVerbsAreSpineGuarded and subcommandGuardSource
            (queue_test.go:319, :346) are the tree's only source-inspecting guard-ordering tests but
            covered queue verbs alone, while the surviving spine verbs are covered by
            repoguard_test.go:40 off processmanual.WorkflowVerbs(). Record that sentence rather than
            leaving it unasked.
          family: deleted-coverage-check
          round: 1
      blocked: true
    - "n": 2
      timestamp: "2026-09-09T12:12:46-07:00"
      agent: claude
      dispose:
        - id: PQ-1
          disposition: not-addressed
          note: |-
            Diagnosed: close_test.go:131 blocks in repolock.Acquire (repolock.go:210) on the
            real .git/sdlc.lock, held by the invoking `sdlc change-code --issue 218`. Not a
            child-process hang. DefaultWaitTimeout is 30m (repolock.go:19), so `-timeout 1800s`
            expires simultaneously and can never report green from inside a transaction.
          round: 2
        - id: PQ-2
          disposition: addressed
          note: |-
            Enumeration table written; all six sites verified present, and "unchanged" correctly
            narrowed to "behavior and coverage unchanged" with the re-anchor rule stated.
          round: 2
        - id: PQ-3
          disposition: addressed
          note: Count dropped from Done-when, with the reason recorded inline.
          round: 2
        - id: PQ-4
          disposition: addressed
          note: |-
            Coverage-loss claim asserted and correct: queue_test.go:319/:346 cover queue subcommands
            only; surviving spine verbs covered by repoguard_test.go off processmanual.WorkflowVerbs().
          round: 2
      blocked: true
    - "n": 3
      timestamp: "2026-09-09T12:18:05-07:00"
      agent: claude
      dispose:
        - id: PQ-1
          disposition: addressed
          note: Diagnosis reproduced live — the lock is held by this gate's own `sdlc change-code --issue 218`; replacement evidence runs green from inside it (build 0.54s, ./cmd/sdlc/internal/... 19.3s all ok).
          round: 3
      findings:
        - id: PQ-5
          severity: Minor
          title: grep sweep clause has ~350 benign residue hits, so it cannot be checked as literally written
          detail: |-
            2nd in family, so the deliverable is the rule, not this instance: a Done-when clause
            must be a runnable command plus its expected output, and when the raw command has
            known-benign hits the clause carries the filter or the expected residue. Measured:
            758 raw hits, ~350 after excluding deleted paths — ~330 in workshop/history (#209
            archives), the rest #218's own files and unrelated BFS-variable senses in
            bootstrap.sh, construct/scripts/list-peers.sh, projectstatus.go:199,
            layergraph/walk.go:41, pkg/vocab/fleetpolicy_test.go:52. I checked every non-history
            hit; none is a queue-feature reference, so the enumeration itself is complete.
          family: verification-not-executable
          round: 3
      blocked: false
    - "n": 4
      timestamp: "2026-09-09T12:22:29-07:00"
      agent: claude
      dispose:
        - id: PQ-5
          disposition: not-addressed
          note: 'Rule adopted, but the filter is inert: grep -rn over "." emits ./-prefixed paths while the three grep -v patterns anchor at ^workshop/ — measured 127 raw hits, 127 filtered. Needs ^\./ and an exclusion for the gate ledger under workshop/plans/000218-.'
          round: 4
      findings:
        - id: PQ-6
          severity: Minor
          title: Re-anchor table is hand-typed rather than derived, and drops trunkfile_test.go:721
          detail: '2nd in family, so the deliverable is the rule: the re-anchor enumeration must be derived from a runnable sweep whose entire residue is dispositioned in the table, not hand-listed as line numbers. Prevalence: four surviving comment sites carry the dead concept (trunkfile.go:452, trunkfile_test.go:15, :83, :721); the table lists three and omits :721, the identical "fine for a queue" sentence as :452, which round 1''s PQ-2 had explicitly named. I ran the derived sweep — of seventeen surviving files matching case-insensitive "queue" outside deleted paths, history and 218/219, the eight not already in the table (layergraph/walk.go, walk_test.go, weave/internal/walk/walk_test.go, projectstatus.go, fleetpolicy_test.go, judge/architecture.md:170, continuation.md:87, introspect/SKILL.md:41, pensive/claude-ecosystem-concepts.md:19) are all benign generic-queue senses, so adding :721 completes it.'
          family: stale-anchor-sweep
          round: 4
      blocked: false
content_hash: c8722968650ec7f0996d33e3ab2011d010532ac3db7742f7cfa00040ff623647
---

# Gate ledger — ariadne#218 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-09-09T11:42:04-07:00 (claude) — BLOCKED

### Raised

- **PQ-1** [Important] `verification-not-executable` Done-when's only mechanical check exceeds go test's default timeout and dies before reporting
  "go test ./cmd/sdlc/... green except the pre-existing ariadne#210 failure" cannot
  produce that result: the cmd/sdlc package blows the default 600s per-package timeout
  (reproduced: FAIL ... 602.502s with a goroutine dump, one goroutine in "syscall, 9
  minutes"), and a -timeout 1800s rerun was still running at 14 minutes. There is no
  Makefile target or CI workflow, so this literal command is the contract. The #210
  failure itself is real and exactly as described (fleet_plan_test.go:14, confirmed by a
  targeted -run). State the required -timeout, and say whether the long tail is expected
  slowness or a hang, or the close gate has no producible evidence.
- **PQ-2** [Important] `stale-anchor-sweep` Plan re-anchors one "built for the queue" reference and freezes or omits its siblings
  ARCH-PURPOSE, instance vs class. The plan correctly re-anchors
  atlas/workflow/sdlc-binary.md:688 but never enumerates the class it belongs to:
  gitx/trunkfile.go:452 ("fine for a queue, wrong for a general primitive that ariadne#207
  will point at"), trunkfile_test.go:15/:83/:721 plus the queue.md fixture path,
  workshop/lessons.md:6-24 whose evidence cites fakeTrunk and two TestQueueEdit_* tests
  deleted here, and workshop/issues/000207-sync-without-worktree.md:175, an OPEN issue
  reasoning from "#209's queue passes an intent-replaying transform" - an undeclared
  cross-issue interaction with the very issue the plan re-anchors onto. Done-when's
  "TrunkFile and its tests are UNCHANGED" forbids fixing the first two, so this needs a
  decision, not a discovery. Write the enumeration and dispose of it in one round.
- **PQ-3** [Minor] `unbacked-count-claim` "its 37 tests" matches no countable set in gitx
  cmd/sdlc/internal/gitx/trunkfile_test.go has 27 top-level Test* functions (plus one
  t.Run); the whole gitx package has 45. A Done-when clause pinned to 37 cannot be checked
  mechanically - drop the number or correct it.
- **PQ-4** [Minor] `deleted-coverage-check` No statement of whether any deleted test covered surviving behavior
  A deletion's substitute for a test-strategy line is the coverage-loss claim, and a green
  suite cannot supply it because a deleted test never fails. I ran the check and it comes
  out clean: TestQueueCmd_WriteVerbsAreSpineGuarded and subcommandGuardSource
  (queue_test.go:319, :346) are the tree's only source-inspecting guard-ordering tests but
  covered queue verbs alone, while the surviving spine verbs are covered by
  repoguard_test.go:40 off processmanual.WorkflowVerbs(). Record that sentence rather than
  leaving it unasked.

## Round 2 — 2026-09-09T12:12:46-07:00 (claude) — BLOCKED

### Disposed

- PQ-1 — not-addressed — Diagnosed: close_test.go:131 blocks in repolock.Acquire (repolock.go:210) on the
real .git/sdlc.lock, held by the invoking `sdlc change-code --issue 218`. Not a
child-process hang. DefaultWaitTimeout is 30m (repolock.go:19), so `-timeout 1800s`
expires simultaneously and can never report green from inside a transaction.
- PQ-2 — addressed — Enumeration table written; all six sites verified present, and "unchanged" correctly
narrowed to "behavior and coverage unchanged" with the re-anchor rule stated.
- PQ-3 — addressed — Count dropped from Done-when, with the reason recorded inline.
- PQ-4 — addressed — Coverage-loss claim asserted and correct: queue_test.go:319/:346 cover queue subcommands
only; surviving spine verbs covered by repoguard_test.go off processmanual.WorkflowVerbs().

## Round 3 — 2026-09-09T12:18:05-07:00 (claude) — passed

### Disposed

- PQ-1 — addressed — Diagnosis reproduced live — the lock is held by this gate's own `sdlc change-code --issue 218`; replacement evidence runs green from inside it (build 0.54s, ./cmd/sdlc/internal/... 19.3s all ok).

### Raised

- **PQ-5** [Minor] `verification-not-executable` grep sweep clause has ~350 benign residue hits, so it cannot be checked as literally written
  2nd in family, so the deliverable is the rule, not this instance: a Done-when clause
  must be a runnable command plus its expected output, and when the raw command has
  known-benign hits the clause carries the filter or the expected residue. Measured:
  758 raw hits, ~350 after excluding deleted paths — ~330 in workshop/history (#209
  archives), the rest #218's own files and unrelated BFS-variable senses in
  bootstrap.sh, construct/scripts/list-peers.sh, projectstatus.go:199,
  layergraph/walk.go:41, pkg/vocab/fleetpolicy_test.go:52. I checked every non-history
  hit; none is a queue-feature reference, so the enumeration itself is complete.

## Round 4 — 2026-09-09T12:22:29-07:00 (claude) — passed

### Disposed

- PQ-5 — not-addressed — Rule adopted, but the filter is inert: grep -rn over "." emits ./-prefixed paths while the three grep -v patterns anchor at ^workshop/ — measured 127 raw hits, 127 filtered. Needs ^\./ and an exclusion for the gate ledger under workshop/plans/000218-.

### Raised

- **PQ-6** [Minor] `stale-anchor-sweep` Re-anchor table is hand-typed rather than derived, and drops trunkfile_test.go:721
  2nd in family, so the deliverable is the rule: the re-anchor enumeration must be derived from a runnable sweep whose entire residue is dispositioned in the table, not hand-listed as line numbers. Prevalence: four surviving comment sites carry the dead concept (trunkfile.go:452, trunkfile_test.go:15, :83, :721); the table lists three and omits :721, the identical "fine for a queue" sentence as :452, which round 1's PQ-2 had explicitly named. I ran the derived sweep — of seventeen surviving files matching case-insensitive "queue" outside deleted paths, history and 218/219, the eight not already in the table (layergraph/walk.go, walk_test.go, weave/internal/walk/walk_test.go, projectstatus.go, fleetpolicy_test.go, judge/architecture.md:170, continuation.md:87, introspect/SKILL.md:41, pensive/claude-ecosystem-concepts.md:19) are all benign generic-queue senses, so adding :721 completes it.

## Open findings

- **PQ-5** [Minor] `verification-not-executable` grep sweep clause has ~350 benign residue hits, so it cannot be checked as literally written
- **PQ-6** [Minor] `stale-anchor-sweep` Re-anchor table is hand-typed rather than derived, and drops trunkfile_test.go:721
