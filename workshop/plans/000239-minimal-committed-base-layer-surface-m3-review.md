# Boundary Review — ariadne#239 (milestone M3)

| field | value |
|-------|-------|
| issue | 239 — Minimal committed base-layer surface |
| repo | ariadne |
| issue file | workshop/issues/000239-minimal-committed-base-layer-surface.md |
| boundary | milestone M3 |
| milestone | M3 |
| window | d0a92cb6da121b60d109c1af08c391901b8ceda5..450b67869efa0a4cd8759aeee35656d43f677e78 |
| command | sdlc milestone-close --issue 239 --milestone M3 |
| reviewer | claude |
| timestamp | 2026-09-20T09:02:22-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

M3's core is right and I verified it beyond the shipped tests: `plan.IgnoreEntries` is a genuinely pure derivation keyed on the manifest verb's ownership class, the `TargetAll` pin is real (I confirmed the lean-target test is non-vacuous), and the conformance test's decisive assertion is falsifiable — making `Mkdir` emit an ignore entry turns it red with exactly the parley.nvim message. I went further and ran the derivation read-only across 12 fleet derivatives: **85–90 entries each, and in every repo zero derived entries correspond to a tracked non-symlink file or to a directory holding tracked files.** The per-path rule holds fleet-wide, which is the load-bearing claim for M4. What blocks SHIP is not the code: the milestone retired a hardcoded list and, in doing so, dropped the one test that pinned `/construct/generated/` to the production path (falsified — deleting `walk.GeneratedRel` from `main.go:674` leaves every Go test *and* the new bash conformance test green); the BR-30 fix to `46-removed-symbol-references.sh` states "one grammar, two uses" and then implements half of it (falsified — a pure reorder of a bare `iota` member false-positives the gate, and ariadne has three such enums); two of the milestone's own prose deliverables did not land; and the ARCH-PURPOSE shadow-sweep missed a surviving second `.gitignore` channel that `base.manifest:171` still ships to every derivative.

## 1. Strengths

- **`IgnoreEntries`'s `default:` case (`gitignore.go:105`) converts a doc promise into an enforced one.** "A new verb joins a class by adding one case" is normally a comment nothing can fail on; here an unhandled `Action` type returns a named error through `planActions`. `TestIgnoreEntriesRejectsUnclassifiedAction` asserts the type name is in the message, so the failure is diagnosable rather than mysterious.
- **The ownership axis is the right rule, and the test that proves it is the right test.** `TestIgnoreEntriesNeverIgnoresScaffoldOrTouch` pins the case the Spec's own flat framing ("every path weave creates is gitignored") would have gotten catastrophically wrong. That the plan caught its own Spec is the best thing in this milestone.
- **ARCH-MOCK called correctly.** `gitignore-surface.test.sh` runs the real `git` binary because the property under test *is* git's pattern semantics — a fake would only assert the author's belief about git. The `--no-index` handling of `check-ignore`'s index-awareness is a real trap avoided, and case 1 mirrors `commitConsumption`'s actual `git ls-files -i -c` rather than approximating it.
- **The live pair#64 instance in ariadne's own tree is genuinely closed.** `/.colima/` and `/construct/scripts/vm-log.sh` are gone from the block (`.gitignore` diff), `.colima/`'s 6 tracked files survive, and `git status` is clean.
- **`expect_red`/`expect_green` plus the self-test in `merge-checks.test.sh:183-193`** — a harness that fails when a missing check script would have reported `ok`. That is the guard on the guard, and it is rare to see it written.

## 2. Critical findings

None.

## 3. Important findings

**I-1 · `cmd/weave/main.go:674` — `/construct/generated/` is no longer pinned by any test.** *(family `verification-cannot-fail`, 7th)* Before M3, `TestGeneratedRuntimeGitignoreCoversConstructGenerated` asserted against the *production* `GeneratedRuntimeGitignoreEntries`. It now asserts against `sampleEntries`, a fixture that passes `walk.GeneratedRel` in itself — so it verifies that `IgnoreEntries` returns an argument it was handed. Falsified: I replaced `[]string{walk.GeneratedRel}` with `nil` at `main.go:674` in a scratch copy; `go test ./cmd/weave/...` stayed green (including `TestCompileEnsuresGitignore`, which derives its own expectation from the same `planActions`) and `gitignore-surface.test.sh` printed PASS. `30-weave-drift.sh` doesn't cover it either — `weave-drift-check` (`Makefile.workflow:249`) only tests dynamic-skill render determinism. `walk/dynamic.go:41-46` still claims the gitignore entry "derives from this constant" and that the four consumers "MUST agree"; nothing now enforces it.

