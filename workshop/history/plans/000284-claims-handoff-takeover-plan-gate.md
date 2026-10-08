---
gate: plan-quality
issue: 284
id_prefix: PQ
rounds:
    - "n": 1
      timestamp: "2026-10-07T14:19:29-07:00"
      agent: claude
      findings:
        - id: PQ-1
          severity: Important
          title: Unclaim has no idempotent rerun when the card write lands but its response is lost
          detail: 'The recovery contract is "rerun; the card decides" (cardpublish.go:26-31), and claim settles reruns through errAlreadyMine. After an unclaim''s card write lands unconfirmed, the card has no claimant, so unclaimDecision refuses the rerun as not mine. For started work the checkout also stays on the issue branch because the switch never ran. Fix: record the releaser''s identity in release (or match release {branch, head} against this HEAD), and treat "already released by me" as success that finishes the switch. The same applies to open unclaim.'
          family: uncertain-outcome-reconciliation
          round: 1
        - id: PQ-2
          severity: Important
          title: D9 widens claim's relocation repair to CanHoldOwner but leaves the model's move statuses as active
          detail: moveRelocation (move.go:128-160) already relocates open cards; it has no status check. But issue.cue:219 declares move's statuses as categories.active, and claim.go:147 derives the repair from that row. Decide the question in the model by changing the move row to holdable, and keep claim deriving from OwnershipEvent("move") instead of switching to CanHoldOwner.
          family: model-is-single-source
          round: 1
        - id: PQ-3
          severity: Minor
          title: Each task lists its test cases in prose instead of naming the functions and one strategy line per risky function
          detail: 'Compress to the functions under test with a strategy line each: claimTimes → fuzz over malformed git log streams; republishDecision → property test over the (local, base, main) triple; ChangeCards → race with injected interleavings.'
          family: test-plan-enumerates-cases
          round: 1
        - id: PQ-4
          severity: Minor
          title: D10 assumes operation tokens are bare kinds and names a takeover token that does not exist
          detail: 'Tokens are verb-hex (trackerenv.go:181), so the parser must match on the prefix. D7 should name the token it uses for takeover. The git log scan is unbounded over tracker history: stream it and stop once every owned card is resolved.'
          family: claim-age-derivation
          round: 1
        - id: PQ-5
          severity: Minor
          title: Issue branches pushed by a started unclaim have no stated removal path
          family: artifact-removal-path
          round: 1
        - id: PQ-6
          severity: Minor
          title: release.branch comes from another machine and should be validated as a ref name before fetch and checkout
          detail: Single-line validation alone is not enough. Require git check-ref-format plus the issue-branch pattern, and keep fetch arguments in argv form.
          family: untrusted-card-input
          round: 1
        - id: PQ-7
          severity: Minor
          title: The issue publish "nothing" outcome must still run the rest fast-forward or branch commit; also state how --commit and --issue share the command
          family: rerun-completes-local-effects
          round: 1
      blocked: true
    - "n": 2
      timestamp: "2026-10-07T14:20:18-07:00"
      agent: claude
      dispose:
        - id: PQ-1
          disposition: addressed
          note: release.by is written on every unclaim; errAlreadyReleased finishes the switch or the rest fast-forward; tested in Tasks 6 and 9.
          round: 2
        - id: PQ-2
          disposition: addressed
          note: The move row becomes holdable in issue.cue; claim.go:147 keeps deriving from OwnershipEvent("move"); pkg/vocab assertion added.
          round: 2
        - id: PQ-3
          disposition: addressed
          note: Revision 3 states one strategy line per risky function; the per-task case lists are demoted to examples.
          round: 2
        - id: PQ-4
          disposition: addressed
          note: Prefix match on verb-hex (trackerenv.go:181); takeover and unclaim tokens named; streamed scan with early stop.
          round: 2
        - id: PQ-5
          disposition: addressed
          note: 'Removed when the branch lands (merge, #286) or by abandon (#286); stated in the unclaim help.'
          round: 2
        - id: PQ-6
          disposition: addressed
          note: check-ref-format plus the issue's own branch name, argv after --, adversarial names tested in Task 8.
          round: 2
        - id: PQ-7
          disposition: addressed
          note: Mode split between tracker and legacy repositories stated; nothing-to-publish still runs the local finish; rerun tested.
          round: 2
      blocked: false
    - "n": 3
      timestamp: "2026-10-07T14:21:02-07:00"
      agent: claude
      dispose:
        - id: PQ-1
          disposition: addressed
          note: Release{by,...} written on every unclaim; errAlreadyReleased lets a rerun finish the remaining work (Revisions 1).
          round: 3
        - id: PQ-2
          disposition: addressed
          note: 'The move row in issue.cue becomes statuses: holdable; claim keeps deriving from the model (Revisions 2).'
          round: 3
        - id: PQ-3
          disposition: addressed
          note: Per-function strategy lines replace the case lists as the contract (Revisions 3).
          round: 3
        - id: PQ-4
          disposition: addressed
          note: Prefix match on operationToken verb-<hex>; takeover and unclaim use operationToken (Revisions 4).
          round: 3
        - id: PQ-5
          disposition: addressed
          note: 'The pushed branch is removed when it lands (merge) or by abandon in #286; the help says so (Revisions 5).'
          round: 3
        - id: PQ-6
          disposition: addressed
          note: check-ref-format, own-branch check, argv after --, plus Task 8 tests (Revisions 6).
          round: 3
        - id: PQ-7
          disposition: addressed
          note: Mode split by repository kind; "nothing" still runs the local finish; rerun test in Task 5 (Revisions 7).
          round: 3
      blocked: false
