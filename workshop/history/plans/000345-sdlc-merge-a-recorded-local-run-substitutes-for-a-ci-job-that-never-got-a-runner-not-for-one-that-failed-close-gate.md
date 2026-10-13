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
    - "n": 2
      timestamp: "2026-10-13T16:17:20-07:00"
      dispatched: "2026-10-13T16:10:14-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: README.md:278-284 names `sdlc local-check run`, .sdlc/local-checks.json and the merge semantics, and links the atlas anchor; matches runCIGate behavior.
          round: 2
        - id: BR-2
          disposition: addressed
          note: isCommitStatus gates recorded[] (cigate.go:94); reverting it turns both new cigate_test.go provenance cases red.
          round: 2
        - id: BR-3
          disposition: not-addressed
          note: for-each-ref --points-at fix is in localcheck_cmd.go:96, but reverting to `branch -r --contains` keeps TestLocalCheckCmd green; add a pushed-but-not-tip (HEAD~1) refusal case.
          round: 2
        - id: BR-4
          disposition: addressed
          note: Presence check now precedes inline runs (cigate.go:452-457); reverting it fails "a missing required verdict refuses before any local run" with calls [zellij@h].
          round: 2
        - id: BR-5
          disposition: addressed
          note: Stray trailing cell removed from the merge row in atlas/workflow/sdlc-binary.md:42.
          round: 2
        - id: BR-6
          disposition: addressed
          note: e2eGH and landingFakeGH embed one nil-safe ciSeamGH (cigate_fake_test.go:112-138); per-fake delegation and the nil deref are gone.
          round: 2
      recipe: milestone-review
      isolation: fenced
      reviewed: 34e4cff2380a59d267d3051ff2c25038a01d1197
      blocked: false
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

## Round 2 — 2026-10-13T16:17:20-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — README.md:278-284 names `sdlc local-check run`, .sdlc/local-checks.json and the merge semantics, and links the atlas anchor; matches runCIGate behavior.
- BR-2 — addressed — isCommitStatus gates recorded[] (cigate.go:94); reverting it turns both new cigate_test.go provenance cases red.
- BR-3 — not-addressed — for-each-ref --points-at fix is in localcheck_cmd.go:96, but reverting to `branch -r --contains` keeps TestLocalCheckCmd green; add a pushed-but-not-tip (HEAD~1) refusal case.
- BR-4 — addressed — Presence check now precedes inline runs (cigate.go:452-457); reverting it fails "a missing required verdict refuses before any local run" with calls [zellij@h].
- BR-5 — addressed — Stray trailing cell removed from the merge row in atlas/workflow/sdlc-binary.md:42.
- BR-6 — addressed — e2eGH and landingFakeGH embed one nil-safe ciSeamGH (cigate_fake_test.go:112-138); per-fake delegation and the nil deref are gone.

## Open findings

- **BR-3** [Minor] `plan-code-drift` local-check --rev accepts any commit contained in any remote branch, not the tip of the current branch on origin as planned
