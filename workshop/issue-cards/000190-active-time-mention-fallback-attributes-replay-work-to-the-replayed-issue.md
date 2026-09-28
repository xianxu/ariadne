---
id: '000190'
status: done
started: 2026-07-29T16:23:09-07:00
created: 2026-07-29
updated: 2026-07-29
estimate_hours: 3.75
actual_hours: 0.86
---

# active-time mention-fallback attributes replay work to the replayed issue

## Problem

The active-time engine's mention-fallback attributes work to whatever issue the session
TALKS about, so replaying or postmortem-ing issue X charges the time to X rather than to the
issue doing the replaying.

Measured live at ariadne#187's close: **46.1m/77% attributed to pair#127** by
`mention fallback without issue commit boundary`, because #187's Task 14 replayed pair#127
and its commits, prose and evidence file cite `#127` constantly. #187 measured 2.29h; the
true figure is higher, and pair#127 — long closed — gained 46 minutes it did not spend.

Both directions are wrong in a way that matters, because these numbers feed velocity
calibration: the replaying issue looks cheaper than it was, and a CLOSED issue's actual
silently grows after the fact.

This is a general hazard, not a #187 quirk. Any issue whose work is *about* another issue —
replays, postmortems, migrations, "fix the thing #N introduced" — hits it. The engine
already knows the difference in principle: it warns `without issue commit boundary`, meaning
it had no commit anchoring the segment and fell back to text mentions.
