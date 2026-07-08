# Topological Settings Merge Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make weave compose settings fragments across the whole selected layer stack, foundation-first, with repo-local settings applied last.

**Architecture:** Keep settings semantics in the pure `settingsx` package and keep filesystem reads/writes in the existing `plan.Apply` IO seam (ARCH-PURE). Change `MergeSettings` from one source to an ordered source chain, and make every consumer of the action derive from that one shape rather than adding a parallel multi-source path (ARCH-DRY). The work is complete only when plan lowering, apply, golden, completeness, and docs all understand the chain (ARCH-PURPOSE).

**Tech Stack:** Go, `cmd/weave/internal/settingsx`, `cmd/weave/internal/plan`, `cmd/weave/internal/golden`, `cmd/weave` integration tests.

---

## Core Concepts

### Pure Entities

| Name | Lives in | Status |
|------|----------|--------|
| `settingsx.MergeChain` | `cmd/weave/internal/settingsx/settingsx.go` | new |
| `settingsx.Merge` | `cmd/weave/internal/settingsx/settingsx.go` | modified |
| `plan.MergeSettings` | `cmd/weave/internal/plan/action.go` | modified |
| `plan.Plan` merge grouping | `cmd/weave/internal/plan/plan.go` | modified |
| `golden.CheckCompleteness` merge coverage | `cmd/weave/internal/golden/completeness.go` | modified |
| `golden.Classify` merge classification | `cmd/weave/internal/golden/golden.go` | modified |

**`settingsx.MergeChain`** - ordered N-source settings fold.
- **Relationships:** 1:N from a `MergeSettings` action to source JSON byte slices; appends optional repo-local bytes as the final source.
- **DRY rationale:** One pure merge engine serves `Apply`, golden classification, and existing two-input `Merge`.
- **Future extensions:** If settings gains per-layer metadata beyond `$merge_keys` and `$remove`, it widens here rather than in action consumers.

**`settingsx.Merge`** - compatibility wrapper for the historical base+local API.
- **Relationships:** 1:1 wrapper over `MergeChain`.
- **DRY rationale:** Existing callers and tests keep the old API while the implementation has one core fold.
- **Future extensions:** Can be retired only after all consumers use `MergeChain` directly.

**`plan.MergeSettings`** - action representing one target settings output from an ordered source list.
- **Relationships:** N:1 from manifest `merge` intents sharing a target to one action.
- **DRY rationale:** `Apply`, prune, golden gather, golden classify, and completeness consume one action shape.
- **Future extensions:** Could later carry a local override path if settings targets stop using the sibling `settings.local.json` convention.

**`plan.Plan` merge grouping** - pure lowering that groups selected `intent.Merge` rows by target while preserving first-seen target order and foundation-first source order.
- **Relationships:** consumes the selected layer intents; emits one `MergeSettings` per target.
- **DRY rationale:** Grouping belongs in the planner because the planner already owns action derivation from layer order.
- **Future extensions:** Multiple settings targets keep independent chains keyed by target.

**`golden.CheckCompleteness` merge coverage** - validates that each selected merge intent is covered by a chain source in the planned action for its target.
- **Relationships:** selected manifest intents to planned `MergeSettings` actions.
- **DRY rationale:** Completeness must not only see "some action writes the target"; it must catch a dropped middle-layer source.
- **Future extensions:** If coverage details are reported in CLI output later, this index can name the missing source.

**`golden.Classify` merge classification** - recomputes expected merged settings from all observed action sources plus optional local and compares semantic JSON.
- **Relationships:** one observed target, N observed action sources, optional local sibling.
- **DRY rationale:** Classification uses `settingsx.MergeChain`, same as `Apply`.
- **Future extensions:** If a source is absent, keep current "unexpected" behavior but name which source is missing.

### Integration Points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `plan.applyMergeSettings` | `cmd/weave/internal/plan/apply.go` | modified | `weavefs.FS` file reads/writes |
| `golden.Gather` merge observation | `cmd/weave/internal/golden/gather.go` | modified | live filesystem observation |
| `weave compile` integration fixture | `cmd/weave/main_test.go` | modified | end-to-end compile over temp repos |
| `atlas/workflow/weave.md` | `atlas/workflow/weave.md` | modified | human architecture map |

