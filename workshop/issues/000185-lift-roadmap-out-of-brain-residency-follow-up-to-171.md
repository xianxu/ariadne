---
id: 000185
status: working
deps: []
github_issue:
created: 2026-07-17
updated: 2026-07-28
estimate_hours:
started: 2026-07-28T17:41:54-07:00
---

# lift roadmap out of brain (residency follow-up to #171)

## Problem

#171 established the residency charter: brain is capture/measurement only and
holds no SDLC process artifacts. Projects lifted to coding-repo
`workshop/projects/`; `roadmap` is the residual SDLC artifact still living in
brain (the AGENTS §Peer-Repo brain line flags it and points here). Same
contradictions #171 named apply: auto-commit sweeps deliberate portfolio
state, and brain's encryption posture couples the sharing decision.

## Spec

Apply #171's model to roadmaps: pick the residency (likely the same
center-of-gravity rule, or a single home repo since a roadmap is inherently
cross-repo), define discovery/navigation if refs point at it, and migrate the
existing record(s). Reuse `DiscoverByIssueRef`-style tooling only if roadmaps
are actually referenced by refs — don't build surface ahead of need
(ARCH-PURPOSE).

## Done when

- No roadmap artifact lives in brain; the AGENTS §Peer-Repo brain line drops
  its roadmap residual clause.
- The roadmap datatype (`construct/vocabulary`/datatype docs) states the
  residency.

## Plan

- [ ] inventory roadmap artifacts in brain + consumers that read them
- [ ] decide residency (brainstorm w/ operator) and migrate
- [ ] docs sweep + propagate-base

## Log

### 2026-07-17

- Filed from #171 M5 boundary review minor: the brain-peer line's "`roadmap`
  remains until it too lifts" pointer needed a real issue to point at.
