# Explicit Operating Constraints Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `ARCH-CONSTRAINTS` as a concise, test-guarded architecture lens that makes material runtime budgets and resource bounds explicit during planning and review.

**Architecture:** Extend the existing embedded Markdown registry; do not add a second configuration object or delivery path. Existing pure marker extraction and registry rendering will carry the new entry into CLI and judge consumers, while focused contract assertions and generated prompt goldens prove both semantics and fan-out.

**Tech Stack:** Markdown registry and atlas, Go 1.x embedded resources and tests, Cobra CLI prompt delivery.

---

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `OperatingEnvelopePrinciple` | `cmd/sdlc/internal/judge/architecture.md` | new |
| `ArchitectureMarkerSet` | `cmd/sdlc/internal/judge/architecture.go` | modified |
| `ArchitectureBlock` | `cmd/sdlc/internal/judge/architecture.go` | modified |

- **`OperatingEnvelopePrinciple`** — the declarative `ARCH-CONSTRAINTS` entry: classify the runtime path, state consequential bounds and their basis, confirm material uncertainty, and define bounded behavior outside the envelope.
  - **Relationships:** N:1 with the registry (one of several principles); 1:N with every rendered architecture block and marker-aware review checklist.
  - **DRY rationale:** The principle body exists only in the registry; consumers receive it through the existing embed/render graph (`ARCH-DRY`, `ARCH-PURPOSE`).
  - **Future extensions:** Domain examples may widen in the same entry when repeated review evidence shows a missing workload class; no per-domain principle hierarchy is introduced now.
- **`ArchitectureMarkerSet`** — the ordered marker list extracted by `ArchitectureMarkers`; its contract gains `ARCH-CONSTRAINTS` without changing the extraction algorithm.
  - **Relationships:** N:1 with the registry; 1:N with boundary-review marker expansion and tests.
  - **DRY rationale:** Marker enumeration continues to derive from headings rather than a production allowlist.
  - **Future extensions:** Further registry headings flow through the same extraction path.
- **`ArchitectureBlock`** — the pure rendered registry plus lens-specific header; its entry count and body change automatically with the registry.
  - **Relationships:** 1:1 with a selected lens per render; 1:N with CLI and judge consumers.
  - **DRY rationale:** Both push and pull delivery reuse the same renderer.
  - **Future extensions:** None for #205; lens vocabulary remains `at-plan` / `at-review`.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `ArchitecturePrincipleDelivery` | `cmd/sdlc/archprinciples.go`, `cmd/sdlc/startplan.go`, `cmd/sdlc/internal/judge/prompts.go`, `cmd/sdlc/internal/judge/review.go` | modified | CLI stdout and fresh agent contexts |
| `ArchitecturePromptGoldens` | `cmd/sdlc/internal/judge/testdata/golden/*.prompt` | modified | byte-level rendered prompt contract |

- **`ArchitecturePrincipleDelivery`** — the existing delivery graph exposes the new registry entry through `sdlc arch-principles`, `sdlc start-plan`, plan-quality, milestone review, DRY/PURE review prompts, and boundary-review marker expansion.
  - **Injected into:** No new injection seam; all consumers already call `ArchitectureBlock` or `ArchitectureMarkers`.
  - **Future extensions:** New consumers must reuse those two existing primitives.
- **`ArchitecturePromptGoldens`** — intentional snapshots of the four architecture-aware generated prompts (`dry`, `pure`, `plan-quality`, `milestone-review`).
  - **Injected into:** `TestBuildPrompt_Golden`; they are regenerated only because the source registry deliberately changes, then reviewed as derived output.
  - **Future extensions:** The existing `AllInjectedCategories` enumeration governs prompt coverage; no new golden mechanism.

No external binary or service is added, so `ARCH-MOCK` requires no fake. The change is declarative and uses existing pure renderers, so `ARCH-PURE` requires no new IO seam.

## Operating envelope (`ARCH-CONSTRAINTS`)

| Parameter | Budget/range | Basis | Behavior when exceeded |
|-----------|--------------|-------|------------------------|
| Registry prompt overhead | One entry, no more than 350 words and comparable to existing principle entries | Operator requested a concise, routinely checked principle; registry is injected into several agent contexts | Remove illustrative prose before weakening required semantics |
| Runtime work | Existing static embed, regex extraction, and string rendering only; no filesystem scan, subprocess, network call, or new loop over project data | Existing delivery architecture and startup/UI lessons from pair #156 | Reject any implementation that adds runtime discovery or blocking work |
| Delivery coverage | Every current architecture-aware CLI and judge seam | `ARCH-PURPOSE` single-source contract and #205 Done-when | Treat any missing consumer/test as incomplete, not follow-up work |
| Test parallelism | At most 20 Go workers and one Go test command at a time | Operator-selected limit for the shared M2 Max development machine | Queue test phases sequentially with `-p 20` |

Non-goals:

