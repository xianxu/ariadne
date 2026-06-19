---
id: 000119
status: working
deps: []
github_issue:
created: 2026-06-18
updated: 2026-06-19
estimate_hours: 7.4
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
5. **Run mode — a first-class dimension; all three accommodated by the responder
   seam, only `autonomous` wired day-one.** `autonomous` (headless, day-one
   build, leaderboard feeder), `interactive` (pinned responder — seam designed
   now, built later), `live` (operator pairing — exhibition only, kept out of
   the leaderboard).

### Framing principle

**The harness is base-layer (`cmd/sdlc`); the task suite + results are
repo-local (`workshop/benchmarks/`).** So `sdlc bench` propagates to every
ariadne-styled repo, letting any repo benchmark agents on *its own* real work,
while each repo curates its own task library. Reuse over rebuild — strongest
pillars: `judge.Dispatch` (the subprocess shim already runs `claude -p` /
`codex exec` / `gemini -p` symmetrically, fresh-subprocess, anti-collusion) and
`gitx`. Two reuses are **partial** and need small changes, flagged at point of
use: `createWorktreeBranch` hardcodes `HEAD` (needs a `base` ref param), and the
`VERDICT` *classifier* is a binary gate that cannot carry Stage-B per-dimension
rankings (only the dispatch shim is reused there, not the contract).

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
  `bench/<task>/<agent>/<runid>` (**extends `createWorktreeBranch`**, which today
  hardcodes `HEAD`, to take a `base` ref — required for the immutability
  guarantee); run `setup` to start green; dispatch with a **constant prompt** —
  *"Solve issue N following this repo's conventions. Commit your work. Do NOT
  merge or open a PR — this branch is the deliverable."* Capture branch/commits/
  diffstat/transcript/wall-clock/CLI-version/exit. **Note:** autonomous dispatch
  is the *first* `judge.Dispatch` caller that needs a write/edit/commit
  `AllowedTools` allowlist **and** a real `context.WithTimeout` turn/time budget —
  every current caller is read-only with no timeout, so this is new plumbing, not
  free reuse.
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
  - **LLM judge:** reuses the `judge.Dispatch` subprocess shim + a new benchmark
    prompt. Stage B output is **net-new structured parsing** (per-dimension
    ranking + rationale + confidence) — *not* the `VERDICT` binary-gate
    classifier, which can't carry it; do not try to extend that contract. The
    judge is a **fixed, pinned model that is NOT one of the contestants**
    (anti-collusion); optionally an ensemble.
  - **Operator judge:** same anonymized packet as a side-by-side review doc;
    records per-dimension rankings/notes via the `🤖` review-convention markers.
    LLM verdict withheld until the operator submits (anti-anchoring), then
    de-anonymize and fold both into the run record.
- **Anonymization is load-bearing** — if a judge can spot Claude, the comparison
  is biased. The transcript leaks style most, so the **primary judging surface is
  diff + artifacts**, transcript only as a summarized supplement. But coding
  *style* (commit-message phrasing, comment density, `🤖` markers, plan shape)
  leaks identity **in the diff itself** — "diff scrubs cleanly" is the single
  assumption most worth proving, not asserting. So the plan must include an
  explicit **anonymization-leak test** (can the pinned judge name the contestant
  from a scrubbed packet above chance?) as a Stage-B acceptance gate. Note too
  that Stage A reads the same git-history surface (commit conventions, trailers)
  that leaks identity — the scrub and the artifact-detection overlap.
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
grader, **Stage-B structured-output parser**, anonymizer, aggregation) +
prompts + per-run **agent CLI-version capture** (`claude --version` etc. —
nothing captures this today). Reuse: `judge.Dispatch` + `gitx` as-is;
`branchcreate` with a new `base`-ref param; `freeze`'s `## Spec` extraction can
lean on the section-anchoring regex already in `close.go`. The `VERDICT`
classifier is *not* reused.

### Rigor controls (recorded per run)

Identical verbatim prompt · identical `base_sha` + setup · pinned agent CLI
versions · pinned judge model · anonymized + randomized-order judging · **each
agent's substrate face recorded** (the intended confound, made explicit) · mode
recorded, live flagged non-leaderboard · turn/time budget.

### Deferred to the plan / open risks

- Interactive-mode `ask_user` interception mechanism (built when interactive is).
- Transcript-scrub imperfection — mitigated by diff+artifacts as primary surface.
- Headless completion of large tasks — task-sizing + partial grading.
- **Nested dispatch.** A contestant running the full SDLC will invoke
  `sdlc milestone-close`/`close`, which *themselves* dispatch a fresh-context
  *review* agent via `judge.Dispatch` (`$AGENT_CMD`). For a non-Claude
  contestant this can mis-attribute the review or hang headless. Decide whether
  sub-dispatched reviews are disabled or pinned during a bench run.
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
- **Anonymization-leak gate passes:** the pinned judge cannot name the contestant
  agent from a scrubbed Stage-B packet above chance — verified as a test, not
  assumed (coding style leaks identity in the diff, not just the transcript).
- `benchmark-task` datatype documented; `workshop/benchmarks/` layout +
  `sdlc bench --help` documented; atlas updated.
