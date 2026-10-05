---
gate: boundary-review
issue: 294
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-05T15:24:50-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Minor
          title: DeclarationReader path makes non-ENOENT construct/deps read errors loud for all Walk consumers
          detail: Under OSFS, permission, ENOTDIR, and symlinked construct/deps now abort Walk in datatype and vocabulary, where they were silently treated as absent. Arguably correct, but not in the Spec or Log; record it and add a test for one non-symlink error.
          family: undocumented-behavior-change
          round: 1
        - id: BR-2
          severity: Minor
          title: Plain-FS fallback reads the whole file before applying DeclarationLimit
          detail: readDeclaration's fallback bounds what is parsed, not what is read into memory. Fine for fakes and weave's FS; the comment should say so.
          family: bounded-read-claim
          round: 1
        - id: BR-3
          severity: Minor
          title: DeclaredSubstrates does not say root must be absolute, unlike Walk
          detail: A relative root gives a relative Owner, contradicting the field doc's absolute-path promise.
          family: api-precondition-doc
          round: 1
      recipe: milestone-review
      blocked: false
    - "n": 2
      timestamp: "2026-10-05T15:29:47-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: Logged in the issue Log; tested by TestUnreadableDeclarationIsLoudAndRootsAreAbsolute (a directory at construct/deps errors for both Walk and DeclaredSubstrates).
          round: 2
        - id: BR-2
          disposition: addressed
          note: The readDeclaration comment now says the fallback bounds what is parsed, not what ReadFile loads.
          round: 2
        - id: BR-3
          disposition: addressed
          note: DeclaredSubstrates calls filepath.Abs before physical(), the doc says so, and a relative-root test after t.Chdir checks that the paths are absolute.
          round: 2
      recipe: milestone-review
      blocked: false
---

# Gate ledger — ariadne#294 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-05T15:24:50-07:00 (claude) — passed

### Raised

- **BR-1** [Minor] `undocumented-behavior-change` DeclarationReader path makes non-ENOENT construct/deps read errors loud for all Walk consumers
  Under OSFS, permission, ENOTDIR, and symlinked construct/deps now abort Walk in datatype and vocabulary, where they were silently treated as absent. Arguably correct, but not in the Spec or Log; record it and add a test for one non-symlink error.
- **BR-2** [Minor] `bounded-read-claim` Plain-FS fallback reads the whole file before applying DeclarationLimit
  readDeclaration's fallback bounds what is parsed, not what is read into memory. Fine for fakes and weave's FS; the comment should say so.
- **BR-3** [Minor] `api-precondition-doc` DeclaredSubstrates does not say root must be absolute, unlike Walk
  A relative root gives a relative Owner, contradicting the field doc's absolute-path promise.

## Round 2 — 2026-10-05T15:29:47-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — Logged in the issue Log; tested by TestUnreadableDeclarationIsLoudAndRootsAreAbsolute (a directory at construct/deps errors for both Walk and DeclaredSubstrates).
- BR-2 — addressed — The readDeclaration comment now says the fallback bounds what is parsed, not what ReadFile loads.
- BR-3 — addressed — DeclaredSubstrates calls filepath.Abs before physical(), the doc says so, and a relative-root test after t.Chdir checks that the paths are absolute.

## Open findings

(none — every finding has been disposed)
