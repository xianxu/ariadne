# Promote `superpowers-writing-plans` as the Canonical Plan Path — Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the adapted `superpowers-writing-plans` skill (landing plans in version-controlled `workshop/plans/`) the constitution's canonical plan path, demoting the Claude Code builtin `EnterPlanMode`'s ephemeral `~/.claude/plans/` file from "the record."

**Architecture:** Mostly constitution + atlas prose (the substance), plus one pure string helper (`planPointer`) printed by `sdlc start-plan` to reinforce the convention at the planning gate. No binary coupling to `~/.claude/plans` — the fallback is *discourage (prose-only)*, keeping the `sdlc` binary agent-agnostic (AGENTS.md §11). Single review boundary (atomic), no `Mx` milestones.

**Tech Stack:** Go (`cmd/sdlc`), Markdown (AGENTS.md, atlas, SKILL.md).

---

## Context

`nous#41` (F6 of the 2026-06-02 sdlc retro) stranded a plan in `~/.claude/plans/` — harness-controlled, ephemeral, not version-controlled — because the agent's default "plan mode" is the builtin `EnterPlanMode`. ariadne already ships `construct/adapted/superpowers-writing-plans/`, which already saves to `workshop/plans/`. So this is **promotion, not invention** (`ARCH-DRY`): name the skill as canonical so durable plans land in-repo by default. This plan dogfoods that path — it *is* a `workshop/plans/` file authored via the skill (the very `~/.claude/plans/validated-puzzling-map.md` scratch copy from this session is the throwaway the issue demotes).

## Scope Check

Single subsystem: the planning-entry gate (`sdlc start-plan`) + the constitution/atlas prose that points at it. Not multi-subsystem; one plan, one review boundary.

## Core concepts

### Pure entities (the conceptual core)

| Name | Lives in | Status |
|------|----------|--------|
| `planPointer` | `cmd/sdlc/startplan.go` | new |

- **planPointer** — pure function `planPointer(issue int) string` rendering the one durable-plan reminder line(s): names the `superpowers-writing-plans` skill, the `workshop/plans/NNNNNN-slug-plan.md` location (zero-padded id when `issue > 0`, else the `NNNNNN` placeholder), and that the builtin `~/.claude/plans/…` file is ephemeral — not the record.
  - **Relationships:** 1:1 with a `start-plan` invocation; sits beside the existing pure helpers `baseContentionSummary` / `issueRef` in the same file.
  - **DRY rationale:** First occurrence of "where the durable plan lives" as binary output; the path grammar (`NNNNNN-slug-plan.md`) is the same one AGENTS.md §1 + SKILL.md state — one convention, three statements of it, pinned by the test asserting the literal.
  - **Future extensions:** If the durable-plan location ever becomes configurable (a flag / config key), this is the single render site to widen.

**Test surface:** colocated `TestPlanPointer` in `cmd/sdlc/startplan_test.go`, no IO mocks (mirrors `TestBaseContentionSummary` / `TestIssueRef`). Table cases: `issue > 0` embeds `000072` + `workshop/plans/` + skill name; `issue == 0` embeds the `NNNNNN` placeholder + skill name.

### Integration points (where pure meets the world)

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `runStartPlan` stdout | `cmd/sdlc/startplan.go` | modified | process stdout |

- **runStartPlan** — the existing thin IO seam that prints the planning-entry payload. The only change: one `fmt.Fprintln(stdout, planPointer(issue))` inserted **between** the architecture block (line 57) and the contention read (line 59). No new IO; the pure `planPointer` is injected into the existing print site.
  - **Injected into:** receives `planPointer`'s string; ordering is *what (architecture) → how/where (pointer) → environment (contention)*.
  - **Future extensions:** none beyond the pointer; the seam already exists.

**No external-service integration** → no process-level fake needed. The remaining changes are pure prose (no entities): AGENTS.md §1/§2, SKILL.md path grammar, start-plan helptext, three atlas files.

