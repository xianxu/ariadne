# Boundary Review — ariadne#239 (milestone M1)

| field | value |
|-------|-------|
| issue | 239 — Minimal committed base-layer surface |
| repo | ariadne |
| issue file | workshop/issues/000239-minimal-committed-base-layer-surface.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 7ffe2cc97ada126ba8b8cf355e27b3b0f9c56af6..221da22f6fbab28faac724f43ca3a19ae31f7ea6 |
| command | sdlc milestone-close --issue 239 --milestone M1 |
| reviewer | claude |
| timestamp | 2026-09-19T21:34:03-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

M1 delivers the `seed-once` verb end-to-end — intent → plan → seam → the seven enumerated switches — with real regression evidence: `portable-makefile.test.sh` PASSes including the new adopter cases, `go test ./cmd/...` is green except the pre-existing `#210` failure (`fleet_plan_test.go:14`, confirmed tracked in `workshop/issues/000210-*.md`), and `weave compile --dry-run` renders `seed-once Makefile -> …/construct/Makefile.seed`. The seed-source split and the `Makefile.workflow` default flip are both correct and the flip actually *removes* a pre-existing inconsistency (every `cmd/sdlc` flag already defaulted to `workshop/issues`; only `Makefile.workflow` said `issues`). What blocks a clean SHIP is one confirmed correctness defect: `golden.classifyAction`'s new `SeedOnce` case uses a different definition of "present" than `applySeedOnce` does, so the drift harness reports MATCH on exactly the fleet state M1 exists to converge — verified live: `weave golden ../nous` prints `MATCH seed-once Makefile — target present — repo-owned, weave would not touch it` while `nous/Makefile` is a symlink `weave compile` would materialize. Plus the README still documents the inverted contract.

## 1. Strengths

- **The presence-check ordering in `applySeedOnce` is right and documented as load-bearing** (`cmd/weave/internal/plan/apply.go:271-273`): `Lstat` before `ReadFile`, so a repo-owned file is never even compared to upstream. The `NOTE the ordering` comment makes the invariant re-checkable.
- **The negative completeness test is the one that can actually fail** (`cmd/weave/internal/golden/completeness_test.go:229-245`). `coverIntent` has no `default:`, so a positive-only test would have passed before the fix and proved nothing — the plan caught this and the code honours it.
- **The enumeration was re-run, not recalled.** I re-ran the grep independently: six action type-switches (`prune.go:68`, `apply.go:51`, `gather.go:95`, `golden.go:120`, `completeness.go:144`, `main.go:774`) plus `walk.isFileShape` and `coverIntent`/`verbName` — every one handles `SeedOnce`. No site was missed.
- **The conformance test's second half is the sharper assertion** (`construct/scripts/test/portable-makefile.test.sh:118-121`): a *later* local edit surviving a second weave is the worse half of defect 2, and it is now pinned against the real binary.
- **`construct/Makefile.seed` is exactly ariadne's root minus the two `WF_*` lines**, verified by `diff` — the split is a genuine ownership fix, not a file move, and ariadne's own `Makefile` is now byte-stable under `seed-once` (confirmed: `weave compile --dry-run` plans the action, the seam no-ops on a regular file).

## 2. Critical findings

**`cmd/weave/internal/golden/golden.go:201-224` — `classifyAction`'s `SeedOnce` case treats a symlink as presence; `applySeedOnce` does not.**

`observePath` (`gather.go:152-157`) sets `Exists: true` for *any* `Lstat` hit, including a symlink, and only then branches on `IsSymlink`. The new classifier switches on bare `dstO.Exists`, so a `Makefile` symlink classifies MATCH. `applySeedOnce` (`apply.go:271`) requires `fi.Mode()&os.ModeSymlink == 0` for presence, so it *removes and materializes* that same slot. The harness whose one job is predicting weave's effect now predicts the opposite.

Confirmed two ways. In a scratch copy of the pinned head:

```
SeedOnce symlink Dst -> MATCH      "target present — repo-owned, weave would not touch it (write-once)"
Seed     symlink Dst -> UNEXPECTED "content drift (live 0 bytes, upstream source 9 bytes)"
```

`Seed` gets it right only incidentally — `observePath` leaves `Content` empty for a symlink, so its content-compare fires. `SeedOnce` compares nothing, so nothing catches it. And against the live fleet (`../nous/Makefile` and `../metis/Makefile` are both `-> ../ariadne/Makefile` today):

```
$ weave golden ../nous
  MATCH      seed-once Makefile — target present — repo-owned, weave would not touch it (write-once)
```

Fix sketch — make the classifier use the *same* presence predicate as the seam:

```go
case dstO.Exists && !dstO.IsSymlink:
    return Divergence{Match, "seed-once", act.Dst, "target present — repo-owned, …"}
case dstO.IsSymlink:
    return Divergence{Unexpected, "seed-once", act.Dst,
        "target is still weave's prior symlink — weave would materialize the template (#225 convergence)"}
default:
    return Divergence{Unexpected, "seed-once", act.Dst, "…target is absent in live"}
```

