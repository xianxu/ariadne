---
id: 000457
status: open
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
collaboration-mode: collaboration
card_mirror: '84de1b28e8737c3b8c7fb6747a32f868a5074bbe' # card fields mirrored from issue-cards; edit via sdlc
---

# Known-flake support for non-Go suites: spec-level registry names and rerun

## Problem

The known-flake registry (#394) only understands Go tests. A registry name must be `<import path>.<TestName>`, and the targeted rerun is `go test -count=1 -v -run '^(…)$' pkgs`. Suites that are not Go have no way in:

- parley#394: `tests/integration/document_registry_spec.lua` (busted spec, the queue/deadline case) failed under parallel suite load in recorded suite cb7588a68bab3b4b10fa36c1e134718c on d6e05403 and passed alone right afterwards.
- pair#685: an nvim route-poke admission test in `test.lua` times out under load.

Neither can be registered, so their intermittent failures block every recorded suite and merge until fixed. The 2026-10-19 pilot found this: of the hub's five candidates, two were Lua specs.

## Spec

Extend the registry and rerun to non-Go suites without weakening the rule. The rule still requires same-head fail→pass evidence, an exact name, one owning issue, and that a rerun covers only the registered tests and must pass.

- A name form for a spec-level test: suite kind, spec file, and case. It must be validated the way Go names are, so it is safe as an argv element and never passed through a shell.
- Attribution: the suite log parser names failing non-Go cases (busted, the nvim harness), so `Unattributed` stays empty only when every failure is named.
- A per-kind rerun command, declared by the repo (for example in its local-check declaration), not hardcoded per repo in sdlc.
- `flake list` and `flake register` show and accept the new names.

## Done when

- A registered Lua spec case that fails alone on a head is rerun alone, and a pass lands with the issue named. A failure the parser cannot attribute still blocks.
- The name form refuses shell metacharacters and paths outside the repo, with tests.
- A parley or pair recorded suite exercises it end to end against a fake suite runner.

## Plan

- [ ]

## Log

### 2026-10-19

Filed from ariadne:2 after the #394 pilot, at the hub's request. It sits beside #451 (CI failures not covered).
