---
id: 000590
status: open
created: 2026-10-24
updated: 2026-10-24
estimate_hours:
github_issue:
---

# Qualify recursive JSON lint against independent oracles and DOM materializers

## Problem

Ariadne pkg/golint json-recursive-scanner reports proven false positives in Pair while adopting 5a17f789619d for pair#838. The matcher treats recursive Token + More + any string-key map as duplicate-validator authority. It flags both an independent differential-test oracle and a standalone DOM materializer. Its advice names cmd/sdlc/internal/jsonstrict, which downstream Go modules cannot import.
