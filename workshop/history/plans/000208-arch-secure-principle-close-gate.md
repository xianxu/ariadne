---
gate: boundary-review
issue: 208
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-09-02T15:50:36-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Important
          title: Deferred-guard vacuity classification has no test and neither branch is reachable in the committed tree
          detail: |-
            architecturedeferred_test.go:71-82 — with ARCH-AUTHORITY present the switch is
            never entered, so both the fatal (sections>0, no markers) and skip (neither)
            branches are dead, and `sections` is used only inside a message string. The
            issue's Log claims "Three probes now pin it — activate/break a heading/leak",
            but the file ships one test and two helpers; those probes were manual and left
            no artifact. deferredMarkers fuses the embed read, the parse and t.Fatalf, so
            classification cannot be exercised without editing the real file (ARCH-PURE).
            Split out pure parseDeferred(text) and classifyDeferred(markers, sections) and
            table-test the three states plus a broken-heading fixture.
          family: unexercised-guard-branch
          round: 1
        - id: BR-2
          severity: Important
          title: Marker extraction is copy-pasted between ArchitectureMarkers and the guard's deferredMarkers
          detail: |-
            architecturedeferred_test.go:37-45 duplicates architecture.go:19-29 —
            the same order-preserving dedupe over archMarkerRE.FindAllStringSubmatch with
            only the input string changed (ARCH-DRY). The guard's correctness depends on the
            forbidden set being extracted identically to the gated set; divergence yields a
            silent false green, the exact failure this issue defends against. Extract
            markersIn(text string) []string in architecture.go and have both call it.
          family: derive-not-restate
          round: 1
        - id: BR-3
          severity: Important
          title: atlas/workflow/sdlc-binary.md not updated and now contradicts the atlas page this diff wrote
          detail: |-
            sdlc-binary.md:880-895 carries the per-entry narrative (lines for #71, #126,
            #205) with no line for #208 and no mention of architecture-deferred.md or the
            documented-but-not-gated concept — new architectural surface. Worse, line 893
            still asserts "Adding an ARCH-* entry flows into every consumer with no other
            edit", which the new architecture-principles.md "Adding an entry" section
            directly contradicts (registry + the one tripwire + the goldens). A reader
            following sdlc-binary.md edits only architecture.md and gets a red suite.
          family: doc-enumeration-drift
          round: 1
        - id: BR-4
          severity: Minor
          title: Done-when claims an unconditional at-least-one-marker assert; the code ships a classified skip, and a truncated file skips silently
          detail: |-
            The implementation deliberately replaced PQ-2's unconditional assert with a
            sections-vs-markers classification (the Log explains why: an unconditional
            assert reds on activating the last entry). The Done-when bullet was never
            amended, so the issue still claims behavior the code does not have. Separately,
            architecturedeferred_test.go:76-81 comments that broken headings are the only
            real disarm vector — an emptied or truncated file also yields a silent t.Skip.
            Require the file's title line before allowing the skip path.
          family: unbacked-existing-behavior
          round: 1
        - id: BR-5
          severity: Minor
          title: Entry-section lookup matches ARCH-PURE inside ARCH-PURPOSE
          detail: |-
            judge_test.go:133 — strings.Index(ArchitectureRegistry, "## "+m) is correct only
            because ARCH-PURE precedes ARCH-PURPOSE in the file. A reorder would silently
            lens-check the wrong entry. Match "## "+m+" " (or the em-dash form) instead.
          family: prefix-match-fragility
          round: 1
        - id: BR-6
          severity: Minor
          title: startplan_test.go holds a sixth hand-written marker list the class sweep missed
          detail: |-
            cmd/sdlc/startplan_test.go:36 spot-checks "ARCH-DRY", "ARCH-CONSTRAINTS" by
            hand. Benign — line 41's Contains(out, judge.ArchitectureRegistry) already
            covers new markers derivedly, which is what satisfies the "start-plan delivers
            it" Done-when — but it is a residual member of the swept class and the two
            literals now assert nothing the registry check doesn't.
          family: marker-enumeration-sites
          round: 1
        - id: BR-7
          severity: Minor
          title: Declared prompt-growth budget is 1.4 KB; measured is 1,991 bytes
          detail: |-
            architecture.md grew 7,519 -> 9,510 bytes, ~39% over the Spec's stated
            "roughly +1.4 KB per dispatch", across four prompts plus start-plan
            (ARCH-CONSTRAINTS at-review: implementation vs declared envelope). The tradeoff
            remains clearly worth it; replace the estimate with the measured value in the
            issue's Operating envelope section and atlas so the ledger records a fact.
          family: operating-envelope-unstated
          round: 1
        - id: BR-8
          severity: Minor
          title: Pre-existing red test blocks any "full suite green" close evidence (out of window)
          detail: |-
            cmd/sdlc/fleet_plan_test.go:14 fails — it reads
            workshop/plans/000200-sdlc-fleet-thread-inventory-plan.md, archived to
            workshop/history/plans/ in dfeba9c, an ancestor of this window's base. Not
            caused by this diff. Scope the close's verification evidence to the packages
            touched (judge + cmd/sdlc archprinciples/startplan are green), and have that
            test resolve the plan through workshop/plans/ or workshop/history/plans/.
          family: stale-path-in-test
          round: 1
      blocked: true
    - "n": 2
      timestamp: "2026-09-02T16:02:59-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: Verified by revert — disabling section counting reds two named subtests; breaking the committed heading reds both the classification and the guard.
          round: 2
        - id: BR-2
          disposition: addressed
          note: markersIn is now the single extraction site; ArchitectureMarkers() and the deferred scan both call it.
          round: 2
        - id: BR-3
          disposition: addressed
          note: sdlc-binary.md now separates runtime consumers (derive) from the test layer, and documents the deferred file.
          round: 2
        - id: BR-4
          disposition: addressed
          note: TestDeferredFileIsGuarding closes the truncated-file skip vector and makes the Done-when claim true — at the cost of the incompatible sibling bullet, raised as a new finding rather than re-raised here.
          round: 2
        - id: BR-5
          disposition: not-addressed
          note: judge_test.go:133 still uses strings.Index(ArchitectureRegistry, "## "+m); the class now has a second member at architecturedeferred_test.go:158 (strings.Contains for marker leak) — the rule is that every marker match must be delimiter-anchored.
          round: 2
        - id: BR-6
          disposition: not-addressed
          note: cmd/sdlc/startplan_test.go:36 and :47 still hand-write ARCH-DRY / ARCH-CONSTRAINTS / ARCH-PURE; the file is untouched in this window.
          round: 2
        - id: BR-7
          disposition: not-addressed
          note: Re-measured at head — architecture.md 7,519 to 9,510 bytes = 1,991; the issue still states "roughly +1.4 KB per dispatch".
          round: 2
        - id: BR-8
          disposition: not-addressed
          note: go test ./... still fails at cmd/sdlc/fleet_plan_test.go:14 on the archived plan path; pre-existing and out of window.
          round: 2
      findings:
        - id: BR-9
          severity: Important
          title: TestDeferredFileIsGuarding duplicates the deferredBroken fatal and its only unique effect is to red the suite on deferredRetired, re-breaking the activation contract
          detail: |-
            Probed against 34f0f59 in a scratch copy: breaking the deferred heading fails BOTH
            TestDeferredPrinciplesReachNoGate (t.Fatalf, architecturedeferred_test.go:136) and
            TestDeferredFileIsGuarding; emptying the file (equivalently, activating the only
            deferred entry) fails ONLY TestDeferredFileIsGuarding. So the new test adds nothing
            on the state that matters and vetoes the state the section-count classification was
            built to bless. The Log records the implementor finding and removing exactly this
            failure in round 1; the BR-1 fix reintroduced it.
            This is the 2nd finding in family unbacked-existing-behavior, so the deliverable is
            the rule, not this test. RULE — the classifier owns the pass/fail/skip action for
            each of its states at ONE site; no second consumer assigns an outcome to a state,
            and no prose restates the mapping. Enumeration of the six sites currently asserting
            the now-false "activation is a pure move": architecture-deferred.md:6;
            architecturedeferred_test.go:30; architecturedeferred_test.go:141-142;
            atlas/workflow/architecture-principles.md:67 and :84; atlas/workflow/sdlc-binary.md:909;
            the issue's Done-when bullet 3 (which contradicts its own bullet 4 — that tension is
            the root). Fix: move the action onto deferredVerdict, narrow or delete
            TestDeferredFileIsGuarding, and rewrite the six sites to point at the one mapping.
            Same finding, secondary: architecturedeferred_test.go:20-29 is an orphaned doc comment
            for the deleted deferredMarkers, fused without a blank line to deferredState's comment,
            so godoc renders the wrong doc for that type. atlas/workflow/architecture-principles.md
            was not touched by the fix commit and now omits both markersIn and TestDeferredFileIsGuarding.
          family: unbacked-existing-behavior
          round: 2
      blocked: true
    - "n": 3
      timestamp: "2026-09-02T16:14:23-07:00"
      agent: claude
      dispose:
        - id: BR-9
          disposition: addressed
          note: Verified by probe at 5903b5d, not by the Log — activating ARCH-AUTHORITY leaves TestDeferredPrinciplesReachNoGate SKIPped green and fails exactly TestArchitectureMarkers + TestBuildPrompt_Golden; breaking the heading reds one test with an actionable message; the orphaned deferredMarkers doc comment is folded and the atlas gained markersIn and the verdict table.
          round: 3
        - id: BR-5
          disposition: not-addressed
          note: judge_test.go:133 still uses strings.Index(ArchitectureRegistry, "## "+m); architecturedeferred_test.go:165 still uses strings.Contains for the leak check. Neither is delimiter-anchored.
          round: 3
        - id: BR-6
          disposition: not-addressed
          note: cmd/sdlc/startplan_test.go:36 still hand-writes ARCH-DRY / ARCH-CONSTRAINTS and :47 ARCH-PURE; the file is untouched across the whole window.
          round: 3
        - id: BR-7
          disposition: not-addressed
          note: Re-measured at head — architecture.md 7,519 to 9,510 bytes = 1,991; the issue's Operating envelope still reads "roughly +1.4 KB per dispatch".
          round: 3
        - id: BR-8
          disposition: not-addressed
          note: go test ./... still fails only at cmd/sdlc/fleet_plan_test.go:14 on the archived plan path; pre-existing, out of window, and the reason close evidence must be package-scoped.
          round: 3
      findings:
        - id: BR-10
          severity: Minor
          title: The fix that single-sourced the verdict mapping restates it in the atlas, and both sites claim it is documented nowhere else
          detail: |-
            atlas/workflow/architecture-principles.md:85-88 states the mapping "is documented there
            and nowhere else, so this page describes it rather than restating the rule" — and lines
            90-94 then restate the full rule as a three-row table. architecturedeferred_test.go:55-56
            makes the same false claim from the Go side ("enforced in exactly one place ... and
            documented nowhere else"). This is the 2nd finding in family doc-enumeration-drift (BR-3
            was atlas/workflow/sdlc-binary.md contradicting the page this diff wrote), so per the
            escalation protocol the deliverable is the rule, not this table. RULE — a doc may name a
            code-owned mapping and point at the owning symbol, or it may reproduce the mapping and
            say so, but it may not claim single-sourcing while reproducing. Prevalence in this issue:
            2 of 2 atlas pages touched. Cheapest resolution: keep the table (it is genuinely useful
            map content) and delete the two "nowhere else" claims, replacing them with "the owning
            symbol is deferredVerdict; this table mirrors it."
          family: doc-enumeration-drift
          round: 3
        - id: BR-11
          severity: Minor
          title: Two derived marker assertions are checked against the source they were derived from, so they cannot fail
          detail: |-
            judge_test.go:124-129 builds `want` from ArchitectureMarkers() and asserts
            strings.Contains(ArchitectureRegistry, w) — but markersIn extracted those exact
            substrings from ArchitectureRegistry, so that branch is unreachable by construction.
            archprinciples_test.go:23 has the same shape, and is additionally subsumed by the
            Contains(out, judge.ArchitectureRegistry) assertion three lines below it. The non-marker
            entries in both loops ("at-plan", "at-review", "principle:") are real and should stay.
            RULE — derive an expectation from the source, then assert it against a CONSUMER; an
            expectation asserted against its own source is a no-op. Adjacent to BR-6's observation
            that the startplan_test literals "now assert nothing the registry check doesn't".
          family: tautological-derivation
          round: 3
      blocked: false
