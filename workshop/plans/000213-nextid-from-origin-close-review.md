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

---

## Re-review — 2026-09-03T18:19:30-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 213 — Allocate issue IDs against origin/main |
| repo | ariadne |
| issue file | workshop/issues/000213-nextid-from-origin.md |
| boundary | whole-issue close |
| milestone | — |
| window | 9078f8d0529b0c9fd4312f435b10f731aa5d3dc2..2a69212a40a0cd3c8addc752828deeaac2d79637 |
| command | sdlc close --issue 213 |
| reviewer | claude |
| timestamp | 2026-09-03T18:19:30-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The three-tree merge model is now genuinely right — I could not construct a shape where `mergedPathsFor`/`introducedCollisions` mis-decides given complete inputs, and the BR-18 primary fix is revert-pinned. What blocks the boundary is that the defect the issue exists to fix still reproduces on a common path: `sdlc issue new` while offline, in any repo that has fetched before, reallocates a **published** id with an empty stderr. `publishedIssueIDs` discards the fetch error (`_, _ = r.Git("fetch", …)`) and then reads whatever stale `origin/main` is lying around; the existing offline test only passes because it deletes `refs/remotes/origin/main` first, which is the one repo state where the warning can fire. I measured it: branch cut, `#2` published from a second clone, origin URL broken → `allocated=000002 stderr=""`. That is the original bug, silently, through the fix, and it violates the Spec's own words ("a silent fallback here recreates the bug it is meant to fix"). Alongside it, three of the open findings' *measured* claims re-reproduced at HEAD unchanged (BR-15's trunk substitution returns a clean verdict on a real collision; BR-18's base-collapse falsely refuses a routine archive with empty stderr; BR-16's script fetch line is untouched), and BR-19 — the only code change in the last commit — is unpinned: reverting `|| len(trunk[id]) > 1` leaves the entire suite green.

## 1. Strengths

- **`mergedPathsFor` is the right abstraction and it is pure** (`cmd/sdlc/issueids.go:138`). Three maps in, one map out; `introducedCollisions` is a thin filter over it; both unit-tested with no IO. I probed ten merge shapes by hand (archive by head, archive by trunk, rename, renumber, cut-then-publish, both-sides-add, head-archives-while-adding-a-collider, trunk-publishes-a-collider, rename-across-a-trunk-collision, pre-existing-grows) and the predicate is correct on every one given complete inputs. ARCH-PURE, cleanly.
- **ARCH-MOCK is honoured properly.** Every allocation test drives a real repo against a real bare origin, and `TestAllocateIssueID_BranchCutBeforePublish` (`issueids_test.go:57`) builds the fixture in the load-bearing order — cut, publish *from a second clone*, allocate. The header comment at `:16` states exactly why a function-call mock cannot express this bug. That decision is what made all five review rounds able to measure rather than argue.
- **`TestIntroducedIDClashes_IndependentOfSlugSortOrder`** (`:432`) asserts both orders, which is the correct structural answer to BR-2 rather than a fixture that happened to sort right.
- **The in-place rationale comments are unusually good** — `issueids.go:94-98` explains why an `ls-tree` failure differs from an absent directory (exit 0, empty output), and `:106-137` records the two *wrong* definitions of "collision" alongside the right one. A future reader will not re-derive those two mistakes.
- **The `sdlc issue lint-ids` verb rather than bash** (`issuelintids.go` header) is the right ARCH-DRY call: filename parsing and the introduced-vs-pre-existing split are decided once, in Go, with tests, and the CI script is a genuine four-line adapter.

## 2. Critical findings

**C1 — `publishedIssueIDs` discards the fetch error, so offline allocation reads a stale trunk and re-allocates a published id, silently.** `cmd/sdlc/issueids.go:85` — `_, _ = r.Git("fetch", …)`. Measured against a real repo + bare origin: branch cut, `000002-published-elsewhere.md` published from a second clone, then `git remote set-url origin <gone>` → `allocateIssueID` returns `000002` with `stderr=""`. `TestAllocateIssueID_OfflineWarnsAndProceeds` (`issueids_test.go:117`) passes only because it runs `git update-ref -d refs/remotes/origin/main` first; any repo that has ever fetched — i.e. every real one — takes the silent path instead. Done-when claims "With origin unreachable, creation still succeeds and emits the warning; a test asserts the warning is present". It does not, on the common shape. **This is the 3rd finding in family `silent-degradation-in-allocator`**, so do not patch this one site: the rule is that *every* read feeding the id space must be verified-fresh-and-complete or announced as degraded, and no path may return a clean verdict or a silent success from a read it could not complete or confirm current. The enumeration is mechanical — eight sites where a git result is discarded or substituted: `issueids.go:85` and `:262` (fetch errors dropped), `:275` (`gitx.Capture` returns `""` on error), `:277-281` (`berr` dropped → `base` collapses to `{}`), `:266`/`:271` (read failure → `return nil`, merge proceeds), `issuelintids.go:122` (`terr` dropped → base substituted for trunk), `:77`/`:91` (read failure → exit 0), and `40-duplicate-issue-id.sh:65-69,80-81` (unresolvable trunk → the BR-1-blind baseline, exit 0). Three rounds have now fixed three of these one at a time. Fix the class with one mechanism — a `freshness{fresh|stale|failed}` threaded out of the read layer, where anything but `fresh` forces the loud warning on the allocator and a non-clean exit on the CI verb (or an explicit `--allow-degraded`).

## 3. Important findings

All are re-raised prior findings; see the dispositions. The load-bearing ones, with what I measured at HEAD:

- **BR-15** — `introducedIDClashes` (`issuelintids.go:122`) drops `terr` and substitutes `baseByID` for the trunk. Probe: real repo + bare origin, branch cut before `000500-theirs.md` published, branch adds `000500-mine.md`; healthy runner returns 1 clash, a runner that fails only the trunk `ls-tree` returns `clashes=[] err=<nil>` — the enforcement gate reports clean on a real collision.
- **BR-18 sweep members** — both reproduce. (1) `refuseDuplicateIssueIDs` with a failing merge-base read refuses a routine `git mv` archive: `#000001 would be claimed by 2 files after merge` naming `workshop/history/issues/000001-first.md` and `workshop/issues/000001-first.md`, with **empty stderr**; the healthy-runner control on the same fixture passes. (2) `sdlc issue lint-ids --base <sha> --head HEAD` on a branch that just introduced a duplicate prints `[!] pre-existing duplicate id #000001` *and then* refuses it as introduced — the same id reported under both labels in one run.
- **BR-16** — `40-duplicate-issue-id.sh:65` is byte-identical to the day it was written (`git fetch --quiet origin main`); only the two Go sites changed. Premise re-measured in a fresh repo with a narrow `remote.origin.fetch`: plain form → `origin/main` UNRESOLVED, explicit refspec → resolved. The issue Log's "both fetch sites use the explicit refspec" is a false evidence claim about the third and only site BR-16 named.
- **BR-19** — the fix is in the code and is correct, but nothing pins it. I reverted `|| len(trunk[id]) > 1` in a scratch copy of `2a69212` and `go test ./cmd/sdlc -run 'TestIntroduced|TestMerged|TestRefuseDuplicate|TestAllocateIssueID|TestDuplicateIDsInRef'` returned `ok`. Per ariadne#194 that is `not-addressed`, and it is the **2nd finding in family `fix-not-pinned-by-a-failing-test`** — the rule is that a change to the gate predicate ships a table row varying the axis it changed. The table at `issueids_test.go:546` still has no row where `trunk[id]` holds two paths.

## 4. Minor findings

- `repoRelativeIDDirs` containment uses `strings.HasPrefix(rel, "..")`, which also refuses an in-repo dir whose name starts with `..` (`filepath.Rel("/repo","/repo/..hidden")` = `"..hidden"` → refused). Reachable via `--issues-dir`/`WF_ISSUES_DIR`; falls back to the local scan.
- BR-11's site is worse than recorded: the 4.6 bypass (`merge.go:339`) prints nothing at all, while 4.5's `--no-validate` branch emits a loud `cwarn`.
- BR-12's class has a second site: `refuseDuplicateIssueIDs` fetches at `issueids.go:262` on the interactive `sdlc merge` path, also unbounded (`execGitRunner.Git` is a bare `exec.Command`, `runner.go:36`).
- The atlas prose that BR-9 flags is trapped in a code fence in `ci-merge-check.md` — in a window whose other half (#211) is entirely about fence-aware parsing.

## 5. Test coverage notes

- `go test ./cmd/...` at HEAD: one failure, `TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory` (missing `workshop/plans/000200-…-plan.md`) — pre-existing #210, unrelated.
- The uncovered axes are exactly where the open findings live: no test drives `runIssueLintIDs` in-process (it calls `exitWithCode(1)` directly, so the pre-existing/introduced labelling and the exit code are only observable through the shell adapter); no test makes the *baseline* read fail (BR-20's `gitx.Capture` escape is the reason); no fixture varies `trunk[id]`'s duplicate count; and the offline test asserts the warning only in the one repo state where it can fire.
- The four fail-injecting runners I wrote to reach these paths (`ls-tree` fails for one named ref) are three lines each and would slot straight into `issueids_test.go` beside `lsTreeFailRunner`.

## 6. Architectural notes

- **ARCH-DRY — flag.** BR-8, unchanged and now larger: three listing parsers (`IDsInTreeListing`, `DuplicateIDsInRef`, `issueFilesByID:329-350`), three `rev-parse + repoRelativeIDDirs + per-dir ls-tree` shells (`publishedIssueIDs`, `issueFilesByID`, `idListing`), and two byte-identical clash-report `Sprintf`s (`issueids.go:285`, `issuelintids.go:129`). One pure `PathsByIDInTreeListing` plus one IO shell returning `(map, freshness, error)` discharges BR-8, BR-15 and C1 together.
- **ARCH-PURE — pass, with one leak.** The decision core is genuinely pure and directly unit-tested. The leak is BR-20: `refuseDuplicateIssueIDs` takes an injected `gitRunner` and then reads its baseline through `gitx.Capture`, which is both an untestable seam and an error-erasing one.
- **ARCH-PURPOSE — flag.** Two deferred purposes, not separable extensions. The propagation run has not happened (`../parley.nvim/scripts/merge-checks.d/` still holds only `.gitkeep`, and parley.nvim carries four of the eight collisions the issue was opened over) — BR-21. And the enforcement half's authoritative baseline can silently degrade to the one BR-1 proved blind, which means the "server-side guarantee" is conditional in a way nothing announces.
- **ARCH-MOCK — pass.** Real repos and a real bare origin throughout, and the fixtures encode the load-bearing ordering. The only gap is that failure modes of the dependency (a fetch that fails, an `ls-tree` that fails) are modelled at one site and not at the others — the class fix in C1 is also what makes them uniformly injectable.
- **ARCH-CONSTRAINTS — flag.** `sdlc issue new` and `sdlc merge` both block on an unbounded `git fetch`; no budget is stated anywhere in the issue. BR-12.
- **ARCH-SECURE — pass with a note.** No credentials in the diff; `repoRelativeIDDirs` is a genuine boundary parse and the reason it exists is documented. The note is that a stale ref is untrusted input treated as well-formed and current — provenance, not location — which is the framing C1 asks for.

## 7. Plan revision recommendations

`workshop/issues/000213-nextid-from-origin.md` needs a `## Revisions` (or Log) entry recording three corrections, because two of these are false evidence claims that a future reader would otherwise trust:

1. **Correct the BR-16 claim.** The round-3 Log says "BR-16 re-raised but verified present — both fetch sites use the explicit `+refs/heads/main:refs/remotes/origin/main` refspec." The site BR-16 named — `scripts/merge-checks.d/40-duplicate-issue-id.sh:65` — was never changed. "Both Go sites" is true; "both fetch sites" is not.
2. **Correct the offline Done-when.** "With origin unreachable, creation still succeeds and emits the warning; a test asserts the warning is present" holds only for a repo with no `refs/remotes/origin/main`. Restate it as the two cases (no ref → warns; stale ref → currently silent, C1) so the gap is on the record rather than claimed closed.
3. **Split the derivative Done-when** into "mechanism declared" (`base.manifest` symlink row — done) and "propagated to derivatives" (open, one `sdlc propagate-base` run), per BR-21. As written the row claims a state that `ls ../parley.nvim/scripts/merge-checks.d/` contradicts.

Also note for the record: this issue has no durable design plan in `workshop/plans/` (only the two generated gate ledgers), so the Core-concepts cross-check has no table to verify against. That is the already-documented `change-code` deviation, not a new gap — but it is why five review rounds have been the only structural check on this design.

```findings
dispose:
  - id: BR-8
    disposition: not-addressed
    note: |
      Unchanged at HEAD: three listing parsers, three rev-parse+ls-tree shells, two byte-identical clash-report Sprintfs (issueids.go:285, issuelintids.go:129).
  - id: BR-9
    disposition: not-addressed
    note: |
      Verified at HEAD — the added prose still sits between the opening fence at ci-merge-check.md:30 and its close at :46.
  - id: BR-10
    disposition: not-addressed
    note: |
      helptext/issue.md SUBCOMMANDS still lists new/sync/set-status/list/show only; lint-ids absent.
  - id: BR-11
    disposition: not-addressed
    note: |
      merge.go:339 still gates on !f.NoValidate, and unlike step 4.5 the bypass branch prints nothing at all; merge.md FLAGS documents neither.
  - id: BR-12
    disposition: not-addressed
    note: |
      execGitRunner.Git is a bare exec.Command (runner.go:36); unbounded at issueids.go:85 and also at :262 on the interactive merge path.
  - id: BR-15
    disposition: not-addressed
    note: |
      Measured at HEAD - a runner failing only the trunk ls-tree makes introducedIDClashes return clashes=[] err=nil on a real collision the healthy control refuses.
  - id: BR-16
    disposition: not-addressed
    note: |
      Script line 65 is byte-identical to its original commit; premise re-measured (narrow refspec - plain form leaves origin/main UNRESOLVED, explicit refspec resolves).
  - id: BR-17
    disposition: not-addressed
    note: |
      ci-merge-check.md still claims the script builds from the checkout under test and lists no-cmd/sdlc as a skip; delivery table still says scaffold; the three-tree predicate has no atlas home.
  - id: BR-18
    disposition: not-addressed
    note: |
      Primary fix confirmed and revert-pinned; both named sweep members reproduce - base collapse falsely refuses a git-mv archive with EMPTY stderr, and lint-ids labels an introduced duplicate pre-existing.
  - id: BR-19
    disposition: not-addressed
    note: |
      Code fix is correct but unpinned - reverting the trunk clause in a scratch copy of 2a69212 leaves the whole suite green. 2nd in family fix-not-pinned-by-a-failing-test.
  - id: BR-20
    disposition: not-addressed
    note: |
      issueids.go:275 still calls gitx.Capture directly while the rest of the function uses the injected runner.
  - id: BR-21
    disposition: not-addressed
    note: |
      ../parley.nvim/scripts/merge-checks.d/ still contains only .gitkeep; no propagate-base run in this window.
  - id: BR-22
    disposition: not-addressed
    note: |
      The +59 lines added to workshop/lessons.md in this window are all 211's; no 213 entry.
findings:
  - id: new
    severity: Critical
    family: silent-degradation-in-allocator
    title: |
      Offline with a stale origin/main, sdlc issue new re-allocates a published id with no warning - the original bug, through the fix
    detail: |
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
  - id: new
    severity: Minor
    family: dir-containment-false-negative
    title: |
      repoRelativeIDDirs tests containment with a string prefix, so an in-repo dir named ..something is refused
    detail: |
      This is the 2nd finding in family dir-containment-false-negative. The rule covering both: a
      containment check compares path COMPONENTS, never a string prefix, and a failure to establish
      containment must not degrade silently. issueids.go:239 uses strings.HasPrefix(rel, ".."), and
      filepath.Rel("/repo", "/repo/..hidden") returns "..hidden" - refused, then silently downgraded to
      the local scan. Reachable via --issues-dir / WF_ISSUES_DIR. Use rel == ".." ||
      strings.HasPrefix(rel, ".."+string(filepath.Separator)).
```

---

## Re-review — 2026-09-03T18:37:57-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 213 — Allocate issue IDs against origin/main |
| repo | ariadne |
| issue file | workshop/issues/000213-nextid-from-origin.md |
| boundary | whole-issue close |
| milestone | — |
| window | 9078f8d0529b0c9fd4312f435b10f731aa5d3dc2..3ad17ffdaf7ed6283ede7d0f243b42a9a25599c4 |
| command | sdlc close --issue 213 |
| reviewer | claude |
| timestamp | 2026-09-03T18:37:57-07:00 |
| verdict | REWORK |

## Review

I've completed the review. Measured everything against the code and a real repo rather than the Log.

```verdict
verdict: REWORK
confidence: high
```

Five rounds in, the three-tree merge model is genuinely right and pinned, and the test discipline (real repo + real bare origin everywhere) is the strongest part of this work. But probing rather than reading turned up a **new, reproducible instance of the issue's own bug**: run `sdlc issue new` from any subdirectory of the repo and the trunk read resolves to `<subdir>/workshop/issues`, returns *nothing*, and the allocator hands out `000001` — then commits and pushes it to `origin/main` under a green `[ok] Issues synced and pushed`. Separately, of the 14 open findings only **BR-15** is actually addressed: BR-16's fix landed at the two Go sites but not at the CI script that the finding names (measured: `git fetch origin main` still leaves `origin/main` unresolvable under a narrow refspec, and the script falls back to the BR-1-blind baseline with exit 0); BR-19's one-line fix is present but **unpinned** — reverting `|| len(trunk[id]) > 1` leaves every test green; BR-23's warning is likewise unpinned — deleting the `cwarn` entirely leaves the suite green — and 7 of its 8 enumerated sites remain. That is the pattern that produced this issue's own round-3 lesson, so it blocks.

## 1. Strengths

- **`mergedPathsFor` is the right model and it is properly pinned.** Reverting the symmetric half (`|| !containsPath(trunk[id], p)`, `cmd/sdlc/issueids.go:142`) in a scratch copy turns `TestMergedPathsFor_ModelsTheMergeResult/trunk_archived_it_while_the_PR_was_open` red with the exact merged-map dump. That is a real revert-verification, and the table row that used to encode the defect as the expectation now encodes the fix.
- **ARCH-MOCK is exemplary here.** `cmd/sdlc/issueids_test.go:16-22` states why a function mock cannot express this bug, and every test drives a real repo against a real bare origin. My own probes reproduced against the same shape without any test-harness scaffolding — that is what a good seam buys.
- **The pure core is genuinely pure** (ARCH-PURE): `NextID`, `IDsInTreeListing`, `DuplicateIDsInRef`, `mergedPathsFor`, `introducedCollisions` all table-test with zero IO. Splitting `introducedIDClashes` out of `runIssueLintIDs` so the decision is testable without cobra or `os.Exit` is the correct seam.
- **Logic in Go, script as a four-line adapter** (`scripts/merge-checks.d/40-duplicate-issue-id.sh`) — filename parsing and directory selection decided once. Every skip condition evaluated before the first side effect is a real ordering fix, correctly ordered in the file.
- **`atlas/workflow/sdlc-binary.md:190-235`** is a strong entry: the three-layer strength table and the trunk-tip-vs-merge-base argument are exactly what a future reader needs.

## 2. Critical findings

**C1 — `sdlc issue new` from a subdirectory allocates `000001` from a vacuously-empty trunk read, misfiles the issue outside the tracker, and publishes it. All layers green.** (`cmd/sdlc/issueids.go:216-247`, `cmd/sdlc/internal/issue/scaffold.go:71`)

This is the **4th finding in family `silent-degradation-in-allocator`**, so per the escalation rule: do not fix this instance. The rule that covers all four — BR-7 (partial `ls-tree`), BR-15 (three read sites), BR-23 (stale ref), and this — is:

> Every read feeding the id space must be verified **fresh**, **complete**, *and* **on-target**, or announced as degraded. A read that resolves to a path the ref does not contain is a **non-answer, not an empty answer**, and must never be unioned in as zero ids.

`repoRelativeIDDirs` joins the caller-supplied relative dirs onto `os.Getwd()`, not onto the repo top-level. From `docs/sub/` that yields `docs/sub/workshop/issues`, which passes the containment guard (it is inside the repo) and then `ls-tree`s to nothing — indistinguishable from a repo with no issues. `ScanLocalIDs` degrades identically via `os.ReadDir` on the same relative path.

Measured, real repo + bare origin, `sdlc` built at `3ad17ff`:

```
$ cd docs/sub && sdlc issue new "subdir run"
  workshop/issues/000001-subdir-run.md
  [ok] Issues synced and pushed to origin/main.

$ git ls-tree -r --name-only origin/main | grep 0000
docs/sub/workshop/issues/000001-subdir-run.md      ← published, misfiled, id 1
workshop/issues/000042-existing.md
workshop/issues/000043-root-run.md
```

No warning on stderr. The printed path reads as the canonical location. All three enforcement layers are blind to it — the CI script `cd`s to the top-level and `ls-tree`s only the three canonical dirs, so `docs/sub/workshop/issues/` is invisible, and the natural repair (`git mv` it into `workshop/issues/`) manufactures exactly the collision this issue exists to prevent. Fix at the class: resolve relative id dirs against `gitx.RepoTopLevel()` (not the cwd) in one place that both `ScanLocalIDs` and `repoRelativeIDDirs` consume, and make "dir does not exist in the ref *and* does not exist on disk" a degraded read that routes to the loud warning rather than contributing zero ids.

**C2 — BR-23 not addressed: 1 of 8 enumerated sites fixed, and that one's deliverable is unpinned.** The behavior change in `allocateIssueID` (`cmd/sdlc/issueids.go:50-62`) is correct, but measured in a scratch copy: disabling the entire stale-warning branch (`if false && ferr == nil && stale != nil`) leaves `go test ./cmd/sdlc -run 'TestAllocateIssueID|TestPublishedIssueIDs'` **green**. `TestPublishedIssueIDs_StaleRefIsReportedStale` asserts the internal `stale` return value, not the warning — and the warning is what Done-when promises. The seven remaining sites are unchanged: `issueids.go:262` (`_, _ = r.Git("fetch", …)`), `:275` (`gitx.Capture` → `""` on error), `:277-281` (`berr` dropped), `:266`/`:271` (read failure → `return nil`, merge proceeds), `issuelintids.go:122` (`terr` dropped, base substituted for trunk), `:77`/`:91` (read failure → exit 0 — a **green required status check** from a read that failed), and the script's `65-69,80-81`.

**C3 — BR-18's class sweep is incomplete.** The Critical predicate is fixed and revert-pinned (see Strengths), but both named sweep members remain. Member 2 measured end to end:

```
$ sdlc issue lint-ids --base $mergebase --trunk origin/main --head HEAD
  [!] pre-existing duplicate id #000002: …000002-aaa.md, …000002-bbb.md   ← the RANGE created both
  this range reuses 1 issue id(s) that already exist at c712544…          ← id 2 exists nowhere at base
```

One id, reported simultaneously as pre-existing *and* as introduced, and the refusal headline names a base the id is absent from. The verb's own `Long` help promises these are different verdicts. Member 1 (`issueids.go:265-271`: an unresolvable or unreadable merge-base leaves `base` empty, erasing all deletion information so every archive is refused) is also unchanged — and BR-20 is why it cannot be pinned through the seam.

## 3. Important findings

**I1 — BR-16 not addressed at the site it names.** The two Go fetches use `+refs/heads/main:refs/remotes/origin/main`; `scripts/merge-checks.d/40-duplicate-issue-id.sh:65` still uses plain `git fetch --quiet origin main`. Reproduced: in a checkout with a narrow `remote.origin.fetch`, the plain form leaves `origin/main` unresolvable while the explicit refspec resolves it. The script then falls through to `--base "$fallback_base"` — the merge-base BR-1 proved structurally blind — prints one stderr line, and **exits 0**. The Log's "both fetch sites use the explicit refspec" is true of the Go sites and false of the CI site. Fix: explicit refspec in the script, and make the no-trunk fallback a loud degraded state, not a pass.

**I2 — BR-19's fix is not pinned by any test.** Reverting `|| len(trunk[id]) > 1` (`cmd/sdlc/issueids.go:196`) to `len(base[id]) > 1` alone leaves `TestIntroducedCollisions_*`, `TestMergedPathsFor_*`, `TestIntroducedIDClashes*`, `TestRefuseDuplicate*` and `TestMergeCheckScript_*` all **PASS**. Per the claimed-fixes protocol that is `not-addressed` regardless of how correct the diff reads. BR-19 named the missing coverage precisely: no table row varies trunk's duplicate count. Add `{base: {1:[a]}, head: {1:[a]}, trunk: {1:[a,b]}, wantIDs: nil}` to `issueids_test.go:546`.

## 4. Minor findings

- BR-8 — three listing parsers + two identical rev-parse/ls-tree shells remain; add a third site: the clash-rendering loop is copy-pasted between `issueids.go:279-284` and `issuelintids.go:126-131` (ARCH-DRY).
- BR-9 — `atlas/workflow/ci-merge-check.md:31-45` still renders inside the fence opened at line 30.
- BR-10 — `lint-ids` still absent from `cmd/sdlc/helptext/issue.md:6-14`.
- BR-11 — still bundled behind `--no-validate`; `merge.md` FLAGS documents neither `--no-validate` nor the gate.
- BR-12 — unbounded `git fetch` on `issue new`, now with a second site: `sdlc merge` step 4.6 adds another to the merge path (ARCH-CONSTRAINTS).
- BR-17 — 0 of 5 named doc homes fixed, and the enumeration is now **8, six wrong**: add `cmd/sdlc/issue.go:200` ("scanning issues/ + history/" — the pre-#213 description), `cmd/sdlc/helptext/fetch.md:17` (same), and `atlas/workflow/sdlc-binary.md` (no mention of the three-tree merge model — the invariant four rounds fought over, and the one future work will break).
- BR-20 — `gitx.Capture` bypasses the injected `gitRunner`; blocks pinning C3's member 1.
- BR-21 — measured: `parley.nvim/scripts/merge-checks.d/` holds only `.gitkeep`, as do all 8 other ariadne-styled peers. The check reaches **zero** derivatives.
- BR-22 — no `workshop/lessons.md` entry for #213; the +59 lines in this window are all #211's.
- BR-24 — `strings.HasPrefix(rel, "..")` at `issueids.go:239` still refuses an in-repo `..something`.

## 5. Test coverage notes

Real-repo/bare-origin discipline is right and the pure tables are the correct shape. The gap is systematic rather than local: **there is no fail-injecting `gitRunner`**, so not one of the seven degradation sites in C2 can be pinned, and the two unpinned fixes (I2, C2) are both cases where the test was written against the same mental model as the fix. Two concrete additions close most of it — a `gitRunner` wrapper that fails a named subcommand, and assertions on `stderr` (the operator-visible deliverable) rather than on internal return values.

## 6. Architectural notes

- **ARCH-DRY — flag** (BR-8, +the clash-render duplication).
- **ARCH-PURE — pass, one note.** The core is genuinely pure. But `repoRelativeIDDirs` fuses `Getwd`/`EvalSymlinks`/`RepoTopLevel` with the containment *decision*, which is why both BR-24 and C1 are bugs in an untestable-without-a-repo function. Extract `containedRel(top, base, dirs) ([]string, error)` as pure.
- **ARCH-PURPOSE — flag.** Shadow-sweep of `sdlc issue lint-ids` as the single source: CI script derives ✓, `sdlc merge` derives ✓, **derivative repos do not** (BR-21, zero), **docs are hand-maintained restatements and six are wrong** (BR-17). And C1 means the stated purpose — never silently allocate a published id — is unfulfilled on a reachable path.
- **ARCH-MOCK — pass**, with the fail-injection gap above.
- **ARCH-CONSTRAINTS — flag** (BR-12, two sites now).
- **ARCH-SECURE — pass.** `ls-tree` output is C-quoted under the default `core.quotePath`, but the quote precedes the last `/` so `IDFromFilename` still matches, and all three reads use identical quoting so path comparison stays consistent. The script builds and runs code from the PR checkout, but the runner already executes `merge-checks.d/*` from it, so no new trust boundary.

## 7. Plan revision recommendations

The issue's `## Plan` has all 8 rows ticked, but `## Done when` claims two things the code does not deliver. Add a `## Revisions` entry:

- **Split the derivatives Done-when row.** "The check reaches derivatives … Verified by running it in parley.nvim" → *mechanism declared* (`base.manifest` symlink row — done) and *propagated* (open; `sdlc propagate-base` not run; measured 0 of 9 peers carry the file).
- **Qualify the offline Done-when row.** "a test asserts the warning is present" is true for the absent-ref shape only; the stale-ref shape — the common one — asserts a return value, and the warning survives deletion with the suite green.
- **Record C1 as a scope item, not a deferral**: allocation is only trunk-backed when the verb runs from the repo root.

```findings
dispose:
  - id: BR-8
    disposition: not-addressed
    note: |
      Three parsers and two IO shells unchanged; clash-rendering is now a third duplicated block.
  - id: BR-9
    disposition: not-addressed
    note: |
      ci-merge-check.md:31-45 still sits inside the fence opened at line 30.
  - id: BR-10
    disposition: not-addressed
    note: |
      helptext/issue.md SUBCOMMANDS still lists only new/sync/set-status/list/show.
  - id: BR-11
    disposition: not-addressed
    note: |
      Still behind --no-validate; merge.md FLAGS documents neither flag nor gate.
  - id: BR-12
    disposition: not-addressed
    note: |
      No bound added; sdlc merge step 4.6 adds a second unbounded fetch.
  - id: BR-15
    disposition: addressed
    note: |
      All three read sites (publishedIssueIDs, issueFilesByID, idListing) now return errors.
  - id: BR-16
    disposition: not-addressed
    note: |
      Go sites fixed; the CI script at line 65 still uses plain `git fetch origin main` and falls back to the blind baseline with exit 0. Reproduced.
  - id: BR-17
    disposition: not-addressed
    note: |
      0 of 5 named homes fixed; enumeration now 8 homes, 6 wrong (add issue.go:200, helptext/fetch.md:17, sdlc-binary.md missing the merge model).
  - id: BR-18
    disposition: not-addressed
    note: |
      Headline predicate fixed and revert-pinned; both named sweep members remain (empty-base erasure; introduced duplicate reported as pre-existing — measured).
  - id: BR-19
    disposition: not-addressed
    note: |
      Code change present and correct, but reverting `|| len(trunk[id]) > 1` leaves every relevant test green — no test pins it.
  - id: BR-20
    disposition: not-addressed
    note: |
      issueids.go:265 still reads the baseline via gitx.Capture outside the injected runner.
  - id: BR-21
    disposition: not-addressed
    note: |
      Measured — parley.nvim and all 8 other peers still hold only .gitkeep; the check reaches zero derivatives.
  - id: BR-22
    disposition: not-addressed
    note: |
      workshop/lessons.md gained only 211's entries in this window; no 213 entry.
  - id: BR-24
    disposition: not-addressed
    note: |
      issueids.go:239 still uses strings.HasPrefix(rel, "..").
findings:
  - id: new
    severity: Critical
    family: silent-degradation-in-allocator
    title: |
      Run from a subdirectory, sdlc issue new reads an empty trunk id space, allocates 000001, misfiles the issue, and pushes it — silently
    detail: |
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
```

---

## Re-review — 2026-09-03T19:06:36-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 213 — Allocate issue IDs against origin/main |
| repo | ariadne |
| issue file | workshop/issues/000213-nextid-from-origin.md |
| boundary | whole-issue close |
| milestone | — |
| window | 9078f8d0529b0c9fd4312f435b10f731aa5d3dc2..3d27a19f6acd330b3516c901f98f3c0151839e4e |
| command | sdlc close --issue 213 |
| reviewer | claude |
| timestamp | 2026-09-03T19:06:36-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The three Criticals this round claims (BR-18, BR-23, BR-25) are genuinely fixed and I verified each by reverting exactly the production change and watching a named test go red — the round-6 collapse to one parser (`issue.PathsByID`), one reader (`refIDSpace`) and one dir resolution (`resolveIDDirs`) is real, and it is the right structural answer to why the same silent-degradation defect had to be found four times. What blocks SHIP is not correctness of the allocator: it is that six rounds have now produced a standing residue that this round did not touch. `docs-lag-new-surface` has four findings and zero were fixed (the atlas prose is still trapped in a code fence, `lint-ids` is still absent from the helptext, `merge.md` documents neither `--no-validate` nor the new gate); the derivative propagation the Done-when claims as verified has still not happened (`../parley.nvim/scripts/merge-checks.d/` holds only `.gitkeep`); the CI script still fetches with the bare refspec BR-16 measured as insufficient and still degrades to the BR-1-blind baseline with a green exit; and BR-19's correct one-line fix is pinned by nothing — I reverted it and the entire `cmd/sdlc` suite stayed green. One new Important: `runIssueNew` now hands the issue-sync an absolute `IssuesDir`, which silently kills the main-worktree cleanliness precheck and breaks publication from a subdirectory — measured on real repos with a real bare origin.

## 1. Strengths

- **`cmd/sdlc/issueids.go:149` `refIDSpace` + `internal/issue/scaffold.go:105` `PathsByID`** — the collapse from three parsers and three readers to one of each is the correct response to a defect family, not to a defect. Confirmed by grep: exactly one `ls-tree` call site remains in non-test code, and `IDsInTreeListing` / `issueFilesByID` / `idListing` / `repoRelativeIDDirs` are gone.
- **`issueids.go:304` `resolveIDDirs`** — anchoring to `gitx.RepoTopLevel()` and returning `Rel`+`Abs` from one resolution is exactly right, and the two tests pin it hard. Reverted to `os.Getwd()`: `TestAllocateIssueID_FromASubdirectoryReadsTheRealIDSpace` fails with `= 000001, want 000043`, and `TestRunIssueNew_FromASubdirectoryWritesToTheRepoIssueDir` fails on all three assertions.
- **`issueids.go:207` `mergedPathsFor`** — the symmetric predicate is pinned. Reverting line 219 to `!containsPath(head[id], p)` alone turns `TestMergedPathsFor_ModelsTheMergeResult/trunk_archived_it_while_the_PR_was_open` red with the exact false refusal. The three-tree model with its two rejected predecessors documented inline is the best-explained code in the diff.
- **`issueids.go:136` fetch-failure capture** — reverting to `_, _ = r.Git("fetch", ...)` turns `TestPublishedIssueIDs_StaleRefIsReportedStale` red. This was the sharpest finding of six rounds and the fix is properly seamed and properly pinned.
- **`atlas/workflow/sdlc-binary.md`** — the four-row BR-7/15/23/25 table is a genuinely good piece of architectural writing: it records the *rule* and the four instances that produced it, which is what makes the family legible to the next reader.

## 2. Critical findings

None outstanding. BR-18, BR-23's named instance and BR-25 are fixed and revert-verified.

## 3. Important findings

**`cmd/sdlc/issue.go:275` — `runIssueNew` hands the issue-sync an `IssuesDir` the sync cannot use.** `f.IssuesDir = writeDir` makes it absolute on *every* `sdlc issue new`, and the sync consumers were not swept. Measured on real repos with a bare origin:

- `git -C <main-worktree> diff --name-only -- /abs/<feature-worktree>/workshop/issues/` → `fatal: ... is outside repository`. `mainHasUncommittedIssueChanges` (`claim.go`) swallows that with `continue // mirror shell || true`, so `mainDirty` is empty and the precheck reports clean from a read it could not perform. Probe: main worktree carrying an uncommitted edit to `000001-one.md`, `sdlc issue new` on `feature` → the "main worktree has uncommitted issue changes. Commit or stash them first" refusal never fires; the operator gets a raw `cannot pull with rebase` instead.
- From a subdirectory the publish route breaks outright. `changedIssueFiles` returns `../../workshop/issues/000002-….md` (git prints `ls-files` paths relative to cwd), and step 6's `filepath.Join(wtRoot, c)` escapes the repo: `read /private/tmp/claude-501/workshop/issues/000002-subdir-on-branch.md: no such file`. It falls back to a local commit, so #82's "a freshly filed issue is tracker state on main" no longer holds on the very path BR-25 made supported. `TestRunIssueNew_FromASubdirectoryWritesToTheRepoIssueDir` asserts the write location but never that the reservation reached origin.

Fix sketch: run the sync with the repo-relative dir *and* from the repo top (or make `changedIssueFiles` emit repo-relative paths), and sweep the enumerable consumer list — `syncPathspec`, `changedIssueFiles`, `mainHasUncommittedIssueChanges`, the step-5 `diff … -- IssuesDir+"/"`, the step-6 copy loop, and the conflict guide's printed `git add <dir>/`. Separately, `mainHasUncommittedIssueChanges` should distinguish "clean" from "could not read".

**BR-16 and BR-19 are re-raised below** — see dispositions.

## 4. Minor findings

- `issueids.go:327` — `strings.HasPrefix(rel, "..")`; the repo already has the correct component-wise form at `reviewwindow.go:103`, and `migrate.go:236,282` carry the same defect (BR-24, class unswept).
- `issueids.go:155` — `git ls-tree --name-only <ref> <dir>/` and `rev-parse --verify --quiet <ref>` pass caller-supplied values with no `--end-of-options` / `--` separator; a ref or `--issues-dir` beginning with `-` is parsed as an option (ARCH-SECURE prefers structural separation to trusting the value).
- `issueids.go:51` and `issue.go:275` both call `resolveIDDirs` for the same invocation; harmless, but the second could take the already-resolved `idDirs`.
- `resolveIDDirs` calls `gitx.RepoTopLevel()` directly rather than through the injected `gitRunner` — same seam leak as BR-20.

## 5. Test coverage notes

- Revert-verified green (fix present, test red without it): BR-18 headline + both sweep members, BR-23, BR-25 ×2.
- Revert-verified **red** (fix present, suite stays green without it): BR-19's `|| len(trunk[id]) > 1`. Full `go test ./cmd/sdlc/` with it reverted → only the pre-existing `TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory` (#210) fails.
- `TestRefuseDuplicateIssueIDs_UnknownBaseSkipsRatherThanRefuses` pins the loud skip, but its orphan-branch fixture does not reproduce the false refusal its own comment describes — with `base = {}` the orphan's identical path dedupes to one claimant. A fixture where head archives an issue would exercise the refusal the comment claims.
- Suite state: `go test ./cmd/sdlc/...` — one failure, `TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory`, pre-existing #210, unrelated.

## 6. Architectural notes

- **ARCH-DRY — pass.** One parser, one reader, one dir resolution, one renderer (`renderClashes`), logic in a Go verb with the CI script as an adapter. Best-executed principle in the diff.
- **ARCH-PURE — flag (Minor).** The decision layer (`NextID`, `PathsByID`, `mergedPathsFor`, `introducedCollisions`, `classifyDuplicates`) is genuinely pure and unit-tested without git. But `gitx.Capture` at `issueids.go:388` and `gitx.RepoTopLevel()` at `:305` escape the injected `gitRunner`, which is why the merge-base failure path cannot be driven by a fake (BR-20).
- **ARCH-PURPOSE — flag (Important).** Shadow-sweep of the single-source resolution: read ✅, write ✅, publish ❌ (new finding §3). Shadow-sweep of the enforcement propagation: manifest row ✅, derivative symlink ❌ (BR-21) — Done-when claims a consumer that does not derive. Shadow-sweep of BR-23's eight-site enumeration: 4 fixed, 4 open.
- **ARCH-MOCK — pass.** Tests drive real repos against a real bare origin throughout, plus `lsTreeFailRunner` for fault injection at the seam. This is the right call — a function-call mock cannot express "a ref exists that the worktree does not contain".
- **ARCH-CONSTRAINTS — flag (Minor).** `sdlc issue new` now performs an unbounded network fetch on an interactive path (BR-12); no budget is stated anywhere.
- **ARCH-SECURE — flag (Minor).** No credentials involved. Untrusted-input handling is sound (`PathsByID` tolerates anything). The gap is structural separation in argv (§4).

## 7. Plan revision recommendations

Add a `## Revisions` entry to `workshop/issues/000213-nextid-from-origin.md` recording:

- **Done-when "The check reaches derivatives" is not delivered.** `base.manifest` declares the symlink; `../parley.nvim/scripts/merge-checks.d/` still contains only `.gitkeep`. Split into "mechanism declared" (done) and "propagated to derivatives" (open), or run `sdlc propagate-base` this round.
- **Done-when "so the guarantee survives a GitHub-UI merge, a bare `gh pr merge`, `--no-validate`"** holds only once `merge-check` is a required status check, which the issue's own next-actions list as not done. State it as conditional.
- **The round-6 Log claims the two families were fixed.** Measured, `silent-degradation-in-allocator` is 4 of 8 sites; the Log should carry the residual enumeration (`issueids.go:370`, `issuelintids.go:76/81/102`, `40-duplicate-issue-id.sh:65-69,76-84`) rather than a completed-family claim.

```findings
dispose:
  - id: BR-8
    disposition: addressed
    note: |
      One parser (issue.PathsByID), one reader (refIDSpace), one dir resolution (resolveIDDirs); grep confirms a single ls-tree call site and the old helpers are gone.
  - id: BR-9
    disposition: not-addressed
    note: |
      atlas/workflow/ci-merge-check.md still opens a fence at line 31 and closes it at line 47, with the added prose at 34-46 inside it.
  - id: BR-10
    disposition: not-addressed
    note: |
      cmd/sdlc/helptext/issue.md SUBCOMMANDS (lines 6-14) still lists new/sync/set-status/list/show only; lint-ids and validate absent.
  - id: BR-11
    disposition: not-addressed
    note: |
      merge.go step 4.6 is still gated on f.NoValidate with no per-gate flag, and merge.md FLAGS documents neither --no-validate nor the id gate.
  - id: BR-12
    disposition: not-addressed
    note: |
      No timeout, deadline or --no-fetch escape on any git fetch in the diff; no stated budget.
  - id: BR-16
    disposition: not-addressed
    note: |
      The Go sites use the explicit refspec, but the finding named the SCRIPT — 40-duplicate-issue-id.sh:65 is still `git fetch --quiet origin main`, and the unresolvable-trunk path at 67-69/80-81 still degrades to the BR-1-blind merge-base baseline and exits 0 green.
  - id: BR-17
    disposition: not-addressed
    note: |
      Three of five doc homes still wrong (BR-9, BR-10, BR-11), ci-merge-check.md:44 still says the script keys on ./cmd/sdlc when it keys on $owner/cmd/sdlc, and a fourth surfaced — the Delivery table at line 19 still calls merge-checks.d/* `scaffold`, contradicting base.manifest's new symlink row.
  - id: BR-18
    disposition: addressed
    note: |
      Revert-verified three ways: asymmetric predicate reddens TestMergedPathsFor .../trunk_archived_it_while_the_PR_was_open; empty-base collapse reddens TestRefuseDuplicateIssueIDs_UnknownBaseSkipsRatherThanRefuses; blanket "pre-existing" reddens TestClassifyDuplicates_IntroducedIsNotCalledPreExisting.
  - id: BR-19
    disposition: not-addressed
    note: |
      The code fix (issueids.go:266 `|| len(trunk[id]) > 1`) is present and correct, but nothing pins it — reverting it and running the full `go test ./cmd/sdlc/` leaves the suite green except the pre-existing 210 failure. Per the claimed-fix rule, unpinned is not addressed; add a row to the issueids_test.go table varying trunk's duplicate count.
  - id: BR-20
    disposition: not-addressed
    note: |
      issueids.go:388 still calls gitx.Capture directly (and :305 gitx.RepoTopLevel), so a failed merge-base read is still indistinguishable from an absent one and cannot be driven through the fake; the "treat empty as degraded" half is done and pinned.
  - id: BR-21
    disposition: not-addressed
    note: |
      Measured this round — ../parley.nvim/scripts/merge-checks.d/ still contains only .gitkeep. The manifest row is a declaration; the propagation run has not happened, and Done-when still claims it as verified.
  - id: BR-22
    disposition: not-addressed
    note: |
      workshop/lessons.md gained 59 lines in this window, all 211's; no entry for the round-3 insight (a test table encoding the defect as its expected value).
  - id: BR-23
    disposition: not-addressed
    note: |
      The named instance is fixed and revert-pinned, and the one-parser/one-reader collapse is the right structural answer — but the finding asked for the CLASS, and 4 of its 8 enumerated sites remain: issueids.go:370 (`_, _ =` on the merge gate's fetch, so a stale trunk lets the gate pass with a confident [ok]); issuelintids.go:76/81/102 (read failure warns then exits 0, which is a GREEN required status check in CI); 40-duplicate-issue-id.sh:65-69,76-84 (see BR-16). No freshness value was introduced and the CI verb still has no non-clean exit for a degraded read.
  - id: BR-24
    disposition: not-addressed
    note: |
      issueids.go:327 is still strings.HasPrefix(rel, ".."). The repo already carries the correct component-wise form at reviewwindow.go:103, and migrate.go:236,282 carry the same defect — the class is enumerable and unswept.
  - id: BR-25
    disposition: addressed
    note: |
      Revert-verified: replacing gitx.RepoTopLevel() with os.Getwd() reddens both TestAllocateIssueID_FromASubdirectoryReadsTheRealIDSpace (000001 vs 000043) and TestRunIssueNew_FromASubdirectoryWritesToTheRepoIssueDir.
findings:
  - id: new
    severity: Important
    family: enforcement-does-not-propagate
    title: |
      runIssueNew's absolute IssuesDir does not reach the sync consumers — the main-worktree cleanliness precheck silently reports clean, and issue new from a subdirectory no longer publishes
    detail: |
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
  - id: new
    severity: Minor
    family: unseparated-argv-options
    title: |
      Caller-supplied refs and directories are passed to git without an option separator
    detail: |
      issueids.go:150-155 runs `git rev-parse --verify --quiet <ref>` and
      `git ls-tree --name-only <ref> <dir>/` with values that come from --base/--head/--trunk
      and --issues-dir/WF_ISSUES_DIR. A value beginning with `-` is parsed as an option rather
      than as data. ARCH-SECURE prefers structural separation over trusting the value: insert
      `--end-of-options` before the ref and `--` before the pathspec. Low exploitability here
      (both are operator-supplied), but it is a one-token fix and the same argv shape recurs
      wherever a ref reaches git.
```
