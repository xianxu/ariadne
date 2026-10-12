---
gate: boundary-review
issue: 336
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-12T08:50:26-07:00"
      dispatched: "2026-10-12T08:49:35-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Important
          title: 'atlas/workflow/sandbox.md agent-rule subsection contradicts the #336 measurements and §15'
          detail: 'Lines 16-22 still say a pipe is only a problem when piped to a command the list doesn''t cover, and that the redirect form was untested. They also omit the 2>&1 exception and the backtick, $((, <( and >( substitution openers. Line 24 of the same page and AGENTS.base.md §15 state the measured rule: any pipe, ; chain or file redirect breaks the exemption, and only 2>&1 passes. Rewrite the subsection to match §15, or reduce it to a pointer.'
          family: single-source-restatement-drift
          round: 1
        - id: BR-2
          severity: Minor
          title: construct/local/couch/SKILL.md partially restates §15 and omits cases
          detail: 'The summary leaves out ''nothing before it'' (cd x &&, a full path), the $(( and >( openers, the 2>&1 exception, and sdlc. Make it a pure pointer to §15 or make it complete. Instances of this family in the window: sandbox.md lines 16-22 and SKILL.md lines 19-22.'
          family: single-source-restatement-drift
          round: 1
      recipe: small-diff-review
      isolation: fenced
      blocked: true
    - "n": 2
      timestamp: "2026-10-12T08:51:58-07:00"
      dispatched: "2026-10-12T08:51:23-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: sandbox.md:17-24 now points at §15; the 10-12 measurement log matches §15 (any pipe, ; or file redirect breaks the exemption, 2>&1 passes, all five openers); the untested-redirect caveat is gone.
          round: 2
        - id: BR-2
          disposition: addressed
          note: SKILL.md:19-21 is now a pure pointer to §15 ("it is not restated here"), with no partial summary left.
          round: 2
      recipe: small-diff-review
      isolation: fenced
      reviewed: b8c75fb0e4dc45ec7e4c93204c7c7d38b8c3eb15
      blocked: false
---

# Gate ledger — ariadne#336 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-12T08:50:26-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `single-source-restatement-drift` atlas/workflow/sandbox.md agent-rule subsection contradicts the #336 measurements and §15
  Lines 16-22 still say a pipe is only a problem when piped to a command the list doesn't cover, and that the redirect form was untested. They also omit the 2>&1 exception and the backtick, $((, <( and >( substitution openers. Line 24 of the same page and AGENTS.base.md §15 state the measured rule: any pipe, ; chain or file redirect breaks the exemption, and only 2>&1 passes. Rewrite the subsection to match §15, or reduce it to a pointer.
- **BR-2** [Minor] `single-source-restatement-drift` construct/local/couch/SKILL.md partially restates §15 and omits cases
  The summary leaves out 'nothing before it' (cd x &&, a full path), the $(( and >( openers, the 2>&1 exception, and sdlc. Make it a pure pointer to §15 or make it complete. Instances of this family in the window: sandbox.md lines 16-22 and SKILL.md lines 19-22.

## Round 2 — 2026-10-12T08:51:58-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — sandbox.md:17-24 now points at §15; the 10-12 measurement log matches §15 (any pipe, ; or file redirect breaks the exemption, 2>&1 passes, all five openers); the untested-redirect caveat is gone.
- BR-2 — addressed — SKILL.md:19-21 is now a pure pointer to §15 ("it is not restated here"), with no partial summary left.

## Open findings

(none — every finding has been disposed)
