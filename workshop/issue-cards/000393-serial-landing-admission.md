---
id: 000393
status: open
created: 2026-10-15
updated: 2026-10-15
estimate_hours:
github_issue:
---

# Serial per-repo landing admission with fleet-blocking priority

## Problem

Concurrent `sdlc merge` calls integrate against the same main; after one lands, the others reintegrate and rerun a roughly nine-minute suite plus queue wait. #376 required three integrations. A fleet-blocking fix (#391) also waited fourth for a suite with no priority lane. #390's operator-approved design chooses serial, per-repo landing admission before any batching.
