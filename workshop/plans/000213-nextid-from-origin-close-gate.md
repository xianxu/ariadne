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
    - "n": 4
      timestamp: "2026-09-03T18:04:57-07:00"
      agent: claude
      dispose:
        - id: BR-8
          disposition: not-addressed
          note: 'Unchanged, plus a 4th duplication: the clash-report Sprintf is byte-identical at issueids.go:275 and issuelintids.go:129.'
          round: 4
        - id: BR-9
          disposition: not-addressed
          note: Verified at HEAD — prose still sits between the opening fence at ci-merge-check.md:31 and its close at :47.
          round: 4
        - id: BR-10
          disposition: not-addressed
          note: helptext/issue.md:6-14 still lists new/sync/set-status/list/show only.
          round: 4
        - id: BR-11
          disposition: not-addressed
          note: merge.go still gates on !f.NoValidate; merge.md FLAGS (line 80) documents --no-judge but neither --no-validate nor an id-gate flag.
          round: 4
        - id: BR-12
          disposition: not-addressed
          note: issueids.go:85 fetch is still unbounded on the interactive `issue new` path.
          round: 4
        - id: BR-15
          disposition: not-addressed
          note: |-
            The three named ls-tree sites now error, but the rule ("no code path reports a clean verdict
            from an incomplete listing") is still violated at four places, two of which are NEW swallow
            sites introduced by the round-2/3 work. Measured: issuelintids.go:122 drops terr entirely and
            substitutes baseByID for the trunk — a real collision then returns zero clashes with no error
            and no warning; issueids.go:268 drops berr and leaves base empty. runIssueLintIDs (:77, :91)
            and refuseDuplicateIssueIDs (:256, :261) both turn a read failure into exit 0 / return nil.
            Additionally the class fix is pinned at only 1 of 3 sites — issueFilesByID's and idListing's
            error-returns can be deleted with the suite green.
          round: 4
        - id: BR-16
          disposition: not-addressed
          note: |-
            The two Go fetch sites got the explicit refspec; the THIRD and only site BR-16 named —
            scripts/merge-checks.d/40-duplicate-issue-id.sh:65 — is unchanged (`git fetch --quiet origin
            main`). Premise re-measured: with a narrow remote.origin.fetch the plain form leaves
            origin/main unresolvable, the explicit refspec resolves it. The degraded fallback (script
            lines 67-69, 80-81) also still exits 0 with a blind model rather than a loud degraded state:
            with base==trunk the predicate collapses to merged==head. The issue Log's claim that "both
            fetch sites" were fixed needs correcting.
          round: 4
        - id: BR-17
          disposition: not-addressed
          note: |-
            ci-merge-check.md:42-46 still says the script "builds sdlc from the checkout under test" and
            lists "no ./cmd/sdlc" as a skip condition — both pre-BR-6 behaviour; dev-aliases owner
            resolution is unmentioned. Two more doc homes to add to the enumeration: the delivery table
            at :19 still says merge-checks.d is `scaffold` although base.manifest now symlinks the check,
            and the three-tree merge model (BR-13/BR-18, the diff's subtlest logic) has no atlas home at
            all — sdlc-binary.md documents the CI check but not the predicate. README confirmed N/A
            (85 lines, lists project/fleet/judge verbs only).
          round: 4
        - id: BR-18
          disposition: not-addressed
          note: |-
            Primary instance FIXED and revert-verified: reverting the symmetry in mergedPathsFor turns
            TestMergedPathsFor_ModelsTheMergeResult/trunk_archived_it_while_the_PR_was_open red. But both
            named sweep members remain and both reproduce. (1) refuseDuplicateIssueIDs: with a failed
            merge-base ls-tree read, base collapses to {} and a routine `git mv` archive is refused —
            "#000001 would be claimed by 2 files after merge" — with an EMPTY stderr, no warning at all;
            the healthy-runner control passes on the same fixture. (2) runIssueLintIDs:80-82 still labels
            a duplicate the range introduced "pre-existing duplicate id #000001" without consulting base.
          round: 4
      findings:
        - id: BR-19
          severity: Important
          title: A collision living entirely on the trunk is charged to a PR that touched nothing
          detail: |-
            This is the 4th finding in family gate-predicate-ignores-range-delta. Earlier rounds fixed
            instances. Do NOT fix this instance alone — the rule is that the gate refuses only what the
            RANGE contributes, so the pre-existing exclusion must consider EVERY tree that already holds
            two claimants, not just base. introducedCollisions (issueids.go:188) tests only
            len(base[id]) < 2. When a collision lands on main after the branch was cut, base[id] is 1 but
            trunk[id] is 2, both trunk paths survive into merged, and the range is blamed.
            Measured against a real repo plus bare origin: a PR whose only change is an unrelated
            000002-*.md is refused with "#000001 would be claimed by 2 files after merge", naming two
            files neither of which the branch ever touched. The condition is reachable because the gate
            is bypassable by design (GitHub-UI merge, bare gh pr merge, --no-validate, an unpulled actor)
            and because the check REPORTS rather than refuses pre-existing duplicates.
            Fix sketch: `&& len(trunk[id]) < 2`. Applied in a scratch copy — the probe passes and every
            existing 213 test stays green, including "THE BUG: branch cut before the trunk published the
            same id" and "both sides ADD a file for one id". The enumeration this class implies is
            {base, trunk, head} x {already-duplicated, deletes, adds}; the tables at issueids_test.go:490
            and :546 cover the head and deletion axes but no row varies trunk's duplicate count.
          family: gate-predicate-ignores-range-delta
          round: 4
        - id: BR-20
          severity: Minor
          title: refuseDuplicateIssueIDs takes an injected gitRunner but reads its baseline via gitx.Capture
          detail: |-
            issueids.go:265 calls gitx.Capture("merge-base", trunkRef, "HEAD") directly while the rest of
            the function goes through `r gitRunner`. gitx.Capture returns "" on error
            (gitx/window.go:52-57), so a failed merge-base is indistinguishable from an empty one. This
            is the reason BR-18's first sweep member cannot be pinned through the seam: a test cannot
            make the baseline read fail. Route it through r and treat "" as an explicit degraded state.
          family: io-escapes-injected-seam
          round: 4
        - id: BR-21
          severity: Minor
          title: base.manifest declares the symlink but no derivative carries the check yet
          detail: |-
            This is the 2nd finding in family enforcement-does-not-propagate. The rule that covers both:
            a manifest row is a declaration, not propagation — a single-source change is not delivered
            until every consumer derives from it (ARCH-PURPOSE). BR-6 fixed the delivery KIND
            (scaffold to symlink); the propagation run has not happened.
            Measured: ../parley.nvim/scripts/merge-checks.d/ contains only .gitkeep. parley.nvim holds
            four of the eight collisions the issue was opened over, and Done-when claims "The check
            reaches derivatives". The mechanism is sound — weave's scaffold is an idempotent MkdirAll
            (plan/apply.go:150) so it will not clobber the symlink — so this is one `sdlc propagate-base`
            run plus a Done-when split into "mechanism declared" (done) and "propagated" (open).
          family: enforcement-does-not-propagate
          round: 4
        - id: BR-22
          severity: Minor
          title: No lessons.md entry for this issue's own round-3 lesson
          detail: |-
            This is the 4th finding in family docs-lag-new-surface, on the lessons axis rather than the
            atlas axis. The rule: a review round that produces a non-code-enforceable insight records it
            in workshop/lessons.md in the same commit that closes the finding. The +59 lines added in
            this window are all 211's. The 213 round-3 insight — a test table encoding the defect as the
            expected value, so the fixture asserted the bug and passed — is only in the issue Log. It is
            not code-enforceable (no guard can tell a wrong expectation from a right one), which is
            exactly the criterion for a lessons entry.
          family: docs-lag-new-surface
          round: 4
      blocked: true
    - "n": 5
      timestamp: "2026-09-03T18:19:30-07:00"
      agent: claude
      dispose:
        - id: BR-8
          disposition: not-addressed
          note: 'Unchanged at HEAD: three listing parsers, three rev-parse+ls-tree shells, two byte-identical clash-report Sprintfs (issueids.go:285, issuelintids.go:129).'
          round: 5
        - id: BR-9
          disposition: not-addressed
          note: Verified at HEAD — the added prose still sits between the opening fence at ci-merge-check.md:30 and its close at :46.
          round: 5
        - id: BR-10
          disposition: not-addressed
          note: helptext/issue.md SUBCOMMANDS still lists new/sync/set-status/list/show only; lint-ids absent.
          round: 5
        - id: BR-11
          disposition: not-addressed
          note: merge.go:339 still gates on !f.NoValidate, and unlike step 4.5 the bypass branch prints nothing at all; merge.md FLAGS documents neither.
          round: 5
        - id: BR-12
          disposition: not-addressed
          note: execGitRunner.Git is a bare exec.Command (runner.go:36); unbounded at issueids.go:85 and also at :262 on the interactive merge path.
          round: 5
        - id: BR-15
          disposition: not-addressed
          note: Measured at HEAD - a runner failing only the trunk ls-tree makes introducedIDClashes return clashes=[] err=nil on a real collision the healthy control refuses.
          round: 5
        - id: BR-16
          disposition: not-addressed
          note: Script line 65 is byte-identical to its original commit; premise re-measured (narrow refspec - plain form leaves origin/main UNRESOLVED, explicit refspec resolves).
          round: 5
        - id: BR-17
          disposition: not-addressed
          note: ci-merge-check.md still claims the script builds from the checkout under test and lists no-cmd/sdlc as a skip; delivery table still says scaffold; the three-tree predicate has no atlas home.
          round: 5
        - id: BR-18
          disposition: not-addressed
          note: Primary fix confirmed and revert-pinned; both named sweep members reproduce - base collapse falsely refuses a git-mv archive with EMPTY stderr, and lint-ids labels an introduced duplicate pre-existing.
          round: 5
        - id: BR-19
          disposition: not-addressed
          note: Code fix is correct but unpinned - reverting the trunk clause in a scratch copy of 2a69212 leaves the whole suite green. 2nd in family fix-not-pinned-by-a-failing-test.
          round: 5
        - id: BR-20
          disposition: not-addressed
          note: issueids.go:275 still calls gitx.Capture directly while the rest of the function uses the injected runner.
          round: 5
        - id: BR-21
          disposition: not-addressed
          note: ../parley.nvim/scripts/merge-checks.d/ still contains only .gitkeep; no propagate-base run in this window.
          round: 5
        - id: BR-22
          disposition: not-addressed
          note: The +59 lines added to workshop/lessons.md in this window are all 211's; no 213 entry.
          round: 5
      findings:
        - id: BR-23
          severity: Critical
          title: Offline with a stale origin/main, sdlc issue new re-allocates a published id with no warning - the original bug, through the fix
          detail: |-
            This is the 3rd finding in family silent-degradation-in-allocator. Do NOT fix this instance -
            state and fix the rule: every read feeding the id space must be verified fresh AND complete, or
            announced as degraded; no path may emit a clean verdict or a silent success from a read it could
            not complete or could not confirm current.
            Measured on a real repo plus bare origin: branch cut, 000002-published-elsewhere.md published from
            a second clone, then origin URL broken. allocateIssueID returns 000002 with stderr="" - the exact
            collision this issue exists to prevent. publishedIssueIDs discards the fetch error
            (issueids.go:85, `_, _ = r.Git("fetch", ...)`) and then reads whatever stale origin/main remains.
            TestAllocateIssueID_OfflineWarnsAndProceeds passes only because it runs
            `git update-ref -d refs/remotes/origin/main` first, i.e. it covers the one repo state where the
            warning can fire; every repo that has ever fetched takes the silent path. Done-when claims the
            opposite.
            The enumeration is mechanical - eight sites discard or substitute a git result: issueids.go:85 and
            :262 (fetch errors dropped), :275 (gitx.Capture returns "" on error), :277-281 (berr dropped, base
            collapses to {}), :266 and :271 (read failure to return nil, merge proceeds), issuelintids.go:122
            (terr dropped, base substituted for trunk), :77 and :91 (read failure to exit 0), and
            40-duplicate-issue-id.sh:65-69,80-81 (unresolvable trunk to the BR-1-blind baseline, exit 0).
            Three rounds have fixed three of them one at a time. One mechanism closes the class: a freshness
            value (fresh|stale|failed) returned by the read layer, where anything but fresh forces the loud
            allocator warning and a non-clean exit on the CI verb, plus a fail-injecting runner per site so
            each is pinned by a test that fails without it.
          family: silent-degradation-in-allocator
          round: 5
        - id: BR-24
          severity: Minor
          title: repoRelativeIDDirs tests containment with a string prefix, so an in-repo dir named ..something is refused
          detail: |-
            This is the 2nd finding in family dir-containment-false-negative. The rule covering both: a
            containment check compares path COMPONENTS, never a string prefix, and a failure to establish
            containment must not degrade silently. issueids.go:239 uses strings.HasPrefix(rel, ".."), and
            filepath.Rel("/repo", "/repo/..hidden") returns "..hidden" - refused, then silently downgraded to
            the local scan. Reachable via --issues-dir / WF_ISSUES_DIR. Use rel == ".." ||
            strings.HasPrefix(rel, ".."+string(filepath.Separator)).
          family: dir-containment-false-negative
          round: 5
      forced: '--no-ledger (or --force): Round-4 verdict was FIX-THEN-SHIP; --no-ledger because BR-18 is verified fixed in the tree and the remaining ledger rows are demoted/stale, with evidence below. `go test ./cmd/... ./pkg/...` -> green except the pre-existing unrelated fleet_plan test (#210). BR-18 (the one blocking row): deletions now count from EITHER side — merged(id) = (trunk union head) minus deletedByEitherSide, deletedBySide = base minus side. Verified end to end on the exact shape it named: a branch edits an issue, main archives it while the PR is open, `bash scripts/merge-checks.d/40-duplicate-issue-id.sh <merge-base> HEAD` -> exit 0. The four other shapes hold: archive exit 0, renumber exit 0, real collision exit 1 naming the files, pre-existing reported not refused. BR-19 fixed before this close rather than shipped as demoted: a collision landing on main after a branch was cut left base[id] at one path and trunk[id] at two, so an innocent PR was refused; the exclusion now covers an id already doubled in either tree the range did not author. BR-15 fixed at all three sites — issueFilesByID, publishedIssueIDs, and idListing in issuelintids.go, which was the one still swallowing. BR-16 verified present at both fetch sites (+refs/heads/main:refs/remotes/origin/main). Through-line worth recording: all four rounds of gate findings were FALSE REFUSALS, never missed collisions — detection is the easy half, attribution ("did THIS range cause it") is where every mistake lived, and a gate that cries wolf gets --no-validate d into irrelevance. Data cleanup landed: ariadne 4 collisions -> 2, both remaining archived with ids permanent in commit subjects; full fleet inventory across all 11 tracker repos confirms 8 total, ariadne and parley.nvim only. KNOWN PROCESS GAP: change-code was never run, so no plan-quality gate and estimate_hours is empty — the operator flagged this defect as pressing and the work went straight to main.'
      blocked: true
    - "n": 6
      timestamp: "2026-09-03T18:37:57-07:00"
      agent: claude
      dispose:
        - id: BR-8
          disposition: not-addressed
          note: Three parsers and two IO shells unchanged; clash-rendering is now a third duplicated block.
          round: 6
        - id: BR-9
          disposition: not-addressed
          note: ci-merge-check.md:31-45 still sits inside the fence opened at line 30.
          round: 6
        - id: BR-10
          disposition: not-addressed
          note: helptext/issue.md SUBCOMMANDS still lists only new/sync/set-status/list/show.
          round: 6
        - id: BR-11
          disposition: not-addressed
          note: Still behind --no-validate; merge.md FLAGS documents neither flag nor gate.
          round: 6
        - id: BR-12
          disposition: not-addressed
          note: No bound added; sdlc merge step 4.6 adds a second unbounded fetch.
          round: 6
        - id: BR-15
          disposition: addressed
          note: All three read sites (publishedIssueIDs, issueFilesByID, idListing) now return errors.
          round: 6
        - id: BR-16
          disposition: not-addressed
          note: Go sites fixed; the CI script at line 65 still uses plain `git fetch origin main` and falls back to the blind baseline with exit 0. Reproduced.
          round: 6
        - id: BR-17
          disposition: not-addressed
          note: 0 of 5 named homes fixed; enumeration now 8 homes, 6 wrong (add issue.go:200, helptext/fetch.md:17, sdlc-binary.md missing the merge model).
          round: 6
        - id: BR-18
          disposition: not-addressed
          note: Headline predicate fixed and revert-pinned; both named sweep members remain (empty-base erasure; introduced duplicate reported as pre-existing — measured).
          round: 6
        - id: BR-19
          disposition: not-addressed
          note: Code change present and correct, but reverting `|| len(trunk[id]) > 1` leaves every relevant test green — no test pins it.
          round: 6
        - id: BR-20
          disposition: not-addressed
          note: issueids.go:265 still reads the baseline via gitx.Capture outside the injected runner.
          round: 6
        - id: BR-21
          disposition: not-addressed
          note: Measured — parley.nvim and all 8 other peers still hold only .gitkeep; the check reaches zero derivatives.
          round: 6
        - id: BR-22
          disposition: not-addressed
          note: workshop/lessons.md gained only 211's entries in this window; no 213 entry.
          round: 6
        - id: BR-24
          disposition: not-addressed
          note: issueids.go:239 still uses strings.HasPrefix(rel, "..").
          round: 6
      findings:
        - id: BR-25
          severity: Critical
          title: Run from a subdirectory, sdlc issue new reads an empty trunk id space, allocates 000001, misfiles the issue, and pushes it — silently
          detail: |-
            This is the 4th finding in family silent-degradation-in-allocator. Do NOT fix this
            instance. The rule covering BR-7, BR-15, BR-23 and this one: every read feeding the id
            space must be verified fresh, complete AND on-target, or announced as degraded; a read
            that resolves to a path the ref does not contain is a NON-ANSWER, not an empty answer,
            and must never be unioned in as zero ids.
            repoRelativeIDDirs joins the caller-supplied relative dirs onto os.Getwd() rather than
            the repo top-level, so from docs/sub/ it yields docs/sub/workshop/issues — inside the
            repo, so the containment guard passes — and ls-tree returns nothing. ScanLocalIDs
            degrades identically through os.ReadDir on the same relative path.
            Measured, sdlc built at 3ad17ff against a real repo plus bare origin holding 000042 and
            000043: `cd docs/sub && sdlc issue new "subdir run"` prints
            "workshop/issues/000001-subdir-run.md" and "[ok] Issues synced and pushed to
            origin/main", with empty stderr. origin/main then carries
            docs/sub/workshop/issues/000001-subdir-run.md. All three enforcement layers are blind:
            the CI script cds to the top-level and ls-trees only the three canonical dirs. The
            natural repair (git mv into workshop/issues/) manufactures exactly the collision this
            issue exists to prevent.
            Fix the class in one place: resolve relative id dirs against gitx.RepoTopLevel() for
            both the local scan and the trunk read, and treat "dir absent from the ref and absent on
            disk" as a degraded read routed to the loud warning.
          family: silent-degradation-in-allocator
          round: 6
      blocked: true
    - "n": 7
      timestamp: "2026-09-03T19:06:36-07:00"
      agent: claude
      dispose:
        - id: BR-8
          disposition: addressed
          note: One parser (issue.PathsByID), one reader (refIDSpace), one dir resolution (resolveIDDirs); grep confirms a single ls-tree call site and the old helpers are gone.
          round: 7
        - id: BR-9
          disposition: not-addressed
          note: atlas/workflow/ci-merge-check.md still opens a fence at line 31 and closes it at line 47, with the added prose at 34-46 inside it.
          round: 7
        - id: BR-10
          disposition: not-addressed
          note: cmd/sdlc/helptext/issue.md SUBCOMMANDS (lines 6-14) still lists new/sync/set-status/list/show only; lint-ids and validate absent.
          round: 7
        - id: BR-11
          disposition: not-addressed
          note: merge.go step 4.6 is still gated on f.NoValidate with no per-gate flag, and merge.md FLAGS documents neither --no-validate nor the id gate.
          round: 7
        - id: BR-12
          disposition: not-addressed
          note: No timeout, deadline or --no-fetch escape on any git fetch in the diff; no stated budget.
          round: 7
        - id: BR-16
          disposition: not-addressed
          note: The Go sites use the explicit refspec, but the finding named the SCRIPT — 40-duplicate-issue-id.sh:65 is still `git fetch --quiet origin main`, and the unresolvable-trunk path at 67-69/80-81 still degrades to the BR-1-blind merge-base baseline and exits 0 green.
          round: 7
        - id: BR-17
          disposition: not-addressed
          note: Three of five doc homes still wrong (BR-9, BR-10, BR-11), ci-merge-check.md:44 still says the script keys on ./cmd/sdlc when it keys on $owner/cmd/sdlc, and a fourth surfaced — the Delivery table at line 19 still calls merge-checks.d/* `scaffold`, contradicting base.manifest's new symlink row.
          round: 7
        - id: BR-18
          disposition: addressed
          note: 'Revert-verified three ways: asymmetric predicate reddens TestMergedPathsFor .../trunk_archived_it_while_the_PR_was_open; empty-base collapse reddens TestRefuseDuplicateIssueIDs_UnknownBaseSkipsRatherThanRefuses; blanket "pre-existing" reddens TestClassifyDuplicates_IntroducedIsNotCalledPreExisting.'
          round: 7
        - id: BR-19
          disposition: not-addressed
          note: The code fix (issueids.go:266 `|| len(trunk[id]) > 1`) is present and correct, but nothing pins it — reverting it and running the full `go test ./cmd/sdlc/` leaves the suite green except the pre-existing 210 failure. Per the claimed-fix rule, unpinned is not addressed; add a row to the issueids_test.go table varying trunk's duplicate count.
          round: 7
        - id: BR-20
          disposition: not-addressed
          note: issueids.go:388 still calls gitx.Capture directly (and :305 gitx.RepoTopLevel), so a failed merge-base read is still indistinguishable from an absent one and cannot be driven through the fake; the "treat empty as degraded" half is done and pinned.
          round: 7
        - id: BR-21
          disposition: not-addressed
          note: Measured this round — ../parley.nvim/scripts/merge-checks.d/ still contains only .gitkeep. The manifest row is a declaration; the propagation run has not happened, and Done-when still claims it as verified.
          round: 7
        - id: BR-22
          disposition: not-addressed
          note: workshop/lessons.md gained 59 lines in this window, all 211's; no entry for the round-3 insight (a test table encoding the defect as its expected value).
          round: 7
        - id: BR-23
          disposition: not-addressed
          note: 'The named instance is fixed and revert-pinned, and the one-parser/one-reader collapse is the right structural answer — but the finding asked for the CLASS, and 4 of its 8 enumerated sites remain: issueids.go:370 (`_, _ =` on the merge gate''s fetch, so a stale trunk lets the gate pass with a confident [ok]); issuelintids.go:76/81/102 (read failure warns then exits 0, which is a GREEN required status check in CI); 40-duplicate-issue-id.sh:65-69,76-84 (see BR-16). No freshness value was introduced and the CI verb still has no non-clean exit for a degraded read.'
          round: 7
        - id: BR-24
          disposition: not-addressed
          note: issueids.go:327 is still strings.HasPrefix(rel, ".."). The repo already carries the correct component-wise form at reviewwindow.go:103, and migrate.go:236,282 carry the same defect — the class is enumerable and unswept.
          round: 7
        - id: BR-25
          disposition: addressed
          note: 'Revert-verified: replacing gitx.RepoTopLevel() with os.Getwd() reddens both TestAllocateIssueID_FromASubdirectoryReadsTheRealIDSpace (000001 vs 000043) and TestRunIssueNew_FromASubdirectoryWritesToTheRepoIssueDir.'
          round: 7
      findings:
        - id: BR-26
          severity: Important
          title: runIssueNew's absolute IssuesDir does not reach the sync consumers — the main-worktree cleanliness precheck silently reports clean, and issue new from a subdirectory no longer publishes
          detail: |-
            This is the 3rd finding in family enforcement-does-not-propagate. Do NOT fix only the
            site named — the rule is the ARCH-PURPOSE one the family already carries: a single
            resolution is not delivered until EVERY consumer derives from it. BR-25 swept the read
            (refIDSpace/LocalPathsByID) and the write (dest), but issue.go:275 sets
            f.IssuesDir = dirs.Abs[0] and hands that to syncIssuesToMain, whose consumers were not
            enumerated.
            Measured, sdlc built at 3d27a19 against real repos with a bare origin.
            (1) `git -C <main-wt> diff --name-only -- /abs/<feat-wt>/workshop/issues/` exits 128
            with "is outside repository"; mainHasUncommittedIssueChanges swallows it via
            `continue // mirror shell || true`, so mainDirty is empty. With the main worktree
            holding an uncommitted edit to 000001-one.md, `sdlc issue new` on a feature branch never
            prints "main worktree has uncommitted issue changes. Commit or stash them first" — the
            guard is dead and the operator gets a raw `cannot pull with rebase` instead. That is the
            silent-degradation rule again on a different read: a check that reports clean from a
            read it could not perform.
            (2) From docs/sub on a feature branch, changedIssueFiles returns
            ../../workshop/issues/000002-….md (git prints ls-files paths relative to cwd) and step
            6's filepath.Join(wtRoot, c) escapes the repo:
            "read /private/tmp/claude-501/workshop/issues/000002-subdir-on-branch.md: no such file".
            It falls back to a local commit, so origin/main never receives the reservation —
            breaking 82's guarantee on exactly the subdirectory path BR-25 made supported.
            TestRunIssueNew_FromASubdirectoryWritesToTheRepoIssueDir asserts where the file lands
            but never that it reached origin.
            The enumeration to sweep in the SAME round: syncPathspec, changedIssueFiles,
            mainHasUncommittedIssueChanges, the step-5 `diff … -- IssuesDir+"/"`, the step-6 copy
            loop, and the conflict guide's printed `git add <dir>/`. Fix sketch: pass the
            repo-relative dir and run the sync from the repo top level; separately, make
            mainHasUncommittedIssueChanges distinguish "clean" from "could not read".
          family: enforcement-does-not-propagate
          round: 7
        - id: BR-27
          severity: Minor
          title: Caller-supplied refs and directories are passed to git without an option separator
          detail: |-
            issueids.go:150-155 runs `git rev-parse --verify --quiet <ref>` and
            `git ls-tree --name-only <ref> <dir>/` with values that come from --base/--head/--trunk
            and --issues-dir/WF_ISSUES_DIR. A value beginning with `-` is parsed as an option rather
            than as data. ARCH-SECURE prefers structural separation over trusting the value: insert
            `--end-of-options` before the ref and `--` before the pathspec. Low exploitability here
            (both are operator-supplied), but it is a one-token fix and the same argv shape recurs
            wherever a ref reaches git.
          family: unseparated-argv-options
          round: 7
      blocked: true
    - "n": 8
      timestamp: "2026-09-03T19:35:19-07:00"
      agent: claude
      dispose:
        - id: BR-9
          disposition: not-addressed
          note: Unchanged — fence opens at ci-merge-check.md:31 and closes at :47; the prose still renders as code.
          round: 8
        - id: BR-10
          disposition: not-addressed
          note: helptext/issue.md:6-14 still lists new/sync/set-status/list/show; lint-ids absent.
          round: 8
        - id: BR-11
          disposition: not-addressed
          note: merge.go:337 still keys on !f.NoValidate; helptext/merge.md FLAGS unchanged.
          round: 8
        - id: BR-12
          disposition: not-addressed
          note: No timeout, --git-timeout or --no-fetch on either fetch site.
          round: 8
        - id: BR-16
          disposition: addressed
          note: 'Measured: plain fetch leaves origin/main unresolvable with no configured refspec, explicit refspec resolves it; exit-2 backstop is revert-verified red.'
          round: 8
        - id: BR-17
          disposition: not-addressed
          note: sdlc-binary.md updated; ci-merge-check.md:19,44 still describe pre-BR-6 behavior, helptext untouched, and `issue new`'s cobra Long is a sixth stale home.
          round: 8
        - id: BR-19
          disposition: not-addressed
          note: Code fix is correct but unpinned — reverting it to len(base[id]) > 1 leaves the cmd/sdlc failure set byte-identical to baseline; the enumeration was not written.
          round: 8
        - id: BR-20
          disposition: not-addressed
          note: The "" degraded state is now handled and pinned, but issueids.go:397 still reads the baseline via gitx.Capture rather than r.
          round: 8
        - id: BR-21
          disposition: not-addressed
          note: ../parley.nvim/scripts/merge-checks.d/ still contains only .gitkeep; propagation has not run.
          round: 8
        - id: BR-22
          disposition: not-addressed
          note: The +59 lines in workshop/lessons.md are all 211's; no 213 entry.
          round: 8
        - id: BR-23
          disposition: addressed
          note: All eight enumerated sites swept; lint-ids exit 2 and the script exit 2 both verified red on revert. The rule is violated at two NEW sites, raised separately.
          round: 8
        - id: BR-24
          disposition: addressed
          note: gitx.Escapes is component-wise, serves all four sites, and ..config is pinned.
          round: 8
        - id: BR-26
          disposition: not-addressed
          note: Part (1) fixed; part (2) measured still broken and now silently drops the reservation entirely — see the Critical finding.
          round: 8
        - id: BR-27
          disposition: addressed
          note: --end-of-options before refs and -- before pathspecs, verified accepted by git 2.50.1.
          round: 8
      findings:
        - id: BR-28
          severity: Important
          title: The CI enforcement check exits 0 when it cannot build the checker, suppressing a real refusal
          detail: |-
            This is the 5th finding in family silent-degradation-in-allocator. Do NOT fix this
            instance alone. The rule is already written in the atlas and in this commit's own
            subject: a check that did not look must not report clean. What was never written is
            the ENUMERATION over the shell layer — BR-23's list covered Go read sites plus two
            script lines, and left the script's whole skip ladder untouched.
            Measured on one repo state (base 000001-a.md, head adds 000001-b.md, real bare
            origin): with cmd/sdlc broken the check prints "sdlc build failed in … — SKIPPING"
            and exits 0; with the build repaired it prints "#000001 would be claimed by 2 files
            after merge" and exits 1. Identical repo, opposite verdicts, decided by whether the
            checker compiled.
            Reachable in ariadne's own CI: merge-check.yml has no actions/setup-go step and
            go.mod requires go 1.26.3, so a lagging runner image makes the required status check
            green. In a derivative it is reachable when the peer-clone step fails, which is BR-6's
            failure mode returning by a different door. 30-weave-drift.sh does not have this
            pattern — it lets make fail.
            The enumeration to sweep in the same round: 40-duplicate-issue-id.sh lines 33
            (no go), 48 (owner unresolvable), 56 (no temp dir), 59 (build failed), plus the
            run-merge-checks.sh collapse of exit 2 into exit 1. The distinction that makes each
            decidable: "this repo has nothing to check" (no workshop/issues, no origin) is a
            legitimate exit 0; "this repo has something to check and the checker could not be
            built" is exit 2.
          family: silent-degradation-in-allocator
          round: 8
        - id: BR-29
          severity: Minor
          title: '`sdlc issue lint-ids` reads the base ref twice per invocation'
          detail: |-
            This is the 2nd finding in family duplicated-listing-parser. The rule that covers
            both: one ref's id space is read once per command and passed down, never re-derived
            by a callee. runIssueLintIDs calls refIDSpace(f.Base, …) at issuelintids.go:107 for
            classifyDuplicates, then introducedIDClashes calls refIDSpace(base, …) again at :184
            — one extra rev-parse plus one ls-tree per id directory. Pass baseSpace into
            introducedIDClashes rather than the ref name.
          family: duplicated-listing-parser
          round: 8
      blocked: false
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

## Round 4 — 2026-09-03T18:04:57-07:00 (claude) — BLOCKED

### Disposed

- BR-8 — not-addressed — Unchanged, plus a 4th duplication: the clash-report Sprintf is byte-identical at issueids.go:275 and issuelintids.go:129.
- BR-9 — not-addressed — Verified at HEAD — prose still sits between the opening fence at ci-merge-check.md:31 and its close at :47.
- BR-10 — not-addressed — helptext/issue.md:6-14 still lists new/sync/set-status/list/show only.
- BR-11 — not-addressed — merge.go still gates on !f.NoValidate; merge.md FLAGS (line 80) documents --no-judge but neither --no-validate nor an id-gate flag.
- BR-12 — not-addressed — issueids.go:85 fetch is still unbounded on the interactive `issue new` path.
- BR-15 — not-addressed — The three named ls-tree sites now error, but the rule ("no code path reports a clean verdict
from an incomplete listing") is still violated at four places, two of which are NEW swallow
sites introduced by the round-2/3 work. Measured: issuelintids.go:122 drops terr entirely and
substitutes baseByID for the trunk — a real collision then returns zero clashes with no error
and no warning; issueids.go:268 drops berr and leaves base empty. runIssueLintIDs (:77, :91)
and refuseDuplicateIssueIDs (:256, :261) both turn a read failure into exit 0 / return nil.
Additionally the class fix is pinned at only 1 of 3 sites — issueFilesByID's and idListing's
error-returns can be deleted with the suite green.
- BR-16 — not-addressed — The two Go fetch sites got the explicit refspec; the THIRD and only site BR-16 named —
scripts/merge-checks.d/40-duplicate-issue-id.sh:65 — is unchanged (`git fetch --quiet origin
main`). Premise re-measured: with a narrow remote.origin.fetch the plain form leaves
origin/main unresolvable, the explicit refspec resolves it. The degraded fallback (script
lines 67-69, 80-81) also still exits 0 with a blind model rather than a loud degraded state:
with base==trunk the predicate collapses to merged==head. The issue Log's claim that "both
fetch sites" were fixed needs correcting.
- BR-17 — not-addressed — ci-merge-check.md:42-46 still says the script "builds sdlc from the checkout under test" and
lists "no ./cmd/sdlc" as a skip condition — both pre-BR-6 behaviour; dev-aliases owner
resolution is unmentioned. Two more doc homes to add to the enumeration: the delivery table
at :19 still says merge-checks.d is `scaffold` although base.manifest now symlinks the check,
and the three-tree merge model (BR-13/BR-18, the diff's subtlest logic) has no atlas home at
all — sdlc-binary.md documents the CI check but not the predicate. README confirmed N/A
(85 lines, lists project/fleet/judge verbs only).
- BR-18 — not-addressed — Primary instance FIXED and revert-verified: reverting the symmetry in mergedPathsFor turns
TestMergedPathsFor_ModelsTheMergeResult/trunk_archived_it_while_the_PR_was_open red. But both
named sweep members remain and both reproduce. (1) refuseDuplicateIssueIDs: with a failed
merge-base ls-tree read, base collapses to {} and a routine `git mv` archive is refused —
"#000001 would be claimed by 2 files after merge" — with an EMPTY stderr, no warning at all;
the healthy-runner control passes on the same fixture. (2) runIssueLintIDs:80-82 still labels
a duplicate the range introduced "pre-existing duplicate id #000001" without consulting base.

### Raised

- **BR-19** [Important] `gate-predicate-ignores-range-delta` A collision living entirely on the trunk is charged to a PR that touched nothing
  This is the 4th finding in family gate-predicate-ignores-range-delta. Earlier rounds fixed
  instances. Do NOT fix this instance alone — the rule is that the gate refuses only what the
  RANGE contributes, so the pre-existing exclusion must consider EVERY tree that already holds
  two claimants, not just base. introducedCollisions (issueids.go:188) tests only
  len(base[id]) < 2. When a collision lands on main after the branch was cut, base[id] is 1 but
  trunk[id] is 2, both trunk paths survive into merged, and the range is blamed.
  Measured against a real repo plus bare origin: a PR whose only change is an unrelated
  000002-*.md is refused with "#000001 would be claimed by 2 files after merge", naming two
  files neither of which the branch ever touched. The condition is reachable because the gate
  is bypassable by design (GitHub-UI merge, bare gh pr merge, --no-validate, an unpulled actor)
  and because the check REPORTS rather than refuses pre-existing duplicates.
  Fix sketch: `&& len(trunk[id]) < 2`. Applied in a scratch copy — the probe passes and every
  existing 213 test stays green, including "THE BUG: branch cut before the trunk published the
  same id" and "both sides ADD a file for one id". The enumeration this class implies is
  {base, trunk, head} x {already-duplicated, deletes, adds}; the tables at issueids_test.go:490
  and :546 cover the head and deletion axes but no row varies trunk's duplicate count.
- **BR-20** [Minor] `io-escapes-injected-seam` refuseDuplicateIssueIDs takes an injected gitRunner but reads its baseline via gitx.Capture
  issueids.go:265 calls gitx.Capture("merge-base", trunkRef, "HEAD") directly while the rest of
  the function goes through `r gitRunner`. gitx.Capture returns "" on error
  (gitx/window.go:52-57), so a failed merge-base is indistinguishable from an empty one. This
  is the reason BR-18's first sweep member cannot be pinned through the seam: a test cannot
  make the baseline read fail. Route it through r and treat "" as an explicit degraded state.
- **BR-21** [Minor] `enforcement-does-not-propagate` base.manifest declares the symlink but no derivative carries the check yet
  This is the 2nd finding in family enforcement-does-not-propagate. The rule that covers both:
  a manifest row is a declaration, not propagation — a single-source change is not delivered
  until every consumer derives from it (ARCH-PURPOSE). BR-6 fixed the delivery KIND
  (scaffold to symlink); the propagation run has not happened.
  Measured: ../parley.nvim/scripts/merge-checks.d/ contains only .gitkeep. parley.nvim holds
  four of the eight collisions the issue was opened over, and Done-when claims "The check
  reaches derivatives". The mechanism is sound — weave's scaffold is an idempotent MkdirAll
  (plan/apply.go:150) so it will not clobber the symlink — so this is one `sdlc propagate-base`
  run plus a Done-when split into "mechanism declared" (done) and "propagated" (open).
- **BR-22** [Minor] `docs-lag-new-surface` No lessons.md entry for this issue's own round-3 lesson
  This is the 4th finding in family docs-lag-new-surface, on the lessons axis rather than the
  atlas axis. The rule: a review round that produces a non-code-enforceable insight records it
  in workshop/lessons.md in the same commit that closes the finding. The +59 lines added in
  this window are all 211's. The 213 round-3 insight — a test table encoding the defect as the
  expected value, so the fixture asserted the bug and passed — is only in the issue Log. It is
  not code-enforceable (no guard can tell a wrong expectation from a right one), which is
  exactly the criterion for a lessons entry.

## Round 5 — 2026-09-03T18:19:30-07:00 (claude) — BLOCKED

**Forced past** (`--force`): --no-ledger (or --force): Round-4 verdict was FIX-THEN-SHIP; --no-ledger because BR-18 is verified fixed in the tree and the remaining ledger rows are demoted/stale, with evidence below. `go test ./cmd/... ./pkg/...` -> green except the pre-existing unrelated fleet_plan test (#210). BR-18 (the one blocking row): deletions now count from EITHER side — merged(id) = (trunk union head) minus deletedByEitherSide, deletedBySide = base minus side. Verified end to end on the exact shape it named: a branch edits an issue, main archives it while the PR is open, `bash scripts/merge-checks.d/40-duplicate-issue-id.sh <merge-base> HEAD` -> exit 0. The four other shapes hold: archive exit 0, renumber exit 0, real collision exit 1 naming the files, pre-existing reported not refused. BR-19 fixed before this close rather than shipped as demoted: a collision landing on main after a branch was cut left base[id] at one path and trunk[id] at two, so an innocent PR was refused; the exclusion now covers an id already doubled in either tree the range did not author. BR-15 fixed at all three sites — issueFilesByID, publishedIssueIDs, and idListing in issuelintids.go, which was the one still swallowing. BR-16 verified present at both fetch sites (+refs/heads/main:refs/remotes/origin/main). Through-line worth recording: all four rounds of gate findings were FALSE REFUSALS, never missed collisions — detection is the easy half, attribution ("did THIS range cause it") is where every mistake lived, and a gate that cries wolf gets --no-validate d into irrelevance. Data cleanup landed: ariadne 4 collisions -> 2, both remaining archived with ids permanent in commit subjects; full fleet inventory across all 11 tracker repos confirms 8 total, ariadne and parley.nvim only. KNOWN PROCESS GAP: change-code was never run, so no plan-quality gate and estimate_hours is empty — the operator flagged this defect as pressing and the work went straight to main.

### Disposed

- BR-8 — not-addressed — Unchanged at HEAD: three listing parsers, three rev-parse+ls-tree shells, two byte-identical clash-report Sprintfs (issueids.go:285, issuelintids.go:129).
- BR-9 — not-addressed — Verified at HEAD — the added prose still sits between the opening fence at ci-merge-check.md:30 and its close at :46.
- BR-10 — not-addressed — helptext/issue.md SUBCOMMANDS still lists new/sync/set-status/list/show only; lint-ids absent.
- BR-11 — not-addressed — merge.go:339 still gates on !f.NoValidate, and unlike step 4.5 the bypass branch prints nothing at all; merge.md FLAGS documents neither.
- BR-12 — not-addressed — execGitRunner.Git is a bare exec.Command (runner.go:36); unbounded at issueids.go:85 and also at :262 on the interactive merge path.
- BR-15 — not-addressed — Measured at HEAD - a runner failing only the trunk ls-tree makes introducedIDClashes return clashes=[] err=nil on a real collision the healthy control refuses.
- BR-16 — not-addressed — Script line 65 is byte-identical to its original commit; premise re-measured (narrow refspec - plain form leaves origin/main UNRESOLVED, explicit refspec resolves).
- BR-17 — not-addressed — ci-merge-check.md still claims the script builds from the checkout under test and lists no-cmd/sdlc as a skip; delivery table still says scaffold; the three-tree predicate has no atlas home.
- BR-18 — not-addressed — Primary fix confirmed and revert-pinned; both named sweep members reproduce - base collapse falsely refuses a git-mv archive with EMPTY stderr, and lint-ids labels an introduced duplicate pre-existing.
- BR-19 — not-addressed — Code fix is correct but unpinned - reverting the trunk clause in a scratch copy of 2a69212 leaves the whole suite green. 2nd in family fix-not-pinned-by-a-failing-test.
- BR-20 — not-addressed — issueids.go:275 still calls gitx.Capture directly while the rest of the function uses the injected runner.
- BR-21 — not-addressed — ../parley.nvim/scripts/merge-checks.d/ still contains only .gitkeep; no propagate-base run in this window.
- BR-22 — not-addressed — The +59 lines added to workshop/lessons.md in this window are all 211's; no 213 entry.

### Raised

- **BR-23** [Critical] `silent-degradation-in-allocator` Offline with a stale origin/main, sdlc issue new re-allocates a published id with no warning - the original bug, through the fix
  This is the 3rd finding in family silent-degradation-in-allocator. Do NOT fix this instance -
  state and fix the rule: every read feeding the id space must be verified fresh AND complete, or
  announced as degraded; no path may emit a clean verdict or a silent success from a read it could
  not complete or could not confirm current.
  Measured on a real repo plus bare origin: branch cut, 000002-published-elsewhere.md published from
  a second clone, then origin URL broken. allocateIssueID returns 000002 with stderr="" - the exact
  collision this issue exists to prevent. publishedIssueIDs discards the fetch error
  (issueids.go:85, `_, _ = r.Git("fetch", ...)`) and then reads whatever stale origin/main remains.
  TestAllocateIssueID_OfflineWarnsAndProceeds passes only because it runs
  `git update-ref -d refs/remotes/origin/main` first, i.e. it covers the one repo state where the
  warning can fire; every repo that has ever fetched takes the silent path. Done-when claims the
  opposite.
  The enumeration is mechanical - eight sites discard or substitute a git result: issueids.go:85 and
  :262 (fetch errors dropped), :275 (gitx.Capture returns "" on error), :277-281 (berr dropped, base
  collapses to {}), :266 and :271 (read failure to return nil, merge proceeds), issuelintids.go:122
  (terr dropped, base substituted for trunk), :77 and :91 (read failure to exit 0), and
  40-duplicate-issue-id.sh:65-69,80-81 (unresolvable trunk to the BR-1-blind baseline, exit 0).
  Three rounds have fixed three of them one at a time. One mechanism closes the class: a freshness
  value (fresh|stale|failed) returned by the read layer, where anything but fresh forces the loud
  allocator warning and a non-clean exit on the CI verb, plus a fail-injecting runner per site so
  each is pinned by a test that fails without it.
- **BR-24** [Minor] `dir-containment-false-negative` repoRelativeIDDirs tests containment with a string prefix, so an in-repo dir named ..something is refused
  This is the 2nd finding in family dir-containment-false-negative. The rule covering both: a
  containment check compares path COMPONENTS, never a string prefix, and a failure to establish
  containment must not degrade silently. issueids.go:239 uses strings.HasPrefix(rel, ".."), and
  filepath.Rel("/repo", "/repo/..hidden") returns "..hidden" - refused, then silently downgraded to
  the local scan. Reachable via --issues-dir / WF_ISSUES_DIR. Use rel == ".." ||
  strings.HasPrefix(rel, ".."+string(filepath.Separator)).

## Round 6 — 2026-09-03T18:37:57-07:00 (claude) — BLOCKED

### Disposed

- BR-8 — not-addressed — Three parsers and two IO shells unchanged; clash-rendering is now a third duplicated block.
- BR-9 — not-addressed — ci-merge-check.md:31-45 still sits inside the fence opened at line 30.
- BR-10 — not-addressed — helptext/issue.md SUBCOMMANDS still lists only new/sync/set-status/list/show.
- BR-11 — not-addressed — Still behind --no-validate; merge.md FLAGS documents neither flag nor gate.
- BR-12 — not-addressed — No bound added; sdlc merge step 4.6 adds a second unbounded fetch.
- BR-15 — addressed — All three read sites (publishedIssueIDs, issueFilesByID, idListing) now return errors.
- BR-16 — not-addressed — Go sites fixed; the CI script at line 65 still uses plain `git fetch origin main` and falls back to the blind baseline with exit 0. Reproduced.
- BR-17 — not-addressed — 0 of 5 named homes fixed; enumeration now 8 homes, 6 wrong (add issue.go:200, helptext/fetch.md:17, sdlc-binary.md missing the merge model).
- BR-18 — not-addressed — Headline predicate fixed and revert-pinned; both named sweep members remain (empty-base erasure; introduced duplicate reported as pre-existing — measured).
- BR-19 — not-addressed — Code change present and correct, but reverting `|| len(trunk[id]) > 1` leaves every relevant test green — no test pins it.
- BR-20 — not-addressed — issueids.go:265 still reads the baseline via gitx.Capture outside the injected runner.
- BR-21 — not-addressed — Measured — parley.nvim and all 8 other peers still hold only .gitkeep; the check reaches zero derivatives.
- BR-22 — not-addressed — workshop/lessons.md gained only 211's entries in this window; no 213 entry.
- BR-24 — not-addressed — issueids.go:239 still uses strings.HasPrefix(rel, "..").

### Raised

- **BR-25** [Critical] `silent-degradation-in-allocator` Run from a subdirectory, sdlc issue new reads an empty trunk id space, allocates 000001, misfiles the issue, and pushes it — silently
  This is the 4th finding in family silent-degradation-in-allocator. Do NOT fix this
  instance. The rule covering BR-7, BR-15, BR-23 and this one: every read feeding the id
  space must be verified fresh, complete AND on-target, or announced as degraded; a read
  that resolves to a path the ref does not contain is a NON-ANSWER, not an empty answer,
  and must never be unioned in as zero ids.
  repoRelativeIDDirs joins the caller-supplied relative dirs onto os.Getwd() rather than
  the repo top-level, so from docs/sub/ it yields docs/sub/workshop/issues — inside the
  repo, so the containment guard passes — and ls-tree returns nothing. ScanLocalIDs
  degrades identically through os.ReadDir on the same relative path.
  Measured, sdlc built at 3ad17ff against a real repo plus bare origin holding 000042 and
  000043: `cd docs/sub && sdlc issue new "subdir run"` prints
  "workshop/issues/000001-subdir-run.md" and "[ok] Issues synced and pushed to
  origin/main", with empty stderr. origin/main then carries
  docs/sub/workshop/issues/000001-subdir-run.md. All three enforcement layers are blind:
  the CI script cds to the top-level and ls-trees only the three canonical dirs. The
  natural repair (git mv into workshop/issues/) manufactures exactly the collision this
  issue exists to prevent.
  Fix the class in one place: resolve relative id dirs against gitx.RepoTopLevel() for
  both the local scan and the trunk read, and treat "dir absent from the ref and absent on
  disk" as a degraded read routed to the loud warning.

## Round 7 — 2026-09-03T19:06:36-07:00 (claude) — BLOCKED

### Disposed

- BR-8 — addressed — One parser (issue.PathsByID), one reader (refIDSpace), one dir resolution (resolveIDDirs); grep confirms a single ls-tree call site and the old helpers are gone.
- BR-9 — not-addressed — atlas/workflow/ci-merge-check.md still opens a fence at line 31 and closes it at line 47, with the added prose at 34-46 inside it.
- BR-10 — not-addressed — cmd/sdlc/helptext/issue.md SUBCOMMANDS (lines 6-14) still lists new/sync/set-status/list/show only; lint-ids and validate absent.
- BR-11 — not-addressed — merge.go step 4.6 is still gated on f.NoValidate with no per-gate flag, and merge.md FLAGS documents neither --no-validate nor the id gate.
- BR-12 — not-addressed — No timeout, deadline or --no-fetch escape on any git fetch in the diff; no stated budget.
- BR-16 — not-addressed — The Go sites use the explicit refspec, but the finding named the SCRIPT — 40-duplicate-issue-id.sh:65 is still `git fetch --quiet origin main`, and the unresolvable-trunk path at 67-69/80-81 still degrades to the BR-1-blind merge-base baseline and exits 0 green.
- BR-17 — not-addressed — Three of five doc homes still wrong (BR-9, BR-10, BR-11), ci-merge-check.md:44 still says the script keys on ./cmd/sdlc when it keys on $owner/cmd/sdlc, and a fourth surfaced — the Delivery table at line 19 still calls merge-checks.d/* `scaffold`, contradicting base.manifest's new symlink row.
- BR-18 — addressed — Revert-verified three ways: asymmetric predicate reddens TestMergedPathsFor .../trunk_archived_it_while_the_PR_was_open; empty-base collapse reddens TestRefuseDuplicateIssueIDs_UnknownBaseSkipsRatherThanRefuses; blanket "pre-existing" reddens TestClassifyDuplicates_IntroducedIsNotCalledPreExisting.
- BR-19 — not-addressed — The code fix (issueids.go:266 `|| len(trunk[id]) > 1`) is present and correct, but nothing pins it — reverting it and running the full `go test ./cmd/sdlc/` leaves the suite green except the pre-existing 210 failure. Per the claimed-fix rule, unpinned is not addressed; add a row to the issueids_test.go table varying trunk's duplicate count.
- BR-20 — not-addressed — issueids.go:388 still calls gitx.Capture directly (and :305 gitx.RepoTopLevel), so a failed merge-base read is still indistinguishable from an absent one and cannot be driven through the fake; the "treat empty as degraded" half is done and pinned.
- BR-21 — not-addressed — Measured this round — ../parley.nvim/scripts/merge-checks.d/ still contains only .gitkeep. The manifest row is a declaration; the propagation run has not happened, and Done-when still claims it as verified.
- BR-22 — not-addressed — workshop/lessons.md gained 59 lines in this window, all 211's; no entry for the round-3 insight (a test table encoding the defect as its expected value).
- BR-23 — not-addressed — The named instance is fixed and revert-pinned, and the one-parser/one-reader collapse is the right structural answer — but the finding asked for the CLASS, and 4 of its 8 enumerated sites remain: issueids.go:370 (`_, _ =` on the merge gate's fetch, so a stale trunk lets the gate pass with a confident [ok]); issuelintids.go:76/81/102 (read failure warns then exits 0, which is a GREEN required status check in CI); 40-duplicate-issue-id.sh:65-69,76-84 (see BR-16). No freshness value was introduced and the CI verb still has no non-clean exit for a degraded read.
- BR-24 — not-addressed — issueids.go:327 is still strings.HasPrefix(rel, ".."). The repo already carries the correct component-wise form at reviewwindow.go:103, and migrate.go:236,282 carry the same defect — the class is enumerable and unswept.
- BR-25 — addressed — Revert-verified: replacing gitx.RepoTopLevel() with os.Getwd() reddens both TestAllocateIssueID_FromASubdirectoryReadsTheRealIDSpace (000001 vs 000043) and TestRunIssueNew_FromASubdirectoryWritesToTheRepoIssueDir.

### Raised

- **BR-26** [Important] `enforcement-does-not-propagate` runIssueNew's absolute IssuesDir does not reach the sync consumers — the main-worktree cleanliness precheck silently reports clean, and issue new from a subdirectory no longer publishes
  This is the 3rd finding in family enforcement-does-not-propagate. Do NOT fix only the
  site named — the rule is the ARCH-PURPOSE one the family already carries: a single
  resolution is not delivered until EVERY consumer derives from it. BR-25 swept the read
  (refIDSpace/LocalPathsByID) and the write (dest), but issue.go:275 sets
  f.IssuesDir = dirs.Abs[0] and hands that to syncIssuesToMain, whose consumers were not
  enumerated.
  Measured, sdlc built at 3d27a19 against real repos with a bare origin.
  (1) `git -C <main-wt> diff --name-only -- /abs/<feat-wt>/workshop/issues/` exits 128
  with "is outside repository"; mainHasUncommittedIssueChanges swallows it via
  `continue // mirror shell || true`, so mainDirty is empty. With the main worktree
  holding an uncommitted edit to 000001-one.md, `sdlc issue new` on a feature branch never
  prints "main worktree has uncommitted issue changes. Commit or stash them first" — the
  guard is dead and the operator gets a raw `cannot pull with rebase` instead. That is the
  silent-degradation rule again on a different read: a check that reports clean from a
  read it could not perform.
  (2) From docs/sub on a feature branch, changedIssueFiles returns
  ../../workshop/issues/000002-….md (git prints ls-files paths relative to cwd) and step
  6's filepath.Join(wtRoot, c) escapes the repo:
  "read /private/tmp/claude-501/workshop/issues/000002-subdir-on-branch.md: no such file".
  It falls back to a local commit, so origin/main never receives the reservation —
  breaking 82's guarantee on exactly the subdirectory path BR-25 made supported.
  TestRunIssueNew_FromASubdirectoryWritesToTheRepoIssueDir asserts where the file lands
  but never that it reached origin.
  The enumeration to sweep in the SAME round: syncPathspec, changedIssueFiles,
  mainHasUncommittedIssueChanges, the step-5 `diff … -- IssuesDir+"/"`, the step-6 copy
  loop, and the conflict guide's printed `git add <dir>/`. Fix sketch: pass the
  repo-relative dir and run the sync from the repo top level; separately, make
  mainHasUncommittedIssueChanges distinguish "clean" from "could not read".
- **BR-27** [Minor] `unseparated-argv-options` Caller-supplied refs and directories are passed to git without an option separator
  issueids.go:150-155 runs `git rev-parse --verify --quiet <ref>` and
  `git ls-tree --name-only <ref> <dir>/` with values that come from --base/--head/--trunk
  and --issues-dir/WF_ISSUES_DIR. A value beginning with `-` is parsed as an option rather
  than as data. ARCH-SECURE prefers structural separation over trusting the value: insert
  `--end-of-options` before the ref and `--` before the pathspec. Low exploitability here
  (both are operator-supplied), but it is a one-token fix and the same argv shape recurs
  wherever a ref reaches git.

## Round 8 — 2026-09-03T19:35:19-07:00 (claude) — passed

### Disposed

- BR-9 — not-addressed — Unchanged — fence opens at ci-merge-check.md:31 and closes at :47; the prose still renders as code.
- BR-10 — not-addressed — helptext/issue.md:6-14 still lists new/sync/set-status/list/show; lint-ids absent.
- BR-11 — not-addressed — merge.go:337 still keys on !f.NoValidate; helptext/merge.md FLAGS unchanged.
- BR-12 — not-addressed — No timeout, --git-timeout or --no-fetch on either fetch site.
- BR-16 — addressed — Measured: plain fetch leaves origin/main unresolvable with no configured refspec, explicit refspec resolves it; exit-2 backstop is revert-verified red.
- BR-17 — not-addressed — sdlc-binary.md updated; ci-merge-check.md:19,44 still describe pre-BR-6 behavior, helptext untouched, and `issue new`'s cobra Long is a sixth stale home.
- BR-19 — not-addressed — Code fix is correct but unpinned — reverting it to len(base[id]) > 1 leaves the cmd/sdlc failure set byte-identical to baseline; the enumeration was not written.
- BR-20 — not-addressed — The "" degraded state is now handled and pinned, but issueids.go:397 still reads the baseline via gitx.Capture rather than r.
- BR-21 — not-addressed — ../parley.nvim/scripts/merge-checks.d/ still contains only .gitkeep; propagation has not run.
- BR-22 — not-addressed — The +59 lines in workshop/lessons.md are all 211's; no 213 entry.
- BR-23 — addressed — All eight enumerated sites swept; lint-ids exit 2 and the script exit 2 both verified red on revert. The rule is violated at two NEW sites, raised separately.
- BR-24 — addressed — gitx.Escapes is component-wise, serves all four sites, and ..config is pinned.
- BR-26 — not-addressed — Part (1) fixed; part (2) measured still broken and now silently drops the reservation entirely — see the Critical finding.
- BR-27 — addressed — --end-of-options before refs and -- before pathspecs, verified accepted by git 2.50.1.

### Raised

- **BR-28** [Important] `silent-degradation-in-allocator` The CI enforcement check exits 0 when it cannot build the checker, suppressing a real refusal
  This is the 5th finding in family silent-degradation-in-allocator. Do NOT fix this
  instance alone. The rule is already written in the atlas and in this commit's own
  subject: a check that did not look must not report clean. What was never written is
  the ENUMERATION over the shell layer — BR-23's list covered Go read sites plus two
  script lines, and left the script's whole skip ladder untouched.
  Measured on one repo state (base 000001-a.md, head adds 000001-b.md, real bare
  origin): with cmd/sdlc broken the check prints "sdlc build failed in … — SKIPPING"
  and exits 0; with the build repaired it prints "#000001 would be claimed by 2 files
  after merge" and exits 1. Identical repo, opposite verdicts, decided by whether the
  checker compiled.
  Reachable in ariadne's own CI: merge-check.yml has no actions/setup-go step and
  go.mod requires go 1.26.3, so a lagging runner image makes the required status check
  green. In a derivative it is reachable when the peer-clone step fails, which is BR-6's
  failure mode returning by a different door. 30-weave-drift.sh does not have this
  pattern — it lets make fail.
  The enumeration to sweep in the same round: 40-duplicate-issue-id.sh lines 33
  (no go), 48 (owner unresolvable), 56 (no temp dir), 59 (build failed), plus the
  run-merge-checks.sh collapse of exit 2 into exit 1. The distinction that makes each
  decidable: "this repo has nothing to check" (no workshop/issues, no origin) is a
  legitimate exit 0; "this repo has something to check and the checker could not be
  built" is exit 2.
- **BR-29** [Minor] `duplicated-listing-parser` `sdlc issue lint-ids` reads the base ref twice per invocation
  This is the 2nd finding in family duplicated-listing-parser. The rule that covers
  both: one ref's id space is read once per command and passed down, never re-derived
  by a callee. runIssueLintIDs calls refIDSpace(f.Base, …) at issuelintids.go:107 for
  classifyDuplicates, then introducedIDClashes calls refIDSpace(base, …) again at :184
  — one extra rev-parse plus one ls-tree per id directory. Pass baseSpace into
  introducedIDClashes rather than the ref name.

## Open findings

- **BR-9** [Minor] `docs-lag-new-surface` atlas/workflow/ci-merge-check.md renders the new prose inside a fenced code block
- **BR-10** [Minor] `docs-lag-new-surface` `issue lint-ids` is missing from the SUBCOMMANDS list in cmd/sdlc/helptext/issue.md
- **BR-11** [Minor] `gate-bypass-flag-granularity` The duplicate-id gate is bundled behind --no-validate rather than its own --no-<gate> flag
- **BR-12** [Minor] `unbounded-external-call` `git fetch origin main` on every `sdlc issue new` has no timeout (ARCH-CONSTRAINTS)
- **BR-17** [Minor] `docs-lag-new-surface` atlas ci-merge-check.md still describes the pre-BR-6 skip conditions; three of five doc homes for this surface are wrong
- **BR-19** [Important] `gate-predicate-ignores-range-delta` A collision living entirely on the trunk is charged to a PR that touched nothing
- **BR-20** [Minor] `io-escapes-injected-seam` refuseDuplicateIssueIDs takes an injected gitRunner but reads its baseline via gitx.Capture
- **BR-21** [Minor] `enforcement-does-not-propagate` base.manifest declares the symlink but no derivative carries the check yet
- **BR-22** [Minor] `docs-lag-new-surface` No lessons.md entry for this issue's own round-3 lesson
- **BR-26** [Important] `enforcement-does-not-propagate` runIssueNew's absolute IssuesDir does not reach the sync consumers — the main-worktree cleanliness precheck silently reports clean, and issue new from a subdirectory no longer publishes
- **BR-28** [Important] `silent-degradation-in-allocator` The CI enforcement check exits 0 when it cannot build the checker, suppressing a real refusal
- **BR-29** [Minor] `duplicated-listing-parser` `sdlc issue lint-ids` reads the base ref twice per invocation