---

## Task 1: `planPointer` — pure helper + test (TDD)

**Files:**
- Modify: `cmd/sdlc/startplan.go` (add `planPointer`; call it in `runStartPlan` between line 57 and line 59)
- Test: `cmd/sdlc/startplan_test.go` (add `TestPlanPointer`)

- [ ] **Step 1: Write the failing test** in `startplan_test.go`

```go
func TestPlanPointer(t *testing.T) {
	withIssue := planPointer(72)
	if !strings.Contains(withIssue, "superpowers-writing-plans") {
		t.Errorf("planPointer(72) should name the skill; got:\n%s", withIssue)
	}
	if !strings.Contains(withIssue, "workshop/plans/000072-") {
		t.Errorf("planPointer(72) should embed the zero-padded durable path; got:\n%s", withIssue)
	}
	if !strings.Contains(withIssue, "~/.claude/plans") {
		t.Errorf("planPointer(72) should demote the ephemeral builtin file; got:\n%s", withIssue)
	}
	noIssue := planPointer(0)
	if !strings.Contains(noIssue, "workshop/plans/NNNNNN-") {
		t.Errorf("planPointer(0) should use the NNNNNN placeholder; got:\n%s", noIssue)
	}
	if !strings.Contains(noIssue, "superpowers-writing-plans") {
		t.Errorf("planPointer(0) should name the skill; got:\n%s", noIssue)
	}
}
```

- [ ] **Step 2: Run it, verify it fails**

Run: `go test ./cmd/sdlc/ -run TestPlanPointer`
Expected: FAIL — `undefined: planPointer`.

- [ ] **Step 3: Implement `planPointer`** in `startplan.go` (beside `issueRef`)

```go
// planPointer renders the durable-plan reminder for the planning gate (#72):
// the canonical plan lands in workshop/plans/ (version-controlled), authored via
// the superpowers-writing-plans skill — NOT the harness builtin's ephemeral
// ~/.claude/plans/ file. Pure: the only input is the issue id (for the slug),
// so the wording is table-testable without IO. Stays agent-agnostic — names the
// skill + repo location, never teaches the binary the Claude-specific path.
func planPointer(issue int) string {
	slug := "NNNNNN-slug"
	if issue > 0 {
		slug = fmt.Sprintf("%06d-slug", issue)
	}
	return fmt.Sprintf("Capture the plan via the superpowers-writing-plans skill →\n"+
		"    workshop/plans/%s-plan.md (version-controlled). The builtin plan-mode\n"+
		"    file (~/.claude/plans/…) is ephemeral — NOT the record.", slug)
}
```

- [ ] **Step 4: Wire it into `runStartPlan`** — insert between the architecture block and the contention read

In `startplan.go`, after `fmt.Fprintln(stdout, judge.ArchitectureBlock("at-plan"))` (line 57) and before the `// #82 M3 / #83` contention block (line 59):

```go
	fmt.Fprintln(stdout)
	cinfo(stdout, planPointer(issue))
```

