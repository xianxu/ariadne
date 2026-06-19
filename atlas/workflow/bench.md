# Multi-agent benchmark harness (`sdlc bench`)

A controlled-experiment harness (#119) for measuring how well different coding
agents (claude, codex, …) do **real ariadne work**. It holds everything constant
— task spec, starting commit, environment, rubric, judges — and varies one thing:
which agent. Two scores per agent, side by side: **output quality** and
**workflow-fit** (how cleanly it rides the SDLC).

**The harness is base-layer (`cmd/sdlc`); the task library + results are
repo-local (`workshop/benchmarks/`).** So every ariadne-styled repo gets
`sdlc bench` and benchmarks agents on *its own* backlog.

## Pipeline

`freeze` (snapshot a live issue → immutable task) → `run` (fan out N agents into
isolated worktrees, per mode, never merging) → `grade` (Stage A measure + Stage B
blind head-to-head judge) → `review` (fold in operator rankings, de-anonymize) →
`report`/`leaderboard` (accumulate).

## Surface

- **Package** `cmd/sdlc/internal/bench/` — a pure core (structs + deterministic
  functions, unit-tested with no IO mocks) wrapped by a thin IO shell
  (`Store`/`Worktreer`/`Runner`/`Measurer`). Per ARCH-PURE, the pure files never
  touch os/exec/time/git; those are injected.
- **CLI** `cmd/sdlc/bench.go` — the `sdlc bench` verb family (helptext in
  `cmd/sdlc/helptext/bench*.md`).
- **Datatype** `benchmark-task` (`construct/datatype/benchmark-task.md`) — the
  immutable frozen task: spec snapshot + `base_sha` + rubric. Agents branch from
  `base_sha` and never merge, so a task replays identically even after main moves.
- **Artifacts** `workshop/benchmarks/{tasks,runs}/`, `leaderboard.md` (see its
  README).

## Key decisions (see #119 spec)

- **Grading is hybrid:** objective dimensions (build/tests/artifacts/metrics) are
  MEASURED agent-neutrally; subjective dimensions (elegance, design reasoning,
  doc quality, gate judgment) are judged HEAD-TO-HEAD **blind** (anonymized
  submissions, randomized order) by a pinned LLM judge that is *not* a contestant,
  plus the operator.
- **Run mode is a first-class dimension** carried by a `Responder` seam:
  `autonomous` (wired day-one), `interactive` (pinned responder — seam only),
  `live` (operator pairing — exhibition, leaderboard-excluded).
- **Effort is wall-clock/turns, not `sdlc actual`** — active-time is
  Claude-transcript-only, so it can't compare across agents.
- **Reuse:** `judge.Dispatch` (already runs claude/codex/gemini symmetrically),
  `gitx`, `issue` frontmatter/section helpers. The `VERDICT` classifier is *not*
  reused — it's a binary gate; Stage B needs a structured-ranking parser.

## State

- **M1 (done):** data model (`Task`/`Rubric` + config-scoped json round-trip),
  `Store`, the `benchmark-task` datatype + `workshop/benchmarks/` layout, and
  `sdlc bench freeze`.
- **M2–M5 (planned):** runner + responder seam; grader Stage A (objective) /
  Stage B (blind judge + operator review); leaderboard + end-to-end demo. See
  `workshop/plans/000119-multi-agent-benchmark-harness-plan.md`.
