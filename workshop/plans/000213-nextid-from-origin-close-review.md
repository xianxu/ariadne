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
