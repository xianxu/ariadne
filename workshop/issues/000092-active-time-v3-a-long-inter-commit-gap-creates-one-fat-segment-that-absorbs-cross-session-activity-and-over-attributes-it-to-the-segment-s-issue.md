---
id: 000092
status: open
deps: []
github_issue:
created: 2026-06-11
updated: 2026-06-11
estimate_hours:
---

# active-time-v3: a long inter-commit gap creates one fat segment that absorbs cross-session activity and over-attributes it to the segment's issue

## Problem

The active-time-v3 engine (`cmd/sdlc/internal/activetime`, #110) is
segment-anchored: it cuts the timeline at every commit in the window and
attributes each segment's active-minutes to that segment-ending commit's issue
refs (at `--commit-weight 1.0`, **100%** by commit). The per-event 15-min gap cap
(`activeMinutes`) correctly
stops a single idle gap from counting as hours — but it does **not** bound a
*segment's* width. When an issue goes a long calendar time between two of its
commits, that one segment spans the whole gap and **vacuums up every transcript
event in the watched dirs (repo + brain) during that span** — including unrelated
other-session activity and brain autosave/sync — and pins all of it on the
segment's issue.

### Concrete evidence (nous#48, the issue that surfaced this)

`sdlc actual --issue 48` reported **14.47h**, vs the per-milestone measurements
of M1 1.11h + M2 4.85h. The v3 per-segment table shows why — one row dominates:

```
33  2026-06-08 16:55 → 2026-06-11 10:03   779.5 min   3fa9414 #48 M2: certify real-Microsoft   #48=779.5m
```

That single segment is **779.5 min ≈ 13h, attributed 100% to #48 — ~90% of #48's
measured total (870 min).** It spans the **2.7-day gap** between the last 06-08
commit (`ee6a387`, 16:55) and the first 06-11 commit (`3fa9414`, 10:03), i.e. the
stretch where the operator stepped away. The properly-anchored, short-span #48
work segments sum to only ~1.4h; the genuine focused effort was ~4h. The other
~10h is **cross-session contamination** absorbed by the fat segment: an
interrupted multi-hour resume session plus unrelated `nous`/`brain` transcript
activity over those 2.7 days, none of it filtered because v3 can't tell a
06-09 brain-autosave event from #48 work — it only sees "an event inside the
#48-anchored segment."

### Why it matters

`sdlc close` treats the v3 number as the authoritative `actual_hours` for velocity
calibration (#68 built v3 precisely so actuals are *measured, not guessed*). A 3–4×
over-count on any issue whose work straddles a multi-day pause silently poisons the
calibration baseline — and "work spanning a pause" is common (review waits,
operator-collaborative steps, weekends). The failure is invisible: the number
*looks* measured. nous#48 only caught it because the operator questioned a
suspiciously high close number; most won't.

The script's own docstring already flags the adjacent gap: *"Parallel-session
dedup not yet implemented for v3 (rare in practice for the operator)."* This is
that gap biting in practice.

## Spec

Bound how much a single segment can absorb, so a long inter-commit gap can't
attribute cross-session activity to the gap's issue. Candidate mechanisms (pick
after measuring against the recorded human baselines — `brain/data/life/42shots/
velocity/baseline-v3.md` — so the fix doesn't *under*-count normal sessions):

1. **Intra-segment gap split (cheapest, most targeted).** Within a segment, split
   into sub-runs wherever the inter-event gap exceeds a "session-break" threshold
   (e.g. ≥ N×`threshold-min`, or a flat 2–4h). Each sub-run is its own active-time
   block; only sub-runs adjacent to the anchoring commit count toward it, or all
   sub-runs count but the segment can no longer be a single 13h block bridging
   distinct work sessions. This keeps a genuinely-continuous long session intact
   while breaking a 2.7-day span into the real bursts it contains.
2. **Session-membership filter.** Treat a transcript session (`.jsonl` file, or a
   maximal sub-15-min-spaced run) as the unit; only count sessions that actually
   contain commits for / mentions of the segment's issue. Drops unrelated
   concurrent sessions that merely overlap the segment's time bounds.
3. **Segment-width cap.** Hard-cap any single segment's counted active-minutes
   (e.g. at the 95th percentile of normal segment widths), logging when the cap
   fires so the truncation is visible (no silent caps — the AGENTS.md rule).

Whatever is chosen, the v3 output / `sdlc actual` should **flag** when one segment
contributes an outsized share (e.g. ">50% of an issue's total from a single
segment spanning >Xh") so a contaminated number is visible at the gate rather than
silently recorded.

Independently, document the **cheap behavioral mitigation** in the SDLC guidance:
commit more often (segments stay short → less to absorb), and don't let an issue
sit mid-work across a multi-day pause without a commit boundary.

## Done when

- A single multi-day inter-commit gap can no longer attribute >~1 session's worth
  of cross-session activity to the gap's issue; re-running v3 over the nous#48
  window attributes ≈ the focused-effort figure (~4h), not 14.47h.
- The fix is validated against the recorded human baselines (`baseline-v3.md`) to
  confirm it doesn't regress normal single-session attribution (the v3 method was
  tuned to within ~5% of human numbers — keep that).
- `sdlc actual` / v3 surfaces a visible flag when one segment dominates an issue's
  total (contaminated-number warning at the close gate).
- SDLC guidance notes the commit-frequency mitigation.

## Plan

- [ ] Reproduce: capture the nous#48 window (`--since 2026-06-08T10:44 --until 2026-06-11T10:18`, `--dir` nous+brain) as a regression fixture; assert the current 779.5-min single-segment behavior.
- [ ] Prototype mechanism (1) intra-segment gap split; measure the nous#48 window + the `baseline-v3.md` human-baselined sessions; compare.
- [ ] If (1) under/over-shoots, evaluate (2) session-membership filter and/or (3) width cap.
- [ ] Add the dominant-segment warning to v3 output + surface it via `sdlc actual` / the close `--actual` suggestion.
- [ ] Document the commit-frequency mitigation in the SDLC guidance.

## Log

### 2026-06-11

Filed from nous#48 (shim(oauth) Microsoft n=2-real). That issue's close measured
14.47h; the operator questioned it; investigation traced ~90% of the figure to one
2.7-day-spanning segment (table above) and corrected nous#48's `actual_hours` to a
labeled judgment of 4.0h. This issue is the upstream fix so the artifact doesn't
silently recur. Source: `cmd/sdlc/internal/activetime/segment.go`
(`attributeSegment` / `buildSegments`); the package doc notes
"parallel-session dedup not yet implemented" as a known gap.