**`plan.applyMergeSettings`** - reads all source files for a chain, reads optional sibling `settings.local.json`, calls `settingsx.MergeChain`, writes the target.
- **Injected into:** `plan.Apply` as the existing IO branch for `MergeSettings`.
- **Future extensions:** Extra local-source conventions would remain here, not in pure merge semantics.

**`golden.Gather` merge observation** - records every chain source, the local sibling, and target with followed-symlink content.
- **Injected into:** `runGolden` and `Classify`.
- **Future extensions:** Can expose per-source observation diagnostics.

**`weave compile` integration fixture** - synthetic base -> mid -> derived repo proving middle settings fragments are applied.
- **Injected into:** no production code; uses real temp files through `weavefs.OSFS`.
- **Future extensions:** Can double as a regression for non-Claude settings targets if more targets appear.

**`atlas/workflow/weave.md`** - documents the settings backend as an N-source chain.
- **Injected into:** human navigation and future issue planning.
- **Future extensions:** If settings gains a target invariant, link it from the atlas entry.

## Chunk 1: Pure Settings Chain

### Task 1: Add failing multi-source merge tests

**Files:**
- Modify: `cmd/weave/internal/settingsx/settingsx_test.go`
- Modify: `cmd/weave/internal/settingsx/settingsx.go`

- [ ] **Step 1: Write `TestMergeChainPreservesMergeKeysAcrossIntermediateSources`**

Add a test that calls the not-yet-existing `MergeChain` with foundation, middle, leaf, and local JSON sources:

```go
func TestMergeChainPreservesMergeKeysAcrossIntermediateSources(t *testing.T) {
	got := runMergeChain(t, []string{
		`{"$merge_keys":["permissions.allow"],"permissions":{"allow":["A"]},"scalar":"base"}`,
		`{"permissions":{"allow":["B"]},"scalar":"mid"}`,
		`{"permissions":{"allow":["C"]},"leaf":true}`,
		`{"permissions":{"allow":["D"]},"scalar":"local"}`,
	})
	want := map[string]any{
		"permissions": map[string]any{"allow": []any{"A", "B", "C", "D"}},
		"scalar": "local",
		"leaf": true,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("MergeChain:\n got=%#v\nwant=%#v", got, want)
	}
}
```

- [ ] **Step 2: Write `TestMergeChainAppliesRemoveFromFinalLocalOnly`**

Prove `$remove` in the final source filters the accumulated base before the final union, while the output still strips all meta keys.

- [ ] **Step 3: Run RED**

Run: `go test ./cmd/weave/internal/settingsx -run 'TestMergeChain' -count=1`

Expected: compile failure because `MergeChain` is undefined.

- [ ] **Step 4: Implement `MergeChain`**

Implementation shape:

```go
func MergeChain(sources [][]byte) ([]byte, error) {
	if len(sources) == 0 {
		return nil, fmt.Errorf("settingsx.MergeChain: no sources")
	}
	objects := make([]map[string]any, 0, len(sources))
	for i, src := range sources {
		var obj map[string]any
		if err := json.Unmarshal(src, &obj); err != nil {
			return nil, fmt.Errorf("settingsx.MergeChain: parse source %d: %w", i, err)
		}
		objects = append(objects, obj)
	}

	mergeKeys := mergeKeySet(objects[0])
	acc := deepCopy(objects[0]).(map[string]any)
	for i := 1; i < len(objects); i++ {
		next := objects[i]
		baseForMerge := acc
		if i == len(objects)-1 {
			if removals, ok := next["$remove"].(map[string]any); ok && len(removals) > 0 {
				baseForMerge = applyRemovals(acc, removals)
			}
		}
		merged := deepMerge(baseForMerge, next, "", mergeKeys)
		acc, _ = merged.(map[string]any)
	}
	result := stripMeta(acc).(map[string]any)
	out, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("settingsx.MergeChain: marshal result: %w", err)
	}
	return append(out, '\n'), nil
}
```

Extract `mergeKeySet(baseObj map[string]any) map[string]bool` from the current `Merge` body.

- [ ] **Step 5: Refactor `Merge` to delegate to `MergeChain`**

Keep the old local-absent behavior:

```go
func Merge(base, local []byte) ([]byte, error) {
	if local == nil {
		return MergeChain([][]byte{base})
	}
	return MergeChain([][]byte{base, local})
}
```

- [ ] **Step 6: Run GREEN**

Run: `go test ./cmd/weave/internal/settingsx -count=1`

Expected: PASS. Existing two-input tests remain green.

## Chunk 2: Action Shape and Apply

