---
id: 000034
status: open
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
