---
id: 000153
status: working
deps: []
github_issue:
created: 2026-07-01
updated: 2026-07-01
estimate_hours: 3.07
started: 2026-07-01T11:43:57-07:00
---

# sdlc retro process manual

Design origin: `workshop/parley/2026-07-01.10-47-05.962_agentic-process-documentation-strategy.md`.

## Problem

The agentic SDLC process is encoded **implicitly**, distributed across layers that
are never documented together: `AGENTS.md`/`CLAUDE.md`, `atlas/`, the `sdlc` binary
(help text + prompts it injects at gates), `.claude/skills/*/SKILL.md`,
`workshop/lessons.md`, and the memories the agent chooses to persist.

Two concrete pains follow:

1. **No single readable process manual.** A human periodically needs to see what the
   process *actually is* — to apply their own judgment about whether it's good — but
   today that means reading source across all those layers. `atlas/` is explicitly
   advisory (a lagging first-glance), so it doesn't serve this.
2. **Unclear when the sdlc-embedded prompts actually fire.** The prompts are
   systematic (they live in the binary) but there's no view of *which* fired in a
   given session, in what order, or whether the agent followed them.

Note on scope: the target is **human process-audit**, not agent-drift policing. Where
"drift" appears below it means the user's sense — *an instruction was injected but the
agent ignored it* — not documentation staleness.

## Spec

Deliverable: an `sdlc retro` verb + an agent skill. **Two outputs**, mapped to the two
pains — ship them as separate milestones.

### Output 1 — static injection catalog (the "process manual")

Deterministically unroll the always-on injection sources into a human-readable markdown
manual, each entry carrying a **live link** back to its source so the human can navigate
in and edit (or drop a `🤖[]` marker to co-edit with the agent — composes with the
review-convention markers / `xx-fix`):

- **sdlc-injected prompts** — enumerate the single builder chokepoint
  `judge.BuildPrompt(category, …)` (`cmd/sdlc/internal/judge/prompts.go`) plus the
  static help-text prompts (`orientation`, `changecode`, `startplan`, …). Because the
  judge prompts funnel through one keyed builder, this slice is a genuine deterministic
  *regeneration*, not a hand-doc that rots.
- **skills** — trigger + body from `.claude/skills/*/SKILL.md`.
- **`lessons.md`**, the **`AGENTS.md`/`CLAUDE.md` chain**, and **persisted memories**.

Cheap, deterministic, always accurate. Doubles as the **baseline** Output 2 diffs against.

### Output 2 — dynamic session reconstruction

Mine a session's transcript to show *which* catalogued injection points actually fired,
in what order, and (harder) whether injected instructions were followed. **Lead with
anomalies** — injected-but-ignored instructions, deviations from the modal sequence — and
relegate the full invocation graph to an appendix (a full per-session dump risks being
write-only; the payload is the exceptions). Output is a markdown report with live links
+ `🤖[]` annotation slots, same collaborative-editing surface as Output 1.

### Data source

