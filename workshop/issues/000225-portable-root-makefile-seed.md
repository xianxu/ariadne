---
id: 000225
status: open
deps: []
github_issue:
created: 2026-09-13
updated: 2026-09-13
estimate_hours:
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
