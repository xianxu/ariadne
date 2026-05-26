---
id: 000034
status: open
deps: [000033]
created: 2026-05-26
updated: 2026-05-26
estimate_hours:
---

# Brainstorming skill: gist vocabulary + default-acceptance signal capture

## Problem

The adapted `superpowers-brainstorming` skill produces clarifying questions as natural prose without structured tags. This blocks two things we want later:

1. **No countable question shapes.** Sessions don't accumulate into a queryable record of "which kinds of questions keep being useful." Without a tag like `scope` or `alternative` on each question, comparing across sessions requires LLM re-parsing of every transcript.

2. **Default-acceptance signal is captured but unstructured.** `AskUserQuestion` already records the user's selection in the transcript, including which option was marked Recommended. But correlating `gist → default-accepted%` requires extra parsing each time, and the gist isn't there to correlate against.

The motivating intuition: we want to accumulate data so that later analysis can answer "for scope questions, the recommended answer is accepted 85% of the time — consider auto-answering or moving from `AskUserQuestion` to `AssumeAndConfirm`." gstack achieves this via named question templates. We want the same end state, but bottom-up: the vocabulary crystallizes from use rather than being curated upfront.

## Spec

Two pieces, sequenced.

### Piece 1 — Starter gist vocabulary as light prose

Add to the brainstorming SKILL.md (via `/construct adapt superpowers`) a small named vocabulary the agent draws from when tagging clarifying questions:

- `scope` — what's in scope / out of scope
- `alternative` — choice between approaches
- `success-criteria` — how we know the work worked
- `constraint` — fixed external limits to honor
- `assumption` — implicit belief worth surfacing
- `prior-art` — existing work / precedent to learn from
- (extend as patterns emerge)

The agent tags each clarifying question inline: `What's the scope here? (gist: scope)`. The vocabulary acts as **suggestion, not enforcement** — agents introduce new tags when nothing fits, and those become candidates for incorporating into the vocabulary later. This is the same crystallization pattern targets use: bottom-up curation from accumulated use.

### Piece 2 — Default-acceptance signal capture

When `AskUserQuestion` is used during a brainstorm, the agent already follows the harness convention of putting the recommended option first with the "(Recommended)" suffix. What we need to add: a lightweight inline log after the user answers, capturing the gist + whether the user picked the default.

Shape (open — pick what's actually parseable later):

    <!-- brainstorm-log: gist=scope chose=default option=A -->
    <!-- brainstorm-log: gist=alternative chose=other option=C -->

Land in the agent's next message after the answer, before continuing the brainstorm. Markers are HTML-comment so they don't render to the user but stay grep-able in transcripts and committed issue files.

### Out of scope (future issues)

- **Analysis tooling.** Once enough data accumulates (~10-20 brainstorm sessions with markers), a separate issue extracts: which gists recur, which gists' recommended answers get accepted at high rates, which gists' alternatives explore meaningfully different ground.
- **Auto-answer mechanism.** gstack-style: if `gist=X` has its default accepted ≥90% of the time, skip the question and assume-and-confirm. Designed only after analysis exists.
- **Per-mode gist variants.** The brainstorming-mode distinction (crystallization / feasibility / domain-learning, per issue exploration on 2026-05-26) may want different starter vocabularies per mode. Defer until single-mode-vocabulary data exists.

## Plan

To be detailed when starting. Rough shape:

- [ ] **M1 — Vocabulary prose.** Through `/construct adapt superpowers`, add the starter gist vocabulary to the brainstorming SKILL.md. Intent file at `construct/intents/superpowers/ariadne.md` captures the rationale. Depends on #33 (adaptation-system narrowing) if that lands first; otherwise just adapt under current `/construct adapt` semantics.
- [ ] **M2 — Default-acceptance marker convention.** Decide marker shape, add to the same skill via the same adaptation, document in the skill prose. Test by running a brainstorm and checking that markers appear in the agent's responses.
- [ ] **M3 — Validation pass.** Run 2-3 real brainstorms end-to-end. Verify the agent applies tags consistently and the markers are grep-able.

## Log

### 2026-05-26 — issue created

Issue extracted from conversation about brainstorming-stage scaffolding (the SDLC walk-through covering stages 1 and 2). The user's stance: superpowers-brainstorming's emergent-question pattern is fine, but it can't accumulate compounding value without structured tags. gstack's curated-question vocabulary has the right end-state but the wrong starting point — bottom-up crystallization beats top-down curation for engineering ideation, which spans more shapes than product ideation.

Deferred items at issue-creation time:
- Brainstorming-mode prose (crystallization / feasibility / domain-learning) — also wants to land via `/construct adapt`; could land in the same adaptation run or separately.
- The whole analysis layer — only after the markers have produced data.
