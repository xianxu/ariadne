---
id: 000107
status: working
deps: []
github_issue:
target: base-layer-mechanics
created: 2026-06-16
updated: 2026-06-16
estimate_hours: 12
---

# weave: Option B — per-harness skill-dir lowering (`.claude/skills` + `.agents/skills`), retire the menu, prose entry files

## Decision + scope (2026-06-16) — Option B, decided

After doc + live-test investigation (see Log + "Design discussion" below),
**Option B is chosen:** lower skills to per-harness skill DIRS — `.claude/skills`
(Claude) + `.agents/skills` (Codex + Gemini, the cross-vendor Agent Skills neutral
path) — **retire the `AGENTS.md` `## Skills` menu**, and make the per-harness entry
files pure prose. Formally: `weave compile` = `⋃_T Compile(C,T)` (Union) is the
default, `--target T` = the lean subset. The `A_T` are SEPARABLE (disjoint paths),
so one shared checkout serves every harness — no contested file, no overlay FS, no
separate checkouts.

**Verified (docs + live test, installed CLIs):**
- Claude reads ONLY `CLAUDE.md` (AGENTS.md invisible unless `@`-imported).
- Codex 0.139.0 + Gemini 0.38.2 both discover project `.agents/skills`, PARSE
  weave's exact `name`+`description` SKILL.md, and FOLLOW SYMLINKS (the weave
  lowering form) — tested with a real + a relative-symlinked fixture skill. Codex
  auto-composes its OWN in-prompt menu from `.agents/skills`, so weave must NOT also
  emit a menu (that's the double-exposure path — reinforces "retire the menu").

**The space is NOT stable.** `.agents/skills` is the Agent Skills open standard
(agentskills.io, Anthropic-originated Dec 2025) but the SCAN-PATH is a per-tool
convention: Gemini ships it "Experimental Preview" (v0.23.0, Jan 2026); Codex had a
real `~/.agents/skills` discovery regression. So we COMMIT the assumptions to a
**per-harness integration assumption test suite** (M1) that fails loudly when a
harness changes behavior, plus an atlas page documenting the integration model for
future harness onboarding / behavior-change triage.

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

## Spec

> **Note (post-rescope):** the `Compile(C,T) → A_T` formalism + the Union/lean
> compositions below still hold under Option B. But the "Today's backends"
> paragraph + the "single missing action" describe the PRE-Option-B menu world and
> are SUPERSEDED by the Decision: under Option B the faces are `.claude/skills`
> (Claude) + `.agents/skills` (Codex/Gemini), there is NO menu, and the cross-target
> prune is **bidirectional**. The authoritative acceptance criteria are "## Done
> when" below (rewritten for Option B).

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

- `weave compile` (Union, default) produces BOTH skill faces — `.claude/skills`
  (Claude) + `.agents/skills` (Codex/Gemini) — from the SAME selected set, plus the
  three pure-prose entry files (`CLAUDE.md`/`AGENTS.md`/`GEMINI.md`); NO `## Skills`
  menu anywhere (Codex auto-composes its own from `.agents/skills`).
- `weave compile --target X` produces ONLY backend-X's face and PRUNES every other
  backend's face — **bidirectional** (a claude compile prunes `.agents/skills`; a
  codex/gemini compile prunes `.claude/skills`). The cross-target prune scans the
  UNION's managed locations with the lean target's produced-set (no per-target
  registry — ARCH-DRY); the existing safety criteria gate removal. Verified both
  directions on a real repo, with a test.
- The per-harness assumption suite (M1) PASSES on the installed CLIs and FAILS on a
  broken assumption.
- The prune's safety invariants are unchanged (no real file/dir or non-weave symlink
  ever removed); `weave golden` / `verify-complete` stay clean.
- `make weave` default = Union (serves every harness); the `.claude/skills` face is
  byte-for-byte unchanged from today.

## Design discussion — Compile(C,T) primitive · Union default · per-harness entry files (2026-06-16)

Worked out with the operator while reasoning about the multi-agent case. Keep
`Compile(C, T) → A_T` as the PRIMITIVE; compose it two ways:

```
weave compile --target T   =  A_T               # LEAN / single-harness (this issue's isolation)
weave compile  (default)   =  ⋃_T Compile(C,T)  # ALL faces at once (the multi-agent mode)
```

Both rest on ONE invariant we engineer and defend: **the `A_T` are SEPARABLE
(pairwise-disjoint paths).** Given separability, the union is a clean superset AND
the isolation is a clean subset — same primitive, two compositions. Isolation is
NOT dropped; it's the subset. The union is what multi-agent wants.

**The Union resolves multi-agent coexistence** (the original "next step"). Several
harnesses sharing ONE repo checkout need every face present at once, in the repo
ROOT where each harness reads its OWN fixed paths. So:
- NO overlay filesystem, NO separate checkouts, NO `weave compile --into <dir>`.
  (`--into` only helps if the agent's CWD *is* `<dir>` — harness read-paths are
  fixed relative to CWD and can't be reliably redirected — so it's purely the
  CWD-swap/overlay enabler, which the Union makes unnecessary.)
