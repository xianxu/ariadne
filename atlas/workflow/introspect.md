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

## The unit is a segment, not a raw Claude session

**A "session" in this pipeline means a segment, not a raw `.jsonl` Claude
session.** A single raw Claude session is typically multi-activity by design
— the user keeps context warm across topic shifts (single-threaded attention
maps better to one long session than to many fresh ones). Classifying the
whole raw session by its *first* message would mislabel everything that
happens after the first pivot.

Segmentation is what makes per-segment activity classification valid.
`normalize.py` slices each raw session into segments on two boundaries:

1. **60-minute idle gap** between events (`GAP_BOUNDARY_SECONDS = 3600` in
   `normalize.py`). Stepping away to think, eat, or sleep produces a fresh
   segment when work resumes.
2. **`away_summary` event** — emitted by Claude Code when `/compact` runs.
   The recap message is appended to the *closing* segment as its last event;
   the next user message starts a new segment.

Empirically, 415 raw sessions × 27 days produced 1367 segments
(avg 3.3 segments/raw-session). The longest raw sessions split into 50-130
segments — those are the multi-day burnt-in conversations the operator keeps
returning to.

Each segment's `first_user_message`, `files_written`, `tool_calls_by_name`,
etc. are **segment-local**, not raw-session-global. That's why downstream
stages (classify, detect, cluster) can treat each segment as an atomic
activity-unit despite the operator's multi-topic session habit.

If you find yourself worrying that "long sessions are multi-activity so
classification is broken" — re-read this section. The pipeline handles it.

## Pipeline shape

```
~/.claude/projects/*.jsonl
        │
        ▼  segment_text + normalize.py
sessions.json              # one row per segment (see "The unit is a segment" above)
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
        ▼  Stage 4.5 — load hints (MANDATORY, every run)
        │   reads ~/.claude/introspect/hints/<activity>/*.md, appends each
        │   as a singleton cluster tagged source: "hint" with retirement
        │   probe against same-activity moments
        ▼
clusters.json              # extracted ∪ hints, retirement-checked
        │
        ▼  Stage 7 — write-back + hint consumption
        │   atomic write to ~/.claude/skills/introspect-<activity>/SKILL.md
        │   then: mv consumed hints → versions/<run-id>/consumed-hints/
~/.claude/skills/introspect-<activity>/SKILL.md
~/.claude/introspect/versions/<run-id>/consumed-hints/<activity>/<slug>.md
```

## Agent-neutral event layer (#173)

The pipeline no longer reads Claude's wire format directly. A canonical
**`NormEvent`** (`scripts/events.py`) is the shape `normalize` aggregation and the
`detect` detectors reason about; per-agent **adapters** map raw transcripts into it:

```
~/.claude/projects/*.jsonl   ─┐
~/.codex/sessions/**/*.jsonl  ─┼─► agent adapter ──► NormEvent stream ──► normalize + detect + segment_text
(future agents) ──────────────┘   claude_events /     (agent-agnostic:
                                  codex_events          aggregation + 4 detectors + render)
```

- **`events.py`** — `NormEvent` (kind ∈ {user_msg, assistant_msg, tool_call,
  tool_result, file_edit, boundary}) + shared `FRICTION_HINTS`. Pure.
- **`agent_claude.py`** — `claude_events(line)` owns the Claude wire-format reads
  (message / tool_use / toolUseResult / away_summary) and **derives `is_error`**
  (Claude flag + text patterns) so detectors read a flag.
- **`agent_codex.py`** — `codex_events(line, raw_sid)` maps codex rollout events
  (`event_msg`/`response_item`/`patch_apply_end`/`compacted`); **derives `is_error`
  from the "Process exited with code N" string** (codex has no error flag).
- **`segment_loader.py`** — `load_segment_norm_events(session)`: agent-keyed raw
  reader (claude project JSONL by sessionId / codex rollout by persisted path) →
  NormEvent stream. Shared by detect + segment_text.
- `normalize`'s `aggregate_norm_event`, the 4 detectors, AND `segment_text`'s
  `render_segment` read **only** `NormEvent` — a new agent is one adapter, consumers
  untouched (ARCH-DRY). Session metadata (cwd/branch, segment boundaries) is a small
  per-agent seam: claude reads per-line (`_apply_line_metadata`, away_summary), codex
  reads `session_meta` once + segments on `compacted`.

Why: introspect was Claude-only while the rest of ariadne is agent-neutral, so
codex taste was invisible. M1 put normalize + detect behind the abstraction
(behavior-preserving — byte-identical `sessions.json`, identical run-3 moment set);
**M2 added the codex adapter + `segment_loader` + lifted `segment_text`**; **M3
dogfooded it over the real codex corpus** (see "M3 finding" below).

## Codex transcript format (single source of truth — shared with #172)

The codex rollout format is documented **here, once**, because two implementations
derive from it and can't share code: introspect is **Python**
(`agent_codex.py`/`normalize.py`); #172's `process-manual --session` is **Go**
(`cmd/sdlc/internal/processmanual/session.go`). The DRY point (ARCH-DRY) is this
spec, not a shared reader. Keep this table and `agent_codex.py` in lockstep.

**Location.** One JSONL file per raw session at
`~/.codex/sessions/YYYY/MM/DD/rollout-<ts>-<uuid>.jsonl`. The sibling SQLite DBs
(`logs_2.sqlite`, `state_5.sqlite`) are tracing/state, **not** the transcript —
ignore them. So ingestion is a **format mapping, not a DB extraction.**

**Record shape.** Every line is `{timestamp, type, payload}`. Codex emits both an
`event_msg/*` stream AND `response_item/*` canonical items — the same turn appears
twice. **Pick ONE canonical source per field** or counts double.

