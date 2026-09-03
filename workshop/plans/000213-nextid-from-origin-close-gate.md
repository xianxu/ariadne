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

## Open findings

- **BR-1** [Critical] `gate-compares-wrong-baseline` The CI check compares against merge-base, so it structurally cannot see the collision this issue exists to catch
- **BR-2** [Critical] `gate-compares-wrong-baseline` issueFilesByID keeps only the first path per id, so an introduced duplicate is detected or missed by slug sort order
- **BR-3** [Important] `dir-containment-false-negative` repoRelativeIDDirs rejects a not-yet-created id dir whenever the repo root is reached through a symlink, silently disabling every layer
- **BR-4** [Important] `gate-compares-wrong-baseline` The merge gate at step 4.6 reads a stale origin/main — merge.go's own comment says the flow has not fetched yet
- **BR-5** [Important] `fix-not-pinned-by-a-failing-test` No test exercises the CI check's refusal path; the one that claims to asserts the SKIP path
- **BR-6** [Important] `enforcement-does-not-propagate` The CI check does not reach derivatives — merge-checks.d is a scaffold row, and the script self-skips without ./cmd/sdlc
- **BR-7** [Important] `silent-degradation-in-allocator` publishedIssueIDs swallows per-directory ls-tree failures, so a partial trunk read allocates a colliding id with no warning
- **BR-8** [Minor] `duplicated-listing-parser` Three near-identical ls-tree listing parsers and two identical rev-parse+ls-tree IO shells (ARCH-DRY)
- **BR-9** [Minor] `docs-lag-new-surface` atlas/workflow/ci-merge-check.md renders the new prose inside a fenced code block
- **BR-10** [Minor] `docs-lag-new-surface` `issue lint-ids` is missing from the SUBCOMMANDS list in cmd/sdlc/helptext/issue.md
- **BR-11** [Minor] `gate-bypass-flag-granularity` The duplicate-id gate is bundled behind --no-validate rather than its own --no-<gate> flag
- **BR-12** [Minor] `unbounded-external-call` `git fetch origin main` on every `sdlc issue new` has no timeout (ARCH-CONSTRAINTS)