- macOS has no per-process mount-namespace + OverlayFS anyway; the Union sidesteps
  the whole need. (Overlays/clonefile stay relevant for OTHER isolation, not this.)

**What makes the `A_T` separable — per-harness entry files.** The only contested
path today is `AGENTS.md` (prose vs prose+menu), currently bridged by `CLAUDE.md =
@AGENTS.md` (so Claude sees whatever AGENTS.md holds, menu included). VERIFIED
(claude-code-guide vs docs.claude.com/en/memory): **Claude Code reads ONLY
`CLAUDE.md`; `AGENTS.md` is invisible to it unless `@`-imported** — no merge, no
precedence, no toggle. Codex reads `AGENTS.md`; Gemini reads `GEMINI.md`. So weave
RETIRES the `@AGENTS.md` bridge and emits DISTINCT per-harness constitutions,
making the faces disjoint:

| entry file | content | harness |
|---|---|---|
| `CLAUDE.md` | prose-only (skills via `.claude/skills/`) | Claude Code |
| `AGENTS.md` | prose + `## Skills` menu | Codex |
| `GEMINI.md` | prose + `## Skills` menu | Gemini CLI |
| `.claude/skills/<name>` | symlinks | Claude-only |

Each harness self-selects by the file it reads; the Union coexists in one root with
no contested path. (Parent-dir `CLAUDE.md` walk-up exists but is fine for this layout.)

**`.claude/skills` cross-read — VERIFIED SAFE** (web check vs official docs).
Codex auto-reads `AGENTS.md` + `~/.codex`/`.codex/config.toml`; Gemini auto-reads
the `GEMINI.md` hierarchy + `.gemini/settings.json`. **Neither scans `.claude/` at
all.** So leaving `.claude/skills` on disk in the Union is inert to them — no
double-load. The "do not scan `.claude`" menu line is therefore NOT needed (keep
only as harmless defense-in-depth, if at all).

**…but the `.agents/skills` finding changes the skill backend (the real news).**
Codex AND Gemini both recognize a SHARED neutral skill-discovery dir
**`.agents/skills`** (Codex also `~/.codex`/builtins; Gemini also `.gemini/skills`).
Consequences:
  1. **#104's premise is now stale.** #104 built the AGENTS.md `## Skills` menu
     because "codex/agy have NO skill-discovery dir." They DO — `.agents/skills`.
     So the menu's original justification no longer holds.
  2. **A cleaner backend than the menu:** weave lowers skills to per-harness skill
     DIRS — `.claude/skills` (Claude) + `.agents/skills` (Codex + Gemini) — and
     every harness discovers NATIVELY. The `## Skills` menu RETIRES; all entry files
     become pure prose. Disjoint paths (`.claude/skills` vs `.agents/skills`) keep
     the Union clean, and this MAXIMIZES separability (the only remaining per-harness
     divergence is harness-specific prose, if any). The skill lowering collapses
     from "symlinks XOR menu" to "symlinks, into the right per-harness dir."
  3. **Reverse caveat:** `.agents/skills` is the path that WOULD double-expose — a
     real `.agents/skills/` read by BOTH Codex and Gemini, IF the same skills are
     also in a menu. The dir-only backend (no menu) avoids that by construction.

So the Union's skill face has two candidate backends: KEEP the #104 menu, OR
**lower to `.agents/skills` + retire the menu** (simpler, native for all, updates
the stale premise). The `.agents/skills` route looks strictly cleaner; decide at
build (could be its own issue). Either way `.claude/skills` coexistence is safe.

**Still genuinely open (the real next step).** Artifact COEXISTENCE is solved by
the Union. What remains: concurrent-WRITE coordination — two agents editing the
SAME repo's source `C` at once (a git/locking concern, orthogonal to artifact
materialization) — and, across the DAG, re-propagating a base-layer edit while
agents are live (ties to [[000106]] propagate-base). Separate from this issue.

## Plan

- [ ] M1 — per-harness integration assumption test suite + atlas page. A runnable
      `scripts/harness-assumptions.test.sh` that builds fixtures and asserts, per
      INSTALLED harness, the contract Option B relies on: Claude reads `CLAUDE.md`
      not `AGENTS.md`; Codex + Gemini discover `.agents/skills` (real AND SYMLINKED)
      in weave's `name`+`description` SKILL.md format and ignore `.claude/`. Skips an
      absent CLI; fails loudly on a broken assumption (deterministic probes: `gemini
      skills list --all`, `codex debug prompt-input`). **Execution home (Finding 4):**
      a `make harness-check` target + a documented trigger (run pre-propagate and
      whenever a harness CLI updates; the atlas page carries the triage runbook) so
      the guard has teeth despite CI lacking the CLIs (it SKIPs there, by design).
      Plus `atlas/workflow/harness-integration.md` documenting the `Compile(C,T)`/Union
      model, the per-harness face map, and the assumptions (linking the suite).
