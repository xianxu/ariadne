---
id: 000437
status: open
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
github_issue:
---

# xx-universal-review edits through couch --session-sync apply; retire deletions: propose-strike

## Problem

Slice 4 of the universal-agent-input-interface project (design: `ops:teams/devinfra/projects/universal-agent-input-interface/design.side/10-s4-edit-path.md`, sections 3, 5 and 8). The shared review skill tells agents to edit documents with `couch --session-sync write` and to hand-write proposal markers under `deletions: propose-strike`. S4 gives edits one path, `apply`, with per-edit reasons, and one knob, `presentation: propose`, for proposals.
