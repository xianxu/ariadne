---
gate: boundary-review
issue: 213
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-09-03T15:35:36-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Critical
          title: The CI check compares against merge-base, so it structurally cannot see the collision this issue exists to catch
          detail: |-
            merge-check.yml passes base = merge-base(base_tip, head). For a branch cut BEFORE
            the colliding id landed on main — the issue's own reproduction — the merge-base
            predates that file too, so baseByID has no entry for the id and introducedIDClashes
            finds nothing. Proven in a scratch repo: branch cut at S, main publishes
            000500-theirs.md, branch adds 000500-mine.md. `lint-ids --base <merge-base>` →
            "[ok] this range introduces no reused issue ids", exit 0. `lint-ids --base <main tip>`
            → refuses, exit 1, naming both paths. The enforcement layer passes the case it was
            built for. Fix: the check must resolve the trunk TIP itself (fetch + `git rev-parse
            origin/main`, falling back to the runner-supplied base) rather than using the range
            base, since the runner contract is fixed at merge-base.
          family: gate-compares-wrong-baseline
          round: 1
        - id: BR-2
          severity: Critical
          title: issueFilesByID keeps only the first path per id, so an introduced duplicate is detected or missed by slug sort order
          detail: |-
            issueFilesByID (cmd/sdlc/issueids.go:190) collapses a within-ref duplicate to the
            first path seen; both refuseDuplicateIssueIDs and introducedIDClashes then compare
            single paths. When the head tree contains BOTH files (a rebased/updated PR, or any
            branch that pulled main after the trunk file landed), head[id] equals base[id] and
            nothing is reported. Proven on the real repo: planting workshop/issues/000213-aaa-
            collision.md → refused, exit 1; planting 000213-planted-collision.md (identical in
            every way except the slug sorts after "nextid") → "[ok] this range introduces no
            reused issue ids", exit 0. Same hole in the merge gate — a scratch test adding
            000001-zzz-later-slug.md next to the trunk's 000001-first.md gets
            "[ok] duplicate-id gate: no reused issue ids". The Spec names this exact hazard
            ("issueFilesByID keeps the first path seen ... which silently collapses exactly the
            state being hunted") but only the within-ref REPORT was fixed, not the comparison.
            Fix: build map[int]map[string]bool (or reuse a shared pure helper alongside
            DuplicateIDsInRef) and flag any head path for an id that base already owns under a
            different path. Both existing tests pass only because their chosen slugs happen to
            sort first.
          family: gate-compares-wrong-baseline
          round: 1
        - id: BR-3
          severity: Important
          title: repoRelativeIDDirs rejects a not-yet-created id dir whenever the repo root is reached through a symlink, silently disabling every layer
          detail: |-
            cmd/sdlc/issueids.go:107 EvalSymlinks the repo top unconditionally but the candidate
            dir only when it exists. A repo under a symlinked path (macOS /tmp → /private/tmp, a
            symlinked workspace) with no workshop/history/ yet yields abs=/tmp/... vs
            top=/private/tmp/..., filepath.Rel returns "../../..", and the dir is refused as
            "outside the current repo". Observed: `sdlc issue lint-ids` printed "id lint skipped:
            workshop/history is outside the current repo" and exited 0 in a fresh fixture;
            creating workshop/history/ made the same command work. In allocateIssueID this path
            warns "origin/main unreachable" and falls back to the local-only scan — the original
            defect, re-armed, on any repo whose history dir does not exist yet. Fix: resolve the
            nearest existing ancestor of abs (or EvalSymlinks the cwd before Abs) so containment
            is decided on comparable paths.
          family: dir-containment-false-negative
          round: 1
        - id: BR-4
          severity: Important
          title: The merge gate at step 4.6 reads a stale origin/main — merge.go's own comment says the flow has not fetched yet
          detail: |-
            refuseDuplicateIssueIDs is called at cmd/sdlc/merge.go:340, but merge.go:464-465
            states "origin/main is stale here (the flow doesn't pull until AFTER deciding,
            below)" and only fetches at :469. A collision published to the trunk since this
            checkout last fetched is invisible to the gate that documents itself as "the last
            point where an id collision is still repairable". Fix: fetch before 4.6, or move the
            gate below the existing fetch at :469.
          family: gate-compares-wrong-baseline
          round: 1
        - id: BR-5
          severity: Important
          title: No test exercises the CI check's refusal path; the one that claims to asserts the SKIP path
          detail: |-
            TestMergeCheckScript_RefusesAPlantedCollision (cmd/sdlc/issueids_test.go:281) plants a
            collision, then runs the script in a fixture with no ./cmd/sdlc and asserts exit 0
            plus "skipping". Its name states the opposite of what it verifies, and it satisfies
            the Plan row "Test the CI check by running it against a real repo with a planted
            collision" with a test that never reaches the check. Nothing goes red if the refusal
            breaks — which is how both Critical findings above shipped with a green suite. The
            exit-1 path is also unreachable in-process because runIssueLintIDs calls
            exitWithCode → os.Exit (the `die` var seam exists for exactly this). Fix: route the
            refusal through a testable seam and add (a) a Go test asserting exit 1 on an
            introduced clash regardless of slug sort order, and (b) a script-level test against a
            repo where the build can actually run; rename or split the existing skip-path test.
          family: fix-not-pinned-by-a-failing-test
          round: 1
        - id: BR-6
          severity: Important
          title: The CI check does not reach derivatives — merge-checks.d is a scaffold row, and the script self-skips without ./cmd/sdlc
          detail: |-
            Done-when claims "it propagates to every derivative through the symlinked runner —
            parley.nvim carries four of the eight known collisions", and atlas/workflow/sdlc-
            binary.md's layer table repeats "propagates to derivatives". Neither holds:
            construct/base.manifest:130 is `scaffold scripts/merge-checks.d` (an empty directory
            per repo — only the runner is symlinked), so the check is an ariadne-local file that
            propagate-base will not carry; and even if copied, scripts/merge-checks.d/40-
            duplicate-issue-id.sh:33 exits 0 when ./cmd/sdlc is absent, which is true of every
            derivative by construction (verified: parley.nvim has no cmd/sdlc and an empty
            merge-checks.d). Fix: either resolve the sdlc module from the cloned upstream peer
            (CI already runs BOOTSTRAP_CLONE_ONLY=1 ./bootstrap.sh, so ../<upstream>/cmd/sdlc
            exists) plus a manifest row that actually delivers the check, or correct the
            Done-when and the atlas table to say ariadne-only.
          family: enforcement-does-not-propagate
          round: 1
        - id: BR-7
          severity: Important
          title: publishedIssueIDs swallows per-directory ls-tree failures, so a partial trunk read allocates a colliding id with no warning
          detail: |-
            cmd/sdlc/issueids.go:88 `continue`s on ls-tree error. A missing directory is already
            exit 0 with empty output (verified), so a non-nil error means a real git failure —
            and dropping that directory's ids silently under-counts the published space, which is
            precisely the silent-fallback failure the Spec forbids ("a silent fallback here
            recreates the bug it is meant to fix"). Fix: return the error so allocateIssueID
            takes its loud-warning path, or warn per directory.
          family: silent-degradation-in-allocator
          round: 1
        - id: BR-8
          severity: Minor
          title: Three near-identical ls-tree listing parsers and two identical rev-parse+ls-tree IO shells (ARCH-DRY)
          detail: |-
            issue.IDsInTreeListing, issue.DuplicateIDsInRef and issueFilesByID each re-implement
            split → trim → LastIndex("/") → IDFromFilename; publishedIssueIDs and idListing each
            re-implement rev-parse + repoRelativeIDDirs + per-dir ls-tree. The duplication is not
            cosmetic: the collapsing map in issueFilesByID exists only because it re-implements
            instead of reusing the id→paths structure DuplicateIDsInRef already builds. One pure
            PathsByIDInTreeListing in the issue package fixes the DRY violation and the
            order-dependence Critical together.
          family: duplicated-listing-parser
          round: 1
        - id: BR-9
          severity: Minor
          title: atlas/workflow/ci-merge-check.md renders the new prose inside a fenced code block
          detail: |-
            The paragraphs added at lines 31-45 sit between the opening ``` at line 30 and its
            closing fence at line 46, so they render as literal code.
          family: docs-lag-new-surface
          round: 1
        - id: BR-10
          severity: Minor
          title: '`issue lint-ids` is missing from the SUBCOMMANDS list in cmd/sdlc/helptext/issue.md'
          detail: |-
            The list at issue.md:6-14 enumerates new/sync/set-status/list/show; the new verb is
            not there. (`validate` is also absent — pre-existing.)
          family: docs-lag-new-surface
          round: 1
        - id: BR-11
          severity: Minor
          title: The duplicate-id gate is bundled behind --no-validate rather than its own --no-<gate> flag
          detail: |-
            AGENTS.md §5 asks for a per-gate flag so a bypass is an explicit acknowledgment of the
            specific gate. Today skipping #124's instance-conformance gate also silently skips the
            id gate, and merge.md's FLAGS block documents neither.
          family: gate-bypass-flag-granularity
          round: 1
        - id: BR-12
          severity: Minor
          title: '`git fetch origin main` on every `sdlc issue new` has no timeout (ARCH-CONSTRAINTS)'
          detail: |-
            allocateIssueID's fetch is best-effort but unbounded. A host that drops packets rather
            than refusing (VPN down, sleeping laptop) blocks an interactive verb for the TCP
            connect timeout. Consider a short --git-timeout / GIT_HTTP_LOW_SPEED_* bound, or an
            explicit --no-fetch escape, and state the budget in the issue.
          family: unbounded-external-call
          round: 1
      blocked: true
    - "n": 2
      timestamp: "2026-09-03T17:00:45-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: Revert-verified — reverting the script to base="$fallback_base" turns TestMergeCheckScript_RefusesGivenMergeBase red with "[ok] no reused ids".
          round: 2
        - id: BR-2
          disposition: addressed
          note: Revert-verified — first-path-only map turns the zzz-sorts-last subtest red while aaa stays green.
          round: 2
        - id: BR-3
          disposition: not-addressed
          note: 'Behavior change is real (measured both binaries under a symlinked root) but NO test pins it — reverting repoRelativeIDDirs to d50a023''s form leaves every #213 test green — and the class survives for absolute dirs: --issues-dir under a symlinked path with workshop/history (or history/issues) absent still prints "outside the current repo" and disables all three layers.'
          round: 2
        - id: BR-4
          disposition: addressed
          note: Fetch added at issueids.go:172, unconditional and reachable; no test pins the ordering (TestRefuseDuplicateIssueIDs' origin is already current), noted in coverage.
          round: 2
        - id: BR-5
          disposition: addressed
          note: The refusal path is now the tested one and it genuinely goes red when the BR-1 fix is reverted; the skip path is a separate, correctly named test.
          round: 2
        - id: BR-6
          disposition: addressed
          note: Verified by weave compile --dry-run in parley.nvim (emits the symlink row) plus a derivative-shaped fixture with no cmd/sdlc; atlas and Done-when corrected. Actual propagation to parley.nvim has not been run yet - operator action.
          round: 2
        - id: BR-7
          disposition: addressed
          note: Fixed and pinned at publishedIssueIDs, but only there - see the new finding for the two sibling sites that still swallow.
          round: 2
        - id: BR-8
          disposition: not-addressed
          note: Three listing parsers and three rev-parse+ls-tree shells remain; PathsByIDInTreeListing was never written, and the shells' error handling has now diverged.
          round: 2
        - id: BR-9
          disposition: not-addressed
          note: atlas/workflow/ci-merge-check.md still opens a fence at line 31 and closes it at 47, so lines 33-46 render as code; that prose is now also stale.
          round: 2
        - id: BR-10
          disposition: not-addressed
          note: cmd/sdlc/helptext/issue.md SUBCOMMANDS still lists new/sync/set-status/list/show only.
          round: 2
        - id: BR-11
          disposition: not-addressed
          note: Still gated on !f.NoValidate at merge.go:339, no --no-dupid, no FLAGS documentation, and no warning printed when the bypass fires.
          round: 2
        - id: BR-12
          disposition: not-addressed
          note: No timeout or --no-fetch escape; merge step 4.6 now adds a second unbounded fetch.
          round: 2
      findings:
        - id: BR-13
          severity: Critical
          title: Archiving or renaming an issue file is refused as a reused id by both the CI check and the merge gate
          detail: |-
            newPathsFor keys on exact path equality, so moving workshop/issues/NNNNNN-x.md to
            workshop/history/issues/NNNNNN-x.md - the archive step AGENTS.md section 1 mandates
            on done - reads as a second file claiming a live id. Replayed against this repo's
            real merged PR 109 (base 008f7e3^1, head 008f7e3^2): exit 1, "this range reuses 1
            issue id(s)", naming 000195 and telling the operator to renumber it. sdlc merge step
            4.6 shares the predicate and dies the same way. Fix: decide "introduced" from the
            range delta - exclude paths that are rename destinations in
            git diff --name-status -M base head over the id dirs - rather than from set-differencing
            two trees' path lists. Verified the corrected predicate still refuses BR-1's
            cut-then-publish shape and still only reports the eight already-merged collisions.
          family: gate-predicate-ignores-range-delta
          round: 2
        - id: BR-14
          severity: Important
          title: A within-ref duplicate the range introduces is labelled pre-existing and passes, contradicting Done-when
          detail: |-
            runIssueLintIDs runs DuplicateIDsInRef over the head listing and reports every result
            as pre-existing; introducedIDClashes cannot see them because base[id] is empty. A
            fixture whose branch adds 000500-agent-a.md and 000500-agent-b.md (neither on the
            trunk) prints "pre-existing duplicate id 000500" followed by "id lint: this range
            introduces no reused issue ids" and exits 0. Done-when states the scan "REFUSES when
            the PR introduces one". Same root cause and same fix as the Critical above: an id is
            introduced-within-ref when head holds two or more distinct paths for it and at least
            one is absent at base.
          family: gate-predicate-ignores-range-delta
          round: 2
        - id: BR-15
          severity: Important
          title: BR-7's class was fixed at one of three sites; issueFilesByID and idListing still swallow ls-tree failures
          detail: |-
            This is the 2nd finding in family silent-degradation-in-allocator. Do NOT fix these
            two sites in isolation. The rule: a partial read of the id space must be an error at
            EVERY site that performs one, and no code path may report a clean verdict from an
            incomplete listing. The enumeration is mechanical - three functions run per-directory
            ls-tree (publishedIssueIDs, issueFilesByID at issueids.go:220, idListing at
            issuelintids.go:143); round 1 fixed the first only. A dropped directory in
            issueFilesByID removes ids from the BASE map, which reads as "this id is new" - a
            false negative inside the enforcement gate. runIssueLintIDs compounds it by turning
            every error into return nil (exit 0) at three places. Fix the class with one shared
            IO shell that returns an error (which also discharges BR-8).
          family: silent-degradation-in-allocator
          round: 2
        - id: BR-16
          severity: Important
          title: git fetch origin main does not guarantee refs/remotes/origin/main, so CI can fall back to the merge-base baseline BR-1 proved blind
          detail: |-
            This is the 4th finding in family gate-compares-wrong-baseline. Earlier rounds fixed
            instances. The rule that covers all of them: a gate must resolve its authoritative
            baseline explicitly and fail loudly when it cannot, never silently substitute a
            baseline that is structurally blind to the class it exists to catch. Measured - with
            a narrow remote.origin.fetch, "git fetch origin main" leaves origin/main unresolvable
            while "git fetch origin +refs/heads/main:refs/remotes/origin/main" resolves it. When
            it fails, 40-duplicate-issue-id.sh falls back to $fallback_base, which is the
            merge-base BR-1 showed cannot see the collision. I did not prove this fires under the
            shim's fetch-depth 0; the explicit refspec removes the dependency on that assumption,
            and the fallback should be a loud degraded state, not a pass.
          family: gate-compares-wrong-baseline
          round: 2
        - id: BR-17
          severity: Minor
          title: atlas ci-merge-check.md still describes the pre-BR-6 skip conditions; three of five doc homes for this surface are wrong
          detail: |-
            This is the 3rd finding in family docs-lag-new-surface. Do NOT fix only this file.
            The rule: a change to a user-facing surface updates that surface's doc homes in the
            SAME commit - helptext, the atlas entry, and README where the verb is listed - and
            corrects prose describing the old behavior. Measured enumeration for 213: helptext
            issue.md SUBCOMMANDS missing lint-ids (BR-10); helptext merge.md FLAGS silent on the
            gate (BR-11); ci-merge-check.md lines 41-45 both trapped in a code fence (BR-9) and
            still claiming the script keys on ./cmd/sdlc, which is exactly what BR-6 changed;
            sdlc-binary.md correct; README not applicable (issue verbs are not listed there).
            Three of five homes wrong - sweep the list.
          family: docs-lag-new-surface
          round: 2
      blocked: true
    - "n": 3
      timestamp: "2026-09-03T17:49:32-07:00"
      agent: claude
      dispose:
        - id: BR-3
          disposition: addressed
          note: 'Verified live — fresh repo under /tmp (→/private/tmp) with no workshop/history: lint reported the planted duplicate instead of skipping.'
          round: 3
        - id: BR-13
          disposition: addressed
          note: The named instance is fixed (real PR 109 replay, base 008f7e3^1 / head 008f7e3^2 → exit 0), but the class is not swept — see the new Critical.
          round: 3
        - id: BR-14
          disposition: addressed
          note: Verified — a branch adding 000500-agent-a.md and 000500-agent-b.md now exits 1 rather than labelling them pre-existing and passing.
          round: 3
        - id: BR-8
          disposition: not-addressed
          note: Three listing parsers and three rev-parse+ls-tree IO shells still present, unchanged across all three rounds.
          round: 3
        - id: BR-9
          disposition: not-addressed
          note: atlas/workflow/ci-merge-check.md lines 30-46 unchanged; the prose still renders inside the code fence.
          round: 3
        - id: BR-10
          disposition: not-addressed
          note: helptext/issue.md SUBCOMMANDS still omits lint-ids, and now also the new --trunk flag added in 65aea14.
          round: 3
        - id: BR-11
          disposition: not-addressed
          note: merge.go:336 still gates on !f.NoValidate; merge.md FLAGS still documents neither gate.
          round: 3
        - id: BR-12
          disposition: not-addressed
          note: Still unbounded, and now at two call sites (issueids.go:83 and :246) rather than one.
          round: 3
        - id: BR-15
          disposition: not-addressed
          note: idListing (issuelintids.go:149) still does `continue`; issueFilesByID's fix is pinned by no test — reverting issueids.go:303 to `continue` leaves the suite green.
          round: 3
        - id: BR-16
          disposition: not-addressed
          note: The explicit refspec landed in Go only; scripts/merge-checks.d/40-duplicate-issue-id.sh:65 still runs `git fetch --quiet origin main`. Re-measured on a single-branch clone with a narrow refspec — plain form leaves origin/main unresolvable, explicit form resolves it. Sibling site — introducedIDClashes at issuelintids.go:120-124 silently substitutes baseByID when the trunk read fails.
          round: 3
        - id: BR-17
          disposition: not-addressed
          note: No doc home moved this round, and 65aea14 added a new one — the --trunk flag and the merge-result model appear in no helptext or atlas entry. sdlc-binary.md:187 still describes only a branch-vs-trunk comparison; ci-merge-check.md still describes the pre-BR-6 skip conditions and build location.
          round: 3
      findings:
        - id: BR-18
          severity: Critical
          title: mergedPathsFor honours deletions by head but not by the trunk, so every PR open across an issue archive on main is falsely refused
          detail: |-
            This is the 3rd finding in family gate-predicate-ignores-range-delta. Earlier rounds fixed
            instances. Do NOT fix this instance — the rule is that the predicate models a three-way merge
            and must be SYMMETRIC in head and trunk — a path at base and absent on EITHER side was deleted
            by that side and cannot survive. cmd/sdlc/issueids.go:142 subtracts only base minus head. The
            formula should read merged(id) = (trunk ∪ head) − (base − head) − (base − trunk).
            Measured with sdlc built at 65aea14 against a real repo plus bare origin — branch cut, then
            main runs `git mv workshop/issues/000007-x.md workshop/history/issues/000007-x.md` (what
            `sdlc merge` does on EVERY close), branch untouched. `sdlc issue lint-ids --base $mergebase
            --trunk origin/main --head HEAD` exits 1 with "#000007 would be claimed by 2 files after
            merge"; the CI adapter 40-duplicate-issue-id.sh exits 1 on the same fixture; `git merge`
            produces exactly one file. refuseDuplicateIssueIDs shares the predicate, so the local gate at
            merge.go step 4.6 dies identically. This is BR-13's failure mode on the mirror axis and would
            fail the required status check on nearly every concurrently-open PR in the fleet.
            The enumeration the class implies is {add, delete, rename/move} x {head, trunk} plus
            both-sides; the table at issueids_test.go:522 has only the head column, and its row "trunk
            archived it while the branch edited it in place" asserts wantIDs [7] — the defect written down
            as the expectation. I verified the symmetric predicate against all ten shapes (head archives,
            trunk archives, head renames, trunk renames, head renumbers, cut-then-publish, second path for
            a live id, pre-existing duplicate, two-files-one-id on a branch, both sides archive): 10/10,
            including the two the current code fails and every one it currently passes.
            Two more members to sweep in the SAME round — refuseDuplicateIssueIDs (issueids.go:255-261)
            leaves base empty when merge-base is unresolvable or its read fails, which erases all deletion
            information and refuses everything; and runIssueLintIDs (issuelintids.go:79-82) labels every
            DuplicateIDsInRef hit on head "pre-existing" without consulting base, so an id the range
            introduces is now reported as both pre-existing and introduced.
          family: gate-predicate-ignores-range-delta
          round: 3
      blocked: true
---

# Gate ledger — ariadne#213 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-09-03T15:35:36-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Critical] `gate-compares-wrong-baseline` The CI check compares against merge-base, so it structurally cannot see the collision this issue exists to catch
  merge-check.yml passes base = merge-base(base_tip, head). For a branch cut BEFORE
  the colliding id landed on main — the issue's own reproduction — the merge-base
  predates that file too, so baseByID has no entry for the id and introducedIDClashes
  finds nothing. Proven in a scratch repo: branch cut at S, main publishes
  000500-theirs.md, branch adds 000500-mine.md. `lint-ids --base <merge-base>` →
  "[ok] this range introduces no reused issue ids", exit 0. `lint-ids --base <main tip>`
  → refuses, exit 1, naming both paths. The enforcement layer passes the case it was
  built for. Fix: the check must resolve the trunk TIP itself (fetch + `git rev-parse
  origin/main`, falling back to the runner-supplied base) rather than using the range
  base, since the runner contract is fixed at merge-base.
- **BR-2** [Critical] `gate-compares-wrong-baseline` issueFilesByID keeps only the first path per id, so an introduced duplicate is detected or missed by slug sort order
  issueFilesByID (cmd/sdlc/issueids.go:190) collapses a within-ref duplicate to the
  first path seen; both refuseDuplicateIssueIDs and introducedIDClashes then compare
  single paths. When the head tree contains BOTH files (a rebased/updated PR, or any
  branch that pulled main after the trunk file landed), head[id] equals base[id] and
  nothing is reported. Proven on the real repo: planting workshop/issues/000213-aaa-
  collision.md → refused, exit 1; planting 000213-planted-collision.md (identical in
  every way except the slug sorts after "nextid") → "[ok] this range introduces no
  reused issue ids", exit 0. Same hole in the merge gate — a scratch test adding
  000001-zzz-later-slug.md next to the trunk's 000001-first.md gets
  "[ok] duplicate-id gate: no reused issue ids". The Spec names this exact hazard
  ("issueFilesByID keeps the first path seen ... which silently collapses exactly the
  state being hunted") but only the within-ref REPORT was fixed, not the comparison.
  Fix: build map[int]map[string]bool (or reuse a shared pure helper alongside
  DuplicateIDsInRef) and flag any head path for an id that base already owns under a
  different path. Both existing tests pass only because their chosen slugs happen to
  sort first.
- **BR-3** [Important] `dir-containment-false-negative` repoRelativeIDDirs rejects a not-yet-created id dir whenever the repo root is reached through a symlink, silently disabling every layer
  cmd/sdlc/issueids.go:107 EvalSymlinks the repo top unconditionally but the candidate
  dir only when it exists. A repo under a symlinked path (macOS /tmp → /private/tmp, a
  symlinked workspace) with no workshop/history/ yet yields abs=/tmp/... vs
  top=/private/tmp/..., filepath.Rel returns "../../..", and the dir is refused as
  "outside the current repo". Observed: `sdlc issue lint-ids` printed "id lint skipped:
  workshop/history is outside the current repo" and exited 0 in a fresh fixture;
  creating workshop/history/ made the same command work. In allocateIssueID this path
  warns "origin/main unreachable" and falls back to the local-only scan — the original
  defect, re-armed, on any repo whose history dir does not exist yet. Fix: resolve the
  nearest existing ancestor of abs (or EvalSymlinks the cwd before Abs) so containment
  is decided on comparable paths.
- **BR-4** [Important] `gate-compares-wrong-baseline` The merge gate at step 4.6 reads a stale origin/main — merge.go's own comment says the flow has not fetched yet
  refuseDuplicateIssueIDs is called at cmd/sdlc/merge.go:340, but merge.go:464-465
  states "origin/main is stale here (the flow doesn't pull until AFTER deciding,
  below)" and only fetches at :469. A collision published to the trunk since this
  checkout last fetched is invisible to the gate that documents itself as "the last
  point where an id collision is still repairable". Fix: fetch before 4.6, or move the
  gate below the existing fetch at :469.
- **BR-5** [Important] `fix-not-pinned-by-a-failing-test` No test exercises the CI check's refusal path; the one that claims to asserts the SKIP path
  TestMergeCheckScript_RefusesAPlantedCollision (cmd/sdlc/issueids_test.go:281) plants a
  collision, then runs the script in a fixture with no ./cmd/sdlc and asserts exit 0
  plus "skipping". Its name states the opposite of what it verifies, and it satisfies
  the Plan row "Test the CI check by running it against a real repo with a planted
  collision" with a test that never reaches the check. Nothing goes red if the refusal
  breaks — which is how both Critical findings above shipped with a green suite. The
  exit-1 path is also unreachable in-process because runIssueLintIDs calls
  exitWithCode → os.Exit (the `die` var seam exists for exactly this). Fix: route the
  refusal through a testable seam and add (a) a Go test asserting exit 1 on an
  introduced clash regardless of slug sort order, and (b) a script-level test against a
  repo where the build can actually run; rename or split the existing skip-path test.
- **BR-6** [Important] `enforcement-does-not-propagate` The CI check does not reach derivatives — merge-checks.d is a scaffold row, and the script self-skips without ./cmd/sdlc
  Done-when claims "it propagates to every derivative through the symlinked runner —
  parley.nvim carries four of the eight known collisions", and atlas/workflow/sdlc-
  binary.md's layer table repeats "propagates to derivatives". Neither holds:
  construct/base.manifest:130 is `scaffold scripts/merge-checks.d` (an empty directory
  per repo — only the runner is symlinked), so the check is an ariadne-local file that
  propagate-base will not carry; and even if copied, scripts/merge-checks.d/40-
  duplicate-issue-id.sh:33 exits 0 when ./cmd/sdlc is absent, which is true of every
  derivative by construction (verified: parley.nvim has no cmd/sdlc and an empty
  merge-checks.d). Fix: either resolve the sdlc module from the cloned upstream peer
  (CI already runs BOOTSTRAP_CLONE_ONLY=1 ./bootstrap.sh, so ../<upstream>/cmd/sdlc
  exists) plus a manifest row that actually delivers the check, or correct the
  Done-when and the atlas table to say ariadne-only.
- **BR-7** [Important] `silent-degradation-in-allocator` publishedIssueIDs swallows per-directory ls-tree failures, so a partial trunk read allocates a colliding id with no warning
  cmd/sdlc/issueids.go:88 `continue`s on ls-tree error. A missing directory is already
  exit 0 with empty output (verified), so a non-nil error means a real git failure —
  and dropping that directory's ids silently under-counts the published space, which is
  precisely the silent-fallback failure the Spec forbids ("a silent fallback here
  recreates the bug it is meant to fix"). Fix: return the error so allocateIssueID
  takes its loud-warning path, or warn per directory.
- **BR-8** [Minor] `duplicated-listing-parser` Three near-identical ls-tree listing parsers and two identical rev-parse+ls-tree IO shells (ARCH-DRY)
  issue.IDsInTreeListing, issue.DuplicateIDsInRef and issueFilesByID each re-implement
  split → trim → LastIndex("/") → IDFromFilename; publishedIssueIDs and idListing each
  re-implement rev-parse + repoRelativeIDDirs + per-dir ls-tree. The duplication is not
  cosmetic: the collapsing map in issueFilesByID exists only because it re-implements
  instead of reusing the id→paths structure DuplicateIDsInRef already builds. One pure
  PathsByIDInTreeListing in the issue package fixes the DRY violation and the
  order-dependence Critical together.
- **BR-9** [Minor] `docs-lag-new-surface` atlas/workflow/ci-merge-check.md renders the new prose inside a fenced code block
  The paragraphs added at lines 31-45 sit between the opening ``` at line 30 and its
  closing fence at line 46, so they render as literal code.
- **BR-10** [Minor] `docs-lag-new-surface` `issue lint-ids` is missing from the SUBCOMMANDS list in cmd/sdlc/helptext/issue.md
  The list at issue.md:6-14 enumerates new/sync/set-status/list/show; the new verb is
  not there. (`validate` is also absent — pre-existing.)
- **BR-11** [Minor] `gate-bypass-flag-granularity` The duplicate-id gate is bundled behind --no-validate rather than its own --no-<gate> flag
  AGENTS.md §5 asks for a per-gate flag so a bypass is an explicit acknowledgment of the
  specific gate. Today skipping #124's instance-conformance gate also silently skips the
  id gate, and merge.md's FLAGS block documents neither.
- **BR-12** [Minor] `unbounded-external-call` `git fetch origin main` on every `sdlc issue new` has no timeout (ARCH-CONSTRAINTS)
  allocateIssueID's fetch is best-effort but unbounded. A host that drops packets rather
  than refusing (VPN down, sleeping laptop) blocks an interactive verb for the TCP
  connect timeout. Consider a short --git-timeout / GIT_HTTP_LOW_SPEED_* bound, or an
  explicit --no-fetch escape, and state the budget in the issue.

## Round 2 — 2026-09-03T17:00:45-07:00 (claude) — BLOCKED

### Disposed

- BR-1 — addressed — Revert-verified — reverting the script to base="$fallback_base" turns TestMergeCheckScript_RefusesGivenMergeBase red with "[ok] no reused ids".
- BR-2 — addressed — Revert-verified — first-path-only map turns the zzz-sorts-last subtest red while aaa stays green.
- BR-3 — not-addressed — Behavior change is real (measured both binaries under a symlinked root) but NO test pins it — reverting repoRelativeIDDirs to d50a023's form leaves every #213 test green — and the class survives for absolute dirs: --issues-dir under a symlinked path with workshop/history (or history/issues) absent still prints "outside the current repo" and disables all three layers.
- BR-4 — addressed — Fetch added at issueids.go:172, unconditional and reachable; no test pins the ordering (TestRefuseDuplicateIssueIDs' origin is already current), noted in coverage.
- BR-5 — addressed — The refusal path is now the tested one and it genuinely goes red when the BR-1 fix is reverted; the skip path is a separate, correctly named test.
- BR-6 — addressed — Verified by weave compile --dry-run in parley.nvim (emits the symlink row) plus a derivative-shaped fixture with no cmd/sdlc; atlas and Done-when corrected. Actual propagation to parley.nvim has not been run yet - operator action.
- BR-7 — addressed — Fixed and pinned at publishedIssueIDs, but only there - see the new finding for the two sibling sites that still swallow.
- BR-8 — not-addressed — Three listing parsers and three rev-parse+ls-tree shells remain; PathsByIDInTreeListing was never written, and the shells' error handling has now diverged.
- BR-9 — not-addressed — atlas/workflow/ci-merge-check.md still opens a fence at line 31 and closes it at 47, so lines 33-46 render as code; that prose is now also stale.
- BR-10 — not-addressed — cmd/sdlc/helptext/issue.md SUBCOMMANDS still lists new/sync/set-status/list/show only.
- BR-11 — not-addressed — Still gated on !f.NoValidate at merge.go:339, no --no-dupid, no FLAGS documentation, and no warning printed when the bypass fires.
- BR-12 — not-addressed — No timeout or --no-fetch escape; merge step 4.6 now adds a second unbounded fetch.

### Raised

- **BR-13** [Critical] `gate-predicate-ignores-range-delta` Archiving or renaming an issue file is refused as a reused id by both the CI check and the merge gate
  newPathsFor keys on exact path equality, so moving workshop/issues/NNNNNN-x.md to
  workshop/history/issues/NNNNNN-x.md - the archive step AGENTS.md section 1 mandates
  on done - reads as a second file claiming a live id. Replayed against this repo's
  real merged PR 109 (base 008f7e3^1, head 008f7e3^2): exit 1, "this range reuses 1
  issue id(s)", naming 000195 and telling the operator to renumber it. sdlc merge step
  4.6 shares the predicate and dies the same way. Fix: decide "introduced" from the
  range delta - exclude paths that are rename destinations in
  git diff --name-status -M base head over the id dirs - rather than from set-differencing
  two trees' path lists. Verified the corrected predicate still refuses BR-1's
  cut-then-publish shape and still only reports the eight already-merged collisions.
- **BR-14** [Important] `gate-predicate-ignores-range-delta` A within-ref duplicate the range introduces is labelled pre-existing and passes, contradicting Done-when
  runIssueLintIDs runs DuplicateIDsInRef over the head listing and reports every result
  as pre-existing; introducedIDClashes cannot see them because base[id] is empty. A
  fixture whose branch adds 000500-agent-a.md and 000500-agent-b.md (neither on the
  trunk) prints "pre-existing duplicate id 000500" followed by "id lint: this range
  introduces no reused issue ids" and exits 0. Done-when states the scan "REFUSES when
  the PR introduces one". Same root cause and same fix as the Critical above: an id is
  introduced-within-ref when head holds two or more distinct paths for it and at least
  one is absent at base.
- **BR-15** [Important] `silent-degradation-in-allocator` BR-7's class was fixed at one of three sites; issueFilesByID and idListing still swallow ls-tree failures
  This is the 2nd finding in family silent-degradation-in-allocator. Do NOT fix these
  two sites in isolation. The rule: a partial read of the id space must be an error at
  EVERY site that performs one, and no code path may report a clean verdict from an
  incomplete listing. The enumeration is mechanical - three functions run per-directory
  ls-tree (publishedIssueIDs, issueFilesByID at issueids.go:220, idListing at
  issuelintids.go:143); round 1 fixed the first only. A dropped directory in
  issueFilesByID removes ids from the BASE map, which reads as "this id is new" - a
  false negative inside the enforcement gate. runIssueLintIDs compounds it by turning
  every error into return nil (exit 0) at three places. Fix the class with one shared
  IO shell that returns an error (which also discharges BR-8).
- **BR-16** [Important] `gate-compares-wrong-baseline` git fetch origin main does not guarantee refs/remotes/origin/main, so CI can fall back to the merge-base baseline BR-1 proved blind
  This is the 4th finding in family gate-compares-wrong-baseline. Earlier rounds fixed
  instances. The rule that covers all of them: a gate must resolve its authoritative
  baseline explicitly and fail loudly when it cannot, never silently substitute a
  baseline that is structurally blind to the class it exists to catch. Measured - with
  a narrow remote.origin.fetch, "git fetch origin main" leaves origin/main unresolvable
  while "git fetch origin +refs/heads/main:refs/remotes/origin/main" resolves it. When
  it fails, 40-duplicate-issue-id.sh falls back to $fallback_base, which is the
  merge-base BR-1 showed cannot see the collision. I did not prove this fires under the
  shim's fetch-depth 0; the explicit refspec removes the dependency on that assumption,
  and the fallback should be a loud degraded state, not a pass.
- **BR-17** [Minor] `docs-lag-new-surface` atlas ci-merge-check.md still describes the pre-BR-6 skip conditions; three of five doc homes for this surface are wrong
  This is the 3rd finding in family docs-lag-new-surface. Do NOT fix only this file.
  The rule: a change to a user-facing surface updates that surface's doc homes in the
  SAME commit - helptext, the atlas entry, and README where the verb is listed - and
  corrects prose describing the old behavior. Measured enumeration for 213: helptext
  issue.md SUBCOMMANDS missing lint-ids (BR-10); helptext merge.md FLAGS silent on the
  gate (BR-11); ci-merge-check.md lines 41-45 both trapped in a code fence (BR-9) and
  still claiming the script keys on ./cmd/sdlc, which is exactly what BR-6 changed;
  sdlc-binary.md correct; README not applicable (issue verbs are not listed there).
  Three of five homes wrong - sweep the list.

## Round 3 — 2026-09-03T17:49:32-07:00 (claude) — BLOCKED

### Disposed

- BR-3 — addressed — Verified live — fresh repo under /tmp (→/private/tmp) with no workshop/history: lint reported the planted duplicate instead of skipping.
- BR-13 — addressed — The named instance is fixed (real PR 109 replay, base 008f7e3^1 / head 008f7e3^2 → exit 0), but the class is not swept — see the new Critical.
- BR-14 — addressed — Verified — a branch adding 000500-agent-a.md and 000500-agent-b.md now exits 1 rather than labelling them pre-existing and passing.
- BR-8 — not-addressed — Three listing parsers and three rev-parse+ls-tree IO shells still present, unchanged across all three rounds.
- BR-9 — not-addressed — atlas/workflow/ci-merge-check.md lines 30-46 unchanged; the prose still renders inside the code fence.
- BR-10 — not-addressed — helptext/issue.md SUBCOMMANDS still omits lint-ids, and now also the new --trunk flag added in 65aea14.
- BR-11 — not-addressed — merge.go:336 still gates on !f.NoValidate; merge.md FLAGS still documents neither gate.
- BR-12 — not-addressed — Still unbounded, and now at two call sites (issueids.go:83 and :246) rather than one.
- BR-15 — not-addressed — idListing (issuelintids.go:149) still does `continue`; issueFilesByID's fix is pinned by no test — reverting issueids.go:303 to `continue` leaves the suite green.
- BR-16 — not-addressed — The explicit refspec landed in Go only; scripts/merge-checks.d/40-duplicate-issue-id.sh:65 still runs `git fetch --quiet origin main`. Re-measured on a single-branch clone with a narrow refspec — plain form leaves origin/main unresolvable, explicit form resolves it. Sibling site — introducedIDClashes at issuelintids.go:120-124 silently substitutes baseByID when the trunk read fails.
- BR-17 — not-addressed — No doc home moved this round, and 65aea14 added a new one — the --trunk flag and the merge-result model appear in no helptext or atlas entry. sdlc-binary.md:187 still describes only a branch-vs-trunk comparison; ci-merge-check.md still describes the pre-BR-6 skip conditions and build location.

### Raised

- **BR-18** [Critical] `gate-predicate-ignores-range-delta` mergedPathsFor honours deletions by head but not by the trunk, so every PR open across an issue archive on main is falsely refused
  This is the 3rd finding in family gate-predicate-ignores-range-delta. Earlier rounds fixed
  instances. Do NOT fix this instance — the rule is that the predicate models a three-way merge
  and must be SYMMETRIC in head and trunk — a path at base and absent on EITHER side was deleted
  by that side and cannot survive. cmd/sdlc/issueids.go:142 subtracts only base minus head. The
  formula should read merged(id) = (trunk ∪ head) − (base − head) − (base − trunk).
  Measured with sdlc built at 65aea14 against a real repo plus bare origin — branch cut, then
  main runs `git mv workshop/issues/000007-x.md workshop/history/issues/000007-x.md` (what
  `sdlc merge` does on EVERY close), branch untouched. `sdlc issue lint-ids --base $mergebase
  --trunk origin/main --head HEAD` exits 1 with "#000007 would be claimed by 2 files after
  merge"; the CI adapter 40-duplicate-issue-id.sh exits 1 on the same fixture; `git merge`
  produces exactly one file. refuseDuplicateIssueIDs shares the predicate, so the local gate at
  merge.go step 4.6 dies identically. This is BR-13's failure mode on the mirror axis and would
  fail the required status check on nearly every concurrently-open PR in the fleet.
  The enumeration the class implies is {add, delete, rename/move} x {head, trunk} plus
  both-sides; the table at issueids_test.go:522 has only the head column, and its row "trunk
  archived it while the branch edited it in place" asserts wantIDs [7] — the defect written down
  as the expectation. I verified the symmetric predicate against all ten shapes (head archives,
  trunk archives, head renames, trunk renames, head renumbers, cut-then-publish, second path for
  a live id, pre-existing duplicate, two-files-one-id on a branch, both sides archive): 10/10,
  including the two the current code fails and every one it currently passes.
  Two more members to sweep in the SAME round — refuseDuplicateIssueIDs (issueids.go:255-261)
  leaves base empty when merge-base is unresolvable or its read fails, which erases all deletion
  information and refuses everything; and runIssueLintIDs (issuelintids.go:79-82) labels every
  DuplicateIDsInRef hit on head "pre-existing" without consulting base, so an id the range
  introduces is now reported as both pre-existing and introduced.

## Open findings

- **BR-8** [Minor] `duplicated-listing-parser` Three near-identical ls-tree listing parsers and two identical rev-parse+ls-tree IO shells (ARCH-DRY)
- **BR-9** [Minor] `docs-lag-new-surface` atlas/workflow/ci-merge-check.md renders the new prose inside a fenced code block
- **BR-10** [Minor] `docs-lag-new-surface` `issue lint-ids` is missing from the SUBCOMMANDS list in cmd/sdlc/helptext/issue.md
- **BR-11** [Minor] `gate-bypass-flag-granularity` The duplicate-id gate is bundled behind --no-validate rather than its own --no-<gate> flag
- **BR-12** [Minor] `unbounded-external-call` `git fetch origin main` on every `sdlc issue new` has no timeout (ARCH-CONSTRAINTS)
- **BR-15** [Important] `silent-degradation-in-allocator` BR-7's class was fixed at one of three sites; issueFilesByID and idListing still swallow ls-tree failures
- **BR-16** [Important] `gate-compares-wrong-baseline` git fetch origin main does not guarantee refs/remotes/origin/main, so CI can fall back to the merge-base baseline BR-1 proved blind
- **BR-17** [Minor] `docs-lag-new-surface` atlas ci-merge-check.md still describes the pre-BR-6 skip conditions; three of five doc homes for this surface are wrong
- **BR-18** [Critical] `gate-predicate-ignores-range-delta` mergedPathsFor honours deletions by head but not by the trunk, so every PR open across an issue archive on main is falsely refused