| codex event | → `NormEvent` | notes |
|---|---|---|
| `session_meta.payload` `{id, cwd, git.branch, forked_from_id}` | session id / cwd / branch (via locator, not an event) | **use the FIRST `session_meta` in the file** — see forks below |
| `event_msg/user_message` | `user_msg` | canonical user source |
| `event_msg/agent_message` | `assistant_msg` | canonical assistant source (NOT `response_item/message` — that's the double-rep) |
| `response_item/function_call` (+ `arguments`) | `tool_call` | `call_id` → `tool_use_id` (correlates to its output) |
| `response_item/custom_tool_call` (+ `input`) | `tool_call` | custom/MCP; args under `input` (dict), not `arguments` |
| `response_item/{web_search_call,tool_search_call}` | `tool_call` | else tool counts undercount |
| `response_item/function_call_output` (+ `custom_tool_call_output`) | `tool_result` | `output` is a **plain string**; `is_error` DERIVED (below) |
| `event_msg/patch_apply_end` `{changes, success}` | one `file_edit` per changed file | `type:"add"`→Write, else Edit; `success` is a structured flag (cleaner than Claude) |
| `compacted` | `boundary` (`compacted`) | explicit segment boundary — better than Claude's lull heuristic |
| `response_item/reasoning`, `event_msg/token_count`, `turn_context`, `response_item/message` | dropped | metadata / double-representation |

**`is_error` is derived, not a flag.** `function_call_output.output` is a plain
string (e.g. `"…Process exited with code 71\n…Operation not permitted"`) — codex has
no structured error field. Mirror Claude's gate: a failing signal (a non-zero
`Process exited with code N`, or an `error:`/`exit code` prefix) **paired with a
`FRICTION_HINT`** (permission / sandbox / operation-not-permitted / blocked). A
non-zero exit ALONE is not friction — grep/sed/ls no-match, `command not found`,
and expected test failures exit non-zero benignly (M3: 106 of 112 raw codex
"friction" moments were this noise). `event_msg/mcp_tool_call_end` (rare) is not
yet mapped.

**⚠️ Multi-agent forks replay the parent — MUST be handled (both languages).** A
forked/sub-agent rollout (pair, parley.nvim multi-agent runs; meta has
`forked_from_id`/`parent_thread_id`/`agent_nickname`) **replays the parent session's
entire transcript**, then adds its own events. Two consequences:

1. It carries **TWO `session_meta` lines** — its own (`id`=fork, `forked_from_id`
   set) FIRST, then the replayed parent's (`id`=parent). Key off the **FIRST** meta,
   or every fork collapses onto its parent's id.
2. Its events **duplicate the parent's**. Processing both double-counts every shared
   moment — the M3 dogfood measured **66% moment inflation** (one redirect counted
   ×11 across a parent + 10 forks).

`normalize.process_codex_file` **skips any rollout whose first meta has
`forked_from_id`** (the user-facing taste — redirects/endorsements — lives in the
parent thread; forks are agent-orchestration). `run.json` reports
`codex_forks_skipped`. Any consumer of this format (incl. #172's Go) must do the
same, or its per-session counts are wrong.

## M3 finding — does codex reopen the taste well? (No)

The M3 dogfood answered #173's headline question: **does the codex corpus yield
more taste than Claude, which hit diminishing returns in #169?** Confound-normalized
answer: **no.** Over 54 root codex sessions (pair / parley.nvim / ariadne), the raw
counts *looked* ~10× richer (112 friction, 198 endorsements) but were **~95%
artifact**: 66% fork-replay duplication + benign-exit friction. Cleaned, the codex
signal is **8 unique redirects (all project-local UX), 12 genuine sandbox-denial
frictions, 0 tool-backed endorsements, 0 new generalizable rules** — the one real
debugging moment ("don't guess, use the logging") was already deployed in
`introspect-debugging`. Same conclusion as #169: the constitution + 22 existing
rules capture the transferable taste regardless of agent; **introspect run cadence
should stretch on both agents.** The engineering (agent-neutral ingest, end-to-end)
is validated; the *hypothesis* that a less-tuned agent reopens the well is not.

## Where things live

| Path | Purpose |
|---|---|
| `construct/local/introspect/SKILL.md` | umbrella skill body |
| `construct/local/introspect/scripts/` | pipeline scripts (normalize, classify, segment_text, aggregate, controller) |
| `construct/local/introspect/scripts/events.py` | canonical `NormEvent` — the agent-neutral core (#173) |
| `construct/local/introspect/scripts/agent_claude.py` · `agent_codex.py` | per-agent raw→NormEvent adapters (#173) |
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

Hints are **one-shot inputs, not durable storage.** A hint lives in
`~/.claude/introspect/hints/` only until the next extract pass renders it
into a deployed `introspect-<activity>/SKILL.md`. At write-back the
originating file is moved to
`~/.claude/introspect/versions/<run-id>/consumed-hints/<activity>/<slug>.md`
as an audit trail — the deployed SKILL.md is the durable taste model.

Hints are **eligible for retirement** at review time. Every pipeline run
probes each hint against same-activity patterns (one LLM call per hint, via
`$PROBE_LLM`). If contradicting evidence is found the hint is flagged
`retirement_candidate: true` with `contradicting_evidence`, and surfaces
first in the user-review step. The user keeps (consume normally), edits
(revise then consume), or retires (hard delete, no SKILL.md write).

A **rejected** hint (operator says "don't write this to SKILL.md") stays
in `hints/` for the next run to re-probe. A hint is consumed only on
positive write-back.

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
