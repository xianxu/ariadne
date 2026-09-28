---
id: '000068'
status: done
created: 2026-06-02
updated: 2026-06-02
estimate_hours: 3
actual_hours: 2.0
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
