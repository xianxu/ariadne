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

---

## Re-review — 2026-09-03T17:00:45-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 213 — Allocate issue IDs against origin/main |
| repo | ariadne |
| issue file | workshop/issues/000213-nextid-from-origin.md |
| boundary | whole-issue close |
| milestone | — |
| window | 9078f8d0529b0c9fd4312f435b10f731aa5d3dc2..41339196eee82ac4a41a9e43bdf7695181f3ca49 |
| command | sdlc close --issue 213 |
| reviewer | claude |
| timestamp | 2026-09-03T17:00:45-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The round-1 Criticals (BR-1, BR-2) are genuinely fixed and I confirmed each by reverting exactly that fix in a scratch clone and watching a named test go red — that is real, and the trunk-tip resolution plus the every-path map are the right shapes. What blocks SHIP is that the gate's *predicate* is still wrong in both directions, and neither direction is a subtlety: it decides "what did this range introduce" by set-differencing two trees' path lists, so (a) moving an issue file to `workshop/history/issues/` — this repo's own documented archive step — is refused as a reused id, and (b) a range that introduces two files claiming one brand-new id is reported as "pre-existing" and passed with the line "this range introduces no reused issue ids", which the Done-when explicitly promises to refuse. I verified (a) against a real merged PR of this repository: replaying `lint-ids` over PR #109's range refuses it and tells the operator to renumber `#000195`. With this check destined to become a required status check, that is a hard block on legitimate merges plus a false-pass on the case the issue's own residual note describes. Four of the twelve prior findings also remain open, one of them (BR-3) both unpinned and still live for absolute issue dirs.

## 1. Strengths

- **ARCH-MOCK is served properly.** `cmd/sdlc/issueids_test.go:16-21` states why, and every test drives a real repo against a real bare origin with the branch cut *before* the second clone publishes. The bug is "a ref exists that this worktree does not contain"; a function-call fake cannot express it, and the tests know that.
- **BR-1's fix is real and pinned.** Reverting `scripts/merge-checks.d/40-duplicate-issue-id.sh:70-76` to `base="$fallback_base"` turns `TestMergeCheckScript_RefusesGivenMergeBase` red with the exact defect in the message (`[ok] id lint: this range introduces no reused issue ids`). The test fixture is built in the collision's defining order — cut, publish, collide — not a convenient one.
- **BR-2's fix is real and pinned.** Reverting `issueFilesByID` to a first-path-only map turns the `000001-zzz-sorts-last.md` subtree of `TestIntroducedIDClashes_IndependentOfSlugSortOrder` red while the `aaa` subtree stays green — i.e. the test is genuinely measuring order-independence, not restating the fix.
- **BR-6's propagation half is verified, not asserted.** `weave compile --dry-run` in `/Users/xianxu/workspace/parley.nvim` now emits `symlink scripts/merge-checks.d/40-duplicate-issue-id.sh -> …/ariadne/…` alongside `mkdir scripts/merge-checks.d`, so the symlink row does survive the scaffold row. The owner resolution via `construct/dev-aliases.sh` is exercised by a fixture that has no `cmd/sdlc` of its own.
- **The atlas correction is honest.** `atlas/workflow/sdlc-binary.md:206-232` retracts the three claims round 1 falsified (propagates / right baseline / independent layers) rather than quietly restating them, and the three-layer table names the local gate as "operator feedback", not enforcement.

## 2. Critical findings

**C1 — A moved or renamed issue file is refused as a reused id (`cmd/sdlc/issueids.go:100-121`, `cmd/sdlc/issuelintids.go:110-127`).**

`newPathsFor` treats any head path for an id that is not byte-identical to a base path as a second claimant. Archiving — `workshop/issues/NNNNNN-x.md` → `workshop/history/issues/NNNNNN-x.md`, the step AGENTS.md §1 mandates on done — changes the path while the id keeps one owner. Replayed against this repo's real PR #109 (`base=008f7e3^1`, `head=008f7e3^2`):

```
this range reuses 1 issue id(s) that already exist at 000f9a39…:
  #000195
      introduced: workshop/history/issues/000195-review-finding-families.md
      already at base: workshop/issues/000195-review-finding-families.md
  … Rename this range's file to a fresh id (and its `id:` frontmatter) before merging.
```

Exit 1. That PR merged cleanly in reality; with this check required it would be blocked, and the remedy it prints would corrupt the tracker. `sdlc merge` step 4.6 shares the predicate and `die`s the same way. Fix sketch below (shared with C2).

**C2 — A within-ref duplicate the range itself introduces is reported as "pre-existing" and passed (`cmd/sdlc/issuelintids.go:77-82`).**

`DuplicateIDsInRef` is run over the head listing and its results are unconditionally labelled pre-existing; `introducedIDClashes` never sees them because `base[id]` is empty, so `newPathsFor` returns `nil`. Measured on a fresh fixture where one branch adds `000500-agent-a.md` and `000500-agent-b.md` (neither on the trunk):

```
[!]  pre-existing duplicate id #000500: workshop/issues/000500-agent-a.md, workshop/issues/000500-agent-b.md
[ok] id lint: this range introduces no reused issue ids
EXIT=0
```

The summary line is false and the exit code is wrong. Done-when says "It REPORTS on the base … and REFUSES when the PR introduces one" — the split is computed by question-type, not by lineage, so the refusal half was never built.

**Shared fix for C1 + C2** — decide "introduced" from the range's delta, not from two trees' path sets:

- `introduced_cross` = head paths for an id absent from base, **minus** paths that appear as rename destinations in `git diff --name-status -M <base> <head> -- <id dirs>`;
- `introduced_within` = ids with ≥2 distinct head paths where at least one path is absent at base;
- refuse on either; report the rest.

Checked against all four shapes: BR-1's cut-then-publish still refuses (head path is new and not a rename dest); the archive passes; the two-new-files case refuses; the eight already-merged collisions still only report. Both `refuseDuplicateIssueIDs` and `introducedIDClashes` should call one shared pure predicate over (base paths-by-id, head paths-by-id, rename map) so they cannot disagree — that also discharges BR-8.

## 3. Important findings

