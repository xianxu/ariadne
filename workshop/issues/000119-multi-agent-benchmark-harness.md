---
id: 000119
status: working
deps: []
github_issue:
created: 2026-06-18
updated: 2026-06-18
estimate_hours:
started: 2026-06-18T23:32:37-07:00
---

# Multi-agent benchmark harness

## Problem

Ariadne has been developed and tested almost entirely against Claude. To claim
the ecosystem is genuinely *multi-agent ready*, we need a controlled way to
measure how well a different coding agent (Codex, Antigravity, Gemini, …) does
**real ariadne work** — not a synthetic coding benchmark, but the actual tasks
in our backlog, judged on both the quality of the result *and* how cleanly the
agent rides the ariadne workflow (claims the issue, plans, creates the right
artifacts, reasons about design subtleties). There is no such harness today.

## Spec

A **controlled experiment harness**: hold everything constant (task spec,
starting commit, environment, grading rubric, judges) and vary one thing —
**which agent**. The reusable asset is the agent-neutral rig, not any single
comparison.

### Decisions (from brainstorm, #119)

1. **What it measures — both, equally.** Each agent gets two scores, reported
   side by side: *output quality* and *workflow-fit*. Neither is the headline.
2. **Cadence — standing suite from day one.** A curated, replayable task library
   with a leaderboard that accumulates over time. Implies pinned base SHA,
   pinned agent versions, and stable/reproducible scoring.
3. **Task source — freeze live backlog issues.** Tasks are real, currently-open
   issues frozen at claim time (spec snapshot + base SHA). No run-time oracle;
   the eventual real solution backfills as an optional *reference* later.
   (Replaying solved history and synthetic probe-tasks were explicitly *not*
   chosen as the seed.)
4. **Grading — hybrid (measure + compare).** Objective dimensions are MEASURED
   automatically and agent-neutrally; subjective dimensions are judged
   HEAD-TO-HEAD blind by an LLM judge + the operator.
5. **Run mode — a first-class dimension, all three supported.** `autonomous`
   (headless, day-one build, leaderboard feeder), `interactive` (pinned
   responder — seam designed now, built later), `live` (operator pairing —
   exhibition only, kept out of the leaderboard).

### Framing principle

**The harness is base-layer (`cmd/sdlc`); the task suite + results are
repo-local (`workshop/benchmarks/`).** So `sdlc bench` propagates to every
ariadne-styled repo, letting any repo benchmark agents on *its own* real work,
while each repo curates its own task library. Reuse over rebuild: the runner and
grader lean on infra that already exists — `judge.Dispatch` (already runs
`claude -p` / `codex exec` / `gemini -p` symmetrically, fresh-subprocess,
anti-collusion), `createWorktreeBranch`, `gitx`, the `VERDICT` output contract.

### Pipeline

`freeze` (snapshot a live issue → immutable task) → `run` (fan out N agents into
isolated worktrees, per mode, no merge) → `grade` (Stage A measure + Stage B
blind head-to-head) → `report` / `leaderboard` (accumulate).

### Component 1 — Data model

Three version-controlled artifacts under `workshop/benchmarks/`:

- **Frozen task** (`tasks/<slug>.md`) — immutable experiment definition. New
  datatype `construct/datatype/benchmark-task.md`. Fields: `base_sha` (the
  reproducibility anchor — agents branch from it, never merge, so it survives
  main moving on), `repo`, `source_issue`, `spec` (the verbatim task prompt
  handed to *every* agent), `setup` (worktree prep), `rubric` (objective checks +
  subjective dimensions + weights), `reference` (optional, backfilled once the
  real issue is solved normally).
- **Run record** (`runs/<task>-<runid>.md`) — one per (task × agent-set ×
  execution). Per agent: branch, commits, diffstat, transcript path, wall-clock,
  exit status, **pinned agent CLI version**, plus the objective scorecard and the
  judge + operator verdicts.
- **Leaderboard** (`leaderboard.md`, generated) — aggregates run records:
  per-agent objective-metric distributions + head-to-head win-rates per
  subjective dimension, keyed by (agent, version).

### Component 2 — The runner

- **`freeze`** copies the issue's `## Spec` verbatim, pins `base_sha = main
  HEAD`, scaffolds the rubric. Task is immutable thereafter.
- **`run`**, per agent: `git worktree add` from `base_sha` on branch
  `bench/<task>/<agent>/<runid>` (reuses `createWorktreeBranch`); run `setup` to
  start green; dispatch with a **constant prompt** — *"Solve issue N following
  this repo's conventions. Commit your work. Do NOT merge or open a PR — this
  branch is the deliverable."* Capture branch/commits/diffstat/transcript/
  wall-clock/CLI-version/exit.
- **What this tests:** we do **not** spoon-feed the SDLC — the agent must
  discover `AGENTS.md`/skills as it would in real life. Same prompt for everyone;
  only the weave'd AGENTS.md *face* differs per agent (the intended confound).
- **Responder seam** — the abstraction that lets all three modes share one
  runner; it answers an agent's questions at a gate:
  - *autonomous* → no responder; headless `judge.Dispatch` (full-auto).
  - *interactive* → a **pinned** responder answers every agent identically (a
    user-simulator LLM seeded with the issue context, or a per-task canned-answer
    script), wired via an `ask_user` tool/MCP the harness backs.
  - *live* → the operator answers; flagged non-reproducible, leaderboard-excluded.
- **No-merge is structural**, not just instructed: harness never merges, branch
  has no push target, each agent isolated in its own worktree. The branch *is*
  the artifact.

### Component 3 — The grader

- **Stage A — objective signals (measured, no LLM, agent-neutral, deterministic),
  run in each worktree:**
  - *Quality:* build passes? pre-existing tests still green (regression)? new
    tests added — and passing? spec completed or stopped partway?
  - *Workflow-fit:* required artifacts present (`## Log` updated, plan ticked,
    atlas touched for new surface)? SDLC gates actually run (detectable from
    commit conventions + `Review-Verdict:` trailers in the branch history)?
  - *Metrics:* wall-clock, diffstat, commit count, turn count, token cost,
    completed-vs-partial.
