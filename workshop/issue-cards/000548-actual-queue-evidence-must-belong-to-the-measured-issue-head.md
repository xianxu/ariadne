---
id: 000548
status: open
created: 2026-10-22
updated: 2026-10-22
estimate_hours:
github_issue:
severity: critical
---

# Actual queue evidence must belong to the measured issue head

## Problem

Echo#9 actual preview refuses because another issue's recorded suite attempt lies in its measurement period while echo:0 holds more than one claim. The suite attempt belongs to echo#3, not echo#9. This blocks echo#9's fix and therefore echo#10, #3 and #5 suites/landings.