---

# Gate ledger — ariadne#208 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-09-02T15:50:36-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `unexercised-guard-branch` Deferred-guard vacuity classification has no test and neither branch is reachable in the committed tree
  architecturedeferred_test.go:71-82 — with ARCH-AUTHORITY present the switch is
  never entered, so both the fatal (sections>0, no markers) and skip (neither)
  branches are dead, and `sections` is used only inside a message string. The
  issue's Log claims "Three probes now pin it — activate/break a heading/leak",
  but the file ships one test and two helpers; those probes were manual and left
  no artifact. deferredMarkers fuses the embed read, the parse and t.Fatalf, so
  classification cannot be exercised without editing the real file (ARCH-PURE).
  Split out pure parseDeferred(text) and classifyDeferred(markers, sections) and
  table-test the three states plus a broken-heading fixture.
- **BR-2** [Important] `derive-not-restate` Marker extraction is copy-pasted between ArchitectureMarkers and the guard's deferredMarkers
  architecturedeferred_test.go:37-45 duplicates architecture.go:19-29 —
  the same order-preserving dedupe over archMarkerRE.FindAllStringSubmatch with
  only the input string changed (ARCH-DRY). The guard's correctness depends on the
  forbidden set being extracted identically to the gated set; divergence yields a
  silent false green, the exact failure this issue defends against. Extract
  markersIn(text string) []string in architecture.go and have both call it.