- [ ] M2 — Option B PRODUCE side. Lower skills to BOTH `.claude/skills` (Claude) +
      `.agents/skills` (Codex/Gemini) symlink farms from the SAME selected set (reuse
      `plan.SkillSymlinks` parameterized by destination dir — ARCH-DRY); RETIRE the
      `## Skills` menu + `IncludeSkillMenu`/`idx.Menu()` compile plumbing (Codex
      auto-composes its own). Per-harness entry files become pure prose: fan the ONE
      `composeAgentsBody` output to three `Dst` (`CLAUDE.md`/`AGENTS.md`/`GEMINI.md`)
      — retire the `CLAUDE.md=@AGENTS.md` bridge. `weave compile` = Union (all faces),
      `--target T` = only T's face. gitignore `/.agents/skills/` + `/GEMINI.md`.
      Tests + golden.
- [ ] M3 — Option B REMOVE side (the original #107 cross-target prune). Generalize
      the prune to scan `ManagedLocations(UNION-actions)` while the produced-set stays
      the lean compile's — so a `--target X` lean compile prunes every NON-selected
      face, **bidirectionally**, reusing the Union primitive M2 builds (NO per-target
      registry; prune.go's derive-not-hardcode invariant intact; NO new delete logic
      — ARCH-DRY, Finding 2). Tests both directions + the safety criteria (a real dir
      / non-weave symlink is never removed).
- [ ] M4 — propagate: re-weave all ariadne-styled repos onto the new lowering; run
      `make harness-check`; `verify-complete` + ancestors byte-pristine (ties to
      [[000106]] — manual loop for now).

## Revisions

- **2026-06-16 (plan-quality gate).** The Option B rescope surfaced 5 findings;
  reconciled: (1) `## Done when` rewritten to Option B (no menu; **bidirectional**
  prune; both skill faces) — `## Spec`'s "Today's backends" marked the pre-Option-B
  baseline. (2) The cross-target prune uses `ManagedLocations(union-actions)` with
  the lean produced-set — NOT a `Target.ExclusiveLoweredLocations()` registry —
  honoring prune.go's derive-not-hardcode and reusing the Union primitive (ARCH-DRY).
  (3) Split the overloaded weave milestone into M2 (PRODUCE: faces + menu-retire +
  prose entry files + union) and M3 (REMOVE: the cross-target prune); propagate → M4.
  (4) Gave the assumption suite an execution home (`make harness-check` + trigger).
  (5) Estimate 8h → 12h.

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
- 2026-06-16: design discussion (operator) — resolved the multi-agent coexistence
  via the **Union framing** (see "Design discussion" above). Keep `Compile(C,T)`
  the primitive; `weave compile` = ⋃ A_T (all faces), `--target T` = lean A_T;
  separability of the A_T (engineered via per-harness entry files
  CLAUDE.md/AGENTS.md/GEMINI.md) is the load-bearing invariant. Verified Claude
  reads only CLAUDE.md (AGENTS.md invisible unless @-imported) → retire the
  CLAUDE.md=@AGENTS.md bridge. Dropped `--into <dir>` (only helps CWD-swap/overlay;
  the Union needs none).
- 2026-06-16: VERIFIED (web/docs) Codex + Gemini do NOT read `.claude/` → the Union
  is safe, no double-load; the "do not scan .claude" belt is unnecessary. Bigger
  find: both recognize a shared neutral skill dir `.agents/skills`, which (a)
  invalidates #104's "codex/agy have no skill-discovery dir" premise and (b) opens a
  cleaner backend — lower skills to `.claude/skills` (Claude) + `.agents/skills`
  (Codex+Gemini), retire the AGENTS.md menu, all entry files become pure prose. The
  path to watch for double-exposure is now `.agents/skills` (read by both), not
  `.claude/skills` (inert to them).
- 2026-06-16: investigated `.agents/skills` from BOTH docs and a live fixture test
  (operator-directed). DOCS: it's the Agent Skills open standard (agentskills.io,
  Anthropic Dec 2025); `name`(≤64, dir-matched)+`description`(≤1024) required,
  optional `license`/`compatibility`/`metadata`/`allowed-tools`; the scan-PATH is a
  per-tool convention — Gemini "Experimental Preview" v0.23.0 (Jan 2026), Codex GA'd
  Dec 2025 but with a reported `~/.agents/skills` discovery regression. TEST (fixture
  = a real + a relative-symlinked skill in weave's SKILL.md format): **Gemini 0.38.2**
  `gemini skills list --all` discovered BOTH (symlink followed); **Codex 0.139.0**
  `codex debug prompt-input` injected BOTH into its auto-built `<skills_instructions>`
  menu (symlink resolved). Format parity + symlink-following + native discovery all
  PASS on installed versions. Codex builds its OWN menu from the dir ⇒ weave must not
  also emit one.
- 2026-06-16: **DECISION — Option B** (operator). Rescoped this issue: M1 the
  per-harness assumption test suite (codify the above + guard against the unstable
  space) + the atlas integration page; M2 the weave Option B lowering (per-harness
  skill dirs, retire the menu, prose entry files, Union default / `--target` lean,
  cross-target prune); M3 propagate. estimate 8h.
