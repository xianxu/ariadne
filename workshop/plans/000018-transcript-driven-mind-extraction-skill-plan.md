---
issue: 000018
status: revised-v1.1
created: 2026-04-30
updated: 2026-04-30
---

## Revision (2026-04-30): heuristic detectors → LLM-direct extraction

The original M3 design used four narrow heuristic detectors (redirect, endorsement, edit-after-edit, friction) feeding a clustering stage. After dogfooding on 218 segments across 2 weeks of real work:

- 181 moments emitted, **0 clusters extracted** at the precision-over-recall threshold.
- Heuristics are *too narrow* for 2-week corpora. They miss most signal — patterns like "prefer commit messages to reference issue numbers" or "always propose a sketch before writing code" don't fit any of the four shapes.
- Lowering thresholds to force clusters out would violate the precision-over-recall principle the user invoked twice during dogfood.

**v1.1 pivot:** the primary extraction primitive becomes **LLM reading one segment at a time and emitting candidate patterns directly.** The heuristic detectors are retained as optional prompt hints ("look for these shapes too, but don't restrict to them"), not as the primary signal source.

**What survives unchanged:** segmentation (`normalize.py`), LLM-direct activity classification (Stage 3), `~/.claude/skills/mind-<activity>` write destination, user-in-the-loop principle.

**What changes:**
- `detect.py` becomes optional; its output isn't required for the pipeline.
- `view_moments.py` is no longer load-bearing.
- New stage **3.5: per-segment pattern extraction** — for each non-skip segment, an LLM reads the raw transcript and emits 0–N candidate patterns.
- New stage **4: LLM clustering** — patterns from all segments aggregated and grouped by theme.
- Threshold relaxes to `≥2 segments` for cluster qualification (the LLM has already done within-segment noise filtering).

