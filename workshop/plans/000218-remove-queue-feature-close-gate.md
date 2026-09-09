---
gate: boundary-review
issue: 218
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-09-09T12:52:48-07:00"
      agent: sdlc
      findings:
        - id: BR-1
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
            (carried from plan-quality PQ-5, deferred to the boundary review)
          family: verification-not-executable
          round: 1
        - id: BR-2
          severity: Minor
          title: Re-anchor table is hand-typed rather than derived, and drops trunkfile_test.go:721
          detail: |-
            2nd in family, so the deliverable is the rule: the re-anchor enumeration must be derived from a runnable sweep whose entire residue is dispositioned in the table, not hand-listed as line numbers. Prevalence: four surviving comment sites carry the dead concept (trunkfile.go:452, trunkfile_test.go:15, :83, :721); the table lists three and omits :721, the identical "fine for a queue" sentence as :452, which round 1's PQ-2 had explicitly named. I ran the derived sweep — of seventeen surviving files matching case-insensitive "queue" outside deleted paths, history and 218/219, the eight not already in the table (layergraph/walk.go, walk_test.go, weave/internal/walk/walk_test.go, projectstatus.go, fleetpolicy_test.go, judge/architecture.md:170, continuation.md:87, introspect/SKILL.md:41, pensive/claude-ecosystem-concepts.md:19) are all benign generic-queue senses, so adding :721 completes it.
            (carried from plan-quality PQ-6, deferred to the boundary review)
          family: stale-anchor-sweep
          round: 1
      boundary: '*'
      no_cap: true
      blocked: false
    - "n": 2
      timestamp: "2026-09-09T12:52:48-07:00"
      agent: claude
      findings:
        - id: BR-3
          severity: Important
          title: Done-when identifier sweep is inert — returns 99 lines, not the claimed one
          detail: |-
            workshop/issues/000218-remove-queue-feature.md:169-183. `grep -rn ... .` emits
            ./-prefixed paths, so all three `grep -v "^workshop/..."` / `"^construct/..."`
            filters never match. Measured 99 lines (86 under ./workshop/history/) against a
            clause asserting "exactly one line". Anchoring the filters at `^\./` makes the
            command return exactly the one deliberate atlas line it names, so the removal is
            sound and only the recorded evidence is wrong. 3rd in family; PQ-5 round 4 stated
            this diagnosis verbatim and it shipped unfixed.
          family: verification-not-executable
          round: 2
        - id: BR-4
          severity: Minor
          title: Re-anchor table omits trunkfile_test.go:719, which the code did re-anchor
          detail: |-
            The Spec table at line 130 lists only `:15,:83`, but the diff correctly
            re-anchored the TestTrunkFile_PreservesFileMode comment at trunkfile_test.go:719
            (the same "fine for a queue" sentence as trunkfile.go:452). PQ-6 asked for this
            row. Code is right; the table under-claims and needs the row.
          family: stale-anchor-sweep
          round: 2
        - id: BR-5
          severity: Minor
          title: '"36 sites" matches no countable set in trunkfile_test.go'
          detail: |-
            workshop/issues/000218-remove-queue-feature.md:129 and the Estimate's
            cross-cutting-refactor rationale both cite 36. Measured on the pre-change file:
            34 `queue.md` occurrences on 34 lines; 42 total case-insensitive `queue` tokens.
            Recurrence of PQ-3's family after "37 tests" was dropped and "~20" was corrected
            to "36".
          family: unbacked-count-claim
          round: 2
        - id: BR-6
          severity: Minor
          title: repoguard.go's restored "exactly the lifecycle verbs" list omits `project close`
          detail: |-
            cmd/sdlc/repoguard.go:10-11 spells WorkflowVerbs as seven verbs, but
            `project close` is in the set (internal/processmanual/gatesig.go:112), calls
            guardSpineRepo (projectclose.go:36), and the drift test enumerates it — I ran
            TestGuardSpineRepo_BrainRefusesAllLifecycleVerbs and the `project_close` subtest
            runs and passes. The set-level claim "exactly" is true; the inline list of
            members is incomplete, in the very sentence this issue set out to make read true.
          family: prose-enumeration-drift
          round: 2
        - id: BR-7
          severity: Minor
          title: '"workshop/queue.md is gone from origin/main" is ticked but only true post-merge'
          detail: |-
            Line 154 and plan item 281. The file is still on origin/main (a722b52, confirmed
            against `git ls-remote`); the deletion is committed on the branch, which is a
            clean descendant of main with no competing commits on that path, so the merge
            will remove it. The clause is not satisfiable at this boundary — re-word to
            "deleted on the branch; verify post-merge."
          family: boundary-claim-premature
          round: 2
      blocked: true
    - "n": 3
      timestamp: "2026-09-09T13:04:13-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: Clause now uses git grep with principled exclusions and states the expected residue plus the raw-grep numbers.
          round: 3
        - id: BR-2
          disposition: addressed
          note: Table demoted to non-authoritative and PreservesFileMode row added; I verified all six rows against the diff.
          round: 3
        - id: BR-3
          disposition: not-addressed
          note: Tool fixed (grep -rn to git grep) but not the exclusion set; measured 5 lines at HEAD, not 1.
          round: 3
        - id: BR-4
          disposition: addressed
          note: The PreservesFileMode row covers it; the code re-anchor is present at trunkfile_test.go:718-721.
          round: 3
        - id: BR-5
          disposition: withdrawn
          note: Retracted — grep -c -i queue on the pre-change trunkfile_test.go is exactly 36, so the count was checkable.
          round: 3
        - id: BR-6
          disposition: addressed
          note: All 8 guardSpineRepo call sites match WorkflowVerbs exactly; all 8 subtests run and pass.
          round: 3
        - id: BR-7
          disposition: not-addressed
          note: Done-when clause rewritten, but the Plan checkbox at :297 that BR-7 also named still claims origin/main.
          round: 3
      findings:
        - id: BR-8
          severity: Minor
          title: atlas says "Its consumer is ariadne#207" while TrunkFile has zero production callers at HEAD
          detail: |-
            atlas/workflow/sdlc-binary.md:649. Verified: git grep 'TrunkFile' over non-test .go files
            outside internal/gitx/trunkfile.go is empty — ~660 lines of production code with no caller.
            2nd live instance in this family (with BR-7's Plan checkbox), so the deliverable is the rule:
            no artifact may state in present or perfect tense what becomes true only after a step outside
            this boundary (the merge, or a future issue's implementation); write the verifiable tense and
            name the step. Measured enumeration for 218 is three sites — the Done-when clause (fixed round
            3), the Plan checkbox at :297, and this atlas line. Sweep all three.
          family: boundary-claim-premature
          round: 3
        - id: BR-9
          severity: Minor
          title: repoguard.go:23 still says "not 7 new per-verb flags" while :9-11 now lists 8 guarded verbs
          detail: |-
            The round-2 fix corrected the verb list one sentence above and left the cardinality below, in
            the same comment block, in the commit whose stated purpose was making that block read true.
            The 7 was already stale before 218 — project close gained the guard in c5b2096 (#180 M4). 2nd
            in family, so the rule rather than the number: a comment must not hand-restate a derived set
            or its cardinality; either point at the derivation or pin the prose with a drift test. The
            comment already added the grep -rn 'guardSpineRepo(' pointer — delete the list and the count
            and keep only that.
          family: prose-enumeration-drift
          round: 3
      blocked: true
---

# Gate ledger — ariadne#218 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-09-09T12:52:48-07:00 (sdlc) — passed

### Raised

- **BR-1** [Minor] `verification-not-executable` grep sweep clause has ~350 benign residue hits, so it cannot be checked as literally written
  2nd in family, so the deliverable is the rule, not this instance: a Done-when clause
  must be a runnable command plus its expected output, and when the raw command has
  known-benign hits the clause carries the filter or the expected residue. Measured:
  758 raw hits, ~350 after excluding deleted paths — ~330 in workshop/history (#209
  archives), the rest #218's own files and unrelated BFS-variable senses in
  bootstrap.sh, construct/scripts/list-peers.sh, projectstatus.go:199,
  layergraph/walk.go:41, pkg/vocab/fleetpolicy_test.go:52. I checked every non-history
  hit; none is a queue-feature reference, so the enumeration itself is complete.
  (carried from plan-quality PQ-5, deferred to the boundary review)
- **BR-2** [Minor] `stale-anchor-sweep` Re-anchor table is hand-typed rather than derived, and drops trunkfile_test.go:721
  2nd in family, so the deliverable is the rule: the re-anchor enumeration must be derived from a runnable sweep whose entire residue is dispositioned in the table, not hand-listed as line numbers. Prevalence: four surviving comment sites carry the dead concept (trunkfile.go:452, trunkfile_test.go:15, :83, :721); the table lists three and omits :721, the identical "fine for a queue" sentence as :452, which round 1's PQ-2 had explicitly named. I ran the derived sweep — of seventeen surviving files matching case-insensitive "queue" outside deleted paths, history and 218/219, the eight not already in the table (layergraph/walk.go, walk_test.go, weave/internal/walk/walk_test.go, projectstatus.go, fleetpolicy_test.go, judge/architecture.md:170, continuation.md:87, introspect/SKILL.md:41, pensive/claude-ecosystem-concepts.md:19) are all benign generic-queue senses, so adding :721 completes it.
  (carried from plan-quality PQ-6, deferred to the boundary review)

## Round 2 — 2026-09-09T12:52:48-07:00 (claude) — BLOCKED

### Raised

- **BR-3** [Important] `verification-not-executable` Done-when identifier sweep is inert — returns 99 lines, not the claimed one
  workshop/issues/000218-remove-queue-feature.md:169-183. `grep -rn ... .` emits
  ./-prefixed paths, so all three `grep -v "^workshop/..."` / `"^construct/..."`
  filters never match. Measured 99 lines (86 under ./workshop/history/) against a
  clause asserting "exactly one line". Anchoring the filters at `^\./` makes the
  command return exactly the one deliberate atlas line it names, so the removal is
  sound and only the recorded evidence is wrong. 3rd in family; PQ-5 round 4 stated
  this diagnosis verbatim and it shipped unfixed.
- **BR-4** [Minor] `stale-anchor-sweep` Re-anchor table omits trunkfile_test.go:719, which the code did re-anchor
  The Spec table at line 130 lists only `:15,:83`, but the diff correctly
  re-anchored the TestTrunkFile_PreservesFileMode comment at trunkfile_test.go:719
  (the same "fine for a queue" sentence as trunkfile.go:452). PQ-6 asked for this
  row. Code is right; the table under-claims and needs the row.
- **BR-5** [Minor] `unbacked-count-claim` "36 sites" matches no countable set in trunkfile_test.go
  workshop/issues/000218-remove-queue-feature.md:129 and the Estimate's
  cross-cutting-refactor rationale both cite 36. Measured on the pre-change file:
  34 `queue.md` occurrences on 34 lines; 42 total case-insensitive `queue` tokens.
  Recurrence of PQ-3's family after "37 tests" was dropped and "~20" was corrected
  to "36".
- **BR-6** [Minor] `prose-enumeration-drift` repoguard.go's restored "exactly the lifecycle verbs" list omits `project close`
  cmd/sdlc/repoguard.go:10-11 spells WorkflowVerbs as seven verbs, but
  `project close` is in the set (internal/processmanual/gatesig.go:112), calls
  guardSpineRepo (projectclose.go:36), and the drift test enumerates it — I ran
  TestGuardSpineRepo_BrainRefusesAllLifecycleVerbs and the `project_close` subtest
  runs and passes. The set-level claim "exactly" is true; the inline list of
  members is incomplete, in the very sentence this issue set out to make read true.
- **BR-7** [Minor] `boundary-claim-premature` "workshop/queue.md is gone from origin/main" is ticked but only true post-merge
  Line 154 and plan item 281. The file is still on origin/main (a722b52, confirmed
  against `git ls-remote`); the deletion is committed on the branch, which is a
  clean descendant of main with no competing commits on that path, so the merge
  will remove it. The clause is not satisfiable at this boundary — re-word to
  "deleted on the branch; verify post-merge."

## Round 3 — 2026-09-09T13:04:13-07:00 (claude) — BLOCKED

### Disposed

- BR-1 — addressed — Clause now uses git grep with principled exclusions and states the expected residue plus the raw-grep numbers.
- BR-2 — addressed — Table demoted to non-authoritative and PreservesFileMode row added; I verified all six rows against the diff.
- BR-3 — not-addressed — Tool fixed (grep -rn to git grep) but not the exclusion set; measured 5 lines at HEAD, not 1.
- BR-4 — addressed — The PreservesFileMode row covers it; the code re-anchor is present at trunkfile_test.go:718-721.
- BR-5 — withdrawn — Retracted — grep -c -i queue on the pre-change trunkfile_test.go is exactly 36, so the count was checkable.
- BR-6 — addressed — All 8 guardSpineRepo call sites match WorkflowVerbs exactly; all 8 subtests run and pass.
- BR-7 — not-addressed — Done-when clause rewritten, but the Plan checkbox at :297 that BR-7 also named still claims origin/main.

### Raised

- **BR-8** [Minor] `boundary-claim-premature` atlas says "Its consumer is ariadne#207" while TrunkFile has zero production callers at HEAD
  atlas/workflow/sdlc-binary.md:649. Verified: git grep 'TrunkFile' over non-test .go files
  outside internal/gitx/trunkfile.go is empty — ~660 lines of production code with no caller.
  2nd live instance in this family (with BR-7's Plan checkbox), so the deliverable is the rule:
  no artifact may state in present or perfect tense what becomes true only after a step outside
  this boundary (the merge, or a future issue's implementation); write the verifiable tense and
  name the step. Measured enumeration for 218 is three sites — the Done-when clause (fixed round
  3), the Plan checkbox at :297, and this atlas line. Sweep all three.
- **BR-9** [Minor] `prose-enumeration-drift` repoguard.go:23 still says "not 7 new per-verb flags" while :9-11 now lists 8 guarded verbs
  The round-2 fix corrected the verb list one sentence above and left the cardinality below, in
  the same comment block, in the commit whose stated purpose was making that block read true.
  The 7 was already stale before 218 — project close gained the guard in c5b2096 (#180 M4). 2nd
  in family, so the rule rather than the number: a comment must not hand-restate a derived set
  or its cardinality; either point at the derivation or pin the prose with a drift test. The
  comment already added the grep -rn 'guardSpineRepo(' pointer — delete the list and the count
  and keep only that.

## Open findings

- **BR-3** [Important] `verification-not-executable` Done-when identifier sweep is inert — returns 99 lines, not the claimed one
- **BR-7** [Minor] `boundary-claim-premature` "workshop/queue.md is gone from origin/main" is ticked but only true post-merge
- **BR-8** [Minor] `boundary-claim-premature` atlas says "Its consumer is ariadne#207" while TrunkFile has zero production callers at HEAD
- **BR-9** [Minor] `prose-enumeration-drift` repoguard.go:23 still says "not 7 new per-verb flags" while :9-11 now lists 8 guarded verbs