- Read the **agent-specific JSONL** transcript directly (Claude
  `~/.claude/projects/<slug>/*.jsonl`; Codex `~/.codex/sessions/…`), behind a **per-agent
  parser seam**. Agent-neutrality (pair's cleaned TTY log) is deferred to v2: TTY is a
  lossy projection of the context window and would cost the completeness that makes the
  analysis worthwhile — the incumbent session-miner (`introspect`) also chose JSONL.
- Reuse `introspect`'s `normalize.py`/segmentation **plumbing** — but not its purpose.
  `introspect` is macro taste-mining that feeds the *agent* (patterns → skills); retro is
  micro provenance-structuring that feeds the *human* (injection → manual). Co-tenants of
  the same raw material, opposite direction.

### Provenance: what's recoverable vs. what needs instrumentation

- *That* a command/skill fired is **free** from JSONL — a Bash tool call for `sdlc …`; a
  Skill tool call whose name → file link is 1:1.
- *Linking rendered injected text to its exact source* is free only for **skills** and
  **static help text**. The judge/review prompts are **interpolated** (`BuildPrompt` takes
  a `PromptInput`), so string-matching rendered→source fails **silently** when a template
  is edited. v1: recover firing + a small lookup for the static prompts, accept fuzzy
  links; instrument only where it bites — emit a provenance id at the `BuildPrompt`
  chokepoint (v2).
- The sdlc **review forks** are already re-joined for us. `close`/`milestone-close`
  dispatch a fresh-context reviewer (`claude -p` / `codex exec`, `judge.Dispatch`), and
  `writeReviewSidecar` (#136) captures its transcript into a **correlated,
  version-controlled** sidecar `workshop/history/NNNNNN-slug-…-review.md` whose header
  carries issue/boundary/window/reviewer/verdict. Retro reads the sidecar, **not** the
  orphan reviewer JSONL.

### Known blind spots (document as stated assumptions)

- `AGENTS.md`/`CLAUDE.md` + persisted memories inject **before** the transcript stream →
  prepend as an implicit "step 0" by convention, based on what files exist, rather than
  detecting them from the log.
- Generic **Task-tool subagent forks** the main agent spawns mid-session leave orphan
  JSONLs with no sidecar → v2 gap; needs cwd + timestamp correlation (precedent exists in
  `cmd/sdlc/actual.go` / `activetime.go`).

## Done when

- `sdlc retro` emits a deterministic markdown **process manual**: every sdlc-injected
  prompt + skill trigger + `lessons.md` + `AGENTS.md` chain + memories, each with a live
  link to its source. (M1)
- A dynamic **per-session report** reconstructs which catalogued injection points fired,
  in order, anomalies first, with `🤖[]` annotation slots. (M2)
- Blind spots (AGENTS.md/memory step-0; Task-tool forks) are documented as stated
  assumptions in the skill.

## Estimate

Scope: **M1 only** (the detailed plan). M2 (dynamic reconstruction) is deferred to
its own `start-plan` and will revise `estimate_hours` upward when its scope is
knowable — estimating it now would be a guess (AGENTS.md: estimate post-design).

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: greenfield-go-module   design=0.8 impl=1.2
item: smaller-go-module      design=0.3 impl=0.3
item: atlas-docs             design=0.0 impl=0.1
item: milestone-review       design=0.0 impl=0.2
design-buffer: 0.15
total: 3.07
```

Derivation: `internal/retro` (pure core + 6 collectors + tests) = greenfield module;
`retro.go` cobra glue + `helptext.FS` accessor = smaller module extending the existing
verb idiom; atlas entry = atlas-docs; M1 is a tagged milestone so it owns a
`milestone-close` review. recomputed = (0.8+0.3+0.0+0.0)×1.15 + (1.2+0.3+0.1+0.2) =
1.265 + 1.8 = 3.065 ≈ 3.07.

**Revision (post-change-code judges):** initial block was 5.45h with v2-scale impl
hours (2.0/1.0/0.5) and no milestone-review line. The estimate-quality gate flagged it
as ~1.5–2× hot vs the v3.1 corpus (greenfield impl clusters 0.3–1.2, smaller 0.15–0.8,
atlas ~0.1) and missing the milestone-review cost. Recalibrated to v3.1 ranges +
design-buffer 0.15 (corpus modal) → 3.07h.

## Plan

Coarse decomposition; the durable detailed plan lands at `start-plan` via
`superpowers-writing-plans` in `workshop/plans/`.

- [x] M1 — static injection catalog (`sdlc retro` unrolls `BuildPrompt` + help text +
      skill triggers + lessons + AGENTS chain + memories → linked markdown manual)
- [ ] M2 — dynamic session reconstruction (JSONL per-agent parser + sidecar join →
      anomaly-first report with annotation surface)

## Log

### 2026-07-01

Issue captured from the design parley. Grounding verified against source this session:
`judge.BuildPrompt(category, PromptInput)` is the single interpolating chokepoint for
gate prompts (`cmd/sdlc/internal/judge/prompts.go:134`); boundary reviews dispatch a
fresh-context `claude -p`/`codex exec` subprocess (`judge.Dispatch`, `dispatch.go`) whose
transcript is persisted to a correlated sidecar by `writeReviewSidecar` (#136,
`reviewsidecar.go` / `milestoneclose.go:476`); both agent transcript stores
(`~/.claude/projects`, `~/.codex/sessions`) are already known to sdlc
(`actual.go` / `activetime.go`). Deferred to planning: exact `sdlc retro` flag surface,
whether the skill or the binary owns the reconstruction pass.

Plan written (`workshop/plans/000153-sdlc-retro-process-manual-plan.md`, M1 detailed /
M2 coarse). Fresh-eyes plan review caught two snippet defects (fixed): the `lessons`
judge category renders `""` → use `judge.LessonsReminder`; the `helptext/embed.go`
accessor collided `var fs embed.FS` with `import "io/fs"` → return concrete `embed.FS`.
`sdlc change-code` gates both **INFO** (non-blocking): plan-quality independently
verified the ARCH-PURPOSE catch (`EstimateQuality` fires at `changecode.go:176` while
`AllCategories()` omits it); estimate-quality flagged the block as ~1.5–2× hot vs the
v3.1 corpus → recalibrated 5.45 → 3.07h (+ `milestone-review` line). Plan-quality nits
folded in: dropped the dead `helptext.Names()` (keep only `FS()`), truthful `When` for
`root` (bare `sdlc --help`). **Base-layer reach:** `cmd/sdlc/{main.go, helptext/embed.go,
retro.go}` are base-layer surface (`construct/base.manifest`), so `sdlc retro` ships to
every downstream ariadne repo — the change is additive + read-only, low risk, but noted
per the workshop-extensions guidance. Branch `000153-sdlc-retro-process-manual` created
in-place; entering implementation.

### 2026-07-01 — M1 implemented (static catalog)

Landed `cmd/sdlc/internal/retro` (pure `InjectionSource` + `renderManual`; IO
collectors for judge prompts, help text, skills, lessons, AGENTS chain, memories),
the `sdlc retro` verb (`retro.go` + `helptext/retro.md`), and `helptext.FS()`.
All collectors TDD'd (fstest.MapFS / temp dirs / pure). Two rendering bugs found +
fixed during the smoke run: (1) absolute/empty links (memories live outside the
repo) — `renderManual` now leaves `/`-absolute links unprefixed and renders empty
links as plain headings; (2) inlined judge prompts carry their own `#`/`##`
headings (ARCH registry, checklists) that hijacked the manual's structure — bodies
with headings are now `~~~`-fenced, and judge prompts are shown as a first-para
gist + link (full inline was ~1100 lines with the ARCH block 4×; gist → 537).

Verified: `go build ./…` ok, `go vet ./cmd/sdlc/…` ok, `go test ./cmd/sdlc/…` all
pass. **Shadow-sweep (ARCH-PURPOSE)** on `sdlc retro`: prompts=8/8 (incl.
estimate-quality that `AllCategories()` omits), help-text=20/20 files,
skills=24/24 dirs, lessons+AGENTS+memory sections all present; exactly the 6
intended `##` sections, no leaked prompt headers. Held at M1 for user testing
before `sdlc milestone-close` (the mandatory fresh-context review boundary).