Better still, hoist the predicate so it cannot drift again (ARCH-DRY / ARCH-ORDER): one exported helper — `plan.SeedOnceSlotIsRepoOwned(mode os.FileMode) bool` — called by both `applySeedOnce` and `classifyAction`. Today the same fact ("is this slot repo-owned?") is written twice and the two copies disagreed on first write. The regression test belongs in `golden` and must go red without the fix: symlinked `Dst` ⇒ `Unexpected`, dangling symlinked `Dst` ⇒ `Unexpected`, regular file ⇒ `Match`.

## 3. Important findings

**`README.md:17-21` — the adoption documentation still states the contract M1 inverted.**

> "A consumer's root `Makefile` is an upstream-owned **seed**: a real file that weave refreshes from ariadne. … Avoid editing the seeded root, since the next weave replaces its contents."

That is now false in both clauses, and it is the *front-door* statement of the thing this milestone changed. `atlas/workflow/base-layer.md` was updated with the new adoption path; README was not, and it is the file a repo adopting ariadne reads first. This is the Docs-update gate's README case. Fix: rewrite the paragraph around `seed-once` — repo-owned after first write, edit it freely, and a repo that already has a `Makefile` adopts with the single `include Makefile.workflow` line.

**`atlas/workflow/setup-and-replication.md:90-91` — the verb count was incremented without fixing the list it counts.**

`**Seven manifest actions:** symlink, tool, scaffold, touch, merge, seed, seed-once` — but `tool` was **retired** in #95 M5 (said so by `atlas/workflow/weave.md:27-29` and by `workshop/targets/base-layer-mechanics.md:97` on the same read), and `prose` + `skill` are missing. `intent.kindByVerb` (`manifest.go:14-25`) is the source of truth and holds eight live verbs. The pre-existing sentence was already wrong; bumping `Six`→`Seven` re-asserted it. For an issue whose thesis is that a hand-maintained restatement of the model is a deferred consumer, restating the verb set by hand is worth fixing while the page is open: `symlink`, `seed`, `seed-once`, `scaffold`, `touch`, `merge`, `prose`, `skill` — with `tool` and `copy` named as retired.

**`atlas/workflow/setup-and-replication.md:119-120` and `construct/base.manifest:34-35` — a not-yet-true invariant stated in the present tense.**

> "Both are **committed** in a derivative … Everything else weave emits is gitignored."

The derived ignore list lands in M3 and the sweep in M4; today a derivative still commits ~45 weave-created paths. `AGENTS.md` §Directory Structure makes `atlas/` the *current* state of the codebase, so an agent reading this during M2/M3 will believe an invariant that does not hold for the whole M1→M4 window. Cheap fix: mark it as the target state with the issue reference ("as of #239 M3/M4 …"), or defer the sentence to the M3 atlas pass where the plan already schedules it (Task 3.4).

## 4. Minor findings

- `cmd/weave/internal/plan/apply.go:290-295` duplicates `applySeed`'s exec-bit block verbatim (only the error prefix differs). The plan's DRY rationale claimed `applySeedOnce` "reuses `applySeed`'s mode-preservation and parent-creation helpers" — `ensureParent`/`removeDestinationSymlink` are reused, the chmod is copy-pasted. Extract `syncExecBit(fs, src, dst, verb string) error` (ARCH-DRY).
- `workshop/targets/base-layer-mechanics.md:90` still heads its section `### file-ops (symlink / seed / scaffold / touch)` — a stale enumeration in the one artifact type whose job is defending an invariant from drift. Add `seed-once`.
- Binary/manifest skew makes `Makefile` prunable in a symlink repo. With an older `weave` reading the new manifest, `ParseManifest` skips the unknown `seed-once` row, so no action targets `Makefile`; I confirmed `PrunePlan` then returns `[Makefile]` for a `Makefile -> /ws/ariadne/Makefile` candidate at a managed root. `make weave` depends on `weave-build`, which closes the window, so this only bites a direct `bin/weave compile` against a stale binary — but the loss is unrecoverable without `./bootstrap.sh`, since a repo with no `Makefile` has no `make weave` target. Worth a line in `## Log` at minimum.
- `gather.go:100-104` reads `Dst` content for `SeedOnce` (`observe(act.Dst, true)`), which `classifyAction` never consults. Harmless, and the Critical fix may well want it — noting only so it is a decision rather than a leftover.
- `applySeedOnce` swallows a `ReadFile(src)` error as a silent skip. It faithfully mirrors `applySeed`'s documented contract, so not a new defect — but the pair now swallows the same class twice with no operator-visible line anywhere.
- The four `SeedOnce` branch tests use `weavefs.OSFS{}` + `t.TempDir()` rather than the fake FS the plan's Core-concepts entry promised. This matches the file's own convention (29 existing `OSFS{}` uses) and gives real symlink semantics, so it is the better call — it just contradicts the plan text.

