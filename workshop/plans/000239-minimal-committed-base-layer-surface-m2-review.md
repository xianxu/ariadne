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
