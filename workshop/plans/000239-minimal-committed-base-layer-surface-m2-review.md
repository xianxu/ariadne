# Boundary Review — ariadne#239 (milestone M2)

| field | value |
|-------|-------|
| issue | 239 — Minimal committed base-layer surface |
| repo | ariadne |
| issue file | workshop/issues/000239-minimal-committed-base-layer-surface.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 315579a6c859f32d9be3ea5da144200a23ec4767..a7609fc062f2b7569037ecdc04d095f08cf0840c |
| command | sdlc milestone-close --issue 239 --milestone M2 |
| reviewer | claude |
| timestamp | 2026-09-19T22:39:54-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

M2 delivers what it promised: `.gitignore` becomes a delimited weave-owned region that replaces its contents wholesale, still carrying the known-good hardcoded nine entries, with the six planned tests landing and the whole suite green. I verified the mechanism against the real tree rather than the commit message — ariadne's live `.gitignore` is a true fixed point of `mergeManagedBlock` (re-ran the transform against the file via a `go test -overlay` probe: `changed=false`), and I enumerated all 18 sibling repos' `.gitignore` files to confirm the one-time migration inverts no repo-owned pattern anywhere in the fleet. Nothing blocks SHIP. What holds it back from a clean SHIP is a cluster of assertions that cannot fail — a new fail-closed guard with no fixture able to reach it, a test rewritten into a tautology while the property it is named for measurably regressed, and a docstring claiming a safety property the code does not have — plus four prose restatements of the mechanism that now all describe the retired append-only behavior, one of them naming a function deleted in the same commit.

## 1. Strengths

- **The migration was verified, not asserted.** `.gitignore:17-24` — the orphaned `/AGENTS.md` explanatory comment was rewritten into a block-aware one rather than left dangling (Task 2.2 Step 5), and I independently confirmed the file round-trips with `changed=false`. Step 6's no-churn claim holds.
- **`TestManagedBlockReplacesWholesaleSoRetiredEntriesDisappear`** (`cmd/weave/internal/plan/gitignore_test.go:211`) pins the single behaviour this milestone exists for, and asserts both directions — the retired entry gone *and* repo-owned lines on both sides surviving.
- **`TestManagedBlockRefusesMalformedMarkers`** (`gitignore_test.go:285`) asserts the error *names the remedy* (`strings.Contains(err.Error(), "make weave")`), not merely that it errored. That is the right assertion for an operator-facing fail-closed path.
- **Milestone ordering held under inspection.** `main.go:687` still emits `plan.EnsureGitignore{Entries: plan.GeneratedRuntimeGitignoreEntries}` target-independently, so the wholesale-replace + lean-`--target` hazard the plan warns about genuinely does not bite until M3. Verified, not assumed.
- **ARCH-PURE is intact**: `mergeManagedBlock` is string-in/string-out and every one of its tests runs with no fake; the only IO change is confined to `applyEnsureGitignore`.

## 2. Critical findings

None.

## 3. Important findings

**I1 — three assertions in this boundary that nothing can falsify** (`verification-cannot-fail`, 4th in family)

- `cmd/weave/internal/plan/gitignore.go:187` — the new fail-closed read-error guard. Measured: `grep -rn "func (.*) ReadFile(" --include="*_test.go" cmd/weave/` returns **nothing**, so no fixture in the package can enter the branch. `materializationFaultFS` (`apply_test.go:647`) already exists and faults `lstat`/`remove`/`write`/`chmod`; adding `ReadFile` is ~4 lines.
- `gitignore_test.go:109` — `TestEnsureGitignoreTextDedupsRepeatedInputEntry` now passes a **single** entry and asserts `Count(got, "/AGENTS.md") != 1`. Tautological. Meanwhile the property it is named for actually regressed: measured `mergeManagedBlock("", []string{"/A","/A"})` → `"…>>>\n/A\n/A\n# <<<…"`. The retired `ensureGitignoreText` guarded this explicitly (`present[entry] = true // guard against a duplicate entry in the input list`). Either restore the guard or delete the test — a green namesake hides the regression until M3's `IgnoreEntries` dedupe.
- `gitignore.go:73` — the docstring claims exact-whole-line matching means "a .gitignore that merely *mentions* a marker (a comment explaining this mechanism) cannot be mistaken for the region itself". Measured false: a `.gitignore` carrying the open marker as a quoted comment line yields `duplicate weave-generated opening marker (line 3) — a merge conflict? — delete the managed block and re-run make weave`, i.e. `make weave` hard-fails and the remedy points at the wrong thing. Whole-line matching does not protect against a line that *is* the marker — which is precisely the `workshop/lessons.md` lesson the comment cites.

Supporting prevalence from the tracker: the issue's `## Log` still states in the past tense that "Both base-layer tests now register as `scripts/merge-checks.d/50-base-layer-tests.sh` (side-quest)". `ls scripts/merge-checks.d/` shows 30/40/45 + README; `git log --all -- 'scripts/merge-checks.d/50*'` is empty. BR-13's disposition already re-scoped this to M3, but the sentence survived and reads as shipped CI coverage.

**Do not fix these three instances one at a time — the rule is what is missing.** Round 3 already stated half of it ("a guard added in answer to a finding is complete only when a test goes red without it"); the residue this round adds is the other half: *a claim in a comment or a Log line is not verification.* Concretely: extend `materializationFaultFS` to the read seam so fail-closed paths are driven like materialization paths already are; and treat a stated safety property as requiring a test naming it (here, the "marker as content" case), or drop the claim.

**I2 — four prose restatements of the gitignore mechanism, all now describing the retired append-only behaviour** (`hand-maintained-restatement-of-model`, 3rd in family)

- `gitignore.go:21` — "the entry LIST + the pure **ensure-text** transform live here"; that transform is gone.
- `gitignore.go:58` — `EnsureGitignore`'s type doc: "**appending the absent ones** (idempotent: a present entry is never duplicated…)". Wholesale replacement now, and duplicates *are* possible (measured above).
- `gitignore.go:177` — `applyEnsureGitignore`'s own doc: "append the missing entries via the pure **`ensureGitignoreText`**" — names a symbol deleted in the same commit, directly above its replacement, and omits both new failure modes.
- `cmd/weave/internal/plan/apply.go:39` — "append the missing fixed entries (idempotent whole-line append, never duplicating), write back only on change." The plan's Task 2.2 **Files** list names `cmd/weave/internal/plan/apply.go:35-39` as a file to modify; `git diff --name-status` for the window shows apply.go **untouched**. A Plan item this boundary claims, not delivered.

The rule is already enforced for the *verb set* by `scripts/merge-checks.d/45-verb-enumeration.sh`. The residual class it cannot see is **a comment naming a top-level identifier the package no longer defines**. Measured prevalence at HEAD: two live instances — `gitignore.go:177` (`ensureGitignoreText`, this milestone) and `golden/golden.go:222` (`plan.SeedOnceSlotIsRepoOwned`, BR-12, still open). That is greppable and enforceable the same way: for each top-level `func`/`var`/`const`/`type` removed by the range, fail if the name still appears in a non-test, non-`workshop/` comment. Build that, then sweep the four sites above in one pass rather than editing the one the next review happens to name.

**I3 — README gate: new adopter-facing surface, no README change in range** (`adopter-facing-surface-undocumented`)

The diff makes weave a co-owner of every repo's `.gitignore`: a maintainer must now know the markers exist, that entries go outside them, that content inside is destroyed on each compile, and that `make weave` — and therefore `make bootstrap` — now **hard-fails** on a `.gitignore` whose markers it cannot parse. `README.md:15-20` already carries the exact sibling fact from M1 ("A consumer's root `Makefile` is **the repo's own**… Edit it freely"), so the shape and the location are settled; this is a 3–4 line addition under the same heading. `atlas/workflow/weave.md` was updated and is good, but the atlas is the codebase map, not the adopter's front door.

## 4. Minor findings