## 5. Test coverage notes

The seam is well covered: four `applySeedOnce` branches (regular file, absent, live symlink, dangling symlink) at `apply_test.go:731-815`, plus lowering (`plan_test.go:187-205`), parsing (`intent_test.go:185-195`), prune membership (`prune_test.go:432-439`), and both completeness directions. The bash conformance test exercises the real binary against a real scratch tree for both the materialize-a-symlink and the preserve-an-adopter halves.

Two gaps, one of which shipped the Critical:

- **`classifyAction`'s `SeedOnce` case has no test at all.** Twenty-four new lines of branching logic in the drift classifier, zero coverage — which is exactly why the symlink branch went out inverted. A `golden` test over the three slot states is the fix's regression guard.
- **The directory-at-`Dst` branch is untested.** The doc comment calls the no-op "deliberate" and contrasts it with `applySeed` erroring; an eight-line test would pin the claim.

One observation on the directory branch worth keeping: `applySeedOnce`'s no-op there is genuinely safer than `applySeed`'s error, and the comment says so — that is the kind of reasoning that should survive into a test rather than living only in prose.

## 6. Architectural notes

Per-marker, at-review lens:

- **ARCH-DRY — flag.** Two instances. The presence predicate is written twice and the copies disagree (the Critical). The exec-bit block is copy-pasted (Minor). Both want one owner.
- **ARCH-PURE — pass.** The planner records paths only; bytes and modes are read in the seam; `classifyAction` is pure over an `Observed` snapshot. `plan.go:121-124`'s comment states the split explicitly.
- **ARCH-PURPOSE — flag (Important).** Shadow-sweep over the consumers of the "root Makefile is upstream-owned" model: weave code ✓ (7 sites), `base.manifest` header ✓, three atlas pages ✓, the conformance test ✓, `bootstrap.sh:5` still accurate ✓ — **`README.md:17-21` missed**. The finding is the class, not just that file: the ownership claim was restated in five prose locations and the sweep reached four.
- **ARCH-MOCK — pass.** `weavefs.FS` is the seam, the unit tests drive it hermetically via `t.TempDir()`, and `portable-makefile.test.sh` is the live conformance check running the real `weave` binary against a real scratch tree — no live peer touched.
- **ARCH-CONSTRAINTS — pass.** One extra `Lstat` per compile on a developer-invoked path; the plan's `< 5 ms` envelope is not remotely threatened.
- **ARCH-SECURE — pass for this boundary.** M1 reads a template from a trusted peer layer; no credentials, no untrusted parse. The untrusted-input surface (`.gitignore` hand-edits, merge-conflicted markers) arrives in M2, where the plan already names fail-closed behaviour.
- **ARCH-ORDER — flag.** `applySeedOnce` models the slot as a four-state space (absent / symlink-live / symlink-dangling / regular-or-dir) and transitions correctly. `classifyAction` models the *same* state space independently and collapses symlink into "present". That is the entry's core failure shape: two components carrying different models of one state, with the difference unwritten. The fix is a shared predicate, not a second fixed switch.
- **ARCH-FUNERAL — pass.** `construct/Makefile.seed` is bounded and ariadne-owned; the per-repo `Makefile` hand-off *is* `seed-once`'s contract. Nothing in M1 grows without bound.

For upcoming work: the seven switches now enumerate the file-shape verbs with **no exhaustiveness enforcement** — `coverIntent` has no `default:` (the plan found this the hard way), and `formatActions` prints `unknown`. Task 3.1 already plans an erroring `default:` for `IgnoreEntries`; extend that to `coverIntent` and `formatActions` in the same pass so verb #9 fails loudly instead of silently reporting "covered".

## 7. Plan revision recommendations

Two `## Revisions` entries for `workshop/plans/000239-minimal-committed-base-layer-surface-plan.md`:

