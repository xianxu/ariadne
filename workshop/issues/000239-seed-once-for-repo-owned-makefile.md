---
id: 000239
status: open
deps: []
github_issue:
created: 2026-09-19
updated: 2026-09-19
estimate_hours:
---

# seed-once for repo-owned Makefile

## Problem

`seed Makefile` silently destroys a pre-existing repo Makefile, on **every**
weave — not just first run.

`applySeed` (`cmd/weave/internal/plan/apply.go:209-241`) has no provenance
check: it reads the upstream source, removes any destination symlink, and if
the bytes differ, overwrites unconditionally. It returns `nil` with no log
line, so there is no warning either.

Before #225 this was survivable — `seed` was write-once, so the clobber was a
one-time adoption event. #225 made `seed` **content-tracking** so derivatives
stranded on a stale `bootstrap.sh` would converge. That was right for
`bootstrap.sh` and swept `Makefile` along with it. Now:

- A repo adopting ariadne loses its build system on first weave.
- A repo that later edits its root Makefile loses that edit on the next weave,
  silently, forever.

**Root cause: one verb, two ownership classes.** `bootstrap.sh` and
`.github/workflows/merge-check.yml` are genuinely upstream-owned and generic —
convergence is correct. `Makefile` is the repo's own front door — convergence
is wrong. Both got `seed`.

### The tell: the seeded root carries repo policy

The 17-line seeded root is not generic. It hardcodes ariadne's own layout:

```make
WF_ISSUES_DIR  = workshop/issues
WF_HISTORY_DIR = workshop/history
```

while `Makefile.workflow:33-34` declares the generic default
`WF_ISSUES_DIR ?= issues`. So ariadne ships its own layout choice to every
derivative through the seed, and a derivative wanting plain `issues/` cannot
say so in the obvious place — the hard `=` runs before the include, making the
overlay's `?=` a no-op, and any edit is clobbered next weave regardless.

A file that encodes per-repo policy and is overwritten by upstream has two
owners. That is the smell.

No test covers this: `construct/scripts/test/portable-makefile.test.sh` (added
by #225) exercises `bootstrap.sh` seeding and the pre-weave owner-resolution
fallbacks, plus "explicit local overlay wins" (`Makefile.local`) — but never a
pre-existing repo-owned root `Makefile`.

## Spec

Split the verb rather than special-casing the file (ARCH-DRY — the policy
belongs to the ownership class, not to one path):

| Verb | Semantics | Rows |
|---|---|---|
| `seed` | content-tracking; converges every weave | `bootstrap.sh`, `.github/workflows/merge-check.yml` |
| `seed-once` | **write-once if absent**; never touched again once present | `Makefile` |

`seed-once` restores pre-#225 semantics for the one file that actually wanted
them, without reverting #225's fix for `bootstrap.sh`.

Consequences:

1. Greenfield repo with no Makefile → still gets a working root for free.
2. Repo with an existing Makefile → keeps it, and adopts ariadne by adding the
   single line `Makefile.workflow` already documents as its contract
   (`Makefile.workflow:1-2`):

   ```make
   # AI issue-based workflow — include from your project Makefile:
   #   include Makefile.workflow
   ```

3. The root `Makefile` becomes **repo-owned and tracked**. Explicitly NOT
   gitignored — it is the repo's own front door; ignoring it would be the same
   two-owners mistake in the other direction.
4. `WF_ISSUES_DIR` / `WF_HISTORY_DIR` move out of the seeded root into each
   repo's own root Makefile, where per-repo policy belongs. `Makefile.workflow`
   keeps the generic `?=` defaults.

### Layering (already exists; only the root is misowned)

- `Makefile.workflow` — ariadne-owned overlay, the includable unit ✓ (symlink)
- `Makefile.local` — repo-owned extensions ✓
- `Makefile` — the wiring root ← **this is the layer to hand back to the repo**

### Naming: keep `Makefile.workflow`

`Makefile.ariadne` was considered — it would match the `settings.<layer>.json`
convention, and generalizes if a mid layer in a 3-deep chain ever needs its own
targets (composable the way settings merge). Rejected for now: the name is
already symlinked fleet-wide and already documents the include contract, so a
rename buys nomenclature at the cost of churn in every repo (Simplicity First).
Revisit only if a mid layer actually needs targets.

## Done when

- A new `seed-once` manifest verb exists, distinct from `seed`: writes when the
  target is absent, no-ops when present (whatever the content), never
  overwrites.
- `Makefile` moved from `seed` to `seed-once` in `construct/base.manifest`;
  `bootstrap.sh` and `.github/workflows/merge-check.yml` stay `seed`.
- A pre-existing repo-owned root `Makefile` survives two consecutive
  `weave compile` runs byte-for-byte — covered by a test in
  `construct/scripts/test/portable-makefile.test.sh`.
- `WF_ISSUES_DIR` / `WF_HISTORY_DIR` no longer ship from the seeded root;
  ariadne's own values live in ariadne's own root Makefile, and a derivative
  can set its own without being clobbered.
- Adoption path documented in `atlas/workflow/base-layer.md`: existing-Makefile
  repos add one `include Makefile.workflow` line.
- Fleet check: derivatives whose root Makefile is currently the seeded copy are
  unaffected (content already identical → `seed-once` no-ops).

## Plan

- [ ] design not yet done — author via `superpowers-writing-plans` at
      `workshop/plans/` after `sdlc claim` + `sdlc start-plan`

## Log

### 2026-09-19

Found while investigating post-`make weave` `git status` churn in pair
(`pair` had 5 dirty weave paths; the fleet showed the same one-time #225
convergence pending in nous + metis).

Investigation notes worth keeping:

- `wf-bootstrap` (`Makefile.workflow:326-331`) runs `weave` **third**, before
  `tools` / `sdlc-install` / `data-deps`. Every symlink referenced without a
  fallback (`scripts/sdlc-install.sh`, `construct/scripts/clone-data-deps.sh`,
  `scripts/parallel-checks.sh`, `scripts/close-issue.py`,
  `scripts/pre-merge-checks.sh`) is consumed strictly *after* weave creates it.
- Everything needed *before* weave already has an owner fallback: `Makefile:12`
  (`$(wildcard Makefile.workflow ../ariadne/Makefile.workflow)`),
  `Makefile.workflow:10` (`wf-helper`), and CI's
  `elif [ -f ../ariadne/scripts/run-merge-checks.sh ]`.
- **Separate follow-up, not this issue:** the justification in
  `cmd/weave/internal/plan/gitignore.go:26-31` for tracking the symlink class
  ("a fresh clone must commit those BEFORE weave can run — the bootstrap
  chicken-and-egg") is **stale**. #225 built the owner-resolution fallbacks
  that dissolved it. The genuinely un-ignorable set is only:
  `.github/workflows/merge-check.yml` (GitHub Actions enumerates workflows from
  the committed tree on GitHub's servers — gitignored means no runner ever
  starts), `bootstrap.sh` (nothing to run on a peerless clone), and
  `construct/deps` (read by bootstrap.sh before any peer exists). Shrinking
  `GeneratedRuntimeGitignoreEntries`' complement to those three would end the
  per-leaf wiring churn — at the cost of the leaf's git history no longer
  recording wiring changes, which wants a `weave --explain` diff or a merge
  check instead. Worth its own issue.