**UNIX kit (this revision's deliverable):** rather than baking the LLM call into the skill, the building blocks are emitted as composable text-on-stdout commands so the same pipeline runs against any model — `claude`, `codex`, `gemini`. The skill body documents composition; the user (or any orchestrating agent) chooses the model.

```
construct/local/mind/scripts/segment_text.py   # one segment → human-readable transcript chunk
construct/local/mind/prompts/extract.md        # system prompt for per-segment extraction
construct/local/mind/prompts/cluster.md        # system prompt for cross-segment clustering
construct/local/mind/scripts/README.md         # composition examples for claude / codex / gemini
```

The rest of this plan reflects the heuristic pipeline as originally designed and is preserved for context. Treat anything below describing detect.py / cluster as the *baseline reference*, not the canonical flow.

---

# Plan: `/construct mind-extract`

## Decisions locked from brainstorming

- **Invocation surface**: umbrella skill `xx-mind` with subcommands. Not a construct subcommand — this is a local-authored skill at `construct/local/mind/SKILL.md`, symlinked to `.claude/skills/xx-mind/`.
  - `/xx-mind extract` — runs the extraction pipeline
  - `/xx-mind load` — loads the `mind-<activity>` matching the current session's activity (close-the-loop step from the spec; replaces the "always-on AGENTS.md" path)
- **No AGENTS.md writes.** Extracted taste lives in `mind-<activity>` files only. They're loaded situationally via `/xx-mind load` rather than always-on. (Rationale: AGENTS.md is high-blast-radius write surface; activity-typed loading scopes the influence to where it applies.)
- **No friction journal in v1**. The extractor is fundamentally a postmortem analyzer of `~/.claude/projects/*.jsonl`. The `friction` signal in Stage 3 already detects what a journal would capture (permission prompts, redirections, wasted-token ratio). A live-capture surface adds UX rabbit holes for marginal v1 value. Revisit only if first real run shows obvious gaps.
- **Activity taxonomy v1**: `code-review`, `brainstorming`, `planning`, `debugging`, `implementation`, `exploration`. Six buckets, rule-based classifier first, LLM fallback for ambiguous sessions.
- **Skill granularity**: one `mind-<activity>.md` per activity, accumulated rules. Re-extraction updates in place. Prior versions retained for diffing.
- **Scope selection at invocation time**: user picks current repo / all projects / select projects. Output destination is derived from scope (see below).
- **Clustering v1**: interactive, in-session with the user. The skill walks the user through the moment list and clusters via conversation. No automated clusterer code. Goal: build manual quality intuition before automating.

## Scope and destinations

**Inputs (scope-dependent):** user picks at invocation time.

| Scope | Transcripts read |
|---|---|
| current repo | `~/.claude/projects/<this-repo-slug>/*.jsonl` |
| all projects | `~/.claude/projects/*/*.jsonl` |
| select | union of selected dirs |

**Outputs (always user-global under `~/.claude/`):**

| Surface | Path |
|---|---|
| `mind-<activity>` skills | `~/.claude/skills/mind-<activity>/SKILL.md` |
| Permission entries | `~/.claude/settings.json` |
| Run state | `~/.claude/mind-state.json` |
| Intermediate cache | `~/.claude/mind-cache/<run-id>/` |
| Version snapshots | `~/.claude/mind-versions/vN/` |

Rationale: it's nondeterministic which repo the user has open when they want to invoke `/xx-mind`. User-global keeps everything reachable from any session and side-steps the "destination derived from scope" complication.

**v1 output surfaces are mind-* skills and permission entries only.** Memory entries and lessons.md updates are dropped from v1 — both are inherently per-project and clash with the user-global storage decision. Revisit in v2 if the v1 outputs prove valuable.

## Pipeline architecture

```
~/.claude/projects/*/*.jsonl  ─┐
git history (per touched repo)─┴──▶ [1 normalize] ──▶ events.parquet
                                                          │
                                                          ▼
                                              [2 activity-classify]
                                                          │
                                                          ▼
                                                [3 moment-detect]
                                                  (six detectors)
                                                          │
                                                          ▼
                                                  [4 cluster]
                                              (within activity bucket)
                                                          │
                                                          ▼
                                                [5 generate drafts]
                                                          │
                                              ┌───────────┼─────────────┬──────────────┐
                                              ▼           ▼             ▼              ▼
                                       mind-<a>.md   AGENTS.md     memory/      permissions
                                                          │
                                                          ▼
                                                [6 review + write]
                                                  (user gate before commit)
                                                          │
                                                          ▼
                                              [7 snapshot version]
                                            ~/.claude/mind-versions/vN/
```

### Stage 1 — Normalize

**Input**: `~/.claude/projects/*/*.jsonl` (filtered by scope selection).
**Output**: structured per-session event stream; each session tagged with repo (cwd), branch, time range, commit-landed?, reverted?, amended?.

Implementation: Python script `construct/scripts/mind-extract/normalize.py` reads JSONL line-by-line, groups by sessionId. For each session:
- start time, end time, message count, tool-call count
- cwd from session metadata or first `Bash` cwd
- repo slug derived from cwd
- git correlation: walk cwd's `git reflog`/`git log --since=<start> --until=<end>` to find commits, reverts, amends in that window. Tag session with landed-commit SHAs.

State file `~/.claude/mind-state.json`: tracks `last_run_at`, `processed_session_ids`, `version_history`. Stage 1 only emits events from sessions not in `processed_session_ids` (or after `last_run_at` for the friction journal).

### Stage 2 — Activity classify

Per-session classifier. Rule-based first pass:

| Activity | Rule signals |
|---|---|
| code-review | `/review`, `/security-review`, `/ultrareview` slash command; `gh pr` reads dominate; few file writes |
| brainstorming | `superpowers-brainstorming` invoked; user message bursts without tool calls; question marks dominate |
| planning | `EnterPlanMode`, plan-mode events; writes to `workshop/plans/`; `superpowers-writing-plans` |
| debugging | `superpowers-systematic-debugging` invoked; high re-edit ratio on same files; many test runs |
| implementation | high write/edit ratio; no plan/brainstorm signals; commits land |
| exploration | many Read/grep, low writes; no commit lands |

Sessions with conflicting signals → defer to LLM single-shot classifier with prompt: "Given these signals (slash commands, file types touched, first user message, commit-landed), classify into one of six." Cheap (Haiku-class).

### Stage 3 — Moment detection (six signal detectors)

Each detector emits `Moment(session_id, type, span, evidence_excerpt, weight)`.

1. **Counterfactual pairs** — assistant proposes edit/file-write/command, next user message contains negation/redirection (regex: `^(no|don't|stop|instead|actually|wait)\b` plus tool-denial events). Pair the proposal with the redirection text. Tool denials are a separate variant.
2. **Edit-after-edit** — same file written then re-edited within N=5 turns by assistant alone (no user message between). The diff is the taste signal.
3. **Endorsement** — user message contains `^(perfect|exactly|yes|good|nice|great)\b` immediately after assistant action; OR user accepts a proposal without modification and no further changes follow. Quiet acceptance is harder — heuristic: assistant did X, user moved to a new topic without comment, X survived to commit.
4. **Friction** — repeated permission prompts for same tool in session, recurring corrections across sessions (cross-session join needed), wasted-token ratio per session = (tokens before redirect) / (tokens total).
5. **Taste fingerprint** — diff between assistant's last write of a file and the file's state at next commit. Bucket the deltas: rename patterns, comment density change, length change.
6. **Process shape** — per-skill ROI: `(skill invoked → outcome quality)` — outcome quality proxied by commit-landed and edit-after-edit-rate. Per-tool ROI similar.

Implementation: each detector is a function in `construct/scripts/mind-extract/detect.py`. First pass deterministic; weight scoring is heuristic.

### Stage 4 — Cluster (interactive, no code)

**Resolved per Fork B**: no clusterer code in v1. The `/xx-mind extract` skill walks the user through clustering interactively in the session.

Flow:
1. Skill presents moments per activity bucket, one bucket at a time, paginated (~30 moments per page)
2. Claude proposes groupings inline with evidence quotes
3. User accepts / merges / splits / discards groups conversationally
4. Surviving groups become rule candidates for stage 5

Why this matters for v1: we need to build intuition about what *should* cluster before we trust any automated grouping. The interactive pass also generates ground-truth labels that an automated clusterer (in some future v2) could be evaluated against.

### Stage 5 — Generate drafts

Per cluster set per activity, the in-session Claude produces:
- Candidate `mind-<activity>` skill body (rules section, when-to-trigger section)
- Candidate permission entries (cross-reference with `fewer-permission-prompts` skill — same data shape)

Each candidate carries `evidence: [moment_id, ...]` provenance so the reviewer can audit.

(Memory entries and lessons.md updates removed from v1 outputs — see Scope and destinations.)

### Stage 6 — Review + write

Present diff-style to user:

```
~/.claude/skills/mind-code-review/SKILL.md (NEW)
+ Rule: ...
   evidence: 4 moments across 3 sessions in brain, ariadne
+ Rule: ...
   evidence: 3 moments in nous

Permission additions (~/.claude/settings.json)
+ Bash(gh pr view:*)   reason: 12 prompts last cycle
```

User can accept all / accept selected / reject. Accepted writes go through.

### Stage 7 — Snapshot version

After successful write:

```
~/.claude/mind-versions/v<N>/
  manifest.json           # what was written, scope, source corpus session IDs, timestamp
  extractor-snapshot.md   # the construct skill's mind-extract section at this run
  mind-<activity>.md ...  # snapshot of each mind-* skill produced
  drafts/                 # rejected candidates, for diagnosis
```

`~/.claude/mind-state.json` updated with new version pointer.

**Re-run-prior-extractor feature**: `/construct mind-extract --rerun-version <N>` reads `versions/v<N>/extractor-snapshot.md`, applies that algorithm to the current corpus, presents diff against `versions/v<N>` outputs. Lets user separate "more data" from "better algorithm".

## Friction journal capture

Out of band of the extractor itself, but the extractor is half-blind without it. Two surfaces:

- `/friction <text>` — slash command (or simple bash alias) that appends `<timestamp> <cwd> <text>\n` to `~/.claude/friction.md`.
- A skill `xx-friction-log` that triggers when the user writes informal "this annoyed me" / "claude got it wrong" mid-session — it writes the entry without ceremony.

**Open fork C**: ship `/friction` capture as part of this issue, or split into a sibling issue? Lean: **ship in this issue** — without capture there's no journal, and the extractor's evidence quality drops by half. Small surface (~30 lines + a SKILL.md). *(Need your call.)*

## Implementation milestones

Each milestone gates a code review per AGENTS.md §3. `BASE_SHA` = previous milestone close.

Code lives in `construct/local/mind/` (skill body + scripts). Symlinked to `.claude/skills/xx-mind/` via construct's local-skill mechanism.

### M1 — Foundation: skill scaffold + normalizer
- `construct/local/mind/SKILL.md` scaffold with `/xx-mind extract` and `/xx-mind load` subcommands
- Symlink registered (auto-handled by construct sync hook)
- `construct/local/mind/scripts/normalize.py` (JSONL → events, git correlation)
- Scope picker in skill body (interactive: current / all / select)
- `~/.claude/mind-state.json` schema + read/write
- **Verification (charon)**: run `/xx-mind extract` with scope=`charon`, confirm normalize output is sane on real transcripts at `~/.claude/projects/-Users-xianxu-workspace-charon/`. Eyeball events for one session.

### M2 — Activity classifier
- Rule-based classifier in `scripts/classify.py`
- LLM fallback handled by the skill itself (single in-session call per ambiguous session)
- **Verification**: classify all sessions across user's 5 project dirs, manually spot-check 10. Target: ≥80% agreement on spot-check.

### M3 — Moment detection
- Six detectors in `scripts/detect.py`
- Per-detector unit tests with synthetic JSONL fixtures
- Output: `moments.jsonl` per scope+activity
- **Verification**: detector recall on a hand-labeled set of 20 known interesting moments from prior sessions. Target: ≥70% recall on the easy four (counterfactual, edit-after-edit, endorsement, friction).

### M4 — Interactive cluster + draft (skill body work)
- Skill instructions for the in-session clustering walkthrough (per Fork B resolution)
- Pagination logic for moment review
- Draft generator instructions per output surface (mind-*.md, memory, permissions, lessons)
- Provenance tracking (evidence trail per candidate)
- **Verification**: walk through clustering on M3 output for one activity bucket. Manual review for false-positive rate.

### M5 — Review + write back
- Diff presentation as markdown in-session
- Per-surface writers (Edit/Write tool calls described in skill body)
- Settings.json permission merge logic (cross-reference `fewer-permission-prompts` patterns)
- Memory dir resolution per repo
- **Verification**: dry-run mode writes to `/tmp/mind-extract-preview/`. Real-write only after explicit user confirm.

### M6 — `/xx-mind load` + close-the-loop
- `/xx-mind load` subcommand: detects current session activity (reuse classifier from M2 over the live session), loads matching `mind-<activity>` skill via Skill tool
- Optional: SessionStart hook hint to auto-suggest `/xx-mind load` when activity becomes detectable
- **Verification**: start a new session, run `/xx-mind load`, confirm right skill loads and content is in context.

### M7 — Versioning + rerun
- Snapshot writer at `~/.claude/mind-versions/vN/`
- `/xx-mind extract --rerun-version <N>` flag
- Diff between two version outputs
- **Verification**: run extractor twice on same corpus, snapshot, then synthetically tweak skill instructions and re-run, confirm diff is sane.

### M8 — First real run + dogfood
- Run `/xx-mind extract` with scope=`charon` (manual test target) on user's full transcript corpus for that project
- Review output, accept/reject, commit produced `mind-*-v1`
- Document the run in the issue's `## Log`
- **Verification**: subjective — does the output feel useful? Are evidence trails legible?
- After dogfood passes on charon, optional second run with scope=all

### M9 — Stabilize + base.manifest
- After two more biweekly runs, if outputs hold up:
  - Add `construct/local/mind/` to `construct/base.manifest` so downstream repos inherit `xx-mind` via `construct/setup.sh`
  - Update `atlas/` with mind-extract entry

## Open forks summary

- ~~Fork A~~ — Resolved: no AGENTS.md writes; taste lives in `mind-<activity>` files loaded via `/xx-mind load`.
- ~~Fork B~~ — Resolved: interactive in-session clustering for v1, no automated clusterer.
- ~~Fork C~~ — Resolved: no friction journal in v1. Extractor's `friction` signal in Stage 3 covers what a journal would capture. Defer live-capture surface until first real run shows obvious gaps.

## Risks and mitigations

| Risk | Mitigation |
|---|---|
| Extractor surfaces noise as rules | Min-cluster size ≥3 across ≥2 sessions; user is the gate |
| Output overfits to past work | Per spec §"Failure mode": each version diff'd against new sessions; retire-rules path in the skill |
| ~~AGENTS.md bloat~~ | N/A — taste lives in activity-loaded `mind-*` skills, not AGENTS.md |
| State file corruption blocks reruns | State file is small JSON, append-only history; recovery via re-scanning corpus |
| Permission entry collisions | Reuse `fewer-permission-prompts` merge logic; never remove existing entries |
| Privacy: transcripts contain secrets | Output is local-only (no upload); evidence excerpts redacted via simple regex pass before user review |
| Cost balloons with corpus size | Stage 1-3 are local; only Stage 4-5 hit LLM API; cap candidate count per activity at 20 |

## Out of scope (per issue spec)

- Live in-session synthesis
- Frozen prompt-level eval set
- Fine-tuning / DPO
- Live friction capture (slash command / passive skill / hook) — deferred to a future iteration if v1 shows gaps

## Success criteria for v1

- `/construct mind-extract` runs end-to-end on user's corpus without manual intervention
- Produces `mind-<activity>-v1` for at least 3 of the 6 activity buckets with non-trivial rule sets
- Evidence trail is legible — user can audit any rule back to source moments
- Re-run produces deterministic output (same corpus + same extractor version → same result)
- The first real run surfaces at least one of: a useful rule, a useful permission entry, or a friction-cluster the user wants to fix