1. **Task 1.4, Step 4 — `golden.classifyAction`.** The plan instructs "classify it in the `Seed` class." The implementation deliberately does *not* (the issue's `## Revisions` explains why, and the reasoning is correct), but the plan still says otherwise. Record the departure — and record that the departure is where the symlink state stopped being handled, so the corrected instruction names all three slot states rather than delegating to Seed's behaviour.
2. **Core concepts → Integration points → `plan.applySeedOnce`.** The entry says "takes `weavefs.FS` so the **fake filesystem** drives every branch, including a dangling symlink." Delivered tests use `weavefs.OSFS{}` + `t.TempDir()`, matching the file's convention and giving real symlink semantics. Record the actual strategy so M2/M3 authors do not reintroduce a fake-FS test for a branch that needs real `Lstat` behaviour.

The M1 rows of the Core-concepts table itself hold up — `intent.SeedOnce`, `plan.SeedOnce`, `construct/Makefile.seed`, `golden.actionIndex` (modified), `Makefile.workflow` (modified), `plan.applySeedOnce` (new), `portable-makefile.test.sh` (modified) all exist at their stated paths with the stated status. No table/code contradiction.

```findings
findings:
  - id: new
    severity: Critical
    family: presence-predicate-written-twice
    title: |
      classifyAction's SeedOnce case calls a symlinked Dst "present", so the drift harness reports MATCH where weave would rewrite the file
    detail: |
      golden.go:201-224 switches on bare dstO.Exists, but observePath (gather.go:152-157)
      sets Exists=true for a symlink too; applySeedOnce (apply.go:271) requires
      !ModeSymlink for presence and materializes the slot. Verified live: `weave golden
      ../nous` prints "MATCH seed-once Makefile — target present — repo-owned, weave
      would not touch it" while ../nous/Makefile is a symlink to ../ariadne/Makefile.
      Seed escapes this only incidentally (its content-compare sees an empty string for a
      symlink). Fix by hoisting one shared repo-owned predicate used by both the seam and
      the classifier, with a golden test covering symlink / dangling symlink / regular
      file (ARCH-DRY, ARCH-ORDER).
  - id: new
    severity: Important
    family: ownership-claim-restated-in-prose
    title: |
      README.md:17-21 still documents the root Makefile as an upstream-owned seed that weave replaces
    detail: |
      "Avoid editing the seeded root, since the next weave replaces its contents" is the
      exact contract M1 inverts, in the file an adopting repo reads first. The shadow-sweep
      over this claim reached base.manifest and three atlas pages but missed README.
      Rewrite around seed-once: repo-owned after first write, and a repo that already has a
      Makefile adopts with one `include Makefile.workflow` line (ARCH-PURPOSE).
  - id: new
    severity: Important
    family: hand-maintained-restatement-of-model
    title: |
      setup-and-replication.md:90 bumps the verb count to seven while still listing retired `tool` and omitting `prose`/`skill`
    detail: |
      intent.kindByVerb (manifest.go:14-25) holds eight live verbs. `tool` was retired in
      -95 M5, as weave.md:27-29 and base-layer-mechanics.md:97 both say. The count was
      incremented without correcting the list it counts. List the eight live verbs and name
      tool/copy as retired (ARCH-DRY).
  - id: new
    severity: Important
    family: atlas-states-future-state-as-current
    title: |
      atlas and base.manifest assert "everything else weave emits is gitignored" — not true until M3/M4
    detail: |
      setup-and-replication.md:119-120 and base.manifest:34-35 state the committed-surface
      invariant in the present tense, but the derived ignore list lands in M3 and the sweep
      in M4. AGENTS.md makes atlas/ the current state of the codebase, so this misleads for
      the whole M1-to-M4 window. Mark it as the target state with the issue reference, or
      defer the sentence to the M3 atlas pass the plan already schedules.
  - id: new
    severity: Minor
    family: presence-predicate-written-twice
    title: |
      applySeedOnce copy-pastes applySeed's exec-bit block instead of sharing a helper
    detail: |
      apply.go:290-295 duplicates apply.go:239-244 verbatim but for the error prefix, while
      the plan's DRY rationale claimed reuse. Extract syncExecBit(fs, src, dst, verb).
  - id: new
    severity: Minor
    family: verb-retirement-orphans-the-slot
    title: |
      Under binary/manifest skew an older weave prunes a derivative's Makefile symlink
    detail: |
      An older weave skips the unknown seed-once row, so nothing targets Makefile;
      PrunePlan then returns [Makefile] for a Makefile -> ariadne symlink at a managed root
      (confirmed in a scratch copy). `make weave` depends on weave-build so the window is
      narrow, but the loss is unrecoverable without ./bootstrap.sh — a repo with no
      Makefile has no `make weave` target. Worth a note in the issue's Log.
  - id: new
    severity: Minor
    family: stale-verb-enumeration
    title: |
      base-layer-mechanics.md:90 still heads its section "file-ops (symlink / seed / scaffold / touch)"
    detail: |
      A stale enumeration in the artifact type whose job is defending an invariant from
      drift. Add seed-once.
```

---

## Re-review — 2026-09-19T21:50:34-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 239 — Minimal committed base-layer surface |
| repo | ariadne |
| issue file | workshop/issues/000239-minimal-committed-base-layer-surface.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 7ffe2cc97ada126ba8b8cf355e27b3b0f9c56af6..d964509a2f14c035bbb7e094ac7067adf6c2f595 |
| command | sdlc milestone-close --issue 239 --milestone M1 |
| reviewer | claude |
| timestamp | 2026-09-19T21:50:34-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

All seven round-1 findings are genuinely addressed, and I verified the load-bearing one rather than taking the commit message for it: reverting the BR-1 fix in a scratch copy turns `TestClassifySeedOnceSymlinkSlotIsNotPresence` red on exactly the two symlink rows (`class = MATCH, want UNEXPECTED`), and reverting the manifest verb to plain `seed` clobbers the adopter's Makefile (`differ: char 1, line 1`), so the conformance test's new half is load-bearing too. M1's declared scope is delivered end-to-end — the verb parses, lowers, applies, prunes, renders, and classifies; the seed source is split; the `Makefile.workflow` flip is safe (all 17 fleet repos already use `workshop/issues`, and the flip removes a pre-existing disagreement with `cmd/sdlc`'s own defaults). What keeps this off a clean SHIP is that the adoption recipe M1 now advertises in three places does not work: a bare `include Makefile.workflow` hard-fails `make` in an adopting repo until the first weave, and that repo has no `make weave` to run — verified live. Plus one undeclared, fleet-propagating change rode along in the BR-fix commit.

