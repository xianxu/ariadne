---
id: 000068
status: open
deps: []
github_issue:
created: 2026-06-02
updated: 2026-06-02
estimate_hours:
---

# fix active-time-v3.py: returns 0 events for these sessions — actuals are fabricated

## Problem

`sdlc close` / `milestone-close` require `--actual <hours>` and tell the agent to
derive it by running `active-time-v3.py` over the issue's commit window. But the
script returns **0 events** for these sessions — flagged as far back as `nous#34`'s
close ("active-time-v3 reported 0 events; session telemetry not captured"), and again
across **~7 closes** in the 2026-06-02 session (shared-brain close-down + `nous#41`),
where every `actual_hours` was a manual guess passed with `--force`/FORCE=1.

Net effect: the velocity-calibration loop the `project` datatype is built around (each
close records `actual_hours` → feeds the velocity-skill validation table) is running on
**fabricated numbers**. The gate has the *form* of calibration with none of the
substance — arguably worse than no gate, because it looks calibrated.

See `brain/data/life/42shots/velocity/{baseline-v3,estimate-logic-v3,SKILL.md}` for the
v3 procedure the script implements.

## Spec / hypotheses to check

- **Operator hypothesis (load-bearing):** the `shared-brain` project went **dormant for
  ~2 weeks** while the operator took detours fixing other things. A look-back / window
  parameter in the script (how far back it scans `~/.claude/projects/.../*.jsonl`
  transcripts, or how it bounds the commit window) may not span that dormancy gap, so the
  per-issue attribution comes up empty. Check the window/look-back params first.
- Confirm whether the script is finding the right transcript directories
  (`~/.claude/projects/-Users-...-{nous,brain,ariadne}/`) — multi-repo sessions (this one
  spanned nous + brain + ariadne) may not all be scanned.
- Confirm the commit-window → transcript-segment join still matches (commit SHAs,
  `--commit-weight`, mention-weighting) after whatever changed since it last produced
  non-zero output.
- Decide the fallback contract: if telemetry is genuinely unavailable for a window, the
  close should say so explicitly and record `actual` as `estimate` (or flag it
  un-calibrated) rather than silently accepting a `--force` guess that pollutes the table.

## Done when

- `active-time-v3.py` returns non-zero, plausible per-issue hours for a real recent
  window that spans a dormancy gap (reproduce against `nous#41`'s window: `c65737a..HEAD`).
- The `--actual` gate gets real data for a normal close, OR a clearly-marked
  "telemetry-unavailable → not calibrated" path exists so the velocity table isn't fed
  guesses.

## Plan

- [ ] Reproduce: run `active-time-v3.py` over `nous#41`'s window + a dormant-project
      window; confirm 0 events and locate where it drops to empty.
- [ ] Fix the window/look-back (and multi-repo transcript scan) so real sessions attribute.
- [ ] Define the telemetry-unavailable fallback contract for `close`.

## Log

### 2026-06-02

Filed from the sdlc tooling retro
(`workshop/pensive/2026-06-02-01-pensive-sdlc-tooling-retro.md`, finding F2). Operator
flagged the dormancy-window hypothesis as the likely cause.
