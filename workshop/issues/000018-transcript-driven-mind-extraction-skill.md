---
id: 000018
status: working
deps: []
created: 2026-04-30
updated: 2026-04-30
---

# Transcript-Driven Mind Extraction Skill

## Problem

Claude session transcripts (the JSONL files under `~/.claude/projects/`) are
counterfactual-rich training data — they record what was proposed-and-rejected,
redirected, or quietly accepted by a tasteful judge (the user). Today this
signal is lost: corrections happen in flight, but no system distills them into
reusable skills, rules, or memories. `workshop/lessons.md` was meant to fill
this role but atrophies in practice because it asks for *synthesis* mid-session,
which is the expensive step.

This issue proposes an **ariadne base-layer skill** (working name
`/xx-mind-extract` or similar) that any ariadne-styled repo's user can invoke
to run a post-hoc extraction pass over their accumulated transcripts and
produce activity-typed `mind-<activity>` skills (mind-code-review,
mind-brainstorm, mind-planning, mind-debugging, …) plus updates to lessons.md,
AGENTS.md, permissions, and the skill inventory of the repo it's invoked from.

Cadence is user-driven: invoke biweekly, or whenever the user feels the agent
isn't picking up on patterns. Each version of the extracted minds should make
future sessions feel a little more like working with someone who knows the
user's taste, without overfitting to past work.

## Spec

### Why ariadne base layer
The pipeline is generic — it operates on `~/.claude/projects/*.jsonl`,
classifies sessions by activity, clusters corrections, and writes back into
the repo's `.claude/skills/` and `AGENTS.md`. Nothing about it is brain-
specific. Living in ariadne means every downstream repo inherits it via
`construct/setup.sh` and gets per-repo `mind-*` skills tuned to that repo's
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
   - New or updated `mind-<activity>.md` skill (situational, invoked only when
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
When cutting `mind-<activity>-v(N+1)`, also re-run v(N)'s extractor on the
*current* corpus and diff. Lets the user separate "algorithm got better" from
"more data." Without this, two confounded axes get optimized blind.

### Failure mode to watch
The extractor will lock in past strong signals. As the user's work moves into
new domains, old minds can hurt rather than help. Each version should be
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
The pipeline produces useful by-products even if no `mind-*` skill ever feels
decisively better in vibe-check:
- Personal eval set from (context, chosen, rejected) triples.
- Friction-cluster list = to-do for tooling/permission/skill cleanup.
- Activity classification of past work = self-knowledge about how time was
  actually spent.

So the win condition is "did running the skill surface useful artifacts", not
"did the produced minds measurably help" — the latter is unmeasurable given
model churn, harness churn, and changing work mix.

## Plan

Detailed design: [workshop/plans/000018-transcript-driven-mind-extraction-skill-plan.md](../plans/000018-transcript-driven-mind-extraction-skill-plan.md)

Decisions locked:
- Invocation: `/xx-mind extract` and `/xx-mind load` (umbrella skill `xx-mind` with subcommands; not a construct subcommand)
- No AGENTS.md writes — extracted taste lives in `mind-<activity>.md` files loaded on demand via `/xx-mind load`
- Friction journal: cross-repo `~/.claude/friction.md`
- Activity taxonomy v1: code-review, brainstorming, planning, debugging, implementation, exploration
- One `mind-<activity>` skill per activity, prior versions retained for diffing
- Scope (current repo / all / select) chosen at invocation; output destination derived from scope
- Clustering v1: interactive in-session with the user, no automated clusterer
- No friction journal in v1: extractor's Stage 3 `friction` signal already covers what a live journal would capture

Milestones:
- [x] M1 — Foundation: skill scaffold + normalizer (state file deferred to M2)
- [ ] M2 — Activity classifier (rule-based + LLM fallback)
- [ ] M3 — Moment detection (six detectors)
- [ ] M4 — Interactive cluster + draft generation
- [ ] M5 — Review + write-back to mind-*, memory, permissions, lessons
- [ ] M6 — `/xx-mind load` + close-the-loop
- [ ] M7 — Versioning + `--rerun-version` diff
- [ ] M8 — First real run on user's corpus; produce `mind-*-v1`
- [ ] M9 — Stabilize, add to `construct/base.manifest`, update atlas

## Log

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

### 2026-04-30 — M1 done
- Scaffolded `construct/local/mind/` (SKILL.md + scripts/normalize.py)
- Symlink `xx-mind → ../../construct/local/mind` registered via construct sync hook
- Normalizer reads JSONL line-by-line, groups by sessionId, extracts:
  cwd, gitBranch, timestamps, user/assistant message counts, tool calls (by
  name), files written/edited/read, bash command count, slash commands,
  first user message, permission modes seen, transcript file names
- Three scope paths verified end-to-end:
  - `--project charon` → 1 dir, 6 sessions, 7347 events
  - `--scope current --cwd .../ariadne` → 1 dir, 9 sessions, 917 events
  - `--scope all` not yet exercised but uses same code path as the others
- Output: `~/.claude/mind-cache/<run-id>/sessions.json` + `run.json`
- Notes for downstream stages:
  - Sessions can be 18+ hours (session-resume preserves sessionId). May want
    `away_summary` event-based subdivision later.
  - One charon session has u=2, a=0 (abandoned/never responded). M2 classifier
    should filter `assistant_message_count == 0`.
  - Git correlation (commits landed in session window) not yet wired —
    deferred until M3 detectors actually need it.
  - State file `~/.claude/mind-state.json` not yet read/written; the skill body
    is the place to do that since multiple stages mutate it.