**I-2 · `scripts/merge-checks.d/46-removed-symbol-references.sh:101` — `still_declared` is blind to bare `iota` members, so a pure reorder fails CI.** *(family `presence-predicate-written-twice`, 7th)* The grouped branch requires an `=`: `^[[:space:]]+$1[[:space:]]*(=|[A-Za-z_*\[][^=]*=)`. The extractor's grouped branch (`:72`) requires only an indented identifier. Members 2..n of an `iota` block have no `=`. Falsified with a fixture: a `const (Alpha Kind = iota; Beta; Gamma)` block with `Beta` and `Gamma` swapped, plus an ordinary comment naming `Beta`, makes the check exit 1 with `pkg/b.go:3 names removed symbol Beta`. This is live in ariadne — `intent.Kind` (`intent.go:26`), `intent.Visibility` (`:68`) and `plan.SlotState` (`apply.go:277`) are all bare-`iota` enums whose members are named in comments throughout, and `intent.go:46` records that this codebase *does* move members within those blocks.

**I-3 · Two of M3's own prose deliverables did not land, and the atlas page they point at is stale in both directions.** *(family `atlas-states-future-state-as-current`, 2nd)* Plan Task 3.4 Step 3 names `atlas/workflow/weave.md:96` as a file to modify; `weave.md` is absent from the diff and `:118` still reads *"The entry list itself is still the fixed set as of M2; #239 M3 derives it from the manifest walk"* — future tense for work that just shipped, on the page the new target section links to. Plan Task 3.1 Step 3 says to rewrite `gitignore.go`'s file header; `:38` still says *"until then the fixed list below stands"* (there is no list below) and `:15` still enumerates *"the `.colima/` VM tree, the vm-log.sh helper"*, both deliberately dropped this milestone. Two neighbours in the same class: `atlas/workflow/base-layer.md:3` — first line of the adoption doc, edited this milestone — still tells adopters to use `construct/setup.sh`, which does not exist; and `action.go:33` says `WriteFile` is lowered "from `intent.Touch` (empty Content)" when `plan.go:114` lowers it to `Touch` — a reader classifying verbs for `IgnoreEntries` from that doc would conclude touch targets get ignored, the exact catastrophe `TestIgnoreEntriesNeverIgnoresScaffoldOrTouch` exists to prevent.

**I-4 · `construct/scripts/apply-gitignore-entries.sh` is a surviving second `.gitignore` channel, shipped fleet-wide by `construct/base.manifest:171`.** *(family `hand-maintained-restatement-of-model`, 6th)* It carries a hand-maintained `GITIGNORE_ENTRIES=(.goto .openshell/.bootstrap/ .openshell/.base-image-digest .DS_Store bin/)`, appends them with `grep -qxF` and never removes — append-only blanket directory globs, including `bin/`, the literal pair#64 pattern the Spec cites. It has **zero callers**: `grep -rn apply-gitignore-entries` over the tree returns the manifest row and its own usage comment, nothing else; its header says it was "extracted from `construct/setup.sh`", which weave retired. This directly contradicts the invariant `workshop/targets/base-layer-mechanics.md` records in this very diff — *"no artifact enters the ignore surface by a second channel either"*.

## 4. Minor findings

