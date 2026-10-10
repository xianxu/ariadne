---
gate: boundary-review
issue: 316
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-09T21:22:47-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Important
          title: Section 5 decodes br in every resolved result, so it can corrupt verbatim X and D text on reject
          detail: 'review-convention.md:88-89 says any br in the result becomes a line break. Reference text (X and D) is quoted verbatim from one line, so a br in it is a literal that is already in the prose, as in markdown table cells. That affects reject-to-X and reject-to-D for the <X>{R}, ~D~, ~D~{N} and ~D~[N] rows, and accept-to-X for <X>[H] and <X>[H]{R}. Rejecting one of these would break a table. Fix: decode only commentary content ({R}, {N}, [N]), restore references verbatim, and say how a literal br in commentary is handled.'
          family: escape-decode-scope
          round: 1
        - id: BR-2
          severity: Important
          title: The xx-fix SKILL.md grammar and resolution steps leave out the single-line rule and br decoding
          detail: 'construct/local/fix/SKILL.md:58-93 (BNF and examples) and step 4 at about :124-140 (substitute D with N, apply R) restate the grammar and never mention single-line markers, writing replies with br, or decoding br on apply, yet xx-fix is the agent writer that motivated the issue. Other restatements I checked are acceptable as they stand: AGENTS.base.md:15 and construct/datatype/target.md:78 link to the full grammar, the docs/vision doc is historical, and scripts/docflow.sh counts markers line by line and is unaffected.'
          family: single-source-consumer-not-derived
          round: 1
        - id: BR-3
          severity: Minor
          title: The review-convention.md frontmatter still has updated 2026-07-22 after the 2026-10-09 revision
          family: frontmatter-updated-stamp
          round: 1
        - id: BR-4
          severity: Minor
          title: Section 3 describes parley.nvim rendering in a grammar doc that says rendering is tool-specific
          detail: Lines 72-74 conflict with the What this is NOT section at line 22. Word it as an example or move it to the Revisions rationale.
          family: grammar-doc-tool-specific-rendering
          round: 1
      recipe: small-diff-review
      blocked: true
    - "n": 2
      timestamp: "2026-10-09T21:25:32-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: 'review-convention.md:89-92 now decodes only R and N and keeps X and D verbatim; SKILL.md:406-408 matches. Still missing: how a literal br inside commentary is written (no escape exists). Worth one sentence.'
          round: 2
        - id: BR-2
          disposition: not-addressed
          note: 'The writing rule (SKILL.md:89-93) and the bulk-resolution decode (:406-408) were added, but step 4 at SKILL.md:137-146 (substitute D with N, apply R) still never says to decode br, and the One line paragraph at :89 does not mention decoding either. Fix: state the decode-on-apply rule once in the :89 paragraph so step 4 and bulk resolution both inherit it.'
          round: 2
        - id: BR-3
          disposition: not-addressed
          note: review-convention.md:6 is still updated 2026-07-22, but the newest Revisions entry is 2026-10-09.
          round: 2
        - id: BR-4
          disposition: not-addressed
          note: review-convention.md:72-75 still describes how parley.nvim renders markers in the grammar section, which conflicts with :22. Present it as an example or move it into the 2026-10-09 Revisions rationale.
          round: 2
      recipe: small-diff-review
      blocked: true
    - "n": 3
      timestamp: "2026-10-09T21:27:14-07:00"
      agent: claude
      dispose:
        - id: BR-2
          disposition: addressed
          note: SKILL.md:89-94 adds the one-line writer rule; step 4 (:137-139) and bulk resolution (:409-411) both decode br in turn text and keep X/D verbatim.
          round: 3
        - id: BR-3
          disposition: addressed
          note: review-convention.md:6 now reads updated 2026-10-09.
          round: 3
        - id: BR-4
          disposition: addressed
          note: 'Lines 72-74 now say an editor (example: thread view) with no parley-specific claim; the leftover Alt+q/render clause is raised separately as a Minor.'
          round: 3
      findings:
        - id: BR-5
          severity: Important
          title: Issue Spec/Plan still describe br encoding inside anchors and a single-file scope; no Revisions entry, empty Log
          detail: Spec says a newline in []/{}/<> is written br and section 5 decodes the resolved text, but the delivered design excludes anchors (BR-1). The Plan lists only review-convention.md, yet SKILL.md was changed (BR-2). Add a Revisions entry plus a plan item and fill in the 2026-10-09 Log heading.
          family: spec-plan-drift-unrecorded
          round: 3
        - id: BR-6
          severity: Important
          title: No way to write a literal br in turn text; accepting a robot edit to a table row containing br splits the row
          detail: '2nd in family escape-decode-scope. Rule: decoding runs only on encoder-produced text and must be invertible, so every token the decoder rewrites needs a literal spelling. Instances: review-convention.md:68-71 and :87-91; SKILL.md:89-92, :137-139, :409-411. Define the literal spelling (or state the limitation) once in the grammar and have both SKILL.md decode sites point to it.'
          family: escape-decode-scope
          round: 3
        - id: BR-7
          severity: Minor
          title: Grammar section 3 still prescribes editor behaviour (render legacy as broken; Alt+q multi-line selection refused)
          detail: '2nd in family. Rule: section 3 states only the grammar; editor and gesture behaviour belongs in the tool-labelled sections 4 and 5. Instance: review-convention.md:75-77. Move the Alt+q sentence into section 4 and drop render-as-broken.'
          family: grammar-doc-tool-specific-rendering
          round: 3
      recipe: small-diff-review
      blocked: true
    - "n": 4
      timestamp: "2026-10-09T21:29:42-07:00"
      agent: claude
      dispose:
        - id: BR-5
          disposition: addressed
          note: Spec now says turn-only decode, anchors verbatim, SKILL.md in scope; Plan item for SKILL.md; Revisions entry; Log filled (stray empty heading noted Minor).
          round: 4
        - id: BR-6
          disposition: not-addressed
          note: Grammar defines the literal \<br>, but SKILL.md step 4 (138-140) and bulk (410-412) restate decode as only <br>; following either literally still splits a table row on accept. Point both sites at the single One line definition and stop restating it (3rd in escape-decode-scope).
          round: 4
        - id: BR-7
          disposition: not-addressed
          note: review-convention.md:76-78 still says render legacy as broken and Alt+q multi-line refused, inside the grammar section 3.
          round: 4
      findings:
        - id: BR-8
          severity: Minor
          title: Issue Log carries an empty duplicate 2026-10-09 heading
          detail: 'The second ### 2026-10-09 heading under Log has no body; delete it. It is the only instance in the window.'
          family: spec-plan-drift-unrecorded
          round: 4
      recipe: small-diff-review
      blocked: false