- The `interactive`/`live` responder seam exists in the runner (interface +
  autonomous impl), even though only autonomous is wired.

## Estimate

```estimate
model: estimate-logic-v2
familiarity: 1.0
item: greenfield-go-module    design=0.3 impl=0.7
item: greenfield-go-module    design=0.3 impl=0.8
item: smaller-go-module       design=0.2 impl=0.7
item: greenfield-go-module    design=0.3 impl=0.9
item: smaller-go-module       design=0.2 impl=0.6
item: atlas-docs              design=0.1 impl=0.4
item: milestone-review        design=0.0 impl=0.3
item: milestone-review        design=0.0 impl=0.3
item: milestone-review        design=0.0 impl=0.3
item: milestone-review        design=0.0 impl=0.3
item: milestone-review        design=0.0 impl=0.3
design-buffer: 0.30
total: 7.4
```

Derivation: one greenfield Go package (`cmd/sdlc/internal/bench/`) built across
five milestones. The three `greenfield-go-module` items are the heavier
single-concern cores — M1 data model + `freeze`, M2 runner/worktree/responder
seam + `run`, M4 grader Stage B (anonymizer + blind judge + leak gate, the
trickiest). The two `smaller-go-module` items extend that core — M3 Stage A
scorecard/measurer + `grade`, M5 leaderboard/report. `atlas-docs` covers the
`benchmark-task` datatype + atlas. Five `milestone-review` items are the
SDLC boundary-review overhead (auto-dispatched, so 0.3 each, not a full manual
pass). recomputed = 1.4×1.30 + 5.6 = 7.42 ≈ total 7.4.

## Plan

Detailed in **`workshop/plans/000119-multi-agent-benchmark-harness-plan.md`**
(Core Concepts → 5 milestone chunks, bite-sized TDD). Milestone skeleton (review
boundaries):

- [x] M1 — Data model + `benchmark-task` datatype + `workshop/benchmarks/` layout + `freeze`
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
- Fresh-eyes spec review (general-purpose subagent, verified reuse claims against
  code): **✅ Approved**. Folded in corrections — two overstated reuse claims
  fixed: `createWorktreeBranch` hardcodes `HEAD` (needs `base` param);
  `VERDICT` classifier is binary-gate-only so Stage-B ranking parsing is net-new
  (not a contract extension). Softened decision #5 wording to match autonomous-
  only day-one scope. Added: anonymization-leak gate to Done-when (coding style
  leaks in the diff, not just transcript); nested-dispatch risk (contestant runs
  `sdlc` which dispatches its own review agent); autonomous dispatch is the first
  `judge.Dispatch` caller needing write allowlist + real timeout plumbing.

### 2026-06-19
- 2026-06-19: closed M1 — M1 data model + freeze: bench package (Task/Rubric config-scoped json round-trip + Store), benchmark-task datatype, sdlc bench freeze registered in main.go + smoke-froze #119 end-to-end; all bench + cmd/sdlc tests green, go build ./... clean; atlas/workflow/bench.md added; review verdict: FIX-THEN-SHIP

- `start-plan` → durable plan written via `superpowers-writing-plans` to
  `workshop/plans/000119-multi-agent-benchmark-harness-plan.md` (Core Concepts:
  ~10 pure entities + ~8 IO seams per ARCH-PURE; 5 milestone chunks, TDD).
- Fresh-eyes plan review (general-purpose subagent, verified every reuse API
  against source): **❌→✅**. Fixed one real bug — `ParseTask` extracted the
  *first* ```json fence over the whole body, so a spec containing its own json
  fence would corrupt the config round-trip; now scoped to the `## Config`
  section (+ regression test with a fenced spec). Folded advisories: leak-gate
  threshold (K=20, ≤60%), no-merge test push-target assertion, `Submission`
  struct fields, `ParseRunRecord` section-scoping, token-cost deferral decision.
- `change-code` gates: plan-quality **INFO**, estimate-quality **INFO** (both
  non-blocking); branch `000119-multi-agent-benchmark-harness` created in-place.
  Judges' actionable INFO folded into plan: extract a single `gitx.WorktreeRoot`
  helper (ARCH-DRY — don't re-derive the worktree-path convention); `freeze`
  warns on `## ` subheadings in a spec (SectionBody truncation guard). Estimate
  7.4h derived (`## Estimate`, reconciles 7.42). Implementing M1 next.
- **M1 done** — `cmd/sdlc/internal/bench/` package: `Task`/`Rubric` structs with
  config-scoped json round-trip (the reviewer's first-block bug fixed + guarded);
  `Store`; `benchmark-task` datatype (`construct/datatype/`, per-layer owned, no
  manifest edit); `workshop/benchmarks/README.md`; `sdlc bench freeze` (registered
  in main.go, helptext wired). Discoveries: issue files are located by `%06d`
  zero-padded glob (matched `branchcreate.go`); `issue.SectionBody` already stops
  at the first `## `, so the truncation risk is at *freeze* (source spec with an
  embedded `## `), not round-trip — warning rewritten to detect a non-canonical
  terminating heading. Smoke-froze #119 itself end-to-end. All bench + cmd/sdlc
  tests green; `go build ./...` clean.