(Use `cinfo` to match the framing line's `==>` styling; the pointer reads as an instruction, like the opening line.)

- [ ] **Step 5: Run the test, verify it passes**

Run: `go test ./cmd/sdlc/ -run TestPlanPointer`
Expected: PASS.

- [ ] **Step 6: Full package + vet**

Run: `go test ./cmd/sdlc/... && go vet ./cmd/sdlc/...`
Expected: all green; existing `TestRunStartPlan_RendersAtPlanLens` still passes (it asserts ARCH content, unaffected by the new line).

- [ ] **Step 7: Commit**

```bash
git add cmd/sdlc/startplan.go cmd/sdlc/startplan_test.go
git commit -m "#72: start-plan prints durable-plan pointer (superpowers-writing-plans → workshop/plans/)"
```

## Task 2: Constitution + skill + helptext prose

**Files:**
- Modify: `AGENTS.md` (§2 plan-mode sentence + Entering-planning bullet; §1 plans bullet)
- Modify: `construct/adapted/superpowers-writing-plans/SKILL.md` (line 18 path grammar)
- Modify: `cmd/sdlc/helptext/start-plan.md` (OUTPUT section)

- [ ] **Step 1: AGENTS.md §2** — rewrite the "Non-trivial task … → plan mode, wait for approval" sentence so "plan mode" means *design via `superpowers-writing-plans`, landing the durable plan in `workshop/plans/NNNNNN-slug-plan.md`*; add the clarifier that the builtin plan-mode's `~/.claude/plans/` file is ephemeral — NOT the record. In the "Entering planning: run `sdlc start-plan`" bullet, note the `(design)` step = author the plan via the skill.

- [ ] **Step 2: AGENTS.md §1** — on the "complex → add `workshop/plans/NNNNNN-slug-plan.md`" bullet, name `superpowers-writing-plans` as the producer (the canonical plan path, cross-ref §2).

- [ ] **Step 3: SKILL.md line 18** — `Save plans to: workshop/plans/<slug>-plan.md` → `workshop/plans/NNNNNN-<slug>-plan.md` (one path grammar across constitution + skill, `ARCH-DRY`).

- [ ] **Step 4: helptext/start-plan.md** — add one line to the OUTPUT section noting start-plan now also points at the durable-plan skill + `workshop/plans/` location (so `--help` matches the new stdout).

- [ ] **Step 5: Verify the AGENTS.md drift guard still passes** (it checks ARCH-* narrative sync, unaffected, but run to be safe)

Run: `go test ./cmd/sdlc/internal/judge/ -run AGENTS`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add AGENTS.md construct/adapted/superpowers-writing-plans/SKILL.md cmd/sdlc/helptext/start-plan.md
git commit -m "#72: AGENTS.md names superpowers-writing-plans as canonical plan path; align SKILL.md slug grammar"
```

## Task 3: Atlas

**Files:**
- Modify: `atlas/workflow/issue-lifecycle.md` (flow diagram, ~line 6)
- Modify: `atlas/workflow/sdlc-binary.md` (start-plan section)
- Modify: `atlas/workflow/artifact-hierarchy.md` (workshop/plans producer)

- [ ] **Step 1: issue-lifecycle.md** — insert `start-plan` into the `claim → … → change-code` flow; name `superpowers-writing-plans` as the durable-plan producer.

- [ ] **Step 2: sdlc-binary.md** — note `start-plan` now also emits the durable-plan pointer (alongside the at-plan lens + contention).

- [ ] **Step 3: artifact-hierarchy.md** — name `superpowers-writing-plans` as the producer of `workshop/plans/` files.

- [ ] **Step 4: Commit**

```bash
git add atlas/
git commit -m "#72: atlas — start-plan in the flow + superpowers-writing-plans as workshop/plans producer"
```

## Verification (end-to-end, before close)

- [ ] `go test ./cmd/sdlc/...` green (incl. `TestPlanPointer`); `go vet ./cmd/sdlc/...` clean.
- [ ] `go run ./cmd/sdlc start-plan --issue 72` (or the built binary) prints, in order: framing → ARCHITECTURE PRINCIPLES block → **durable-plan pointer naming the skill + `workshop/plans/000072-slug-plan.md`** → base-contention line.
- [ ] Read AGENTS.md §1/§2 top-to-bottom: "plan mode" unambiguously means `superpowers-writing-plans` → `workshop/plans/`; the ephemeral builtin file is explicitly demoted.
- [ ] `sdlc close --issue 72 --actual <h> --verified '<evidence>'` — the mandatory fresh-eyes boundary review runs here; fix Critical/Important before crossing.

## Out of scope
- No binary coupling to `~/.claude/plans` (rejected fallback — agent-agnosticism, §11).
- No `sdlc plan import` bridge verb (YAGNI; revisit only if the ephemeral-plan bridge is actually hit).
