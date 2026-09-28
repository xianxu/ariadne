---
id: '000065'
status: done
created: 2026-06-02
updated: 2026-06-02
estimate_hours: 0.5
actual_hours: 0.5
---

# single-pass atomic work should use plain checkboxes not an M1 tag — avoids redundant milestone-close + issue-close double-log

## Problem

Closing #63 (a single-pass, atomic task) produced two near-identical `## Log`
lines:

```
- 2026-06-02: closed — <verified>          (sdlc close)
- 2026-06-02: closed M1 — <verified>       (sdlc milestone-close --milestone M1)
```

Root cause is **guidance, not a tool bug**. Both verbs funnel through
`runClose` (`milestoneclose.go:99` → `close.go:361-368`), which appends one
`- <date>: closed [Mx] — <verified>` line per call when `--verified` is set.
#63's Plan carried a single `- [x] M1 — …` row; the issue-close guard requires
*each milestone listed in `## Plan`* to carry a `Review-Verdict:` trailer, and
only `milestone-close` produces that trailer. So tagging the one-shot task `M1`
*forced* a `milestone-close M1` (log line #1) before the issue `close` (log
line #2). For a multi-milestone issue this reads as a clean progression
(`closed M1`, `closed M2`, … then `closed`); with exactly one milestone the two
lines carry the same verified text and sit adjacent → reads as a dup.

AGENTS.md §3 actively invites this: *"Single-pass work → one milestone, not
three"* — which says to tag atomic work as `M1`, the very thing that causes the
double-log.
