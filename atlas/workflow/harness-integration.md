# Harness integration — how weave serves many agent harnesses from one checkout

ariadne is harness-agnostic: ONE repo checkout serves Claude Code, OpenAI Codex,
Gemini CLI, … concurrently. This page is the durable model + the **per-harness
assumption ledger** — read it when onboarding a NEW harness or triaging a harness
behavior change. The mechanism is weave (`cmd/weave`); the build is issue #107
(Option B); the math lives in [base-layer-mechanics](../../workshop/targets/base-layer-mechanics.md).

## The model — `Compile(C, T)`, Union, separability

weave compiles the repo's base-layer source `C` into per-harness artifacts:

```
Compile(C, T) → A_T          T ∈ { claude, codex, gemini, … };  A_T = T's harness-specific artifacts
weave compile --target T  =  A_T               # LEAN: one harness's face
weave compile  (default)  =  ⋃_T Compile(C,T)  # UNION: every harness's face at once
```

The Union is what lets one checkout serve every harness. It works because the
`A_T` are **separable — pairwise-disjoint paths**. We ENGINEER that separability;
it is the load-bearing invariant. Everything else (the shared, target-INDEPENDENT
base — settings merge, generic symlinks, scaffolds) is in no `A_T` and is always
present.

Two consequences:
- A harness reads its OWN fixed paths (relative to its CWD). You can't reliably
  redirect a harness to another dir, so there is NO overlay filesystem / separate
  checkout / `--into <dir>` — the Union just puts every face in the repo root.
- A lean `--target T` must REMOVE every other face (`⋃_{T′≠T} A_T′`) — the
  cross-target prune (#107 M3), bidirectional.

## Per-harness face map (Option B)

Each harness gets a **pure-prose entry file** + discovers skills from its **own
skill dir** (no `## Skills` menu — Codex/Gemini auto-compose their own from the
dir; a weave-emitted menu would double-expose).

| Harness | Entry file (prose) | Skill dir | Bodies |
|---|---|---|---|
| Claude Code | `CLAUDE.md` | `.claude/skills/<name>/` | on demand via the dir |
| OpenAI Codex | `AGENTS.md` | `.agents/skills/<name>/` | auto-menu + on demand |
| Gemini CLI | `GEMINI.md` | `.agents/skills/<name>/` | `gemini skills` + on demand |

`.agents/skills` is the **Agent Skills open standard** neutral path
([agentskills.io](https://agentskills.io/specification), Anthropic-originated Dec
2025): `SKILL.md` with YAML frontmatter `name` (≤64, `[a-z0-9-]`, must match the
dir) + `description` (≤1024); optional `license`/`compatibility`/`metadata`/
`allowed-tools`. weave's existing SKILL.md format is conformant, so weave lowers
the SAME skill set into both `.claude/skills` and `.agents/skills` as symlink farms
(reuse `plan.SkillSymlinks` parameterized by dir). The faces are disjoint paths →
the Union is clean.

## Assumption ledger — what we rely on, and the guard

These are the integration assumptions Option B depends on. The **unstable** ones
(the `.agents/skills` scan-path is a per-tool convention, NOT the stable part of
the standard) are runtime-guarded by `scripts/harness-assumptions.test.sh`
(`make harness-check`); Claude's (stable, Anthropic-owned) are doc-asserted.

| Assumption | Verified | Guard |
|---|---|---|
| Claude reads `CLAUDE.md`, NOT `AGENTS.md` (unless `@`-imported) | docs ([memory](https://code.claude.com/docs/en/memory)) | doc-asserted (no CLI hook) |
| Codex discovers `.agents/skills` (real + **symlinked**), weave SKILL.md format | live test (`codex debug prompt-input`) | `make harness-check` |
| Gemini discovers `.agents/skills` (real + **symlinked**), weave SKILL.md format | live test (`gemini skills list --all`) | `make harness-check` |
| Codex + Gemini do NOT read `.claude/` | docs + live test | `make harness-check` |
| Codex auto-composes its OWN menu from `.agents/skills` ⇒ weave must NOT emit a menu | live test | (design constraint) |

**Open:** does Claude ALSO read `.agents/skills` (the neutral alias)? If yes, weave
could drop `.claude/skills` and use only `.agents/skills` for all harnesses. Manual
check; not yet resolved.

## Stability — assume the space MOVES

The standard is young and the scan-paths are per-tool conventions:
- **Gemini** ships `.agents/skills` as **Experimental Preview** (v0.23.0, Jan 2026).
- **Codex** GA'd skills Dec 2025 but had a reported `~/.agents/skills` user-scope
  discovery regression.

So an assumption WILL eventually break. That's why the contract is a runnable test,
not just prose.

## Runbooks

**Onboard a new harness H:**
1. Determine H's entry file + skill dir + SKILL.md format (docs + a fixture probe).
2. Add H's row to the face map + the assumption ledger above.
3. Add H's assertions to `scripts/harness-assumptions.test.sh` (discovery, symlink-
   following, format, isolation from other harnesses' dirs).
4. Teach weave to lower H's face (entry-file prose + skill dir) and the cross-target
   prune to cover it (it already scans the Union's managed locations — no per-target
   registry).
5. `make harness-check` green → propagate.

**Triage a behavior change (`make harness-check` FAILs):**
1. The FAIL names the harness + the broken assumption.
2. Confirm against the harness's current docs/changelog (did the scan-path, format,
   or symlink behavior change?).
3. Update the model here + the weave lowering + the suite to the new reality, OR
   pin the harness to a known-good CLI version until weave catches up.
4. Re-run `make harness-check`; re-propagate.

## Pointers
- Suite: `scripts/harness-assumptions.test.sh` · `make harness-check`
- Build: issue #107 (Option B) + `workshop/plans/000107-option-b-plan.md`
- Skill subsystem: [skill-system](../../workshop/targets/skill-system.md) · weave: [weave.md](weave.md)
