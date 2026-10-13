---
id: 000350
status: open
deps: []
github_issue:
created: 2026-10-13
updated: 2026-10-13
estimate_hours:
card_mirror: 'a88c98687f32f300004793be7d7b3d67f4281e61' # card fields mirrored from issue-cards; edit via sdlc
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

- [ ]

## Log

### 2026-10-13
