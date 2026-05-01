# Introspection — extracting taste from past Claude sessions

## What it is

A periodic postmortem that mines your accumulated `~/.claude/projects/*.jsonl` transcripts for **taste signals** — patterns the user pushed back on, endorsed, or struggled with — and turns them into activity-typed skill files that auto-load in future sessions.

Family layout:

```
xx-introspect                  ← umbrella skill (manual control)
├── /xx-introspect extract     ← run the full pipeline against new corpus
└── /xx-introspect load        ← force-load a specific introspect-<activity>

introspect-<activity>          ← five auto-loading sub-skills (one per activity)
├── introspect-implementation
├── introspect-planning
├── introspect-debugging
├── introspect-brainstorming
└── introspect-exploration
```

The sub-skills auto-load via Claude Code's standard description-based discovery — each description names the activity-shape signals that should trigger it.

## Pipeline shape

```
~/.claude/projects/*.jsonl
        │
        ▼  segment_text + normalize.py
sessions.json              # one row per "segment" (away_summary / 60-min gap boundaries)
        │
        ▼  LLM-direct activity classify (skill body + user approval)
classified.json            # segment → activity (six taxonomy buckets + skip/oos/ambiguous)
        │
        ▼  segment_text.py + prompts/extract.md
patterns/<seg-id>.json     # per-segment taste-pattern JSON (LLM call per segment)
        │
        ▼  aggregate_patterns.py (stable IDs, decoration, fence stripping)
patterns.json              # combined array
        │
        ▼  prompts/cluster.md
clusters.json              # rules grouped by activity, ≥2-distinct-segments threshold
        │
        ▼  manual write-back (Stage 7 not yet a script)
~/.claude/skills/introspect-<activity>/SKILL.md
```

## Where things live

| Path | Purpose |
|---|---|
| `construct/local/introspect/SKILL.md` | umbrella skill body |
| `construct/local/introspect/scripts/` | pipeline scripts (normalize, classify, segment_text, aggregate, controller) |
| `construct/local/introspect/prompts/` | system prompts for extract + cluster passes |
| `construct/local/introspect/scripts/README.md` | composition recipes for claude / codex / gemini / local |
| `~/.claude/skills/introspect-*/SKILL.md` | deployed sub-skills (auto-load) |
| `~/.claude/introspect/cache/<run-id>/` | per-run intermediates (sessions, classified, patterns, clusters) |
| `~/.claude/introspect/hints/<activity>/<slug>.md` | human-authored hints (issue#19, eligible for retirement) |
| `~/.claude/introspect/versions/v<N>/` | post-run snapshots (versioning deferred — M7 is unbuilt) |
| `~/.claude/introspect-cache` | backward-compat symlink → `~/.claude/introspect/cache/` |

## Operating principles

- **Text-on-stdout building blocks.** Every script emits to stdout; the controller (`introspect-extract.sh`) is shell-glue. Same pipeline runs against any model — claude / codex / gemini / local — by overriding `EXTRACT_LLM` / `CLUSTER_LLM` env vars.
- **User-in-the-loop on every model judgment.** Activity classification, cluster acceptance, and write-back all surface to the user before disk writes.
- **Precision over recall.** Both the classifier and the cluster pass leave uncertain rows as `ambiguous`. Stage 4 filters them. Forcing weak clusters pollutes the resulting rules.
- **Cache-safe re-runs.** Per-segment LLM outputs cached at `<run-dir>/patterns/<seg-id>.json` — Ctrl-C and resume picks up where it stopped, full re-runs are cheap on unchanged segments.
- **Cadence is "full run every couple of weeks."** Delta/merge logic (M7) is intentionally not built; the cache makes a full re-run effectively incremental in compute terms.

## When you'd touch this

| Scenario | What to do |
|---|---|
| Recurring postmortem (biweekly) | `EXTRACT_LLM='claude --print --system-prompt "$1" --tools "" --model opus' bash construct/local/introspect/scripts/introspect-extract.sh ~/.claude/introspect/cache/<new-run-id>` then walk through clusters and run the manual write-back script. |
| Auto-load fired at the wrong activity | `/xx-introspect load <activity>` (or just `/introspect-<activity>` directly) |
| Tune a sub-skill's auto-trigger | Edit the `description:` line in `~/.claude/skills/introspect-<activity>/SKILL.md` |
| Tune the extraction prompt | Edit `construct/local/introspect/prompts/extract.md` and rerun with `--force` |
| Roll back a bad run | `~/.claude/introspect/versions/` (M7 unbuilt — for now, manually restore from previous cache dir's clusters.json) |

## v1.0 → v1.1 history

The v1.0 architecture was a heuristic-detector pipeline (redirect / endorsement / edit-after-edit / friction regexes feeding rule-based clustering). On a 2-week dogfood corpus it produced 181 moments and **0 clusters** — the heuristics were too narrow. v1.1 swaps that primary primitive for LLM-direct extraction, keeping the heuristic detectors as optional baselines.

The first real run on a 2-week corpus (45 raw sessions → 227 segments → 437 patterns → 17 clusters → 5 deployed sub-skills) shipped in `c9b79f2`.

## Pointers

- Issue & log: `workshop/history/000018-transcript-driven-introspection.md` (or `workshop/issues/` if still open)
- Plan with v1.0 → v1.1 revision header: `workshop/plans/000018-transcript-driven-introspection-plan.md`
- Composition recipes: `construct/local/introspect/scripts/README.md`
- Memory entry on the precision-over-recall principle: `~/.claude/projects/-Users-xianxu-workspace-ariadne/memory/feedback_precision_over_recall.md`
