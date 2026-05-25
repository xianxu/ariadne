---
type: pensive
date: 2026-05-25
topic: sdlc-checkpoint-binary
mode: thesis
description: Markdown-index progressive disclosure leaks under `**/*.md` globbing; binary-backed skills (gmail/charon pattern) keep content inside the executable and force explicit `<bin> --help` requests. The same binary surface lifts naturally to workflow-checkpoint guards: SKILL.md describes states, subcommands enforce transitions. The right unit is the checkpoint, not the FSM — we don't model the journey, we defend known commit moments against drift. An `sdlc` binary collects ariadne's existing Makefile targets (close-issue, fetch, worktree, push, pull-request, merge) under one verb namespace with embedded help, structured state inspection, and LLM-judge subcommands that run in fresh, non-colluding context for essence checks.
references: [/Users/xianxu/workspace/brain/memory/feedback_deterministic_shell.md, /Users/xianxu/workspace/brain/memory/feedback_consistency_prosthesis.md, /Users/xianxu/workspace/brain/memory/feedback_under_specify.md, /Users/xianxu/workspace/ariadne/scripts/close-issue.py, /Users/xianxu/workspace/ariadne/Makefile.workflow, /Users/xianxu/workspace/ariadne/atlas/workflow/issue-lifecycle.md]
---

# Pensive: skill binaries and the SDLC checkpoint binary

## The progressive-disclosure leak

Skill systems use an index-markdown-with-references pattern to manage context budget: a summary file points at deeper sub-files, the agent loads what it needs. The intent is lazy disclosure — pay for content only when context demands it.

The leak is the agent itself. An agent that decides to "just be thorough" runs `**/*.md` and pulls the whole tree into context, defeating the disclosure boundary. The boundary is conventional, not enforced.

## Binary-backed skills: enforced disclosure

The `gmail` / `charon` pattern moves skill content from the filesystem into a binary. The SKILL.md becomes a thin pointer: "to use this skill, run `<bin> --help`." The actual guidance lives inside the binary, embedded into the `--help` surface. Deeper guidance is reachable only via `<bin> <subcommand> --help`. Convention over invention — agents try `--help` reflexively; we don't need a custom verb to do what the standard flag already does.

What changes:

- Globbing markdown can't reach the content — it isn't on disk.
- The skill content and the behavior ship together — no drift between what the docs say and what the binary does.
- Disclosure becomes capability-based: each layer requires an explicit request.

The cost is discoverability and grepability. Agents must know the binary exists; humans can't `rg` the prose offline. Mitigation: keep the index markdown as a thin, greppable signpost that names the binary and lists its top-level verbs.

Right pattern when: (a) the skill is large enough that lazy loading matters, (b) the skill is tied to a binary that has to exist anyway, (c) instructions and behavior change in lockstep. Pure-prose skills (style guides, conventions) stay markdown.

## From workflow FSMs to checkpoint guards

A tempting next step is to formalize workflow skills as finite state machines: SKILL.md describes states, the binary enforces transitions via `bin from:to` subcommands.

The discipline that survives the bait: **we are not modeling workflows, we are codifying checkpoint guards.** `close-issue` doesn't pretend to model the SDLC. It defends one transition — claiming "done" — against omission. It refuses without verification, actual hours, and an atlas touch. Everything between checkpoints stays in prose where it belongs.

This matters because most workflow knowledge is tacit and emergent. Sitting down to design the FSM ossifies what hasn't settled yet. Checkpoint guards are incremental: add a guard when the same drift recurs at a specific transition. The trigger is empirical (drift), not architectural ("we should formalize this step").

Heuristic: when the same drift gets caught at review twice, promote it from a `lessons.md` entry to a subcommand check. First time = prose. Second time = code.

## The `sdlc` binary

Ariadne already has the proto-pattern as Makefile targets:

| Target | Defends what |
|---|---|
| `make fetch <N>` | Issue file shape, next-ID assignment, frontmatter |
| `make worktree [name]` | Single-untracked-issue auto-detect, pre-commit before branch |
| `make pull-request` | Branch ≠ main, links touched issues into PR body |
| `make merge` | Upstream synced, ahead=0, undone-issue scan, irreversible-action confirm |
| `make push` | Clean tree, pre-merge checks pass, archive done issues |
| `make close-issue ISSUE=N ACTUAL=h VERIFIED='...'` | Evidence required, milestone plan ticked, atlas touched in window |
| `make issue-sync` | Issue file landed on main as parallelization lock |
| `make check-*` | LLM-judge: DRY, PURE, plan, specs, lessons |

The Make-target surface is a primitive command dispatcher with weak arg parsing, env-export shims, and no way to ship embedded help. A real binary lifts this with:

- Typed subcommand args (`sdlc close-issue --issue 15 --actual 2.5 --verified '...'`)
- Rich, embedded `--help` per subcommand (prose + flags + examples, not just usage)
- Progressive disclosure: `sdlc --help` → top-level shape + verb list, `sdlc close --help` → checkpoint depth
- Structured state: `sdlc state` reports current branch, open worktrees, working issues, where in the lifecycle the agent sits

The last one matters most for compaction recovery. "Where am I in this workflow" becomes a command instead of re-inference from the issue file.

## Form and essence: two guard layers

Checkpoint guards defend against **omission** (claiming done without doing). They don't defend against **theater** (doing the form without substance). `close-issue` forces `VERIFIED='...'`; it can't tell whether the verification was real or ritual.

The second layer is LLM-as-judge with **anti-collusion via fresh context.** The doer's context is contaminated — it has incentive (and just attention bias) to declare success. A judge dispatched as a clean subagent has none. The pre-merge `check-*` family is already the structural anti-collusion move, just not named that way. The axis isn't model identity (same model is fine); it's context contamination.

The layering matters: form first because it's cheap and deterministic (no judgment, just "did you produce X"), essence second because judge invocations are expensive and probabilistic. Form filters easy-to-catch failures so the judge spends its attention on substantive ones.

Practical surface: `sdlc judge plan|specs|lessons|milestone-review` — each dispatched fresh, returns structured findings. The doer never invokes the judge in-session; the harness does.

## What this is not

- **Not a full SDLC model.** No state graph, no enumerated transitions, no `from:to` syntax. Adding subcommands is incremental, driven by observed drift.
- **Not a replacement for prose skills.** `xx-issues` describes the *why* of the issue file shape; `sdlc fetch` enforces the *invariant* that newly-created files match it. Both keep their role.
- **Not autonomous.** Some checkpoints (`make merge`) prompt for irreversibility confirmation. The binary can expose `--yes` for autonomous runs, but the default posture is: agent invokes, human approves at the irreversible step.

## Open questions

- How fragmented should the binary be? One `sdlc` binary with many subcommands keeps the namespace coherent but risks growing into a god-binary. Split candidates: `sdlc` (SDLC primitives), `construct` (substrate management — adapt/promote/upgrade — which already lives as a slash-command surface), `velocity` (estimate/calibrate, currently `active-time-v3.py`). The `<verb> --help` discoverability works either way; the question is which boundary minimizes cross-cutting flag passing.
- What's the right state surface for `sdlc state`? Terse human-readable ("on branch X, issue 42 working, 1 worktree") or structured JSON for tool composition? Probably both — `--json` for tooling, default human-readable.
- How do checkpoint guards interact with the autonomy gradient? Different policy per checkpoint: `close-issue` autonomous-friendly (no irreversibility), `merge` agent-confirmed (irreversible cleanup), `push` somewhere between.
- Anti-collusion strength: same model in fresh context is the cheap baseline. Is there marginal value in cross-model judging (Sonnet judges Opus's work)? Cost is calibration drift; gain is independence of failure modes.
- What stops the `sdlc` binary from itself becoming a target of context-pollution? The judge subcommands invoke fresh agents — but `sdlc` is invoked from within the doer's session. If the doer is dishonest with `VERIFIED='...'`, the binary cooperates. Mitigation lives upstream (judge sees the same artifacts the doer claims to have produced), not inside the form-layer guard.