### Task 2: Convert `MergeSettings` to ordered sources

**Files:**
- Modify: `cmd/weave/internal/plan/action.go`
- Modify: `cmd/weave/internal/plan/plan.go`
- Modify: `cmd/weave/internal/plan/plan_test.go`
- Modify: `cmd/weave/internal/plan/apply.go`
- Modify: `cmd/weave/internal/plan/apply_test.go`
- Modify: `cmd/weave/internal/plan/prune.go`
- Modify: `cmd/weave/main.go`

- [ ] **Step 1: Update planner tests first**

Change `TestPlanMergeLowering` to expect:

```go
MergeSettings{Sources: []string{"/ws/ariadne/.claude/settings.ariadne.json"}, Target: ".claude/settings.json"}
```

Add `TestPlanGroupsMergeRowsByTargetFoundationFirst` with base, mid, and leaf layers. It should expect one `MergeSettings` for `.claude/settings.json` whose `Sources` are absolute layer-joined paths in layer order:

```go
MergeSettings{
	Sources: []string{
		"/ws/base/.claude/settings.base.json",
		"/ws/mid/.claude/settings.mid.json",
		"/ws/leaf/.claude/settings.leaf.json",
	},
	Target: ".claude/settings.json",
}
```

Run: `go test ./cmd/weave/internal/plan -run 'TestPlan.*Merge' -count=1`

Expected: FAIL because `MergeSettings.Source` still exists and grouping is not implemented.

- [ ] **Step 2: Change the action type**

In `action.go`, replace `Source string` with:

```go
Sources []string
Target  string
```

The `Sources` entries should be absolute layer-joined paths, matching `Symlink.Src` and `Seed.Src`; this lets a downstream repo read each ancestor's real fragment instead of requiring every fragment to be present in the leaf checkout.

- [ ] **Step 3: Group merge intents in `Plan`**

In `Plan`, collect selected `intent.Merge` rows into an ordered map during the existing layer/intents scan, and append the grouped merge actions after the ordinary file-op actions. Merge writes a generated target and does not feed later pure planning, so preserving exact interleaving is unnecessary; preserving source order inside each target chain is the important behavior.

Sketch:

```go
type mergeGroup struct {
	target  string
	sources []string
}
mergeGroups := map[string]*mergeGroup{}
var mergeOrder []string
```

When seeing a merge intent, append `joinPath(l.Path, in.Source)` to the group for `in.Target`; do not append an action immediately. After the intent scan, append one `MergeSettings` per `mergeOrder`.

- [ ] **Step 4: Update `applyMergeSettings`**

Read every `act.Sources` path directly. Because sources are absolute, do not join them with `repoRoot`.

```go
sources := make([][]byte, 0, len(act.Sources)+1)
for _, sourcePath := range act.Sources {
	data, err := fs.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("apply merge: read source %s: %w", sourcePath, err)
	}
	sources = append(sources, data)
}
if data, lerr := fs.ReadFile(localPath); lerr == nil {
	sources = append(sources, data)
}
merged, err := settingsx.MergeChain(sources)
```

Return an explicit error if `len(act.Sources) == 0`.

- [ ] **Step 5: Update apply tests**

Existing tests now pass absolute source paths in `MergeSettings.Sources`. Add `TestApplyMergeSettingsMultipleSourcesWithLocal` proving base, middle, and local compose into one target.

- [ ] **Step 6: Update action printers and prune**

Replace every `act.Source` reference in dry-run output, prune managed-location scans, and similar action fan-out code with `act.Sources`. Keep default unknown-action branches unchanged.

- [ ] **Step 7: Run GREEN**

Run: `go test ./cmd/weave/internal/plan -count=1`

Expected: PASS.

## Chunk 3: Golden and Completeness Consumers

### Task 3: Teach golden about source chains

**Files:**
- Modify: `cmd/weave/internal/golden/gather.go`
- Modify: `cmd/weave/internal/golden/gather_test.go`
- Modify: `cmd/weave/internal/golden/golden.go`
- Modify: `cmd/weave/internal/golden/golden_test.go`
- Modify: `cmd/weave/internal/golden/completeness.go`
- Modify: `cmd/weave/internal/golden/completeness_test.go`

- [ ] **Step 1: Update golden tests first**

Change existing `MergeSettings` fixtures to use `Sources`.

Add `TestMergeSettingsChainSemanticMatch` in `golden_test.go` where observed base, mid, local, and target prove semantic match.

