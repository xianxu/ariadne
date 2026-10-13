# Boundary Review — ariadne#235 (whole-issue close)

| field | value |
|-------|-------|
| issue | 235 — Durable-plan lookup is exact-name: a plan under another slug is invisible |
| repo | ariadne |
| issue file | workshop/issues/000235-plan-lookup-by-id.md |
| boundary | whole-issue close |
| milestone | — |
| window | ad739caf8086d16cafcb67f2912bead482a1218e..1ffc38208e773a21271f8fdf4fd043c4eac152d4 |
| command | sdlc close --issue 235 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-12T18:00:27-07:00 |
| verdict | SHIP |

## Review

Diff read; now running the tests.

```verdict
verdict: SHIP
confidence: high
```

The change does what the issue asked for. `planArtifactPath` now finds the plan by issue id. It globs `familyPattern(id)` (`cmd/sdlc/resolve.go:252`), keeps only files that `classifyFamily` marks `-plan.md`, and passes them to a pure `pickPlanArtifact` (`cmd/sdlc/changecode.go:633`). I traced all six readers to it: change-code step 2, the design checkpoint, `recordChangeCodeFlow`, close's quick-flow measure, `judge plan-quality`, and the review manifest plus close snapshot via `reviewPlanPaths`. No code builds `<stem>-plan.md` on its own any more. Grepping `cmd/` and `pkg/` for `"-plan` finds only the canonical-name construction inside the lookup and the classifier rule. When two plans exist, every reader returns the error or dies on it, so an ambiguous plan never reads as "no plan". `go build ./...` passed. The five new tests and the existing `TestResolveBoundaryReviewManifest_*` tests pass at the pinned head. Nothing blocks the ship; I have three Minor notes only.

1. **Strengths**
   - Sidecar exclusion reuses the rule `sdlc resolve` already has (`classifyFamily`'s `-plan.md` suffix, `resolve.go:146`) rather than writing a new one (ARCH-DRY). The test table covers the `-plan-gate`, `-close-gate`, `-m1-review` and `-close-review` sidecars, a longer id sharing the prefix (`0002830-`) and a neighbouring id.
   - The decision logic is a small pure function with a separate IO step, as the plan described.
   - The error is passed up through readers that used to swallow it: `readOptionalPlanFile` now returns `(string, error)` (`changecode.go:585`), and `closeflow.go:80` and `changecode_flow.go:107` die on it. A missing file still reads as "", as before.
   - `planArtifactPathIn(searchDir, plansDir, …)` separates the directory being searched from the path shown to the user. That fixes the root-relative glob in `reviewPlanPaths` cleanly.
   - `TestChangeCodeFindsOffSlugPlan` uses pair#283's real file names and checks every Done-when item end to end: the flow is inferred as full, the note appears, the plan text reaches the plan-quality prompt and `judge plan-quality --dry-run`, and the design checkpoint commits the off-slug plan.

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings**
   - **Duplicated canonical name (ARCH-DRY):** `filepath.Join(plansDir, name+"-plan.md")` is built twice, at `changecode.go:609` and `changecode.go:644`. A `canonicalPlanPath(plansDir, name)` helper would keep the note and the lookup from drifting apart.
   - **Glob metacharacters:** `filepath.Glob` treats `[`, `*` and `?` in `searchDir` as wildcards. If the repo root or plans path contains one, an off-slug plan quietly reads as absent, which is the old behaviour. A canonically named plan would also be missed, because the lookup still returns the canonical path when the glob finds nothing. Fixing this would need `os.ReadDir` plus a prefix match instead of `Glob`. It's a rare edge, not worth doing now.
   - **Close paths not tested for ambiguity:** nothing checks that `closeFlowStep` dies, or that `captureCloseReviewSnapshot` errors, when two plans exist. Both call the tested helpers, so the risk is low.

5. **Test coverage notes:** The table test checks both properties the plan promised: the result is never a non-`-plan.md` file, and two plans are never resolved to one silently. The manifest test checks the off-slug case and the two-plan refusal. I did not do my own revert-and-rerun mutation check, since this round has no prior fixes to verify. Reverting to an exact-name lookup would obviously fail the "pair#283's off-slug" table row and the manifest test.

6. **Architectural notes**
   - ARCH-DRY: pass, apart from the Minor above. `familyPattern` and `classifyFamily` are shared.
   - ARCH-PURE: pass. `pickPlanArtifact` is pure, and the glob is a thin IO step.
   - ARCH-PURPOSE: pass. All readers go through the lookup. Archive/history lookup and auto-rename are stated non-goals and aren't the point of the issue.
   - ARCH-MOCK: N/A. It's a local filesystem glob with no external dependency. The manifest tests use the existing `fakeBoundaryGit`.
   - ARCH-CONSTRAINTS: pass. It's one glob over a small directory per command call.
   - ARCH-SECURE: pass. Plan file names are files in the repo; ambiguity surfaces as an error rather than a made-up choice.
   - ARCH-ORDER: N/A. Nothing holds state between events.
   - ARCH-FUNERAL: N/A. It creates no files. The `-plan-gate.md` sidecar in this range comes from the existing gate ledger and is archived by `archivePlanArtifacts`.
   - Docs: the atlas (`sdlc-binary.md:1314`) and `helptext/change-code.md` are both updated. There is no new command, flag or config key, so README needs no change.

7. **Plan revision recommendations:** none. The plan matches the code.

```findings
findings:
  - id: new
    severity: Minor
    family: single-source-derivation
    title: |
      Canonical plan name is built twice (planArtifactPathIn and planNameNote)
    detail: |
      changecode.go:609 and :644 both build filepath.Join(plansDir, name+"-plan.md"); a canonicalPlanPath helper would stop the note and the lookup drifting apart (ARCH-DRY).
  - id: new
    severity: Minor
    family: path-as-glob-pattern
    title: |
      Plan lookup globs a path that may contain glob metacharacters
    detail: |
      filepath.Glob over searchDir treats [ * ? in the repo root or plans dir as wildcards, so an off-slug plan quietly reads as absent there; os.ReadDir plus a prefix match would avoid it.
  - id: new
    severity: Minor
    family: refusal-path-coverage
    title: |
      No test that close's measure or review snapshot refuses two plans
    detail: |
      closeflow.go:80 and close.go:1603 now pass the ambiguity error up, but no fixture reaches them with two plans; only the shared helpers are tested.
```
