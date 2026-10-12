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
    - "n": 2
      timestamp: "2026-10-12T16:47:27-07:00"
      dispatched: "2026-10-12T16:46:47-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: atlas/workflow/issue-lifecycle.md 3a and the intent's Conversation 12 (narrative and Verify) now read "`change-code` plan-quality gate → code"; TestLifecycleAtlasNamesReviewChain pins the atlas to the reviewChain constant; no "plan-quality → code" remains in the tree.
          round: 2
        - id: BR-2
          disposition: addressed
          note: spec-document-reviewer-prompt.md:9 reads "Dispatch (optional, once) after" and plan-document-reviewer-prompt.md:9 reads "Dispatch (optional, at most once per chunk) after"; no other "Dispatch after" line remains.
          round: 2
      recipe: small-diff-review
      isolation: checkout
      reviewed: dfe05cbd2be40066e4b17bcd0937287e9675fe34
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

## Round 2 — 2026-10-12T16:47:27-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — atlas/workflow/issue-lifecycle.md 3a and the intent's Conversation 12 (narrative and Verify) now read "`change-code` plan-quality gate → code"; TestLifecycleAtlasNamesReviewChain pins the atlas to the reviewChain constant; no "plan-quality → code" remains in the tree.
- BR-2 — addressed — spec-document-reviewer-prompt.md:9 reads "Dispatch (optional, once) after" and plan-document-reviewer-prompt.md:9 reads "Dispatch (optional, at most once per chunk) after"; no other "Dispatch after" line remains.

## Open findings

(none — every finding has been disposed)
