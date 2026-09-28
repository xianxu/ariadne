---
id: '000139'
status: done
started: 2026-06-30T12:11:03-07:00
created: 2026-06-29
updated: 2026-06-30
estimate_hours: 0.63
actual_hours: 1.70
---

# sdlc close should finalize metadata after review verdict

## Problem

`sdlc close` currently mutates the issue file before the boundary review has
returned. In pair#84, the first close attempt flipped the issue to `status:
done`, wrote `actual_hours`, and appended a "closed" log line, then the boundary
review returned `REWORK`.

That leaves the repo in an awkward intermediate state:

- the issue says `done` while the boundary says "do not cross yet";
- rerunning close requires `--no-reclose-guard`;
- repeated close attempts append stale close log lines that the operator must
  clean up manually;
- the close bookkeeping itself becomes tangled with the code-review loop.

The status flip should be the last successful act of close, not something that
happens before the gate has established that the boundary can be crossed.
