# Option B — per-harness skill-dir lowering — Implementation Plan (#107)

**Goal:** Lower skills to per-harness skill DIRS — `.claude/skills` (Claude) +
`.agents/skills` (Codex + Gemini, the Agent Skills neutral path) — retire the
`AGENTS.md` `## Skills` menu, and make per-harness entry files pure prose. One
shared checkout serves every harness because the `A_T` are separable (disjoint
paths). `weave compile` = `⋃_T Compile(C,T)` (Union, default); `--target T` = the
lean subset. Full rationale + the doc/test verification: issue 000107.

**Architecture:** `Compile(C,T) → A_T` is the primitive (ARCH-PURE: the lowering is
a pure derivation from the selected skill set; IO confined to the apply seam). The
two new skill faces (`.claude/skills`, `.agents/skills`) are TWO renderings of the
SAME `skill.SelectVisible` set — no second discovery (ARCH-DRY, reuse `plan.SkillSymlinks`
parameterized by dir). The menu composer + the `IncludeSkillMenu`/`EmitSkillSymlinks`
target split go away.

**Why a test suite first (M1 before M2):** the `.agents/skills` SCAN-PATH is a
per-tool convention, not the stable part of the standard (Gemini "Experimental
Preview" v0.23.0; Codex had a `~/.agents/skills` regression). M2 RELIES on these
assumptions. M1 codifies them as a runnable guard that (a) confirms they hold now
and (b) fails loudly when a harness changes — so a future weave/harness break is
caught at the integration boundary, not in production. (feedback: model external
services for E2E — the deliverable is the integration + a faithful contract test.)

---

## M1 — per-harness integration assumption test suite + atlas page

**Outcome:** `scripts/harness-assumptions.test.sh` asserts ariadne's per-harness
skill-discovery contract against the INSTALLED CLIs; `atlas/workflow/harness-integration.md`
documents the model + the assumptions (linking the suite).

### Task 1 — `scripts/harness-assumptions.test.sh`
- Build a temp fixture (mktemp; mirror weave's lowering form):
  - `.agents/skills/probe-real/SKILL.md` — weave-format frontmatter (`name`+`description`).
  - `.agents/skills/probe-link → ../../<src>/probe-link` — a RELATIVE symlink to a
    real skill dir (the exact form weave lowers; the load-bearing symlink case).
  - `.claude/skills/probe-claudeonly/SKILL.md` — a skill ONLY under `.claude/skills`
    (to assert Codex/Gemini IGNORE `.claude/`).
  - `CLAUDE.md` (marker `CLAUDE-MARKER`) + `AGENTS.md` (marker `AGENTS-MARKER`).
- Per harness, skip if its CLI is absent (graceful, prints SKIP). Assertions:
  - **gemini** (`gemini skills list --all` in fixture): output CONTAINS `probe-real`
    + `probe-link` (discovery + symlink + format) and does NOT contain
    `probe-claudeonly` (ignores `.claude/`). Also RECORD whether it reads
    `.agents/skills` vs `.gemini/skills`.
  - **codex** (`codex debug prompt-input` in fixture): the rendered prompt CONTAINS
    `probe-real` + `probe-link` and NOT `probe-claudeonly`.
  - **claude**: assert what's deterministically testable + RECORD the rest. Probe
    whether claude discovers `.claude/skills` (its native dir) and whether it ALSO
    reads `.agents/skills` (the neutral alias — open question that affects M2: if
    claude reads `.agents/skills` too, weave may not need `.claude/skills`). The
    `CLAUDE.md`-vs-`AGENTS.md` precedence is doc-verified (Claude reads only
    CLAUDE.md); test at runtime if a deterministic hook exists, else assert-by-doc
    with a clear NOTE.
- Output a per-harness PASS / FAIL / SKIP matrix; exit non-zero on any FAIL.
- ARCH-DRY: one `setup_fixture` + one `assert_contains`/`assert_absent` helper pair
  reused across harnesses (no per-harness copy-paste). Mirror `scripts/docflow.test.sh`
  conventions.
- **Execution home (Finding 4):** add a `make harness-check` target running the suite;
  document the trigger — run pre-propagate (M4) and whenever a harness CLI updates;
  the atlas page (Task 2) carries the triage runbook. CI lacks the CLIs ⇒ it SKIPs
  there by design; the value is the dev / pre-propagate gate, not CI.

### Task 2 — `atlas/workflow/harness-integration.md`
- Document: the `Compile(C,T)`/Union model; the per-harness face map (entry file +
  skill dir per harness); the ASSUMPTIONS table (what each harness reads, verified
  how, doc URL, the test that guards it); the instability note (per-tool scan-path,
  preview/regression history); how to onboard a NEW harness (add its row + a suite
  assertion) and how to triage a behavior change (the suite fails → update model +
  weave). Link from `atlas/index.md`.

### Task 3 — M1 close
- Run the suite (expect PASS on the installed CLIs — assumptions verified 2026-06-16).
- `sdlc milestone-close --issue 107 --milestone M1`.

---

## M2 — Option B PRODUCE side (design; detail at M2 change-code)

- **Skill faces (ARCH-DRY).** Replace the `EmitSkillSymlinks`/`IncludeSkillMenu`
  target split with: lower the selected set to BOTH `.claude/skills/<name>` and
  `.agents/skills/<name>` symlink farms — reuse `plan.SkillSymlinks` parameterized by
  destination dir (one renderer, two dirs; no second discovery). RETIRE `idx.Menu()`
  + the menu composition in the compile path (Codex auto-composes its own from
  `.agents/skills`; an explicit menu double-exposes).
- **Entry files (ARCH-DRY).** Retire the `CLAUDE.md = @AGENTS.md` bridge; FAN the ONE
  `composeAgentsBody` output to three `Dst` — `CLAUDE.md`, `AGENTS.md`, `GEMINI.md` —
  pure prose, no menu, composed ONCE (not three times). (Open: real copies vs
  symlinks for the 2nd/3rd file.)
- **Union vs lean.** `weave compile` (no `--target`) = Union (produce ALL faces).
  `--target T` = produce only T's face. (The REMOVE side is M3.)
- **gitignore.** Add `/.agents/skills/` + `/GEMINI.md` to `GeneratedRuntimeGitignoreEntries`.
- **Tests + golden.** Union produces both skill dirs + 3 prose entry files; the
  `.claude/skills` face is byte-unchanged from today.
- **Targets/atlas:** skill-system + base-layer-mechanics + weave.md.

## M3 — Option B REMOVE side: the cross-target prune (design)

- The original #107 bug: a lean `--target X` compile must prune every NON-selected
  face. **Reuse the Union primitive (ARCH-DRY, Finding 2) — NOT a hardcoded
  registry.** Compute the UNION's actions (what all faces would produce for these
  layers), scan `ManagedLocations(union-actions)`, but keep the PRODUCED-set as the
  lean compile's actual output. `shouldPrune` then removes any face the lean compile
  didn't produce — **bidirectionally** — with ZERO new delete logic and ZERO second
  source of truth (honors prune.go criterion-2: "derived from produced actions, never
  hardcode").
- **Don't over-prune.** The existing criteria already protect real dirs + non-weave
  symlinks; pin with a test (a hand-authored real `.claude/skills/<x>` dir / an
  unrelated symlink survives a `--target codex` compile).
- **Tests.** claude→codex prunes `.claude/skills`; codex→claude prunes
  `.agents/skills`; union prunes neither; safety criteria hold.

## M4 — propagate (sketch)

- Re-weave all 10 ariadne-styled repos onto the new lowering; run `make harness-check`;
  `verify-complete` + ancestors byte-pristine; the brains via the sandbox-safe path.
  Manual loop for now (the first-class flow is [[000106]] propagate-base).

---

## Open questions (carry into the milestone that hits them)
- **Does Claude read `.agents/skills`?** (M1 probes it.) If yes, weave might drop
  `.claude/skills` and use ONLY `.agents/skills` for all harnesses — even simpler.
- **Entry files: real copies vs symlinks** for `AGENTS.md`/`GEMINI.md` → the prose.
- **`make weave` default** = Union (all faces) so a fresh weave serves every harness;
  confirm that's the right default vs `--target claude`.
