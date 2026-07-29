---
id: 000185
status: done
deps: []
github_issue:
created: 2026-07-17
updated: 2026-07-28
estimate_hours: 0.84
started: 2026-07-28T17:41:54-07:00
actual_hours: 0.22
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

Decision: roadmaps use the same center-of-gravity repo rule as projects. A
roadmap instance lives under the product's center-of-gravity repo at
`workshop/projects/roadmap/<YYYYMM>/<product>.md`. This keeps planning artifacts
with the portfolio surface instead of in brain, while preserving the existing
one-roadmap-per-product-month shape. Parley/project cross-repo discovery is the
navigation model for now; do not add roadmap-specific resolver code until a real
reference workflow needs it (ARCH-PURPOSE, ARCH-DRY).

## Done when

- No roadmap artifact lives in brain; the AGENTS §Peer-Repo brain line drops
  its roadmap residual clause.
- The roadmap datatype states the residency; if a `construct/vocabulary`
  roadmap model exists, its discovery/residency agrees with the datatype.

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: typed-data-prototype design=0.10 impl=0.08
item: atlas-docs design=0.20 impl=0.08
item: cross-cutting-refactor design=0.15 impl=0.08
item: milestone-review design=0.00 impl=0.08
design-buffer: 0.15
total: 0.84
```

Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against
`baseline-v3.1.md`. Method A only. The typed-data design is discounted because
the operator resolved the residency choice before implementation; v3.1 scales
implementation hours to 40% of the v2/v2.1 primitive table.

## Plan

- [x] inventory roadmap artifacts in brain + consumers that read them; migrate
      named files or log the empty-inventory no-op
- [x] decide residency (brainstorm w/ operator) and update the roadmap datatype
- [x] docs sweep + propagate-base

## Log

### 2026-07-28
- 2026-07-28: closed — Updated roadmap residency contract to center-of-gravity coding repo under workshop/projects/roadmap; regenerated ariadne harness docs with make weave; refreshed 42shots/parley.nvim with make weave after temporary stashes and restored their local changes; verified kbench read-only already has new clause; verified no brain roadmap artifacts with find /Users/xianxu/workspace/brain -path '*/data/roadmap/*' -print and rg -n '^type: roadmap\b' /Users/xianxu/workspace/brain -g '*.md'; verified no roadmap vocabulary model with rg --files construct/vocabulary | rg roadmap; ran sdlc issue validate --issue 185, sdlc issue validate --issue 15, make weave-drift-check, git diff --check, and stale-clause rg sweep.; review verdict: FIX-THEN-SHIP

- Claimed, planned, and entered implementation. Operator chose the same
  center-of-gravity repo residency as projects; roadmap instances live under
  `workshop/projects/roadmap/<YYYYMM>/<product>.md` (ARCH-PURPOSE, ARCH-DRY).
- Brain roadmap inventory is currently empty: `find
  /Users/xianxu/workspace/brain -path '*/data/roadmap/*' -print` produced no
  paths, and `rg -n "^type: roadmap\b" /Users/xianxu/workspace/brain -g
  '*.md'` found no matches. Migration is a no-op unless the final sweep finds a
  missed artifact.
- Updated roadmap/product datatype docs, atlas data-artifacts, AGENTS.base, and
  live issue #15. Ran `make weave`; generated `AGENTS.md`, `CLAUDE.md`, and
  `GEMINI.md` now carry the roadmaps-in-coding-repos wording.
- `sdlc propagate-base --ref ariadne#185` re-wove and verified 10 dependents
  with no changes, but refused to touch dirty working trees in `42shots`,
  `parley.nvim`, and `kbench`. Read-only sweep confirmed those three skipped
  repos still carry the old roadmap residual clause in generated harness docs.
  Next action: clean/stash those peer changes, rerun propagation, then close.
- Per operator direction, stashed/restored `42shots` and `parley.nvim` local
  changes around `make weave`. Both repos now verify cleanly for this base-layer
  change and no longer carry the stale roadmap residual clause. Left `kbench`
  unmodified per operator direction; read-only verification shows its generated
  harness docs already carry the new roadmaps-in-coding-repos clause.
- Close review returned FIX-THEN-SHIP. Fixed the roadmap month search recipe to
  distinguish current-repo listing from the proto-company sibling-repo view,
  fixed product link searches to include `workshop/projects/`, and revised the
  durable plan's Core Concepts table with INTEGRATION kind values.

### 2026-07-17

- Filed from #171 M5 boundary review minor: the brain-peer line's roadmap
  residual pointer needed a real issue to point at.
