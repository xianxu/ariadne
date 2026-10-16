---
id: 000407
status: open
created: 2026-10-15
updated: 2026-10-15
estimate_hours:
github_issue:
---

# A curated Go lint catalog: golangci-lint + shared ruleguard rules for known mistakes (per-test builds, sleeps, wall-clock asserts, temp dirs, strict wire decode)

## Problem

Recurring mechanical mistakes are caught by humans after they cost something: 11 pair tests `go build` the binary
per run (pair#602), sleeps and wall-clock assertions behind most of 10-15's flakes, temp dirs outside TMPDIR (parley
#342, ariadne#342), strict decoding of cross-version wire messages (pair#594). Neither pair nor ariadne runs a Go
linter; the only lint is ariadne's hand-written `pkg/shelllint`, and lessons files keep rules as unenforced prose.
Operator (2026-10-15): "we need to curate a list of 'stupid mistakes'". Part of ops project test-economy.
