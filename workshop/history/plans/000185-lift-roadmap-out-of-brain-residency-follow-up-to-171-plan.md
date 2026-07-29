# Lift Roadmap Out Of Brain Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move roadmap residency from brain-era `data/roadmap` language to the center-of-gravity repo model under `workshop/projects`.

**Architecture:** This is a docs-contract change. The roadmap prototype becomes the source that tells agents where roadmap instances live; the base constitution drops the explicit brain exception; atlas and generated harness entry files are verified as consumers. `ARCH-PURPOSE` drives the scope: do not stop at "no files found in brain" while the contract still says the lift is pending.

**Tech Stack:** Markdown datatype prototypes, weave-generated harness entry files, `sdlc` workflow gates, `rg` verification.

---

## Core Concepts

| Name | Kind | Lives in | Status |
|------|------|----------|--------|
| `RoadmapResidency` | INTEGRATION | `construct/datatype/roadmap.md` | modified |
| `BrainResidencyCharter` | INTEGRATION | `AGENTS.base.md` | modified |
| `DatatypeAtlasRow` | INTEGRATION | `atlas/workflow/data-artifacts.md` | modified |

- **RoadmapResidency** — contract for where roadmap instances live and how agents discover prior roadmaps.
  - **Relationships:** One roadmap belongs to one product/month pair; it resides in the center-of-gravity repo for that product, under `workshop/projects/roadmap/<YYYYMM>/<product>.md`.
  - **DRY rationale:** The datatype prototype is the source that agents apply; derived docs should mirror it instead of inventing a second path (`ARCH-DRY`).
  - **Future extensions:** If parley or `sdlc resolve` gains direct roadmap navigation, it should consume this residency contract rather than hardcoding brain paths.

- **BrainResidencyCharter** — constitution clause that defines brain as capture/measurement only.
  - **Relationships:** Exported from `AGENTS.base.md` into `AGENTS.md`, `CLAUDE.md`, and `GEMINI.md` by weave.
  - **DRY rationale:** The base file is the one source; generated consumers must be recompiled, not hand-maintained.
  - **Future extensions:** New SDLC artifact types should name their non-brain residency here only if they are exceptions worth calling out.

- **DatatypeAtlasRow** — atlas summary of the roadmap type.
  - **Relationships:** Mirrors `construct/datatype/roadmap.md` for humans scanning the data-artifacts map.
  - **DRY rationale:** It is a map, not the contract; keep it short and point at the same path language (`ARCH-PURPOSE`).
  - **Future extensions:** Broader roadmap tooling should get a separate atlas subsection when code exists.

## Integration Points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `WeaveCompile` | `make weave` | existing | generated harness entry files |
| `RoadmapSweep` | terminal verification commands | existing | filesystem search |

- **WeaveCompile** — regenerates `AGENTS.md`, `CLAUDE.md`, and `GEMINI.md` from the edited base constitution.
  - **Injected into:** No code; this is the IO step that proves generated consumers follow the source.
  - **Future extensions:** If the base manifest changes, this remains the regeneration point.

- **RoadmapSweep** — `rg`/`find` checks for stale roadmap residency text and live brain roadmap artifacts.
  - **Injected into:** Verification only.
  - **Future extensions:** Can become a deterministic lint if roadmap residency drifts again.

## Revisions

### 2026-07-28

- Close review returned FIX-THEN-SHIP. Added `Kind` to the Core Concepts table
  and classified the docs-contract entities as INTEGRATION because their
  behavior is verified through repository files and generated consumers. Also
  corrected roadmap/product search recipes so they match the center-of-gravity
  repo residency instead of current-repo-only or data-only searches.

## Chunk 1: Contract And Verification

**Files:**
- Modify: `workshop/issues/000185-lift-roadmap-out-of-brain-residency-follow-up-to-171.md`
- Modify: `construct/datatype/roadmap.md`
- Modify: `AGENTS.base.md`
- Modify: `atlas/workflow/data-artifacts.md`
- Generated: `AGENTS.md`
- Generated: `CLAUDE.md`
- Generated: `GEMINI.md`

- [x] **Step 1: Update the issue spec and estimate**
  - Record the operator decision: roadmaps use center-of-gravity repo residency, under `workshop/projects/roadmap/<YYYYMM>/<product>.md`.
  - Add the required `## Estimate` block using `estimate-logic-v3.1`.

- [x] **Step 2: Update the roadmap datatype prototype**
  - Change the description, authoring instructions, search recipes, and rules from `data/roadmap/...` to `workshop/projects/roadmap/...`.
  - State that the repo is the product's center-of-gravity repo, matching project residency.

- [x] **Step 3: Inventory and handle migration**
  - Run `find /Users/xianxu/workspace/brain -path '*/data/roadmap/*' -print` and `rg -n "^type: roadmap\\b" /Users/xianxu/workspace/brain -g '*.md'`.
  - If files are found, list each source path in the issue Log, determine its product/month from frontmatter and filename, and identify the product's center-of-gravity repo.
  - Before writing any peer repo, follow AGENTS.md peer safety: read that peer's `AGENTS.local.md` and `MEMORY.md`; run `git -C <repo> branch --show-current` and `git -C <repo> status --porcelain`; proceed only when it is on the expected integration branch and clean. If a target peer is dirty or mid-feature, stop and log the deferred migration path instead of writing there.
  - Move each eligible roadmap to `<target-repo>/workshop/projects/roadmap/<YYYYMM>/<product>.md`. Prefer `sdlc migrate` only if it can express the source/target move; otherwise use a named manual move and log the reason. Do not use `git clean` or broad destructive cleanup in any peer.
  - If no files are found, record the migration as a no-op in the issue Log with the exact commands used as evidence.

- [x] **Step 4: Update constitution and atlas docs**
  - Remove the "`roadmap` remains until it too lifts" residual clause from `AGENTS.base.md`.
  - Update `atlas/workflow/data-artifacts.md` so the roadmap row uses the new residency.
  - Search consumers with `rg -n "data/roadmap|workshop/roadmap|type: roadmap|roadmap.*brain" construct/datatype atlas AGENTS.base.md docs workshop/issues` and update only live contract/docs surfaces, not historical records.
  - Verify `rg --files construct/vocabulary | rg 'roadmap'` returns no roadmap vocabulary model; if one appears, update its discovery/residency field in the same pass.

- [x] **Step 5: Regenerate derived harness docs**
  - Run `make weave`.
  - Verify `AGENTS.md`, `CLAUDE.md`, and `GEMINI.md` no longer contain the residual roadmap clause.

- [x] **Step 6: Sweep and verify**
  - Run `rg -n "roadmap.*brain|remains until it too lifts|data/roadmap|workshop/projects/roadmap" AGENTS*.md CLAUDE.md GEMINI.md construct/datatype atlas/workflow workshop/issues/000185-lift-roadmap-out-of-brain-residency-follow-up-to-171.md`.
  - Run `find /Users/xianxu/workspace/brain -path '*/data/roadmap/*' -print` and `rg -n "^type: roadmap\\b" /Users/xianxu/workspace/brain -g '*.md'`.
  - Run `git diff --check`.