## 1. Strengths

- **The BR-1 class fix is the right shape, not just the right patch.** `plan.SeedOnceSlotIsRepoOwned` (`cmd/weave/internal/plan/apply.go:274`) is one exported predicate called by both the seam (`apply.go:302`) and the classifier (`golden.go:235`), and I swept the class: no other "who owns this slot" decision is written twice — the remaining `ModeSymlink` tests in `prune.go:217`, `apply.go:134`, `apply.go:365`, `gather.go:159` each answer a genuinely different question. Proportionate.
- **The regression test is the negative direction, deliberately.** `TestCheckCompletenessFlagsUncoveredSeedOnce` (`completeness_test.go:229`) asserts the *uncovered* case because `coverIntent` has no `default:` — a positive-only test would have passed before the fix. Same discipline in the four-case table test at `golden_test.go:427`, which covers live symlink / dangling symlink / regular file / absent.
- **The enumeration was re-run, not recalled, and it is complete.** I independently grepped every `Seed` site: all seven switches handle `SeedOnce`, and `weave verify-complete .` reports 0 unplanned paths while `weave golden .` renders `MATCH seed-once Makefile — target present — repo-owned`.
- **`construct/Makefile.seed` is exactly ariadne's root Makefile minus the two `WF_*` lines** (`diff -u` confirms) — the split really is an ownership change, not a rewrite that could smuggle in behavior.
- **`syncExecBit` (`apply.go:249`) is guarded where it matters.** The existing `applySeed` exec-bit subtests (`apply_test.go:250-310`) still pass, so the BR-5 extraction is covered for the caller whose seeded artifact (`bootstrap.sh`) actually needs the bit.

## 2. Critical findings

None.

## 3. Important findings

**I1 — the advertised one-line adoption path breaks `make` in the repo it is written for.** `README.md:24-27`, `atlas/workflow/base-layer.md:39-43`, `construct/base.manifest:120-123`.

M1's user-facing promise is "already have a Makefile? adopting ariadne needs one line: `include Makefile.workflow`." Before the first weave, `Makefile.workflow` is a weave-created symlink that does not exist yet, so a hard `include` aborts *every* target, including the repo's own:

```
Makefile:4: Makefile.workflow: No such file or directory
make: *** No rule to make target `Makefile.workflow'.  Stop.
```

The adopting repo therefore cannot run `make weave` to create the file the include needs. The same diff ships the correct form 20 lines away in `construct/Makefile.seed:17-18` — `WF_WORKFLOW := $(firstword $(wildcard Makefile.workflow ../ariadne/Makefile.workflow))` + `-include $(WF_WORKFLOW)` — which I verified resolves `make weave` against the sibling ariadne pre-weave. Fix: make all three doc sites quote the template's resolver form (or point at `construct/Makefile.seed` as the single executable snippet) rather than the `Makefile.workflow:1-2` header, and correct that header too, since it is the fourth copy and the one the docs cite as authority. Note this originated in BR-2's own prescribed wording, so it is a defect in the finding-fix loop, not just in the edit. (ARCH-PURPOSE — the adoption path *is* the point of `seed-once`.)

**I2 — an out-of-scope, fleet-propagating change rode along in the BR-fix commit.** `.claude/settings.ariadne.json:57`.

`d964509` ("#239 M1: fix the 7 boundary-review findings") adds `api.anthropic.com` to the sandbox network allowlist. That file is `symlink`ed and `merge`d into every derivative (`construct/base.manifest:87-88`), so this widens sandbox egress fleet-wide. It is mentioned nowhere — not in the commit body, the issue `## Log`, or `## Revisions` — and it is not a boundary-review finding. I confirmed it also puts the tree in weave-drift: `weave golden` on a scratch tree at HEAD reports `UNEXPECTED merge .claude/settings.json`, while the same tree with only this line reverted reports `MATCH`. Fix: either split it to its own commit with a stated reason, or drop it from this boundary. (ARCH-SECURE — a trust-boundary widening should never be invisible.)

