---
id: '000135'
status: done
started: 2026-06-26T18:16:39-07:00
created: 2026-06-26
updated: 2026-06-26
estimate_hours: 2.65
actual_hours: 0.76
---

# sdlc close actual not applicable

## Problem

`sdlc close --no-actual` can currently produce an invalid closed issue: it flips `status: done` while omitting `actual_hours`, but the issue schema requires `actual_hours` for done issues. This happened during pair#72 and was caught by the boundary review.

There are legitimate closes where focused dev-hours are not applicable or cannot be measured: roadmap-only records, wontfix/punt-like administrative closes, or cases where telemetry is unavailable and recording a made-up number would pollute velocity calibration.

The schema and close command need an explicit, valid "not applicable" representation instead of relying on a missing field.