- **BR-3** [Important] `doc-enumeration-drift` atlas/workflow/sdlc-binary.md not updated and now contradicts the atlas page this diff wrote
  sdlc-binary.md:880-895 carries the per-entry narrative (lines for #71, #126,
  #205) with no line for #208 and no mention of architecture-deferred.md or the
  documented-but-not-gated concept — new architectural surface. Worse, line 893
  still asserts "Adding an ARCH-* entry flows into every consumer with no other
  edit", which the new architecture-principles.md "Adding an entry" section
  directly contradicts (registry + the one tripwire + the goldens). A reader
  following sdlc-binary.md edits only architecture.md and gets a red suite.
- **BR-4** [Minor] `unbacked-existing-behavior` Done-when claims an unconditional at-least-one-marker assert; the code ships a classified skip, and a truncated file skips silently
  The implementation deliberately replaced PQ-2's unconditional assert with a
  sections-vs-markers classification (the Log explains why: an unconditional
  assert reds on activating the last entry). The Done-when bullet was never
  amended, so the issue still claims behavior the code does not have. Separately,
  architecturedeferred_test.go:76-81 comments that broken headings are the only
  real disarm vector — an emptied or truncated file also yields a silent t.Skip.
  Require the file's title line before allowing the skip path.
- **BR-5** [Minor] `prefix-match-fragility` Entry-section lookup matches ARCH-PURE inside ARCH-PURPOSE
  judge_test.go:133 — strings.Index(ArchitectureRegistry, "## "+m) is correct only
  because ARCH-PURE precedes ARCH-PURPOSE in the file. A reorder would silently
  lens-check the wrong entry. Match "## "+m+" " (or the em-dash form) instead.
- **BR-6** [Minor] `marker-enumeration-sites` startplan_test.go holds a sixth hand-written marker list the class sweep missed
  cmd/sdlc/startplan_test.go:36 spot-checks "ARCH-DRY", "ARCH-CONSTRAINTS" by
  hand. Benign — line 41's Contains(out, judge.ArchitectureRegistry) already
  covers new markers derivedly, which is what satisfies the "start-plan delivers
  it" Done-when — but it is a residual member of the swept class and the two
  literals now assert nothing the registry check doesn't.
- **BR-7** [Minor] `operating-envelope-unstated` Declared prompt-growth budget is 1.4 KB; measured is 1,991 bytes
  architecture.md grew 7,519 -> 9,510 bytes, ~39% over the Spec's stated
  "roughly +1.4 KB per dispatch", across four prompts plus start-plan
  (ARCH-CONSTRAINTS at-review: implementation vs declared envelope). The tradeoff
  remains clearly worth it; replace the estimate with the measured value in the
  issue's Operating envelope section and atlas so the ledger records a fact.
- **BR-8** [Minor] `stale-path-in-test` Pre-existing red test blocks any "full suite green" close evidence (out of window)
  cmd/sdlc/fleet_plan_test.go:14 fails — it reads
  workshop/plans/000200-sdlc-fleet-thread-inventory-plan.md, archived to
  workshop/history/plans/ in dfeba9c, an ancestor of this window's base. Not
  caused by this diff. Scope the close's verification evidence to the packages
  touched (judge + cmd/sdlc archprinciples/startplan are green), and have that
  test resolve the plan through workshop/plans/ or workshop/history/plans/.

## Round 2 — 2026-09-02T16:02:59-07:00 (claude) — BLOCKED

### Disposed

- BR-1 — addressed — Verified by revert — disabling section counting reds two named subtests; breaking the committed heading reds both the classification and the guard.
- BR-2 — addressed — markersIn is now the single extraction site; ArchitectureMarkers() and the deferred scan both call it.
- BR-3 — addressed — sdlc-binary.md now separates runtime consumers (derive) from the test layer, and documents the deferred file.
- BR-4 — addressed — TestDeferredFileIsGuarding closes the truncated-file skip vector and makes the Done-when claim true — at the cost of the incompatible sibling bullet, raised as a new finding rather than re-raised here.
- BR-5 — not-addressed — judge_test.go:133 still uses strings.Index(ArchitectureRegistry, "## "+m); the class now has a second member at architecturedeferred_test.go:158 (strings.Contains for marker leak) — the rule is that every marker match must be delimiter-anchored.
- BR-6 — not-addressed — cmd/sdlc/startplan_test.go:36 and :47 still hand-write ARCH-DRY / ARCH-CONSTRAINTS / ARCH-PURE; the file is untouched in this window.
- BR-7 — not-addressed — Re-measured at head — architecture.md 7,519 to 9,510 bytes = 1,991; the issue still states "roughly +1.4 KB per dispatch".
- BR-8 — not-addressed — go test ./... still fails at cmd/sdlc/fleet_plan_test.go:14 on the archived plan path; pre-existing and out of window.

### Raised

- **BR-9** [Important] `unbacked-existing-behavior` TestDeferredFileIsGuarding duplicates the deferredBroken fatal and its only unique effect is to red the suite on deferredRetired, re-breaking the activation contract
  Probed against 34f0f59 in a scratch copy: breaking the deferred heading fails BOTH
  TestDeferredPrinciplesReachNoGate (t.Fatalf, architecturedeferred_test.go:136) and
  TestDeferredFileIsGuarding; emptying the file (equivalently, activating the only
  deferred entry) fails ONLY TestDeferredFileIsGuarding. So the new test adds nothing
  on the state that matters and vetoes the state the section-count classification was
  built to bless. The Log records the implementor finding and removing exactly this
  failure in round 1; the BR-1 fix reintroduced it.
  This is the 2nd finding in family unbacked-existing-behavior, so the deliverable is
  the rule, not this test. RULE — the classifier owns the pass/fail/skip action for
  each of its states at ONE site; no second consumer assigns an outcome to a state,
  and no prose restates the mapping. Enumeration of the six sites currently asserting
  the now-false "activation is a pure move": architecture-deferred.md:6;
  architecturedeferred_test.go:30; architecturedeferred_test.go:141-142;
  atlas/workflow/architecture-principles.md:67 and :84; atlas/workflow/sdlc-binary.md:909;
  the issue's Done-when bullet 3 (which contradicts its own bullet 4 — that tension is
  the root). Fix: move the action onto deferredVerdict, narrow or delete
  TestDeferredFileIsGuarding, and rewrite the six sites to point at the one mapping.
  Same finding, secondary: architecturedeferred_test.go:20-29 is an orphaned doc comment
  for the deleted deferredMarkers, fused without a blank line to deferredState's comment,
  so godoc renders the wrong doc for that type. atlas/workflow/architecture-principles.md
  was not touched by the fix commit and now omits both markersIn and TestDeferredFileIsGuarding.