**I3 — six in-code enumerations of the verb/action set were not updated.** `walk.go:112`, `plan.go:25`, `action.go:13`, `intent.go:8`, `gather.go:21`, `golden.go:364`.

> **This is the 2nd finding in family `stale-verb-enumeration`.** Earlier rounds fixed instances. Do NOT fix this instance — state the rule that covers all of them, and fix that.

**The rule:** *prose must name the source, not restate the set.* Every hand-written list of the verb/action set is a copy of `intent.kindByVerb` / the `Action` sum type that derives from nothing, so it goes stale on every verb change — which is defect 3 of this very issue, one level down. The BR-3 fix already demonstrated the correct move on `setup-and-replication.md`: it now says "the live set in `intent.kindByVerb` … the source of truth — check there, not here." Apply that rule to the residual sites rather than extending each list: delete the enumeration and name the source. Measured prevalence: round 1 fixed 5 doc sites; 6 remain, all in Go doc comments, including `walk.go:112` — `"(symlink/seed/scaffold/touch)"` sitting directly above the `isFileShape` line this diff edited to add `SeedOnce`. The implementor did update `Apply`'s behaviour list (`apply.go:32-35`) because the plan named that one file; the class was never enumerated. If you want it enforceable, a merge check that fails a file listing ≥3 verb names without naming `kindByVerb` is a ten-line grep.

## 4. Minor findings

- **M1 — the `Observed → FileMode` bridge is lossy, which re-opens the BR-1 class.** `golden.go:227-230` sets only `os.ModeSymlink` from `dstO.IsSymlink`, but `Observed` also carries `IsDir` (`golden.go:72`). The comment claims "widening the predicate updates both callers" — true today, false the moment the predicate reads any bit other than `ModeSymlink`. *This is the 3rd finding in family `presence-predicate-written-twice`.* The rule, not the instance: the harness must not reconstruct a partial mode at all — give `Observed` a `Mode os.FileMode` field populated by `observePath`, which already holds the real `fi.Mode()` (`gather.go:151`), and delete the reconstruction. That removes the divergence surface instead of patching one bit of it.
- **M2 — `gather.go:101-102` asserts a content comparison `classifyAction` does not do.** The comment says "classifyAction compares the live target against the upstream source bytes for both", but the `SeedOnce` case reads only `Exists`/`IsSymlink` — deliberately, since content drift is the intended end state. Both probes' content reads (`observe(act.Dst, true)`, `observeAbs`) are consumed by nothing. *This is the 2nd finding in family `hand-maintained-restatement-of-model`.* Rule: a comment restating a sibling function's behavior derives from nothing and drifts — state what *this* function needs (existence + symlink-ness) and let the classifier document itself. Worse than cosmetic here: it invites the next reader to "fix" the classifier into a content compare, re-asserting the two-owners claim `seed-once` retires.
- **M3 — the plan's Core concepts table does not list `plan.SeedOnceSlotIsRepoOwned`**, a newly exported cross-package API introduced by the BR-1 fix. Same family/rule as M2; see §7.
- **M4 — `weave golden` runs in no CI seam.** `grep` finds no invocation in `scripts/`, `.github/`, or `Makefile.workflow` — the drift harness BR-1 repaired is hand-run only. `30-weave-drift.sh` is a *determinism* check, unrelated. Worth a line in `## Log` so the harness's real coverage is not overestimated; M3's `50-base-layer-tests.sh` registration is the natural home.

## 5. Test coverage notes

Coverage of the shipped behavior is good: all four slot states are pinned at both the seam (`apply_test.go:727-815`, real FS in `t.TempDir()` — the right call, since real symlink semantics are the thing under test) and the classifier (`golden_test.go:427`), plus lowering, prune, and both completeness directions. `portable-makefile.test.sh` PASSes and I confirmed it is falsifiable. `go test ./cmd/...` is green except `TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory`, which I confirmed is the pre-existing `#210` hardcoded-path failure (issue file present).

Gaps, all low-impact:
- `applySeedOnce`'s **directory-in-slot** branch has no test, though the doc comment calls it deliberate and contrasts it with `applySeed` erroring.
- `applySeedOnce`'s **missing-src non-fatal skip** is untested (the `applySeed` equivalent is, via its own path).
- `classifyAction`'s `!srcO.Exists` branch (`golden.go:232`) is untested — the table test holds the source present in all four rows.
- `syncExecBit` has no test through the `seed-once` caller. Harmless while the only `seed-once` row is a non-executable Makefile; it becomes a gap the first time the verb is reused for an executable, which the plan's "future extensions" explicitly anticipates (`.envrc`, starter scripts).

## 6. Architectural notes for upcoming work

