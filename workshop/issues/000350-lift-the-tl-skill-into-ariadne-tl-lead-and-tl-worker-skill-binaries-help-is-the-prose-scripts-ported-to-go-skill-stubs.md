---
id: 000350
status: working
deps: []
github_issue:
created: 2026-10-13
updated: 2026-10-13
estimate_hours: 4.83
card_mirror: '928c386e1933063f3d32198c4013409015eaff2c' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-13T17:19:46-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
---

# lift the TL skill into ariadne: tl-lead and tl-worker skill binaries (help is the prose), scripts ported to Go, skill stubs

## Problem

The TL skill lives as a draft in ops (`ops/skills-draft/xx-tl/SKILL.md`, ~340 lines, plus three Python scripts: `ss.py` 201 lines for guarded session-file writes, `tl-status.py` 132, `render-trains.py` 146). It was iterated through two projects (ariadne-robustness-1, couch-broadcast-web, 10-09..13) and has proven its shape, but: it isn't delivered by ariadne's skill system (a new TL agent must be told to read it, kink 78); its prose mixes the lead's job with the rules every worker needs, which the TL re-sends by hand in each dispatch and which drift (rules added mid-day never reached slots briefed earlier); and its scripts are Python beside Go tooling that already parses the same project files (`sdlc project`).

## Spec

The operator (2026-10-13):
1. **Split into two skills.** `tl-lead`: the hub's job (the eight-step playbook in 'Running a project with a TL', talking to the operator, dispatch judgment, plan review as product steward with the operator's two design rules, the smoke loop via `sdlc stage`, attention and load, project lifecycle). `tl-worker`: what every slot in a TL-run fleet must do (report to and ask the hub, never in-pane prompts; couch/sdlc calls as the whole shell call; no permission needed for tests, ask for windows/:0/machine changes/scope; real hashes from git rev-parse; a cited known failure never ends the suite early (make -k); retry a failed hub send; stage with `sdlc stage --build`; on SHIP land yourself, report, stop; never edit ops project files). Parts that apply to any agent in a Couch fleet move down into the base layer (AGENTS §15) and the couch skill instead of being duplicated.
2. **Skill-binary pattern** (like `sdlc --help`, `couch --skill`): `tl-lead --help` prints the lead prose, `tl-worker --help` the worker prose; the prose ships inside the binaries so it can't drift from the tools.
3. **Port the scripts to Go subcommands:** `ss.py`/`tl-status.py` → `tl-lead status …` (guarded session-file writes: submit, section, thread, owed, seen, with the operator-lock refusal); `render-trains.py` → `tl-lead trains`, built on `sdlc`'s project parser (no second parser). Python retired.
4. **Skill stubs** wired into ariadne's skill delivery (weave), like the couch pointer stub: a few lines saying 'run `tl-lead --help` / `tl-worker --help` and follow it'. `couch --prepare` (and the dispatch) loads `tl-worker` explicitly in worker slots; a hub loads `tl-lead`.
5. Compress while lifting: keep judgment, drop incident narratives (they stay in ops learnings); every repo reference in full `repo#NNN` form.
The ops draft stays the editing surface until this lands, then becomes a pointer.

## Done when

- `tl-lead --help` and `tl-worker --help` print the split prose; their stubs are delivered by weave; a worker slot prepared by Couch has `tl-worker` loaded; the TL's session-file and trains commands run from `tl-lead` with the Python removed; tests cover the ported subcommands (including the operator-lock refusal) and that the stubs point at the binaries.
- The ops draft is reduced to a pointer, and a TL dispatch shrinks to the issue plus TL-only context.

## Plan

Durable plan: `workshop/plans/000350-lift-the-tl-skill-into-ariadne-tl-lead-and-tl-worker-skill-binaries-help-is-the-prose-scripts-ported-to-go-skill-stubs-plan.md` (three review boundaries):

