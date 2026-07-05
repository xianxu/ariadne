---
id: 000144
status: codecomplete
deps: []
created: 2026-06-29
updated: 2026-07-05
started: 2026-07-05T10:18:05-07:00
estimate_hours: 1.8
actual_hours: 2.0
---

# sdlc resolve — read-only artifact-reference resolver

## Problem

ariadne artifacts (issues, plans, review sidecars, targets) refer to each other —
and across peer repos — with **symbolic** refs (`ariadne#11`, `#15 M4`, `pair#84`),
but there's no mechanism to turn a ref into the file it names. Consumers (a human in
parley, an agent, the CLI) each have to glob/guess the path.

This issue **was** framed as "files should carry stored cross-links" — rejected:
the id is stable but the path is not (slug renames; `issues/ → history/` on
close/merge, which ariadne#160 made happen on every merge), so stored links rot on
archive. The fix is **read-time resolution** — keep the symbolic ref canonical and
resolve the path on demand. This issue is the **ariadne slice** of that: a
read-only `sdlc resolve`. The editor UX is **parley#160** (`navigate ariadne
artifact references`), which shells to this.

## Spec

Add a **read-only** `sdlc resolve <ref>` that maps a symbolic ref → the current
file path(s).

- **Read-only ⟹ no git transaction lock.** Unlike mutating verbs (`issue new`,
  `close`), it never takes `.git/sdlc.lock`, so it's not subject to lock-contention
  slowness — cost is just process spawn + a glob. This is the property that makes it
  acceptable to route parley navigation through the binary (single source of truth)
  rather than re-implementing a resolver in Lua.
- **Derive locations from the models, don't hardcode.** Use the `discovery:` blocks
  in the vocab/datatype models (parley already sources the issue home from
  `issue.cue`, ariadne#116) so resolution tracks ariadne's structure — incl. the
  `issues/ → history/` mirror and `plans/` — and can't drift.
- **Single-source the ref grammar** here so parley + agents can't diverge:
  - `repo#id` → that repo's workshop issue (`<parent>/<repo>/workshop/{issues,history}/<id>-*.md`).
  - `#id` → the current repo.
  - `repo#id Mx` → the issue + milestone context (the `Mx` row / review sidecar).
  - By 6-digit id, resolve the whole **family**: issue + `<id>-*-plan.md` +
    `<id>-*-m*-review.md` (this is the original "interlink" ask — surfaced as a
    resolvable set, not stored links).
  - Disambiguate the GitHub inbox from the workshop tracker (sdlc already splits
    `--issue` vs `--github-issue`) — pick a form for GitHub refs (e.g. `repo gh#id`).
- Output: path(s), machine-readable (for parley/agents) — consider `--json` and a
  human `sdlc open <ref>` sugar that opens `$EDITOR`.

## Done when

- [x] `sdlc resolve ariadne#11` prints the current path, correct after the file
      archived (`issues/ → history/`) and across sibling repos. *(verified: `#160`
      archived family, `parley#160` cross-repo)*
- [x] It's read-only (no lock; provable by running under a held lock). *(structural
      `commandNeedsRepoLock`==false + runtime `repolock.Acquire` held-lock test)*
- [x] Locations + grammar derive from the models (no hardcoded artifact paths).
      *(`issue.cue` discovery + `vocab.Discovery()`; grammar = `parseRef`)*
- [x] Resolving by id returns the family (issue + plan + reviews).
- [x] Grammar is documented/single-sourced so parley#160 + agents consume one spec.
      *(`parseRef` is the sole parser; `helptext/resolve.md` + doc-drift test)*

## Plan

- [x] Define the ref grammar (single source — `parseRef`, the sole parser).
- [x] `sdlc resolve` reading `discovery:` from the models; family resolution by id.
- [x] Read-only guarantee (no lock) + a test proving it resolves under a held lock.
- [x] `--json` output + `sdlc open` sugar.
- [x] Tests: archived-file resolution, cross-repo, milestone/family refs.

Full durable plan: `workshop/plans/000144-sdlc-resolve-plan.md` (2 milestones,
TDD, fresh-eyes reviewed).

## Estimate

