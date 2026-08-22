---
gate: boundary-review
issue: 203
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-08-22T10:34:51-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Important
          title: Doc-surface class enumerated by hand; three unrouted findings surfaces remain, two of them doc twins of code sites this diff routed
          detail: |-
            A paragraph-level scan of cmd/sdlc/helptext/*.md for findings-word + directive-verb finds
            close.md:13 (POST-VERDICT PROTOCOL, "fix the findings NOW" — twin of the routed close.go:1809),
            close.md:59 ("Fix the findings, or have the next review dispose them" — twin of the routed
            close.go:1194), and milestone-close.md:38 ("fix the findings before committing"; the file was
            not touched at all). Spec bullet 2 says EVERY surface routes. The Go class got a mechanical
            scan that found 8 where a hand pass found 4; the doc class got a hand pass. ARCH-PURPOSE.
          family: hand-enumerated-class
          round: 1
        - id: BR-2
          severity: Important
          title: Enumeration guard's class signature is narrower than the class it claims; four synthetic ninth sites ship green
          detail: |-
            Verified by appending unrouted sites to cmd/sdlc/state.go and re-running the guard. Blind to
            (a) a message split across adjacent string literals — the tree's own prevailing style, e.g.
            close.go:1194 is 3 literals and changecode.go:554 became 2 in this diff; (b) a package-level
            const, since pass 2 walks only FuncDecl bodies; (c) a directive verb outside "fix "/"address "/
            "review above"; and (d) a second unrouted line inside an already-routing builder func, because
            pass 2 is function-granular — the exact thing pass 1's own comment declares insufficient.
            atlas/workflow/gate-state.md:127 and Done-when 4 claim "a ninth refusal site cannot ship
            unrouted". Fix: concatenate a call's literals before matching; make pass 2 count-based.
          family: guard-narrower-than-claimed-class
          round: 1
        - id: BR-3
          severity: Minor
          title: New ARCH-PURPOSE citations in helptext are unguarded, unlike the routing line itself
          detail: |-
            helptext/close.md:79 and helptext/change-code.md:37 cite ARCH-PURPOSE with no drift guard;
            TestFixTheClassLine_RoutesToArchPrinciples covers only fixTheClassLine, and
            TestArchitecture_NarrativeRoutesToArchPrinciples reads AGENTS.md only.
          family: guard-narrower-than-claimed-class
          round: 1
        - id: BR-4
          severity: Minor
          title: No test asserts a real gate path's stderr actually contains the routing line
          detail: |-
            Coverage is source-level plus string-level with nothing joining them. changecode_test.go:364
            already drives the blocking runPlanQualityJudge path with a stub judge; swapping ioDiscard()
            for a bytes.Buffer and asserting on "ARCH-PURPOSE" is a two-line end-to-end pin.
          family: coverage-stops-at-the-seam
          round: 1
        - id: BR-5
          severity: Minor
          title: The "\n  " join prefix is hand-written at six of the eight call sites
          detail: |-
            changecode.go:554,572,722, close.go:1238, judge.go:172, milestoneclose.go:626 each spell the
            same newline+indent. A fixTheClassNote() returning the pre-joined form would leave only the
            two genuinely different indents (close.go:1194, close.go:1809).
          family: repeated-join-prefix
          round: 1
        - id: BR-6
          severity: Minor
          title: The guard's scan scope (cmd/sdlc package dir only) is undocumented, and cmd/doc-review has an unruled sibling
          detail: |-
            gatefindings_test.go:70 globs "./*.go", excluding subpackages; one comment naming that as the
            deliberate class boundary would prevent a wrong assumption. cmd/doc-review/review.go:125 prints
            "triage each finding" to the main agent — the same framing in another binary, deserving an
            explicit in-or-out ruling.
          family: undocumented-scan-boundary
          round: 1
      blocked: true
    - "n": 2
      timestamp: "2026-08-22T12:33:48-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: TestEveryFixerFacingHelptextRoutes verified red on revert, naming exactly close.md:13, close.md:59, milestone-close.md:38.
          round: 2
        - id: BR-2
          disposition: addressed
          note: All four named shapes plus a fifth I constructed now fail the guard; verified in a scratch copy.
          round: 2
        - id: BR-3
          disposition: not-addressed
          note: close.md:79's citation is guarded; change-code.md:37's is not — removing it fails no test, since neither it nor its predecessor paragraph carries a directive verb.
          round: 2
        - id: BR-4
          disposition: addressed
          note: 'Verified red: stripping fixTheClassNote from classifyFallback fails TestGatePathStderrCarriesRoutingLine on real stderr.'
          round: 2
        - id: BR-5
          disposition: addressed
          note: fixTheClassNote owns the join at six sites; the two remaining indents are argued at gatefindings.go:33.
          round: 2
        - id: BR-6
          disposition: addressed
          note: Scan boundary documented at the test; doc-review ruled OUT in Non-goals with reasoning.
          round: 2
      findings:
        - id: BR-7
          severity: Important
          title: Third round of the same family — two further shapes of the guard's class ship green; state the rule, do not patch F and H
          detail: |-
            Verified green in a scratch copy: (F) a fixer-facing message split across adjacent literals in a
            call that is not cwarn/die — pass 2 never concatenates, and cmd/sdlc non-test files hold ~200
            fmt.Fprint* calls and ~70 adjacent-literal concatenations; (H) a func whose cwarn routes (claimed
            by pass 1) and which also builds a second unrouted fixer-facing line — countRoutingRefs credits the
            claimed call's ref to the unclaimed literal, and finalizeBoundaryReview already carries two such
            credits. Rounds 1 and 2 each widened to the shapes the finding named. The RULE: match every
            fixer-facing message as a whole string-valued expression (fold +-joined BinaryExpr chains wherever
            they occur, not only inside two hardcoded emitter names) and attribute each match to a routing ref
            in its own syntactic unit (count only unclaimed refs in pass 2). One pass over string-valued
            expressions attributed to the nearest enclosing statement makes F and H the same case, and the same
            rule covers the doc scan, which shares fixerFacingMessage. Alternative, equally valid at class
            level: soften atlas/workflow/gate-state.md:127 and Done-when 4 to the declared boundary rather than
            keeping the strong claim over the weak mechanism. ARCH-PURPOSE.
          family: guard-narrower-than-claimed-class
          round: 2
        - id: BR-8
          severity: Minor
          title: The issue's own "Doc surfaces" enumeration is stale — names 2, delivered 5 across 3 files
          detail: |-
            Second in this family. The section whose point is being mechanical is still hand-maintained: it
            names close.md FINDING FAMILIES and change-code.md THE PLAN GATE, while BR-1's fix routed five
            paragraphs across close.md, change-code.md and milestone-close.md; Done-when 5 omits
            milestone-close.md. Do not just correct the list — the rule is that a hand-written enumeration of a
            class the guards now compute must ROUTE at the guard (the #128 pattern this issue applies
            everywhere else) instead of restating sites that drift. Same rule covers the "eight sites" table,
            which will go stale the moment a ninth lands.
          family: hand-enumerated-class
          round: 2
        - id: BR-9
          severity: Minor
          title: superpowers-receiving-code-review is an unruled sibling that actively contradicts the issue's thesis
          detail: |-
            Second in this family. construct/adapted/superpowers-receiving-code-review/SKILL.md is the repo's
            canonical findings-RECEPTION surface — invoked exactly when a gate hands findings over — and its
            response pattern ends "6. IMPLEMENT: One item at a time, test each", the per-site patching #203
            exists to stop. It escapes the doc scan only because it says "feedback"/"items", never "finding".
            The rule: each guard declares its own glob, but the SET of globs (which instruction surfaces are in
            the class) is never declared, so Non-goals rules siblings in or out from memory — doc-review by
            name, this one not at all. Declare the surface set, then rule each member.
          family: undocumented-scan-boundary
          round: 2
      blocked: true
    - "n": 3
      timestamp: "2026-08-22T12:50:38-07:00"
      agent: claude
      dispose:
        - id: BR-3
          disposition: not-addressed
          note: 'Re-verified: deleting the ARCH-PURPOSE paragraph at change-code.md:37 fails no test — neither it nor its neighbours carry a directive verb, so fixerFacingMessage never reaches them. Class fix: a guard that every ARCH-* marker cited in helptext exists in the registry.'
          round: 3
        - id: BR-7
          disposition: addressed
          note: 'Verified in a scratch worktree: shapes F and H both go red, along with a package-level const and an off-list verb; the two-pass name-keyed structure is deleted, not widened.'
          round: 3
        - id: BR-8
          disposition: addressed
          note: 'The section it named now routes at the guards. Note: Plan items 4-5 still hand-name "the eight sites" and two doc surfaces, but those are checked-off historical steps rather than a live enumeration, and `## Done when` was removed entirely along with its restatements.'
          round: 3
        - id: BR-9
          disposition: addressed
          note: Surface set declared; SKILL.md guard verified red on revert on both assertions; the .claude/skills symlink confirms the edit reaches the loaded skill. Its durability is a separate new finding.
          round: 3
      findings:
        - id: BR-10
          severity: Important
          title: Statement attribution is membership, not counting — two messages sharing one statement with one routing ref ship green
          detail: |-
            4th in this family, and not a new shape: it is the half of BR-7's own stated rule that round 3
            dropped ("count only unclaimed refs"). Verified green in a scratch copy — an append carrying
            "1. Fix the findings NOW..." + fixTheClassLine() + "2. Any remaining findings — address them"
            passes. Reachable in formatFixThenShipProtocol specifically, because round 3's fix put the
            routing ref into that same append. THE RULE, and why to stop widening here: a syntactic
            approximation cannot back an absolute claim — counting closes this residue, the next (two refs
            both joined to one message) is one level further, without bound. So fix the CLAIM:
            atlas/workflow/gate-state.md:127 still says "a ninth refusal site cannot ship unrouted", and the
            paragraph under it describes this exact defect with "function" where "statement" now belongs.
            State what the guard approximates (whole +-folded string expressions, per-statement attribution,
            statically visible literals only). I also prototyped the counting change: ~5 lines, closes the
            residue, ZERO false positives on the real tree — free, but optional; the claim is the deliverable.
            ARCH-PURPOSE.
          family: guard-narrower-than-claimed-class
          round: 3
        - id: BR-11
          severity: Important
          title: BR-9's SKILL.md fix edits render output; /construct upgrade regenerates it from an intent transcript that never mentions the skill
          detail: |-
            construct/adapted/ is generated. /construct promote step 4 is delete-then-copy (rm -rf
            construct/adapted/superpowers-*/ then copy staging), and /construct upgrade renders staging from
            construct/sources/<version>/ + construct/intents/superpowers.md, where "skills not mentioned in
            the intent → copy new source as-is". grep for receiving-code-review in that transcript returns
            nothing, so the GENERALIZE step, the ARCH-PURPOSE routing, and the family:-as-worklist paragraph
            are all scheduled for deletion by the substrate's own pipeline. The precedent is exact: a547179
            (#71), the last ARCH-registry change touching an adapted skill, appended a Conversation entry
            with Verify clauses in the same commit. TestReceivingCodeReviewSkillGeneralizes makes the loss
            loud rather than silent, which is why this is Important not Critical — but a red test with no
            intent record leaves the next reader reconstructing wording from three strings.Contains calls.
            This is the issue's own principle one level out: the deliverable landed at the compiled consumer
            instead of the source it derives from. ARCH-PURPOSE.
          family: fix-at-consumer-not-source
          round: 3
        - id: BR-12
          severity: Minor
          title: The routing line is 120 chars — 129 at the FIX-THEN-SHIP indent, against ~83 for every neighbouring line
          detail: |-
            close.go:1812 places it in a block whose widest existing line is 83. It hard-wraps mid-sentence in
            an 80/100-col terminal. Splitting after the colon would match the surrounding rhythm.
          family: refusal-line-width
          round: 3
      blocked: true
    - "n": 4
      timestamp: "2026-08-22T13:07:21-07:00"
      agent: claude
      dispose:
        - id: BR-3
          disposition: not-addressed
          note: 'Re-verified a third time — deleting the ARCH-PURPOSE paragraph at helptext/change-code.md:37 adds zero test failures over the baseline. Class fix remains: a guard that every ARCH-* marker cited in helptext exists in the registry.'
          round: 4
        - id: BR-10
          disposition: addressed
          note: Counting verified load-bearing both ways (probe red with counting, green with membership restored); gate-state.md's absolute claim replaced with a stated over-approximation naming all three residues.
          round: 4
        - id: BR-11
          disposition: addressed
          note: Pipeline premise re-derived from construct/skill/construct/SKILL.md; source still holds "One item at a time"; Conv 8's Verify clauses match the shipped skill; a547179 precedent confirmed to have touched intents in-commit.
          round: 4
        - id: BR-12
          disposition: not-addressed
          note: Unchanged — measured 120 chars, 129 at the FIX-THEN-SHIP indent, 122 on a cwarn continuation.
          round: 4
      findings:
        - id: BR-13
          severity: Minor
          title: superpowers-requesting-code-review is a live in-class sibling the declared surface set neither guards nor rules out
          detail: |-
            This is the 3rd finding in family `undocumented-scan-boundary` — do NOT rule this
            instance and stop. SKILL.md:52-56 says "3. Act on feedback: Fix Critical issues
            immediately / Fix Important issues before proceeding / Note Minor issues for later":
            a fixer-facing directive, in a skill symlinked live at .claude/skills/, kept in play
            by AGENTS.md section 3 for ad-hoc reviews, in the same directory as the surface BR-9
            named, escaping the doc scan by the identical mechanism (it says "feedback"/"issues",
            never "findings"). THE RULE: the surface SET is hand-declared and unverified, so a
            member can be missing without any test firing — BR-9 wrote the hand enumeration down
            rather than computing it. Measured prevalence: 14 adapted skills, exactly 2 carry
            fixer-facing directives, the declared set names 1. Apply BR-10's own remedy one level
            up rather than widening again: gate-state.md:148 and the fixerFacingSurfaces header
            both claim the "whole surface set", which is an absolute claim over a hand-maintained
            artifact — say instead that the scans compute the sites while the set of surfaces is
            declared by hand. (Extending the doc glob over construct/adapted/*/SKILL.md would not
            reach this one without widening fixerFacingMessage past "finding", which is the other
            family's residue.) ARCH-PURPOSE.
          family: undocumented-scan-boundary
          round: 4
        - id: BR-14
          severity: Minor
          title: referencesRoutingLine is dead code — the superseded membership predicate sits beside the live counting one
          detail: |-
            cmd/sdlc/gatefindings_test.go:426, zero call sites since round 4 replaced membership
            with counting at :197. It is the exact predicate BR-10 named as the defect, left
            immediately below its replacement, so a future edit reaching for the wrong helper
            silently re-opens the residue. Round 3's Log claims the redesign deleted the whole
            two-pass/claimed structure; this is the piece that survived. Delete it. ARCH-DRY.
          family: superseded-mechanism-left-in-tree
          round: 4
      blocked: false
---

# Gate ledger — ariadne#203 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-08-22T10:34:51-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `hand-enumerated-class` Doc-surface class enumerated by hand; three unrouted findings surfaces remain, two of them doc twins of code sites this diff routed
  A paragraph-level scan of cmd/sdlc/helptext/*.md for findings-word + directive-verb finds
  close.md:13 (POST-VERDICT PROTOCOL, "fix the findings NOW" — twin of the routed close.go:1809),
  close.md:59 ("Fix the findings, or have the next review dispose them" — twin of the routed
  close.go:1194), and milestone-close.md:38 ("fix the findings before committing"; the file was
  not touched at all). Spec bullet 2 says EVERY surface routes. The Go class got a mechanical
  scan that found 8 where a hand pass found 4; the doc class got a hand pass. ARCH-PURPOSE.
- **BR-2** [Important] `guard-narrower-than-claimed-class` Enumeration guard's class signature is narrower than the class it claims; four synthetic ninth sites ship green
  Verified by appending unrouted sites to cmd/sdlc/state.go and re-running the guard. Blind to
  (a) a message split across adjacent string literals — the tree's own prevailing style, e.g.
  close.go:1194 is 3 literals and changecode.go:554 became 2 in this diff; (b) a package-level
  const, since pass 2 walks only FuncDecl bodies; (c) a directive verb outside "fix "/"address "/
  "review above"; and (d) a second unrouted line inside an already-routing builder func, because
  pass 2 is function-granular — the exact thing pass 1's own comment declares insufficient.
  atlas/workflow/gate-state.md:127 and Done-when 4 claim "a ninth refusal site cannot ship
  unrouted". Fix: concatenate a call's literals before matching; make pass 2 count-based.
- **BR-3** [Minor] `guard-narrower-than-claimed-class` New ARCH-PURPOSE citations in helptext are unguarded, unlike the routing line itself
  helptext/close.md:79 and helptext/change-code.md:37 cite ARCH-PURPOSE with no drift guard;
  TestFixTheClassLine_RoutesToArchPrinciples covers only fixTheClassLine, and
  TestArchitecture_NarrativeRoutesToArchPrinciples reads AGENTS.md only.
- **BR-4** [Minor] `coverage-stops-at-the-seam` No test asserts a real gate path's stderr actually contains the routing line
  Coverage is source-level plus string-level with nothing joining them. changecode_test.go:364
  already drives the blocking runPlanQualityJudge path with a stub judge; swapping ioDiscard()
  for a bytes.Buffer and asserting on "ARCH-PURPOSE" is a two-line end-to-end pin.
- **BR-5** [Minor] `repeated-join-prefix` The "\n  " join prefix is hand-written at six of the eight call sites
  changecode.go:554,572,722, close.go:1238, judge.go:172, milestoneclose.go:626 each spell the
  same newline+indent. A fixTheClassNote() returning the pre-joined form would leave only the
  two genuinely different indents (close.go:1194, close.go:1809).
- **BR-6** [Minor] `undocumented-scan-boundary` The guard's scan scope (cmd/sdlc package dir only) is undocumented, and cmd/doc-review has an unruled sibling
  gatefindings_test.go:70 globs "./*.go", excluding subpackages; one comment naming that as the
  deliberate class boundary would prevent a wrong assumption. cmd/doc-review/review.go:125 prints
  "triage each finding" to the main agent — the same framing in another binary, deserving an
  explicit in-or-out ruling.

## Round 2 — 2026-08-22T12:33:48-07:00 (claude) — BLOCKED

### Disposed

- BR-1 — addressed — TestEveryFixerFacingHelptextRoutes verified red on revert, naming exactly close.md:13, close.md:59, milestone-close.md:38.
- BR-2 — addressed — All four named shapes plus a fifth I constructed now fail the guard; verified in a scratch copy.
- BR-3 — not-addressed — close.md:79's citation is guarded; change-code.md:37's is not — removing it fails no test, since neither it nor its predecessor paragraph carries a directive verb.
- BR-4 — addressed — Verified red: stripping fixTheClassNote from classifyFallback fails TestGatePathStderrCarriesRoutingLine on real stderr.
- BR-5 — addressed — fixTheClassNote owns the join at six sites; the two remaining indents are argued at gatefindings.go:33.
- BR-6 — addressed — Scan boundary documented at the test; doc-review ruled OUT in Non-goals with reasoning.

### Raised

- **BR-7** [Important] `guard-narrower-than-claimed-class` Third round of the same family — two further shapes of the guard's class ship green; state the rule, do not patch F and H
  Verified green in a scratch copy: (F) a fixer-facing message split across adjacent literals in a
  call that is not cwarn/die — pass 2 never concatenates, and cmd/sdlc non-test files hold ~200
  fmt.Fprint* calls and ~70 adjacent-literal concatenations; (H) a func whose cwarn routes (claimed
  by pass 1) and which also builds a second unrouted fixer-facing line — countRoutingRefs credits the
  claimed call's ref to the unclaimed literal, and finalizeBoundaryReview already carries two such
  credits. Rounds 1 and 2 each widened to the shapes the finding named. The RULE: match every
  fixer-facing message as a whole string-valued expression (fold +-joined BinaryExpr chains wherever
  they occur, not only inside two hardcoded emitter names) and attribute each match to a routing ref
  in its own syntactic unit (count only unclaimed refs in pass 2). One pass over string-valued
  expressions attributed to the nearest enclosing statement makes F and H the same case, and the same
  rule covers the doc scan, which shares fixerFacingMessage. Alternative, equally valid at class
  level: soften atlas/workflow/gate-state.md:127 and Done-when 4 to the declared boundary rather than
  keeping the strong claim over the weak mechanism. ARCH-PURPOSE.
- **BR-8** [Minor] `hand-enumerated-class` The issue's own "Doc surfaces" enumeration is stale — names 2, delivered 5 across 3 files
  Second in this family. The section whose point is being mechanical is still hand-maintained: it
  names close.md FINDING FAMILIES and change-code.md THE PLAN GATE, while BR-1's fix routed five
  paragraphs across close.md, change-code.md and milestone-close.md; Done-when 5 omits
  milestone-close.md. Do not just correct the list — the rule is that a hand-written enumeration of a
  class the guards now compute must ROUTE at the guard (the #128 pattern this issue applies
  everywhere else) instead of restating sites that drift. Same rule covers the "eight sites" table,
  which will go stale the moment a ninth lands.
- **BR-9** [Minor] `undocumented-scan-boundary` superpowers-receiving-code-review is an unruled sibling that actively contradicts the issue's thesis
  Second in this family. construct/adapted/superpowers-receiving-code-review/SKILL.md is the repo's
  canonical findings-RECEPTION surface — invoked exactly when a gate hands findings over — and its
  response pattern ends "6. IMPLEMENT: One item at a time, test each", the per-site patching #203
  exists to stop. It escapes the doc scan only because it says "feedback"/"items", never "finding".
  The rule: each guard declares its own glob, but the SET of globs (which instruction surfaces are in
  the class) is never declared, so Non-goals rules siblings in or out from memory — doc-review by
  name, this one not at all. Declare the surface set, then rule each member.

## Round 3 — 2026-08-22T12:50:38-07:00 (claude) — BLOCKED

### Disposed

- BR-3 — not-addressed — Re-verified: deleting the ARCH-PURPOSE paragraph at change-code.md:37 fails no test — neither it nor its neighbours carry a directive verb, so fixerFacingMessage never reaches them. Class fix: a guard that every ARCH-* marker cited in helptext exists in the registry.
- BR-7 — addressed — Verified in a scratch worktree: shapes F and H both go red, along with a package-level const and an off-list verb; the two-pass name-keyed structure is deleted, not widened.
- BR-8 — addressed — The section it named now routes at the guards. Note: Plan items 4-5 still hand-name "the eight sites" and two doc surfaces, but those are checked-off historical steps rather than a live enumeration, and `## Done when` was removed entirely along with its restatements.
- BR-9 — addressed — Surface set declared; SKILL.md guard verified red on revert on both assertions; the .claude/skills symlink confirms the edit reaches the loaded skill. Its durability is a separate new finding.

### Raised

- **BR-10** [Important] `guard-narrower-than-claimed-class` Statement attribution is membership, not counting — two messages sharing one statement with one routing ref ship green
  4th in this family, and not a new shape: it is the half of BR-7's own stated rule that round 3
  dropped ("count only unclaimed refs"). Verified green in a scratch copy — an append carrying
  "1. Fix the findings NOW..." + fixTheClassLine() + "2. Any remaining findings — address them"
  passes. Reachable in formatFixThenShipProtocol specifically, because round 3's fix put the
  routing ref into that same append. THE RULE, and why to stop widening here: a syntactic
  approximation cannot back an absolute claim — counting closes this residue, the next (two refs
  both joined to one message) is one level further, without bound. So fix the CLAIM:
  atlas/workflow/gate-state.md:127 still says "a ninth refusal site cannot ship unrouted", and the
  paragraph under it describes this exact defect with "function" where "statement" now belongs.
  State what the guard approximates (whole +-folded string expressions, per-statement attribution,
  statically visible literals only). I also prototyped the counting change: ~5 lines, closes the
  residue, ZERO false positives on the real tree — free, but optional; the claim is the deliverable.
  ARCH-PURPOSE.
- **BR-11** [Important] `fix-at-consumer-not-source` BR-9's SKILL.md fix edits render output; /construct upgrade regenerates it from an intent transcript that never mentions the skill
  construct/adapted/ is generated. /construct promote step 4 is delete-then-copy (rm -rf
  construct/adapted/superpowers-*/ then copy staging), and /construct upgrade renders staging from
  construct/sources/<version>/ + construct/intents/superpowers.md, where "skills not mentioned in
  the intent → copy new source as-is". grep for receiving-code-review in that transcript returns
  nothing, so the GENERALIZE step, the ARCH-PURPOSE routing, and the family:-as-worklist paragraph
  are all scheduled for deletion by the substrate's own pipeline. The precedent is exact: a547179
  (#71), the last ARCH-registry change touching an adapted skill, appended a Conversation entry
  with Verify clauses in the same commit. TestReceivingCodeReviewSkillGeneralizes makes the loss
  loud rather than silent, which is why this is Important not Critical — but a red test with no
  intent record leaves the next reader reconstructing wording from three strings.Contains calls.
  This is the issue's own principle one level out: the deliverable landed at the compiled consumer
  instead of the source it derives from. ARCH-PURPOSE.
- **BR-12** [Minor] `refusal-line-width` The routing line is 120 chars — 129 at the FIX-THEN-SHIP indent, against ~83 for every neighbouring line
  close.go:1812 places it in a block whose widest existing line is 83. It hard-wraps mid-sentence in
  an 80/100-col terminal. Splitting after the colon would match the surrounding rhythm.

## Round 4 — 2026-08-22T13:07:21-07:00 (claude) — passed

### Disposed

- BR-3 — not-addressed — Re-verified a third time — deleting the ARCH-PURPOSE paragraph at helptext/change-code.md:37 adds zero test failures over the baseline. Class fix remains: a guard that every ARCH-* marker cited in helptext exists in the registry.
- BR-10 — addressed — Counting verified load-bearing both ways (probe red with counting, green with membership restored); gate-state.md's absolute claim replaced with a stated over-approximation naming all three residues.
- BR-11 — addressed — Pipeline premise re-derived from construct/skill/construct/SKILL.md; source still holds "One item at a time"; Conv 8's Verify clauses match the shipped skill; a547179 precedent confirmed to have touched intents in-commit.
- BR-12 — not-addressed — Unchanged — measured 120 chars, 129 at the FIX-THEN-SHIP indent, 122 on a cwarn continuation.

### Raised

- **BR-13** [Minor] `undocumented-scan-boundary` superpowers-requesting-code-review is a live in-class sibling the declared surface set neither guards nor rules out
  This is the 3rd finding in family `undocumented-scan-boundary` — do NOT rule this
  instance and stop. SKILL.md:52-56 says "3. Act on feedback: Fix Critical issues
  immediately / Fix Important issues before proceeding / Note Minor issues for later":
  a fixer-facing directive, in a skill symlinked live at .claude/skills/, kept in play
  by AGENTS.md section 3 for ad-hoc reviews, in the same directory as the surface BR-9
  named, escaping the doc scan by the identical mechanism (it says "feedback"/"issues",
  never "findings"). THE RULE: the surface SET is hand-declared and unverified, so a
  member can be missing without any test firing — BR-9 wrote the hand enumeration down
  rather than computing it. Measured prevalence: 14 adapted skills, exactly 2 carry
  fixer-facing directives, the declared set names 1. Apply BR-10's own remedy one level
  up rather than widening again: gate-state.md:148 and the fixerFacingSurfaces header
  both claim the "whole surface set", which is an absolute claim over a hand-maintained
  artifact — say instead that the scans compute the sites while the set of surfaces is
  declared by hand. (Extending the doc glob over construct/adapted/*/SKILL.md would not
  reach this one without widening fixerFacingMessage past "finding", which is the other
  family's residue.) ARCH-PURPOSE.
- **BR-14** [Minor] `superseded-mechanism-left-in-tree` referencesRoutingLine is dead code — the superseded membership predicate sits beside the live counting one
  cmd/sdlc/gatefindings_test.go:426, zero call sites since round 4 replaced membership
  with counting at :197. It is the exact predicate BR-10 named as the defect, left
  immediately below its replacement, so a future edit reaching for the wrong helper
  silently re-opens the residue. Round 3's Log claims the redesign deleted the whole
  two-pass/claimed structure; this is the piece that survived. Delete it. ARCH-DRY.

## Open findings

- **BR-3** [Minor] `guard-narrower-than-claimed-class` New ARCH-PURPOSE citations in helptext are unguarded, unlike the routing line itself
- **BR-12** [Minor] `refusal-line-width` The routing line is 120 chars — 129 at the FIX-THEN-SHIP indent, against ~83 for every neighbouring line
- **BR-13** [Minor] `undocumented-scan-boundary` superpowers-requesting-code-review is a live in-class sibling the declared surface set neither guards nor rules out
- **BR-14** [Minor] `superseded-mechanism-left-in-tree` referencesRoutingLine is dead code — the superseded membership predicate sits beside the live counting one
