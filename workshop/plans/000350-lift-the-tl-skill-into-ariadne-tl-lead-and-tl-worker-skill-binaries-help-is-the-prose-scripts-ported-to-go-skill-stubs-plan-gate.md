---
gate: plan-quality
issue: 350
id_prefix: PQ
rounds:
    - "n": 1
      timestamp: "2026-10-13T17:30:01-07:00"
      agent: claude
      findings:
        - id: PQ-1
          severity: Minor
          title: Seen-set file renamed from .ss-seen to .tl-seen with no cleanup of the old file
          detail: ss.py:57-59 writes ".<file>.ss-seen"; the plan names ".tl-seen". Either keep the name or have the cutover delete the old files, and log the switch from grow-only (ss.py:62-65) to prune-to-owed as a deliberate behaviour change.
          family: artifact-removal-path
          round: 1
        - id: PQ-2
          severity: Minor
          title: M3 status tests are listed case by case in prose instead of as a strategy line
          detail: 'Compress to one line per risky function, e.g. "write/status: fakeSessionSync injects stale or lock at call k; assert write count and exit code". The table-driven code will carry the individual cases.'
          family: test-plan-enumeration
          round: 1
        - id: PQ-3
          severity: Minor
          title: No explicit Non-goals section; the deliberate exclusions are scattered through Log and Plan
          detail: 'Collect them in one place: no Couch change, tl-status.py retired rather than ported, the ops pointer and Python deletion left to the TL, no milestone field added to the sdlc observation.'
          family: missing-non-goals
          round: 1
        - id: PQ-4
          severity: Minor
          title: The §15 bullet in AGENTS.base.md tells every downstream Couch slot to run tl-worker without a fallback for a missing binary
          detail: Base-layer text reaches every downstream repo. Make the bullet say what to do when tl-worker is not installed (or point at the stub), and confirm dev-aliases puts tl-worker on PATH in the environments of non-ariadne repos.
          family: base-layer-downstream-impact
          round: 1
      blocked: false
content_hash: d8df071a8e696c6792fa0b86c270791622a12545b4879938a18eadb809d63dad
---

# Gate ledger — ariadne#350 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-13T17:30:01-07:00 (claude) — passed

### Raised

- **PQ-1** [Minor] `artifact-removal-path` Seen-set file renamed from .ss-seen to .tl-seen with no cleanup of the old file
  ss.py:57-59 writes ".<file>.ss-seen"; the plan names ".tl-seen". Either keep the name or have the cutover delete the old files, and log the switch from grow-only (ss.py:62-65) to prune-to-owed as a deliberate behaviour change.
- **PQ-2** [Minor] `test-plan-enumeration` M3 status tests are listed case by case in prose instead of as a strategy line
  Compress to one line per risky function, e.g. "write/status: fakeSessionSync injects stale or lock at call k; assert write count and exit code". The table-driven code will carry the individual cases.
- **PQ-3** [Minor] `missing-non-goals` No explicit Non-goals section; the deliberate exclusions are scattered through Log and Plan
  Collect them in one place: no Couch change, tl-status.py retired rather than ported, the ops pointer and Python deletion left to the TL, no milestone field added to the sdlc observation.
- **PQ-4** [Minor] `base-layer-downstream-impact` The §15 bullet in AGENTS.base.md tells every downstream Couch slot to run tl-worker without a fallback for a missing binary
  Base-layer text reaches every downstream repo. Make the bullet say what to do when tl-worker is not installed (or point at the stub), and confirm dev-aliases puts tl-worker on PATH in the environments of non-ariadne repos.

## Open findings

- **PQ-1** [Minor] `artifact-removal-path` Seen-set file renamed from .ss-seen to .tl-seen with no cleanup of the old file
- **PQ-2** [Minor] `test-plan-enumeration` M3 status tests are listed case by case in prose instead of as a strategy line
- **PQ-3** [Minor] `missing-non-goals` No explicit Non-goals section; the deliberate exclusions are scattered through Log and Plan
- **PQ-4** [Minor] `base-layer-downstream-impact` The §15 bullet in AGENTS.base.md tells every downstream Couch slot to run tl-worker without a fallback for a missing binary
