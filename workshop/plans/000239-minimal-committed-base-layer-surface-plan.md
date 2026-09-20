# Minimal Committed Base-Layer Surface — Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A derivative commits only its bootstrap core plus its own source; every path `make weave` re-derives is gitignored, from a list weave computes from the manifest walk and maintains inside a delimited block it owns.

**Architecture:** One insight collapses Pieces A–C into a single rule. A manifest verb already declares **who owns the bytes after weave runs**, and that is exactly the commit/ignore axis:

| Verb → Action | Who owns the bytes afterwards | Consequence |
|---|---|---|
| `symlink` → `Symlink`, `prose` → `WriteFile`, `merge` → `MergeSettings`, skill-dir links → `Symlink` | **weave** — recomputed byte-for-byte on every compile | **ignore** (reproducible from the substrate) |
| `scaffold` → `Mkdir`, `touch` → `Touch`, `seed` → `Seed`, `seed-once` → `SeedOnce` | **the repo** (or, for `seed`, upstream-before-substrate) — provisioned once, then never clobbered | **track** |

So the rule is one sentence: **weave ignores what it re-derives, and tracks what it merely provisions.** The bootstrap core falls out of it rather than being listed — `bootstrap.sh` and `merge-check.yml` are `seed` rows (they exist *because* they must work before any substrate), `Makefile` becomes `seed-once` (Piece A), and `construct/deps` is written by the `weave link` operator verb and is not a weave action at all. No hardcoded exclusion list, no second declaration channel (ARCH-DRY, and the `base-layer-mechanics` spine invariant that "no artifact enters the composition by any other channel").

The rule also repairs a defect the flat "ignore what weave creates" framing would have introduced: `scaffold workshop/issues` and `touch workshop/lessons.md` are weave-created, and ignoring them would untrack every issue file and the lessons log.

**Tech Stack:** Go (`cmd/weave`, `cmd/sdlc`), bash conformance tests (`construct/scripts/test/`), `base.manifest` as the declaration surface.

**Milestones (review boundaries).** Ordered so the riskier change always lands on machinery already proven:

- **M1 — `seed-once`:** the verb + the seed-source split. Independent of the gitignore work.
- **M2 — managed block:** `.gitignore` becomes a delimited weave-owned region, still carrying *today's* hardcoded 9 entries. Block machinery proven against a known-good list.
- **M3 — derive the list:** swap the hardcoded `[]string` for the manifest-walk derivation. Only now does the list grow to ~95 per-path entries, on machinery that can already retire a line.
- **M4 — fleet untrack:** the irreversible sweep, behind a green CI on one derivative.

M2 **must** precede M3: appending ~95 derived entries through today's append-only `ensureGitignoreText` is exactly the "actively dangerous" case the issue names — a retired manifest row would leave a permanent stale ignore line in every repo.

---

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `intent.SeedOnce` | `cmd/weave/internal/intent/intent.go` | new |
| `plan.SeedOnce` | `cmd/weave/internal/plan/action.go` | new |
| `construct/Makefile.seed` | `construct/Makefile.seed` | new |
| `plan.IgnoreEntries` | `cmd/weave/internal/plan/gitignore.go` | new |
| `plan.mergeManagedBlock` | `cmd/weave/internal/plan/gitignore.go` | new |
| `plan.ensureGitignoreText` | `cmd/weave/internal/plan/gitignore.go` | deleted |
| `plan.GeneratedRuntimeGitignoreEntries` | `cmd/weave/internal/plan/gitignore.go` | deleted |
| `golden.CheckCompleteness` | `cmd/weave/internal/golden/completeness.go` | modified |