---

# Gate ledger — ariadne#316 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-09T21:22:47-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `escape-decode-scope` Section 5 decodes br in every resolved result, so it can corrupt verbatim X and D text on reject
  review-convention.md:88-89 says any br in the result becomes a line break. Reference text (X and D) is quoted verbatim from one line, so a br in it is a literal that is already in the prose, as in markdown table cells. That affects reject-to-X and reject-to-D for the <X>{R}, ~D~, ~D~{N} and ~D~[N] rows, and accept-to-X for <X>[H] and <X>[H]{R}. Rejecting one of these would break a table. Fix: decode only commentary content ({R}, {N}, [N]), restore references verbatim, and say how a literal br in commentary is handled.
- **BR-2** [Important] `single-source-consumer-not-derived` The xx-fix SKILL.md grammar and resolution steps leave out the single-line rule and br decoding
  construct/local/fix/SKILL.md:58-93 (BNF and examples) and step 4 at about :124-140 (substitute D with N, apply R) restate the grammar and never mention single-line markers, writing replies with br, or decoding br on apply, yet xx-fix is the agent writer that motivated the issue. Other restatements I checked are acceptable as they stand: AGENTS.base.md:15 and construct/datatype/target.md:78 link to the full grammar, the docs/vision doc is historical, and scripts/docflow.sh counts markers line by line and is unaffected.
- **BR-3** [Minor] `frontmatter-updated-stamp` The review-convention.md frontmatter still has updated 2026-07-22 after the 2026-10-09 revision
- **BR-4** [Minor] `grammar-doc-tool-specific-rendering` Section 3 describes parley.nvim rendering in a grammar doc that says rendering is tool-specific
  Lines 72-74 conflict with the What this is NOT section at line 22. Word it as an example or move it to the Revisions rationale.

## Round 2 — 2026-10-09T21:25:32-07:00 (claude) — BLOCKED

### Disposed

