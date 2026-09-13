---
gate: boundary-review
issue: 225
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-09-13T12:41:58-07:00"
      agent: codex
      findings:
        - id: BR-1
          severity: Critical
          title: Workflow discovery is classified PURE despite filesystem-dependent behavior
          detail: 'Plan line 21 contradicts Makefile line 11 and its filesystem/subprocess test. ARCH-PURE: classify workflow discovery as INTEGRATION and append a plan revision; no runtime redesign is required.'
          family: core-concept-classification
          round: 1
        - id: BR-2
          severity: Important
          title: README update is missing for portable consumer setup
          detail: The workflow introduces executable scripts/ci-setup.sh at line 56, but README.md is unchanged in the review range. Document seed ownership, Makefile.local, bootstrap, and the hook's execution order and failure behavior.
          family: user-surface-documentation
          round: 1
      blocked: true
    - "n": 2
      timestamp: "2026-09-13T12:45:10-07:00"
      agent: codex
      dispose:
        - id: BR-1
          disposition: not-addressed
          note: The plan now correctly classifies workflow discovery as integration and includes a revision. No regression test pins that classification; reverting the documentation correction leaves the existing runtime tests unchanged. The required failing-without-fix evidence is missing.
          round: 2
        - id: BR-2
          disposition: not-addressed
          note: README now documents seed ownership, Makefile.local, bootstrap, and CI hook ordering and failure behavior. Neither portable test reads README, so removing this documentation leaves those tests unchanged. The required failing-without-fix evidence is missing.
          round: 2
      blocked: true
    - "n": 3
      timestamp: "2026-09-13T12:47:19-07:00"
      agent: codex
      dispose:
        - id: BR-1
          disposition: not-addressed
          note: The plan correctly classifies workflow discovery as INTEGRATION and records the correction. No test pins that correction; plan lines 88–93 explicitly decline that evidence. The requested fails-without-the-fix condition remains unmet.
          round: 3
        - id: BR-2
          disposition: not-addressed
          note: README lines 15–34 document seed ownership, Makefile.local, bootstrap, and hook ordering/failure behavior. No test pins this documentation correction, so the requested fails-without-the-fix condition remains unmet.
          round: 3
      blocked: true
---

# Gate ledger — ariadne#225 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-09-13T12:41:58-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `core-concept-classification` Workflow discovery is classified PURE despite filesystem-dependent behavior
  Plan line 21 contradicts Makefile line 11 and its filesystem/subprocess test. ARCH-PURE: classify workflow discovery as INTEGRATION and append a plan revision; no runtime redesign is required.
- **BR-2** [Important] `user-surface-documentation` README update is missing for portable consumer setup
  The workflow introduces executable scripts/ci-setup.sh at line 56, but README.md is unchanged in the review range. Document seed ownership, Makefile.local, bootstrap, and the hook's execution order and failure behavior.

## Round 2 — 2026-09-13T12:45:10-07:00 (codex) — BLOCKED

### Disposed

- BR-1 — not-addressed — The plan now correctly classifies workflow discovery as integration and includes a revision. No regression test pins that classification; reverting the documentation correction leaves the existing runtime tests unchanged. The required failing-without-fix evidence is missing.
- BR-2 — not-addressed — README now documents seed ownership, Makefile.local, bootstrap, and CI hook ordering and failure behavior. Neither portable test reads README, so removing this documentation leaves those tests unchanged. The required failing-without-fix evidence is missing.

## Round 3 — 2026-09-13T12:47:19-07:00 (codex) — BLOCKED

### Disposed

- BR-1 — not-addressed — The plan correctly classifies workflow discovery as INTEGRATION and records the correction. No test pins that correction; plan lines 88–93 explicitly decline that evidence. The requested fails-without-the-fix condition remains unmet.
- BR-2 — not-addressed — README lines 15–34 document seed ownership, Makefile.local, bootstrap, and hook ordering/failure behavior. No test pins this documentation correction, so the requested fails-without-the-fix condition remains unmet.

## Open findings

- **BR-1** [Critical] `core-concept-classification` Workflow discovery is classified PURE despite filesystem-dependent behavior
- **BR-2** [Important] `user-surface-documentation` README update is missing for portable consumer setup