- `gitignore.go:119` — repo lines positioned *after* the managed block are moved *above* it, so git's last-match-wins silently flips in weave's favour. Measured: `mergeManagedBlock(block + "!/CLAUDE.md\n", …)` returns `!/CLAUDE.md` above the block. I enumerated all 18 sibling repos: none currently has a pattern after the weave entries that the nine entries touch, so nothing breaks today — but `atlas/workflow/weave.md:111` says "Outside is the repo's own, **preserved verbatim**", which is true of content and not of position. Worth one clause in the atlas and a line in `## Log` so M3/M4 don't re-derive it.
- `gitignore.go:122` — a CRLF `.gitignore` defeats the marker match (`\r` survives `strings.Split(current, "\n")`). Measured: weave appends a **second**, LF block and can never recognise or retire the original — a permanent orphan whose line numbers would also mislead M4's `commitConsumption` provenance filter, which keys on the block's line range. No fleet repo uses CRLF; one `strings.TrimRight(line, "\r")` at the compare closes it.
- **(`edit-splice-leaves-orphan-clause`, 2nd in family)** The absorb orphans each derivative's own explanatory comment. ariadne's was hand-removed; `parley.nvim/.gitignore` carries a 4-line "# weave-generated runtime artifacts (lowered by `make weave`, like AGENTS.md above)" comment whose eight entries will all migrate into the block at that repo's next weave, stranding the comment above `.mdbg*` with a now-false "above" back-reference. Do not patch parley.nvim. The rule from round 3 generalises with the actor swapped: *after a splice, read the RENDERED result, not the hunk* — here the splicer is code, so the migration must be rehearsed against each repo's real `.gitignore`, not only ariadne's. M4's per-repo pilot is the natural place to make that a step.
- `gitignore.go:187` uses `os.IsNotExist` rather than `errors.Is(err, fs.ErrNotExist)`. Correct for `OSFS` (returns a bare `*PathError`) and consistent with `apply.go:312`/`:412`, so no change needed — noting only that a future `weavefs.FS` wrapping its errors with `%w` would turn "no `.gitignore`" into a hard weave failure.

## 5. Test coverage notes

Six planned block tests landed and all six pin real behaviour. Gaps: (a) no seam-level test for the read-error branch (I1); (b) no seam-level test that a malformed block propagates out of `Apply` — only the pure function is covered, and the wrapper at `gitignore.go:196` is untested; (c) the duplicate-entry contract has no test that can fail (I1). `TestCompileEnsuresGitignore` (`main_test.go:198`) survives translation intact and still proves the end-to-end create + idempotence. Real-`git` conformance for the block lands with M3's `gitignore-surface.test.sh`; until then nothing checks that the block's patterns behave as git patterns, which is acceptable at M2 because the nine entries are unchanged.

Environment caveat: `scripts/merge-checks.d/30-weave-drift.sh` could not run in my sandbox (`mktemp: Operation not permitted`), and it reports that failure as `dynamic-skill render is NON-DETERMINISTIC` — a misleading diagnosis, but pre-existing and outside this window. `45-verb-enumeration.sh` passes on the range; `go build ./...` and `go test ./cmd/weave/...` are green.

## 6. Architecture (all eight, at-review)

- **ARCH-DRY — flag.** `mergeManagedBlock` is correctly the one owner of "which lines are weave's"; flagged only for the four prose copies of its contract (I2).
- **ARCH-PURE — pass.** Pure transform, thin seam, no mock needed to test the core.
- **ARCH-PURPOSE — flag.** M2's scope (keep the hardcoded nine) is the plan's deliberate design, so no under-delivery there. But the plan named two stale-doc sites and the diff fixed one (`gitignore.go:26-31`) while leaving its enumerable siblings — the instance, not the class (I2).
- **ARCH-MOCK — flag.** The fs seam has a stateful fake, and the new fail-closed path is outside what it can drive (I1). `git`'s own behaviour against the block has no conformance check until M3.
- **ARCH-CONSTRAINTS — pass.** O(lines) over a <10 KB file, once per compile; comfortably inside the plan's <5 ms budget; no concurrency, fan-out, or network.
- **ARCH-SECURE — flag.** The instinct is right — untrusted hand-edited input, fail closed on both unparseable markers and read errors, remedy carried in the message. Undercut by one claimed property that does not hold (I1c) and one guard no test reaches (I1a).
- **ARCH-ORDER — pass.** `.gitignore` is durable state across weave runs and the transition set is explicit: (no block | valid pair | malformed) × entries → (block | error), enumerated in a single switch with the invalid combinations rejected before the splice. Idempotence verified on the live file and in `TestManagedBlockRoundTripsRepoOwnedNegations`. The CRLF variant is an unmodelled fifth state (Minor).
- **ARCH-FUNERAL — pass.** The block *is* the removal path this milestone exists to add, bounded by the manifest; `legacyBlanketEntries` (`gitignore.go:92`) names its own end. Note the deletion trigger currently lives only in a code comment — the plan asks for it in the issue `## Log` at M4 close, which is still outstanding.

## 7. Plan revision recommendations

1. **`## Revisions` — M2 boundary review.** Core concepts gains `managedBlockOpen`/`managedBlockClose` and `legacyBlanketEntries` as new package-level entities (the same omission BR-12 flagged for `plan.SeedOnceSlotIsRepoOwned`); and Task 2.2's **Files** list names `cmd/weave/internal/plan/apply.go:35-39` as modified when apply.go is untouched in the range — either do the edit or drop the claim.
2. **Task 3.1** should record that `mergeManagedBlock` emits `entries` verbatim (measured), so `IgnoreEntries`' dedupe is load-bearing rather than tidy, and the M3 replacement for `TestEnsureGitignoreTextDedupsRepeatedInputEntry` must pass an actually-repeated entry.
3. **Task 4.0a** should name the CRLF/duplicate-block state as a precondition: the provenance filter keys on the managed block's line range, and an unrecognised stale block would give it a second range to key against. Close it in M3 or state the assumption.

```findings
findings:
  - id: new
    severity: Important
    family: verification-cannot-fail
    title: |
      Three assertions added this boundary cannot fail — the fail-closed read guard, the rewritten dedup test, and the docstring's marker-as-content claim
    detail: |
      This is the 4th finding in family verification-cannot-fail, so the deliverable is the
      RULE, not these three sites. Round 3 stated half of it; the residue is the other half —
      a claim in a comment or a Log line is not verification. Measured instances.
      (1) gitignore.go:187's read-error guard: grep for a test FS overriding ReadFile across
      cmd/weave returns nothing, so no fixture can enter the branch, while materializationFaultFS
      (apply_test.go:647) already faults four other ops and would take ReadFile in four lines.
      (2) gitignore_test.go:109 was rewritten to pass ONE entry and assert count==1 — tautological —
      while the property it is named for regressed - measured mergeManagedBlock("", 2x "/A")
      emits /A twice, where the retired ensureGitignoreText guarded it explicitly.
      (3) gitignore.go:73 claims a .gitignore that merely mentions a marker "cannot be mistaken
      for the region itself"; measured false — a quoted open-marker line returns "duplicate
      weave-generated opening marker (line 3)" and hard-fails make weave, with a remedy pointing
      at the wrong thing. Whole-line matching does not protect a line that IS the marker, which
      is exactly the lessons.md lesson the comment cites. Supporting prevalence in the tracker -
      the issue Log still states in the past tense that both base-layer tests register as
      scripts/merge-checks.d/50-base-layer-tests.sh; the file does not exist and git log --all
      for it is empty. Enforcement: extend materializationFaultFS to the read seam so fail-closed
      paths are driven like materialization paths already are, and require a named test for any
      stated safety property (ARCH-SECURE, ARCH-MOCK).
  - id: new
    severity: Important
    family: hand-maintained-restatement-of-model
    title: |
      Four prose restatements of the gitignore mechanism all still describe the retired append-only behaviour, one naming a function deleted in the same commit
    detail: |
      This is the 3rd finding in family hand-maintained-restatement-of-model, so the deliverable
      is the RULE. Sites - gitignore.go:21 ("the entry LIST + the pure ensure-text transform"),
      gitignore.go:58 (EnsureGitignore's type doc, "appending the absent ones (idempotent - a
      present entry is never duplicated)" — both halves now false), gitignore.go:177
      (applyEnsureGitignore's own doc, "append the missing entries via the pure ensureGitignoreText",
      naming a symbol this commit deleted and omitting both new failure modes), and apply.go:39,
      which the plan's Task 2.2 Files list explicitly names as a file to modify while
      git diff --name-status shows apply.go untouched in the window. 45-verb-enumeration.sh
      already enforces this rule for the VERB set; the residual class it cannot see is a comment
      naming a top-level identifier the package no longer defines. Measured prevalence at HEAD -
      two live, gitignore.go:177 (ensureGitignoreText, this milestone) and golden/golden.go:222
      (plan.SeedOnceSlotIsRepoOwned, BR-12, still open). Enforceable the same greppable way -
      for each top-level func/var/const/type removed by the range, fail if the name still appears
      in a non-test, non-workshop comment. Build that, then sweep all four sites in one pass
      (ARCH-DRY, ARCH-PURPOSE).
  - id: new
    severity: Important
    family: adopter-facing-surface-undocumented
    title: |
      README not updated for the managed block — weave is now a co-owner of every repo's .gitignore and make weave can hard-fail on it
    detail: |
      The diff introduces an adopter-facing convention - markers a maintainer must not edit,
      entries that must go outside them, content inside destroyed each compile, and a NEW way
      for make weave (hence make bootstrap) to hard-fail. README.md:15-20 already carries the
      exact sibling fact from M1 ("A consumer's root Makefile is the repo's own ... Edit it
      freely"), so the heading and shape are settled; this is a 3-4 line addition there.
      atlas/workflow/weave.md was updated and is good, but the atlas is the codebase map, not
      the adopter's front door.
  - id: new
    severity: Minor
    family: block-position-changes-pattern-precedence
    title: |
      Repo lines positioned after the block move above it, silently flipping git's last-match-wins in weave's favour
    detail: |
      Measured - mergeManagedBlock(block + "!/CLAUDE.md\n", ...) returns !/CLAUDE.md above the
      block. I enumerated all 18 sibling repos' .gitignore files - none currently has a pattern
      after the weave entries that the nine entries touch, so nothing breaks today.
      atlas/workflow/weave.md:111 says outside entries are "preserved verbatim", true of content
      but not of position. One clause in the atlas plus a Log line so M3/M4 do not re-derive it.
  - id: new
    severity: Minor
    family: block-position-changes-pattern-precedence
    title: |
      A CRLF .gitignore defeats marker matching — weave appends a second block and can never retire the first
    detail: |
      \r survives strings.Split(current, "\n"), so no line equals managedBlockOpen. Measured - a
      CRLF file gains a second LF block while the original becomes a permanent orphan the
      wholesale-replace path can never reach, and its line numbers would also mislead M4's
      commitConsumption provenance filter, which keys on the block's line range. No fleet repo
      uses CRLF today; one strings.TrimRight(line, "\r") at the compare closes it.
  - id: new
    severity: Minor
    family: edit-splice-leaves-orphan-clause
    title: |
      The one-time absorb orphans each derivative's own explanatory comment, verified in parley.nvim
    detail: |
      This is the 2nd finding in family edit-splice-leaves-orphan-clause, so the deliverable is
      the RULE, not parley.nvim. ariadne's orphan was hand-removed at .gitignore:17-24;
      parley.nvim/.gitignore carries a four-line "# weave-generated runtime artifacts (lowered
      by make weave, like AGENTS.md above)" comment whose eight entries all migrate into the
      block at that repo's next weave, stranding it above .mdbg* with a false "above"
      back-reference. The round-3 rule generalises with the actor swapped - after a splice, read
      the RENDERED result rather than the hunk; here the splicer is code, so the migration must
      be rehearsed against each repo's real .gitignore, not only ariadne's. M4's per-repo pilot
      is the natural place to make that a step.
```

