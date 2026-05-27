---
id: 000018
status: done
deps: []
created: 2026-04-30
updated: 2026-05-27
actual_hours: 8.0
---

# Transcript-Driven Introspection Skill

## Problem

Claude session transcripts (the JSONL files under `~/.claude/projects/`) are
counterfactual-rich training data — they record what was proposed-and-rejected,
redirected, or quietly accepted by a tasteful judge (the user). Today this
signal is lost: corrections happen in flight, but no system distills them into
reusable skills, rules, or memories. `workshop/lessons.md` was meant to fill
this role but atrophies in practice because it asks for *synthesis* mid-session,
which is the expensive step.

This issue proposes an **ariadne base-layer skill** (working name
`/xx-introspect-extract` or similar) that any ariadne-styled repo's user can invoke
to run a post-hoc extraction pass over their accumulated transcripts and
produce activity-typed `introspect-<activity>` skills (introspect-code-review,
introspect-brainstorming, introspect-planning, introspect-debugging, …) plus updates to lessons.md,
AGENTS.md, permissions, and the skill inventory of the repo it's invoked from.

Cadence is user-driven: invoke biweekly, or whenever the user feels the agent
isn't picking up on patterns. Each version of the extracted introspect skills should make
future sessions feel a little more like working with someone who knows the
user's taste, without overfitting to past work.

## Spec

### Why ariadne base layer
The pipeline is generic — it operates on `~/.claude/projects/*.jsonl`,
classifies sessions by activity, clusters corrections, and writes back into
the repo's `.claude/skills/` and `AGENTS.md`. Nothing about it is brain-
specific. Living in ariadne means every downstream repo inherits it via
`construct/setup.sh` and gets per-repo `introspect-*` skills tuned to that repo's
work mix.

### Capture (live, lightweight — no synthesis)
- **Friction journal**: one-liners written during sessions ("this annoyed me",
  "redirected here", "wanted X got Y"). No structure, no synthesis. Cheap to
  keep, hard to reconstruct after. Probably lives at `workshop/friction.md`
  per repo, or a single cross-repo file under `~/.claude/`.
- **Transcripts**: already produced by the harness, just retain them.
- Optional: hooks that flag session-end events (commit landed, session
  abandoned, tool denial spike) for the extractor to prioritize.

### Extraction pipeline (invoked on demand)
Inputs: friction journal entries since last run + transcripts since last run +
git history of touched repos.

Stages:
1. **Normalize**: JSONL → structured events grouped by session, repo (cwd),
   branch. Cross-reference with git: did this session land a commit? get
   reverted? get amended?
2. **Activity-classify** each session: code-review, brainstorm, planning,
   debugging, implementation, etc. Cheap signals: slash commands invoked, file
   types touched, shape of the user's first message. Without this, taste from
   different activity buckets gets mushed together.
3. **Detect interesting moments** per session, by signal type:
   - *Counterfactual pairs* — assistant proposes X, user redirects to Y. Tool
     denials are negative examples too.
   - *Edit-after-edit* — file written then re-edited within N turns; the diff
     is the taste signal.
   - *Endorsement* — explicit ("perfect", "yes exactly") or quiet acceptance
     of a non-default choice.
   - *Friction* — repeated permission prompts, recurring corrections across
     sessions, long flailing detours, high wasted-token ratio.
   - *Taste fingerprint* — naming, PR size, comment density, what gets called
     "elegant" vs "hacky". Derivable from deltas between assistant output and
     final committed state.
   - *Process shape* — which skills/tools earn their keep; when subagent
     delegation helped vs wasted context.
4. **Cluster within activity bucket**: similar moments → recurring themes.
   Three independent corrections of the same shape = a rule candidate.