## Round 3 — 2026-09-02T16:14:23-07:00 (claude) — passed

### Disposed

- BR-9 — addressed — Verified by probe at 5903b5d, not by the Log — activating ARCH-AUTHORITY leaves TestDeferredPrinciplesReachNoGate SKIPped green and fails exactly TestArchitectureMarkers + TestBuildPrompt_Golden; breaking the heading reds one test with an actionable message; the orphaned deferredMarkers doc comment is folded and the atlas gained markersIn and the verdict table.
- BR-5 — not-addressed — judge_test.go:133 still uses strings.Index(ArchitectureRegistry, "## "+m); architecturedeferred_test.go:165 still uses strings.Contains for the leak check. Neither is delimiter-anchored.
- BR-6 — not-addressed — cmd/sdlc/startplan_test.go:36 still hand-writes ARCH-DRY / ARCH-CONSTRAINTS and :47 ARCH-PURE; the file is untouched across the whole window.
- BR-7 — not-addressed — Re-measured at head — architecture.md 7,519 to 9,510 bytes = 1,991; the issue's Operating envelope still reads "roughly +1.4 KB per dispatch".
- BR-8 — not-addressed — go test ./... still fails only at cmd/sdlc/fleet_plan_test.go:14 on the archived plan path; pre-existing, out of window, and the reason close evidence must be package-scoped.

