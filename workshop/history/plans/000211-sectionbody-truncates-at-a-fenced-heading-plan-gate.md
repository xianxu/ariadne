---
gate: plan-quality
issue: 211
id_prefix: PQ
rounds:
    - "n": 1
      timestamp: "2026-09-02T18:25:08-07:00"
      agent: claude
      findings:
        - id: PQ-1
          severity: Critical
          title: SectionBody inherits scanMarkdownLines' unterminated-fence policy, which hides Plan/Done-when on issue 211 itself
          detail: 'scanMarkdownLines (project/doc.go:118) treats an unterminated opener as fenced-to-EOF. Line 70 of the 211 issue file is a bare four-backtick run in prose that fenceMarker reads as a width-4 opener; simulating the scanner over that file leaves only `## Problem` and `## Spec` visible. checkPlan (structural.go:159) would then report "no ## Plan section found" (change-code refuses on its own issue) and the close.go:563 guard takes the m==nil branch and silently skips, reproducing the false PASS the issue exists to fix. `sdlc issue validate --issue 211` reports conforms today, so the Done-when prediction of zero corpus verdict flips is already falsified. State which unterminated-fence setting SectionBody takes (prose, per ARCH-SECURE degrade-visibly) and why.'
          family: fence-semantics-delta-unstated
          round: 1
        - id: PQ-2
          severity: Important
          title: PlanSectionRE has five production consumers plus a test, not the four the Spec names
          detail: 'Missing from the enumeration: CountPlanItems (plan.go:30), which feeds state.go:255 and carries the identical truncation bug, and close_test.go:440, which stops compiling when PlanSectionRE is deleted. Also stale after this change: the section.go:10-11 comment asserting checkPlan needs byte offsets. Run the grep and sweep the class, not the four sites the analysis happened to reach.'
          family: consumer-enumeration-incomplete
          round: 1
        - id: PQ-3
          severity: Important
          title: SplitFences re-base changes line-anchoring, indent rule and byte-exactness, not just the unterminated policy
          detail: SplitFences (structural.go:247) locates fences with an unanchored, indent-blind strings.Index, while fenceMarker requires line-start with at most three spaces of indent — so a 4-space-indented triple-backtick line flips from fenced to prose and migrate would start rewriting refs inside indented code blocks. SplitFences also guarantees byte-exact reassembly, which a line-visitor that drops fence lines cannot provide without deliberate reconstruction. Enumerate the divergence axes (line-anchoring, indent rule, tilde/width, unterminated policy, byte-exactness) and state each rebuilt consumer's setting.
          family: fence-semantics-delta-unstated
          round: 1
        - id: PQ-4
          severity: Important
          title: Test plan enumerates fence forms instead of naming functions plus one adversarial strategy line each
          detail: Name fenceMarker, scanMarkdownLines, SectionBody, stripCodeFences, SplitFences and the close plan-unchecked path, with one strategy line per risky function. The scanner wants a corpus-seeded property test over the 378 workshop markdown files rather than a hand-written form table — the enumerated five forms missed the form that is live in the corpus today. Fold the manual before/after corpus diff into that test so the invariant is mechanical.
          family: test-strategy-underspecified
          round: 1
      blocked: true
    - "n": 2
      timestamp: "2026-09-02T18:29:36-07:00"
      agent: claude
      dispose:
        - id: PQ-1
          disposition: addressed
          note: 'Policy is now an explicit per-consumer parameter (prose for SectionBody), pinned by a Done-when test; re-simulated the scanner over 378 files and #211 no longer self-truncates.'
          round: 2
        - id: PQ-2
          disposition: addressed
          note: Six-site table matches grep exactly, including CountPlanItems and close_test.go:440, plus both stale comments.
          round: 2
        - id: PQ-3
          disposition: addressed
          note: Six-axis divergence table verified against structural.go:247-269; line-anchoring flagged as an explicit migrate behaviour decision with a test either way.
          round: 2
        - id: PQ-4
          disposition: addressed
          note: Tests named per function with one strategy line each; scanner gets a corpus-seeded property test with the before/after diff folded in.
          round: 2
      blocked: false
    - "n": 3
      timestamp: "2026-09-02T18:34:14-07:00"
      agent: claude
      dispose:
        - id: PQ-1
          disposition: addressed
          note: Unterminated-fence policy is now an explicit per-consumer parameter with a table; SectionBody takes prose.
          round: 3
        - id: PQ-2
          disposition: addressed
          note: Spec table now enumerates six sites by grep; verified against the tree, it matches exactly.
          round: 3
        - id: PQ-3
          disposition: addressed
          note: Six divergence axes tabled, including line-anchoring and byte-exact reassembly, with the migrate impact called out.
          round: 3
        - id: PQ-4
          disposition: addressed
          note: Test row now names functions with one strategy line each rather than an enumerated case list.
          round: 3
      findings:
        - id: PQ-5
          severity: Minor
          title: Plan row 3 restates a stale consumer count ("four") that the Spec table has already superseded with six
          detail: |-
            2nd in family. Do not fix the instance: the rule is that the consumer set is
            enumerated in exactly one place, the grep-produced Spec table, and every other
            mention refers to it without restating a count. Measured prevalence 1 of 3
            references non-compliant (Plan row 3; Spec table and Estimate row 2 both comply).
          family: consumer-enumeration-incomplete
          round: 3
        - id: PQ-6
          severity: Minor
          title: Corpus-seeded property test does not separate stable invariants from an oracle that goes stale as issues are authored
          detail: |-
            "Concatenating visited + skipped reproduces the input" is content-independent;
            "no file loses a real section" over 406 live workshop files needs an oracle, and
            a golden captured today reddens on the next issue that quotes a fence. State that
            the corpus supplies inputs while assertions stay invariants, and name how the test
            in cmd/sdlc/internal/issue resolves the repo-root path.
          family: mutable-corpus-as-test-oracle
          round: 3
      blocked: false
