---
id: 000125
status: working
deps: [ariadne#122]
github_issue:
created: 2026-06-25
updated: 2026-06-25
estimate_hours: 1.56
started: 2026-06-25T14:33:20-07:00
---

# sdlc embeds a help-text fragment GENERATED from the issue vocabulary (stop hand-maintaining lifecycle prose)

## Problem

`sdlc`'s embedded help text (`cmd/sdlc/helptext/*.md`) hand-restates lifecycle facts and
**drifts**. #122 M4 made `set-status.md`'s "All other transitions are allowed without
guards" *false* (the lifecycle gate now refuses non-modeled flips); nothing caught it
automatically — the #122 whole-issue fresh-eyes review did (FIX-THEN-SHIP), and it was
patched by hand. So the operator-facing prose is a **hand-maintained shadow** of the
model, not derived — the exact drift the vocabulary layer (#122) exists to kill, for the
one consumer we never wired. We already generate-from-source where the target is
*templated* — code → `issue.json` via `go:embed`; the vocabulary skill `SKILL.md` via the
`.dynamic-skill` renderer — the gap is the *free-form help prose*.

## Spec

Generate the model-derived portion of the relevant help text from the vocabulary, so it
can't drift; keep the hand-written framing prose. **Generate the facts, reference the
source** (don't re-enumerate edges by hand). Likely shape:

- Extend the vocabulary renderer (the one that already emits the skill breadcrumb) to
  emit a **help fragment** for the issue lifecycle — the legal transitions / the gate
  description — as markdown.
- `sdlc` includes that fragment in `set-status` help (and any other help that states
  lifecycle facts — `claim`/`close`) via `go:embed` of the generated fragment, regenerated
  by the same generic `go generate` / `make vocab-embed`, with the generic git-diff
  freshness check (like the embedded JSON, #122 M3/M4).
- Reuses the per-language binding insight: the help fragment is just another generated
  face of the model.

## Done when

- Changing `construct/vocabulary/issue.cue`'s lifecycle (add/remove an edge) + regenerating
  updates `sdlc issue set-status --help` with **no hand edit**.
- A stale help fragment fails the freshness check (CI-catchable).
- The hand-maintained lifecycle claims (e.g. the "all other transitions allowed" class)
  are gone — the facts are model-derived.

## Estimate

```estimate
model: estimate-logic-v2
familiarity: 1.0
item: smaller-go-module    design=0.2 impl=0.7
item: milestone-review     design=0.0 impl=0.6
design-buffer: 0.30
total: 1.56
```

Single-pass atomic (no Mx): the `renderLong` substitution seam + pure render helpers + the
FULL shadow-sweep (set-status.md + issue.md + claim.md) + tests (smaller-go-module), plus the
one close review. (Scope widened from the first cut after the plan-quality judge's complete
shadow-sweep — issue.md is the biggest restatement.) estimate-logic-v2 runs ~2.3× high vs
ship-wall-clock (#127); honest un-compensated v2.

## Plan

Design decided (operator-confirmed): **RUNTIME render, not a generated file.** The help consumer
is Go (`sdlc`), which already has the model via `pkg/vocab` at runtime — so render the lifecycle
facts live and template the `.md` (the `{{ARCH_STAR}}` idiom): no cached fragment, no new
freshness gate, can't drift. Freshness of the underlying `issue.json` is already gated by
`make vocab-embed`.

- [x] **Substitution seam (Critical fix):** a single `renderLong(name string) string` in `cmd/sdlc` = `lifecycleReplacer.Replace(helptext.MustGet(name))` (a `strings.NewReplacer`, the `{{ARCH_STAR}}` idiom). Route EVERY command-Long load through it — `main.go:49/64/89/95` + **`issue.go:46`** (the canonical `sdlc issue set-status`) + the hidden alias `main.go:95`. (No-op when no placeholder present.) `add()` alone misses all three set-status sites.
- [x] **Render helpers** (pure, on `IssueModel`, reuse `AllStatuses()` ordering — ARCH-DRY): `RenderLifecycleHelp()` = the full block (STATUSES with `When` + LEGAL TRANSITIONS from `LegalTransitions`); plus small views for the inline shadows (status-name list, the `·` gloss). Byte-stable (unit-tested like `renderSkill`).
- [x] **Full shadow-sweep** (ARCH-PURPOSE — every prose restatement, not a subset): `set-status.md` → `{{LIFECYCLE}}` (replace the STATUSES block + the lifecycle-graph enumeration; keep `→working`/`→done`/`reopen` policy + the `--force` framing). `issue.md` → derive `:23` (status names) + `:50-51` (the `when` gloss — currently a hand-paraphrase). `claim.md:21` → **reword** the `(working/blocked/punt/wontfix/done)` list to "anything other than open" (*reference* the source, don't enumerate — drift-proof without a placeholder).
- [x] Tests: render content (every status + its `when`; the legal edges) + byte-stability; the built `set-status` **and `issue`** Longs have NO surviving `{{` and render every model status; the hand-maintained "all other transitions allowed" claim is gone (Done-when). A guard mirroring `estimate_helptext_test.go` (the existing drift-test for this class).

## Log

### 2026-06-25

- Filed as a #122 follow-up (the prose-consumer half of "compiled to consumers"). Motivated
  by the M4 help-text drift the close review caught — `issue.cue` was, for the help surface,
  still just-documentation that didn't derive from the model. The general principle (lessons):
  *every prose surface that restates the model is a shadow; generate it or reference it.*
- **Plan-quality FAILURE addressed (2026-06-25):** (A, Critical) the substitution seam was wrong — `set-status`'s Long is set at `issue.go:46` + the alias `main.go:95`, NOT through `add()`, so `add()`-substitution would ship a literal `{{LIFECYCLE}}`. Fixed: a single `renderLong` wrapping every `helptext.MustGet` command-Long site. (B, Important — ARCH-PURPOSE) the sweep missed `issue.md` (`:23` + `:50-51`), the *biggest* status/`when` restatement — extended the sweep to all surfaces (set-status + issue.md derive; claim.md references). (C) reuse `AllStatuses()` ordering. Scope + estimate widened (1.3 → 1.56).
- **Implemented (impl fork) + verified (2026-06-25):** `pkg/vocab` got 3 pure render methods (`RenderLifecycleHelp` + `StatusNames`/`StatusGloss`, neutral/data-only, reusing `AllStatuses`); `renderLong` seam in cmd/sdlc routes every command-Long load (incl. the `issue.go:46` site `add()` missed — the Critical fix); set-status.md + issue.md derive via `{{LIFECYCLE}}`/`{{STATUS_NAMES}}`/`{{STATUS_GLOSS}}`, claim.md references ("anything other than open"). A wiring guard `TestNoCommandLongHasSurvivingPlaceholder` walks the real `buildRoot()` tree asserting no `{{` survives in any command's Long (regression guard for the Critical finding). `go test ./...` + `go vet` green. **Dogfooded the core Done-when:** added a temp `open→blocked` edge to issue.cue → `vocabulary export` regen → `set-status --help` showed `open → working, blocked, …` with NO hand edit, then reverted. Freshness: no separate fragment — the help derives from `issue.json`, gated by `make vocab-embed`.