Add a completeness test where two selected merge intents share a target but the planned action includes only one source. Expected: uncovered entry naming the missing source. This prevents the old "same target means covered" under-production gap.

Run: `go test ./cmd/weave/internal/golden -run 'MergeSettings|Completeness' -count=1`

Expected: FAIL until consumers inspect `Sources`.

- [ ] **Step 2: Update gather**

For a `MergeSettings`, call `observeMerge` for every source in `act.Sources`. Sources are absolute; use `observeAbs` or a new helper that follows symlinks and records content by absolute path. Continue observing `act.Target` and local sibling.

- [ ] **Step 3: Update classify**

Build the chain from `act.Sources` in order:

```go
var chain [][]byte
for _, source := range act.Sources {
	sourceO := obs[source]
	if !sourceO.Exists {
		return Divergence{Unexpected, "merge", act.Target, fmt.Sprintf("weave would merge %s, but source %s absent in live", act.Target, source)}
	}
	chain = append(chain, []byte(sourceO.Content))
}
if localO := obs[localAbs]; localO.Exists {
	chain = append(chain, []byte(localO.Content))
}
merged, err := settingsx.MergeChain(chain)
```

Preserve semantic JSON comparison with `settingsx.SemanticEqual`.

- [ ] **Step 4: Update completeness**

Index merge actions as `target -> set(source)`, not only `target -> bool`. `coverIntent` should require a `MergeSettings` for the target and the joined source path for that intent.

Because `coverIntent` currently receives only the `intent.Intent`, widen it to receive the layer path or compute expected source during the layer loop before calling the helper.

- [ ] **Step 5: Run GREEN**

Run: `go test ./cmd/weave/internal/golden -count=1`

Expected: PASS.

## Chunk 4: End-to-End Weave Fixture and Docs

### Task 4: Prove compile applies a middle settings fragment

**Files:**
- Modify: `cmd/weave/main_test.go`
- Modify: `atlas/workflow/weave.md`
- Modify: `workshop/issues/000097-weave-topo-settings-merge.md`

- [ ] **Step 1: Write failing integration test**

Add `TestCompileMergesSettingsAcrossLayerChain` to `main_test.go`.

Fixture:
- `base` manifest declares `merge .claude/settings.base.json .claude/settings.json`.
- `mid` depends on base and declares `merge .claude/settings.mid.json .claude/settings.json`.
- `derived` depends on mid and declares no settings source or declares its own `merge .claude/settings.derived.json .claude/settings.json` if the test needs leaf-source coverage.
- `derived/.claude/settings.local.json` adds a local override/removal.

Expected final `derived/.claude/settings.json`:
- union array contains base, mid, derived/local entries in order.
- scalar from the highest layer/local wins.
- no `$merge_keys` or `$remove` leaks.

Run: `go test ./cmd/weave -run TestCompileMergesSettingsAcrossLayerChain -count=1`

Expected: FAIL before the implementation is wired end to end.

- [ ] **Step 2: Make the integration test pass**

Fix any absolute-vs-relative source handling gaps exposed by the integration test. Do not add a real metis/nous consumer in this issue; the purpose is compiler capability, not a downstream policy decision.

- [ ] **Step 3: Update atlas**

In `atlas/workflow/weave.md`, change the settings backend description from "reads `.claude/settings.ariadne.json` + optional local" to "groups selected merge rows by target and folds ordered sources foundation-first plus optional local."

- [ ] **Step 4: Update issue log and plan checkboxes**

Mark the issue `## Plan` items as complete as work lands. Add log entries with verification commands and ARCH markers where decisions mattered.

- [ ] **Step 5: Run full verification**

Run:

```bash
go test ./cmd/weave/internal/settingsx -count=1
go test ./cmd/weave/internal/plan -count=1
go test ./cmd/weave/internal/golden -count=1
go test ./cmd/weave -count=1
go test ./... 
git diff --check
```

Expected: all pass.

## Execution Notes

- Follow TDD for each task: write a failing test, run it, implement, rerun.
- Do not introduce a second merge implementation in `plan` or `golden`; all semantic comparison must flow through `settingsx.MergeChain` (ARCH-DRY).
- Keep `settings.local.json` as the target-sibling local convention. The issue is about layer-source topology, not changing local settings discovery (Simplicity First).
- Do not add milestone tags unless the work is split into multiple close boundaries. This plan is intended as one close boundary.
