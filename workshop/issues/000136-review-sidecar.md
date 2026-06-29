---
id: 000136
status: working
deps: []
github_issue:
created: 2026-06-26
updated: 2026-06-29
estimate_hours:
started: 2026-06-29T15:42:51-07:00
---

# sdlc boundary review sidecar

## Problem

`sdlc` boundary reviews can produce more output than an interactive agent can
reliably keep in scrollback. During pair#81 retro, pair#72 boundary-review
output was effectively a transient terminal artifact: the close/milestone gate
could make a decision, but the agent did not have a stable file to reopen for
details after truncation or context compaction.

Boundary reviews are workflow evidence. They should be persisted as first-class
sidecar artifacts alongside the issue/plan record instead of existing only as
TTY output and commit trailers.

## Spec

`sdlc` boundary-review paths should write the full review transcript to a
durable sidecar file under `workshop/plans/`, following the existing durable
planning artifact convention.

- Milestone review sidecars use
  `workshop/plans/NNNNNN-slug-m2-review.md` for milestone `M2`.
- Issue-close review sidecars use an equally predictable name, such as
  `workshop/plans/NNNNNN-slug-close-review.md`.
- If a review is re-run for the same boundary, the command must not silently
  destroy prior evidence. Either append a timestamped revision section or choose
  a deterministic collision-safe suffix; the chosen behavior must be documented.
- The sidecar includes enough metadata to orient a fresh reader:
  issue id/title, repo identity, issue file path, boundary kind, milestone when
  present, base/head SHAs, command invoked, reviewer agent/model if known,
  timestamp, verdict, and the full review body.
- Terminal output from `sdlc close`, `sdlc milestone-close`, and the underlying
  boundary-review judge should print a compact verdict/summary plus the sidecar
  path, rather than relying on the terminal as the only durable surface for the
  full review.
- Existing behavior remains intact: `Review-Verdict:` trailers, close gates,
  issue log expectations, and failure behavior continue to work.
- The sidecar is intended for agents to read after the gate runs, including
  after scrollback loss, context compaction, or a follow-up session.

## Done when

- Boundary review output is persisted under `workshop/plans/` for both
  milestone close and full issue close.
- The naming convention covers milestone and issue-close boundaries and handles
  re-runs without silent overwrite.
- The terminal prints the verdict and sidecar path in a compact form.
- Tests cover path naming, required metadata, preserved review body, and
  existing verdict/trailer behavior.
- Help or atlas documentation tells agents where to find the sidecar.

## Plan

- [ ] Locate the close/milestone-close boundary review dispatch and output path.
- [ ] Define sidecar naming, re-run semantics, and metadata fields.
- [ ] Persist the full review body atomically under `workshop/plans/`.
- [ ] Adjust CLI output to show compact verdict plus sidecar path.
- [ ] Add regression tests for milestone close, issue close, and trailer/gate preservation.
- [ ] Update help or atlas docs with the sidecar convention.

## Log

### 2026-06-26

- Created from pair#81 retro point 6: boundary reviews need a durable sidecar
  file, e.g. `workshop/plans/NNNNNN-slug-m2-review.md`, so agents can read the
  full review after the gate runs.
