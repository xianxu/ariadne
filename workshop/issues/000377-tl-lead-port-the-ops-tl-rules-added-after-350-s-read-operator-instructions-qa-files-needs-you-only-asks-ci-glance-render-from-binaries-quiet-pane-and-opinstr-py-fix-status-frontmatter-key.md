---
id: 000377
status: working
deps: []
github_issue:
created: 2026-10-14
updated: 2026-10-14
estimate_hours: 1.259
card_mirror: 'd030951424b0a69bc961e763d46dfe42cdd56cbc' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-14T20:39:55-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
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

Durable plan: `workshop/plans/000377-tl-lead-port-the-ops-tl-rules-added-after-350-s-read-operator-instructions-qa-files-needs-you-only-asks-ci-glance-render-from-binaries-quiet-pane-and-opinstr-py-fix-status-frontmatter-key-plan.md` (single pass, one close).

- [x] help.md: rules G1–G8, the frontmatter key, the instructions verbs
- [x] pure parser and linker (`instructions.go`) with golden tests
- [x] `status instructions list|link` verbs with fake-backed tests
- [x] help names every verb (test)
- [x] atlas tl-skills.md
- [x] G10 (verify slots after a disruption) and G11 (issue filing), text from the TL (plan Revisions)
- [x] fresh comparison of the draft against help.md, recorded in the Log

## Estimate

Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against
`baseline-v3.1.md`. Method A only. Source flagged stale; treat values as provisional.
Existing Go stack (cobra, the SessionSync seam and its stateful fake, the help
guards), so familiarity 1.0; no library discount applies. The approved plan is
thorough, so design hours are small and the design buffer is 15%. Impl values are
40% of the v2 ranges.

Items in order: help.md rule port (G1–G8, G10, G11); the pure parser and linker;
the list/link verbs with the single-shot write; help guards derived from the tree;
atlas; the fresh comparison; one close review. Design 0.26h + impl 0.96h; buffered
total 1.259h.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: smaller-go-module design=0.06 impl=0.20
item: smaller-go-module design=0.04 impl=0.16
item: smaller-go-module design=0.04 impl=0.16
item: smaller-go-module design=0.02 impl=0.08
item: atlas-docs design=0.04 impl=0.08
item: smaller-go-module design=0.02 impl=0.08
item: milestone-review design=0.04 impl=0.20
design-buffer: 0.15
total: 1.259
```

## Log

### 2026-10-14
- Compared ops draft e7ea8e5 with help.md at main 18915f6c: gaps G1–G9 in the plan. `couch --skill` names `couch-owner:` canonical (`owner:` legacy), so the help keeps it. parley.nvim#315 landed (PR #228): the `holder: agent` workaround is not ported.
- The TL approved the plan (G1–G9, the hash guard, `couch-owner:`), pinned `--flow full`, and supplied the text for G10 and G11 (plan Revisions).
- plan-quality INFO, 2 Minors (help rule guards; the single-shot link write), folded into the plan's Revisions. The plan file was renamed to its canonical name.
- Built in 677b2a33. The help guards cover one phrase per rule (G1–G8, G10, G11); the verb test walks the command tree. Parity: on a copy of ops's real status file with its answer links stripped, `instructions list` output matches `opinstr.py list` byte for byte (3 bullets). `go test ./cmd/tl-lead ./cmd/tl-worker` green.
- Fresh comparison (a fresh-context subagent, the draft at e7ea8e5 against tl-lead/tl-worker help plus the cited `couch --skill`, `sdlc --help` and AGENTS §5/§15): about 140 rules covered, no load-bearing rule missing. It found 5 minor omissions, now folded in with guards: in a dispatch the repo prefix is enough; update the status file on every thread change; don't ask Couch to encode one repo's usage; stage with `sdlc stage`, never `sdlc move`; waves are fine, written as the trains table. One contradiction is deliberate: known-red CI. The draft lets it land; tl-lead (since #350) and AGENTS §5 (#309) say `sdlc merge` refuses red CI, fix it in its own issue. It was confirmed in practice on #350 (no CI waiver). Also dropped as obsolete: the until-ariadne#308 stray-tick trap, the ariadne#318 sandbox bypass, `tools/watch_issues.py`, 'by hand until sdlc learns directories'.
