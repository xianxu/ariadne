---
gate: boundary-review
issue: 194
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-08-20T19:17:37-07:00"
      agent: claude
      boundary: M2
      blocked: true
      protocol_error: no valid findings block
    - "n": 2
      timestamp: "2026-08-20T19:27:06-07:00"
      agent: claude
      boundary: M2
      blocked: true
      protocol_error: no valid findings block
    - "n": 3
      timestamp: "2026-08-20T21:27:45-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Critical
          title: Prior findings never reach the reviewer — reviewThenFinalizeLocked blanks PlansDir, so the ledger blocks on findings it cannot dispose
          detail: |-
            close.go:1128 sets dispatchParams.PlansDir = "" before dispatch, and
            boundaryledger.go:69 returns "" on exactly that field, so PriorFindings is
            empty in every live close / milestone-close review (the unlocked
            reviewThenFinalize path never dispatches: close --milestone is refused since
            146, and the milestone short-circuits are --no-judge / --force / --dry-run).
            persistBoundaryRound still runs with the un-blanked params, so the ledger is
            written and enforced. A Critical BR finding therefore wedges the boundary
            permanently: the next round is shown "no prior rounds", cannot emit a
            dispose entry for an id it was never handed, and BlocksPastCap keeps Critical
            blocking forever. Escape is only --no-judge / --force, which skip the review
            entirely. Fix by reading the block under the repo lock beside
            captureCloseReviewSnapshot and carrying it on boundaryReviewParams, rather
            than overloading PlansDir as a may-I-write flag.
          round: 3
        - id: BR-2
          severity: Important
          title: A corrupt boundary ledger silently disables the gate and drops the round
          detail: |-
            boundaryledger.go:141-143 warns and returns gatestate.Decision{} on a read
            error, so Block is false, this round's findings are discarded, the corrupt
            file is left unwritten, and the close finalizes. The plan gate does the
            opposite (changecode.go:424-429 halts), and readGateLedger's own doc says a
            silent reset is worse than the status quo because it would look like it
            worked. Route the read failure to closeHalt instead.
          round: 3
        - id: BR-3
          severity: Important
          title: Round.Blocked is never stamped, so the boundary ledger records "passed" for rounds that blocked
          detail: |-
            changecode.go:536-537 stamps ledger.Rounds[last].Blocked = d.Block after
            Decide; boundaryledger.go:159-171 never does. render.go:82-85 then prints
            "— passed" and the frontmatter writes blocked false for a round that refused
            the boundary. ledger.go:300-304 (PassesUnchanged) reads that field, which is
            what 183's --fixed-to-ship pass-through will depend on at this same gate.
          round: 3
        - id: BR-4
          severity: Important
          title: atlas/workflow/gate-state.md not updated; it now asserts the mechanism this diff deleted
          detail: |-
            Lines 75-78 still claim code-review.md instructs the boundary reviewer to
            read the ledger's Open findings — that section was deleted in this diff and
            replaced by seeding. Also stale: line 22 (only -plan-gate.md), line 73 (only
            WF_PLAN_ROUND_CAP), line 153's code-map row, and no row for boundaryledger.go;
            ledger-landscape.md:77 likewise names only *-plan-gate.md. This is the second
            instance of M1 I1's family, so the durable fix is the rule: when a mechanism
            moves, grep atlas/ for its name in the same commit.
          round: 3
        - id: BR-5
          severity: Important
          title: 'Docs gate: helptext documents no part of the new boundary gate'
          detail: |-
            close.md / milestone-close.md cover M1's anchor but omit the -close-gate.md
            artifact, the new "verdict SHIP but the gate ledger still has open blocking
            finding(s)" refusal (a close that refuses despite a passing review), the
            dispose-before-raising contract, and WF_BOUNDARY_ROUND_CAP (change-code.md:85
            documents the sibling knob). close.md's BYPASSING A GATE table also enumerates
            a flag per gate while the ledger block has none, so it now under-reports.
          round: 3
        - id: BR-6
          severity: Important
          title: construct/generated/vocabulary/finding.json is stale — the export no longer derives from finding.cue
          detail: |-
            finding.cue:68 says *-gate.md; construct/generated/vocabulary/finding.json:35
            still says *-plan-gate.md. Verified: `go run ./cmd/vocabulary check --output
            construct/generated/vocabulary` reports STALE and exits 1. pkg/vocab/finding.json
            was updated and the published export beside it was not. This repo is the base
            layer, so construct/generated propagates downstream (ARCH-PURPOSE). The plan's
            Verification list already calls for running this target. Fix with `make weave`
            and read the resulting diff.
          round: 3
        - id: BR-7
          severity: Important
          title: The plan's D4 heading states the opposite of the shipped behavior, and the plan has no Revisions section
          detail: |-
            plan.md:157 reads "a boundary protocol miss halts" while its own body and the
            code implement warn-and-persist, so grepping the D-headings returns the inverse
            of the truth. The plan also carries two mid-stream revisions (D4's reversal, the
            Core-concepts correction) as in-place blockquotes with no appended Revisions
            section, which AGENTS.md requires and M1's review recommended.
          round: 3
        - id: BR-8
          severity: Minor
          title: The BoundaryAll seed round consumes a round-cap slot at every boundary
          detail: |-
            Decide counts len(l.Rounds) and FilterBoundary retains every "*" round, so a
            seeded issue gets 2 real rounds before Important findings demote, not 3.
          round: 3
        - id: BR-9
          severity: Minor
          title: A dispatch failure persists a blocked round with an empty protocol_error and empty agent
          detail: |-
            milestoneclose.go:566,573 return res(...) with Round nil, ProtocolError "" and
            Agent "", so persistBoundaryRound records a round for a review that never ran,
            indistinguishable in the frontmatter from a reviewer that emitted no fence.
          round: 3
        - id: BR-10
          severity: Minor
          title: One bad disposition id nullifies a whole round's valid dispositions
          detail: |-
            ApplyChecked returns on the first unknown or unmodeled disposition and
            boundaryledger.go:166 then drops all of them, at the gate whose entire purpose
            is disposal. Same shape at the plan gate, so the fix belongs in gatestate:
            return the offending ids and drop only those.
          round: 3
        - id: BR-11
          severity: Minor
          title: persistBoundaryRound's new operator lines have no assertNoGatesigCollision guard
          detail: |-
            The unconditional cwarn/cok/cinfo lines added here skip the derived guard that
            this same issue added to formatAnchorDocsOnly one milestone ago (M1 I5), so a
            future GateCatalog row can silently collide with them.
          round: 3
        - id: BR-12
          severity: Minor
          title: 'previousReviewBoundary greps Review-Verdict: unanchored over the whole commit message'
          detail: |-
            milestoneclose.go:342 uses --grep=Review-Verdict:, which matches a commit BODY.
            Commit 23d5b8a in this very window came one character from matching in prose.
            Anchoring it is a one-token hardening, adjacent to 197 and the same class as
            the lessons.md entry this diff added.
          round: 3
        - id: BR-13
          severity: Minor
          title: seedFromPlanGate mints ids by index instead of via nextIDSeq
          detail: |-
            boundaryledger.go:107-114 formats BR-<i+1> directly; correct only because the
            function runs on an empty ledger, and nothing pins that precondition.
          round: 3
      boundary: M2
      blocked: false
