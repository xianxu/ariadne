# Boundary Review — ariadne#213 (whole-issue close)

| field | value |
|-------|-------|
| issue | 213 — Allocate issue IDs against origin/main |
| repo | ariadne |
| issue file | workshop/issues/000213-nextid-from-origin.md |
| boundary | whole-issue close |
| milestone | — |
| window | 9078f8d0529b0c9fd4312f435b10f731aa5d3dc2..36181f885ddfbc4bd2e00144832ce2dd7442bd9a |
| command | sdlc close --issue 213 |
| reviewer | claude |
| timestamp | 2026-09-03T15:35:36-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

```findings
findings:
  - id: new
    severity: Critical
    family: gate-compares-wrong-baseline
    title: |
      The CI check compares against merge-base, so it structurally cannot see the collision this issue exists to catch
    detail: |
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
  - id: new
    severity: Critical
    family: gate-compares-wrong-baseline
    title: |
      issueFilesByID keeps only the first path per id, so an introduced duplicate is detected or missed by slug sort order
    detail: |
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
  - id: new
    severity: Important
    family: dir-containment-false-negative
    title: |
      repoRelativeIDDirs rejects a not-yet-created id dir whenever the repo root is reached through a symlink, silently disabling every layer
    detail: |
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
  - id: new
    severity: Important
    family: gate-compares-wrong-baseline
    title: |
      The merge gate at step 4.6 reads a stale origin/main — merge.go's own comment says the flow has not fetched yet
    detail: |
      refuseDuplicateIssueIDs is called at cmd/sdlc/merge.go:340, but merge.go:464-465
      states "origin/main is stale here (the flow doesn't pull until AFTER deciding,
      below)" and only fetches at :469. A collision published to the trunk since this
      checkout last fetched is invisible to the gate that documents itself as "the last
      point where an id collision is still repairable". Fix: fetch before 4.6, or move the
      gate below the existing fetch at :469.
  - id: new
    severity: Important
    family: fix-not-pinned-by-a-failing-test
    title: |
      No test exercises the CI check's refusal path; the one that claims to asserts the SKIP path
    detail: |
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
  - id: new
    severity: Important
    family: enforcement-does-not-propagate
    title: |
      The CI check does not reach derivatives — merge-checks.d is a scaffold row, and the script self-skips without ./cmd/sdlc
    detail: |
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
  - id: new
    severity: Important
    family: silent-degradation-in-allocator
    title: |
      publishedIssueIDs swallows per-directory ls-tree failures, so a partial trunk read allocates a colliding id with no warning
    detail: |
      cmd/sdlc/issueids.go:88 `continue`s on ls-tree error. A missing directory is already
      exit 0 with empty output (verified), so a non-nil error means a real git failure —
      and dropping that directory's ids silently under-counts the published space, which is
      precisely the silent-fallback failure the Spec forbids ("a silent fallback here
      recreates the bug it is meant to fix"). Fix: return the error so allocateIssueID
      takes its loud-warning path, or warn per directory.
  - id: new
    severity: Minor
    family: duplicated-listing-parser
    title: |
      Three near-identical ls-tree listing parsers and two identical rev-parse+ls-tree IO shells (ARCH-DRY)
    detail: |
      issue.IDsInTreeListing, issue.DuplicateIDsInRef and issueFilesByID each re-implement
      split → trim → LastIndex("/") → IDFromFilename; publishedIssueIDs and idListing each
      re-implement rev-parse + repoRelativeIDDirs + per-dir ls-tree. The duplication is not
      cosmetic: the collapsing map in issueFilesByID exists only because it re-implements
      instead of reusing the id→paths structure DuplicateIDsInRef already builds. One pure
      PathsByIDInTreeListing in the issue package fixes the DRY violation and the
      order-dependence Critical together.
  - id: new
    severity: Minor
    family: docs-lag-new-surface
    title: |
      atlas/workflow/ci-merge-check.md renders the new prose inside a fenced code block
    detail: |
      The paragraphs added at lines 31-45 sit between the opening ``` at line 30 and its
      closing fence at line 46, so they render as literal code.
  - id: new
    severity: Minor
    family: docs-lag-new-surface
    title: |
      `issue lint-ids` is missing from the SUBCOMMANDS list in cmd/sdlc/helptext/issue.md
    detail: |
      The list at issue.md:6-14 enumerates new/sync/set-status/list/show; the new verb is
      not there. (`validate` is also absent — pre-existing.)
  - id: new
    severity: Minor
    family: gate-bypass-flag-granularity
    title: |
      The duplicate-id gate is bundled behind --no-validate rather than its own --no-<gate> flag
    detail: |
      AGENTS.md §5 asks for a per-gate flag so a bypass is an explicit acknowledgment of the
      specific gate. Today skipping #124's instance-conformance gate also silently skips the
      id gate, and merge.md's FLAGS block documents neither.
  - id: new
    severity: Minor
    family: unbounded-external-call
    title: |
      `git fetch origin main` on every `sdlc issue new` has no timeout (ARCH-CONSTRAINTS)
    detail: |
      allocateIssueID's fetch is best-effort but unbounded. A host that drops packets rather
      than refusing (VPN down, sleeping laptop) blocks an interactive verb for the TCP
      connect timeout. Consider a short --git-timeout / GIT_HTTP_LOW_SPEED_* bound, or an
      explicit --no-fetch escape, and state the budget in the issue.
