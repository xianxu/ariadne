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
    - "n": 4
      timestamp: "2026-08-20T21:52:24-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: Read under the lock beside captureCloseReviewSnapshot and carried on PriorFindings; TestCloseCommand_LiveReviewSeesPriorFindings drives the real close command.
          round: 4
        - id: BR-2
          disposition: addressed
          note: blockOnLedgerFailure returns Block true with a Reason; the caller branches on empty OpenBlocking so the message fits the actual failure.
          round: 4
        - id: BR-3
          disposition: addressed
          note: boundaryledger.go:180 stamps Blocked before the write. Residual data only - the existing ledger still records round 3 as blocked false, which no longer matches its Critical finding.
          round: 4
        - id: BR-4
          disposition: addressed
          note: Lines 22, 73, 75-88 and the code-map row all corrected; ledger-landscape.md:77-79 too. A newly-stale section and index.md:14 are raised separately below.
          round: 4
        - id: BR-5
          disposition: addressed
          note: close.md and milestone-close.md now cover the ledger artifact, the passing-verdict refusal, dispose-before-raising, WF_BOUNDARY_ROUND_CAP, and the bypass table row.
          round: 4
        - id: BR-6
          disposition: addressed
          note: Verified - go run ./cmd/vocabulary check --output construct/generated/vocabulary exits 0.
          round: 4
        - id: BR-7
          disposition: addressed
          note: D4's heading now reads "warns"; a Revisions section with four entries exists. The Core-concepts half is raised separately as a Minor.
          round: 4
        - id: BR-8
          disposition: addressed
          note: Seed round carries NoCap and Decide uses CountedRounds; TestCountedRounds_ExcludesRoundsThatConsumedNoReview pins it.
          round: 4
        - id: BR-9
          disposition: addressed
          note: protocol_error and no_cap now distinguish the two cases in the frontmatter. Residual - Agent is still empty on a dispatch error even where opts.Agent is resolved.
          round: 4
        - id: BR-10
          disposition: not-addressed
          note: Fixed for the boundary gate, but the plan gate half named in the finding still holds - changecode.go:525-531 discards ApplyChecked's `applied` and hand-builds a round with no dispositions at all, so one typo still nullifies every valid disposal there.
          round: 4
        - id: BR-11
          disposition: addressed
          note: TestBoundaryGateOperatorLines_NoGatesigCollision covers all four new lines through the derived guard.
          round: 4
        - id: BR-12
          disposition: addressed
          note: Anchored to ^Review-Verdict:. Verified empirically that git's --grep matches line-wise, so the three real trailer commits still resolve - a tightening, not a regression.
          round: 4
        - id: BR-13
          disposition: addressed
          note: Ids now come from AssignIDsAt. The empty-ledger precondition is still unpinned by a test, but the code no longer depends on it.
          round: 4
      findings:
        - id: BR-14
          severity: Important
          title: The gate-ledger refusal, --no-ledger, and blockOnLedgerFailure have no test at any level
          detail: |-
            grep finalizeBoundaryReview cmd/sdlc/*_test.go returns nothing. Untested - a
            finalizing verdict plus an open blocking finding refuses without running
            applyClose; the error names the findings; --no-ledger waives exactly that
            refusal; the unusable-ledger branch prints Reason instead of the open-findings
            message. This is D4 and Task 2.2 Step 5, the behavior close.md advertises as
            surprising, and it is the same coverage shape that let BR-1 ship. The new
            GateCatalog no-ledger Ack/Refusal patterns are also never matched against real
            emitted output.
          round: 4
        - id: BR-15
          severity: Important
          title: NoCap does not implement the case it names as its motivation; doc, commit message and issue Log disagree with the code
          detail: |-
            gatestate/ledger.go:74-89 claims three NoCap kinds; only two assignment sites
            exist (boundaryledger.go:120 and :163). "A round persisted before a non-review
            refusal" is never set. Worse, the cited motivation - two reviews killed by host
            sleep - lands as ProtocolError "no valid findings block" with NoCap false and
            still counts. Proof - this issue's own close-gate ledger rounds 1 and 2 carry
            no no_cap key, so M2 is at 3 counted rounds and this round trips the cap. The
            issue's Log still records the question as "Not decided here" while the commit
            message says it was settled. Make all three agree.
          round: 4
        - id: BR-16
          severity: Important
          title: atlas gate-state.md now asserts the superseded cap rule, and NoCap/CountedRounds is undocumented
          detail: |-
            Not BR-4 - those lines are fixed. This staleness was created by the fix commit.
            gate-state.md:105-111 "Protocol misses still count" grounds the cap in
            len(Rounds), but Decide now uses CountedRounds and a never-dispatched round is
            persisted with protocol_error and does NOT count. No atlas file mentions
            no_cap, CountedRounds, or "the cap counts review cycles" - a new persisted YAML
            field with no map entry. Separately atlas/index.md:14 still calls the
            boundary-review carry-forward consumer "intended (#183)" after #194 delivered
            it. Third instance of the family this diff's own lessons.md entry names.
          round: 4
        - id: BR-17
          severity: Minor
          title: Round.Forced is never stamped on a boundary round, unlike the plan gate
          detail: |-
            boundaryledger.go:180 mirrors changecode.go:537 but not :538, so a --force or
            --no-ledger bypass at the boundary leaves no durable record - the same field
            that feeds closeMetrics' "N forced" for the plan gate. ARCH-DRY - the two
            persist tails diverge again at the one line nobody notices.
          round: 4
        - id: BR-18
          severity: Minor
          title: The plan's Core-concepts tables never gained M2's entities, and Revisions omits two shared-gatestate behavior changes
          detail: |-
            Every existing row verifies clean against the filesystem, so there is no
            table/code contradiction - but gateLedgerKind, readGateLedger/writeGateLedger,
            seedFromPlanGate, persistBoundaryRound, boundaryPriorFindings,
            blockOnLedgerFailure, roundCapFromEnvVar, AssignIDsAt, CountedRounds and
            Round.NoCap are all absent, so the table stops being the greppable index it
            exists to be. Revisions also omits the no_cap schema field and ApplyChecked's
            per-disposition semantics, both of which change code plan-quality also runs.
          round: 4
        - id: BR-19
          severity: Minor
          title: A seeded BoundaryAll finding's disposal is boundary-scoped, so it re-opens at every later boundary
          detail: |-
            FilterBoundary retains the BoundaryAll seed round at every boundary but drops
            the M1 round that disposed it, so OpenFindings shows the seed open again at M2
            and at the whole-issue close. Cheap in practice (one dispose entry per
            boundary, cleared in the same round) and arguably intended, but D5's wording
            says "until disposed" where the code means "until disposed at each boundary".
            Decide it explicitly and say so in gate-state.md.
          round: 4
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

## Round 4 — 2026-08-20T21:52:24-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — Read under the lock beside captureCloseReviewSnapshot and carried on PriorFindings; TestCloseCommand_LiveReviewSeesPriorFindings drives the real close command.
- BR-2 — addressed — blockOnLedgerFailure returns Block true with a Reason; the caller branches on empty OpenBlocking so the message fits the actual failure.
- BR-3 — addressed — boundaryledger.go:180 stamps Blocked before the write. Residual data only - the existing ledger still records round 3 as blocked false, which no longer matches its Critical finding.
- BR-4 — addressed — Lines 22, 73, 75-88 and the code-map row all corrected; ledger-landscape.md:77-79 too. A newly-stale section and index.md:14 are raised separately below.
- BR-5 — addressed — close.md and milestone-close.md now cover the ledger artifact, the passing-verdict refusal, dispose-before-raising, WF_BOUNDARY_ROUND_CAP, and the bypass table row.
- BR-6 — addressed — Verified - go run ./cmd/vocabulary check --output construct/generated/vocabulary exits 0.
- BR-7 — addressed — D4's heading now reads "warns"; a Revisions section with four entries exists. The Core-concepts half is raised separately as a Minor.
- BR-8 — addressed — Seed round carries NoCap and Decide uses CountedRounds; TestCountedRounds_ExcludesRoundsThatConsumedNoReview pins it.
- BR-9 — addressed — protocol_error and no_cap now distinguish the two cases in the frontmatter. Residual - Agent is still empty on a dispatch error even where opts.Agent is resolved.
- BR-10 — not-addressed — Fixed for the boundary gate, but the plan gate half named in the finding still holds - changecode.go:525-531 discards ApplyChecked's `applied` and hand-builds a round with no dispositions at all, so one typo still nullifies every valid disposal there.
- BR-11 — addressed — TestBoundaryGateOperatorLines_NoGatesigCollision covers all four new lines through the derived guard.
- BR-12 — addressed — Anchored to ^Review-Verdict:. Verified empirically that git's --grep matches line-wise, so the three real trailer commits still resolve - a tightening, not a regression.
- BR-13 — addressed — Ids now come from AssignIDsAt. The empty-ledger precondition is still unpinned by a test, but the code no longer depends on it.

### Raised

- **BR-14** [Important] The gate-ledger refusal, --no-ledger, and blockOnLedgerFailure have no test at any level
  grep finalizeBoundaryReview cmd/sdlc/*_test.go returns nothing. Untested - a
  finalizing verdict plus an open blocking finding refuses without running
  applyClose; the error names the findings; --no-ledger waives exactly that
  refusal; the unusable-ledger branch prints Reason instead of the open-findings
  message. This is D4 and Task 2.2 Step 5, the behavior close.md advertises as
  surprising, and it is the same coverage shape that let BR-1 ship. The new
  GateCatalog no-ledger Ack/Refusal patterns are also never matched against real
  emitted output.
- **BR-15** [Important] NoCap does not implement the case it names as its motivation; doc, commit message and issue Log disagree with the code
  gatestate/ledger.go:74-89 claims three NoCap kinds; only two assignment sites
  exist (boundaryledger.go:120 and :163). "A round persisted before a non-review
  refusal" is never set. Worse, the cited motivation - two reviews killed by host
  sleep - lands as ProtocolError "no valid findings block" with NoCap false and
  still counts. Proof - this issue's own close-gate ledger rounds 1 and 2 carry
  no no_cap key, so M2 is at 3 counted rounds and this round trips the cap. The
  issue's Log still records the question as "Not decided here" while the commit
  message says it was settled. Make all three agree.
- **BR-16** [Important] atlas gate-state.md now asserts the superseded cap rule, and NoCap/CountedRounds is undocumented
  Not BR-4 - those lines are fixed. This staleness was created by the fix commit.
  gate-state.md:105-111 "Protocol misses still count" grounds the cap in
  len(Rounds), but Decide now uses CountedRounds and a never-dispatched round is
  persisted with protocol_error and does NOT count. No atlas file mentions
  no_cap, CountedRounds, or "the cap counts review cycles" - a new persisted YAML
  field with no map entry. Separately atlas/index.md:14 still calls the
  boundary-review carry-forward consumer "intended (#183)" after #194 delivered
  it. Third instance of the family this diff's own lessons.md entry names.
- **BR-17** [Minor] Round.Forced is never stamped on a boundary round, unlike the plan gate
  boundaryledger.go:180 mirrors changecode.go:537 but not :538, so a --force or
  --no-ledger bypass at the boundary leaves no durable record - the same field
  that feeds closeMetrics' "N forced" for the plan gate. ARCH-DRY - the two
  persist tails diverge again at the one line nobody notices.
- **BR-18** [Minor] The plan's Core-concepts tables never gained M2's entities, and Revisions omits two shared-gatestate behavior changes
  Every existing row verifies clean against the filesystem, so there is no
  table/code contradiction - but gateLedgerKind, readGateLedger/writeGateLedger,
  seedFromPlanGate, persistBoundaryRound, boundaryPriorFindings,
  blockOnLedgerFailure, roundCapFromEnvVar, AssignIDsAt, CountedRounds and
  Round.NoCap are all absent, so the table stops being the greppable index it
  exists to be. Revisions also omits the no_cap schema field and ApplyChecked's
  per-disposition semantics, both of which change code plan-quality also runs.
- **BR-19** [Minor] A seeded BoundaryAll finding's disposal is boundary-scoped, so it re-opens at every later boundary
  FilterBoundary retains the BoundaryAll seed round at every boundary but drops
  the M1 round that disposed it, so OpenFindings shows the seed open again at M2
  and at the whole-issue close. Cheap in practice (one dispose entry per
  boundary, cleared in the same round) and arguably intended, but D5's wording
  says "until disposed" where the code means "until disposed at each boundary".
  Decide it explicitly and say so in gate-state.md.

## Open findings

- **BR-10** [Minor] One bad disposition id nullifies a whole round's valid dispositions
- **BR-14** [Important] The gate-ledger refusal, --no-ledger, and blockOnLedgerFailure have no test at any level
- **BR-15** [Important] NoCap does not implement the case it names as its motivation; doc, commit message and issue Log disagree with the code
- **BR-16** [Important] atlas gate-state.md now asserts the superseded cap rule, and NoCap/CountedRounds is undocumented
- **BR-17** [Minor] Round.Forced is never stamped on a boundary round, unlike the plan gate
- **BR-18** [Minor] The plan's Core-concepts tables never gained M2's entities, and Revisions omits two shared-gatestate behavior changes
- **BR-19** [Minor] A seeded BoundaryAll finding's disposal is boundary-scoped, so it re-opens at every later boundary
