---
id: 000169
status: codecomplete
deps: []
github_issue:
created: 2026-07-13
updated: 2026-07-13
estimate_hours:
started: 2026-07-13T16:10:13-07:00
actual_hours: 1.03
---

# ariadne stack introspection 3

time to run an /xx-introspect for the last couple of months sessions.

## Problem

## Spec

## Done when

- introspect run-3 executed over the last-couple-months corpus, with the human in
  the loop at each judgment checkpoint.
- Cluster synthesis reviewed; skill changes applied or explicitly declined
  (precision over recall — no force-fitting).
- `introspect-state.json` updated; pending hints resolved.

## Plan

- [x] Stages 1–4: scope → normalize → classify (fan-out) → detect
- [x] Stage 5 synthesis reviewed with user
- [x] Decision: no skill changes (diminishing returns); hint consumed; state bumped

## Log

### 2026-07-13
- 2026-07-13: closed — introspect run-3 completed over 449 real sessions (all-projects, since 2026-05-27), Stages 1-7 with user in loop. 543 moments detected; 39 high-value (11 redirect + 28 endorsement) all reinforced EXISTING rules, 0 new robust rules, 0 friction → per precision-over-recall, user chose option A: NO skill changes. Consumed probe-before-rm hint (already deployed); bumped introspect-state.json. Meta-finding: introspect at diminishing returns (logged to #169/#170). No repo/architectural surface (outputs are user-global ~/.claude) → --no-atlas. Follow-ups filed: #172, #173.; review verdict: SHIP

**introspect run-3 in progress.** run-id `20260713T161752` (cache under
`~/.claude/introspect/cache/20260713T161752/`).

- **Stage 1 scope:** `all` projects (matches run-2). Note: introspect reads only
  Claude Code transcripts (`~/.claude/projects`) — **codex/other-agent sessions
  are not in the corpus** (known gap; relevant to #170).
- **Stage 2 normalize:** `--since 2026-05-27T18:35:00Z` (run-2's `last_run_at`).
  2071 files / 164,857 events / 2743 segments.
- **Corpus scoping (decided w/ user):** excluded the ephemeral temp stores
  (`/private/var/folders/.../T`, `/private/tmp/claude-501`) = **1,081 utility-
  prompt segments, pure noise** (orientation-slug / changelog-extraction calls,
  1-2 msgs each). Then applied a **depth floor amc≥15** → **970 real-repo
  segments** (`filtered.json`). Per-project: ariadne 242, pair 224, brain 213,
  parley 152, metis 92, kbench 22, kaggle 13, others ~12.
- **Stage 3 classify:** fanned out to 10 subagents (`cls-batch-01..10.json`),
  ~97 segments each. Aggregated clean (970 records, 0 dup/missing, all valid).
  Raw distribution: **code-review 495**, implementation 278, planning 63,
  debugging 42, exploration 40, brainstorming 26, ambiguous 25, out-of-scope 1.
- **Big finding — the 495 "code-review" sessions are auto-dispatched SDLC judge
  prompts firing as their own Claude sessions** (verbatim `milestone-review` /
  `plan-quality` / `specs` prompts; no human turns → no taste signal).
  **Excluded** (→ skip), along with ambiguous/out-of-scope. Real taste corpus =
  **449 sessions** across exactly the 5 deployed skills (impl 278, planning 63,
  debugging 42, exploration 40, brainstorming 26).
- **Cross-finding for #170/#172:** 495 judge-sessions + 1,081 utility-prompts =
  **1,576 / 2,743 segments (57%) are machine-driven automation, not human work.**
  The SDLC review machinery alone spawned ~495 sessions in 7 weeks — ties to the
  `--no-judge` bypass being the #1 gate skip (#172 baseline).
- **Stage 4 detect:** 543 moments from 449 sessions — but **93% edit-after-edit
  (504)**, only **11 redirect + 28 endorsement**, **0 friction**. edit-after-edit
  is iteration noise (top files `main.go`/`init.lua`/`Makefile` — expected churn,
  no anomalous pattern).
- **Stage 5 synthesis:** the 39 high-value moments **reinforce existing rules**,
  add ~nothing new. Endorsements are routine "yes, go ahead" (→ reinforces impl
  *Default to action on readiness*); redirects map to existing rules (*Report-
  symptoms ≠ done*, *Verify API before coding*, *converge by attrition*). One
  borderline candidate (impl *split/defer tangential work to its own ticket*) was
  **rejected** — overlaps the constitution's scope-discipline / `side-quest:`
  (fails the D1 dedup rule from #170).
- **Decision (user, option A): no skill changes.** Precision over recall — don't
  manufacture rules. Consumed the `probe-before-rm` hint (already reflected in the
  deployed `introspect-debugging`) → `versions/20260713T161752/consumed-hints/`.
  Bumped `introspect-state.json` `last_run_at`; recorded run.

**Meta-finding (the real result): introspect has hit strong diminishing returns.**
7 weeks × 449 substantial sessions → ~0 robust new rules, 0 friction. Evidence the
constitution + existing 22 rules capture the taste well now. Feeds #170 Q2.6
(is distilled knowledge useful) and argues the **run cadence should stretch**.
Validates #170's D1 downgrade (introspect rarely surfaces anything not already
codified).

Also filed **#173** (make introspect ingest codex transcripts) — this run's corpus
was Claude-only by construction, an agent-neutrality gap.

Scale note: this 7-week window is ~10× prior runs (very heavy usage); interactive
stages were adapted via subagent fan-out rather than one-context classification.
