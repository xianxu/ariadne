---
id: 000121
status: working
deps: []
github_issue:
created: 2026-06-19
updated: 2026-06-20
estimate_hours:
started: 2026-06-20T10:43:05-07:00
---

# Writing-assistant skill — pair review workbench agent protocol

## Problem

The `xx-fix` skill (`construct/local/fix/`) is the **agent half** of the pair agentic
review workbench (pair **#000066**). M0–M3 built the pair-side document surface (the
record/apply/undo spine, projection/markers, the review window + toggle + poke + bar).
M4 needs the agent half: today the workbench pokes a bare `/xx-fix <path>` and the
dumb agent guesses (it ran `doc-review`/fact-check, edited nothing). The skill must
become a review-mode-aware, record-driven collaborative writing assistant.

This is the cross-repo counterpart of pair #000066 M4 — the contract both sides honor
is `pair/workshop/targets/review-protocol.md` (the agent↔nvim state machine). pair #66
M4 (the nvim consumer / seam / bar / menu) **depends on** this issue.

Note: `xx-fix` has outlived its name (it's no longer "fix small things from `🤖[]`" —
it's a collaborative editor). Rename to `writing-assistant` is in scope but can be the
last step; keep the `xx-fix` name working until pair's `REVIEW_TRIGGER` poke is swapped.

## Spec

Per `review-protocol.md` (the seam + invariants) and pair #66's M4 design:

- **Review-mode recognition.** When poked from the workbench ("please review …"),
  engage the **record-handoff flow** — propose `{old, occurrence, new, explain}` records
  to the handoff file (not xx-fix's file-write-in-place; the pane applies them undo-ably).
- **Agent owns all git.** Create the `review/<slug>` branch in the doc's repo; commit
  human + agent rounds via `docflow` — *after* the nvim's "applied" poke (apply can drop
  records). Unwinds the M1 scaffolding where the nvim shelled docflow. `ship` on "ship it".
- **Modes** (3 editing postures, mutually exclusive) + **fact-check** (orthogonal pass):
  **Generate** / **Copy Edit** / **Proofread** as `modes/<name>.md` (the `mode.lua` form),
  described up front; the active mode tracked as session state (the `review-<tag>.mode`
  seam, agent-written). **Fact-check** = free-text-triggered; dispatch read-only
  `doc-review` → integrate the note as edits via the record flow (fold doc-review in here).
- **Voice.** A doc's `voice: <slug>` frontmatter → load `~/.personal/<slug>-writing-style.md`
  for Generate + Copy Edit (Proofread + Fact-check are voice-neutral).
- **Memory discovery.** Reach into brain/pensives/repos for the review (the original win
  over parley's amnesiac one-shot).
- **Agnosticism.** The skill is the contract; pair/Claude is one *consumer* — keep it
  agent-agnostic where practical.

## Done when

- Poked from the workbench, the agent runs the record-handoff review flow (not doc-review),
  and the pane applies/commits rounds — a faithful process-level round-trip with pair #66.
- All git (branch / human+agent rounds / ship) is agent-side; the pair nvim writes none.
- Generate / Copy Edit / Proofread drive the agent; mode is switchable + reflected in the
  seam; fact-check dispatches doc-review and integrates via records.
- `voice: <slug>` is honored by the voice-relevant modes.
- (rename) `xx-fix` → `writing-assistant`, with pair's `REVIEW_TRIGGER` swapped in lockstep.

## Plan

Mirrors pair #66 M4's sub-milestones (the agent-side rows):

- [ ] M4a — review-mode recognition + record-handoff flow + agent-owns-git (branch/rounds);
  memory discovery. (The spine; pairs with pair #66 M4a.)
- [ ] M4b — modes: `modes/{generate,copy-edit,proofread}.md` + `mode.directives` wired;
  write the active mode to the `review-<tag>.mode` seam.
- [ ] M4c — voice (`voice:` frontmatter loading) + fact-check pass (doc-review dispatch +
  integrate).
- [ ] M4d — ship ("ship it" → `docflow ship`) + the faithful e2e demo; rename to
  `writing-assistant`.

## Log

### 2026-06-19

- Filed as the agent half of pair #000066 M4 (operator chose split (A): own ariadne issue
  + pair #66 deps on it). Contract = `pair/workshop/targets/review-protocol.md`.
