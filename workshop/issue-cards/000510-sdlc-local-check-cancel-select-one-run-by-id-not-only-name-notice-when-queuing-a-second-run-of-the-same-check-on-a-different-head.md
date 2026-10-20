---
id: 000510
status: open
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
github_issue:
---

# sdlc local-check cancel: select one run by id (not only --name); notice when queuing a second run of the same check on a different head

## Problem

`sdlc local-check cancel` selects runs only by `--name`. On 10-20, pair:6 had two queued `suite` runs from one
checkout (an original on an older head, and an affected-check fallback to the full suite on the current head, the
pair#738 case) and could not cancel just one: `--name suite` cancels both and drops the original's FIFO place, so it
had to cancel both and requeue.
