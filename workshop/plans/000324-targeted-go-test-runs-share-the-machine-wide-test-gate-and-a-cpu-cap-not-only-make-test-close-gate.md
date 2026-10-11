---
gate: boundary-review
issue: 324
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-10T18:51:22-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Minor
          title: -test.parallel cap is applied even when the gate stands aside (make test bypass, WF_TEST_LOCK=off)
          detail: capParallel runs before enter()'s bypass/off checks (testgate.go:108), contradicting the atlas/package-doc "stands aside"; the HELD_BY helper test asserts only held=false. Impact is small today (2 t.Parallel calls in gated packages).
          family: bypass-path-still-applies-side-effects
          round: 1
        - id: BR-2
          severity: Minor
          title: writeRecord forks git rev-parse on every slot acquire, before PreferRealGit
          detail: An extra (xcrun-shimmed) git fork in a gate meant to curb git forks; the repo name could come from the cwd or be dropped (ARCH-MOCK).
          family: external-call-outside-seam
          round: 1
        - id: BR-3
          severity: Minor
          title: weave's gated packages are hand-listed with no source guard like TestGitDrivingPackagesPreferRealGit
          detail: A new git-driving cmd/weave package would go ungated silently; current coverage matches today's git-spawning packages.
          family: coverage-enforced-by-guard
          round: 1
        - id: BR-4
          severity: Minor
          title: gofmt-only whitespace change in cmd/sdlc/publishgate_test.go rides along in the diff
          family: unrelated-churn
          round: 1
      recipe: milestone-review
      reviewed: 3a8bb9c28e6932b608696ee7476cdba98ed1cbc3
      blocked: false
---

# Gate ledger — ariadne#324 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-10T18:51:22-07:00 (claude) — passed

### Raised

- **BR-1** [Minor] `bypass-path-still-applies-side-effects` -test.parallel cap is applied even when the gate stands aside (make test bypass, WF_TEST_LOCK=off)
  capParallel runs before enter()'s bypass/off checks (testgate.go:108), contradicting the atlas/package-doc "stands aside"; the HELD_BY helper test asserts only held=false. Impact is small today (2 t.Parallel calls in gated packages).
- **BR-2** [Minor] `external-call-outside-seam` writeRecord forks git rev-parse on every slot acquire, before PreferRealGit
  An extra (xcrun-shimmed) git fork in a gate meant to curb git forks; the repo name could come from the cwd or be dropped (ARCH-MOCK).
- **BR-3** [Minor] `coverage-enforced-by-guard` weave's gated packages are hand-listed with no source guard like TestGitDrivingPackagesPreferRealGit
  A new git-driving cmd/weave package would go ungated silently; current coverage matches today's git-spawning packages.
- **BR-4** [Minor] `unrelated-churn` gofmt-only whitespace change in cmd/sdlc/publishgate_test.go rides along in the diff

## Open findings

- **BR-1** [Minor] `bypass-path-still-applies-side-effects` -test.parallel cap is applied even when the gate stands aside (make test bypass, WF_TEST_LOCK=off)
- **BR-2** [Minor] `external-call-outside-seam` writeRecord forks git rev-parse on every slot acquire, before PreferRealGit
- **BR-3** [Minor] `coverage-enforced-by-guard` weave's gated packages are hand-listed with no source guard like TestGitDrivingPackagesPreferRealGit
- **BR-4** [Minor] `unrelated-churn` gofmt-only whitespace change in cmd/sdlc/publishgate_test.go rides along in the diff
