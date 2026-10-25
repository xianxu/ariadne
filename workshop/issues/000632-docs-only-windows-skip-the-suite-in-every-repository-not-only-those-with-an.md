---
id: 000632
status: open
deps: []
github_issue:
created: 2026-10-25
updated: 2026-10-25
estimate_hours:
card_mirror: '9f2c34406ed110e2a080f20b8b20b435ec67771f' # card fields mirrored from issue-cards; edit via sdlc
---

# Docs-only windows skip the suite in every repository, not only those with an affected declaration

## Problem

## Spec

**A docs-only window runs the full suite in repositories without an affected declaration (ops-echo:0, 10-25).**
echo#39 changes only a design plan and its issue file, yet `sdlc local-check run --pre-close` selected the full echo
suite, because echo has no `.sdlc/affected.json` (and trains are off); it waited fourth in the machine queue behind
ariadne and pair suites to land a document. Where a repository declares affected inputs, sdlc already records a
docs-only pass without invoking the runner; that shortcut shouldn't depend on the declaration.

Recognize a docs-only window (only `workshop/`, `atlas/`, `docs/`, Markdown and similar non-code paths, using the same
classification the close gate's docs-only rule already uses) in every repository, and record the pass without running
the suite. Anything that isn't clearly docs falls back to today's behavior.

## Done when

- In a repository without `.sdlc/affected.json`, a docs-only window's pre-close records a pass without a suite run;
  a window touching code still runs the full suite. Tests for both, plus the declared-affected case unchanged.
- The docs-only classification is one shared function used by the close gate and local-check.

## Plan

- [ ]

## Log

### 2026-10-25
