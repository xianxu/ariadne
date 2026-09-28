---
gate: boundary-review
issue: 257
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-09-27T23:39:56-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Minor
          title: ReadCutoverMarkerAt treats any cat-file -e failure (bad ref, git error) as "marker absent"
          detail: cutover.go:104 swallows the error; ReadCutoverMarker distinguishes ErrNotExist. Fails closed in checkCutover but Presence's stale branch would answer unmarked. Same pattern pre-exists at issuelintids.go:231. Verify commit^{commit} first, then treat only a missing path as absent.
          family: absent-vs-error-conflation
          round: 1
        - id: BR-2
          severity: Minor
          title: Marker presence probed twice for lint-ids, with a third commit-tree reader in issuemigrate
          detail: issuelintids.go:231 cat-file -e duplicates ReadCutoverMarkerAt's probe; issuemigrate.go:132-137 is a parallel commit-tree marker reader via mainView. Route lint-ids' mode decision through tracker.ReadCutoverMarkerAt (ARCH-DRY).
          family: single-marker-reader
          round: 1
        - id: BR-3
          severity: Minor
          title: Spec says lint-ids without --head reads the checkout; it reads HEAD's committed tree
          detail: --head defaults to HEAD and cardlessAdditions always calls GuardCutoverAt, so a working-tree-only marker edit is not seen. Consistent with line 231, but the Spec sentence should say so.
          family: doc-claim-drift
          round: 1
        - id: BR-4
          severity: Minor
          title: checkCutover's markerAt-absent refusal (cutover.go:130) has no test and no reaching caller
          detail: lint-ids checks marker presence at head before guarding, so this branch is unreachable today; a tracker-package unit test would pin it for future GuardCutoverAt callers.
          family: unexercised-branch
          round: 1
      recipe: small-diff-review
      blocked: false
---

# Gate ledger — ariadne#257 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-09-27T23:39:56-07:00 (claude) — passed

### Raised

- **BR-1** [Minor] `absent-vs-error-conflation` ReadCutoverMarkerAt treats any cat-file -e failure (bad ref, git error) as "marker absent"
  cutover.go:104 swallows the error; ReadCutoverMarker distinguishes ErrNotExist. Fails closed in checkCutover but Presence's stale branch would answer unmarked. Same pattern pre-exists at issuelintids.go:231. Verify commit^{commit} first, then treat only a missing path as absent.
- **BR-2** [Minor] `single-marker-reader` Marker presence probed twice for lint-ids, with a third commit-tree reader in issuemigrate
  issuelintids.go:231 cat-file -e duplicates ReadCutoverMarkerAt's probe; issuemigrate.go:132-137 is a parallel commit-tree marker reader via mainView. Route lint-ids' mode decision through tracker.ReadCutoverMarkerAt (ARCH-DRY).
- **BR-3** [Minor] `doc-claim-drift` Spec says lint-ids without --head reads the checkout; it reads HEAD's committed tree
  --head defaults to HEAD and cardlessAdditions always calls GuardCutoverAt, so a working-tree-only marker edit is not seen. Consistent with line 231, but the Spec sentence should say so.
- **BR-4** [Minor] `unexercised-branch` checkCutover's markerAt-absent refusal (cutover.go:130) has no test and no reaching caller
  lint-ids checks marker presence at head before guarding, so this branch is unreachable today; a tracker-package unit test would pin it for future GuardCutoverAt callers.

## Open findings

- **BR-1** [Minor] `absent-vs-error-conflation` ReadCutoverMarkerAt treats any cat-file -e failure (bad ref, git error) as "marker absent"
- **BR-2** [Minor] `single-marker-reader` Marker presence probed twice for lint-ids, with a third commit-tree reader in issuemigrate
- **BR-3** [Minor] `doc-claim-drift` Spec says lint-ids without --head reads the checkout; it reads HEAD's committed tree
- **BR-4** [Minor] `unexercised-branch` checkCutover's markerAt-absent refusal (cutover.go:130) has no test and no reaching caller
