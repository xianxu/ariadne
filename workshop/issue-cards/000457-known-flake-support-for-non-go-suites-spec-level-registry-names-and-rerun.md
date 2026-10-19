---
id: 000457
status: open
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
github_issue:
---

# Known-flake support for non-Go suites: spec-level registry names and rerun

## Problem

The known-flake registry (#394) only understands Go tests. A registry name must be `<import path>.<TestName>`, and the targeted rerun is `go test -count=1 -v -run '^(…)$' pkgs`. Suites that are not Go have no way in:

- parley#394: `tests/integration/document_registry_spec.lua` (busted spec, the queue/deadline case) failed under parallel suite load in recorded suite cb7588a68bab3b4b10fa36c1e134718c on d6e05403 and passed alone right afterwards.
- pair#685: an nvim route-poke admission test in `test.lua` times out under load.

Neither can be registered, so their intermittent failures block every recorded suite and merge until fixed. The 2026-10-19 pilot found this: of the hub's five candidates, two were Lua specs.
