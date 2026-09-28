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


### 2026-09-27 — Fleet adoption findings (from #255 Phase B)

Found while running `weave compile` across the fleet for the #252 tracker
cutover (#255). Every failure below reproduces with the pre-#252 `weave`.

- **Symlinked root Makefiles.** metis, kaggle, kbench, you-decide, nous,
  42shots and astro commit `Makefile -> ../ariadne/Makefile`. Since #239 the
  root Makefile is seed-once and ariadne's own builds ariadne's tools, so the
  startup `make tools` through the symlink builds `./cmd/sdlc` in a repo
  without it and fails every dependent's compile. `startup.Tools` checks
  `IsRegular` via `Stat`, which follows the link; `Lstat` would refuse clearly.
- **Pre-inventory generated output.** Repos last compiled before ownership
  inventories (metis Jul 29, parli Sep 19) have `construct/generated/` but no
  `ownership.json`, so compile refuses `vocabulary/.source-sha` as an "authored
  replacement". Nothing adopts such output; neither compile nor
  `propagate-base` will retire what it cannot prove it owns.
- **A refused compile is not atomic.** It had already rewritten `.gitignore`
  (managed block replacing the old list), exposing ignored generated files as
  untracked; hand-restored in metis and parli.
- `propagate-base` is all-dependents-or-nothing and skips dirty trees, so it
  cannot adopt one repository at a time.

**Manual adoption procedure** (proved on metis 43a885a, kaggle b9f5b6e, kbench
77deb07, foundation first; committed locally, not pushed):
1. Replace a symlinked Makefile with the seed (`construct/Makefile.seed`) plus
   the repo's `WF_ISSUES_DIR`/`WF_HISTORY_DIR` and a `tools` target building
   its dependents' commands (metis: `metis`; kaggle: `kaggle`,
   `kaggle-download`, `kaggle-submit` — replacing kbench's hand `go build`).
2. Move the pre-inventory `construct/generated/` aside; `weave compile`
   rewrites it and records ownership (metis: 96 outputs).
3. Untrack exactly the tracked paths git now ignores whose link target matches
   the ownership record (25 per repo; kbench's force-tracked run data and
   `AGENTS.local.md` stay), `git rm` the two links into files ariadne retired
   (`construct/scripts/bootstrap-peers.sh`, `scripts/issue-sync.sh`), and take
   the new seeded `bootstrap.sh` and `merge-check.yml`.
4. Verify: a second compile is a no-op, `weave verify-complete` clean, the
   repo's own checks pass (kaggle `make go-check`).

Candidate fixes: an explicit adopt mode (step 2–3 with the ownership proof),
per-repository `propagate-base`, `Lstat` in `startup.Tools`, and a compile
that stages its `.gitignore` edit with the outputs it covers.
Remaining: you-decide, nous, 42shots, astro (symlink + pre-inventory), parli
(pre-inventory), tools, xianxu.dev (work in flight); pair and ducks compile.

### 2026-09-27 — Fleet adoption progress (manual, per operator: one-off)

Operator decision: adopt by hand, one repository at a time, after a branch
cleanup in each; no adopt mode needed for a one-off.

- Adopted (committed locally, unpushed): metis 43a885a, kaggle b9f5b6e,
  kbench 77deb07, you-decide 8b5085a, nous 8bac006 (`tools` builds `nous` for
  the brain repos; `README-NOUS.md` is nous's own manifest link), 42shots
  43e4b72, astro 30440c8, parli 1ab4bee. Each: 25–26 owned links untracked,
  dead links into retired ariadne files removed, second compile a no-op.
- New finding (parli): weave's "add the workflow include to an existing
  Makefile" does not recognise the older indirect form
  (`-include $(WF_WORKFLOW)` with a `../ariadne` fallback) and prepends a
  second include, above the repo's `WF_*` settings: `make` then warns about
  redefined targets and the next compile would prepend again. Fixed by hand by
  moving parli to the seed's form.
- Homebrew now refuses untrusted third-party taps: nous's Brewfile needs
  `brew trust mutagen-io/mutagen` (operator approved) before compile runs.
- Deferred, work in flight on `:0`: tools (#80, #81 working; local main holds
  their unpublished designs) and xianxu.dev (#4 working, close review
  verdict unknown on 2026-09-19). Adopt after they land.
- Not yet checked: pair and ducks compile, but may still track old links.

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

### 2026-09-25

- #250 added an interim consumer fallback to the seeded merge-check workflow:
  when `brew tap xianxu/ariadne` fails it builds weave from ariadne source. When
  this issue publishes the tap, delete that `elif` branch (and its
  `portable-ci.test.sh` tap-unpublished case) in the same change.
  #250's close review also noted the step name "Build candidate gateway for
  ariadne source CI" is stale while the fallback exists; it becomes accurate
  again once the branch is deleted.

