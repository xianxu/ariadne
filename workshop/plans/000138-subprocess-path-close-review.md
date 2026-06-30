# Boundary Review — ariadne#138 (whole-issue close)

| field | value |
|-------|-------|
| issue | 138 — sdlc subprocess path resolution |
| repo | ariadne |
| issue file | workshop/issues/000138-subprocess-path.md |
| boundary | whole-issue close |
| milestone | — |
| window | d191427e827df2d7e46f32cb1581e3ae0fdd66e0..HEAD |
| command | sdlc close --issue 138 |
| reviewer | claude |
| timestamp | 2026-06-29T22:37:33-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

I have enough to complete the review. Tests pass, the implementation is clean, and I've cross-checked the plan's Core concepts table against the code, traced the Done-when items, and worked the ARCH lenses. Let me note one more check on the error-path test before writing up.

The key findings: `TestDispatch_LaunchError_Surfaces` (judge_test.go:794) asserts only `err != nil`, not the new diagnostic content; and the plan's Core concepts table documents a `binAugmentedEnv(execPath ...)` signature plus omits `ownerBinDir`, both of which drifted from the shipped code.

VERDICT: FIX-THEN-SHIP (confidence: high)

The change does exactly what #138 asks: it injects the owner `bin/` (derived from `os.Executable()`) onto the subprocess `PATH` at the single `Run` seam, so a fresh review/judge agent resolves `sdlc` without depending on the user's interactive shell startup. The pure/IO split is textbook, the pure helper has a real table test, and `TestMinimalPathResolvesSdlc` is a genuine process-level proof of the narrow-PATH claim. Nothing blocks shipping. Two Important items should be cleaned up before the next boundary: the plan's Core-concepts table now documents a signature and entity set the code no longer matches (it was refined — for the better — into `ownerBinDir` + `binAugmentedEnv(binDir,…)`), and the new launch-failure diagnostic (a stated Done-when) has no test pinning its content.

**1. Strengths**
- `ownerBinDir()` (dispatch.go:41) single-sources "where sibling tools live," consumed by both `Run` and `Dispatch` — exactly the ARCH-DRY consolidation the design called for; no per-caller PATH plumbing. **PASS ARCH-DRY.**
- `binAugmentedEnv` (dispatch.go:53) is genuinely pure (no `os.Executable`, no exec), and its test (dispatch_test.go:12) runs with zero IO — the IO (`os.Executable`, `os.Environ`, exec) stays in the thin `Run` seam. **PASS ARCH-PURE.**
- `TestMinimalPathResolvesSdlc` (dispatch_test.go:56) is the right kind of fixture: a real `sh -c 'command -v sdlc'` under a deliberately narrow base PATH, proving end-to-end resolution rather than re-asserting the implementation.
- Best-effort posture is correct: `Run` leaves `cmd.Env` nil when `ownerBinDir()` errors, preserving today's inherit-parent-env behavior rather than aborting a review over a PATH nicety (D4).

**2. Critical findings**
None.

**3. Important findings**
- **Plan Core-concepts table contradicts the shipped code** — `workshop/plans/000138-subprocess-path-plan.md` documents `binAugmentedEnv(execPath string, env []string)` doing `filepath.Dir(execPath)` internally, and explicitly justifies `execPath` as "a *parameter* … precisely so it's testable." The code instead ships `binAugmentedEnv(binDir string, env []string)` and hoists `filepath.Dir(os.Executable())` into a new `ownerBinDir()` that the table never lists. The refactor is an improvement (it's what enables the DRY single-source), but the plan now claims an API the code doesn't have and omits a real integration-point entity. Fix via a plan `## Revisions` entry (see §7) — the cross-check gate exists to keep the plan from documenting what the code doesn't deliver.
- **New launch-failure diagnostic is untested** (dispatch.go:154; test at judge_test.go:794) — the Done-when "Error output includes the attempted executable path and PATH when resolution fails" is delivered in the error string, but `TestDispatch_LaunchError_Surfaces` only asserts `err != nil`. A regression that dropped the `dispatch %s (PATH includes owner bin %q)` framing would pass silently. Cheap fix: in that test, also assert the error string contains the agent name (`"claude"`) and the `owner bin` substring (`Run` is already stubbed there, so no real spawn needed).

**4. Minor findings**
- Literal Done-when says the failure message should include "the PATH used"; the error includes only the prepended owner-bin *dir* (`%q`), not the full effective PATH. The owner-bin dir is the diagnostically relevant piece, so this is acceptable — but it's a slightly narrower delivery than the words promise; either tighten the wording in the issue or include `PATH` if you want literal parity.
- On a `Dispatch` launch failure where `ownerBinDir()` itself errors, `dir, _ := ownerBinDir()` (dispatch.go:153) yields `owner bin ""` in the message — harmless but mildly confusing; a `"?"` fallback would read better.
- `strings.CutPrefix(e, "PATH=")` is case-sensitive; on Windows `os.Environ()` yields `Path=`, so it would synthesize a duplicate `PATH=` rather than prepend. Out of practical scope (darwin/linux dev tool), noting only for completeness.

**5. Test coverage notes**
- Pure helper: prepend / synthesize / empty-and-`.` no-op all covered (dispatch_test.go:12). Good coverage of the branch matrix.
- End-to-end narrow-PATH resolution: covered by the real-spawn fixture.
- Gaps: (a) the launch-failure diagnostic content (see Important #2); (b) the `Run` one-liner wiring is unasserted — acknowledged in the Log and acceptable, since `os.Executable()` is uncontrollable in-test and the logic is covered by the pure helper + fixture (correct ARCH-PURE reasoning).

**6. Architectural notes for upcoming work**
- ARCH-DRY ✓, ARCH-PURE ✓, ARCH-PURPOSE ✓ (the shadow-sweep: both consumers — every boundary review *and* `sdlc judge* — flow through the single `Run` seam, so PATH injection covers all `sdlc` invocations the agent makes on its own initiative, not just scripted ones; the "no prompt rewrite" reasoning in the Log holds because no prompt instructs the agent to run `sdlc`). The `binAugmentedEnv` "future extension: prepend multiple dirs" note is a clean seam if a later agent needs a peer's `bin/`.

**7. Plan revision recommendations**
- Add a `## Revisions` entry to `workshop/plans/000138-subprocess-path-plan.md` recording: (1) `binAugmentedEnv`'s signature changed from `(execPath string, …)` to `(binDir string, …)`; (2) `filepath.Dir(os.Executable())` was extracted into a new `ownerBinDir()` integration helper (consumed by both `Run` and `Dispatch`), which should be added to the Core concepts "Integration points" table as new; (3) reason: ARCH-DRY single-sourcing of the owner-bin dir so the launch-failure diagnostic and the PATH build share one derivation. Update the `Run`/`Dispatch (error)` line references (`:36`/`:105` → current `:77`/`:149`) while you're there.
