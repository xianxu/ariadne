---
gate: boundary-review
issue: 228
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-12T16:45:58-07:00"
      dispatched: "2026-10-12T16:45:02-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Minor
          title: The review chain is worded differently in the atlas and intent than in the two skills
          detail: |-
            Both skills (and the reviewChain test constant) say "`change-code` plan-quality gate → code".
            atlas/workflow/issue-lifecycle.md step 3a and construct/intents/superpowers.md Conversation 12
            (narrative and Verify clause) say "`change-code` plan-quality → code". Only the skills are
            pinned by the test. Use one wording everywhere (ARCH-DRY).
          family: review-chain-single-wording
          round: 1
        - id: BR-2
          severity: Minor
          title: The reviewer templates' "Dispatch after" lines still read as if every dispatch is expected
          detail: |-
            plan-document-reviewer-prompt.md:9 "Dispatch after: Each plan chunk is written" implies a review
            after every chunk, while SKILL.md makes it optional, at most once per chunk. The same pattern is
            at spec-document-reviewer-prompt.md:9. Reword both, e.g. "Dispatch (optional) after …".
          family: advisory-reviewer-mandatory-wording
          round: 1
      recipe: small-diff-review
      isolation: checkout
      reviewed: ae1e0785ecaa0a6b7978544bfdcb9696ae63cd17
      blocked: false
---

# Gate ledger — ariadne#228 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-12T16:45:58-07:00 (claude) — passed

### Raised

- **BR-1** [Minor] `review-chain-single-wording` The review chain is worded differently in the atlas and intent than in the two skills
  Both skills (and the reviewChain test constant) say "`change-code` plan-quality gate → code".
  atlas/workflow/issue-lifecycle.md step 3a and construct/intents/superpowers.md Conversation 12
  (narrative and Verify clause) say "`change-code` plan-quality → code". Only the skills are
  pinned by the test. Use one wording everywhere (ARCH-DRY).
- **BR-2** [Minor] `advisory-reviewer-mandatory-wording` The reviewer templates' "Dispatch after" lines still read as if every dispatch is expected
  plan-document-reviewer-prompt.md:9 "Dispatch after: Each plan chunk is written" implies a review
  after every chunk, while SKILL.md makes it optional, at most once per chunk. The same pattern is
  at spec-document-reviewer-prompt.md:9. Reword both, e.g. "Dispatch (optional) after …".

## Open findings

- **BR-1** [Minor] `review-chain-single-wording` The review chain is worded differently in the atlas and intent than in the two skills
- **BR-2** [Minor] `advisory-reviewer-mandatory-wording` The reviewer templates' "Dispatch after" lines still read as if every dispatch is expected
