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
    - "n": 2
      timestamp: "2026-09-27T23:46:19-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: cutover.go:106 verifies commit^{commit} before ls-tree; only an empty listing is absent. cutover_test.go:147 (no-such-commit must error) goes red on the old cat-file -e probe.
          round: 2
        - id: BR-2
          disposition: addressed
          note: issuelintids.go:231 and issuemigrate.go:132 both use tracker.ReadCutoverMarkerAt; the guard reads via Repository.readMarker. Residual reconcile re-read raised separately.
          round: 2
        - id: BR-3
          disposition: addressed
          note: Spec now reads "from the tree of its head (--head, default HEAD), never the checkout's files", matching cardlessAdditions.
          round: 2
        - id: BR-4
          disposition: addressed
          note: TestGuardCutoverAtJudgesTheCommit asserts "commit <sha> has no" for an unmarked commit beside a marked checkout; without the cutover.go:139 branch the generic checkout message fails it.
          round: 2
      findings:
        - id: BR-5
          severity: Minor
          title: reconcile re-reads main's marker right after migratedAlready read it, and drops the parse error
          detail: '2nd finding in family single-marker-reader. Rule: read a commit''s marker once through ReadCutoverMarkerAt and pass the root along. At issuemigrate.go:518-522, mainView.Read plus a ParseCutoverMarker whose error is discarded repeats migratedAlready''s read. Fix: have migratedAlready return (root, done, err). Enumeration in this window: issuemigrate.go:518 (this one); issuelintids.go:231 followed by the GuardCutoverAt re-read (same reader, acceptable).'
          family: single-marker-reader
          round: 2
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

## Round 2 — 2026-09-27T23:46:19-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — cutover.go:106 verifies commit^{commit} before ls-tree; only an empty listing is absent. cutover_test.go:147 (no-such-commit must error) goes red on the old cat-file -e probe.
- BR-2 — addressed — issuelintids.go:231 and issuemigrate.go:132 both use tracker.ReadCutoverMarkerAt; the guard reads via Repository.readMarker. Residual reconcile re-read raised separately.
- BR-3 — addressed — Spec now reads "from the tree of its head (--head, default HEAD), never the checkout's files", matching cardlessAdditions.
- BR-4 — addressed — TestGuardCutoverAtJudgesTheCommit asserts "commit <sha> has no" for an unmarked commit beside a marked checkout; without the cutover.go:139 branch the generic checkout message fails it.

### Raised

- **BR-5** [Minor] `single-marker-reader` reconcile re-reads main's marker right after migratedAlready read it, and drops the parse error
  2nd finding in family single-marker-reader. Rule: read a commit's marker once through ReadCutoverMarkerAt and pass the root along. At issuemigrate.go:518-522, mainView.Read plus a ParseCutoverMarker whose error is discarded repeats migratedAlready's read. Fix: have migratedAlready return (root, done, err). Enumeration in this window: issuemigrate.go:518 (this one); issuelintids.go:231 followed by the GuardCutoverAt re-read (same reader, acceptable).

## Open findings

- **BR-5** [Minor] `single-marker-reader` reconcile re-reads main's marker right after migratedAlready read it, and drops the parse error