---

## Re-review — 2026-09-19T22:57:28-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 239 — Minimal committed base-layer surface |
| repo | ariadne |
| issue file | workshop/issues/000239-minimal-committed-base-layer-surface.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 315579a6c859f32d9be3ea5da144200a23ec4767..382c4b9702c2a8b8b84df5d2cdce389bd2e41c9f |
| command | sdlc milestone-close --issue 239 --milestone M2 |
| reviewer | claude |
| timestamp | 2026-09-19T22:57:28-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

M2's mechanism is sound and I verified it against the tree rather than the commit message: `mergeManagedBlock` is a true fixed point of ariadne's real committed `.gitignore` (`changed=false`, measured), and both behavior-changing BR-19 fixes are genuinely falsifiable — I reverted each in a scratch export of HEAD and watched the named test go red. BR-20's rule artifact (`46-removed-symbol-references.sh`) also falsifies on its primary case and runs in 0.14s over the whole branch. What holds it back from SHIP is that two of the three disposals are class-incomplete in a measurable way: the tautological test BR-19 named by file:line (`gitignore_test.go:109`) is still in the tree unchanged, still passes with the dedupe guard reverted, and now carries a comment the code contradicts; and the check built to enforce BR-20 derives its removed-symbol set from `git diff BASE HEAD`, so a symbol born and buried inside the range is invisible — which is exactly the shape of BR-20's *second* measured site (`golden.go`/`SeedOnceSlotIsRepoOwned`), and the check reports green over the range CI actually passes it. Nothing Critical; the three carried Minors (BR-22/23/24) are all still live, BR-23 more silently than the finding described.

## 1. Strengths

- **The two behavior-changing fixes carry real regression evidence.** Reverting the dedupe loop (`gitignore.go:176-186`) to `append([]string{managedBlockOpen}, entries...)` turns `TestManagedBlockDedupsRepeatedInputEntry` red; stubbing out the `!os.IsNotExist` arm (`gitignore.go:213-219`) turns `TestApplyEnsureGitignoreFailsClosedOnReadError` red. Both measured in a scratch `git archive` of HEAD, not inferred.
- **`scripts/merge-checks.d/46-removed-symbol-references.sh` falsifies on its motivating case.** Over `315579a..be53a71` it flags `gitignore.go:178 names removed symbol ensureGitignoreText` and exits 1; over HEAD it is green; over the full branch it runs in 0.14s — fast enough not to get routed around, which was the stated bar from 45-.
- **The historical-mention allowance is the right design and is actually exercised.** `gitignore.go:176` ("The retired `ensureGitignoreText` guarded this…") survives the check. A rule that forbade every mention would have deleted the single most useful comment in the file.
- **The live-migration claim holds at HEAD.** Feeding ariadne's real `.gitignore` to `mergeManagedBlock(current, GeneratedRuntimeGitignoreEntries)` returns `changed=false` — the committed block is a fixed point, so the next `make weave` on ariadne is a genuine no-op.
- **README.md:39-55** lands the adopter contract under the heading M1 already established, and names the two consequences an adopter actually needs: content inside the markers is destroyed each compile, and unparseable markers fail `make weave` — hence `make bootstrap` — rather than guessing.

## 2. Critical findings

None.

## 3. Important findings

**(a) `46-…sh` cannot see BR-20's second motivating site — `cmd/weave/internal/golden/golden.go` / `scripts/merge-checks.d/46-removed-symbol-references.sh:37-46`.** `removed` comes from `git diff "$BASE" "$HEAD"`, which collapses a symbol added and deleted inside the range. `SeedOnceSlotIsRepoOwned` was introduced at `fd195a1` and removed at `361e00a`, both on this branch. Measured: over `merge-base(main,HEAD)..382c4b9` — the range `merge-check.yml` passes — the extracted set is exactly `{ensureGitignoreText}` and the check prints green; over `361e00a^..361e00a` it flags `golden.go:222` and exits 1. So in the mode it will actually run, the check enforces one of the two findings it was built for. Fix sketch: union per-commit removals across `git rev-list "$BASE".."$HEAD"`, or make it range-free by asserting every backticked/package-qualified identifier in a Go comment is declared at HEAD.

**(b) BR-19's named site is unchanged — `cmd/weave/internal/plan/gitignore_test.go:109-121`.** See the disposition below; re-raised there by id rather than as a new finding.

## 4. Minor findings

- `scripts/merge-checks.d/46-…sh:71` — the `scripts/merge-checks.d/46-*` arm of the `case "$f"` filter can never fire: `git grep` is restricted to `'*.go' '*.md'`, so a `.sh` path never reaches it. Dead guard; drop it or extend the grep globs.
- `cmd/sdlc` is red on `main` for an unrelated reason: `TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory` reads `workshop/plans/000200-sdlc-fleet-thread-inventory-plan.md`, archived away at `dfeba9c`. Absent at the window base too, so it is not this boundary's — but it will bite at merge if the gate runs `go test ./...`.
- `mergeManagedBlock` builds two sets over `entries` (`absorb` at :151-156, `emitted` at :183) that could be one.

## 5. Test coverage notes

`go test ./cmd/weave/...` is green. The six planned M2 cases plus the three BR-19 additions all pin real behavior, and `TestManagedBlockRefusesMalformedMarkers` asserts the message *names the remedy*, which is the right assertion for an operator-facing fail-closed path. Two gaps the suite cannot currently see, both measured live:

- No case places a repo-owned line **after** the block, so the round-trip test's `strings.HasPrefix(second, repo)` only certifies "preserved verbatim" for the prefix case. `mergeManagedBlock(block + "!/CLAUDE.md\n", …)` returns `!/CLAUDE.md` *above* the block (BR-22).
- No case feeds CRLF. A CRLF `.gitignore` yields two opening markers after one merge, and the **second** pass returns `changed=false, err=nil` — silent and permanent, not fail-closed (BR-23).

