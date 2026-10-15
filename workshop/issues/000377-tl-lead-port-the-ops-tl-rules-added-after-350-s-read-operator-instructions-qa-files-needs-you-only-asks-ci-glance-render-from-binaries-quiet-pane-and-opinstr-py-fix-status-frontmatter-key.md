---
id: 000377
status: working
deps: []
github_issue:
created: 2026-10-14
updated: 2026-10-14
estimate_hours:
card_mirror: '61dfaff9de95f84900765daee472045008c3b800' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-14T20:39:55-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
---

# tl-lead: port the ops TL rules added after #350's read (Operator instructions + QA files, Needs-you only asks, CI glance, render-from-binaries, quiet pane) and opinstr.py; fix status frontmatter key

## Problem

ariadne#350 (landed 10-14, PR #243) lifted the ops TL skill draft (`~/workspace/ops/skills-draft/xx-tl/SKILL.md`) into `tl-lead` / `tl-worker`, reading the draft mid-day. The operator and TL kept changing it afterwards, and a read-only comparison (10-14 afternoon) found rules missing from `cmd/tl-lead/help.md`, so the ops pointer and the retirement of the ops Python can't happen yet:
- **`## Operator instructions` and QA files:** the operator-owned section 0; the TL only appends `→ [answer](qa/…)` to every bullet, question or instruction, and never edits, deletes or marks the operator's lines; answers in `teams/<team>/qa/<operator>/YYYY-MM-DD-NN-<slug>.md`, gitignored; `process-instruction: "[free-form] answer the operator's 🤖 comments in place"` frontmatter; a plain relative backlink, no `#anchor` (parley.nvim#323); an answer file answers only the question asked.
- **Needs you holds only asks:** no status; remove an item the moment it's done.
- **Glance at CI on main every pass** (`gh run list --branch main`); a red main gets an issue and a slot that pass.
- **Render sections from binaries wherever possible:** Quota from `sdlc usage` (#365); facts from `sdlc`; "a fact retyped every pass is the next piece to lift".
- **Quiet agent pane:** status-file and QA input is answered in files (at most one line in the pane); routine worker reports go to the status file and digest.
- **A lock with no submit:** mention in chat that updates are waiting.
- **The multi-line bullet parser** `ops/skills-draft/xx-tl/scripts/opinstr.py` (`list`, `link`; bullets by position, since the operator reuses numbers) has no Go port.
- Contradictions: the status frontmatter key (ops uses `owner:`, which Couch reads; B says `couch-owner:`); the section order (Operator instructions first).

## Spec

Port each missing rule into `tl-lead --help` (rules only, no stories); port `opinstr.py` as `tl-lead status instructions list|link FILE …` with its golden tests (multi-line bullets, blank lines inside, repeated numbers, image lines); fix the frontmatter key and the section order. Then the TL replaces the ops draft with a pointer to `tl-lead --help` and deletes `ops/skills-draft/xx-tl/scripts/*.py`.

## Done when

- `tl-lead --help` contains every rule above; `tl-lead status instructions` passes the ported parser tests; a fresh comparison of the ops draft against tl-lead finds no missing rule.

## Plan

Durable plan: `workshop/plans/000377-tl-lead-ops-rules-catch-up-plan.md` (single pass, one close).

- [ ] help.md: rules G1–G8, the frontmatter key, the instructions verbs
- [ ] pure parser and linker (`instructions.go`) with golden tests
- [ ] `status instructions list|link` verbs with fake-backed tests
- [ ] help names every verb (test)
- [ ] atlas tl-skills.md
- [ ] G10/G11 (slot verification after disruption, filing from a resting checkout) once the hub sends the source
- [ ] fresh comparison of the draft against help.md, recorded in the Log

## Log

### 2026-10-14
- Compared ops draft e7ea8e5 with help.md at main 18915f6c: gaps G1–G9 in the plan. `couch --skill` names `couch-owner:` canonical (`owner:` legacy), so the help keeps it. parley.nvim#315 landed (PR #228): the `holder: agent` workaround is not ported.
