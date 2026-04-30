---
name: xx-mind
description: Use when the user wants to extract reusable taste signals from past Claude Code sessions, or to load a previously-extracted activity-typed mind skill into the current session. Invoked as `/xx-mind extract` (run the postmortem extraction pipeline) or `/xx-mind load` (load mind-<activity> matching the current session). Operates on `~/.claude/projects/*.jsonl` transcripts; all outputs land in user-global `~/.claude/`. See `workshop/issues/000018-...md` for full design context.
---

# xx-mind — postmortem mind extraction

Two subcommands:
- `/xx-mind extract` — run the extraction pipeline against accumulated transcripts.
- `/xx-mind load` — detect the current session's activity and load the matching `mind-<activity>` skill.

## Operating principle

**The orchestrating Claude (you, in the user's session) executes every command on the user's behalf and surfaces every model judgment to the user for approval before writing.** This skill is about extracting the user's taste — not Claude's. Don't run silent disambiguation, silent clustering, or silent file writes. Don't ask the user to copy-paste shell invocations either; you have the Bash tool, run them yourself.

The only steps that don't need user approval are deterministic, no-op-on-failure reads (normalize, classify, detect, view). Everything that involves judgment (3a disambiguation, 5 clustering, 6 drafting) or that writes user-facing artifacts (7 write-back) is a checkpoint where the user decides.

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

#### 3a. Disambiguate `ambiguous` rows (with user in the loop)

The `ambiguous` rows are the ones rule-scoring couldn't decide. You — the orchestrating Claude in the user's session — resolve them by reasoning over each session's evidence and surfacing your proposed bucket to the user. The user is the final judge; do not write back silently.

**Procedure:**
1. Load `classified.json`, find rows with `activity: "ambiguous"`.
2. For each ambiguous row, pull the corresponding session record from `sessions.json` for context (first user message, slash commands, tool counts, file activity, top_two rule guesses).
3. Reason about the right bucket for each, picking from `code-review`, `brainstorming`, `planning`, `debugging`, `implementation`, `exploration`, or `out-of-scope` if the session is genuinely off-taxonomy (personal/non-code work).
4. Present the full table to the user — one row per ambiguous session — with: short id, first user message excerpt, your proposed activity, and a one-line rationale citing the strongest signal you used. Format compactly so the table fits on a screen.
5. Ask the user to accept-all, accept-with-edits, or override specific rows. Apply their corrections.
6. Atomically write back to `classified.json`: write `classified.json.tmp`, then `mv` into place. Set `confidence: "llm"` for unedited rows, `confidence: "user"` for rows the user overrode, and **leave the original `ambiguous`/`low`** unchanged for any row where neither you nor the user feels confident — precision over recall.

**Heuristic for your bucket choice:** the first user message is the primary intent signal; tool/file counts are secondary. Long sessions that started as exploration but did real work should still go to the originating intent — the rule-scoring already biases toward intent for exactly this reason.

**Skip out-of-scope rows downstream.** Stage 4+ should treat `out-of-scope` and `unknown` the same as `skip`.

### 4. Moment detection

```
python3 $REPO_ROOT/construct/local/mind/scripts/detect.py \
  --cache-dir ~/.claude/mind-cache/<run-id>/
```

Walks the raw JSONL for each non-skip session in `classified.json`, runs four detectors, emits `moments.jsonl` (one record per line) plus `moments-summary.json`.

**Detector types:**
- `redirect` — user negates/redirects after assistant action
- `endorsement` — user reacts positively to assistant action
- `edit-after-edit` — assistant re-edits same file ≥3 times within 5-turn window with no user message between (one moment per file with count, not per pair)
- `friction` — same tool gets ≥3 explicit errors (`is_error: true` or `Exit code N` + friction-keyword)

Two more detectors (`taste-fingerprint` requires git-diff correlation; `process-shape` requires cross-session aggregation) are deferred.

Each moment carries `{session_id, project_slug, activity, type, ts, weight, evidence}`. The `evidence` shape is type-specific.

### 5. Interactive cluster walkthrough (in-session, with the user)

This stage is a guided conversation. Do not write code that auto-clusters. The point of v1 is to build user-confirmed clusters by hand so we know what *should* group together before automating.

**Preconditions:**
1. Stage 3a has run. After 3a, rows have one of: a six-taxonomy activity, `out-of-scope`, `unknown`, `ambiguous`, or `skip`. **`ambiguous` is allowed to persist** — for any row where the user (or you reasoning on their behalf) couldn't confidently pick a bucket, leaving it ambiguous is correct. Precision over recall: clustering operates on signal we trust, not signal we hope.
2. `out-of-scope`, `unknown`, `ambiguous`, and `skip` rows are all filtered out before the cluster loop begins. Their moments are excluded from any `mind-<activity>` skill draft.

**Iteration order: outer loop is activity, inner loop is type.**

Process each in-taxonomy activity in descending session-count order (most data first). Within each activity, walk type buckets in this order — highest taste signal first:

1. `redirect` — explicit user correction
2. `friction` — actionable tool/permission failures
3. `endorsement` with `weight=2` — tool-backed acceptance (skip `weight=1` text-only rows in v1)
4. `edit-after-edit` — only cluster if a recurring file/area pattern is visible

Skip `(activity, type)` buckets that have fewer than 3 moments OR fewer than 2 distinct sessions — see "Skip thresholds" below.

**Pagination loop:**

For each `(activity, type)` bucket with ≥3 moments:

```
python3 $REPO_ROOT/construct/local/mind/scripts/view_moments.py \
  --cache-dir <run-dir> \
  --activity <activity> \
  --type <type> \
  --limit 12
```

Read the page. For each page, propose 1-3 candidate cluster names that group similar moments, citing moment IDs. Format:

```
Cluster proposal 1: "user pushes back when assistant assumes file structure without checking"
  evidence: [m_4c6e82bd4b, m_1bfd21350a, m_4b20568e3c]
  rule sketch: Before writing to a path, verify the file exists and the
    enclosing directory layout matches what the user expects.
```

Ask the user to: (a) accept, (b) merge with another proposal, (c) split off a moment, (d) discard. After each page, page forward (`--offset`) until the bucket is exhausted.

**Cross-bucket merging:** at the end of an activity (after walking all four types), ask the user whether any cross-type clusters within this activity should merge (e.g., a `redirect` cluster about "verify before writing" and a `friction` cluster about Bash permission failures might both signal "check before acting"). Merged clusters keep one combined `moment_ids` list and stay assigned to the current activity — there is no cross-activity merging in v1, since each activity will produce its own `mind-<activity>` skill anyway.

**Persist clusters:** at end of each activity, write the accepted cluster set to `<run-dir>/clusters/<activity>.json`:

```json
{
  "activity": "implementation",
  "clusters": [
    {
      "id": "c_impl_1",
      "name": "...",
      "rule_sketch": "...",
      "moment_ids": ["m_4c6e82bd4b", "m_1bfd21350a", ...],
      "moment_count": 3,
      "session_count": 2
    }
  ]
}
```

**Skip thresholds:**
- Skip clusters with fewer than 3 moments OR fewer than 2 distinct sessions. Per plan: three independent corrections of the same shape = a rule candidate. Two from one session = within-session correction, not yet a recurring pattern.

### 6. Draft generation (in-session)

For each activity that has ≥1 accepted cluster, draft:

**`~/.claude/skills/mind-<activity>/SKILL.md`** (only the draft — write-back is Stage 7):

```markdown
---
name: mind-<activity>
description: Use when the current session is doing <activity> work — extracted from past sessions where the user redirected, endorsed, or struggled. Loaded by /xx-mind load when activity is detected.
version: <N>
generated_from_run: <run-id>
generated_at: <iso-ts>
---

# Notes from past <activity> sessions

## Rule: <rule name from cluster>

<the rule, written as a directive to a future Claude. Include the *why*
when the cluster's evidence makes it clear.>

**Evidence:** `<moment-id>`, `<moment-id>`, ... (3 moments, 2 sessions)

## Rule: <next>
...
```

**Permission additions** (one entry per friction cluster targeting the same tool/command):

```json
{
  "permissions": {
    "allow": [
      "Bash(gh pr view:*)",
      ...
    ]
  }
}
```

Each permission entry carries an inline comment-style note in the draft showing the friction count and example error, so the user can audit.

**Provenance file:** `<run-dir>/drafts/<activity>.evidence.json` (sibling, NOT inside the SKILL.md draft — two YAML frontmatters in one file is malformed). Schema:

```json
{
  "activity": "implementation",
  "rules": [
    {
      "rule_name": "...",
      "moment_ids": ["m_4c6e82bd4b", "m_1bfd21350a", ...],
      "moment_excerpts": [
        {"id": "m_4c6e82bd4b", "type": "redirect", "session": "74cf212a", "excerpt": "..."}
      ]
    }
  ]
}
```

The user can audit any rule back to its source moments before accepting in Stage 7.

### 7. Stage 7 (write-back) — M5

Not yet implemented. After drafts are generated, present diff-style and let the user accept/reject per cluster. Accepted drafts get atomic-write to `~/.claude/skills/mind-<activity>/SKILL.md` and `~/.claude/settings.json`.

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
