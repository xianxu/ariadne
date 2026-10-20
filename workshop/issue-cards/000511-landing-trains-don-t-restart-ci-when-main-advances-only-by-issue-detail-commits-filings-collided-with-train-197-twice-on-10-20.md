---
id: 000511
status: open
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
github_issue:
---

# landing trains: don't restart CI when main advances only by issue-detail commits (filings collided with train 197 twice on 10-20)

## Problem

`sdlc issue new` and `sdlc issue publish` commit issue details straight to main. On 10-20, two such commits (pair#745
and #746 filings) landed while pair train 197 waited on CI; each time the conductor re-integrated main and restarted
CI, delaying a train that had already passed its full suite. Filing and landing collide.