- Do not define universal latency or concurrency defaults; budgets remain workload- and operator-specific.
- Do not build a typed constraint schema, parser, linter, benchmark harness, or new judge category.
- Do not restate the principle body in AGENTS.md, help text, prompt templates, or atlas; those surfaces route to or describe the registry.
- Do not alter `ArchitectureMarkers`, `ArchitectureBlock`, or prompt plumbing algorithms; their existing derivation is the mechanism this change exercises.

## Chunk 1: Registry contract and derived delivery

### Task 1: Guard the new semantics and marker fan-out

**Files:**

- Modify: `cmd/sdlc/internal/judge/judge_test.go`
- Modify: `cmd/sdlc/archprinciples_test.go`
- Modify: `cmd/sdlc/startplan_test.go`

- [ ] **Step 1: Write the failing registry contract test**

Add a test-only section extractor and `TestArchitectureRegistry_ConstraintsContract` beside `TestArchitectureRegistry_Content`. Scope every assertion to the `ARCH-CONSTRAINTS` entry (from its heading to the next `## ARCH-` heading or EOF), so wording from another principle cannot mask a deletion. Assert small stable phrases for the complete required semantic contract, not the whole paragraph:

```go
func architectureEntry(t *testing.T, marker string) string {
	t.Helper()
	heading := "## " + marker
	start := strings.Index(ArchitectureRegistry, heading)
	if start < 0 {
		t.Fatalf("ArchitectureRegistry missing %s heading", marker)
	}
	entry := ArchitectureRegistry[start:]
	rest := entry[len(heading):]
	if next := strings.Index(rest, "\n## ARCH-"); next >= 0 {
		entry = entry[:len(heading)+next]
	}
	return entry
}

func TestArchitectureRegistry_ConstraintsContract(t *testing.T) {
	entry := architectureEntry(t, "ARCH-CONSTRAINTS")
	for _, want := range []string{
		"operating envelope",
		"workload and interaction path",
		"latency",
		"workload and growth",
		"CPU",
		"memory",
		"disk/network IO",
		"concurrency",
		"target environment and co-tenancy",
		"overload behavior",
		"keystroke",
		"UI response",
		"startup/shutdown",
		"online request",
		"batch",
		"training/inference",
		"budget/range",
		"measured fact",
		"requirement",
		"domain-informed assumption",
		"operator choice",
		"N/A",
		"universal defaults",
		"material uncertainty",
		"operator",
		"bounded behavior",
		"representative measurements or tests",
		"blocking optional work",
		"unbounded concurrency or fan-out",
		"repeated expensive work",
		"cached or incremental",
		"resource monopolization",
		"unsupported performance claims",
		"outside the stated bounds",
	} {
		if !strings.Contains(entry, want) {
			t.Errorf("ARCH-CONSTRAINTS contract missing %q", want)
		}
	}
}
```

Also append `ARCH-CONSTRAINTS` to the existing exact expectations in `TestArchitectureRegistry_Content`, `TestArchitectureMarkers`, and `TestCodeReviewBody_Renders`. This pins the source, ordered marker extraction, and the complete boundary-review checklist.

- [ ] **Step 2: Guard both CLI delivery paths**

Append `ARCH-CONSTRAINTS` to `TestRunArchPrinciples_RendersRegistry` and `TestRunStartPlan_RendersAtPlanLens`. The former pins the standalone pull path; the latter pins the planning-entry push path.

- [ ] **Step 3: Run the focused tests and verify RED**

Run one test process with the operator's concurrency ceiling:

```bash
go test -p 20 ./cmd/sdlc/internal/judge ./cmd/sdlc -run 'TestArchitectureRegistry_(Content|ConstraintsContract|EmbeddedInPrompts)|TestArchitectureMarkers|TestCodeReviewBody_Renders|TestRunArchPrinciples_RendersRegistry|TestRunStartPlan_RendersAtPlanLens' -count=1
```

Expected: FAIL because `ARCH-CONSTRAINTS` and its semantic phrases are absent from the registry and therefore absent from both CLI outputs and marker expansion.

### Task 2: Add the single-source principle

**Files:**

- Modify: `cmd/sdlc/internal/judge/architecture.md`

- [ ] **Step 1: Add the concise registry entry**

Append this entry after `ARCH-MOCK`, preserving the registry's `principle` / `at-plan` / `at-review` shape and the 350-word envelope:

```markdown
## ARCH-CONSTRAINTS — Design to an explicit operating envelope

- **principle:** Runtime behavior is part of the architecture. Before choosing a
  mechanism, identify the small set of external constraints that can materially
  shape it: latency, workload and growth, CPU, memory, disk/network IO,
  concurrency, target environment and co-tenancy, and overload behavior. Make
  consequential expectations explicit instead of leaving them as hidden
  assumptions.
- **at-plan:** Classify the workload and interaction path (for example keystroke,
  UI response, startup/shutdown, online request, batch, or training/inference),
  then give each relevant constraint a budget/range, basis (measured fact,
  requirement, domain-informed assumption, or operator choice), and bounded
  behavior when exceeded. Mark irrelevant categories `N/A`; do not fill a
  ceremonial checklist or invent universal defaults. Make an educated initial
  estimate when useful, but confirm material uncertainty with the operator.
- **at-review:** Check that the implementation enforces the declared operating
  envelope and that representative measurements or tests exercise the relevant
  environment and workload. Flag blocking optional work on a critical UI path,
  unbounded concurrency or fan-out, repeated expensive work that should be
  cached or incremental, resource monopolization, unsupported performance
  claims, and behavior that silently operates outside the stated bounds.
```