- BR-1 — addressed — review-convention.md:89-92 now decodes only R and N and keeps X and D verbatim; SKILL.md:406-408 matches. Still missing: how a literal br inside commentary is written (no escape exists). Worth one sentence.
- BR-2 — not-addressed — The writing rule (SKILL.md:89-93) and the bulk-resolution decode (:406-408) were added, but step 4 at SKILL.md:137-146 (substitute D with N, apply R) still never says to decode br, and the One line paragraph at :89 does not mention decoding either. Fix: state the decode-on-apply rule once in the :89 paragraph so step 4 and bulk resolution both inherit it.
- BR-3 — not-addressed — review-convention.md:6 is still updated 2026-07-22, but the newest Revisions entry is 2026-10-09.
- BR-4 — not-addressed — review-convention.md:72-75 still describes how parley.nvim renders markers in the grammar section, which conflicts with :22. Present it as an example or move it into the 2026-10-09 Revisions rationale.

## Round 3 — 2026-10-09T21:27:14-07:00 (claude) — BLOCKED

### Disposed

- BR-2 — addressed — SKILL.md:89-94 adds the one-line writer rule; step 4 (:137-139) and bulk resolution (:409-411) both decode br in turn text and keep X/D verbatim.
- BR-3 — addressed — review-convention.md:6 now reads updated 2026-10-09.
- BR-4 — addressed — Lines 72-74 now say an editor (example: thread view) with no parley-specific claim; the leftover Alt+q/render clause is raised separately as a Minor.

### Raised

- **BR-5** [Important] `spec-plan-drift-unrecorded` Issue Spec/Plan still describe br encoding inside anchors and a single-file scope; no Revisions entry, empty Log
  Spec says a newline in []/{}/<> is written br and section 5 decodes the resolved text, but the delivered design excludes anchors (BR-1). The Plan lists only review-convention.md, yet SKILL.md was changed (BR-2). Add a Revisions entry plus a plan item and fill in the 2026-10-09 Log heading.
- **BR-6** [Important] `escape-decode-scope` No way to write a literal br in turn text; accepting a robot edit to a table row containing br splits the row
  2nd in family escape-decode-scope. Rule: decoding runs only on encoder-produced text and must be invertible, so every token the decoder rewrites needs a literal spelling. Instances: review-convention.md:68-71 and :87-91; SKILL.md:89-92, :137-139, :409-411. Define the literal spelling (or state the limitation) once in the grammar and have both SKILL.md decode sites point to it.
- **BR-7** [Minor] `grammar-doc-tool-specific-rendering` Grammar section 3 still prescribes editor behaviour (render legacy as broken; Alt+q multi-line selection refused)
  2nd in family. Rule: section 3 states only the grammar; editor and gesture behaviour belongs in the tool-labelled sections 4 and 5. Instance: review-convention.md:75-77. Move the Alt+q sentence into section 4 and drop render-as-broken.

## Round 4 — 2026-10-09T21:29:42-07:00 (claude) — passed

### Disposed

- BR-5 — addressed — Spec now says turn-only decode, anchors verbatim, SKILL.md in scope; Plan item for SKILL.md; Revisions entry; Log filled (stray empty heading noted Minor).
- BR-6 — not-addressed — Grammar defines the literal \<br>, but SKILL.md step 4 (138-140) and bulk (410-412) restate decode as only <br>; following either literally still splits a table row on accept. Point both sites at the single One line definition and stop restating it (3rd in escape-decode-scope).
- BR-7 — not-addressed — review-convention.md:76-78 still says render legacy as broken and Alt+q multi-line refused, inside the grammar section 3.

### Raised

- **BR-8** [Minor] `spec-plan-drift-unrecorded` Issue Log carries an empty duplicate 2026-10-09 heading
  The second ### 2026-10-09 heading under Log has no body; delete it. It is the only instance in the window.

## Open findings

- **BR-6** [Important] `escape-decode-scope` No way to write a literal br in turn text; accepting a robot edit to a table row containing br splits the row
- **BR-7** [Minor] `grammar-doc-tool-specific-rendering` Grammar section 3 still prescribes editor behaviour (render legacy as broken; Alt+q multi-line selection refused)
- **BR-8** [Minor] `spec-plan-drift-unrecorded` Issue Log carries an empty duplicate 2026-10-09 heading
