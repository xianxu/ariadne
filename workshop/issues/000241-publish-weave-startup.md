---
id: 000241
status: open
deps: [ariadne#239]
github_issue:
target: base-layer-mechanics
created: 2026-09-20
updated: 2026-09-20
estimate_hours:
---

# Publish weave and cut over startup

## Problem

Weave must be available as the standalone entrypoint for ariadne-style repos.
The startup implementation in #239 needs to merge before a public release and
consumer cutover can be verified. Tracking delivery here avoids making #239's
close depend on a release that itself depends on #239 being merged.

## Spec

User-specified public names: Homebrew tap `xianxu/ariadne`, formula
`xianxu/ariadne/weave`, installed command `weave`.

Publish the tested binaries and Homebrew formula prepared by #239, using a
reviewed merged commit. Proposed backing repo is `xianxu/homebrew-ariadne`;
source and release assets remain in `xianxu/ariadne`. Native macOS and Linux
bootstrap use the same release artifacts. Version/asset/checksum metadata must
come from the release tooling, not be independently maintained here.

Then cut consumers over to the shared startup contract: new repos use install,
`weave link <source>`, and `weave compile`; existing derivative clones use
`./bootstrap.sh`. Pilot parley.nvim and nous, verify real fresh-clone setup and
CI, then migrate remaining applicable consumers. Follow the preparation and
ownership rules in #239's
[restart plan](../plans/000239-minimal-committed-base-layer-surface-plan.md#restart-standalone-weave-startup).
Actual peer changes use peer-owned issues/workflows; brain/data repos retain
their capture/commit rhythm. This issue owns delivery evidence, not a second
startup implementation.

Before publishing, inspect the concrete release artifacts and formula; do not
claim distribution from a successful local build. Obtain the operator's license
choice before adding a project license grant; no license is inferred here.

## Done when

- A versioned weave release with the tested platform binaries and matching
  checksums is publicly available from ariadne.
- `brew tap xianxu/ariadne` and `brew install xianxu/ariadne/weave` work in a
  clean environment, and the installed executable reports the intended version.
- New-repo remote link/compile and existing-derivative bootstrap are exercised
  using the published gateway, without pre-existing layer/toolchain checkouts.
- Pilot consumers complete fresh-clone setup and real CI; remaining applicable
  consumers are migrated with recorded per-repo evidence.
- Generated inherited wiring is untracked only where ownership is proven;
  authored files, unrelated tracked-but-ignored files and local ignore rules
  survive. No production-signed/service binary is replaced by an implicit
  development build.
- Release instructions and consumer setup documentation match the published
  commands and observed behavior.

## Plan

- [ ] After #239 merges, inspect its release candidate, native tests, formula,
      bootstrap version floor and migration tooling; resolve publication metadata.
- [ ] Publish the versioned release from the merged commit and the corresponding
      formula in the tap; verify installation from the public endpoints.
- [ ] Apply the prepared pilot migrations through peer workflows, verify cold
      startup and real CI, then roll out remaining applicable consumers.
- [ ] Record public release/install links and per-consumer evidence, update
      documentation, and close only after delivery is verified.

## Log

### 2026-09-20

Created during #239 restart planning. The operator explicitly allowed publication
in a separate ticket. Fresh-context plan review identified the close/ship cycle;
this dependent delivery issue makes the sequence explicit. #239 prepares/tests
startup and release/migration tools, closes and merges; #241 publishes and
verifies delivery afterward. No release or consumer mutation has started.


## Revisions

### 2026-09-20 — tap repository and command availability confirmed

**Reason:** operator accepted the conventional backing repo and clarified that
shell integration belongs to the user/layer, not generic bootstrap.

**Delta:** `xianxu/homebrew-ariadne` is the agreed tap backing repository, not a
pending option. Distribution exposes the gateway `weave`; layer binaries remain
in their owners' bin directories and users explicitly add those directories to
PATH. Verify that documented step and that bootstrap does not edit shell profiles
or perform sdlc-specific installation. No `weave exec` or `weave env` command is
part of the delivery contract.
