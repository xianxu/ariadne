---
gate: boundary-review
issue: 255
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-09-28T09:16:13-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Important
          title: metis.bak and test-repo are ariadne-layered, not cut over, and not recorded as excluded
          detail: Done-when 1 requires every ariadne-layer repo in ~/workspace to be cut over or recorded as deliberately excluded. Both have construct/ and workshop/ but no marker and no remote. Add a Log exclusion line or delete them.
          family: done-when-enumeration-incomplete
          round: 1
        - id: BR-2
          severity: Important
          title: Cutover checklist plan left unticked; contradicts the issue Plan it is archived with
          detail: workshop/plans/000255-fleet-tracker-cutover-checklist.md lines 242-319 still show every C-row, C12 decide and Phase D as open with blank digests. Tick them, mark them superseded by the Log, and point Phase D at 258.
          family: durable-artifact-drift
          round: 1
        - id: BR-3
          severity: Minor
          title: Per-repo Log lacks the dry-run counts and apply digests the Spec asked for (only ariadne has them)
          family: spec-record-incomplete
          round: 1
        - id: BR-4
          severity: Minor
          title: Sweep line lists sdlc-smoke as on the tracker although it was later deleted
          family: durable-artifact-drift
          round: 1
        - id: BR-5
          severity: Minor
          title: Legacy-mode removal in 258 is punted with no trigger or date for "a while"
          family: lifecycle-end-unnamed
          round: 1
      recipe: milestone-review
      blocked: true
    - "n": 2
      timestamp: "2026-09-28T09:18:30-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: Log "Excluded" bullet and checklist Excluded section name metis.bak and test-repo; verified on disk (1 commit, no remote, 0 issues each).
          round: 2
        - id: BR-2
          disposition: addressed
          note: Checklist C1-C13 rows ticked with digest/root Result lines, C11 marked superseded by A4, C12 decided, Phase D rows point at 258.
          round: 2
        - id: BR-3
          disposition: addressed
          note: Per-repository results table in the 255 Log carries counts, digest and tracker root for all 15 repos.
          round: 2
        - id: BR-4
          disposition: addressed
          note: Appended Excluded bullet corrects the sweep line; sdlc-smoke confirmed deleted from the workspace.
          round: 2
        - id: BR-5
          disposition: addressed
          note: 258 Log now says revisit by 2026-11-30 or when an sdlc change next touches a legacy path.
          round: 2
      recipe: milestone-review
      blocked: false
---

# Gate ledger — ariadne#255 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-09-28T09:16:13-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `done-when-enumeration-incomplete` metis.bak and test-repo are ariadne-layered, not cut over, and not recorded as excluded
  Done-when 1 requires every ariadne-layer repo in ~/workspace to be cut over or recorded as deliberately excluded. Both have construct/ and workshop/ but no marker and no remote. Add a Log exclusion line or delete them.
- **BR-2** [Important] `durable-artifact-drift` Cutover checklist plan left unticked; contradicts the issue Plan it is archived with
  workshop/plans/000255-fleet-tracker-cutover-checklist.md lines 242-319 still show every C-row, C12 decide and Phase D as open with blank digests. Tick them, mark them superseded by the Log, and point Phase D at 258.
- **BR-3** [Minor] `spec-record-incomplete` Per-repo Log lacks the dry-run counts and apply digests the Spec asked for (only ariadne has them)
- **BR-4** [Minor] `durable-artifact-drift` Sweep line lists sdlc-smoke as on the tracker although it was later deleted
- **BR-5** [Minor] `lifecycle-end-unnamed` Legacy-mode removal in 258 is punted with no trigger or date for "a while"

## Round 2 — 2026-09-28T09:18:30-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — Log "Excluded" bullet and checklist Excluded section name metis.bak and test-repo; verified on disk (1 commit, no remote, 0 issues each).
- BR-2 — addressed — Checklist C1-C13 rows ticked with digest/root Result lines, C11 marked superseded by A4, C12 decided, Phase D rows point at 258.
- BR-3 — addressed — Per-repository results table in the 255 Log carries counts, digest and tracker root for all 15 repos.
- BR-4 — addressed — Appended Excluded bullet corrects the sweep line; sdlc-smoke confirmed deleted from the workspace.
- BR-5 — addressed — 258 Log now says revisit by 2026-11-30 or when an sdlc change next touches a legacy path.

## Open findings

(none — every finding has been disposed)