---

# Gate ledger — ariadne#194 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-08-20T19:17:37-07:00 (claude) — BLOCKED

**Protocol error:** no valid findings block — this round contributed no findings.

## Round 2 — 2026-08-20T19:27:06-07:00 (claude) — BLOCKED

**Protocol error:** no valid findings block — this round contributed no findings.

## Round 3 — 2026-08-20T21:27:45-07:00 (claude) — passed

### Raised

- **BR-1** [Critical] Prior findings never reach the reviewer — reviewThenFinalizeLocked blanks PlansDir, so the ledger blocks on findings it cannot dispose
  close.go:1128 sets dispatchParams.PlansDir = "" before dispatch, and
  boundaryledger.go:69 returns "" on exactly that field, so PriorFindings is
  empty in every live close / milestone-close review (the unlocked
  reviewThenFinalize path never dispatches: close --milestone is refused since
  146, and the milestone short-circuits are --no-judge / --force / --dry-run).
  persistBoundaryRound still runs with the un-blanked params, so the ledger is
  written and enforced. A Critical BR finding therefore wedges the boundary
  permanently: the next round is shown "no prior rounds", cannot emit a
  dispose entry for an id it was never handed, and BlocksPastCap keeps Critical
  blocking forever. Escape is only --no-judge / --force, which skip the review
  entirely. Fix by reading the block under the repo lock beside
  captureCloseReviewSnapshot and carrying it on boundaryReviewParams, rather
  than overloading PlansDir as a may-I-write flag.
- **BR-2** [Important] A corrupt boundary ledger silently disables the gate and drops the round
  boundaryledger.go:141-143 warns and returns gatestate.Decision{} on a read
  error, so Block is false, this round's findings are discarded, the corrupt
  file is left unwritten, and the close finalizes. The plan gate does the
  opposite (changecode.go:424-429 halts), and readGateLedger's own doc says a
  silent reset is worse than the status quo because it would look like it
  worked. Route the read failure to closeHalt instead.
- **BR-3** [Important] Round.Blocked is never stamped, so the boundary ledger records "passed" for rounds that blocked
  changecode.go:536-537 stamps ledger.Rounds[last].Blocked = d.Block after
  Decide; boundaryledger.go:159-171 never does. render.go:82-85 then prints
  "— passed" and the frontmatter writes blocked false for a round that refused
  the boundary. ledger.go:300-304 (PassesUnchanged) reads that field, which is
  what 183's --fixed-to-ship pass-through will depend on at this same gate.
- **BR-4** [Important] atlas/workflow/gate-state.md not updated; it now asserts the mechanism this diff deleted
  Lines 75-78 still claim code-review.md instructs the boundary reviewer to
  read the ledger's Open findings — that section was deleted in this diff and
  replaced by seeding. Also stale: line 22 (only -plan-gate.md), line 73 (only
  WF_PLAN_ROUND_CAP), line 153's code-map row, and no row for boundaryledger.go;
  ledger-landscape.md:77 likewise names only *-plan-gate.md. This is the second
  instance of M1 I1's family, so the durable fix is the rule: when a mechanism
  moves, grep atlas/ for its name in the same commit.
