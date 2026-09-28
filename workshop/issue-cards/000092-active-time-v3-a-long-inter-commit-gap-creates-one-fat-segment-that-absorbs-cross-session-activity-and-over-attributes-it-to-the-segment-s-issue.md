---
id: '000092'
status: done
started: 2026-06-29T13:23:59-07:00
created: 2026-06-11
updated: 2026-06-29
estimate_hours: 5.95
actual_hours: 2.86
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
