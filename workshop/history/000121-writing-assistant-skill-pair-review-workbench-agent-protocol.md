---
id: 000121
status: done
deps: []
github_issue:
created: 2026-06-19
updated: 2026-06-22
estimate_hours: 3.51
started: 2026-06-20T10:43:05-07:00
actual_hours: 2.80
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
it's a collaborative editor). Rename is deferred to a follow-up; the likely
user-facing name is `review`. Keep the `xx-fix` name working until the rename lands
in lockstep with any downstream trigger changes.

## Spec

Per `review-protocol.md` (the seam + invariants) and pair #66's M4 design:

- **Review-mode recognition.** When poked from the workbench ("please review …"),
  engage the **record-handoff flow** — propose `{old, occurrence, new, explain}` records
  to the handoff file (not xx-fix's file-write-in-place; the pane applies them undo-ably).
- **Agent owns all git.** Create the `review/<slug>` branch in the doc's repo; commit
  human + agent rounds via `docflow` — *after* the nvim's "applied" poke (apply can drop
  records). Unwinds the M1 scaffolding where the nvim shelled docflow. `ship` on "ship it".
- **Modes** (3 editing postures, mutually exclusive) + **fact-check** (orthogonal pass):
  **Generate** / **Edit** / **Proofread**, described directly in `SKILL.md` so pair
  only needs UI metadata; the active mode is tracked in pair's `review-<tag>.mode`
  seam. **Fact-check** = free-text-triggered; dispatch read-only `doc-review` →
  integrate the note as edits via the record flow.
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
- Generate / Edit / Proofread drive the agent; mode is switchable + reflected in the
  seam; fact-check dispatches doc-review and integrates via records.
- `voice: <slug>` is honored by the voice-relevant modes.
- Rename is deliberately deferred; `xx-fix` remains the deployed skill name for this
  shipment, with `review` the likely follow-up name.

## Estimate

```estimate
model: estimate-logic-v2
familiarity: 1.0
item: skill-or-dispatcher    design=0.4 impl=1.0
item: skill-or-dispatcher    design=0.3 impl=1.0
item: milestone-review       design=0.0 impl=0.6
design-buffer: 0.30
total: 3.51
```

`skill-or-dispatcher` ×2 — the core review-workbench protocol (M4a) + the rest of the
SKILL (M4b–d: modes / voice / fact-check / ship / rename); `milestone-review` for the
boundary reviews. Mostly prose (SKILL.md), so impl-weighted; the live smoke (not a
headless suite) is the real verification, with `fake-agent-v2` as the deterministic
reference. familiarity 1.0 (the protocol is already specced in `review-protocol.md`).

## Plan

Mirrors pair #66 M4's **structure-first re-slice** (2026-06-21) — the agent-side rows:

- [x] M4a — review-mode recognition + record-handoff flow + agent-owns-git (branch/rounds);
  memory discovery. *(Done; the spine. Pairs with pair #66 M4a. Close pends the live smoke.)*
- [x] M4a' — the prep: on the propose poke, run `pair-review-readiness <file>`, act per the
  case (stop/track/resume/new/interact), mark the target `ready`; on resume → reestablish
  context from the round commits. *(Pairs with pair #66 M4a'.)*
- [x] M4b — **skeleton** (agent side of the structure slice): the 🤖[] fulfill/punt +
  accept/reject (parley §5) handling + a **default posture** + **ship** ("ship it" →
  `docflow ship`).
- [x] M4c/M4d — **thicken** (tuning): Generate/Edit/Proofread mode prose in
  `construct/local/fix/SKILL.md`; `review-<tag>.mode` seam owned by pair; voice
  (`voice:` frontmatter → `~/.personal/<slug>-writing-style.md`); fact-check pass
  (`doc-review` → records); copy-edit minimal marker contract; direct agents away from
  stale fake harness guidance. Rename deferred to follow-up naming work.

## Log

### 2026-06-22
- 2026-06-22: closed — go test ./...; go test ./cmd/doc-review ./cmd/sdlc ./cmd/sdlc/internal/issue; git diff --check; live pair review workbench used to revise and publish the binary-skill blog post; hosted xx-fix protocol covered review prep, record handoff, agent-owned git/ship, Generate/Edit/Proofread, doc-review fact-check guidance, and minimal Edit markers; rename intentionally deferred; --no-verdict because M4a/M4b/M4c were historical sub-boundaries folded into final live-acceptance close; --no-judge because the retrospective and final Claude judge dispatches hung without returning; review verdict: not-run

- Final acceptance: operator used the pair review workbench and this hosted `xx-fix`
  protocol for a real revision of the binary-skill blog post, then posted the article.
  The skill now covers the hosted pair protocol, local review prep guidance, agent-owned
  git/ship, Generate/Edit/Proofread mode semantics, fact-check via `doc-review`, voice
  guidance, and minimal marker proposals for Edit. Rename is intentionally not part of
  this shipment; `review` is the likely successor name to evaluate separately.
- Folded fact-check review into `xx-fix` and removed standalone fresh-context-review
  guidance from the hosted path; agents should treat fake review-agent scripts as
  harness fixtures, not runtime behavior.
- Tightened the hosted Copy Edit contract after pair #66 M4d smoke feedback: one-round
  menu instructions are not sticky, and copy-edit proposals should use minimal inline
  marker records (`new = "🤖<old>{new}"` / `new = "🤖{new}"`) rather than paragraph-sized
  direct replacements. Updated `review-convention` so `🤖<X>{Y}` is documented alongside
  the already-implemented pair accept/reject behavior.

### 2026-06-19

- Filed as the agent half of pair #000066 M4 (operator chose split (A): own ariadne issue
  + pair #66 deps on it). Contract = `pair/workshop/targets/review-protocol.md`.