- **BR-5** [Important] Docs gate: helptext documents no part of the new boundary gate
  close.md / milestone-close.md cover M1's anchor but omit the -close-gate.md
  artifact, the new "verdict SHIP but the gate ledger still has open blocking
  finding(s)" refusal (a close that refuses despite a passing review), the
  dispose-before-raising contract, and WF_BOUNDARY_ROUND_CAP (change-code.md:85
  documents the sibling knob). close.md's BYPASSING A GATE table also enumerates
  a flag per gate while the ledger block has none, so it now under-reports.
- **BR-6** [Important] construct/generated/vocabulary/finding.json is stale — the export no longer derives from finding.cue
  finding.cue:68 says *-gate.md; construct/generated/vocabulary/finding.json:35
  still says *-plan-gate.md. Verified: `go run ./cmd/vocabulary check --output
  construct/generated/vocabulary` reports STALE and exits 1. pkg/vocab/finding.json
  was updated and the published export beside it was not. This repo is the base
  layer, so construct/generated propagates downstream (ARCH-PURPOSE). The plan's
  Verification list already calls for running this target. Fix with `make weave`
  and read the resulting diff.
- **BR-7** [Important] The plan's D4 heading states the opposite of the shipped behavior, and the plan has no Revisions section
  plan.md:157 reads "a boundary protocol miss halts" while its own body and the
  code implement warn-and-persist, so grepping the D-headings returns the inverse
  of the truth. The plan also carries two mid-stream revisions (D4's reversal, the
  Core-concepts correction) as in-place blockquotes with no appended Revisions
  section, which AGENTS.md requires and M1's review recommended.
- **BR-8** [Minor] The BoundaryAll seed round consumes a round-cap slot at every boundary
  Decide counts len(l.Rounds) and FilterBoundary retains every "*" round, so a
  seeded issue gets 2 real rounds before Important findings demote, not 3.
- **BR-9** [Minor] A dispatch failure persists a blocked round with an empty protocol_error and empty agent
  milestoneclose.go:566,573 return res(...) with Round nil, ProtocolError "" and
  Agent "", so persistBoundaryRound records a round for a review that never ran,
  indistinguishable in the frontmatter from a reviewer that emitted no fence.
- **BR-10** [Minor] One bad disposition id nullifies a whole round's valid dispositions
  ApplyChecked returns on the first unknown or unmodeled disposition and
  boundaryledger.go:166 then drops all of them, at the gate whose entire purpose
  is disposal. Same shape at the plan gate, so the fix belongs in gatestate:
  return the offending ids and drop only those.
- **BR-11** [Minor] persistBoundaryRound's new operator lines have no assertNoGatesigCollision guard
  The unconditional cwarn/cok/cinfo lines added here skip the derived guard that
  this same issue added to formatAnchorDocsOnly one milestone ago (M1 I5), so a
  future GateCatalog row can silently collide with them.
- **BR-12** [Minor] previousReviewBoundary greps Review-Verdict: unanchored over the whole commit message
  milestoneclose.go:342 uses --grep=Review-Verdict:, which matches a commit BODY.
  Commit 23d5b8a in this very window came one character from matching in prose.
  Anchoring it is a one-token hardening, adjacent to 197 and the same class as
  the lessons.md entry this diff added.
- **BR-13** [Minor] seedFromPlanGate mints ids by index instead of via nextIDSeq
  boundaryledger.go:107-114 formats BR-<i+1> directly; correct only because the
  function runs on an empty ledger, and nothing pins that precondition.

## Open findings

- **BR-1** [Critical] Prior findings never reach the reviewer — reviewThenFinalizeLocked blanks PlansDir, so the ledger blocks on findings it cannot dispose
- **BR-2** [Important] A corrupt boundary ledger silently disables the gate and drops the round
- **BR-3** [Important] Round.Blocked is never stamped, so the boundary ledger records "passed" for rounds that blocked
- **BR-4** [Important] atlas/workflow/gate-state.md not updated; it now asserts the mechanism this diff deleted
- **BR-5** [Important] Docs gate: helptext documents no part of the new boundary gate
- **BR-6** [Important] construct/generated/vocabulary/finding.json is stale — the export no longer derives from finding.cue
- **BR-7** [Important] The plan's D4 heading states the opposite of the shipped behavior, and the plan has no Revisions section
- **BR-8** [Minor] The BoundaryAll seed round consumes a round-cap slot at every boundary
- **BR-9** [Minor] A dispatch failure persists a blocked round with an empty protocol_error and empty agent
- **BR-10** [Minor] One bad disposition id nullifies a whole round's valid dispositions
- **BR-11** [Minor] persistBoundaryRound's new operator lines have no assertNoGatesigCollision guard
- **BR-12** [Minor] previousReviewBoundary greps Review-Verdict: unanchored over the whole commit message
- **BR-13** [Minor] seedFromPlanGate mints ids by index instead of via nextIDSeq