Per-marker: **ARCH-DRY** flag (I3, M1, M2 — the shared-predicate and `syncExecBit` extractions are right, the prose restatements are not). **ARCH-PURE** pass — the planner records paths, the seam reads bytes, the classifier is a pure function of (action, observation), and no test needs a mock to run a pure entity. **ARCH-PURPOSE** flag (I1 — the adoption path is the purpose, and the recipe doesn't run; I3 — the shadow-sweep over the verb-set consumers stopped at the doc sites the plan happened to name). **ARCH-MOCK** pass — the FS is the external surface, `weavefs.FS` is the seam, and `portable-makefile.test.sh` is a real-binary conformance check I verified both green and falsifiable. **ARCH-CONSTRAINTS** pass — O(1) added work on a developer-invoked path; envelope declared. **ARCH-SECURE** flag (I2). **ARCH-ORDER** pass — the slot's legal states are now an explicit named predicate rather than a boolean constellation, and all four are exercised at both ends; the one unenumerated interleaving is a `WriteFile` failure after `removeDestinationSymlink` leaving the slot empty with no rollback, which is identical to the pre-existing `applySeed` and so not a regression. **ARCH-FUNERAL** pass — `seed-once` *is* a hand-off, the plan's Lifecycle section names it, and this boundary creates no growing artifact family.

For M2–M4: M2's `mergeManagedBlock` will parse a `.gitignore` the process did not write — the plan's ARCH-SECURE section already commits to failing closed on an unmatched or duplicated marker pair, and that is the highest-value thing to hold it to at the next boundary. The `Observed`-carries-`Mode` fix in M1 above is cheap now and gets more expensive once M3's derivation adds a second mode-sensitive consumer. Separately, `cmd/sdlc` restates the `workshop/issues` default at ~24 sites via `envOr("WF_ISSUES_DIR", "workshop/issues")` while `validategate.go:48` alone derives it from `vocab.Issue().Discovery().Home` — pre-existing, outside this window, but it is the same single-source rule this issue is built on and a natural companion sweep.

## 7. Plan revision recommendations

`workshop/plans/000239-minimal-committed-base-layer-surface-plan.md` was not touched in this window and now under-describes the code. One `## Revisions` entry covering:

- **Core concepts / Pure entities** gains `plan.SeedOnceSlotIsRepoOwned` (`cmd/weave/internal/plan/apply.go` — new, exported, consumed by `golden`) and `plan.syncExecBit` (`apply.go` — new, unexported). Both were introduced by the M1 boundary-review fixes; the table currently claims an inventory the code exceeds.
- **Task 1.3's stated test strategy** says `applySeedOnce` "takes `weavefs.FS` so the fake filesystem drives every branch." The delivered tests use `weavefs.OSFS{}` over `t.TempDir()`, matching the package convention and exercising real symlink semantics. The delivered choice is better; the plan should say so rather than describe a fake that was not used.
- **Task 1.6 / the adoption prose**: record that the `include Makefile.workflow` one-liner prescribed by BR-2 does not work pre-weave (I1), and that the canonical snippet is `construct/Makefile.seed`'s resolver form.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Shared exported predicate plan.SeedOnceSlotIsRepoOwned (apply.go:274) called by seam (apply.go:302) and classifier (golden.go:235); reverting it in a scratch copy turns the 4-case table test red on both symlink rows.
  - id: BR-2
    disposition: addressed
    note: |
      README.md:17-32 rewritten around seed-once and repo ownership; the new adoption line it introduces is a separate new finding, not a re-raise.
  - id: BR-3
    disposition: addressed
    note: |
      setup-and-replication.md:90 now says eight, lists exactly kindByVerb's eight live verbs, names copy and tool as retired, and points at manifest.go as the source.
  - id: BR-4
    disposition: addressed
    note: |
      Target-state blockquote with the issue reference in setup-and-replication.md:119-123 and "not yet true" in base.manifest:36-37.
  - id: BR-5
    disposition: addressed
    note: |
      syncExecBit(fs, src, dst, verb) at apply.go:249 shared by both seams; the existing applySeed exec-bit subtests still pass, so the extraction is guarded.
  - id: BR-6
    disposition: addressed
    note: |
      Logged in the issue's Revisions with cause, the narrow weave-build window, and ./bootstrap.sh as recovery — which is what the finding asked for.
  - id: BR-7
    disposition: addressed
    note: |
      base-layer-mechanics.md:90 heading now reads symlink / seed / seed-once / scaffold / touch.
findings:
  - id: new
    severity: Important
    family: doc-recipe-never-executed
    title: |
      The advertised one-line adoption `include Makefile.workflow` hard-fails make in an adopting repo until the first weave, and there is no make weave to run
    detail: |
      README.md:26, atlas/workflow/base-layer.md:41 and construct/base.manifest:123 all
      present a bare `include Makefile.workflow` as the complete adoption contract. Before
      the first weave that file is an absent weave symlink, so make aborts every target:
      "Makefile.workflow: No such file or directory. Stop." — verified in a scratch two-layer
      tree. The repo cannot run `make weave` to fix it. The correct form ships 20 lines away
      in construct/Makefile.seed:17-18 (firstword wildcard + ../ariadne fallback, then
      `-include`), which I verified resolves `make weave` pre-weave. Fix all three sites plus
      Makefile.workflow:1-2, the header they cite as authority, by quoting the template's
      resolver form or pointing at construct/Makefile.seed as the one executable snippet.
      Originated in BR-2's own prescribed wording (ARCH-PURPOSE).
  - id: new
    severity: Important
    family: out-of-scope-change-rides-along
    title: |
      api.anthropic.com added to the fleet-propagated sandbox egress allowlist inside the boundary-review fix commit, mentioned nowhere
    detail: |
      .claude/settings.ariadne.json:57 gains api.anthropic.com in d964509 ("fix the 7
      boundary-review findings"). That file is symlinked AND merged into every derivative
      (base.manifest:87-88), so this widens sandbox egress fleet-wide. It is absent from the
      commit body, the issue Log, and Revisions, and is not one of the seven findings. It
      also leaves the tree in weave-drift: `weave golden` on a scratch HEAD tree reports
      UNEXPECTED merge .claude/settings.json, and the same tree with only this line reverted
      reports MATCH. Split it to its own commit with a reason, or drop it from this boundary
      (ARCH-SECURE).
  - id: new
    severity: Important
    family: stale-verb-enumeration
    title: |
      Six in-code enumerations of the verb/action set were not updated for seed-once, including the doc comment directly above the isFileShape line this diff edited
    detail: |
      This is the 2nd finding in family stale-verb-enumeration, so the deliverable is the
      RULE: prose must NAME the source, not restate the set. Every hand-written verb list is
      a copy of intent.kindByVerb / the Action sum type that derives from nothing, which is
      defect 3 of this issue one level down. The BR-3 fix already demonstrated the move on
      setup-and-replication.md ("the source of truth — check there, not here"); apply it to
      the residual sites by DELETING the enumeration rather than extending each list.
      Measured prevalence: round 1 fixed 5 doc sites, 6 remain — walk.go:112
      ("symlink/seed/scaffold/touch", two lines above the edited isFileShape case),
      plan.go:25, action.go:13, intent.go:8, gather.go:21, golden.go:364. The implementor
      updated Apply's behaviour list (apply.go:32-35) because the plan named that one file;
      the class was never enumerated. Enforceable as a grep merge check: fail a file listing
      3+ verb names without naming kindByVerb (ARCH-DRY).
  - id: new
    severity: Minor
    family: presence-predicate-written-twice
    title: |
      The Observed-to-FileMode bridge in classifyAction carries only ModeSymlink and silently drops IsDir, re-opening the BR-1 divergence for any future widening
    detail: |
      This is the 3rd finding in family presence-predicate-written-twice, so the deliverable
      is the RULE, not this bit: the harness must not reconstruct a partial mode at all.
      golden.go:227-230 builds dstMode from dstO.IsSymlink only, while Observed also carries
      IsDir (golden.go:72). The adjacent comment claims "widening the predicate updates both
      callers" — true only while the predicate reads exactly ModeSymlink. Class fix: add a
      Mode os.FileMode field to Observed, populated by observePath which already holds the
      real fi.Mode() (gather.go:151), and delete the reconstruction. No current bug; this
      removes the divergence surface instead of patching one bit of it (ARCH-DRY, ARCH-ORDER).
  - id: new
    severity: Minor
    family: hand-maintained-restatement-of-model
    title: |
      gather.go's SeedOnce comment asserts a content comparison classifyAction deliberately does not do, and the plan's Core concepts table omits the new exported predicate
    detail: |
      This is the 2nd finding in family hand-maintained-restatement-of-model, so the
      deliverable is the RULE: a comment or table that restates a sibling's model derives
      from nothing and drifts — state what THIS site needs, or make the restatement derived.
      Instances measured: gather.go:101-102 claims "classifyAction compares the live target
      against the upstream source bytes for both", but the SeedOnce case reads only
      Exists/IsSymlink and both probes' content reads are consumed by nothing — and the false
      claim invites a future reader to "fix" the classifier into a content compare, which
      would re-assert the two-owners claim seed-once retires. Second instance: the plan's
      Core concepts table lists no plan.SeedOnceSlotIsRepoOwned, a newly exported
      cross-package API the BR-1 fix introduced.
  - id: new
    severity: Minor
    family: verification-cannot-fail
    title: |
      weave golden runs in no CI seam, so the drift harness BR-1 repaired is hand-run only
    detail: |
      grep finds no `weave golden` or verify-complete invocation in scripts/, .github/ or
      Makefile.workflow; 30-weave-drift.sh is a dynamic-skill determinism check, unrelated.
      The harness whose correctness BR-1 restored therefore gates nothing today. Worth a line
      in the issue Log so its coverage is not overestimated; M3's planned
      scripts/merge-checks.d/50-base-layer-tests.sh is the natural registration point.
```