**I1 — BR-7's class was fixed at one of three sites (`cmd/sdlc/issueids.go:220`, `cmd/sdlc/issuelintids.go:143`).** *This is the 2nd finding in family `silent-degradation-in-allocator`.* Do not fix these two sites and stop. **Rule:** *a partial read of the id space must be an error at every site that performs one; no code path may report a clean verdict from an incomplete listing.* The enumeration is mechanical — three functions run per-directory `ls-tree` (`publishedIssueIDs`, `issueFilesByID`, `idListing`); round 1 fixed `publishedIssueIDs` only, and the other two still `continue` on error. In `issueFilesByID` a dropped directory removes ids from the **base** map, which reads as "this id is new" — a false negative in the enforcement gate itself. Compounding it, `runIssueLintIDs` converts every error into `return nil` (exit 0) at three places, so a broken read is a passing CI check. Fix the class: one shared IO shell that returns an error, and let the verb refuse-or-announce rather than pass.

**I2 — `git fetch --quiet origin main` does not guarantee `refs/remotes/origin/main`, so the check can fall back to the baseline BR-1 proved blind (`scripts/merge-checks.d/40-duplicate-issue-id.sh:69-75`).** *This is the 4th finding in family `gate-compares-wrong-baseline`.* Earlier rounds fixed instances. **Rule:** *a gate must resolve its authoritative baseline explicitly and fail loudly when it cannot — never silently substitute a baseline that is structurally blind to the class the gate exists to catch.* Measured: with `remote.origin.fetch` set narrowly (as `actions/checkout` does for some PR configurations), `git fetch origin main` leaves `origin/main` unresolvable, while `git fetch origin '+refs/heads/main:refs/remotes/origin/main'` resolves it. When it fails the script falls back to `$fallback_base` = merge-base, which is exactly BR-1. I did not prove this fires under `fetch-depth: 0`; the one-word explicit-refspec fix removes the dependency on that assumption, and the fallback should be treated as a degraded state (loud, and arguably a non-zero "could not verify") rather than a pass.

## 4. Minor findings

- **M1 — Stale atlas prose about the check's skip conditions (`atlas/workflow/ci-merge-check.md:41-45`).** *This is the 3rd finding in family `docs-lag-new-surface`.* **Rule:** *a change to a user-facing surface updates that surface's three doc homes in the same commit — helptext (`cmd/sdlc/helptext/*.md`), the atlas entry, and README where the verb is listed — and corrects any prose that describes the old behavior.* The enumeration for #213 is: helptext `issue.md` SUBCOMMANDS (missing — BR-10), helptext `merge.md` FLAGS (the gate is undocumented — BR-11), `atlas/workflow/ci-merge-check.md` (broken fence — BR-9 — *and* still says the script keys on `./cmd/sdlc`, the exact thing BR-6 changed), `atlas/workflow/sdlc-binary.md` (done), README (N/A — `issue *` verbs are not listed there). Three of five homes are wrong; sweep the list, don't patch the one named here.
- **M2 —** `sdlc merge` prints no warning when `--no-validate` skips the duplicate-id gate (`cmd/sdlc/merge.go:339`), unlike the #124 gate immediately above it — so the bypass is silent as well as bundled.
- **M3 —** `refuseDuplicateIssueIDs` compares local `HEAD`, but merge is server-side; on a stale local checkout the gate blesses a tree that is not what merges.

## 5. Test coverage notes

