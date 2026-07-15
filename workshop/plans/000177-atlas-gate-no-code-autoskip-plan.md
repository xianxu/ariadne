# atlas gate auto-satisfies on no-code windows — Implementation Plan

> **For agentic workers:** single-pass atomic change (one review boundary — no Mx tags). TDD; checkbox steps.

**Goal:** the close/milestone-close atlas gate auto-satisfies (loud info line) when the review window contains no code paths — a docs/workshop-only delta has no surface to map, so demanding an atlas edit (or a `--no-atlas` acknowledgment) is incoherent there. Correctness fix, not a volume fix (#172 M4 diffstat: ~1 of the 11 observed `--no-atlas` closes was code-free; the rest belong to #174).

**Architecture:** one new arm in the existing gate block (`close.go` `computeClose`, shared by both verbs — a docs-only *milestone* window auto-satisfies too, the case #172's M4 nearly hit). Pure predicate `hasCodePath([]string) bool` with the single docs classifier: `*.md` anywhere, or any path under `workshop/`, `atlas/`, `docs/`; everything else — including `Makefile`, `.gitignore`, extensionless files — conservatively counts as code and keeps the refusal (build files ARE surface). Empty windows auto-satisfy (zero files → no code), which also softens the #174 bookkeeping-re-close edge. The info line must collide with no `GateCatalog` pattern (the #172 instrument must not count auto-satisfactions as gate events) — same guard shape as #178's.

## Tasks

### Task 1: predicate + gate arm (TDD)
**Files:** `cmd/sdlc/close.go`; new `cmd/sdlc/close_atlasskip_test.go`.
- [ ] **Step 1: Failing tests** — `hasCodePath` table: `.go`→true; `README.md`→false; `workshop/issues/x.md`→false; `docs/vision/x.md`→false; `Makefile`→true; extensionless `LICENSE`→true; empty list→false; `atlas/x.md`→false (single classifier — atlas is peeled off before the call but the definition stays whole). Info-line format test. Gatesig no-collision test (cinfo-rendered line vs every AckPat/RefusalPat).
- [ ] **Step 2–4:** implement the arm (`no code → cinfo auto-satisfied`; `code → existing refusal/bypass`); `go test ./cmd/sdlc/...` PASS.
- [ ] **Step 5: Commit** — `#177: atlas gate auto-satisfies on no-code windows`.

### Task 2: docs + verify + close
- [ ] Helptext (`close.md` atlas section + `milestone-close.md` if it restates the gate) + `AGENTS.base.md` §5 close bullet ("atlas update" clause gains the docs-only carve-out) + atlas sdlc-binary.md gate description; shadow-sweep `grep -rn "no-atlas"` for other restatements.
- [ ] Live verify: a real docs-only window close attempt shows the auto-satisfy line pre-mutation (the omit+no-verified trick from #178); code-window behavior unchanged (existing tests).
- [ ] Close: `sdlc close --issue 177 --verified '<evidence>'` (adopt path records the actual).

## ARCH notes
- **ARCH-PURE:** the predicate is pure + table-tested; the gate arm is three lines of glue.
- **ARCH-DRY:** one docs classifier, defined once and named, aligned with the #172 windowstat study's classification.
- **Root Cause / Simplicity:** removes an incoherent demand rather than adding a flag or a config; refusal semantics for code windows untouched.
