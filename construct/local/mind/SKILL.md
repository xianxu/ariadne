---
name: xx-mind
description: Use when the user wants to extract reusable taste signals from past Claude Code sessions, or to load a previously-extracted activity-typed mind skill into the current session. Invoked as `/xx-mind extract` (run the postmortem extraction pipeline) or `/xx-mind load` (load mind-<activity> matching the current session). Operates on `~/.claude/projects/*.jsonl` transcripts; all outputs land in user-global `~/.claude/`. See `workshop/issues/000018-...md` for full design context.
---

# xx-mind — postmortem mind extraction

Two subcommands:
- `/xx-mind extract` — run the extraction pipeline against accumulated transcripts.
- `/xx-mind load` — detect the current session's activity and load the matching `mind-<activity>` skill.

## Storage layout (all user-global)

```
~/.claude/skills/mind-<activity>/SKILL.md  # produced output, loaded on demand
~/.claude/mind-state.json                  # run history + processed-session pointers
~/.claude/mind-cache/<run-id>/             # intermediate stages of one run
~/.claude/mind-versions/vN/                # post-run snapshots for diffing
~/.claude/settings.json                    # permission entries written here
```

## `/xx-mind extract`

### 1. Scope picker

Ask the user which transcripts to read. Three options:

```
[1] current repo  → ~/.claude/projects/<repo-slug-of-cwd>/*.jsonl
[2] all projects  → ~/.claude/projects/*/*.jsonl
[3] select        → list project dirs, user picks subset
```

If `cwd` doesn't have a corresponding `~/.claude/projects/<slug>/` (slug = cwd path with `/` → `-`), `current repo` is unavailable and the user must pick from `all` or `select`.

For dogfood/testing, the user may pass an explicit slug: `/xx-mind extract --project charon` (resolves to `-Users-xianxu-workspace-charon`).

### 2. Run normalize

```
python3 $REPO_ROOT/construct/local/mind/scripts/normalize.py \
  --scope <choice> \
  [--project <slug>] \
  [--cwd "$PWD"]                   # only when --scope current; defaults to os.getcwd()
  [--since <last_run_at-from-state>] \
  --out ~/.claude/mind-cache/<run-id>/
```

Outputs:
- `sessions.json` — one record per session: id, start, end, cwd, gitBranch, message counts, tool counts, slash commands invoked, files touched
- `run.json` — meta-record of the run (scope, projects, file/event counts, since filter)

A flat `events.jsonl` stream will be added in M2/M3 once detectors need to walk events outside of session aggregates. Until then the raw JSONL files remain the source of truth for downstream stages.

Run-id format: `YYYYMMDDTHHMMSS`.

### 3. Activity classify

```
python3 $REPO_ROOT/construct/local/mind/scripts/classify.py \
  --in  ~/.claude/mind-cache/<run-id>/sessions.json \
  --out ~/.claude/mind-cache/<run-id>/classified.json
```

Output is a list of `{session_id, activity, confidence, scores, evidence, top_two}` records.

`activity` values:
- one of `code-review`, `brainstorming`, `planning`, `debugging`, `implementation`, `exploration` — confidently rule-classified
- `ambiguous` — rules didn't produce a clear winner; the orchestrating skill should disambiguate via a single LLM call (see step 3a below)
- `skip` — degenerate session (e.g., no assistant messages); excluded from downstream stages

#### 3a. Disambiguate `ambiguous` rows

The `ambiguous` rows are the ones rule-scoring couldn't decide. Resolve them in a **single batched in-session call**, not row-by-row, to keep cost low.

**Build the prompt:**

For each ambiguous row, gather a compact record from `sessions.json` (look up by `session_id`):

```
{
  "id": "<short-prefix-of-session-id>",
  "first_user_message": "<first 400 chars>",
  "slash_commands": [...],
  "tool_calls_by_name": {...},
  "files_written": <count>,
  "files_edited": <count>,
  "files_read": <count>,
  "user_message_count": ...,
  "assistant_message_count": ...,
  "rule_scores": <classified.scores>,
  "rule_top_two": <classified.top_two>
}
```

Send this prompt:

> Classify each session below into exactly one of these six activity buckets:
> `code-review`, `brainstorming`, `planning`, `debugging`, `implementation`, `exploration`.
>
> If the session truly doesn't fit any (e.g., personal/non-code content), return `out-of-scope`.
>
> Use the first_user_message as the primary intent signal; tool/file counts as
> secondary. Long sessions often start as one activity and evolve — go with the
> *originating* intent unless the user explicitly redirected.
>
> Output strict JSON: `[{"id": "<id>", "activity": "<bucket>"}, ...]`.
> One object per session, no prose.
>
> Sessions:
> ```json
> [<records>]
> ```

**Parse the response:**
- JSON array, one object per ambiguous row.
- Validate `activity` is one of the seven legal values (six buckets + `out-of-scope`).
- If a value is illegal or the row is missing → fall back to `unknown` and log it.

**Write back:**
- Update `classified.json` in place: for each disambiguated row, set `activity` to the LLM's choice and `confidence` to `"llm"` (or `"llm-out-of-scope"` / `"llm-unknown"` for the fallback cases).
- Atomically: write to `classified.json.tmp`, then `mv` into place.

**Batch sizing:**
- Up to ~25 sessions per call. If more than 25 ambiguous rows, split into batches and merge results before write-back. Never re-classify a row twice.

**Skip out-of-scope rows downstream.** Stage 4+ should treat `out-of-scope` and `unknown` the same as `skip`.

### 4. Stages 3-7 (M3 onwards)

Not yet implemented. After classification, present a summary:

```
Processed N sessions, classified into:
  code-review:    X
  brainstorming:  Y
  ...
Skipped Z degenerate sessions.
Wrote: ~/.claude/mind-cache/<run-id>/classified.json
Next stages (detect, cluster, draft, write) not yet implemented.
```

## `/xx-mind load`

Not yet implemented (M6). Placeholder: report "load subcommand pending — once mind-<activity> skills exist at ~/.claude/skills/, this will detect activity and Skill-invoke the right one."

## State file schema

`~/.claude/mind-state.json`:

```json
{
  "schema_version": 1,
  "last_run_at": "2026-04-30T18:00:00Z",
  "processed_session_ids": ["uuid1", "uuid2"],
  "runs": [
    {
      "id": "20260430T180000",
      "ts": "2026-04-30T18:00:00Z",
      "scope": "charon",
      "session_count": 6,
      "version_pointer": null
    }
  ]
}
```

Initialize as `{"schema_version": 1, "last_run_at": null, "processed_session_ids": [], "runs": []}` if the file doesn't exist.

## Key rules

- All outputs land under `~/.claude/`. Never write to repo-local `.claude/skills/` from this skill.
- Never overwrite an existing `mind-<activity>` skill without an explicit user accept.
- The `mind-cache/<run-id>/` directory is keep-forever for now (small JSON). M7 versioning will introduce pruning.
- For the M1 implementation, only stage 1 (normalize) runs. Stages 2-7 should be scaffolded as TODOs in the skill body, not silently no-op.