- `README.md:40-56` documents the managed block's *mechanism* (M2) but never says what determines its contents; an adopter can't learn that a manifest row now changes their `.gitignore`. *(family `adopter-facing-surface-undocumented`, 2nd — same rule as I-3: the adopter-facing surface has no derivation from the model.)*
- ARCH-CONSTRAINTS measurement from plan Task 3.2 Step 5 was not recorded in `## Log`. Measured here: `compile --dry-run` and `--dry-run --target claude` both ≈0.00 s warm despite the lean path now doing three full lowerings (was two); `50-base-layer-tests.sh` adds ~11 s to CI. Comfortably inside the declared envelope — worth writing down at close.
- `IgnoreEntries` emits derived paths verbatim into `.gitignore`'s glob language. A skill-directory name (discovered from the filesystem, not the manifest) containing `[`, `*`, `?`, or a leading `!`/`#` yields a pattern that matches something other than the literal path. ARCH-SECURE's "typed value at the boundary" lens; no escaping seam exists. *(new family `derived-value-unescaped-in-target-grammar`)*
- `50-base-layer-tests.sh:24,28` print `✓` and exit 0 when `cmd/weave` or `go` is absent. Defensible, and unreachable in CI (`merge-check.yml` sets up Go first), but it is the vacuous-pass shape the file's own header condemns — same rule as I-1.
- Absorbing a derivative's loose entries leaves their introducing comment block orphaned: parley.nvim's `.gitignore` has a four-line *"weave-generated runtime artifacts … skill symlinks, the merged settings.json, the .colima symlinks, and the vm-log.sh symlink"* comment that will head nothing after the migration, in ~12 repos. *(family `edit-splice-leaves-orphan-clause`, 3rd — weave correctly refuses to edit repo prose outside the block, so the rule is that the M4 sweep, not weave, owns removing the comment it orphans; add it to M4's per-repo checklist rather than teaching `mergeManagedBlock` to delete human text.)*
- `TestGeneratedRuntimeGitignoreCoversConstructGenerated` (`gitignore_test.go:104`) still carries the retired symbol's name.

## 5. Test coverage notes

Everything I could falsify, I did. `go test ./cmd/weave/...` green; `gitignore-surface.test.sh` PASS; `50-base-layer-tests.sh` green in ~11 s across all three suites; `46` clean over the M3 range. The conformance test's decisive assertion is real — patching `Mkdir` to emit `"/" + path + "/"` produces `FAIL: the sweep would untrack repo-owned scripts/merge-checks.d/20-vocabulary.sh`. The gap is I-1: the suite has no test that fails when the production wiring drops an entry source, because both the unit fixture and `TestCompileEnsuresGitignore` derive their expectations from the code under test. The cheapest fix is one assertion in `gitignore-surface.test.sh` on a literal — `grep -qxF '/construct/generated/' .gitignore` — which no refactor of the Go side can satisfy tautologically; that is the rule I-1 asks for, applied to every entry source rather than to `construct/generated` alone.

## 6. Architectural notes

**ARCH-DRY** — pass on the headline (one switch owns the rule; the `default:` makes it enforceable), flag on I-2 and I-4. For I-2, state the rule rather than adding a third regex: `extract_removed` and `still_declared` are two implementations of one grammar, "what is a Go declaration of NAME". Make them one — run the same awk over `git show $HEAD:<file>` (or over the `+`/context side) instead of a second regex pair — and make `merge-checks.test.sh`'s matrix the grammar's own enumeration (column-0, grouped-with-`=`, grouped-bare-`iota`, method receiver), so teaching the grammar a shape necessarily adds a fixture. A second regex only moves the boundary to the next shape.

**ARCH-PURE** — pass. `IgnoreEntries` is a total function over `[]Action`; its tests need no filesystem, no mocks. `planActionsCore` is the thin lowering, `applyEnsureGitignore` the only IO. Correct shape.

**ARCH-PURPOSE** — flag, and this is the one that matters for M4. I ran the shadow-sweep the lens asks for, fleet-wide: the *derived* consumer is clean in all 12 repos. But the enumeration was never written down, so `apply-gitignore-entries.sh` (I-4) survived it. The rule for I-3 and I-4 together: a doc line or script that restates a derived model needs either a derivation or a funeral. Two enforcements are cheap and sit beside checks that already exist — (a) fail a backticked path in `atlas/**.md` or `README.md` that does not resolve in the tree, which catches `construct/setup.sh`; (b) fail a `#<issue> M<n>`-tensed claim in atlas or Go comments once that milestone is closed, which catches `weave.md:118` and `gitignore.go:38`. Both are siblings of `46-`, which already greps `*.go` and `*.md` and already has the historical-mention allowlist to reuse.

**ARCH-MOCK** — pass, notably well (see Strengths).

**ARCH-CONSTRAINTS** — pass. Measured, inside budget; only the recording is missing.

**ARCH-SECURE** — pass on the load-bearing part: `.gitignore` is input weave did not produce and `mergeManagedBlock` fails closed on unparseable markers with the remedy in the message. Minor flag on unescaped entry emission.

**ARCH-ORDER** — pass. M3 adds no cross-event state. I checked the one ordering claim that could bite: a derivative jumping pre-M2 → post-M3. Verified against parley.nvim's real `.gitignore` — `/AGENTS.md`, `/CLAUDE.md`, `/GEMINI.md`, `/.claude/settings.json`, `/construct/scripts/vm-log.sh` are exact matches of derived entries (absorbed), and `/.claude/skills/`, `/.agents/skills/`, `/.colima/` are in `legacyBlanketEntries`. No loose weave entry is stranded.

**ARCH-FUNERAL** — flag on I-4. `legacyBlanketEntries` names its own end properly (`gitignore.go:152-154`, with the check anchored at #239's close) — that is the standard. `apply-gitignore-entries.sh` is the counterexample: created, orphaned, and still symlinked into every derivative with no removal path in the diff or in any routine. Retiring `base.manifest:171` is the funeral; do it in M4's sweep, where the manifest row's removal will also drop the symlink from all 12 repos in one pass.

For M4 specifically: the fleet shadow-sweep above is a reusable read-only gate — `weave compile --dry-run` + `git ls-files -s` per derived path, no writes — and it currently reports clean for all 12 repos. Running it as the pre-flight before each `propagate-base --repo` would turn Task 4.1's scratch-clone proof from a one-repo sample into a fleet-wide precondition.

## 7. Plan revision recommendations

- **Task 3.4**: record that Step 3 (`atlas/workflow/weave.md:96`) did not land, and that Step 1–2 did. As written the plan claims an atlas page this milestone did not touch.
- **Task 3.1 Step 3**: record that the file-header rewrite ("rewrite the file header's ownership paragraph") did not land; `gitignore.go:13-38` still describes the retired list.
- **Task 3.1 Step 1**: the plan's own test list promised `TestGeneratedRuntimeGitignoreCoversConstructGenerated` would be kept "now against `IgnoreEntries(nil, []string{walk.GeneratedRel})`". That is what shipped, and it is what makes the test tautological — note the consequence so M4 does not inherit the belief that `/construct/generated/` is covered.
- **Chunk 4 (M4)**: add `construct/base.manifest:171` / `construct/scripts/apply-gitignore-entries.sh` to the sweep as an explicit retirement, and add the orphaned-comment cleanup to the per-repo checklist.

```findings
findings:
  - id: new
    severity: Important
    family: verification-cannot-fail
    title: |
      /construct/generated/ is no longer pinned by any test — deleting walk.GeneratedRel from main.go:674 leaves the whole suite green
    detail: |
      This is the 7th finding in family `verification-cannot-fail`. Earlier rounds fixed
      instances; do NOT fix this instance alone. The RULE: a test whose expected value is
      derived by calling the code under test (or by passing the fixture in itself) cannot
      fail when the production wiring changes — every entry SOURCE needs one literal
      assertion somewhere no refactor can satisfy tautologically. Measured: replacing
      `[]string{walk.GeneratedRel}` with `nil` at main.go:674 in a scratch copy left
      `go test ./cmd/weave/...` green (TestCompileEnsuresGitignore derives wantEntries from
      the same planActions; TestGeneratedRuntimeGitignoreCoversConstructGenerated at
      gitignore_test.go:104 asserts sampleEntries contains an argument sampleEntries passed
      in) and `gitignore-surface.test.sh` printed PASS. 30-weave-drift.sh does not cover it
      — weave-drift-check (Makefile.workflow:249) only tests dynamic-skill render
      determinism. walk/dynamic.go:41-46 still asserts the gitignore entry "derives from
      this constant" and that the consumers "MUST agree". Same rule covers
      50-base-layer-tests.sh:24,28, whose skip-guards print a green checkmark and exit 0.
      Cheapest application of the rule: one literal `grep -qxF '/construct/generated/'`
      per entry source in gitignore-surface.test.sh.
  - id: new
    severity: Important
    family: presence-predicate-written-twice
    title: |
      46-removed-symbol-references.sh:101 still cannot see bare iota members, so a pure reorder false-positives CI
    detail: |
      This is the 7th finding in family `presence-predicate-written-twice`. Do NOT patch
      still_declared with a third regex. The RULE: `extract_removed` (:54-75) and
      `still_declared` (:99-102) are two implementations of one grammar — "what is a Go
      declaration of NAME" — and must be a single function parameterised by which side of
      the diff it reads (run the same awk over `git show $HEAD:<file>`). The BR-30 comment
      at :90-98 states exactly this rule and then implements half of it: the extractor's
      grouped branch needs only an indented identifier, the predicate's needs an `=`.
      Members 2..n of an iota block have no `=`. Falsified: const (Alpha Kind = iota; Beta;
      Gamma) with Beta and Gamma swapped, plus `// Beta is lowered by the planner.` in
      another file, exits 1 with "pkg/b.go:3 names removed symbol Beta". Live in this repo
      — intent.Kind (intent.go:26), intent.Visibility (:68), plan.SlotState (apply.go:277)
      are all bare-iota, and intent.go:46 records that members do get moved within them.
      Also make merge-checks.test.sh's matrix the grammar's own enumeration so a new shape
      cannot be taught to one half only.
  - id: new
    severity: Important
    family: atlas-states-future-state-as-current
    title: |
      Two of M3's prose deliverables did not land; weave.md:118 and gitignore.go:38 still describe the retired fixed list
    detail: |
      This is the 2nd finding in family `atlas-states-future-state-as-current` (mirror
      direction: current state stated as future). State the RULE rather than fixing the
      four sites: a doc or comment that names a milestone as pending, or names an artifact
      by path, is a claim with an expiry date that nothing enforces. Two enforcements sit
      beside 46- and reuse its allowlist — (a) fail a backticked path in atlas/**.md or
      README.md that does not resolve in the tree; (b) fail a `#<issue> M<n>` future-tense
      claim once that milestone is closed. Sites: atlas/workflow/weave.md:118 "still the
      fixed set as of M2; #239 M3 derives it" — plan Task 3.4 Step 3 named this file and it
      is absent from the diff, and it is the page the new target section links to;
      cmd/weave/internal/plan/gitignore.go:38 "until then the fixed list below stands" and
      :15 enumerating ".colima/ VM tree, the vm-log.sh helper" — plan Task 3.1 Step 3 named
      this rewrite; atlas/workflow/base-layer.md:3 sends adopters to construct/setup.sh,
      which does not exist, on the first line of the page this milestone edited;
      cmd/weave/internal/plan/action.go:33 says WriteFile is lowered "from intent.Touch
      (empty Content)" when plan.go:114 lowers it to Touch — a reader classifying verbs for
      IgnoreEntries from that doc reaches the catastrophe
      TestIgnoreEntriesNeverIgnoresScaffoldOrTouch exists to prevent.
  - id: new
    severity: Important
    family: hand-maintained-restatement-of-model
    title: |
      construct/scripts/apply-gitignore-entries.sh is a surviving second gitignore channel, shipped fleet-wide by base.manifest:171
    detail: |
      This is the 6th finding in family `hand-maintained-restatement-of-model`. The RULE to
      fix: a single-source change is not done until the CONSUMER ENUMERATION is written
      down and swept — the ARCH-PURPOSE shadow-sweep found the derived consumer clean in
      all 12 derivatives but never enumerated the non-derived writers, so this one
      survived. The script carries a hand-maintained GITIGNORE_ENTRIES=(.goto
      .openshell/.bootstrap/ .openshell/.base-image-digest .DS_Store bin/), appends with
      `grep -qxF` and never removes — append-only blanket directory globs, including `bin/`,
      the literal pair#64 pattern the Spec cites as the motivating hazard. It has zero
      callers: `grep -rn apply-gitignore-entries` over the tree returns only base.manifest:171
      and its own usage comment; its header says it was extracted from construct/setup.sh,
      which weave retired. This contradicts the invariant
      workshop/targets/base-layer-mechanics.md records in this same diff — "no artifact
      enters the ignore surface by a second channel either". ARCH-FUNERAL: retire the
      manifest row in M4's sweep so the symlink drops from all 12 repos in one pass.
  - id: new
    severity: Minor
    family: adopter-facing-surface-undocumented
    title: |
      README.md:40-56 documents the managed block's mechanism but never what determines its contents
    detail: |
      This is the 2nd finding in family `adopter-facing-surface-undocumented`. Same rule as
      the atlas finding above: the adopter-facing surface restates the model instead of
      deriving from it. An adopter reading README cannot learn that adding a manifest row
      now changes their .gitignore, nor that the entries became per-path. One sentence
      naming plan.IgnoreEntries and the ownership rule closes it.
  - id: new
    severity: Minor
    family: derived-value-unescaped-in-target-grammar
    title: |
      IgnoreEntries emits derived paths verbatim into git's glob language with no escaping
    detail: |
      ARCH-SECURE's "parse into a typed value at the boundary" lens. Skill directory names
      are discovered from the filesystem, not the manifest; one containing `[`, `*`, `?` or
      a leading `!`/`#` yields a pattern matching something other than the literal path —
      potentially a repo-owned file. No escaping seam exists between the derivation and the
      .gitignore writer.
  - id: new
    severity: Minor
    family: edit-splice-leaves-orphan-clause
    title: |
      Absorbing a derivative's loose entries orphans the comment block that introduced them, in ~12 repos
    detail: |
      This is the 3rd finding in family `edit-splice-leaves-orphan-clause`. The RULE here
      resolves in weave's favour and should be recorded as such: weave correctly refuses to
      edit repo prose outside the managed block (ARCH-SECURE), so the owner of the orphan is
      the M4 sweep, not mergeManagedBlock. Verified live — parley.nvim/.gitignore has a
      four-line "weave-generated runtime artifacts … skill symlinks, the merged
      settings.json, the .colima symlinks, and the vm-log.sh symlink" comment whose eight
      entry lines are all absorbed (five by exact match, three by legacyBlanketEntries),
      leaving the comment heading nothing. Add it to M4's per-repo checklist.
  - id: new
    severity: Minor
    family: verification-cannot-fail
    title: |
      ARCH-CONSTRAINTS measurement required by plan Task 3.2 Step 5 was not recorded in the Log
    detail: |
      Measured during this review, inside the declared envelope: `weave compile --dry-run`
      and `--dry-run --target claude` both ~0.00 s warm, despite the lean path now doing
      three full lowerings (was two); scripts/merge-checks.d/50-base-layer-tests.sh adds
      ~11 s to CI across its three suites; ariadne's block is 56 entries (predicted ~55), a
      derivative's 85-90 (predicted ~95). Worth writing into the Log at close so the budget
      has a datum rather than a prediction.
```

---

## Re-review — 2026-09-20T09:13:59-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 239 — Minimal committed base-layer surface |
| repo | ariadne |
| issue file | workshop/issues/000239-minimal-committed-base-layer-surface.md |
| boundary | milestone M3 |
| milestone | M3 |
| window | d0a92cb6da121b60d109c1af08c391901b8ceda5..384190b581062d24e80323b9f607dee439bc38f9 |
| command | sdlc milestone-close --issue 239 --milestone M3 |
| reviewer | claude |
| timestamp | 2026-09-20T09:13:59-07:00 |
| verdict | REWORK |

## Review

All inspections complete. Here is the review.

```verdict
verdict: REWORK
confidence: high
```

M3's core deliverable is solid and verified: `plan.IgnoreEntries` derives the ignore set from the planned actions by verb-ownership class, `GeneratedRuntimeGitignoreEntries` is gone, the derivation is pinned to `TargetAll`, the real-git conformance test passes, and ariadne's own block migrated to 55 per-path entries — killing the live `/.colima/` blanket-glob hazard that started the issue. Two of the four prior Important findings are genuinely fixed with falsified regression tests. What blocks SHIP is BR-37: the patch applied to `still_declared` did exactly what the finding said not to do (a third regex instead of one shared grammar), and the new regex is stateless where the extractor is stateful — so ariadne's own `46-removed-symbol-references.sh` now **silently passes on a violation it caught before this commit** (measured below). BR-38 was answered at 2 of its 5 sites with none of the enforcement it asked for, and the stale-path class it named is ~16 sites across 6 atlas pages. One new Important: the managed `.gitignore` block is now a *committed* generated artifact and nothing in CI checks it is current — I deleted 25 entries from it and the entire suite stayed green.

## 1. Strengths

- **`cmd/weave/internal/plan/gitignore.go:79` — the ownership rule is the right axis.** Deriving from the verb class (`Symlink`/`WriteFile`/`MergeSettings` → ignore; `Mkdir`/`Touch`/`Seed`/`SeedOnce` → track) makes the bootstrap core fall out of the rule instead of being an exclusion list, and the `default:` returning an error (not a panic) makes the "one case per verb" promise enforceable. `TestIgnoreEntriesNeverIgnoresScaffoldOrTouch` pins the one mistake that would have been catastrophic.
- **`cmd/weave/main.go:665-676` — the `TargetAll` pin is real and falsified.** `planActions`/`planActionsCore` split with no recursion (ARCH-DRY), and `TestIgnoreEntriesIdenticalAcrossTargets` has a non-vacuity guard (`hasAgents`) so it cannot pass on an empty list.
- **BR-36's fix survives falsification.** I reverted `[]string{walk.GeneratedRel}` → `nil` at `main.go:674` in a scratch copy at HEAD: `TestCompileIgnoresTheDynamicSkillGeneratedTree` goes red (`main_test.go:1103`). The expected value comes from the constant, not from the code under test — the rule BR-36 asked for.
- **`construct/scripts/test/merge-checks.test.sh:27-57` — `expect_red`/`expect_green` plus the self-test on a missing check script.** The guard-on-the-guard at :196 is the right shape; 14/14 green, and I confirmed the `46` run over this range is non-vacuous (`GeneratedRuntimeGitignoreEntries` is genuinely in the extracted removed set).
- **BR-39's retirement is complete at the source.** `construct/scripts/apply-gitignore-entries.sh` deleted, `base.manifest:171` row removed with the reasoning inline; `grep -rn apply-gitignore-entries` over the tree now returns only review artifacts. `ManagedLocations` (`prune.go:94`) does mark `construct/scripts` managed in a derivative, so the `PruneOrphans` claim in the manifest comment holds.
- **`workshop/targets/base-layer-mechanics.md:90-117`** states the invariant as a table with the "why not 'did weave create it'" counterexample, and names its test binding.

## 2. Critical findings

None raised as new — but see the BR-37 disposition: its severity has escalated in effect from a false-*positive* risk to a silent false-*negative* in ariadne's own merge gate. Treat it as blocking.

## 3. Important findings

**New — `.gitignore`'s managed block is committed generated output with no drift gate** (`scripts/merge-checks.d/30-weave-drift.sh:1`, `.gitignore:25-81`).

Measured: in a scratch copy of HEAD I deleted all 25 `/.agents/skills/*` lines from the managed block (55 → 30 entries). `go test ./cmd/weave/...`, `gitignore-surface.test.sh` and `merge-checks.test.sh` all stayed green. `30-weave-drift.sh` only tests dynamic-skill render determinism, and its own header argues the staleness job "evaporated" because weave's outputs are gitignored — M3 makes one output *committed* again, reviving exactly that staleness. Consequence: a manifest row retired without a `make weave && git commit` leaves a stale ignore line in the committed tree forever, which is the append-only hazard the issue exists to remove, now reintroduced one level up. The scope for such a gate is already written down — it is `IgnoreEntries`' own track case.

**Not-addressed — BR-37** (`scripts/merge-checks.d/46-removed-symbol-references.sh:106-109`). Detail in the disposition block; the short version is that the second `git grep -qE "^[[:space:]]+$1..."` is not stateful, so any indented identifier anywhere in HEAD — a struct field, a local — satisfies "still declared".

**Not-addressed — BR-38** (`cmd/weave/internal/plan/gitignore.go:14-16,38`, `atlas/workflow/base-layer.md:3`, `cmd/weave/internal/plan/action.go:33`).

## 4. Minor findings

- **New** — retiring `base.manifest:171` in M3 (rather than in M4's sweep, as BR-39 recommended) orphans a *tracked* `120000` symlink in 12 repos — including `brain`, `brain-family`, `brain-private` on the auto-commit rhythm. Their next `make weave` prunes it and stages a deletion no checklist owns.
- BR-40, BR-41, BR-42, BR-43 all remain open — see dispositions.
- `gitignore_test.go:104` `TestGeneratedRuntimeGitignoreCoversConstructGenerated` is still tautological (`sampleEntries` passes `walk.GeneratedRel` in, the test asserts it comes out). Harmless now that `main_test.go` carries the real pin, but the name no longer matches anything.
- `IgnoreEntries`' track case lists `EnsureGitignore`, which can never appear in its input (it is appended after). Dead but defensible as future-proofing.
- Environment note, not a finding: `30-weave-drift.sh` fails in my session on `mktemp: Operation not permitted` under `/var/folders/...`. That is this review sandbox, not the diff.

## 5. Test coverage notes

`go test ./cmd/weave/...` green; `gitignore-surface.test.sh` PASS; `merge-checks.test.sh` 14/14; `portable-makefile.test.sh` PASS via `50-base-layer-tests.sh`. `bash scripts/run-merge-checks.sh d0a92cb 384190b` runs 40/45/46/50 green (30 fails on the sandbox artifact above). Registration works — exec bits are `100755` in the index, so `run-merge-checks.sh:37`'s `-x` filter picks it up.

The gap the suite still cannot catch is the one above: every test builds its own scratch fixture, so ariadne's *own* committed artifacts are unasserted. Secondary: `50-base-layer-tests.sh:23,28`'s skip guards print `✓` and `exit 0`. Low practical risk (`cmd/weave` always exists here; `merge-check.yml:52` sets up Go), but a green checkmark on a skip is the same shape BR-36 named.

## 6. Architectural notes

- **ARCH-DRY** — flag, on `46-removed-symbol-references.sh:106`. `extract_removed` (:54) and `still_declared` (:106) are two implementations of one grammar and now disagree in a second way. One awk over `git show $HEAD:<file>` is the consolidation.
- **ARCH-PURE** — pass. `IgnoreEntries` is a total function over `[]Action`, tested with no fakes; IO stays in `applyEnsureGitignore`.
- **ARCH-PURPOSE** — flag. Shadow-sweep of the ignore-model's consumers: derivation ✓, `main.planActions` ✓, the block ✓, `apply-gitignore-entries.sh` retired ✓, `atlas/workflow/weave.md` ✓, the target ✓, `base-layer.md`'s new section ✓ — but `gitignore.go`'s own file header still restates the retired fixed list (naming `.colima/` and `vm-log.sh`, the two entries this milestone *removed*), README still doesn't say what determines the block, and BR-38 was answered at the instance, not the class.
- **ARCH-MOCK** — pass. `gitignore-surface.test.sh` runs the real `git` binary for the one external behaviour the invariant rests on, and `50-base-layer-tests.sh` gives it a per-PR cadence. Correct call not to fake `git check-ignore`'s index-awareness.
- **ARCH-CONSTRAINTS** — pass on the mechanism, flag on recording. Measured: `weave compile --dry-run` 0.00s ×3 warm, well inside the <5ms-added budget; block 55 entries against a predicted ~55. `50-base-layer-tests.sh` adds ~10.7s to every PR unconditionally. None of it is in `## Log`.
- **ARCH-SECURE** — flag (BR-41). `IgnoreEntries` emits filesystem-discovered skill directory names verbatim into git's glob grammar with no escaping seam.
- **ARCH-ORDER** — pass. Nothing new carries state between events; `mergeManagedBlock`'s marker parse fails closed on unterminated/duplicated/inverted pairs with the remedy in the message.
- **ARCH-FUNERAL** — flag. `legacyBlanketEntries:145-159` names its own end properly and covers all three blanket entries the old 9-item list needed (I checked the migration against each). The counterexample is the orphaned tracked symlink in 12 repos with no collection step.

## 7. Plan revision recommendations

1. **Chunk 4 (M4)** — add an explicit step: `construct/scripts/apply-gitignore-entries.sh` was retired in **M3**, so every repo's first post-M3 weave stages a deletion of a tracked `120000` path. Name the 12 repos and fold the deletion into the sweep commit; note the three brain repos where the auto-commit rhythm will pick it up unattended.
2. **Chunk 4 (M4)** — add the orphaned-comment cleanup to the per-repo checklist (BR-42): `parley.nvim/.gitignore`'s four-line "weave-generated runtime artifacts…" comment heads eight entries that are all now absorbed into the block.
3. **Chunk 3 / new task** — the plan never says who guarantees the *committed* block stays current. Add the drift gate (extend `30-weave-drift.sh`, whose name already claims it) scoped by `IgnoreEntries`' own track case, and state it in `## Core concepts` → `Lifecycle (ARCH-FUNERAL)`, which currently says supersession removes retired lines "on the next weave" without naming what forces that weave.
4. **`## Revisions`** — the entry added at `workshop/issues/000239-…md` under "M3 boundary review: 4 findings, all fixed" claims BR-38 was "two prose sites"; it named five, across two rules. Correct the record before it becomes the durable account.

```findings
dispose:
  - id: BR-36
    disposition: addressed
    note: |
      Falsified at HEAD: reverting main.go:674 to `nil` turns TestCompileIgnoresTheDynamicSkillGeneratedTree red; expected value comes from walk.GeneratedRel, not from planActions. Residual (not blocking): gitignore_test.go:104 is still tautological, and 50-base-layer-tests.sh:23,28 still print a green checkmark on a skip.
  - id: BR-37
    disposition: not-addressed
    note: |
      The finding said do NOT patch still_declared with a third regex; 46-removed-symbol-references.sh:106-109 adds exactly that, and it REGRESSED the check. The extractor's grouped branch is stateful (only between a group opener and its `)`); the new predicate `^[[:space:]]+NAME[[:space:]]*($|[=[:space:]])` is not, so any indented identifier satisfies it. Measured live: a repo removing top-level `func removedFunc` while `pkg/b.go` holds a struct field `removedFunc string`, with a stale comment `// removedFunc did the thing.` — the pre-patch check (450b678) exits 1 and names pkg/a.go:3; the patched check at HEAD prints "✓ no comment names a removed symbol" and exits 0. One grammar over `git show $HEAD:<file>` is still the fix, and merge-checks.test.sh's matrix should enumerate the grammar rather than list hand-picked shapes (ARCH-DRY).
  - id: BR-38
    disposition: not-addressed
    note: |
      2 of 5 sites, and neither enforcement. Fixed: atlas/workflow/weave.md:118, gitignore.go:14 ("fixed"→"DERIVED"). Still stale: gitignore.go:38 "until then the fixed list below stands"; gitignore.go:15-16 still enumerates ".colima/ VM tree, the vm-log.sh helper" — the two entries THIS milestone removed from the block; atlas/workflow/base-layer.md:3 still sends adopters to construct/setup.sh (absent — and the class is ~16 sites, incl. the dead link atlas/index.md:35 `[setup.sh](../construct/setup.sh)`, atlas/index.md:28, workflow/index.md:17, construct-adaptation.md:62, and all of setup-and-replication.md); action.go:33 still says WriteFile is lowered "from intent.Touch (empty Content)" when plan.go:110-114 lowers it to Touch with a comment saying explicitly NOT WriteFile. The measured 16-site spread is why the finding asked for the rule, not the sites.
  - id: BR-39
    disposition: addressed
    note: |
      Script deleted, base.manifest:171 retired with the reasoning inline; `grep -rn apply-gitignore-entries` over the tree now returns only review artifacts. ManagedLocations (prune.go:94) does mark construct/scripts managed in a derivative, so the PruneOrphans claim holds. The consumer enumeration the RULE asked for is still not written down as an artifact — see the new finding on the orphaned tracked slot.
  - id: BR-40
    disposition: not-addressed
    note: |
      README.md untouched in this window (last touched at 382c4b9, M2). :40-56 still documents the marker mechanism only; nothing names plan.IgnoreEntries or tells an adopter that a manifest row now changes their .gitignore.
  - id: BR-41
    disposition: not-addressed
    note: |
      IgnoreEntries (gitignore.go:88-96) still emits `"/" + filepath.Clean(dst)` verbatim; no escaping seam between the derivation and the .gitignore writer.
  - id: BR-42
    disposition: not-addressed
    note: |
      No M4 per-repo checklist entry was added; the plan's only orphan-comment step is Task 2.1's ariadne-local one at plan line 1066.
  - id: BR-43
    disposition: not-addressed
    note: |
      The Log records the 55-entry block but no timing. Re-measured this round: `weave compile --dry-run` 0.00s ×3 warm; 50-base-layer-tests.sh 10.7s standalone; full run-merge-checks.sh over the M3 range 15.0s.
findings:
  - id: new
    severity: Important
    family: verification-cannot-fail
    title: |
      The committed .gitignore managed block has no drift gate — deleting 25 entries leaves the whole suite green
    detail: |
      This is the 9th finding in family `verification-cannot-fail`. Do NOT fix
      this instance. The RULE: weave outputs that stay COMMITTED must be
      regenerated-and-diffed in CI; every other weave output is gitignored, so
      `30-weave-drift.sh`'s header reasons the staleness job "evaporated" — M3
      made one output committed again and the gate was not reinstated. Measured:
      in a scratch copy of 384190b I removed all 25 `/.agents/skills/*` lines
      from .gitignore (55 → 30 entries); `go test ./cmd/weave/...`,
      gitignore-surface.test.sh and merge-checks.test.sh all stayed green. The
      enumeration the gate needs is already written: IgnoreEntries' track case at
      gitignore.go:100 IS the set of weave targets that stay tracked, so the
      drift check derives its own scope from the same switch rather than naming
      .gitignore by hand. Consequence without it: a retired manifest row leaves a
      stale ignore line in the committed tree until someone remembers to weave
      and commit — the append-only hazard this issue exists to remove, one level up.
  - id: new
    severity: Minor
    family: verb-retirement-orphans-the-slot
    title: |
      Retiring base.manifest:171 in M3 orphans a tracked symlink in 12 repos with no step owning the deletion
    detail: |
      This is the 2nd finding in family `verb-retirement-orphans-the-slot`. The
      RULE rather than the instance: retiring a manifest row must name the
      TRACKED slot it orphans in every derivative and where that orphan is
      collected — the row's removal is the funeral for ariadne's copy only.
      Verified live: 12 sibling repos still carry
      `construct/scripts/apply-gitignore-entries.sh` at mode 120000 in the index
      (42shots, astro, brain, brain-family, brain-private, kaggle, kbench, metis,
      nous, pair, parli, robotics). PruneOrphans deletes the dangling link on
      each repo's next `make weave`, staging a deletion that M4's per-repo
      checklist does not mention; three of the twelve are brain repos on the
      auto-commit rhythm, so it lands unattended. BR-39 recommended doing the
      retirement inside M4's sweep for exactly this reason.
```
