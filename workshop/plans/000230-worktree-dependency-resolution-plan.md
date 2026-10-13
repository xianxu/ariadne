# Worktree-aware Dependency Resolution Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A relative `substrate` row declared by a linked Git worktree resolves against
its primary checkout when, and only when, the worktree-local target is absent, and every
layer-graph consumer (weave compile, weave's read-only verbs, vocabulary, datatype,
sdlc's substrate walk) makes that choice through one resolver.

**Architecture:** A pure chooser (`ChooseSubstrate`) decides local-vs-fallback from two
probes. One Git boundary (`workspace.DependencyAnchor`) answers "what is this layer's
fallback anchor, or why does it have none". A small `layergraph.Resolver` composes the
two (lazy: Git runs only when a relative target is absent) and is used by `layergraph`'s
walk, by weave's `acquire.Restore`, and by sdlc's `substrateWalk`.

**Tech Stack:** Go; real `git worktree` fixtures in tests.

---

## Current behavior (measured 2026-10-12, before any change)

The issue was filed when `weave compile` used `layergraph.Walk`'s present-skip. Since
then compile goes through `acquire.Restore` (#296). Reproduced in a scratch fleet
(`fleet/base`, `fleet/leaf` declaring `substrate ../base`, a linked worktree at
`fleet/worktree/leaf/br`), with a weave built from `864aeef1`:

- sourceless row: `weave compile` fails with `missing substrate …/worktree/leaf/base
  declared in …/br: record its source in construct/deps (no source declared; no numbered
  environment to recover it from)`. Nothing is pruned, but the worktree cannot compile.
- sourced row: Restore would **clone a duplicate** `base` into `fleet/worktree/leaf/base`
  (dry run reports `missing source …/worktree/leaf/base`), a stale copy nobody owns.
- `layergraph.Walk` consumers (vocabulary validation, datatype, weave
  verify-complete/golden/skills) still silently skip the absent row, so the vocabulary
  that sdlc's merge validation needs (the parley.nvim#254 failure) is missing.

So spec item 6 (no prune before an error) already holds for compile; this plan keeps it
under test, and fixes the resolution itself.

## Decisions

1. **Fallback anchor = Git's primary working tree** (`git worktree list` first entry,
   via the existing `workspace` topology loader), only when the declaring layer is
   exactly a linked worktree's root. No search of other worktrees (spec 4).
2. **Local wins when present**, valid or not. A present local target without a manifest
   stays a `NotLayerError` (spec 5); a present-but-invalid fallback target is also a
   `NotLayerError` rather than being skipped.
3. **Absolute rows never fall back** (spec 5).
4. **Numbered environments keep slot-local provisioning.** A slot host (e.g.
   `worktree/pair-slot1/pair`, a linked worktree of `~/workspace/pair`) must not borrow
   the fleet primary's siblings: `acquire.Policy` restricts dependencies to the
   environment and #296 recovers sources slot-locally. `DependencyAnchor` therefore
   returns no anchor when the vantage is in a numbered environment and the primary lies
   outside it. A feature worktree of an environment's independent dependency clone
   (`<env>/.worktrees/ariadne/br`, primary `<env>/ariadne`) still falls back, inside the
   environment. This refines the 2026-09-21 log entry ("otherwise use the primary
   checkout"), which predates the environment policy; recorded in the issue's Revisions.
5. **No-anchor reasons are values, not failures.** `DependencyAnchor` returns an error
   naming why there is no anchor (not a checkout root, is the primary, bare primary,
   numbered-environment host, Git failure). `Walk` keeps its present-skip for an absent
   row (existing compatibility) and records the reason on `DeclaredSubstrate`;
   `Restore`'s missing-substrate error appends it, so the operator sees why the
   fallback did not apply.
6. **Reporting:** `Restore` returns `Fallbacks`; `weave compile` prints one line per
   fallback (declared row, owner, chosen path, primary). `DeclaredSubstrate.Via` carries
   the same provenance for read-only consumers.
7. **Out of scope:** `data` rows (their destination is `dir(owner)/<name>`, a clone weave
   provisions; a worktree-local data clone is redundant but correct), and the bash
   bootstrap parser `construct/scripts/lib-deps.sh` (runs before weave exists). Both
   noted in the issue Log.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `Probe` | `pkg/layergraph/substrate.go` | new |
| `Resolution` | `pkg/layergraph/substrate.go` | new |
| `ChooseSubstrate` | `pkg/layergraph/substrate.go` | new |
| `DeclaredSubstrate` (+`Declared`, `Via`, `Unanchored`) | `pkg/layergraph/walk.go` | modified |
| `acquire.Result` (+`Fallbacks`), `acquire.Fallback` | `cmd/weave/internal/acquire/acquire.go` | modified / new |

- **ChooseSubstrate(owner, target, primary string, probe func(string) Probe) Resolution**
  — the whole selection rule, no IO. Tests in `substrate_test.go` are a table over probe
  stubs.
  - **Relationships:** 1 resolution per (owner, row).
  - **DRY rationale:** today three resolvers (`substrateTargets`, `acquire.Restore`,
    `sdlc substrateWalk`) each join owner+target; the fallback rule must exist once.
  - **Future extensions:** a second anchor kind (e.g. an explicit slot map) widens
    `primary` into an ordered anchor list.
- **Resolution** `{Path, Local, Via string; Probe Probe}` — `Path` is the chosen
  lexical path, `Via` the primary used (`""` when local).

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `workspace.DependencyAnchor` | `pkg/workspace/anchor.go` | new | `git rev-parse`, `git worktree list` (via `loadVantage`) |
| `workspace.ExecGit` | `pkg/workspace/anchor.go` | new | `os/exec` git |
| `layergraph.Anchors` / `GitAnchors` | `pkg/layergraph/anchor.go` | new | `DependencyAnchor` |
| `layergraph.Resolver` | `pkg/layergraph/anchor.go` | new | `FS` probes + `Anchors`, memoized per owner |

- **DependencyAnchor(git GitReader, dir string) (string, error)** — returns the primary
  root, or an error naming why there is none. It first `Lstat`s `dir/.git`: without it
  the layer is not a checkout root (a non-Git layer, or a subdirectory whose enclosing
  repository must not lend its primary) and Git is not run.
  - **Injected into:** `GitAnchors`; tested with real Git and `workspacetest.Fake`.
- **Resolver** — `Substrate(owner, target) (Resolution, error)` runs the local probe;
  only on `Absent` for a relative row does it ask `Anchors` (memoized per owner for one
  walk). The returned error is the no-anchor reason, never fatal by itself.
  - **Injected into:** `declaredGraph`, `acquire.Client.Restore` (field `Anchors`,
    default `GitAnchors`), sdlc `substrateWalk`. A test FS may implement `Anchors` itself
    (same optional-interface pattern as `DeclarationReader`), so in-memory fakes stay
    hermetic.

**Lifecycle (ARCH-FUNERAL):** creates nothing durable: the per-owner anchor memo lives in
one `Resolver` value for one walk. The change removes a durable artifact (the duplicate
clone Restore would make beside a worktree).

## Tasks

### Task 1: Pure chooser

**Files:** Create `pkg/layergraph/substrate.go`, `pkg/layergraph/substrate_test.go`.

- [ ] **Step 1: failing table test** `TestChooseSubstrate` with rows (each asserts `Path`,
  `Via`, `Probe`):
  - local Layer, primary set → local, `Via==""`, probe never asked for the candidate
    (stub records calls).
  - local NotLayer, primary set (candidate Layer) → local NotLayer (invalid local is not
    bypassed).
  - local Absent, primary `""` → local Absent.
  - local Absent, candidate Layer → candidate, `Via==primary`.
  - local Absent, candidate NotLayer → candidate NotLayer, `Via==primary`.
  - local Absent, candidate Absent → local Absent, `Via==""`.
  - absolute target, local Absent, candidate would be Layer → target itself, no fallback.
  - target with spaces (`../my base`) → candidate `primary/../my base` joined correctly.
- [ ] **Step 2:** `go test ./pkg/layergraph -run TestChooseSubstrate` → FAIL (undefined).
- [ ] **Step 3: implement**

```go
// Probe is what a resolver observed at one candidate path.
type Probe int

const (
	Absent   Probe = iota // nothing on disk
	NotLayer              // present, but no construct/base.manifest
	Layer                 // present and ships construct/base.manifest
)

// Resolution is where one substrate row resolves. Path is lexical; the
// boundary canonicalizes it. Via is the primary checkout used, "" when local.
type Resolution struct {
	Path, Local, Via string
	Probe            Probe
}

// ChooseSubstrate is #230's selection rule, pure. A relative target resolves
// against owner first; only when nothing is there and owner has a fallback
// anchor (its Git primary checkout) is the same relative path tried from the
// anchor. A present local target wins even when invalid, so a broken local
// layer is reported, never bypassed. Absolute targets never fall back.
func ChooseSubstrate(owner, target, primary string, probe func(string) Probe) Resolution {
	if filepath.IsAbs(target) {
		return Resolution{Path: target, Local: target, Probe: probe(target)}
	}
	local := filepath.Join(owner, target)
	r := Resolution{Path: local, Local: local, Probe: probe(local)}
	if r.Probe != Absent || primary == "" {
		return r
	}
	candidate := filepath.Join(primary, target)
	if p := probe(candidate); p != Absent {
		return Resolution{Path: candidate, Local: local, Via: primary, Probe: p}
	}
	return r
}
```

- [ ] **Step 4:** test PASS. **Step 5:** commit `#230: layergraph: pure substrate chooser`.

### Task 2: Git anchor boundary

**Files:** Create `pkg/workspace/anchor.go`, `pkg/workspace/anchor_test.go`.

- [ ] **Step 1: failing tests** (real Git, `realGit` helper from `conformance_test.go`;
  temp root canonicalized):
  - `TestDependencyAnchorLinkedWorktree`: `repo` primary + `git worktree add
    ../worktree/repo/br` → returns canonical `repo`.
  - `…Primary`: called on `repo` → error containing `primary checkout`.
  - `…Subdirectory`: `repo/sub` (no `.git` entry) → error `not a Git checkout root`,
    and the fake GitReader records zero calls.
  - `…NonGit`: plain dir → same error, zero Git calls.
  - `…Bare`: `git init --bare b.git`, `git -C b.git worktree add ../w` → error
    mentioning `bare`.
  - `…NumberedEnvironmentHost`: the conformance fixture's slot
    `worktree/repo-slot1/repo` (linked worktree of `repo`) → error `numbered environment`.
  - `…SpacesAndSymlink`: fleet root `my fleet`, worktree reached through a symlinked
    path → returns the canonical primary.
- [ ] **Step 2:** `go test ./pkg/workspace -run TestDependencyAnchor` → FAIL.
- [ ] **Step 3: implement**

```go
// ExecGit is the production GitReader: git run in dir, combined output.
type ExecGit struct{}

func (ExecGit) GitInDir(dir string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return out, fmt.Errorf("git %v in %q: %w", args, dir, err)
	}
	return out, nil
}

// DependencyAnchor is a layer's fallback anchor for relative dependencies
// (#230): its repository's primary checkout, when dir is exactly the root of
// a linked worktree. The error names why there is none. A numbered
// environment provisions dependencies inside itself, so a vantage whose
// primary lies outside its environment has no anchor.
func DependencyAnchor(git GitReader, dir string) (string, error) {
	if _, err := os.Lstat(filepath.Join(dir, ".git")); err != nil {
		return "", fmt.Errorf("%s is not a Git checkout root", dir)
	}
	v, err := NormalizeVantage(git, dir)
	if err != nil {
		return "", err // includes "bare primary worktree"
	}
	root, err := CanonicalPath(dir)
	if err != nil {
		return "", err
	}
	switch {
	case v.WorktreeRoot != root:
		return "", fmt.Errorf("%s is not a Git checkout root", dir)
	case v.PrimaryRoot == root:
		return "", fmt.Errorf("%s is its repository's primary checkout", dir)
	case v.EnvironmentHost != nil && !within(v.EnvironmentRoot, v.PrimaryRoot):
		return "", fmt.Errorf("%s is in numbered environment %s, which provisions its own dependencies", dir, v.EnvironmentRoot)
	}
	return v.PrimaryRoot, nil
}
```

  (`within` = equal or prefix with separator; reuse an existing helper if `pkg/workspace`
  has one.)
- [ ] **Step 4:** PASS. **Step 5:** commit `#230: workspace: dependency anchor`.

### Task 3: Resolver, and the layergraph walk on it

**Files:** Create `pkg/layergraph/anchor.go`; modify `pkg/layergraph/walk.go`
(`declaredGraph`, `substrateTargets`, `DeclaredSubstrate`); create
`pkg/layergraph/worktree_test.go` (real Git); `pkg/layergraph` gets a `TestMain` calling
`testgate.Enter()` if the testgate guard requires one for packages that exec Git.

- [ ] **Step 1: failing integration tests** in `worktree_test.go`, fixture
  `fleet/{base,leaf}` with manifests, `leaf/construct/deps` = `substrate ../base`,
  linked worktree `fleet/worktree/leaf/br`:
  - `TestWalkLinkedWorktreeFallsBackToPrimarySibling`: `Walk(OSFS{}, br)` ==
    `[fleet/base, br]`; `DeclaredSubstrates` row has `Declared=="../base"`,
    `Via==fleet/leaf`, `Present`. The deps file and the worktree's parent are unchanged
    (no symlink created: `Lstat(fleet/worktree/leaf/base)` is NotExist).
  - `…LocalSiblingWins`: also create a valid `fleet/worktree/leaf/base` layer → Walk
    picks it, `Via==""`.
  - `…InvalidLocalNotBypassed`: `fleet/worktree/leaf/base` exists without manifest →
    `*NotLayerError` with that path.
  - `…Transitive`: `base` declares `substrate ../core`; `fleet/core` exists → order
    `[core, base, br]` (core resolves beside the primary-anchored base).
  - `…Absolute`: deps row with the absolute path of an absent dir → skipped, no fallback.
  - `…Spaces`: the whole fleet under `my fleet/`.
  - `…NoAnchorKeepsPresentSkip`: leaf primary (not a worktree), base absent → Walk ==
    `[leaf]`, `DeclaredSubstrate.Unanchored` mentions `primary checkout`.
  - `…Vocabulary` (the parley.nvim#254 regression): `base/construct/vocabulary/issue.cue`
    exists; `MergeByName` over Walk's roots from `br` contains `issue.cue` — the
    resolution `cmd/vocabulary` performs before sdlc's merge validation.
- [ ] **Step 2:** run `go test ./pkg/layergraph -run 'TestWalkLinkedWorktree'` → FAIL
  (falls into present-skip).
- [ ] **Step 3: implement** `anchor.go`:

```go
// Anchors finds a declaring layer's fallback anchor (#230). An error is the
// reason there is none; resolution then stays local.
type Anchors interface {
	PrimaryCheckout(layer string) (string, error)
}

// GitAnchors is the production Anchors: workspace.DependencyAnchor over git.
type GitAnchors struct{}

func (GitAnchors) PrimaryCheckout(layer string) (string, error) {
	return workspace.DependencyAnchor(workspace.ExecGit{}, layer)
}

// Resolver resolves substrate rows for one walk: FS probes, plus an anchor
// looked up only when a relative target is absent, memoized per owner.
type Resolver struct {
	FS      FS
	Anchors Anchors
	memo    map[string]anchorResult
}

type anchorResult struct {
	primary string
	err     error
}

// NewResolver uses fs's own Anchors when it implements one (test fakes),
// else GitAnchors.
func NewResolver(fs FS) *Resolver {
	a, ok := fs.(Anchors)
	if !ok {
		a = GitAnchors{}
	}
	return &Resolver{FS: fs, Anchors: a}
}

// Substrate resolves one row; the error is the no-anchor reason when the
// local target was absent and no fallback applied (informational).
func (r *Resolver) Substrate(owner, target string) (Resolution, error) {
	probe := func(p string) Probe {
		switch {
		case hasManifest(r.FS, p):
			return Layer
		case pathExists(r.FS, p):
			return NotLayer
		}
		return Absent
	}
	if filepath.IsAbs(target) || probe(filepath.Join(owner, target)) != Absent {
		return ChooseSubstrate(owner, target, "", probe), nil
	}
	a := r.anchor(owner)
	return ChooseSubstrate(owner, target, a.primary, probe), a.err
}
```

  `walk.go`: `declaredGraph` builds one `NewResolver(fs)`; `substrateTargets(res, cur)`
  calls `res.Substrate(repoRoot, row.Path)`, canonicalizes `Resolution.Path` exactly as
  today (`physicalOrEmpty(parent)`, skip when empty), and carries `Declared`, `Via`,
  `Unanchored` (the error's text, only when the row stayed absent) into
  `substrateRow`/`DeclaredSubstrate`. The `NotLayerError` check is unchanged and now
  also covers an invalid fallback target. Update the `Walk` doc comment: relative rows
  fall back to the declaring layer's primary checkout (#230).
- [ ] **Step 4:** the new tests and the whole package pass:
  `go test ./pkg/layergraph ./cmd/vocabulary ./cmd/datatype`.
- [ ] **Step 5:** commit `#230: layergraph: worktree-aware substrate resolution`.

### Task 4: weave compile (acquire.Restore) on the same resolver

**Files:** modify `cmd/weave/internal/acquire/acquire.go` (`Client.Anchors`, `Result.Fallbacks`,
substrate branch of `Restore`), `cmd/weave/migrate.go` (`restoreSetup` reporting); tests in
`cmd/weave/internal/acquire/acquire_test.go` and `cmd/weave/main_test.go`.

- [ ] **Step 1: failing tests**
  - acquire: `TestRestoreLinkedWorktreeUsesPrimarySibling` — the Task 3 fixture, row
    `substrate ../base <bogus file:// URL that does not exist>`; `Restore(br, false)`
    succeeds, `Layers==[fleet/base, br]`, `Fallbacks` has one entry
    `{Owner: br, Path: "../base", Resolved: fleet/base, Primary: fleet/leaf}`, and
    `fleet/worktree/leaf/base` does not exist (no clone attempted — the bogus URL would
    fail one). Use a sourceless row variant too (no "record its source" error).
  - acquire: `TestRestoreMissingEverywhereNamesAnchorReason` — base absent at both
    places, sourceless → error still contains `record its source` and now
    `worktree fallback:`.
  - weave: `TestCompileLinkedWorktreePreservesOutputOnMissingDependency` — compile once
    with base present (generated output exists), remove `fleet/base`, compile again →
    error, previously generated files byte-identical.
  - weave: `TestCompileLinkedWorktreeReportsFallback` — compile output contains
    `resolved via primary checkout`.
- [ ] **Step 2:** run `go test ./cmd/weave/internal/acquire -run 'LinkedWorktree|AnchorReason'`
  and `go test ./cmd/weave -run 'CompileLinkedWorktree'` → FAIL.
- [ ] **Step 3: implement** in `Restore`'s substrate branch, replacing the bare
  owner-join for relative rows:

```go
if !filepath.IsAbs(dest) {
	res, reason := c.resolver().Substrate(owner, row.Path)
	dest = res.Path
	if res.Via != "" {
		result.Fallbacks = append(result.Fallbacks, Fallback{Owner: owner, Path: row.Path, Resolved: canonical(dest), Primary: res.Via})
	} else if reason != nil && res.Probe == Absent {
		anchorReason = reason // appended to this row's missing-substrate error
	}
}
```

  `c.resolver()` returns `layergraph.NewResolver(layergraph.OSFS{})` with `c.Anchors`
  when set (one per `Restore` call, so the memo is per walk). Policy validation,
  `checkExisting`, recovery and cloning then run on `dest` unchanged; an environment
  never sees an out-of-environment fallback because `DependencyAnchor` refuses one.
  Wrap the two missing-substrate errors (`recoverSource` failure, and `Ensure`'s
  "missing repository") with `; worktree fallback: <reason>` when `anchorReason` is set.
  `restoreSetup` prints per fallback:
  `weave: substrate <Path> declared in <Owner> is absent beside this worktree; resolved via primary checkout <Primary>: <Resolved>`.
- [ ] **Step 4:** tests PASS; `go test ./cmd/weave/... -run 'Restore|Compile|Walk'` green.
- [ ] **Step 5:** commit `#230: weave: compile resolves worktree deps via the primary`.

### Task 5: sdlc substrate walk on the same resolver

**Files:** modify `cmd/sdlc/startplan.go` (`substrateWalk`); test in
`cmd/sdlc/startplan_test.go` (or the existing `substrateWalk` test file).

- [ ] **Step 1: failing test** `TestSubstrateWalkLinkedWorktree`: the Task 3 fixture;
  `substrateWalk(br)` returns `order==[fleet/base]`, `absent` empty.
- [ ] **Step 2:** `go test ./cmd/sdlc -run 'TestSubstrateWalk'` → FAIL.
- [ ] **Step 3:** replace the owner-join in `substrateWalk` with one
  `layergraph.NewResolver(layergraph.OSFS{})` per walk and `res.Path`; keep its
  canonicalization, seen-set and present test (it accepts non-layer dirs; unchanged).
- [ ] **Step 4:** PASS, plus existing `TestSubstrateWalk*`/`TestParseSubstrateTargets`.
- [ ] **Step 5:** commit `#230: sdlc: substrate walk shares worktree resolution`.

### Task 6: atlas, docs, verify

- [ ] Update `atlas/` (the layer-graph / weave dependency page that documents
  `construct/deps` resolution — find with `grep -rln "present-skip\|construct/deps" atlas`)
  with the resolution order, the environment exception, and the compile notice; keep
  `atlas/index.md` links valid.
- [ ] Grep every restatement of "relative to the repo root" / `deps_substrate_targets`
  resolution in Go comments and help (`weave --help` text, `cmd/weave/main.go` comments)
  and update them.
- [ ] Issue Log: record the out-of-scope `data` rows and `lib-deps.sh`, and the
  environment refinement (Revisions).
- [ ] Full suite in the background (`make test`), then `sdlc close`.