- [ ] **M1 — tl-worker.** Inventory every worker-facing rule in the ops draft and in TL dispatches (ops `skills-draft/xx-tl/SKILL.md`, "Running batches" dispatch checklist; recent dispatch messages in ops `teams/devinfra/projects/*/project.md` Logs). Sort each into: base layer (AGENTS §15, any Couch-fleet agent), couch skill (messaging mechanics), or `tl-worker` (TL-run fleet only). New `cmd/tl-worker`: `--help` prints the embedded prose (go:embed of a markdown file under the command, the same shape as `sdlc --help`/`couch --skill`). Skill stub under ariadne's skill delivery (weave) pointing at the binary. Couch-side load: coordinate with pair so `couch --prepare` tells a worker slot to load `tl-worker` (a pair issue, filed by the implementer, if Couch needs a change). Tests: the stub names the binary; `--help` contains each rule; AGENTS §15 and the couch skill aren't contradicted.
- [ ] **M2 — tl-lead prose.** New `cmd/tl-lead`; `--help` prints the lead prose rewritten from the ops draft as current rules and procedure: the eight-step playbook, talking to the operator, dispatch judgment, product stewardship with the operator's two design rules, the smoke loop via `sdlc stage`, attention and load, the project lifecycle. Every reference in full `repo#NNN` form.
- [ ] **M3 — Go subcommands; retire Python.** `tl-lead status submit|section|thread|owed|seen|trains` replaces `ss.py`/`tl-status.py` (same session-file format, the operator-lock refusal, the owed/fold bookkeeping); `tl-lead trains` replaces `render-trains.py` using `sdlc`'s project parser (shared package, no second parser). Golden tests against the current ops status file and project files; a lock-refusal test. Then the ops draft becomes a pointer (an ops change the TL makes), and the TL switches its own calls over.

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: skill-or-dispatcher     design=0.6 impl=0.14
item: skill-or-dispatcher     design=0.8 impl=0.16
item: cross-cutting-refactor  design=0.3 impl=0.14
item: greenfield-go-module    design=0.8 impl=0.24
item: smaller-go-module       design=0.3 impl=0.16
item: real-api-discovery      design=0.0 impl=0.18
item: atlas-docs              design=0.1 impl=0.05
item: milestone-review        design=0.0 impl=0.14
item: milestone-review        design=0.0 impl=0.14
item: milestone-review        design=0.0 impl=0.14
design-buffer: 0.15
total: 4.83
```

Items in order: M1 tl-worker (binary, prose, stub, §15 line); M2 tl-lead prose; the markdoc/trains extraction to pkg/; `tl-lead status` over the SessionSync seam with its stateful fake; `tl-lead trains` over the SDLC seam; the couch session-sync live conformance; atlas; three milestone reviews.

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only.*

## Log

### 2026-10-13

- Design (plan written): `tl-status.py` is dead (no caller in ops; the draft routes every write through `ss.py` over `couch --session-sync`, which owns the operator-lock refusal, exit 3). It is retired, not ported; `tl-lead status` passes exit 3 through as the lock refusal.
- ARCH-DRY: sdlc's trains parser (`cmd/sdlc/internal/project/trains.go`) and plan/section/fence scanners (`internal/issue`) are internal to cmd/sdlc; M3 moves them to `pkg/trains` and `pkg/markdoc` with forwarding aliases, so `tl-lead` and sdlc's gates share one parser.
- `sdlc issue show --json` consumers reject unknown keys, so no milestone field is added to the observation; car status comes from it, the bar from `pkg/markdoc` on the claimant checkout's file, off-train slots from `sdlc fleet inventory --json`.
- tl-worker load: one AGENTS §15 line (`tl-worker --help` per dispatched item; the hub runs `tl-lead --help`). `couch --prepare` already reloads the context onto the constitution, so no pair change is needed for the load.

- TL (ops:0) approved the plan at 043d866e, all six design calls as proposed (delivery dc0d1c8e, receipt verified).

## Revisions

- 2026-10-13 — the operator: **don't copy stories or lessons** into the binaries' prose. The prose states current
  rules and procedure only; incident history stays in ops learnings, and new kinks are addressed going forward as rule
  changes in the skill binaries (each fix lands as a prose or code change there), not as narratives.
