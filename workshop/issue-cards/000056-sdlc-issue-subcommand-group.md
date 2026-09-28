---
id: '000056'
status: done
created: 2026-05-31
updated: 2026-05-31
estimate_hours: 7
actual_hours: 1.5
---

# Lift the issue subsystem into `sdlc issue`

## Problem

Issue creation/editing is agent-driven prose. The `xx-issues` skill tells the
agent to *"find the highest existing ID and increment by 1"* and write the
template by hand — a deterministic step done manually, which is racy under
parallel workstreams and off-by-one prone. The deterministic core already
exists inside `fetch.go` (`nextIssueID` + `renderFetchedIssue`) but isn't
surfaced as a reusable verb. The subsystem is mature (100s of issues) and
deserves a first-class binary surface, the way SDLC checkpoints got one.
