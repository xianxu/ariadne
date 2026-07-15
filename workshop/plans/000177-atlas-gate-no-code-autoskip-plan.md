# atlas gate auto-satisfies on no-code windows — Implementation Plan

> **For agentic workers:** single-pass atomic change (one review boundary — no Mx tags). TDD; checkbox steps.

**Goal:** the close/milestone-close atlas gate auto-satisfies (loud info line) when the review window contains no code paths — a docs/workshop-only delta has no surface to map, so demanding an atlas edit (or a `--no-atlas` acknowledgment) is incoherent there. Correctness fix, not a volume fix (#172 M4 diffstat: ~1 of the 11 observed `--no-atlas` closes was code-free; the rest belong to #174).

**Architecture:** one new arm in the existing gate block (`close.go` `computeClose`, shared by both verbs — a docs-only *milestone* window auto-satisfies too, the case #172's M4 nearly hit). Pure predicate `hasCodePath([]string) bool` with the single docs classifier: `*.md` anywhere, or any path under `workshop/`, `atlas/`, `docs/`; everything else — including `Makefile`, `.gitignore`, extensionless files — conservatively counts as code and keeps the refusal (build files ARE surface). Empty windows auto-satisfy (zero files → no code), which also softens the #174 bookkeeping-re-close edge. The info line must collide with no `GateCatalog` pattern (the #172 instrument must not count auto-satisfactions as gate events) — same guard shape as #178's.

## Tasks

### Task 1: predicate + gate arm (TDD)
**Files:** `cmd/sdlc/close.go`; new `cmd/sdlc/close_atlasskip_test.go`.
- [x] **Step 1: Failing tests** — `hasCodePath` table: `.go`→true; `README.md`→false; `workshop/issues/x.md`→false; `docs/vision/x.md`→false; `Makefile`→true; extensionless `LICENSE`→true; empty list→false; `atlas/x.md`→false (single classifier — atlas is peeled off before the call but the definition stays whole). Info-line format test. Gatesig no-collision test (cinfo-rendered line vs every AckPat/RefusalPat).
- [x] **Step 2–4:** implement the arm (`no code → cinfo auto-satisfied`; `code → existing refusal/bypass`); `go test ./cmd/sdlc/...` PASS.
- [x] **Step 5: Commit** — `#177: atlas gate auto-satisfies on no-code windows`.

### Task 2: docs + verify + close
- [x] Helptext (`close.md` atlas section + `milestone-close.md` if it restates the gate) + `AGENTS.base.md` §5 close bullet ("atlas update" clause gains the docs-only carve-out) + atlas sdlc-binary.md gate description; shadow-sweep `grep -rn "no-atlas"` for other restatements.
- [x] Live verify: a real docs-only window close attempt shows the auto-satisfy line pre-mutation (the omit+no-verified trick from #178); code-window behavior unchanged (existing tests).
- [x] Close: `sdlc close --issue 177 --verified '<evidence>'` (adopt path records the actual).

## ARCH notes
- **ARCH-PURE:** the predicate is pure + table-tested; the gate arm is three lines of glue.
- **ARCH-DRY:** one docs classifier, defined once and named, aligned with the #172 windowstat study's classification.
- **Root Cause / Simplicity:** removes an incoherent demand rather than adding a flag or a config; refusal semantics for code windows untouched.

## Revisions

### 2026-07-14 — executed (both tasks complete)

As planned, no shape deviations. Live verify used a hermetic repo driven
through `milestone-close --dry-run` (runs the real computeClose gates, skips
mutation + judge): a docs-only window printed
"atlas gate: no code surface in window (2 doc/workshop file(s)) — auto-satisfied";
adding one `.go` file to the same window restored the exact old refusal.
Doc sweep: helptext close.md (2 spots) + milestone-close.md, AGENTS.base.md §5
(+ generated AGENTS.md in place), atlas sdlc-binary.md gate paragraph.

### 2026-07-14 — close-review fixes (FIX-THEN-SHIP)

Important #1: the pre-existing swallowed `gitx.DiffNames` error would have
flipped fail-closed → fail-open under the new arm (git failure → nil files →
auto-satisfy); the gate now DIES on a window-diff failure, consistent with the
path's other git failures. The legitimate empty window (nil, nil) still
auto-satisfies. Important #2: the weave-generated entry files (CLAUDE.md et al)
now derive — `weave compile` re-run; they'd been stale since before #178's
sweep (the settings.json half of compile was sandbox-blocked; prose faces
landed). Minors: the gatesig collision guard is now the shared
`assertNoGatesigCollision` helper and matches ANSI-STRIPPED lines (the way the
classifier does) — fixes #178's copy too. Case-sensitive `.md` left as-is
(conservative miss direction).