content_hash: 7a9110df457e9f8ec5817e2b44367f4f1ffa0d9a470ca8fb51143af82841356c
---

# Gate ledger — ariadne#284 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-07T14:19:29-07:00 (claude) — BLOCKED

### Raised

- **PQ-1** [Important] `uncertain-outcome-reconciliation` Unclaim has no idempotent rerun when the card write lands but its response is lost
  The recovery contract is "rerun; the card decides" (cardpublish.go:26-31), and claim settles reruns through errAlreadyMine. After an unclaim's card write lands unconfirmed, the card has no claimant, so unclaimDecision refuses the rerun as not mine. For started work the checkout also stays on the issue branch because the switch never ran. Fix: record the releaser's identity in release (or match release {branch, head} against this HEAD), and treat "already released by me" as success that finishes the switch. The same applies to open unclaim.
- **PQ-2** [Important] `model-is-single-source` D9 widens claim's relocation repair to CanHoldOwner but leaves the model's move statuses as active
  moveRelocation (move.go:128-160) already relocates open cards; it has no status check. But issue.cue:219 declares move's statuses as categories.active, and claim.go:147 derives the repair from that row. Decide the question in the model by changing the move row to holdable, and keep claim deriving from OwnershipEvent("move") instead of switching to CanHoldOwner.
- **PQ-3** [Minor] `test-plan-enumerates-cases` Each task lists its test cases in prose instead of naming the functions and one strategy line per risky function
  Compress to the functions under test with a strategy line each: claimTimes → fuzz over malformed git log streams; republishDecision → property test over the (local, base, main) triple; ChangeCards → race with injected interleavings.
- **PQ-4** [Minor] `claim-age-derivation` D10 assumes operation tokens are bare kinds and names a takeover token that does not exist
  Tokens are verb-hex (trackerenv.go:181), so the parser must match on the prefix. D7 should name the token it uses for takeover. The git log scan is unbounded over tracker history: stream it and stop once every owned card is resolved.
- **PQ-5** [Minor] `artifact-removal-path` Issue branches pushed by a started unclaim have no stated removal path
- **PQ-6** [Minor] `untrusted-card-input` release.branch comes from another machine and should be validated as a ref name before fetch and checkout
  Single-line validation alone is not enough. Require git check-ref-format plus the issue-branch pattern, and keep fetch arguments in argv form.
- **PQ-7** [Minor] `rerun-completes-local-effects` The issue publish "nothing" outcome must still run the rest fast-forward or branch commit; also state how --commit and --issue share the command

## Round 2 — 2026-10-07T14:20:18-07:00 (claude) — passed

### Disposed

- PQ-1 — addressed — release.by is written on every unclaim; errAlreadyReleased finishes the switch or the rest fast-forward; tested in Tasks 6 and 9.
- PQ-2 — addressed — The move row becomes holdable in issue.cue; claim.go:147 keeps deriving from OwnershipEvent("move"); pkg/vocab assertion added.
- PQ-3 — addressed — Revision 3 states one strategy line per risky function; the per-task case lists are demoted to examples.
- PQ-4 — addressed — Prefix match on verb-hex (trackerenv.go:181); takeover and unclaim tokens named; streamed scan with early stop.
- PQ-5 — addressed — Removed when the branch lands (merge, #286) or by abandon (#286); stated in the unclaim help.
- PQ-6 — addressed — check-ref-format plus the issue's own branch name, argv after --, adversarial names tested in Task 8.
- PQ-7 — addressed — Mode split between tracker and legacy repositories stated; nothing-to-publish still runs the local finish; rerun tested.

## Round 3 — 2026-10-07T14:21:02-07:00 (claude) — passed

### Disposed

- PQ-1 — addressed — Release{by,...} written on every unclaim; errAlreadyReleased lets a rerun finish the remaining work (Revisions 1).
- PQ-2 — addressed — The move row in issue.cue becomes statuses: holdable; claim keeps deriving from the model (Revisions 2).
- PQ-3 — addressed — Per-function strategy lines replace the case lists as the contract (Revisions 3).
- PQ-4 — addressed — Prefix match on operationToken verb-<hex>; takeover and unclaim use operationToken (Revisions 4).
- PQ-5 — addressed — The pushed branch is removed when it lands (merge) or by abandon in #286; the help says so (Revisions 5).
- PQ-6 — addressed — check-ref-format, own-branch check, argv after --, plus Task 8 tests (Revisions 6).
- PQ-7 — addressed — Mode split by repository kind; "nothing" still runs the local finish; rerun test in Task 5 (Revisions 7).

## Open findings

(none — every finding has been disposed)
