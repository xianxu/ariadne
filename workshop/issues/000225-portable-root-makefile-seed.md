---
id: 000225
status: working
deps: []
github_issue:
created: 2026-09-13
updated: 2026-09-13
estimate_hours: 1.03
started: 2026-09-13T12:29:20-07:00
---

# Publish a portable root Makefile as a safe seed

## Problem

parley.nvim#208 needs a standalone contributor checkout with a real root
Makefile. The inherited `symlink Makefile` declaration replaces a portable
root on the next weave. A leaf seed cannot counteract it safely: current
`applySeed` follows destination symlinks, potentially overwriting an ancestor.
Fresh clones also lack ignored Makefile.workflow, so bootstrap's handoff to
`make bootstrap` cannot work merely by making that include optional.

## Spec

- Keep the generic root Makefile upstream-owned, published via `seed Makefile`.
  Product-specific targets remain in each consumer's Makefile.local.
- Make workflow inclusion optional for a public checkout, with a sibling
  ariadne overlay fallback after bootstrap has cloned peers. Normal product
  targets and help work without maintainer tools. Existing maintainer targets
  remain available after bootstrap, even when all local helper links are absent.
- Materializing a seed unlinks a destination symlink before writing or chmod,
  including identical-content links; never change its old target's bytes/mode.
  Reuse the safe regular-file materialization pattern already in applyWriteFile.
- Preserve generic CI ownership too: its seeded merge-check workflow discovers
  the bootstrapped upstream runner when the local helper link is absent and
  invokes an optional executable repo-owned scripts/ci-setup.sh before checks.
  Consumer-specific tool provisioning lives in that hook; a repeated weave
  must not erase the effective CI setup. Parley needs Go/CUE/vocabulary for
  its generated-runtime-data drift check.
- No local restoration script or copy-after-weave workaround. Migration is
  idempotent for linked and already-seeded consumers.

## Done when

- Scratch leaf with real Makefile.local has working local targets and help
  without a sibling; bootstrap discovers its overlay after peers are present.
- Existing linked consumer becomes a real seeded Makefile; ancestor bytes and
  permissions remain unchanged. Repeated weave leaves content unchanged.
- Regression cases cover matching, differing, and dangling destination symlinks
  and unavailable/failed materialization paths.
- A parley-like scratch consumer survives maintainer refresh and still runs its
  product tests. Source/manifest/atlas documentation describe seed ownership.

## Plan

- [ ] Design safe seed materialization and portable generic Makefile behavior.
- [ ] Add regressions, implement the seed migration and bootstrap path.
- [ ] Validate representative consumers and close with measured evidence.

## Log

### 2026-09-13

Discovered during fresh-eyes review of parley.nvim#208's deployment plan.
This is a prerequisite for its maintainer-link cleanup, not a runtime package
dependency. No implementation or estimate yet; plan approval comes first.

## Revisions

### 2026-09-13 — CI ownership sweep

Reason: the generic merge-check workflow is also upstream-seeded. Delta: add
runner discovery and a repo-owned setup hook to the portable-maintainer boundary
so the consumer does not fork a workflow that weave would overwrite.

## Estimate

Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against
`baseline-v3.1.md`. Method A only. Calibration flagged stale by estimate-source,
so provisional. Existing OSFS and Make infrastructure eliminate library discovery.
Thorough plan discounts design by 0.2; implementation uses v3.1's 0.4 scaling.
Smaller Go extension (0.2/0.5), two cross-cutting surfaces (each 0.3/0.5),
docs (0.1/0.1), and one review (0.1/0.4), before those multipliers.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: smaller-go-module design=0.04 impl=0.20
item: cross-cutting-refactor design=0.06 impl=0.20
item: cross-cutting-refactor design=0.06 impl=0.20
item: atlas-docs design=0.02 impl=0.04
item: milestone-review design=0.02 impl=0.16
design-buffer: 0.15
total: 1.03
```
