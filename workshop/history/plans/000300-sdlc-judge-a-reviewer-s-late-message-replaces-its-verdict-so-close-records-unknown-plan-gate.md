---
gate: plan-quality
issue: 300
id_prefix: PQ
rounds:
    - "n": 1
      timestamp: "2026-10-09T20:22:37-07:00"
      agent: claude
      findings:
        - id: PQ-1
          severity: Important
          title: Dispatch's return contract and the owner of the D3 retry decision are unstated
          detail: Dispatch returns a string and leaves classification to its three callers (milestoneclose.go:663-673, judge.go:187, planningreview.go:107,116). D3 needs a verdict predicate inside Dispatch, and RunVerdict returning findings adds a judge->gatestate import. The plan should state whether Dispatch returns a Run or joined text, who decides the retry and with which predicate per caller, and why per-message parsing beats the existing last-block-wins parsers (classify.go:182, gatestate/parse.go:34) per ARCH-DRY.
          family: unstated-seam-change
          round: 1
        - id: PQ-2
          severity: Important
          title: D6 says plan-quality already blocks protocol-error rounds; it falls back to the verdict token
          detail: changecode.go:673-677 sets blocked = classifyFallback(...) != nil, so a fenceless CLEAN round is unblocked today. stampAndPersist (gatepersist.go:50) is shared and feeds PassesUnchanged (ledger.go:400). Scope D6 to the boundary path, where boundaryledger.go:167's Blocked:true is overwritten by d.Block, or state that plan-quality's fallback changes too.
          family: unbacked-existing-behavior-claim
          round: 1
        - id: PQ-3
          severity: Important
          title: D5 matches API-failure strings anywhere in the output, so reviews quoting them read as "did not run"
          detail: '#300''s own diff will contain those strings, so codex and gemini rounds of this issue would be classified "review did not run". Limit detection to the stream result''s is_error and stderr, and on the text path only when no verdict parsed.'
          family: classifier-scope
          round: 1
        - id: PQ-4
          severity: Minor
          title: The plan doesn't say whether the sized timeout applies per attempt or across the D3 retry
          detail: The worst case could double to 4h (ARCH-CONSTRAINTS).
          family: operating-envelope-unstated
          round: 1
        - id: PQ-5
          severity: Minor
          title: The existing live conformance test breaks under stream-json, and a partial-stream fallback discards parsed messages
          detail: live_stream_conformance_test.go:47 asserts stdout == STREAM_OK for claude. The text fallback on a truncated stream loses messages that already parsed, and JSON-escaped blocks don't parse as text.
          family: stream-fallback-lossy
          round: 1
        - id: PQ-6
          severity: Minor
          title: Tasks 2 and 6 list test cases; replace with one strategy line per risky function
          family: test-prose-enumeration
          round: 1
      blocked: true
    - "n": 2
      timestamp: "2026-10-09T20:24:34-07:00"
      agent: claude
      dispose:
        - id: PQ-1
          disposition: addressed
          note: Dispatch keeps a string contract (joined messages); Dispatch owns the retry; ParseVerdictToken goes last-match. The predicate it names is wrong for plan-quality — see the new finding.
          round: 2
        - id: PQ-2
          disposition: addressed
          note: 'Scoped to the boundary path. Minor residue: stampAndPersist overwrites Blocked from d.Block (gatepersist.go:50), so the plan should say whether the returned Decision blocks or only the ledger field changes.'
          round: 2
        - id: PQ-3
          disposition: addressed
          round: 2
        - id: PQ-4
          disposition: addressed
          round: 2
        - id: PQ-5
          disposition: addressed
          round: 2
        - id: PQ-6
          disposition: addressed
          round: 2
      findings:
        - id: PQ-7
          severity: Important
          title: D3's retry predicate ParseVerdict(text) != VerdictUnknown is always Unknown for plan-quality CLEAN/INFO/FAILURE, so every plan-quality round would be retried
          detail: 'verdictFor (classify.go:148-153) accepts only SHIP/FIX-THEN-SHIP/REWORK (verdict.cue:16-17), and classify.go:215 notes that CLEAN/INFO/FAILURE fall through. 2nd finding in this family. The rule for the class: every existing-behavior claim a decision rests on carries a file:line. For this instance, the predicate must match the recipe''s verdict set (a predicate passed in DispatchOptions, or ParseVerdictToken ok or ParseVerdictBlock ok).'
          family: unbacked-existing-behavior-claim
          round: 2
      blocked: true
    - "n": 3
      timestamp: "2026-10-09T20:25:18-07:00"
      agent: claude
      dispose:
        - id: PQ-7
          disposition: addressed
          note: D3 now uses HasVerdict, which is ParseVerdictBlock ok or ParseVerdictToken ok. The token pattern at classify.go:56 covers CLEAN/INFO/FAILURE, and every claim in the decisions is now backed by a file:line.
          round: 3
      blocked: false
    - "n": 4
      timestamp: "2026-10-09T20:25:57-07:00"
      agent: claude
      blocked: false
      protocol_error: no valid findings block
