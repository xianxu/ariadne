---
gate: plan-quality
issue: 288
id_prefix: PQ
rounds:
    - "n": 1
      timestamp: "2026-10-02T13:23:45-07:00"
      agent: claude
      blocked: true
      protocol_error: no valid findings block
    - "n": 2
      timestamp: "2026-10-02T13:24:29-07:00"
      agent: claude
      findings:
        - id: PQ-1
          severity: Minor
          title: ARCH-ORDER not addressed; state "holds no state between events because inventory is a single-shot read"
          detail: The principle requires a stated claim rather than an omission, even when it is effectively N/A.
          family: arch-order-unstated
          round: 2
        - id: PQ-2
          severity: Minor
          title: Claim placement compares canonRoot output with workspace.CanonicalPath output
          detail: claimant.worktree comes from canonRoot (propagatebase.go:41) while tree_path comes from workspace.CanonicalPath (fleet/gitpaths.go:17); if they diverge, live claims read as dangling. Use one helper and test with a symlinked root.
          family: path-identity-single-source
          round: 2
        - id: PQ-3
          severity: Minor
          title: Task 2/4 test bullets enumerate cases in prose; compress to one strategy line per risky function
          family: test-prose-enumeration
          round: 2
      blocked: false
    - "n": 3
      timestamp: "2026-10-02T13:25:50-07:00"
      agent: claude
      dispose:
        - id: PQ-1
          disposition: addressed
          note: Decisions now state "holds no state between events because inventory is a single-shot read over one pinned snapshot".
          round: 3
        - id: PQ-2
          disposition: addressed
          note: LookupRepoClaims re-canonicalizes with the fleet's canonicalPath (workspace.CanonicalPath), the same helper that produces tree_path; only one helper decides placement.
          round: 3
        - id: PQ-3
          disposition: not-addressed
          note: 'Task 4 fixed (cross product); Task 2 e2e still enumerates six verb checks. Rule: one bullet = function + adversarial class + invariant guarded.'
          round: 3
      findings:
        - id: PQ-4
          severity: Minor
          title: Task 2 site table is a line-numbered call-site inventory; four lines do not match a Card == nil check
          detail: close.go:515, actual.go:173, trackercompletion.go:39 and projectstatus.go:304 were confirmed; push.go:561, issuefiles.go:85, issue.go:606 and observe.go:50 did not match a grep for Card == nil. Find the sites by rec.Card use rather than trusting the line numbers.
          family: test-prose-enumeration
          round: 3
      blocked: false
content_hash: d7ebaa248f2e450e6f810dcb8e4e6c90dcce80efaa4f6ae4435618aa2b6697de
---

# Gate ledger — ariadne#288 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-02T13:23:45-07:00 (claude) — BLOCKED

**Protocol error:** no valid findings block — this round contributed no findings.

## Round 2 — 2026-10-02T13:24:29-07:00 (claude) — passed

### Raised

- **PQ-1** [Minor] `arch-order-unstated` ARCH-ORDER not addressed; state "holds no state between events because inventory is a single-shot read"
  The principle requires a stated claim rather than an omission, even when it is effectively N/A.
- **PQ-2** [Minor] `path-identity-single-source` Claim placement compares canonRoot output with workspace.CanonicalPath output
  claimant.worktree comes from canonRoot (propagatebase.go:41) while tree_path comes from workspace.CanonicalPath (fleet/gitpaths.go:17); if they diverge, live claims read as dangling. Use one helper and test with a symlinked root.
- **PQ-3** [Minor] `test-prose-enumeration` Task 2/4 test bullets enumerate cases in prose; compress to one strategy line per risky function

## Round 3 — 2026-10-02T13:25:50-07:00 (claude) — passed

### Disposed

- PQ-1 — addressed — Decisions now state "holds no state between events because inventory is a single-shot read over one pinned snapshot".
- PQ-2 — addressed — LookupRepoClaims re-canonicalizes with the fleet's canonicalPath (workspace.CanonicalPath), the same helper that produces tree_path; only one helper decides placement.
- PQ-3 — not-addressed — Task 4 fixed (cross product); Task 2 e2e still enumerates six verb checks. Rule: one bullet = function + adversarial class + invariant guarded.

### Raised

- **PQ-4** [Minor] `test-prose-enumeration` Task 2 site table is a line-numbered call-site inventory; four lines do not match a Card == nil check
  close.go:515, actual.go:173, trackercompletion.go:39 and projectstatus.go:304 were confirmed; push.go:561, issuefiles.go:85, issue.go:606 and observe.go:50 did not match a grep for Card == nil. Find the sites by rec.Card use rather than trusting the line numbers.

## Open findings

- **PQ-3** [Minor] `test-prose-enumeration` Task 2/4 test bullets enumerate cases in prose; compress to one strategy line per risky function
- **PQ-4** [Minor] `test-prose-enumeration` Task 2 site table is a line-numbered call-site inventory; four lines do not match a Card == nil check
