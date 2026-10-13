---
gate: boundary-review
issue: 345
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-13T14:14:22-07:00"
      dispatched: "2026-10-13T14:08:54-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Important
          title: README update appears missing for the new `sdlc local-check run` verb and .sdlc/local-checks.json
          detail: README.md:269-276 covers sdlc pr / sdlc merge landing but says nothing about the new user-typed verb or that a repo can declare local substitutes; atlas and help are updated. Add a sentence plus a link to atlas/workflow/ci-merge-check.md.
          family: docs-gate-readme-surface
          round: 1
        - id: BR-2
          severity: Minor
          title: decideCIGate accepts any check named sdlc-local/<x> as a recorded pass, not only a commit status
          detail: cigate.go:85-89 keys recorded[] on c.Name, so an Actions job or third-party check run with that name satisfies a substitute or required check; the local-gates guard on .github/workflows limits this. Require Workflow=="" and a non-Actions link (ARCH-SECURE).
          family: evidence-provenance-by-name
          round: 1
        - id: BR-3
          severity: Minor
          title: local-check --rev accepts any commit contained in any remote branch, not the tip of the current branch on origin as planned
          family: plan-code-drift
          round: 1
        - id: BR-4
          severity: Minor
          title: runCIGate runs inline local checks before the cheap required-check presence test (ciMissing)
          detail: A missing local-gates-guard/verdict refuses only after up to timeout_minutes of local runs; check required presence before running inline.
          family: gate-check-ordering
          round: 1
        - id: BR-5
          severity: Minor
          title: atlas/workflow/sdlc-binary.md merge row gained a stray trailing empty cell
          family: doc-table-format
          round: 1
        - id: BR-6
          severity: Minor
          title: e2eGH and landingFakeGH each add delegation boilerplate for the new seam methods; landingFakeGH derefs a nil ci
          family: test-fake-duplication
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#345 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-13T14:14:22-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `docs-gate-readme-surface` README update appears missing for the new `sdlc local-check run` verb and .sdlc/local-checks.json
  README.md:269-276 covers sdlc pr / sdlc merge landing but says nothing about the new user-typed verb or that a repo can declare local substitutes; atlas and help are updated. Add a sentence plus a link to atlas/workflow/ci-merge-check.md.
- **BR-2** [Minor] `evidence-provenance-by-name` decideCIGate accepts any check named sdlc-local/<x> as a recorded pass, not only a commit status
  cigate.go:85-89 keys recorded[] on c.Name, so an Actions job or third-party check run with that name satisfies a substitute or required check; the local-gates guard on .github/workflows limits this. Require Workflow=="" and a non-Actions link (ARCH-SECURE).
- **BR-3** [Minor] `plan-code-drift` local-check --rev accepts any commit contained in any remote branch, not the tip of the current branch on origin as planned
- **BR-4** [Minor] `gate-check-ordering` runCIGate runs inline local checks before the cheap required-check presence test (ciMissing)
  A missing local-gates-guard/verdict refuses only after up to timeout_minutes of local runs; check required presence before running inline.
- **BR-5** [Minor] `doc-table-format` atlas/workflow/sdlc-binary.md merge row gained a stray trailing empty cell
- **BR-6** [Minor] `test-fake-duplication` e2eGH and landingFakeGH each add delegation boilerplate for the new seam methods; landingFakeGH derefs a nil ci

## Open findings

- **BR-1** [Important] `docs-gate-readme-surface` README update appears missing for the new `sdlc local-check run` verb and .sdlc/local-checks.json
- **BR-2** [Minor] `evidence-provenance-by-name` decideCIGate accepts any check named sdlc-local/<x> as a recorded pass, not only a commit status
- **BR-3** [Minor] `plan-code-drift` local-check --rev accepts any commit contained in any remote branch, not the tip of the current branch on origin as planned
- **BR-4** [Minor] `gate-check-ordering` runCIGate runs inline local checks before the cheap required-check presence test (ciMissing)
- **BR-5** [Minor] `doc-table-format` atlas/workflow/sdlc-binary.md merge row gained a stray trailing empty cell
- **BR-6** [Minor] `test-fake-duplication` e2eGH and landingFakeGH each add delegation boilerplate for the new seam methods; landingFakeGH derefs a nil ci