### Raised

- **BR-10** [Minor] `doc-enumeration-drift` The fix that single-sourced the verdict mapping restates it in the atlas, and both sites claim it is documented nowhere else
  atlas/workflow/architecture-principles.md:85-88 states the mapping "is documented there
  and nowhere else, so this page describes it rather than restating the rule" — and lines
  90-94 then restate the full rule as a three-row table. architecturedeferred_test.go:55-56
  makes the same false claim from the Go side ("enforced in exactly one place ... and
  documented nowhere else"). This is the 2nd finding in family doc-enumeration-drift (BR-3
  was atlas/workflow/sdlc-binary.md contradicting the page this diff wrote), so per the
  escalation protocol the deliverable is the rule, not this table. RULE — a doc may name a
  code-owned mapping and point at the owning symbol, or it may reproduce the mapping and
  say so, but it may not claim single-sourcing while reproducing. Prevalence in this issue:
  2 of 2 atlas pages touched. Cheapest resolution: keep the table (it is genuinely useful
  map content) and delete the two "nowhere else" claims, replacing them with "the owning
  symbol is deferredVerdict; this table mirrors it."
- **BR-11** [Minor] `tautological-derivation` Two derived marker assertions are checked against the source they were derived from, so they cannot fail
  judge_test.go:124-129 builds `want` from ArchitectureMarkers() and asserts
  strings.Contains(ArchitectureRegistry, w) — but markersIn extracted those exact
  substrings from ArchitectureRegistry, so that branch is unreachable by construction.
  archprinciples_test.go:23 has the same shape, and is additionally subsumed by the
  Contains(out, judge.ArchitectureRegistry) assertion three lines below it. The non-marker
  entries in both loops ("at-plan", "at-review", "principle:") are real and should stay.
  RULE — derive an expectation from the source, then assert it against a CONSUMER; an
  expectation asserted against its own source is a no-op. Adjacent to BR-6's observation
  that the startplan_test literals "now assert nothing the registry check doesn't".

## Open findings

- **BR-5** [Minor] `prefix-match-fragility` Entry-section lookup matches ARCH-PURE inside ARCH-PURPOSE
- **BR-6** [Minor] `marker-enumeration-sites` startplan_test.go holds a sixth hand-written marker list the class sweep missed
- **BR-7** [Minor] `operating-envelope-unstated` Declared prompt-growth budget is 1.4 KB; measured is 1,991 bytes
- **BR-8** [Minor] `stale-path-in-test` Pre-existing red test blocks any "full suite green" close evidence (out of window)
- **BR-10** [Minor] `doc-enumeration-drift` The fix that single-sourced the verdict mapping restates it in the atlas, and both sites claim it is documented nowhere else
- **BR-11** [Minor] `tautological-derivation` Two derived marker assertions are checked against the source they were derived from, so they cannot fail
