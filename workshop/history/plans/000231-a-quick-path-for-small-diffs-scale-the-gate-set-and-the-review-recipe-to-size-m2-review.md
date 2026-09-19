# Boundary Review — ariadne#231 (milestone M2)

| field | value |
|-------|-------|
| issue | 231 — A quick path for small diffs: scale the gate set and the review recipe to size |
| repo | ariadne |
| issue file | workshop/issues/000231-a-quick-path-for-small-diffs-scale-the-gate-set-and-the-review-recipe-to-size.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 88b34b640cbb5832edd0030655108cad0cc6f1a8..60962e2687a71c8067c1806c28f73647aabe8a4d |
| command | sdlc milestone-close --issue 231 --milestone M2 |
| reviewer | claude |
| timestamp | 2026-09-17T21:51:13-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

M2 does what its Plan claims. The shell is a pure, edge-tested core. The upgrade is written only at finalize. The shared-surface declaration is read as committed at base ∪ head. The `quick-flow:` field in the registry sets which principles the recipe uses, and full-flow prompts stay byte-identical. The end-to-end test runs the real verbs.

I ran `go test ./cmd/sdlc/... ./cmd/vocabulary/... -count=1` at 60962e2. Everything passes except `TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory`, which already fails on main (#210).

Two Important findings stop this from being a clean SHIP:

1. **A full-review REWORK doesn't carry over.** If the fix makes the net diff smaller, the next round goes back to the small-diff recipe, and the issue stays recorded as quick. I reproduced this in a scratch copy (details below).
2. **A quick branch in ariadne can loosen its own shell.** The shared-surface declaration protects itself, but not the code that enforces the shell. The operator's `sdlc` function rebuilds from the checkout, so a branch is judged by the binary it just changed.

Neither is a correctness crash, and both are cheap to fix.

**1. Strengths**
- The pure shell: `Measure` takes the file list apart from numstat, so a binary still counts as a file (`flow/shell.go:22`). `Size.Crossings` (`:45`) is pinned at "exactly at the limit" and "one past" for every limit by `TestCrossings`.
- The upgrade is composed into `newFM` (`close.go:581`) and written only by `applyClose`. A REWORK leaves the issue byte-identical, which `TestCloseQuickReworkThenReclose` asserts.
- `committedSurfaces`/`declarationAt` (`closeflow.go:136,160`) tell "absent" from "unreadable". An unreadable declaration becomes a crossing, so it fails toward the full flow. `TestCloseSurfacesUncommittedEditIgnored` shows the working tree is never read.
- The registry split is lossless. `ArchitectureBlock` delegates to `ArchitectureBlockFor` over every marker (`architecture.go:58-66`), so full prompts are unchanged, and a missing, duplicated or invalid `quick-flow:` field fails a test.
- `{{BOUNDARY_TAIL}}` removes the duplicated review tail. `hasCodePath` and `publishGateHasCodeSurface` are now loops over one per-path rule set. I checked them against the old logic and they behave the same.

**2. Critical findings**
None.

**3. Important findings**
- **The REWORK decision isn't sticky (ARCH-ORDER).** `closeflow.go:36,60` looks only at the current `base..HEAD` net diff. `gatestate.Round` doesn't record which recipe ran.
  - Scratch reproduction: 3 code files → full review → REWORK → a commit deleting `cmd/c.go` → the re-close runs the small-diff recipe, and the record stays `quick`.
  - The existing test only covers the case where no commit lands between rounds.
  - Fix: record the recipe on each boundary round. If an earlier round of this boundary ran the full recipe, treat that as a crossing. Add a test for REWORK → a fix that shrinks the diff → re-close.
- **The shared-surface declaration is incomplete (ARCH-SECURE).** It doesn't cover `flow/limits.go`, `shell.go`, `surfaces.go`, `churn/classify.go` or `closeflow.go`.
  - `close.md:210` and `.sdlc/shared-surfaces:8` both claim "a branch cannot loosen its own shell". That is false in ariadne, because `sdlc` is rebuilt from the working tree.
  - Fix: declare those files, and have `TestRepoDeclarationParses` find the shell's implementation files itself and require each to match.

**4. Minor findings**
- The shell's wording says "tests and docs excluded" and "docs are never code", but `cmd/**/*.md` counts as code.
- A surface pattern like `lua/parley/*/` is accepted by the parser but never matches anything.
- `declarationAt` repeats an existing git helper and drops `--end-of-options`.
- `f.skip("done-when")` names a gate key that `skip()` has no case for.
- The upgrade Log line lists every code file on one line, so it has no length limit.
- The wider `isTestPath` now treats `.colima/test/` as a test directory.

**5. Test coverage notes**
- The REWORK test covers only one ordering (see Important finding 1).
- "Exactly at the limit" is tested only in the pure `Size` tests, not end to end through real numstat. That's acceptable.
- Nothing tests `--force` bypassing the empty-Done-when refusal.
- `FuzzParseSurfaces` only checks for panics. That's low-risk: `path.Match` validates the whole pattern even against an empty name (checked on go1.27.1).

**6. Architectural notes for upcoming work (M3)**
- The `upgraded` ledger column should come from the round history once finding 1 is fixed. Otherwise an issue that had a full-recipe round enters calibration as a pure quick row.
- The `isTestPath` cut-over belongs in `ledger-landscape.md` as planned.

**7. Plan revision recommendations**
- **Core concepts:**
  - `judge.ArchitectureSection` is really the unexported `architectureSections` / `mustArchitectureSections`.
  - `Measure` takes a `surfacesErr` parameter, and `Crossings` is the method `Size.Crossings()`.
  - Add `flow.Upgrade`, `flow.Union` and `flow.DeclarationPath` to the table.
  - The close step lives in `cmd/sdlc/closeflow.go`, not inline in `close.go`.
  - The M2 integration tests are in `closeflow_test.go`, not `closereview_test.go`.
- **Decisions ("upgrade is recorded at finalize"), Spec §1 ("re-derives the upgrade from the same window") and `close.md`:** revise to match however finding 1 is resolved.
- **ARCH-SECURE bullet ("a branch cannot loosen its own shell"):** qualify it, or extend it to cover the shell's implementation.

```findings
findings:
  - id: new
    severity: Important
    family: round-decision-not-sticky
    title: |
      A full-recipe REWORK is not sticky: a fix that shrinks the net diff sends the next round back to the small-diff recipe, and the issue stays recorded quick
    detail: |
      closeFlowStep/decideCloseFlow (cmd/sdlc/closeflow.go:36,60) re-measure only the current base..HEAD net diff. A REWORK persists its round to the ledger, but gatestate.Round records nothing about which recipe ran. Reproduced in a scratch copy of 60962e2: 3 code files -> full review -> REWORK -> a commit deleting cmd/c.go -> the re-close dispatches the small-diff recipe, and the record stays {kind: quick}. The Spec/plan claim that the re-close "re-derives the upgrade from the same window" holds only when no commit lands between rounds. That is the one ordering TestCloseQuickReworkThenReclose exercises (ARCH-ORDER: one observable interleaving). Effects: the round after a deletion is judged under 3 principles, not 8. Open full-recipe findings go to a judge told to raise nothing outside those 3. M3 will count the issue as a pure quick row. Fix: stamp the recipe on each boundary round. Make "a prior round of this boundary ran the full recipe" a crossing, recorded at finalize like the others. Add a test for REWORK -> shrinking fix -> re-close. If the operator keeps today's behaviour, record that as a Revision and correct the "same window" claim in the Spec, the plan Decisions and close.md.
  - id: new
    severity: Important
    family: hand-listed-guard-incomplete
    title: |
      .sdlc/shared-surfaces guards the declaration but not the code that enforces the shell, and ariadne's sdlc is rebuilt from the working tree, so a quick branch can loosen its own shell
    detail: |
      This is the 2nd finding in family hand-listed-guard-incomplete. Rule: a guard that claims "the change cannot loosen X" must cover every artifact that defines X, and a test must derive that set rather than trust a list. Here X is the shell. Only the declaration is covered (base union head). Undeclared: flow/limits.go (MaxCodeFiles, MaxChangedLines), flow/shell.go and surfaces.go, churn/classify.go (IsCodeFile, isTestPath) and cmd/sdlc/closeflow.go. The operator's sdlc shell function runs go build ./cmd/sdlc from the checkout, where the default in-place branch lives. So a one-line quick edit raising MaxChangedLines closes under the binary it just changed. This contradicts close.md:210, .sdlc/shared-surfaces:8 and the plan's ARCH-SECURE bullet. The header's own criterion ("what the judges are told ... the boundary-review procedure") also omits boundary-tail.md, which this diff created and which carries the boundary contract, and judge/prompts/. Fix: declare cmd/sdlc/internal/flow/, cmd/sdlc/internal/churn/classify.go and cmd/sdlc/closeflow.go, and decide on the judge prompt files. Extend TestRepoDeclarationParses to glob the shell's implementation files and require Match for each.
  - id: new
    severity: Minor
    family: doc-claim-contradicts-code
    title: |
      The shell's prose says "tests and docs excluded" and "docs are never code", but IsCodeFile counts cmd/**/*.md, which IsDoc itself classifies as docs
    detail: |
      This is the 4th finding in family doc-claim-contradicts-code. Rule: prose restating what a classifier or gate does must be rendered from the owning package, or pinned by a test against that classifier's own exemplars. It must not be hand-written, and a test must not pin the wording alone. Instances: ShellSummary (flow/limits.go:20), rendered into start-plan, change-code help, close help and flowInfoLine; close.md:207-208; the limits.go:11 comment; and TestShellSummaryReadsTheConstants (flow_test.go:152), which pins the inaccurate "tests and docs". TestIsDoc marks cmd/sdlc/helptext/close.md as a doc, while TestIsCodeFile counts cmd/**/*.md as code. The plan's Core concepts table drift is the same class (see plan revisions). Fix: own the exclusion clause in churn beside IsCodeFile, with a test checking each clause against IsCodeFile exemplars, including the embedded-markdown exception. ShellSummary should use it, and close.md should use QUICK_SHELL instead of restating it.
  - id: new
    severity: Minor
    family: accepted-input-never-effective
    title: |
      A trailing-slash surface pattern containing a glob (lua/parley/*/) passes ParseSurfaces but Match compares it as a literal prefix, so it never matches
    detail: |
      surfaces.go validates the pattern with path.Match, but Match uses strings.HasPrefix for the trailing-slash form, so the declared guard silently guards nothing. Verified: "lua/*/" validates without error, and HasPrefix of lua/parley/x.lua is false. Reject glob metacharacters in the directory form at parse time (that becomes a crossing, failing toward full), or match the directory form segment-wise.
  - id: new
    severity: Minor
    family: existing-helper-not-reused
    title: |
      declarationAt re-implements the ls-tree presence plus blob read that already exists (gitx TrunkFile.entryOf/readFrom, pathTrackedAtHEAD)
    detail: |
      ARCH-DRY. closeflow.go:160 is the third copy of the question "is ref:path present, absent, or unknown" (gitx/trunkfile.go:255-287, synctrunk.go:319). It is also the only copy without --end-of-options. Extract gitx.FileAt(ref, path) and route all three through it.
  - id: new
    severity: Minor
    family: gate-key-undeclared
    title: |
      checkQuickDoneWhen calls f.skip("done-when"), a key closeFlags.skip has no case for, so it means "--force only" by silently falling through
    detail: |
      closeflow.go:85 against close.go:111-137. The behaviour is intended, but it rests on the switch default, so a reader takes done-when for a real gate key, and a typo'd key elsewhere fails the same way. Write f.Force explicitly, or add a case so TestCloseFlags_Skip covers it.
```

---

## Re-review — 2026-09-17T22:12:00-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 231 — A quick path for small diffs: scale the gate set and the review recipe to size |
| repo | ariadne |
| issue file | workshop/issues/000231-a-quick-path-for-small-diffs-scale-the-gate-set-and-the-review-recipe-to-size.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 88b34b640cbb5832edd0030655108cad0cc6f1a8..8225b15803582b2e02c2f1bc16bdf7372b1221ca |
| command | sdlc milestone-close --issue 231 --milestone M2 |
| reviewer | claude |
| timestamp | 2026-09-17T22:12:00-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

Five of the six open findings are fixed, and I confirmed each behaviour fix against a test that fails without it. For BR-18 and BR-21 I reverted the fix in a scratch worktree and watched the tests go red. For BR-19 I narrowed the declaration and watched `TestRepoDeclarationCoversTheShell` go red. `go test ./cmd/sdlc/... ./cmd/vocabulary/...` passes except `TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory`. That one fails on main too, because the `000200` plan file is missing. The one leftover is BR-22: `readFrom` still has its own blob read.

Two new Important findings stand in the way of SHIP, and both are cheap to fix:

- **A renamed shared surface escapes the shell.** The shell only sees the new path of a renamed file. I reproduced two cases at `sdlc close`. Moving a file out of a declared surface stays quick. Renaming `.sdlc/shared-surfaces` itself also stays quick, which quietly switches the shell off for every later branch once it merges.
- **The "gates don't change under the gate" rule stops at `cmd/sdlc/`.** The binary also compiles `pkg/frontmatter` and depends on `go.mod`/`go.sum`, and none of those are declared.

### 1. Strengths
- **Sticky flow across rounds.** The fix for BR-18 is modelled as state, not patched at one site. Each round records its `Recipe` (`boundaryledger.go:181-183`), and an earlier full round counts as a crossing. Tests cover both directions: a full REWORK followed by a shrinking fix stays full, and a small-diff REWORK stays quick.
- **Unstamped rounds read correctly.** `fullRoundIn` treats a round with no recipe (written before #231) as full and skips rounds that never ran. `TestFullRoundIn` pins that table.
- **Glob check at parse time.** `ParseSurfaces` now refuses a glob inside a directory entry, so the error becomes a crossing and fails toward the full flow.
- **`gitx.EntryAt` is shared.** It is the one ls-tree presence check, and the close step's copy now gets `--end-of-options`. Adding `--full-tree` to `TrunkFile.entryOf` changes nothing, because that dir is always the repo root (`synctrunk.go:41`).
- **Code-file rule lives in one place.** `churn.CodeFileRule` sits beside `IsCodeFile` and is tested against its examples, including markdown under `cmd/`. `ShellSummary` reads it, so the help text can't drift.

### 2. Critical findings
None.

### 3. Important findings
- **`cmd/sdlc/closeflow.go:124` → `flow/shell.go:35`: the surface check misses rename sources.** `gitx.DiffNames` (`git diff --name-only`) lists only a rename's new path. So `Measure` never sees the old path leaving a surface.
  - I ran both probe fixtures through the real close path. Both closed as `{kind: quick}` with the small-diff recipe.
  - This breaks the promise, made in the `Surfaces.Match` doc comment, the file header and `close.md`, that the declaration always matches itself.
  - It also hits the one limit the Spec says would have caught #263.
  - This is the first finding in the new `diff-paths-drop-rename-source` family. Enumeration: within this diff, the only consumer is `Measure` (code files and surfaces). Counting code files by the new path is correct; only the surface match needs both paths. The atlas gate's `atlasChanged` has the same issue, but it predates this diff.
  - **Fix:** match surfaces against both paths of every rename or copy (a `--no-renames --name-only` list, or `--name-status` with both paths). Keep counting code files by the new path. Add two fixtures: one moving a file out of a surface, one renaming the declaration.
- **`.sdlc/shared-surfaces:16` and `flow/surfaces_test.go:117`: the declaration and its test stop at `cmd/sdlc/`.**
  - **This is the 3rd finding in family `hand-listed-guard-incomplete`.**
  - `go list -deps ./cmd/sdlc` also includes `pkg/frontmatter`, which `internal/issue/frontmatter.go` uses to split the frontmatter that holds the flow record. The binary also depends on `go.mod`/`go.sum`. None of these are declared.
  - The test only walks `cmd/sdlc`, so the set it derives is still one the author chose.
  - **Rule:** derive a guard's protected set from the real dependency graph, not from a directory someone picked.
  - **Fix:** declare `pkg/frontmatter/` (or `pkg/`), `go.mod` and `go.sum`. Derive the test's set from `go list -deps -f '{{.Dir}}' ./cmd/sdlc`, filtered to this module, plus the module files.

### 4. Minor findings
- BR-22 remains: `gitx/trunkfile.go:283` runs the same `cat-file blob` call as the new `gitx.BlobAt`. Route it through `BlobAt`.
- After a sticky upgrade, the Log reason says only "an earlier round of this close already ran the full review". The original size reason is not repeated. That's acceptable.

### 5. Test coverage notes
- The tests for each limit, the rework cases, the stale and empty Done-when checks, uncommitted declaration edits, milestone-close upgrades and the end-to-end verb sequence all exist and pass.
- The gap is renames; see the first Important finding.

### 6. Architectural notes
- **ARCH-DRY:** flag (BR-22 leftover). Otherwise it passes: there is one classifier set per path, one shared prompt tail, `windowFileStats` is shared with the churn report, and `recipe()` is resolved in one place.
- **ARCH-PURE:** pass. `decideCloseFlow`, `fullRoundIn`, `Measure`, `Crossings` and `ParseSurfaces` are pure; the IO sits in `closeFlowStep` and its helpers.
- **ARCH-PURPOSE:** pass for M2. One thing for the operator to confirm: declaring `cmd/sdlc/` means nearly all of ariadne's own code work takes the full flow. Record that as a Decision.
- **ARCH-MOCK:** pass. Tests use real git in temp repos, with the judge stubbed at `judge.Run`.
- **ARCH-CONSTRAINTS:** pass. Each close adds one numstat, two ls-tree plus cat-file pairs and one ledger read. Crossing lists are capped.
- **ARCH-SECURE:** flag, for the rename bypass and the incomplete build set. Keeping the sticky state in the ledger file carries the same trust as the round cap, which is acceptable.
- **ARCH-ORDER:** pass. Stickiness is explicit per-round state, tested in both orders, and I confirmed both tests fail with the fix removed.
- **ARCH-FUNERAL:** pass. `Recipe` adds one field per round, bounded by the round cap and archived with the issue.

### 7. Plan revision recommendations
- **Spec §1:** it still says code files exclude `*.md`, but markdown under `cmd/` counts as code. Add a Revisions entry.
- **Plan Decisions:** record that ariadne declares `cmd/sdlc/` (widened to the full build set once the second finding is fixed), and that sdlc work in ariadne therefore always takes the full flow.
- **Plan Decisions, ARCH-SECURE bullet:** it says "a branch cannot loosen its own shell". Once the rename fix lands, add a clause saying renames are matched on both paths.

```findings
dispose:
  - id: BR-18
    disposition: addressed
    note: |
      Recipe stamp (boundaryledger.go:182) + EarlierFullReview crossing (closeflow.go:80); dropping the crossing reddens TestCloseQuickReworkShrinkThenReclose, dropping the stamp reddens TestCloseQuickSmallDiffReworkStaysQuick; Spec/plan/close.md corrected.
  - id: BR-19
    disposition: addressed
    note: |
      cmd/sdlc/ declared (covers flow, churn, closeflow, judge prompts, boundary-tail.md); narrowing it reddens TestRepoDeclarationCoversTheShell. The build closure outside cmd/sdlc is raised as a new finding.
  - id: BR-20
    disposition: addressed
    note: |
      churn.CodeFileRule beside IsCodeFile, TestCodeFileRule checks exemplars incl. cmd/**/*.md; ShellSummary renders it; close.md no longer restates it.
  - id: BR-21
    disposition: addressed
    note: |
      surfaces.go refuses a glob in a directory entry; removing the check reddens TestParseSurfacesRejectsGlobInDirectory.
  - id: BR-22
    disposition: not-addressed
    note: |
      entryOf and declarationAt now share EntryAt, but TrunkFile.readFrom (trunkfile.go:283) still issues its own cat-file blob, identical to the new BlobAt; route it through BlobAt.
  - id: BR-23
    disposition: addressed
    note: |
      closeflow.go:92 checks f.Force explicitly; behaviour unchanged and still pinned by TestCloseQuickEmptyDoneWhenRefuses.
findings:
  - id: new
    severity: Important
    family: diff-paths-drop-rename-source
    title: |
      The shell's shared-surface check sees only rename destinations, so moving a file out of a surface, or renaming .sdlc/shared-surfaces itself, stays quick
    detail: |
      Measure matches Surfaces against gitx.DiffNames (git diff --name-only), which lists a rename's destination only. Reproduced at sdlc close: git mv pkg/vocab/x.go other/x.go under a pkg/vocab/ declaration, and git mv of the declaration to .sdlc/shared-surfaces.old, both close {kind: quick} with the small-diff recipe. The second defeats the declaration self-match that Surfaces.Match, the file header and close.md promise, and disables the shell for later branches once merged. Enumeration in this window: Measure is the only DiffNames consumer the diff adds; code-file counting by destination is right, surface matching needs both sides (the atlas gate's atlas/ split has the same class, pre-existing). Fix: match surfaces over both sides of renames/copies (--no-renames name list, or name-status with old and new paths), keep destination-only counting, add rename-out and rename-declaration fixtures.
  - id: new
    severity: Important
    family: hand-listed-guard-incomplete
    title: |
      The gate-machinery declaration stops at cmd/sdlc/, but sdlc's build closure also includes pkg/frontmatter and go.mod/go.sum, and the covering test only walks cmd/sdlc
    detail: |
      This is the 3rd finding in family hand-listed-guard-incomplete. Rule: the set a guard protects must be derived from the protected artifact's real dependency graph, never from a directory the author picked. For the gate binary that is go list -deps ./cmd/sdlc restricted to this module, plus go.mod and go.sum. pkg/frontmatter (Split, used by internal/issue/frontmatter.go to read the frontmatter holding flow:) is compiled into sdlc and undeclared, so a quick branch editing it closes under the binary it changed, which contradicts the declaration header and close.md. Fix: declare pkg/frontmatter/ (or pkg/), go.mod, go.sum, and make TestRepoDeclarationCoversTheShell derive its set from go list -deps instead of WalkDir over cmd/sdlc.
```

---

## Re-review — 2026-09-17T22:29:42-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 231 — A quick path for small diffs: scale the gate set and the review recipe to size |
| repo | ariadne |
| issue file | workshop/issues/000231-a-quick-path-for-small-diffs-scale-the-gate-set-and-the-review-recipe-to-size.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 88b34b640cbb5832edd0030655108cad0cc6f1a8..a5f4c65c32ff117aea4f032150d7ce1d81ff9a92 |
| command | sdlc milestone-close --issue 231 --milestone M2 |
| reviewer | claude |
| timestamp | 2026-09-17T22:29:42-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

All three open findings from earlier rounds are fixed, and I checked each one against the code rather than the commit messages. For BR-24 and BR-25, I reverted each fix in a scratch copy of the repo, and its regression test failed. M2 delivers what the plan asks for, and the Done-when items for this milestone all have tests. The suite passes except for one test that was already failing on main (#210, noted in the Log). Two cheap Important findings remain. The first is that `.sdlc/shared-surfaces` accepts three kinds of pattern that `Match` can never use, so a downstream declaration can silently protect nothing — the same class as BR-21. The second is that three prose passages list things the code owns, and this window changed those things without updating the prose. That is the fifth finding in the `doc-claim-contradicts-code` family. Neither one blocks.

**Architecture principles, one by one:**
- **ARCH-DRY:** pass, with nits (see notes).
- **ARCH-PURE:** pass. `decideCloseFlow`, `fullRoundIn`, `Measure` and `Crossings` are pure. The IO is limited to `measureCloseWindow`, `committedSurfaces` and `earlierFullReview`.
- **ARCH-PURPOSE:** flagged. The sweep over the declaration's syntax and the sweep over prose restatements are incomplete (findings 1 and 2).
- **ARCH-MOCK:** pass. git runs for real in temp repos, and the judge is replaced at the `judge.Run` seam.
- **ARCH-CONSTRAINTS:** pass. Close gains one more `git diff --name-only`.
- **ARCH-SECURE:** pass. The declaration is read only as committed, at base and head. `EntryAt` uses `--end-of-options`.
- **ARCH-ORDER:** pass. The `Recipe` stamp makes the upgrade stick across a REWORK, and the ordering is tested in both directions.
- **ARCH-FUNERAL:** pass. Each round gains one `recipe:` line and is archived with the issue. The declaration is a single file.

## 1. Strengths
- **Close measures what a window changed separately from what it touched.** Code files are counted by rename destination; surfaces are matched on both sides of a rename (`closeflow.go:119-132`, `shell.go:24-45`). Both rename fixtures fail when `touched` is swapped back to `diffFiles`.
- **`TestRepoDeclarationCoversTheShell` now works out the protected set from `go list -deps`.** It includes embedded files, go.mod and go.sum. Dropping `pkg/` or `go.mod` from the declaration makes it fail, listing each uncovered file.
- **REWORK is handled by stamping each ledger round with the recipe it ran** (`boundaryledger.go:180-183`, `fullRoundIn`). It is tested both ways: a full-review REWORK followed by a fix that shrinks the diff stays full, and a small-diff REWORK stays quick.
- **The registry decides which principles the quick flow checks.** Each entry's `quick-flow:` field is enforced by a parser that fails the build on a missing or bad value (`architecture.go:108-145`). The render is lossless, so the full-flow goldens change only by the new lines.
- **The end-to-end test drives the real `claim → change-code → close` sequence without any flow flag,** and checks both the recorded flow and the recipe that was dispatched.

## 2. Critical findings
None.

## 3. Important findings
- **`flow/surfaces.go:23-64`: the parser accepts patterns that never take effect (family `accepted-input-never-effective`, 2nd finding).** I ran each case through a probe test at head:
  - `pkg/vocab` (a directory without the trailing slash) parses, but `Match("pkg/vocab/x.go")` is false.
  - `lua/parley/**` parses but matches only one level deep; `lua/parley/a/b.lua` is false.
  - `!pkg/x.go` parses and never matches.

  All three fail open: a quick issue touching the surface keeps the small-diff review. parley's declaration is the next one to be written, and #263 is the case this file exists to catch. `FuzzParseSurfaces` even uses `a/**/b` as a seed, but its only check is "doesn't panic".
  - **Rule:** every pattern form the parser accepts must take the effect its reader expects, or be refused when parsed.
  - **Fix:** refuse `**` and a leading `!`. Let a pattern with no glob characters match the exact path or anything under it, as gitignore does. Optionally, treat a pattern that matches no tracked path at base or head as a crossing. Add one table test covering all of these forms.

- **Prose still lists facts this window changed in code (family `doc-claim-contradicts-code`, 5th finding).** Earlier rounds fixed individual instances, so this should be fixed as a rule. Three instances are in this window:
  - `cmd/sdlc/helptext/close.md:210-213` still says "the shell's own code must be declared too; ariadne declares `cmd/sdlc/`". That is the scope BR-25 replaced; ariadne now declares the whole build closure.
  - `AGENTS.base.md:52` lists close's guards and their `--no-*` flags without `--no-done-when-fresh`, which this diff added. It also lacks the older `--no-ledger`. This file is the base layer, so the gap spreads to other repos.
  - `atlas/workflow/sdlc-binary.md:891-895` still says "`close` has 8 gates" and gives a flag list that is now incomplete.

  **Rule:** prose must not restate a set, count or scope that code owns. It should either render from the owner (as `{{QUICK_SHELL}}` does, or a token built from `processmanual.GateCatalog` for the flags) or point to it (`sdlc close --help`, `.sdlc/shared-surfaces`) without repeating it. Any fix that changes such a fact should grep for the old fact in the same round.

## 4. Minor findings
- **`surfaces_test.go`: `TestRepoDeclarationCoversTheShell` calls `t.Skipf` when `go list` fails,** so the BR-25 guard can turn itself off without anyone noticing. This is the 3rd finding in `test-oracle-weaker-than-plan`. **Rule:** a guard test must fail when the thing it depends on fails; use `t.Fatalf`.
- **Two files that `go build` reads are not covered: `go.work`/`go.work.sum` and `vendor/`.** If a branch adds either, the binary changes without touching any declared path. This is the 4th finding in `hand-listed-guard-incomplete`. **Rule:** the protected set is everything the go command reads in module mode, including files that don't exist today. Add `go.work*` and `vendor/`.
- **`Measure` gets paths in two different formats.** `files` comes from `DiffNames`, which quotes non-ASCII names (`"docs/caf\303\251.md"`); `touched` uses `-z`, which does not. I confirmed the quoting in a scratch repo. A doc or test with a non-ASCII name is therefore counted as a code file, which can upgrade a quick issue for no reason. The error is on the safe side (toward full), but the `-z` lesson was applied to one listing and not the other.

## 5. Test coverage notes
- The edge cases around each limit (`TestCrossings`), the REWORK sequences in both directions, both rename cases, uncommitted declaration edits, the operator pin, the Done-when refusals and the gate-flag catalogue are all covered with real git and a stubbed judge.
- **Mutation checks I ran on a scratch copy:**
  - With `touched` replaced by `diffFiles`, both rename tests fail.
  - With `pkg/` removed from the declaration, 12 files are reported uncovered.
  - With `go.mod` removed, `go.mod` is reported uncovered.
- **Suite status:** at head, the suite passes apart from `TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory`, which already fails on main (#210). `TestArchitecture_NarrativeRoutesToArchPrinciples` fails only in an exported copy, because `AGENTS.md` is gitignored; it passes in the real checkout.
- **Missing:** a test over the declaration's pattern forms that checks `Match` results, not just that parsing doesn't panic (finding 1).

## 6. Architectural notes for upcoming work
- **BR-22 leftover:** `declarationAt` and `TrunkFile.readFrom` both call `EntryAt` then `BlobAt`. A `gitx.FileAt(dir, ref, path) ([]byte, bool, error)` helper would make this one call.
- **`DiffPathsBothSides` is in `entry.go`,** while `DiffNames` and `DiffNameStatus` are in `window.go` and run git differently. Keep the diff listers together.
- **`!IsDoc(p) || IsEmbedded(p)` appears in both `IsCodeFile` and `publishGateHasCodeSurface`.** It could be one predicate, `churn.IsCodeSurface`.
- **M3:** the ledger's `Recipe` values and flow columns should use the same `judge.Category` strings.

## 7. Plan revision recommendations
- Record that `Measure`'s signature is now `Measure(files, touched, stats, s, surfacesErr, milestones)`. The Core-concepts bullet still shows the four-argument form, and the BR-24 revision names `DiffPathsBothSides` but not the new `touched` parameter.
- Once finding 1 is fixed, record the declaration grammar in the plan's Surfaces bullet: which forms are accepted, what each matches, and which are refused.

```findings
dispose:
  - id: BR-22
    disposition: addressed
    note: |
      gitx.EntryAt/BlobAt (with --end-of-options) now serve TrunkFile.entryOf, TrunkFile.readFrom and close's declarationAt; pathTrackedAtHEAD keeps its injected gitRunner seam by design (plan Revisions). Residual EntryAt-then-BlobAt composition is a nit.
  - id: BR-24
    disposition: addressed
    note: |
      Measure matches surfaces over gitx.DiffPathsBothSides (--no-renames -z) and counts code over destinations; reverting to diffFiles reddens TestCloseRenameOutOfASurfaceUpgrades and TestCloseRenamingTheDeclarationUpgrades.
  - id: BR-25
    disposition: addressed
    note: |
      Declaration adds pkg/, go.mod, go.sum; TestRepoDeclarationCoversTheShell derives from go list -deps (Go + embed files) plus go.mod/go.sum; dropping pkg/ or go.mod reddens it. Residuals (Skip on go-list failure, go.work/vendor) raised separately as Minor.
findings:
  - id: new
    severity: Important
    family: accepted-input-never-effective
    title: |
      ParseSurfaces accepts a bare directory, a double-star glob and a leading-bang pattern, none of which Match ever honours, so a declaration can guard nothing
    detail: |
      Probed at head: "pkg/vocab" parses but Match("pkg/vocab/x.go") is false; "lua/parley/**" matches one level only (lua/parley/a/b.lua false); "!pkg/x.go" parses and never matches. Fail-open on exactly the #263 case, for the next declaration (parley). 2nd in family after BR-21. Rule: every form the parser accepts must take the effect its reader expects, or be refused at parse. Fix: refuse double-star and a leading bang, let a glob-free literal match itself or anything under it, optionally treat a pattern matching no tracked path at base or head as a crossing, and pin the forms with a Match table test (the fuzz oracle only checks no-panic).
  - id: new
    severity: Important
    family: doc-claim-contradicts-code
    title: |
      Prose hand-restates the declared build-closure scope and the close gate-flag set, and this window changed both without sweeping them
    detail: |
      5th in family. Instances in this window: helptext/close.md:210-213 says ariadne declares cmd/sdlc/ as the shell's own code (BR-25 made it the whole build closure); AGENTS.base.md:52 lists close guards and flags without --no-done-when-fresh (added here) or --no-ledger; atlas/workflow/sdlc-binary.md:891-895 says close has 8 gates, with a stale flag list. Rule: prose never restates a set, count or scope that code owns. It renders from the owner (a help token, e.g. from processmanual.GateCatalog) or points at it without restating, and any fix that changes such a fact sweeps restatements keyed on the old fact in the same round.
  - id: new
    severity: Minor
    family: test-oracle-weaker-than-plan
    title: |
      TestRepoDeclarationCoversTheShell skips when go list fails, so the BR-25 guard can silently turn off
    detail: |
      3rd in family. Rule: a guard test fails when the derivation it rests on fails; replace t.Skipf with t.Fatalf. The plan says the closure is derived and pinned by this test, and a skip is neither.
  - id: new
    severity: Minor
    family: hand-listed-guard-incomplete
    title: |
      go.work, go.work.sum and vendor/ change what go build compiles but are not declared shared surfaces
    detail: |
      4th in family. go list -deps can only see inputs that exist today; a branch that adds go.work with a replace directive, or a vendor/ tree, changes sdlc's binary without touching go.mod, go.sum or a package directory. Rule: the protected set is every input the go command reads in module mode, including files that do not exist yet. Declare go.work* and vendor/.
  - id: new
    severity: Minor
    family: porcelain-output-parsed-as-data
    title: |
      Measure gets code-file paths from quoted DiffNames output but surface paths from -z output, so non-ASCII doc and test paths count as code
    detail: |
      git diff --name-only (and --numstat) quote non-ASCII paths ("docs/caf\303\251.md"), which IsDoc and isTestPath then fail to recognise, so they count as code and can upgrade a quick issue for nothing. The error is on the safe side (toward full), but the -z lesson cited on DiffPathsBothSides was applied to only one of the two listings Measure pairs. Use -z (or core.quotePath=false) for the code-file list too.
```

---

## Re-review — 2026-09-17T22:50:41-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 231 — A quick path for small diffs: scale the gate set and the review recipe to size |
| repo | ariadne |
| issue file | workshop/issues/000231-a-quick-path-for-small-diffs-scale-the-gate-set-and-the-review-recipe-to-size.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 88b34b640cbb5832edd0030655108cad0cc6f1a8..7baba3d43654cbf3cf1de27ea075174843e59405 |
| command | sdlc milestone-close --issue 231 --milestone M2 |
| reviewer | claude |
| timestamp | 2026-09-17T22:50:41-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

Round 3 fixed the code sites BR-26, BR-28, BR-29 and BR-30 named, and I confirmed each by reverting it in a scratch export of HEAD. Reverting the new literal-match branch fails `TestSurfaceForms`. Removing the `**` or `!` refusals fails the rejection table. Removing `go.work` or `vendor/` from the declaration fails `TestRepoDeclarationCoversTheShell`. Reverting `DiffNames -z` fails `TestCloseNonASCIIPathsClassifyAsThemselves`. What remains is the doc/test class, not code correctness:
- **BR-27 is only partly fixed.** The new `TestGateFlagListsInHelpAreComplete` checks a whole help page, not each list. Two partial gate-flag lists pass it, and one of them was written in this round.
- **The surface-pattern description is stale.** BR-26's fix changed how a pattern is read, but `.sdlc/shared-surfaces` and the atlas still describe the old rules.
- **The numstat change has no test.** Reverting it leaves the whole suite green. `core.quotePath=false` also still quotes some filenames: a probe shows a 150-line code file named `cmd/a"b.go` closes on the quick flow at HEAD.

None of this makes the gates wrong on realistic diffs, so it is FIX-THEN-SHIP, not REWORK.

**1. Strengths**
- **The protected set is derived, not listed, and fails closed.** `TestRepoDeclarationCoversTheShell` builds it from `go list -deps`, including embedded files and module inputs that don't exist yet. It now fails instead of skipping when `go list` fails (`cmd/sdlc/internal/flow/surfaces_test.go:118-153`). Removing a declared entry turns it red with the uncovered file named.
- **A declaration line that can't be parsed is refused, and the refusal counts as a crossing.** A broken `.sdlc/shared-surfaces` sends the close to the full flow (`surfaces.go:40-53`, `closeflow.go:committedSurfaces`).
- **`DiffNames -z` fixes the quoting in the one helper.** All three callers (close's atlas gate, the publish gate, `reviewanchor`) are corrected at once, rather than one caller at a time (`gitx/window.go:451-464`).
- **The recipe stamp is written in one place for both persist branches.** `neverRan` is computed before the branch (`boundaryledger.go:160-183`). `fullRoundIn` treats a round with no recipe stamp as a full review, which is the conservative reading.

**2. Critical findings:** none.

**3. Important findings**
- **BR-27, not addressed** (details in the findings block).
  - `cmd/sdlc/helptext/close.md:262-276` FLAGS omits `--no-ledger`.
  - `cmd/sdlc/helptext/milestone-close.md:78-81`, a hand-written list added this round, omits `--no-ledger`.
  - Both pass the test because `--no-ledger` appears elsewhere on each page (`close.md:65,177`, `milestone-close.md:125`).
  - `atlas/workflow/sdlc-binary.md:893-894` says every help list is pinned complete, which the test does not do.
- **New, 6th in `doc-claim-contradicts-code`:** the description of the surface-pattern rules is out of date in two places.
  - `.sdlc/shared-surfaces:5-6` and `atlas/workflow/sdlc-binary.md:299` still describe two forms.
  - The code now has three forms plus refusals (`surfaces.go:18-31,68-84`).

**4. Minor findings**
- **The numstat half of BR-30 has no failing-without-it test, and some paths are still quoted.** Probes:
  - `cmd/über.go` with 150 lines closes quick once the `quotePath` change is reverted, and the suite stays green.
  - `cmd/a"b.go` with 150 lines closes quick at HEAD, because `quotePath=false` still quotes `"`, `\` and control characters.
- **Some accepted patterns still match nothing.** `/`, `./`, `.`, `pkg/./vocab`, `pkg//vocab/` and `pkg/vocab/.` all parse and match nothing (probe).
- **Duplicated split loop (ARCH-DRY, not in the findings block).** `DiffNames` and `DiffPathsBothSides` now carry the same NUL-split loop and nearly the same git call. One `nameListZ(args…)` helper in gitx would remove the copy.

**5. Test coverage notes**
- I ran `go test ./...` on a scratch export of HEAD. Everything passes except three failures caused by the environment or by older code:
  - `TestProjectCloseRejectsDuplicateLogicalMVPScopeRefs` needs the checkout directory to be named `ariadne`; it passes after renaming.
  - `TestArchitecture_NarrativeRoutesToArchPrinciples` reads the generated `AGENTS.md`, which is gitignored and so absent from the export.
  - `TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory` reads `workshop/plans/000200-…-plan.md`, which is missing at both base and HEAD. It predates this window.
- `FuzzParseSurfaces` still only checks for no panic. The property "an accepted glob-free pattern matches itself" would find the Minor surface-pattern forms above mechanically.

**6. Architectural notes (ARCH-\*)**
- **ARCH-DRY: flag.**
  - The gate-flag lists are restated in several places (BR-27).
  - The surface-pattern rules are restated in three places (new finding).
  - The NUL-split loop is duplicated in gitx.
- **ARCH-PURE: pass.** `flow`, `churn` and `judge` hold the decisions. `closeflow.go` is a thin IO shell.
- **ARCH-PURPOSE: flag.** BR-27 was answered page by page, not list by list.
- **ARCH-MOCK: pass.** git runs for real in temp repos, and the judge is faked at `judge.Run`.
- **ARCH-CONSTRAINTS: pass.** Close adds a few git calls on one small window.
- **ARCH-SECURE: flag (Minor).** Filenames authored on the branch still reach the gate through quoted numstat output. That errs toward the quick flow.
- **ARCH-ORDER: pass.** Recipe stickiness goes through the ledger. A round that never ran is excluded (`NoCap`).
- **ARCH-FUNERAL: pass.** One field per ledger round, archived with the issue.

**7. Plan revision recommendations**
- **Round-3 BR-27 entry:** correct "all three are completed". The milestone-close list omits `--no-ledger`, and the test checks pages, not lists. Record whichever fix is chosen.
- **Round-3 BR-30 entry:** `TestCloseNonASCIIPathsClassifyAsThemselves` covers `DiffNames` only. Name the test that covers the numstat half once it exists.
- **Surfaces core concept:** record the single owner of the surface-pattern rules once it is created.

```findings
dispose:
  - id: BR-26
    disposition: addressed
    note: |
      Literal-or-subtree branch and the refusals of a leading bang, double-star and globbed dir entries are pinned by TestSurfaceForms and the extended rejection table; reverting either reddens its test (verified in a scratch export). Residual degenerate forms raised separately.
  - id: BR-27
    disposition: not-addressed
    note: |
      Named sites fixed, class not: the new test is page-granular. close.md FLAGS (262-276) and milestone-close.md FLAGS (78-81, written this round) both omit --no-ledger and pass because it appears elsewhere on each page; atlas sdlc-binary.md:893-894 claims every list is pinned complete; sdlc-binary.md:754 still restates "12 gates / 16 sigs" against a 20-row GateCatalog. Render per-command gate-flag lists from GateCatalog via a help token, or make the oracle list-granular.
  - id: BR-28
    disposition: addressed
    note: |
      surfaces_test.go:125 now t.Fatalf on a go list failure.
  - id: BR-29
    disposition: addressed
    note: |
      go.work, go.work.sum and vendor/ declared and in the test's shell list; removing go.work and vendor/ from the declaration reddens TestRepoDeclarationCoversTheShell.
  - id: BR-30
    disposition: addressed
    note: |
      DiffNames uses -z; reverting it reddens TestCloseNonASCIIPathsClassifyAsThemselves. The numstat half is untested and raised separately.
findings:
  - id: new
    severity: Important
    family: doc-claim-contradicts-code
    title: |
      BR-26 changed the shared-surface grammar, but the declaration header and the atlas still describe the old two-form grammar
    detail: |
      This is the 6th finding in family doc-claim-contradicts-code. .sdlc/shared-surfaces:5-6 says every non-directory line is a path.Match glob, so pkg/vocab would mean exactly that path, and it omits the refusals; atlas/workflow/sdlc-binary.md:299 says the same. The code (surfaces.go:18-31,68-84) now treats a glob-free literal as path-or-subtree and refuses the double-star and leading-bang forms. The rule BR-27 stated (a fix that changes a fact sweeps restatements keyed on it in the same round) was violated by the same commit. Rule-level fix: give the grammar one owner, e.g. a flow.SurfaceForms string rendered into sdlc close --help via a token and pinned clause-by-clause against TestSurfaceForms (the churn.CodeFileRule / TestCodeFileRule pattern). Point the declaration header and atlas at it, and sweep in the same round with git grep for path.Match, trailing-slash and for-a-subtree phrasing.
  - id: new
    severity: Minor
    family: porcelain-output-parsed-as-data
    title: |
      The window numstat is still parsed as quoted output: its BR-30 change has no red test, and core.quotePath=false still quotes some paths
    detail: |
      This is the 2nd finding in family porcelain-output-parsed-as-data. Reverting the quotePath change in windowFileStats (closeflow.go:142) leaves the whole suite green; a probe with cmd/über.go at 150 lines then closes quick. The plan's "mutation-checked" claim covers only DiffNames. At HEAD, cmd/a"b.go at 150 lines closes quick, because quotePath=false still quotes a double-quote, a backslash and control characters, and the quoted row never joins the -z code-file list. Rule: every git listing a gate reads as data uses -z. Use diff --numstat -z (renames arrive as separate NUL fields) with a -z parse, and add a close test with a git-quoted code-file name over the line limit.
  - id: new
    severity: Minor
    family: accepted-input-never-effective
    title: |
      ParseSurfaces still accepts root and non-canonical patterns (/, ./, ., pkg/./vocab, pkg//vocab/) that match nothing
    detail: |
      This is the 3rd finding in family accepted-input-never-effective. Probed at head: each parses without error and matches no path. The forms are refused by hand-enumeration while the fuzz oracle still checks only no-panic, so the next form slips too. Rule: accept only canonical, non-empty patterns (path.Clean(p) == p and p is not "."). Make the oracle a property instead of a table: FuzzParseSurfaces asserts that an accepted glob-free pattern matches itself, and a directory entry matches pattern + "x". Also decide whether a glob naming a directory (pkg/v*) covers its subtree the way the literal form now does, or document the asymmetry.
```
