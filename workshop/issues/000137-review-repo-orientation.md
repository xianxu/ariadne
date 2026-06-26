---
id: 000137
status: open
deps: []
github_issue:
created: 2026-06-26
updated: 2026-06-26
estimate_hours:
---

# sdlc boundary review repo orientation

## Problem

Boundary review prompts can misorient a fresh reviewer about which repository is
under review. During pair#81 retro, review context for pair work was observed to
refer to an `ariadne#...` issue shape, even though the operating repository was
`pair`.

This is risky because the boundary reviewer is intentionally fresh-context. If
the prompt names or implies the wrong repo, the reviewer can inspect the wrong
tracker, apply ariadne base-repo assumptions to a downstream repo, or report
findings against the wrong issue surface.

## Spec

Tighten `sdlc` boundary-review orientation so every close and milestone review
prompt identifies the current repo and issue owner explicitly and accurately.

- The prompt must include the repository under review, derived from the current
  git root/remote/cwd context rather than a hardcoded `ariadne` label.
- The prompt must include enough concrete anchors for a fresh reviewer:
  repo slug/name, repo root path, issue reference such as `pair#72`, issue file
  path, boundary kind, milestone when present, base SHA, and head SHA.
- The prompt must distinguish base-repo work from downstream/peer repo work. A
  reviewer operating in `pair` should be told that `pair` is the reviewed repo;
  ariadne should appear only when it is actually the reviewed repo or explicitly
  relevant as a dependency.
- Existing boundary-review behavior remains intact: verdict format, trailers,
  gates, and side effects should not change except for clearer orientation.
- This should apply consistently to `sdlc close`, `sdlc milestone-close`, and
  the underlying boundary-review judge/prompt construction path.

## Done when

- Boundary review prompts name the current reviewed repo accurately for ariadne
  and for a downstream repo fixture.
- Prompt text includes repo root, issue file, issue reference, base/head SHAs,
  boundary kind, and milestone when applicable.
- Tests fail if prompt construction falls back to a hardcoded `ariadne#N`
  reference for non-ariadne repos.
- Existing verdict/trailer/gate tests continue to pass.
- Help, atlas, or prompt comments document the repo-orientation contract.

## Plan

- [ ] Locate boundary-review prompt construction for close and milestone-close.
- [ ] Define a single repo-orientation data structure shared by boundary prompts.
- [ ] Derive repo slug/name from the active git root/remote/cwd context.
- [ ] Render explicit issue and repo anchors in all boundary-review prompts.
- [ ] Add tests for ariadne and downstream-repo prompt orientation.
- [ ] Update documentation or inline prompt comments for the orientation contract.

## Log

### 2026-06-26

- Created from pair#81 retro point 7: boundary review instructions need to
  orient the reviewer to the repo being operated on, especially for downstream
  repos like `pair`.