## 6. Architectural notes for upcoming work

- **ARCH-DRY** pass — this round's real DRY win is the rule artifact, not the five swept comments. **ARCH-PURE** pass — `mergeManagedBlock` unit-tests with no IO; `materializationFaultFS.ReadFile` extends the existing `weavefs.FS` seam rather than introducing a second one. **ARCH-MOCK** pass — production and test share the same boundary. **ARCH-CONSTRAINTS** pass — one read + conditional write per compile; check measured at 0.14s branch-wide. **ARCH-ORDER** pass for M2 — the transform carries no state between events, and the one cross-version ordering (a derivative jumping pre-M2 → post-M3) is modeled explicitly by `legacyBlanketEntries` and tested. **ARCH-FUNERAL** pass — that list names its end and the trigger is a real plan step (Task 4.3 Step 5).
- **ARCH-SECURE flag** — `.gitignore` is input weave did not produce, and the atlas now claims the transform "**fails closed** on any marker shape it cannot parse" (`atlas/workflow/weave.md:115-117`). Measured false for CRLF: it degrades invisibly instead of visibly. That is the entry's own failure mode ("failure path degrades visibly rather than … substituting a fabricated value downstream code will read as evidence") — and M4's `commitConsumption` provenance filter keys on the block's line range, so the orphaned first block is exactly the evidence it would misread.
- **ARCH-PURPOSE flag** — both open disposals are the instance rather than the class; detail in the findings block.
- For M3: ariadne's own `/.colima/` blanket ignore disappears when per-path derivation lands (the rows self-reference on the self-walk). I checked — `.colima/` holds only six tracked source files and no ignored runtime content, so the drop is safe and is in fact the pair#64 fix the plan predicts. Worth one Log line so M3 does not re-derive it.

## 7. Plan revision recommendations

`workshop/plans/000239-…-plan.md` was not touched in this window and now contradicts the code in one place and under-specifies M4 in another:

- **Task 2.2 Step 3 (plan:1056)** — "the new block emits `entries` verbatim, so a repeated input entry would appear twice. Dedupe inside `IgnoreEntries` … and repoint this test there in M3." BR-19 moved the dedupe into `mergeManagedBlock` (`gitignore.go:176-186`). Add a `## Revisions` entry recording that the block now owns dedupe and saying whether `IgnoreEntries` still dedupes at the source in M3 (belt-and-braces) or defers to the block.
- **Task 3.1's `IgnoreEntries` docstring draft (plan:1234)** — "deduped and sorted" should name which of the two layers is authoritative, so M3 does not re-litigate it.
- **M4 / Task 4.1-4.3** — add BR-24's step explicitly: rehearse `mergeManagedBlock` against **each repo's real `.gitignore`** and read the rendered result, not the hunk. The issue's `## Revisions` records it as "advisory, carried to M4", but the durable plan has no such step.
- Tick Task 2.1 / 2.2's checkboxes at the M2 close (AGENTS.md §8).

```findings
dispose:
  - id: BR-19
    disposition: not-addressed
    note: |
      Two of three sites fully addressed with red-without-fix evidence (dedupe guard and read guard each reverted in a scratch HEAD export; named test goes red). The residue is the site the finding measured by line: gitignore_test.go:109 is unchanged, still passes ONE entry, still PASSES with the dedupe guard reverted, and its comment now asserts "de-duplication is the entry LIST's job — IgnoreEntries dedupes at the source in M3", which gitignore.go:176-186 contradicts. The rule half ("a claim in a comment is not verification") was applied to three instances but the enumeration it implies was never swept, and no durable artifact records it — workshop/lessons.md is untouched in the window and its nearest entry (line 1242, guard-nested assertions) does not cover a namesake test fed input that cannot violate its property. Fix: delete or repoint gitignore_test.go:109-121 and drop the stale comment.
  - id: BR-20
    disposition: addressed
    note: |
      All five sites swept (gitignore.go:21/58/177, apply.go:39 — the plan item that was claimed but undelivered — and golden.go:222), and the rule exists as scripts/merge-checks.d/46-removed-symbol-references.sh, falsified both ways by me: red at 315579a..be53a71 flagging gitignore.go:178, green at HEAD, 0.14s branch-wide. Its range blind spot is raised separately below rather than re-opening this id.
  - id: BR-21
    disposition: addressed
    note: |
      README.md:39-55 documents the managed region under the heading M1 established, with the markers, the wholesale replace, where to put your own entries, and the hard-fail path through make weave into make bootstrap. Inspected against gitignore.go's actual behaviour; the one inaccuracy ("preserved verbatim") is BR-22/BR-24's substance, noted there.
  - id: BR-22
    disposition: not-addressed
    note: |
      Still live and now restated in a second place. Measured at HEAD: mergeManagedBlock(block + "!/CLAUDE.md\n", ["/CLAUDE.md"]) returns !/CLAUDE.md ABOVE the block. atlas/workflow/weave.md:111 still says outside lines are "preserved verbatim", and README.md:49 (added this round) now says the same — so the round's new adopter-facing doc inherits the claim. No issue Log line either. One clause in both docs plus a Log line.
  - id: BR-23
    disposition: not-addressed
    note: |
      Still live and worse than described. Measured: a CRLF .gitignore carrying a block gains a second LF block on the first merge, and the SECOND pass returns changed=false, err=nil — silent and permanent, with no fail-closed path. That contradicts the new atlas claim (weave.md:115) that the transform "fails closed on any marker shape it cannot parse", and the orphaned first block is exactly what M4's line-range-keyed commitConsumption filter would misread. One strings.TrimRight(line, "\r") at the compare, plus a CRLF test case.
  - id: BR-24
    disposition: not-addressed
    note: |
      Recorded in the issue's Revisions as "advisory, carried to M4", but the durable plan is untouched in this window and M4's tasks contain no rehearsal step — a note in the issue is not the plan step the finding asked for. The behaviour is confirmed at HEAD: mergeManagedBlock("# my own vm tree\n/.colima/\nmine/\n", ["/AGENTS.md"]) deletes /.colima/ and strands its comment. Minor, non-blocking; add the per-repo rendered-result rehearsal to the M4 pilot task.
findings:
  - id: new
    severity: Important
    family: verification-cannot-fail
    title: |
      46-removed-symbol-references.sh misses symbols born and buried inside its own range, so it never sees BR-20's second motivating site
    detail: |
      This is the 5th finding in family verification-cannot-fail (BR-19 was the 4th). Do not fix this instance alone — the RULE is: a check is not evidence until it has been run, AT THE RANGE GRANULARITY CI WILL USE, against EVERY site that motivated it, and observed to go red on each; falsifying one site in a scratch repo is a sample of size one. The enumeration that implies is greppable and cheap: for each finding a check claims to enforce, list the finding's measured sites and re-run the check over merge-base..head; any site it passes is an unenforced site. Sweep that enumeration this round, for 45- as well as 46-. Measured prevalence for this instance: 46-…sh:37-46 derives `removed` from `git diff "$BASE" "$HEAD"`, which collapses a symbol added and deleted inside the range. SeedOnceSlotIsRepoOwned was introduced at fd195a1 and removed at 361e00a, both on this branch, so over merge-base(main,HEAD)..382c4b9 — the range merge-check.yml passes — the extracted set is exactly {ensureGitignoreText} and the check prints green; over 361e00a^..361e00a it flags golden.go:222 and exits 1. Fix sketch: union per-commit removals across `git rev-list "$BASE".."$HEAD"`, or make it range-free by asserting every backticked or package-qualified identifier in a Go comment is declared at HEAD.
```

---

## Re-review — 2026-09-19T23:14:57-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 239 — Minimal committed base-layer surface |
| repo | ariadne |
| issue file | workshop/issues/000239-minimal-committed-base-layer-surface.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 315579a6c859f32d9be3ea5da144200a23ec4767..a62d20ad3f9ec88ac5adc5d035852eee0fb346f6 |
| command | sdlc milestone-close --issue 239 --milestone M2 |
| reviewer | claude |
| timestamp | 2026-09-19T23:14:57-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

