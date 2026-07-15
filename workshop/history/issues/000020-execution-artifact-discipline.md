---
id: 000020
status: done
deps: []
github_issue:
created: 2026-05-04
updated: 2026-05-04
estimate_hours: 2
actual_hours: 1
---

# Execution artifact discipline

## Problem

Execution worked well across charon-launch-push (#13/#14/#15/#16, four issues closed in 5 days under the v2-era workflow), but the retrospective surfaced consistent thin spots in how *context for future-you* gets captured. Specifically:

1. **Status drift across artifact layers.** The brain portfolio file (`data/project/charon-launch-push.md`), the per-issue `## Log` sections, and `atlas/` updates fall out of sync under load. Example from charon-launch-push: M4 was committed (`fd4daf2`) and e2e-verified before the brain project file ticked the box.
2. **Side quests are real and unbudgeted.** During charon-launch-push: the `make dev` signed-binary trap (~30 min), OSC 8 hyperlinks (~40 min, not in any plan), in-session per-screen cursor memory (~45 min, also not planned), action-hint visibility tweak. These were the *right* things to do during e2e but appear in no estimate, no plan, no traceable artifact. Aggregate roughly ~20% of project effort.
3. **Mid-stream scope events overwrite original intent.** When `workshop/plans/*-plan.md` gets revised mid-stream (e.g., #15's M4b broadening), the original section is rewritten in place. Future readers can't reconstruct what *would* have shipped under the original scope vs. what got pulled in. For estimate calibration that's a real loss.
4. **Meta-discussion lives only in transcripts.** The 16-hour conversation that drove charon-launch-push's M4b posture flip, OSC 8 vs ctrl+o trade-off, default-preserve revoke decision, etc. is captured in `~/.claude/projects/*.jsonl` — not in any artifact except as outcomes (code comments, issue `## Log`). Re-entering cold requires reading transcripts.
5. **Estimates without `actual_hours` are calibration-dead.** v2.1 calibration (just landed) only worked because actuals could be derived from transcripts. Going forward, capturing `actual_hours` on issue close should be mechanical, not archaeological. The `construct/datatype/project.md` datatype already mentions this; xx-issues skill doesn't enforce it on close.
6. **Atlas updates land at end-of-project, not per-milestone.** The constitution says "as you go" but each milestone's delta doesn't *feel* big enough; updates accumulate to a final docs sweep. Future agents pay the cost when they read a stale atlas.
7. **Self-improvement-loop pattern is implicit.** The velocity skill (paired versioned files + provenance + validation log + bump rules) is a pattern that generalizes — threat models, prompt templates, decision logs — but ariadne doesn't name the pattern.

Source for these findings: a retrospective on charon-launch-push closure (May 4 2026) recorded in the conversation transcript at `~/.claude/projects/-Users-xianxu-workspace-brain/b0f3bfdd-*.jsonl`. The findings are durable enough to formalize as ariadne discipline, not project-specific noise.

## Spec

Land 7 changes across ariadne's propagation surfaces (`AGENTS.md`, `construct/local/issues/SKILL.md`, `construct/datatype/project.md`). After landing, descendants pick up changes via `make refresh`.

### Mapping (one row per finding)

| # | Finding | Target file | Change shape |
|---|---|---|---|
| 1 | Status drift | `AGENTS.md` §8 (Maintenance of Atlas) — extend with parallel "Maintenance of Project File" rule. New §5b "Closing checklist" — flips status, ticks tasks, updates parent project file, records `actual_hours`, atlas update, all in one sweep. | Existing-rule sharpen + new checklist |
| 2 | Side quests budgeted | `construct/local/issues/SKILL.md` — new optional `## Side quests` subsection within `## Log`. One line per side quest: name + ~time + commit ref. Recommended for multi-day issues, not required for one-sitting issues. Pairs with #6 commit convention. | New convention in xx-issues |
| 3 | Mid-stream scope events without intent traces | `AGENTS.md` §1 (Artifact hierarchy) — when revising `workshop/plans/*-plan.md`, append a `## Revisions` section with timestamp + reason + delta, **don't overwrite the original section in place**. Cross-reference in `construct/datatype/project.md` so scope events are visible at portfolio level too. | New rule + datatype reinforcement |
| 4 | Meta-discussion lives only in transcripts | `construct/local/issues/SKILL.md` — formalize a `## Log` sub-format **"Session summary"**: per major sitting (or milestone close), one paragraph naming what was attempted, what landed, what got deferred, in-flight design decisions. Replaces "you'd have to read transcripts" with explicit handoff. | New sub-format in xx-issues |
| 5 | `actual_hours` on close | `construct/local/issues/SKILL.md` "Closing an issue" — make `actual_hours` REQUIRED at close. `AGENTS.md` §5 (Verification Before Done) — add to "done" criteria: actual hours recorded, validation log entry written. | Enforce existing convention |
| 6 | Side-quest commit tagging | `AGENTS.md` — new short §12 "Commit conventions". Pattern: `<area>: <verb>: <subject>` where `<verb>` includes `side-quest:` for unbudgeted work. `git log --grep "side-quest:"` becomes cheap. | New section |
| 7 | Atlas per-milestone (not end-of-project) | `AGENTS.md` §8 — sharpen "continuously update" → "**at each milestone close**, before moving to next milestone." Pair with closing checklist from #1. | Existing-rule sharpen |
| 8 | Self-improvement-loop pattern | `AGENTS.md` §4 (Self-Improvement Loop) — extend with a generalization: *for any policy/playbook deliverable likely to drift (estimates, prompt templates, threat models, decision logs), follow the **paired-versioned-files + calibration-loop pattern**: paired `<thing>-vN.md` and `<thing>-baseline-vN.md` files; provenance string on every produced output; validation log appended on each project close; bump rules documented.* Brain's velocity skill is the canonical instantiation that lives in personal state because the calibration data is personal; descendants build their own when they need one. | Generalize existing rule |

### Why velocity skill stays in brain (CON #4 from the conversation)

The trio architecture — **charon = outbound capability / nous = task capability / brain = personal state** — places calibration data in personal state by design. Baseline files contain timestamps, project specifics, and individual coding pace; they're not portable across operators. AGENTS.md §4 describes the *pattern* abstractly so any future ariadne adopter can build their own calibration without inheriting `xianxu`'s constants. Brain's `data/life/42shots/velocity/` is one instantiation, cited as canonical.

## Plan

Land in 4 sequential commits, all referencing this issue:

- [x] **Commit 1** `5ace498`: `AGENTS.md` updates — closing checklist (in §5), commit conventions (new §12), atlas/project-file/scope-revision rule sharpens (§1, §8), self-improvement-loop generalization (§4).
- [x] **Commit 2** `bb037d3`: `construct/local/issues/SKILL.md` updates — `actual_hours` required at close, `## Side quests` subsection format, `## Log` "Session summary" sub-format.
- [x] **Commit 3** `2cb066c`: `construct/datatype/project.md` reinforcement — scope events visible in project file; reference back to AGENTS.md §1 revisions rule.
- [x] **Commit 4** (this): atlas/workflow/issue-lifecycle.md + atlas/index.md touch-ups pointing at the new conventions.
- [ ] After landing, run `make refresh` in brain / charon / nous so descendants pick up. Verify with one cycle (e.g., next charon issue authored after this lands). **(Operator follow-up)**

## Estimate

Range: 1.5–4 hr. Best guess: ~2.5 hr.

Produced via `brain/data/life/42shots/velocity/estimate-logic-v2.1.md` against `baseline-v2.1.md`. Method A only.

| Item | Primitive | Spec quality | Design (hr) | Impl (hr) | Total |
|---|---|---|---|---|---|
| AGENTS.md edits (4 sub-changes) | Atlas/docs maintenance | ×0.2 (this issue IS the spec) | 0.05–0.2 | 0.3–0.6 | 0.35–0.8 |
| xx-issues SKILL update | Atlas/docs maintenance | ×0.2 | 0.05–0.2 | 0.2–0.5 | 0.25–0.7 |
| project datatype touch | Atlas/docs maintenance | ×0.2 | 0.05–0.2 | 0.1–0.3 | 0.15–0.5 |
| README + atlas index | Atlas/docs maintenance ×2 | ×0.2 | 0.05–0.2 | 0.1–0.3 | 0.15–0.5 |
| `make refresh` in 3 descendants + sanity check | Smaller Go module (mirror — refresh is mechanical) | ×0.2 | 0 | 0.2–0.5 | 0.2–0.5 |
| **Subtotal (design / impl)** | | | **0.2–1** | **0.9–2.2** | 1.1–3.2 |
| **+15% on design (thorough plan: this Spec)** | | | +0.03–0.15 | n/a | +0.03–0.15 |
| **Total** | | | | | **1.5–4 hr** |

Caveats:
- All work is documentation/convention — no risky code changes.
- Real risk is in **wording precision**: rules that are too soft get ignored, rules that are too rigid get worked around. Budget for one round of user review on AGENTS.md text.

## Log

### 2026-05-04 — session summary

Filed and implemented in one sitting following the charon-launch-push retrospective. Spec was thorough (worked from a target-file mapping prepared during the retrospective conversation), so all 4 commits landed mechanically — no design pivots, no scope events.

Token cost: AGENTS.md grew **2185 → 3256** (cl100k_base, +1071 tokens, +50%). Operator flagged this as worth shrinking later; content density preserved over brevity for now.

Wording-precision concern noted in the estimate caveats — the 4-commit set is a single review surface for the operator. Self-review pass before close: rules read as "discipline you follow" rather than "fences that block you," and they cross-reference each other consistently (closing checklist in §5 → atlas/project-file rules in §8 → versioned-playbook pattern in §4 → commit conventions in §12 → xx-issues skill enforcement → project datatype scope-event logging). Each rule names the *why* (the charon-launch-push observation that prompted it) so future-you can decide whether the cost is worth keeping.

Velocity calibration: estimated 1.5–4 hr (mid 2.5) under v2.1 Method A; actual ~1 hr. Below the low end. Drivers: spec quality was unusually high (this issue's Spec section IS the design), the work was 100% docs/convention with zero discovery, and no review iteration was needed. v2.1 over-estimated this issue by ~2.5×; logged as a v2.1 validation entry below.

### 2026-05-04 — Commits in flight

- `b17da01` issue filed
- `5ace498` AGENTS.md (Commit 1)
- `bb037d3` xx-issues SKILL (Commit 2)
- `2cb066c` project datatype (Commit 3)
- `(this commit)` atlas touch-ups + close (Commit 4)

## Validation log entry

For appending to `brain/data/life/42shots/velocity/estimate-logic-v2.1.md`'s validation table — the first post-v2.1 actual.

| Project | Estimated (low–high) | Actual hours | Off by | Notes |
|---|---|---|---|---|
| ariadne #20 (execution artifact discipline) | 1.5–4 hr | ~1 hr | 2.5× over (vs mid 2.5 hr) | All docs/convention work; spec quality ×0.2 across all primitives; no review iteration. v2.1 still slightly conservative for pure-docs issues with thorough specs — confirms the "plan-doc-thoroughness rewarding" open question from baseline-v2.1. Single data point; revisit when 2nd post-v2.1 actual lands. |
