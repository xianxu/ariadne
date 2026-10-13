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

---

## Re-review — 2026-10-12T18:04:07-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 235 — Durable-plan lookup is exact-name: a plan under another slug is invisible |
| repo | ariadne |
| issue file | workshop/issues/000235-plan-lookup-by-id.md |
| boundary | whole-issue close |
| milestone | — |
| window | ad739caf8086d16cafcb67f2912bead482a1218e..ab683fcc8743e0185df90f8b49ada6815f700c96 |
| command | sdlc close --issue 235 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-12T18:04:07-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

**Verdict: SHIP.** #235 does what it set out to do. `planArtifactPath` now finds the durable plan by issue id, using `familyPattern` plus `classifyFamily`'s `-plan.md` rule. The decision is a pure `pickPlanArtifact`, and every reader the Plan listed now goes through this one lookup: change-code's plan read and design checkpoint, `recordChangeCodeFlow`, close's measure and review snapshot, the review manifest, and `judge plan-quality`. When two plans share an id, the error is passed up to every caller, so an ambiguous plan can never be read as "no plan". All three prior Minors are fixed. All seven #235 tests pass and `go vet` is clean. The targeted run also hit one unrelated failure (`TestArchiveDoneIssuesInDir_SweepsPlanArtifacts`): the test reports "no issue tracker is reachable", it is outside the diff, and it is an environment failure in this checkout. I'm raising one new Minor finding, the second in the `path-as-glob-pattern` family. It does not block the boundary.

1. **Strengths**
   - `cmd/sdlc/changecode.go:601-149` (range): one lookup (`planArtifactPathIn`), one pure decision (`pickPlanArtifact`) and one canonical-name helper (`canonicalPlanPath`). Each fact has one owner (ARCH-DRY, ARCH-PURE).
   - Sidecars are excluded by the same `classifyFamily` rule that `sdlc resolve` uses, not by a new suffix check. `familyPattern` is shared with `familyFiles` (`resolve.go:250`).
   - The rule that a lookup must refuse ambiguity rather than pick a plan holds everywhere. `readOptionalPlanFile` now returns `(string, error)`, and every caller dies or returns on that error (`changecode_flow.go:107`, `closeflow.go:80`, `close.go:1603`, `reviewwindow.go:81`, `judge.go:292`).
   - `TestChangeCodeFindsOffSlugPlan` reproduces pair#283 end to end. It checks that the full flow is inferred, the canonical-name note is printed, the plan reaches the plan-quality prompt and the `judge plan-quality` dry-run, and the design checkpoint commits the plan.
   - `TestPlanLookupToleratesGlobMetacharacters` uses a directory named `repo[1]*?`, where `[1]` really does fail to match under `filepath.Glob`. That makes it a genuine regression test for BR-2.

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings**
   - **Second finding in family `path-as-glob-pattern`.** BR-2 was fixed only at the site it named. The diff also adds the general rule to `workshop/lessons.md:54` ("Do not `filepath.Glob` a path joined from a directory you did not write"). But 11 non-test sites in `cmd/sdlc` still do exactly that. One is `familyFiles` (`resolve.go:235`), which this diff touches and the Spec names as the thing to reuse. Others are `actualsplit.go:37`, `issuefiles.go:43,148`, `push.go:320,577`, `setstatus.go:209`, `changecode.go:572`, `issue.go:150` and `project.go:135`.
     - **The rule:** every "files matching a pattern in directory D" lookup should go through one helper, for example `matchIn(dir, pattern)` built on `os.ReadDir` plus `filepath.Match` against the base name. `planArtifactPathIn` already has the body.
     - **How to enforce it:** a source-scan test that rejects `filepath.Glob(filepath.Join(`.
     - **Scope:** out of #235's scope, so file it as a follow-up issue. Fixing one more site would just repeat the instance-only fix.

5. **Test coverage notes**
   - Every Done-when item is pinned:
     - the canonical name and the off-slug name
     - two plans refused, naming both
     - sidecars and a longer id (`0002830-`) ignored
     - the pair#283 end-to-end run
     - the review manifest finding an off-slug plan and refusing two
   - The death path at `closeflow.go:80` is tested only through `readOptionalPlanFile`'s refusal. It is a one-line `die`, so that is acceptable.
   - I did not run a revert check for BR-3. Undoing the fix changes `reviewPlanPaths`' return type, so the test would no longer compile, which is evidence enough.

6. **Architectural notes**
   - ARCH-DRY: pass. The canonical name, family pattern and classification each have one source.
   - ARCH-PURE: pass. `pickPlanArtifact` and `planNameNote` are pure; the directory IO is confined to `planArtifactPathIn`.
   - ARCH-PURPOSE: pass. All the readers the Plan listed are rerouted, and the shadow sweep found no exact-name readers left. The one gap is the glob-safety rule above, which applies across the codebase, not just to this lookup.
   - ARCH-MOCK: N/A. The only external interaction is listing a local directory. The existing tests use a temp directory and the tracker fake.
   - ARCH-CONSTRAINTS: pass. One `ReadDir` of the plans directory per read, a few hundred entries at most.
   - ARCH-SECURE: pass. Filenames are matched, never interpreted. An ambiguous set fails closed, with a visible error.
   - ARCH-ORDER: N/A. Nothing is held between events: each call re-reads the directory.
   - ARCH-FUNERAL: N/A. Nothing durable is created. The note goes to stderr only, and nothing is renamed automatically.
   - Docs: the atlas (`sdlc-binary.md:1314`) and `helptext/change-code.md` are updated. README has no new surface to cover.

7. **Plan revision recommendations**
   - Optionally, add a one-line Revisions entry. The Plan says the lookup is "a glob of `plansDir`", but since the close round 1 fix it lists the directory with `ReadDir`. The Log already records this, so it is not urgent.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      canonicalPlanPath (changecode.go) is now the single builder; planArtifactPathIn and planNameNote both call it.
  - id: BR-2
    disposition: addressed
    note: |
      planArtifactPathIn uses os.ReadDir + filepath.Match on base names; TestPlanLookupToleratesGlobMetacharacters (repo[1]*?) passes and would miss under Glob.
  - id: BR-3
    disposition: addressed
    note: |
      TestReviewPlanPathsRefusesTwoPlans covers the seam close.go:1603 reads through; closeflow.go:80 dies on readOptionalPlanFile's refusal, which TestReadOptionalPlanFileRefusesAmbiguity pins.
findings:
  - id: new
    severity: Minor
    family: path-as-glob-pattern
    title: |
      Glob-over-joined-directory rule fixed at one site; 11 sibling filepath.Glob(filepath.Join(dir, ...)) sites remain
    detail: |
      2nd finding in this family. The rule (now in workshop/lessons.md) is: every directory-scoped pattern lookup goes through one ReadDir+Match helper. Prevalence: 11 non-test sites, including familyFiles at resolve.go:235, which this diff touched. Fix the rule, not another instance: extract the helper from planArtifactPathIn, route all sites through it, and add a source-scan test that rejects filepath.Glob(filepath.Join(. File it as a follow-up; it is outside #235's scope.
```
