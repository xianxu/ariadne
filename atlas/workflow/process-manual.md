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

It has two modes: the default **static catalog** (what *can* inject) and, with
`--session`, a **dynamic reconstruction** (what *did* fire in one session, in order
— #157, below).

## The six source kinds

| Kind | Source | Collector |
|------|--------|-----------|
| `sdlc-injected prompts` | `cmd/sdlc/internal/judge/prompts/*.md` (embedded templates, rendered by `judge.BuildPrompt`; all 8 categories, incl. change-code-time `estimate-quality` that `AllCategories()` omits) | `judgeSources` (pure) |
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
- Judge prompts (#153 M2) are single-sourced as embedded `judge/prompts/*.md`
  templates (placeholder substitution in `BuildPrompt`, byte-fidelity pinned by
  `judge/golden_test.go`) — so `judgeSources` links straight to the **readable `.md`**
  (like help text/skills), not to Go code. The body is a **first-paragraph gist** +
  link by default; `--full` inlines the complete rendered prompt (fenced, so the
  outline is unchanged — each prompt otherwise runs hundreds of lines, re-embedding
  the ARCH registry ~4×). The `When` says where each fires.
- `cmd/sdlc/processmanual.go` is the cobra glue: `--out <path>` writes to a file (links
  re-based to that file), else stdout. Reuses `gitx.RepoTopLevel()` for the root.

## Dynamic reconstruction (`--session`, #157)

`sdlc process-manual --session <jsonl|current>` reads a Claude session transcript
and reconstructs which catalogued injection points actually **fired**, in timestamp
order, matched back to the catalog above. It is **Go-native** (not a shell-out to
introspect's Python) precisely so it matches against the in-process `InjectionSource`
catalog rather than serializing across a boundary (ARCH-DRY). All in
`cmd/sdlc/internal/processmanual/session.go`:

- **Pure core** (fixture-tested, no IO): `parseEvents(data, validVerbs)` tolerantly
  scans the JSONL (unknown record types skipped), keeps the fired injections via
  `classifyToolUse`, and recovers `close`/`milestone-close` **verdicts** from the
  following `tool_result`'s stdout — linked by `tool_use_id`, parsed with the exact
  `judge.ParseVerdict` that `close` itself uses. `segmentEvents` splits on a >60-min
  lull (constant ported from introspect's `normalize.py`) or an `away_summary`
  boundary. `renderSessionReport` emits linked, segmented markdown.
- **Injection detection is precision-first.** A Bash `sdlc <verb>` only counts when
  `sdlc` sits at a **command boundary** AND the verb is a **real, linkable verb**
  (validated against the catalog's help-text titles — so "classified" ⟺ "in the
  catalog"). This drops prose mentions (`git commit -m "…sdlc close…"`), flags
  (`cmd/sdlc --include=…`), and dev-time `go run ./cmd/sdlc <verb>` smoke calls —
  none of which are real workflow events.
- **IO shell**: `locateSessionJSONL` (prefers `$CLAUDE_CODE_SESSION_ID`, else newest
  `*.jsonl` by mtime under `claudeProjectSlug`) + `SessionReport` (the composer
  mirroring `Manual`). `runProcessManual` branches on `--session`.

**Two hard limits, rendered into the output** (not silently omitted — ARCH-PURPOSE):
(1) agents-chain (AGENTS/CLAUDE.md) + memory are session-start *system-prompt*
injections that never appear in a transcript — availability is knowable, firing is
not; (2) forked review *prompts* aren't in the transcript, only their *output* (the
recovered verdict). **Deferred by design**: anomaly / "injected-but-ignored"
detection — undetectable for agents-chain/memory, and an LLM-judge problem for "was
the guidance followed?".

## Stated blind spots (M1)

- **Persisted memories are agent-specific (Claude), private, and live outside the
  repo.** They are **redacted by default** — inlining them would write absolute home
  paths + personal content into a (committable) file. `--include-memory` shows them
  for local inspection only and is refused together with `--out`. When shown, they are
  located by convention (`claudeProjectSlug`); an absent dir yields a "none found" note.
- **The default (no `--session`) is the static catalog (#153 M1+M2: catalog +
  judge-prompts-as-markdown).** The dynamic *fired-in-a-session* pass is delivered
  (**#157**, `--session`, above); the further *whether the agent followed the guidance*
  pass (an LLM-judge / anomaly problem) remains deferred.

## Base-layer note

`cmd/sdlc/{main.go, helptext/embed.go, processmanual.go}` + `internal/processmanual/`
(incl. `session.go`) are base-layer surface, so `sdlc process-manual` (both the static
manual and `--session`) ships to every downstream ariadne repo (additive + read-only).
