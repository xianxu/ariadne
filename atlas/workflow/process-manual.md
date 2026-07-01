# sdlc process-manual — the process manual

## What it is

`sdlc process-manual` regenerates a single markdown **process manual**: every always-on
injection source that shapes an agent's behavior, unrolled into one document with
a live link back to each source. It answers "what *is* the process, and where does
each piece live?" — the counterpart to `atlas/` (advisory prose) grounded instead
in the binary + files that actually inject context (#153).

It is a **deterministic regeneration**, not a hand-maintained doc: re-run it rather
than editing its output. The sdlc-prompt slice comes straight from the binary, so
that part never drifts from what actually fires.

## The six source kinds

| Kind | Source | Collector |
|------|--------|-----------|
| `sdlc-injected prompts` | `judge.BuildPrompt` (all 8 categories, incl. change-code-time `estimate-quality` that `AllCategories()` omits) | `judgeSources` (pure) |
| `help text` | `cmd/sdlc/helptext/*.md` (embedded) | `helptextSources(helptext.FS())` |
| `skills` | `.claude/skills/*/SKILL.md` triggers | `skillSources` |
| `lessons` | `workshop/lessons.md` | `fileSources` |
| `AGENTS chain` | `AGENTS.md` + `.base`/`.local`, `CLAUDE.md`/`GEMINI.md` | `fileSources` |
| `persisted memories` | `~/.claude/projects/<slug>/memory` | `memorySources` (best-effort) |

## Design

Pure core + thin IO shell (ARCH-PURE), all in `cmd/sdlc/internal/processmanual`:

- `InjectionSource` + `renderManual(sources, linkPrefix)` are pure — grouping,
  ordering, link-prefixing, and heading-safe **fencing** of bodies (an inlined
  prompt's own `#` headings are fenced so they can't hijack the manual's structure).
- Each collector is injected with its IO seam (`fs.FS`, a dir, or `$HOME`) so it
  tests against `fstest.MapFS` / a temp dir with no mocks. `judgeSources` is pure
  because `judge.BuildPrompt` is pure.
- Judge prompts are shown as a **first-paragraph gist** + link, not inlined in full
  (each rendered prompt runs to hundreds of lines and re-embeds the ARCH registry —
  inlining all 8 would bloat the manual ~4×). The `When` says where each fires.
- `cmd/sdlc/processmanual.go` is the cobra glue: `--out <path>` writes to a file (links
  re-based to that file), else stdout. Reuses `gitx.RepoTopLevel()` for the root.

## Stated blind spots (M1)

- **Persisted memories are agent-specific (Claude), private, and live outside the
  repo.** They are **redacted by default** — inlining them would write absolute home
  paths + personal content into a (committable) file. `--include-memory` shows them
  for local inspection only and is refused together with `--out`. When shown, they are
  located by convention (`claudeProjectSlug`); an absent dir yields a "none found" note.
- **This is the static catalog (M1).** The dynamic pass — which of these actually
  *fired* in a given session, in what order, and whether the agent followed them —
  is a separate milestone (M2), which consumes this catalog as its baseline.

## Base-layer note

`cmd/sdlc/{main.go, helptext/embed.go, processmanual.go}` are base-layer surface, so
`sdlc process-manual` ships to every downstream ariadne repo (additive + read-only).
