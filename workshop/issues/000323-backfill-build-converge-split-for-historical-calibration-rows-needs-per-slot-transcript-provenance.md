---
id: 000323
status: open
deps: []
github_issue:
created: 2026-10-10
updated: 2026-10-10
estimate_hours:
card_mirror: 'b6223cb995d52969f50ab44ac1c4e0655e332726' # card fields mirrored from issue-cards; edit via sdlc
---

# Backfill build/converge split for historical calibration rows (needs per-slot transcript provenance)

## Problem

ariadne#234 splits each issue's measured actual into **build** and
**converge** time at its boundary reviews. It does that for new closes only.
The 762 historical calibration-ledger rows keep `split_source: -`, so
`k = converge / build` and est-vs-build drift have no history.

Split out of #234 (TL decision 2026-10-10) because historical actuals cannot
be reproduced:
- they were measured from slot worktree cwds;
- transcript selection is per cwd (`internal/transcripts`);
- which slot's transcript dir a measurement read was never recorded.

Re-measuring from :0 would read the wrong or no transcripts, and the rows
also predate #317/#321's attribution fixes.

## Spec

- First decide whether a faithful backfill is possible. Options:
  - re-measure over every `<repo>-slot*` transcript dir plus the repo's own
    dir, with #270/#321 scoping limiting cross-slot bleed;
  - or apply the converge fraction from a re-measure to the recorded actual.

  Either way, rows are labelled with their method, and no number is guessed.
- Cut-point evidence, best first: a stamped `dispatched` round (#234); the
  gate ledger's first round per boundary (finalize time, so it is an
  approximation); the review sidecar; the first `Review-Verdict:` trailer.
  Rows with none stay `split_source: -`.
- Run a dry-run on a sample, have the operator check it, then the full
  ledger (AGENTS.md §10). Document the method in brain's `velocity/SKILL.md`.
- Old-rule actuals (milestone tails included) must not mix silently with
  #234's new-rule actuals in drift.

## Done when

- The method is decided, and its fidelity is evidenced on a sample.
- Rows with recoverable evidence carry a labelled split; the rest stay `-`.

## Plan

- [ ]

## Log

### 2026-10-10

- Filed by ariadne:3 from #234 at the TL's direction: backfill out of #234,
  not in the MVP.