content_hash: db5ada09a839ccd585173f6f2887a7a730c68ab6bec288a396a0ae88ad29482f
---

# Gate ledger — ariadne#211 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-09-02T18:25:08-07:00 (claude) — BLOCKED

### Raised

- **PQ-1** [Critical] `fence-semantics-delta-unstated` SectionBody inherits scanMarkdownLines' unterminated-fence policy, which hides Plan/Done-when on issue 211 itself
  scanMarkdownLines (project/doc.go:118) treats an unterminated opener as fenced-to-EOF. Line 70 of the 211 issue file is a bare four-backtick run in prose that fenceMarker reads as a width-4 opener; simulating the scanner over that file leaves only `## Problem` and `## Spec` visible. checkPlan (structural.go:159) would then report "no ## Plan section found" (change-code refuses on its own issue) and the close.go:563 guard takes the m==nil branch and silently skips, reproducing the false PASS the issue exists to fix. `sdlc issue validate --issue 211` reports conforms today, so the Done-when prediction of zero corpus verdict flips is already falsified. State which unterminated-fence setting SectionBody takes (prose, per ARCH-SECURE degrade-visibly) and why.
- **PQ-2** [Important] `consumer-enumeration-incomplete` PlanSectionRE has five production consumers plus a test, not the four the Spec names
  Missing from the enumeration: CountPlanItems (plan.go:30), which feeds state.go:255 and carries the identical truncation bug, and close_test.go:440, which stops compiling when PlanSectionRE is deleted. Also stale after this change: the section.go:10-11 comment asserting checkPlan needs byte offsets. Run the grep and sweep the class, not the four sites the analysis happened to reach.
- **PQ-3** [Important] `fence-semantics-delta-unstated` SplitFences re-base changes line-anchoring, indent rule and byte-exactness, not just the unterminated policy
  SplitFences (structural.go:247) locates fences with an unanchored, indent-blind strings.Index, while fenceMarker requires line-start with at most three spaces of indent — so a 4-space-indented triple-backtick line flips from fenced to prose and migrate would start rewriting refs inside indented code blocks. SplitFences also guarantees byte-exact reassembly, which a line-visitor that drops fence lines cannot provide without deliberate reconstruction. Enumerate the divergence axes (line-anchoring, indent rule, tilde/width, unterminated policy, byte-exactness) and state each rebuilt consumer's setting.
- **PQ-4** [Important] `test-strategy-underspecified` Test plan enumerates fence forms instead of naming functions plus one adversarial strategy line each
  Name fenceMarker, scanMarkdownLines, SectionBody, stripCodeFences, SplitFences and the close plan-unchecked path, with one strategy line per risky function. The scanner wants a corpus-seeded property test over the 378 workshop markdown files rather than a hand-written form table — the enumerated five forms missed the form that is live in the corpus today. Fold the manual before/after corpus diff into that test so the invariant is mechanical.

## Round 2 — 2026-09-02T18:29:36-07:00 (claude) — passed

### Disposed

- PQ-1 — addressed — Policy is now an explicit per-consumer parameter (prose for SectionBody), pinned by a Done-when test; re-simulated the scanner over 378 files and #211 no longer self-truncates.
- PQ-2 — addressed — Six-site table matches grep exactly, including CountPlanItems and close_test.go:440, plus both stale comments.
- PQ-3 — addressed — Six-axis divergence table verified against structural.go:247-269; line-anchoring flagged as an explicit migrate behaviour decision with a test either way.
- PQ-4 — addressed — Tests named per function with one strategy line each; scanner gets a corpus-seeded property test with the before/after diff folded in.

## Round 3 — 2026-09-02T18:34:14-07:00 (claude) — passed

### Disposed

- PQ-1 — addressed — Unterminated-fence policy is now an explicit per-consumer parameter with a table; SectionBody takes prose.
- PQ-2 — addressed — Spec table now enumerates six sites by grep; verified against the tree, it matches exactly.
- PQ-3 — addressed — Six divergence axes tabled, including line-anchoring and byte-exact reassembly, with the migrate impact called out.
- PQ-4 — addressed — Test row now names functions with one strategy line each rather than an enumerated case list.

### Raised

- **PQ-5** [Minor] `consumer-enumeration-incomplete` Plan row 3 restates a stale consumer count ("four") that the Spec table has already superseded with six
  2nd in family. Do not fix the instance: the rule is that the consumer set is
  enumerated in exactly one place, the grep-produced Spec table, and every other
  mention refers to it without restating a count. Measured prevalence 1 of 3
  references non-compliant (Plan row 3; Spec table and Estimate row 2 both comply).
- **PQ-6** [Minor] `mutable-corpus-as-test-oracle` Corpus-seeded property test does not separate stable invariants from an oracle that goes stale as issues are authored
  "Concatenating visited + skipped reproduces the input" is content-independent;
  "no file loses a real section" over 406 live workshop files needs an oracle, and
  a golden captured today reddens on the next issue that quotes a fence. State that
  the corpus supplies inputs while assertions stay invariants, and name how the test
  in cmd/sdlc/internal/issue resolves the repo-root path.

## Open findings

- **PQ-5** [Minor] `consumer-enumeration-incomplete` Plan row 3 restates a stale consumer count ("four") that the Spec table has already superseded with six
- **PQ-6** [Minor] `mutable-corpus-as-test-oracle` Corpus-seeded property test does not separate stable invariants from an oracle that goes stale as issues are authored