- **`intent.SeedOnce`** — the manifest verb `seed-once`: write when the target slot is absent (or holds a weave symlink), no-op forever after, whatever the content.
  - **Relationships:** 1:1 with `plan.SeedOnce`; sibling of `intent.Seed` in `kindByVerb` and in `isFileShape`'s destructive-op set.
  - **DRY rationale:** Splits the one `seed` verb that carries **two ownership classes** (`bootstrap.sh`/`merge-check.yml` are upstream-owned and must converge; `Makefile` is the repo's front door and must not) into two verbs, rather than special-casing a path inside `applySeed`. A path special-case would be the same "one verb, two owners" defect one level down.
  - **Future extensions:** Any artifact where upstream wants to hand a working default over to the repo permanently — a starter `Makefile.local`, a starter `.envrc`.

- **`plan.SeedOnce`** — the Action: `{Src, Dst}`, identical in shape to `plan.Seed`, different in convergence semantics.
  - **Relationships:** 1:1 with `intent.SeedOnce`. Joins `ProducedPathSet` (prune must never treat a seeded-once file as an orphan) and `IgnoreEntries`'s tracked class.
  - **DRY rationale:** Reuses `applySeed`'s mode-preservation and parent-creation helpers; only the presence guard differs.

- **`construct/Makefile.seed`** — the generic root-Makefile template, split out of ariadne's own root `Makefile`.
  - **DRY rationale:** This is the root cause of defect 2, not a side-effect of it. Today `seed Makefile` means "the source *is* ariadne's own front door", which is precisely why the template hardcodes `WF_ISSUES_DIR = workshop/issues`. One file cannot be both a generic template and one repo's policy. After the split ariadne owns its root `Makefile` like every other repo, and the template holds no repo's layout.
  - **Future extensions:** If a mid layer in a 3-deep chain ever needs its own root targets, this is the file that would gain a `Makefile.<layer>` sibling (the naming the issue's Spec considered and deferred).

- **`plan.IgnoreEntries`** — `(actions []Action, generatedRoots []string) []string`: the derivation. Maps each action to a repo-relative ignore entry **iff weave re-derives its bytes**, then dedupes and sorts.
  - **Relationships:** N:1 with the action list `planActions` already computes. Consumed by exactly one caller (`main.planActions`).
  - **DRY rationale:** Retires `GeneratedRuntimeGitignoreEntries` — a hand-maintained restatement of what the manifest already says, i.e. a deferred consumer of the model (ARCH-PURPOSE). After this, adding or retiring a manifest row changes every repo's `.gitignore` with no code edit.
  - **Future extensions:** A new manifest verb joins the ignore or the track class by adding one `case` to the switch, which is the whole decision.
  - **`generatedRoots`** is the one weave-generated tree that is *not* an Action: `construct/generated/`, materialized by the `.dynamic-skill` exec stage that runs before planning. It is passed from `walk.GeneratedRel`, the constant that already owns it — a derivation from the owner, not a second hand-list. It is the only parameter of its kind, and a second one would be a signal that the dynamic-skill stage should emit Actions instead.

- **`plan.mergeManagedBlock`** — `(current string, entries []string) (next string, changed bool, err error)`: the pure `.gitignore` transform. Replaces the delimited region wholesale, preserves everything outside it verbatim and in place, absorbs loose duplicates of block-owned entries, and appends a fresh block when none exists.
  - **Relationships:** 1:1 replacement for `ensureGitignoreText`; same pure-string shape, so that function's existing direct unit tests translate rather than disappear.
  - **DRY rationale:** One owner for "which of these lines are weave's". Today ownership is implicit in an append that can never be undone.
  - **Future extensions:** The marker pair is the natural anchor if weave ever needs a second managed region (e.g. a managed `.gitattributes`); the parse/splice would generalize on the marker text.

**Test surface.** Every entity above is pure and gets a colocated `_test.go` running without IO mocks: `intent_test.go`, `action`/`plan_test.go`, `gitignore_test.go`, `completeness` coverage in the golden package. Two bash conformance tests (below) exercise the real binary against a real tree.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `plan.applySeedOnce` | `cmd/weave/internal/plan/apply.go` | new | filesystem (`weavefs.FS`) |
| `plan.applyEnsureGitignore` | `cmd/weave/internal/plan/gitignore.go` | modified | filesystem (`weavefs.FS`) |
| `main.planActions` | `cmd/weave/main.go:655` | modified | the compile lowering |
| `portable-makefile.test.sh` | `construct/scripts/test/portable-makefile.test.sh` | modified | real `weave` over a real scratch tree |
| `gitignore-surface.test.sh` | `construct/scripts/test/gitignore-surface.test.sh` | new | real `weave` + real `git` over a real scratch tree |

- **`plan.applySeedOnce`** — presence-guarded write. A **regular file in the slot is sacrosanct** (no read of `Src`, no compare, no write). A **symlink** in the slot is *not* presence: it is weave's own prior `symlink Makefile` lowering, and materializing it is the #225 convergence the fleet still needs (`pair/Makefile` is a tracked symlink to `../ariadne/Makefile` today). Absent slot → write with `applySeed`'s mode preservation.
  - **Injected into:** `plan.Apply`'s type switch; takes `weavefs.FS` so the fake filesystem drives every branch.

- **`plan.applyEnsureGitignore`** — unchanged responsibility (read, transform, write only on change); the transform it calls becomes `mergeManagedBlock`, and it now propagates that function's error for a malformed block.

- **`main.planActions`** — computes the ignore entries from the action list it just built. **Must derive from `plan.TargetAll`, never from the lean `--target`.** This hazard is created by M2: an append-only list could not lose an entry, but a wholesale-replaced block compiled under `--target claude` would drop every `.agents/skills/*` line and silently re-expose Codex's skill symlinks to `git status`. `runVerifyComplete` already demonstrates the second-plan pattern (`main.go:545` plans `TargetAll` alongside a lean target).

- **`gitignore-surface.test.sh`** — the stateful conformance test for the whole invariant, over a real two-layer scratch tree with a real `git init`: repo-owned entries and negations round-trip across two weaves; a retired manifest row loses its ignore line; a repo-owned file beside a weave symlink is never ignored; only the bootstrap core remains committed.
  - **Injected into:** `scripts/parallel-checks.sh` (the pre-merge check runner), beside the existing `portable-makefile.test.sh`.
  - **ARCH-MOCK:** weave's dependency surface here is the filesystem and `git`. The filesystem already has the `weavefs.FS` seam with a fake for unit tests; `git`'s behavior (what `check-ignore` and `ls-files -i -c` actually do with per-path patterns and negations) is the thing under test and **cannot** be faked without testing our own assumption — so this test runs the real `git` binary against a real scratch repo. That is the live conformance check for the one external behavior the whole issue rests on.

### Operating envelope (ARCH-CONSTRAINTS)

- **Interaction path:** developer-invoked `make weave` / `weave compile`, plus a CI invocation per PR. Not a keystroke or request path.
- **Latency budget:** the added work is `O(|actions| + |skills|)` string manipulation plus one `.gitignore` read/write already in the path — target **< 5 ms** added to a compile that runs in ~0.3–1 s today. Basis: measured action count below, all in-memory. Re-measure with `time ./bin/weave compile` before and after M3 (recorded in `## Log`).
- **Workload scale:** ariadne's manifest is ~45 rows; the fleet has 16 derivatives and 25 skills. The derived block is therefore **~95 lines** (~45 action-derived + 25 `.claude/skills/*` + 25 `.agents/skills/*`), against 9 today. Growth is linear in manifest rows + skills, both O(100) and operator-controlled.
- **Overload behavior:** none needed — no concurrency, no fan-out, no network. The block is bounded by the manifest, which is a committed file.
- **N/A:** memory, disk IO, co-tenancy — the transform holds one `.gitignore` (< 10 KB) in memory.

### Lifecycle (ARCH-FUNERAL)

- **The managed block** is the only durable artifact this work creates. Created by `weave compile`, last needed by the repo's `git`, removed by **supersession**: every compile replaces the region wholesale, so a retired manifest row's line disappears on the next weave. Bound: ~95 lines, linear in manifest rows + skills. This is precisely the removal path the current append-only mechanism lacks.
- **`construct/Makefile.seed`** creates nothing durable beyond the one-time `Makefile` per repo, which the repo then owns forever — that hand-off *is* `seed-once`'s contract.
- **The untracked paths** (M4) are removed from the index once; the files stay on disk and are regenerated by every weave.

### Trust boundaries (ARCH-SECURE)

`.gitignore` is an input weave did not produce: it is hand-edited by humans, written by older weave versions, and merged by git. The block parse must therefore fail closed on every shape it cannot interpret rather than guess at a splice point — a wrong guess deletes a repo's own ignore rules. Specifically: an open marker with no close, or more than one marker pair, is an **error** naming the file, not a best-effort repair. This is the direct lesson from `workshop/lessons.md` — *"a search that keys on content cannot see the content that describes it"* — three splices in #207/#218 cut at the wrong place because the marker was also discussable content. Markers are matched as **exact whole lines** only.

---

## Chunk 1: M1 — the `seed-once` verb

### Task 1.1: `intent.SeedOnce` kind

**Files:**
- Modify: `cmd/weave/internal/intent/intent.go`
- Modify: `cmd/weave/internal/intent/manifest.go:13-21`
- Test: `cmd/weave/internal/intent/intent_test.go`

- [ ] **Step 1: Write the failing test**

```go
func TestParseManifestSeedOnce(t *testing.T) {
	got, err := ParseManifest("seed-once construct/Makefile.seed Makefile\n")
	if err != nil {
		t.Fatalf("ParseManifest: %v", err)
	}
	want := []Intent{{Kind: SeedOnce, Visibility: Export, Source: "construct/Makefile.seed", Target: "Makefile"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}
```

- [ ] **Step 2: Run it and watch it fail**

Run: `go test ./cmd/weave/internal/intent/ -run SeedOnce -v`
Expected: FAIL — `undefined: SeedOnce`.

- [ ] **Step 3: Add the Kind and the verb**

In `intent.go`, append to the `const` block after `Seed` (append, never insert — `Kind` is an `iota` enum and inserting renumbers every later kind):

```go
	// SeedOnce — write-once real-file copy: created when the target slot is
	// absent, NEVER touched again whatever its content. The ownership sibling of
	// Seed: Seed's content is upstream-owned and converges every compile
	// (bootstrap.sh, merge-check.yml); a SeedOnce target is handed to the REPO on
	// first write and is the repo's from then on (the root Makefile — its own
	// front door, which upstream must not overwrite). One verb per ownership
	// class, rather than a path special-case inside the seam (#239).
	SeedOnce
```

In `manifest.go`, add to `kindByVerb`:

```go
	"seed-once": SeedOnce,
```

- [ ] **Step 4: Run the test**

Run: `go test ./cmd/weave/internal/intent/ -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add cmd/weave/internal/intent/
git commit -m "#239 M1: intent: add the seed-once verb"
```

### Task 1.2: `plan.SeedOnce` action + lowering

**Files:**
- Modify: `cmd/weave/internal/plan/action.go`
- Modify: `cmd/weave/internal/plan/plan.go:112-120`
- Test: `cmd/weave/internal/plan/plan_test.go`

- [ ] **Step 1: Write the failing test**

```go
func TestPlanLowersSeedOnce(t *testing.T) {
	layers := []layer.Layer{{Path: "/up", Intents: []intent.Intent{
		{Kind: intent.SeedOnce, Source: "construct/Makefile.seed", Target: "Makefile"},
	}}}
	actions, err := Plan(layers, nil)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	want := []Action{SeedOnce{Src: "/up/construct/Makefile.seed", Dst: "Makefile"}}
	if !reflect.DeepEqual(actions, want) {
		t.Fatalf("got %+v, want %+v", actions, want)
	}
}
```

- [ ] **Step 2: Run it and watch it fail**

Run: `go test ./cmd/weave/internal/plan/ -run SeedOnce -v`
Expected: FAIL — `undefined: SeedOnce`.

- [ ] **Step 3: Add the Action and the lowering**

In `action.go`, after the `Seed` type:

```go
// SeedOnce is a WRITE-ONCE real-file copy of an upstream Src into Dst — the
// ownership sibling of Seed. Seed TRACKS upstream (content-tracking, converges
// on every compile) because its content is upstream-owned; SeedOnce hands the
// slot to the REPO on first write and never touches it again, whatever it later
// contains. It exists for the root Makefile: a repo's own front door, which a
// greenfield repo should get for free but an adopting repo must keep (#239).
//
// "Present" means a REGULAR FILE. A SYMLINK in the slot is NOT presence — it is
// weave's own pre-#239 `symlink Makefile` lowering, and materializing it is the
// #225 convergence derivatives still need. See applySeedOnce.
type SeedOnce struct {
	Src string
	Dst string
}
```

and its marker beside the others:

```go
func (SeedOnce) isAction() {}
```

In `plan.go`, add a case beside `intent.Seed`:

```go
			case intent.SeedOnce:
				// Same path FACTS as Seed (ARCH-PURE: the planner records paths,
				// the seam reads bytes); the seam's presence guard is what differs.
				actions = append(actions, SeedOnce{Src: joinPath(l.Path, in.Source), Dst: in.Target})
```

- [ ] **Step 4: Run the tests**

Run: `go test ./cmd/weave/internal/plan/ -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add cmd/weave/internal/plan/action.go cmd/weave/internal/plan/plan.go cmd/weave/internal/plan/plan_test.go
git commit -m "#239 M1: plan: lower seed-once to a SeedOnce action"
```

### Task 1.3: `applySeedOnce` — the presence guard

**Files:**
- Modify: `cmd/weave/internal/plan/apply.go:44-70` (the type switch), plus the new function
- Test: `cmd/weave/internal/plan/apply_test.go`

- [ ] **Step 1: Write the failing tests — all three branches**

```go
// A repo-owned root Makefile is sacrosanct: seed-once must not touch it, whatever
// the upstream template says. This is the #239 defect-2 regression guard.
func TestApplySeedOncePreservesExistingRegularFile(t *testing.T) {
	root := t.TempDir()
	up := t.TempDir()
	mustWrite(t, filepath.Join(up, "Makefile.seed"), "UPSTREAM\n")
	mustWrite(t, filepath.Join(root, "Makefile"), "MY OWN BUILD SYSTEM\n")

	act := []Action{SeedOnce{Src: filepath.Join(up, "Makefile.seed"), Dst: "Makefile"}}
	if err := Apply(weavefs.OSFS{}, root, act); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if got := mustRead(t, filepath.Join(root, "Makefile")); got != "MY OWN BUILD SYSTEM\n" {
		t.Fatalf("seed-once clobbered a repo-owned file: %q", got)
	}
}

func TestApplySeedOnceCreatesWhenAbsent(t *testing.T) {
	root := t.TempDir()
	up := t.TempDir()
	mustWrite(t, filepath.Join(up, "Makefile.seed"), "UPSTREAM\n")

	act := []Action{SeedOnce{Src: filepath.Join(up, "Makefile.seed"), Dst: "Makefile"}}
	if err := Apply(weavefs.OSFS{}, root, act); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if got := mustRead(t, filepath.Join(root, "Makefile")); got != "UPSTREAM\n" {
		t.Fatalf("got %q, want the upstream template", got)
	}
}

// A symlink is weave's OWN prior lowering, not repo content: materialize it
// (the #225 convergence pair/nous/metis still need), and never write THROUGH it.
func TestApplySeedOnceMaterializesSymlinkWithoutFollowingIt(t *testing.T) {
	root := t.TempDir()
	up := t.TempDir()
	src := filepath.Join(up, "Makefile.seed")
	mustWrite(t, src, "UPSTREAM\n")
	ancestor := filepath.Join(up, "Makefile")
	mustWrite(t, ancestor, "ANCESTOR OWN\n")
	if err := os.Symlink(ancestor, filepath.Join(root, "Makefile")); err != nil {
		t.Fatal(err)
	}

	act := []Action{SeedOnce{Src: src, Dst: "Makefile"}}
	if err := Apply(weavefs.OSFS{}, root, act); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if fi, _ := os.Lstat(filepath.Join(root, "Makefile")); fi.Mode()&os.ModeSymlink != 0 {
		t.Fatal("still a symlink — the #225 convergence did not happen")
	}
	if got := mustRead(t, filepath.Join(root, "Makefile")); got != "UPSTREAM\n" {
		t.Fatalf("got %q, want the upstream template", got)
	}
	if got := mustRead(t, ancestor); got != "ANCESTOR OWN\n" {
		t.Fatalf("wrote THROUGH the symlink into the ancestor: %q", got)
	}
}
```

Reuse the existing helpers in `apply_test.go` if `mustWrite`/`mustRead` already exist under other names; match the file's established style rather than adding parallel helpers.

- [ ] **Step 2: Run them and watch them fail**

Run: `go test ./cmd/weave/internal/plan/ -run SeedOnce -v`
Expected: FAIL — `apply: unknown action type plan.SeedOnce`.

- [ ] **Step 3: Implement the seam**

In `apply.go`'s type switch, beside `case Seed:`:

```go
		case SeedOnce:
			err = applySeedOnce(fs, act.Src, filepath.Join(repoRoot, act.Dst))
```

and the function, beside `applySeed`:

```go
// applySeedOnce is the WRITE-ONCE half of the seed pair (#239). Where applySeed
// converges on upstream every compile, this one writes the slot at most once and
// then hands it to the repo permanently:
//
//   - A REGULAR FILE in the slot → no-op, with NO read of src and no comparison.
//     The repo owns it, whatever it now contains. This is the whole point: a repo
//     adopting ariadne keeps its own root Makefile, and a repo that later edits
//     its root Makefile keeps that edit across every subsequent weave.
//   - A SYMLINK in the slot → NOT presence. It is weave's own pre-#239 `symlink
//     Makefile` lowering; removing it and materializing a real file is the #225
//     convergence (pair/Makefile is such a link today). removeDestinationSymlink
//     also guarantees we never write THROUGH it into the ancestor's own Makefile.
//   - Absent → write src's bytes, preserving its executable bits exactly as
//     applySeed does.
//
// A missing src is a non-fatal skip, matching applySeed: weave cannot read the
// template, so it leaves the slot alone rather than erroring the walk.
func applySeedOnce(fs weavefs.FS, src, dst string) error {
	if fi, err := fs.Lstat(dst); err == nil && fi.Mode()&os.ModeSymlink == 0 {
		return nil // repo-owned regular file — sacrosanct, never read src
	}
	data, err := fs.ReadFile(src)
	if err != nil {
		return nil // template missing/unreadable → non-fatal skip (applySeed's contract)
	}
	if err := removeDestinationSymlink(fs, dst); err != nil {
		return err
	}
	if err := ensureParent(fs, dst); err != nil {
		return err
	}
	if err := fs.WriteFile(dst, data); err != nil {
		return fmt.Errorf("apply seed-once: write %s: %w", dst, err)
	}
	if fi, serr := fs.Stat(src); serr == nil && fi.Mode().Perm()&0o111 != 0 {
		if err := fs.Chmod(dst, fi.Mode().Perm()); err != nil {
			return fmt.Errorf("apply seed-once: chmod %s: %w", dst, err)
		}
	}
	return nil
}
```

Note the `Lstat` ordering: the presence check runs **before** the `src` read, so a repo-owned file is never even compared against upstream.

- [ ] **Step 4: Run the tests**

Run: `go test ./cmd/weave/internal/plan/ -v`
Expected: PASS, all three.

- [ ] **Step 5: Commit**

```bash
git add cmd/weave/internal/plan/apply.go cmd/weave/internal/plan/apply_test.go
git commit -m "#239 M1: weave: seed-once never overwrites a repo-owned file"
```

### Task 1.4: teach the walk, prune and completeness about `seed-once`

Three existing switches enumerate the file-shape verbs. Each one that misses `SeedOnce` is a silent defect, and they do not fail loudly — this is the *"making a gate conditional means re-qualifying every sentence that asserts it"* lesson, so the enumeration comes from a grep, not from memory.

**Files:**
- Modify: `cmd/weave/internal/walk/walk.go:112-122` (`isFileShape`)
- Modify: `cmd/weave/internal/plan/prune.go` (`ProducedPathSet`)
- Modify: `cmd/weave/internal/golden/completeness.go` (`coverIntent`, `verbName`)
- Modify: `cmd/weave/internal/golden/golden.go` (the action classifier, if it switches on action type)

- [ ] **Step 1: Derive the enumeration, don't recall it**

Run and record the output in the issue's `## Log`:

```bash
grep -rn "intent\.Seed\b\|case Seed:\|plan\.Seed\b" --include="*.go" cmd/ | grep -v _test
```

Every hit is a switch that must also handle `SeedOnce`. Expected: `walk.isFileShape`, `plan.ProducedPathSet`, `golden.coverIntent`, `golden.verbName`, `golden.classify`.

- [ ] **Step 2: Write the failing tests**

```go
// prune must never treat a seeded-once file as an orphan.
func TestProducedPathSetIncludesSeedOnce(t *testing.T) {
	set := ProducedPathSet([]Action{SeedOnce{Src: "/up/construct/Makefile.seed", Dst: "Makefile"}})
	if !set["Makefile"] {
		t.Fatalf("SeedOnce target missing from the produced set: %v", set)
	}
}
```

```go
// A seed-once row must not report as under-produced.
func TestCheckCompletenessCoversSeedOnce(t *testing.T) {
	layers := []layer.Layer{{Path: "/up", Intents: []intent.Intent{
		{Kind: intent.SeedOnce, Source: "construct/Makefile.seed", Target: "Makefile"},
	}}}
	actions := []plan.Action{plan.SeedOnce{Src: "/up/construct/Makefile.seed", Dst: "Makefile"}}
	if got := CheckCompleteness(layers, actions); len(got) != 0 {
		t.Fatalf("seed-once reported under-produced: %+v", got)
	}
}
```

- [ ] **Step 3: Run them and watch them fail**

Run: `go test ./cmd/weave/... -run 'SeedOnce' -v`
Expected: FAIL on both.

- [ ] **Step 4: Add `SeedOnce` to each switch the grep found**

`walk.isFileShape` — add `intent.SeedOnce` to the destructive-op case. `ProducedPathSet` — add `case SeedOnce: set[filepath.Clean(act.Dst)] = true`. `golden.coverIntent` — cover a `seed-once` intent with a `plan.SeedOnce` of the same `Dst`. `golden.verbName` — return `"seed-once"`. `golden.classify` — classify it as the existing `Seed` class does.

- [ ] **Step 5: Run the full weave suite**

Run: `go test ./cmd/weave/...`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add cmd/weave/
git commit -m "#239 M1: teach walk/prune/completeness the seed-once verb"
```

### Task 1.5: split the seed source from ariadne's own Makefile

This is the ownership fix, not a file move: ariadne's root `Makefile` stops being every repo's template.

**Files:**
- Create: `construct/Makefile.seed`
- Modify: `Makefile` (ariadne's own — unchanged content, now repo-owned)
- Modify: `construct/base.manifest:108-114`
- Modify: `Makefile.workflow:33-34`

- [ ] **Step 1: Create the generic template**

`construct/Makefile.seed` — ariadne's current root `Makefile` **minus** the two `WF_*` lines, plus a header saying what it is:

```make
# Generic root Makefile, seeded ONCE into a consuming repo (`seed-once`, #239).
# After the first weave this file belongs to YOUR repo: edit it freely, weave
# will never overwrite it. Adopting ariadne in a repo that already has a
# Makefile needs no seeding at all — just add the include below.
#
# Per-repo policy goes HERE, above the include, where `?=` defaults can still
# see it — e.g. a repo whose issues live under workshop/:
#     WF_ISSUES_DIR  = workshop/issues
#     WF_HISTORY_DIR = workshop/history

# Canonical repo name from git remote (portable across worktrees and containers)
REPO_NAME := $(shell git remote get-url origin 2>/dev/null | sed 's|.*/||; s|\.git$$||')

# Public checkouts use product targets without the maintainer overlay. After
# bootstrap clones peers, the sibling fallback supplies the first weave.
.DEFAULT_GOAL := help
WF_WORKFLOW := $(firstword $(wildcard Makefile.workflow ../ariadne/Makefile.workflow))
-include $(WF_WORKFLOW)
-include Makefile.local

.PHONY: help
help: $(WF_HELP_TARGETS)
	@true
```

- [ ] **Step 2: Point the manifest at it**

In `construct/base.manifest`, replace `seed      Makefile` with:

```
# The root Makefile is the REPO'S OWN front door: seeded ONCE so a greenfield
# repo gets a working root for free, then never touched again (#239). A repo that
# already has a Makefile keeps it and adopts ariadne with one line —
# `include Makefile.workflow`, the contract Makefile.workflow:1-2 documents.
# The SOURCE is construct/Makefile.seed, NOT ariadne's own root Makefile: one
# file cannot be both a generic template and one repo's layout policy, which is
# exactly why the old template hardcoded WF_ISSUES_DIR = workshop/issues.
seed-once construct/Makefile.seed Makefile
```

Leave ariadne's own root `Makefile` exactly as it is — it is now ariadne's, and its `WF_ISSUES_DIR = workshop/issues` lines are ariadne's policy, correctly stated in ariadne's own file.

- [ ] **Step 3: Confirm `Makefile.workflow` already holds the generic defaults**

`Makefile.workflow:33-34` already reads `WF_ISSUES_DIR ?= issues` / `WF_HISTORY_DIR ?= history`. No change needed — verify with `sed -n '30,36p' Makefile.workflow`. The `?=` is now genuinely overridable: the repo owns the root Makefile, so a `WF_ISSUES_DIR = issues` above the include is no longer clobbered next weave.

- [ ] **Step 4: Check every derivative's root Makefile carries its own `WF_*`**

Each fleet repo nests issues under `workshop/`, and after this change the seeded template no longer supplies that. A derivative whose root `Makefile` is still a **symlink** gets the template materialized without the `WF_*` lines and would silently fall back to `issues/`.

```bash
for d in ../*/; do
  [ -f "$d/construct/deps" ] || continue
  printf '%-16s ' "$(basename "$d")"
  if [ -L "$d/Makefile" ]; then echo "SYMLINK — needs WF_* after materialization"
  elif grep -q WF_ISSUES_DIR "$d/Makefile" 2>/dev/null; then echo "ok (own WF_*)"
  else echo "REGULAR FILE, no WF_* — check"; fi
done
```

Record the table in the issue's `## Log`. Every repo reported `SYMLINK` or `REGULAR FILE, no WF_*` gets its two `WF_*` lines added to its own root Makefile as part of M4's sweep (a one-line-per-repo edit, and that repo owns it forever after).

- [ ] **Step 5: Update the portable-makefile conformance test**

`construct/scripts/test/portable-makefile.test.sh` copies `$SOURCE/Makefile` as both the ancestor's and the leaf's. Three edits:

- line 12: the leaf's starting Makefile stays `$SOURCE/Makefile` (a repo-owned root) — unchanged.
- line 82: `cp "$SOURCE/Makefile" "$SCRATCH/ariadne/Makefile"` → also `mkdir -p "$SCRATCH/ariadne/construct" && cp "$SOURCE/construct/Makefile.seed" "$SCRATCH/ariadne/construct/Makefile.seed"`.
- line 86: the `awk` manifest filter selects on `$2`; a `seed-once` row's `$2` is now `construct/Makefile.seed`, so extend the pattern: `$2 == "construct/Makefile.seed"`.
- lines 96-97: `cmp "$SCRATCH/ancestor-before" "$SCRATCH/ariadne/Makefile"` still holds (the ancestor's own Makefile is untouched); `cmp "$SCRATCH/ariadne/Makefile" "$SCRATCH/leaf/Makefile"` must become `cmp "$SCRATCH/ariadne/construct/Makefile.seed" "$SCRATCH/leaf/Makefile"` — the leaf now materializes the *template*, not the ancestor's own root.

- [ ] **Step 6: Add the defect-2 regression case to the same test**

Append, after the existing two-weave convergence block:

```bash
# #239: a repo-owned root Makefile survives weave byte-for-byte, forever.
# Pre-#239 `seed Makefile` was content-tracking and silently destroyed it on the
# first weave, and any later edit to it on every subsequent weave.
mkdir -p "$SCRATCH/adopter/construct"
printf 'substrate ../ariadne\n' > "$SCRATCH/adopter/construct/deps"
: > "$SCRATCH/adopter/construct/base.manifest"
printf 'MY OWN BUILD SYSTEM\ninclude Makefile.workflow\n' > "$SCRATCH/adopter/Makefile"
cp "$SCRATCH/adopter/Makefile" "$SCRATCH/adopter-before"
(cd "$SCRATCH/adopter" && "$SCRATCH/real-weave" compile)
cmp "$SCRATCH/adopter-before" "$SCRATCH/adopter/Makefile"
printf 'AND A LATER LOCAL EDIT\n' >> "$SCRATCH/adopter/Makefile"
cp "$SCRATCH/adopter/Makefile" "$SCRATCH/adopter-edited"
(cd "$SCRATCH/adopter" && "$SCRATCH/real-weave" compile)
cmp "$SCRATCH/adopter-edited" "$SCRATCH/adopter/Makefile"
```

The second half is the part the issue calls out as the worse defect: not just the one-time adoption loss, but the edit lost on *every* weave.

- [ ] **Step 7: Run the conformance test**

Run: `bash construct/scripts/test/portable-makefile.test.sh`
Expected: `PASS portable Make product/overlay/bootstrap ordering and real weave convergence`.

- [ ] **Step 8: Weave ariadne itself and confirm no churn**

Run: `make weave && git status --short`
Expected: ariadne's root `Makefile` unchanged (seed-once no-ops on a regular file); no new dirty paths.

- [ ] **Step 9: Commit**

```bash
git add construct/Makefile.seed construct/base.manifest construct/scripts/test/portable-makefile.test.sh
git commit -m "#239 M1: split the seed template from ariadne's own root Makefile

The root Makefile carried two owners: a generic template for every derivative
AND ariadne's own layout policy (WF_ISSUES_DIR = workshop/issues), which is why
the 'generic' template hardcoded one repo's directory names. Splitting the
source makes each repo's root Makefile its own."
```

- [ ] **Step 10: Close the milestone**

```bash
sdlc milestone-close --issue 239 --milestone M1
```

Fix Critical/Important findings before crossing; log the verdict in `## Log`.

---

## Chunk 2: M2 — `.gitignore` as a weave-managed block

Entries stay the existing hardcoded nine. Only the *mechanism* changes, so a regression here is visible against a known-good list.

### Task 2.1: the pure `mergeManagedBlock` transform

**Files:**
- Modify: `cmd/weave/internal/plan/gitignore.go`
- Test: `cmd/weave/internal/plan/gitignore_test.go`

- [ ] **Step 1: Write the failing tests**

These five cases are the contract. The negation round-trip is the one the issue names explicitly (pair's `bin/*` + `!bin/*.sh`), and the malformed-block case is the ARCH-SECURE fail-closed guard.

```go
func TestManagedBlockAppendsWhenAbsent(t *testing.T) {
	got, changed, err := mergeManagedBlock("mine/\n", []string{"/CLAUDE.md"})
	if err != nil || !changed {
		t.Fatalf("err=%v changed=%v", err, changed)
	}
	want := "mine/\n" + managedBlockOpen + "\n/CLAUDE.md\n" + managedBlockClose + "\n"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

// Retiring a manifest row must REMOVE its ignore line — the append-only
// mechanism this replaces could not, and a stale line can silently untrack a
// repo-owned file that later takes that path.
func TestManagedBlockReplacesWholesaleSoRetiredEntriesDisappear(t *testing.T) {
	current := "mine/\n" + managedBlockOpen + "\n/CLAUDE.md\n/RETIRED.md\n" + managedBlockClose + "\ntail/\n"
	got, changed, err := mergeManagedBlock(current, []string{"/CLAUDE.md"})
	if err != nil || !changed {
		t.Fatalf("err=%v changed=%v", err, changed)
	}
	if strings.Contains(got, "/RETIRED.md") {
		t.Fatalf("retired entry survived: %q", got)
	}
	if !strings.Contains(got, "mine/") || !strings.Contains(got, "tail/") {
		t.Fatalf("repo-owned entries lost: %q", got)
	}
}

// pair's bin/* + !bin/*.sh block must survive verbatim, IN PLACE — the #64
// regression where a blanket ignore made tracked shell scripts look disposable.
func TestManagedBlockRoundTripsRepoOwnedNegationsInPlace(t *testing.T) {
	repo := "bin/*\n!bin/*.sh\n!bin/pair-dev\ncache/\n"
	first, _, err := mergeManagedBlock(repo, []string{"/CLAUDE.md"})
	if err != nil {
		t.Fatal(err)
	}
	second, changed, err := mergeManagedBlock(first, []string{"/CLAUDE.md"})
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Fatal("second weave rewrote an already-current .gitignore")
	}
	if second != first {
		t.Fatalf("not idempotent:\n%q\n%q", first, second)
	}
	if !strings.HasPrefix(second, repo) {
		t.Fatalf("repo entries moved or changed: %q", second)
	}
}

// Migration: a loose entry the block now owns is ABSORBED, not duplicated.
func TestManagedBlockAbsorbsLooseDuplicates(t *testing.T) {
	got, _, err := mergeManagedBlock("/CLAUDE.md\nmine/\n", []string{"/CLAUDE.md"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(got, "/CLAUDE.md") != 1 {
		t.Fatalf("duplicate entry: %q", got)
	}
	if !strings.Contains(got, "mine/") {
		t.Fatalf("repo entry lost: %q", got)
	}
}

// ARCH-SECURE: .gitignore is an input weave did not produce. A block it cannot
// parse is an error naming the file, never a guessed splice — a wrong guess
// deletes the repo's own rules.
func TestManagedBlockRefusesMalformedMarkers(t *testing.T) {
	for name, current := range map[string]string{
		"unterminated": managedBlockOpen + "\n/CLAUDE.md\n",
		"doubled":      managedBlockOpen + "\n" + managedBlockClose + "\n" + managedBlockOpen + "\n" + managedBlockClose + "\n",
		"close_first":  managedBlockClose + "\n" + managedBlockOpen + "\n",
	} {
		t.Run(name, func(t *testing.T) {
			if _, _, err := mergeManagedBlock(current, []string{"/CLAUDE.md"}); err == nil {
				t.Fatal("expected an error, got nil")
			}
		})
	}
}
```

- [ ] **Step 2: Run them and watch them fail**

Run: `go test ./cmd/weave/internal/plan/ -run ManagedBlock -v`
Expected: FAIL — `undefined: mergeManagedBlock`.

- [ ] **Step 3: Implement the transform**

Replace `ensureGitignoreText` in `gitignore.go`:

```go
// The managed region's delimiters. Matched as EXACT WHOLE LINES — never as a
// substring — so a .gitignore that merely *mentions* a marker (a comment
// explaining this mechanism) cannot be mistaken for the region itself. That is
// the workshop/lessons.md splice lesson: a search keyed on a token is wrong
// exactly where the token appears as content rather than as structure.
const (
	managedBlockOpen  = "# >>> weave-generated — managed by `make weave`, do not edit >>>"
	managedBlockClose = "# <<< weave-generated <<<"
)

// mergeManagedBlock is the pure transform behind applyEnsureGitignore: given a
// .gitignore's current content and the entries weave owns, it returns the next
// content, whether anything changed, and an error if the existing block cannot
// be parsed.
//
//   - INSIDE the markers: replaced WHOLESALE, so an entry weave no longer
//     produces (a retired manifest row) loses its ignore line. The append-only
//     predecessor could not do this; a stale line is dangerous once the list is
//     manifest-derived, because it can silently untrack a repo-owned file that
//     later takes that path.
//   - OUTSIDE the markers: preserved verbatim, IN PLACE. A repo's own entries and
//     negations (pair's `bin/*` + `!bin/*.sh`) must round-trip untouched.
//   - MIGRATION: a loose line outside the block that exactly matches an entry the
//     block now owns is absorbed, not duplicated.
//   - No block yet → append one at the end.
//
// Fails closed (ARCH-SECURE): an unterminated or duplicated marker pair is an
// error, never a guessed splice point — .gitignore is hand-edited input weave
// did not produce, and a wrong guess deletes a repo's own rules.
func mergeManagedBlock(current string, entries []string) (string, bool, error) {
	lines := strings.Split(current, "\n")
	open, close := -1, -1
	for i, line := range lines {
		switch line {
		case managedBlockOpen:
			if open != -1 {
				return "", false, fmt.Errorf(".gitignore: duplicate %q marker at line %d", managedBlockOpen, i+1)
			}
			open = i
		case managedBlockClose:
			if close != -1 {
				return "", false, fmt.Errorf(".gitignore: duplicate %q marker at line %d", managedBlockClose, i+1)
			}
			close = i
		}
	}
	switch {
	case open == -1 && close != -1:
		return "", false, fmt.Errorf(".gitignore: %q marker with no opening marker", managedBlockClose)
	case open != -1 && close == -1:
		return "", false, fmt.Errorf(".gitignore: %q marker with no closing marker", managedBlockOpen)
	case open != -1 && close < open:
		return "", false, fmt.Errorf(".gitignore: managed-block markers are inverted")
	}

	// The repo's own lines: everything outside the block, minus any loose
	// duplicate of an entry the block now owns (the migration absorb).
	owned := map[string]bool{}
	for _, e := range entries {
		owned[e] = true
	}
	var outside []string
	for i, line := range lines {
		if open != -1 && i >= open && i <= close {
			continue
		}
		if owned[line] {
			continue // absorbed into the block
		}
		outside = append(outside, line)
	}
	// Split.../Join round-trips a trailing newline as a final empty element;
	// drop trailing blanks so the block appends cleanly, then rebuild.
	for len(outside) > 0 && outside[len(outside)-1] == "" {
		outside = outside[:len(outside)-1]
	}

	block := append([]string{managedBlockOpen}, entries...)
	block = append(block, managedBlockClose)

	next := strings.Join(append(outside, block...), "\n") + "\n"
	return next, next != current, nil
}
```

The `outside`/`block` split keeps the repo's entries first and the block last. A repo whose block sat mid-file gets it moved to the end once, then it is stable — worth stating in the commit message.

- [ ] **Step 4: Run the tests**

Run: `go test ./cmd/weave/internal/plan/ -run ManagedBlock -v`
Expected: PASS, all five.

- [ ] **Step 5: Commit**

```bash
git add cmd/weave/internal/plan/gitignore.go cmd/weave/internal/plan/gitignore_test.go
git commit -m "#239 M2: weave owns a delimited .gitignore block

Append-only could never retire an entry. Wholesale replacement inside markers
can, and repo-owned entries outside them round-trip verbatim."
```

### Task 2.2: wire the seam and translate the old tests

**Files:**
- Modify: `cmd/weave/internal/plan/gitignore.go` (`applyEnsureGitignore`)
- Modify: `cmd/weave/internal/plan/gitignore_test.go` (the `ensureGitignoreText` tests)
- Modify: `cmd/weave/internal/plan/apply.go:35-39` (the doc comment)

- [ ] **Step 1: Propagate the error through the seam**

```go
	next, changed, err := mergeManagedBlock(current, entries)
	if err != nil {
		return fmt.Errorf("apply ensure-gitignore: %s: %w", gitignorePath, err)
	}
	if !changed {
		return nil
	}
```

- [ ] **Step 2: Translate, don't delete, the existing unit tests**

`TestEnsureGitignoreText*` and `TestApplyEnsureGitignore*` each still assert something true of the new mechanism (appends when absent, idempotent when current, preserves existing, dedupes a repeated input entry, adds a trailing newline). Rewrite each against `mergeManagedBlock`/`Apply` rather than dropping it — the behaviours are still the contract; only the layout changed.

`TestEnsureGitignoreTextDedupsRepeatedInputEntry` needs a decision: the new block emits `entries` verbatim, so a repeated input entry would appear twice. Dedupe inside `IgnoreEntries` (Task 3.1, which sorts and dedupes anyway) and keep this test pointed at that function.

- [ ] **Step 3: Run the full suite**

Run: `go test ./cmd/weave/...`
Expected: PASS.

- [ ] **Step 4: Weave ariadne and inspect the migration**

Run: `make weave && git diff .gitignore`
Expected: the nine loose entries absorbed into a block at the end of the file; `.goto`, `bin/`, `/couch`… and every other ariadne-owned line untouched and in place. Hand-remove the now-orphaned explanatory comment at `.gitignore:17-21` (it described `/AGENTS.md`, which has moved into the block) in the same commit.

- [ ] **Step 5: Weave twice and confirm no churn**

Run: `make weave && make weave && git status --short .gitignore`
Expected: empty after the first commit — the second weave writes nothing.

- [ ] **Step 6: Commit**

```bash
git add cmd/weave/ .gitignore
git commit -m "#239 M2: migrate .gitignore entries into the managed block"
```

- [ ] **Step 7: Close the milestone**

```bash
sdlc milestone-close --issue 239 --milestone M2
```

---

## Chunk 3: M3 — derive the entry list from the manifest walk

### Task 3.1: `IgnoreEntries` — the derivation

**Files:**
- Modify: `cmd/weave/internal/plan/gitignore.go`
- Test: `cmd/weave/internal/plan/gitignore_test.go`

- [ ] **Step 1: Write the failing tests**

The first two encode the *rule*; the next three encode the hazards the issue proved are real, not hypothetical.

```go
// The rule: weave ignores what it RE-DERIVES, and tracks what it merely
// PROVISIONS. Symlink/WriteFile/MergeSettings are recomputed byte-for-byte every
// compile; Mkdir/Touch/Seed/SeedOnce are created once and then owned elsewhere.
func TestIgnoreEntriesIgnoresOnlyRederivedPaths(t *testing.T) {
	got := IgnoreEntries([]Action{
		Symlink{Src: "/up/scripts/lib.sh", Dst: "scripts/lib.sh"},
		WriteFile{Path: "CLAUDE.md", Content: "x"},
		MergeSettings{Sources: []string{"/up/a.json"}, Target: ".claude/settings.json"},
		Mkdir{Path: "workshop/issues"},
		Touch{Path: "workshop/lessons.md"},
		Seed{Src: "/up/bootstrap.sh", Dst: "bootstrap.sh"},
		SeedOnce{Src: "/up/construct/Makefile.seed", Dst: "Makefile"},
	}, nil)
	want := []string{"/.claude/settings.json", "/CLAUDE.md", "/scripts/lib.sh"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

// Catastrophic if wrong: `scaffold workshop/issues` and `touch
// workshop/lessons.md` are weave-CREATED, and ignoring them would untrack every
// issue file and the lessons log. The flat "ignore what weave creates" framing
// gets this wrong; the ownership framing gets it right.
func TestIgnoreEntriesNeverIgnoresScaffoldOrTouch(t *testing.T) {
	got := IgnoreEntries([]Action{
		Mkdir{Path: "workshop/issues"},
		Mkdir{Path: "atlas"},
		Touch{Path: "workshop/lessons.md"},
	}, nil)
	if len(got) != 0 {
		t.Fatalf("scaffold/touch leaked into the ignore list: %v", got)
	}
}

// The bootstrap core falls OUT of the rule rather than being listed: both seed
// verbs mean "must exist before the substrate does", which means committed.
func TestIgnoreEntriesNeverIgnoresTheBootstrapCore(t *testing.T) {
	got := IgnoreEntries([]Action{
		Seed{Src: "/up/bootstrap.sh", Dst: "bootstrap.sh"},
		Seed{Src: "/up/.github/workflows/merge-check.yml", Dst: ".github/workflows/merge-check.yml"},
		SeedOnce{Src: "/up/construct/Makefile.seed", Dst: "Makefile"},
	}, nil)
	if len(got) != 0 {
		t.Fatalf("bootstrap core ignored: %v", got)
	}
}

// Per-path, never a directory glob. parley.nvim/scripts/merge-checks.d/
// 20-vocabulary.sh is a repo-owned check living beside the weave symlink
// 40-duplicate-issue-id.sh; a blanket `scripts/merge-checks.d/` would untrack it.
// This is the pair#64 pattern verbatim.
func TestIgnoreEntriesIsPerPathNotPerDirectory(t *testing.T) {
	got := IgnoreEntries([]Action{
		Mkdir{Path: "scripts/merge-checks.d"},
		Symlink{Src: "/up/scripts/merge-checks.d/40-duplicate-issue-id.sh", Dst: "scripts/merge-checks.d/40-duplicate-issue-id.sh"},
	}, nil)
	want := []string{"/scripts/merge-checks.d/40-duplicate-issue-id.sh"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

// A DIRECTORY symlink (symlink .tart/scripts) must get NO trailing slash: git's
// `foo/` pattern does not match a symlink named foo, so a trailing slash would
// silently fail to ignore it. Only a real generated DIRECTORY gets one.
func TestIgnoreEntriesTrailingSlashOnlyForGeneratedRoots(t *testing.T) {
	got := IgnoreEntries([]Action{
		Symlink{Src: "/up/.tart/scripts", Dst: ".tart/scripts"},
	}, []string{"construct/generated"})
	want := []string{"/.tart/scripts", "/construct/generated/"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}
```

- [ ] **Step 2: Run them and watch them fail**

Run: `go test ./cmd/weave/internal/plan/ -run IgnoreEntries -v`
Expected: FAIL — `undefined: IgnoreEntries`.

- [ ] **Step 3: Implement the derivation, delete the hardcoded list**

```go
// IgnoreEntries derives the paths weave's .gitignore block owns, from the
// ACTIONS weave planned — one source of truth with the manifest, automatically
// correct when a row is added or retired (ARCH-DRY, and the base-layer-mechanics
// spine invariant that no artifact enters the composition by another channel).
// It replaces a hardcoded []string, which was a hand-maintained restatement of
// the model — a deferred consumer, not a finished one (ARCH-PURPOSE).
//
// The rule is the manifest verb's OWNERSHIP class, because a verb already
// declares who owns the bytes after weave runs:
//
//	weave RE-DERIVES them every compile → IGNORE
//	  Symlink (symlink rows + the lowered skill-dir links), WriteFile (the
//	  composed per-harness entry files), MergeSettings (the settings cascade).
//	weave merely PROVISIONS the slot, then someone else owns it → TRACK
//	  Mkdir (scaffold: an empty container for the REPO's content — ignoring
//	  workshop/issues would untrack every issue file), Touch (create-if-missing,
//	  never clobbered — workshop/lessons.md accumulates real content), Seed and
//	  SeedOnce (both mean "must work BEFORE any substrate exists", which is
//	  exactly why they must be committed — the bootstrap core falls out of the
//	  rule instead of being listed).
//
// generatedRoots carries the one weave-generated tree that is NOT an Action:
// construct/generated/, materialized by the .dynamic-skill exec stage that runs
// before planning. It is passed from walk.GeneratedRel, the constant that already
// owns that path — a derivation from the owner, not a second hand-list.
//
// Entries are repo-root-anchored with a leading slash, deduped and sorted
// lexicographically, so a manifest REORDER produces no .gitignore churn.
//
// PER-PATH, never a directory glob: scripts/, construct/scripts/, .claude/ and
// scripts/merge-checks.d/ all mix weave-created and repo-owned files
// (parley.nvim/scripts/merge-checks.d/20-vocabulary.sh sits beside a weave
// symlink; every repo's scripts/ci-setup.sh sits among them). A blanket ignore
// there is the pair#64 regression, where a `bin/` glob made tracked shell scripts
// look disposable and a propagate-base sweep git-rm'd them.
//
// TRAILING SLASH only for generatedRoots. An action-derived entry gets none:
// git's `foo/` pattern does not match a SYMLINK named foo, so `symlink
// .tart/scripts` would silently go un-ignored with one. Pure.
func IgnoreEntries(actions []Action, generatedRoots []string) []string {
	seen := map[string]bool{}
	var out []string
	add := func(entry string) {
		if !seen[entry] {
			seen[entry] = true
			out = append(out, entry)
		}
	}
	for _, a := range actions {
		switch act := a.(type) {
		case Symlink:
			add("/" + filepath.Clean(act.Dst))
		case WriteFile:
			add("/" + filepath.Clean(act.Path))
		case MergeSettings:
			add("/" + filepath.Clean(act.Target))
		case Mkdir, Touch, Seed, SeedOnce, EnsureGitignore:
			// Provisioned once, then owned by the repo (or, for the seeds, the
			// pre-substrate bootstrap core). Tracked — never ignored.
		}
	}
	for _, root := range generatedRoots {
		add("/" + filepath.Clean(root) + "/")
	}
	sort.Strings(out)
	return out
}
```

Delete `GeneratedRuntimeGitignoreEntries` and rewrite the file header comment: the "What is NOT ignored — the pre-weave BOOTSTRAP scaffolding" paragraph is the stale justification #225 obsoleted (defect 1) and must be replaced by the ownership rule above, not merely trimmed.

- [ ] **Step 4: Run the tests**

Run: `go test ./cmd/weave/internal/plan/ -v`
Expected: PASS. `cmd/weave/main_test.go:210` and `gitignore_test.go:49,65,71,100,108,119,148` reference the deleted variable — update each to call `IgnoreEntries` over a planned action set instead. `gitignore_test.go:65-71` asserts `/construct/generated/` is present; keep that assertion, now against `IgnoreEntries(..., []string{walk.GeneratedRel})`.

- [ ] **Step 5: Commit**

```bash
git add cmd/weave/internal/plan/
git commit -m "#239 M3: derive the gitignore entries from the manifest walk

A hardcoded list of what weave generates is a second declaration channel, which
the base-layer-mechanics spine forbids. The verb already declares ownership:
weave ignores what it re-derives and tracks what it merely provisions."
```

### Task 3.2: wire `planActions` — and pin it to `TargetAll`

**Files:**
- Modify: `cmd/weave/main.go:655-690`
- Test: `cmd/weave/main_test.go`

- [ ] **Step 1: Write the failing test**

```go
// A lean --target must NOT shrink the block. Append-only could not lose an entry;
// wholesale replacement can, so a `weave compile --target claude` would drop
// every /.agents/skills/* line and silently re-expose Codex's symlinks to
// `git status`. The ignore list is a property of the REPO, not of the face being
// compiled — so it is always derived from TargetAll.
func TestIgnoreEntriesIdenticalAcrossTargets(t *testing.T) {
	fs := weavefs.OSFS{}
	layers := mustWalkFixture(t, fs)
	union := ignoreEntriesFor(t, fs, layers, plan.TargetAll)
	lean := ignoreEntriesFor(t, fs, layers, plan.TargetClaude)
	if !reflect.DeepEqual(union, lean) {
		t.Fatalf("lean target changed the ignore list:\n union=%v\n  lean=%v", union, lean)
	}
	var hasAgents bool
	for _, e := range lean {
		if strings.HasPrefix(e, "/.agents/skills/") {
			hasAgents = true
		}
	}
	if !hasAgents {
		t.Fatal("lean target dropped the .agents/skills entries")
	}
}
```

`ignoreEntriesFor` is a test helper that runs `planActions` for the target and pulls the `EnsureGitignore` action's `Entries` out of the result. Reuse `main_test.go`'s existing fixture-walk helper if one is present.

- [ ] **Step 2: Run it and watch it fail**

Run: `go test ./cmd/weave/ -run IgnoreEntriesIdenticalAcrossTargets -v`
Expected: FAIL — the lean plan lacks the `.agents/skills` symlinks.

- [ ] **Step 3: Wire it**

In `planActions`, replace the hardcoded append:

```go
	// The ignore list is a property of the REPO, not of the face being compiled.
	// It must therefore be derived from the UNION plan even on a lean --target:
	// the managed block is replaced WHOLESALE (#239 M2), so deriving it from a
	// lean action set would DELETE the other harnesses' entries and silently
	// re-expose their symlinks to `git status`. Append-only could not lose an
	// entry; wholesale replacement can, so this is a hazard M2 created and M3
	// must close. runVerifyComplete (main.go:545) uses the same second-plan shape.
	ignoreActions := actions
	if target != plan.TargetAll {
		if ignoreActions, err = planActionsForIgnore(fs, layers); err != nil {
			return nil, fmt.Errorf("plan ignore entries: %w", err)
		}
	}
	actions = append(actions, plan.EnsureGitignore{
		Entries: plan.IgnoreEntries(ignoreActions, []string{walk.GeneratedRel}),
	})
```

where `planActionsForIgnore` is the Union plan without its own `EnsureGitignore` (extract the body of `planActions` above the append into a helper both call, so there is no recursion and no duplicated lowering — ARCH-DRY).

- [ ] **Step 4: Run the tests**

Run: `go test ./cmd/weave/...`
Expected: PASS.

- [ ] **Step 5: Measure the envelope**

```bash
go build -o bin/weave ./cmd/weave
time ./bin/weave compile
./bin/weave compile --dry-run | grep -c .
sed -n "/$(printf '%s' '# >>> weave-generated')/,/# <<< weave-generated/p" .gitignore | wc -l
```

Record compile wall time and block line count in the issue's `## Log` against the plan's budget (< 5 ms added; ~95 lines). A block materially larger than ~95 lines means the derivation is picking up something it should not — investigate before proceeding.

- [ ] **Step 6: Weave ariadne and read the diff carefully**

Run: `make weave && git diff .gitignore`

Verify specifically:
- `/.colima/` is **gone**, replaced by nothing in ariadne. ariadne OWNS `.colima/` (6 tracked real files), and the blanket entry silently ignored every *new* file there — `git check-ignore -v .colima/NEWFILE` reports `.gitignore:28:/.colima/`. This is the pair#64 hazard live in ariadne's own tree, and per-path derivation fixes it: the `.colima/*` rows are self-referential on ariadne's self-walk, so `walk.loadLayer` drops them and they produce no actions. Confirm with `git check-ignore -v .colima/NEWFILE` after the weave — expected: no match, exit 1.
- `/.claude/skills/` is replaced by one `/.claude/skills/<name>` line per skill, and likewise `/.agents/skills/<name>`.
- `/AGENTS.md`, `/CLAUDE.md`, `/GEMINI.md`, `/.claude/settings.json`, `/construct/generated/` all survive.
- `git status --short` is clean apart from `.gitignore` itself.

- [ ] **Step 7: Commit**

```bash
git add cmd/weave/ .gitignore
git commit -m "#239 M3: weave derives its ignore block from the Union plan

Deriving from a lean --target would delete the other harnesses' entries, since
the block is replaced wholesale. Also drops the blanket /.colima/ entry, which
silently ignored new files in a directory ariadne itself owns."
```

### Task 3.3: the surface conformance test

**Files:**
- Create: `construct/scripts/test/gitignore-surface.test.sh`
- Modify: `scripts/parallel-checks.sh` (register it)

- [ ] **Step 1: Write the test**

A real two-layer scratch tree with a real `git init` — the invariant is about what `git` does with these patterns, and a fake would only test our assumption about git (ARCH-MOCK: this is the live conformance check).

```bash
#!/usr/bin/env bash
# #239: the committed-surface invariant, end to end, against real git.
# A derivative commits only its bootstrap core plus its own source; everything
# `make weave` re-derives is ignored, per-path.
set -euo pipefail
SOURCE="$(cd "$(dirname "$0")/../../.." && pwd)"
SCRATCH="$(mktemp -d "${TMPDIR:-/tmp}/gitignore-surface.XXXXXX")"
trap 'rm -rf "$SCRATCH"' EXIT
go build -o "$SCRATCH/weave" "$SOURCE/cmd/weave"

# Upstream: a miniature ariadne carrying one row of each ownership class.
mkdir -p "$SCRATCH/up/construct/scripts" "$SCRATCH/up/scripts/merge-checks.d"
printf 'UP\n' > "$SCRATCH/up/scripts/lib.sh"
printf 'CHECK\n' > "$SCRATCH/up/scripts/merge-checks.d/40-dup.sh"
printf 'BOOT\n' > "$SCRATCH/up/bootstrap.sh"
printf 'TEMPLATE\n' > "$SCRATCH/up/construct/Makefile.seed"
cat > "$SCRATCH/up/construct/base.manifest" <<'EOF'
symlink   scripts/lib.sh
symlink   scripts/merge-checks.d/40-dup.sh
scaffold  scripts/merge-checks.d
scaffold  workshop/issues
touch     workshop/lessons.md
seed      bootstrap.sh
seed-once construct/Makefile.seed Makefile
EOF

# Leaf: a derivative with its OWN files in the mixed directories, plus its own
# .gitignore entries and negations (pair's bin/* + !bin/*.sh shape).
mkdir -p "$SCRATCH/leaf/construct" "$SCRATCH/leaf/scripts/merge-checks.d" "$SCRATCH/leaf/bin"
printf 'substrate ../up\n' > "$SCRATCH/leaf/construct/deps"
: > "$SCRATCH/leaf/construct/base.manifest"
printf 'MINE\n' > "$SCRATCH/leaf/scripts/merge-checks.d/20-vocabulary.sh"
printf 'MINE\n' > "$SCRATCH/leaf/scripts/ci-setup.sh"
printf 'MINE\n' > "$SCRATCH/leaf/bin/helper.sh"
printf 'bin/*\n!bin/*.sh\ncache/\n' > "$SCRATCH/leaf/.gitignore"
cp "$SCRATCH/leaf/.gitignore" "$SCRATCH/repo-entries-before"
(cd "$SCRATCH/leaf" && git init -q . && git add -A && git -c user.email=t@t -c user.name=t commit -qm init)

(cd "$SCRATCH/leaf" && "$SCRATCH/weave" compile)

cd "$SCRATCH/leaf"
# 1. Repo-owned files beside weave symlinks are NEVER ignored (pair#64 / the
#    parley.nvim 20-vocabulary.sh hazard the issue proved is real).
for f in scripts/merge-checks.d/20-vocabulary.sh scripts/ci-setup.sh bin/helper.sh; do
  ! git check-ignore -q "$f" || { echo "FAIL: repo-owned $f is ignored"; exit 1; }
done
# 2. Scaffold/touch targets are NEVER ignored — these hold the repo's content.
for f in workshop/issues workshop/lessons.md; do
  ! git check-ignore -q "$f" || { echo "FAIL: $f is ignored"; exit 1; }
done
# 3. The bootstrap core stays committable.
for f in bootstrap.sh Makefile construct/deps; do
  ! git check-ignore -q "$f" || { echo "FAIL: bootstrap core $f is ignored"; exit 1; }
done
# 4. Re-derived paths ARE ignored.
for f in scripts/lib.sh scripts/merge-checks.d/40-dup.sh CLAUDE.md; do
  git check-ignore -q "$f" || { echo "FAIL: weave-generated $f is not ignored"; exit 1; }
done
# 5. The repo's own entries and negations round-trip verbatim, in place.
head -n 3 .gitignore | cmp - "$SCRATCH/repo-entries-before"
# 6. A second weave writes nothing.
cp .gitignore "$SCRATCH/after-first"
"$SCRATCH/weave" compile
cmp "$SCRATCH/after-first" .gitignore
# 7. RETIRING a manifest row removes its ignore line (append-only could not).
grep -v '40-dup' "$SCRATCH/up/construct/base.manifest" > "$SCRATCH/m" && mv "$SCRATCH/m" "$SCRATCH/up/construct/base.manifest"
"$SCRATCH/weave" compile
! grep -q '40-dup' .gitignore || { echo "FAIL: retired row left a stale ignore line"; exit 1; }
grep -q 'bin/\*' .gitignore || { echo "FAIL: repo entries lost on retire"; exit 1; }
echo 'PASS gitignore committed-surface invariant'
```

- [ ] **Step 2: Run it**

Run: `bash construct/scripts/test/gitignore-surface.test.sh`
Expected: `PASS gitignore committed-surface invariant`.

- [ ] **Step 3: Register it in the pre-merge checks**

Add it beside `portable-makefile.test.sh` in `scripts/parallel-checks.sh`. Confirm with `scripts/parallel-checks.sh` and check the new row appears in the output.

- [ ] **Step 4: Commit**

```bash
git add construct/scripts/test/gitignore-surface.test.sh scripts/parallel-checks.sh
git commit -m "#239 M3: conformance test for the committed-surface invariant"
```

### Task 3.4: record the invariant in the target and the atlas

**Files:**
- Modify: `workshop/targets/base-layer-mechanics.md`
- Modify: `atlas/workflow/base-layer.md`
- Modify: `atlas/workflow/weave.md:96`
- Modify: `construct/base.manifest:8-35` (the verb documentation header)

- [ ] **Step 1: Add the committed-surface invariant to the target**

Under the file-ops section, add a subsection stating the invariant and the ownership rule, in the target's existing register:

> **The committed surface.** A derivative commits only its bootstrap core plus its own source; every path `make weave` re-derives is gitignored. The rule is the verb's ownership class: weave **ignores what it re-derives** (`symlink`, `prose`, `merge`, the lowered skill links) and **tracks what it merely provisions** (`scaffold`, `touch`, `seed`, `seed-once`). The bootstrap core follows from that rather than being listed — both seed verbs mean "must work before any substrate exists". The ignore list is derived from the planned actions and maintained inside a weave-owned `.gitignore` block, so no artifact enters the *ignore* surface by a second channel either.

Also split the file-ops formula's verb list to name `seed` and `seed-once` as the two ownership classes, and update the `## Open questions` entry on file-op collision precedence if `seed-once` bears on it.

- [ ] **Step 2: Document the adoption path in the atlas**

`atlas/workflow/base-layer.md` gains a short section: a repo that already has a `Makefile` keeps it — `seed-once` never overwrites — and adopts ariadne by adding `include Makefile.workflow`, the contract `Makefile.workflow:1-2` already documents. Note that `WF_ISSUES_DIR`/`WF_HISTORY_DIR` now go in the repo's own root Makefile **above** the include, and that `Makefile.workflow`'s `?=` defaults are `issues`/`history`.

- [ ] **Step 3: Update the manifest verb documentation**

`construct/base.manifest`'s header block documents each verb. Add `seed-once` beside `seed`, stating the ownership split in one sentence each. This file is read by every agent that touches a manifest; a verb absent from it is a verb nobody uses.

- [ ] **Step 4: Update the weave atlas line**

`atlas/workflow/weave.md:96` describes `plan.EnsureGitignore` as weave owning a fixed entry set. Rewrite for the derived list + managed block.

- [ ] **Step 5: Commit and close the milestone**

```bash
git add workshop/targets/ atlas/ construct/base.manifest
git commit -m "#239 M3: record the committed-surface invariant"
sdlc milestone-close --issue 239 --milestone M3
```

---

## Chunk 4: M4 — the fleet untrack

Irreversible. It lands last, behind a green CI on one derivative.

### Task 4.1: prove the untrack set on one derivative, in a scratch clone

**Files:** none in ariadne — this is an operator procedure whose output goes in the issue's `## Log`.

- [ ] **Step 1: Pick the pilot and clone it to scratch**

`pair` is the right pilot: it carries the `bin/*` + `!bin/*.sh` negations, 28 tracked weave symlinks, a tracked `Makefile` symlink (so it exercises the `seed-once` materialization), and a retired-row orphan (`scripts/issue-sync.sh` is tracked but absent from today's manifest).

```bash
git clone ../pair "$TMPDIR/pair-pilot" && cd "$TMPDIR/pair-pilot"
printf 'substrate ../../workspace/ariadne\n' > construct/deps   # keep the clone pointed at the real ancestor
```

- [ ] **Step 2: Weave and enumerate what would be untracked**

```bash
../../workspace/ariadne/bin/weave compile
git ls-files -i -c --exclude-standard | sort > "$TMPDIR/untrack-set"
wc -l "$TMPDIR/untrack-set"; cat "$TMPDIR/untrack-set"
```

- [ ] **Step 3: Verify no repo-owned file is in the set**

The structural argument is that per-path derivation makes this impossible — a path weave never produces can never enter the block. Verify it anyway, because the argument is the thing under test:

```bash
comm -12 "$TMPDIR/untrack-set" <(git ls-files | sort) > "$TMPDIR/tracked-and-ignored"
# Every line must be a weave-produced path. Cross-check against the plan:
../../workspace/ariadne/bin/weave compile --dry-run > "$TMPDIR/plan.txt"
while read -r f; do grep -qF "$f" "$TMPDIR/plan.txt" || echo "NOT WEAVE-PRODUCED: $f"; done < "$TMPDIR/tracked-and-ignored"
```

Expected: no `NOT WEAVE-PRODUCED` line. Explicitly confirm `scripts/ci-setup.sh` and every `bin/*.sh` are absent from the untrack set. Record the full set in `## Log`.

- [ ] **Step 4: Confirm the fresh-clone bootstrap still works with only the core present**

This is the load-bearing claim of the whole issue — that #225's owner-resolution fallbacks really did dissolve the chicken-and-egg:

```bash
cd "$TMPDIR/pair-pilot" && git rm --cached -q $(cat "$TMPDIR/untrack-set") && git commit -qm untrack
git clone "$TMPDIR/pair-pilot" "$TMPDIR/pair-fresh"
cd "$TMPDIR/pair-fresh"
git ls-files | grep -E 'Makefile|bootstrap|merge-check|construct/deps'
ls Makefile.workflow scripts/lib.sh 2>&1   # expected: No such file — nothing but the core survived
make bootstrap 2>&1 | tail -40
```

Expected: `bootstrap` completes. The pre-weave resolutions that must carry it, each already verified in the issue's `## Log`: `Makefile:11`'s `$(wildcard Makefile.workflow ../ariadne/Makefile.workflow)`, `Makefile.workflow:10`'s `wf-helper`, and CI's `elif [ -f ../ariadne/scripts/run-merge-checks.sh ]`. If any resolution is missing, **stop and re-plan** — that is the design assumption failing, not a bug to push through.

- [ ] **Step 5: Record the result in the issue Log**

Whether it passed or failed, and the exact untrack set. This is the evidence M4's close gate needs.

### Task 4.2: land it on the pilot and prove CI

- [ ] **Step 1: Weave and untrack in the real pilot repo**

`sdlc propagate-base` already does exactly this: `commitConsumption` (`cmd/sdlc/propagatebase.go:243-259`) runs `git ls-files -i -c --exclude-standard` and `git rm --cached` each result, precisely to avoid the inert-gitignore trap. Use the verb rather than hand-rolling git (AGENTS.md: do not route around an sdlc verb):

```bash
cd ../pair && git status --short   # MUST be clean — commitConsumption's precondition
sdlc propagate-base --repo pair    # confirm the real flag names with `sdlc propagate-base --help`
```

If the verb cannot express "sweep this one repo", that is a genuine gap in `sdlc` — fix it there and re-run, per the workflow contract.

- [ ] **Step 2: Add pair's own `WF_*` lines**

After the `Makefile` symlink materializes into a real file, pair's root Makefile no longer carries `WF_ISSUES_DIR`. Add both lines above the include, per Task 1.5 Step 4's table, and commit in pair.

- [ ] **Step 3: Open a PR in the pilot and watch CI**

The Done-when requires CI green on a derivative PR — it proves `merge-check.yml` plus the `bootstrap.sh` CLONE_ONLY path carry the whole runner resolution with nothing else committed.

```bash
cd ../pair && sdlc pr
gh pr checks --watch
```

Expected: green. A failure here means the committed core is short one path — add it to the tracked class by making it a seed row, do not add an ad-hoc ignore exception.

- [ ] **Step 4: Record in the issue Log**

PR link, CI result, and the count of paths untracked.

### Task 4.3: sweep the remaining derivatives

- [ ] **Step 1: Enumerate the fleet from `construct/deps`, not from memory**

```bash
for d in ../*/; do [ -f "$d/construct/deps" ] && grep -q '^substrate' "$d/construct/deps" && basename "$d"; done
```

Expected (2026-09-19): `42shots astro brain-family brain-private brain kaggle kbench metis nous pair parley.nvim parli robotics tools xianxu.dev you-decide`. **Skip `brain`, `brain-family`, `brain-private`** — brain repos are capture-only and carry no SDLC surface; confirm each with `test -d .brain` before touching it.

- [ ] **Step 2: Per repo, run Task 4.1 Steps 2–3 then propagate-base**

Per repo, in this order: confirm the tree is clean, weave, enumerate the untrack set, verify no repo-owned path is in it, `sdlc propagate-base`, add the repo's own `WF_*` lines if its root Makefile just materialized. Do them one at a time and record each in `## Log` — a repo whose untrack set contains a surprise stops the sweep.

- [ ] **Step 3: Explicitly verify the two named files survive**

```bash
cd ../parley.nvim && git ls-files --error-unmatch scripts/merge-checks.d/20-vocabulary.sh
for d in ../*/; do [ -f "$d/scripts/ci-setup.sh" ] && (cd "$d" && git ls-files --error-unmatch scripts/ci-setup.sh >/dev/null && echo "ok $(basename "$d")"); done
```

Expected: every one still tracked. These are the two the issue names as the acceptance condition for "no repo-owned file is untracked by the sweep".

- [ ] **Step 4: Confirm the committed weave surface is now the core alone**

```bash
for d in ../*/; do
  [ -f "$d/construct/deps" ] || continue
  printf '%-16s symlinks=%s\n' "$(basename "$d")" "$(cd "$d" && git ls-files -s | awk '$1==120000' | wc -l)"
done
```

Expected: `0` for every derivative (pair was 28). Record the before/after table in `## Log` — it is the issue's headline result.

### Task 4.4: close

- [ ] **Step 1: Run the full check suite**

```bash
cd ../ariadne && make check
bash construct/scripts/test/portable-makefile.test.sh
bash construct/scripts/test/gitignore-surface.test.sh
go test ./cmd/...
```

- [ ] **Step 2: Walk the Done-when list**

Check each of the issue's ten criteria against evidence recorded in `## Log`, not against memory. Any unmet criterion is unfinished work, not a follow-up.

- [ ] **Step 3: Add the lessons entry**

Per AGENTS.md §4, add to `workshop/lessons.md` only what is not already enforced by code. The candidate: *a justification comment is not a test* — `gitignore.go:26-31` stated a bootstrap chicken-and-egg that #225 had already dissolved, and nothing failed when it went stale, so it survived months of manifest edits. The rule: when a comment explains why a set is what it is, the set gets a test that fails when the justification stops holding. (`gitignore-surface.test.sh` step 3 is that test here — so check whether the new test already enforces it before writing the entry.)

- [ ] **Step 4: Close the issue**

```bash
sdlc close --issue 239 --verified '<evidence: pilot CI link, fleet symlink counts before/after, conformance tests passing>'
```

Omit `--actual` — close measures and adopts the hours itself (#178).

---

## Risks and how this plan answers them

| Risk | Answer |
|---|---|
| A blanket ignore untracks a repo-owned file (pair#64, parley.nvim's `20-vocabulary.sh`) | Per-path derivation makes it structurally impossible: a path weave never produces can never enter the block. Tested in `IgnoreEntries` unit tests, the conformance test, and verified per repo in M4. |
| Ignoring `workshop/issues/` or `lessons.md` | The ownership rule puts `scaffold`/`touch` in the tracked class. Tested directly. |
| A lean `--target` shrinks the wholesale-replaced block | `planActions` always derives from `TargetAll`. This is a hazard M2 *creates*; M3 Task 3.2 closes it with a test. |
| A derivative's fresh clone stops bootstrapping | M4 Task 4.1 Step 4 clones the untracked pilot and runs `make bootstrap` before anything irreversible lands; a failure stops the plan. |
| A derivative silently falls back to `issues/` after its Makefile materializes | Task 1.5 Step 4 enumerates every repo's root Makefile before the sweep; M4 adds the two lines per repo. |
| A hand-edited `.gitignore` gets spliced wrong | `mergeManagedBlock` fails closed on any marker shape it cannot parse, and matches markers as exact whole lines. |
| The git history stops recording wiring changes in leaves | Accepted tradeoff, stated in the issue's Spec. If the audit trail is wanted back it belongs in `weave --explain` or a merge check, not in ~45 tracked symlinks per repo. |
