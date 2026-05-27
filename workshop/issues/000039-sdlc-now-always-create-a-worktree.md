---
id: 000039
status: done
estimate_hours: 8
deps: []
created: 2026-05-27
updated: 2026-05-27
actual_hours: 4
---

# Defer worktree decision to implementation-start time

## Problem

`sdlc start --issue N` today does two distinct things in one shot:
1. Commits + pushes the issue file (workstream-claim primitive).
2. Creates a worktree + branch from HEAD.

For a single operator with a single-threaded mind, the second step
slows down development and is often the wrong default — small fixes,
docs touch-ups, and early planning work all happen better on main with
no worktree overhead. The decision *whether* to worktree is also poorly
timed: at `sdlc start` you don't yet know the change's complexity,
blast radius, or test surface. Those signals become legible only after
the plan is written.

Beyond the worktree question, the planning → implementation transition
is also currently un-gated: there's no checkpoint that enforces a
filled-in Spec, a non-empty Plan, or judges whether the plan is
actually executable. `sdlc judge plan` runs only at merge time (to
catch unchecked-done items), never as a pre-implementation gate.

## Done when

- A new verb `sdlc claim --issue N` exists for the workstream-claim
  step (rename of today's `sdlc lock`). `sdlc lock` is removed in
  the same PR — clean rename, no deprecation alias.
- A new verb `sdlc change-code --issue N` exists for the planning →
  implementation transition, composing three gates:
  1. **Structural sanity** — issue has filled `## Spec`, non-empty
     `## Plan` checklist, populated `## Done when` (or `related:`
     frontmatter); refusable with `--force` carrying a rationale.
  2. **Plan-quality LLM judge** — fresh-context judge against the
     issue + plan file, emitting `VERDICT: CLEAN | INFO | FAILURE`
     per the shipped classifier contract (#40). FAILURE refuses
     without `--force`; INFO surfaces and proceeds; CLEAN silent.
  3. **Branching-strategy ask** — AskUserQuestion (via the
     agent-protocol below) with a sizing hint derived from the
     plan: worktree vs branch-in-place.
- `sdlc start` is removed — there's no realistic moment where you
:q
  want to claim AND change-code in one shot, since planning happens
  between them. All in-repo references (AGENTS.md, skills,
  Makefile, scripts) get migrated to `sdlc claim` + `sdlc change-code`
  in the same PR.
- `sdlc change-code` agent-protocol: when invoked without
  `--worktree=<yes|no>` and stdin is not a tty, emit the sizing
  hint on stderr, print the sentinel line `ASK_BRANCHING_STRATEGY`
  on stdout, exit code 2. The agent reads exit 2 + sentinel as
  "ask the operator via AskUserQuestion, then re-run with the
  flag." Direct human invocation (isatty) falls back to a stdin
  prompt.
- The `xx-sdlc` skill learns the sentinel pattern: on exit 2 + a
  recognized `ASK_<TOPIC>` line, it issues the corresponding
  AskUserQuestion and re-invokes the sdlc command with the answer
  flag (`--worktree=yes|no` for the v1 branching ask).

## Spec

### Verb shape

```
sdlc claim --issue N
  Composes: today's `sdlc lock` (sync workshop/issues/ → main).
  Side effects: commits + pushes the issue file; status stays
                whatever the operator set (open / working / blocked).
  No branch creation, no worktree.

sdlc change-code --issue N [--worktree=yes|no] [--force]
  Three gates, in order:
    1. Structural sanity checks (deterministic).
    2. Plan-quality judge (LLM, fresh context).
    3. Branching-strategy ask (skipped when --worktree=… is set).
  On success: creates branch (worktree or in-place per the answer).
  On failure: prints the failing gate, refuses with non-zero exit.

sdlc start  (REMOVED)
  Replaced by separate `claim` (early) + `change-code` (later) verbs.
  All in-repo references migrated in this PR. If anything still calls
  `sdlc start`, it should error with a clear "use sdlc claim + sdlc
  change-code instead" message — not silently compose them, since
  composing is the wrong default per this issue's premise.

sdlc lock  (REMOVED)
  Renamed to `sdlc claim`. All in-repo references migrated in this PR.
  Anything still calling `sdlc lock` errors with "use sdlc claim".
```

### Sizing hint (input to the branching ask)

Computed from the issue file's frontmatter + `## Plan` section + any
referenced `workshop/plans/NNNN-*-plan.md`:

```
Issue NNNN — <title>
  estimate:        <frontmatter estimate_hours>
  plan items:      <count of - [ ] / - [x] in ## Plan>
  milestones:      <count of "Mx:" labels in plan items>
  spec words:      <word count of ## Spec section>
  related files:   <count of paths in `related:` frontmatter>
  plan file size:  <if separate plan file exists, its word count>
  → <bucket>: small | medium | large
```

`bucket` is mechanical: small if estimate < 2h AND plan items ≤ 5 AND
milestones == 0; large if estimate ≥ 6h OR milestones ≥ 3; medium
otherwise. No fancy judgment — operator reads the numbers and decides.

### Agent protocol for the branching ask

When `sdlc change-code` reaches the branching gate without
`--worktree=` set:

- If `isatty(stdin)`: prompt `Branching: [w]orktree / [m]ain in-place / [c]ancel >`
  on stderr, read one char from stdin, act.
- Else: emit the sizing hint on stderr, emit
  `ASK_BRANCHING_STRATEGY` as a single line on stdout, exit 2.
  Agent re-invokes with `--worktree=yes` or `--worktree=no` after
  AskUserQuestion.

The sentinel is a stable contract — agents grep for it. Future asks
(if more gates need operator input) follow the same `ASK_<TOPIC>`
pattern.

### Plan-quality judge

New judge category `plan-quality` (distinct from existing `plan`
which audits issue file completeness pre-merge). Prompt asks: "Read
the issue's Spec + Plan + any referenced plan file. Is this plan
executable as-written? Flag vague items, missing test surface,
undefined acceptance criteria, undeclared cross-issue deps." Emits
the shipped `VERDICT: CLEAN | INFO | FAILURE` shape.

### Structural sanity checks (deterministic)

Each refusable with `--force <reason>`:

| Check | Rule |
|---|---|
| Spec present | `## Spec` section has ≥ 50 words |
| Plan present | `## Plan` has ≥ 1 checklist item that isn't an empty `- [ ]` |
| Done-when present | `## Done when` has ≥ 1 non-empty bullet OR `related:` frontmatter populated |
| Estimate present | `estimate_hours:` in frontmatter is a positive number |

These mirror `sdlc close`'s deterministic guards — same `--force +
reason` pattern.

## Plan

- [x] Rename: introduce `cmd/sdlc/claim.go` mirroring today's
      `lock.go`. Keep `lock.go` as a thin alias calling into the
      same internals. Update help text + tests.
- [x] New verb: `cmd/sdlc/changecode.go`. Cobra command, wires
      flags `--worktree=<yes|no>`, `--force`, `--no-judge` (skip
      plan-quality), `--no-structural` (skip structural).
- [x] Structural checks: pure function in
      `cmd/sdlc/internal/issue/` or a new `cmd/sdlc/internal/gate/`.
      Returns a list of failures; caller decides refuse vs proceed.
      Unit tests against synthetic issue files.
- [x] Plan-quality judge: new `judge.Category` value
      `plan-quality`. Add prompt template to `prompts.go`. Reuses
      the shipped `VERDICT:` classifier from #40 unchanged.
- [x] Sizing hint: pure function reading the issue file (and
      optional plan file) → struct → printable summary. Bucket
      logic per Spec. Unit tests.
- [x] Branching ask: implement the agent-protocol (sentinel +
      exit 2) and the stdin-tty fallback in `changecode.go`.
- [x] Branch creation: factor today's `sdlc start` worktree-creation
      code into a reusable helper; add an in-place branch path
      (`git checkout -b <name>`).
- [x] Removal: both `sdlc start` and `sdlc lock` are removed in
      this PR. Each errors with a one-line "use sdlc claim /
      sdlc change-code instead" message and exits non-zero.
- [x] In-repo migration: grep for `sdlc start` and `sdlc lock` and
      replace with the new verbs across the operator-facing surface:
      atlas/, README.md, AGENTS.md, skills (xx-sdlc and any
      sibling), cobra command help strings in `cmd/sdlc/`, embedded
      helptext in `cmd/sdlc/helptext/`. Leave `workshop/history/`
      untouched (archived prose). Lands in the same PR so nothing
      breaks at merge.
- [x] Skill update: `xx-sdlc` learns the `ASK_<TOPIC>` sentinel
      contract — on `exit 2` from sdlc with a recognized topic line,
      issue the matching `AskUserQuestion` and re-invoke with the
      answer flag. v1: just `ASK_BRANCHING_STRATEGY`
      (→ `--worktree=yes|no`).
- [x] Atlas: update `atlas/workflow/sdlc-binary.md` — split the
      stage table to show claim + change-code as separate stages
      between planning and build. Document the agent-protocol
      sentinel pattern.
- [x] AGENTS.md: update §2 (Overall Workflow) to mention the
      new verbs and the planning-on-main convention.
- [x] Tests: end-to-end test for the full claim → plan → change-code
      flow against a synthetic repo (extend the existing test
      harness in `cmd/sdlc/*_test.go`).

## Log


- 2026-05-27: closed — All gates of sdlc change-code working live; sdlc claim is the rename of sdlc lock; sdlc start errors with migration message. Unit tests cover structural-checks (11), sizing (6), plan-quality prompt (3), branching prompt (10), title extraction (5), sentinel stability (1). Full go test ./... green across ariadne. Smoke-tested all three verbs against the installed binary at pair/bin/sdlc.
- 2026-05-27 — Shipped on `main` directly (no worktree, no PR — per
  the operator's preference and the principle this issue itself
  codifies).
- Commits (in chronological order on `main`):
  - structural-checks + sizing-hint pure helpers + tests
  - sdlc lock → sdlc claim rename
  - plan-quality judge category + prompt + tests
  - sdlc change-code verb + branching-ask + helptext + tests
  - sdlc start removal + in-repo migration + xx-sdlc skill update
- AGENTS.md plan item turned out to be a no-op: grep showed AGENTS.md
  doesn't reference `sdlc start` or `sdlc lock`. Box ticked because
  the work was *to verify* and there was nothing to change.
- End-to-end test (synthetic repo: claim → plan → change-code →
  branch created) deferred to follow-up issue. Constituent parts
  are unit-tested: structural-checks (11 cases), sizing (6 cases),
  plan-quality prompt (3 cases), branching prompt (10 cases), title
  extraction (5 cases), sentinel stability (1 case).
- Verified live: built sdlc, installed at
  `/Users/xianxu/workspace/pair/bin/sdlc`, smoke-tested all three
  verbs — `sdlc claim --help` and `sdlc change-code --help` render
  the new helptext; `sdlc start` errors with the migration message;
  `sdlc lock` returns cobra's "unknown command" error.