- Full suite: green except the known, unrelated `TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory` (#210, archived-plan path) — matches the Log's qualifier.
- **BR-3's fix has no pin.** Reverting `repoRelativeIDDirs` to its pre-fix form (`git show d50a023`) leaves every #213 test green. The behavior change is real — I built both binaries and, under a symlinked repo root with `PWD` set to the link, the reverted one prints `id lint skipped: workshop/history is outside the current repo` while HEAD prints the verdict — but nothing in the suite would catch its removal. A test needs a fixture whose root is reached through a symlink *and* which lacks `workshop/history`.
- **BR-4's fix has no pin.** The fetch at `issueids.go:172` is unconditional and reachable, but `TestRefuseDuplicateIssueIDs` pushes before branching, so `origin/main` is already current and the test passes with or without it. A recording `gitRunner` asserting `fetch` precedes `ls-tree` is a few lines.
- Nothing pins the `base.manifest` symlink row; I verified it by hand with `weave compile --dry-run` in `parley.nvim`. A `weave` golden/completeness assertion on that row would make BR-6 self-defending.
- C1 and C2 are both trivially testable at the `introducedIDClashes` level: one fixture that `git mv`s an issue into `workshop/history/issues/` and asserts zero clashes, one that adds two same-id files on the branch and asserts a refusal.

## 6. Architectural notes

- **ARCH-DRY — flag.** BR-8 stands unchanged: `IDsInTreeListing`, `DuplicateIDsInRef` and `issueFilesByID` each re-implement split → trim → `LastIndex("/")` → `IDFromFilename`, and `publishedIssueIDs`/`issueFilesByID`/`idListing` each re-implement rev-parse + `repoRelativeIDDirs` + per-directory `ls-tree`. The cost is no longer cosmetic: the three shells now have *divergent* error handling (one returns, two swallow), which is I1. One pure `PathsByIDInTreeListing` plus one IO shell collapses BR-8, I1 and the C1/C2 predicate together.
- **ARCH-PURE — partial flag.** `NextID`, `IDsInTreeListing`, `DuplicateIDsInRef` and `newPathsFor` are properly pure. But the gate *predicate* — the thing that is wrong in C1 and C2 — is fused into `issueFilesByID`'s IO, so it can only be exercised by building a git repo. Extract the decision over (basePaths, headPaths, renames) and the two failures above become table tests.
- **ARCH-PURPOSE — flag.** Shadow-sweep of the "one source, compiled to consumers" claim: the CI script genuinely derives from `sdlc issue lint-ids` (four-line adapter — good), but `sdlc merge`'s gate does **not** derive; it is a second implementation with its own message and its own scope, sharing only two helpers. That is why C2 exists in one and not the other. Separately, C2 and I1 are both "the instance was fixed, the class was not" — the Done-when's refusal half and BR-7's two sibling sites.
- **ARCH-MOCK — pass.** Real repo + real bare origin throughout, and the script is driven end-to-end in a derivative-shaped fixture. The one unmodelled dependency is GitHub Actions' checkout refspec (I2) — worth a note in the issue rather than a fake.
- **ARCH-CONSTRAINTS — flag.** BR-12 is untouched and `sdlc merge` now adds a second unbounded `git fetch` at step 4.6, so two interactive verbs can hang for the TCP connect timeout on a sleeping-VPN host. The CI check also pays a full `go build ./cmd/sdlc` per PR with no cache. None of these has a stated budget.
- **ARCH-SECURE — pass, with one recorded assumption.** The check builds and executes Go from the PR checkout and resolves its owner path from a repo-controlled script, but it runs under `pull_request` (no secrets, read-only token), the same profile as the existing `30-weave-drift.sh`. Untrusted input is handled well: `ls-tree` output is parsed through `IDFromFilename`, which returns 0 rather than fabricating an id, and a hand-edited or truncated filename degrades to "not an issue file". No credential touches a log, argv, or fixture.

## 7. Plan revision recommendations

The issue's `## Plan` rows are all ticked but two of them do not deliver what `## Done when` states. Add a `## Revisions` entry recording:

- **Done-when drift (C2):** "A within-ref scan … REFUSES when the PR introduces one" is not implemented — every within-ref duplicate at head is classified pre-existing and reported. Either implement the refusal or restate the Done-when.
- **New Done-when clause (C1):** the gate must not treat an archive or slug rename of an id it already owns as a collision; name PR #109 as the measured regression case.
- **Correct the round-1 Log entry for BR-3:** it reads as closed, but nothing pins it and the class survives for absolute `--issues-dir`/`WF_ISSUES_DIR` values under a symlinked root.
- **Record the remaining doc debt (M1)** as an explicit enumeration rather than three separate open findings, so the next round can check it mechanically.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Revert-verified — reverting the script to base="$fallback_base" turns TestMergeCheckScript_RefusesGivenMergeBase red with "[ok] no reused ids".
  - id: BR-2
    disposition: addressed
    note: |
      Revert-verified — first-path-only map turns the zzz-sorts-last subtest red while aaa stays green.
  - id: BR-3
    disposition: not-addressed
    note: |
      Behavior change is real (measured both binaries under a symlinked root) but NO test pins it — reverting repoRelativeIDDirs to d50a023's form leaves every #213 test green — and the class survives for absolute dirs: --issues-dir under a symlinked path with workshop/history (or history/issues) absent still prints "outside the current repo" and disables all three layers.
  - id: BR-4
    disposition: addressed
    note: |
      Fetch added at issueids.go:172, unconditional and reachable; no test pins the ordering (TestRefuseDuplicateIssueIDs' origin is already current), noted in coverage.
  - id: BR-5
    disposition: addressed
    note: |
      The refusal path is now the tested one and it genuinely goes red when the BR-1 fix is reverted; the skip path is a separate, correctly named test.
  - id: BR-6
    disposition: addressed
    note: |
      Verified by weave compile --dry-run in parley.nvim (emits the symlink row) plus a derivative-shaped fixture with no cmd/sdlc; atlas and Done-when corrected. Actual propagation to parley.nvim has not been run yet - operator action.
  - id: BR-7
    disposition: addressed
    note: |
      Fixed and pinned at publishedIssueIDs, but only there - see the new finding for the two sibling sites that still swallow.
  - id: BR-8
    disposition: not-addressed
    note: |
      Three listing parsers and three rev-parse+ls-tree shells remain; PathsByIDInTreeListing was never written, and the shells' error handling has now diverged.
  - id: BR-9
    disposition: not-addressed
    note: |
      atlas/workflow/ci-merge-check.md still opens a fence at line 31 and closes it at 47, so lines 33-46 render as code; that prose is now also stale.
  - id: BR-10
    disposition: not-addressed
    note: |
      cmd/sdlc/helptext/issue.md SUBCOMMANDS still lists new/sync/set-status/list/show only.
  - id: BR-11
    disposition: not-addressed
    note: |
      Still gated on !f.NoValidate at merge.go:339, no --no-dupid, no FLAGS documentation, and no warning printed when the bypass fires.
  - id: BR-12
    disposition: not-addressed
    note: |
      No timeout or --no-fetch escape; merge step 4.6 now adds a second unbounded fetch.
findings:
  - id: new
    severity: Critical
    family: gate-predicate-ignores-range-delta
    title: |
      Archiving or renaming an issue file is refused as a reused id by both the CI check and the merge gate
    detail: |
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
  - id: new
    severity: Important
    family: gate-predicate-ignores-range-delta
    title: |
      A within-ref duplicate the range introduces is labelled pre-existing and passes, contradicting Done-when
    detail: |
      runIssueLintIDs runs DuplicateIDsInRef over the head listing and reports every result
      as pre-existing; introducedIDClashes cannot see them because base[id] is empty. A
      fixture whose branch adds 000500-agent-a.md and 000500-agent-b.md (neither on the
      trunk) prints "pre-existing duplicate id 000500" followed by "id lint: this range
      introduces no reused issue ids" and exits 0. Done-when states the scan "REFUSES when
      the PR introduces one". Same root cause and same fix as the Critical above: an id is
      introduced-within-ref when head holds two or more distinct paths for it and at least
      one is absent at base.
  - id: new
    severity: Important
    family: silent-degradation-in-allocator
    title: |
      BR-7's class was fixed at one of three sites; issueFilesByID and idListing still swallow ls-tree failures
    detail: |
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
  - id: new
    severity: Important
    family: gate-compares-wrong-baseline
    title: |
      git fetch origin main does not guarantee refs/remotes/origin/main, so CI can fall back to the merge-base baseline BR-1 proved blind
    detail: |
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
  - id: new
    severity: Minor
    family: docs-lag-new-surface
    title: |
      atlas ci-merge-check.md still describes the pre-BR-6 skip conditions; three of five doc homes for this surface are wrong
    detail: |
      This is the 3rd finding in family docs-lag-new-surface. Do NOT fix only this file.
      The rule: a change to a user-facing surface updates that surface's doc homes in the
      SAME commit - helptext, the atlas entry, and README where the verb is listed - and
      corrects prose describing the old behavior. Measured enumeration for 213: helptext
      issue.md SUBCOMMANDS missing lint-ids (BR-10); helptext merge.md FLAGS silent on the
      gate (BR-11); ci-merge-check.md lines 41-45 both trapped in a code fence (BR-9) and
      still claiming the script keys on ./cmd/sdlc, which is exactly what BR-6 changed;
      sdlc-binary.md correct; README not applicable (issue verbs are not listed there).
      Three of five homes wrong - sweep the list.
```

---

## Re-review — 2026-09-03T17:49:32-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 213 — Allocate issue IDs against origin/main |
| repo | ariadne |
| issue file | workshop/issues/000213-nextid-from-origin.md |
| boundary | whole-issue close |
| milestone | — |
| window | 9078f8d0529b0c9fd4312f435b10f731aa5d3dc2..65aea14c2632156106f0662f6396cbd2af0e9c8c |
| command | sdlc close --issue 213 |
| reviewer | claude |
| timestamp | 2026-09-03T17:49:32-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The three-tree merge model added in `65aea14` is the right idea and it does fix BR-13's named instance (replayed against the real merged PR #109: exit 0, no false refusal). But it was only made symmetric on the **head** side: `mergedPathsFor` subtracts paths head deleted relative to base and never subtracts paths the **trunk** deleted. Measured end to end through the CI adapter — a branch cut before another issue's close, with `main` archiving `workshop/issues/000007-x.md` → `workshop/history/issues/000007-x.md`, is refused with "#000007 would be claimed by 2 files after merge" while `git merge` produces exactly one file. Since `sdlc merge` archives on every close, this fails `sdlc merge` step 4.6 **and** the required-status CI check for essentially every PR that is open across any issue close on main — BR-13's failure mode re-armed on the mirror axis. That blocks. Alongside it, three of the four "addressed" claims in commit `65aea14`'s message do not hold at the site the finding named: `idListing` still swallows `ls-tree` failures (BR-15 claimed "all three now error"), the CI script still uses the bare `git fetch origin main` the commit says it replaced (BR-16), and none of the five doc homes moved (BR-9/10/11/17), including for the brand-new `--trunk` flag.

## 1. Strengths

- **`ARCH-MOCK` is genuinely satisfied.** Every allocation test drives a real repo against a real bare origin with the branch cut *before* a second clone publishes the colliding id (`cmd/sdlc/issueids_test.go:56`). A function-call fake cannot express "a ref exists that this worktree does not contain", and the tests say so out loud. `TestMergeCheckScript_RefusesGivenMergeBase` executes the actual shell adapter, so the CI layer is exercised, not read.
- **`ARCH-PURE` on the decision layer.** `issue.NextID(idSets ...[]int)`, `IDsInTreeListing`, `DuplicateIDsInRef`, `mergedPathsFor` and `introducedCollisions` are all pure and table-tested with zero IO; the git calls sit in `issueids.go`'s shell behind an injected `gitRunner`. `TestNextID_IsPure` and `TestMergedPathsFor_ModelsTheMergeResult` run without a repo.
- **BR-13's instance is properly pinned.** Reverting `mergedPathsFor` to "head has a path base lacks" turns the archive/rename/renumber rows of `TestIntroducedCollisions_ArchiveAndRenameAreNotCollisions` and `TestMergedPathsFor_ModelsTheMergeResult` red. Verified against the real repo too (base `008f7e3^1`, head `008f7e3^2` → exit 0).
- **BR-14 is fixed and verified.** A branch adding `000500-agent-a.md` and `000500-agent-b.md`, neither at base, now exits 1 rather than passing with a "pre-existing" label.
- **BR-3 is fixed and verified live.** Ran `sdlc issue lint-ids` in a fresh repo under `/tmp` (→ `/private/tmp`) with no `workshop/history/` at all: it reported the planted duplicate instead of printing "outside the current repo" and skipping.
- **The DRY argument for keeping logic in Go** (`issuelintids.go:5`) is right and the script really is a thin adapter — filename parsing and directory selection are decided once.

## 2. Critical findings

**C1 — `mergedPathsFor` ignores deletions on the trunk side, so any PR open across an issue archive on `main` is falsely refused (`cmd/sdlc/issueids.go:130`, the loop at `:142`).**

This is the **3rd finding in family `gate-predicate-ignores-range-delta`**. Rounds 1 and 2 each fixed the instance named. Per ARCH-PURPOSE, do not fix this instance — fix the class.

*The rule:* the predicate models a three-way merge, so it must be **symmetric in `head` and `trunk`**. A path present at `base` and absent on *either* side was deleted by that side and cannot survive the merge. The formula in the doc comment at `:123` is half-written:

```
merged(id) = (trunk(id) ∪ head(id)) − (base(id) − head(id)) − (base(id) − trunk(id))
                                       ^ implemented              ^ missing
```

*Measured failure* (`sdlc` built at `65aea14`, real repo + bare origin, branch cut before the archive):

```
main:    git mv workshop/issues/000007-x.md workshop/history/issues/000007-x.md
feature: untouched, one unrelated commit
$ sdlc issue lint-ids --base $(git merge-base origin/main HEAD) --trunk origin/main --head HEAD
  #000007 would be claimed by 2 files after merge: ...            exit 1
$ bash scripts/merge-checks.d/40-duplicate-issue-id.sh $MB HEAD   exit 1
$ git merge origin/main   → workshop/history/issues/000007-x.md   (one file)
```

`refuseDuplicateIssueIDs` (`issueids.go:243`) shares the predicate and dies the same way, so both the local gate and the required-status CI check refuse a merge git performs cleanly.

*The enumeration the class implies* — {add, delete, rename/move} × {head, trunk}, plus both-sides. The test table at `issueids_test.go:522` has rows only for the head column; write the trunk column and the fix follows. I verified the symmetric predicate against all ten shapes (head archives, **trunk archives**, head renames, **trunk renames**, head renumbers, cut-then-publish, second path added for a live id, pre-existing duplicate, two-files-one-id on a branch, both sides archive): 10/10 correct, including the two the current code gets wrong and every one it currently gets right.

Two more members of the same class to sweep in the same round rather than the next:

- `refuseDuplicateIssueIDs` (`issueids.go:255-261`): when `merge-base` returns empty or its `issueFilesByID` read fails, `base` stays `{}`, which erases *all* deletion information and turns every trunk∪head pair into a refusal — the same defect with no fixture behind it.
- `runIssueLintIDs` (`issuelintids.go:79-82`) labels every `DuplicateIDsInRef` hit on `head` "pre-existing" without consulting `base`. It no longer *passes* one the range introduced (BR-14 is fixed by the other path), but the operator is now told the same id is both pre-existing and newly introduced. Same rule: a range verdict cannot be computed from one tree.

## 3. Important findings

All four are re-raised prior findings; see the `findings` block for dispositions. Summarised:

- **BR-15 (`silent-degradation-in-allocator`, 3rd) — not addressed.** `idListing` at `cmd/sdlc/issuelintids.go:149-151` still does `if err != nil { continue }`, so a dropped directory makes `sdlc issue lint-ids` print `[ok]` having read a fraction of the id space. `65aea14`'s message claims "the ls-tree swallow class fixed at the two remaining sites"; one of the two was not touched. The `issueFilesByID` half *was* changed, but **no test pins it** — `lsTreeFailRunner` is used only by `TestPublishedIssueIDs_PartialReadIsAnError`, so reverting `issueids.go:303` to `continue` leaves the suite green. Per the claimed-fix rule that is `not-addressed` however plausible the diff reads.
- **BR-16 (`gate-compares-wrong-baseline`, 4th) — not addressed at the site the finding named.** The explicit refspec landed in `issueids.go:83` and `:246` (Go), but `scripts/merge-checks.d/40-duplicate-issue-id.sh:65` still runs `git fetch --quiet origin main`. Re-measured on a `--single-branch` clone with a narrow `remote.origin.fetch`: plain form → `origin/main` unresolvable; explicit refspec → resolved. The degraded path then runs `lint-ids --base "$fallback_base"` and exits 0 — a green check computed from the merge-base BR-1 proved blind. `introducedIDClashes` (`issuelintids.go:120-124`) has the same shape in Go: a failed trunk read silently sets `trunkByID = baseByID`.
- **BR-17 (`docs-lag-new-surface`, 3rd) — not addressed; now four of five homes wrong.** `atlas/workflow/ci-merge-check.md` is unchanged (still fenced, still says the script "builds sdlc from the checkout under test" and skips on "no `./cmd/sdlc`" — both describe the pre-BR-6 script). `helptext/issue.md` SUBCOMMANDS still omits `lint-ids`. `helptext/merge.md` FLAGS still silent on the gate. And `65aea14` added a **new user-facing flag `--trunk`** plus the merge-result semantics with no doc home at all: `atlas/workflow/sdlc-binary.md:187` still describes only "a branch-vs-trunk comparison".
- **BR-11 (`gate-bypass-flag-granularity`, 1st) — not addressed.** `merge.go:336` still gates on `!f.NoValidate`, so bypassing #124's instance-conformance gate silently bypasses the id gate too, contrary to AGENTS.md §5's per-gate `--no-<gate>` convention.

## 4. Minor findings

- BR-8 (`duplicated-listing-parser`) still open: `IDsInTreeListing`, `DuplicateIDsInRef` and `issueFilesByID` each re-implement split → trim → `LastIndex("/")` → `IDFromFilename`; `publishedIssueIDs`/`issueFilesByID`/`idListing` each re-implement rev-parse + `repoRelativeIDDirs` + per-dir `ls-tree`. One `PathsByIDInTreeListing` + one IO shell returning an error discharges BR-8 and BR-15 together (`ARCH-DRY`).
- BR-9: `atlas/workflow/ci-merge-check.md:30-46` — the new prose still renders as literal code inside the fence.
- BR-12 (`unbounded-external-call`) still open, and now at two call sites (`issueids.go:83`, `:246`) rather than one.
- `allocateIssueID` reports `repoRelativeIDDirs` failures as "`origin/main` unreachable" (`issueids.go:48`), which misattributes a path-containment refusal to the network.
- `runIssueLintIDs`'s `stdout` parameter is unused; every line goes to stderr.
- `ls-tree --name-only` output is parsed line-by-line; `-z` would remove the (currently theoretical) dependency on filenames without newlines.

## 5. Test coverage notes

The bar the checklist sets — "the kind of bug this diff could ship is covered" — is met for allocation and for the head-side merge shapes, and not met for the trunk side. `TestMergedPathsFor_ModelsTheMergeResult` (`issueids_test.go:522`) even *encodes* the defect: the row "trunk archived it while the branch edited it in place" asserts `wantIDs: []int{7}` with the comment "both survive, which IS a real contradiction to surface". Git disagrees — I ran the merge; one file survives. That row is the failing case written down as the expectation, which is exactly the "test written from the same mental model as the fix" the claimed-fixes rule warns about. Fix the predicate and that row flips to `nil`, plus a trunk-rename row.

Two other gaps: no test pins `issueFilesByID`'s or `idListing`'s partial-read behavior (BR-15), and no test drives `refuseDuplicateIssueIDs`/the script with an empty or unresolvable merge-base. Suite state: `go test ./cmd/sdlc/...` is green except the known unrelated `TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory` (#210, missing archived plan path).

## 6. Architecture notes

- **ARCH-DRY — flag.** BR-8, unchanged across three rounds. The consolidation also removes the duplication that produced BR-2 and BR-15.
- **ARCH-PURE — pass.** The decision layer is pure and directly unit-tested; git lives behind `gitRunner`. `TestMergedPathsFor_ModelsTheMergeResult` runs with no repo at all.
- **ARCH-PURPOSE — flag (C1, BR-15, BR-16, BR-17).** Four findings this round were answered at the single site each named while enumerable siblings of the same class stayed in the tree: head-side move but not trunk-side, two of three `ls-tree` sites, the Go fetch but not the script fetch, one doc home of five. The shadow-sweep for the id-space single source is otherwise good — allocation, the local gate and CI all derive from `sdlc issue lint-ids`, with no hand-maintained restatement.
- **ARCH-MOCK — pass.** Real bare-origin fixtures throughout, and the shell adapter runs for real. The one absence is a live conformance check on the runner contract itself; the check runs in CI, which is adequate.
- **ARCH-CONSTRAINTS — flag (BR-12).** Two unbounded `git fetch` calls, one of them on the interactive `sdlc issue new` path. The per-PR `go build ./cmd/sdlc` in the CI script is bounded and fine.
- **ARCH-SECURE — pass with a note.** No credentials, no untrusted network bodies. `ls-tree` output is process-local and trusted; `repoRelativeIDDirs`'s containment check is the right instinct (it stops a `--issues-dir` elsewhere from reading this repo's trunk as its own id space). The degraded paths, however, substitute a *fabricated* baseline (`trunkByID = baseByID`, empty `base`) that downstream code reads as evidence and reports `[ok]` from — the at-review lens asks that parse failures degrade visibly, which they currently don't.

## 7. Plan revision recommendations

`workshop/issues/000213-nextid-from-origin.md` needs a `## Revisions` entry (or a round-3 Log section) recording:

1. The round-2 Log's "All three now error" (BR-15) and "Explicit refspec now" (BR-16) are **incorrect as written** — `idListing` still swallows, and the script still uses the bare fetch form. State what was actually changed.
2. The "Verified on all four shapes end to end" claim covers only the head-side shapes; the trunk-side archive and trunk-side rename were never tried, and both fail. Record the measured reproduction.
3. Done-when needs a row for the merge-model invariant, phrased symmetrically: *"a file the trunk archived or renamed while a branch was open does not read as a collision"* — otherwise the next round can satisfy the existing wording without covering the class.
4. `## Plan` currently shows all eight rows ticked while a Critical is open in the delivered predicate; add a row for the symmetric-merge sweep and the doc-home sweep rather than leaving the plan claiming completion.

```findings
dispose:
  - id: BR-3
    disposition: addressed
    note: |
      Verified live — fresh repo under /tmp (→/private/tmp) with no workshop/history: lint reported the planted duplicate instead of skipping.
  - id: BR-13
    disposition: addressed
    note: |
      The named instance is fixed (real PR 109 replay, base 008f7e3^1 / head 008f7e3^2 → exit 0), but the class is not swept — see the new Critical.
  - id: BR-14
    disposition: addressed
    note: |
      Verified — a branch adding 000500-agent-a.md and 000500-agent-b.md now exits 1 rather than labelling them pre-existing and passing.
  - id: BR-8
    disposition: not-addressed
    note: |
      Three listing parsers and three rev-parse+ls-tree IO shells still present, unchanged across all three rounds.
  - id: BR-9
    disposition: not-addressed
    note: |
      atlas/workflow/ci-merge-check.md lines 30-46 unchanged; the prose still renders inside the code fence.
  - id: BR-10
    disposition: not-addressed
    note: |
      helptext/issue.md SUBCOMMANDS still omits lint-ids, and now also the new --trunk flag added in 65aea14.
  - id: BR-11
    disposition: not-addressed
    note: |
      merge.go:336 still gates on !f.NoValidate; merge.md FLAGS still documents neither gate.
  - id: BR-12
    disposition: not-addressed
    note: |
      Still unbounded, and now at two call sites (issueids.go:83 and :246) rather than one.
  - id: BR-15
    disposition: not-addressed
    note: |
      idListing (issuelintids.go:149) still does `continue`; issueFilesByID's fix is pinned by no test — reverting issueids.go:303 to `continue` leaves the suite green.
  - id: BR-16
    disposition: not-addressed
    note: |
      The explicit refspec landed in Go only; scripts/merge-checks.d/40-duplicate-issue-id.sh:65 still runs `git fetch --quiet origin main`. Re-measured on a single-branch clone with a narrow refspec — plain form leaves origin/main unresolvable, explicit form resolves it. Sibling site — introducedIDClashes at issuelintids.go:120-124 silently substitutes baseByID when the trunk read fails.
  - id: BR-17
    disposition: not-addressed
    note: |
      No doc home moved this round, and 65aea14 added a new one — the --trunk flag and the merge-result model appear in no helptext or atlas entry. sdlc-binary.md:187 still describes only a branch-vs-trunk comparison; ci-merge-check.md still describes the pre-BR-6 skip conditions and build location.
findings:
  - id: new
    severity: Critical
    family: gate-predicate-ignores-range-delta
    title: |
      mergedPathsFor honours deletions by head but not by the trunk, so every PR open across an issue archive on main is falsely refused
    detail: |
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
```

---

## Re-review — 2026-09-03T18:04:57-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 213 — Allocate issue IDs against origin/main |
| repo | ariadne |
| issue file | workshop/issues/000213-nextid-from-origin.md |
| boundary | whole-issue close |
| milestone | — |
| window | 9078f8d0529b0c9fd4312f435b10f731aa5d3dc2..0beea8fbfad2fff1d2e3703433a337627cdf9c39 |
| command | sdlc close --issue 213 |
| reviewer | claude |
| timestamp | 2026-09-03T18:04:57-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The Critical that dominated round 3 is genuinely fixed and genuinely pinned: I reverted `mergedPathsFor`'s symmetry in a scratch copy and `TestMergedPathsFor_ModelsTheMergeResult/trunk_archived_it_while_the_PR_was_open` goes red with exactly the false refusal BR-18 described. The three-tree merge model is correct across every shape I could construct, and the test tables now encode the right answers. What blocks a clean SHIP is that the class-sweep instructions attached to three open findings were not carried out: BR-18 named two sibling sites to fix in the *same* round and both are still there (I reproduced both); BR-16's one-word fix landed in the two Go call sites but **not** in `40-duplicate-issue-id.sh`, which is the only site BR-16 actually named — and the issue Log states "both fetch sites use the explicit `+refs/heads/main:...` refspec", which is a false evidence claim about the site that matters. Beyond the open set I found one new measured member of `gate-predicate-ignores-range-delta`: a collision living entirely on the trunk, published after the branch was cut, refuses a PR that touched nothing. All five open Minors (BR-8/9/10/11/12) are untouched.

## 1. Strengths

- **`mergedPathsFor` is the right abstraction, and it is pure.** `cmd/sdlc/issueids.go:138` takes three maps and returns one; `introducedCollisions` is a thin filter over it. Both are unit-tested with no IO (`issueids_test.go:490`, `:546`). This is textbook ARCH-PURE and it is why the round-3 fix was a two-condition change rather than a rewrite.
- **The revert-verification is real this time.** Restoring `if !containsPath(head[id], p)` alone produces `introducedCollisions = [7], want []` with the merged map printed in the failure. The fix is pinned by a test that fails without it — the ariadne#194 bar, met.
- **`ARCH-MOCK` is honoured properly.** Every allocation test drives a real repo against a real bare origin, and `TestAllocateIssueID_BranchCutBeforePublish` builds the fixture in the load-bearing order (cut → publish from a *second clone* → allocate). A function-call mock cannot express "a ref exists that this worktree does not contain", and the header comment at `issueids_test.go:16` says exactly that.
- **`TestIntroducedIDClashes_IndependentOfSlugSortOrder` asserts both orders** (`issueids_test.go:432`), which is the correct response to BR-2 — the previous test passed only because its slug happened to sort first.
- **The BR-7 error-return decision is well-argued in-place.** The comment at `issueids.go:94-98` explains *why* an `ls-tree` failure differs from an absent directory (exit 0, empty output), which is the non-obvious fact a future reader needs.

## 2. Critical findings

None new. (BR-18 is disposed `not-addressed` for its unswept members, not for its primary instance.)

## 3. Important findings

**I1 — `introducedCollisions` excludes on `base` only, so a collision that lives entirely on the trunk is charged to an innocent PR.**
`cmd/sdlc/issueids.go:188` — `if len(paths) > 1 && len(base[id]) < 2`. When a collision lands on `main` *after* a branch is cut, `base[id]` is 1 (or 0) but `trunk[id]` is 2, so both trunk paths survive into `merged` and the range is blamed. Measured against a real repo + bare origin: a PR whose only change is adding an unrelated `000002-*.md` is refused with `#000001 would be claimed by 2 files after merge`. Fix sketch: `&& len(trunk[id]) < 2`. I applied that one-liner in a scratch copy — the probe passes and every existing `#213` test stays green (`TestIntroduced*`, `TestMerged*`, `TestRefuseDuplicate*`, `TestMergeCheckScript*`, `TestAllocateIssueID*`, `TestDuplicateIDsInRef*`), including the two shapes that define the issue.

**I2 — the derivative propagation run has not happened, so the enforcement half reaches no repo that carries collisions (ARCH-PURPOSE).**
`construct/base.manifest:131` declares `symlink scripts/merge-checks.d/40-duplicate-issue-id.sh` and the mechanism is sound (weave's `scaffold` is an idempotent `MkdirAll`, `apply.go:150`, so it will not clobber the symlink). But `../parley.nvim/scripts/merge-checks.d/` contains only `.gitkeep`. The issue's Problem section opens with "half the measured collisions are in `parley.nvim`, not here", and Done-when claims "The check reaches derivatives". A manifest row is the source; a derivative that does not yet derive from it is a deferred consumer, not a finished one. One `sdlc propagate-base` run closes it — this is the thing that is the point, not a separable extension.

## 4. Minor findings

- **M1** — `refuseDuplicateIssueIDs` takes an injected `r gitRunner` and then calls `gitx.Capture("merge-base", …)` directly (`issueids.go:265`), so its baseline read escapes the seam. `Capture` also returns `""` on error (`gitx/window.go:52`), making failure indistinguishable from success. This is *why* the degraded path in BR-18's sweep member 1 cannot be pinned through the runner. Route it through `r`.
- **M2** — `workshop/lessons.md` gained 59 lines, all of them #211's. #213's round-3 insight — "a table only pins what you believed when you wrote it"; the reviewer's own fixture asserted the defect as expected behaviour — has no lessons.md entry, and it is not code-enforceable, so AGENTS §4 applies.
- **M3** — the clash-report `fmt.Sprintf("  #%06d would be claimed by %d files after merge:\n      %s", …)` is byte-identical at `issueids.go:275` and `issuelintids.go:129`. Same consolidation as BR-8 (ARCH-DRY).
- **M4** — `sdlc merge` now fetches `origin/main` twice per invocation: `issueids.go:252` at step 4.6 and `merge.go:469` at the merged-check. Not hot-path, but it is a second network round-trip on an interactive verb.

## 5. Test coverage notes

- **The BR-15 class fix is unpinned at 2 of 3 sites.** `TestPublishedIssueIDs_PartialReadIsAnError` (`issueids_test.go:462`) covers `publishedIssueIDs`. `issueFilesByID`'s and `idListing`'s new `return …, fmt.Errorf(…)` have no test at all — `lsTreeFailRunner` is used once. Deleting either error-return leaves the suite green. This is the `fix-not-pinned-by-a-failing-test` family (2nd instance): the rule is that a class fix needs one test *per member*, not one test for the member that happened to be found first.
- **No test exercises any degraded read path.** Both swallow sites (`issueids.go:268`, `issuelintids.go:122`) and both warn-and-exit-0 consumers are uncovered; I had to write throwaway probes to reach them. A single `refLsTreeFailRunner{badRef: …}` helper — the pattern I used — makes all four cheap to pin.
- **No test exercises the narrow-refspec CI shape.** `TestMergeCheckScript_RefusesGivenMergeBase` passes because the fixture clone carries git's default `+refs/heads/*` refspec, so `git fetch origin main` incidentally populates `origin/main`. A fixture that sets `remote.origin.fetch` to the PR branch only would go red today; I confirmed the underlying git behaviour directly (plain form → `origin/main` unresolvable; explicit refspec → resolves).
- Otherwise the pure surface is well covered: `NextID`, `IDsInTreeListing`, `DuplicateIDsInRef`, `mergedPathsFor`, `introducedCollisions` all have IO-free tables.
- Suite state: `go test ./...` is green except `TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory`, which I confirmed is pre-existing and unrelated — the plan file was archived by `dfeba9c`, before this window's base.

## 6. Architectural notes

- **ARCH-DRY — flag.** BR-8 unaddressed, plus M3. Three parsers (`IDsInTreeListing`, `DuplicateIDsInRef`, `issueFilesByID`) all do split → trim → `LastIndex("/")` → `IDFromFilename`; three IO shells (`publishedIssueIDs`, `issueFilesByID`, `idListing`) all do rev-parse → `repoRelativeIDDirs` → per-dir ls-tree. One `PathsByIDInTreeListing` + one shared shell returning `(map[int][]string, error)` collapses all six and discharges BR-8, BR-15's remaining sites, and M3 together.
- **ARCH-PURE — pass, with M1.** The decision layer is genuinely pure and directly tested. The single leak is `gitx.Capture` inside a runner-injected function.
- **ARCH-PURPOSE — flag (I2).** The issue's stated purpose is enforcement across the fleet; the fleet does not yet carry the check. Also relevant to the sweep instructions: BR-15, BR-16 and BR-18 each explicitly named the class and the sibling sites, and each was answered at a subset. Four families now sit at 3–4 findings. Writing the enumeration down as a checklist in `## Plan` before the next round would break the pattern that reading the finding text apparently does not.
- **ARCH-MOCK — pass.** Real git, real bare origin, production and test flow share the boundary. No live-conformance cadence is declared for `ls-tree`/`fetch` behaviour, but these are stable plumbing commands and the tests run the real binary, so I do not think one is owed here.
- **ARCH-CONSTRAINTS — flag.** BR-12 unaddressed: `git fetch` on every `sdlc issue new` is unbounded, on an interactive verb. Add a `GIT_HTTP_LOW_SPEED_*` bound or a `--no-fetch` escape and state the budget.
- **ARCH-SECURE — pass.** The check script executes `construct/dev-aliases.sh` from the repo under test and builds Go from the path it names, but the workflow is `pull_request` (fork code, read-only token, ephemeral runner), so this is the exposure CI already has. No credentials, no untrusted persisted format. One provenance note for future work: at the local `pre-push` call site the verdict depends on an out-of-tree sibling checkout's code, which is worth stating in the atlas when the propagation lands.

## 7. Plan revision recommendations

There is no `workshop/plans/000213-*-plan.md` (the Log records that `change-code` was skipped), so there is no Core-concepts table to cross-check and nothing to revise there. Two artifacts do need edits:

- **`workshop/issues/000213-nextid-from-origin.md`, `## Log`, round-3 entry** — correct the sentence "**BR-16 re-raised but verified present** — both fetch sites use the explicit `+refs/heads/main:refs/remotes/origin/main` refspec." There are three fetch sites; the third (`scripts/merge-checks.d/40-duplicate-issue-id.sh:65`) still uses the plain form, and it is the CI site BR-16 named. Per this repo's own lesson ("a Log line asserting evidence names the command run and the observed result"), replace the claim with the measured state.
- **`## Done when`** — "The check reaches derivatives … Verified by running it in `parley.nvim`" describes a manual run, not propagation. Either tick it only after `sdlc propagate-base`, or split it into "mechanism declared in `base.manifest`" (done) and "propagated to derivatives" (open, and already listed under Immediate next actions).

```findings
dispose:
  - id: BR-8
    disposition: not-addressed
    note: |
      Unchanged, plus a 4th duplication: the clash-report Sprintf is byte-identical at issueids.go:275 and issuelintids.go:129.
  - id: BR-9
    disposition: not-addressed
    note: |
      Verified at HEAD — prose still sits between the opening fence at ci-merge-check.md:31 and its close at :47.
  - id: BR-10
    disposition: not-addressed
    note: |
      helptext/issue.md:6-14 still lists new/sync/set-status/list/show only.
  - id: BR-11
    disposition: not-addressed
    note: |
      merge.go still gates on !f.NoValidate; merge.md FLAGS (line 80) documents --no-judge but neither --no-validate nor an id-gate flag.
  - id: BR-12
    disposition: not-addressed
    note: |
      issueids.go:85 fetch is still unbounded on the interactive `issue new` path.
  - id: BR-15
    disposition: not-addressed
    note: |
      The three named ls-tree sites now error, but the rule ("no code path reports a clean verdict
      from an incomplete listing") is still violated at four places, two of which are NEW swallow
      sites introduced by the round-2/3 work. Measured: issuelintids.go:122 drops terr entirely and
      substitutes baseByID for the trunk — a real collision then returns zero clashes with no error
      and no warning; issueids.go:268 drops berr and leaves base empty. runIssueLintIDs (:77, :91)
      and refuseDuplicateIssueIDs (:256, :261) both turn a read failure into exit 0 / return nil.
      Additionally the class fix is pinned at only 1 of 3 sites — issueFilesByID's and idListing's
      error-returns can be deleted with the suite green.
  - id: BR-16
    disposition: not-addressed
    note: |
      The two Go fetch sites got the explicit refspec; the THIRD and only site BR-16 named —
      scripts/merge-checks.d/40-duplicate-issue-id.sh:65 — is unchanged (`git fetch --quiet origin
      main`). Premise re-measured: with a narrow remote.origin.fetch the plain form leaves
      origin/main unresolvable, the explicit refspec resolves it. The degraded fallback (script
      lines 67-69, 80-81) also still exits 0 with a blind model rather than a loud degraded state:
      with base==trunk the predicate collapses to merged==head. The issue Log's claim that "both
      fetch sites" were fixed needs correcting.
  - id: BR-17
    disposition: not-addressed
    note: |
      ci-merge-check.md:42-46 still says the script "builds sdlc from the checkout under test" and
      lists "no ./cmd/sdlc" as a skip condition — both pre-BR-6 behaviour; dev-aliases owner
      resolution is unmentioned. Two more doc homes to add to the enumeration: the delivery table
      at :19 still says merge-checks.d is `scaffold` although base.manifest now symlinks the check,
      and the three-tree merge model (BR-13/BR-18, the diff's subtlest logic) has no atlas home at
      all — sdlc-binary.md documents the CI check but not the predicate. README confirmed N/A
      (85 lines, lists project/fleet/judge verbs only).
  - id: BR-18
    disposition: not-addressed
    note: |
      Primary instance FIXED and revert-verified: reverting the symmetry in mergedPathsFor turns
      TestMergedPathsFor_ModelsTheMergeResult/trunk_archived_it_while_the_PR_was_open red. But both
      named sweep members remain and both reproduce. (1) refuseDuplicateIssueIDs: with a failed
      merge-base ls-tree read, base collapses to {} and a routine `git mv` archive is refused —
      "#000001 would be claimed by 2 files after merge" — with an EMPTY stderr, no warning at all;
      the healthy-runner control passes on the same fixture. (2) runIssueLintIDs:80-82 still labels
      a duplicate the range introduced "pre-existing duplicate id #000001" without consulting base.
findings:
  - id: new
    severity: Important
    family: gate-predicate-ignores-range-delta
    title: |
      A collision living entirely on the trunk is charged to a PR that touched nothing
    detail: |
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
  - id: new
    severity: Minor
    family: io-escapes-injected-seam
    title: |
      refuseDuplicateIssueIDs takes an injected gitRunner but reads its baseline via gitx.Capture
    detail: |
      issueids.go:265 calls gitx.Capture("merge-base", trunkRef, "HEAD") directly while the rest of
      the function goes through `r gitRunner`. gitx.Capture returns "" on error
      (gitx/window.go:52-57), so a failed merge-base is indistinguishable from an empty one. This
      is the reason BR-18's first sweep member cannot be pinned through the seam: a test cannot
      make the baseline read fail. Route it through r and treat "" as an explicit degraded state.
  - id: new
    severity: Minor
    family: enforcement-does-not-propagate
    title: |
      base.manifest declares the symlink but no derivative carries the check yet
    detail: |
      This is the 2nd finding in family enforcement-does-not-propagate. The rule that covers both:
      a manifest row is a declaration, not propagation — a single-source change is not delivered
      until every consumer derives from it (ARCH-PURPOSE). BR-6 fixed the delivery KIND
      (scaffold to symlink); the propagation run has not happened.
      Measured: ../parley.nvim/scripts/merge-checks.d/ contains only .gitkeep. parley.nvim holds
      four of the eight collisions the issue was opened over, and Done-when claims "The check
      reaches derivatives". The mechanism is sound — weave's scaffold is an idempotent MkdirAll
      (plan/apply.go:150) so it will not clobber the symlink — so this is one `sdlc propagate-base`
      run plus a Done-when split into "mechanism declared" (done) and "propagated" (open).
  - id: new
    severity: Minor
    family: docs-lag-new-surface
    title: |
      No lessons.md entry for this issue's own round-3 lesson
    detail: |
      This is the 4th finding in family docs-lag-new-surface, on the lessons axis rather than the
      atlas axis. The rule: a review round that produces a non-code-enforceable insight records it
      in workshop/lessons.md in the same commit that closes the finding. The +59 lines added in
      this window are all 211's. The 213 round-3 insight — a test table encoding the defect as the
      expected value, so the fixture asserted the bug and passed — is only in the issue Log. It is
      not code-enforceable (no guard can tell a wrong expectation from a right one), which is
      exactly the criterion for a lessons entry.
```
