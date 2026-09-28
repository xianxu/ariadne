---
id: '000239'
status: codecomplete
started: 2026-09-19T18:21:15-07:00
created: 2026-09-19
updated: 2026-09-20
estimate_hours: 4.151
actual_hours: 14.84
---

# Minimal committed base-layer surface

## Problem

**A derivative should commit only what it needs to bootstrap. Everything else
`make weave` produces should be gitignored.** Today it commits ~45 weave-created
paths, and the justification for doing so is stale.

Three defects, escalating:

### 1. The tracked-symlink set has no live justification

`cmd/weave/internal/plan/gitignore.go:26-31` explains why the symlink class is
NOT ignored:

> *the pre-weave BOOTSTRAP scaffolding (bootstrap.sh, Makefile, Makefile.workflow,
> construct/scripts/{...}.sh, the construct/\* dir symlinks,
> .claude/settings.ariadne.json). A fresh clone must commit those BEFORE weave
> can run (the bootstrap chicken-and-egg)*

That was true before #225. **#225 built the owner-resolution fallbacks that
dissolved the chicken-and-egg, and the not-ignored list was never shrunk.**
Evidence:

- `wf-bootstrap` (`Makefile.workflow:326-331`) runs `weave` **third**, before
  `tools` / `sdlc-install` / `data-deps`. Every symlink referenced without a
  fallback — `scripts/sdlc-install.sh`, `construct/scripts/clone-data-deps.sh`,
  `scripts/parallel-checks.sh`, `scripts/close-issue.py`,
  `scripts/pre-merge-checks.sh` — is consumed strictly *after* weave creates it.
- Everything needed *before* weave already resolves from its owner:
  `Makefile:12` (`$(wildcard Makefile.workflow ../ariadne/Makefile.workflow)`),
  `Makefile.workflow:10` (`wf-helper`, commented *"Before first weave, helper
  links may be absent"*), and CI's
  `elif [ -f ../ariadne/scripts/run-merge-checks.sh ]`.

Cost of the stale list: the manifest states the wiring once, then ~45 committed
paths restate it in every leaf, drifting on every manifest edit. That churn is
what surfaced this (pair had 5 dirty weave paths; nous + metis have the same
#225 convergence still pending).

### 2. `seed Makefile` silently destroys a repo's own Makefile

`applySeed` (`cmd/weave/internal/plan/apply.go:209-241`) has no provenance
check: read upstream source, remove any destination symlink, and if bytes
differ, overwrite unconditionally. Returns `nil` with no log line.

Pre-#225 `seed` was write-once, so this was a one-time adoption event. #225 made
it **content-tracking** so derivatives stranded on a stale `bootstrap.sh` would
converge — right for `bootstrap.sh`, and it swept `Makefile` along. Now a repo
adopting ariadne loses its build system on first weave, and a repo that later
edits its root Makefile loses that edit on **every** subsequent weave, silently.

**Root cause: one verb, two ownership classes.** `bootstrap.sh` and
`merge-check.yml` are genuinely upstream-owned — convergence is correct.
`Makefile` is the repo's own front door — convergence is wrong.

The tell: the seeded root is not generic. It hardcodes ariadne's own layout —

```make
WF_ISSUES_DIR  = workshop/issues
WF_HISTORY_DIR = workshop/history
```

— while `Makefile.workflow:33-34` declares the generic default
`WF_ISSUES_DIR ?= issues`. A derivative wanting plain `issues/` cannot say so in
the obvious place: the hard `=` runs before the include, making the overlay's
`?=` a no-op, and the edit is clobbered next weave anyway. A file encoding
per-repo policy that upstream overwrites has two owners.

No test covers it — `construct/scripts/test/portable-makefile.test.sh` (#225)
exercises `bootstrap.sh` seeding, the pre-weave fallbacks, and "explicit local
overlay wins" (`Makefile.local`), but never a pre-existing repo-owned root
`Makefile`.

### 3. The gitignore list is a hand-maintained second channel, and append-only

`GeneratedRuntimeGitignoreEntries` is a hardcoded `[]string` of 9 paths. This
target's spine says *"base.manifest is the single source of truth for what a
layer contributes and to whom — no artifact enters the composition by any other
channel."* A hand-maintained list of what weave generates **is** a second
channel, and a hand-maintained restatement of the model is a deferred consumer,
not a finished one (ARCH-PURPOSE, ARCH-DRY).

It is also **append-only**: `ensureGitignoreText` (`gitignore.go:75-92`) appends
absent entries and never removes. Harmless at 9 hardcoded entries; actively
dangerous at ~45 manifest-derived ones, because a retired manifest row leaves a
stale ignore line that can silently untrack a repo-owned file later taking that
path.