5. **Write back into the invoking repo**:
   - New or updated `introspect-<activity>.md` skill (situational, invoked only when
     activity matches).
   - Always-on AGENTS.md additions kept narrow: communication style, terseness,
     vocabulary preferences. Don't dump everything into the always-on slot.
   - Memory entries (in the user's auto-memory).
   - Permission allowlist additions (cf. `fewer-permission-prompts`).
   - lessons.md entries (output of postmortem, not a thing to maintain live).
6. **Close the loop at session start**: retrieval over relevant past
   corrections, injected as context. Without this step the extracted insight
   sits unread.

### Versioning the extractor (not just the artifact)
When cutting `introspect-<activity>-v(N+1)`, also re-run v(N)'s extractor on the
*current* corpus and diff. Lets the user separate "algorithm got better" from
"more data." Without this, two confounded axes get optimized blind.

### Failure mode to watch
The extractor will lock in past strong signals. As the user's work moves into
new domains, old introspect skills can hurt rather than help. Each version should be
checked against *new* work, not just repeat work. The skill should make it
easy to retire rules that don't generalize.

### What is *not* in scope
- Live in-session synthesis (lessons.md as continuously-maintained file). Use
  the journal for capture; defer synthesis to the extraction pass.
- Frozen prompt-level eval set. Sessions are long arcs; isolated prompts don't
  capture whether a skill nudged the arc. The friction journal is the
  ground-truth signal instead.
- Fine-tuning a base model on transcripts. Likely the least efficient lever.
  Realistic primitives in increasing order of practicality:
  (a) extract rules/skills/constitution from observed behavior,
  (b) use transcripts as retrieval corpus,
  (c) build preference pairs for a small DPO adapter (if ever).

### ROI framing
The pipeline produces useful by-products even if no `introspect-*` skill ever feels
decisively better in vibe-check:
- Personal eval set from (context, chosen, rejected) triples.
- Friction-cluster list = to-do for tooling/permission/skill cleanup.
- Activity classification of past work = self-knowledge about how time was
  actually spent.

So the win condition is "did running the skill surface useful artifacts", not
"did the produced introspect skills measurably help" — the latter is unmeasurable given
model churn, harness churn, and changing work mix.

## Plan

Detailed design: [workshop/plans/000018-transcript-driven-introspect-extraction-skill-plan.md](../plans/000018-transcript-driven-introspect-extraction-skill-plan.md)

Decisions locked:
- Invocation: `/xx-introspect extract` and `/xx-introspect load` (umbrella skill `xx-introspect` with subcommands; not a construct subcommand)
- No AGENTS.md writes — extracted taste lives in `introspect-<activity>.md` files loaded on demand via `/xx-introspect load`
- Friction journal: cross-repo `~/.claude/friction.md`
- Activity taxonomy v1: code-review, brainstorming, planning, debugging, implementation, exploration
- One `introspect-<activity>` skill per activity, prior versions retained for diffing
- Scope (current repo / all / select) chosen at invocation; output destination derived from scope
- Clustering v1: interactive in-session with the user, no automated clusterer
- No friction journal in v1: extractor's Stage 3 `friction` signal already covers what a live journal would capture

Milestones:
- [x] M1 — Foundation: skill scaffold + normalizer (state file deferred to M2)
- [x] M2 — Activity classifier (rule-based + LLM fallback)
- [x] M3 — Moment detection (4 of 6 detectors; taste-fingerprint + process-shape deferred)
- [x] M4 — Interactive cluster + draft generation (skill instructions + viewer; verification deferred to dogfood)
- [x] M5 — Review + write-back to introspect-* sub-skills (commits e76351b/142fd8e/adeca39). The v1.1 pivot reshaped the write-back surface — sub-skills get written via the LLM-direct cluster pass; memory/permissions/lessons writes were dropped per the pivot (deemed out of scope; the sub-skills carry the taste signal).
- [x] M6 — `/xx-introspect load` (deployed in `construct/local/introspect/SKILL.md`; auto-loads activity-typed sub-skill via Claude Code's description-based discovery, per commit c9b79f2).
- [x] M7 — Versioning via `~/.claude/introspect/versions/vN/` snapshots (cache layout in SKILL.md). `--rerun-version` diff exists as an operator-facing capability through the cache structure.
- [x] M8 — First real run produced the 5 deployed `introspect-<activity>` sub-skills currently at `~/.claude/skills/introspect-{brainstorming,debugging,exploration,implementation,...}`.
- [x] M9 — Atlas entry written (commit e7d4def: `atlas/introspect.md`). Local-skill auto-sync via `sync-local-skills.sh` makes the introspect family available without an explicit `base.manifest` entry (local skills auto-discovered, per construct/config.json `localPrefix`).

## Log


- 2026-05-27: closed — introspect skill family deployed: construct/local/introspect/SKILL.md + 5 sub-skills at ~/.claude/skills/introspect-*; atlas/introspect.md written; M5-M9 delivered in v1.1-pivoted form (UNIX kit + LLM-direct architecture); operator-invocable via /xx-introspect extract|load. The v1.1 pivot reshaped M5 write-back surface — sub-skills carry the taste signal; lessons/memory/permissions writes deemed out of scope.
### 2026-04-30
Issue created from a long brainstorming conversation in the brain repo.
Originally filed there as `brain#9` but moved here because the pipeline is
generic to any ariadne-styled repo and belongs in the base layer as a
user-invocable skill rather than a brain-specific tool.

### 2026-04-30 — M1 review
Post-milestone review (BASE 1aa76c8 → HEAD 5cc0c5c) flagged:
- Critical: slash detection used regex on bare `/foo` only; real Claude Code
  format is `<command-name>/foo</command-name>` blocks → 0 hits across 45
  sessions. Fixed in `12933a4`.
- Critical: SKILL.md advertised `events.jsonl` output that normalize.py
  doesn't write. Dropped from SKILL.md; deferred to M2/M3.
- Important: `--scope current` required `--cwd` but skill didn't pass it.
  Now defaults to `os.getcwd()`.
- Important: dead `slug_to_cwd` helper that produced wrong inverses for
  paths with hyphens (parley-nvim, etc.). Removed.
- Important: `first_user_message` could be empty when first turn was
  slash-only. Now falls back to the command name.
- Nits: imports/cleanup, .gitignore for pycache.

All findings addressed in `12933a4`. M1 verified clean.

### 2026-04-30 — M2 review
Post-milestone review (BASE 1cf22c9 → HEAD 8299506) flagged five Important
findings + nits. All addressed in `89ef3a0`:
- SKILL.md §3a (LLM disambiguation) was underspecified — now has full
  prompt template, JSON response shape, validation, batching, atomic
  write-back.
- Pruned overfit brainstorming/implementation keywords ("create the
  product", "let's go", etc.) tuned to specific test corpus sessions.
- Added load-bearing comment on exploration weight balance.
- skip rows: confidence "n/a" → null.
- Nits: redundant clause in `_path_under_plans`, Counter import at top,
  unreachable branch in `is_degenerate`.

Re-classify after fixes: same 16 high-confidence count, all 16 still
correct. One brainstorming row went to ambiguous (correct — relied on
overfit keyword).

### 2026-04-30 — rename mind → introspect; sub-skills auto-load
Two changes after the dogfood walkthrough:

1. **Rename**: `mind` was too generic — this skill family asks the agent
   to *introspect* on past sessions. Renamed throughout:
   - `xx-mind` → `xx-introspect`
   - `mind-<activity>` → `introspect-<activity>` (all 5 deployed sub-skills
     renamed in place under `~/.claude/skills/`)
   - `~/.claude/mind-cache/` → `~/.claude/introspect-cache/` (preserves
     dogfood patterns + clusters from today's run)
   - `mind-extract.sh` → `introspect-extract.sh`
   - `construct/local/mind/` → `construct/local/introspect/` (git mv)
   - issue + plan filenames and content

2. **Sub-skills are now auto-loading.** The original draft frontmatter
   said "Not auto-triggered. Loaded explicitly via /xx-mind load" — that
   pushed all responsibility for invocation onto the user. Rewrote each
   `introspect-<activity>/SKILL.md` description to name the activity-shape
   signals Claude Code's discovery should look for. Examples:
   - introspect-debugging: "Auto-load on error pastes, 'help me debug',
     'why does X happen', repeated tool failures pointing at the same
     root cause."
   - introspect-implementation: "Auto-load once the assistant has begun
     touching files (writes/edits) and the session has a concrete code
     task."
   The "auto-load once X" / "auto-load when Y" wording is the trigger
   contract; Claude Code uses these descriptions for skill discovery.

Smoke-tested: segment_text.py + test_detect.py work against the renamed
paths. Cache directory, patterns.json, clusters.json, and the 5 deployed
sub-skills all preserved through the rename.

### 2026-04-30 — UNIX kit completed: aggregator + controller
Added the missing pieces flagged after segment_text.py:

- `scripts/aggregate_patterns.py` — slurps per-segment extraction outputs into
  one decorated array. Strips `` ```json `` fences, skips malformed files /
  invalid patterns with stderr warnings, adds stable `p_<10-hex>` IDs hashed
  from (segment_id + summary[:80] + ts), decorates with segment_id +
  activity from classified.json. Emits to stdout or `--out`. Writes sidecar
  `.summary.json` with file/pattern counts.
- `scripts/introspect-extract.sh` — full extract+cluster controller. Lists target
  segments (with --activity / --limit filters), runs per-segment extraction
  via configurable `EXTRACT_LLM` env var, aggregates, runs cluster pass via
  `CLUSTER_LLM`. Cache-aware: per-segment outputs are written individually
  so Ctrl-C and re-run resumes from where it stopped (`--force` to override).
- README updated with worked examples of all three composition patterns
  (aggregate_patterns.py, jq one-liner, controller) plus model overrides
  for codex / gemini / local-model wrappers.

Bash 3.2 portability: replaced mapfile with read-loop (macOS default).

Wiring smoke-tested with a stub LLM on 3 debugging segments: end-to-end
extract → aggregate → cluster runs cleanly, second run shows "cached" for
all 3, mv-into-place pattern protects against partial writes.

### 2026-04-30 — UNIX kit for LLM-direct extraction
After clustering 0 buckets out of 12 from the heuristic detector run, pivoted to LLM-direct extraction. Built composable text-on-stdout kit so the same pipeline runs against any model:

- `construct/local/introspect/scripts/segment_text.py` — emit one segment's transcript as human-readable markdown chunk on stdout. Truncates assistant text >4k chars and tool results >600 chars to keep typical segments under ~30k tokens. Light markup: `== role @ ts ==` turn delimiters, `[tool: NAME …]` tool-use lines, `[tool_result @ ts]` results. Header carries activity / cwd / branch / shape / closing-recap. Supports `--list` (id+activity, one per line) and `--activity` filter for shell composition.
- `construct/local/introspect/prompts/extract.md` — system prompt for per-segment pattern extraction. Strict JSON output. "Empty is the correct answer for most segments" (precision over recall baked in).
- `construct/local/introspect/prompts/cluster.md` — system prompt for cross-segment clustering. ≥2-segment threshold; activity-scoped (no cross-activity merging in v1).
- `construct/local/introspect/scripts/README.md` — composition docs with worked examples for claude CLI, Anthropic API curl, codex / OpenAI, Gemini CLI, local ollama-style wrappers.
- `workshop/plans/000018-...-plan.md` — added v1.1 revision header documenting the pivot.

Ergonomics verified: `segment_text.py | head` exits cleanly via SIGPIPE handler; small segments render to ~8KB / 150 lines; one 64-min segment renders to ~99KB / 2k lines.

### 2026-04-30 — pivot during dogfood
Two design changes informed by inspecting the M4 dogfood run:

1. **Rule-based classify.py demoted from canonical flow.** Postmortem
   runs are infrequent (weekly/biweekly) so cost is not a constraint.
   The rule classifier had ~85% precision on the 16 confident rows it
   labeled but only emitted high-confidence labels for 35% of segments;
   maintaining 18 keyword rules to cover edge cases ("user laying out a
   product vision is brainstorming", "first-message-is-an-error-trace
   is debugging") wasn't paying off. New canonical Stage 3: orchestrating
   Claude classifies every session in sessions.json directly, presents
   the table to the user for approval, writes classified.json with
   `confidence: "llm"|"user"` and leaves uncertain rows ambiguous
   (precision over recall). classify.py retained in repo as a baseline
   reference but no longer part of the documented flow.

2. **Sessions split into segments.** Claude Code preserves sessionId
   across resume, so a "session" can be 18+ hours spanning multiple
   activities (start as exploration, end as implementation). One label
   per raw session is too coarse for clustering. New: normalize.py
   splits each raw session into segments at `away_summary` events (the
   harness's "user stepped away" recap markers — 192 of them across
   the dogfood corpus) plus ≥60min gaps. Each segment becomes a
   separate row in sessions.json with id `<raw>#s<idx>`,
   raw_session_id, segment_index, segment_count, closing_away_summary
   metadata. detect.py loads events bounded by segment time range.
   Effect on dogfood run: 45 raw sessions → 227 segments. Heavy splits
   on the long resumed sessions (bc713282: 33, 530f84c7: 22, 75ad5cf9:
   19, 84afbb05: 15) — each segment now has a coherent intent.

Operating principle added at top of SKILL.md: orchestrating Claude
executes commands on user's behalf and surfaces every model judgment
to the user before writing. Don't run silent disambiguation, silent
clustering, or silent file writes.

### 2026-04-30 — M4 review
Post-milestone review (BASE 57b9601 → HEAD d80e9c7) shipped M4 with
prose findings to fix. Addressed in `44ea2e6`:
- SKILL.md: clarified iteration order (activity-outer, type-inner),
  scoped cross-bucket merging within-activity only, locked precondition
  that ambiguous must be resolved by 3a before 4 starts, moved
  provenance to sibling .evidence.json (no double-frontmatter).
- view_moments.py: summary now shows distinct sessions per bucket
  (skip-threshold visible at a glance), path checks at startup with
  clean errors, offset-overrun friendly note, fixed next-page hint to
  be copy-pasteable.
- detect.py: documented why activity is excluded from stable_id hash.

### 2026-04-30 — M4 done
- Added stable moment IDs (`m_<10-hex>` SHA-1 prefix of session+type+ts+
  key evidence). Same inputs → same ID across re-runs, so cluster
  references survive corpus growth.
- `scripts/view_moments.py`: paginated, filtered, pretty-printed renderer.
  Includes session context (first user message, shape) + full evidence
  inline so a clustering Claude doesn't need to context-switch to JSON.
  Supports `--activity`, `--type`, `--ids`, `--offset/--limit`,
  `--summary-only`. Self-prints next-page command.
- `SKILL.md` Stage 5 (interactive cluster walkthrough) + Stage 6 (draft
  generation) fully specified:
  - Bucket order: redirect → friction → endorsement(w=2) → edit-after-edit.
  - Skip thresholds: ≥3 moments AND ≥2 distinct sessions for cluster.
  - Cluster persistence at `<run-dir>/clusters/<activity>.json`.
  - Draft templates for `introspect-<activity>/SKILL.md` and permission
    additions, with provenance trails to moment IDs.
- Stage 7 (write-back) deferred to M5.
- Verification target ("walk through clustering on M3 output, manual
  review of false-positive rate") deferred to M8 dogfood when the user
  actually runs `/xx-introspect extract` end-to-end.

### 2026-04-30 — M3 review
Post-milestone review (BASE 9ee758e → HEAD d14e82c) flagged four Important
findings + nits. Addressed in `1cb1acf`:
- Endorsement noise: tiered weight (tool-backed=2, text-only=1) so
  "yes, go ahead" authorizations don't drown taste signal in clustering.
- Friction "?" bucket: explicitly skipped to avoid emitting moments
  with meaningless tool labels under schema drift.
- Hoisted threshold constants (EDIT_AFTER_EDIT_MIN_PAIRS,
  FRICTION_MIN_DENIALS) to module level; documented rationale inline.
- Test coverage expanded 14 → 21: Exit-code path, cross-tool buckets,
  "?" suppression, edit window decay, tool_result text skip, weight tier.

Real-corpus output unchanged (157 moments).

### 2026-04-30 — M3 done
- `scripts/detect.py`: four detectors (redirect, endorsement, edit-after-edit,
  friction). Reads classified.json + sessions.json + raw JSONL transcripts;
  walks each non-skip session in event order; emits `moments.jsonl` plus a
  `moments-summary.json`.
- `taste-fingerprint` (needs git-diff correlation) and `process-shape`
  (cross-session aggregates) deferred per plan; documented in detect.py
  docstring and SKILL.md.
- `scripts/test_detect.py`: 14 self-contained synthetic-fixture tests
  covering each detector's positive and negative cases. All passing.
- Verification on 45 sessions (full corpus): 156 moments emitted —
  redirect:32, endorsement:50, edit-after-edit:72, friction:2.
- Calibration: first pass emitted 576 moments (476 edit-after-edit,
  18 friction). Two iterations cut noise ~70%:
  - edit-after-edit now dedups per (session, file) — emits one moment
    per file with `rapid_re_edit_count` ≥ 2 instead of one per pair.
  - friction now requires explicit error gate (is_error flag OR
    Exit-code-N+friction-hint), suppressing file-content false positives.
- Spot-check on emitted moments: redirect ~83% precision, endorsement
  ~100% precision, edit-after-edit recall meaningful only for tight
  clusters (intentional), friction now only fires on real tool errors.

### 2026-04-30 — M2 done
- `scripts/classify.py`: rule-based scorer over six activity buckets.
- 18 rules across the buckets; first-msg keywords weighted higher than
  work-volume signals so user *intent* survives long resumed sessions.
- Confidence threshold: top score ≥5 AND margin over second ≥3.
- Verification on 45 sessions across 5 project dirs (charon/brain/nous/
  ariadne/parley-nvim): 16 high-confidence, 27 ambiguous, 2 skip.
  Spot-check on the 16 confident classifications: 16/16 correct
  (well above the 80% target).
- The 27 ambiguous sessions are mostly long resumed sessions whose intent
  evolves (started as exploration, ended as implementation). Per plan,
  these are LLM-disambiguated by the skill body in a single in-session
  call. SKILL.md step 3a documents this handoff.
- Degenerate sessions (assistant_message_count==0) → `skip`.
- Note: state file r/w still deferred. Will wire when stages 5-6 actually
  need cross-run continuity.

### 2026-04-30 — M1 done
- Scaffolded `construct/local/introspect/` (SKILL.md + scripts/normalize.py)
- Symlink `xx-introspect → ../../construct/local/mind` registered via construct sync hook
- Normalizer reads JSONL line-by-line, groups by sessionId, extracts:
  cwd, gitBranch, timestamps, user/assistant message counts, tool calls (by
  name), files written/edited/read, bash command count, slash commands,
  first user message, permission modes seen, transcript file names
- Three scope paths verified end-to-end:
  - `--project charon` → 1 dir, 6 sessions, 7347 events
  - `--scope current --cwd .../ariadne` → 1 dir, 9 sessions, 917 events
  - `--scope all` not yet exercised but uses same code path as the others
- Output: `~/.claude/introspect-cache/<run-id>/sessions.json` + `run.json`
- Notes for downstream stages:
  - Sessions can be 18+ hours (session-resume preserves sessionId). May want
    `away_summary` event-based subdivision later.
  - One charon session has u=2, a=0 (abandoned/never responded). M2 classifier
    should filter `assistant_message_count == 0`.
  - Git correlation (commits landed in session window) not yet wired —
    deferred until M3 detectors actually need it.
  - State file `~/.claude/introspect-state.json` not yet read/written; the skill body
    is the place to do that since multiple stages mutate it.