M2's mechanism is real and I verified it against the tree rather than the commit messages: ariadne's live `.gitignore` is a true fixed point of `mergeManagedBlock` (`changed=false`, measured), and all three BR-19 guards go **red** when reverted in a scratch copy — the fail-closed read guard (`gitignore_test.go:301` fires), the input dedupe (`/AGENTS.md` emitted twice), and the quoted-marker error. BR-25's union fix is likewise measured: reintroducing the born-and-buried `SeedOnceSlotIsRepoOwned` comment at `golden.go:222` now exits 1 over `merge-base(main,HEAD)..HEAD`, the exact range `merge-check.yml` passes. Nothing blocks SHIP. What holds it back is that the *enforcement sweep the last round promised* is itself partly unmeasured: I reverted `gitignore.go:21` to its literal pre-fix text and check 46 printed green, and I found two further shapes it structurally cannot see — so the family's rule ("a check is evidence only once it has gone red on every motivating site") is written down but still executed by hand and by memory. Plus the atlas page that enumerates ariadne's merge checks is now two entries stale, including the check this round added. Three prior Minors (BR-22/23/24 position, CRLF, splice-orphan) were not touched; two are one-line fixes.

## 1. Strengths

- **Every BR-19 guard was falsified, not asserted.** I independently reverted each in a scratch tree: the read guard (`gitignore.go:209-217`), the dedupe (`gitignore.go:181-189`), and the duplicate-open-marker error (`gitignore.go:132-136`) each turn its named test red. The retired tautological namesake is genuinely gone, not left beside the real test.
- **`materializationFaultFS.ReadFile`** (`apply_test.go:670-677`) is the right shape — it extends the existing fault seam rather than inventing a second one, and matches its siblings' `operation`-only convention.
- **BR-25's union fix is correct at CI granularity.** `46-…sh:54-58` unions per-commit removals across `git rev-list`; I confirmed the two-point diff would still report zero and the union reports the stale comment. 0.6s over 20 commits — cheap enough to keep.
- **ARCH-PURE is clean.** `mergeManagedBlock` is string-in/string-out and every unit test runs with no fake; the only IO change is confined to `applyEnsureGitignore`. The docstring at `gitignore.go:113-115` is the one restatement that *does* name the position side-effect honestly.
- **`legacyBlanketEntries` names its own end** (`gitignore.go:96-98`, "delete once `git grep` finds no repo carrying these lines… checked at #239's close") — a textbook ARCH-FUNERAL entry for a migration aid.

## 2. Critical findings

None.

## 3. Important findings

**I1 — `scripts/merge-checks.d/46-removed-symbol-references.sh`: three shapes it is credited with, all measured green** (`verification-cannot-fail`, 6th in family)

Per the family protocol I am **not** asking for these three sites to be patched. The rule is: *a check's falsification must live in the repo as a fixture the runner re-executes, not as a one-time manual sweep recorded in prose.* Neither `45-` nor `46-` has a red/green fixture, so the ledger's ✓ column is produced by remembering, and one row is already wrong. Measurements:

- `46-…sh` vs the ledger's "`gitignore.go` header ✓" (issue Revisions, round 6). I reverted `cmd/weave/internal/plan/gitignore.go:21` to its exact base text (`315579a:21` — `// ensure-text transform live here as data + a string function; the actual`), committed, and ran over `merge-base(main,HEAD)..HEAD`: **exit 0**. That site never named a declared symbol, so 46 cannot see it and never could.
- `46-…sh:81-84` — the historical-mention allowlist matches bare substrings `*remov*`/`*replac*`/`*delet*`. I injected `// ensureGitignoreText appends absent entries and never removes.` — a stale restatement of precisely the append-only mechanism the check exists for — and it was **exempted** (exit 0) because the line contains "removes".
- `46-…sh:46-52` — `extract_removed` only matches declarations at column 0. `managedBlockOpen`/`managedBlockClose` are declared inside a grouped `const ( … )` at `gitignore.go:83-86`, in the very file that motivated the check; a rename emits `-\tmanagedBlockOpen  = …` with no `-const ` prefix. Run against a synthetic hunk, the awk program extracted `legacyBlanketEntries` and `ReadFile` and **not** `managedBlockOpen`.

Supporting scope note: `46-…sh:79` excludes `*_test.go`, so four functions named `TestEnsureGitignoreText*` (`gitignore_test.go:33,44,63,96`) survive at HEAD for a function this range deleted.

Fix sketch (the class, not the instances): add a fixture harness — a Go test or `scripts/merge-checks.d/test/` — that builds a throwaway git repo per motivating site, asserts the check exits 1 on each and 0 on a clean tree, and registers it so CI re-runs the falsification. Then correct the round-6 ledger row to state what 46 covers (comments naming a *declared, ungrouped* symbol) and what it does not (prose describing retired behaviour without naming one) instead of a ✓. Narrowing the allowlist to adjacent retirement phrases and teaching `extract_removed` grouped declarations are the two cheap sub-fixes that fixture would drive.

**I2 — `atlas/workflow/ci-merge-check.md:32` hand-enumerates ariadne's merge checks and is two entries stale** (`hand-maintained-restatement-of-model`, 4th in family)

Again the rule, not the instance. `**Checks in ariadne today:** 30-weave-drift.sh … and 40-duplicate-issue-id.sh` restates the contents of `scripts/merge-checks.d/`, which now also holds `45-verb-enumeration.sh` (M1) and `46-removed-symbol-references.sh` (this round). That is #239's own thesis one level down: a hand-maintained restatement of a model is a deferred consumer, and it went stale across two consecutive boundaries without anything failing. The fix that covers the class is to stop enumerating — have the page name `scripts/merge-checks.d/` as the source and describe only the *contract* and the one check (`40-`) whose rationale is load-bearing, or generate the list. Rider while editing: the fence opened at `:31` does not close until `:47`, so this entire paragraph currently renders as a code block (pre-existing, out of window).

## 4. Minor findings

- `46-…sh:28-31` — invoked with no range it prints `✓ … nothing to compare` and exits 0, while its sibling `45-` falls back to scanning the whole tree. A bare local run reports success having checked nothing.
- `46-…sh:77` — `IFS=: read -r _rev f lineno line` truncates on a path containing `:`. Harmless for `*.go`/`*.md` today.
- `45-…sh:62` excludes `workshop/{plans,issues}` + `lessons.md`; `46-…sh:79` excludes all of `workshop/*`. The two siblings disagree on scope for no stated reason.
- `go test ./...` is red at HEAD: `cmd/sdlc/fleet_plan_test.go:14` opens `workshop/plans/000200-sdlc-fleet-thread-inventory-plan.md`, archived to history at `dfeba9c`. Out of window and not M2's doing, but it means the suite-wide green this boundary leans on is not actually clean.

## 5. Test coverage notes

- The pure transform is well covered and non-tautological: `TestManagedBlockReplacesWholesaleSoRetiredEntriesDisappear` feeds a real `/RETIRED.md` and asserts both directions; `TestManagedBlockRefusesMalformedMarkers` asserts the error *names the remedy*.
- **`TestCompileEnsuresGitignore` (`main_test.go:198`) cannot distinguish M2 from what it replaced.** It asserts entry presence (`strings.Contains(got, entry+"\n")`) and byte-idempotence — both true of the retired append-only mechanism. The milestone's headline behaviour has no end-to-end assertion; the plan predicted this test would "survive unchanged", so it is not a deviation, but one `strings.Contains(got, managedBlockOpen)` would close it.
- No test covers a `.gitignore` whose repo-owned lines sit *below* the block, or a CRLF file — see BR-22/BR-23 below, both still measurable at HEAD.

## 6. Architectural notes for upcoming work

- **ARCH-DRY** — pass on the code. Flagged at I2 for the atlas enumeration. `mergeManagedBlock`'s `absorb` and `emitted` maps look similar but answer different questions; not duplication.
- **ARCH-PURE** — pass. Pure transform, IO confined to the seam, no mocks needed to run the unit tests.
- **ARCH-PURPOSE** — flagged at I1. The shadow-sweep over "what does weave own in `.gitignore`" finds four prose consumers (`.gitignore:17-24`, `README.md:40-56`, `atlas/workflow/weave.md:101-118`, `gitignore.go:105-125`) and they now disagree on precision: only the code docstring states the position side-effect. M3/M4's `commitConsumption` provenance filter will be the first consumer that *derives* from the block — that is the one that matters.
- **ARCH-MOCK** — pass for the FS seam. The two merge checks shell directly to `git` with no seam or fixture; that gap is exactly what I1 asks for.
- **ARCH-CONSTRAINTS** — pass. 46 is O(commits) git invocations, 0.6s over this 20-commit range, bounded by PR size since `merge-check.yml` passes `merge-base..head`.
- **ARCH-SECURE** — mostly pass: `.gitignore` is correctly treated as untrusted (fail closed on unreadable, fail closed on unparseable markers, remedy named). The one hole is CRLF, where malformed-for-this-parser input is silently *accepted* into a duplicated-block state instead of failing closed — BR-23, still open.
- **ARCH-ORDER** — the `.gitignore` is state carried across weave runs, and the CRLF path reaches a state the transform can never leave (second pass returns `changed=false` with two blocks). That is a legal-state gap, not just a formatting nit.
- **ARCH-FUNERAL** — pass, and well done on `legacyBlanketEntries`. The one unnamed residue is the orphaned CRLF block: a durable artifact the wholesale-replace path cannot reach.

