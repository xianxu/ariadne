---
gate: plan-quality
issue: 241
id_prefix: PQ
rounds:
    - "n": 1
      timestamp: "2026-09-28T18:58:17-07:00"
      agent: claude
      findings:
        - id: PQ-1
          severity: Important
          title: Plan ignores Homebrew's untrusted-tap refusal (noted in its own Log) before removing the source fallback
          detail: 'The Log records that Homebrew now refuses untrusted third-party taps, and nous needed brew trust. xianxu/ariadne is third-party, so a plain brew install in a clean env or CI may fail. Removing the #250 elif then breaks every consumer CI. Verify trust behaviour in a clean environment and put any required brew trust step in bootstrap.sh, the seeded workflow or the docs before deleting the fallback.'
          family: external-dependency-behavior-unverified
          round: 1
        - id: PQ-2
          severity: Important
          title: Docs step names 3 targets, but README.md and atlas/ hold about 11 until-241 references (ARCH-PURPOSE)
          detail: README.md:35/135/162 and atlas/workflow/base-layer.md:9/108/132 are missed by the named list. Sweep every 241 and "until published" reference in README.md and atlas/ so the docs Done-when holds.
          family: class-sweep-not-enumerated
          round: 1
        - id: PQ-3
          severity: Minor
          title: Plan builds release archives locally even though weave-release.yml already prepares a candidate from the tag
          detail: State which build's archives and SHA256SUMS get uploaded (ARCH-DRY), so the published binaries are traceable to one build.
          family: build-of-record-unstated
          round: 1
        - id: PQ-4
          severity: Minor
          title: Linux consumer CI check offers nous, whose Linux bootstrap the atlas records as unverified
          detail: setup-and-replication.md:223 notes launchd plus an amd64-only Mutagen formula for nous. Prefer parley.nvim, or define what counts as passing.
          family: verification-target-ambiguous
          round: 1
        - id: PQ-5
          severity: Minor
          title: 'No non-goals: adopt mode, per-repo propagate-base, Lstat and atomic .gitignore fixes from the Log are silently excluded'
          family: missing-non-goals
          round: 1
        - id: PQ-6
          severity: Minor
          title: State that the branch must not be rebased or amended after tagging, so the tagged SHA still reaches main
          family: release-provenance-invariant
          round: 1
      blocked: true
    - "n": 2
      timestamp: "2026-09-28T19:00:19-07:00"
      agent: claude
      dispose:
        - id: PQ-1
          disposition: addressed
          note: Explicit-name install analysis (bootstrap.sh:13,15), clean-store verification gate, brew trust contingency, README docs.
          round: 2
        - id: PQ-2
          disposition: addressed
          note: Sweep enumerated plus grep-at-close rule; setup-and-replication.md:223 (nous, now a non-goal) is caught by that rule.
          round: 2
        - id: PQ-3
          disposition: addressed
          note: prepare-weave-release artifact at the tag is the build of record; no local upload.
          round: 2
        - id: PQ-4
          disposition: addressed
          note: parley.nvim chosen with an explicit pass criterion; nous excluded.
          round: 2
        - id: PQ-5
          disposition: addressed
          round: 2
        - id: PQ-6
          disposition: addressed
          round: 2
      findings:
        - id: PQ-7
          severity: Minor
          title: prepare-weave-release dispatch omits the required tag input (-f tag=weave-v0.1.0)
          detail: weave-release.yml:7-10 declares inputs.tag as required, so passing only --ref weave-v0.1.0 fails the dispatch. Name the full gh workflow run command.
          family: verification-target-ambiguous
          round: 2
      blocked: false
    - "n": 3
      timestamp: "2026-09-28T19:01:30-07:00"
      agent: claude
      dispose:
        - id: PQ-7
          disposition: addressed
          note: M2 now names the full command with -f tag=weave-v0.1.0, matching weave-release.yml:7-10.
          round: 3
      blocked: false
content_hash: 2567e9ee044bdd6e1bc343fe439e854562b5689b7a3a1ac3ee6d9d911a9e5e61
---

# Gate ledger — ariadne#241 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-09-28T18:58:17-07:00 (claude) — BLOCKED

### Raised

- **PQ-1** [Important] `external-dependency-behavior-unverified` Plan ignores Homebrew's untrusted-tap refusal (noted in its own Log) before removing the source fallback
  The Log records that Homebrew now refuses untrusted third-party taps, and nous needed brew trust. xianxu/ariadne is third-party, so a plain brew install in a clean env or CI may fail. Removing the #250 elif then breaks every consumer CI. Verify trust behaviour in a clean environment and put any required brew trust step in bootstrap.sh, the seeded workflow or the docs before deleting the fallback.
- **PQ-2** [Important] `class-sweep-not-enumerated` Docs step names 3 targets, but README.md and atlas/ hold about 11 until-241 references (ARCH-PURPOSE)
  README.md:35/135/162 and atlas/workflow/base-layer.md:9/108/132 are missed by the named list. Sweep every 241 and "until published" reference in README.md and atlas/ so the docs Done-when holds.
- **PQ-3** [Minor] `build-of-record-unstated` Plan builds release archives locally even though weave-release.yml already prepares a candidate from the tag
  State which build's archives and SHA256SUMS get uploaded (ARCH-DRY), so the published binaries are traceable to one build.
- **PQ-4** [Minor] `verification-target-ambiguous` Linux consumer CI check offers nous, whose Linux bootstrap the atlas records as unverified
  setup-and-replication.md:223 notes launchd plus an amd64-only Mutagen formula for nous. Prefer parley.nvim, or define what counts as passing.
- **PQ-5** [Minor] `missing-non-goals` No non-goals: adopt mode, per-repo propagate-base, Lstat and atomic .gitignore fixes from the Log are silently excluded
- **PQ-6** [Minor] `release-provenance-invariant` State that the branch must not be rebased or amended after tagging, so the tagged SHA still reaches main

## Round 2 — 2026-09-28T19:00:19-07:00 (claude) — passed

### Disposed

- PQ-1 — addressed — Explicit-name install analysis (bootstrap.sh:13,15), clean-store verification gate, brew trust contingency, README docs.
- PQ-2 — addressed — Sweep enumerated plus grep-at-close rule; setup-and-replication.md:223 (nous, now a non-goal) is caught by that rule.
- PQ-3 — addressed — prepare-weave-release artifact at the tag is the build of record; no local upload.
- PQ-4 — addressed — parley.nvim chosen with an explicit pass criterion; nous excluded.
- PQ-5 — addressed
- PQ-6 — addressed

### Raised

- **PQ-7** [Minor] `verification-target-ambiguous` prepare-weave-release dispatch omits the required tag input (-f tag=weave-v0.1.0)
  weave-release.yml:7-10 declares inputs.tag as required, so passing only --ref weave-v0.1.0 fails the dispatch. Name the full gh workflow run command.

## Round 3 — 2026-09-28T19:01:30-07:00 (claude) — passed

### Disposed

- PQ-7 — addressed — M2 now names the full command with -f tag=weave-v0.1.0, matching weave-release.yml:7-10.

## Open findings

(none — every finding has been disposed)
