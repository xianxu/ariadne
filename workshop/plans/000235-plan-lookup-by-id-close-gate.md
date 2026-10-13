---
gate: boundary-review
issue: 235
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-12T18:00:27-07:00"
      dispatched: "2026-10-12T17:48:36-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Minor
          title: Canonical plan name is built twice (planArtifactPathIn and planNameNote)
          detail: changecode.go:609 and :644 both build filepath.Join(plansDir, name+"-plan.md"); a canonicalPlanPath helper would stop the note and the lookup drifting apart (ARCH-DRY).
          family: single-source-derivation
          round: 1
        - id: BR-2
          severity: Minor
          title: Plan lookup globs a path that may contain glob metacharacters
          detail: filepath.Glob over searchDir treats [ * ? in the repo root or plans dir as wildcards, so an off-slug plan quietly reads as absent there; os.ReadDir plus a prefix match would avoid it.
          family: path-as-glob-pattern
          round: 1
        - id: BR-3
          severity: Minor
          title: No test that close's measure or review snapshot refuses two plans
          detail: closeflow.go:80 and close.go:1603 now pass the ambiguity error up, but no fixture reaches them with two plans; only the shared helpers are tested.
          family: refusal-path-coverage
          round: 1
      recipe: milestone-review
      isolation: fenced
      reviewed: 1ffc38208e773a21271f8fdf4fd043c4eac152d4
      blocked: false
    - "n": 2
      timestamp: "2026-10-12T18:04:07-07:00"
      dispatched: "2026-10-12T18:02:46-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: canonicalPlanPath (changecode.go) is now the single builder; planArtifactPathIn and planNameNote both call it.
          round: 2
        - id: BR-2
          disposition: addressed
          note: planArtifactPathIn uses os.ReadDir + filepath.Match on base names; TestPlanLookupToleratesGlobMetacharacters (repo[1]*?) passes and would miss under Glob.
          round: 2
        - id: BR-3
          disposition: addressed
          note: TestReviewPlanPathsRefusesTwoPlans covers the seam close.go:1603 reads through; closeflow.go:80 dies on readOptionalPlanFile's refusal, which TestReadOptionalPlanFileRefusesAmbiguity pins.
          round: 2
      findings:
        - id: BR-4
          severity: Minor
          title: Glob-over-joined-directory rule fixed at one site; 11 sibling filepath.Glob(filepath.Join(dir, ...)) sites remain
          detail: '2nd finding in this family. The rule (now in workshop/lessons.md) is: every directory-scoped pattern lookup goes through one ReadDir+Match helper. Prevalence: 11 non-test sites, including familyFiles at resolve.go:235, which this diff touched. Fix the rule, not another instance: extract the helper from planArtifactPathIn, route all sites through it, and add a source-scan test that rejects filepath.Glob(filepath.Join(. File it as a follow-up; it is outside #235''s scope.'
          family: path-as-glob-pattern
          round: 2
      recipe: milestone-review
      isolation: fenced
      reviewed: ab683fcc8743e0185df90f8b49ada6815f700c96
      blocked: false
---

# Gate ledger — ariadne#235 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-12T18:00:27-07:00 (claude) — passed

### Raised

- **BR-1** [Minor] `single-source-derivation` Canonical plan name is built twice (planArtifactPathIn and planNameNote)
  changecode.go:609 and :644 both build filepath.Join(plansDir, name+"-plan.md"); a canonicalPlanPath helper would stop the note and the lookup drifting apart (ARCH-DRY).
- **BR-2** [Minor] `path-as-glob-pattern` Plan lookup globs a path that may contain glob metacharacters
  filepath.Glob over searchDir treats [ * ? in the repo root or plans dir as wildcards, so an off-slug plan quietly reads as absent there; os.ReadDir plus a prefix match would avoid it.
- **BR-3** [Minor] `refusal-path-coverage` No test that close's measure or review snapshot refuses two plans
  closeflow.go:80 and close.go:1603 now pass the ambiguity error up, but no fixture reaches them with two plans; only the shared helpers are tested.

## Round 2 — 2026-10-12T18:04:07-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — canonicalPlanPath (changecode.go) is now the single builder; planArtifactPathIn and planNameNote both call it.
- BR-2 — addressed — planArtifactPathIn uses os.ReadDir + filepath.Match on base names; TestPlanLookupToleratesGlobMetacharacters (repo[1]*?) passes and would miss under Glob.
- BR-3 — addressed — TestReviewPlanPathsRefusesTwoPlans covers the seam close.go:1603 reads through; closeflow.go:80 dies on readOptionalPlanFile's refusal, which TestReadOptionalPlanFileRefusesAmbiguity pins.

### Raised

- **BR-4** [Minor] `path-as-glob-pattern` Glob-over-joined-directory rule fixed at one site; 11 sibling filepath.Glob(filepath.Join(dir, ...)) sites remain
  2nd finding in this family. The rule (now in workshop/lessons.md) is: every directory-scoped pattern lookup goes through one ReadDir+Match helper. Prevalence: 11 non-test sites, including familyFiles at resolve.go:235, which this diff touched. Fix the rule, not another instance: extract the helper from planArtifactPathIn, route all sites through it, and add a source-scan test that rejects filepath.Glob(filepath.Join(. File it as a follow-up; it is outside #235's scope.

## Open findings

- **BR-4** [Minor] `path-as-glob-pattern` Glob-over-joined-directory rule fixed at one site; 11 sibling filepath.Glob(filepath.Join(dir, ...)) sites remain