## 7. Plan revision recommendations

`workshop/plans/000239-…-plan.md` was not modified in this window and now over-claims in one place:

- **Task 2.2, Step 3** still says `TestEnsureGitignoreTextDedupsRepeatedInputEntry` "needs a decision: … Dedupe inside `IgnoreEntries` (Task 3.1 …) and repoint this test there in M3." The dedupe instead landed in `mergeManagedBlock` (`gitignore.go:181-189`) in M2, driven by BR-19, and the namesake test was deleted rather than repointed. Add a `## Revisions` entry recording that the dedupe is now the block emitter's, so Task 3.1 does not add a second one.
- **Task 2.1, Step 3's code listing** shows `block := append([]string{managedBlockOpen}, entries...)`, superseded by the same change. Same Revisions entry.
- **Task 4.3 (per-repo pilot)** should gain the `parley.nvim` rehearsal as a checkable step — it currently exists only as an advisory paragraph in the issue's Revisions (BR-24).

```findings
dispose:
  - id: BR-19
    disposition: addressed
    note: |
      All three falsified in a scratch copy of a62d20a — read guard, input dedupe and quoted-marker error each go red when reverted; the tautological namesake is deleted and the issue Log's 50-base-layer-tests.sh claim now reads "deferred to M3".
  - id: BR-22
    disposition: not-addressed
    note: |
      Measured at HEAD - mergeManagedBlock(block + "!/CLAUDE.md\n", ["/CLAUDE.md"]) still returns !/CLAUDE.md above the block; atlas/workflow/weave.md still says "Outside is the repo's own, preserved verbatim" with no position clause, and this round ADDED a second restatement carrying the same gap at README.md:49 ("lines there are preserved verbatim, negations included"). No Log line. Only gitignore.go:113-115 states it.
  - id: BR-23
    disposition: not-addressed
    note: |
      Measured at HEAD - a CRLF .gitignore gains a second LF block (2 open markers in the result) and the second pass returns changed=false, so the original is a permanent orphan the wholesale-replace path can never reach. gitignore.go:131 still compares raw lines with no TrimRight(line, "\r").
  - id: BR-24
    disposition: addressed
    note: |
      The rule is recorded durably in the issue's Revisions (000239…md:683-687) with the actor swapped as the finding asked; it is not yet a checkable step in the plan's Task 4.3, which is the plan-revision recommendation above rather than an open finding.
  - id: BR-25
    disposition: addressed
    note: |
      Union fix verified at CI granularity - reintroducing the born-and-buried SeedOnceSlotIsRepoOwned comment at golden.go:222 exits 1 over merge-base(main,HEAD)..HEAD, and the sweep it demanded produced a real fix in 45- (case-insensitivity for the CamelCase Action sum type). The one overstated ledger row is raised separately below rather than re-raised here.
findings:
  - id: new
    severity: Important
    family: verification-cannot-fail
    title: |
      46-removed-symbol-references.sh prints green on three shapes it is credited with, including one the round-6 ledger marks verified
    detail: |
      This is the 6th finding in family verification-cannot-fail. Do NOT fix these three sites — the RULE is that a check's falsification must live in the repo as a fixture the runner re-executes, not as a one-time manual sweep recorded in prose; neither 45- nor 46- has a red/green fixture, which is why one ledger row is already wrong. Measured, all over merge-base(main,HEAD)..HEAD in a scratch clone.
      (1) The ledger's "46 (BR-20's sites) - gitignore.go header ✓" - reverting gitignore.go:21 to its exact base text (315579a:21, "ensure-text transform") and re-running exits 0. That site never named a declared symbol.
      (2) 46-…sh:81-84 exempts any line containing the bare substring "remov"/"replac"/"delet"; the injected stale restatement "ensureGitignoreText appends absent entries and never removes." exits 0.
      (3) 46-…sh:46-52 extracts only column-0 declarations, so managedBlockOpen/managedBlockClose - grouped in const ( … ) at gitignore.go:83-86, the file that motivated the check - are invisible to a rename. Verified against a synthetic hunk - legacyBlanketEntries and ReadFile extracted, managedBlockOpen not.
      Scope rider - 46-…sh:79 excludes *_test.go, so four TestEnsureGitignoreText* functions (gitignore_test.go:33,44,63,96) still name the function this range deleted. Fix - a fixture harness that builds a throwaway repo per motivating site and asserts exit 1 on each / 0 clean, then state 46's real coverage in the ledger instead of a ✓.
  - id: new
    severity: Important
    family: hand-maintained-restatement-of-model
    title: |
      atlas/workflow/ci-merge-check.md hand-enumerates ariadne's merge checks and is two entries stale, including the one added this round
    detail: |
      This is the 4th finding in family hand-maintained-restatement-of-model. Do NOT just append the two names — the RULE is that the atlas names the source rather than restating its contents. ci-merge-check.md:32 says "Checks in ariadne today - 30-weave-drift.sh … and 40-duplicate-issue-id.sh"; scripts/merge-checks.d/ also holds 45-verb-enumeration.sh (M1) and 46-removed-symbol-references.sh (this round), so the list went stale across two consecutive boundaries with nothing failing — #239's own thesis one level down. Have the page point at the directory and keep only 40-'s load-bearing rationale, or generate the list. Rider while editing - the fence opened at :31 does not close until :47, so the whole paragraph currently renders as code (pre-existing, out of window).
  - id: new
    severity: Minor
    family: check-invocation-modes-diverge
    title: |
      46-…sh with no range is a silent no-op pass while its sibling 45- scans the whole tree
    detail: |
      46-…sh:28-31 exits 0 with "no commit range given — nothing to compare", so a bare local invocation reports success having checked nothing; 45-…sh:51-55 falls back to the full tree in the same situation. The two siblings also disagree on exclusions (45- excludes workshop/{plans,issues} + lessons.md; 46- excludes all of workshop/*) with no stated reason.
  - id: new
    severity: Minor
    family: archived-artifact-breaks-pinned-reference
    title: |
      go test ./... is red at HEAD on an out-of-window failure, so the suite-wide signal this boundary leans on is not clean
    detail: |
      cmd/sdlc/fleet_plan_test.go:14 (TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory) opens workshop/plans/000200-sdlc-fleet-thread-inventory-plan.md, which was archived to workshop/history at dfeba9c. Not caused by this range — the test file last changed at c1bae9b — but it means "go test ./... green" cannot be cited as evidence at this or any later boundary until the test reads the archived path or is retired.
```

---

## Re-review — 2026-09-19T23:30:24-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 239 — Minimal committed base-layer surface |
| repo | ariadne |
| issue file | workshop/issues/000239-minimal-committed-base-layer-surface.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 315579a6c859f32d9be3ea5da144200a23ec4767..95333045220aafb9d3f359487704a6700c01cc8a |
| command | sdlc milestone-close --issue 239 --milestone M2 |
| reviewer | claude |
| timestamp | 2026-09-19T23:30:24-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

M2 delivers its mechanism and I verified it rather than reading about it: `mergeManagedBlock` is a genuine pure transform, ariadne's live `.gitignore` is an exact fixed point of it (`changed=false, identical=true`, probed against the real file from a scratch clone at 9533304), and all three BR-19 guards go **red** when reverted in a scratch copy (read-fault guard, input dedupe, quoted-marker pin). BR-26's demanded falsification harness is real: I broke the grouped-declaration rules, the per-commit union, and the allowlist one at a time and `construct/scripts/test/merge-checks.test.sh` reported `46/grouped-const-decl`, `46/born-and-buried`, and three cases FAIL respectively. What holds this back from SHIP is that the verification layer added this round has two measured holes of its own — `46`'s declaration grammar is written twice and the halves disagree in *both* directions (it false-positives on a grouped-const reorder and is still blind to a grouped rename deeper than the diff's context, which is the exact `managedBlockOpen` shape it was built for), and the new harness's red assertions report `ok` even when every fixture fails to build and nothing is ever checked. Both prior Minors (BR-22 block position, BR-23 CRLF) remain live and measured at HEAD for a third round.

## 1. Strengths