- [ ] **Step 2: Run the focused tests and verify GREEN**

Run the exact command from Task 1 Step 3.

Expected: PASS. `ArchitectureMarkers` discovers the fifth heading, `ArchitectureBlock` reports five entries, both CLI paths render the body, and the boundary procedure expands the complete marker list.

- [ ] **Step 3: Commit the tested registry contract**

```bash
git add cmd/sdlc/internal/judge/architecture.md cmd/sdlc/internal/judge/judge_test.go cmd/sdlc/archprinciples_test.go cmd/sdlc/startplan_test.go
git commit -m "architecture: #205 add explicit operating constraints"
```

### Task 3: Refresh derived prompt evidence and architecture maps

**Files:**

- Modify: `cmd/sdlc/internal/judge/testdata/golden/dry.prompt`
- Modify: `cmd/sdlc/internal/judge/testdata/golden/pure.prompt`
- Modify: `cmd/sdlc/internal/judge/testdata/golden/plan-quality.prompt`
- Modify: `cmd/sdlc/internal/judge/testdata/golden/milestone-review.prompt`
- Modify: `atlas/workflow/architecture-principles.md`
- Modify: `atlas/workflow/sdlc-binary.md`
- Modify: `atlas/index.md`

- [ ] **Step 1: Deliberately regenerate prompt goldens from the changed registry**

```bash
go test -p 20 ./cmd/sdlc/internal/judge -run '^TestBuildPrompt_Golden$' -update-golden -count=1
```

Expected: PASS. Review `git diff --name-only -- cmd/sdlc/internal/judge/testdata/golden`; exactly the four architecture-aware prompt files listed above change. Inspect their diffs to confirm the fifth entry and complete marker enumeration derive from the registry; do not hand-edit generated prompt bodies.

- [ ] **Step 2: Update the architecture atlas**

In `atlas/workflow/architecture-principles.md`, add a concise `ARCH-CONSTRAINTS` map and extend the key-consumer list to name `sdlc start-plan` as well as the pull command. In the `Architecture principles (#75)` section of `atlas/workflow/sdlc-binary.md`, record the fifth principle and its explicit operating-envelope purpose. Update the `atlas/index.md` Architecture Principles row so it no longer spotlights only `ARCH-MOCK`.

- [ ] **Step 3: Run the package and golden verification**

```bash
go test -p 20 ./cmd/sdlc/internal/judge ./cmd/sdlc -count=1
```

Expected: PASS, including `TestBuildPrompt_Golden` without the update flag.

- [ ] **Step 4: Commit derived evidence and maps**

```bash
git add cmd/sdlc/internal/judge/testdata/golden atlas/workflow/architecture-principles.md atlas/workflow/sdlc-binary.md atlas/index.md
git commit -m "architecture: #205 map operating constraint delivery"
```

### Task 4: Verify the complete single-pass change

**Files:**

- Modify: `workshop/issues/000205-make-operating-constraints-explicit.md`
- Modify: `workshop/plans/000205-make-operating-constraints-explicit-plan.md`

- [ ] **Step 1: Run the full Go test suite**

```bash
go test -p 20 ./... -count=1
```

Expected: PASS with no more than 20 package workers and no concurrent Go test runner.

- [ ] **Step 2: Run static verification**

```bash
go vet -p 20 ./...
git diff --check
```

Expected: both commands exit 0.

- [ ] **Step 3: Smoke both operator-facing lenses**

```bash
go run ./cmd/sdlc arch-principles
go run ./cmd/sdlc arch-principles --lens at-review
```

Expected: both outputs name `ARCH-CONSTRAINTS`; the first foregrounds `at-plan`, the second `at-review`, and each header reports five entries.

- [ ] **Step 4: Update durable execution state and commit**

Check every completed issue/plan checkbox and add a dated `## Log` entry with the registry, consumer sweep, golden, atlas, and verification evidence. If execution changes any approved plan detail, append a timestamped `## Revisions` section rather than rewriting the original plan claim.

```bash
git add workshop/issues/000205-make-operating-constraints-explicit.md workshop/plans/000205-make-operating-constraints-explicit-plan.md
git commit -m "architecture: #205 record operating constraint verification"
```

- [ ] **Step 5: Cross the single review boundary**

Run:

```bash
sdlc close --issue 205 --verified 'ARCH-CONSTRAINTS registry contract, derived CLI/judge delivery, prompt goldens, atlas maps, go test -p 20 ./... -count=1, go vet -p 20 ./..., CLI at-plan/at-review smoke, and git diff --check all verified'
```

Expected: the binary measures actual time and dispatches the mandatory fresh-context whole-issue review. Resolve any Critical/Important finding by class before accepting the boundary verdict; do not run a second manual review at this SDLC boundary.
