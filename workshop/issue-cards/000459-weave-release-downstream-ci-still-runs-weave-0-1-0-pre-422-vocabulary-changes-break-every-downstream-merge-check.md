---
id: 000459
status: open
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
github_issue:
---

# weave release: downstream CI still runs weave 0.1.0 (pre-#422); vocabulary changes break every downstream merge-check

## Problem

Downstream repos' merge-checks install the published weave release 0.1.0, which predates ariadne#422 (ship and sidecar ownership). On 2026-10-19 ariadne#394 changed construct/vocabulary/issue.cue; weave 0.1.0 then treated parley's regenerated construct/generated/vocabulary/issue.json as an "authored replacement" and failed every parley PR's merge-check (parley.nvim#396 regenerates to unblock). Any later vocabulary or schema change will trip the old released compiler the same way, in every downstream repo whose CI pins the release.
