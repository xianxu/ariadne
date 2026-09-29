---
id: 000264
status: open
deps: []
github_issue:
target: base-layer-mechanics
created: 2026-09-28
updated: 2026-09-28
estimate_hours:
card_mirror: '4aa7ef59a348cd7f7430b4514c168a8aa3f37ab4' # card fields mirrored from issue-cards; edit via sdlc
---

# weave compile: a refused compile must not rewrite .gitignore

## Problem

`weave compile` reconciles two ownership scopes in sequence
(`cmd/weave/main.go`): data mounts per owner (`ApplyManaged(…, ScopeData)`),
then the leaf's artifacts (`ApplyManaged(…, ScopeArtifacts)`). Each call
rewrites the root `.gitignore` via `managedIgnoreText`, which also *migrates*:
it strips every line of the historical fixed list
(`GeneratedRuntimeGitignoreEntries`) and writes a managed block holding only
the outputs recorded so far.

When the artifacts pass then refuses (e.g. `preserving authored replacement`
on pre-inventory `construct/generated/vocabulary/.source-sha`), the data pass
has already published a `.gitignore` whose legacy list is gone and whose block
covers only `/construct/generated/weave/`. Every generated file (AGENTS.md,
CLAUDE.md, skill links, `.colima/`, …) surfaces as untracked. The operator sees
a failed compile plus an unexplained pile of "new" files.

Observed 2026-09-28 in parley.nvim (and earlier in metis, parli, xianxu.dev per
ariadne#241's log, each hand-restored). Named in #241's "Candidate fixes" as "a
compile that stages its `.gitignore` edit with the outputs it covers"; split
out here so #241 stays about publication.

## Spec

A compile that does not reach a successful artifacts apply leaves `.gitignore`
byte-identical. Legacy-list migration happens only in the same step that
records the outputs replacing it — the ignore edit is part of the artifacts
publication, not of each scope's pass. Candidate shape: the data scope
contributes entries but does not write `.gitignore`; the artifacts pass writes
it once, after its preflight has cleared.

## Done when

- A regression test: pre-inventory repo with a conflicting generated file →
  `weave compile` fails with the existing refusal AND `.gitignore` is unchanged
  (fails on current main).
- Successful compiles still produce the same managed block and legacy
  migration (existing ownership/managed-ignore tests pass).

## Plan

- [ ]

## Log

### 2026-09-28

Filed from the parley.nvim investigation: `ownership.json` held
`"outputs": []` (written by the data pass) while the artifacts pass had refused
`vocabulary/.source-sha`; the `.gitignore` diff showed the legacy list removed
and a one-entry block.