```

# Review — ariadne#213, whole-issue close

The allocation half is genuinely good work: `NextID` is properly pure, the union semantics are right, and every allocation test drives a real repo against a real bare origin — an honest ARCH-MOCK setup that a function-call fake could not have expressed. The enforcement half does not hold up. I ran the shipped artifacts rather than reading them, and the CI check — the layer the issue itself calls "the enforceable one" — passes the exact scenario the issue was written to catch: with `base = merge-base` (what `merge-check.yml` passes), a branch cut before the colliding id landed on main has no id at base, so nothing is reported. Separately, detection of an introduced duplicate flips on slug alphabetical order, because `issueFilesByID` keeps only the first path per id — a hazard the Spec names in prose and then leaves in the comparison. Both holes are invisible to the suite because the only test of the CI script asserts its *skip* path while being named `...RefusesAPlantedCollision`. That combination — an enforcement gate that fails open on its motivating case, with no test that would go red — is what blocks SHIP.

## 1. Strengths

- `cmd/sdlc/internal/issue/scaffold.go:52` — `NextID(idSets ...[]int)` is a clean ARCH-PURE split: the decision takes id sets, the directory read is `ScanLocalIDs`, and the trunk read is the caller's. The table test at `issueids_test.go:200` runs with no IO at all.
- `issueids_test.go:56` `TestAllocateIssueID_BranchCutBeforePublish` is the right fixture and the revert-verification claim checks out: it cuts the branch first, publishes #2 through a *second clone of the bare origin*, and asserts the fixture's own premise (`t.Fatal("fixture broken: the branch should not contain #2")`). That is ARCH-MOCK done properly.
- `issueids_test.go:111` asserts the offline *warning text*, not merely that creation succeeded — the deliverable the Spec actually named.
- `repoRelativeIDDirs` (`issueids.go:99`) is a real insight found by the tests, not by review: git runs against cwd while the dirs are caller-supplied. The containment check is the right idea (see the Important finding on its symlink edge, which is an implementation slip, not a design error).
- `issue.DuplicateIDsInRef` correctly treats "the same path twice" as one claimant (`scaffold.go:139`), with a test pinning it — the failure mode that would have flagged every clean repo.
- Verified green: `go test ./cmd/sdlc/...` passes except the pre-existing `TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory` (missing `workshop/plans/000200-*-plan.md`, i.e. #210) — matching the Log's qualifier.

## 2. Critical

**C1 — CI check compares against the wrong baseline** (`scripts/merge-checks.d/40-duplicate-issue-id.sh:56`, `.github/workflows/merge-check.yml`). Evidence, scratch repo: cut branch at S; main publishes `000500-theirs.md`; branch adds `000500-mine.md`.
```
lint-ids --base <merge-base>  → [ok] this range introduces no reused issue ids   EXIT=0
lint-ids --base <main tip>    → this range reuses 1 issue id(s) ... EXIT=1
```
The runner contract fixes base at merge-base, so the script must resolve the trunk tip itself (`git fetch -q origin main` + `git rev-parse origin/main`, falling back to `$1`).

**C2 — order-dependent duplicate detection** (`cmd/sdlc/issueids.go:190`). Evidence on the real repo at HEAD, same base, two plants differing only in slug:
```
000213-aaa-collision.md     → refused, EXIT=1
000213-planted-collision.md → [ok] ... no reused issue ids, EXIT=0
```
And a scratch test against the merge gate: a branch adding `000001-zzz-later-slug.md` alongside the trunk's `000001-first.md` gets `[ok] duplicate-id gate: no reused issue ids`. Fix by carrying all paths per id (one shared pure helper with `DuplicateIDsInRef`) and flagging any head path an id doesn't own at base.

## 3. Important

See the findings block: symlink false-negative in `repoRelativeIDDirs` (silently disables all three layers on a fresh repo), stale `origin/main` at merge step 4.6, no test pinning the refusal path, non-propagation to derivatives contradicting Done-when + atlas, and swallowed `ls-tree` errors in the allocator.

## 4. Minor

Duplicated listing parsers (ARCH-DRY); atlas prose inside a code fence; `lint-ids` missing from `helptext/issue.md`; gate bundled under `--no-validate`; unbounded fetch on `issue new`.

## 5. Test coverage notes

Allocation is well covered (branch-cut, union, offline-warning, history-on-trunk). Detection is not: every duplicate-id test happens to pick a slug that sorts before the incumbent, so C2 is invisible; no test drives `base = merge-base`, so C1 is invisible; no test reaches the script's refusal path, so the adapter's exit code is unpinned (I confirmed by hand, with a stubbed `mktemp`, that the adapter *does* propagate exit 1 — the plumbing is fine, the semantics are not). Minimum additions: a slug-order pair (`-aaa-` and `-zzz-`) for both `introducedIDClashes` and `refuseDuplicateIssueIDs`; a merge-base-shaped range test; a script test in a repo where the build can run.

## 6. Architectural notes

- **ARCH-DRY** — flag. Three listing parsers, two IO shells; the collapsing map that causes C2 exists only because `issueFilesByID` re-implements rather than reuses.
- **ARCH-PURE** — mostly pass. `NextID`/`IDsInTreeListing`/`DuplicateIDsInRef` are clean. Flag: `introducedIDClashes` fuses comparison, formatting and git IO, and `runIssueLintIDs` hard-wires `claimRunner` and exits via `os.Exit` — which is why the decision that is wrong (C1/C2) has no unit test.
- **ARCH-PURPOSE** — flag. The purpose is a guaranteed id space; delivered is a gate that misses its own motivating case (C1), a detector that coin-flips on slug order (C2), and enforcement that stops at ariadne's boundary while the Done-when claims fleet coverage. The Spec identified the collapsing-map class and only the report instance was fixed — the instance, not the class.
- **ARCH-MOCK** — pass for allocation (real repo + real bare origin throughout, correctly argued). Flag for the CI layer: production flow (build + refuse) and test flow (skip) do not share the boundary.
- **ARCH-CONSTRAINTS** — flag (minor). Unbounded `git fetch` on an interactive verb; `go build ./cmd/sdlc` per PR is acceptable.
- **ARCH-SECURE** — mostly pass. Input is parsed with a strict `^(\d{6})-` regex and unmatched names are dropped; no credentials touched; tests write only to temp dirs. Two notes: `ls-tree --name-only` C-quotes non-ASCII paths (harmless for the id prefix, but `-z` would be structurally safer), and the gate builds and runs the PR's own code to judge that PR — bounded here since `pull_request` grants no secrets, but worth naming if fork PRs ever land.

## 7. Plan revision recommendations

There is no `workshop/plans/000213-*-plan.md` and no Core-concepts table, so this applies to the issue file's `## Done when`. Add a `## Revisions` entry recording:

1. **Done-when "sdlc merge refuses a branch introducing an id that already exists on the trunk, naming both paths"** — qualify: as implemented this holds only when the branch does not also contain the trunk's file, and only against a possibly-stale `origin/main`.
2. **Done-when "scripts/merge-checks.d/40-duplicate-issue-id.sh performs the same refusal in CI"** — the check runs against merge-base and therefore does not refuse the branch-cut-before-publish case; state the trunk-tip comparison as the requirement.
3. **Done-when "propagates to every derivative through the symlinked runner"** — false as built (`scaffold scripts/merge-checks.d`, plus the `./cmd/sdlc` self-skip). Either scope it to ariadne or make the propagation real and re-tick.
4. **Plan row "Test the CI check by running it against a real repo with a planted collision"** — untick; the existing test asserts the skip path.
5. Record the workflow deviation already documented in `## Log` (no `change-code`, no plan-quality gate, no estimate) as a Revisions entry so it is part of the record rather than only the narrative.