- **Stage B — subjective head-to-head (blind):** for code elegance
  (DRY/PURE/root-cause-vs-hack), **UI/design-subtlety reasoning** (read from the
  agent's spec+plan+diff), doc quality, gate judgment.
  - Each solution → **anonymized submission** (diff + artifacts + transcript
    summary), all identity scrubbed (`Co-Authored-By` trailers, `bench/<agent>/…`
    branch names, self-references), relabeled Submission A/B/C in randomized
    order.
  - **LLM judge:** reuses `judge.Dispatch` + a new benchmark prompt on the
    existing `VERDICT` contract, extended to emit a structured per-dimension
    ranking + rationale + confidence. The judge is a **fixed, pinned model that
    is NOT one of the contestants** (anti-collusion); optionally an ensemble.
  - **Operator judge:** same anonymized packet as a side-by-side review doc;
    records per-dimension rankings/notes via the `🤖` review-convention markers.
    LLM verdict withheld until the operator submits (anti-anchoring), then
    de-anonymize and fold both into the run record.
- **Anonymization is load-bearing** — if a judge can spot Claude, the comparison
  is biased. Diff+artifacts scrub cleanly; transcripts leak style, so the
  **primary judging surface is diff + artifacts**, transcript only as a
  summarized supplement.
- **The rubric** (per-task, with defaults) maps every dimension into the two
  groups — *output quality* and *workflow-fit* — yielding objective scorecard
  (absolute) + head-to-head rankings (comparative).

### Component 4 — CLI surface

`sdlc bench` verb family (base-layer, mirrors the `claim → … → close` cadence):

| Verb | Does |
|---|---|
| `sdlc bench freeze --issue N` | Snapshot live issue → immutable task |
| `sdlc bench run --task <slug> --agents claude,codex [--mode autonomous] [--parallel]` | Fan out worktrees, dispatch per mode, capture |
| `sdlc bench grade --run <id>` | Stage A measure + anonymize + dispatch LLM judge; emit operator review doc |
| `sdlc bench review --run <id>` | Await operator rankings, then de-anonymize + merge verdicts |
| `sdlc bench report --run <id>` / `sdlc bench leaderboard` | Aggregate → `leaderboard.md` |

Net-new code is one package, `cmd/sdlc/internal/bench/` (task model, runner,
grader, anonymizer, aggregation) + prompts. Everything else is reuse: `judge`,
`gitx`, `branchcreate`, `issue`.

### Rigor controls (recorded per run)

Identical verbatim prompt · identical `base_sha` + setup · pinned agent CLI
versions · pinned judge model · anonymized + randomized-order judging · **each
agent's substrate face recorded** (the intended confound, made explicit) · mode
recorded, live flagged non-leaderboard · turn/time budget.

### Deferred to the plan / open risks

- Interactive-mode `ask_user` interception mechanism (built when interactive is).
- Transcript-scrub imperfection — mitigated by diff+artifacts as primary surface.
- Headless completion of large tasks — task-sizing + partial grading.
- Cross-agent effort metric is **wall-clock/turns, not `sdlc actual`**
  (active-time is Claude-transcript-only — asymmetric across agents).
- Whether the *leaderboard* migrates to brain when it goes cross-repo (the
  harness stays base-layer regardless).

## Done when

- `sdlc bench freeze/run/grade/report` run end-to-end in **autonomous** mode on
  one real frozen backlog issue, comparing **claude vs codex**, producing a
  graded run record + a leaderboard entry.
- Objective Stage A scorecard is computed deterministically and agent-neutrally
  (build/tests/artifacts/metrics).
- Stage B produces an anonymized head-to-head packet, an LLM-judge verdict, and
  an operator review doc; `review` de-anonymizes and records both.
- `base_sha` immutability + no-merge isolation verified (a frozen task replays
  identically after main has advanced).
- `benchmark-task` datatype documented; `workshop/benchmarks/` layout +
  `sdlc bench --help` documented; atlas updated.
- The `interactive`/`live` responder seam exists in the runner (interface +
  autonomous impl), even though only autonomous is wired.

## Plan

_To be detailed via `superpowers-writing-plans` after `start-plan`. Milestone
skeleton (review boundaries):_

- [ ] M1 — Data model + `benchmark-task` datatype + `workshop/benchmarks/` layout + `freeze`
- [ ] M2 — Runner (autonomous mode) + responder seam interface + no-merge isolation
- [ ] M3 — Grader Stage A (objective scorecard) + metrics capture
- [ ] M4 — Grader Stage B (anonymizer + LLM judge + operator review doc) + `review`
- [ ] M5 — `report` / `leaderboard` aggregation + end-to-end claude-vs-codex demo + atlas

## Log

### 2026-06-18

- Brainstormed via `superpowers-brainstorming` (#119). Four scoping decisions +
  run-mode decision captured in `## Spec`. Mapped reusable infra first
  (`judge.Dispatch` already multi-agent; `createWorktreeBranch`; `VERDICT`
  contract; `sdlc actual` is Claude-transcript-only → use wall-clock cross-agent).
  Spec written; fresh-eyes spec review next.
