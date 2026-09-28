---
id: '000140'
status: done
started: 2026-07-01T00:27:31-07:00
created: 2026-06-29
updated: 2026-07-01
estimate_hours: 0.67
actual_hours: 0.75
---

# sdlc boundary review should show progress while waiting

## Problem

Boundary reviews can run silently for several minutes. In pair#84, `sdlc close`
printed:

```text
dispatching boundary review (...HEAD) via claude ...
```

and then produced no progress output for multiple 60-second polling intervals.
The only way to tell "still working" from "wedged" was to inspect the process
tree and infer that a `claude -p` child still existed.

This creates operational friction for agents and humans:

- long silent waits look like hangs;
- operators do not know whether the subprocess is making network/model progress;
- agents may be tempted to interrupt a valid review;
- genuine stalls are hard to distinguish from normal review latency.