Best guess: ~1.8 hr (ship wall-clock, AI-paired).

Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: greenfield-go-module   design=0.5 impl=0.32
item: smaller-go-module      design=0.1 impl=0.18
item: atlas-docs             design=0.1 impl=0.15
item: milestone-review       design=0.0 impl=0.18
item: milestone-review       design=0.0 impl=0.18
design-buffer: 0.15
total: 1.82
```

Derivation (design kept from v2; impl = 40% of the v2/v2.1 primitive-table impl
per estimate-logic-v3.1; +15% design buffer for a thorough plan doc):
- **greenfield-go-module** — `resolve.go`: `parseRef` + `classifyFamily` + the two
  IO seams (`resolveRepoDir`, `familyFiles`) + `resolve`/`open` commands + ~10 tests.
  The meaty new module. design 0.5 (parser edge cases carry real design density
  even with the plan); impl 0.32 (top of the greenfield band × 0.4).
- **smaller-go-module** — `pkg/vocab` `Discovery()` accessor + `issue.cue` discovery
  extension + JSON regen. Extend-existing.
- **atlas-docs** — `helptext/resolve.md` (the single-source grammar doc) + `open.md`
  + `atlas/workflow/sdlc-binary.md`.
- **milestone-review ×2** — the M1 + M2 fresh-context boundary reviews (process overhead).

recomputed = Σdesign(0.7)×1.15 + Σimpl(1.01)×1.0 = 0.805 + 1.01 = 1.815 ≈ total 1.82.

## Log

### 2026-06-29

### 2026-07-03 — reframed (NOT wontfix): this is the ariadne `sdlc resolve` slice

Design discussion with the operator: the original "files carry stored cross-links"
premise is rejected (link-rot on archive). The replacement — **read-time resolution**
— keeps the symbolic ref canonical and resolves on demand. The editor UX lives in
parley (parley#160), but the **resolver is ariadne work** (`sdlc resolve`, base-layer
Go). Operator caught that I'd wrongly wontfix'd this: since sdlc owns the resolver,
this issue stays as the ariadne slice. Reframed off the stored-link premise onto
`sdlc resolve`; parley#160 `deps: [ariadne#144]`.

### 2026-07-05 — implemented `sdlc resolve` / `sdlc open`
- 2026-07-05: closed — sdlc resolve/open implemented + verified end-to-end (go build): resolve #144 -> issue+plan (active); #160 -> full 6-file family in history/ (archive-correct, even with divergent plan slug); parley#160 -> cross-repo parley.nvim; #160 M2 -> m2-review; --json #144 structured; gh#7 -> github label; open #160 M3 -> m3-review; distinct errors for unknown-id/missing-milestone/bad-ref. Lock-free proven structurally (commandNeedsRepoLock==false) + at runtime (held repolock.Acquire). go test ./... green. ACTUAL: tool suggested 7.88h is a known multi-day-window over-attribution (#92/#162) spanning 29 issues incl 2026-07-02/03 cross-session activity; scoped to today (claim 10:18 -> HEAD 12:27 ~2.2h wall-clock, nearly all active, background-agent spans counted per #118) the measured actual is ~2.0h, validating the 1.8h estimate.; review verdict: SHIP

Durable plan `workshop/plans/000144-sdlc-resolve-plan.md`, fresh-eyes reviewed
(plan-quality: CLEAN; a codebase-assumption sweep caught several plan errors —
`make vocab-embed` not bare `go generate`, `vocabulary vet` not `cue vet ./...`,
no `writeJSON`/`acquireTestLock` helpers — all folded in before coding).

**Structure** (single close boundary; M1/M2 kept as logical phases — see plan Revisions):
- **Model:** extended `issue.cue` `discovery:` with `archive` (`workshop/history`) +
  `plans` (`workshop/plans`); regen `pkg/vocab/issue.json`; `vocab.Discovery()` accessor.
  So resolve derives every location from the model — zero hardcoded artifact paths (ARCH-DRY).
- **Pure core:** `parseRef` (the SINGLE-SOURCE grammar parser — parley#160 shells to it,
  never re-encodes) + `classifyFamily` (issue→plan→reviews ordering). No IO.
- **IO shell:** `resolveRepoDir` (sibling repo: exact basename then unique-prefix, so
  `parley`→`parley.nvim`), `familyFiles` (model-derived 3-dir glob, archive-inclusive).
- **Commands:** `sdlc resolve` (family output / `--json` / `Mx` narrow / `gh#` label) +
  `sdlc open` sugar; grammar in `helptext/resolve.md`; a doc-drift test binds every
  documented example to `parseRef`.

**Verification (real repo, `go build`):**
- `resolve '#144'` → issue + plan (active). ✓
- `resolve '#160'` → the 6-file family in `workshop/history/` (issue+plan+M1/M2/M3+close),
  correct after archiving AND despite #160's plan slug differing from its issue slug. ✓
- `resolve 'parley#160'` → `../parley.nvim/workshop/issues/000160-*.md` (cross-repo). ✓
- `resolve '#160 M2'` → the m2-review sidecar; `--json '#144'` → structured; `gh#7` →
  `github:ariadne#7`; `open '#160 M3'` (EDITOR=echo) → the m3-review. ✓
- Errors distinct: unknown id, "exists but has no M7 review sidecar", bad-ref grammar. ✓
- Lock-free: structural (`commandNeedsRepoLock`==false) + runtime (resolves under a
  held `repolock.Acquire`). ✓
- `go test ./cmd/sdlc/... ./pkg/vocab/...` green.
