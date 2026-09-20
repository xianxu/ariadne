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
