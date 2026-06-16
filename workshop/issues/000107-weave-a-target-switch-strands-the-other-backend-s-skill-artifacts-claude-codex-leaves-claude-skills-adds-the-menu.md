---
id: 000107
status: open
deps: []
github_issue:
target: skill-system
created: 2026-06-16
updated: 2026-06-16
estimate_hours:
---

# weave: a target switch strands the other backend's skill artifacts (claude→codex leaves .claude/skills + adds the menu)

## Problem

Skills lower through TWO mutually-exclusive per-harness backends (`plan/target.go`):
- **symlink backend** (`--target claude`): `.claude/skills/<name>` symlinks; AGENTS.md prose-only.
- **menu backend** (`--target codex|agy`): the `## Skills` section composed INTO AGENTS.md; NO symlinks.

A single `weave compile` lowers exactly ONE backend. But **switching targets on a
repo strands the other backend's artifacts** — verified on ariadne (24 live
`.claude/skills/` links):

```
weave compile --target claude   # 24 .claude/skills/<name> symlinks + prose-only AGENTS.md
weave compile --target codex --dry-run   # writefile AGENTS.md (15228 bytes, WITH the menu);
                                          # ZERO .claude/skills actions — no prune, no symlink
```

So after `claude → codex` the repo carries BOTH skill faces: the menu in AGENTS.md
AND the 24 stale `.claude/skills/` symlinks.

**Root cause — the prune is target-myopic.** `plan/prune.go`'s scan set is
`ManagedLocations(actions)` = "dirs weave produced a symlink into THIS run." The
codex run produces no symlink under `.claude/skills` → it isn't a managed location
→ `ScanManagedSymlinks` never looks there → the claude-era links are
orphaned-but-unscanned. The prune can only GC locations the CURRENT target writes
into; a backend the current target stops emitting can't be cleaned.

**Aggravators:**
- The stale links are **gitignored** (`EnsureGitignore` owns `/.claude/skills/`),
  so they're invisible to `git status` — silent cruft.
- It compounds: parked on the menu target, a renamed/removed skill leaves a
  stale/dangling `.claude/skills/` link that nothing cleans until you switch back.
- **Duplication risk** if any reader in a codex/agy session scans `.claude/skills`
  (the design ASSUMES menu-only harnesses don't — true today, fragile tomorrow).

**Asymmetry — only `claude → codex` leaks.** `codex → claude` self-heals: AGENTS.md
is a full `WriteFile` (claude overwrites it prose-only → menu gone), and
`.claude/skills` IS a managed location for claude (orphans pruned + regenerated).
The menu backend's artifact lives INSIDE the regenerated AGENTS.md, so it cleans
itself; the symlink backend's artifacts are separate files that only get pruned
when scanned. Hence the leak is one-directional.

**Latent today:** every repo is woven `--target claude` (`make weave` hardcodes
it); codex/agy aren't exercised yet (#104 noted "only the claude target is ever
exercised"). This bites the moment they are.

## Spec

The invariant: **a `weave compile --target X` must leave EXACTLY backend-X's
exclusive artifacts present and EVERY other backend's exclusive artifacts ABSENT**
— "+X −(every other backend)" — on top of the shared, target-independent file-ops
(prose body, settings merge, generic symlinks, scaffold/touch/seed), which are
always present regardless of target.

Precisely (today there are two backends, but state it generally so a third —
say a future `.cursor/skills` — is covered for free):
- `weave compile` knows the full set of **skill-backend exclusive lowered
  locations** (claude → `.claude/skills/`; menu → the `## Skills` AGENTS.md section).
- It produces the selected backend's artifacts AND prunes every NON-selected
  backend's exclusive artifacts.
- For the symlink backends this means the prune must **scan `.claude/skills`
  regardless of the current target** — on a codex/agy run, all its links are
  orphans (the run produces none there) and they already satisfy the existing
  safety criteria (symlink + not-in-producedSet + target points into a source
  root), so they're safe to remove.
- The menu backend needs no separate prune step — it lives in the regenerated
  AGENTS.md, already overwritten each run.

### Design questions (resolve in the plan)
- **Where does the "all backend locations" set come from** without violating
  `prune.go`'s deliberate *derive-from-produced-actions, never hardcode* principle?
  Options: (a) compute what the OTHER backend(s) WOULD produce for these layers and
  fold their locations into the managed-scan set; (b) a small explicit
  backend-location registry on `Target` (e.g. `Target.ExclusiveLoweredRoots()`),
  scanned-but-not-produced; (c) generalize `ManagedLocations` to "all targets'
  potential lowered locations." Lean (b) — smallest, keeps the safety criteria
  intact (only symlinks pointing into source roots get pruned).
- **Don't over-prune** — the criteria already protect real files/dirs and
  non-weave symlinks; confirm a menu-target run can't nuke a hand-authored
  `.claude/skills/<x>` (it can't: not a symlink into a source root ⇒ KEPT).
- **`make weave` stays claude** — the fix is for correctness when targets ARE
  switched; it shouldn't change the default flow.

## Done when

- After `weave compile --target X` on a repo previously woven for backend Y, ONLY
  backend-X's skill artifacts remain — backend-Y's exclusive artifacts are gone.
  Verified BOTH directions (claude→codex prunes `.claude/skills`; codex→claude
  already clean) on a real repo.
- A test pins it: a fixture woven claude (has `.claude/skills/*`), re-woven codex,
  asserts the `.claude/skills` links are pruned AND the menu is in AGENTS.md.
- The prune's safety invariants are unchanged (no real file/dir or non-weave
  symlink ever removed); `weave golden`/`verify-complete` stay clean.
- `make weave` (claude) behavior is unchanged.

## Plan

- [ ]

## Log

### 2026-06-16
- Filed from an operator question while reviewing #104's multi-target lowering:
  "running `--target claude` then `--target codex` — won't the non-conflicting
  lowered artifacts (`.claude/skills`, …) be left on disk, so a codex session that
  reads `.claude/skills` sees duplicated skills?" Confirmed by dry-run (codex plan
  on ariadne: writefile AGENTS.md WITH menu, ZERO `.claude/skills` prune). Root
  cause: `ManagedLocations` derives the prune-scan set from the CURRENT run's
  produced symlinks, so a backend the target stops emitting goes un-pruned. The
  fix is the "+selected −all-other-backends" invariant (operator's framing). Note
  `.claude/settings.json` is NOT affected — the `merge` intent is
  target-independent, re-merged identically by every target.
