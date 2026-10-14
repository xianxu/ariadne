---
gate: boundary-review
issue: 333
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-13T23:14:58-07:00"
      dispatched: "2026-10-13T23:13:28-07:00"
      agent: codex
      findings:
        - id: BR-1
          severity: Important
          title: README update appears missing for the host-preflight hook
          detail: 'Makefile.workflow:20–27 introduces two user-facing surfaces: automatic tests/host-preflight.sh discovery and the WF_TEST_PREFLIGHT override. Both are documented in atlas/workflow/sdlc-binary.md:1635, but README.md is unchanged. Add a concise README usage paragraph covering the convention, setting the override before including Makefile.workflow, and failure before lock acquisition, with an atlas link. These are all instances of this family in the window.'
          family: user-facing-surface-documentation
          round: 1
      recipe: small-diff-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#333 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-13T23:14:58-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Important] `user-facing-surface-documentation` README update appears missing for the host-preflight hook
  Makefile.workflow:20–27 introduces two user-facing surfaces: automatic tests/host-preflight.sh discovery and the WF_TEST_PREFLIGHT override. Both are documented in atlas/workflow/sdlc-binary.md:1635, but README.md is unchanged. Add a concise README usage paragraph covering the convention, setting the override before including Makefile.workflow, and failure before lock acquisition, with an atlas link. These are all instances of this family in the window.

## Open findings

- **BR-1** [Important] `user-facing-surface-documentation` README update appears missing for the host-preflight hook