- **The falsification harness genuinely falsifies.** I reverted three separate mechanisms in `scripts/merge-checks.d/46-removed-symbol-references.sh` in a scratch clone and each produced a distinct FAIL row. The `46/plain-removed-func` case also went red when I reverted the allowlist to bare substrings — confirming the "strip the symbol name first" rule (46:111) is load-bearing, not decorative. This is the first artifact in this issue that answers `verification-cannot-fail` with something executable.
- **BR-19's three guards are all reachable and all falsifiable** — `TestApplyEnsureGitignoreFailsClosedOnReadError` (`gitignore_test.go:293`) fails without the `os.IsNotExist` branch, `TestManagedBlockDedupsRepeatedInputEntry` (`:311`) emits `/AGENTS.md` twice without the dedupe, `TestManagedBlockTreatsAQuotedMarkerAsAMarker` (`:330`) fails without the duplicate-open guard. Verified individually.
- **BR-27's repair is the right shape, not an append.** `atlas/workflow/ci-merge-check.md:35` now says *which checks exist is `ls scripts/merge-checks.d/`* and keeps only `40-duplicate-issue-id.sh`'s rationale — the one thing a filename cannot carry. The fence that had been rendering the section as code is closed (exactly two fences, lines 31 and 33).
- **ARCH-PURE is intact.** `mergeManagedBlock` is string→string; every one of its ten tests runs with no fake, and the only IO change is confined to `applyEnsureGitignore` with the fault injected through `weavefs.FS`.
- **`TestManagedBlockIdempotentWhenAllPresent` (`:63`) builds its fixture from `GeneratedRuntimeGitignoreEntries`**, so M3's list swap cannot silently desync it.

## 2. Critical findings

None.

## 3. Important findings

**I1 — `46`'s declaration grammar is written twice and the two halves disagree** (`scripts/merge-checks.d/46-removed-symbol-references.sh:69-73` vs `:91`). *5th in family `presence-predicate-written-twice` — do not fix the instance.* `extract_removed` was made group-aware this round; `still_declared` still greps only column-0 declarations. Measured both directions in throwaway repos:

- *False positive.* A pure reorder of two members inside `const ( … )` — nothing removed — makes the check exit 1: `pkg/a.go:8  names removed symbol alpha`, flagging a correct comment. On ariadne's own PRs that is a red CI for a cosmetic change, which `45`'s own header calls the failure mode that gets a check routed around.
- *False negative.* Renaming member 8 of a 12-member `const ( … )` group prints `✓ no top-level symbols removed in range`, because the `const (` opener falls outside the 3-line diff context so `ingroup` is never set. That is the `managedBlockOpen`/`managedBlockClose` shape at `gitignore.go:83-86` — the site the fix exists for. The harness's `46/grouped-const-decl` fixture passes only because its group has one member and its opener is inside the hunk.

The rule: **"line L declares symbol N" must have one definition, applied to both the removed side and the still-declared side, and resolved against the file at each commit (`git show <c>:<path>`) rather than against hunk context.** The fixture harness must then carry each grammar shape in *both* directions (removed-and-gone → red, removed-but-still-declared → green) at a size where context does not reach the group opener. Building the fixture to the implementation's minimum is what let this ship green.

**I2 — the falsification harness is itself unrun, and its red assertions cannot fail for the right reason** (`construct/scripts/test/merge-checks.test.sh:51`). *7th in family `verification-cannot-fail` — do not fix the instance.* BR-26's rule was "a fixture **the runner re-executes**". The fixture landed; the runner half did not, and the assertion half is weaker than it looks.

- *No runner.* `grep -rn "scripts/test" Makefile Makefile.workflow .github/ scripts/` finds no invocation of any of the **nine** files in `construct/scripts/test/`. The plan's Task 3.3 Step 3 still says `50-base-layer-tests.sh` runs "**both** tests" (portable-makefile + gitignore-surface), so M3 registers 2 of 9 and this harness is orphaned by construction. It has no Core-concepts row and no atlas mention, unlike `base-layer.md:214` and `setup-and-replication.md:150` which do inventory their siblings.
- *Red assertions pass on a broken fixture.* `run46` discards stdout/stderr and treats any non-zero exit as "the check fired". I broke every fixture build (made `git commit` fail) leaving the checks untouched: the harness reported **`== 6 passed, 4 failed ==`** — all six red-direction cases said `ok` while nothing was ever checked. Only the green-direction cases noticed.

The rule: **registration derives from `construct/scripts/test/*.test.sh` rather than naming files** — the same derive-don't-enumerate repair BR-27 just applied one level up, applied to the class instead of the instance — **and a red assertion must match the check's own failure signal** (its message, or exit 1 specifically), with the fixture build asserted before the check runs. Pinning ambient git state (`-c commit.gpgsign=false`, `-c init.defaultBranch=main`) belongs in the same edit.

## 4. Minor findings

- `mergeManagedBlock` does not validate `entries`: an empty member makes `absorb[""]` true, silently deleting **every blank line** outside the block — measured, `"# group one\nbin/\n\n# group two\ncache/\n"` loses its separator. Unreachable at M2's hardcoded nine; reachable once M3 derives the list. One-line guard.
- `46` globs `'*.md'` but its match regex requires a line starting with `//` or `*`, so Markdown is effectively uncovered — the atlas-prose restatement class (BR-27's own site) could never be caught by it.
- `45` and `46` still disagree on exclusions with no stated reason (`45:52,54` excludes only `workshop/history/`; `46:102` excludes all `workshop/*`) — the BR-28 rider.
- The entry dedupe landed in `mergeManagedBlock` (`gitignore.go:176-186`), but plan Task 2.2 Step 3 and Task 3.1 still place it in `IgnoreEntries`; after M3 the same normalization exists twice.

## 5. Test coverage notes

Scoped `go test ./cmd/weave/...` is green. Suite-wide `go test ./...` is **red** at HEAD on `TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory` (`cmd/sdlc/fleet_plan_test.go:14`) — confirmed by running it; out of window, tracked as `workshop/issues/000210-fleet-plan-test-hardcoded-path.md`. `46` costs ~1.05s over this branch's 21 commits (O(commits) `git diff` invocations) — fine for a PR range (ARCH-CONSTRAINTS pass). The gap the diff could ship and does not cover: no test asserts where the block lands relative to repo-owned lines (I1's sibling, BR-22), and none feeds a CRLF file (BR-23).

## 6. Architectural notes

**ARCH-DRY** flag (I1, and the hand-enumerated registration list in I2). **ARCH-PURE** pass — pure transform, injected IO, mock-free unit tests. **ARCH-PURPOSE** flag — BR-27's derive-don't-enumerate repair was applied to the atlas but the identical hand-enumeration survives in Task 3.3; that is the instance, not the class. **ARCH-MOCK** pass with a note — 45/46 fixtures drive real `git` in throwaway repos, the right seam, but the harness does not pin the ambient gitconfig it inherits. **ARCH-CONSTRAINTS** pass. **ARCH-SECURE** flag — `.gitignore` is untrusted input and the read/parse paths do fail closed with a remedy, but CRLF is a representable input that silently produces a second block instead of failing, contradicting `atlas/workflow/weave.md:115`'s claim that the transform "fails closed on any marker shape it cannot parse". **ARCH-ORDER** flag — the transform's state space over its input is (no block, one block, malformed); CRLF adds a fourth that maps to "no block" because the alphabet is not normalized at the boundary. **ARCH-FUNERAL** pass — `legacyBlanketEntries` names its end and Task 4.3 Step 5 checks it; the block is replaced wholesale so it cannot grow unbounded; the harness traps its scratch dir.

For M4: BR-23's orphan block matters more than its Minor severity suggests — `commitConsumption`'s provenance filter keys on the managed block's **line range**, and a CRLF repo would present two blocks with the wrong one authoritative.

## 7. Plan revision recommendations

1. `## Revisions` entry for the M1/M2 review rounds, adding Core-concepts rows for `scripts/merge-checks.d/45-verb-enumeration.sh`, `scripts/merge-checks.d/46-removed-symbol-references.sh` and `construct/scripts/test/merge-checks.test.sh` — all three are new repo files, two of them CI gates that can fail any ariadne PR, and none appears anywhere in the plan today.
2. Task 3.3 Step 3: replace "running **both** tests" with a glob over `construct/scripts/test/*.test.sh` (nine files exist, two are named).
3. Task 2.2 Step 3 / Task 3.1: record that the dedupe landed in `mergeManagedBlock`, and name the single owner so M3 does not implement it twice.
4. Record BR-22 (block position inverts precedence for lines placed after it) and BR-23 (CRLF produces a permanent orphan block) as known limitations of `mergeManagedBlock`, with the M4 dependency on the line-range provenance filter stated.