content_hash: c7df715bd60c7bb5f349fd703643be7d72cbaf13ef463c58d7d2f2c1599b2793
---

# Gate ledger — ariadne#300 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-09T20:22:37-07:00 (claude) — BLOCKED

### Raised

- **PQ-1** [Important] `unstated-seam-change` Dispatch's return contract and the owner of the D3 retry decision are unstated
  Dispatch returns a string and leaves classification to its three callers (milestoneclose.go:663-673, judge.go:187, planningreview.go:107,116). D3 needs a verdict predicate inside Dispatch, and RunVerdict returning findings adds a judge->gatestate import. The plan should state whether Dispatch returns a Run or joined text, who decides the retry and with which predicate per caller, and why per-message parsing beats the existing last-block-wins parsers (classify.go:182, gatestate/parse.go:34) per ARCH-DRY.
- **PQ-2** [Important] `unbacked-existing-behavior-claim` D6 says plan-quality already blocks protocol-error rounds; it falls back to the verdict token
  changecode.go:673-677 sets blocked = classifyFallback(...) != nil, so a fenceless CLEAN round is unblocked today. stampAndPersist (gatepersist.go:50) is shared and feeds PassesUnchanged (ledger.go:400). Scope D6 to the boundary path, where boundaryledger.go:167's Blocked:true is overwritten by d.Block, or state that plan-quality's fallback changes too.
- **PQ-3** [Important] `classifier-scope` D5 matches API-failure strings anywhere in the output, so reviews quoting them read as "did not run"
  #300's own diff will contain those strings, so codex and gemini rounds of this issue would be classified "review did not run". Limit detection to the stream result's is_error and stderr, and on the text path only when no verdict parsed.
- **PQ-4** [Minor] `operating-envelope-unstated` The plan doesn't say whether the sized timeout applies per attempt or across the D3 retry
  The worst case could double to 4h (ARCH-CONSTRAINTS).
- **PQ-5** [Minor] `stream-fallback-lossy` The existing live conformance test breaks under stream-json, and a partial-stream fallback discards parsed messages
  live_stream_conformance_test.go:47 asserts stdout == STREAM_OK for claude. The text fallback on a truncated stream loses messages that already parsed, and JSON-escaped blocks don't parse as text.
- **PQ-6** [Minor] `test-prose-enumeration` Tasks 2 and 6 list test cases; replace with one strategy line per risky function

## Round 2 — 2026-10-09T20:24:34-07:00 (claude) — BLOCKED

### Disposed

- PQ-1 — addressed — Dispatch keeps a string contract (joined messages); Dispatch owns the retry; ParseVerdictToken goes last-match. The predicate it names is wrong for plan-quality — see the new finding.
- PQ-2 — addressed — Scoped to the boundary path. Minor residue: stampAndPersist overwrites Blocked from d.Block (gatepersist.go:50), so the plan should say whether the returned Decision blocks or only the ledger field changes.
- PQ-3 — addressed
- PQ-4 — addressed
- PQ-5 — addressed
- PQ-6 — addressed

### Raised

- **PQ-7** [Important] `unbacked-existing-behavior-claim` D3's retry predicate ParseVerdict(text) != VerdictUnknown is always Unknown for plan-quality CLEAN/INFO/FAILURE, so every plan-quality round would be retried
  verdictFor (classify.go:148-153) accepts only SHIP/FIX-THEN-SHIP/REWORK (verdict.cue:16-17), and classify.go:215 notes that CLEAN/INFO/FAILURE fall through. 2nd finding in this family. The rule for the class: every existing-behavior claim a decision rests on carries a file:line. For this instance, the predicate must match the recipe's verdict set (a predicate passed in DispatchOptions, or ParseVerdictToken ok or ParseVerdictBlock ok).

## Round 3 — 2026-10-09T20:25:18-07:00 (claude) — passed

### Disposed

- PQ-7 — addressed — D3 now uses HasVerdict, which is ParseVerdictBlock ok or ParseVerdictToken ok. The token pattern at classify.go:56 covers CLEAN/INFO/FAILURE, and every claim in the decisions is now backed by a file:line.

## Round 4 — 2026-10-09T20:25:57-07:00 (claude) — passed

**Protocol error:** no valid findings block — this round contributed no findings.

## Open findings

(none — every finding has been disposed)
