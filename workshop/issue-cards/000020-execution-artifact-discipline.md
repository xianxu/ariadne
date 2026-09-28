---
id: '000020'
status: done
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