```findings
dispose:
  - id: BR-22
    disposition: not-addressed
    note: |
      Measured at HEAD in a scratch clone - mergeManagedBlock(block + "!/CLAUDE.md\n", ["/CLAUDE.md"]) still returns !/CLAUDE.md ABOVE the block; atlas/workflow/weave.md:112 (text added THIS round) says "Outside is the repo's own, preserved verbatim" with no position clause, README.md:49 says the same, and there is still no issue Log line. The .gitignore comment added this round ("Add your own entries ABOVE it") is the nearest approach but never says a line placed below is relocated above and can have its precedence inverted.
  - id: BR-23
    disposition: not-addressed
    note: |
      Measured at HEAD - a CRLF .gitignore carrying a block gains a second LF block (2 open markers) on pass 1, and pass 2 returns changed=false err=nil, so the original is a permanent silent orphan. gitignore.go:131 still compares raw lines with no TrimRight(line, "\r"). This also contradicts atlas/workflow/weave.md:115's new claim that the transform "fails closed on any marker shape it cannot parse".
  - id: BR-26
    disposition: addressed
    note: |
      construct/scripts/test/merge-checks.test.sh exists and genuinely falsifies - I reverted the grouped-decl awk rules, the per-commit union, and the allowlist separately in a scratch clone and got 46/grouped-const-decl FAIL, 46/born-and-buried FAIL, and 3 FAILs respectively. 46's real coverage is stated as fixtures in the Revisions rather than a prose tick, and the four TestEnsureGitignoreText* functions are now TestManagedBlock*. Two residues raised as new findings (harness unrun; red assertions pass on a broken fixture).
  - id: BR-27
    disposition: addressed
    note: |
      atlas/workflow/ci-merge-check.md:35 now says which checks exist is "ls scripts/merge-checks.d/" and keeps only 40-duplicate-issue-id.sh's rationale; the stale two-name list is gone and the fence opened at :31 closes at :33 (exactly two fences in the file), so the section no longer renders as code.
  - id: BR-28
    disposition: addressed
    note: |
      46-...sh:32-39 now derives merge-base(origin/main|main, HEAD) when no range is given and exits 1 if it cannot, so a bare run can no longer print a vacuous green. The exclusion-divergence rider is still open (45:52,54 excludes only workshop/history/; 46:102 excludes all workshop/*) and is re-noted as a Minor rather than re-raised.
  - id: BR-29
    disposition: addressed
    note: |
      Confirmed still red at HEAD - go test ./... fails on TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory (cmd/sdlc/fleet_plan_test.go:14), out of window, and now tracked as workshop/issues/000210-fleet-plan-test-hardcoded-path.md. The Revisions record that suite-wide green is not citable at this or any later boundary and that the scoped go test ./cmd/weave/... (verified green) is what is being cited instead. Prose correction, inspected against the source.
findings:
  - id: new
    severity: Important
    family: presence-predicate-written-twice
    title: |
      46's declaration grammar is written twice and the halves disagree - it false-positives on a grouped-const reorder and is still blind to the grouped rename it was built for
    detail: |
      This is the 5th finding in family presence-predicate-written-twice. Do NOT fix these two
      sites - the RULE is that "line L declares symbol N" must have ONE definition applied to
      both the removed-side extraction and the still-declared lookup, resolved against the FILE
      at each commit (git show <c>:<path>) rather than against hunk context. Measured in
      throwaway repos, both directions.
      (1) FALSE POSITIVE - extract_removed (46:69-73) was made group-aware this round but
      still_declared (46:91) greps only column-0 declarations. A pure reorder of two members
      inside const ( ... ) removes nothing, yet the check exits 1 with "pkg/a.go:8  names removed
      symbol alpha", failing CI on a correct comment. 45's own header calls a check that fires on
      legitimate prose the failure mode that gets a check routed around.
      (2) FALSE NEGATIVE - renaming member 8 of a 12-member const group prints
      "no top-level symbols removed in range", because the const ( opener falls outside the
      3-line diff context so ingroup is never set. That is exactly the
      managedBlockOpen/managedBlockClose shape at gitignore.go:83-86 that motivated the fix.
      The harness's 46/grouped-const-decl case passes only because its group has one member and
      its opener sits inside the hunk - the fixture was built to the implementation's minimum,
      not to the shape at real size. The harness therefore needs each grammar shape in BOTH
      directions and at a size where context does not reach the opener.
  - id: new
    severity: Important
    family: verification-cannot-fail
    title: |
      The falsification harness BR-26 demanded is itself unrun, and its red assertions report ok even when every fixture fails to build
    detail: |
      This is the 7th finding in family verification-cannot-fail. Do NOT fix the instance - the
      RULE has two halves and this round delivered one and a half of them. BR-26 stated it as
      "a fixture THE RUNNER RE-EXECUTES"; add to it "and a red assertion must match the check's
      own failure signal, not merely a non-zero exit". Two measured instances.
      (1) NO RUNNER. grep -rn "scripts/test" over Makefile, Makefile.workflow, .github/ and
      scripts/ finds no invocation of any of the NINE files in construct/scripts/test/. The
      plan's Task 3.3 Step 3 still says 50-base-layer-tests.sh runs "both tests"
      (portable-makefile + gitignore-surface), so M3 registers 2 of 9 and merge-checks.test.sh
      is orphaned by construction. It has no Core-concepts row and no atlas mention, unlike
      base-layer.md:214 and setup-and-replication.md:150 which inventory their siblings. The
      class fix is to glob construct/scripts/test/*.test.sh - the same derive-don't-enumerate
      repair BR-27 just applied to ci-merge-check.md, applied to the class this time.
      (2) RED ASSERTIONS CANNOT FAIL FOR THE RIGHT REASON. run46 (merge-checks.test.sh:51)
      discards stdout/stderr and treats ANY non-zero exit as "the check fired". I broke every
      fixture build (made git commit fail) while leaving both checks untouched - the harness
      reported "== 6 passed, 4 failed ==", every red-direction case saying ok while nothing was
      ever checked. Only the four green-direction cases noticed. Pinning the ambient gitconfig
      the harness inherits (-c commit.gpgsign=false, -c init.defaultBranch=main) belongs in the
      same edit.
  - id: new
    severity: Minor
    family: degenerate-input-not-rejected
    title: |
      mergeManagedBlock does not validate entries - an empty member silently deletes every blank line outside the block
    detail: |
      gitignore.go:159-165 builds the absorb set straight from entries, so an empty member makes
      absorb[""] true and every blank line outside weave's region is dropped. Measured -
      "# group one\nbin/\n\n# group two\ncache/\n" comes back with its separator gone, a
      reformat of a file weave does not own. A member equal to a marker is the same class: it
      lands inside the block and makes the NEXT run hard-fail on a duplicate marker. Unreachable
      at M2's hardcoded nine, reachable once M3 derives entries from the action list. One guard
      at the top of the entry loop.
  - id: new
    severity: Minor
    family: hand-maintained-restatement-of-model
    title: |
      46 globs '*.md' but its match regex requires a // or * line prefix, so Markdown prose is effectively uncovered
    detail: |
      The pathspec at 46:119 includes '*.md' and the header speaks of "comment lines", but the
      regex "^[[:space:]]*(//|\\*).*NAME" only matches Go comments and Markdown bullet lines. A
      stale atlas paragraph naming a deleted symbol - the exact class BR-27 was about - passes.
      Either drop '*.md' from the pathspec or give Markdown its own line predicate, so the
      check's stated scope and its real scope agree.
  - id: new
    severity: Minor
    family: check-invocation-modes-diverge
    title: |
      45 and 46 still disagree on which paths they exclude, with no stated reason
    detail: |
      45:52 and 45:54 exclude only workshop/history/; 46:102 excludes all of workshop/*. BR-28's
      main claim is addressed (46 now derives its range) but this rider is not. Siblings
      enforcing the same rule on the same tree should share one exclusion set, or each should
      say why it differs.
  - id: new
    severity: Minor
    family: presence-predicate-written-twice
    title: |
      The entry dedupe landed in mergeManagedBlock but the plan still assigns it to IgnoreEntries, so M3 will implement it twice
    detail: |
      gitignore.go:176-186 dedupes the entry list (the BR-19 repair). Plan Task 2.2 Step 3 says
      "Dedupe inside IgnoreEntries (Task 3.1 dedupes and sorts anyway)" and Task 3.1's body at
      plan line 1279 still calls sort.Strings over a deduped slice. Pick one owner and record it
      in the plan's Revisions before M3 lands both.
```
