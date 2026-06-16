---
id: 000107
status: open
deps: []
github_issue:
target: base-layer-mechanics
created: 2026-06-16
updated: 2026-06-16
estimate_hours:
---

# weave: target-isolated compile — `compile --target T` clears every other target's artifacts before producing T's

## Problem

`weave compile --target T` should leave a repo holding EXACTLY one target's
artifacts. Today it doesn't — switching targets STRANDS the previous target's
artifacts. Verified on ariadne (24 live `.claude/skills/` links):

```
weave compile --target claude   → 24 .claude/skills/<name> symlinks + prose-only AGENTS.md
weave compile --target codex    → AGENTS.md WITH the `## Skills` menu, and the 24
                                   .claude/skills links LEFT UNTOUCHED (no prune)
```

After `claude → codex` the repo carries BOTH skill faces: the menu in AGENTS.md
AND the 24 stale `.claude/skills/` symlinks.

**Root cause — the prune is target-myopic.** `plan/prune.go`'s scan set is
`ManagedLocations(actions)` = "dirs weave produced a symlink into THIS run." A
backend the current target stops emitting (codex produces nothing under
`.claude/skills`) isn't a managed location → `ScanManagedSymlinks` never looks
there → the claude-era links are orphaned-but-unscanned. The prune can only GC
locations the CURRENT target actively writes into.

**Aggravators:** the stale links are gitignored (`EnsureGitignore` owns
`/.claude/skills/`) → invisible to `git status`; the staleness compounds (a
renamed/removed skill leaves a dangling link until you switch back); and a
menu-session reader that scans `.claude/skills` would see every skill twice.

## Spec — target-isolated lowering

Treat the compile as a pure function over (base-layer source, target):

```
Compile(C, T) → A_T
    C   = the repo's construct / base-layer source (committed, in the repo)
    T   = target ∈ { claude, codex, agy, … }
    A_T = T's target-SPECIFIC (backend-exclusive) artifacts
```

**Invariant.** `weave compile --target T` must, in order:

```
1.  remove   ⋃_{T′ ≠ T}  A_T′      (clear every other target's specific artifacts)
2.  produce  A_T                    (lower the selected target)
```

so after any compile the repo holds **exactly `A_T`** plus the
target-INDEPENDENT shared base (prose AGENTS.md body, settings merge, generic
symlinks, scaffolds, seeds) — which is in NO `A_T` and is always present.

- **Across all base layers.** The cleanup is over the invoking repo's FULL lowered
  surface: an `A_T′` artifact may have been contributed by ANY layer in the DAG
  (e.g. an ancestor's skill lowered into the leaf's `.claude/skills`), so removal
  must cover the whole compiled surface, not just the leaf's own contributions.
- **Today's backends.** `A_claude` = `.claude/skills/<name>` symlinks;
  `A_codex` = `A_agy` = the `## Skills` section IN AGENTS.md. The menu artifact
  SELF-CLEANS (AGENTS.md is a full `WriteFile`, regenerated every compile, so
  switching away from the menu erases it for free); the symlink artifact is
  SEPARATE FILES that only vanish when scanned. So the concrete gap today is a
  single missing action: **a menu-target compile must prune `.claude/skills`.**
- **The leftovers already pass the prune's safety criteria** (symlink + not in the
  run's produced-set + target points into a source root ⇒ safe to remove). The
  only missing piece is SCANNING the non-selected backends' locations — not new
  safety logic.

### Design questions (resolve in the plan)
- **Where does the `{T → A_T exclusive locations}` registry come from** without
  breaking `prune.go`'s deliberate *derive-from-produced-actions, never hardcode*
  principle? Smallest fit: a `Target.ExclusiveLoweredLocations()` (scanned-but-not-
  produced) folded into the managed-scan set; the existing criteria still gate what
  is actually removed. (Alt: compute what the OTHER backends WOULD produce for
  these layers and union their locations.)
- **Don't over-prune.** A hand-authored real `.claude/skills/<x>` dir or a
  non-weave symlink must never be removed — the criteria already protect them
  (real dir ⇒ not a candidate; symlink not pointing into a source root ⇒ KEPT).
  Pin with a test.
- **`make weave` stays `--target claude`** — this is correctness when targets ARE
  switched, not a change to the default flow.

## Done when

- After `weave compile --target X` on a repo previously woven for backend Y, ONLY
  `A_X` remains; every `A_Y` (Y ≠ X) is gone. Verified BOTH directions on a real
  repo (claude→codex prunes `.claude/skills`; codex→claude already clean).
- A test pins it: a fixture woven claude (has `.claude/skills/*`), re-woven codex,
  asserts `.claude/skills` pruned AND the menu present in AGENTS.md.
- The prune's safety invariants are unchanged (no real file/dir or non-weave
  symlink ever removed); `weave golden` / `verify-complete` stay clean.
- `make weave` (claude) behavior is unchanged.

## Out of scope — the next step

**Multi-repo / multi-agent coexistence.** How a GROUP of repos worked by DIFFERENT
agents that compile for DIFFERENT targets coexist safely (e.g. one agent on the
claude target, another on codex, across sibling repos or shared state) — without
one agent's compile clobbering another's expected artifacts. This issue is the
prerequisite: **single-repo target isolation**, the foundation that keeps each
repo's base-layer artifacts clean for ONE target at a time. The concurrency /
per-agent-target model builds on it.

## Plan

- [ ]

## Log

### 2026-06-16
- Filed from an operator question on #104's multi-target lowering: "running
  `--target claude` then `--target codex` — won't the non-conflicting lowered
  artifacts (`.claude/skills`, …) be left on disk, so a codex session reading
  `.claude/skills` sees duplicated skills?" Confirmed by dry-run (codex plan on
  ariadne: writefile AGENTS.md WITH the menu, ZERO `.claude/skills` prune).
- Generalized (operator framing) from the skill-specific bug to the systematic
  invariant: target-isolated lowering — `Compile(C,T)` removes ⋃ A_T′ before
  producing A_T — with skills (`.claude/skills` vs the AGENTS.md menu) as the
  concrete, verified instance. `.claude/settings.json` is NOT affected (the `merge`
  intent is target-independent — re-merged identically by every target). The
  multi-repo/multi-agent-different-targets coexistence is the explicit next step.
