# Introspection — extracting taste from past Claude sessions

## What it is

A periodic postmortem that mines your accumulated `~/.claude/projects/*.jsonl` transcripts for **taste signals** — patterns the user pushed back on, endorsed, or struggled with — and turns them into activity-typed skill files that auto-load in future sessions.

Family layout:

```
xx-introspect                  ← umbrella skill (manual control)
├── /xx-introspect extract     ← run the full pipeline against new corpus
├── /xx-introspect hint        ← author/list/retire human hints (issue#19)
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
clusters.json              # extracted clusters: rules grouped by activity, ≥2-distinct-segments threshold
        │
        ▼  read_hints.py --merge-into       (issue#19)
        │   reads ~/.claude/introspect/hints/<activity>/*.md, appends each
        │   as its own singleton cluster tagged source: "hint"
        ▼
        ▼  hint_retire_check.py             (issue#19)
        │   per-hint contradiction probe (one LLM call) against same-activity
        │   patterns; flags retirement candidates with contradicting_evidence
        ▼
clusters.json              # extracted ∪ hints, retirement-checked
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

- **Text-on-stdout building blocks.** Every script emits to stdout; the controller (`introspect-extract.sh`) is shell-glue. Same pipeline runs against any model — claude / codex / gemini / local — by overriding `EXTRACT_LLM` / `CLUSTER_LLM` / `PROBE_LLM` env vars (PROBE_LLM gates retirement-candidate detection on hints; cheap models are appropriate).
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
| Inject a hint without waiting for a pipeline run | `/xx-introspect hint <activity> "<rule>"` — writes `~/.claude/introspect/hints/<activity>/<slug>.md`. Picked up at next pipeline run as a singleton cluster. |
| Hint feels stale | Either edit the file under `~/.claude/introspect/hints/` or wait for the retirement probe to flag it on the next run |
| Roll back a bad run | `~/.claude/introspect/versions/` (M7 unbuilt — for now, manually restore from previous cache dir's clusters.json) |

## v1.0 → v1.1 history

The v1.0 architecture was a heuristic-detector pipeline (redirect / endorsement / edit-after-edit / friction regexes feeding rule-based clustering). On a 2-week dogfood corpus it produced 181 moments and **0 clusters** — the heuristics were too narrow. v1.1 swaps that primary primitive for LLM-direct extraction, keeping the heuristic detectors as optional baselines.

The first real run on a 2-week corpus (45 raw sessions → 227 segments → 437 patterns → 17 clusters → 5 deployed sub-skills) shipped in `c9b79f2`.

## Human hints (issue#19)

Hints are user-authored rules that act as **strong, single-shot signals** in
the cluster pass. Each hint becomes its own pre-formed cluster, bypassing
the ≥2-distinct-segments threshold that gates extracted patterns.

File format at `~/.claude/introspect/hints/<activity>/<slug>.md`:

```markdown
---
activity: debugging
created: 2026-05-01
---

## Rule: <imperative title>

<one-to-three sentence directive>

**Why:** <optional rationale>
```

Hints are **eligible for retirement** — every pipeline run probes each hint
against same-activity patterns (one LLM call per hint, via `$PROBE_LLM`).
If contradicting evidence is found, the hint is flagged
`retirement_candidate: true` with `contradicting_evidence`, and surfaces
first in the user-review step. The user keeps / edits / retires; retirement
is a hard delete on the file.

Authoring paths (equivalent — both produce the same file shape):
1. `/xx-introspect hint <activity> "<rule>"` — slash command via the
   umbrella skill writes the file with a slug derived from the rule title.
2. Drop a markdown file under `hints/<activity>/` by hand.

## Pointers

- Issue & log: `workshop/history/000018-transcript-driven-introspection.md` (or `workshop/issues/` if still open)
- Hints feature issue: `workshop/issues/000019-introspect-human-hints.md`
- Plan with v1.0 → v1.1 revision header: `workshop/plans/000018-transcript-driven-introspection-plan.md`
- Composition recipes: `construct/local/introspect/scripts/README.md`
- Memory entry on the precision-over-recall principle: `~/.claude/projects/-Users-xianxu-workspace-ariadne/memory/feedback_precision_over_recall.md`
